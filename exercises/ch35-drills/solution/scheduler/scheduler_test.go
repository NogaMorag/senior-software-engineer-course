package scheduler

import (
	"context"
	"errors"
	"slices"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"example.com/ch35-drills/clock"
)

var t0 = time.Date(2026, 10, 5, 9, 0, 0, 0, time.UTC)

// start runs a scheduler on a fake clock for the duration of the test.
func start(t *testing.T, workers int) (*Scheduler, *clock.Fake) {
	t.Helper()
	clk := clock.NewFake(t0)
	s := New(clk, workers)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- s.Run(ctx) }()
	t.Cleanup(func() {
		cancel()
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Error("Run did not return after its context was cancelled")
		}
	})
	return s, clk
}

// eventually polls cond in real time. Time in the scheduler is fake; only the goroutines
// need a moment to react.
func eventually(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for: %s", what)
		}
		time.Sleep(time.Millisecond)
	}
}

// settle gives the scheduler a real moment in which it must NOT do anything.
func settle() { time.Sleep(20 * time.Millisecond) }

func status(t *testing.T, s *Scheduler, id string) Status {
	t.Helper()
	st, err := s.Status(id)
	if err != nil {
		t.Fatalf("Status(%q): %v", id, err)
	}
	return st
}

func waitState(t *testing.T, s *Scheduler, id string, want State) Status {
	t.Helper()
	var st Status
	eventually(t, id+" "+want.String(), func() bool { st = status(t, s, id); return st.State == want })
	return st
}

func ok(context.Context) error { return nil }

var errBoom = errors.New("boom")

func TestRunsAtItsTimeNotBefore(t *testing.T) {
	s, clk := start(t, 1)
	var calls atomic.Int32
	task := func(context.Context) error { calls.Add(1); return nil }
	if err := s.Submit("a", t0.Add(time.Minute), task, Options{}); err != nil {
		t.Fatal(err)
	}
	clk.Advance(59 * time.Second)
	settle()
	if st := status(t, s, "a"); st.State != Pending || calls.Load() != 0 {
		t.Fatalf("after 59s: state %s, calls %d; want pending, 0", st.State, calls.Load())
	}
	clk.Advance(time.Second)
	waitState(t, s, "a", Succeeded)
	if calls.Load() != 1 {
		t.Fatalf("calls = %d, want 1", calls.Load())
	}
}

func TestPastTimeRunsImmediately(t *testing.T) {
	s, _ := start(t, 1)
	if err := s.Submit("a", t0.Add(-time.Hour), ok, Options{}); err != nil {
		t.Fatal(err)
	}
	waitState(t, s, "a", Succeeded)
}

func TestRetryWaitsForBackoff(t *testing.T) {
	s, clk := start(t, 1)
	var calls atomic.Int32
	task := func(context.Context) error {
		if calls.Add(1) < 3 {
			return errBoom
		}
		return nil
	}
	opts := Options{MaxAttempts: 5, Backoff: Exponential(time.Second, time.Minute)}
	if err := s.Submit("a", t0, task, opts); err != nil {
		t.Fatal(err)
	}

	var st Status
	eventually(t, "first attempt failed", func() bool { st = status(t, s, "a"); return st.Attempts == 1 && st.State == Pending })
	if !st.NextRun.Equal(t0.Add(time.Second)) || !errors.Is(st.Err, errBoom) {
		t.Fatalf("after attempt 1: NextRun %v, Err %v; want t0+1s, boom", st.NextRun, st.Err)
	}
	clk.Advance(999 * time.Millisecond)
	settle()
	if calls.Load() != 1 {
		t.Fatal("retried before the backoff of 1s had passed")
	}
	clk.Advance(time.Millisecond)
	eventually(t, "second attempt failed", func() bool { st = status(t, s, "a"); return st.Attempts == 2 && st.State == Pending })
	if want := clk.Now().Add(2 * time.Second); !st.NextRun.Equal(want) {
		t.Fatalf("after attempt 2: NextRun %v, want %v (backoff doubles)", st.NextRun, want)
	}
	clk.Advance(2 * time.Second)
	st = waitState(t, s, "a", Succeeded)
	if st.Attempts != 3 {
		t.Fatalf("Attempts = %d, want 3", st.Attempts)
	}
}

