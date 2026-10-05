// Package testdb gives each test its own throw-away database on the server named by
// DATABASE_URL, at schema version 0. Tests skip when DATABASE_URL is unset.
package testdb

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var seq atomic.Int64

// New creates an empty database next to the one in DATABASE_URL, returns its URL, and drops
// it when the test ends (closing any connection still open to it).
func New(t testing.TB) string {
	t.Helper()
	admin := os.Getenv("DATABASE_URL")
	if admin == "" {
		t.Skip("integration test: set DATABASE_URL to the chapter 30 database; see README")
	}
	u, err := url.Parse(admin)
	if err != nil {
		t.Fatalf("parse DATABASE_URL: %v", err)
	}
	name := fmt.Sprintf("ch30_test_%d_%d", os.Getpid(), seq.Add(1))
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, admin)
	if err != nil {
		t.Fatalf("connect to DATABASE_URL: %v", err)
	}
	defer conn.Close(ctx)
	if _, err := conn.Exec(ctx, "CREATE DATABASE "+name); err != nil {
		t.Fatalf("create database: %v", err)
	}
	t.Cleanup(func() {
		c, err := pgx.Connect(context.Background(), admin)
		if err != nil {
			t.Errorf("drop %s: %v", name, err)
			return
		}
		defer c.Close(context.Background())
		if _, err := c.Exec(context.Background(), "DROP DATABASE "+name+" WITH (FORCE)"); err != nil {
			t.Errorf("drop %s: %v", name, err)
		}
	})
	u.Path = "/" + name
	return u.String()
}

// WaitForLockWaiters blocks until at least n sessions in this database are waiting for a
// lock. Tests use it to know, without sleeping, that a statement is really blocked.
func WaitForLockWaiters(t testing.TB, pool *pgxpool.Pool, n int) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		var waiting int
		err := pool.QueryRow(context.Background(), `
			SELECT count(*) FROM pg_stat_activity
			WHERE datname = current_database() AND wait_event_type = 'Lock'`).Scan(&waiting)
		if err != nil {
			t.Fatalf("pg_stat_activity: %v", err)
		}
		if waiting >= n {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("fewer than %d sessions started waiting for a lock within 10s", n)
}
