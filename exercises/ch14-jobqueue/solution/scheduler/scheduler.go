// Package scheduler is an in-process periodic scheduler that is safe to run in every
// replica: each tick enqueues the job for the current time slot with a unique key, so
// two replicas ticking in the same slot create one job, not two.
package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"example.com/ch14-jobqueue/queue"
)

// Slot is the start of the interval of length every that contains t.
func Slot(t time.Time, every time.Duration) time.Time {
	return t.UTC().Truncate(every)
}

// SlotKey names one run of a schedule: "nightly-report@2026-10-05T02:00:00Z".
func SlotKey(name string, slot time.Time) string {
	return fmt.Sprintf("%s@%s", name, slot.UTC().Format(time.RFC3339))
}

// EnqueueDue enqueues the run for the slot containing now. Every replica may call it;
// the unique key lets exactly one call create the job. It reports whether this call did.
func EnqueueDue(ctx context.Context, db *pgxpool.Pool, name, kind string, payload []byte, every time.Duration, now time.Time) (int64, bool, error) {
	slot := Slot(now, every)
	return queue.Enqueue(ctx, db, kind, payload, queue.Options{UniqueKey: SlotKey(name, slot)})
}

// Run ticks every interval until ctx is cancelled. The scheduler only enqueues; workers
// do the work, so a slow job never delays the next tick, and a crashed scheduler loses
// at most the ticks it was down for.
func Run(ctx context.Context, db *pgxpool.Pool, name, kind string, payload []byte, every time.Duration) error {
	t := time.NewTicker(every)
	defer t.Stop()
	for {
		if _, _, err := EnqueueDue(ctx, db, name, kind, payload, every, time.Now()); err != nil && ctx.Err() == nil {
			return err
		}
		select {
		case <-ctx.Done():
			return nil
		case <-t.C:
		}
	}
}
