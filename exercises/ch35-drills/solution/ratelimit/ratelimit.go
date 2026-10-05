// Package ratelimit is drill 1: a token bucket per key, in memory, with an injected clock.
package ratelimit

import (
	"errors"
	"math"
	"sync"
	"time"
)

// Decision is the answer for one request.
type Decision struct {
	Allowed    bool
	RetryAfter time.Duration // when refused: how long until one token has accrued
}

// Limiter gives every key its own bucket of Burst tokens that refills at Rate tokens per
// second. It is safe for concurrent use.
type Limiter struct {
	rate  float64
	burst float64
	now   func() time.Time

	mu      sync.Mutex
	buckets map[string]*bucket
}

type bucket struct {
	tokens float64
	last   time.Time
}

// ErrConfig is returned by New for a rate or burst that cannot work.
var ErrConfig = errors.New("ratelimit: rate must be > 0 and burst >= 1")

// New returns a limiter. now is the clock: time.Now in production, a fake in tests.
func New(rate float64, burst int, now func() time.Time) (*Limiter, error) {
	if rate <= 0 || math.IsNaN(rate) || math.IsInf(rate, 0) || burst < 1 || now == nil {
		return nil, ErrConfig
	}
	return &Limiter{rate: rate, burst: float64(burst), now: now, buckets: make(map[string]*bucket)}, nil
}

// Allow takes one token from key's bucket, or refuses and says when to come back.
func (l *Limiter) Allow(key string) Decision {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	b, ok := l.buckets[key]
	if !ok {
		// A new key starts full: a client that has never been seen gets its whole burst.
		b = &bucket{tokens: l.burst, last: now}
		l.buckets[key] = b
	}
	l.refill(b, now)
	if b.tokens >= 1 {
		b.tokens--
		return Decision{Allowed: true}
	}
	wait := time.Duration(math.Ceil((1 - b.tokens) / l.rate * float64(time.Second)))
	return Decision{RetryAfter: wait}
}

// refill adds the tokens earned since the bucket was last touched. There is no timer:
// the bucket is brought up to date only when someone asks. A clock that went backwards
// earns nothing and does not move last, so it cannot be used to earn tokens twice.
func (l *Limiter) refill(b *bucket, now time.Time) {
	if elapsed := now.Sub(b.last); elapsed > 0 {
		b.tokens = math.Min(l.burst, b.tokens+elapsed.Seconds()*l.rate)
		b.last = now
	}
}

// Prune forgets every key whose bucket has refilled completely, and returns how many it
// forgot. A full bucket behaves exactly like a key that was never seen, so forgetting it
// changes no decision; it only bounds memory to the keys that were active recently.
func (l *Limiter) Prune() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	n := 0
	for key, b := range l.buckets {
		l.refill(b, now)
		if b.tokens >= l.burst {
			delete(l.buckets, key)
			n++
		}
	}
	return n
}

// Len reports how many keys the limiter currently remembers.
func (l *Limiter) Len() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.buckets)
}
