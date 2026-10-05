// Package orders is the chapter 10 repository: customers' orders and their items,
// read from PostgreSQL with pgx. Copied unchanged from exercises/ch10-orders-db/solution;
// chapter 26 adds integration tests that start their own PostgreSQL.
package orders

import (
	"context"
	_ "embed"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Schema is the DDL in schema.sql. The tests apply it to a private schema;
// in psql you run the file itself.
//
//go:embed schema.sql
var Schema string

// Config parses a connection URL such as
// postgres://course:change-me@localhost:5410/orders and sets explicit pool limits.
// Callers may adjust the returned config (tests add a tracer) before Connect.
func Config(url string) (*pgxpool.Config, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}
	cfg.MaxConns = 10
	cfg.MaxConnIdleTime = 5 * time.Minute
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

// QueryCounter is a pgx tracer that counts the statements sent to the server.
// Install it in Config().ConnConfig.Tracer to see an N+1 pattern as a number.
type QueryCounter struct {
	n atomic.Int64
}

// TraceQueryStart implements pgx.QueryTracer.
func (c *QueryCounter) TraceQueryStart(ctx context.Context, _ *pgx.Conn, _ pgx.TraceQueryStartData) context.Context {
	c.n.Add(1)
	return ctx
}

// TraceQueryEnd implements pgx.QueryTracer.
func (c *QueryCounter) TraceQueryEnd(context.Context, *pgx.Conn, pgx.TraceQueryEndData) {}

// Count returns the number of statements since the last Reset.
func (c *QueryCounter) Count() int64 { return c.n.Load() }

// Reset sets the count back to zero.
func (c *QueryCounter) Reset() { c.n.Store(0) }
