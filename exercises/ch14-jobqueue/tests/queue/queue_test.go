package queue_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"example.com/ch14-jobqueue/queue"
	"example.com/ch14-jobqueue/testenv"
)

var ctx = context.Background()

func enqueue(t *testing.T, db *pgxpool.Pool, opt queue.Options) int64 {
	t.Helper()
	id, _, err := queue.Enqueue(ctx, db, "test", nil, opt)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func claim(t *testing.T, db *pgxpool.Pool, worker string, lease time.Duration) *queue.Job {
	t.Helper()
	j, err := queue.Claim(ctx, db, worker, lease)
	if err != nil {
		t.Fatal(err)
	}
	return j
}

func get(t *testing.T, db *pgxpool.Pool, id int64) queue.Job {
	t.Helper()
	j, err := queue.Get(ctx, db, id)
	if err != nil {
		t.Fatal(err)
	}
	return j
}

func TestEnqueueClaimComplete(t *testing.T) {
	db := testenv.Postgres(t)
	id := enqueue(t, db, queue.Options{})
	j := claim(t, db, "w1", time.Minute)
	if j == nil || j.ID != id || j.State != queue.Running || j.Attempts != 1 || j.LockedBy != "w1" {
		t.Fatalf("claimed %+v, want job %d running, attempt 1, held by w1", j, id)
	}
	if again := claim(t, db, "w2", time.Minute); again != nil {
		t.Fatalf("second claim got job %d while it is leased", again.ID)
	}
	if err := queue.Complete(ctx, db, j); err != nil {
		t.Fatal(err)
	}
	if got := get(t, db, id); got.State != queue.Done {
		t.Errorf("state = %s, want done", got.State)
	}
}

func TestClaimRespectsRunAt(t *testing.T) {
	db := testenv.Postgres(t)
	enqueue(t, db, queue.Options{RunAt: time.Now().Add(time.Hour)})
	if j := claim(t, db, "w1", time.Minute); j != nil {
		t.Fatalf("claimed job %d an hour early", j.ID)
	}
	now := enqueue(t, db, queue.Options{})
	if j := claim(t, db, "w1", time.Minute); j == nil || j.ID != now {
		t.Fatalf("claimed %+v, want the due job %d", j, now)
	}
}

func TestUniqueKeyEnqueuesOnce(t *testing.T) {
	db := testenv.Postgres(t)
	first, created1, err := queue.Enqueue(ctx, db, "report", nil, queue.Options{UniqueKey: "report@2026-10-05"})
	if err != nil {
		t.Fatal(err)
	}
	second, created2, err := queue.Enqueue(ctx, db, "report", nil, queue.Options{UniqueKey: "report@2026-10-05"})
	if err != nil {
		t.Fatal(err)
	}
	if first != second || !created1 || created2 {
		t.Errorf("ids %d and %d, created %v and %v; want one job, created once", first, second, created1, created2)
	}
}

func TestConcurrentClaimsNeverShareAJob(t *testing.T) {
	db := testenv.Postgres(t)
	const jobs, workers = 60, 8
	for range jobs {
		enqueue(t, db, queue.Options{})
	}
	var mu sync.Mutex
	seen := map[int64]string{}
	var wg sync.WaitGroup
	for w := range workers {
		name := string(rune('a' + w))
		wg.Go(func() {
			for {
				j, err := queue.Claim(ctx, db, name, time.Minute)
				if err != nil {
					t.Error(err)
					return
				}
				if j == nil {
					return
				}
				mu.Lock()
				if other, dup := seen[j.ID]; dup {
					t.Errorf("job %d claimed by %s and %s", j.ID, other, name)
				}
				seen[j.ID] = name
				mu.Unlock()
			}
		})
	}
	wg.Wait()
	if len(seen) != jobs {
		t.Errorf("claimed %d distinct jobs, want %d", len(seen), jobs)
	}
}

func TestFailReschedulesLater(t *testing.T) {
	db := testenv.Postgres(t)
	id := enqueue(t, db, queue.Options{MaxAttempts: 3})
	j := claim(t, db, "w1", time.Minute)
	state, err := queue.Fail(ctx, db, j, errors.New("receiver returned 503"), time.Hour)
	if err != nil || state != queue.Pending {
		t.Fatalf("Fail = %s, %v; want pending", state, err)
	}
	got := get(t, db, id)
	if until := time.Until(got.RunAt); until < 59*time.Minute {
		t.Errorf("run_at is %v from now, want about an hour", until)
	}
	if got.LastError != "receiver returned 503" || got.Attempts != 1 {
		t.Errorf("job = %+v, want last_error recorded and attempts 1", got)
	}
	if j := claim(t, db, "w1", time.Minute); j != nil {
		t.Errorf("claimed the job before its retry time")
	}
}

func TestLifecycleToDeadAndReplay(t *testing.T) {
	tests := []struct {
		name        string
		maxAttempts int
		cause       error
		wantStates  []queue.State // state after each failed attempt
	}{
		{"transient errors use every attempt", 3, errors.New("timeout"), []queue.State{queue.Pending, queue.Pending, queue.Dead}},
		{"one attempt allowed", 1, errors.New("timeout"), []queue.State{queue.Dead}},
		{"permanent error dies at once", 5, queue.Permanent(errors.New("410 gone")), []queue.State{queue.Dead}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := testenv.Postgres(t)
			id := enqueue(t, db, queue.Options{MaxAttempts: tt.maxAttempts})
			for i, want := range tt.wantStates {
				j := claim(t, db, "w1", time.Minute)
				if j == nil {
					t.Fatalf("attempt %d: nothing to claim", i+1)
				}
				got, err := queue.Fail(ctx, db, j, tt.cause, 0)
				if err != nil || got != want {
					t.Fatalf("attempt %d: Fail = %s, %v; want %s", i+1, got, err, want)
				}
			}
			if j := claim(t, db, "w1", time.Minute); j != nil {
				t.Fatalf("a dead job was claimed")
			}
			dead, err := queue.ListDead(ctx, db, 10)
			if err != nil || len(dead) != 1 || dead[0].ID != id {
				t.Fatalf("ListDead = %+v, %v; want job %d", dead, err, id)
			}

			if err := queue.Replay(ctx, db, id); err != nil {
				t.Fatal(err)
			}
			if err := queue.Replay(ctx, db, id); !errors.Is(err, queue.ErrNotDead) {
				t.Errorf("second Replay = %v, want ErrNotDead", err)
			}
			j := claim(t, db, "w2", time.Minute)
			if j == nil || j.ID != id || j.Attempts != 1 {
				t.Fatalf("after replay claimed %+v, want job %d on a fresh attempt 1", j, id)
			}
			if err := queue.Complete(ctx, db, j); err != nil {
				t.Fatal(err)
			}
			history, err := queue.Attempts(ctx, db, id)
			if err != nil {
				t.Fatal(err)
			}
			if len(history) != len(tt.wantStates)+1 || history[len(history)-1].Error != "" {
				t.Errorf("history = %+v, want %d failures then a success", history, len(tt.wantStates))
			}
		})
	}
}

