// Package scheduler is drill 2: an in-process task scheduler with a worker limit, pause
// and resume, retries with backoff, and cancellation through context.
// Time box: 35 minutes. The tests in scheduler_test.go are the specification.
package scheduler

import (
	"context"
	"errors"
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

// Permanent marks an error that must not be retried. errors.Is and errors.As must still
// see the error it wraps.
func Permanent(err error) error {
	panic("TODO: drill 2, see chapter 35, section 35.3")
}

// Exponential returns base, 2*base, 4*base, ... capped at limit.
func Exponential(base, limit time.Duration) func(int) time.Duration {
	panic("TODO: drill 2, see chapter 35, section 35.3")
}

// Scheduler runs submitted tasks at their time, at most workers at once. Submit, Cancel,
// Pause, Resume and Status may be called from any goroutine, before or during Run.
type Scheduler struct {
	// TODO: your fields.
}

// New returns a scheduler that uses clk and runs at most workers tasks at once.
func New(clk clock.Clock, workers int) *Scheduler {
	panic("TODO: drill 2, see chapter 35, section 35.3")
}

// Submit adds a task whose first attempt may start at at (or immediately, if at has passed).
func (s *Scheduler) Submit(id string, at time.Time, task Task, opts Options) error {
	panic("TODO: drill 2, see chapter 35, section 35.3")
}

// Cancel stops a task. A pending task never runs again; a running one has its context
// cancelled and is marked Cancelled when the attempt returns, whatever it returns.
func (s *Scheduler) Cancel(id string) error {
	panic("TODO: drill 2, see chapter 35, section 35.3")
}

// Pause stops new attempts from starting. Attempts already running continue.
func (s *Scheduler) Pause() {
	panic("TODO: drill 2, see chapter 35, section 35.3")
}

// Resume lets attempts start again, beginning with every task that became due meanwhile.
func (s *Scheduler) Resume() {
	panic("TODO: drill 2, see chapter 35, section 35.3")
}

// Status returns a snapshot of the task.
func (s *Scheduler) Status(id string) (Status, error) {
	panic("TODO: drill 2, see chapter 35, section 35.3")
}

// Done returns a channel that is closed when the task reaches a final state, or nil for an
// unknown id.
func (s *Scheduler) Done(id string) <-chan struct{} {
	panic("TODO: drill 2, see chapter 35, section 35.3")
}

// Run starts due tasks until ctx is cancelled, then cancels running attempts, waits for
// them and returns. An attempt interrupted this way does not count: the task is Pending
// again, so a later Run would retry it.
func (s *Scheduler) Run(ctx context.Context) error {
	panic("TODO: drill 2, see chapter 35, section 35.3")
}
