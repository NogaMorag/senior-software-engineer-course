package retry_test

import (
	"context"
	"errors"
	"math/rand/v2"
	"slices"
	"testing"
	"time"

	"example.com/ch15-resilience-lab/retry"
)

// top returns the largest value rnd may return, so the tests see each ceiling exactly.
func top(k int64) int64 { return k - 1 }

func TestBackoffCeilings(t *testing.T) {
	base, ceiling := 100*time.Millisecond, time.Second
	tests := []struct {
		n    int
		want time.Duration
	}{
		{0, 0},
		{1, 100*time.Millisecond - 1},
		{2, 200*time.Millisecond - 1},
		{3, 400*time.Millisecond - 1},
		{4, 800*time.Millisecond - 1},
		{5, time.Second - 1},  // 1.6s capped
		{80, time.Second - 1}, // no overflow
	}
	for _, tt := range tests {
		if got := retry.Backoff(tt.n, base, ceiling, top); got != tt.want {
			t.Errorf("Backoff(%d) with the top random value = %v, want %v", tt.n, got, tt.want)
		}
	}
}

func TestBackoffIsSpreadOut(t *testing.T) {
	r := rand.New(rand.NewPCG(1, 2))
	seen := map[time.Duration]bool{}
	for range 1000 {
		d := retry.Backoff(3, 100*time.Millisecond, time.Second, r.Int64N)
		if d < 0 || d >= 400*time.Millisecond {
			t.Fatalf("Backoff(3) = %v, outside [0, 400ms)", d)
		}
		seen[d/(10*time.Millisecond)] = true
	}
	if len(seen) < 30 {
		t.Fatalf("1000 draws fell into %d of 40 buckets: that is not full jitter", len(seen))
	}
}

func TestBudgetCapsRetries(t *testing.T) {
	b := retry.NewBudget(0.1, 5)
	retries := 0
	for range 1000 {
		b.Deposit()
		if b.Withdraw() {
			retries++
		}
	}
	// 5 tokens to start with, then one token per 10 calls.
	if retries < 100 || retries > 105 {
		t.Fatalf("retries allowed over 1000 calls = %d, want about 10%% (100 to 105)", retries)
	}
}

var errFlaky = errors.New("503")

func noSleep(context.Context, time.Duration) error { return nil }

func TestDo(t *testing.T) {
	permanent := errors.New("404")
	tests := []struct {
		name         string
		policy       retry.Policy
		failures     int // fn fails this many times, then succeeds
		failWith     error
		wantAttempts int
		wantErr      error
	}{
		{"success first time", retry.Policy{MaxAttempts: 3}, 0, errFlaky, 1, nil},
		{"success on the third", retry.Policy{MaxAttempts: 3}, 2, errFlaky, 3, nil},
		{"attempts run out", retry.Policy{MaxAttempts: 3}, 10, errFlaky, 3, errFlaky},
		{"not retryable", retry.Policy{MaxAttempts: 3, Retryable: func(err error) bool { return err != permanent }},
			10, permanent, 1, permanent},
		{"budget empty", retry.Policy{MaxAttempts: 3, Budget: retry.NewBudget(0.1, 0)}, 10, errFlaky, 1, retry.ErrBudgetExhausted},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.policy.Sleep = noSleep
			calls := 0
			n, err := retry.Do(context.Background(), tt.policy, func(context.Context) error {
				calls++
				if calls <= tt.failures {
					return tt.failWith
				}
				return nil
			})
			if n != tt.wantAttempts || calls != tt.wantAttempts {
				t.Errorf("attempts = %d (fn called %d times), want %d", n, calls, tt.wantAttempts)
			}
			if !errors.Is(err, tt.wantErr) || (tt.wantErr == nil && err != nil) {
				t.Errorf("err = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestDoStopsBeforeTheDeadline(t *testing.T) {
	now := time.Now() // the fake clock starts at real time so ctx itself is not already expired
	ctx, cancel := context.WithDeadline(context.Background(), now.Add(150*time.Millisecond))
	defer cancel()
	var slept []time.Duration
	p := retry.Policy{
		MaxAttempts: 10, Base: 100 * time.Millisecond, Cap: time.Second, Rand: top,
		Now:   func() time.Time { return now },
		Sleep: func(_ context.Context, d time.Duration) error { slept = append(slept, d); now = now.Add(d); return nil },
	}
	n, err := retry.Do(ctx, p, func(context.Context) error { return errFlaky })
	// Waits of ~100ms then ~200ms: the second would end after the deadline, so Do stops.
	if n != 2 || !errors.Is(err, errFlaky) || len(slept) != 1 {
		t.Fatalf("attempts = %d, err = %v, sleeps = %v; want 2 attempts and one sleep", n, err, slept)
	}
}

func TestDoRespectsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	n, _ := retry.Do(ctx, retry.Policy{MaxAttempts: 5, Sleep: noSleep}, func(context.Context) error {
		cancel() // the caller gives up during the first attempt
		return errFlaky
	})
	if n != 1 {
		t.Fatalf("attempts = %d after the caller canceled, want 1", n)
	}
}

// slowDown is a failure that asks for a wait, like a 429 or 503 with Retry-After.
type slowDown struct{ after time.Duration }

func (e slowDown) Error() string              { return "slow down" }
func (e slowDown) ServerDelay() time.Duration { return e.after }

func TestDoHonorsRetryAfter(t *testing.T) {
	tests := []struct {
		name         string
		after        time.Duration // what the server asks for on every failure
		deadline     time.Duration // from the fake now; 0 means no deadline
		budget       *retry.Budget
		wantAttempts int
		wantSlept    []time.Duration
		wantErr      error
	}{
		{"longer than the backoff: wait what the server asked", 2 * time.Second, 0, nil,
			2, []time.Duration{2 * time.Second}, nil},
		{"shorter than the backoff: the backoff wins", time.Millisecond, 0, nil,
			2, []time.Duration{100*time.Millisecond - 1}, nil},
		{"fits before the deadline", 2 * time.Second, 3 * time.Second, nil,
			2, []time.Duration{2 * time.Second}, nil},
		{"past the deadline: give up without sleeping", 10 * time.Second, 3 * time.Second, nil,
			1, nil, retry.ErrRetryAfterTooLong},
		{"budget empty: the budget still decides", 2 * time.Second, 0, retry.NewBudget(0.1, 0),
			1, nil, retry.ErrBudgetExhausted},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			now := time.Now()
			ctx := context.Background()
			if tt.deadline > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithDeadline(ctx, now.Add(tt.deadline))
				defer cancel()
			}
			var slept []time.Duration
			p := retry.Policy{
				MaxAttempts: 2, Base: 100 * time.Millisecond, Cap: time.Second, Rand: top, Budget: tt.budget,
				Now:   func() time.Time { return now },
				Sleep: func(_ context.Context, d time.Duration) error { slept = append(slept, d); now = now.Add(d); return nil },
			}
			calls := 0
			n, err := retry.Do(ctx, p, func(context.Context) error {
				calls++
				if calls == 1 {
					return slowDown{tt.after}
				}
				return nil
			})
			if n != tt.wantAttempts || !slices.Equal(slept, tt.wantSlept) {
				t.Errorf("attempts = %d, sleeps = %v; want %d attempts, sleeps %v", n, slept, tt.wantAttempts, tt.wantSlept)
			}
			if !errors.Is(err, tt.wantErr) || (tt.wantErr == nil && err != nil) {
				t.Errorf("err = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
