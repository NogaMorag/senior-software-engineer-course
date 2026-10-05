package worker_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"example.com/ch14-jobqueue/backoff"
	"example.com/ch14-jobqueue/queue"
	"example.com/ch14-jobqueue/testenv"
	"example.com/ch14-jobqueue/worker"
)

func newPool(db *pgxpool.Pool, h worker.Handler) *worker.Pool {
	return &worker.Pool{
		DB:       db,
		Handlers: map[string]worker.Handler{"test": h},
		Name:     "w",
		Workers:  2,
		Lease:    10 * time.Second,
		Poll:     20 * time.Millisecond,
		Backoff:  backoff.Policy{Base: time.Millisecond, Cap: 5 * time.Millisecond},
		Log:      slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
}

// waitState polls until job id reaches want, or fails after a few seconds.
func waitState(t *testing.T, db *pgxpool.Pool, id int64, want queue.State) queue.Job {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for {
		j, err := queue.Get(context.Background(), db, id)
		if err != nil {
			t.Fatal(err)
		}
		if j.State == want {
			return j
		}
		if time.Now().After(deadline) {
			t.Fatalf("job %d is %s after 10s, want %s", id, j.State, want)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func run(t *testing.T, p *worker.Pool) (stop func()) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		_ = p.Run(ctx)
		close(done)
	}()
	return func() {
		cancel()
		<-done
	}
}

func TestRetriesUntilSuccess(t *testing.T) {
	db := testenv.Postgres(t)
	var calls atomic.Int32
	p := newPool(db, func(_ context.Context, j queue.Job) error {
		calls.Add(1)
		if j.Attempts < 3 {
			return errors.New("flaky")
		}
		return nil
	})
	id, _, err := queue.Enqueue(context.Background(), db, "test", nil, queue.Options{MaxAttempts: 5})
	if err != nil {
		t.Fatal(err)
	}
	stop := run(t, p)
	defer stop()
	j := waitState(t, db, id, queue.Done)
	if j.Attempts != 3 || calls.Load() != 3 {
		t.Errorf("attempts = %d, calls = %d; want 3 and 3", j.Attempts, calls.Load())
	}
}

func TestAlwaysFailingJobGoesDead(t *testing.T) {
	db := testenv.Postgres(t)
	p := newPool(db, func(context.Context, queue.Job) error { return errors.New("down") })
	id, _, err := queue.Enqueue(context.Background(), db, "test", nil, queue.Options{MaxAttempts: 4})
	if err != nil {
		t.Fatal(err)
	}
	stop := run(t, p)
	defer stop()
	j := waitState(t, db, id, queue.Dead)
	if j.Attempts != 4 || j.LastError != "down" {
		t.Errorf("dead job = %+v, want 4 attempts and last_error \"down\"", j)
	}
}

func TestPanicAndUnknownKindDoNotKillTheWorker(t *testing.T) {
	db := testenv.Postgres(t)
	p := newPool(db, func(context.Context, queue.Job) error { panic("boom") })
	ctx := context.Background()
	panics, _, err := queue.Enqueue(ctx, db, "test", nil, queue.Options{MaxAttempts: 1})
	if err != nil {
		t.Fatal(err)
	}
	unknown, _, err := queue.Enqueue(ctx, db, "no-such-kind", nil, queue.Options{MaxAttempts: 5})
	if err != nil {
		t.Fatal(err)
	}
	stop := run(t, p)
	defer stop()
	if j := waitState(t, db, panics, queue.Dead); j.LastError != "handler panic: boom" {
		t.Errorf("panicking job last_error = %q", j.LastError)
	}
	if j := waitState(t, db, unknown, queue.Dead); j.Attempts != 1 {
		t.Errorf("unknown kind used %d attempts, want 1 (permanent)", j.Attempts)
	}
}

// Shutdown starts while a job is running: Run must wait for it, record it as done, and
// claim nothing new.
func TestGracefulDrainWaitsForTheClaimedJob(t *testing.T) {
	db := testenv.Postgres(t)
	started := make(chan int64, 1)
	release := make(chan struct{})
	p := newPool(db, func(ctx context.Context, j queue.Job) error {
		started <- j.ID
		select {
		case <-release:
			return nil
		case <-ctx.Done(): // must NOT fire on shutdown; only at the lease deadline
			return ctx.Err()
		}
	})
	p.Workers = 1
	bg := context.Background()
	first, _, err := queue.Enqueue(bg, db, "test", nil, queue.Options{})
	if err != nil {
		t.Fatal(err)
	}
	second, _, err := queue.Enqueue(bg, db, "test", nil, queue.Options{})
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(bg)
	returned := make(chan struct{})
	go func() {
		_ = p.Run(ctx)
		close(returned)
	}()
	select {
	case got := <-started:
		if got != first {
			t.Fatalf("started job %d, want %d", got, first)
		}
	case <-time.After(10 * time.Second):
		cancel()
		t.Fatal("no job started within 10s")
	}
	cancel() // SIGTERM arrives while the job runs
	select {
	case <-returned:
		t.Fatal("Run returned while a claimed job was still running")
	case <-time.After(200 * time.Millisecond):
	}
	close(release)
	select {
	case <-returned:
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return after the job finished")
	}
	if j, _ := queue.Get(bg, db, first); j.State != queue.Done {
		t.Errorf("drained job state = %s, want done (its result was not recorded)", j.State)
	}
	if j, _ := queue.Get(bg, db, second); j.State != queue.Pending {
		t.Errorf("second job state = %s, want pending (claimed during shutdown)", j.State)
	}
}
