package ratelimit

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"example.com/ch35-drills/clock"
)

var t0 = time.Date(2026, 10, 5, 9, 0, 0, 0, time.UTC)

func newLimiter(t *testing.T, rate float64, burst int) (*Limiter, *clock.Fake) {
	t.Helper()
	clk := clock.NewFake(t0)
	l, err := New(rate, burst, clk.Now)
	if err != nil {
		t.Fatalf("New(%v, %d): %v", rate, burst, err)
	}
	return l, clk
}

func TestNewRejectsBadConfig(t *testing.T) {
	cases := []struct {
		name  string
		rate  float64
		burst int
		now   func() time.Time
	}{
		{"zero rate", 0, 1, time.Now},
		{"negative rate", -1, 1, time.Now},
		{"NaN rate", math.NaN(), 1, time.Now},
		{"infinite rate", math.Inf(1), 1, time.Now},
		{"zero burst", 1, 0, time.Now},
		{"nil clock", 1, 1, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := New(tc.rate, tc.burst, tc.now); !errors.Is(err, ErrConfig) {
				t.Fatalf("err = %v, want ErrConfig", err)
			}
		})
	}
}

func TestBurstThenRefuse(t *testing.T) {
	l, _ := newLimiter(t, 2, 3)
	for i := range 3 {
		if d := l.Allow("tenant-a"); !d.Allowed {
			t.Fatalf("request %d refused inside the burst", i+1)
		}
	}
	d := l.Allow("tenant-a")
	if d.Allowed {
		t.Fatal("fourth request allowed with burst 3 and no time passing")
	}
	if d.RetryAfter != 500*time.Millisecond {
		t.Fatalf("RetryAfter = %v, want 500ms (one token at 2 per second)", d.RetryAfter)
	}
}

func TestRefillOverTime(t *testing.T) {
	cases := []struct {
		name      string
		elapsed   time.Duration
		wantAllow int
	}{
		{"nothing earned yet", 499 * time.Millisecond, 0},
		{"one token at exactly 500ms", 500 * time.Millisecond, 1},
		{"two tokens after one second", time.Second, 2},
		{"never more than burst", time.Hour, 3},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			l, clk := newLimiter(t, 2, 3)
			for range 3 {
				l.Allow("k")
			}
			clk.Advance(tc.elapsed)
			got := 0
			for range 10 {
				if l.Allow("k").Allowed {
					got++
				}
			}
			if got != tc.wantAllow {
				t.Fatalf("allowed %d after %v, want %d", got, tc.elapsed, tc.wantAllow)
			}
		})
	}
}

func TestRetryAfterShrinksWithPartialToken(t *testing.T) {
	l, clk := newLimiter(t, 2, 1)
	l.Allow("k")
	clk.Advance(250 * time.Millisecond) // half a token earned
	d := l.Allow("k")
	if d.Allowed || d.RetryAfter != 250*time.Millisecond {
		t.Fatalf("got %+v, want refused with RetryAfter 250ms", d)
	}
	clk.Advance(d.RetryAfter)
	if !l.Allow("k").Allowed {
		t.Fatal("refused after waiting exactly RetryAfter")
	}
}

func TestKeysAreIndependent(t *testing.T) {
	l, _ := newLimiter(t, 1, 1)
	if !l.Allow("a").Allowed {
		t.Fatal("a refused")
	}
	if l.Allow("a").Allowed {
		t.Fatal("a allowed twice with burst 1")
	}
	if !l.Allow("b").Allowed {
		t.Fatal("b refused because a used its token")
	}
}

func TestClockGoingBackwardsEarnsNothing(t *testing.T) {
	l, clk := newLimiter(t, 1, 1)
	l.Allow("k")
	clk.Advance(-time.Hour)
	if l.Allow("k").Allowed {
		t.Fatal("allowed after the clock went back")
	}
	clk.Advance(time.Hour) // back to the moment of the first request: still nothing earned
	if l.Allow("k").Allowed {
		t.Fatal("allowed after returning to the same instant")
	}
	clk.Advance(time.Second)
	if !l.Allow("k").Allowed {
		t.Fatal("refused one second after the last real request")
	}
}

func TestConcurrentCallersShareOneBucket(t *testing.T) {
	l, _ := newLimiter(t, 1, 50)
	var allowed atomic.Int64
	var wg sync.WaitGroup
	for range 100 {
		wg.Go(func() {
			for range 10 {
				if l.Allow("shared").Allowed {
					allowed.Add(1)
				}
			}
		})
	}
	wg.Wait()
	if got := allowed.Load(); got != 50 {
		t.Fatalf("allowed %d of 1000 concurrent requests, want exactly the burst, 50", got)
	}
}

func TestPruneForgetsOnlyFullBuckets(t *testing.T) {
	l, clk := newLimiter(t, 1, 2)
	for i := range 5 {
		l.Allow(fmt.Sprintf("idle-%d", i))
	}
	clk.Advance(time.Second) // idle keys have 2 of 2 tokens again
	l.Allow("busy")
	l.Allow("busy") // busy has 0 of 2
	if n := l.Prune(); n != 5 {
		t.Fatalf("Prune removed %d keys, want the 5 idle ones", n)
	}
	if n := l.Len(); n != 1 {
		t.Fatalf("Len = %d after Prune, want 1 (busy)", n)
	}
	if l.Allow("busy").Allowed {
		t.Fatal("Prune reset the busy key")
	}
	if !l.Allow("idle-0").Allowed {
		t.Fatal("a pruned key must behave like a new key: full")
	}
}
