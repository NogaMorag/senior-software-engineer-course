package links_test

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"example.com/ch24-telemetry/internal/links"
)

const base = "http://sho.rt"

func newMux() *http.ServeMux {
	mux := http.NewServeMux()
	links.NewHandler(links.NewMemoryStore(), base, 1024, slog.New(slog.DiscardHandler)).Register(mux)
	return mux
}

// do sends one request to h and returns the recorder.
func do(h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(method, path, strings.NewReader(body)))
	return rec
}

func decode(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &m); err != nil {
		t.Fatalf("body is not JSON: %q", rec.Body.String())
	}
	return m
}

func TestCreate(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantError  string // substring of the "error" field, when not 201
	}{
		{"random code", `{"url": "https://go.dev/doc/"}`, http.StatusCreated, ""},
		{"custom code", `{"url": "https://go.dev/", "code": "go-home"}`, http.StatusCreated, ""},
		{"missing url", `{}`, http.StatusBadRequest, "url is required"},
		{"relative url", `{"url": "/just/a/path"}`, http.StatusBadRequest, "absolute"},
		{"ftp url", `{"url": "ftp://example.com/file"}`, http.StatusBadRequest, "absolute"},
		{"bad code", `{"url": "https://go.dev/", "code": "no spaces"}`, http.StatusBadRequest, "letters, digits"},
		{"reserved code", `{"url": "https://go.dev/", "code": "healthz"}`, http.StatusBadRequest, "reserved"},
		{"unknown field", `{"url": "https://go.dev/", "ttl": 5}`, http.StatusBadRequest, "unknown field"},
		{"too large", `{"url": "https://go.dev/?q=` + strings.Repeat("a", 2000) + `"}`, http.StatusRequestEntityTooLarge, "larger than"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := do(newMux(), http.MethodPost, "/v1/links", tt.body)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body %s", rec.Code, tt.wantStatus, rec.Body.String())
			}
			got := decode(t, rec)
			if tt.wantStatus != http.StatusCreated {
				if msg, _ := got["error"].(string); !strings.Contains(msg, tt.wantError) {
					t.Errorf("error = %q, want it to contain %q", msg, tt.wantError)
				}
				return
			}
			code, _ := got["code"].(string)
			if got["short_url"] != base+"/"+code || rec.Header().Get("Location") != "/v1/links/"+code {
				t.Errorf("short_url = %v, Location = %q, code %q", got["short_url"], rec.Header().Get("Location"), code)
			}
		})
	}
}

func TestCreateDuplicateCodeIsConflict(t *testing.T) {
	mux := newMux()
	body := `{"url": "https://go.dev/", "code": "dup"}`
	if rec := do(mux, http.MethodPost, "/v1/links", body); rec.Code != http.StatusCreated {
		t.Fatalf("first create = %d", rec.Code)
	}
	if rec := do(mux, http.MethodPost, "/v1/links", body); rec.Code != http.StatusConflict {
		t.Errorf("second create = %d, want 409", rec.Code)
	}
}

func TestResolveAndStats(t *testing.T) {
	mux := newMux()
	do(mux, http.MethodPost, "/v1/links", `{"url": "https://go.dev/blog/", "code": "blog"}`)

	for range 2 {
		rec := do(mux, http.MethodGet, "/blog", "")
		if rec.Code != http.StatusFound || rec.Header().Get("Location") != "https://go.dev/blog/" {
			t.Fatalf("GET /blog = %d Location %q, want 302 to the stored URL", rec.Code, rec.Header().Get("Location"))
		}
	}
	rec := do(mux, http.MethodGet, "/v1/links/blog", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("stats = %d", rec.Code)
	}
	if got := decode(t, rec); got["hits"] != float64(2) || got["url"] != "https://go.dev/blog/" {
		t.Errorf("stats = %v, want 2 hits", got)
	}
}

func TestRoutingErrors(t *testing.T) {
	tests := []struct {
		name, method, path string
		wantStatus         int
	}{
		{"unknown code", http.MethodGet, "/nothing-here", http.StatusNotFound},
		{"unknown stats", http.MethodGet, "/v1/links/nothing-here", http.StatusNotFound},
		{"wrong method", http.MethodGet, "/v1/links", http.StatusMethodNotAllowed},
		{"delete is not a route", http.MethodDelete, "/v1/links/abc", http.StatusMethodNotAllowed},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := do(newMux(), tt.method, tt.path, "")
			if rec.Code != tt.wantStatus {
				t.Errorf("%s %s = %d, want %d", tt.method, tt.path, rec.Code, tt.wantStatus)
			}
			if tt.wantStatus == http.StatusMethodNotAllowed && rec.Header().Get("Allow") == "" {
				t.Error("405 without an Allow header")
			}
		})
	}
}
