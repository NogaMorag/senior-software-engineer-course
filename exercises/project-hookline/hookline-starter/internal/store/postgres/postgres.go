// Package postgres implements store.Store on PostgreSQL with pgx (milestone M2). It satisfies the same
// interfaces and passes the same conformance suite as the memory store; the differences are that rows
// survive a restart, several API replicas share them, and the database's constraints, not a Go mutex,
// decide duplicates.
package postgres

import (
	"context"

	"time"

	"github.com/jackc/pgx/v5"

	"github.com/jackc/pgx/v5/pgxpool"

	"example.com/hookline/internal/domain"
	"example.com/hookline/internal/page"
	"example.com/hookline/internal/store"
)

// SQLSTATE codes this package turns into domain errors.
const (
	uniqueViolation     = "23505"
	foreignKeyViolation = "23503"
)

// Store is the PostgreSQL store. The pool is shared and safe for concurrent use.
type Store struct {
	pool *pgxpool.Pool
}

var _ store.Store = (*Store)(nil)

// New returns a store over pool; the schema must already be migrated (hookline-admin migrate).
func New(pool *pgxpool.Pool) *Store { panic("TODO: see chapter 38") }

func sqlState(err error) string {
	panic("TODO: see chapter 38")

}

// one maps "no rows" to domain.ErrNotFound and wraps anything else with what failed.
func one(err error, what string) error {
	panic("TODO: see chapter 38")

}

// CreateTenant implements store.Tenants; a key hash that exists already is domain.ErrConflict.
func (s *Store) CreateTenant(ctx context.Context, t domain.Tenant) error {
	panic("TODO: see chapter 38")

}

// TenantByKeyHash implements store.Tenants. The lookup uses the unique index on api_key_hash.
func (s *Store) TenantByKeyHash(ctx context.Context, h domain.APIKeyHash) (domain.Tenant, error) {
	panic("TODO: see chapter 38")

}

const endpointColumns = `id, tenant_id, url, secret, enabled, created_at`

func scanEndpoint(row pgx.Row) (domain.Endpoint, error) {
	panic("TODO: see chapter 38")

}

// CreateEndpoint implements store.Endpoints.
func (s *Store) CreateEndpoint(ctx context.Context, e domain.Endpoint) error {
	panic("TODO: see chapter 38")

}

// GetEndpoint implements store.Endpoints. tenant_id is part of every WHERE: that is the BOLA guard.
func (s *Store) GetEndpoint(ctx context.Context, tenantID, id string) (domain.Endpoint, error) {
	panic("TODO: see chapter 38")

}

// ListEndpoints implements store.Endpoints with keyset pagination: the row comparison
// (created_at, id) < ($2, $3) matches the order of endpoints_tenant_page_idx, so PostgreSQL reads only
// the rows of the page, however deep the page is. It asks for limit+1 rows to learn whether more exist.
func (s *Store) ListEndpoints(ctx context.Context, tenantID string, after *page.Cursor, limit int) ([]domain.Endpoint, bool, error) {
	panic("TODO: see chapter 38")

}

// DisableEndpoint implements store.Endpoints.
func (s *Store) DisableEndpoint(ctx context.Context, tenantID, id string) (domain.Endpoint, error) {
	panic("TODO: see chapter 38")

}

const subscriptionColumns = `id, tenant_id, endpoint_id, event_type, created_at`

func scanSubscription(row pgx.Row) (domain.Subscription, error) {
	panic("TODO: see chapter 38")

}

// CreateSubscription implements store.Subscriptions. The composite foreign key
// (tenant_id, endpoint_id) -> endpoints (tenant_id, id) rejects another tenant's endpoint, and the unique
// constraint rejects a second subscription for the same type: no read-then-write race is possible.
func (s *Store) CreateSubscription(ctx context.Context, sub domain.Subscription) error {
	panic("TODO: see chapter 38")

}

// GetSubscription implements store.Subscriptions.
func (s *Store) GetSubscription(ctx context.Context, tenantID, id string) (domain.Subscription, error) {
	panic("TODO: see chapter 38")

}

// ListSubscriptions implements store.Subscriptions; an empty endpointID means every endpoint.
func (s *Store) ListSubscriptions(ctx context.Context, tenantID, endpointID string, after *page.Cursor, limit int) ([]domain.Subscription, bool, error) {
	panic("TODO: see chapter 38")

}

// DeleteSubscription implements store.Subscriptions.
func (s *Store) DeleteSubscription(ctx context.Context, tenantID, id string) error {
	panic("TODO: see chapter 38")

}

// ActiveSubscriptions implements store.Subscriptions: the join keeps only enabled endpoints.
func (s *Store) ActiveSubscriptions(ctx context.Context, tenantID string) ([]domain.Subscription, error) {
	panic("TODO: see chapter 38")

}

// IngestEvent implements store.Events in one read-committed transaction:
//
//  1. INSERT ... ON CONFLICT (tenant_id, idempotency_key) DO NOTHING. A concurrent insert of the same key
//     makes this statement wait on the unique index until the other transaction ends.
//  2. A returned row means the event is new: insert its outbox row and commit both together.
//  3. No row means the key exists. A new statement sees the committed winner; compare body hashes.
//
// The response is a function of the stored row alone, so storing the event is storing the response.
func (s *Store) IngestEvent(ctx context.Context, e domain.Event) (domain.Event, bool, error) {
	panic("TODO: see chapter 38")

}

const eventColumns = `id, tenant_id, type, payload, idempotency_key, body_hash, created_at`

func scanEvent(row pgx.Row) (domain.Event, error) {
	panic("TODO: see chapter 38")

}

// GetEvent implements store.Events.
func (s *Store) GetEvent(ctx context.Context, tenantID, id string) (domain.Event, error) {
	panic("TODO: see chapter 38")

}

const deliveryColumns = `id, tenant_id, event_id, endpoint_id, status, attempts, attempt_base, next_attempt_at,
	COALESCE(last_error, ''), delivered_at, created_at`

func scanDelivery(row pgx.Row) (domain.Delivery, error) {
	panic("TODO: see chapter 38")

}

// EventDeliveries implements store.Deliveries: the deliveries of one event, then every attempt of
// those deliveries in a second query, attached in Go.
func (s *Store) EventDeliveries(ctx context.Context, tenantID, eventID string) ([]domain.Delivery, error) {
	panic("TODO: see chapter 38")

}

// DeadDeliveries implements store.Deliveries over the partial index deliveries_dead_page_idx.
func (s *Store) DeadDeliveries(ctx context.Context, tenantID string, after *page.Cursor, limit int) ([]domain.Delivery, bool, error) {
	panic("TODO: see chapter 38")

}

// ReplayDelivery implements store.Deliveries. attempts is not reset: attempt_base remembers where the new
// round starts, so attempts.n stays unique and Hookline-Attempt keeps rising (chapter 29, section 29.5).
// The status condition is in the UPDATE itself, so two replays at once move the delivery once; the
// second finds it pending and gets domain.ErrNotDead.
func (s *Store) ReplayDelivery(ctx context.Context, tenantID, id string) (domain.Delivery, error) {
	panic("TODO: see chapter 38")

}

// cursorArgs turns an optional cursor into the two SQL parameters; NULL means "from the start".
func cursorArgs(after *page.Cursor) (*time.Time, *string) {
	panic("TODO: see chapter 38")

}

// trim drops the extra row that was fetched only to answer "are there more?".
func trim[T any](items []T, limit int) ([]T, bool, error) {
	panic("TODO: see chapter 38")

}
