// Package queue is a durable job queue on PostgreSQL: jobs with a run_at time, claimed
// with FOR UPDATE SKIP LOCKED under a lease, retried until max_attempts, then dead until
// an operator replays them. Every finished attempt is recorded in job_attempts.
package queue

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Schema creates the jobs and job_attempts tables.
//
//go:embed schema.sql
var Schema string

// State is a job's lifecycle state.
type State string

// The four states. pending: waiting for run_at. running: claimed under a lease.
// done: finished. dead: out of attempts (or failed permanently); waits for Replay.
const (
	Pending State = "pending"
	Running State = "running"
	Done    State = "done"
	Dead    State = "dead"
)

var (
	// ErrLeaseLost means the job is no longer ours: the lease expired and another worker
	// claimed it (or it was finished by someone else). The result must be discarded.
	ErrLeaseLost = errors.New("lease lost")
	// ErrNotDead means Replay was called on a job that is not dead.
	ErrNotDead = errors.New("job is not dead")
)

// Job is one row of the jobs table.
type Job struct {
	ID          int64
	Kind        string
	Payload     []byte
	State       State
	RunAt       time.Time
	Attempts    int // includes the current one while running
	MaxAttempts int
	LockedBy    string
	LastError   string
}

// Attempt is one row of job_attempts.
type Attempt struct {
	Attempt int
	Worker  string
	Error   string // "" means success
}

// Options tune Enqueue. The zero value means: run now, 5 attempts, no unique key.
type Options struct {
	RunAt       time.Time
	MaxAttempts int
	// UniqueKey, when set, makes Enqueue idempotent: a second Enqueue with the same key
	// returns the first job's id instead of adding a job.
	UniqueKey string
}

// Querier is satisfied by *pgxpool.Pool and by pgx.Tx, so a job can be enqueued inside
// the same transaction as the business write that caused it.
type Querier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// Enqueue adds a job. It reports whether a new row was created (false only when
// UniqueKey matched an existing job).
func Enqueue(ctx context.Context, q Querier, kind string, payload []byte, opt Options) (int64, bool, error) {
	if payload == nil {
		payload = []byte("{}")
	}
	maxAttempts := opt.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 5
	}
	var runAt, key any // NULL unless set
	if !opt.RunAt.IsZero() {
		runAt = opt.RunAt
	}
	if opt.UniqueKey != "" {
		key = opt.UniqueKey
	}
	var id int64
	err := q.QueryRow(ctx, `
		INSERT INTO jobs (kind, payload, run_at, max_attempts, unique_key)
		VALUES ($1, $2, COALESCE($3::timestamptz, now()), $4, $5)
		ON CONFLICT (unique_key) DO NOTHING
		RETURNING id`, kind, payload, runAt, maxAttempts, key).Scan(&id)
	if err == nil {
		return id, true, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return 0, false, fmt.Errorf("enqueue %s: %w", kind, err)
	}
	// The unique key already exists: return that job.
	if err := q.QueryRow(ctx, `SELECT id FROM jobs WHERE unique_key = $1`, key).Scan(&id); err != nil {
		return 0, false, fmt.Errorf("enqueue %s: find existing job: %w", kind, err)
	}
	return id, false, nil
}

const jobColumns = `id, kind, payload, state, run_at, attempts, max_attempts, COALESCE(locked_by, ''), COALESCE(last_error, '')`

func scanJob(row pgx.Row) (Job, error) {
	var j Job
	err := row.Scan(&j.ID, &j.Kind, &j.Payload, &j.State, &j.RunAt, &j.Attempts, &j.MaxAttempts, &j.LockedBy, &j.LastError)
	return j, err
}

// Claim takes one runnable job for worker and leases it for lease. Runnable means
// pending with run_at in the past, or running with an expired lease (its worker died).
// It returns nil, nil when there is nothing to do. SKIP LOCKED lets many workers claim at
// once without waiting on each other and without two of them taking the same job.
//
// The claim commits at once; the handler then runs with no transaction open. The lease,
// not a row lock, is what keeps other workers away while the job runs.
func Claim(ctx context.Context, db *pgxpool.Pool, worker string, lease time.Duration) (*Job, error) {
	var claimed *Job
	err := pgx.BeginFunc(ctx, db, func(tx pgx.Tx) error {
		for {
			var id int64
			var state State
			var attempts, maxAttempts int
			// TODO(day 41): select ONE runnable job and lock it, skipping rows other workers
			// have locked. Runnable: pending with run_at <= now(), or running with
			// locked_until < now(). Oldest run_at first, then id.
			err := tx.QueryRow(ctx, `
				SELECT id, state, attempts, max_attempts FROM jobs
				WHERE false`).Scan(&id, &state, &attempts, &maxAttempts)
			if errors.Is(err, pgx.ErrNoRows) {
				return nil
			}
			if err != nil {
				return fmt.Errorf("claim: %w", err)
			}
			if state == Running {
				// The previous holder's lease ran out: that attempt failed without a word.
				if _, err := tx.Exec(ctx, `
					INSERT INTO job_attempts (job_id, attempt, worker, started_at, error)
					SELECT id, attempts, locked_by, locked_at, 'lease expired' FROM jobs WHERE id = $1`, id); err != nil {
					return fmt.Errorf("claim: record expired attempt: %w", err)
				}
				if attempts >= maxAttempts {
					if _, err := tx.Exec(ctx, `
						UPDATE jobs SET state = 'dead', last_error = 'lease expired on the last attempt',
						       locked_by = NULL, locked_until = NULL, finished_at = now()
						WHERE id = $1`, id); err != nil {
						return fmt.Errorf("claim: bury job %d: %w", id, err)
					}
					continue // look for another job
				}
			}
			j, err := scanJob(tx.QueryRow(ctx, `
				UPDATE jobs SET state = 'running', attempts = attempts + 1, locked_by = $2,
				       locked_at = now(), locked_until = now() + $3::interval
				WHERE id = $1
				RETURNING `+jobColumns, id, worker, lease))
			if err != nil {
				return fmt.Errorf("claim job %d: %w", id, err)
			}
			claimed = &j
			return nil
		}
	})
	if err != nil {
		return nil, err
	}
	return claimed, nil
}

