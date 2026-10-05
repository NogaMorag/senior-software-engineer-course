package postgres_test

import (
	"context"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"example.com/hookline/internal/domain"
	"example.com/hookline/internal/store"
	"example.com/hookline/internal/store/postgres"
	"example.com/hookline/internal/store/storetest"
	"example.com/hookline/internal/testdb"
)

func TestMain(m *testing.M) { os.Exit(testdb.Main(m)) }

// The same suite the memory store passes, now against PostgreSQL.
func TestConformance(t *testing.T) {
	storetest.Run(t, func(t *testing.T) store.Store { return postgres.New(testdb.Pool(t)) })
}

func TestActiveSubscriptions(t *testing.T) {
	storetest.ActiveSubscriptions(t, postgres.New(testdb.Pool(t)))
}

func count(t *testing.T, query string, args ...any) int {
	t.Helper()
	var n int
	if err := testdb.Pool(t).QueryRow(context.Background(), query, args...).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

// F3 at the database: many concurrent requests with one key leave exactly one event row and exactly
// one outbox row, and every request gets the same event back.
func TestConcurrentIngestLeavesOneEventAndOneOutboxRow(t *testing.T) {
	pool := testdb.Pool(t)
	s := postgres.New(pool)
	tn, _ := storetest.NewTenant(t, s, "Acme")
	const n = 24
	ids := make([]string, n)
	errs := make([]error, n)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := range n {
		wg.Go(func() {
			<-start
			e := storetest.NewEvent(tn.ID, "order-42", "invoice.paid", `{"amount": 4200}`, domain.Now(time.Now))
			got, _, err := s.IngestEvent(context.Background(), e)
			ids[i], errs[i] = got.ID, err
		})
	}
	close(start)
	wg.Wait()
	for i := range n {
		if errs[i] != nil {
			t.Fatalf("request %d: %v", i, errs[i])
		}
		if ids[i] != ids[0] {
			t.Fatalf("request %d got event %s, request 0 got %s", i, ids[i], ids[0])
		}
	}
	if got := count(t, `SELECT count(*) FROM events WHERE tenant_id = $1`, tn.ID); got != 1 {
		t.Fatalf("events rows = %d, want 1", got)
	}
	if got := count(t, `SELECT count(*) FROM outbox o JOIN events e ON e.id = o.event_id WHERE e.tenant_id = $1 AND o.published_at IS NULL`, tn.ID); got != 1 {
		t.Fatalf("outbox rows = %d, want 1 (written in the event's transaction, never for a replay)", got)
	}
}

// A rejected request (another body under a used key) writes nothing.
func TestMismatchWritesNothing(t *testing.T) {
	s := postgres.New(testdb.Pool(t))
	tn, _ := storetest.NewTenant(t, s, "Acme")
	now := domain.Now(time.Now)
	if _, _, err := s.IngestEvent(context.Background(), storetest.NewEvent(tn.ID, "k", "a.b", `{"v": 1}`, now)); err != nil {
		t.Fatal(err)
	}
	_, _, err := s.IngestEvent(context.Background(), storetest.NewEvent(tn.ID, "k", "a.b", `{"v": 2}`, now))
	if err != domain.ErrIdempotencyMismatch {
		t.Fatalf("err = %v, want ErrIdempotencyMismatch", err)
	}
	if got := count(t, `SELECT count(*) FROM outbox o JOIN events e ON e.id = o.event_id WHERE e.tenant_id = $1`, tn.ID); got != 1 {
		t.Fatalf("outbox rows = %d, want 1", got)
	}
}

// The keyset query must use the pagination index, not sort the whole table.
func TestListEndpointsUsesTheIndex(t *testing.T) {
	pool := testdb.Pool(t)
	var plan string
	rows, err := pool.Query(context.Background(), `EXPLAIN
		SELECT id FROM endpoints
		WHERE tenant_id = '00000000-0000-0000-0000-000000000001'
		  AND (created_at, id) < (now(), '00000000-0000-0000-0000-000000000001')
		ORDER BY created_at DESC, id DESC LIMIT 21`)
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		var line string
		if err := rows.Scan(&line); err != nil {
			t.Fatal(err)
		}
		plan += line + "\n"
	}
	if rows.Err() != nil {
		t.Fatal(rows.Err())
	}
	if !strings.Contains(plan, "endpoints_tenant_page_idx") {
		t.Fatalf("plan does not use endpoints_tenant_page_idx:\n%s", plan)
	}
}
