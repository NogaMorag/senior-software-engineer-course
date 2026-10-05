package db_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"example.com/ch11-isolation-lab/db"
	"example.com/ch11-isolation-lab/internal/testdb"
)

func TestMain(m *testing.M) { os.Exit(testdb.Main(m)) }

var serializationFailure = &pgconn.PgError{Code: db.SerializationFailure, Message: "could not serialize access (simulated)"}

func TestRetrySerializable(t *testing.T) {
	pool := testdb.Pool(t)
	boom := errors.New("boom")
	tests := []struct {
		name         string
		failures     int   // fn fails with 40001 this many times, then returns 42
		other        error // or fails once with this error
		wantAttempts int
		wantErr      bool
	}{
		{"success on the first attempt", 0, nil, 1, false},
		{"two serialization failures, then success", 2, nil, 3, false},
		{"a deadlock is retried too", -1, nil, 2, false},
		{"other errors are not retried", 0, boom, 1, true},
		{"gives up after MaxAttempts", 100, nil, db.MaxAttempts, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			attempts := 0
			got, err := db.RetrySerializable(context.Background(), pool, func(ctx context.Context, tx pgx.Tx) (int, error) {
				attempts++
				switch {
				case tc.other != nil:
					return 0, tc.other
				case tc.failures == -1 && attempts == 1:
					return 0, &pgconn.PgError{Code: db.DeadlockDetected}
				case attempts <= tc.failures:
					return 0, serializationFailure
				}
				return 42, nil
			})
			if attempts != tc.wantAttempts {
				t.Fatalf("fn ran %d times, want %d", attempts, tc.wantAttempts)
			}
			if tc.wantErr {
				if err == nil {
					t.Fatal("want an error, got nil")
				}
				if tc.other != nil && !errors.Is(err, tc.other) {
					t.Fatalf("err = %v, want %v", err, tc.other)
				}
				if tc.other == nil && !db.IsRetryable(err) {
					t.Fatalf("err = %v should still carry SQLSTATE 40001", err)
				}
				return
			}
			if err != nil || got != 42 {
				t.Fatalf("got %d, %v; want 42, nil", got, err)
			}
		})
	}
}

func TestRetrySerializableRunsAtSerializable(t *testing.T) {
	pool := testdb.Pool(t)
	level, err := db.RetrySerializable(context.Background(), pool, func(ctx context.Context, tx pgx.Tx) (string, error) {
		var level string
		err := tx.QueryRow(ctx, `SHOW transaction_isolation`).Scan(&level)
		return level, err
	})
	if err != nil || level != "serializable" {
		t.Fatalf("isolation = %q, %v; want serializable", level, err)
	}
}

func TestRetrySerializableStopsWhenContextEnds(t *testing.T) {
	pool := testdb.Pool(t)
	ctx, cancel := context.WithCancel(context.Background())
	attempts := 0
	_, err := db.RetrySerializable(ctx, pool, func(context.Context, pgx.Tx) (int, error) {
		attempts++
		cancel() // the caller gives up while we are failing
		return 0, serializationFailure
	})
	if !errors.Is(err, context.Canceled) || attempts != 1 {
		t.Fatalf("err = %v after %d attempts, want context.Canceled after 1", err, attempts)
	}
}

// The pool never opens more than MaxConns connections; the next caller waits, and gives
// up when its context does. This is what "pool exhausted" looks like from Go.
func TestPoolWaitsAtMaxConns(t *testing.T) {
	pool := testdb.Pool(t)
	cfg, err := db.Config(pool.Config().ConnString())
	if err != nil {
		t.Fatal(err)
	}
	cfg.MaxConns = 2
	small, err := db.Connect(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer small.Close()

	ctx := context.Background()
	c1, err := small.Acquire(ctx)
	if err != nil {
		t.Fatal(err)
	}
	c2, err := small.Acquire(ctx)
	if err != nil {
		t.Fatal(err)
	}
	short, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()
	if _, err := small.Acquire(short); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("third Acquire: err = %v, want context.DeadlineExceeded", err)
	}
	c1.Release()
	c3, err := small.Acquire(ctx)
	if err != nil {
		t.Fatalf("Acquire after a release: %v", err)
	}
	c3.Release()
	c2.Release()
	if s := small.Stat(); s.MaxConns() != 2 || s.TotalConns() > 2 {
		t.Fatalf("stat: max %d total %d", s.MaxConns(), s.TotalConns())
	}
}
