package telemetry_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"

	"example.com/hookline/internal/telemetry"
)

// TestSetupExportsAllThreeSignals points the exporters at a fake collector
// and checks that a span, a metric and a log line all arrive over OTLP/HTTP
// by the time Shutdown returns.
func TestSetupExportsAllThreeSignals(t *testing.T) {
	var mu sync.Mutex
	got := map[string]int{}
	collector := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		mu.Lock()
		got[r.URL.Path]++
		mu.Unlock()
	}))
	defer collector.Close()
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", collector.URL)

	ctx := context.Background()
	shutdown, err := telemetry.Setup(ctx, "telemetry-test", "v0")
	if err != nil {
		t.Fatalf("Setup: %v", err)
	}

	_, span := otel.Tracer("test").Start(ctx, "work")
	span.End()
	counter, err := otel.Meter("test").Int64Counter("test.things")
	if err != nil {
		t.Fatal(err)
	}
	counter.Add(ctx, 1)
	otelslog.NewLogger("test").InfoContext(ctx, "hello")

	flushCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := shutdown(flushCtx); err != nil {
		t.Fatalf("shutdown: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	for _, path := range []string{"/v1/traces", "/v1/metrics", "/v1/logs"} {
		if got[path] == 0 {
			t.Errorf("nothing posted to %s; got %v", path, got)
		}
	}
}
