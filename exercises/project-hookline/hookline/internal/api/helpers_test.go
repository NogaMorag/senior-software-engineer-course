package api_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"example.com/hookline/internal/api"
	"example.com/hookline/internal/store"
	"example.com/hookline/internal/store/storetest"
)

// clock is a test clock that moves one millisecond per reading, so resources made in a row get
// distinct, increasing timestamps. Handlers run concurrently in some tests, hence the mutex.
type clock struct {
	mu sync.Mutex
	t  time.Time
}

func (c *clock) now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.t = c.t.Add(time.Millisecond)
	return c.t
}

// harness is one API under test: a fresh store (memory, or PostgreSQL when DATABASE_URL is set; see
// store_test.go) and a handler with a small body limit.
type harness struct {
	t     *testing.T
	store store.Store
	h     http.Handler
}

func newHarness(t *testing.T, devMode bool) *harness {
	t.Helper()
	st := newStore(t)
	c := &clock{t: time.Date(2026, 10, 5, 9, 0, 0, 0, time.UTC)}
	h := api.New(st, api.Options{MaxBodyBytes: 4096, DevMode: devMode, Now: c.now})
	return &harness{t: t, store: st, h: h}
}

// tenant creates a tenant and returns its API key.
func (hs *harness) tenant(name string) string {
	hs.t.Helper()
	_, key := storetest.NewTenant(hs.t, hs.store, name)
	return key
}

// req describes one request. Body is sent as application/json unless ContentType says otherwise.
type req struct {
	Method, Path, Key, Body string
	ContentType             string
	Header                  map[string]string
}

func (hs *harness) do(r req) *httptest.ResponseRecorder {
	hs.t.Helper()
	var body io.Reader
	if r.Body != "" {
		body = strings.NewReader(r.Body)
	}
	hr := httptest.NewRequest(r.Method, r.Path, body)
	if r.Body != "" {
		ct := r.ContentType
		if ct == "" {
			ct = "application/json"
		}
		hr.Header.Set("Content-Type", ct)
	}
	if r.Key != "" {
		hr.Header.Set("Authorization", "Bearer "+r.Key)
	}
	for k, v := range r.Header {
		hr.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	hs.h.ServeHTTP(rec, hr)
	return rec
}

// decode parses a JSON response body into a map.
func decode(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &m); err != nil {
		t.Fatalf("response is not a JSON object: %v\n%s", err, rec.Body)
	}
	return m
}

// wantProblem checks the RFC 9457 shape: status, media type, type URI, title, status member, instance.
func wantProblem(t *testing.T, rec *httptest.ResponseRecorder, status int, slug string) map[string]any {
	t.Helper()
	if rec.Code != status {
		t.Fatalf("status = %d, want %d; body %s", rec.Code, status, rec.Body)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/problem+json" {
		t.Fatalf("Content-Type = %q, want application/problem+json", ct)
	}
	p := decode(t, rec)
	if p["type"] != "https://hookline.example.com/problems/"+slug {
		t.Errorf("type = %v, want problems/%s", p["type"], slug)
	}
	if p["status"] != float64(status) || p["title"] == "" || p["title"] == nil {
		t.Errorf("problem = %v, want status %d and a title", p, status)
	}
	if inst, _ := p["instance"].(string); !strings.HasPrefix(inst, "/") {
		t.Errorf("instance = %v, want the request path", p["instance"])
	}
	return p
}

// pointers returns the JSON pointers of a validation problem's errors.
func pointers(p map[string]any) []string {
	var out []string
	errs, _ := p["errors"].([]any)
	for _, e := range errs {
		if m, ok := e.(map[string]any); ok {
			out = append(out, m["pointer"].(string))
		}
	}
	return out
}

// createEndpoint creates an endpoint for key and returns its JSON.
func (hs *harness) createEndpoint(key, url string) map[string]any {
	hs.t.Helper()
	rec := hs.do(req{Method: "POST", Path: "/v1/endpoints", Key: key, Body: `{"url":"` + url + `"}`})
	if rec.Code != http.StatusCreated {
		hs.t.Fatalf("create endpoint = %d %s", rec.Code, rec.Body)
	}
	return decode(hs.t, rec)
}

func (hs *harness) createSubscription(key, endpointID, eventType string) map[string]any {
	hs.t.Helper()
	rec := hs.do(req{Method: "POST", Path: "/v1/subscriptions", Key: key,
		Body: `{"endpoint_id":"` + endpointID + `","event_type":"` + eventType + `"}`})
	if rec.Code != http.StatusCreated {
		hs.t.Fatalf("create subscription = %d %s", rec.Code, rec.Body)
	}
	return decode(hs.t, rec)
}

func (hs *harness) ingest(key, idemKey, body string) *httptest.ResponseRecorder {
	hs.t.Helper()
	return hs.do(req{Method: "POST", Path: "/v1/events", Key: key, Body: body, Header: map[string]string{"Idempotency-Key": idemKey}})
}
