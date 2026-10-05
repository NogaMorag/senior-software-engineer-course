package backoff_test

import (
	"testing"
)

// Property tests for the backoff policy with pgregory.net/rapid (day 80, step 7).
// Replace each t.Fatal with a rapid.Check whose generators cover every valid policy and attempt.

// Bounds: every delay lies in [0, Ceiling(attempt)], and no ceiling exceeds Cap.
func TestPropertyDelayWithinBounds(t *testing.T) {
	t.Fatal("TODO: write the bounds property; README day 80, step 7")
}

// Monotone up to the cap: the ceiling doubles until it reaches Cap, then stays there.
func TestPropertyCeilingDoublesThenStaysAtCap(t *testing.T) {
	t.Fatal("TODO: write the monotone-cap property; README day 80, step 7")
}

// Distribution sanity: many seeded draws average about Ceiling/2 and fill both halves.
func TestPropertyDelayIsSpreadOverTheRange(t *testing.T) {
	t.Fatal("TODO: write the distribution property; README day 80, step 7")
}
