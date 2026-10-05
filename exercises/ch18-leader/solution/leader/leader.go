// Package leader elects one active instance among several with a PostgreSQL session-level
// advisory lock. The lock lives exactly as long as the database session that took it, so a
// leader that crashes, or whose session is terminated, loses it and a follower takes over.
//
// Every new leader also gets a term, a number one higher than the previous leader's, to
// pass to storage as a fencing token: between the moment a leader's session dies and the
// moment its heartbeat notices, it still believes it leads, and only the token stops its
// late writes.
package leader

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Schema is the table that numbers the terms. Run it once (EnsureSchema, or copy it into a
// migration).
const Schema = `CREATE TABLE IF NOT EXISTS leader_terms (
	lock_key bigint PRIMARY KEY,
	term     bigint NOT NULL
)`

// EnsureSchema creates the terms table if it is missing.
func EnsureSchema(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, Schema)
	return err
}

// Config configures an Elector. Zero durations get the defaults.
type Config struct {
	Key       int64         // advisory lock key, the same for every instance of one role
	Heartbeat time.Duration // how often the leader checks that its session is alive; default 1s
	Retry     time.Duration // how often a follower tries for the lock; default 1s
	Logger    *slog.Logger  // default slog.Default()
}

// Elector campaigns for leadership on behalf of one process.
type Elector struct {
	pool *pgxpool.Pool
	cfg  Config
}

// New returns an Elector that takes its sessions from pool.
func New(pool *pgxpool.Pool, cfg Config) *Elector {
	if cfg.Heartbeat <= 0 {
		cfg.Heartbeat = time.Second
	}
	if cfg.Retry <= 0 {
		cfg.Retry = time.Second
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	return &Elector{pool: pool, cfg: cfg}
}

// Run campaigns until ctx is done. Whenever this process wins, Run calls lead with the new
// term and a context that is cancelled as soon as leadership is in doubt: the heartbeat
// failed, or ctx ended. When lead returns, for any reason, this process steps down and
// campaigns again after Retry. Run returns ctx's error once ctx is done and lead has
// returned; the lock is released before Run returns.
func (e *Elector) Run(ctx context.Context, lead func(ctx context.Context, term int64) error) error {
	for {
		won, err := e.campaign(ctx, lead)
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err != nil {
			e.cfg.Logger.Warn("leader: campaign failed", "err", err)
		} else if won {
			e.cfg.Logger.Info("leader: stepped down", "key", e.cfg.Key)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(e.cfg.Retry):
		}
	}
}

// campaign makes one attempt. It reports whether it won; if it did, it returns only after
// the term is over.
func (e *Elector) campaign(ctx context.Context, lead func(context.Context, int64) error) (bool, error) {
	pc, err := e.pool.Acquire(ctx)
	if err != nil {
		return false, err
	}
	// A session lock belongs to this one session. Take the connection out of the pool so it
	// can never be handed to other code (or returned still holding the lock), and close it at
	// the end of the term: closing the session releases the lock even if unlocking failed.
	conn := pc.Hijack()
	defer conn.Close(context.Background())

	var got bool
	if err := conn.QueryRow(ctx, `SELECT pg_try_advisory_lock($1)`, e.cfg.Key).Scan(&got); err != nil {
		return false, fmt.Errorf("pg_try_advisory_lock: %w", err)
	}
	if !got {
		return false, nil // someone else leads; try again later
	}
	var term int64
	err = conn.QueryRow(ctx, `
		INSERT INTO leader_terms (lock_key, term) VALUES ($1, 1)
		ON CONFLICT (lock_key) DO UPDATE SET term = leader_terms.term + 1
		RETURNING term`, e.cfg.Key).Scan(&term)
	if err != nil {
		return false, fmt.Errorf("next term: %w", err) // the deferred Close releases the lock
	}
	log := e.cfg.Logger.With("key", e.cfg.Key, "term", term, "backend_pid", conn.PgConn().PID())
	log.Info("leader: elected")
	e.serve(ctx, conn, term, lead, log)
	return true, nil
}

// serve runs one term: lead in its own goroutine, the heartbeat here. It returns when ctx
// ends, lead returns, or the heartbeat fails, and only after lead has returned.
func (e *Elector) serve(ctx context.Context, conn *pgx.Conn, term int64,
	lead func(context.Context, int64) error, log *slog.Logger) {
	leadCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- lead(leadCtx, term) }()

	ticker := time.NewTicker(e.cfg.Heartbeat)
	defer ticker.Stop()
	alive := true
	var leadErr error
	for running := true; running; {
		select {
		case <-ctx.Done():
			running = false
		case leadErr = <-done:
			done = nil // already received
			running = false
		case <-ticker.C:
			hbCtx, cancelHB := context.WithTimeout(ctx, e.cfg.Heartbeat)
			_, err := conn.Exec(hbCtx, `SELECT 1`)
			cancelHB()
			if err != nil && ctx.Err() == nil {
				log.Warn("leader: heartbeat failed, leadership lost", "err", err)
				alive, running = false, false
			}
		}
	}
	cancel() // tell lead to stop, then wait for it: never two terms in one process
	if done != nil {
		leadErr = <-done
	}
	if leadErr != nil && !errors.Is(leadErr, context.Canceled) {
		log.Warn("leader: lead returned an error", "err", leadErr)
	}
	if alive { // a clean step-down: give the lock back now rather than at Close
		unlockCtx, cancelUnlock := context.WithTimeout(context.Background(), e.cfg.Heartbeat)
		defer cancelUnlock()
		if _, err := conn.Exec(unlockCtx, `SELECT pg_advisory_unlock($1)`, e.cfg.Key); err != nil {
			log.Warn("leader: unlock failed; closing the session releases the lock", "err", err)
		}
	}
}

// HolderPID returns the backend process ID of the session that holds the advisory lock key
// in the current database, and false if nobody holds it. pg_locks shows a bigint key split
// in two (high half in classid, low half in objid, objsubid = 1); the query puts it back.
func HolderPID(ctx context.Context, pool *pgxpool.Pool, key int64) (int32, bool, error) {
	var pid int32
	err := pool.QueryRow(ctx, `
		SELECT pid FROM pg_locks
		WHERE locktype = 'advisory' AND granted AND objsubid = 1
		  AND database = (SELECT oid FROM pg_database WHERE datname = current_database())
		  AND ((classid::bigint << 32) | objid::bigint) = $1`, key).Scan(&pid)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, false, nil
	}
	return pid, err == nil, err
}
