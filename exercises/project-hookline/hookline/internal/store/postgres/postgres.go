// Package postgres implements store.Store on PostgreSQL with pgx (milestone M2). It satisfies the same
// interfaces and passes the same conformance suite as the memory store; the differences are that rows
// survive a restart, several API replicas share them, and the database's constraints, not a Go mutex,
// decide duplicates.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
func New(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

func sqlState(err error) string {
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
		return pgErr.Code
	}
	return ""
}

// one maps "no rows" to domain.ErrNotFound and wraps anything else with what failed.
func one(err error, what string) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, pgx.ErrNoRows):
		return domain.ErrNotFound
	default:
		return fmt.Errorf("%s: %w", what, err)
	}
}

// CreateTenant implements store.Tenants; a key hash that exists already is domain.ErrConflict.
func (s *Store) CreateTenant(ctx context.Context, t domain.Tenant) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO tenants (id, name, api_key_hash, rate_per_sec, burst, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		t.ID, t.Name, t.APIKeyHash[:], t.RatePerSec, t.Burst, t.CreatedAt)
	if sqlState(err) == uniqueViolation {
		return domain.ErrConflict
	}
	return one(err, "insert tenant")
}

// TenantByKeyHash implements store.Tenants. The lookup uses the unique index on api_key_hash.
func (s *Store) TenantByKeyHash(ctx context.Context, h domain.APIKeyHash) (domain.Tenant, error) {
	var t domain.Tenant
	var hash []byte
	err := s.pool.QueryRow(ctx, `
		SELECT id, name, api_key_hash, rate_per_sec, burst, created_at
		FROM tenants WHERE api_key_hash = $1`, h[:]).
		Scan(&t.ID, &t.Name, &hash, &t.RatePerSec, &t.Burst, &t.CreatedAt)
	if err != nil {
		return domain.Tenant{}, one(err, "select tenant")
	}
	copy(t.APIKeyHash[:], hash)
	t.CreatedAt = t.CreatedAt.UTC()
	return t, nil
}

const endpointColumns = `id, tenant_id, url, secret, enabled, created_at`

func scanEndpoint(row pgx.Row) (domain.Endpoint, error) {
	var e domain.Endpoint
	err := row.Scan(&e.ID, &e.TenantID, &e.URL, &e.Secret, &e.Enabled, &e.CreatedAt)
	e.CreatedAt = e.CreatedAt.UTC()
	return e, err
}

// CreateEndpoint implements store.Endpoints.
func (s *Store) CreateEndpoint(ctx context.Context, e domain.Endpoint) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO endpoints (`+endpointColumns+`) VALUES ($1, $2, $3, $4, $5, $6)`,
		e.ID, e.TenantID, e.URL, e.Secret, e.Enabled, e.CreatedAt)
	return one(err, "insert endpoint")
}

// GetEndpoint implements store.Endpoints. tenant_id is part of every WHERE: that is the BOLA guard.
func (s *Store) GetEndpoint(ctx context.Context, tenantID, id string) (domain.Endpoint, error) {
	e, err := scanEndpoint(s.pool.QueryRow(ctx, `
		SELECT `+endpointColumns+` FROM endpoints WHERE tenant_id = $1 AND id = $2`, tenantID, id))
	return e, one(err, "select endpoint")
}

// ListEndpoints implements store.Endpoints with keyset pagination: the row comparison
// (created_at, id) < ($2, $3) matches the order of endpoints_tenant_page_idx, so PostgreSQL reads only
// the rows of the page, however deep the page is. It asks for limit+1 rows to learn whether more exist.
func (s *Store) ListEndpoints(ctx context.Context, tenantID string, after *page.Cursor, limit int) ([]domain.Endpoint, bool, error) {
	at, id := cursorArgs(after)
	rows, err := s.pool.Query(ctx, `
		SELECT `+endpointColumns+` FROM endpoints
		WHERE tenant_id = $1 AND ($2::timestamptz IS NULL OR (created_at, id) < ($2, $3::uuid))
		ORDER BY created_at DESC, id DESC
		LIMIT $4`, tenantID, at, id, limit+1)
	if err != nil {
		return nil, false, fmt.Errorf("list endpoints: %w", err)
	}
	items, err := pgx.CollectRows(rows, func(r pgx.CollectableRow) (domain.Endpoint, error) { return scanEndpoint(r) })
	if err != nil {
		return nil, false, fmt.Errorf("list endpoints: %w", err)
	}
	return trim(items, limit)
}

// DisableEndpoint implements store.Endpoints.
func (s *Store) DisableEndpoint(ctx context.Context, tenantID, id string) (domain.Endpoint, error) {
	e, err := scanEndpoint(s.pool.QueryRow(ctx, `
		UPDATE endpoints SET enabled = false WHERE tenant_id = $1 AND id = $2
		RETURNING `+endpointColumns, tenantID, id))
	return e, one(err, "disable endpoint")
}

const subscriptionColumns = `id, tenant_id, endpoint_id, event_type, created_at`

func scanSubscription(row pgx.Row) (domain.Subscription, error) {
	var sub domain.Subscription
	err := row.Scan(&sub.ID, &sub.TenantID, &sub.EndpointID, &sub.EventType, &sub.CreatedAt)
	sub.CreatedAt = sub.CreatedAt.UTC()
	return sub, err
}

// CreateSubscription implements store.Subscriptions. The composite foreign key
// (tenant_id, endpoint_id) -> endpoints (tenant_id, id) rejects another tenant's endpoint, and the unique
// constraint rejects a second subscription for the same type: no read-then-write race is possible.
func (s *Store) CreateSubscription(ctx context.Context, sub domain.Subscription) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO subscriptions (`+subscriptionColumns+`) VALUES ($1, $2, $3, $4, $5)`,
		sub.ID, sub.TenantID, sub.EndpointID, sub.EventType, sub.CreatedAt)
	switch sqlState(err) {
	case foreignKeyViolation:
		return domain.ErrNotFound
	case uniqueViolation:
		return domain.ErrConflict
	}
	return one(err, "insert subscription")
}

