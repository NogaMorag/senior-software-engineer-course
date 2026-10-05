// Package testdb gives a test package its own throw-away database on the server named by
// DATABASE_URL. Advisory locks are per database, so packages that run at the same time
// never see each other's locks.
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
)

var dbURL string

// Main is called from TestMain: os.Exit(testdb.Main(m)). With DATABASE_URL set it creates
// a database, runs the tests and drops the database; without it, every test skips.
func Main(m *testing.M) int {
	admin := os.Getenv("DATABASE_URL")
	if admin == "" {
		return m.Run()
	}
	ctx := context.Background()
	u, err := url.Parse(admin)
	if err != nil {
		fmt.Fprintln(os.Stderr, "testdb: parse DATABASE_URL:", err)
		return 1
	}
	name := fmt.Sprintf("ch18_test_%d_%d", os.Getpid(), time.Now().UnixNano())
	if err := exec(ctx, admin, "CREATE DATABASE "+name); err != nil {
		fmt.Fprintln(os.Stderr, "testdb:", err)
		return 1
	}
	defer func() {
		if err := exec(ctx, admin, "DROP DATABASE "+name+" WITH (FORCE)"); err != nil {
			fmt.Fprintln(os.Stderr, "testdb: drop:", err)
		}
	}()
	u.Path = "/" + name
	dbURL = u.String()
	return m.Run()
}

func exec(ctx context.Context, url, sql string) error {
	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)
	_, err = conn.Exec(ctx, sql)
	return err
}

// Pool opens a new pool on the package's database, closed when the test ends. Each call is
// a separate pool, as two processes would have; the test skips without DATABASE_URL.
func Pool(t testing.TB) *pgxpool.Pool {
	t.Helper()
	if dbURL == "" {
		t.Skip("integration test: set DATABASE_URL to the chapter 18 database; see README")
	}
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	return pool
}
