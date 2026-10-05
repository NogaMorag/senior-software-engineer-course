// Package scheduler is drill 2: an in-process task scheduler with a worker limit, pause
// and resume, retries with backoff, and cancellation through context.
package scheduler

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"example.com/ch35-drills/clock"
)

// Task is one unit of work. It must return soon after ctx is cancelled.
type Task func(ctx context.Context) error

// State is where a task is in its life.
type State int

const (
	Pending   State = iota // waiting for its time, for a free worker, or for Resume
	Running                // an attempt is in progress
	Succeeded              // an attempt returned nil
	Failed                 // the last allowed attempt failed, or an attempt failed permanently
	Cancelled              // Cancel was called before the task finished
)

func (s State) String() string {
	return [...]string{"pending", "running", "succeeded", "failed", "cancelled"}[s]
}

// Options control retries. The zero value means one attempt.
type Options struct {
	MaxAttempts int                             // attempts in total, including the first; < 1 means 1
	Backoff     func(attempt int) time.Duration // delay after failed attempt n (1-based); nil means no delay
}

// Status is a snapshot of one task.
type Status struct {
	State    State
	Attempts int       // attempts started so far
	NextRun  time.Time // when Pending: the earliest time of the next attempt
	Err      error     // the error of the last failed attempt, if any
}

var (
	ErrDuplicate = errors.New("scheduler: task id already exists")
	ErrNotFound  = errors.New("scheduler: no such task")
	ErrFinished  = errors.New("scheduler: task already finished")
)

// Permanent marks an error that must not be retried.
func Permanent(err error) error { return permanentError{err} }

type permanentError struct{ err error }

func (p permanentError) Error() string { return "permanent: " + p.err.Error() }
func (p permanentError) Unwrap() error { return p.err }

// Exponential returns base, 2*base, 4*base, ... capped at limit.
func Exponential(base, limit time.Duration) func(int) time.Duration {
	return func(attempt int) time.Duration {
		d := base
		for i := 1; i < attempt; i++ {
			d *= 2
			if d >= limit || d <= 0 { // d <= 0: overflow
				return limit
			}
		}
		return min(d, limit)
	}
}

type entry struct {
	seq     int // submission order breaks ties between equal NextRun
	task    Task
	opts    Options
	state   State
	attempt int
	next    time.Time
	err     error
	cancel  context.CancelFunc // set while Running
	stopped bool               // Cancel was called while Running
	done    chan struct{}      // closed when the task reaches a final state
}

// Scheduler runs submitted tasks at their time, at most workers at once. Submit, Cancel,
// Pause, Resume and Status may be called from any goroutine, before or during Run.
type Scheduler struct {
	clk     clock.Clock
	workers int

	mu      sync.Mutex
	tasks   map[string]*entry
	seq     int
	paused  bool
	running int
	wake    chan struct{} // capacity 1: "something changed, look again"
}

// New returns a scheduler that uses clk and runs at most workers tasks at once.
func New(clk clock.Clock, workers int) *Scheduler {
	return &Scheduler{clk: clk, workers: max(workers, 1), tasks: make(map[string]*entry), wake: make(chan struct{}, 1)}
}

// Submit adds a task whose first attempt may start at at (or immediately, if at has passed).
func (s *Scheduler) Submit(id string, at time.Time, task Task, opts Options) error {
	if opts.MaxAttempts < 1 {
		opts.MaxAttempts = 1
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.tasks[id]; ok {
		return fmt.Errorf("%w: %q", ErrDuplicate, id)
	}
	s.seq++
	s.tasks[id] = &entry{seq: s.seq, task: task, opts: opts, next: at, done: make(chan struct{})}
	s.notify()
	return nil
}

// Cancel stops a task. A pending task never runs again; a running one has its context
// cancelled and is marked Cancelled when the attempt returns, whatever it returns.
func (s *Scheduler) Cancel(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.tasks[id]
	if !ok {
		return fmt.Errorf("%w: %q", ErrNotFound, id)
	}
	switch e.state {
	case Pending:
		s.finish(e, Cancelled)
	case Running:
		e.stopped = true
		e.cancel()
	default:
		return fmt.Errorf("%w: %q is %s", ErrFinished, id, e.state)
	}
	return nil
}

// Pause stops new attempts from starting. Attempts already running continue.
func (s *Scheduler) Pause() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.paused = true
}

// Resume lets attempts start again, beginning with every task that became due meanwhile.
func (s *Scheduler) Resume() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.paused = false
	s.notify()
}