func TestExpiredLeaseIsReclaimedAndFenced(t *testing.T) {
	db := testenv.Postgres(t)
	id := enqueue(t, db, queue.Options{MaxAttempts: 3})
	stale := claim(t, db, "w1", 100*time.Millisecond)
	time.Sleep(300 * time.Millisecond) // w1 "hangs" past its lease
	fresh := claim(t, db, "w2", time.Minute)
	if fresh == nil || fresh.ID != id || fresh.Attempts != 2 {
		t.Fatalf("reclaim = %+v, want job %d on attempt 2", fresh, id)
	}
	if err := queue.Complete(ctx, db, stale); !errors.Is(err, queue.ErrLeaseLost) {
		t.Errorf("stale worker's Complete = %v, want ErrLeaseLost", err)
	}
	if err := queue.Complete(ctx, db, fresh); err != nil {
		t.Fatal(err)
	}
	history, err := queue.Attempts(ctx, db, id)
	if err != nil {
		t.Fatal(err)
	}
	want := []queue.Attempt{{Attempt: 1, Worker: "w1", Error: "lease expired"}, {Attempt: 2, Worker: "w2"}}
	if len(history) != 2 || history[0] != want[0] || history[1] != want[1] {
		t.Errorf("history = %+v, want %+v", history, want)
	}
}

func TestExpiredLeaseOnLastAttemptGoesDead(t *testing.T) {
	db := testenv.Postgres(t)
	id := enqueue(t, db, queue.Options{MaxAttempts: 1})
	claim(t, db, "w1", 100*time.Millisecond)
	time.Sleep(300 * time.Millisecond)
	if j := claim(t, db, "w2", time.Minute); j != nil {
		t.Fatalf("claimed job %d after its last attempt expired", j.ID)
	}
	if got := get(t, db, id); got.State != queue.Dead {
		t.Errorf("state = %s, want dead", got.State)
	}
}
