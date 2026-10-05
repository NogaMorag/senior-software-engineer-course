package replica

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// LSN is a position in the primary's write-ahead log, kept in PostgreSQL's text form
// ("0/3000148"). Comparisons happen on the server, as pg_lsn, never as strings.
type LSN string

// Source names the server that answered a read.
type Source string

const (
	FromPrimary Source = "primary"
	FromReplica Source = "replica"
)

// ErrNotFound means the row is not there on the server that was asked.
var ErrNotFound = errors.New("not found")

// CurrentLSN returns the primary's current WAL write position. Taken right after a commit,
// it is at or past that commit's record, so a replica that has replayed it has the write.
func (p Pair) CurrentLSN(ctx context.Context) (LSN, error) {
	var lsn string
	if err := p.Primary.QueryRow(ctx, `SELECT pg_current_wal_lsn()::text`).Scan(&lsn); err != nil {
		return "", fmt.Errorf("pg_current_wal_lsn: %w", err)
	}
	return LSN(lsn), nil
}

// CaughtUp reports whether the replica has replayed WAL up to lsn.
func (p Pair) CaughtUp(ctx context.Context, lsn LSN) (bool, error) {
	var ok bool
	err := p.Replica.QueryRow(ctx,
		`SELECT COALESCE(pg_last_wal_replay_lsn() >= $1::pg_lsn, false)`, string(lsn)).Scan(&ok)
	if err != nil {
		return false, fmt.Errorf("pg_last_wal_replay_lsn: %w", err)
	}
	return ok, nil
}

// WaitForLSN polls the replica every interval until it has replayed lsn or ctx ends.
func (p Pair) WaitForLSN(ctx context.Context, lsn LSN, interval time.Duration) error {
	for {
		ok, err := p.CaughtUp(ctx, lsn)
		if err != nil || ok {
			return err
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("replica has not replayed %s: %w", lsn, ctx.Err())
		case <-time.After(interval):
		}
	}
}

// WriteNote inserts a note on the primary and returns its id and the LSN token the client
// keeps (in its session, a cookie, or the response) for its next read.
func (p Pair) WriteNote(ctx context.Context, note string) (int64, LSN, error) {
	var id int64
	err := p.Primary.QueryRow(ctx,
		`INSERT INTO probe_writes (note) VALUES ($1) RETURNING id`, note).Scan(&id)
	if err != nil {
		return 0, "", fmt.Errorf("write note: %w", err)
	}
	lsn, err := p.CurrentLSN(ctx)
	return id, lsn, err
}

// ReadNoteFromReplica is the naive read: whatever the replica has right now.
func (p Pair) ReadNoteFromReplica(ctx context.Context, id int64) (string, error) {
	return readNote(ctx, p.Replica, id)
}

// ReadNote gives read-your-writes: it reads from the replica once the replica has replayed
// the client's token, waiting at most maxWait; after that it reads from the primary, which
// always has the write. An empty token means the client has written nothing it must see.
func (p Pair) ReadNote(ctx context.Context, id int64, token LSN, maxWait time.Duration) (string, Source, error) {
	panic("TODO(day 50): wait for the token on the replica, at most maxWait, else read the primary; see section 17.4")
}

type querier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func readNote(ctx context.Context, q querier, id int64) (string, error) {
	var note string
	err := q.QueryRow(ctx, `SELECT note FROM probe_writes WHERE id = $1`, id).Scan(&note)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	return note, err
}
