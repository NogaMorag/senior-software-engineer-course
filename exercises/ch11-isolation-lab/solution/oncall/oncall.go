// Package oncall is the write-skew lab: every shift must keep at least one doctor on call.
package oncall

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"example.com/ch11-isolation-lab/db"
)

// ErrLastOnCall is returned when the doctor is the only one left on call.
var ErrLastOnCall = errors.New("oncall: cannot go off call, nobody else is on call")

// CountOnCall counts the doctors on call for the shift.
func CountOnCall(ctx context.Context, tx pgx.Tx, shift int) (int, error) {
	var n int
	err := tx.QueryRow(ctx, `SELECT count(*) FROM oncall WHERE shift_id = $1 AND on_call`, shift).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("count on call for shift %d: %w", shift, err)
	}
	return n, nil
}

// GoOffCall checks the rule, then takes the doctor off call. The check reads every row
// of the shift but the write touches only the doctor's own row: under REPEATABLE READ two
// doctors can both pass the check and both leave (write skew).
func GoOffCall(ctx context.Context, tx pgx.Tx, shift int, doctor string) error {
	n, err := CountOnCall(ctx, tx, shift)
	if err != nil {
		return err
	}
	if n < 2 {
		return ErrLastOnCall
	}
	_, err = tx.Exec(ctx, `UPDATE oncall SET on_call = false WHERE shift_id = $1 AND doctor = $2`, shift, doctor)
	if err != nil {
		return fmt.Errorf("take %s off call: %w", doctor, err)
	}
	return nil
}

// GoOffCallSerializable runs GoOffCall as a SERIALIZABLE transaction and retries it on
// serialization failure. Of two doctors who leave at the same moment, one succeeds and
// the other gets ErrLastOnCall on the retry, when the count already sees the change.
func GoOffCallSerializable(ctx context.Context, pool *pgxpool.Pool, shift int, doctor string) error {
	_, err := db.RetrySerializable(ctx, pool, func(ctx context.Context, tx pgx.Tx) (struct{}, error) {
		return struct{}{}, GoOffCall(ctx, tx, shift, doctor)
	})
	return err
}
