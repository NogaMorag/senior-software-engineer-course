// Package replica measures replication lag between a PostgreSQL primary and a streaming
// hot-standby replica, and gives a client read-your-writes on top of that replica.
package replica

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Pair is one primary and one of its streaming replicas.
type Pair struct {
	Primary *pgxpool.Pool
	Replica *pgxpool.Pool
}

// Connect opens a pool to each server.
func Connect(ctx context.Context, primaryURL, replicaURL string) (Pair, error) {
	primary, err := pgxpool.New(ctx, primaryURL)
	if err != nil {
		return Pair{}, fmt.Errorf("primary: %w", err)
	}
	replica, err := pgxpool.New(ctx, replicaURL)
	if err != nil {
		primary.Close()
		return Pair{}, fmt.Errorf("replica: %w", err)
	}
	return Pair{Primary: primary, Replica: replica}, nil
}

// Close closes both pools.
func (p Pair) Close() {
	p.Primary.Close()
	p.Replica.Close()
}

const schema = `
CREATE TABLE IF NOT EXISTS probe_writes (
	id         bigserial PRIMARY KEY,
	note       text NOT NULL,
	written_at timestamptz NOT NULL DEFAULT clock_timestamp()
);
CREATE TABLE IF NOT EXISTS probe_load (
	id      bigserial PRIMARY KEY,
	payload text NOT NULL
)`

// Setup creates the probe tables on the primary (DDL is replicated like any other write)
// and returns once the replica has replayed them.
func (p Pair) Setup(ctx context.Context) error {
	if _, err := p.Primary.Exec(ctx, schema); err != nil {
		return fmt.Errorf("create probe tables: %w", err)
	}
	lsn, err := p.CurrentLSN(ctx)
	if err != nil {
		return err
	}
	return p.WaitForLSN(ctx, lsn, 10*time.Millisecond)
}

// Observation is one probe: a row written on the primary and polled for on the replica.
type Observation struct {
	ID    int64
	Lag   time.Duration // from the primary's commit returning to the first replica read that saw the row
	Polls int           // replica reads made, including the one that saw the row
}

// Probe writes one row on the primary, then reads the replica every interval until the row
// is visible there. The lag it reports includes one replica round trip and up to one
// interval of polling, so it is an upper bound on the true replication delay.
func (p Pair) Probe(ctx context.Context, note string, interval time.Duration) (Observation, error) {
	panic("TODO(day 50): write a row on the primary, then poll the replica until it is visible; see section 17.3")
}

// Summary describes a run of probes.
type Summary struct {
	N                     int
	Min, Median, P95, Max time.Duration
	MissedFirstRead       int // probes whose first replica read did not see the row
}

// Summarize computes nearest-rank percentiles over the observed lags.
func Summarize(obs []Observation) Summary {
	panic("TODO(day 50): sort the lags and take nearest-rank percentiles; see section 17.3")
}
