package ratelimit_test

import (
	"math/rand/v2"
	"testing"
	"time"

	"golang.org/x/time/rate"

	"example.com/hookline/internal/ratelimit"
)

// TestSameAnswersAsXTimeRate feeds one timeline to our bucket and to the standard
// extended library's limiter. AllowN takes the time as an argument there, so the
// comparison needs no sleeping either.
func TestSameAnswersAsXTimeRate(t *testing.T) {
	c := newClock()
	ours := ratelimit.NewBucket(10, 20, c.Now)
	theirs := rate.NewLimiter(rate.Limit(10), 20)
	rnd := rand.New(rand.NewPCG(7, 7))
	allowed := 0
	for i := range 5000 {
		// Bursts and pauses: mostly 0-40ms apart, sometimes a 2s pause.
		gap := time.Duration(rnd.IntN(40_000)) * time.Microsecond
		if rnd.IntN(100) == 0 {
			gap = 2 * time.Second
		}
		c.Advance(gap)
		got := ours.Allow().Allowed
		want := theirs.AllowN(c.Now(), 1)
		if got != want {
			t.Fatalf("event %d at +%v: ours allowed=%v, x/time/rate allowed=%v", i, gap, got, want)
		}
		if got {
			allowed++
		}
	}
	if allowed == 0 || allowed == 5000 {
		t.Fatalf("allowed %d of 5000: the timeline never exercised the limit", allowed)
	}
}