// Status returns a snapshot of the task.
func (s *Scheduler) Status(id string) (Status, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.tasks[id]
	if !ok {
		return Status{}, fmt.Errorf("%w: %q", ErrNotFound, id)
	}
	return Status{State: e.state, Attempts: e.attempt, NextRun: e.next, Err: e.err}, nil
}

// Done returns a channel that is closed when the task reaches a final state, or nil for an
// unknown id.
func (s *Scheduler) Done(id string) <-chan struct{} {
	s.mu.Lock()
	defer s.mu.Unlock()
	if e, ok := s.tasks[id]; ok {
		return e.done
	}
	return nil
}

// Run starts due tasks until ctx is cancelled, then cancels running attempts, waits for
// them and returns. An attempt interrupted this way does not count: the task is Pending
// again, so a later Run would retry it.
func (s *Scheduler) Run(ctx context.Context) error {
	var wg sync.WaitGroup
	defer wg.Wait()
	for {
		timer := s.startDue(ctx, &wg)
		var fire <-chan time.Time
		if timer != nil {
			fire = timer.C()
		}
		select {
		case <-ctx.Done():
			if timer != nil {
				timer.Stop()
			}
			return nil
		case <-s.wake:
		case <-fire:
		}
		if timer != nil {
			timer.Stop()
		}
	}
}

// startDue starts every attempt that may start now and returns a timer for the earliest
// pending task that is not due yet (nil if there is none, or while paused).
func (s *Scheduler) startDue(ctx context.Context, wg *sync.WaitGroup) clock.Timer {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.paused {
		return nil
	}
	now := s.clk.Now()
	pending := make([]*entry, 0, len(s.tasks))
	for _, e := range s.tasks {
		if e.state == Pending {
			pending = append(pending, e)
		}
	}
	// Earliest first, then first submitted: the order is deterministic. A heap would make
	// this O(log n) per change instead of O(n log n) per wake-up.
	sort.Slice(pending, func(i, j int) bool {
		if !pending[i].next.Equal(pending[j].next) {
			return pending[i].next.Before(pending[j].next)
		}
		return pending[i].seq < pending[j].seq
	})
	for _, e := range pending {
		if e.next.After(now) {
			// Not due: sleep until it is. Workers that free up later send a wake-up.
			return s.clk.NewTimer(e.next.Sub(now))
		}
		if s.running == s.workers {
			return nil // a finishing attempt will wake the loop
		}
		s.start(ctx, wg, e)
	}
	return nil
}

// start runs one attempt of e in its own goroutine. Called with s.mu held.
func (s *Scheduler) start(ctx context.Context, wg *sync.WaitGroup, e *entry) {
	actx, cancel := context.WithCancel(ctx)
	e.state, e.cancel = Running, cancel
	e.attempt++
	s.running++
	wg.Go(func() {
		err := call(actx, e.task)
		cancel()
		s.attemptDone(ctx, e, err)
	})
}

// call runs the task and turns a panic into an error, so one bad task cannot kill the
// process or leak its worker slot.
func call(ctx context.Context, task Task) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = Permanent(fmt.Errorf("task panicked: %v", r))
		}
	}()
	return task(ctx)
}

func (s *Scheduler) attemptDone(runCtx context.Context, e *entry, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.running--
	e.cancel = nil
	if err != nil {
		e.err = err
	}
	var permanent permanentError
	switch {
	case e.stopped:
		s.finish(e, Cancelled)
	case err == nil:
		s.finish(e, Succeeded)
	case runCtx.Err() != nil:
		// Shutdown, not a failure of the task: give the attempt back.
		e.state, e.next = Pending, s.clk.Now()
		e.attempt--
	case errors.As(err, &permanent) || e.attempt >= e.opts.MaxAttempts:
		s.finish(e, Failed)
	default:
		var delay time.Duration
		if e.opts.Backoff != nil {
			delay = e.opts.Backoff(e.attempt)
		}
		e.state, e.next = Pending, s.clk.Now().Add(delay)
	}
	s.notify()
}

// finish moves e to a final state. Called with s.mu held.
func (s *Scheduler) finish(e *entry, st State) {
	e.state = st
	e.next = time.Time{}
	close(e.done)
}

// notify wakes the Run loop without ever blocking: one pending wake-up is enough, because
// the loop re-reads all state when it wakes.
func (s *Scheduler) notify() {
	select {
	case s.wake <- struct{}{}:
	default:
	}
}
