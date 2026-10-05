// Package consumer is the receiving side: a dedupe table that turns at-least-once
// delivery into an effect applied once, and a poll loop that commits offsets only after
// processing.
package consumer

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/twmb/franz-go/pkg/kgo"

	"example.com/ch13-outbox/orders"
	"example.com/ch13-outbox/outbox"
)

// Message is what the consumer needs from a record.
type Message struct {
	ID      string // from the message-id header
	Key     string
	Payload []byte
}

// FromRecord extracts a Message from a Kafka record.
func FromRecord(r *kgo.Record) (Message, error) {
	for _, h := range r.Headers {
		if h.Key == outbox.HeaderMessageID {
			return Message{ID: string(h.Value), Key: string(r.Key), Payload: r.Value}, nil
		}
	}
	return Message{}, fmt.Errorf("record %s/%d@%d has no %s header", r.Topic, r.Partition, r.Offset, outbox.HeaderMessageID)
}

// Effect applies the business effect of one message inside tx.
type Effect func(ctx context.Context, tx pgx.Tx, m Message) error

// Process applies effect at most once per (consumer, message ID). The dedupe row and the
// effect commit in the same transaction: if the effect fails, the dedupe row rolls back
// too, and the next delivery tries again. It reports whether this call applied the effect
// (false means the message was a duplicate).
func Process(ctx context.Context, db *pgxpool.Pool, consumer string, m Message, effect Effect) (bool, error) {
	// TODO(day 38): in ONE transaction:
	//  1. INSERT INTO processed_messages (consumer, message_id) VALUES ($1, $2) ON CONFLICT DO NOTHING
	//  2. if no row was inserted (tag.RowsAffected() == 0), it is a duplicate: commit, return false
	//  3. otherwise run effect(ctx, tx, m); its error rolls back the dedupe row as well
	// Wrap errors as "process message %s: %w".
	return false, errors.New("TODO(day 38): dedupe and apply")
}

// Ship is the effect for orders.placed: one shipment row per order.
func Ship(ctx context.Context, tx pgx.Tx, m Message) error {
	p, err := orders.DecodePlaced(m.Payload)
	if err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO shipments (order_id) VALUES ($1)`, p.OrderID); err != nil {
		return fmt.Errorf("insert shipment for order %d: %w", p.OrderID, err)
	}
	return nil
}

// Run polls cl until ctx is done or handle fails, and commits offsets only for records
// that handle finished. A record whose handling failed is not committed, so after a
// restart the group delivers it again: at least once, never "at most once".
//
// cl must be a group consumer built with DisableAutoCommit and BlockRebalanceOnPoll
// (bus.NewGroupConsumer does this): the rebalance waits for AllowRebalance, so offsets
// are never committed for partitions this member no longer owns.
func Run(ctx context.Context, cl *kgo.Client, handle func(context.Context, *kgo.Record) error) error {
	for {
		fetches := cl.PollRecords(ctx, 100)
		if ctx.Err() != nil {
			cl.AllowRebalance()
			return ctx.Err()
		}
		if fetches.IsClientClosed() {
			return nil
		}
		if errs := fetches.Errors(); len(errs) > 0 {
			cl.AllowRebalance()
			return fmt.Errorf("fetch %s/%d: %w", errs[0].Topic, errs[0].Partition, errs[0].Err)
		}
		var done []*kgo.Record
		var handleErr error
		iter := fetches.RecordIter()
		for !iter.Done() {
			r := iter.Next()
			if err := handle(ctx, r); err != nil {
				handleErr = fmt.Errorf("handle %s/%d@%d: %w", r.Topic, r.Partition, r.Offset, err)
				break
			}
			done = append(done, r)
		}
		// TODO(day 37): commit the offsets of the records in done, and only those, with
		// cl.CommitRecords. Give the commit its own 10-second deadline built from
		// context.WithoutCancel(ctx), so that work finished just before a shutdown is kept.
		// On a commit error: cl.AllowRebalance(), then return "commit: %w".
		_ = done
		_ = time.Second
		cl.AllowRebalance()
		if handleErr != nil {
			return handleErr
		}
	}
}
