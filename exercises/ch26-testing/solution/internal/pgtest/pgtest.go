// Package pgtest gives integration tests a real PostgreSQL without any setup outside Go:
// one container per test binary (testcontainers-go), and one empty database per test inside it.
//
// Use it from TestMain:
//
//	func TestMain(m *testing.M) { os.Exit(pgtest.Main(m)) }
//
// and in each test:
//
//	url := pgtest.NewDatabase(t)
//
// `go test -short` skips every test that asks for a database, without starting Docker.
package pgtest

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"os"
	"sync/atomic"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

// Image is the server the tests run against: the same tag as the course's compose files.
const Image = "postgres:18.6"

var (
	adminURL string        // connection URL of the container's maintenance database
	admin    *pgxpool.Pool // creates and drops the per-test databases
	nextID   atomic.Int64  // numbers the per-test databases
)

// Main starts the container, runs the tests, and removes the container, in that order.
// It returns the exit code for os.Exit. If the binary dies before the deferred removal runs
// (a panic in another goroutine, a kill), Ryuk, the reaper container that testcontainers starts
// next to yours, removes the container once this process's connection to it is gone.
func Main(m *testing.M) int {
	flag.Parse() // testing.Short reads the -short flag
	if testing.Short() {
		return m.Run()
	}
	ctx := context.Background()

	ctr, err := postgres.Run(ctx, Image,
		postgres.WithDatabase("pgtest"),
		postgres.WithUsername("course"),
		postgres.WithPassword("change-me"), // a throw-away test server, not a secret
		postgres.BasicWaitStrategies(),
	)
	// Terminate even when Run failed: a container that started but never became ready
	// still exists. TerminateContainer accepts a nil container.
	defer func() {
		if err := testcontainers.TerminateContainer(ctr); err != nil {
			fmt.Fprintln(os.Stderr, "pgtest: terminate container:", err)
		}
	}()
	if err != nil {
		fmt.Fprintln(os.Stderr, "pgtest: start PostgreSQL (is Docker running? use -short to skip):", err)
		return 1
	}

	adminURL, err = ctr.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		fmt.Fprintln(os.Stderr, "pgtest: connection string:", err)
		return 1
	}
	admin, err = pgxpool.New(ctx, adminURL)
	if err != nil {
		fmt.Fprintln(os.Stderr, "pgtest: connect:", err)
		return 1
	}
	defer admin.Close()
	return m.Run()
}

// NewDatabase creates an empty database for this test and returns its connection URL.
// The database is dropped when the test ends, so tests that call it may run in parallel.
func NewDatabase(t testing.TB) string {
	t.Helper()
	if admin == nil {
		if testing.Short() {
			t.Skip("pgtest: integration test skipped under -short")
		}
		t.Fatal("pgtest: no server; call os.Exit(pgtest.Main(m)) from TestMain")
	}
	ctx := context.Background()
	name := fmt.Sprintf("test_%d", nextID.Add(1))
	ident := pgx.Identifier{name}.Sanitize()
	if _, err := admin.Exec(ctx, "CREATE DATABASE "+ident); err != nil {
		t.Fatalf("pgtest: create database: %v", err)
	}
	t.Cleanup(func() {
		// FORCE closes connections a test forgot to close, instead of failing the drop.
		if _, err := admin.Exec(context.Background(), "DROP DATABASE "+ident+" WITH (FORCE)"); err != nil {
			t.Errorf("pgtest: drop database: %v", err)
		}
	})

	u, err := url.Parse(adminURL)
	if err != nil {
		t.Fatalf("pgtest: parse url: %v", err)
	}
	u.Path = "/" + name
	return u.String()
}
