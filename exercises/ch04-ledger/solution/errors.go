package ledger

import (
	"errors"
	"fmt"
)

// Sentinel errors: fixed values that callers compare with errors.Is.
var (
	ErrAccountNotFound = errors.New("account not found")
	ErrAccountExists   = errors.New("account already exists")
	ErrInvalidAmount   = errors.New("amount must be positive")
)

// InsufficientFundsError carries data the caller may want, so it is a type,
// not a sentinel. Callers extract it with errors.As or errors.AsType.
type InsufficientFundsError struct {
	Account   AccountID
	Balance   Cents
	Requested Cents
}

// Error has a value receiver, so both InsufficientFundsError and
// *InsufficientFundsError satisfy error. The code returns the value form.
func (e InsufficientFundsError) Error() string {
	return fmt.Sprintf("insufficient funds in %s: balance %s, requested %s", e.Account, e.Balance, e.Requested)
}

// Shortfall is how much is missing.
func (e InsufficientFundsError) Shortfall() Cents {
	return e.Requested - e.Balance
}
