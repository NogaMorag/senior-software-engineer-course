package backoff_test

import (
	"math/rand/v2"
	"testing"
	"time"

	"pgregory.net/rapid"

	"example.com/hookline/internal/backoff"
)

// TestPolicyProperties is the property test that spec F6 asks for: for any policy and any attempt,
// the delay lies in [0, Ceiling], the ceiling never exceeds Cap, and it never shrinks as attempts grow.
func TestPolicyProperties(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		p := backoff.Policy{
			Base: time.Duration(rapid.Int64Range(1, int64(time.Hour)).Draw(t, "base")),
			Cap:  time.Duration(rapid.Int64Range(1, int64(24*time.Hour)).Draw(t, "cap")),
		}
		attempt := rapid.IntRange(-3, 1000).Draw(t, "attempt")
		seed := rapid.Uint64().Draw(t, "seed")

		c := p.Ceiling(attempt)
		if c < 0 || c > p.Cap {
			t.Fatalf("Ceiling(%d) = %v, outside [0, Cap=%v]", attempt, c, p.Cap)
		}
		if next := p.Ceiling(attempt + 1); next < c {
			t.Fatalf("Ceiling shrinks: %v at %d, %v at %d", c, attempt, next, attempt+1)
		}
		d := p.Delay(attempt, rand.New(rand.NewPCG(seed, 0)))
		if d < 0 || d > c {
			t.Fatalf("Delay(%d) = %v, outside [0, %v]", attempt, d, c)
		}
	})
}
