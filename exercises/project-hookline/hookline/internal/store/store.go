// Package store defines what the API needs from storage, as interfaces. Milestone M1 implements them in
// memory (package memory); milestone M2 implements the same interfaces on PostgreSQL (package postgres),
// and the same conformance suite (package storetest) checks both.
//
// Every method that reads or changes a tenant's resource takes the tenant id and filters by it. A store
// never returns another tenant's row: for the caller it does not exist (domain.ErrNotFound).
package store

import (
	"context"

	"example.com/hookline/internal/domain"
	"example.com/hookline/internal/page"
)

// Tenants are created by the operator (hookline-admin) and looked up by the auth middleware.
type Tenants interface {
	CreateTenant(ctx context.Context, t domain.Tenant) error
	// TenantByKeyHash returns the tenant whose API key hashes to h, or domain.ErrNotFound.
	TenantByKeyHash(ctx context.Context, h domain.APIKeyHash) (domain.Tenant, error)
}

// Endpoints are a tenant's receiver URLs.
type Endpoints interface {
	CreateEndpoint(ctx context.Context, e domain.Endpoint) error
	GetEndpoint(ctx context.Context, tenantID, id string) (domain.Endpoint, error)
	// ListEndpoints returns up to limit endpoints newest first, strictly after the cursor (from the
	// start when after is nil), and whether more exist.
	ListEndpoints(ctx context.Context, tenantID string, after *page.Cursor, limit int) ([]domain.Endpoint, bool, error)
	// DisableEndpoint sets enabled to false and returns the endpoint. Disabling twice is not an error.
	DisableEndpoint(ctx context.Context, tenantID, id string) (domain.Endpoint, error)
}

// Subscriptions connect an endpoint to an event type.
type Subscriptions interface {
	// CreateSubscription returns domain.ErrNotFound when the endpoint is not the tenant's, and
	// domain.ErrConflict when the endpoint already has a subscription for the event type.
	CreateSubscription(ctx context.Context, s domain.Subscription) error
	GetSubscription(ctx context.Context, tenantID, id string) (domain.Subscription, error)
	// ListSubscriptions is like ListEndpoints; a non-empty endpointID keeps only that endpoint's.
	ListSubscriptions(ctx context.Context, tenantID, endpointID string, after *page.Cursor, limit int) ([]domain.Subscription, bool, error)
	DeleteSubscription(ctx context.Context, tenantID, id string) error
	// ActiveSubscriptions returns every subscription of the tenant whose endpoint is enabled: what the
	// dispatcher (milestone M4) matches events against, through the cache of milestone M3.
	ActiveSubscriptions(ctx context.Context, tenantID string) ([]domain.Subscription, error)
}

// Events are ingested once per (tenant, idempotency key).
type Events interface {
	// IngestEvent stores e unless the tenant already used e.IdempotencyKey. Then, with the same
	// BodyHash, it returns the stored event and replayed=true; with another BodyHash it returns
	// domain.ErrIdempotencyMismatch. The decision is atomic: concurrent calls with one key store
	// exactly one event and all return it. The PostgreSQL store also writes the outbox row (M4) in the
	// same transaction.
	IngestEvent(ctx context.Context, e domain.Event) (stored domain.Event, replayed bool, err error)
	GetEvent(ctx context.Context, tenantID, id string) (domain.Event, error)
}

// Deliveries are created by the dispatcher and moved by the workers (package delivery); the API
// only reads them and replays dead ones (milestone M5).
type Deliveries interface {
	// EventDeliveries returns the deliveries of one event with their attempt history, or
	// domain.ErrNotFound when the event is not the tenant's.
	EventDeliveries(ctx context.Context, tenantID, eventID string) ([]domain.Delivery, error)
	// DeadDeliveries pages through the tenant's dead deliveries, newest first.
	DeadDeliveries(ctx context.Context, tenantID string, after *page.Cursor, limit int) ([]domain.Delivery, bool, error)
	// ReplayDelivery makes a dead delivery pending again with a fresh budget of attempts, without
	// resetting the attempt count (it sets attempt_base). It returns
	// domain.ErrNotFound for an unknown id and domain.ErrNotDead for a delivery in any other status.
	ReplayDelivery(ctx context.Context, tenantID, id string) (domain.Delivery, error)
}

// Store is everything the API uses.
type Store interface {
	Tenants
	Endpoints
	Subscriptions
	Events
	Deliveries
}
