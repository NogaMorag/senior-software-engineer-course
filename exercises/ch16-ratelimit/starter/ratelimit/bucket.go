// Package ratelimit is a token bucket in three forms: one bucket in memory, a bucket per
// key in memory, and a bucket per key in Redis shared by every instance of a service.
package ratelimit

import (
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
	// TODO(day 48): under the lock:
	//  1. Refill lazily: if now is after b.last, add elapsed.Seconds()*rate tokens, at most
	//     burst, and move b.last to now. A clock that went backwards earns nothing.
	//  2. n above burst: refuse (Decision with Remaining only, no RetryAfter).
	//  3. Enough tokens: take n and allow. Otherwise refuse with RetryAfter = the time
	//     until n tokens exist, rounded up (math.Ceil) to a whole nanosecond.
	//  Remaining is always int(b.tokens) after the decision.
	panic("TODO(day 48): Bucket.AllowN")
}
