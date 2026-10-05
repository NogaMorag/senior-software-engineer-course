package postgres_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"example.com/hookline/internal/dispatch"
	"example.com/hookline/internal/domain"
	"example.com/hookline/internal/outbox"
	"example.com/hookline/internal/page"
	"example.com/hookline/internal/store/postgres"
	"example.com/hookline/internal/testdb"
	"example.com/hookline/internal/testenv"
)

// deadDeliveries makes n events with one dead delivery each, with two attempts in the history.
func deadDeliveries(t *testing.T, s *postgres.Store, tenantID string, n int) []domain.Event {
	t.Helper()
	pool := testdb.Pool(t)
	ep := testenv.Endpoint(t, s, tenantID, "https://receiver.example/hooks", testenv.SinkSecret)
	var evs []domain.Event
	for range n {
		ev := testenv.Event(t, s, tenantID, "invoice.paid", map[string]int{"amount": 1})
		if _, err := dispatch.Fanout(context.Background(), pool, outbox.Message{EventID: ev.ID, TenantID: tenantID, Type: ev.Type}, []string{ep.ID}); err != nil {
			t.Fatal(err)
		}
		if _, err := pool.Exec(context.Background(), `
			WITH d AS (UPDATE deliveries SET status = 'dead', attempts = 2, last_error = 'receiver answered 503'
			           WHERE event_id = $1 RETURNING id)
			INSERT INTO attempts (delivery_id, n, started_at, duration_ms, status_code, error)
			SELECT d.id, n, now(), 5, 503, 'receiver answered 503' FROM d, generate_series(1, 2) AS n`, ev.ID); err != nil {
			t.Fatal(err)
		}
		evs = append(evs, ev)
	}
	return evs
}

func TestEventDeliveriesCarryTheirHistory(t *testing.T) {
	s := postgres.New(testdb.Pool(t))
	tn := testenv.Tenant(t, s)
	other := testenv.Tenant(t, s)
	ev := deadDeliveries(t, s, tn.ID, 1)[0]
	ds, err := s.EventDeliveries(context.Background(), tn.ID, ev.ID)
	if err != nil || len(ds) != 1 {
		t.Fatalf("EventDeliveries = %+v, %v", ds, err)
	}
	if d := ds[0]; d.Status != domain.Dead || len(d.History) != 2 || d.History[1].N != 2 || d.History[1].StatusCode != 503 {
		t.Fatalf("delivery = %+v", d)
	}
	if _, err := s.EventDeliveries(context.Background(), other.ID, ev.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("another tenant's event = %v, want ErrNotFound", err)
	}
}

func TestDeadDeliveriesPage(t *testing.T) {
	s := postgres.New(testdb.Pool(t))
	tn := testenv.Tenant(t, s)
	deadDeliveries(t, s, tn.ID, 5)
	var seen int
	var after *page.Cursor
	for pages := 0; ; pages++ {
		items, more, err := s.DeadDeliveries(context.Background(), tn.ID, after, 2)
		if err != nil || pages > 5 {
			t.Fatalf("page %d: %v", pages, err)
		}
		seen += len(items)
		if !more {
			break
		}
		last := items[len(items)-1]
		after = &page.Cursor{CreatedAt: last.CreatedAt, ID: last.ID}
	}
	if seen != 5 {
		t.Fatalf("paged through %d dead deliveries, want 5", seen)
	}
}

// TestConcurrentReplaysMoveTheDeliveryOnce: two operators click "replay" at the same moment; one wins,
// the other gets ErrNotDead, and attempts is not reset (attempt_base takes the old count).
func TestConcurrentReplaysMoveTheDeliveryOnce(t *testing.T) {
	s := postgres.New(testdb.Pool(t))
	tn := testenv.Tenant(t, s)
	ev := deadDeliveries(t, s, tn.ID, 1)[0]
	ds, _ := s.EventDeliveries(context.Background(), tn.ID, ev.ID)
	id := ds[0].ID
	var wg sync.WaitGroup
	errs := make([]error, 8)
	for i := range errs {
		wg.Go(func() { _, errs[i] = s.ReplayDelivery(context.Background(), tn.ID, id) })
	}
	wg.Wait()
	wins := 0
	for _, err := range errs {
		switch {
		case err == nil:
			wins++
		case !errors.Is(err, domain.ErrNotDead):
			t.Fatalf("replay: %v", err)
		}
	}
	if wins != 1 {
		t.Fatalf("%d replays succeeded, want exactly 1", wins)
	}
	d, _ := s.EventDeliveries(context.Background(), tn.ID, ev.ID)
	if d[0].Status != domain.Pending || d[0].Attempts != 2 || d[0].AttemptBase != 2 || d[0].LastError != "" ||
		d[0].NextAttemptAt.After(time.Now().Add(time.Second)) {
		t.Fatalf("after replay: %+v", d[0])
	}
	if _, err := s.ReplayDelivery(context.Background(), tn.ID, domain.NewID(time.Now())); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("unknown id = %v, want ErrNotFound", err)
	}
}
