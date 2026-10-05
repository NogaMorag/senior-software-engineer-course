// Package backoff computes retry delays: exponential growth, a cap, and full jitter.
// It is a pure function of its inputs; the randomness is injected so that tests can pin it.
package backoff

import (
	"time"
)

// Rand is the one method of *math/rand/v2.Rand that Delay needs. Production code passes
// Global; tests pass a seeded *rand.Rand or a stub.
type Rand interface {
	// Int64N returns a value in [0, n). It panics if n <= 0.
	Int64N(n int64) int64
}

// Global draws from math/rand/v2's top-level source, which is safe for concurrent use
// (a *rand.Rand is not).
var Global Rand = global{}

type global struct{}

func (global) Int64N(n int64) int64 { panic("TODO: see chapter 38") }

// Policy describes the delays. Base is the ceiling of the first retry's delay; each
// later attempt doubles the ceiling until it reaches Cap.
type Policy struct {
	Base time.Duration
	Cap  time.Duration
}

// Ceiling is the largest delay attempt may get: min(Cap, Base * 2^(attempt-1)).
// Attempt 1 is the first retry. Attempts below 1 count as 1. The doubling stops as soon
// as it reaches Cap, so a huge attempt number cannot overflow.
func (p Policy) Ceiling(attempt int) time.Duration {
	panic("TODO: see chapter 38")

}

// Delay is "full jitter": a uniformly random duration in [0, Ceiling(attempt)]. Spreading
// retries over the whole interval, instead of waiting exactly the ceiling, keeps many
// clients that failed together from retrying together.
func (p Policy) Delay(attempt int, rnd Rand) time.Duration {
	panic("TODO: see chapter 38")

}
