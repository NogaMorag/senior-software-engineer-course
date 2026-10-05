package orders

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Order is one row of orders, optionally with its items.
type Order struct {
	ID         int64
	CustomerID int64
	Status     string
	TotalCents int64
	CreatedAt  time.Time
	Items      []Item
}

// Item is one line of an order.
type Item struct {
	ProductID  int64
	Quantity   int32
	PriceCents int64
}

// Cursor marks the last row of a page: "continue after this one".
// (created_at, id) is unique even when two orders share a timestamp.
type Cursor struct {
	CreatedAt time.Time
	ID        int64
}

// Repo reads orders. It holds a pool, never a single connection.
type Repo struct {
	pool *pgxpool.Pool
}

// New returns a repository over pool.
func New(pool *pgxpool.Pool) *Repo { return &Repo{pool: pool} }

const orderColumns = `id, customer_id, status, total_cents, created_at`

func scanOrder(row pgx.CollectableRow) (Order, error) {
	var o Order
	err := row.Scan(&o.ID, &o.CustomerID, &o.Status, &o.TotalCents, &o.CreatedAt)
	return o, err
}

// RecentOrders returns the customer's newest orders, newest first (query Q3).
func (r *Repo) RecentOrders(ctx context.Context, customerID int64, limit int) ([]Order, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+orderColumns+`
		FROM orders
		WHERE customer_id = $1
		ORDER BY created_at DESC, id DESC
		LIMIT $2`, customerID, limit)
	if err != nil {
		return nil, fmt.Errorf("recent orders: %w", err)
	}
	return pgx.CollectRows(rows, scanOrder)
}

// OrdersWithItemsNPlusOne is the slow version, kept for comparison: one query for
// the orders, then one more query per order for its items (1 + N round trips).
func (r *Repo) OrdersWithItemsNPlusOne(ctx context.Context, customerID int64, limit int) ([]Order, error) {
	orders, err := r.RecentOrders(ctx, customerID, limit)
	if err != nil {
		return nil, err
	}
	for i := range orders {
		rows, err := r.pool.Query(ctx, `
			SELECT product_id, quantity, price_cents
			FROM order_items
			WHERE order_id = $1
			ORDER BY id`, orders[i].ID)
		if err != nil {
			return nil, fmt.Errorf("items of order %d: %w", orders[i].ID, err)
		}
		items, err := pgx.CollectRows(rows, pgx.RowToStructByPos[Item])
		if err != nil {
			return nil, fmt.Errorf("items of order %d: %w", orders[i].ID, err)
		}
		orders[i].Items = items
	}
	return orders, nil
}

// OrdersWithItems returns the same result as OrdersWithItemsNPlusOne in exactly
// two queries: the orders, then all their items at once with = ANY($1).
func (r *Repo) OrdersWithItems(ctx context.Context, customerID int64, limit int) ([]Order, error) {
	orders, err := r.RecentOrders(ctx, customerID, limit)
	if err != nil || len(orders) == 0 {
		return orders, err
	}
	ids := make([]int64, len(orders))
	byID := make(map[int64]*Order, len(orders))
	for i := range orders {
		ids[i] = orders[i].ID
		orders[i].Items = []Item{}
		byID[orders[i].ID] = &orders[i]
	}
	rows, err := r.pool.Query(ctx, `
		SELECT order_id, product_id, quantity, price_cents
		FROM order_items
		WHERE order_id = ANY($1)
		ORDER BY order_id, id`, ids)
	if err != nil {
		return nil, fmt.Errorf("items of %d orders: %w", len(ids), err)
	}
	var orderID int64
	var it Item
	_, err = pgx.ForEachRow(rows, []any{&orderID, &it.ProductID, &it.Quantity, &it.PriceCents}, func() error {
		o := byID[orderID]
		o.Items = append(o.Items, it)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("items of %d orders: %w", len(ids), err)
	}
	return orders, nil
}

// OrdersPage returns one page of the customer's orders, newest first, using keyset
// pagination: after == nil means the first page. The returned cursor is nil on the
// last page. Each page costs the same no matter how deep it is, because the index
// on (customer_id, created_at, id) lets PostgreSQL start right after the cursor.
func (r *Repo) OrdersPage(ctx context.Context, customerID int64, after *Cursor, limit int) ([]Order, *Cursor, error) {
	if limit <= 0 {
		return nil, nil, fmt.Errorf("orders page: limit must be positive, got %d", limit)
	}
	// Ask for one extra row: if it exists, there is a next page.
	var rows pgx.Rows
	var err error
	if after == nil {
		rows, err = r.pool.Query(ctx, `
			SELECT `+orderColumns+`
			FROM orders
			WHERE customer_id = $1
			ORDER BY created_at DESC, id DESC
			LIMIT $2`, customerID, limit+1)
	} else {
		rows, err = r.pool.Query(ctx, `
			SELECT `+orderColumns+`
			FROM orders
			WHERE customer_id = $1 AND (created_at, id) < ($2, $3)
			ORDER BY created_at DESC, id DESC
			LIMIT $4`, customerID, after.CreatedAt, after.ID, limit+1)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("orders page: %w", err)
	}
	page, err := pgx.CollectRows(rows, scanOrder)
	if err != nil {
		return nil, nil, fmt.Errorf("orders page: %w", err)
	}
	if len(page) <= limit {
		return page, nil, nil
	}
	page = page[:limit]
	last := page[limit-1]
	return page, &Cursor{CreatedAt: last.CreatedAt, ID: last.ID}, nil
}
