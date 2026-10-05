package api_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"example.com/hookline/internal/api"
	"example.com/hookline/internal/ratelimit"
	"example.com/hookline/internal/shed"
	"example.com/hookline/internal/store/storetest"
)

// fixedClock is a ratelimit clock that only moves when told to.
type fixedClock struct {
	mu sync.Mutex
	t  time.Time
}

func (c *fixedClock) now() time.Time { c.mu.Lock(); defer c.mu.Unlock(); return c.t }
func (c *fixedClock) add(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.t = c.t.Add(d)
}

func ingestN(t *testing.T, h http.Handler, key string, n int, prefix string) []*httptest.ResponseRecorder {
	t.Helper()
	var out []*httptest.ResponseRecorder
	for i := range n {
		hs := &harness{t: t, h: h}
		out = append(out, hs.ingest(key, prefix+strconv.Itoa(i), paid))
	}
	return out
}

// TestTenantOverItsLimitGets429OthersDoNot is spec N5: tenant A (burst 100 from the fixture, the
// default) exhausts its bucket; tenant B, at the same moment, is unaffected.
func TestTenantOverItsLimitGets429OthersDoNot(t *testing.T) {
	st := newStore(t)
	clock := &fixedClock{t: time.Date(2026, 10, 5, 9, 0, 0, 0, time.UTC)}
	h := api.New(st, api.Options{MaxBodyBytes: 4096, Limiter: &ratelimit.Local{Now: clock.now}})
	_, a := storetest.NewTenant(t, st, "A") // rate 50/s, burst 100
	_, b := storetest.NewTenant(t, st, "B")

	recs := ingestN(t, h, a, 101, "a-")
	for i, rec := range recs[:100] {
		if rec.Code != http.StatusAccepted {
			t.Fatalf("request %d within the burst = %d %s", i, rec.Code, rec.Body)
		}
	}
	last := recs[100]
	wantProblem(t, last, http.StatusTooManyRequests, "rate-limited")
	if last.Header().Get("Retry-After") != "1" || last.Header().Get("RateLimit") != `"ingest";r=0` {
		t.Errorf("429 headers: Retry-After %q, RateLimit %q", last.Header().Get("Retry-After"), last.Header().Get("RateLimit"))
	}
	if rec := ingestN(t, h, b, 1, "b-")[0]; rec.Code != http.StatusAccepted || rec.Header().Get("RateLimit") != `"ingest";r=99` {
		t.Fatalf("tenant B while A is limited = %d, RateLimit %q", rec.Code, rec.Header().Get("RateLimit"))
	}
	clock.add(time.Second) // 50 tokens later
	if rec := ingestN(t, h, a, 1, "a2-")[0]; rec.Code != http.StatusAccepted {
		t.Fatalf("tenant A after a second = %d", rec.Code)
	}
}

func TestOnlyIngestIsRateLimited(t *testing.T) {
	st := newStore(t)
	clock := &fixedClock{t: time.Now()}
	h := api.New(st, api.Options{Limiter: &ratelimit.Local{Now: clock.now}})
	_, key := storetest.NewTenant(t, st, "A")
	hs := &harness{t: t, h: h}
	for range 150 {
		if rec := hs.do(req{Method: "GET", Path: "/v1/tenant", Key: key}); rec.Code != http.StatusOK {
			t.Fatalf("GET /v1/tenant = %d: reads are not limited", rec.Code)
		}
	}
}

type brokenLimiter struct{}

func (brokenLimiter) Allow(context.Context, string, float64, int) (ratelimit.Decision, error) {
	return ratelimit.Decision{}, errors.New("redis: connection refused")
}

func TestLimiterFailureFailsOpenAndIsReported(t *testing.T) {
	st := newStore(t)
	var reported atomic.Int64
	h := api.New(st, api.Options{Limiter: brokenLimiter{}, OnLimiterError: func(error) { reported.Add(1) }})
	_, key := storetest.NewTenant(t, st, "A")
	for i, rec := range ingestN(t, h, key, 3, "x-") {
		if rec.Code != http.StatusAccepted {
			t.Fatalf("request %d with the limiter down = %d; the decision is to fail open", i, rec.Code)
		}
	}
	if reported.Load() != 3 {
		t.Fatalf("reported %d limiter errors, want 3", reported.Load())
	}
}

// TestOverloadShedsV1ButNotHealth is spec N6: with every slot taken, /v1 answers 503 with
// Retry-After at once, and the health route still answers.
func TestOverloadShedsV1ButNotHealth(t *testing.T) {
	st := newStore(t)
	sh := shed.New(shed.Options{Limit: 1, RetryAfter: 2 * time.Second})
	release := make(chan struct{})
	entered := make(chan struct{})
	slow := sh.Wrap(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		close(entered)
		<-release
	}))
	go slow.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/v1/anything", nil)) // holds the one slot
	<-entered
	defer close(release)

	h := api.New(st, api.Options{Shedder: sh})
	_, key := storetest.NewTenant(t, st, "A")
	hs := &harness{t: t, h: h}
	rec := hs.do(req{Method: "GET", Path: "/v1/tenant", Key: key})
	wantProblem(t, rec, http.StatusServiceUnavailable, "overloaded")
	if rec.Header().Get("Retry-After") != "2" {
		t.Errorf("Retry-After = %q, want 2", rec.Header().Get("Retry-After"))
	}
	if rec := hs.do(req{Method: "GET", Path: "/healthz"}); rec.Code != http.StatusOK {
		t.Fatalf("/healthz while overloaded = %d; probes must not be shed", rec.Code)
	}
}
