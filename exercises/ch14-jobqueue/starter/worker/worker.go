// Package worker runs queue jobs with a fixed number of goroutines, retries failures with
// backoff, and drains on shutdown: when its context is cancelled it stops claiming, lets
// every claimed job finish and records its result, and only then returns.
package worker

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"example.com/ch14-jobqueue/backoff"
	"example.com/ch14-jobqueue/queue"
)

// Handler runs one job. It must be idempotent: after a crash or an expired lease the same
// job runs again. Return queue.Permanent(err) for failures that retrying cannot fix.
type Handler func(ctx context.Context, j queue.Job) error

// Pool is a set of workers sharing one database pool.
type Pool struct {
	DB       *pgxpool.Pool
	Handlers map[string]Handler // by job kind
	Name     string             // worker name prefix, recorded in locked_by
	Workers  int                // goroutines; 0 means 1
	Lease    time.Duration      // how long a claim lasts; also the handler's deadline
	Poll     time.Duration      // idle wait when there is no job
	Backoff  backoff.Policy
	Rand     backoff.Rand // nil means backoff.Global
	Log      *slog.Logger // nil means slog.Default()
}

// Run starts the workers and blocks until ctx is cancelled and every worker has finished
// its current job. It returns nil after a clean drain.
func (p *Pool) Run(ctx context.Context) error {
	n := max(p.Workers, 1)
	var wg sync.WaitGroup
	for i := range n {
		name := fmt.Sprintf("%s-%d", p.Name, i)
		wg.Go(func() { p.loop(ctx, name) })
	}
	wg.Wait()
	return nil
}

func (p *Pool) loop(ctx context.Context, name string) {
	for ctx.Err() == nil { // stop claiming as soon as shutdown starts
		j, err := queue.Claim(ctx, p.DB, name, p.Lease)
		if err != nil && ctx.Err() == nil {
			p.log().Error("claim failed", "worker", name, "err", err)
		}
		if j == nil {
			select { // nothing to do (or an error): wait, unless shutting down
			case <-ctx.Done():
			case <-time.After(p.Poll):
			}
			continue
		}
		p.runOne(ctx, name, j)
	}
}

// runOne runs a claimed job to the end even if shutdown starts meanwhile. The handler's
// context is detached from ctx (it is not cancelled by SIGTERM) but bounded by the lease:
// after the lease another worker may take the job, so running longer is pointless.
func (p *Pool) runOne(ctx context.Context, name string, j *queue.Job) {
	// TODO(day 42): drain. Detach the handler from shutdown: build jobCtx from
	// context.WithoutCancel(ctx), still bounded by p.Lease.
	jobCtx, cancel := context.WithTimeout(ctx, p.Lease)
	defer cancel()
	err := p.call(jobCtx, *j)

	// Record the result even during shutdown: that is the point of draining.
	// TODO(day 42): the same here, or the result is lost when ctx is already cancelled.
	rec, cancelRec := context.WithTimeout(ctx, 10*time.Second)
	defer cancelRec()
	log := p.log().With("worker", name, "job", j.ID, "kind", j.Kind, "attempt", j.Attempts)
	if err == nil {
		if err := queue.Complete(rec, p.DB, j); err != nil {
			log.Error("complete failed", "err", err)
			return
		}
		log.Info("job done")
		return
	}
	delay := p.Backoff.Delay(j.Attempts, p.rand())
	state, ferr := queue.Fail(rec, p.DB, j, err, delay)
	if ferr != nil {
		log.Error("recording failure failed", "err", ferr, "cause", err)
		return
	}
	log.Warn("job failed", "err", err, "state", state, "retry_in", delay)
}

// call runs the handler and turns a missing handler or a panic into a failure, so one bad
// job cannot kill the worker.
func (p *Pool) call(ctx context.Context, j queue.Job) (err error) {
	h, ok := p.Handlers[j.Kind]
	if !ok {
		return queue.Permanent(fmt.Errorf("no handler for kind %q", j.Kind))
	}
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("handler panic: %v", r)
		}
	}()
	return h(ctx, j)
}

func (p *Pool) rand() backoff.Rand {
	if p.Rand == nil {
		return backoff.Global
	}
	return p.Rand
}

func (p *Pool) log() *slog.Logger {
	if p.Log == nil {
		return slog.Default()
	}
	return p.Log
}
