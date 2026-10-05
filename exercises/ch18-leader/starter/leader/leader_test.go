package leader

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"example.com/ch18-leader/fencing"
	"example.com/ch18-leader/internal/testdb"
)

func TestMain(m *testing.M) { os.Exit(testdb.Main(m)) }

// event is what a lead function reports: who leads, in which term, starting or stopping.
type event struct {
	who     string
	term    int64
	started bool
}

// leadUntilCancelled reports its start, works until its context ends, reports its stop.
func leadUntilCancelled(who string, events chan<- event) func(context.Context, int64) error {
	return func(ctx context.Context, term int64) error {
		events <- event{who, term, true}
		<-ctx.Done()
		events <- event{who, term, false}
		return ctx.Err()
	}
}

// testLogger sends the elector's log lines to the test log (visible with -v).
func testLogger(t *testing.T, who string) *slog.Logger {
	return slog.New(slog.NewTextHandler(logWriter{t}, nil)).With("who", who)
}

type logWriter struct{ t *testing.T }

func (w logWriter) Write(p []byte) (int, error) {
	w.t.Log(strings.TrimSuffix(string(p), "\n"))
	return len(p), nil
}

// start runs an Elector with its own pool, as a separate process would. The returned stop
// function cancels it and returns Run's error; cleanup stops it too.
func start(t *testing.T, who string, cfg Config, lead func(context.Context, int64) error) (stop func() error) {
	t.Helper()
	pool := testdb.Pool(t)
	cfg.Logger = testLogger(t, who)
	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan error, 1)
	go func() { result <- New(pool, cfg).Run(ctx, lead) }()
	var err error
	stopped := false
	stop = func() error {
		if !stopped {
			cancel()
			err, stopped = <-result, true
		}
		return err
	}
	t.Cleanup(func() { _ = stop() })
	return stop
}

// setup creates the tables and forgets earlier terms of key, so that every run (also with
// -count=N) starts at term 1.
func setup(t *testing.T, key int64) *pgxpool.Pool {
	t.Helper()
	admin := testdb.Pool(t)
	ctx := context.Background()
	if err := EnsureSchema(ctx, admin); err != nil {
		t.Fatal(err)
	}
	if _, err := admin.Exec(ctx, fencing.Schema); err != nil {
		t.Fatal(err)
	}
	if _, err := admin.Exec(ctx, `DELETE FROM leader_terms WHERE lock_key = $1`, key); err != nil {
		t.Fatal(err)
	}
	if _, err := admin.Exec(ctx, `DELETE FROM fenced_resources`); err != nil {
		t.Fatal(err)
	}
	return admin
}

// next waits up to d for the next event.
func next(t *testing.T, events <-chan event, d time.Duration) event {
	t.Helper()
	select {
	case e := <-events:
		return e
	case <-time.After(d):
		t.Fatalf("no leadership event within %v", d)
		return event{}
	}
}

func terminateHolder(t *testing.T, admin *pgxpool.Pool, key int64) {
	t.Helper()
	ctx := context.Background()
	pid, ok, err := HolderPID(ctx, admin, key)
	if err != nil || !ok {
		t.Fatalf("HolderPID = %d, %v, %v; want the leader's backend", pid, ok, err)
	}
	var terminated bool
	// The timeout (ms) makes pg_terminate_backend wait until the backend has really exited.
	if err := admin.QueryRow(ctx, `SELECT pg_terminate_backend($1, 5000)`, pid).Scan(&terminated); err != nil || !terminated {
		t.Fatalf("pg_terminate_backend(%d) = %v, %v", pid, terminated, err)
	}
	t.Logf("terminated the leader's backend, pid %d", pid)
}

func TestFollowerTakesOverWhenTheLeaderSessionIsKilled(t *testing.T) {
	const key = 18001
	admin := setup(t, key)
	cfg := Config{Key: key, Heartbeat: 100 * time.Millisecond, Retry: 50 * time.Millisecond}
	events := make(chan event, 10)

	start(t, "A", cfg, leadUntilCancelled("A", events))
	if e := next(t, events, 5*time.Second); e != (event{"A", 1, true}) {
		t.Fatalf("first event %+v, want A leading term 1", e)
	}
	start(t, "B", cfg, leadUntilCancelled("B", events))
	select {
	case e := <-events:
		t.Fatalf("event %+v while A leads; want B to stay a follower", e)
	case <-time.After(500 * time.Millisecond): // B tried about ten times and lost each time
	}

	killed := time.Now()
	terminateHolder(t, admin, key)
	// Two things happen, in either order: B wins term 2, and A's heartbeat notices.
	got := map[event]bool{}
	for range 2 {
		got[next(t, events, 5*time.Second)] = true
	}
	if !got[event{"A", 1, false}] || !got[event{"B", 2, true}] {
		t.Fatalf("events %v, want A stopping term 1 and B leading term 2", got)
	}
	t.Logf("takeover completed %v after the kill", time.Since(killed).Round(time.Millisecond))
}

