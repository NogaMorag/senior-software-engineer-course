// Package retry retries a failed call with capped exponential backoff and full jitter,
// inside the caller's deadline and, optionally, inside a shared retry budget.
package retry

import (
	"context"
	"errors"
	"math/rand/v2"
	"sync"
	"time"
)

// ErrBudgetExhausted wraps the last error when the budget refused a retry.
var ErrBudgetExhausted = errors.New("retry: budget exhausted")

// Backoff returns the full-jitter delay before retry number n (n = 1 is the first retry):
// a uniform random duration in [0, min(cap, base*2^(n-1))). rnd(k) returns a value in [0, k).
func Backoff(n int, base, cap time.Duration, rnd func(int64) int64) time.Duration {
	// TODO(day 46): full jitter. Return 0 for n < 1 or non-positive base or cap.
	// The ceiling is min(cap, base<<(n-1)); guard the shift so a large n cannot overflow
	// (n-1 >= 62, or a shifted value <= 0, means "use cap"). Return time.Duration(rnd(ceiling)).
	panic("TODO(day 46): retry.Backoff")
}

// Budget is a token bucket for retries. Every call deposits Ratio tokens, every retry
// withdraws one, and the balance never exceeds Max. With Ratio 0.1, retries settle at
// about 10% of calls, however many layers of callers are failing at once.
type Budget struct {
	mu     sync.Mutex
	ratio  float64
	max    float64
	tokens float64
}

// NewBudget returns a full budget.
func NewBudget(ratio, max float64) *Budget {
	return &Budget{ratio: ratio, max: max, tokens: max}
}

// Deposit records one call (not a retry).
func (b *Budget) Deposit() {
	// TODO(day 47): under the lock, tokens = min(max, tokens+ratio).
	panic("TODO(day 47): Budget.Deposit")
}

// Withdraw takes one token for a retry and reports whether there was one.
func (b *Budget) Withdraw() bool {
	// TODO(day 47): under the lock, take one token if there is a whole one.
	panic("TODO(day 47): Budget.Withdraw")
}

// Tokens reports the current balance.
func (b *Budget) Tokens() float64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.tokens
}

// Policy configures Do. Zero values: one attempt, no delay, no budget.
type Policy struct {
	MaxAttempts int           // attempts including the first; 3 means up to two retries
	Base, Cap   time.Duration // backoff base and ceiling
	Budget      *Budget       // shared by every caller of one dependency; nil means unlimited
	// Retryable reports whether an error is worth another attempt (default: every error).
	Retryable func(error) bool
	Rand      func(int64) int64                          // default rand.Int64N
	Sleep     func(context.Context, time.Duration) error // default: a timer that respects ctx
	Now       func() time.Time                           // default time.Now; used against ctx's deadline
}

// Do calls fn until it succeeds, the error is not retryable, attempts run out, the next
// wait would cross ctx's deadline, or the budget refuses. It returns the number of
// attempts made and the last error.
func Do(ctx context.Context, p Policy, fn func(context.Context) error) (int, error) {
	// TODO(day 46):
	//  p = p.withDefaults(); if there is a budget, Deposit once (this is one call).
	//  Loop over attempts starting at 1:
	//   - err := fn(ctx); success returns (attempt, nil).
	//   - Stop and return (attempt, err) when attempts are used up, the error is not
	//     retryable, or ctx is already done.
	//   - wait := Backoff(attempt, ...). If ctx has a deadline and p.Now().Add(wait) is after
	//     it, stop: sleeping would only burn the time the caller has left.
	//   - With a budget, Withdraw; on refusal return fmt.Errorf("%w: %w", ErrBudgetExhausted, err).
	//   - p.Sleep(ctx, wait); if it fails, return (attempt, err).
	panic("TODO(day 46): retry.Do")
}

func (p Policy) withDefaults() Policy {
	if p.MaxAttempts < 1 {
		p.MaxAttempts = 1
	}
	if p.Retryable == nil {
		p.Retryable = func(error) bool { return true }
	}
	if p.Rand == nil {
		p.Rand = rand.Int64N
	}
	if p.Sleep == nil {
		p.Sleep = sleep
	}
	if p.Now == nil {
		p.Now = time.Now
	}
	return p
}

func sleep(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return ctx.Err()
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
