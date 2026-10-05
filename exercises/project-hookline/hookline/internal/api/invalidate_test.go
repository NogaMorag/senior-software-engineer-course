package api_test

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"sync"
	"testing"
)

// recorder is an api.Invalidator that remembers which tenants it was told about.
type recorder struct {
	mu      sync.Mutex
	tenants []string
	fail    bool
}

func (r *recorder) Invalidate(_ context.Context, tenantID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tenants = append(r.tenants, tenantID)
	if r.fail {
		return errors.New("redis down")
	}
	return nil
}

func (r *recorder) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.tenants)
}

// TestChangesInvalidateTheSubscriptionCache: every committed change that can alter who receives an
// event (subscribe, unsubscribe, disable an endpoint) drops the tenant's cache entry; reads and failed
// writes do not.
func TestChangesInvalidateTheSubscriptionCache(t *testing.T) {
	rec := &recorder{}
	hs := newHarnessWith(t, false, rec)
	key := hs.tenant("Acme")
	tenantID := decode(t, hs.do(req{Method: "GET", Path: "/v1/tenant", Key: key}))["id"].(string)
	ep := hs.createEndpoint(key, "https://example.com/hooks")["id"].(string)
	if rec.count() != 0 {
		t.Fatalf("creating an endpoint without subscriptions invalidated %d times", rec.count())
	}
	sub := hs.createSubscription(key, ep, "invoice.paid")["id"].(string)
	hs.do(req{Method: "GET", Path: "/v1/subscriptions/" + sub, Key: key})
	hs.do(req{Method: "POST", Path: "/v1/subscriptions", Key: key, Body: `{"endpoint_id":"` + ep + `","event_type":"invoice.paid"}`}) // 409
	hs.do(req{Method: "DELETE", Path: "/v1/subscriptions/" + sub, Key: key})
	hs.do(req{Method: "POST", Path: "/v1/endpoints/" + ep + "/disable", Key: key})
	if rec.count() != 3 || !slices.Equal(rec.tenants, []string{tenantID, tenantID, tenantID}) {
		t.Fatalf("invalidations = %v, want the tenant three times (subscribe, delete, disable)", rec.tenants)
	}
}

func TestInvalidationFailureDoesNotFailTheRequest(t *testing.T) {
	hs := newHarnessWith(t, false, &recorder{fail: true})
	key := hs.tenant("Acme")
	ep := hs.createEndpoint(key, "https://example.com/hooks")["id"].(string)
	rec := hs.do(req{Method: "POST", Path: "/v1/subscriptions", Key: key, Body: `{"endpoint_id":"` + ep + `","event_type":"*"}`})
	if rec.Code != http.StatusCreated {
		t.Fatalf("subscribe with the cache down = %d %s; the change is committed and must succeed", rec.Code, rec.Body)
	}
}
