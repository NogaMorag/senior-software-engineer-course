// Package queue is a minimal work queue on PostgreSQL: many workers, no job taken twice.
package queue

import (
	"context"
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
	// TODO(day 27): in one transaction: SELECT the oldest queued job FOR UPDATE SKIP LOCKED (none: return false), run handle, mark it done. See chapter 11, section 11.4.
	panic("TODO: Work, see chapter 11, section 11.4")
}
