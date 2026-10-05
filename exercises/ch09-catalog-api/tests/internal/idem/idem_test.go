package idem_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"example.com/ch09-catalog-api/internal/idem"
)

type clock struct{ t time.Time }

func (c *clock) now() time.Time { return c.t }

// counting is a handler that answers 201 with a growing counter, so a replay is easy to recognise.
func counting(calls *atomic.Int64) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		n := calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Location", "/v1/items/"+string(rune('0'+n)))
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{"call": n, "echo": string(body)})
	})
}

func post(h http.Handler, path, key, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if key != "" {
		req.Header.Set(idem.Header, key)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func problemType(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	if ct := rec.Header().Get("Content-Type"); ct != "application/problem+json" {
		t.Fatalf("Content-Type = %q, want application/problem+json (body %s)", ct, rec.Body)
	}
	var p struct{ Type string }
	_ = json.Unmarshal(rec.Body.Bytes(), &p)
	return p.Type[strings.LastIndex(p.Type, "/")+1:]
}

func TestMissingOrInvalidKey(t *testing.T) {
	var calls atomic.Int64
	h := idem.NewStore(time.Hour, nil).Middleware(counting(&calls))
	for _, key := range []string{"", `""`, strings.Repeat("k", 256), "has space"} {
		rec := post(h, "/v1/items", key, `{}`)
		if rec.Code != http.StatusBadRequest || problemType(t, rec) != "idempotency-key-missing" {
			t.Errorf("key %q: status %d body %s, want 400 idempotency-key-missing", key, rec.Code, rec.Body)
		}
	}
	if calls.Load() != 0 {
		t.Fatalf("handler ran %d times without a valid key", calls.Load())
	}
}

func TestSameKeySameBodyReplays(t *testing.T) {
	var calls atomic.Int64
	h := idem.NewStore(time.Hour, nil).Middleware(counting(&calls))
	first := post(h, "/v1/items", "k-1", `{"sku":"A"}`)
	again := post(h, "/v1/items", `"k-1"`, `{"sku":"A"}`) // the quoted Structured Fields form is the same key
	if calls.Load() != 1 {
		t.Fatalf("handler ran %d times, want 1", calls.Load())
	}
	if again.Code != first.Code || again.Body.String() != first.Body.String() {
		t.Fatalf("replay = %d %s, want %d %s", again.Code, again.Body, first.Code, first.Body)
	}
	if again.Header().Get("Location") != first.Header().Get("Location") {
		t.Fatalf("replay lost the Location header")
	}
	if first.Header().Get(idem.ReplayedHeader) != "" || again.Header().Get(idem.ReplayedHeader) != "true" {
		t.Fatalf("only the replay carries %s", idem.ReplayedHeader)
	}
}

func TestSameKeyDifferentBodyIs422(t *testing.T) {
	var calls atomic.Int64
	h := idem.NewStore(time.Hour, nil).Middleware(counting(&calls))
	post(h, "/v1/items", "k-1", `{"sku":"A"}`)
	rec := post(h, "/v1/items", "k-1", `{"sku":"B"}`)
	if rec.Code != http.StatusUnprocessableEntity || problemType(t, rec) != "idempotency-key-reused" {
		t.Fatalf("status %d body %s, want 422 idempotency-key-reused", rec.Code, rec.Body)
	}
	if calls.Load() != 1 {
		t.Fatalf("handler ran %d times, want 1", calls.Load())
	}
}

func TestKeyIsScopedToRoute(t *testing.T) {
	var calls atomic.Int64
	h := idem.NewStore(time.Hour, nil).Middleware(counting(&calls))
	post(h, "/v1/items", "k-1", `{}`)
	rec := post(h, "/v1/orders", "k-1", `{}`)
	if rec.Code != http.StatusCreated || calls.Load() != 2 {
		t.Fatalf("same key on another route: status %d, calls %d; want 201 and 2", rec.Code, calls.Load())
	}
}

// Two requests with one key at the same time: the first runs, the second gets 409 while it runs,
// and after it finishes the same request is a replay.
func TestConcurrentSameKeyIs409(t *testing.T) {
	entered, release := make(chan struct{}), make(chan struct{})
	var calls atomic.Int64
	slow := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 { // only the first call blocks, so a broken middleware fails instead of hanging
			close(entered)
			<-release
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{"id":"itm_00000001"}`)
	})
	h := idem.NewStore(time.Hour, nil).Middleware(slow)

	done := make(chan *httptest.ResponseRecorder)
	go func() { done <- post(h, "/v1/items", "k-1", `{}`) }()
	<-entered
	second := post(h, "/v1/items", "k-1", `{}`)
	if second.Code != http.StatusConflict || problemType(t, second) != "idempotency-in-progress" {
		t.Fatalf("concurrent retry: status %d body %s, want 409 idempotency-in-progress", second.Code, second.Body)
	}
	if second.Header().Get("Retry-After") == "" {
		t.Fatalf("409 in progress should carry Retry-After")
	}
	close(release)
	first := <-done
	if first.Code != http.StatusCreated {
		t.Fatalf("first request: status %d, want 201", first.Code)
	}
	third := post(h, "/v1/items", "k-1", `{}`)
	if third.Code != http.StatusCreated || third.Header().Get(idem.ReplayedHeader) != "true" || calls.Load() != 1 {
		t.Fatalf("after completion: status %d, replayed %q, calls %d; want a replay of the one call",
			third.Code, third.Header().Get(idem.ReplayedHeader), calls.Load())
	}
}

func TestServerErrorIsNotStored(t *testing.T) {
	var calls atomic.Int64
	flaky := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusCreated)
	})
	h := idem.NewStore(time.Hour, nil).Middleware(flaky)
	if rec := post(h, "/v1/items", "k-1", `{}`); rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("first: %d, want 503", rec.Code)
	}
	if rec := post(h, "/v1/items", "k-1", `{}`); rec.Code != http.StatusCreated || calls.Load() != 2 {
		t.Fatalf("retry after 5xx: status %d, calls %d; want 201 and 2 (the key was released)", rec.Code, calls.Load())
	}
}

func TestPanicReleasesKey(t *testing.T) {
	var calls atomic.Int64
	boom := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			panic("boom")
		}
		w.WriteHeader(http.StatusCreated)
	})
	keys := idem.NewStore(time.Hour, nil)
	h := keys.Middleware(boom)
	func() {
		defer func() { _ = recover() }()
		post(h, "/v1/items", "k-1", `{}`)
	}()
	if keys.Len() != 0 {
		t.Fatalf("a panicking request left its key claimed")
	}
	if rec := post(h, "/v1/items", "k-1", `{}`); rec.Code != http.StatusCreated {
		t.Fatalf("retry after panic: %d, want 201", rec.Code)
	}
}

func TestKeysExpire(t *testing.T) {
	clk := &clock{t: time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC)}
	var calls atomic.Int64
	keys := idem.NewStore(time.Hour, clk.now)
	h := keys.Middleware(counting(&calls))
	post(h, "/v1/items", "k-1", `{"sku":"A"}`)
	post(h, "/v1/items", "k-2", `{"sku":"A"}`)
	clk.t = clk.t.Add(59 * time.Minute)
	if rec := post(h, "/v1/items", "k-1", `{"sku":"B"}`); rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("before expiry: %d, want 422", rec.Code)
	}
	clk.t = clk.t.Add(2 * time.Minute)
	if rec := post(h, "/v1/items", "k-1", `{"sku":"B"}`); rec.Code != http.StatusCreated {
		t.Fatalf("after expiry the key is free: %d, want 201", rec.Code)
	}
	if keys.Len() != 1 {
		t.Fatalf("after the sweep %d keys remain, want 1 (k-2 expired)", keys.Len())
	}
}

func TestBodyOverLimitIs413(t *testing.T) {
	var calls atomic.Int64
	h := idem.NewStore(time.Hour, nil).Middleware(counting(&calls))
	limited := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 8)
		h.ServeHTTP(w, r)
	})
	rec := post(limited, "/v1/items", "k-1", `{"sku":"far too long"}`)
	if rec.Code != http.StatusRequestEntityTooLarge || problemType(t, rec) != "body-too-large" {
		t.Fatalf("status %d body %s, want 413 body-too-large", rec.Code, rec.Body)
	}
}
