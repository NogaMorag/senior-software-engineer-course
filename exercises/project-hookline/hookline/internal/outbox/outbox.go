// Package outbox is the relay of the transactional outbox (chapter 13). The API already writes one
// outbox row per event in the event's own transaction (store/postgres IngestEvent); the relay reads the
// rows that are not published yet, publishes them to Redpanda in order, and marks them.
//
// The guarantee is "at least once": a crash after Publish and before the mark commits publishes the
// same rows again on the next run. The dispatcher's unique (event_id, endpoint_id) makes that harmless.
package outbox

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Topic is where events go. Its key is the tenant id, so one tenant's events stay in order.
const Topic = "hookline.events"

// RelayLockKey is the advisory lock key the relays campaign for (milestone M6, chapter 18): one relay
// holds it and publishes; the others wait. One active relay keeps each tenant's publish order simple.
const RelayLockKey int64 = 0x686f6f6b6c696e65 // "hookline" in ASCII

// ErrNotLeader means the mark was refused because a newer relay term exists: this relay lost the lock
// without noticing yet (a pause, a lost session). Its publish already happened, which is a duplicate the
// dispatcher absorbs; the mark must not happen, and the relay must step down.
var ErrNotLeader = errors.New("outbox: a newer relay term exists")

// HeaderEventID is the record header that carries the event id, the dedupe key downstream. The Kafka
// offset is not a dedupe key: a duplicate publish gets a new offset.
const HeaderEventID = "hookline-event-id"

// Message is one outbox row on its way to the log: the event's identity, not its payload. Workers
// read the payload from the events table when they send, so the log stays small.
type Message struct {
	Seq      int64  `json:"-"` // outbox.id: the publish order
	EventID  string `json:"event_id"`
	TenantID string `json:"tenant_id"`
	Type     string `json:"type"`
	// TraceContext is the ingest request's traceparent; it travels in the record headers (kotel), not
	// in the value.
	TraceContext string `json:"-"`
}

// Value is the record value: the message as JSON.
func (m Message) Value() []byte {
	b, _ := json.Marshal(m) // three strings: cannot fail
	return b
}

// Decode parses a record value made by Value.
func Decode(value []byte) (Message, error) {
	var m Message
	if err := json.Unmarshal(value, &m); err != nil {
		return Message{}, fmt.Errorf("decode event message: %w", err)
	}
	if m.EventID == "" || m.TenantID == "" || m.Type == "" {
		return Message{}, fmt.Errorf("decode event message: missing field in %s", value)
	}
	return m, nil
}

// Publisher sends a batch and returns only when the broker has acknowledged all of it, in order.
// bus.KafkaPublisher is the real one; tests use a fake.
type Publisher interface {
	Publish(ctx context.Context, msgs []Message) error
}

// Relay moves pending outbox rows to the log.
type Relay struct {
	DB        *pgxpool.Pool
	Publisher Publisher
	BatchSize int // rows per run; 0 means 100

	// AfterPublish, when set, runs after Publish succeeds and before the rows are marked. Production
	// code leaves it nil. Tests set it to simulate a crash at that exact point.
	AfterPublish func(ctx context.Context, published []Message) error

	// Term, when positive, is this relay's leader term (package leader) and fences the mark: the
	// UPDATE succeeds only while leader_terms still holds this term for RelayLockKey.
	Term int64
}

func (r *Relay) batch() int {
	if r.BatchSize <= 0 {
		return 100
	}
	return r.BatchSize
}

