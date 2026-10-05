package outbox_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"example.com/hookline/internal/outbox"
	"example.com/hookline/internal/store/postgres"
	"example.com/hookline/internal/testdb"
	"example.com/hookline/internal/testenv"
)

func TestMain(m *testing.M) { os.Exit(testdb.Main(m)) }

// fakePublisher records every message it was asked to publish, like a broker would.
type fakePublisher struct {
	mu   sync.Mutex
	sent []outbox.Message
	fail error
}

func (p *fakePublisher) Publish(_ context.Context, msgs []outbox.Message) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.fail != nil {
		return p.fail
	}
	p.sent = append(p.sent, msgs...)
	return nil
}

func (p *fakePublisher) events() []string {
	p.mu.Lock()
	defer p.mu.Unlock()
	var ids []string
	for _, m := range p.sent {
		ids = append(ids, m.EventID)
	}
	return ids
}

// drain runs the relay until the outbox is empty. The outbox is shared by every test in the package,
// and the tests run one after another, so after drain nothing of this test is pending.
func drain(t *testing.T, r *outbox.Relay) int {
	t.Helper()
	total := 0
	for {
		n, err := r.RunOnce(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if n == 0 {
			return total
		}
		total += n
	}
}

func ingest(t *testing.T, pool *pgxpool.Pool, n int) []string {
	t.Helper()
	st := postgres.New(pool)
	tn := testenv.Tenant(t, st)
	var ids []string
	for i := range n {
		ids = append(ids, testenv.Event(t, st, tn.ID, "invoice.paid", map[string]int{"n": i}).ID)
	}
	return ids
}

func TestMessageRoundTrip(t *testing.T) {
	m := outbox.Message{Seq: 7, EventID: "e", TenantID: "t", Type: "invoice.paid"}
	got, err := outbox.Decode(m.Value())
	if err != nil || got.EventID != "e" || got.TenantID != "t" || got.Type != "invoice.paid" {
		t.Fatalf("Decode(Value) = %+v, %v", got, err)
	}
	for _, bad := range []string{`not json`, `{"event_id":"e"}`, `{}`} {
		if _, err := outbox.Decode([]byte(bad)); err == nil {
			t.Errorf("Decode(%s) succeeded", bad)
		}
	}
}

func TestRelayPublishesInOrderAndMarks(t *testing.T) {
	pool := testdb.Pool(t)
	want := ingest(t, pool, 7)
	pub := &fakePublisher{}
	drain(t, &outbox.Relay{DB: pool, Publisher: pub, BatchSize: 3})
	got := pub.events()
	// Events of earlier tests may be published too; ours must all be there, in ingest order.
	var mine []string
	for _, id := range got {
		if slices.Contains(want, id) {
			mine = append(mine, id)
		}
	}
	if !slices.Equal(mine, want) {
		t.Fatalf("published %v\nwant      %v (outbox order)", mine, want)
	}
	if p, err := outbox.Pending(context.Background(), pool); err != nil || p != 0 {
		t.Fatalf("pending = %d, %v; want 0", p, err)
	}
}

func TestPublishFailureMarksNothing(t *testing.T) {
	pool := testdb.Pool(t)
	drain(t, &outbox.Relay{DB: pool, Publisher: &fakePublisher{}}) // start empty
	ingest(t, pool, 2)
	broker := errors.New("broker unavailable")
	if _, err := (&outbox.Relay{DB: pool, Publisher: &fakePublisher{fail: broker}}).RunOnce(context.Background()); !errors.Is(err, broker) {
		t.Fatalf("RunOnce = %v, want %v", err, broker)
	}
	if p, _ := outbox.Pending(context.Background(), pool); p != 2 {
		t.Fatalf("pending = %d, want 2: a row was marked without being published", p)
	}
	drain(t, &outbox.Relay{DB: pool, Publisher: &fakePublisher{}})
}

// TestCrashBetweenPublishAndMarkPublishesAgain is the crash point of chapter 37, section 37.6: the broker
// has the messages, the mark never commits, the next relay publishes them again. The dispatcher's
// unique (event_id, endpoint_id) makes that harmless (dispatch tests).
func TestCrashBetweenPublishAndMarkPublishesAgain(t *testing.T) {
	pool := testdb.Pool(t)
	drain(t, &outbox.Relay{DB: pool, Publisher: &fakePublisher{}})
	ids := ingest(t, pool, 3)
	pub := &fakePublisher{}
	crash := errors.New("relay killed")
	crashing := &outbox.Relay{DB: pool, Publisher: pub, AfterPublish: func(context.Context, []outbox.Message) error { return crash }}
	if _, err := crashing.RunOnce(context.Background()); !errors.Is(err, crash) {
		t.Fatalf("crashing relay = %v", err)
	}
	if p, _ := outbox.Pending(context.Background(), pool); p != 3 {
		t.Fatalf("after the crash pending = %d, want 3", p)
	}
	drain(t, &outbox.Relay{DB: pool, Publisher: pub})
	if got := pub.events(); !slices.Equal(got, append(slices.Clone(ids), ids...)) {
		t.Fatalf("published %v, want every event twice: once before the crash, once after", got)
	}
}

func TestRunStopsOnCancelAndRetriesOnError(t *testing.T) {
	pool := testdb.Pool(t)
	drain(t, &outbox.Relay{DB: pool, Publisher: &fakePublisher{}})
	ids := ingest(t, pool, 2)
	pub := &fakePublisher{fail: errors.New("down")}
	relay := &outbox.Relay{DB: pool, Publisher: pub}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- relay.Run(ctx, 10*time.Millisecond, slog.New(slog.NewTextHandler(io.Discard, nil))) }()
	time.Sleep(50 * time.Millisecond)
	pub.mu.Lock()
	pub.fail = nil // the broker is back
	pub.mu.Unlock()
	deadline := time.Now().Add(5 * time.Second)
	for len(pub.events()) < len(ids) && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	cancel()
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(pub.events(), ids) {
		t.Fatalf("after the broker came back: published %v, want %v", pub.events(), ids)
	}
}

