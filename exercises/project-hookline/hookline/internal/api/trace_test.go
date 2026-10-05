package api_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"example.com/hookline/internal/oteltest"
)

// TestIngestStoresTheTraceOfTheRequest: the event remembers the server span's trace, so the relay, the
// dispatcher and the workers can continue it (milestone M8). A client's traceparent is the parent.
func TestIngestStoresTheTraceOfTheRequest(t *testing.T) {
	spans := oteltest.Traces(t)
	hs := newHarness(t, false)
	key := hs.tenant("Acme")
	const traceID = "4bf92f3577b34da6a3ce929d0e0e4736"
	rec := hs.do(req{Method: "POST", Path: "/v1/events", Key: key, Body: paid, Header: map[string]string{
		"Idempotency-Key": "traced-1", "traceparent": oteltest.Traceparent(traceID, "00f067aa0ba902b7"),
	}})
	if rec.Code != http.StatusAccepted {
		t.Fatalf("ingest = %d %s", rec.Code, rec.Body)
	}
	id := decode(t, rec)["id"].(string)
	ev, err := hs.store.GetEvent(context.Background(), tenantID(t, hs, key), id)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(ev.TraceContext, "00-"+traceID+"-") {
		t.Fatalf("stored trace context %q, want the client's trace %s", ev.TraceContext, traceID)
	}
	var server bool
	for _, s := range spans.Ended() {
		if s.Name() == "POST /v1/events" && s.SpanContext().TraceID().String() == traceID {
			server = true
		}
	}
	if !server {
		var names []string
		for _, s := range spans.Ended() {
			names = append(names, s.Name())
		}
		t.Fatalf("no server span named after the route in trace %s; got %v", traceID, names)
	}
}

func TestHealthIsNotTraced(t *testing.T) {
	spans := oteltest.Traces(t)
	hs := newHarness(t, false)
	hs.do(req{Method: "GET", Path: "/healthz"})
	if n := len(spans.Ended()); n != 0 {
		t.Fatalf("/healthz made %d spans; probes would flood the traces", n)
	}
}

func tenantID(t *testing.T, hs *harness, key string) string {
	t.Helper()
	return decode(t, hs.do(req{Method: "GET", Path: "/v1/tenant", Key: key}))["id"].(string)
}