// RunOnce publishes up to BatchSize pending rows in outbox order and marks them published. It returns
// how many rows it marked. The rows stay locked (FOR UPDATE) from the read until the commit, so a
// second relay waits instead of publishing them too; milestone M6 makes one relay the only active one.
func (r *Relay) RunOnce(ctx context.Context) (int, error) {
	marked := 0
	err := pgx.BeginFunc(ctx, r.DB, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, `
			SELECT o.id, e.id, e.tenant_id, e.type, COALESCE(e.trace_context, '')
			FROM outbox o JOIN events e ON e.id = o.event_id
			WHERE o.published_at IS NULL
			ORDER BY o.id
			LIMIT $1
			FOR UPDATE OF o`, r.batch())
		if err != nil {
			return fmt.Errorf("read outbox: %w", err)
		}
		msgs, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (Message, error) {
			var m Message
			err := row.Scan(&m.Seq, &m.EventID, &m.TenantID, &m.Type, &m.TraceContext)
			return m, err
		})
		if err != nil {
			return fmt.Errorf("read outbox: %w", err)
		}
		if len(msgs) == 0 {
			return nil
		}
		// 1. Publish. If this fails, nothing is marked and the rows are retried later.
		if err := r.Publisher.Publish(ctx, msgs); err != nil {
			return fmt.Errorf("publish %d messages: %w", len(msgs), err)
		}
		// The dangerous gap: the broker has the messages, the database does not know yet.
		if r.AfterPublish != nil {
			if err := r.AfterPublish(ctx, msgs); err != nil {
				return err
			}
		}
		// 2. Mark, only after the broker acknowledged every message.
		seqs := make([]int64, len(msgs))
		for i, m := range msgs {
			seqs[i] = m.Seq
		}
		tag, err := tx.Exec(ctx, `
			UPDATE outbox SET published_at = now()
			WHERE id = ANY($1)
			  AND ($2::bigint = 0 OR (SELECT term FROM leader_terms WHERE lock_key = $3) = $2)`,
			seqs, r.Term, RelayLockKey)
		if err != nil {
			return fmt.Errorf("mark published: %w", err)
		}
		if tag.RowsAffected() == 0 {
			return ErrNotLeader // the fencing token was refused; the transaction rolls back
		}
		marked = len(msgs)
		return nil
	})
	if err != nil {
		return 0, err
	}
	return marked, nil
}

// Run calls RunOnce until ctx is done, or until the mark is fenced off (ErrNotLeader): at once again after a full batch, after idle when there was
// nothing to do, and with a doubling wait (up to 30 times idle) while publishing fails, for example
// while Redpanda is down. Ingest keeps working meanwhile; the rows wait in the outbox.
func (r *Relay) Run(ctx context.Context, idle time.Duration, log *slog.Logger) error {
	wait := idle
	for {
		n, err := r.RunOnce(ctx)
		switch {
		case ctx.Err() != nil:
			return nil
		case errors.Is(err, ErrNotLeader):
			return err // step down; the elector campaigns again
		case err != nil:
			log.Warn("relay run failed; retrying", slog.String("error", err.Error()), slog.Duration("retry_in", wait))
			wait = min(wait*2, 30*idle)
		case n == r.batch():
			wait = idle
			continue // probably more waiting: no pause
		default:
			wait = idle
			if n > 0 {
				log.Debug("relay published", slog.Int("messages", n))
			}
		}
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(wait):
		}
	}
}

// Pending counts rows that are not yet published.
func Pending(ctx context.Context, db *pgxpool.Pool) (int, error) {
	n, _, err := Stats(ctx, db)
	return int(n), err
}

// Stats returns how many rows wait and how old the oldest is (0 when none): the two outbox gauges of
// milestone M8, and the signal of the OutboxStalled alert. The partial index on unpublished rows keeps
// this cheap however long the published history grows.
func Stats(ctx context.Context, db *pgxpool.Pool) (int64, time.Duration, error) {
	var n int64
	var oldest float64
	err := db.QueryRow(ctx, `
		SELECT count(*), COALESCE(EXTRACT(EPOCH FROM now() - min(created_at)), 0)::float8
		FROM outbox WHERE published_at IS NULL`).Scan(&n, &oldest)
	return n, time.Duration(oldest * float64(time.Second)), err
}
