package queue_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"example.com/ch11-isolation-lab/internal/testdb"
	"example.com/ch11-isolation-lab/queue"
)

func TestMain(m *testing.M) { os.Exit(testdb.Main(m)) }

func reset(t *testing.T, pool *pgxpool.Pool, n int) {
	t.Helper()
	ctx := context.Background()
	if _, err := pool.Exec(ctx, `TRUNCATE jobs`); err != nil {
		t.Fatal(err)
	}
	for i := range n {
		if _, err := queue.Enqueue(ctx, pool, fmt.Sprintf("job-%d", i)); err != nil {
			t.Fatal(err)
		}
	}
}

func count(t *testing.T, pool *pgxpool.Pool, query string) int {
	t.Helper()
	var n int
	if err := pool.QueryRow(context.Background(), query).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

// Eight workers drain 300 jobs. Every job is handled exactly once, and the work is shared.
func TestWorkersNeverShareAJob(t *testing.T) {
	pool := testdb.Pool(t)
	reset(t, pool, 300)

	var mu sync.Mutex
	seen := map[int64]int{}
	handle := func(_ context.Context, _ pgx.Tx, j queue.Job) error {
		mu.Lock()
		seen[j.ID]++
		mu.Unlock()
		time.Sleep(time.Millisecond) // a little work, so workers overlap
		return nil
	}
	var wg sync.WaitGroup
	for w := range 8 {
		wg.Go(func() {
			for {
				ok, err := queue.Work(context.Background(), pool, fmt.Sprintf("worker-%d", w), handle)
				if err != nil {
					t.Errorf("worker %d: %v", w, err)
					return
				}
				if !ok {
					return
				}
			}
		})
	}
	wg.Wait()

	if len(seen) != 300 {
		t.Fatalf("%d distinct jobs handled, want 300", len(seen))
	}
	for id, n := range seen {
		if n != 1 {
			t.Fatalf("job %d handled %d times", id, n)
		}
	}
	if got := count(t, pool, `SELECT count(*) FROM jobs WHERE status = 'done'`); got != 300 {
		t.Fatalf("%d jobs done, want 300", got)
	}
	if got := count(t, pool, `SELECT count(DISTINCT done_by) FROM jobs`); got < 2 {
		t.Fatalf("only %d worker did any work; SKIP LOCKED should let them share", got)
	}
}

// While one worker holds the first job, a second worker gets the next one at once
// instead of waiting: that is SKIP LOCKED.
func TestSecondWorkerSkipsALockedJob(t *testing.T) {
	pool := testdb.Pool(t)
	reset(t, pool, 2)
	ctx := context.Background()

	holding := make(chan int64)
	release := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		_, err := queue.Work(ctx, pool, "slow", func(_ context.Context, _ pgx.Tx, j queue.Job) error {
			holding <- j.ID
			<-release
			return nil
		})
		done <- err
	}()
	first := <-holding

	quick, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var second int64
	ok, err := queue.Work(quick, pool, "fast", func(_ context.Context, _ pgx.Tx, j queue.Job) error {
		second = j.ID
		return nil
	})
	close(release)
	if err != nil || !ok {
		t.Fatalf("second worker: ok=%v err=%v", ok, err)
	}
	if second == first {
		t.Fatalf("both workers got job %d", first)
	}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}

// A failing handler rolls the claim back: the job is queued again, not lost.
func TestFailedJobGoesBackToTheQueue(t *testing.T) {
	pool := testdb.Pool(t)
	reset(t, pool, 1)
	boom := errors.New("boom")
	ok, err := queue.Work(context.Background(), pool, "w", func(context.Context, pgx.Tx, queue.Job) error { return boom })
	if ok || !errors.Is(err, boom) {
		t.Fatalf("ok=%v err=%v, want false and boom", ok, err)
	}
	if got := count(t, pool, `SELECT count(*) FROM jobs WHERE status = 'queued'`); got != 1 {
		t.Fatalf("%d queued jobs, want 1", got)
	}
}

func TestWorkOnEmptyQueue(t *testing.T) {
	pool := testdb.Pool(t)
	reset(t, pool, 0)
	ok, err := queue.Work(context.Background(), pool, "w", func(context.Context, pgx.Tx, queue.Job) error {
		t.Fatal("handler called on an empty queue")
		return nil
	})
	if ok || err != nil {
		t.Fatalf("ok=%v err=%v, want false and nil", ok, err)
	}
}
