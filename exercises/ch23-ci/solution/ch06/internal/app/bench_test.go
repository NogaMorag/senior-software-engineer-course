package app_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"example.com/ch06-shorty/internal/app"
	"example.com/ch06-shorty/internal/links"
	"example.com/ch06-shorty/internal/platform"
)

// BenchmarkResolve measures the whole resolve path in process: middleware
// (request ID, access log as JSON, recover), routing, the store and the
// redirect. No network, so it measures shorty, not the loopback interface.
//
//	go test -run '^$' -bench Resolve -benchmem ./internal/app/
func BenchmarkResolve(b *testing.B) {
	store := links.NewMemoryStore()
	if err := store.Create(context.Background(), links.Link{Code: "go", URL: "https://go.dev/"}); err != nil {
		b.Fatal(err)
	}
	log := platform.NewLogger(io.Discard, slog.LevelInfo) // real JSON encoding, output discarded
	h := app.NewHandler(app.Config{BaseURL: "http://sho.rt", MaxBodyBytes: 4096}, store, log)
	req := httptest.NewRequest(http.MethodGet, "/go", nil)

	for b.Loop() {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusFound {
			b.Fatalf("status %d", rec.Code)
		}
	}
}
