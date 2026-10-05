package orders_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// epoch is the time of order 0; anOrder(n) is created n hours later unless the test says otherwise.
var epoch = time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)

// orderBuilder is a test data builder. anOrder(n) is already a valid order (customer 1, delivered,
// one item); a test calls only the methods for the fields it cares about, so the fixture says
// what matters to the test and nothing else. A new NOT NULL column changes insert, not every test.
type orderBuilder struct {
	id         int64
	customerID int64
	status     string
	createdAt  time.Time
	items      int
}

func anOrder(id int64) *orderBuilder {
	return &orderBuilder{id: id, customerID: 1, status: "delivered", createdAt: epoch.Add(time.Duration(id) * time.Hour), items: 1}
}

func (b *orderBuilder) forCustomer(id int64) *orderBuilder { b.customerID = id; return b }
func (b *orderBuilder) at(t time.Time) *orderBuilder       { b.createdAt = t; return b }
func (b *orderBuilder) withItems(n int) *orderBuilder      { b.items = n; return b }

// insert writes the order and its items. Item k (1-based) is product k, quantity k, 100*k cents,
// so a test can predict every item from the count alone. Products 1..5 must exist (seedCatalog).
func (b *orderBuilder) insert(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()
	if _, err := pool.Exec(ctx, `
		INSERT INTO orders (id, customer_id, status, total_cents, created_at)
		VALUES ($1, $2, $3, 0, $4)`, b.id, b.customerID, b.status, b.createdAt); err != nil {
		t.Fatalf("insert order %d: %v", b.id, err)
	}
	for k := 1; k <= b.items; k++ {
		if _, err := pool.Exec(ctx, `
			INSERT INTO order_items (order_id, product_id, quantity, price_cents)
			VALUES ($1, $2, $3, $4)`, b.id, k, k, 100*k); err != nil {
			t.Fatalf("insert item %d of order %d: %v", k, b.id, err)
		}
	}
}

// seedCatalog inserts customers 1 and 2 and products 1..5: what every order fixture refers to.
func seedCatalog(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	if _, err := pool.Exec(context.Background(), `
		INSERT INTO customers (id, email, name, country, created_at) VALUES
			(1, 'a@example.com', 'A', 'IL', '2026-01-01T00:00:00Z'),
			(2, 'b@example.com', 'B', 'IL', '2026-01-01T00:00:00Z');
		INSERT INTO products (id, name, price_cents)
			SELECT p, 'P' || p, 100 * p FROM generate_series(1, 5) AS p`); err != nil {
		t.Fatalf("seed catalog: %v", err)
	}
}

// itemsOf is the number of items the standard fixture gives order n.
func itemsOf(n int64) int { return 1 + int(n%3) }

// standardFixture is the chapter 10 dataset: customer 1 has orders 1..25, one per hour, except
// that order 12 shares order 11's timestamp (the keyset tie-breaker case). Customer 2 has none.
func standardFixture(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	seedCatalog(t, pool)
	for n := int64(1); n <= 25; n++ {
		b := anOrder(n).withItems(itemsOf(n))
		if n == 12 {
			b.at(epoch.Add(11 * time.Hour))
		}
		b.insert(t, pool)
	}
}
