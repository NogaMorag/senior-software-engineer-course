package hop_test

import (
	"context"
	"errors"
	"math"
	"net/http"
	"net/http/httptest"
	"slices"
	"sync/atomic"
	"testing"
	"time"

	"example.com/ch15-resilience-lab/breaker"
	"example.com/ch15-resilience-lab/flaky"
	"example.com/ch15-resilience-lab/hop"
	"example.com/ch15-resilience-lab/retry"
)

func noSleep(context.Context, time.Duration) error { return nil }

// chain starts upstream and three hops in front of it (a calls b calls c calls upstream)
// and returns a Caller for the client, the hops and the upstream.
func chain(t *testing.T, up *flaky.Server, policy func() retry.Policy) (*hop.Caller, []*hop.Hop, *flaky.Server) {
	t.Helper()
	srv := httptest.NewServer(up)
	t.Cleanup(srv.Close)
	next := srv.URL
	var hops []*hop.Hop
	for range 3 {
		h := &hop.Hop{
			Next:      &hop.Caller{URL: next, Client: srv.Client(), Retry: policy()},
			Max:       5 * time.Second,
			Reserve:   5 * time.Millisecond,
			Propagate: true,
		}
		s := httptest.NewServer(h.Handler())
		t.Cleanup(s.Close)
		hops = append([]*hop.Hop{h}, hops...)
		next = s.URL
	}
	client := &hop.Caller{URL: next, Client: srv.Client(), Retry: retry.Policy{MaxAttempts: 1}}
	return client, hops, up
}

func TestRetryStormAmplifies(t *testing.T) {
	tests := []struct {
		name         string
		policy       func() retry.Policy
		wantUpstream int64
	}{
		{"three attempts at each of three layers", func() retry.Policy {
			return retry.Policy{MaxAttempts: 3, Sleep: noSleep}
		}, 27},
		{"one attempt everywhere", func() retry.Policy {
			return retry.Policy{MaxAttempts: 1}
		}, 1},
		{"three attempts, empty budgets", func() retry.Policy {
			return retry.Policy{MaxAttempts: 3, Sleep: noSleep, Budget: retry.NewBudget(0.1, 0)}
		}, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, hops, up := chain(t, flaky.New(1.0, 0, 1), tt.policy)
			if err := client.Call(context.Background()); err == nil {
				t.Fatal("call through a dead upstream succeeded")
			}
			if got := up.Counts().Received; got != tt.wantUpstream {
				t.Errorf("upstream received %d requests for one user request, want %d (hops saw %d, %d, %d)",
					got, tt.wantUpstream, hops[0].Received(), hops[1].Received(), hops[2].Received())
			}
		})
	}
}

func TestBudgetCutsOffTheCallWhenItRunsOut(t *testing.T) {
	up := flaky.New(0, 2*time.Second, 1) // healthy but far too slow
	client, _, _ := chain(t, up, func() retry.Policy { return retry.Policy{MaxAttempts: 1} })

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	start := time.Now()
	err := client.Call(ctx)
	elapsed := time.Since(start)
	// Either the client's own deadline fired, or an inner hop (whose budget ends a few
	// milliseconds earlier because of the reserves) answered 504 first.
	var se *hop.StatusError
	if !errors.Is(err, context.DeadlineExceeded) && !(errors.As(err, &se) && se.Code == http.StatusGatewayTimeout) {
		t.Fatalf("Call = %v, want a deadline error or 504", err)
	}
	if elapsed > time.Second {
		t.Fatalf("Call took %v with a 200ms budget: the deadline did not cut it off", elapsed)
	}
	// The upstream notices that nobody waits any more and stops working.
	waitFor(t, func() bool { return up.Counts().Abandoned == 1 })
}

func TestBreakerStopsCallingADeadUpstream(t *testing.T) {
	up := flaky.New(1.0, 0, 1)
	srv := httptest.NewServer(up)
	defer srv.Close()
	c := &hop.Caller{
		URL: srv.URL, Client: srv.Client(), Retry: retry.Policy{MaxAttempts: 1},
		Breaker: breaker.New(breaker.Settings{FailureThreshold: 5, OpenTimeout: time.Hour}),
	}
	for range 100 {
		_ = c.Call(context.Background())
	}
	if got := up.Counts().Received; got != 5 {
		t.Fatalf("upstream received %d of 100 calls, want 5: the breaker opens after 5 failures", got)
	}
	if c.Refused() != 95 {
		t.Fatalf("refused = %d, want 95", c.Refused())
	}
}