// TestStaleRelayCannotMark is the fencing of chapter 18: a relay whose term is no longer the newest
// may still publish (a duplicate), but its mark is refused, so it cannot mark rows the new leader owns.
func TestStaleRelayCannotMark(t *testing.T) {
	pool := testdb.Pool(t)
	ctx := context.Background()
	drain(t, &outbox.Relay{DB: pool, Publisher: &fakePublisher{}})
	if _, err := pool.Exec(ctx, `
		INSERT INTO leader_terms (lock_key, term) VALUES ($1, 7)
		ON CONFLICT (lock_key) DO UPDATE SET term = 7`, outbox.RelayLockKey); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM leader_terms WHERE lock_key = $1`, outbox.RelayLockKey)
	})
	ingest(t, pool, 2)

	stale := &outbox.Relay{DB: pool, Publisher: &fakePublisher{}, Term: 6}
	if _, err := stale.RunOnce(ctx); !errors.Is(err, outbox.ErrNotLeader) {
		t.Fatalf("stale relay RunOnce = %v, want ErrNotLeader", err)
	}
	if err := stale.Run(ctx, 10*time.Millisecond, slog.New(slog.NewTextHandler(io.Discard, nil))); !errors.Is(err, outbox.ErrNotLeader) {
		t.Fatalf("stale relay Run = %v, want it to step down with ErrNotLeader", err)
	}
	if p, _ := outbox.Pending(ctx, pool); p != 2 {
		t.Fatalf("pending = %d after the stale relay, want 2", p)
	}
	if n := drain(t, &outbox.Relay{DB: pool, Publisher: &fakePublisher{}, Term: 7}); n != 2 {
		t.Fatalf("the current leader marked %d, want 2", n)
	}
}
