// Package fencing is the storage side of a fencing token. Every write carries the writer's
// token (a leader's term); the store refuses a write whose token is lower than one it has
// already accepted for that resource. A lock tells a process it may write; only the
// storage can tell that a write comes from a process that has since lost the lock.
package fencing

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Schema holds one row per resource: its value and the highest token that wrote it.
const Schema = `CREATE TABLE IF NOT EXISTS fenced_resources (
	name  text PRIMARY KEY,
	value text NOT NULL,
	token bigint NOT NULL
)`

// ErrStaleToken means a writer with a higher token has already written this resource.
var ErrStaleToken = errors.New("fencing: stale token")

// DB is what Store needs: a *pgxpool.Pool, a *pgx.Conn or a pgx.Tx.
type DB interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// Store is a table of fenced resources.
type Store struct{ DB DB }

// Write sets name to value if token is at least the highest token seen for name. The check
// and the write are one statement, so two writers cannot both pass the check.
func (s Store) Write(ctx context.Context, name, value string, token int64) error {
	tag, err := s.DB.Exec(ctx, `
		INSERT INTO fenced_resources (name, value, token) VALUES ($1, $2, $3)
		ON CONFLICT (name) DO UPDATE SET value = EXCLUDED.value, token = EXCLUDED.token
		WHERE fenced_resources.token <= EXCLUDED.token`, name, value, token)
	if err != nil {
		return fmt.Errorf("fenced write %q: %w", name, err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("write %q with token %d: %w", name, token, ErrStaleToken)
	}
	return nil
}

// Read returns the value of name and the token that wrote it.
func (s Store) Read(ctx context.Context, name string) (string, int64, error) {
	var value string
	var token int64
	err := s.DB.QueryRow(ctx, `SELECT value, token FROM fenced_resources WHERE name = $1`, name).Scan(&value, &token)
	return value, token, err
}
