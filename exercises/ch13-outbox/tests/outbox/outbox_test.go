package outbox_test

import (
	"context"
	"errors"
	"slices"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"example.com/ch13-outbox/consumer"
	"example.com/ch13-outbox/orders"
	"example.com/ch13-outbox/outbox"
	"example.com/ch13-outbox/testenv"
)

// fakePublisher records every message it was asked to publish, like a broker would.
type fakePublisher struct {
	mu   sync.Mutex
	sent []outbox.Message
	fail error
}

func (p *fakePublisher) Publish(_ context.Context, msgs []outbox.Message) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.fail != nil {
		return p.fail
	}
	p.sent = append(p.sent, msgs...)
	return nil
}

func count(t *testing.T, db *pgxpool.Pool, query string, args ...any) int {
	t.Helper()
	var n int
	if err := db.QueryRow(context.Background(), query, args...).Scan(&n); err != nil {
		t.Fatalf("%s: %v", query, err)
	}
	return n
}

func TestPlaceWritesOrderAndOutboxTogether(t *testing.T) {
	db := testenv.Postgres(t)
	ctx := context.Background()
	id, err := orders.Place(ctx, db, orders.Order{Customer: "alice", SKU: "book", Quantity: 2})
	if err != nil {
		t.Fatal(err)
	}
	if got := count(t, db, `SELECT count(*) FROM orders WHERE id = $1`, id); got != 1 {
		t.Fatalf("orders rows = %d, want 1", got)
	}
	var topic, key string
	var orderID int64
	if err := db.QueryRow(ctx, `SELECT topic, key, (payload->>'order_id')::bigint FROM outbox`).Scan(&topic, &key, &orderID); err != nil {
		t.Fatalf("outbox row: %v", err)
	}
	if topic != orders.Topic || key != "alice" || orderID != id {
		t.Fatalf("outbox row = (%s, %s, %d), want (%s, alice, %d)", topic, key, orderID, orders.Topic, id)
	}
}

func TestFailedPlaceWritesNeither(t *testing.T) {
	db := testenv.Postgres(t)
	// quantity 0 violates the CHECK constraint, so the transaction rolls back.
	if _, err := orders.Place(context.Background(), db, orders.Order{Customer: "alice", SKU: "book", Quantity: 0}); err == nil {
		t.Fatal("Place with quantity 0 succeeded, want an error")
	}
	if n := count(t, db, `SELECT count(*) FROM orders`); n != 0 {
		t.Errorf("orders rows = %d, want 0", n)
	}
	if n := count(t, db, `SELECT count(*) FROM outbox`); n != 0 {
		t.Errorf("outbox rows = %d, want 0: a message without its business row", n)
	}
}

func TestRelayPublishesInOrderAndMarks(t *testing.T) {
	db := testenv.Postgres(t)
	ctx := context.Background()
	for _, c := range []string{"alice", "bob", "alice", "carol", "alice"} {
		if _, err := orders.Place(ctx, db, orders.Order{Customer: c, SKU: "book", Quantity: 1}); err != nil {
			t.Fatal(err)
		}
	}
	pub := &fakePublisher{}
	relay := &outbox.Relay{DB: db, Publisher: pub, BatchSize: 3}
	total := 0
	for {
		n, err := relay.RunOnce(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if n == 0 {
			break
		}
		total += n
	}
	if total != 5 || len(pub.sent) != 5 {
		t.Fatalf("marked %d and published %d, want 5 and 5", total, len(pub.sent))
	}
	seqs := make([]int64, len(pub.sent))
	for i, m := range pub.sent {
		seqs[i] = m.Seq
		if m.ID == "" {
			t.Errorf("message %d has no ID", i)
		}
	}
	if !slices.IsSorted(seqs) {
		t.Errorf("published out of outbox order: %v", seqs)
	}
	if p, err := outbox.Pending(ctx, db); err != nil || p != 0 {
		t.Errorf("pending = %d, %v; want 0", p, err)
	}
}

func TestRelayPublishFailureMarksNothing(t *testing.T) {
	db := testenv.Postgres(t)
	ctx := context.Background()
	if _, err := orders.Place(ctx, db, orders.Order{Customer: "alice", SKU: "book", Quantity: 1}); err != nil {
		t.Fatal(err)
	}
	broker := errors.New("broker unavailable")
	relay := &outbox.Relay{DB: db, Publisher: &fakePublisher{fail: broker}}
	if _, err := relay.RunOnce(ctx); !errors.Is(err, broker) {
		t.Fatalf("RunOnce error = %v, want it to wrap %v", err, broker)
	}
	if p, _ := outbox.Pending(ctx, db); p != 1 {
		t.Errorf("pending = %d, want 1: an unpublished row was marked", p)
	}
}

// The crash point: the broker has the message, the mark never commits. The next relay
// publishes it again; the consumer's dedupe table applies it once.
func TestCrashBetweenPublishAndMarkDuplicatesButAppliesOnce(t *testing.T) {
	db := testenv.Postgres(t)
	ctx := context.Background()
	orderID, err := orders.Place(ctx, db, orders.Order{Customer: "alice", SKU: "book", Quantity: 1})
	if err != nil {
		t.Fatal(err)
	}
	pub := &fakePublisher{}
	crash := errors.New("relay killed")
	crashing := &outbox.Relay{DB: db, Publisher: pub, AfterPublish: func(context.Context, []outbox.Message) error { return crash }}
	if _, err := crashing.RunOnce(ctx); !errors.Is(err, crash) {
		t.Fatalf("crashing relay error = %v, want %v", err, crash)
	}
	if p, _ := outbox.Pending(ctx, db); p != 1 {
		t.Fatalf("after the crash pending = %d, want 1", p)
	}
	restarted := &outbox.Relay{DB: db, Publisher: pub}
	if n, err := restarted.RunOnce(ctx); err != nil || n != 1 {
		t.Fatalf("restarted relay = %d, %v; want 1, nil", n, err)
	}
	if len(pub.sent) != 2 || pub.sent[0].ID != pub.sent[1].ID {
		t.Fatalf("published %d messages, want the same message twice", len(pub.sent))
	}

	applied := 0
	for _, m := range pub.sent {
		ok, err := consumer.Process(ctx, db, "shipping", consumer.Message{ID: m.ID, Key: m.Key, Payload: m.Payload}, consumer.Ship)
		if err != nil {
			t.Fatal(err)
		}
		if ok {
			applied++
		}
	}
	if applied != 1 {
		t.Errorf("applied %d times, want 1", applied)
	}
	if n := count(t, db, `SELECT count(*) FROM shipments WHERE order_id = $1`, orderID); n != 1 {
		t.Errorf("shipments = %d, want exactly 1", n)
	}
	if n := count(t, db, `SELECT count(*) FROM processed_messages WHERE consumer = 'shipping'`); n != 1 {
		t.Errorf("dedupe rows = %d, want 1", n)
	}
}
