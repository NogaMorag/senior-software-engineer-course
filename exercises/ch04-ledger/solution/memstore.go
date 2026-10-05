package ledger

import "slices"

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
	s.lazyInit()
	if _, ok := s.accounts[a.ID]; ok {
		return ErrAccountExists
	}
	s.accounts[a.ID] = a
	return nil
}

func (s *MemStore) Account(id AccountID) (Account, error) {
	a, ok := s.accounts[id] // works on a nil map too
	if !ok {
		return Account{}, ErrAccountNotFound
	}
	return a, nil
}

func (s *MemStore) Record(tx Transaction) (Transaction, Account, error) {
	a, err := s.Account(tx.Account)
	if err != nil {
		return Transaction{}, Account{}, err
	}
	a, err = apply(a, tx)
	if err != nil {
		return Transaction{}, Account{}, err
	}
	s.lazyInit()
	s.lastID++
	tx.ID = s.lastID
	s.accounts[a.ID] = a
	s.txs[a.ID] = append(s.txs[a.ID], tx)
	return tx, a, nil
}

// Transactions returns a copy, oldest first: the caller may append to it
// without corrupting the store (the aliasing trap from chapter 3).
func (s *MemStore) Transactions(id AccountID) ([]Transaction, error) {
	if _, err := s.Account(id); err != nil {
		return nil, err
	}
	return slices.Clone(s.txs[id]), nil
}
