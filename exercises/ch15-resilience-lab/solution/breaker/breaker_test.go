package breaker_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"example.com/ch15-resilience-lab/breaker"
)

// clock is a fake time source: tests move it forward by hand, so nothing sleeps.
type clock struct {
	mu  sync.Mutex
	now time.Time
}

func (c *clock) Now() time.Time      { c.mu.Lock(); defer c.mu.Unlock(); return c.now }
func (c *clock) Add(d time.Duration) { c.mu.Lock(); defer c.mu.Unlock(); c.now = c.now.Add(d) }

var errDown = errors.New("upstream down")

func newBreaker(t *testing.T) (*breaker.Breaker, *clock, *[]string) {
	t.Helper()
	c := &clock{now: time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)}
	var changes []string
	b := breaker.New(breaker.Settings{
		FailureThreshold: 3,
		OpenTimeout:      10 * time.Second,
		HalfOpenProbes:   2,
		Now:              c.Now,
		OnStateChange: func(from, to breaker.State) {
			changes = append(changes, from.String()+">"+to.String())
		},
	})
	return b, c, &changes
}

func fail() error    { return errDown }
func succeed() error { return nil }

func trip(t *testing.T, b *breaker.Breaker) {
	t.Helper()
	for range 3 {
		if err := b.Execute(fail); !errors.Is(err, errDown) {
			t.Fatalf("closed breaker returned %v, want the call's own error", err)
		}
	}
	if got := b.State(); got != breaker.Open {
		t.Fatalf("after 3 failures state = %v, want open", got)
	}
}

func TestClosedCountsConsecutiveFailures(t *testing.T) {
	b, _, _ := newBreaker(t)
	// Two failures, a success, two failures: never three in a row.
	for _, fn := range []func() error{fail, fail, succeed, fail, fail} {
		_ = b.Execute(fn)
	}
	if got := b.State(); got != breaker.Closed {
		t.Fatalf("state = %v, want closed: a success resets the count", got)
	}
}

func TestOpenFailsFastWithoutCalling(t *testing.T) {
	b, _, _ := newBreaker(t)
	trip(t, b)
	called := false
	err := b.Execute(func() error { called = true; return nil })
	if !errors.Is(err, breaker.ErrOpen) || called {
		t.Fatalf("open breaker: err = %v, called = %v; want ErrOpen and no call", err, called)
	}
}

func TestStateTransitions(t *testing.T) {
	tests := []struct {
		name      string
		probes    []func() error
		wantState breaker.State
		wantLog   []string
	}{
		{"two successful probes close it", []func() error{succeed, succeed}, breaker.Closed,
			[]string{"closed>open", "open>half-open", "half-open>closed"}},
		{"a failed probe opens it again", []func() error{fail}, breaker.Open,
			[]string{"closed>open", "open>half-open", "half-open>open"}},
		{"one success is not enough", []func() error{succeed}, breaker.HalfOpen,
			[]string{"closed>open", "open>half-open"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, c, log := newBreaker(t)
			trip(t, b)
			c.Add(9 * time.Second)
			if got := b.State(); got != breaker.Open {
				t.Fatalf("9s after opening state = %v, want still open", got)
			}
			c.Add(time.Second)
			if got := b.State(); got != breaker.HalfOpen {
				t.Fatalf("10s after opening state = %v, want half-open", got)
			}
			for _, p := range tt.probes {
				_ = b.Execute(p)
			}
			if got := b.State(); got != tt.wantState {
				t.Errorf("state = %v, want %v", got, tt.wantState)
			}
			if len(*log) != len(tt.wantLog) {
				t.Fatalf("transitions = %v, want %v", *log, tt.wantLog)
			}
			for i := range tt.wantLog {
				if (*log)[i] != tt.wantLog[i] {
					t.Errorf("transitions = %v, want %v", *log, tt.wantLog)
					break
				}
			}
		})
	}
}

func TestHalfOpenLimitsProbes(t *testing.T) {
	b, c, _ := newBreaker(t)
	trip(t, b)
	c.Add(10 * time.Second)
	// Two slow probes are in flight; a third caller is turned away.
	done1, err1 := b.Allow()
	done2, err2 := b.Allow()
	_, err3 := b.Allow()
	if err1 != nil || err2 != nil || !errors.Is(err3, breaker.ErrTooManyProbes) {
		t.Fatalf("Allow errors = %v, %v, %v; want nil, nil, ErrTooManyProbes", err1, err2, err3)
	}
	done1(nil)
	done2(nil)
	if got := b.State(); got != breaker.Closed {
		t.Fatalf("state = %v, want closed after two successful probes", got)
	}
}

func TestStaleResultIsIgnored(t *testing.T) {
	b, c, _ := newBreaker(t)
	// A slow call starts while closed ...
	slowDone, err := b.Allow()
	if err != nil {
		t.Fatal(err)
	}
	trip(t, b) // ... meanwhile three fast failures open the breaker ...
	c.Add(10 * time.Second)
	if got := b.State(); got != breaker.HalfOpen {
		t.Fatalf("state = %v, want half-open", got)
	}
	slowDone(errDown) // ... and the slow call's failure arrives late.
	if got := b.State(); got != breaker.HalfOpen {
		t.Fatalf("state = %v, want half-open: a result from an older generation must not count", got)
	}
}

func TestCanceledCallsDoNotCount(t *testing.T) {
	b, _, _ := newBreaker(t)
	for range 5 {
		_ = b.Execute(func() error { return context.Canceled })
	}
	if got := b.State(); got != breaker.Closed {
		t.Fatalf("state = %v, want closed: our caller giving up is not the dependency failing", got)
	}
}

func TestPanicCountsAsFailure(t *testing.T) {
	b, _, _ := newBreaker(t)
	for range 3 {
		func() {
			defer func() { _ = recover() }()
			_ = b.Execute(func() error { panic("boom") })
		}()
	}
	if got := b.State(); got != breaker.Open {
		t.Fatalf("state = %v, want open after three panics", got)
	}
}

func TestConcurrentUse(t *testing.T) {
	b, c, _ := newBreaker(t)
	var wg sync.WaitGroup
	for i := range 50 {
		wg.Go(func() {
			if i%2 == 0 {
				_ = b.Execute(fail)
			} else {
				_ = b.Execute(succeed)
			}
			c.Add(time.Second)
			_ = b.State()
		})
	}
	wg.Wait() // run with -race: the point is that nothing races
}