func TestCleanStepDownHandsOverQuickly(t *testing.T) {
	const key = 18002
	setup(t, key)
	cfg := Config{Key: key, Heartbeat: 100 * time.Millisecond, Retry: 50 * time.Millisecond}
	events := make(chan event, 10)

	stopA := start(t, "A", cfg, leadUntilCancelled("A", events))
	if e := next(t, events, 5*time.Second); e != (event{"A", 1, true}) {
		t.Fatalf("first event %+v, want A leading term 1", e)
	}
	start(t, "B", cfg, leadUntilCancelled("B", events))
	if err := stopA(); !errors.Is(err, context.Canceled) {
		t.Fatalf("A's Run returned %v, want context.Canceled", err)
	}
	if e := next(t, events, time.Second); e != (event{"A", 1, false}) {
		t.Fatalf("event %+v, want A stopping", e)
	}
	if e := next(t, events, 5*time.Second); e != (event{"B", 2, true}) {
		t.Fatalf("event %+v, want B leading term 2", e)
	}
}

// A leader that is paused (here: blocked on a channel, in life: a GC pause or a stopped VM)
// loses its session, a new leader is elected, and then the old one wakes up and writes.
// The lock cannot stop that write; the fencing token can.
func TestPausedLeaderIsFenced(t *testing.T) {
	const key = 18003
	admin := setup(t, key)
	store := fencing.Store{DB: admin}
	events := make(chan event, 10)
	resume := make(chan struct{})
	lateWrite := make(chan error, 1)

	// A heartbeats rarely, so it has no idea that its session is gone.
	start(t, "A", Config{Key: key, Heartbeat: time.Hour, Retry: 50 * time.Millisecond},
		func(ctx context.Context, term int64) error {
			events <- event{"A", term, true}
			select { // the pause
			case <-resume:
			case <-ctx.Done(): // the test gave up early; do not hang its cleanup
				return ctx.Err()
			}
			lateWrite <- store.Write(context.Background(), "relay-cursor", "written by A", term)
			<-ctx.Done()
			return ctx.Err()
		})
	if e := next(t, events, 5*time.Second); e != (event{"A", 1, true}) {
		t.Fatalf("first event %+v, want A leading term 1", e)
	}
	start(t, "B", Config{Key: key, Heartbeat: 100 * time.Millisecond, Retry: 50 * time.Millisecond},
		func(ctx context.Context, term int64) error {
			events <- event{"B", term, true}
			if err := store.Write(ctx, "relay-cursor", "written by B", term); err != nil {
				t.Errorf("B's write in term %d: %v", term, err)
			}
			<-ctx.Done()
			return ctx.Err()
		})

	terminateHolder(t, admin, key)
	if e := next(t, events, 5*time.Second); e != (event{"B", 2, true}) {
		t.Fatalf("event %+v, want B leading term 2", e)
	}
	time.Sleep(100 * time.Millisecond) // let B's write land
	close(resume)                      // A wakes up, still sure it leads
	if err := <-lateWrite; !errors.Is(err, fencing.ErrStaleToken) {
		t.Fatalf("A's late write: err = %v, want fencing.ErrStaleToken", err)
	}
	value, token, err := store.Read(context.Background(), "relay-cursor")
	if err != nil || value != "written by B" || token != 2 {
		t.Fatalf("resource = %q (token %d), %v; want B's write", value, token, err)
	}
}

func TestHolderPIDWhenNobodyLeads(t *testing.T) {
	admin := setup(t, 18999)
	pid, ok, err := HolderPID(context.Background(), admin, 18999)
	if err != nil || ok {
		t.Fatalf("HolderPID = %d, %v, %v; want nobody", pid, ok, err)
	}
}
