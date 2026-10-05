package ttlcache

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"example.com/ch35-drills/clock"
)

var t0 = time.Date(2026, 10, 5, 9, 0, 0, 0, time.UTC)

func newCache(t *testing.T, capacity int, ttl time.Duration) (*Cache[string, int], *clock.Fake) {
	t.Helper()
	clk := clock.NewFake(t0)
	c, err := New[string, int](capacity, ttl, clk.Now)
	if err != nil {
		t.Fatal(err)
	}
	return c, clk
}

func mustGet(t *testing.T, c *Cache[string, int], key string, want int) {
	t.Helper()
	got, ok := c.Get(key)
	if !ok || got != want {
		t.Fatalf("Get(%q) = %d, %v; want %d, true", key, got, ok, want)
	}
}

func mustMiss(t *testing.T, c *Cache[string, int], key string) {
	t.Helper()
	if got, ok := c.Get(key); ok {
		t.Fatalf("Get(%q) = %d, true; want a miss", key, got)
	}
}

func TestNewRejectsBadConfig(t *testing.T) {
	for _, tc := range []struct {
		name     string
		capacity int
		ttl      time.Duration
		now      func() time.Time
	}{
		{"zero capacity", 0, time.Minute, time.Now},
		{"zero ttl", 1, 0, time.Now},
		{"negative ttl", 1, -time.Second, time.Now},
		{"nil clock", 1, time.Minute, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := New[string, int](tc.capacity, tc.ttl, tc.now); !errors.Is(err, ErrConfig) {
				t.Fatalf("err = %v, want ErrConfig", err)
			}
		})
	}
}

func TestSetGetDelete(t *testing.T) {
	c, _ := newCache(t, 10, time.Minute)
	mustMiss(t, c, "a")
	c.Set("a", 1)
	mustGet(t, c, "a", 1)
	c.Set("a", 2)
	mustGet(t, c, "a", 2)
	if c.Len() != 1 {
		t.Fatalf("Len = %d after overwriting one key, want 1", c.Len())
	}
	if !c.Delete("a") || c.Delete("a") {
		t.Fatal("Delete must report true once, then false")
	}
	mustMiss(t, c, "a")
}

func TestLazyExpiry(t *testing.T) {
	c, clk := newCache(t, 10, time.Minute)
	c.Set("a", 1)
	clk.Advance(time.Minute - time.Nanosecond)
	mustGet(t, c, "a", 1)
	clk.Advance(time.Nanosecond) // exactly at the expiry: gone
	if c.Len() != 1 {
		t.Fatalf("Len = %d before the expired entry is touched, want 1", c.Len())
	}
	mustMiss(t, c, "a")
	if c.Len() != 0 {
		t.Fatalf("Len = %d after Get of an expired entry, want 0: Get must remove it", c.Len())
	}
}

func TestPerEntryTTLAndOverwriteRenews(t *testing.T) {
	c, clk := newCache(t, 10, time.Minute)
	c.SetWithTTL("short", 1, time.Second)
	c.Set("renewed", 2)
	clk.Advance(30 * time.Second)
	c.Set("renewed", 3) // a new value gets a new full TTL
	clk.Advance(45 * time.Second)
	mustMiss(t, c, "short")
	mustGet(t, c, "renewed", 3)
	c.SetWithTTL("renewed", 4, 0)
	mustMiss(t, c, "renewed")
}

func TestLRUEviction(t *testing.T) {
	c, _ := newCache(t, 3, time.Hour)
	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("c", 3)
	mustGet(t, c, "a", 1) // a is now the most recent; b is the least recent
	c.Set("d", 4)
	mustMiss(t, c, "b")
	mustGet(t, c, "a", 1)
	mustGet(t, c, "c", 3)
	mustGet(t, c, "d", 4)
	c.Set("c", 30) // an overwrite also counts as use: a is now the oldest
	c.Set("e", 5)
	mustMiss(t, c, "a")
	if c.Len() != 3 {
		t.Fatalf("Len = %d, want the capacity, 3", c.Len())
	}
}

func TestCapacityOne(t *testing.T) {
	c, _ := newCache(t, 1, time.Hour)
	c.Set("a", 1)
	c.Set("b", 2)
	mustMiss(t, c, "a")
	mustGet(t, c, "b", 2)
}

func TestDeleteExpiredIsTheActiveHalf(t *testing.T) {
	c, clk := newCache(t, 100, time.Minute)
	for i := range 10 {
		c.Set(fmt.Sprintf("old-%d", i), i)
	}
	clk.Advance(30 * time.Second)
	for i := range 5 {
		c.Set(fmt.Sprintf("new-%d", i), i)
	}
	clk.Advance(30 * time.Second) // old-* expire now, new-* in 30s
	if n := c.DeleteExpired(); n != 10 {
		t.Fatalf("DeleteExpired removed %d, want 10", n)
	}
	if c.Len() != 5 {
		t.Fatalf("Len = %d, want 5", c.Len())
	}
	if n := c.DeleteExpired(); n != 0 {
		t.Fatalf("second DeleteExpired removed %d, want 0", n)
	}
}

func TestJanitorSweepsOnItsInterval(t *testing.T) {
	c, clk := newCache(t, 100, time.Minute)
	for i := range 3 {
		c.Set(fmt.Sprintf("k%d", i), i)
	}
	ctx, cancel := context.WithCancel(context.Background())
	stopped := make(chan struct{})
	go func() { c.RunJanitor(ctx, clk, 10*time.Second); close(stopped) }()
	defer func() { cancel(); <-stopped }()

	// The janitor sets a relative timer; wait until it exists before moving time.
	clk.BlockUntil(1)
	clk.Advance(50 * time.Second) // one sweep, nothing expired yet
	clk.BlockUntil(1)
	if c.Len() != 3 {
		t.Fatalf("Len = %d before expiry, want 3", c.Len())
	}
	clk.Advance(10 * time.Second) // now at the TTL: the next sweep removes all three
	deadline := time.Now().Add(2 * time.Second)
	for c.Len() != 0 {
		if time.Now().After(deadline) {
			t.Fatalf("Len = %d, want 0: the janitor did not sweep", c.Len())
		}
		time.Sleep(time.Millisecond)
	}
}

func TestConcurrentUseKeepsTheBound(t *testing.T) {
	c, clk := newCache(t, 64, time.Second)
	var wg sync.WaitGroup
	for g := range 16 {
		wg.Go(func() {
			for i := range 2000 {
				key := fmt.Sprintf("k%d", (g*7+i)%200)
				switch i % 4 {
				case 0, 1:
					c.Set(key, i)
				case 2:
					c.Get(key)
				default:
					c.Delete(key)
				}
				if i%500 == 0 {
					c.DeleteExpired()
				}
				if n := c.Len(); n > 64 {
					t.Errorf("Len = %d, above capacity 64", n)
					return
				}
			}
		})
	}
	wg.Go(func() {
		for range 100 {
			clk.Advance(10 * time.Millisecond)
		}
	})
	wg.Wait()
}
