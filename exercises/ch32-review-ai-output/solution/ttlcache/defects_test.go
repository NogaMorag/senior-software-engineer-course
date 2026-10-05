package ttlcache_test

// One test per planted defect. Each fails on the starter and passes on the solution.
// Run with -race: the first one is only visible to the race detector.

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"example.com/ch32-review-ai-output/ttlcache"
)

// Defect 1: Get increments the counters while holding only the read lock, so two
// concurrent readers write the same memory. The race detector reports it.
func TestConcurrentGetsDoNotRace(t *testing.T) {
	c := ttlcache.New(time.Minute, newClock().Now)
	c.Set("a", []byte("1"))
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			for range 200 {
				c.Get("a")
				c.Get("missing")
			}
		})
	}
	wg.Wait()
	if s := c.Stats(); s.Hits != 8*200 || s.Misses != 8*200 {
		t.Fatalf("stats = %+v, want %d hits and %d misses", s, 8*200, 8*200)
	}
}

// Defect 2: the janitor must stop when its context is cancelled. A goroutine that never
// returns is a leak: one per cache, forever, holding the cache in memory.
func TestJanitorStopsOnCancel(t *testing.T) {
	c := ttlcache.New(time.Minute, nil)
	ctx, cancel := context.WithCancel(context.Background())
	done := c.StartJanitor(ctx, time.Millisecond)
	time.Sleep(5 * time.Millisecond) // let it sweep a few times
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("janitor still running one second after its context was cancelled")
	}
}

// Defect 3: an entry set at t with TTL d must be gone at exactly t+d, and Get must agree
// with Sweep about it. The AI's own expiry test advanced the clock by 2d, far from the edge.
func TestEntryExpiresExactlyAtTTL(t *testing.T) {
	tests := []struct {
		name    string
		advance time.Duration
		want    bool
	}{
		{"just before the TTL", time.Minute - time.Nanosecond, true},
		{"exactly at the TTL", time.Minute, false},
		{"after the TTL", time.Minute + time.Nanosecond, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clk := newClock()
			c := ttlcache.New(time.Minute, clk.Now)
			c.Set("a", []byte("1"))
			clk.Advance(tt.advance)
			if _, ok := c.Get("a"); ok != tt.want {
				t.Fatalf("Get after %v: found = %v, want %v", tt.advance, ok, tt.want)
			}
			swept := c.Sweep() == 1
			if swept == tt.want {
				t.Fatalf("Get and Sweep disagree after %v: Get found = %v, Sweep removed = %v", tt.advance, tt.want, swept)
			}
		})
	}
}

// Defect 4: a body over the limit must be rejected with 413 and must not be stored. The
// starter ignores the read error and stores the first maxBody bytes as if all was well.
func TestPutTooLargeIs413AndStoresNothing(t *testing.T) {
	c := ttlcache.New(time.Minute, nil)
	srv := httptest.NewServer(ttlcache.NewHandler(c, 8))
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodPut, srv.URL+"/cache/big", strings.NewReader("0123456789abcdef"))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Errorf("PUT of 16 bytes with a limit of 8: status = %d, want 413", resp.StatusCode)
	}
	if v, ok := c.Get("big"); ok {
		t.Errorf("a rejected PUT stored %q", v)
	}
}
