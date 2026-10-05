// Package retry retries a failed call with capped exponential backoff and full jitter,
// inside the caller's deadline and, optionally, inside a shared retry budget.
package retry

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"sync"
	"time"
)

// ErrBudgetExhausted wraps the last error when the budget refused a retry.
var ErrBudgetExhausted = errors.New("retry: budget exhausted")

// Backoff returns the full-jitter delay before retry number n (n = 1 is the first retry):
// a uniform random duration in [0, min(cap, base*2^(n-1))). rnd(k) returns a value in [0, k).
func Backoff(n int, base, cap time.Duration, rnd func(int64) int64) time.Duration {
	if n < 1 || base <= 0 || cap <= 0 {
		return 0
	}
	ceil := cap
	if n-1 < 62 { // past 2^62 the shift overflows; the cap has long been reached
		if d := base << (n - 1); d > 0 && d < cap {
			ceil = d
		}
	}
	return time.Duration(rnd(int64(ceil)))
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
	b.mu.Lock()
	defer b.mu.Unlock()
	b.tokens = min(b.max, b.tokens+b.ratio)
}

// Withdraw takes one token for a retry and reports whether there was one.
func (b *Budget) Withdraw() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
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
	p = p.withDefaults()
	if p.Budget != nil {
		p.Budget.Deposit()
	}
	for attempt := 1; ; attempt++ {
		err := fn(ctx)
		if err == nil {
			return attempt, nil
		}
		if attempt >= p.MaxAttempts || !p.Retryable(err) || ctx.Err() != nil {
			return attempt, err
		}
		wait := Backoff(attempt, p.Base, p.Cap, p.Rand)
		if dl, ok := ctx.Deadline(); ok && p.Now().Add(wait).After(dl) {
			return attempt, err // sleeping would spend the budget the caller has left
		}
		if p.Budget != nil && !p.Budget.Withdraw() {
			return attempt, fmt.Errorf("%w: %w", ErrBudgetExhausted, err)
		}
		if serr := p.Sleep(ctx, wait); serr != nil {
			return attempt, err
		}
	}
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
