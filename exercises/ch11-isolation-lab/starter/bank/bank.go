// Package bank is the lost-update and deadlock lab: accounts with a balance.
package bank

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrInsufficientFunds is returned when a withdrawal would make the balance negative.
var ErrInsufficientFunds = errors.New("bank: insufficient funds")

// ReadBalance reads the balance with a plain SELECT: no lock is taken.
func ReadBalance(ctx context.Context, tx pgx.Tx, id int64) (int64, error) {
	// TODO(day 26): a plain SELECT balance ... WHERE id = $1, scanned into an int64. See chapter 11, section 11.3.
	panic("TODO: ReadBalance, see chapter 11, section 11.3")
}

// ReadBalanceForUpdate reads the balance and locks the row until the transaction ends.
// A second transaction that asks for the same lock waits here.
func ReadBalanceForUpdate(ctx context.Context, tx pgx.Tx, id int64) (int64, error) {
	// TODO(day 27): like ReadBalance, with FOR UPDATE at the end of the SELECT. See chapter 11, section 11.4.
	panic("TODO: ReadBalanceForUpdate, see chapter 11, section 11.4")
}

// SetBalance overwrites the balance with a value computed in Go.
func SetBalance(ctx context.Context, tx pgx.Tx, id, balance int64) error {
	// TODO(day 26): UPDATE accounts SET balance = $2 WHERE id = $1. See chapter 11, section 11.3.
	panic("TODO: SetBalance, see chapter 11, section 11.3")
}

// AddToBalance changes the balance by delta inside the database.
func AddToBalance(ctx context.Context, tx pgx.Tx, id, delta int64) error {
	if _, err := tx.Exec(ctx, `UPDATE accounts SET balance = balance + $2 WHERE id = $1`, id, delta); err != nil {
		return fmt.Errorf("add %d to %d: %w", delta, id, err)
	}
	return nil
}

// Withdraw is the bug this chapter starts from: read, compute in Go, write back.
// Two concurrent calls can both read the same balance, and one withdrawal is lost.
func Withdraw(ctx context.Context, pool *pgxpool.Pool, id, amount int64) error {
	return pgx.BeginFunc(ctx, pool, func(tx pgx.Tx) error {
		balance, err := ReadBalance(ctx, tx, id)
		if err != nil {
			return err
		}
		if balance < amount {
			return ErrInsufficientFunds
		}
		return SetBalance(ctx, tx, id, balance-amount)
	})
}

// WithdrawLocked is Withdraw with the read turned into SELECT ... FOR UPDATE: the second
// caller waits for the first to commit and then reads the new balance.
func WithdrawLocked(ctx context.Context, pool *pgxpool.Pool, id, amount int64) error {
	// TODO(day 27): Withdraw, but read with ReadBalanceForUpdate. See chapter 11, section 11.4.
	panic("TODO: WithdrawLocked, see chapter 11, section 11.4")
}

// WithdrawAtomic needs no explicit lock: the check and the change are one UPDATE, and
// PostgreSQL re-checks the WHERE clause on the newest row version if it had to wait.
func WithdrawAtomic(ctx context.Context, pool *pgxpool.Pool, id, amount int64) error {
	tag, err := pool.Exec(ctx, `
		UPDATE accounts SET balance = balance - $2
		WHERE id = $1 AND balance >= $2`, id, amount)
	if err != nil {
		return fmt.Errorf("withdraw from %d: %w", id, err)
	}
	if tag.RowsAffected() == 0 {
		return ErrInsufficientFunds
	}
	return nil
}

// TransferNaive debits from, then credits to. Two opposite transfers that run at the
// same time lock the two rows in opposite orders and can deadlock (SQLSTATE 40P01).
func TransferNaive(ctx context.Context, pool *pgxpool.Pool, from, to, amount int64) error {
	return pgx.BeginFunc(ctx, pool, func(tx pgx.Tx) error {
		if err := AddToBalance(ctx, tx, from, -amount); err != nil {
			return err
		}
		return AddToBalance(ctx, tx, to, amount)
	})
}

// Transfer locks both rows in ascending id order before changing either, so every
// transfer takes its locks in the same global order and no cycle can form.
func Transfer(ctx context.Context, pool *pgxpool.Pool, from, to, amount int64) error {
	// TODO(day 27): lock min(from, to) then max(from, to) with ReadBalanceForUpdate, then the two AddToBalance calls. See chapter 11, section 11.5.
	panic("TODO: Transfer, see chapter 11, section 11.5")
}
