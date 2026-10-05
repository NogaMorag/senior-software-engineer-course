package api_test

import (
	"net/http"
	"testing"
)

// TestTenantIsolation is the broken-object-level-authorization suite (OWASP API1): tenant B holds valid
// ids of tenant A's resources and tries every route with them. Each answer must be the same 404 an
// unknown id gets, so B cannot even learn that the id exists, and A's resources must be unchanged.
func TestTenantIsolation(t *testing.T) {
	hs := newHarness(t, false)
	a, b := hs.tenant("A"), hs.tenant("B")
	ep := hs.createEndpoint(a, "https://a.example/hooks")["id"].(string)
	sub := hs.createSubscription(a, ep, "invoice.paid")["id"].(string)
	ev := decode(t, hs.ingest(a, "evt-a", paid))["id"].(string)

	for _, r := range []req{
		{Method: "GET", Path: "/v1/endpoints/" + ep},
		{Method: "POST", Path: "/v1/endpoints/" + ep + "/disable"},
		{Method: "GET", Path: "/v1/subscriptions/" + sub},
		{Method: "DELETE", Path: "/v1/subscriptions/" + sub},
		{Method: "GET", Path: "/v1/events/" + ev},
		{Method: "POST", Path: "/v1/subscriptions", Body: `{"endpoint_id":"` + ep + `","event_type":"*"}`},
		// A pair A already has: must still be 404, never 409, or B learns that it exists.
		{Method: "POST", Path: "/v1/subscriptions", Body: `{"endpoint_id":"` + ep + `","event_type":"invoice.paid"}`},
	} {
		t.Run(r.Method+" "+r.Path, func(t *testing.T) {
			r.Key = b
			wantProblem(t, hs.do(r), http.StatusNotFound, "not-found")
		})
	}

	for _, path := range []string{"/v1/endpoints", "/v1/subscriptions", "/v1/subscriptions?endpoint_id=" + ep} {
		body := decode(t, hs.do(req{Method: "GET", Path: path, Key: b}))
		if n := len(body["items"].([]any)); n != 0 {
			t.Errorf("B lists %d of A's resources at %s", n, path)
		}
	}

	// A's resources survived B's attempts.
	if decode(t, hs.do(req{Method: "GET", Path: "/v1/endpoints/" + ep, Key: a}))["enabled"] != true {
		t.Error("B disabled A's endpoint")
	}
	if rec := hs.do(req{Method: "GET", Path: "/v1/subscriptions/" + sub, Key: a}); rec.Code != http.StatusOK {
		t.Error("B deleted A's subscription")
	}
}
