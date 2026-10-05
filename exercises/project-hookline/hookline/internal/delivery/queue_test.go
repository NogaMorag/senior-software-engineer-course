package delivery_test

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"example.com/hookline/internal/delivery"
	"example.com/hookline/internal/dispatch"
	"example.com/hookline/internal/domain"
	"example.com/hookline/internal/outbox"
	"example.com/hookline/internal/store/postgres"
	"example.com/hookline/internal/testdb"
	"example.com/hookline/internal/testenv"
)

func TestMain(m *testing.M) { os.Exit(testdb.Main(m)) }

// world is a tenant with endpoints in a migrated database.
type world struct {
	t      *testing.T
	pool   *pgxpool.Pool
	st     *postgres.Store
	tenant domain.Tenant
}

func newWorld(t *testing.T) *world {
	pool := testdb.Pool(t)
	st := postgres.New(pool)
	return &world{t: t, pool: pool, st: st, tenant: testenv.Tenant(t, st)}
}

func (w *world) endpoint(url string) domain.Endpoint {
	return testenv.Endpoint(w.t, w.st, w.tenant.ID, url, testenv.SinkSecret)
}

// deliveries ingests one event and fans it out to the endpoints, as the relay and dispatcher would.
func (w *world) deliveries(endpoints ...domain.Endpoint) domain.Event {
	w.t.Helper()
	ev := testenv.Event(w.t, w.st, w.tenant.ID, "invoice.paid", map[string]int{"amount": 4200})
	ids := make([]string, len(endpoints))
	for i, e := range endpoints {
		ids[i] = e.ID
	}
	if _, err := dispatch.Fanout(context.Background(), w.pool, outbox.Message{EventID: ev.ID, TenantID: w.tenant.ID, Type: ev.Type}, ids); err != nil {
		w.t.Fatal(err)
	}
	return ev
}

func (w *world) history(ev domain.Event) []domain.Delivery {
	w.t.Helper()
	ds, err := w.st.EventDeliveries(context.Background(), w.tenant.ID, ev.ID)
	if err != nil {
		w.t.Fatal(err)
	}
	return ds
}

// claimMine claims until it gets a delivery of this test's tenant: other tests share the database.
func (w *world) claimMine(lease time.Duration, maxAttempts int) *delivery.Job {
	w.t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		j, err := delivery.Claim(context.Background(), w.pool, lease, maxAttempts)
		if err != nil {
			w.t.Fatal(err)
		}
		if j != nil && j.TenantID == w.tenant.ID {
			return j
		}
		if j == nil {
			time.Sleep(20 * time.Millisecond)
		}
	}
	w.t.Fatal("no delivery of this test became claimable")
	return nil
}

func ok(code int) delivery.Result {
	return delivery.Result{StatusCode: code, Started: time.Now(), Duration: 12 * time.Millisecond}
}

func failed(code int) delivery.Result {
	r := ok(code)
	r.Retryable, r.Err = delivery.Classify(code)
	return r
}

func TestClaimCompleteRecordsTheAttempt(t *testing.T) {
	w := newWorld(t)
	ep := w.endpoint("https://receiver.example/hooks")
	ev := w.deliveries(ep)
	j := w.claimMine(time.Minute, 5)
	if j.Attempt != 1 || j.Round != 1 || j.URL != ep.URL || j.Secret != testenv.SinkSecret || j.EventID != ev.ID || !j.Enabled {
		t.Fatalf("claimed %+v", j)
	}
	if err := delivery.Complete(context.Background(), w.pool, j, ok(204)); err != nil {
		t.Fatal(err)
	}
	d := w.history(ev)[0]
	if d.Status != domain.Delivered || d.DeliveredAt == nil || len(d.History) != 1 || d.History[0].N != 1 || d.History[0].StatusCode != 204 {
		t.Fatalf("after Complete: %+v", d)
	}
}

func TestRetryThenDeadAfterTheLastAttempt(t *testing.T) {
	w := newWorld(t)
	ev := w.deliveries(w.endpoint("https://receiver.example/hooks"))
	for attempt, want := range []string{"pending", "dead"} {
		j := w.claimMine(time.Minute, 2)
		if j.Attempt != attempt+1 {
			t.Fatalf("attempt = %d, want %d", j.Attempt, attempt+1)
		}
		status, err := delivery.Fail(context.Background(), w.pool, j, failed(503), 2, 0)
		if err != nil || status != want {
			t.Fatalf("Fail #%d = %q, %v; want %s", attempt+1, status, err, want)
		}
	}
	d := w.history(ev)[0]
	if d.Status != domain.Dead || d.Attempts != 2 || d.LastError == "" || len(d.History) != 2 {
		t.Fatalf("after two 503s with max 2: %+v", d)
	}
}

