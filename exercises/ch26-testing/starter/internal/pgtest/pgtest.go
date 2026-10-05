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
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Image is the server the tests run against: the same tag as the course's compose files.
const Image = "postgres:18.6"

var admin *pgxpool.Pool // creates and drops the per-test databases; set by Main

// Main starts the container, runs the tests, and removes the container, in that order.
// It returns the exit code for os.Exit.
//
// TODO (day 80, step 2):
//   - flag.Parse(), then return m.Run() at once under testing.Short().
//   - postgres.Run(ctx, Image, WithDatabase, WithUsername, WithPassword, BasicWaitStrategies()).
//   - defer testcontainers.TerminateContainer(ctr) BEFORE checking the error of Run.
//   - ConnectionString(ctx, "sslmode=disable"), open admin with pgxpool.New, defer admin.Close().
//   - return m.Run().
func Main(m *testing.M) int {
	panic("TODO: pgtest.Main; see the README, day 80, step 2")
}

// NewDatabase creates an empty database for this test and returns its connection URL.
// The database is dropped when the test ends, so tests that call it may run in parallel.
//
// TODO (day 80, step 3): skip under -short when admin is nil; CREATE DATABASE with a unique
// name (pgx.Identifier{name}.Sanitize()); t.Cleanup with DROP DATABASE ... WITH (FORCE);
// return the admin URL with its path replaced by "/" + name.
func NewDatabase(t testing.TB) string {
	t.Helper()
	_ = admin
	panic("TODO: pgtest.NewDatabase; see the README, day 80, step 3")
}
