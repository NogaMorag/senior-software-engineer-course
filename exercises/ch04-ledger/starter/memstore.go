package ledger

import "errors"

// MemStore keeps everything in maps. Its zero value is ready to use.
// It is not safe for concurrent use; chapter 5 adds the locking.
type MemStore struct {
	accounts map[AccountID]Account
	txs      map[AccountID][]Transaction
	lastID   int64
}

// Compile-time check that *MemStore implements Store.
var _ Store = (*MemStore)(nil)

// lazyInit allocates the maps on first write. Reading a nil map is fine; writing panics.
func (s *MemStore) lazyInit() {
	if s.accounts == nil {
		s.accounts = make(map[AccountID]Account)
		s.txs = make(map[AccountID][]Transaction)
	}
}

func (s *MemStore) CreateAccount(a Account) error {
	// TODO(day 6): call s.lazyInit(), return ErrAccountExists if a.ID is taken, else store a.
	return errors.New("TODO: CreateAccount")
}

func (s *MemStore) Account(id AccountID) (Account, error) {
	a, ok := s.accounts[id] // works on a nil map too
	if !ok {
		return Account{}, ErrAccountNotFound
	}
	return a, nil
}

func (s *MemStore) Record(tx Transaction) (Transaction, Account, error) {
	// TODO(day 6): look the account up (s.Account), run apply, and only if both
	// succeed: give tx the next ID (s.lastID), save the account and append tx.
	return Transaction{}, Account{}, errors.New("TODO: Record")
}

// Transactions returns a copy, oldest first: the caller may append to it
// without corrupting the store (the aliasing trap from chapter 3).
func (s *MemStore) Transactions(id AccountID) ([]Transaction, error) {
	if _, err := s.Account(id); err != nil {
		return nil, err
	}
	// TODO(day 6): return a copy (slices.Clone), not the stored slice itself.
	return s.txs[id], nil
}
