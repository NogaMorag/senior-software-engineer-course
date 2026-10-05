package api_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"slices"
	"sync"
	"testing"
)

const paid = `{"type":"invoice.paid","payload":{"invoice":"inv_1","amount":4200}}`

func TestIngestAcceptsAndStoresAnEvent(t *testing.T) {
	hs := newHarness(t, false)
	key := hs.tenant("Acme")
	rec := hs.ingest(key, "evt-1", paid)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("POST /v1/events = %d %s", rec.Code, rec.Body)
	}
	if rec.Header().Get("Idempotent-Replayed") != "" {
		t.Error("a first request is not a replay")
	}
	ev := decode(t, rec)
	id := ev["id"].(string)
	if rec.Header().Get("Location") != "/v1/events/"+id || ev["type"] != "invoice.paid" {
		t.Fatalf("event = %v, Location %q", ev, rec.Header().Get("Location"))
	}
	got := decode(t, hs.do(req{Method: "GET", Path: "/v1/events/" + id, Key: key}))
	want := map[string]any{"invoice": "inv_1", "amount": float64(4200)}
	if got["id"] != id || !reflect.DeepEqual(got["payload"], want) || got["created_at"] != ev["created_at"] {
		t.Fatalf("GET = %v", got)
	}
}

func TestIngestIdempotency(t *testing.T) {
	hs := newHarness(t, false)
	key := hs.tenant("Acme")
	first := decode(t, hs.ingest(key, "evt-1", paid))

	retry := hs.ingest(key, "evt-1", paid)
	if retry.Code != http.StatusAccepted || retry.Header().Get("Idempotent-Replayed") != "true" {
		t.Fatalf("retry = %d, replayed header %q", retry.Code, retry.Header().Get("Idempotent-Replayed"))
	}
	if again := decode(t, retry); !reflect.DeepEqual(again, first) {
		t.Fatalf("retry returned %v, want the original %v", again, first)
	}

	changed := `{"type":"invoice.paid","payload":{"invoice":"inv_1","amount":9900}}`
	wantProblem(t, hs.ingest(key, "evt-1", changed), http.StatusUnprocessableEntity, "idempotency-key-reused")

	// The key is scoped to the tenant: another tenant's evt-1 is its own event.
	other := decode(t, hs.ingest(hs.tenant("Beta"), "evt-1", paid))
	if other["id"] == first["id"] {
		t.Fatal("two tenants share an idempotency key space")
	}
}

func TestIngestConcurrentRetriesGetOneEvent(t *testing.T) {
	hs := newHarness(t, false)
	key := hs.tenant("Acme")
	const n = 12
	recs := make([]*httptest.ResponseRecorder, n)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := range n {
		wg.Go(func() {
			<-start
			recs[i] = hs.ingest(key, "same-key", paid)
		})
	}
	close(start)
	wg.Wait()
	ids, fresh := map[string]bool{}, 0
	for _, rec := range recs {
		if rec.Code != http.StatusAccepted {
			t.Fatalf("concurrent ingest = %d %s", rec.Code, rec.Body)
		}
		ids[decode(t, rec)["id"].(string)] = true
		if rec.Header().Get("Idempotent-Replayed") == "" {
			fresh++
		}
	}
	if len(ids) != 1 || fresh != 1 {
		t.Fatalf("%d concurrent requests: %d event ids, %d first responses; want 1 and 1", n, len(ids), fresh)
	}
}

func TestIngestRejects(t *testing.T) {
	tests := []struct {
		name, key, body string
		status          int
		slug            string
		pointers        []string
	}{
		{"no idempotency key", "", paid, 400, "idempotency-key-missing", nil},
		{"key with a space", "a b", paid, 400, "idempotency-key-missing", nil},
		{"bad type and missing payload", "k", `{"type":"Invoice Paid"}`, 400, "validation", []string{"#/type", "#/payload"}},
		{"wildcard is not an event type", "k", `{"type":"*","payload":{}}`, 400, "validation", []string{"#/type"}},
		{"payload not an object", "k", `{"type":"invoice.paid","payload":[1,2]}`, 400, "validation", []string{"#/payload"}},
		{"payload null", "k", `{"type":"invoice.paid","payload":null}`, 400, "validation", []string{"#/payload"}},
		{"unknown field", "k", `{"type":"invoice.paid","payload":{},"tenant_id":"x"}`, 400, "validation", []string{"#/tenant_id"}},
		{"malformed", "k", `{"type":`, 400, "malformed-json", nil},
		{"too large", "k", fmt.Sprintf(`{"type":"a.b","payload":{"x":%q}}`, make([]byte, 5000)), 413, "body-too-large", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hs := newHarness(t, false)
			key := hs.tenant("Acme")
			r := req{Method: "POST", Path: "/v1/events", Key: key, Body: tt.body}
			if tt.key != "" {
				r.Header = map[string]string{"Idempotency-Key": tt.key}
			}
			p := wantProblem(t, hs.do(r), tt.status, tt.slug)
			if tt.pointers != nil && !slices.Equal(pointers(p), tt.pointers) {
				t.Errorf("pointers = %v, want %v", pointers(p), tt.pointers)
			}
		})
	}
}

func TestIngestAcceptsAQuotedKeyAndWhitespace(t *testing.T) {
	hs := newHarness(t, false)
	key := hs.tenant("Acme")
	body := "{\n  \"type\": \"invoice.paid\",\n  \"payload\":   {\"n\": 1}\n}"
	first := hs.ingest(key, `"8e03978e-40d5-43e8-bc93-6894a57f9324"`, body)
	if first.Code != http.StatusAccepted {
		t.Fatalf("= %d %s", first.Code, first.Body)
	}
	var ev struct {
		Payload json.RawMessage `json:"payload"`
	}
	_ = json.Unmarshal(first.Body.Bytes(), &ev)
	var payload map[string]any
	if err := json.Unmarshal(ev.Payload, &payload); err != nil || payload["n"] != float64(1) {
		t.Fatalf("payload = %s", ev.Payload)
	}
}
