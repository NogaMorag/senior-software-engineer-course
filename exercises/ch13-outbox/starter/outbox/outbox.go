// Package outbox implements the transactional outbox: Enqueue writes a message row inside
// the caller's transaction, and a Relay publishes pending rows and then marks them.
//
// The guarantee is "at least once": a crash after Publish and before the mark commits
// publishes the same rows again on the next run. Consumers dedupe by Message.ID.
package outbox

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Message is one outbox row on its way to the log.
type Message struct {
	Seq     int64  // outbox.id: the publish order
	ID      string // outbox.message_id: the dedupe key consumers use
	Topic   string
	Key     string
	Payload []byte // JSON
}

// Publisher sends a batch of messages and returns only when the broker has acknowledged
// all of them, in order. bus.KafkaPublisher is the real one; tests use a fake.
type Publisher interface {
	Publish(ctx context.Context, msgs []Message) error
}

// Enqueue inserts a message into the outbox inside tx, the caller's transaction. It does
// not talk to the broker: if tx rolls back, the message never existed.
func Enqueue(ctx context.Context, tx pgx.Tx, topic, key string, payload []byte) (string, error) {
	var id string
	err := tx.QueryRow(ctx, `
		INSERT INTO outbox (topic, key, payload) VALUES ($1, $2, $3)
		RETURNING message_id::text`, topic, key, payload).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("outbox enqueue: %w", err)
	}
	return id, nil
}

// Relay moves pending outbox rows to the log.
type Relay struct {
	DB        *pgxpool.Pool
	Publisher Publisher
	BatchSize int // rows per run; 0 means 100

	// AfterPublish, when set, runs after Publish succeeds and before the rows are marked.
	// Production code leaves it nil. Tests set it to simulate a crash at that exact point.
	AfterPublish func(ctx context.Context, published []Message) error
}

// RunOnce publishes up to BatchSize pending rows in outbox order and marks them published.
// It returns how many rows it marked. The rows stay locked (FOR UPDATE) from the read
// until the commit, so a second relay waits instead of publishing the same rows, and the
// mark happens in the same transaction that read them.
func (r *Relay) RunOnce(ctx context.Context) (int, error) {
	limit := r.BatchSize
	if limit <= 0 {
		limit = 100
	}
	marked := 0
	err := pgx.BeginFunc(ctx, r.DB, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, `
			SELECT id, message_id::text, topic, key, payload::text FROM outbox
			WHERE published_at IS NULL
			ORDER BY id
			LIMIT $1
			FOR UPDATE`, limit)
		if err != nil {
			return fmt.Errorf("read outbox: %w", err)
		}
		msgs, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (Message, error) {
			var m Message
			var payload string
			err := row.Scan(&m.Seq, &m.ID, &m.Topic, &m.Key, &payload)
			m.Payload = []byte(payload)
			return m, err
		})
		if err != nil {
			return fmt.Errorf("read outbox: %w", err)
		}
		if len(msgs) == 0 {
			return nil
		}
		// TODO(day 38):
		//  1. r.Publisher.Publish(ctx, msgs); on error return "publish %d messages: %w".
		//  2. if r.AfterPublish != nil, call it and return its error (the crash point).
		//  3. UPDATE outbox SET published_at = now() WHERE id = ANY($1) with the Seq values.
		return errors.New("TODO(day 38): publish, then mark")
	})
	if err != nil {
		return 0, err
	}
	return marked, nil
}

// Pending counts rows that are not yet marked published.
func Pending(ctx context.Context, db *pgxpool.Pool) (int, error) {
	var n int
	err := db.QueryRow(ctx, `SELECT count(*) FROM outbox WHERE published_at IS NULL`).Scan(&n)
	return n, err
}
