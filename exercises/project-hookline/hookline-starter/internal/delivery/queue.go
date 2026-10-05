package delivery

import (
	"context"
	"errors"

	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrLeaseLost means the delivery is no longer ours: the lease expired and another worker claimed it.
// The attempt's result must be discarded; the other worker records its own.
var ErrLeaseLost = errors.New("lease lost")

// Claim takes one due delivery and leases it until now+lease (chapter 14). Due means pending with
// next_attempt_at in the past, or in flight with an expired lease (its worker died). It returns nil, nil
// when there is nothing to do. SKIP LOCKED lets many workers claim at once without waiting on each
// other and without two of them taking the same row.
//
// The claim commits at once; the HTTP attempt runs with no transaction open. The lease, not a row lock,
// keeps other workers away while it runs, and attempts is the fencing token that finish checks.
func Claim(ctx context.Context, db *pgxpool.Pool, lease time.Duration, maxAttempts int) (*Job, error) {
	panic("TODO: see chapter 38")

}

// record moves a delivery we still hold to its next status and appends the attempt, in one transaction.
// The attempts value is the fencing token: if another worker has reclaimed the delivery since, attempts
// has moved on and nothing changes.
func record(ctx context.Context, db *pgxpool.Pool, j *Job, res Result, set string, args ...any) error {
	panic("TODO: see chapter 38")

}

// Complete marks the delivery delivered and records the attempt. ErrLeaseLost if it is no longer ours.
func Complete(ctx context.Context, db *pgxpool.Pool, j *Job, res Result) error {
	panic("TODO: see chapter 38")

}

// Fail records a failed attempt. A retryable failure with attempts left in the round goes back to
// pending at now+retryIn; anything else goes dead and waits for a replay. It returns the new status.
func Fail(ctx context.Context, db *pgxpool.Pool, j *Job, res Result, maxAttempts int, retryIn time.Duration) (string, error) {
	panic("TODO: see chapter 38")

}

// Postpone gives a claimed delivery back without spending an attempt (milestone M6): the endpoint's
// breaker is open, or its bulkhead is full, so no request was sent. attempts goes back to its value
// before the claim, nothing is recorded, and the delivery is due again after `after`.
func Postpone(ctx context.Context, db *pgxpool.Pool, j *Job, after time.Duration) error {
	panic("TODO: see chapter 38")

}
