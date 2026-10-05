// Package ratelimit is drill 1: a token bucket per key, in memory, with an injected clock.
// Time box: 25 minutes. The tests in ratelimit_test.go are the specification.
package ratelimit

import (
	"errors"
	"time"
)

// Decision is the answer for one request.
type Decision struct {
	Allowed    bool
	RetryAfter time.Duration // when refused: how long until one token has accrued
}

// Limiter gives every key its own bucket of burst tokens that refills at rate tokens per
// second. It must be safe for concurrent use.
type Limiter struct {
	// TODO: your fields. Hint: what does one bucket need to remember between calls?
}

// ErrConfig is returned by New for a rate or burst that cannot work.
var ErrConfig = errors.New("ratelimit: rate must be > 0 and burst >= 1")

// New returns a limiter. now is the clock: time.Now in production, a fake in tests.
func New(rate float64, burst int, now func() time.Time) (*Limiter, error) {
	panic("TODO: drill 1, see chapter 35, section 35.2")
}

// Allow takes one token from key's bucket, or refuses and says when to come back.
func (l *Limiter) Allow(key string) Decision {
	panic("TODO: drill 1, see chapter 35, section 35.2")
}

// Prune forgets every key whose bucket has refilled completely, and returns how many it
// forgot.
func (l *Limiter) Prune() int {
	panic("TODO: drill 1, see chapter 35, section 35.2")
}

// Len reports how many keys the limiter currently remembers.
func (l *Limiter) Len() int {
	panic("TODO: drill 1, see chapter 35, section 35.2")
}
