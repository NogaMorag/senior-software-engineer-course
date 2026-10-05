package orders_test

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"example.com/ch26-testing/internal/pgtest"
	orders "example.com/ch26-testing/orders"
)

// One PostgreSQL container for every test in this package; each test gets its own database.
func TestMain(m *testing.M) { os.Exit(pgtest.Main(m)) }

// setup returns a repository over a fresh database with the schema applied, the raw pool
// (for fixtures), and the query counter installed on the repository's pool.
func setup(t *testing.T) (*orders.Repo, *pgxpool.Pool, *orders.QueryCounter) {
	t.Helper()
	t.Parallel() // safe: no two tests share a database
	cfg, err := orders.Config(pgtest.NewDatabase(t))
	if err != nil {
		t.Fatal(err)
	}
	counter := &orders.QueryCounter{}
	cfg.ConnConfig.Tracer = counter
	pool, err := orders.Connect(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	if _, err := pool.Exec(context.Background(), orders.Schema); err != nil {
		t.Fatalf("apply schema: %v", err)
	}
	return orders.New(pool), pool, counter
}

func ids(list []orders.Order) []int64 {
	out := make([]int64, len(list))
	for i, o := range list {
		out[i] = o.ID
	}
	return out
}

func TestRecentOrders(t *testing.T) {
	repo, pool, _ := setup(t)
	standardFixture(t, pool)
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

// The builder makes a fixture that states only what this test is about: whose order it is.
func TestRecentOrdersOnlyThatCustomer(t *testing.T) {
	repo, pool, _ := setup(t)
	seedCatalog(t, pool)
	anOrder(1).insert(t, pool)
	anOrder(2).forCustomer(2).insert(t, pool)
	anOrder(3).insert(t, pool)

	got, err := repo.RecentOrders(context.Background(), 2, 10)
	if err != nil {
		t.Fatal(err)
	}
	if want := []int64{2}; !reflect.DeepEqual(ids(got), want) {
		t.Fatalf("ids = %v, want %v", ids(got), want)
	}
}

func TestOrdersWithItemsUsesTwoQueries(t *testing.T) {
	repo, pool, counter := setup(t)
	standardFixture(t, pool)
	ctx := context.Background()

	counter.Reset()
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
		if want := itemsOf(o.ID); len(o.Items) != want {
			t.Errorf("order %d has %d items, want %d", o.ID, len(o.Items), want)
		}
	}
}

func TestOrdersWithItemsNoOrders(t *testing.T) {
	repo, pool, counter := setup(t)
	standardFixture(t, pool)
	counter.Reset()
	got, err := repo.OrdersWithItems(context.Background(), 2, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 || counter.Count() != 1 {
		t.Fatalf("got %d orders with %d queries, want 0 orders with 1 query", len(got), counter.Count())
	}
}

func TestOrdersPageWalksEverythingOnce(t *testing.T) {
	repo, pool, _ := setup(t)
	standardFixture(t, pool)
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

// Only a real server enforces the schema's constraints; a fake repository would accept this.
func TestSchemaRejectsUnknownStatus(t *testing.T) {
	_, pool, _ := setup(t)
	seedCatalog(t, pool)
	_, err := pool.Exec(context.Background(), `
		INSERT INTO orders (customer_id, status, total_cents, created_at)
		VALUES (1, 'lost', 0, now())`)
	if err == nil {
		t.Fatal("status 'lost' accepted, want a CHECK violation")
	}
}
