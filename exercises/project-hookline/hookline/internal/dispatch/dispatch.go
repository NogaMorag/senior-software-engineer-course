// Package dispatch is the dispatcher (milestone M4): a consumer-group member on hookline.events that
// turns each event into one delivery row per matching, enabled endpoint, and commits the offset only
// after the rows are committed (chapter 13). Reprocessing an event, after a relay duplicate or a crash
// before the commit, inserts nothing new: deliveries has UNIQUE (event_id, endpoint_id).
package dispatch

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/twmb/franz-go/pkg/kgo"

	"example.com/hookline/internal/outbox"
)

// Matcher answers "which endpoints get this event?"; *subcache.Cache is the real one.
type Matcher interface {
	Endpoints(ctx context.Context, tenantID, eventType string) ([]string, error)
}

// Dispatcher fans events out into deliveries.
type Dispatcher struct {
	DB      *pgxpool.Pool
	Matcher Matcher
	Log     *slog.Logger
}

// Fanout inserts one pending delivery per endpoint and reports how many rows are new. uuidv7() is
// built into PostgreSQL 18. The insert is one statement, so either all rows of the event exist or none.
func Fanout(ctx context.Context, db *pgxpool.Pool, m outbox.Message, endpointIDs []string) (int, error) {
	if len(endpointIDs) == 0 {
		return 0, nil
	}
	tag, err := db.Exec(ctx, `
		INSERT INTO deliveries (id, tenant_id, event_id, endpoint_id)
		SELECT uuidv7(), $1, $2, endpoint_id FROM unnest($3::uuid[]) AS endpoint_id
		ON CONFLICT (event_id, endpoint_id) DO NOTHING`, m.TenantID, m.EventID, endpointIDs)
	if err != nil {
		return 0, fmt.Errorf("fan out event %s: %w", m.EventID, err)
	}
	return int(tag.RowsAffected()), nil
}

// Handle processes one record. A record that cannot be decoded is logged and skipped: retrying it
// would block its partition forever (a poison message). Any other error is returned, and Run retries.
func (d *Dispatcher) Handle(ctx context.Context, r *kgo.Record) error {
	m, err := outbox.Decode(r.Value)
	if err != nil {
		d.Log.Error("skipping undecodable record", slog.String("topic", r.Topic), slog.Int("partition", int(r.Partition)),
			slog.Int64("offset", r.Offset), slog.String("error", err.Error()))
		return nil
	}
	ids, err := d.Matcher.Endpoints(ctx, m.TenantID, m.Type)
	if err != nil {
		return fmt.Errorf("match event %s: %w", m.EventID, err)
	}
	n, err := Fanout(ctx, d.DB, m, ids)
	if err != nil {
		return err
	}
	d.Log.Debug("dispatched", slog.String("event_id", m.EventID), slog.Int("endpoints", len(ids)), slog.Int("new", n))
	return nil
}

// Run polls cl until ctx is done and commits offsets only for records that handle finished. A failing
// handle (the database is down) is retried in place with a doubling wait up to 5 seconds, so records
// are neither skipped nor reordered. cl must come from bus.NewGroupConsumer (no auto-commit, rebalances
// blocked while a poll's records are being handled).
func Run(ctx context.Context, cl *kgo.Client, handle func(context.Context, *kgo.Record) error, log *slog.Logger) error {
	for {
		fetches := cl.PollRecords(ctx, 100)
		if ctx.Err() != nil {
			cl.AllowRebalance()
			return nil
		}
		if fetches.IsClientClosed() {
			return nil
		}
		if errs := fetches.Errors(); len(errs) > 0 {
			cl.AllowRebalance()
			return fmt.Errorf("fetch %s/%d: %w", errs[0].Topic, errs[0].Partition, errs[0].Err)
		}
		var done []*kgo.Record
		iter := fetches.RecordIter()
		for !iter.Done() && ctx.Err() == nil {
			r := iter.Next()
			if retry(ctx, log, func() error { return handle(ctx, r) }) {
				done = append(done, r)
			}
		}
		// Commit what was handled, with its own deadline so work finished just before a shutdown still
		// commits. Anything not handled is delivered again after a restart: at least once.
		if len(done) > 0 {
			cctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
			err := cl.CommitRecords(cctx, done...)
			cancel()
			if err != nil {
				cl.AllowRebalance()
				return fmt.Errorf("commit: %w", err)
			}
		}
		cl.AllowRebalance()
	}
}

// retry runs fn until it succeeds (true) or ctx is done (false).
func retry(ctx context.Context, log *slog.Logger, fn func() error) bool {
	wait := 100 * time.Millisecond
	for {
		err := fn()
		if err == nil {
			return true
		}
		log.Warn("dispatch failed; retrying", slog.String("error", err.Error()), slog.Duration("retry_in", wait))
		select {
		case <-ctx.Done():
			return false
		case <-time.After(wait):
		}
		wait = min(wait*2, 5*time.Second)
	}
}
