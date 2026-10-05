package ledger

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSentinelThroughWrapping(t *testing.T) {
	l := New(&MemStore{}, fixedClock())
	_, err := l.Withdraw("ghost", 5_00, "")

	if err == ErrAccountNotFound {
		t.Fatal("err == ErrAccountNotFound: the ledger should add context, not return the bare sentinel")
	}
	if !errors.Is(err, ErrAccountNotFound) {
		t.Fatalf("errors.Is(%v, ErrAccountNotFound) = false", err)
	}
	if want := "withdrawal of 5.00 on ghost: account not found"; err.Error() != want {
		t.Errorf("message = %q, want %q", err.Error(), want)
	}
}

func TestTypedErrorThroughWrapping(t *testing.T) {
	l := New(&MemStore{}, fixedClock())
	mustOpen(t, l, "alice")
	_, err := l.Withdraw("alice", 1_00, "")

	// The classic form: a target variable and a pointer to it.
	var target InsufficientFundsError
	if !errors.As(err, &target) {
		t.Fatalf("errors.As found no InsufficientFundsError in %v", err)
	}
	// The generic form (Go 1.26 and later): no target variable.
	got, ok := errors.AsType[InsufficientFundsError](err)
	if !ok || got != target {
		t.Fatalf("errors.AsType = %+v, %v; want %+v, true", got, ok, target)
	}
	if got.Account != "alice" || got.Balance != 0 || got.Requested != 1_00 {
		t.Errorf("fields = %+v", got)
	}
	// One more layer of wrapping, by a caller, still works.
	outer := fmt.Errorf("monthly run: %w", err)
	if _, ok := errors.AsType[InsufficientFundsError](outer); !ok {
		t.Error("errors.AsType lost the typed error after a second wrap")
	}
}

func TestInvalidAmount(t *testing.T) {
	l := New(&MemStore{}, fixedClock())
	mustOpen(t, l, "alice")
	for _, amount := range []Cents{0, -1_00} {
		if _, err := l.Deposit("alice", amount, ""); !errors.Is(err, ErrInvalidAmount) {
			t.Errorf("Deposit(%v) error = %v, want ErrInvalidAmount", amount, err)
		}
	}
}

func TestCorruptFileKeepsTheJSONError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ledger.json")
	if err := os.WriteFile(path, []byte(`{"accounts": [`), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := OpenFileStore(path)
	if err == nil {
		t.Fatal("OpenFileStore on a truncated file succeeded")
	}
	if !strings.Contains(err.Error(), path) {
		t.Errorf("error %q does not name the file", err)
	}
	var syntax *json.SyntaxError
	if !errors.As(err, &syntax) {
		t.Errorf("error %v (%T) does not wrap a *json.SyntaxError", err, errors.Unwrap(err))
	}
}
