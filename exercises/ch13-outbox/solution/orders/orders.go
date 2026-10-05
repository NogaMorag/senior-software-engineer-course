// Package orders is the business side of the exercise: placing an order writes the order
// row and its "order placed" message in one transaction.
package orders

import (
	"context"
	"encoding/json"
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
	var orderID int64
	err := pgx.BeginFunc(ctx, db, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, `
			INSERT INTO orders (customer, sku, quantity) VALUES ($1, $2, $3)
			RETURNING id`, o.Customer, o.SKU, o.Quantity).Scan(&orderID); err != nil {
			return fmt.Errorf("insert order: %w", err)
		}
		payload, err := json.Marshal(Placed{OrderID: orderID, Customer: o.Customer, SKU: o.SKU, Quantity: o.Quantity})
		if err != nil {
			return err
		}
		_, err = outbox.Enqueue(ctx, tx, Topic, o.Customer, payload)
		return err
	})
	if err != nil {
		return 0, fmt.Errorf("place order: %w", err)
	}
	return orderID, nil
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
