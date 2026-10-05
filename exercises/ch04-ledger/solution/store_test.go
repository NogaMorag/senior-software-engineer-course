package ledger

import (
	"errors"
	"path/filepath"
	"testing"
	"time"
)

// fixedClock returns a clock that always says 2026-10-05 09:00 UTC.
func fixedClock() func() time.Time {
	t := time.Date(2026, time.October, 5, 9, 0, 0, 0, time.UTC)
	return func() time.Time { return t }
}

// testStore is the shared suite: every Store implementation must pass it.
// newStore returns a fresh, empty store for each subtest.
func testStore(t *testing.T, newStore func(t *testing.T) Store) {
	t.Run("open and read back", func(t *testing.T) {
		l := New(newStore(t), fixedClock())
		if _, err := l.Open("alice", "Alice"); err != nil {
			t.Fatal(err)
		}
		b, err := l.Balance("alice")
		if err != nil || b != 0 {
			t.Fatalf("Balance = %v, %v; want 0, nil", b, err)
		}
	})

	t.Run("open twice is ErrAccountExists", func(t *testing.T) {
		l := New(newStore(t), fixedClock())
		if _, err := l.Open("alice", "Alice"); err != nil {
			t.Fatal(err)
		}
		_, err := l.Open("alice", "Alice again")
		if !errors.Is(err, ErrAccountExists) {
			t.Fatalf("second Open error = %v, want ErrAccountExists", err)
		}
	})

	t.Run("deposit and withdraw", func(t *testing.T) {
		l := New(newStore(t), fixedClock())
		mustOpen(t, l, "alice")
		if _, err := l.Deposit("alice", 120_00, "משכורת"); err != nil {
			t.Fatal(err)
		}
		a, err := l.Withdraw("alice", 45_50, "groceries")
		if err != nil {
			t.Fatal(err)
		}
		if a.Balance != 74_50 {
			t.Fatalf("balance = %v, want 74.50", a.Balance)
		}
	})

	t.Run("missing account is ErrAccountNotFound", func(t *testing.T) {
		l := New(newStore(t), fixedClock())
		_, err := l.Deposit("nobody", 1_00, "")
		if !errors.Is(err, ErrAccountNotFound) {
			t.Fatalf("error = %v, want ErrAccountNotFound", err)
		}
		if _, err := l.History("nobody", 0, 10); !errors.Is(err, ErrAccountNotFound) {
			t.Fatalf("History error = %v, want ErrAccountNotFound", err)
		}
	})

	t.Run("overdraft is InsufficientFundsError and changes nothing", func(t *testing.T) {
		l := New(newStore(t), fixedClock())
		mustOpen(t, l, "alice")
		if _, err := l.Deposit("alice", 10_00, ""); err != nil {
			t.Fatal(err)
		}
		_, err := l.Withdraw("alice", 25_00, "")
		var insufficient InsufficientFundsError
		if !errors.As(err, &insufficient) {
			t.Fatalf("error = %v, want InsufficientFundsError", err)
		}
		if insufficient.Shortfall() != 15_00 {
			t.Errorf("Shortfall = %v, want 15.00", insufficient.Shortfall())
		}
		if b, _ := l.Balance("alice"); b != 10_00 {
			t.Errorf("balance after failed withdrawal = %v, want 10.00", b)
		}
		page, _ := l.History("alice", 0, 10)
		if len(page.Items) != 1 {
			t.Errorf("history has %d transactions, want 1 (the failed one must not be stored)", len(page.Items))
		}
	})

	t.Run("history pages in order", func(t *testing.T) {
		l := New(newStore(t), fixedClock())
		mustOpen(t, l, "alice")
		for i := 1; i <= 5; i++ {
			if _, err := l.Deposit("alice", Cents(i*100), ""); err != nil {
				t.Fatal(err)
			}
		}
		var seen []Cents
		cursor := 0
		for {
			page, err := l.History("alice", cursor, 2)
			if err != nil {
				t.Fatal(err)
			}
			for _, tx := range page.Items {
				seen = append(seen, tx.Amount)
			}
			if !page.HasMore {
				break
			}
			cursor = page.Next
		}
		if len(seen) != 5 || seen[0] != 1_00 || seen[4] != 5_00 {
			t.Fatalf("amounts across pages = %v, want 1.00 .. 5.00 in order", seen)
		}
	})

	t.Run("returned history cannot corrupt the store", func(t *testing.T) {
		s := newStore(t)
		l := New(s, fixedClock())
		mustOpen(t, l, "alice")
		if _, err := l.Deposit("alice", 1_00, ""); err != nil {
			t.Fatal(err)
		}
		txs, _ := s.Transactions("alice")
		txs[0].Amount = 999_00 // the caller scribbles on its copy
		again, _ := s.Transactions("alice")
		if again[0].Amount != 1_00 {
			t.Fatalf("store changed through a returned slice: %v", again[0].Amount)
		}
	})
}

func mustOpen(t *testing.T, l *Ledger, id AccountID) {
	t.Helper()
	if _, err := l.Open(id, string(id)); err != nil {
		t.Fatal(err)
	}
}

func TestMemStore(t *testing.T) {
	testStore(t, func(t *testing.T) Store { return &MemStore{} })
}

func TestFileStore(t *testing.T) {
	testStore(t, func(t *testing.T) Store {
		s, err := OpenFileStore(filepath.Join(t.TempDir(), "ledger.json"))
		if err != nil {
			t.Fatal(err)
		}
		return s
	})
}

func TestFileStoreSurvivesReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ledger.json")
	s, err := OpenFileStore(path)
	if err != nil {
		t.Fatal(err)
	}
	l := New(s, fixedClock())
	mustOpen(t, l, "alice")
	if _, err := l.Deposit("alice", 30_00, "מתנה"); err != nil {
		t.Fatal(err)
	}

	reopened, err := OpenFileStore(path)
	if err != nil {
		t.Fatal(err)
	}
	l2 := New(reopened, fixedClock())
	if b, err := l2.Balance("alice"); err != nil || b != 30_00 {
		t.Fatalf("after reopen: balance %v, %v; want 30.00", b, err)
	}
	a, err := l2.Deposit("alice", 1_00, "")
	if err != nil {
		t.Fatal(err)
	}
	page, _ := l2.History("alice", 0, 10)
	if a.Balance != 31_00 || len(page.Items) != 2 || page.Items[1].ID != 2 || page.Items[0].Memo != "מתנה" {
		t.Fatalf("after reopen: balance %v, history %+v; want 31.00, IDs 1 and 2, Hebrew memo kept", a.Balance, page.Items)
	}
}
