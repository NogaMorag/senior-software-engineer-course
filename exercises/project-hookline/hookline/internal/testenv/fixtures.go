package testenv

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"testing"
	"time"

	"example.com/hookline/internal/domain"
	"example.com/hookline/internal/store"
)

// SinkSecret is the secret the compose sink verifies every hook with (RECEIVER_DEFAULT_SECRET).
const SinkSecret = "whsec_change-me"

// Tenant creates a tenant through st.
func Tenant(t testing.TB, st store.Tenants) domain.Tenant {
	t.Helper()
	_, hash := domain.NewAPIKey()
	now := domain.Now(time.Now)
	tn := domain.Tenant{ID: domain.NewID(now), Name: "test", APIKeyHash: hash, RatePerSec: 50, Burst: 100, CreatedAt: now}
	if err := st.CreateTenant(context.Background(), tn); err != nil {
		t.Fatal(err)
	}
	return tn
}

// Endpoint creates an enabled endpoint with a known secret, as the API would but without generating
// the secret, so that a test receiver can verify the signature.
func Endpoint(t testing.TB, st store.Endpoints, tenantID, url, secret string) domain.Endpoint {
	t.Helper()
	now := domain.Now(time.Now)
	e := domain.Endpoint{ID: domain.NewID(now), TenantID: tenantID, URL: url, Secret: secret, Enabled: true, CreatedAt: now}
	if err := st.CreateEndpoint(context.Background(), e); err != nil {
		t.Fatal(err)
	}
	return e
}

// Subscribe subscribes the endpoint to eventType.
func Subscribe(t testing.TB, st store.Subscriptions, tenantID, endpointID, eventType string) {
	t.Helper()
	now := domain.Now(time.Now)
	s := domain.Subscription{ID: domain.NewID(now), TenantID: tenantID, EndpointID: endpointID, EventType: eventType, CreatedAt: now}
	if err := st.CreateSubscription(context.Background(), s); err != nil {
		t.Fatal(err)
	}
}

// Event ingests one event (and, in PostgreSQL, its outbox row) with a fresh idempotency key.
func Event(t testing.TB, st store.Events, tenantID, eventType string, payload any) domain.Event {
	t.Helper()
	p, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	now := domain.Now(time.Now)
	e := domain.Event{ID: domain.NewID(now), TenantID: tenantID, Type: eventType, Payload: p,
		IdempotencyKey: Unique("key"), BodyHash: sha256.Sum256(p), CreatedAt: now}
	stored, _, err := st.IngestEvent(context.Background(), e)
	if err != nil {
		t.Fatal(err)
	}
	return stored
}
