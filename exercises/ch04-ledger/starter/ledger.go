// Package ledger keeps accounts and their transactions behind a Store interface.
package ledger

import (
	"fmt"
	"time"
)

// Ledger holds the rules; a Store holds the data. Ledger depends only on the
// interface, so the same rules run on memory, on a file, and later on PostgreSQL.
type Ledger struct {
	store Store
	now   func() time.Time
}

// New returns a Ledger over store. A nil now means time.Now; tests pass a fixed clock.
func New(store Store, now func() time.Time) *Ledger {
	if now == nil {
		now = time.Now
	}
	return &Ledger{store: store, now: now}
}

// Open creates an account with a zero balance.
func (l *Ledger) Open(id AccountID, owner string) (Account, error) {
	a := Account{ID: id, Owner: owner, Opened: l.now().UTC()}
	if err := l.store.CreateAccount(a); err != nil {
		return Account{}, fmt.Errorf("open %s: %w", id, err)
	}
	return a, nil
}

// Deposit adds amount to the account.
func (l *Ledger) Deposit(id AccountID, amount Cents, memo string) (Account, error) {
	return l.post(id, Deposit, amount, memo)
}

// Withdraw takes amount from the account, or fails with InsufficientFundsError.
func (l *Ledger) Withdraw(id AccountID, amount Cents, memo string) (Account, error) {
	return l.post(id, Withdrawal, amount, memo)
}

func (l *Ledger) post(id AccountID, kind Kind, amount Cents, memo string) (Account, error) {
	if amount <= 0 {
		return Account{}, fmt.Errorf("%v of %s on %s: %w", kind, amount, id, ErrInvalidAmount)
	}
	tx := Transaction{Account: id, Kind: kind, Amount: amount, At: l.now().UTC(), Memo: memo}
	_, a, err := l.store.Record(tx)
	if err != nil {
		// TODO(day 6): wrap err with context in the same shape as the amount check
		// above ("withdrawal of 5.00 on ghost: ..."), using %w so the cause survives.
		return Account{}, err
	}
	return a, nil
}

// Balance returns the current balance.
func (l *Ledger) Balance(id AccountID) (Cents, error) {
	a, err := l.store.Account(id)
	if err != nil {
		return 0, fmt.Errorf("balance of %s: %w", id, err)
	}
	return a.Balance, nil
}

// History returns one page of the account's transactions, oldest first.
func (l *Ledger) History(id AccountID, cursor, limit int) (Page[Transaction], error) {
	txs, err := l.store.Transactions(id)
	if err != nil {
		return Page[Transaction]{}, fmt.Errorf("history of %s: %w", id, err)
	}
	page, err := Paginate(txs, cursor, limit)
	if err != nil {
		return Page[Transaction]{}, fmt.Errorf("history of %s: %w", id, err)
	}
	return page, nil
}
