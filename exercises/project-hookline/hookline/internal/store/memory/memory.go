// Package memory is the in-memory implementation of store.Store: milestone M1, dev mode and fast tests.
// It is safe for concurrent use. Lists sort on every call (O(n log n)); fine for development, which is
// all it is for.
package memory

import (
	"cmp"
	"context"
	"slices"
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
	deliveries    map[string]domain.Delivery // only a test or a future in-process dispatcher adds any
	byIdemKey     map[idemKey]string         // (tenant, idempotency key) -> event id: the unique constraint of M2
}

type idemKey struct{ tenantID, key string }

var _ store.Store = (*Store)(nil)

// New returns an empty store.
func New() *Store {
	return &Store{
		tenants:       map[string]domain.Tenant{},
		byKeyHash:     map[domain.APIKeyHash]string{},
		endpoints:     map[string]domain.Endpoint{},
		subscriptions: map[string]domain.Subscription{},
		events:        map[string]domain.Event{},
		deliveries:    map[string]domain.Delivery{},
		byIdemKey:     map[idemKey]string{},
	}
}

// CreateTenant adds t; a second tenant with the same key hash is domain.ErrConflict.
func (s *Store) CreateTenant(_ context.Context, t domain.Tenant) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, dup := s.byKeyHash[t.APIKeyHash]; dup {
		return domain.ErrConflict
	}
	s.tenants[t.ID] = t
	s.byKeyHash[t.APIKeyHash] = t.ID
	return nil
}

// TenantByKeyHash implements store.Tenants.
func (s *Store) TenantByKeyHash(_ context.Context, h domain.APIKeyHash) (domain.Tenant, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id, ok := s.byKeyHash[h]
	if !ok {
		return domain.Tenant{}, domain.ErrNotFound
	}
	return s.tenants[id], nil
}

// CreateEndpoint implements store.Endpoints.
func (s *Store) CreateEndpoint(_ context.Context, e domain.Endpoint) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.endpoints[e.ID] = e
	return nil
}

// GetEndpoint implements store.Endpoints.
func (s *Store) GetEndpoint(_ context.Context, tenantID, id string) (domain.Endpoint, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.endpoints[id]
	if !ok || e.TenantID != tenantID {
		return domain.Endpoint{}, domain.ErrNotFound
	}
	return e, nil
}

// ListEndpoints implements store.Endpoints.
func (s *Store) ListEndpoints(_ context.Context, tenantID string, after *page.Cursor, limit int) ([]domain.Endpoint, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	items, more := pageOf(s.endpoints, func(e domain.Endpoint) bool { return e.TenantID == tenantID },
		func(e domain.Endpoint) page.Cursor { return page.Cursor{CreatedAt: e.CreatedAt, ID: e.ID} }, after, limit)
	return items, more, nil
}

// DisableEndpoint implements store.Endpoints.
func (s *Store) DisableEndpoint(_ context.Context, tenantID, id string) (domain.Endpoint, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.endpoints[id]
	if !ok || e.TenantID != tenantID {
		return domain.Endpoint{}, domain.ErrNotFound
	}
	e.Enabled = false
	s.endpoints[id] = e
	return e, nil
}

// CreateSubscription implements store.Subscriptions.
func (s *Store) CreateSubscription(_ context.Context, sub domain.Subscription) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if e, ok := s.endpoints[sub.EndpointID]; !ok || e.TenantID != sub.TenantID {
		return domain.ErrNotFound
	}
	for _, other := range s.subscriptions {
		if other.EndpointID == sub.EndpointID && other.EventType == sub.EventType {
			return domain.ErrConflict
		}
	}
	s.subscriptions[sub.ID] = sub
	return nil
}

// GetSubscription implements store.Subscriptions.
func (s *Store) GetSubscription(_ context.Context, tenantID, id string) (domain.Subscription, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sub, ok := s.subscriptions[id]
	if !ok || sub.TenantID != tenantID {
		return domain.Subscription{}, domain.ErrNotFound
	}
	return sub, nil
}

// ListSubscriptions implements store.Subscriptions.
func (s *Store) ListSubscriptions(_ context.Context, tenantID, endpointID string, after *page.Cursor, limit int) ([]domain.Subscription, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	keep := func(sub domain.Subscription) bool {
		return sub.TenantID == tenantID && (endpointID == "" || sub.EndpointID == endpointID)
	}
	items, more := pageOf(s.subscriptions, keep,
		func(sub domain.Subscription) page.Cursor { return page.Cursor{CreatedAt: sub.CreatedAt, ID: sub.ID} }, after, limit)
	return items, more, nil
}

