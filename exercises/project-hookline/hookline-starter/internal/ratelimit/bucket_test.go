package ratelimit_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"example.com/hookline/internal/ratelimit"
)

// clock is a fake time source: a test of "half a second later" moves it, not the world.
type clock struct {
	mu  sync.Mutex
	now time.Time
}

func newClock() *clock                   { return &clock{now: time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)} }
func (c *clock) Now() time.Time          { c.mu.Lock(); defer c.mu.Unlock(); return c.now }
func (c *clock) Advance(d time.Duration) { c.mu.Lock(); defer c.mu.Unlock(); c.now = c.now.Add(d) }

func TestBucket(t *testing.T) {
	type step struct {
		advance   time.Duration
		n         int
		allowed   bool
		remaining int
		retry     time.Duration
	}
	tests := []struct {
		name  string
		steps []step
	}{
		{"burst then empty", []step{
			{0, 1, true, 4, 0}, {0, 1, true, 3, 0}, {0, 1, true, 2, 0}, {0, 1, true, 1, 0}, {0, 1, true, 0, 0},
			{0, 1, false, 0, 500 * time.Millisecond}, // 2 tokens per second: the next one is 0.5s away
		}},
		{"refill", []step{
			{0, 5, true, 0, 0},
			{250 * time.Millisecond, 1, false, 0, 250 * time.Millisecond}, // half a token so far
			{250 * time.Millisecond, 1, true, 0, 0},
			{time.Second, 1, true, 1, 0}, // two more tokens, one taken
		}},
		{"refill stops at burst", []step{
			{time.Hour, 1, true, 4, 0},
		}},
		{"more than burst is never allowed", []step{
			{time.Hour, 6, false, 5, 0},
		}},
		{"retry-after for several tokens", []step{
			{0, 5, true, 0, 0},
			{0, 3, false, 0, 1500 * time.Millisecond},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newClock()
			b := ratelimit.NewBucket(2, 5, c.Now) // 2 tokens per second, room for 5
			for i, s := range tt.steps {
				c.Advance(s.advance)
				got := b.AllowN(s.n)
				want := ratelimit.Decision{Allowed: s.allowed, Remaining: s.remaining, RetryAfter: s.retry}
				if got != want {
					t.Fatalf("step %d (+%v, n=%d): got %+v, want %+v", i, s.advance, s.n, got, want)
				}
			}
		})
	}
}

func TestClockGoingBackwardsEarnsNothing(t *testing.T) {
	c := newClock()
	b := ratelimit.NewBucket(1, 1, c.Now)
	b.Allow()
	c.Advance(-time.Hour)
	if d := b.Allow(); d.Allowed {
		t.Fatalf("after the clock went back an hour: %+v, want refused", d)
	}
}

func TestLocalKeysAreIndependent(t *testing.T) {
	c := newClock()
	l := &ratelimit.Local{Now: c.Now}
	ctx := context.Background()
	for i, want := range []bool{true, true, false} {
		if d, _ := l.Allow(ctx, "tenant-a", 1, 2); d.Allowed != want {
			t.Fatalf("tenant-a request %d allowed = %v, want %v", i, d.Allowed, want)
		}
	}
	if d, _ := l.Allow(ctx, "tenant-b", 1, 2); !d.Allowed {
		t.Fatal("tenant-b was refused because tenant-a used its own bucket up")
	}
}
