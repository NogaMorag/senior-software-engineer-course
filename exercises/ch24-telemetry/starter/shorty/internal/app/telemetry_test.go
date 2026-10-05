package app_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"example.com/ch24-telemetry/internal/app"
	"example.com/ch24-telemetry/internal/links"
	"example.com/ch24-telemetry/internal/platform"
)

// The instrumentation uses the global providers, as it does in production.
// The tests install in-memory ones once, before any handler is built, and
// reset the recorded spans in every test. They do not run in parallel.
var (
	spans  = tracetest.NewInMemoryExporter()
	reader = sdkmetric.NewManualReader()
)

func TestMain(m *testing.M) {
	otel.SetTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithSyncer(spans)))
	otel.SetMeterProvider(sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader)))
	os.Exit(m.Run())
}

// instrumented returns a handler and the buffer its JSON log goes to.
func instrumented(t *testing.T, publish links.PublisherFunc) (http.Handler, *bytes.Buffer) {
	t.Helper()
	spans.Reset()
	var buf bytes.Buffer
	cfg := app.Config{BaseURL: "http://sho.rt", MaxBodyBytes: 4096}
	return app.NewHandler(cfg, links.NewMemoryStore(), platform.NewLogger(&buf, slog.LevelInfo), publish), &buf
}

func serve(h http.Handler, method, path, body string) int {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(method, path, strings.NewReader(body)))
	return rec.Code
}

func TestServerSpansAreNamedByRoute(t *testing.T) {
	h, _ := instrumented(t, nil)
	serve(h, http.MethodPost, "/v1/links", `{"url": "https://go.dev/", "code": "gohome"}`)
	serve(h, http.MethodGet, "/gohome", "")
	serve(h, http.MethodGet, "/healthz", "")

	got := spans.GetSpans()
	if len(got) != 2 {
		t.Fatalf("got %d spans, want 2 (healthz is filtered out): %v", len(got), names(got))
	}
	for i, want := range []string{"POST /v1/links", "GET /{code}"} {
		if got[i].Name != want {
			t.Errorf("span %d = %q, want %q", i, got[i].Name, want)
		}
		if route := attr(got[i].Attributes, "http.route"); route != strings.SplitN(want, " ", 2)[1] {
			t.Errorf("span %q has http.route %q", want, route)
		}
	}
}

func TestRequestMetricsCarryRouteAndStatus(t *testing.T) {
	h, _ := instrumented(t, nil)
	serve(h, http.MethodGet, "/nosuchcode", "") // 404 from the handler: the route still matched

	var rm metricdata.ResourceMetrics
	if err := reader.Collect(context.Background(), &rm); err != nil {
		t.Fatal(err)
	}
	hist, ok := find(rm, "http.server.request.duration").(metricdata.Histogram[float64])
	if !ok {
		t.Fatalf("no http.server.request.duration histogram in %v", rm.ScopeMetrics)
	}
	for _, dp := range hist.DataPoints {
		route, _ := dp.Attributes.Value("http.route")
		status, _ := dp.Attributes.Value("http.response.status_code")
		if route.AsString() == "/{code}" && status.AsInt64() == 404 && dp.Count > 0 {
			return
		}
	}
	t.Errorf("no data point with http.route=/{code} and status 404: %+v", hist.DataPoints)
}

func TestAccessLogCarriesTheTraceID(t *testing.T) {
	h, buf := instrumented(t, nil)
	serve(h, http.MethodGet, "/nosuchcode", "")

	var line map[string]any
	if err := json.Unmarshal(buf.Bytes(), &line); err != nil {
		t.Fatalf("log = %q: %v", buf.String(), err)
	}
	got := spans.GetSpans()
	if len(got) != 1 {
		t.Fatalf("got %d spans, want 1", len(got))
	}
	if line["trace_id"] != got[0].SpanContext.TraceID().String() {
		t.Errorf("access log trace_id = %v, span trace id = %s", line["trace_id"], got[0].SpanContext.TraceID())
	}
}

func TestFailedPublishIsA503AndAnErrorSpan(t *testing.T) {
	h, _ := instrumented(t, func(context.Context, links.Link) error { return errors.New("unknown topic") })
	if code := serve(h, http.MethodPost, "/v1/links", `{"url": "https://go.dev/"}`); code != http.StatusServiceUnavailable {
		t.Fatalf("create with a failing publisher = %d, want 503", code)
	}
	got := spans.GetSpans()
	if len(got) != 1 || got[0].Status.Code != codes.Error {
		t.Errorf("spans = %v, want one span with status Error", got)
	}
}

func names(s tracetest.SpanStubs) []string {
	var out []string
	for _, span := range s {
		out = append(out, span.Name)
	}
	return out
}

func attr(kvs []attribute.KeyValue, key string) string {
	for _, kv := range kvs {
		if string(kv.Key) == key {
			return kv.Value.String()
		}
	}
	return ""
}

func find(rm metricdata.ResourceMetrics, name string) metricdata.Aggregation {
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name == name {
				return m.Data
			}
		}
	}
	return nil
}