func TestGivesUpAfterMaxAttempts(t *testing.T) {
	s, _ := start(t, 1)
	var calls atomic.Int32
	task := func(context.Context) error { calls.Add(1); return errBoom }
	if err := s.Submit("a", t0, task, Options{MaxAttempts: 3}); err != nil {
		t.Fatal(err)
	}
	st := waitState(t, s, "a", Failed)
	if st.Attempts != 3 || calls.Load() != 3 || !errors.Is(st.Err, errBoom) {
		t.Fatalf("got %+v with %d calls; want 3 attempts, 3 calls, Err boom", st, calls.Load())
	}
}

func TestPermanentErrorIsNotRetried(t *testing.T) {
	s, _ := start(t, 1)
	task := func(context.Context) error { return Permanent(errBoom) }
	if err := s.Submit("a", t0, task, Options{MaxAttempts: 5}); err != nil {
		t.Fatal(err)
	}
	st := waitState(t, s, "a", Failed)
	if st.Attempts != 1 || !errors.Is(st.Err, errBoom) {
		t.Fatalf("got %+v; want 1 attempt and an error that wraps boom", st)
	}
}

func TestPanicFailsTheTaskNotTheScheduler(t *testing.T) {
	s, _ := start(t, 1)
	bad := func(context.Context) error { panic("bug in task") }
	if err := s.Submit("bad", t0, bad, Options{MaxAttempts: 3}); err != nil {
		t.Fatal(err)
	}
	if st := waitState(t, s, "bad", Failed); st.Attempts != 1 || st.Err == nil {
		t.Fatalf("got %+v; want Failed after 1 attempt with an error", st)
	}
	if err := s.Submit("good", t0, ok, Options{}); err != nil {
		t.Fatal(err)
	}
	waitState(t, s, "good", Succeeded) // the worker slot was given back
}

func TestCancelPendingTask(t *testing.T) {
	s, clk := start(t, 1)
	var calls atomic.Int32
	task := func(context.Context) error { calls.Add(1); return nil }
	if err := s.Submit("a", t0.Add(time.Minute), task, Options{}); err != nil {
		t.Fatal(err)
	}
	if err := s.Cancel("a"); err != nil {
		t.Fatal(err)
	}
	select {
	case <-s.Done("a"):
	default:
		t.Fatal("Done not closed after cancelling a pending task")
	}
	clk.Advance(time.Hour)
	settle()
	if st := status(t, s, "a"); st.State != Cancelled || calls.Load() != 0 {
		t.Fatalf("state %s, calls %d; want cancelled, 0", st.State, calls.Load())
	}
}

func TestCancelRunningTaskStopsItWithoutRetry(t *testing.T) {
	s, _ := start(t, 1)
	entered := make(chan struct{})
	task := func(ctx context.Context) error {
		close(entered)
		<-ctx.Done()
		return ctx.Err()
	}
	if err := s.Submit("a", t0, task, Options{MaxAttempts: 3}); err != nil {
		t.Fatal(err)
	}
	<-entered
	if err := s.Cancel("a"); err != nil {
		t.Fatal(err)
	}
	st := waitState(t, s, "a", Cancelled)
	if st.Attempts != 1 {
		t.Fatalf("Attempts = %d, want 1: a cancelled task is not retried", st.Attempts)
	}
}

func TestErrors(t *testing.T) {
	s, _ := start(t, 1)
	if err := s.Submit("a", t0, ok, Options{}); err != nil {
		t.Fatal(err)
	}
	if err := s.Submit("a", t0, ok, Options{}); !errors.Is(err, ErrDuplicate) {
		t.Fatalf("second Submit: %v, want ErrDuplicate", err)
	}
	waitState(t, s, "a", Succeeded)
	if err := s.Cancel("a"); !errors.Is(err, ErrFinished) {
		t.Fatalf("Cancel finished: %v, want ErrFinished", err)
	}
	if err := s.Cancel("nope"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Cancel unknown: %v, want ErrNotFound", err)
	}
	if _, err := s.Status("nope"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Status unknown: %v, want ErrNotFound", err)
	}
	if s.Done("nope") != nil {
		t.Fatal("Done of an unknown id must be nil")
	}
}

