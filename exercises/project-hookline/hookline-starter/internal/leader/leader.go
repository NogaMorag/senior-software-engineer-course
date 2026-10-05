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

	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Schema is the table that numbers the terms. Hookline creates it in migration 00006; EnsureSchema
// stays for tests and tools that run without the migrations.
const Schema = `CREATE TABLE IF NOT EXISTS leader_terms (
	lock_key bigint PRIMARY KEY,
	term     bigint NOT NULL
)`

// EnsureSchema creates the terms table if it is missing.
func EnsureSchema(ctx context.Context, pool *pgxpool.Pool) error {
	panic("TODO: see chapter 38")

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
	panic("TODO: see chapter 38")

}

// Run campaigns until ctx is done. Whenever this process wins, Run calls lead with the new
// term and a context that is cancelled as soon as leadership is in doubt: the heartbeat
// failed, or ctx ended. When lead returns, for any reason, this process steps down and
// campaigns again after Retry. Run returns ctx's error once ctx is done and lead has
// returned; the lock is released before Run returns.
func (e *Elector) Run(ctx context.Context, lead func(ctx context.Context, term int64) error) error {
	panic("TODO: see chapter 38")

}

// campaign makes one attempt. It reports whether it won; if it did, it returns only after
// the term is over.
func (e *Elector) campaign(ctx context.Context, lead func(context.Context, int64) error) (bool, error) {
	panic("TODO: see chapter 38")

}

// serve runs one term: lead in its own goroutine, the heartbeat here. It returns when ctx
// ends, lead returns, or the heartbeat fails, and only after lead has returned.
func (e *Elector) serve(ctx context.Context, conn *pgx.Conn, term int64,
	lead func(context.Context, int64) error, log *slog.Logger) {
	panic("TODO: see chapter 38")

}

// HolderPID returns the backend process ID of the session that holds the advisory lock key
// in the current database, and false if nobody holds it. pg_locks shows a bigint key split
// in two (high half in classid, low half in objid, objsubid = 1); the query puts it back.
func HolderPID(ctx context.Context, pool *pgxpool.Pool, key int64) (int32, bool, error) {
	panic("TODO: see chapter 38")

}
