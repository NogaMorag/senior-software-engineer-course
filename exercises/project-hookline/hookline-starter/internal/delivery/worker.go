package delivery

import (
	"context"

	"log/slog"

	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"example.com/hookline/internal/backoff"
)

// Pool is a set of delivery workers sharing one database pool (chapter 14's worker, on deliveries).
type Pool struct {
	DB          *pgxpool.Pool
	Sender      Sender
	Workers     int            // goroutines; 0 means 1
	Lease       time.Duration  // how long a claim lasts; must exceed Sender.Timeout
	Poll        time.Duration  // idle wait when nothing is due
	MaxAttempts int            // per round; a replay starts a new round
	Backoff     backoff.Policy // delay before retry n of a round
	Rand        backoff.Rand   // nil means backoff.Global
	Log         *slog.Logger   // nil discards
	// Guards are the per-endpoint breakers and bulkheads (milestone M6). nil means none.
	Guards *Guards
}

// Run starts the workers and blocks until ctx is cancelled and every worker has finished and recorded
// its current attempt (drain). It returns nil after a clean drain.
func (p *Pool) Run(ctx context.Context) error {
	panic("TODO: see chapter 38")

}

func (p *Pool) timeout() time.Duration {
	panic("TODO: see chapter 38")

}

func (p *Pool) loop(ctx context.Context, worker int) {
	panic("TODO: see chapter 38")

}

// RunOne sends one claimed delivery and records the result, even if shutdown starts meanwhile: the
// attempt's context is detached from ctx but bounded by the lease, because after the lease another
// worker may take the delivery and our result no longer counts.
func (p *Pool) RunOne(ctx context.Context, j *Job) {
	panic("TODO: see chapter 38")

}

// guardedSend sends j through its endpoint's bulkhead and breaker. When either refuses, the delivery is
// postponed without spending an attempt, and sent is false: a slow or dead endpoint costs this worker a
// database round trip, not a worker blocked for the whole timeout.
func (p *Pool) guardedSend(ctx context.Context, j *Job) (res Result, sent bool) {
	panic("TODO: see chapter 38")

}

func (p *Pool) postpone(ctx context.Context, j *Job, after time.Duration, why error) {
	panic("TODO: see chapter 38")

}

func (p *Pool) rand() backoff.Rand {
	panic("TODO: see chapter 38")

}

func (p *Pool) log() *slog.Logger {
	panic("TODO: see chapter 38")

}