// DeleteSubscription implements store.Subscriptions.
func (s *Store) DeleteSubscription(_ context.Context, tenantID, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	sub, ok := s.subscriptions[id]
	if !ok || sub.TenantID != tenantID {
		return domain.ErrNotFound
	}
	delete(s.subscriptions, id)
	return nil
}

// ActiveSubscriptions implements store.Subscriptions.
func (s *Store) ActiveSubscriptions(_ context.Context, tenantID string) ([]domain.Subscription, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	items, _ := pageOf(s.subscriptions, func(sub domain.Subscription) bool {
		return sub.TenantID == tenantID && s.endpoints[sub.EndpointID].Enabled
	}, func(sub domain.Subscription) page.Cursor { return page.Cursor{CreatedAt: sub.CreatedAt, ID: sub.ID} }, nil, len(s.subscriptions))
	return items, nil
}

// IngestEvent implements store.Events. The mutex makes check-and-insert atomic, as the unique
// constraint on (tenant_id, idempotency_key) does in PostgreSQL.
func (s *Store) IngestEvent(_ context.Context, e domain.Event) (domain.Event, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	k := idemKey{e.TenantID, e.IdempotencyKey}
	if id, ok := s.byIdemKey[k]; ok {
		stored := s.events[id]
		if stored.BodyHash != e.BodyHash {
			return domain.Event{}, false, domain.ErrIdempotencyMismatch
		}
		return stored, true, nil
	}
	e.Payload = slices.Clone(e.Payload) // the caller's buffer is not ours to keep
	s.events[e.ID] = e
	s.byIdemKey[k] = e.ID
	return e, false, nil
}

// GetEvent implements store.Events.
func (s *Store) GetEvent(_ context.Context, tenantID, id string) (domain.Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.events[id]
	if !ok || e.TenantID != tenantID {
		return domain.Event{}, domain.ErrNotFound
	}
	return e, nil
}

// The memory store has no dispatcher, so in dev mode no deliveries exist: these methods answer
// truthfully about an empty set. The PostgreSQL store is where deliveries live.

// EventDeliveries implements store.Deliveries.
func (s *Store) EventDeliveries(_ context.Context, tenantID, eventID string) ([]domain.Delivery, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if e, ok := s.events[eventID]; !ok || e.TenantID != tenantID {
		return nil, domain.ErrNotFound
	}
	items, _ := pageOf(s.deliveries, func(d domain.Delivery) bool { return d.EventID == eventID },
		func(d domain.Delivery) page.Cursor { return page.Cursor{CreatedAt: d.CreatedAt, ID: d.ID} }, nil, len(s.deliveries))
	return items, nil
}

// DeadDeliveries implements store.Deliveries.
func (s *Store) DeadDeliveries(_ context.Context, tenantID string, after *page.Cursor, limit int) ([]domain.Delivery, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	items, more := pageOf(s.deliveries, func(d domain.Delivery) bool { return d.TenantID == tenantID && d.Status == domain.Dead },
		func(d domain.Delivery) page.Cursor { return page.Cursor{CreatedAt: d.CreatedAt, ID: d.ID} }, after, limit)
	return items, more, nil
}

// ReplayDelivery implements store.Deliveries.
func (s *Store) ReplayDelivery(_ context.Context, tenantID, id string) (domain.Delivery, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	d, ok := s.deliveries[id]
	switch {
	case !ok || d.TenantID != tenantID:
		return domain.Delivery{}, domain.ErrNotFound
	case d.Status != domain.Dead:
		return domain.Delivery{}, domain.ErrNotDead
	}
	d.Status, d.AttemptBase, d.LastError = domain.Pending, d.Attempts, ""
	s.deliveries[id] = d
	return d, nil
}

// pageOf selects the values that keep accepts, sorts them newest first by (CreatedAt, ID), and returns
// the page after the cursor. It is the in-memory twin of the keyset query in package postgres.
func pageOf[T any](all map[string]T, keep func(T) bool, key func(T) page.Cursor, after *page.Cursor, limit int) ([]T, bool) {
	var items []T
	for _, v := range all {
		if keep(v) && (after == nil || less(key(v), *after)) {
			items = append(items, v)
		}
	}
	slices.SortFunc(items, func(a, b T) int { return compare(key(b), key(a)) }) // descending
	if len(items) > limit {
		return items[:limit], true
	}
	return items, false
}

// compare orders cursors by CreatedAt, then ID, ascending.
func compare(a, b page.Cursor) int {
	if c := a.CreatedAt.Compare(b.CreatedAt); c != 0 {
		return c
	}
	return cmp.Compare(a.ID, b.ID)
}

// less reports whether a sorts strictly before b in ascending order, the Go form of (a.t, a.id) < (b.t, b.id).
func less(a, b page.Cursor) bool { return compare(a, b) < 0 }
