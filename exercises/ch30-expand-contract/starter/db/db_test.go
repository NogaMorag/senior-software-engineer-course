package db_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"example.com/ch30-expand-contract/db"
	"example.com/ch30-expand-contract/internal/testdb"
)

func open(t *testing.T, url string, cfg func(string) (*pgxpool.Config, error)) *pgxpool.Pool {
	t.Helper()
	c, err := cfg(url)
	if err != nil {
		t.Fatal(err)
	}
	pool, err := db.Connect(context.Background(), c)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func TestMigrationConnectionsHaveALockTimeout(t *testing.T) {
	url := testdb.New(t)
	ctx := context.Background()
	var appTimeout, migTimeout string
	if err := open(t, url, db.AppConfig).QueryRow(ctx, "SHOW lock_timeout").Scan(&appTimeout); err != nil {
		t.Fatal(err)
	}
	if err := open(t, url, db.MigrationConfig).QueryRow(ctx, "SHOW lock_timeout").Scan(&migTimeout); err != nil {
		t.Fatal(err)
	}
	if appTimeout != "0" || migTimeout != "2s" {
		t.Fatalf("lock_timeout: app %q, migrations %q; want \"0\" and \"2s\"", appTimeout, migTimeout)
	}
}

// TestWaitingDDLBlocksReaders shows the lock queue: an ALTER TABLE that waits behind a long
// transaction makes every later query on the table wait behind the ALTER, even a plain SELECT.
// With lock_timeout the ALTER gives up, and the readers behind it go on.
func TestWaitingDDLBlocksReaders(t *testing.T) {
	url := testdb.New(t)
	ctx := context.Background()
	appPool := open(t, url, db.AppConfig)
	migPool := open(t, url, db.MigrationConfig)
	if err := db.MigrateTo(ctx, migPool, 1); err != nil {
		t.Fatal(err)
	}

	// A long transaction that read the table holds ACCESS SHARE until it ends.
	long, err := appPool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = long.Rollback(ctx) }()
	if _, err := long.Exec(ctx, "SELECT count(*) FROM users"); err != nil {
		t.Fatal(err)
	}

	ddl := make(chan error, 1)
	start := time.Now()
	go func() {
		_, err := migPool.Exec(ctx, "ALTER TABLE users ADD COLUMN nickname text")
		ddl <- err
	}()
	testdb.WaitForLockWaiters(t, appPool, 1) // the ALTER waits for the long transaction

	read := make(chan error, 1)
	go func() {
		_, err := appPool.Exec(ctx, "SELECT count(*) FROM users")
		read <- err
	}()
	testdb.WaitForLockWaiters(t, appPool, 2) // and now a plain SELECT waits behind the ALTER

	err = <-ddl
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "55P03" {
		t.Fatalf("ALTER returned %v; want lock_timeout (SQLSTATE 55P03)", err)
	}
	if waited := time.Since(start); waited < db.LockTimeout {
		t.Fatalf("ALTER gave up after %v, before lock_timeout", waited)
	}
	select {
	case err := <-read:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("the SELECT is still blocked after the ALTER gave up")
	}
	t.Logf("ALTER gave up after %v: %s", time.Since(start).Round(100*time.Millisecond), pgErr.Message)
}