// finish moves a running job that we still hold to its next state and records the
// attempt, in one transaction. The attempts value is the fencing token: if another
// worker has reclaimed the job since, attempts has moved on and nothing is changed.
func finish(ctx context.Context, db *pgxpool.Pool, j *Job, set string, args []any, cause string) error {
	return pgx.BeginFunc(ctx, db, func(tx pgx.Tx) error {
		var lockedBy string
		var lockedAt time.Time
		err := tx.QueryRow(ctx, `
			SELECT locked_by, locked_at FROM jobs
			WHERE id = $1 AND state = 'running' AND attempts = $2
			FOR UPDATE`, j.ID, j.Attempts).Scan(&lockedBy, &lockedAt)
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("job %d attempt %d: %w", j.ID, j.Attempts, ErrLeaseLost)
		}
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `UPDATE jobs SET `+set+`, locked_by = NULL, locked_until = NULL WHERE id = $1`,
			append([]any{j.ID}, args...)...); err != nil {
			return err
		}
		var errText any
		if cause != "" {
			errText = cause
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO job_attempts (job_id, attempt, worker, started_at, error)
			VALUES ($1, $2, $3, $4, $5)`, j.ID, j.Attempts, lockedBy, lockedAt, errText)
		return err
	})
}

// Complete marks a job done. It returns ErrLeaseLost if the job is no longer ours.
func Complete(ctx context.Context, db *pgxpool.Pool, j *Job) error {
	if err := finish(ctx, db, j, `state = 'done', finished_at = now()`, nil, ""); err != nil {
		return fmt.Errorf("complete: %w", err)
	}
	return nil
}

// Fail records a failed attempt. If attempts remain and cause is not permanent, the job
// goes back to pending with run_at = now() + retryIn; otherwise it goes dead. It returns
// the job's new state.
func Fail(ctx context.Context, db *pgxpool.Pool, j *Job, cause error, retryIn time.Duration) (State, error) {
	// TODO(day 41): the retry branch. Call finish with
	//   set  = `state = 'pending', run_at = now() + $2::interval, last_error = $3`
	//   args = []any{retryIn, cause.Error()}
	// and return Pending. Wrap errors as "fail: %w".
	// TODO(day 42): the dead branch. When IsPermanent(cause) or j.Attempts >= j.MaxAttempts:
	//   set  = `state = 'dead', finished_at = now(), last_error = $2`
	//   args = []any{cause.Error()}
	// and return Dead.
	return "", errors.New("TODO(day 41): fail")
}

// Replay makes a dead job runnable again with a fresh attempt budget. Its history in
// job_attempts stays. It returns ErrNotDead for a job in any other state, so replaying a
// job twice, or a job that is running, cannot duplicate work.
func Replay(ctx context.Context, db *pgxpool.Pool, id int64) error {
	// TODO(day 42): one UPDATE that sets state 'pending', attempts 0, run_at now(),
	// finished_at NULL, but only WHERE id = $1 AND state = 'dead'. Zero rows affected
	// means "not dead": return an error wrapping ErrNotDead.
	return errors.New("TODO(day 42): replay")
}

// Get returns one job.
func Get(ctx context.Context, db *pgxpool.Pool, id int64) (Job, error) {
	j, err := scanJob(db.QueryRow(ctx, `SELECT `+jobColumns+` FROM jobs WHERE id = $1`, id))
	if err != nil {
		return Job{}, fmt.Errorf("get job %d: %w", id, err)
	}
	return j, nil
}

// ListDead returns up to limit dead jobs, oldest first: the operator's view of the
// dead-letter queue.
func ListDead(ctx context.Context, db *pgxpool.Pool, limit int) ([]Job, error) {
	rows, err := db.Query(ctx, `SELECT `+jobColumns+` FROM jobs WHERE state = 'dead' ORDER BY id LIMIT $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("list dead: %w", err)
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (Job, error) { return scanJob(row) })
}

// Attempts returns a job's attempt history in order.
func Attempts(ctx context.Context, db *pgxpool.Pool, id int64) ([]Attempt, error) {
	rows, err := db.Query(ctx, `
		SELECT attempt, COALESCE(worker, ''), COALESCE(error, '') FROM job_attempts
		WHERE job_id = $1 ORDER BY id`, id)
	if err != nil {
		return nil, fmt.Errorf("attempts of job %d: %w", id, err)
	}
	return pgx.CollectRows(rows, pgx.RowToStructByPos[Attempt])
}
