// Package e2e_test runs the whole chain against the real stack: orders, outbox, relay,
// Redpanda, consumer group, dedupe table.
package e2e_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"

	"example.com/ch13-outbox/bus"
	"example.com/ch13-outbox/consumer"
	"example.com/ch13-outbox/orders"
	"example.com/ch13-outbox/outbox"
	"example.com/ch13-outbox/testenv"
)

func TestRelayCrashAfterPublishBeforeMark(t *testing.T) {
	db := testenv.Postgres(t)
	brokers := testenv.Brokers(t)
	topic := testenv.Topic(t, brokers, 3)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	for _, c := range []string{"alice", "bob", "carol"} {
		if _, err := orders.Place(ctx, db, orders.Order{Customer: c, SKU: "book", Quantity: 1}); err != nil {
			t.Fatal(err)
		}
	}
	// Route this test's rows to its private topic instead of the shared orders.Topic.
	if _, err := db.Exec(ctx, `UPDATE outbox SET topic = $1`, topic); err != nil {
		t.Fatal(err)
	}

	producer, err := bus.NewProducer(brokers)
	if err != nil {
		t.Fatal(err)
	}
	defer producer.Close()
	pub := bus.KafkaPublisher{Client: producer}

	// Run 1: publish succeeds, then the relay "dies" before the mark commits.
	crash := errors.New("relay killed between publish and mark")
	first := &outbox.Relay{DB: db, Publisher: pub, AfterPublish: func(context.Context, []outbox.Message) error { return crash }}
	if _, err := first.RunOnce(ctx); !errors.Is(err, crash) {
		t.Fatalf("first relay: %v, want the simulated crash", err)
	}
	if p, _ := outbox.Pending(ctx, db); p != 3 {
		t.Fatalf("pending after crash = %d, want 3 (no business row lost its message)", p)
	}
	// Run 2: a restarted relay publishes the same rows again and marks them.
	second := &outbox.Relay{DB: db, Publisher: pub}
	if n, err := second.RunOnce(ctx); err != nil || n != 3 {
		t.Fatalf("restarted relay = %d, %v; want 3, nil", n, err)
	}

	// The consumer group reads everything on the topic and applies each message once.
	cl, err := bus.NewGroupConsumer(brokers, testenv.Group("shipping"), topic)
	if err != nil {
		t.Fatal(err)
	}
	defer cl.Close()
	runCtx, stop := context.WithCancel(ctx)
	records, applied := 0, 0
	ids := map[string]int{}
	err = consumer.Run(runCtx, cl, func(ctx context.Context, r *kgo.Record) error {
		m, err := consumer.FromRecord(r)
		if err != nil {
			return err
		}
		records++
		ids[m.ID]++
		ok, err := consumer.Process(ctx, db, "shipping", m, consumer.Ship)
		if err != nil {
			return err
		}
		if ok {
			applied++
		}
		if records == 6 {
			stop()
		}
		return nil
	})
	if !errors.Is(err, context.Canceled) || ctx.Err() != nil {
		t.Fatalf("consumer stopped with %v after %d records, want 6 records", err, records)
	}

	if len(ids) != 3 {
		t.Errorf("distinct message IDs on the topic = %d, want 3", len(ids))
	}
	for id, n := range ids {
		if n != 2 {
			t.Errorf("message %s published %d times, want 2 (the duplicate from the crash)", id, n)
		}
	}
	if applied != 3 {
		t.Errorf("effects applied = %d, want 3", applied)
	}
	var shipments, dedupe, ordersWithoutOne int
	q := db.QueryRow(ctx, `
		SELECT (SELECT count(*) FROM shipments),
		       (SELECT count(*) FROM processed_messages WHERE consumer = 'shipping'),
		       (SELECT count(*) FROM orders o WHERE (SELECT count(*) FROM shipments s WHERE s.order_id = o.id) <> 1)`)
	if err := q.Scan(&shipments, &dedupe, &ordersWithoutOne); err != nil {
		t.Fatal(err)
	}
	if shipments != 3 || dedupe != 3 || ordersWithoutOne != 0 {
		t.Errorf("shipments=%d dedupe=%d orders without exactly one shipment=%d; want 3, 3, 0", shipments, dedupe, ordersWithoutOne)
	}
}
