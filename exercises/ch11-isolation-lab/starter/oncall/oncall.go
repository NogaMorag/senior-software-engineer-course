// Package oncall is the write-skew lab: every shift must keep at least one doctor on call.
package oncall

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrLastOnCall is returned when the doctor is the only one left on call.
var ErrLastOnCall = errors.New("oncall: cannot go off call, nobody else is on call")

// CountOnCall counts the doctors on call for the shift.
func CountOnCall(ctx context.Context, tx pgx.Tx, shift int) (int, error) {
	// TODO(day 26): SELECT count(*) of on_call rows for the shift. See chapter 11, section 11.3.
	panic("TODO: CountOnCall, see chapter 11, section 11.3")
}

// GoOffCall checks the rule, then takes the doctor off call. The check reads every row
// of the shift but the write touches only the doctor's own row: under REPEATABLE READ two
// doctors can both pass the check and both leave (write skew).
func GoOffCall(ctx context.Context, tx pgx.Tx, shift int, doctor string) error {
	// TODO(day 26): count; fewer than 2 means ErrLastOnCall; otherwise UPDATE only this doctor's row. See chapter 11, section 11.3.
	panic("TODO: GoOffCall, see chapter 11, section 11.3")
}

// GoOffCallSerializable runs GoOffCall as a SERIALIZABLE transaction and retries it on
// serialization failure. Of two doctors who leave at the same moment, one succeeds and
// the other gets ErrLastOnCall on the retry, when the count already sees the change.
func GoOffCallSerializable(ctx context.Context, pool *pgxpool.Pool, shift int, doctor string) error {
	// TODO(day 28): wrap GoOffCall in db.RetrySerializable. See chapter 11, section 11.8.
	panic("TODO: GoOffCallSerializable, see chapter 11, section 11.8")
}
