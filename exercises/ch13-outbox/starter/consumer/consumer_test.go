package consumer_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"example.com/ch13-outbox/consumer"
	"example.com/ch13-outbox/orders"
	"example.com/ch13-outbox/testenv"
)

func placed(t *testing.T, db *pgxpool.Pool) (int64, []byte) {
	t.Helper()
	ctx := context.Background()
	var id int64
	if err := db.QueryRow(ctx, `INSERT INTO orders (customer, sku, quantity) VALUES ('alice', 'book', 1) RETURNING id`).Scan(&id); err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(orders.Placed{OrderID: id, Customer: "alice", SKU: "book", Quantity: 1})
	if err != nil {
		t.Fatal(err)
	}
	return id, payload
}

func shipments(t *testing.T, db *pgxpool.Pool) int {
	t.Helper()
	var n int
	if err := db.QueryRow(context.Background(), `SELECT count(*) FROM shipments`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func TestProcess(t *testing.T) {
	const msgA = "00000000-0000-4000-8000-00000000000a"
	const msgB = "00000000-0000-4000-8000-00000000000b"
	tests := []struct {
		name         string
		deliveries   []string // message IDs in delivery order
		wantApplied  []bool
		wantShipment int
	}{
		{"one delivery", []string{msgA}, []bool{true}, 1},
		{"duplicate delivery", []string{msgA, msgA}, []bool{true, false}, 1},
		{"three deliveries of one message", []string{msgA, msgA, msgA}, []bool{true, false, false}, 1},
		{"two different messages", []string{msgA, msgB}, []bool{true, true}, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := testenv.Postgres(t)
			_, payload := placed(t, db)
			for i, id := range tt.deliveries {
				got, err := consumer.Process(context.Background(), db, "shipping", consumer.Message{ID: id, Payload: payload}, consumer.Ship)
				if err != nil {
					t.Fatal(err)
				}
				if got != tt.wantApplied[i] {
					t.Errorf("delivery %d applied = %v, want %v", i, got, tt.wantApplied[i])
				}
			}
			if n := shipments(t, db); n != tt.wantShipment {
				t.Errorf("shipments = %d, want %d", n, tt.wantShipment)
			}
		})
	}
}

func TestDedupeIsPerConsumer(t *testing.T) {
	db := testenv.Postgres(t)
	_, payload := placed(t, db)
	m := consumer.Message{ID: "00000000-0000-4000-8000-00000000000c", Payload: payload}
	noop := func(context.Context, pgx.Tx, consumer.Message) error { return nil }
	for _, name := range []string{"shipping", "billing"} {
		ok, err := consumer.Process(context.Background(), db, name, m, noop)
		if err != nil || !ok {
			t.Fatalf("consumer %s: applied = %v, %v; want true: each consumer dedupes on its own", name, ok, err)
		}
	}
}

func TestFailedEffectLeavesNoDedupeRow(t *testing.T) {
	db := testenv.Postgres(t)
	ctx := context.Background()
	// An event for an order that does not exist: the shipment insert violates the FK.
	payload, _ := json.Marshal(orders.Placed{OrderID: 999999, Customer: "ghost", SKU: "x", Quantity: 1})
	m := consumer.Message{ID: "00000000-0000-4000-8000-00000000000d", Payload: payload}
	if _, err := consumer.Process(ctx, db, "shipping", m, consumer.Ship); err == nil {
		t.Fatal("Process succeeded, want the foreign-key error")
	}
	var n int
	if err := db.QueryRow(ctx, `SELECT count(*) FROM processed_messages`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Errorf("dedupe rows = %d, want 0: a failed effect must not be remembered as done", n)
	}
}
