// Package testenv gives each test a private PostgreSQL schema with the chapter 14 tables,
// and skips the test when DATABASE_URL is not set. It is test support, not exercise code.
package testenv

import (
	"context"
	"fmt"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"example.com/ch14-jobqueue/notify"
	"example.com/ch14-jobqueue/queue"
)

var counter atomic.Int64

// Postgres returns a pool whose search_path is a fresh schema with queue.Schema and
// notify.Schema applied. The schema is dropped when the test ends.
func Postgres(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("integration test: set DATABASE_URL to the chapter 14 database; see README")
	}
	ctx := context.Background()
	schema := fmt.Sprintf("ch14_test_%d_%d_%d", os.Getpid(), time.Now().UnixNano(), counter.Add(1))
	admin, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(admin.Close)
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := admin.Exec(context.Background(), "DROP SCHEMA "+schema+" CASCADE"); err != nil {
			t.Errorf("drop schema: %v", err)
		}
	})
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		t.Fatal(err)
	}
	cfg.ConnConfig.RuntimeParams["search_path"] = schema
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	for _, s := range []string{queue.Schema, notify.Schema} {
		if _, err := pool.Exec(ctx, s); err != nil {
			t.Fatalf("apply schema: %v", err)
		}
	}
	return pool
}