// GetSubscription implements store.Subscriptions.
func (s *Store) GetSubscription(ctx context.Context, tenantID, id string) (domain.Subscription, error) {
	sub, err := scanSubscription(s.pool.QueryRow(ctx, `
		SELECT `+subscriptionColumns+` FROM subscriptions WHERE tenant_id = $1 AND id = $2`, tenantID, id))
	return sub, one(err, "select subscription")
}

// ListSubscriptions implements store.Subscriptions; an empty endpointID means every endpoint.
func (s *Store) ListSubscriptions(ctx context.Context, tenantID, endpointID string, after *page.Cursor, limit int) ([]domain.Subscription, bool, error) {
	at, id := cursorArgs(after)
	var endpoint *string
	if endpointID != "" {
		endpoint = &endpointID
	}
	rows, err := s.pool.Query(ctx, `
		SELECT `+subscriptionColumns+` FROM subscriptions
		WHERE tenant_id = $1
		  AND ($2::uuid IS NULL OR endpoint_id = $2)
		  AND ($3::timestamptz IS NULL OR (created_at, id) < ($3, $4::uuid))
		ORDER BY created_at DESC, id DESC
		LIMIT $5`, tenantID, endpoint, at, id, limit+1)
	if err != nil {
		return nil, false, fmt.Errorf("list subscriptions: %w", err)
	}
	items, err := pgx.CollectRows(rows, func(r pgx.CollectableRow) (domain.Subscription, error) { return scanSubscription(r) })
	if err != nil {
		return nil, false, fmt.Errorf("list subscriptions: %w", err)
	}
	return trim(items, limit)
}

// DeleteSubscription implements store.Subscriptions.
func (s *Store) DeleteSubscription(ctx context.Context, tenantID, id string) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM subscriptions WHERE tenant_id = $1 AND id = $2`, tenantID, id)
	if err != nil {
		return fmt.Errorf("delete subscription: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
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
	var stored domain.Event
	var replayed bool
	err := pgx.BeginTxFunc(ctx, s.pool, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, `
			INSERT INTO events (id, tenant_id, type, payload, idempotency_key, body_hash, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (tenant_id, idempotency_key) DO NOTHING
			RETURNING `+eventColumns,
			e.ID, e.TenantID, e.Type, []byte(e.Payload), e.IdempotencyKey, e.BodyHash[:], e.CreatedAt)
		got, err := scanEvent(row)
		if err == nil {
			stored = got
			_, err = tx.Exec(ctx, `INSERT INTO outbox (event_id, created_at) VALUES ($1, $2)`, got.ID, got.CreatedAt)
			return err
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		stored, err = scanEvent(tx.QueryRow(ctx, `
			SELECT `+eventColumns+` FROM events WHERE tenant_id = $1 AND idempotency_key = $2`,
			e.TenantID, e.IdempotencyKey))
		if err != nil {
			return err
		}
		if stored.BodyHash != e.BodyHash {
			return domain.ErrIdempotencyMismatch
		}
		replayed = true
		return nil
	})
	if errors.Is(err, domain.ErrIdempotencyMismatch) {
		return domain.Event{}, false, err
	}
	if err != nil {
		return domain.Event{}, false, fmt.Errorf("ingest event: %w", err)
	}
	return stored, replayed, nil
}

const eventColumns = `id, tenant_id, type, payload, idempotency_key, body_hash, created_at`

func scanEvent(row pgx.Row) (domain.Event, error) {
	var e domain.Event
	var payload, hash []byte
	err := row.Scan(&e.ID, &e.TenantID, &e.Type, &payload, &e.IdempotencyKey, &hash, &e.CreatedAt)
	e.Payload = payload
	copy(e.BodyHash[:], hash)
	e.CreatedAt = e.CreatedAt.UTC()
	return e, err
}

// GetEvent implements store.Events.
func (s *Store) GetEvent(ctx context.Context, tenantID, id string) (domain.Event, error) {
	e, err := scanEvent(s.pool.QueryRow(ctx, `
		SELECT `+eventColumns+` FROM events WHERE tenant_id = $1 AND id = $2`, tenantID, id))
	return e, one(err, "select event")
}

// cursorArgs turns an optional cursor into the two SQL parameters; NULL means "from the start".
func cursorArgs(after *page.Cursor) (*time.Time, *string) {
	if after == nil {
		return nil, nil
	}
	return &after.CreatedAt, &after.ID
}

// trim drops the extra row that was fetched only to answer "are there more?".
func trim[T any](items []T, limit int) ([]T, bool, error) {
	if len(items) > limit {
		return items[:limit], true, nil
	}
	return items, false, nil
}