func TestStatusCodes(t *testing.T) {
	for _, code := range []int{200, 404, 429, 500, 503} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(code) }))
		c := &hop.Caller{URL: srv.URL, Client: srv.Client()}
		err := c.Call(context.Background())
		var se *hop.StatusError
		isFailure := errors.As(err, &se)
		if want := code >= 500 || code == 429; isFailure != want {
			t.Errorf("status %d: failure = %v (err %v), want %v", code, isFailure, err, want)
		}
		srv.Close()
	}
}

func waitFor(t *testing.T, cond func() bool) {
	t.Helper()
	for range 200 {
		if cond() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("condition not reached within 2s")
}

func TestParseRetryAfter(t *testing.T) {
	now := time.Date(1999, time.December, 31, 23, 59, 29, 0, time.UTC)
	tests := []struct {
		value  string
		want   time.Duration
		wantOK bool
	}{
		{"120", 120 * time.Second, true},
		{"0", 0, true},
		{" 5 ", 5 * time.Second, true},
		{"Fri, 31 Dec 1999 23:59:59 GMT", 30 * time.Second, true},  // IMF-fixdate, the form servers send
		{"Friday, 31-Dec-99 23:59:59 GMT", 30 * time.Second, true}, // obsolete RFC 850 form
		{"Fri Dec 31 23:59:59 1999", 30 * time.Second, true},       // obsolete asctime form
		{"Fri, 31 Dec 1999 23:00:00 GMT", 0, true},                 // already past: no wait
		{"99999999999999999999", math.MaxInt64, true},              // overflows: as long as possible
		{"", 0, false},
		{"-5", 0, false},
		{"+5", 0, false},
		{"1.5", 0, false},
		{"soon", 0, false},
	}
	for _, tt := range tests {
		got, ok := hop.ParseRetryAfter(tt.value, now)
		if got != tt.want || ok != tt.wantOK {
			t.Errorf("ParseRetryAfter(%q) = %v, %v; want %v, %v", tt.value, got, ok, tt.want, tt.wantOK)
		}
	}
}

func TestCallHonorsRetryAfter(t *testing.T) {
	start := time.Now().Truncate(time.Second) // whole seconds, like an HTTP-date
	tests := []struct {
		name         string
		code         int
		retryAfter   string // "" sends no header
		deadline     time.Duration
		budget       *retry.Budget
		wantRequests int64
		wantSlept    []time.Duration
		wantErr      error
	}{
		{"429 with seconds", 429, "2", 0, nil, 2, []time.Duration{2 * time.Second}, nil},
		{"503 with an HTTP-date", 503, start.Add(3 * time.Second).UTC().Format(http.TimeFormat), 0, nil,
			2, []time.Duration{3 * time.Second}, nil},
		{"429 without the header: plain backoff", 429, "", 0, nil, 2, []time.Duration{time.Millisecond - 1}, nil},
		{"500 ignores Retry-After", 500, "2", 0, nil, 2, []time.Duration{time.Millisecond - 1}, nil},
		{"429 asking past the deadline", 429, "10", 5 * time.Second, nil, 1, nil, retry.ErrRetryAfterTooLong},
		{"budget empty: no retry despite Retry-After", 429, "2", 0, retry.NewBudget(0.1, 0),
			1, nil, retry.ErrBudgetExhausted},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var requests atomic.Int64
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if requests.Add(1) > 1 {
					return // 200
				}
				if tt.retryAfter != "" {
					w.Header().Set("Retry-After", tt.retryAfter)
				}
				w.WriteHeader(tt.code)
			}))
			defer srv.Close()
			now := start
			var slept []time.Duration
			c := &hop.Caller{URL: srv.URL, Client: srv.Client(), Retry: retry.Policy{
				MaxAttempts: 3, Base: time.Millisecond, Cap: time.Millisecond, Rand: top, Budget: tt.budget,
				Now:   func() time.Time { return now },
				Sleep: func(_ context.Context, d time.Duration) error { slept = append(slept, d); now = now.Add(d); return nil },
			}}
			ctx := context.Background()
			if tt.deadline > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithDeadline(ctx, start.Add(tt.deadline))
				defer cancel()
			}
			err := c.Call(ctx)
			if got := requests.Load(); got != tt.wantRequests || !slices.Equal(slept, tt.wantSlept) {
				t.Errorf("requests = %d, sleeps = %v; want %d, %v", got, slept, tt.wantRequests, tt.wantSlept)
			}
			if !errors.Is(err, tt.wantErr) || (tt.wantErr == nil && err != nil) {
				t.Errorf("err = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func top(k int64) int64 { return k - 1 }
