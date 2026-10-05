package delivery

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
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
	var job *Job
	err := pgx.BeginFunc(ctx, db, func(tx pgx.Tx) error {
		for {
			var id, status string
			var attempts, base int
			err := tx.QueryRow(ctx, `
				SELECT id, status, attempts, attempt_base FROM deliveries
				WHERE (status = 'pending' AND next_attempt_at <= now())
				   OR (status = 'in_flight' AND lease_until < now())
				ORDER BY next_attempt_at
				LIMIT 1
				FOR UPDATE SKIP LOCKED`).Scan(&id, &status, &attempts, &base)
			if errors.Is(err, pgx.ErrNoRows) {
				return nil
			}
			if err != nil {
				return fmt.Errorf("claim: %w", err)
			}
			if status == "in_flight" {
				// The previous holder's lease ran out: that attempt failed without a word. Its start is
				// estimated as lease_until minus the lease, the lease every worker of this version uses.
				if _, err := tx.Exec(ctx, `
					INSERT INTO attempts (delivery_id, n, started_at, duration_ms, error)
					SELECT id, attempts, lease_until - $2::interval, $3, 'lease expired' FROM deliveries WHERE id = $1`,
					id, lease, lease.Milliseconds()); err != nil {
					return fmt.Errorf("claim: record expired attempt: %w", err)
				}
				if attempts-base >= maxAttempts {
					if _, err := tx.Exec(ctx, `
						UPDATE deliveries SET status = 'dead', lease_until = NULL,
						       last_error = 'lease expired on the last attempt'
						WHERE id = $1`, id); err != nil {
						return fmt.Errorf("claim: bury delivery %s: %w", id, err)
					}
					continue // look for another delivery
				}
			}
			var j Job
			err = tx.QueryRow(ctx, `
				WITH claimed AS (
					UPDATE deliveries SET status = 'in_flight', attempts = attempts + 1,
					       lease_until = now() + $2::interval
					WHERE id = $1
					RETURNING id, tenant_id, event_id, endpoint_id, attempts, attempt_base, lease_until
				)
				SELECT c.id, c.tenant_id, c.event_id, c.endpoint_id, c.attempts, c.attempts - c.attempt_base,
				       c.lease_until, ep.url, ep.secret, ep.enabled, ev.type, ev.created_at, ev.payload
				FROM claimed c
				JOIN endpoints ep ON ep.id = c.endpoint_id
				JOIN events ev ON ev.id = c.event_id`, id, lease).Scan(
				&j.DeliveryID, &j.TenantID, &j.EventID, &j.EndpointID, &j.Attempt, &j.Round, &j.LeaseExpiry,
				&j.URL, &j.Secret, &j.Enabled, &j.EventType, &j.EventAt, &j.Payload)
			if err != nil {
				return fmt.Errorf("claim delivery %s: %w", id, err)
			}
			j.EventAt = j.EventAt.UTC()
			job = &j
			return nil
		}
	})
	if err != nil {
		return nil, err
	}
	return job, nil
}

// record moves a delivery we still hold to its next status and appends the attempt, in one transaction.
// The attempts value is the fencing token: if another worker has reclaimed the delivery since, attempts
// has moved on and nothing changes.
func record(ctx context.Context, db *pgxpool.Pool, j *Job, res Result, set string, args ...any) error {
	return pgx.BeginFunc(ctx, db, func(tx pgx.Tx) error {
		var held bool
		err := tx.QueryRow(ctx, `
			SELECT true FROM deliveries WHERE id = $1 AND status = 'in_flight' AND attempts = $2
			FOR UPDATE`, j.DeliveryID, j.Attempt).Scan(&held)
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("delivery %s attempt %d: %w", j.DeliveryID, j.Attempt, ErrLeaseLost)
		}
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `UPDATE deliveries SET `+set+`, lease_until = NULL WHERE id = $1`,
			append([]any{j.DeliveryID}, args...)...); err != nil {
			return err
		}
		var status, errText any // NULL unless known
		if res.StatusCode != 0 {
			status = res.StatusCode
		}
		if res.Err != nil {
			errText = res.Err.Error()
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO attempts (delivery_id, n, started_at, duration_ms, status_code, error)
			VALUES ($1, $2, $3, $4, $5, $6)`,
			j.DeliveryID, j.Attempt, res.Started, res.Duration.Milliseconds(), status, errText)
		return err
	})
}

// Complete marks the delivery delivered and records the attempt. ErrLeaseLost if it is no longer ours.
func Complete(ctx context.Context, db *pgxpool.Pool, j *Job, res Result) error {
	if err := record(ctx, db, j, res, `status = 'delivered', delivered_at = now(), last_error = NULL`); err != nil {
		return fmt.Errorf("complete: %w", err)
	}
	return nil
}

// Fail records a failed attempt. A retryable failure with attempts left in the round goes back to
// pending at now+retryIn; anything else goes dead and waits for a replay. It returns the new status.
func Fail(ctx context.Context, db *pgxpool.Pool, j *Job, res Result, maxAttempts int, retryIn time.Duration) (string, error) {
	cause := "failed"
	if res.Err != nil {
		cause = res.Err.Error()
	}
	next, set, args := "pending", `status = 'pending', next_attempt_at = now() + $2::interval, last_error = $3`, []any{retryIn, cause}
	if !res.Retryable || j.Round >= maxAttempts {
		next, set, args = "dead", `status = 'dead', last_error = $2`, []any{cause}
	}
	if err := record(ctx, db, j, res, set, args...); err != nil {
		return "", fmt.Errorf("fail: %w", err)
	}
	return next, nil
}

// Postpone gives a claimed delivery back without spending an attempt (milestone M6): the endpoint's
// breaker is open, or its bulkhead is full, so no request was sent. attempts goes back to its value
// before the claim, nothing is recorded, and the delivery is due again after `after`.
func Postpone(ctx context.Context, db *pgxpool.Pool, j *Job, after time.Duration) error {
	tag, err := db.Exec(ctx, `
		UPDATE deliveries SET status = 'pending', attempts = attempts - 1, lease_until = NULL,
		       next_attempt_at = now() + $3::interval
		WHERE id = $1 AND status = 'in_flight' AND attempts = $2`, j.DeliveryID, j.Attempt, after)
	if err != nil {
		return fmt.Errorf("postpone: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("postpone delivery %s attempt %d: %w", j.DeliveryID, j.Attempt, ErrLeaseLost)
	}
	return nil
}
