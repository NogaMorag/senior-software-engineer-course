package ratelimit_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"example.com/ch16-ratelimit/ratelimit"
)

func ok() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
}

func byHeader(r *http.Request) string { return r.Header.Get("X-Tenant") }

func TestMiddlewareRefusesWith429(t *testing.T) {
	c := newClock()
	h := ratelimit.Middleware(ratelimit.Options{
		Limiter: &ratelimit.Local{Rate: 0.5, Burst: 2, Now: c.Now}, Key: byHeader, Rate: 0.5, Burst: 2,
	}, ok())
	send := func(tenant string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/v1/events", nil)
		req.Header.Set("X-Tenant", tenant)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec
	}
	tests := []struct {
		tenant     string
		code       int
		rateLimit  string
		retryAfter string
	}{
		{"a", 200, `"default";r=1`, ""},
		{"a", 200, `"default";r=0`, ""},
		{"a", 429, `"default";r=0`, "2"}, // half a token per second: the next one is 2s away
		{"b", 200, `"default";r=1`, ""},  // another tenant, another bucket
	}
	for i, tt := range tests {
		rec := send(tt.tenant)
		if rec.Code != tt.code {
			t.Fatalf("request %d (tenant %s): status %d, want %d", i, tt.tenant, rec.Code, tt.code)
		}
		if got := rec.Header().Get("RateLimit"); got != tt.rateLimit {
			t.Errorf("request %d: RateLimit = %q, want %q", i, got, tt.rateLimit)
		}
		if got := rec.Header().Get("Retry-After"); got != tt.retryAfter {
			t.Errorf("request %d: Retry-After = %q, want %q", i, got, tt.retryAfter)
		}
		if got := rec.Header().Get("RateLimit-Policy"); got != `"default";q=2;w=4` {
			t.Errorf("request %d: RateLimit-Policy = %q", i, got)
		}
	}
}

type broken struct{}

func (broken) Allow(context.Context, string) (ratelimit.Decision, error) {
	return ratelimit.Decision{}, errors.New("redis: connection refused")
}

func TestMiddlewareWhenTheLimiterFails(t *testing.T) {
	for _, failOpen := range []bool{true, false} {
		errs := 0
		h := ratelimit.Middleware(ratelimit.Options{
			Limiter: broken{}, Key: byHeader, Rate: 1, Burst: 1, FailOpen: failOpen,
			OnError: func(error) { errs++ },
		}, ok())
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
		want := http.StatusServiceUnavailable
		if failOpen {
			want = http.StatusOK
		}
		if rec.Code != want || errs != 1 {
			t.Errorf("failOpen=%v: status %d, errors reported %d; want %d and 1", failOpen, rec.Code, errs, want)
		}
	}
}

func TestRetryAfterSeconds(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{0, "1"}, {time.Millisecond, "1"}, {time.Second, "1"}, {1001 * time.Millisecond, "2"}, {90 * time.Second, "90"},
	}
	for _, tt := range tests {
		if got := ratelimit.RetryAfterSeconds(tt.d); got != tt.want {
			t.Errorf("RetryAfterSeconds(%v) = %q, want %q", tt.d, got, tt.want)
		}
	}
}
