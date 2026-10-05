// Package catalog is the chapter 12 exercise: a products lookup in PostgreSQL
// (the products table from chapter 10) and two caches in front of it.
package catalog

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Schema is the DDL and seed in schema.sql.
//
//go:embed schema.sql
var Schema string

// ErrNotFound means there is no product with that id.
var ErrNotFound = errors.New("product not found")

// Product is one row of products. The JSON tags are the cached form.
type Product struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	PriceCents int64     `json:"price_cents"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Config parses a connection URL such as
// postgres://course:change-me@localhost:5412/catalog and sets explicit pool limits,
// as in chapter 10.
func Config(url string) (*pgxpool.Config, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}
	cfg.MaxConns = 10
	cfg.ConnConfig.ConnectTimeout = 5 * time.Second
	return cfg, nil
}

// Connect opens the pool and checks that the database answers.
func Connect(ctx context.Context, cfg *pgxpool.Config) (*pgxpool.Pool, error) {
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("open pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return pool, nil
}

// Store reads and writes products. It counts its reads so tests and the demo can
// see how many requests reached the database.
type Store struct {
	pool  *pgxpool.Pool
	calls atomic.Int64
	// Latency, when set before use, makes every read also run pg_sleep for that long:
	// a stand-in for an expensive query, so that a stampede is visible.
	Latency time.Duration
}

// NewStore returns a store over pool.
func NewStore(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

// GetProduct reads one product: the source of truth behind both caches.
func (s *Store) GetProduct(ctx context.Context, id int64) (Product, error) {
	s.calls.Add(1)
	var p Product
	err := s.pool.QueryRow(ctx, `
		SELECT id, name, price_cents, updated_at
		FROM products, pg_sleep($2)
		WHERE id = $1`, id, s.Latency.Seconds()).Scan(&p.ID, &p.Name, &p.PriceCents, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Product{}, ErrNotFound
	}
	if err != nil {
		return Product{}, fmt.Errorf("get product %d: %w", id, err)
	}
	return p, nil
}

// UpdatePrice changes a price in a transaction. When it returns nil, the change is
// committed: only then may a cache entry for the product be deleted.
func (s *Store) UpdatePrice(ctx context.Context, id, priceCents int64) error {
	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, `
			UPDATE products SET price_cents = $2, updated_at = now()
			WHERE id = $1`, id, priceCents)
		if err != nil {
			return fmt.Errorf("update price of %d: %w", id, err)
		}
		if tag.RowsAffected() == 0 {
			return ErrNotFound
		}
		return nil
	})
}

// Calls returns the number of GetProduct calls since the last ResetCalls.
func (s *Store) Calls() int64 { return s.calls.Load() }

// ResetCalls sets the read count back to zero.
func (s *Store) ResetCalls() { s.calls.Store(0) }
