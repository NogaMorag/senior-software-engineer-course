// Package testdb gives each test package its own throw-away database on the server
// named by DATABASE_URL, migrated to the latest version. Tests skip when it is unset.
package testdb

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"example.com/ch11-isolation-lab/db"
)

var shared *pgxpool.Pool

// Main is called from TestMain: os.Exit(testdb.Main(m)). With DATABASE_URL set it
// creates a database, migrates it, runs the tests and drops the database.
func Main(m *testing.M) int {
	admin := os.Getenv("DATABASE_URL")
	if admin == "" {
		return m.Run() // every test calls Pool and skips
	}
	ctx := context.Background()
	dbURL, drop, err := CreateDatabase(ctx, admin)
	if err != nil {
		fmt.Fprintln(os.Stderr, "testdb:", err)
		return 1
	}
	defer func() {
		if err := drop(); err != nil {
			fmt.Fprintln(os.Stderr, "testdb: drop:", err)
		}
	}()
	cfg, err := db.Config(dbURL)
	if err != nil {
		fmt.Fprintln(os.Stderr, "testdb:", err)
		return 1
	}
	pool, err := db.Connect(ctx, cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "testdb:", err)
		return 1
	}
	defer pool.Close()
	if err := db.Migrate(ctx, pool); err != nil {
		fmt.Fprintln(os.Stderr, "testdb:", err)
		return 1
	}
	shared = pool
	return m.Run()
}

// Pool returns the package's migrated pool, or skips the test without DATABASE_URL.
func Pool(t testing.TB) *pgxpool.Pool {
	t.Helper()
	if shared == nil {
		t.Skip("integration test: set DATABASE_URL to the chapter 11 database; see README")
	}
	return shared
}

// CreateDatabase creates an empty database next to the one in adminURL and returns its
// URL and a function that drops it (closing any connection still open to it).
func CreateDatabase(ctx context.Context, adminURL string) (string, func() error, error) {
	u, err := url.Parse(adminURL)
	if err != nil {
		return "", nil, fmt.Errorf("parse DATABASE_URL: %w", err)
	}
	name := fmt.Sprintf("ch11_test_%d_%d", os.Getpid(), time.Now().UnixNano())
	conn, err := pgx.Connect(ctx, adminURL)
	if err != nil {
		return "", nil, err
	}
	defer conn.Close(ctx)
	if _, err := conn.Exec(ctx, "CREATE DATABASE "+name); err != nil {
		return "", nil, err
	}
	u.Path = "/" + name
	drop := func() error {
		c, err := pgx.Connect(context.Background(), adminURL)
		if err != nil {
			return err
		}
		defer c.Close(context.Background())
		_, err = c.Exec(context.Background(), "DROP DATABASE "+name+" WITH (FORCE)")
		return err
	}
	return u.String(), drop, nil
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
	t.Fatalf("no session started waiting for a lock within 10s")
}
