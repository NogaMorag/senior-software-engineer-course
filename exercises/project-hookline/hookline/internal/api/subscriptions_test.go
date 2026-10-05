package api_test

import (
	"net/http"
	"slices"
	"testing"
)

func TestSubscriptions(t *testing.T) {
	hs := newHarness(t, false)
	key := hs.tenant("Acme")
	ep := hs.createEndpoint(key, "https://example.com/hooks")["id"].(string)
	exact := hs.createSubscription(key, ep, "invoice.paid")
	all := hs.createSubscription(key, ep, "*")
	if exact["endpoint_id"] != ep || exact["event_type"] != "invoice.paid" || all["event_type"] != "*" {
		t.Fatalf("created %v and %v", exact, all)
	}

	dup := hs.do(req{Method: "POST", Path: "/v1/subscriptions", Key: key, Body: `{"endpoint_id":"` + ep + `","event_type":"invoice.paid"}`})
	wantProblem(t, dup, http.StatusConflict, "conflict")

	id := exact["id"].(string)
	got := decode(t, hs.do(req{Method: "GET", Path: "/v1/subscriptions/" + id, Key: key}))
	if got["id"] != id || got["event_type"] != "invoice.paid" {
		t.Fatalf("GET = %v", got)
	}
	if rec := hs.do(req{Method: "DELETE", Path: "/v1/subscriptions/" + id, Key: key}); rec.Code != http.StatusNoContent || rec.Body.Len() != 0 {
		t.Fatalf("DELETE = %d %q", rec.Code, rec.Body)
	}
	wantProblem(t, hs.do(req{Method: "DELETE", Path: "/v1/subscriptions/" + id, Key: key}), http.StatusNotFound, "not-found")
	wantProblem(t, hs.do(req{Method: "GET", Path: "/v1/subscriptions/" + id, Key: key}), http.StatusNotFound, "not-found")
}

func TestCreateSubscriptionValidation(t *testing.T) {
	hs := newHarness(t, false)
	key := hs.tenant("Acme")
	ep := hs.createEndpoint(key, "https://example.com/hooks")["id"].(string)
	tests := []struct {
		name, body string
		pointers   []string
	}{
		{"both missing", `{}`, []string{"#/endpoint_id", "#/event_type"}},
		{"bad type", `{"endpoint_id":"` + ep + `","event_type":"Invoice Paid"}`, []string{"#/event_type"}},
		{"partial wildcard", `{"endpoint_id":"` + ep + `","event_type":"invoice.*"}`, []string{"#/event_type"}},
		{"bad endpoint id", `{"endpoint_id":"ep_1","event_type":"invoice.paid"}`, []string{"#/endpoint_id"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := wantProblem(t, hs.do(req{Method: "POST", Path: "/v1/subscriptions", Key: key, Body: tt.body}), 400, "validation")
			if !slices.Equal(pointers(p), tt.pointers) {
				t.Errorf("pointers = %v, want %v (every invalid field, in one response)", pointers(p), tt.pointers)
			}
		})
	}
	unknown := `{"endpoint_id":"0199b0a2-6c1e-7a00-8000-000000000042","event_type":"invoice.paid"}`
	wantProblem(t, hs.do(req{Method: "POST", Path: "/v1/subscriptions", Key: key, Body: unknown}), http.StatusNotFound, "not-found")
}

func TestListSubscriptionsByEndpoint(t *testing.T) {
	hs := newHarness(t, false)
	key := hs.tenant("Acme")
	e1 := hs.createEndpoint(key, "https://one.example/hooks")["id"].(string)
	e2 := hs.createEndpoint(key, "https://two.example/hooks")["id"].(string)
	for _, typ := range []string{"a.one", "a.two", "a.three"} {
		hs.createSubscription(key, e1, typ)
	}
	hs.createSubscription(key, e2, "*")

	body := decode(t, hs.do(req{Method: "GET", Path: "/v1/subscriptions?endpoint_id=" + e1 + "&limit=2", Key: key}))
	items := body["items"].([]any)
	if len(items) != 2 || body["next_cursor"] == nil {
		t.Fatalf("first page = %v", body)
	}
	for _, it := range items {
		if it.(map[string]any)["endpoint_id"] != e1 {
			t.Fatalf("filter by endpoint_id returned %v", it)
		}
	}
	all := decode(t, hs.do(req{Method: "GET", Path: "/v1/subscriptions", Key: key}))
	if n := len(all["items"].([]any)); n != 4 {
		t.Errorf("all subscriptions = %d, want 4", n)
	}
	p := wantProblem(t, hs.do(req{Method: "GET", Path: "/v1/subscriptions?endpoint_id=nope", Key: key}), 400, "validation")
	if !slices.Equal(pointers(p), []string{"query:endpoint_id"}) {
		t.Errorf("pointers = %v", pointers(p))
	}
}
