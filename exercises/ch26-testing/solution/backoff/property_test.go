package backoff_test

import (
	"math"
	"math/rand/v2"
	"testing"
	"time"

	"pgregory.net/rapid"

	"example.com/ch26-testing/backoff"
)

// drawnRand is a backoff.Rand whose every draw comes from rapid. rapid then explores the
// edges of the range (0 and n-1) on purpose, and shrinks a failure to the smallest draw.
type drawnRand struct{ t *rapid.T }

func (r drawnRand) Int64N(n int64) int64 {
	return rapid.Int64Range(0, n-1).Draw(r.t, "draw")
}

// policyGen generates any valid policy: positive Base and Cap, in either order, up to the
// largest Duration. Base above Cap is legal; the policy clamps it.
var policyGen = rapid.Custom(func(t *rapid.T) backoff.Policy {
	return backoff.Policy{
		Base: time.Duration(rapid.Int64Range(1, math.MaxInt64).Draw(t, "base")),
		Cap:  time.Duration(rapid.Int64Range(1, math.MaxInt64).Draw(t, "cap")),
	}
})

// attemptGen covers the attempts that matter: below 1, the doubling range, and far beyond it.
var attemptGen = rapid.OneOf(rapid.IntRange(-5, 70), rapid.IntRange(70, math.MaxInt32))

// Bounds: every delay lies in [0, Ceiling(attempt)], and no ceiling exceeds Cap.
func TestPropertyDelayWithinBounds(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		p := policyGen.Draw(t, "policy")
		attempt := attemptGen.Draw(t, "attempt")
		ceiling := p.Ceiling(attempt)
		if ceiling < min(p.Base, p.Cap) || ceiling > p.Cap {
			t.Fatalf("Ceiling(%d) = %v, want within [%v, %v]", attempt, ceiling, min(p.Base, p.Cap), p.Cap)
		}
		if d := p.Delay(attempt, drawnRand{t}); d < 0 || d > ceiling {
			t.Fatalf("Delay(%d) = %v outside [0, %v]", attempt, d, ceiling)
		}
	})
}

// Monotone up to the cap: the next ceiling is exactly double until it reaches Cap, and once it
// is Cap it stays Cap. Together these say "exponential, then flat", with no overflow wrap-around.
func TestPropertyCeilingDoublesThenStaysAtCap(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		p := policyGen.Draw(t, "policy")
		attempt := rapid.IntRange(1, 200).Draw(t, "attempt")
		cur, next := p.Ceiling(attempt), p.Ceiling(attempt+1)
		switch {
		case next < cur:
			t.Fatalf("Ceiling(%d) = %v < Ceiling(%d) = %v: not monotone", attempt+1, next, attempt, cur)
		case cur == p.Cap && next != p.Cap:
			t.Fatalf("Ceiling left the cap: %v then %v", cur, next)
		case next < p.Cap && next != 2*cur:
			t.Fatalf("below the cap the ceiling must double: %v then %v", cur, next)
		}
	})
}

// Distribution sanity, not a statistics course: 4000 full-jitter draws from a seeded source
// average close to Ceiling/2, and both halves of the range get draws. The mean of n uniform
// draws on [0, c] has a standard deviation of c/sqrt(12n), about 0.0046c here; the 0.05c
// tolerance is over ten of those, so a correct Delay never fails this by chance.
func TestPropertyDelayIsSpreadOverTheRange(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		p := backoff.Policy{
			Base: time.Duration(rapid.Int64Range(int64(time.Millisecond), int64(time.Second)).Draw(t, "base")),
			Cap:  time.Duration(rapid.Int64Range(int64(time.Second), int64(time.Hour)).Draw(t, "cap")),
		}
		attempt := rapid.IntRange(1, 30).Draw(t, "attempt")
		rnd := rand.New(rand.NewPCG(rapid.Uint64().Draw(t, "seed"), 1))
		c := float64(p.Ceiling(attempt))

		const n = 4000
		var sum float64
		var low int
		for range n {
			d := float64(p.Delay(attempt, rnd))
			sum += d
			if d < c/2 {
				low++
			}
		}
		if mean := sum / n; math.Abs(mean-c/2) > 0.05*c {
			t.Fatalf("mean delay %.0f, want about %.0f (ceiling %.0f)", mean, c/2, c)
		}
		if low < n/4 || low > 3*n/4 {
			t.Fatalf("%d of %d draws in the low half; full jitter should give about half", low, n)
		}
	})
}
