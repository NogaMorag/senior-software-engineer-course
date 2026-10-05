package ledger

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
)

// FileStore is a MemStore that saves itself to a JSON file after every change.
// It embeds MemStore: Account and Transactions are promoted unchanged, while
// CreateAccount and Record are redefined here to call the embedded version and then save.
type FileStore struct {
	MemStore
	path string
}

var _ Store = (*FileStore)(nil)

// snapshot is the on-disk format.
type snapshot struct {
	LastID       int64         `json:"last_id"`
	Accounts     []Account     `json:"accounts"`
	Transactions []Transaction `json:"transactions"`
}

// OpenFileStore loads path, or starts empty if the file does not exist yet.
func OpenFileStore(path string) (*FileStore, error) {
	s := &FileStore{path: path}
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return nil, fmt.Errorf("open ledger file: %w", err)
	}
	var snap snapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return nil, fmt.Errorf("load %s: %w", path, err)
	}
	s.lazyInit()
	s.lastID = snap.LastID
	for _, a := range snap.Accounts {
		s.accounts[a.ID] = a
	}
	for _, tx := range snap.Transactions {
		s.txs[tx.Account] = append(s.txs[tx.Account], tx)
	}
	return s, nil
}

// CreateAccount shadows MemStore.CreateAccount. If saving fails, memory is
// ahead of the file; the error tells the caller to stop using this store.
func (s *FileStore) CreateAccount(a Account) error {
	if err := s.MemStore.CreateAccount(a); err != nil {
		return err
	}
	return s.save()
}

// Record shadows MemStore.Record in the same way.
func (s *FileStore) Record(tx Transaction) (Transaction, Account, error) {
	tx, a, err := s.MemStore.Record(tx)
	if err != nil {
		return Transaction{}, Account{}, err
	}
	return tx, a, s.save()
}

// save writes a sorted snapshot to a temporary file and renames it over the
// real one, so a crash never leaves a half-written ledger.
func (s *FileStore) save() error {
	snap := snapshot{LastID: s.lastID}
	for _, a := range s.accounts {
		snap.Accounts = append(snap.Accounts, a)
	}
	for _, txs := range s.txs {
		snap.Transactions = append(snap.Transactions, txs...)
	}
	// Map iteration order is random; sort so the file is identical on every run.
	slices.SortFunc(snap.Accounts, func(a, b Account) int { return cmp.Compare(a.ID, b.ID) })
	slices.SortFunc(snap.Transactions, func(a, b Transaction) int { return cmp.Compare(a.ID, b.ID) })

	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return fmt.Errorf("encode ledger: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(s.path), ".ledger-*.tmp")
	if err != nil {
		return fmt.Errorf("save ledger: %w", err)
	}
	defer os.Remove(tmp.Name()) // no-op after a successful rename
	if _, err := tmp.Write(append(data, '\n')); err != nil {
		tmp.Close()
		return fmt.Errorf("save ledger: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("save ledger: %w", err)
	}
	if err := os.Rename(tmp.Name(), s.path); err != nil {
		return fmt.Errorf("save ledger: %w", err)
	}
	return nil
}
