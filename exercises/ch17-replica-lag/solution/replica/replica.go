// Package replica measures replication lag between a PostgreSQL primary and a streaming
// hot-standby replica, and gives a client read-your-writes on top of that replica.
package replica

import (
	"context"
	"fmt"
	"math"
	"slices"
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
	var id int64
	err := p.Primary.QueryRow(ctx,
		`INSERT INTO probe_writes (note) VALUES ($1) RETURNING id`, note).Scan(&id)
	if err != nil {
		return Observation{}, fmt.Errorf("write on primary: %w", err)
	}
	committed := time.Now() // autocommit: the row is durable on the primary from here on
	for polls := 1; ; polls++ {
		var seen bool
		err := p.Replica.QueryRow(ctx,
			`SELECT EXISTS (SELECT 1 FROM probe_writes WHERE id = $1)`, id).Scan(&seen)
		if err != nil {
			return Observation{}, fmt.Errorf("read on replica: %w", err)
		}
		if seen {
			return Observation{ID: id, Lag: time.Since(committed), Polls: polls}, nil
		}
		select {
		case <-ctx.Done():
			return Observation{}, fmt.Errorf("row %d not visible on the replica after %d polls: %w", id, polls, ctx.Err())
		case <-time.After(interval):
		}
	}
}

// Summary describes a run of probes.
type Summary struct {
	N                     int
	Min, Median, P95, Max time.Duration
	MissedFirstRead       int // probes whose first replica read did not see the row
}

// Summarize computes nearest-rank percentiles over the observed lags.
func Summarize(obs []Observation) Summary {
	if len(obs) == 0 {
		return Summary{}
	}
	lags := make([]time.Duration, len(obs))
	s := Summary{N: len(obs)}
	for i, o := range obs {
		lags[i] = o.Lag
		if o.Polls > 1 {
			s.MissedFirstRead++
		}
	}
	slices.Sort(lags)
	rank := func(q float64) time.Duration { // nearest rank: the smallest value with at least q of the data at or below it
		i := int(math.Ceil(q*float64(len(lags)))) - 1
		return lags[max(0, i)]
	}
	s.Min, s.Max = lags[0], lags[len(lags)-1]
	s.Median, s.P95 = rank(0.5), rank(0.95)
	return s
}
