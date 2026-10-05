package ledger

import (
	"fmt"
	"time"
)

// AccountID names an account, for example "alice-main".
type AccountID string

// Account is the current state of one account.
type Account struct {
	ID      AccountID `json:"id"`
	Owner   string    `json:"owner"`
	Balance Cents     `json:"balance"`
	Opened  time.Time `json:"opened"`
}

// Kind says which way money moved.
type Kind int

const (
	Deposit Kind = iota + 1 // start at 1 so the zero Kind is visibly invalid
	Withdrawal
)

func (k Kind) String() string {
	switch k {
	case Deposit:
		return "deposit"
	case Withdrawal:
		return "withdrawal"
	default:
		return fmt.Sprintf("Kind(%d)", int(k))
	}
}

// Transaction is one movement of money. Amount is always positive; Kind gives the sign.
type Transaction struct {
	ID      int64     `json:"id"`
	Account AccountID `json:"account"`
	Kind    Kind      `json:"kind"`
	Amount  Cents     `json:"amount"`
	At      time.Time `json:"at"`
	Memo    string    `json:"memo,omitzero"`
}

// Store is everything the ledger needs from storage. Any type with these
// four methods is a Store; it never has to say so.
type Store interface {
	CreateAccount(a Account) error
	Account(id AccountID) (Account, error)
	// Record applies tx to its account, stores it with a fresh ID, and returns
	// the stored transaction and the updated account.
	Record(tx Transaction) (Transaction, Account, error)
	Transactions(id AccountID) ([]Transaction, error)
}

// apply is the one business rule both stores share: what a transaction does to a balance.
func apply(a Account, tx Transaction) (Account, error) {
	switch tx.Kind {
	case Deposit:
		a.Balance += tx.Amount
	case Withdrawal:
		if tx.Amount > a.Balance {
			return a, InsufficientFundsError{Account: a.ID, Balance: a.Balance, Requested: tx.Amount}
		}
		a.Balance -= tx.Amount
	default:
		return a, fmt.Errorf("unknown transaction kind %v", tx.Kind)
	}
	return a, nil
}
