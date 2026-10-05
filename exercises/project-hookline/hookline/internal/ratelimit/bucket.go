// Package ratelimit is a token bucket in three forms: one bucket in memory, a bucket per
// key in memory, and a bucket per key in Redis shared by every instance of a service.
package ratelimit

import (
	"math"
	"sync"
	"time"
)

// Decision is the answer for one request.
type Decision struct {
	Allowed    bool
	Remaining  int           // whole tokens left after this decision
	RetryAfter time.Duration // when refused: how long until enough tokens have accrued
}

// Bucket holds up to burst tokens and gains rate tokens per second. A request takes
// tokens; with too few, it is refused. Bucket is safe for concurrent use.
type Bucket struct {
	rate  float64 // tokens per second
	burst float64
	now   func() time.Time

	mu     sync.Mutex
	tokens float64
	last   time.Time
}

// NewBucket returns a full bucket. now is the clock; pass time.Now in production and a
// fake clock in tests, so that a test of "one minute later" takes no time at all.
func NewBucket(rate float64, burst int, now func() time.Time) *Bucket {
	return &Bucket{rate: rate, burst: float64(burst), now: now, tokens: float64(burst), last: now()}
}

// Allow asks for one token.
func (b *Bucket) Allow() Decision { return b.AllowN(1) }

// AllowN asks for n tokens. A request for more than burst is never allowed.
func (b *Bucket) AllowN(n int) Decision {
	b.mu.Lock()
	defer b.mu.Unlock()
	now := b.now()
	// Refill lazily: no timer, just the tokens earned since the last call. A clock that
	// went backwards earns nothing.
	if elapsed := now.Sub(b.last); elapsed > 0 {
		b.tokens = math.Min(b.burst, b.tokens+elapsed.Seconds()*b.rate)
		b.last = now
	}
	need := float64(n)
	if need > b.burst {
		return Decision{Remaining: int(b.tokens)}
	}
	if b.tokens >= need {
		b.tokens -= need
		return Decision{Allowed: true, Remaining: int(b.tokens)}
	}
	wait := time.Duration(math.Ceil((need - b.tokens) / b.rate * float64(time.Second)))
	return Decision{Remaining: int(b.tokens), RetryAfter: wait}
}
