package db

import (
	"context"
	"errors"
	"math/rand/v2"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// SQLSTATE codes that mean "the transaction lost a race; running it again may succeed".
const (
	SerializationFailure = "40001"
	DeadlockDetected     = "40P01"
)

// MaxAttempts bounds RetrySerializable: the first run plus four retries.
const MaxAttempts = 5

const (
	baseDelay = 5 * time.Millisecond
	maxDelay  = 200 * time.Millisecond
)

// TxBeginner is what RetrySerializable needs: *pgxpool.Pool and *pgx.Conn both have it.
type TxBeginner interface {
	BeginTx(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, error)
}

// SQLState returns the PostgreSQL error code inside err, or "" if there is none.
func SQLState(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	return ""
}

// IsRetryable reports whether err is a serialization failure or a deadlock.
func IsRetryable(err error) bool {
	code := SQLState(err)
	return code == SerializationFailure || code == DeadlockDetected
}

// RetrySerializable runs fn inside a SERIALIZABLE transaction and commits it. When
// PostgreSQL aborts the transaction with 40001 or 40P01, the WHOLE transaction runs
// again, including fn's reads and decisions, after a short randomized wait. Any other
// error is returned at once. fn may therefore run several times: it must not have
// effects outside tx (no HTTP calls, no messages, no writes to shared Go state that a
// retry would duplicate).
func RetrySerializable[T any](ctx context.Context, db TxBeginner, fn func(ctx context.Context, tx pgx.Tx) (T, error)) (T, error) {
	// TODO(day 28): loop: pgx.BeginTxFunc with IsoLevel pgx.Serializable; return on success or on a non-retryable error; stop after MaxAttempts; sleep(ctx, backoff(attempt)) between attempts. See chapter 11, section 11.8.
	_, _ = backoff, sleep // you will call these between attempts
	panic("TODO: RetrySerializable, see chapter 11, section 11.8")
}

// backoff is exponential with full jitter: a random wait between 0 and
// min(maxDelay, baseDelay * 2^(attempt-1)), so competing transactions spread out.
func backoff(attempt int) time.Duration {
	ceiling := min(maxDelay, baseDelay<<(attempt-1))
	return rand.N(ceiling + 1)
}

// sleep waits for d or until ctx is done, whichever comes first.
func sleep(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
