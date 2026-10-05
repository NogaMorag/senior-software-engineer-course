package api_test

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"example.com/hookline/internal/dispatch"
	"example.com/hookline/internal/outbox"
	"example.com/hookline/internal/testdb"
)

// deliveriesFor fans an ingested event out to the endpoints, as the dispatcher would. Deliveries exist
// only in PostgreSQL, so these tests skip on the memory store.
func deliveriesFor(t *testing.T, tenantID, eventID string, endpointIDs ...string) {
	t.Helper()
	pool := testdb.Available()
	if _, err := dispatch.Fanout(context.Background(), pool, outbox.Message{EventID: eventID, TenantID: tenantID, Type: "invoice.paid"}, endpointIDs); err != nil {
		t.Fatal(err)
	}
}

func requirePostgres(t *testing.T) {
	t.Helper()
	if testdb.Available() == nil {
		t.Skip("integration test: deliveries live in PostgreSQL; set DATABASE_URL (make up)")
	}
}

// kill marks every delivery of the event dead, the state a worker leaves after the last attempt. It is
// a fixture shortcut; package delivery tests the real path.
func kill(t *testing.T, eventID string) {
	t.Helper()
	if _, err := testdb.Available().Exec(context.Background(), `
		UPDATE deliveries SET status = 'dead', attempts = 3, last_error = 'receiver answered 503' WHERE event_id = $1`, eventID); err != nil {
		t.Fatal(err)
	}
}

func TestEventDeliveriesDeadListAndReplay(t *testing.T) {
	requirePostgres(t)
	hs := newHarness(t, false)
	key := hs.tenant("Acme")
	tenantID := decode(t, hs.do(req{Method: "GET", Path: "/v1/tenant", Key: key}))["id"].(string)
	e1 := hs.createEndpoint(key, "https://one.example/hooks")["id"].(string)
	e2 := hs.createEndpoint(key, "https://two.example/hooks")["id"].(string)
	ev := decode(t, hs.ingest(key, "evt-1", paid))["id"].(string)

	empty := decode(t, hs.do(req{Method: "GET", Path: "/v1/events/" + ev + "/deliveries", Key: key}))
	if n := len(empty["items"].([]any)); n != 0 {
		t.Fatalf("before the dispatcher: %d deliveries", n)
	}
	deliveriesFor(t, tenantID, ev, e1, e2)
	list := decode(t, hs.do(req{Method: "GET", Path: "/v1/events/" + ev + "/deliveries", Key: key}))["items"].([]any)
	if len(list) != 2 || list[0].(map[string]any)["status"] != "pending" {
		t.Fatalf("deliveries = %v", list)
	}

	kill(t, ev)
	page1 := decode(t, hs.do(req{Method: "GET", Path: "/v1/deliveries?status=dead&limit=1", Key: key}))
	next, _ := page1["next_cursor"].(string)
	if len(page1["items"].([]any)) != 1 || next == "" {
		t.Fatalf("first dead page = %v", page1)
	}
	page2 := decode(t, hs.do(req{Method: "GET", Path: "/v1/deliveries?status=dead&limit=1&cursor=" + url.QueryEscape(next), Key: key}))
	if len(page2["items"].([]any)) != 1 || page2["next_cursor"] != nil {
		t.Fatalf("second dead page = %v", page2)
	}

	id := page1["items"].([]any)[0].(map[string]any)["id"].(string)
	rec := hs.do(req{Method: "POST", Path: "/v1/deliveries/" + id + "/replay", Key: key})
	if rec.Code != http.StatusAccepted {
		t.Fatalf("replay = %d %s", rec.Code, rec.Body)
	}
	d := decode(t, rec)
	if d["status"] != "pending" || d["attempts"] != float64(3) || d["attempt_base"] != float64(3) {
		t.Fatalf("replayed = %v; want pending, attempts kept at 3, attempt_base 3", d)
	}
	wantProblem(t, hs.do(req{Method: "POST", Path: "/v1/deliveries/" + id + "/replay", Key: key}), http.StatusConflict, "delivery-not-dead")
}

func TestDeliveryRoutesValidateAndIsolate(t *testing.T) {
	hs := newHarness(t, false)
	a, b := hs.tenant("A"), hs.tenant("B")
	p := wantProblem(t, hs.do(req{Method: "GET", Path: "/v1/deliveries", Key: a}), http.StatusBadRequest, "validation")
	if ptrs := pointers(p); len(ptrs) != 1 || ptrs[0] != "query:status" {
		t.Errorf("pointers = %v", ptrs)
	}
	ev := decode(t, hs.ingest(a, "evt-a", paid))["id"].(string)
	wantProblem(t, hs.do(req{Method: "GET", Path: "/v1/events/" + ev + "/deliveries", Key: b}), http.StatusNotFound, "not-found")
	wantProblem(t, hs.do(req{Method: "POST", Path: "/v1/deliveries/0199b0a2-6c1e-7a00-8000-000000000042/replay", Key: a}), http.StatusNotFound, "not-found")
	wantProblem(t, hs.do(req{Method: "POST", Path: "/v1/deliveries/nope/replay", Key: a}), http.StatusNotFound, "not-found")
	if items := decode(t, hs.do(req{Method: "GET", Path: "/v1/deliveries?status=dead", Key: b}))["items"].([]any); len(items) != 0 {
		t.Errorf("B lists %d dead deliveries", len(items))
	}
}

// TestReplayIsTenantScoped: B cannot replay A's dead delivery, and learns nothing about it.
func TestReplayIsTenantScoped(t *testing.T) {
	requirePostgres(t)
	hs := newHarness(t, false)
	a, b := hs.tenant("A"), hs.tenant("B")
	tenantA := decode(t, hs.do(req{Method: "GET", Path: "/v1/tenant", Key: a}))["id"].(string)
	ep := hs.createEndpoint(a, "https://one.example/hooks")["id"].(string)
	ev := decode(t, hs.ingest(a, "evt-1", paid))["id"].(string)
	deliveriesFor(t, tenantA, ev, ep)
	kill(t, ev)
	id := decode(t, hs.do(req{Method: "GET", Path: "/v1/deliveries?status=dead", Key: a}))["items"].([]any)[0].(map[string]any)["id"].(string)
	wantProblem(t, hs.do(req{Method: "POST", Path: "/v1/deliveries/" + id + "/replay", Key: b}), http.StatusNotFound, "not-found")
}