func TestPauseHoldsNewStartsAndResumeReleasesThem(t *testing.T) {
	s, _ := start(t, 2)
	release := make(chan struct{})
	entered := make(chan struct{})
	long := func(ctx context.Context) error {
		close(entered)
		select {
		case <-release:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	if err := s.Submit("long", t0, long, Options{}); err != nil {
		t.Fatal(err)
	}
	<-entered
	s.Pause()
	if err := s.Submit("later", t0, ok, Options{}); err != nil {
		t.Fatal(err)
	}
	settle()
	if st := status(t, s, "later"); st.State != Pending || st.Attempts != 0 {
		t.Fatalf("while paused: %+v, want pending with 0 attempts", st)
	}
	close(release)
	waitState(t, s, "long", Succeeded) // a running attempt continues through a pause
	s.Resume()
	waitState(t, s, "later", Succeeded)
}

func TestWorkerLimit(t *testing.T) {
	s, _ := start(t, 2)
	release := make(chan struct{})
	var running, peak atomic.Int32
	task := func(ctx context.Context) error {
		n := running.Add(1)
		for {
			p := peak.Load()
			if n <= p || peak.CompareAndSwap(p, n) {
				break
			}
		}
		defer running.Add(-1)
		select {
		case <-release:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	ids := []string{"a", "b", "c", "d", "e"}
	for _, id := range ids {
		if err := s.Submit(id, t0, task, Options{}); err != nil {
			t.Fatal(err)
		}
	}
	eventually(t, "two tasks running", func() bool { return running.Load() == 2 })
	settle()
	if p := peak.Load(); p != 2 {
		t.Fatalf("peak concurrency %d, want 2", p)
	}
	close(release)
	for _, id := range ids {
		waitState(t, s, id, Succeeded)
	}
	if p := peak.Load(); p != 2 {
		t.Fatalf("peak concurrency %d, want 2", p)
	}
}

func TestDueTasksStartEarliestFirstThenInSubmitOrder(t *testing.T) {
	clk := clock.NewFake(t0)
	s := New(clk, 1)
	var mu sync.Mutex
	var order []string
	record := func(id string) Task {
		return func(context.Context) error {
			mu.Lock()
			defer mu.Unlock()
			order = append(order, id)
			return nil
		}
	}
	for _, sub := range []struct {
		id string
		at time.Duration
	}{{"a", 2 * time.Second}, {"b", time.Second}, {"c", time.Second}} {
		if err := s.Submit(sub.id, t0.Add(sub.at), record(sub.id), Options{}); err != nil {
			t.Fatal(err)
		}
	}
	clk.Advance(3 * time.Second) // all three are due before Run starts
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = s.Run(ctx) }()
	waitState(t, s, "a", Succeeded)
	mu.Lock()
	defer mu.Unlock()
	if want := []string{"b", "c", "a"}; !slices.Equal(order, want) {
		t.Fatalf("order %v, want %v", order, want)
	}
}

func TestRunCancelsAttemptsAndGivesThemBack(t *testing.T) {
	clk := clock.NewFake(t0)
	s := New(clk, 1)
	entered := make(chan struct{})
	task := func(ctx context.Context) error {
		close(entered)
		<-ctx.Done()
		return ctx.Err()
	}
	if err := s.Submit("a", t0, task, Options{}); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- s.Run(ctx) }()
	<-entered
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run returned %v, want nil", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return after cancel")
	}
	// Run waited for the attempt, so the state is final for this run.
	if st := status(t, s, "a"); st.State != Pending || st.Attempts != 0 {
		t.Fatalf("after shutdown: %+v, want pending with 0 attempts (the interrupted attempt is given back)", st)
	}
}

func TestExponential(t *testing.T) {
	b := Exponential(time.Second, 10*time.Second)
	cases := []struct {
		attempt int
		want    time.Duration
	}{{1, time.Second}, {2, 2 * time.Second}, {3, 4 * time.Second}, {4, 8 * time.Second}, {5, 10 * time.Second}, {100, 10 * time.Second}}
	for _, tc := range cases {
		if got := b(tc.attempt); got != tc.want {
			t.Errorf("attempt %d: %v, want %v", tc.attempt, got, tc.want)
		}
	}
}
