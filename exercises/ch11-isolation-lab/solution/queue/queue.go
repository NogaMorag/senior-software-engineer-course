// Package queue is a minimal work queue on PostgreSQL: many workers, no job taken twice.
package queue

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Job is one queued unit of work.
type Job struct {
	ID      int64
	Payload string
}

// Enqueue adds a job and returns its id.
func Enqueue(ctx context.Context, pool *pgxpool.Pool, payload string) (int64, error) {
	var id int64
	err := pool.QueryRow(ctx, `INSERT INTO jobs (payload) VALUES ($1) RETURNING id`, payload).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("enqueue: %w", err)
	}
	return id, nil
}

// Work claims the oldest queued job that no other worker is holding, runs handle inside
// the same transaction, and marks the job done when handle succeeds. It returns false
// when there is nothing to claim. If handle fails, the transaction rolls back and the
// job is queued again for any worker. The row lock lasts as long as handle runs, so
// handle must be short; chapter 14 replaces this with a lease for long jobs.
func Work(ctx context.Context, pool *pgxpool.Pool, worker string, handle func(context.Context, pgx.Tx, Job) error) (bool, error) {
	claimed := false
	err := pgx.BeginFunc(ctx, pool, func(tx pgx.Tx) error {
		var job Job
		err := tx.QueryRow(ctx, `
			SELECT id, payload FROM jobs
			WHERE status = 'queued'
			ORDER BY id
			LIMIT 1
			FOR UPDATE SKIP LOCKED`).Scan(&job.ID, &job.Payload)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("claim job: %w", err)
		}
		if err := handle(ctx, tx, job); err != nil {
			return fmt.Errorf("job %d: %w", job.ID, err)
		}
		if _, err := tx.Exec(ctx, `
			UPDATE jobs SET status = 'done', done_by = $2, done_at = now()
			WHERE id = $1`, job.ID, worker); err != nil {
			return fmt.Errorf("finish job %d: %w", job.ID, err)
		}
		claimed = true
		return nil
	})
	return claimed, err
}