func TestRetryWaitsForNextAttemptAt(t *testing.T) {
	w := newWorld(t)
	w.deliveries(w.endpoint("https://receiver.example/hooks"))
	j := w.claimMine(time.Minute, 5)
	if _, err := delivery.Fail(context.Background(), w.pool, j, failed(500), 5, time.Hour); err != nil {
		t.Fatal(err)
	}
	for range 3 {
		if j, err := delivery.Claim(context.Background(), w.pool, time.Minute, 5); err != nil || (j != nil && j.TenantID == w.tenant.ID) {
			t.Fatalf("claimed %+v, %v an hour before its next attempt", j, err)
		}
	}
}

func TestPermanentFailureIsDeadAtOnce(t *testing.T) {
	w := newWorld(t)
	ev := w.deliveries(w.endpoint("https://receiver.example/hooks"))
	j := w.claimMine(time.Minute, 10)
	if status, err := delivery.Fail(context.Background(), w.pool, j, failed(400), 10, time.Second); err != nil || status != "dead" {
		t.Fatalf("Fail(400) = %q, %v; want dead with 9 attempts left", status, err)
	}
	if d := w.history(ev)[0]; d.Status != domain.Dead || d.History[0].StatusCode != 400 {
		t.Fatalf("%+v", d)
	}
}

// TestExpiredLeaseIsReclaimedAndTheOldWorkerFenced: a worker that stalls past its lease loses the
// delivery; its late result is refused, and the lost attempt is in the history.
func TestExpiredLeaseIsReclaimedAndTheOldWorkerFenced(t *testing.T) {
	w := newWorld(t)
	ev := w.deliveries(w.endpoint("https://receiver.example/hooks"))
	stalled := w.claimMine(300*time.Millisecond, 5)
	time.Sleep(400 * time.Millisecond)
	fresh := w.claimMine(time.Minute, 5)
	if fresh.Attempt != 2 || fresh.DeliveryID != stalled.DeliveryID {
		t.Fatalf("reclaimed %+v, want attempt 2 of the same delivery", fresh)
	}
	if err := delivery.Complete(context.Background(), w.pool, stalled, ok(200)); !errors.Is(err, delivery.ErrLeaseLost) {
		t.Fatalf("late Complete = %v, want ErrLeaseLost", err)
	}
	if err := delivery.Complete(context.Background(), w.pool, fresh, ok(200)); err != nil {
		t.Fatal(err)
	}
	d := w.history(ev)[0]
	if len(d.History) != 2 || d.History[0].Error != "lease expired" || d.History[1].StatusCode != 200 {
		t.Fatalf("history = %+v", d.History)
	}
}

// TestReplayKeepsCountingAttempts: attempts.n stays unique and Hookline-Attempt keeps rising, while the
// new round gets the full budget (attempt_base; chapter 29, section 29.5).
func TestReplayKeepsCountingAttempts(t *testing.T) {
	w := newWorld(t)
	ev := w.deliveries(w.endpoint("https://receiver.example/hooks"))
	ctx := context.Background()
	for range 2 { // two attempts, max 2: dead
		j := w.claimMine(time.Minute, 2)
		if _, err := delivery.Fail(ctx, w.pool, j, failed(503), 2, 0); err != nil {
			t.Fatal(err)
		}
	}
	d := w.history(ev)[0]
	replayed, err := w.st.ReplayDelivery(ctx, w.tenant.ID, d.ID)
	if err != nil || replayed.Status != domain.Pending || replayed.Attempts != 2 || replayed.AttemptBase != 2 {
		t.Fatalf("replay = %+v, %v", replayed, err)
	}
	if _, err := w.st.ReplayDelivery(ctx, w.tenant.ID, d.ID); !errors.Is(err, domain.ErrNotDead) {
		t.Fatalf("second replay = %v, want ErrNotDead", err)
	}
	j := w.claimMine(time.Minute, 2)
	if j.Attempt != 3 || j.Round != 1 {
		t.Fatalf("after replay: attempt %d round %d, want 3 and 1", j.Attempt, j.Round)
	}
	if status, err := delivery.Fail(ctx, w.pool, j, failed(503), 2, 0); err != nil || status != "pending" {
		t.Fatalf("first failure of the new round = %q, %v; want pending (one attempt left)", status, err)
	}
	if d := w.history(ev)[0]; len(d.History) != 3 || d.History[2].N != 3 {
		t.Fatalf("history = %+v", d.History)
	}
}

func TestConcurrentClaimsNeverShareADelivery(t *testing.T) {
	w := newWorld(t)
	var eps []domain.Endpoint
	for range 30 {
		eps = append(eps, w.endpoint("https://receiver.example/hooks"))
	}
	w.deliveries(eps...)
	var mu sync.Mutex
	seen := map[string]int{}
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			for {
				j, err := delivery.Claim(context.Background(), w.pool, time.Minute, 5)
				if err != nil {
					t.Error(err)
					return
				}
				if j == nil {
					return
				}
				mu.Lock()
				seen[j.DeliveryID]++
				mu.Unlock()
			}
		})
	}
	wg.Wait()
	mine := 0
	for id, n := range seen {
		if n != 1 {
			t.Errorf("delivery %s claimed %d times", id, n)
		}
		mine++
	}
	if mine < 30 {
		t.Fatalf("claimed %d deliveries, want at least this test's 30", mine)
	}
}
