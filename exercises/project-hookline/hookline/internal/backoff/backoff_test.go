package backoff_test

import (
	"math/rand/v2"
	"testing"
	"time"

	"example.com/hookline/internal/backoff"
)

// fixed is a stub Rand: it returns the top of the range (max) or 0.
type fixed struct{ top bool }

func (f fixed) Int64N(n int64) int64 {
	if n <= 0 {
		panic("Int64N with n <= 0")
	}
	if f.top {
		return n - 1
	}
	return 0
}

var policy = backoff.Policy{Base: 100 * time.Millisecond, Cap: 2 * time.Second}

func TestCeiling(t *testing.T) {
	tests := []struct {
		attempt int
		want    time.Duration
	}{
		{-3, 100 * time.Millisecond}, // below 1 counts as the first retry
		{0, 100 * time.Millisecond},
		{1, 100 * time.Millisecond},
		{2, 200 * time.Millisecond},
		{3, 400 * time.Millisecond},
		{4, 800 * time.Millisecond},
		{5, 1600 * time.Millisecond},
		{6, 2 * time.Second}, // 3.2 s capped
		{7, 2 * time.Second},
		{64, 2 * time.Second}, // 2^63 would overflow without the early stop
		{1 << 30, 2 * time.Second},
	}
	for _, tt := range tests {
		if got := policy.Ceiling(tt.attempt); got != tt.want {
			t.Errorf("Ceiling(%d) = %v, want %v", tt.attempt, got, tt.want)
		}
	}
}

func TestCeilingEdgePolicies(t *testing.T) {
	tests := []struct {
		name string
		p    backoff.Policy
		want time.Duration // for attempt 3
	}{
		{"zero policy means no wait", backoff.Policy{}, 0},
		{"base above cap is clamped", backoff.Policy{Base: time.Minute, Cap: time.Second}, time.Second},
		{"cap not a power-of-two multiple", backoff.Policy{Base: 300 * time.Millisecond, Cap: time.Second}, time.Second},
		{"largest duration as cap", backoff.Policy{Base: time.Hour, Cap: time.Duration(1<<63 - 1)}, 4 * time.Hour},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.Ceiling(3); got != tt.want {
				t.Errorf("Ceiling(3) = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDelayEndsOfTheRange(t *testing.T) {
	for attempt := 1; attempt <= 10; attempt++ {
		if got := policy.Delay(attempt, fixed{top: false}); got != 0 {
			t.Errorf("attempt %d, lowest draw: Delay = %v, want 0", attempt, got)
		}
		if got, want := policy.Delay(attempt, fixed{top: true}), policy.Ceiling(attempt); got != want {
			t.Errorf("attempt %d, highest draw: Delay = %v, want the ceiling %v", attempt, got, want)
		}
	}
}

// A property checked over many seeded draws: every delay lies in [0, ceiling], the cap is
// never exceeded, and the draws really spread (full jitter, not a fixed wait).
func TestDelayStaysInBoundsAndSpreads(t *testing.T) {
	rnd := rand.New(rand.NewPCG(1, 2)) // fixed seed: the test is deterministic
	for attempt := 1; attempt <= 12; attempt++ {
		ceiling := policy.Ceiling(attempt)
		var lowHalf, highHalf int
		for range 2000 {
			d := policy.Delay(attempt, rnd)
			if d < 0 || d > ceiling || d > policy.Cap {
				t.Fatalf("attempt %d: Delay = %v outside [0, %v]", attempt, d, ceiling)
			}
			if d < ceiling/2 {
				lowHalf++
			} else {
				highHalf++
			}
		}
		// Uniform draws land in each half about 1000 times; 800 is a loose floor.
		if lowHalf < 800 || highHalf < 800 {
			t.Errorf("attempt %d: %d draws in the low half and %d in the high half; not uniform", attempt, lowHalf, highHalf)
		}
	}
}

func TestZeroPolicyNeverCallsRand(t *testing.T) {
	// fixed panics on n <= 0; a zero policy must return 0 without drawing.
	if got := (backoff.Policy{}).Delay(5, fixed{}); got != 0 {
		t.Errorf("Delay = %v, want 0", got)
	}
}
