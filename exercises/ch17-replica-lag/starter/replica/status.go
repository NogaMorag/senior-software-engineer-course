package replica

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// Sender is one row of pg_stat_replication on the primary: one connected standby.
type Sender struct {
	Name           string         // application_name of the standby's walreceiver
	State          string         // "streaming" once it has caught up with the backlog
	SyncState      string         // "async" unless synchronous_standby_names names it
	ReplayLagBytes int64          // WAL the primary has written that the standby has not replayed
	ReplayLag      *time.Duration // nil when the view has nothing recent to report
}

// Senders reads pg_stat_replication on the primary.
func (p Pair) Senders(ctx context.Context) ([]Sender, error) {
	rows, err := p.Primary.Query(ctx, `
		SELECT application_name, state, sync_state,
		       COALESCE(pg_wal_lsn_diff(pg_current_wal_lsn(), replay_lsn), 0)::bigint,
		       EXTRACT(EPOCH FROM replay_lag)::float8
		FROM pg_stat_replication
		ORDER BY application_name`)
	if err != nil {
		return nil, fmt.Errorf("pg_stat_replication: %w", err)
	}
	defer rows.Close()
	var out []Sender
	for rows.Next() {
		var s Sender
		var lagSeconds *float64
		if err := rows.Scan(&s.Name, &s.State, &s.SyncState, &s.ReplayLagBytes, &lagSeconds); err != nil {
			return nil, err
		}
		if lagSeconds != nil {
			d := time.Duration(*lagSeconds * float64(time.Second))
			s.ReplayLag = &d
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// ReplayAge is now() minus pg_last_xact_replay_timestamp() on the replica: how old the last
// replayed transaction is. Under steady writes it approximates lag; on an idle primary it
// keeps growing although nothing is behind. Nil before any transaction has been replayed.
func (p Pair) ReplayAge(ctx context.Context) (*time.Duration, error) {
	var seconds *float64
	err := p.Replica.QueryRow(ctx,
		`SELECT EXTRACT(EPOCH FROM now() - pg_last_xact_replay_timestamp())::float8`).Scan(&seconds)
	if err != nil || seconds == nil {
		return nil, err
	}
	d := time.Duration(*seconds * float64(time.Second))
	return &d, nil
}

// SetApplyDelay sets recovery_min_apply_delay on the replica (zero resets it) and waits
// until the server reports the new value. The replica then applies each commit only once
// its own clock is at least d past the commit time recorded on the primary.
// It needs a superuser connection to the replica; ALTER SYSTEM works on a standby because it
// edits postgresql.auto.conf on that server and writes no WAL.
func (p Pair) SetApplyDelay(ctx context.Context, d time.Duration) error {
	stmt := "ALTER SYSTEM RESET recovery_min_apply_delay"
	if d > 0 {
		stmt = fmt.Sprintf("ALTER SYSTEM SET recovery_min_apply_delay = %d", d.Milliseconds())
	}
	if _, err := p.Replica.Exec(ctx, stmt); err != nil {
		return fmt.Errorf("%s: %w", stmt, err)
	}
	if _, err := p.Replica.Exec(ctx, "SELECT pg_reload_conf()"); err != nil {
		return fmt.Errorf("reload: %w", err)
	}
	want := strconv.FormatInt(d.Milliseconds(), 10)
	for { // the reload is asynchronous: poll until the setting is visible
		var got string
		err := p.Replica.QueryRow(ctx,
			`SELECT setting FROM pg_settings WHERE name = 'recovery_min_apply_delay'`).Scan(&got)
		if err != nil {
			return err
		}
		if got == want {
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("recovery_min_apply_delay is %s, want %s: %w", got, want, ctx.Err())
		case <-time.After(20 * time.Millisecond):
		}
	}
}

// Load runs writers goroutines that each insert batches of 1 KB rows on the primary until
// ctx is done, generating WAL for the replica to replay. It returns the rows written and the
// first error that was not caused by ctx ending.
func (p Pair) Load(ctx context.Context, writers int) (int64, error) {
	var total atomic.Int64
	errs := make(chan error, writers)
	var wg sync.WaitGroup
	for range writers {
		wg.Go(func() {
			for ctx.Err() == nil {
				tag, err := p.Primary.Exec(ctx,
					`INSERT INTO probe_load (payload) SELECT repeat('x', 1000) FROM generate_series(1, 500)`)
				if err != nil {
					if ctx.Err() == nil {
						errs <- err
					}
					return
				}
				total.Add(tag.RowsAffected())
			}
		})
	}
	wg.Wait()
	close(errs)
	return total.Load(), <-errs // a closed, empty channel yields nil
}
