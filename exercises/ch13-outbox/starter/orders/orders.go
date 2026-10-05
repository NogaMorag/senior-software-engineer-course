// Package orders is the business side of the exercise: placing an order writes the order
// row and its "order placed" message in one transaction.
package orders

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"example.com/ch13-outbox/outbox"
)

// Topic is where "order placed" events go.
const Topic = "orders.placed"

// Order is what a customer submits.
type Order struct {
	Customer string `json:"customer"`
	SKU      string `json:"sku"`
	Quantity int    `json:"quantity"`
}

// Placed is the event payload.
type Placed struct {
	OrderID  int64  `json:"order_id"`
	Customer string `json:"customer"`
	SKU      string `json:"sku"`
	Quantity int    `json:"quantity"`
}

// Place inserts the order and enqueues its event in the same transaction, keyed by
// customer so that one customer's events stay in order. There is no "dual write": the
// broker is never called here.
func Place(ctx context.Context, db *pgxpool.Pool, o Order) (int64, error) {
	// TODO(day 38): in ONE transaction (pgx.BeginFunc), insert the order RETURNING id,
	// marshal a Placed event, and call outbox.Enqueue(ctx, tx, Topic, o.Customer, payload).
	// Wrap any error as "place order: %w".
	_, _, _ = pgx.BeginFunc, json.Marshal, outbox.Enqueue
	return 0, errors.New("TODO(day 38): place order")
}

// DecodePlaced parses an event payload.
func DecodePlaced(payload []byte) (Placed, error) {
	var p Placed
	if err := json.Unmarshal(payload, &p); err != nil {
		return Placed{}, fmt.Errorf("decode %s event: %w", Topic, err)
	}
	if p.OrderID <= 0 {
		return Placed{}, fmt.Errorf("decode %s event: bad order_id %d", Topic, p.OrderID)
	}
	return p, nil
}
