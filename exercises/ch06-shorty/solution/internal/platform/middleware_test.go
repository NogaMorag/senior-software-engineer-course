package platform_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"example.com/ch06-shorty/internal/platform"
)

// syncBuffer is a bytes.Buffer safe for the concurrent writes of a server.
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

// lines parses the JSON log lines written so far.
func (b *syncBuffer) lines(t *testing.T) []map[string]any {
	t.Helper()
	b.mu.Lock()
	defer b.mu.Unlock()
	var out []map[string]any
	for _, line := range strings.Split(strings.TrimSpace(b.buf.String()), "\n") {
		if line == "" {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Fatalf("log line is not JSON: %q", line)
		}
		out = append(out, m)
	}
	return out
}

// newStack builds a mux with the routes that register adds, behind the same
// middleware order as shorty, and returns it with the buffer the logs go to.
func newStack(register func(mux *http.ServeMux, log *slog.Logger)) (http.Handler, *syncBuffer) {
	logs := &syncBuffer{}
	log := platform.NewLogger(logs, slog.LevelDebug)
	mux := http.NewServeMux()
	register(mux, log)
	return platform.Chain(mux, platform.RequestID, platform.AccessLog(log), platform.Recover(log)), logs
}

// one registers a single handler for every path.
func one(h http.HandlerFunc) func(*http.ServeMux, *slog.Logger) {
	return func(mux *http.ServeMux, _ *slog.Logger) { mux.Handle("/", h) }
}

func TestRequestIDIsGeneratedAndEchoed(t *testing.T) {
	var seen string
	h, _ := newStack(one(func(w http.ResponseWriter, r *http.Request) {
		seen = platform.RequestIDFrom(r.Context())
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	got := rec.Header().Get(platform.RequestIDHeader)
	if len(got) != 16 {
		t.Errorf("response %s = %q, want 16 hex characters", platform.RequestIDHeader, got)
	}
	if seen != got {
		t.Errorf("handler saw %q, response carries %q", seen, got)
	}
}

func TestRequestIDFromClient(t *testing.T) {
	tests := []struct {
		name, incoming string
		kept           bool
	}{
		{"well formed", "abc-123.x_y", true},
		{"too long", strings.Repeat("a", 65), false},
		{"newline injection", "abc\nfake log line", false},
		{"spaces", "a b", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, _ := newStack(one(func(http.ResponseWriter, *http.Request) {}))
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set(platform.RequestIDHeader, tt.incoming)
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			got := rec.Header().Get(platform.RequestIDHeader)
			if (got == tt.incoming) != tt.kept {
				t.Errorf("incoming %q became %q; kept = %v, want %v", tt.incoming, got, got == tt.incoming, tt.kept)
			}
		})
	}
}

func TestEveryLogLineCarriesTheRequestID(t *testing.T) {
	h, logs := newStack(func(mux *http.ServeMux, log *slog.Logger) {
		mux.HandleFunc("GET /tea", func(w http.ResponseWriter, r *http.Request) {
			log.InfoContext(r.Context(), "brewing")
			log.DebugContext(r.Context(), "still brewing")
			w.WriteHeader(http.StatusTeapot)
		})
	})

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/tea", nil))
	id := rec.Header().Get(platform.RequestIDHeader)

	lines := logs.lines(t)
	if len(lines) != 3 {
		t.Fatalf("got %d log lines, want 3 (two from the handler, one access log)", len(lines))
	}
	for i, l := range lines {
		if l["request_id"] != id {
			t.Errorf("line %d request_id = %v, want %s", i, l["request_id"], id)
		}
	}
	access := lines[2]
	if access["msg"] != "request" || access["status"] != float64(http.StatusTeapot) || access["pattern"] != "GET /tea" {
		t.Errorf("access log = %v", access)
	}
}

func TestRecoverTurnsPanicInto500AndKeepsServing(t *testing.T) {
	h, logs := newStack(one(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/boom" {
			panic("something went very wrong")
		}
		_, _ = w.Write([]byte("fine"))
	}))
	srv := httptest.NewServer(h)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/boom")
	if err != nil {
		t.Fatalf("GET /boom: %v (a panic must not drop the connection)", err)
	}
	var body map[string]string
	_ = json.NewDecoder(resp.Body).Decode(&body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusInternalServerError || body["error"] != "internal error" {
		t.Errorf("GET /boom = %d %v, want 500 {error: internal error}", resp.StatusCode, body)
	}

	resp, err = http.Get(srv.URL + "/after")
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /after = %v, %v; the server must keep serving", resp, err)
	}
	resp.Body.Close()

	var panicLine, accessLine map[string]any
	for _, l := range logs.lines(t) {
		if l["msg"] == "panic in handler" {
			panicLine = l
		}
		if l["msg"] == "request" && l["path"] == "/boom" {
			accessLine = l
		}
	}
	if panicLine == nil || panicLine["panic"] != "something went very wrong" || panicLine["request_id"] == nil {
		t.Errorf("panic log line = %v", panicLine)
	}
	if accessLine == nil || accessLine["status"] != float64(500) {
		t.Errorf("access log line for /boom = %v, want status 500", accessLine)
	}
}
