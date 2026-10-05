// Package memory is the in-memory implementation of store.Store: milestone M1, dev mode and fast tests.
// It is safe for concurrent use. Lists sort on every call (O(n log n)); fine for development, which is
// all it is for.
package memory

import (
	"context"

	"sync"

	"example.com/hookline/internal/domain"
	"example.com/hookline/internal/page"
	"example.com/hookline/internal/store"
)

// Store keeps every resource in maps behind one mutex.
type Store struct {
	mu            sync.Mutex
	tenants       map[string]domain.Tenant
	byKeyHash     map[domain.APIKeyHash]string
	endpoints     map[string]domain.Endpoint
	subscriptions map[string]domain.Subscription
	events        map[string]domain.Event
	byIdemKey     map[idemKey]string // (tenant, idempotency key) -> event id: the unique constraint of M2
}

type idemKey struct{ tenantID, key string }

var _ store.Store = (*Store)(nil)

// New returns an empty store.
func New() *Store {
	panic("TODO: see chapter 38")

}

// CreateTenant adds t; a second tenant with the same key hash is domain.ErrConflict.
func (s *Store) CreateTenant(_ context.Context, t domain.Tenant) error {
	panic("TODO: see chapter 38")

}

// TenantByKeyHash implements store.Tenants.
func (s *Store) TenantByKeyHash(_ context.Context, h domain.APIKeyHash) (domain.Tenant, error) {
	panic("TODO: see chapter 38")

}

// CreateEndpoint implements store.Endpoints.
func (s *Store) CreateEndpoint(_ context.Context, e domain.Endpoint) error {
	panic("TODO: see chapter 38")

}

// GetEndpoint implements store.Endpoints.
func (s *Store) GetEndpoint(_ context.Context, tenantID, id string) (domain.Endpoint, error) {
	panic("TODO: see chapter 38")

}

// ListEndpoints implements store.Endpoints.
func (s *Store) ListEndpoints(_ context.Context, tenantID string, after *page.Cursor, limit int) ([]domain.Endpoint, bool, error) {
	panic("TODO: see chapter 38")

}

// DisableEndpoint implements store.Endpoints.
func (s *Store) DisableEndpoint(_ context.Context, tenantID, id string) (domain.Endpoint, error) {
	panic("TODO: see chapter 38")

}

// CreateSubscription implements store.Subscriptions.
func (s *Store) CreateSubscription(_ context.Context, sub domain.Subscription) error {
	panic("TODO: see chapter 38")

}

// GetSubscription implements store.Subscriptions.
func (s *Store) GetSubscription(_ context.Context, tenantID, id string) (domain.Subscription, error) {
	panic("TODO: see chapter 38")

}

// ListSubscriptions implements store.Subscriptions.
func (s *Store) ListSubscriptions(_ context.Context, tenantID, endpointID string, after *page.Cursor, limit int) ([]domain.Subscription, bool, error) {
	panic("TODO: see chapter 38")

}

// DeleteSubscription implements store.Subscriptions.
func (s *Store) DeleteSubscription(_ context.Context, tenantID, id string) error {
	panic("TODO: see chapter 38")

}

// IngestEvent implements store.Events. The mutex makes check-and-insert atomic, as the unique
// constraint on (tenant_id, idempotency_key) does in PostgreSQL.
func (s *Store) IngestEvent(_ context.Context, e domain.Event) (domain.Event, bool, error) {
	panic("TODO: see chapter 38")

}

// GetEvent implements store.Events.
func (s *Store) GetEvent(_ context.Context, tenantID, id string) (domain.Event, error) {
	panic("TODO: see chapter 38")

}

// pageOf selects the values that keep accepts, sorts them newest first by (CreatedAt, ID), and returns
// the page after the cursor. It is the in-memory twin of the keyset query in package postgres.
func pageOf[T any](all map[string]T, keep func(T) bool, key func(T) page.Cursor, after *page.Cursor, limit int) ([]T, bool) {
	panic("TODO: see chapter 38")

}

// compare orders cursors by CreatedAt, then ID, ascending.
func compare(a, b page.Cursor) int {
	panic("TODO: see chapter 38")

}

// less reports whether a sorts strictly before b in ascending order, the Go form of (a.t, a.id) < (b.t, b.id).
func less(a, b page.Cursor) bool { panic("TODO: see chapter 38") }
