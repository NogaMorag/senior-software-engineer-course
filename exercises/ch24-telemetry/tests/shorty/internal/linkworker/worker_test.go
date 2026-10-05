package linkworker_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"

	"example.com/ch24-telemetry/internal/events"
	"example.com/ch24-telemetry/internal/linkworker"
)

var (
	spans  = tracetest.NewInMemoryExporter()
	reader = sdkmetric.NewManualReader()
)

func TestMain(m *testing.M) {
	otel.SetTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithSyncer(spans)))
	otel.SetMeterProvider(sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader)))
	otel.SetTextMapPropagator(propagation.TraceContext{})
	os.Exit(m.Run())
}

// TestOneTraceAcrossTheKafkaHop plays both sides of the broker with the kotel
// hooks: the producer hook injects the context into the record headers, then
// a fresh record with only those headers (what the consumer fetches) goes
// through the fetch hook and the worker. No broker is needed.
func TestOneTraceAcrossTheKafkaHop(t *testing.T) {
	spans.Reset()
	var gotTraceparent, gotSignature string
	receiver := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotTraceparent = r.Header.Get("traceparent")
		gotSignature = r.Header.Get("Hookline-Signature")
		_, _ = io.Copy(io.Discard, r.Body)
	}))
	defer receiver.Close()

	_, tracer := events.Hooks()
	w, err := linkworker.New(tracer, receiver.URL+"/hooks/links", "change-me", slog.New(slog.DiscardHandler))
	if err != nil {
		t.Fatal(err)
	}

	// Producer side: a parent span (shorty's handler) and the publish hook.
	ctx, parent := otel.Tracer("test").Start(context.Background(), "POST /v1/links")
	produced := &kgo.Record{Topic: "links.created", Value: []byte(`{"code":"abc","url":"https://go.dev/"}`), Context: ctx}
	tracer.OnProduceRecordBuffered(produced)
	tracer.OnProduceRecordUnbuffered(produced, nil)
	parent.End()

	// Consumer side: only the headers survive the trip through the broker.
	fetched := &kgo.Record{Topic: produced.Topic, Value: produced.Value, Headers: produced.Headers, Timestamp: time.Now()}
	tracer.OnFetchRecordBuffered(fetched)
	if err := w.Handle(fetched); err != nil {
		t.Fatalf("Handle: %v", err)
	}
	tracer.OnFetchRecordUnbuffered(fetched, true)

	traceID := parent.SpanContext().TraceID()
	byName := map[string]tracetest.SpanStub{}
	for _, s := range spans.GetSpans() {
		byName[s.Name] = s
		if s.SpanContext.TraceID() != traceID {
			t.Errorf("span %q is in trace %s, want %s", s.Name, s.SpanContext.TraceID(), traceID)
		}
	}
	for _, name := range []string{"POST /v1/links", "links.created publish", "links.created receive", "links.created process", "HTTP POST"} {
		if _, ok := byName[name]; !ok {
			t.Errorf("no span %q; have %v", name, keys(byName))
		}
	}
	if p := byName["links.created process"]; p.Parent.SpanID() != byName["links.created receive"].SpanContext.SpanID() {
		t.Errorf("process span's parent is not the receive span")
	}
	if p := byName["links.created process"]; p.SpanKind != trace.SpanKindConsumer {
		t.Errorf("process span kind = %v, want consumer", p.SpanKind)
	}
	if !strings.Contains(gotTraceparent, traceID.String()) {
		t.Errorf("receiver got traceparent %q, want trace %s", gotTraceparent, traceID)
	}
	if !strings.HasPrefix(gotSignature, "t=") {
		t.Errorf("receiver got signature %q", gotSignature)
	}
}

func TestUSEMetrics(t *testing.T) {
	failing := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer failing.Close()
	_, tracer := events.Hooks()
	w, err := linkworker.New(tracer, failing.URL, "change-me", slog.New(slog.DiscardHandler))
	if err != nil {
		t.Fatal(err)
	}
	rec := &kgo.Record{Topic: "links.created", Value: []byte(`{"code":"abc"}`), Timestamp: time.Now().Add(-2 * time.Second)}
	if err := w.Handle(rec); err == nil {
		t.Fatal("Handle succeeded against a receiver that answers 500")
	}

	var rm metricdata.ResourceMetrics
	if err := reader.Collect(context.Background(), &rm); err != nil {
		t.Fatal(err)
	}
	got := map[string]metricdata.Aggregation{}
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			got[m.Name] = m.Data
		}
	}
	deliveries, _ := got["linkworker.deliveries"].(metricdata.Sum[int64])
	var errorsSeen int64
	for _, dp := range deliveries.DataPoints {
		if v, _ := dp.Attributes.Value("outcome"); v.AsString() == "error" {
			errorsSeen += dp.Value
		}
	}
	if errorsSeen < 1 {
		t.Errorf("linkworker.deliveries{outcome=error} = %d, want at least 1", errorsSeen)
	}
	wait, _ := got["linkworker.wait"].(metricdata.Histogram[float64])
	if len(wait.DataPoints) == 0 {
		t.Fatalf("no linkworker.wait data: %+v", got["linkworker.wait"])
	}
	if maxWait, ok := wait.DataPoints[0].Max.Value(); !ok || maxWait < 2 {
		t.Errorf("linkworker.wait max = %.2fs, want at least the 2s the record waited", maxWait)
	}
	if _, ok := got["linkworker.busy"].(metricdata.Sum[float64]); !ok {
		t.Errorf("no linkworker.busy counter")
	}
}

func TestSignMatchesTheSharedVector(t *testing.T) {
	// The first vector of exercises/ch08-receiver/data/signature-vectors.json.
	body := `{"id":"6f1c2a9e-3b7d-4c1a-9e2f-1a2b3c4d5e6f","type":"invoice.paid","data":{"amount_cents":1999}}`
	want := "t=1791201600,v1=57665317204099d1fd6e02057a3c1c93560ada672e3e5aaa1ca693b7182e6189"
	if got := linkworker.Sign([]byte("whsec_change-me"), 1791201600, []byte(body)); got != want {
		t.Errorf("Sign = %s\nwant   %s", got, want)
	}
}

func keys(m map[string]tracetest.SpanStub) []string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	return out
}
