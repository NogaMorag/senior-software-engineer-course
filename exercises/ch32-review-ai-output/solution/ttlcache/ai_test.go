package ttlcache_test

// These are the tests the AI assistant wrote along with the module. They all pass on the
// starter: read them and ask what they do not check.

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"example.com/ch32-review-ai-output/ttlcache"
)

// clock is a fake time source shared by every test in the package.
type clock struct {
	mu  sync.Mutex
	now time.Time
}

func newClock() *clock                   { return &clock{now: time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)} }
func (c *clock) Now() time.Time          { c.mu.Lock(); defer c.mu.Unlock(); return c.now }
func (c *clock) Advance(d time.Duration) { c.mu.Lock(); defer c.mu.Unlock(); c.now = c.now.Add(d) }

func TestSetGet(t *testing.T) {
	c := ttlcache.New(time.Minute, newClock().Now)
	c.Set("a", []byte("1"))
	v, ok := c.Get("a")
	if !ok || string(v) != "1" {
		t.Fatalf("Get(a) = %q, %v; want 1, true", v, ok)
	}
	if _, ok := c.Get("missing"); ok {
		t.Fatal("Get(missing) found a value")
	}
}

func TestExpiry(t *testing.T) {
	clk := newClock()
	c := ttlcache.New(time.Minute, clk.Now)
	c.Set("a", []byte("1"))
	clk.Advance(2 * time.Minute)
	if _, ok := c.Get("a"); ok {
		t.Fatal("entry still there two minutes after a one-minute TTL")
	}
}

func TestSweepRemovesExpired(t *testing.T) {
	clk := newClock()
	c := ttlcache.New(time.Minute, clk.Now)
	c.Set("old", []byte("1"))
	clk.Advance(2 * time.Minute)
	c.Set("new", []byte("2"))
	if n := c.Sweep(); n != 1 {
		t.Fatalf("Sweep removed %d, want 1", n)
	}
	if got := c.Stats().Entries; got != 1 {
		t.Fatalf("entries after sweep = %d, want 1", got)
	}
}

func TestStatsCount(t *testing.T) {
	c := ttlcache.New(time.Minute, newClock().Now)
	c.Set("a", []byte("1"))
	c.Get("a")
	c.Get("b")
	if s := c.Stats(); s.Hits != 1 || s.Misses != 1 {
		t.Fatalf("stats = %+v, want 1 hit and 1 miss", s)
	}
}

func TestHandlerRoundTrip(t *testing.T) {
	srv := httptest.NewServer(ttlcache.NewHandler(ttlcache.New(time.Minute, nil), 1024))
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodPut, srv.URL+"/cache/greeting", strings.NewReader("hello"))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("PUT status = %d, want 204", resp.StatusCode)
	}

	resp, err = http.Get(srv.URL + "/cache/greeting")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET status = %d, want 200", resp.StatusCode)
	}

	resp2, err := http.Get(srv.URL + "/cache/nothing")
	if err != nil {
		t.Fatal(err)
	}
	resp2.Body.Close()
	if resp2.StatusCode != http.StatusNotFound {
		t.Fatalf("GET missing status = %d, want 404", resp2.StatusCode)
	}
}
