package ledger

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
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
	// TODO(day 7): after the embedded call succeeds, call s.save().
	return s.MemStore.CreateAccount(a)
}

// Record shadows MemStore.Record in the same way.
func (s *FileStore) Record(tx Transaction) (Transaction, Account, error) {
	// TODO(day 7): same as CreateAccount: save after a successful embedded call.
	return s.MemStore.Record(tx)
}

// save writes a sorted snapshot to a temporary file and renames it over the
// real one, so a crash never leaves a half-written ledger.
func (s *FileStore) save() error {
	// TODO(day 7):
	//   1. build a snapshot from s.lastID, s.accounts and s.txs;
	//   2. sort both slices by ID (slices.SortFunc + cmp.Compare): map order is random;
	//   3. json.MarshalIndent it;
	//   4. write to os.CreateTemp(filepath.Dir(s.path), ".ledger-*.tmp"), close it,
	//      and os.Rename it over s.path; wrap every error with fmt.Errorf("save ledger: %w", err).
	return errors.New("TODO: save")
}
