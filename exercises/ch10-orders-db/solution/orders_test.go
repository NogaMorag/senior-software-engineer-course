package orders_test

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	orders "example.com/ch10-orders-db"
)

// setup creates a private schema with a small, known dataset and returns a repository
// over it plus the query counter installed on its pool. The schema is dropped at the end.
func setup(t *testing.T) (*orders.Repo, *orders.QueryCounter) {
	t.Helper()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("integration test: set DATABASE_URL to the chapter 10 database; see README")
	}
	ctx := context.Background()
	schema := fmt.Sprintf("ch10_test_%d_%d", os.Getpid(), time.Now().UnixNano())

	admin, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(admin.Close)
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := admin.Exec(context.Background(), "DROP SCHEMA "+schema+" CASCADE"); err != nil {
			t.Errorf("drop schema: %v", err)
		}
	})

	cfg, err := orders.Config(url)
	if err != nil {
		t.Fatal(err)
	}
	cfg.ConnConfig.RuntimeParams["search_path"] = schema
	counter := &orders.QueryCounter{}
	cfg.ConnConfig.Tracer = counter
	pool, err := orders.Connect(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)

	if _, err := pool.Exec(ctx, orders.Schema); err != nil {
		t.Fatalf("apply schema: %v", err)
	}
	// Customer 1: 25 orders, one per hour, except that orders 11 and 12 share a timestamp
	// (the keyset tie-breaker case). Order n has 1 + n%3 items. Customer 2: no orders.
	if _, err := pool.Exec(ctx, `
		INSERT INTO customers (id, email, name, country, created_at) VALUES
			(1, 'a@example.com', 'A', 'IL', '2026-01-01T00:00:00Z'),
			(2, 'b@example.com', 'B', 'IL', '2026-01-01T00:00:00Z');
		INSERT INTO products (id, name, price_cents)
			SELECT p, 'P' || p, 100 * p FROM generate_series(1, 5) AS p;
		INSERT INTO orders (id, customer_id, status, total_cents, created_at)
			SELECT n, 1, 'delivered', 0,
			       TIMESTAMPTZ '2026-09-01T00:00:00Z' + make_interval(hours => CASE WHEN n = 12 THEN 11 ELSE n END)
			FROM generate_series(1, 25) AS n;
		INSERT INTO order_items (order_id, product_id, quantity, price_cents)
			SELECT n, k, k, 100 * k FROM generate_series(1, 25) AS n, generate_series(1, 3) AS k
			WHERE k <= 1 + n % 3;`); err != nil {
		t.Fatalf("insert fixture: %v", err)
	}
	counter.Reset()
	return orders.New(pool), counter
}

func ids(list []orders.Order) []int64 {
	out := make([]int64, len(list))
	for i, o := range list {
		out[i] = o.ID
	}
	return out
}

func TestRecentOrders(t *testing.T) {
	repo, _ := setup(t)
	tests := []struct {
		name     string
		customer int64
		limit    int
		want     []int64
	}{
		{"newest first", 1, 4, []int64{25, 24, 23, 22}},
		{"same timestamp breaks ties by id", 1, 16, []int64{25, 24, 23, 22, 21, 20, 19, 18, 17, 16, 15, 14, 13, 12, 11, 10}},
		{"customer without orders", 2, 10, []int64{}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := repo.RecentOrders(context.Background(), tc.customer, tc.limit)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(ids(got), tc.want) {
				t.Fatalf("ids = %v, want %v", ids(got), tc.want)
			}
		})
	}
}

func TestOrdersWithItemsUsesTwoQueries(t *testing.T) {
	repo, counter := setup(t)
	ctx := context.Background()

	slow, err := repo.OrdersWithItemsNPlusOne(ctx, 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	if got := counter.Count(); got != 11 {
		t.Fatalf("N+1 version sent %d queries, want 1 + 10", got)
	}

	counter.Reset()
	fast, err := repo.OrdersWithItems(ctx, 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	if got := counter.Count(); got != 2 {
		t.Fatalf("OrdersWithItems sent %d queries, want 2", got)
	}
	if !reflect.DeepEqual(fast, slow) {
		t.Fatalf("results differ:\n fast %+v\n slow %+v", fast, slow)
	}
	for _, o := range fast {
		if want := 1 + int(o.ID%3); len(o.Items) != want {
			t.Errorf("order %d has %d items, want %d", o.ID, len(o.Items), want)
		}
	}
}

func TestOrdersWithItemsNoOrders(t *testing.T) {
	repo, counter := setup(t)
	got, err := repo.OrdersWithItems(context.Background(), 2, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 || counter.Count() != 1 {
		t.Fatalf("got %d orders with %d queries, want 0 orders with 1 query", len(got), counter.Count())
	}
}

func TestOrdersPageWalksEverythingOnce(t *testing.T) {
	repo, _ := setup(t)
	ctx := context.Background()
	for _, size := range []int{1, 7, 25, 100} {
		t.Run(fmt.Sprintf("page size %d", size), func(t *testing.T) {
			var all []int64
			var after *orders.Cursor
			pages := 0
			for {
				page, next, err := repo.OrdersPage(ctx, 1, after, size)
				if err != nil {
					t.Fatal(err)
				}
				pages++
				if len(page) > size {
					t.Fatalf("page has %d rows, limit %d", len(page), size)
				}
				all = append(all, ids(page)...)
				if next == nil {
					break
				}
				after = next
			}
			want, err := repo.RecentOrders(ctx, 1, 1000)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(all, ids(want)) {
				t.Fatalf("walked %v, want %v", all, ids(want))
			}
			if wantPages := (25 + size - 1) / size; pages != wantPages {
				t.Fatalf("took %d pages, want %d", pages, wantPages)
			}
		})
	}
}

func TestOrdersPageRejectsBadLimit(t *testing.T) {
	repo, _ := setup(t)
	if _, _, err := repo.OrdersPage(context.Background(), 1, nil, 0); err == nil {
		t.Fatal("limit 0 accepted, want an error")
	}
}
