package bank_test

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"example.com/ch11-isolation-lab/bank"
	"example.com/ch11-isolation-lab/db"
	"example.com/ch11-isolation-lab/internal/testdb"
)

func TestMain(m *testing.M) { os.Exit(testdb.Main(m)) }

// reset leaves exactly the given accounts in the table.
func reset(t *testing.T, pool *pgxpool.Pool, balances map[int64]int64) {
	t.Helper()
	ctx := context.Background()
	if _, err := pool.Exec(ctx, `TRUNCATE accounts`); err != nil {
		t.Fatal(err)
	}
	for id, b := range balances {
		if _, err := pool.Exec(ctx, `INSERT INTO accounts (id, owner, balance) VALUES ($1, 'owner', $2)`, id, b); err != nil {
			t.Fatal(err)
		}
	}
}

func balance(t *testing.T, pool *pgxpool.Pool, id int64) int64 {
	t.Helper()
	var b int64
	if err := pool.QueryRow(context.Background(), `SELECT balance FROM accounts WHERE id = $1`, id).Scan(&b); err != nil {
		t.Fatal(err)
	}
	return b
}

func begin(t *testing.T, pool *pgxpool.Pool, level pgx.TxIsoLevel) pgx.Tx {
	t.Helper()
	tx, err := pool.BeginTx(context.Background(), pgx.TxOptions{IsoLevel: level})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = tx.Rollback(context.Background()) }) // no-op after Commit
	return tx
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

// Two withdrawals of 30 from 100, interleaved by hand under READ COMMITTED:
// both read 100, both write 70. The final balance should be 40; it is 70.
func TestLostUpdateUnderReadCommitted(t *testing.T) {
	pool := testdb.Pool(t)
	reset(t, pool, map[int64]int64{1: 100})
	ctx := context.Background()

	tx1 := begin(t, pool, pgx.ReadCommitted)
	tx2 := begin(t, pool, pgx.ReadCommitted)
	b1, err := bank.ReadBalance(ctx, tx1, 1)
	must(t, err)
	b2, err := bank.ReadBalance(ctx, tx2, 1)
	must(t, err)
	must(t, bank.SetBalance(ctx, tx1, 1, b1-30))
	must(t, tx1.Commit(ctx))
	must(t, bank.SetBalance(ctx, tx2, 1, b2-30)) // does not wait: tx1 already committed
	must(t, tx2.Commit(ctx))

	if got := balance(t, pool, 1); got != 70 {
		t.Fatalf("balance = %d; the lost update should leave 70 (40 would be correct)", got)
	}
}

// The same interleaving under REPEATABLE READ: PostgreSQL refuses the second write.
func TestRepeatableReadRejectsTheLostUpdate(t *testing.T) {
	pool := testdb.Pool(t)
	reset(t, pool, map[int64]int64{1: 100})
	ctx := context.Background()

	tx1 := begin(t, pool, pgx.RepeatableRead)
	tx2 := begin(t, pool, pgx.RepeatableRead)
	b1, err := bank.ReadBalance(ctx, tx1, 1)
	must(t, err)
	b2, err := bank.ReadBalance(ctx, tx2, 1)
	must(t, err)
	must(t, bank.SetBalance(ctx, tx1, 1, b1-30))
	must(t, tx1.Commit(ctx))
	err = bank.SetBalance(ctx, tx2, 1, b2-30)
	if db.SQLState(err) != db.SerializationFailure {
		t.Fatalf("second write: err = %v, want SQLSTATE 40001", err)
	}
	if got := balance(t, pool, 1); got != 70 {
		t.Fatalf("balance = %d, want 70 (only the first withdrawal happened)", got)
	}
}

// With FOR UPDATE the second reader waits until the first commits, then sees 70.
func TestForUpdateMakesTheSecondReaderWait(t *testing.T) {
	pool := testdb.Pool(t)
	reset(t, pool, map[int64]int64{1: 100})
	ctx := context.Background()

	tx1 := begin(t, pool, pgx.ReadCommitted)
	tx2 := begin(t, pool, pgx.ReadCommitted)
	b1, err := bank.ReadBalanceForUpdate(ctx, tx1, 1)
	must(t, err)

	type result struct {
		balance int64
		err     error
	}
	second := make(chan result, 1)
	go func() {
		b, err := bank.ReadBalanceForUpdate(ctx, tx2, 1)
		second <- result{b, err}
	}()
	testdb.WaitForLockWaiters(t, pool, 1) // tx2 is now blocked on tx1's row lock
	select {
	case r := <-second:
		t.Fatalf("second reader did not wait: %+v", r)
	default:
	}

	must(t, bank.SetBalance(ctx, tx1, 1, b1-30))
	must(t, tx1.Commit(ctx))
	r := <-second
	must(t, r.err)
	if r.balance != 70 {
		t.Fatalf("second reader saw %d, want 70", r.balance)
	}
	must(t, bank.SetBalance(ctx, tx2, 1, r.balance-30))
	must(t, tx2.Commit(ctx))
	if got := balance(t, pool, 1); got != 40 {
		t.Fatalf("balance = %d, want 40", got)
	}
}

// 25 concurrent withdrawals of 5 from 100: exactly 20 succeed and the balance is 0.
func TestConcurrentWithdrawalsAreExact(t *testing.T) {
	pool := testdb.Pool(t)
	tests := []struct {
		name     string
		withdraw func(context.Context, *pgxpool.Pool, int64, int64) error
	}{
		{"FOR UPDATE", bank.WithdrawLocked},
		{"atomic UPDATE", bank.WithdrawAtomic},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reset(t, pool, map[int64]int64{1: 100})
			var wg sync.WaitGroup
			var mu sync.Mutex
			ok, refused := 0, 0
			for range 25 {
				wg.Go(func() {
					err := tc.withdraw(context.Background(), pool, 1, 5)
					mu.Lock()
					defer mu.Unlock()
					switch {
					case err == nil:
						ok++
					case errors.Is(err, bank.ErrInsufficientFunds):
						refused++
					default:
						t.Errorf("withdraw: %v", err)
					}
				})
			}
			wg.Wait()
			if ok != 20 || refused != 5 {
				t.Fatalf("%d succeeded and %d were refused, want 20 and 5", ok, refused)
			}
			if got := balance(t, pool, 1); got != 0 {
				t.Fatalf("balance = %d, want 0", got)
			}
		})
	}
}

// Two opposite transfers, each holding one row and asking for the other's:
// PostgreSQL detects the cycle and aborts exactly one of them with 40P01.
func TestOppositeTransfersDeadlock(t *testing.T) {
	pool := testdb.Pool(t)
	reset(t, pool, map[int64]int64{1: 100, 2: 100})
	ctx := context.Background()

	tx1 := begin(t, pool, pgx.ReadCommitted)
	tx2 := begin(t, pool, pgx.ReadCommitted)
	must(t, bank.AddToBalance(ctx, tx1, 1, -10)) // tx1 holds row 1
	must(t, bank.AddToBalance(ctx, tx2, 2, -10)) // tx2 holds row 2

	first := make(chan error, 1)
	go func() { first <- bank.AddToBalance(ctx, tx1, 2, 10) }() // tx1 waits for row 2
	testdb.WaitForLockWaiters(t, pool, 1)
	err2 := bank.AddToBalance(ctx, tx2, 1, 10) // tx2 waits for row 1: a cycle
	err1 := <-first

	deadlocks := 0
	for _, err := range []error{err1, err2} {
		if db.SQLState(err) == db.DeadlockDetected {
			deadlocks++
		} else if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if deadlocks != 1 {
		t.Fatalf("got %d deadlock errors (err1=%v, err2=%v), want exactly 1", deadlocks, err1, err2)
	}
	t.Logf("PostgreSQL said: %v", errors.Join(err1, err2))
}

// Transfer locks rows in id order, so opposite transfers queue up instead of deadlocking.
func TestOrderedTransfersNeverDeadlock(t *testing.T) {
	pool := testdb.Pool(t)
	reset(t, pool, map[int64]int64{1: 1000, 2: 1000})
	var wg sync.WaitGroup
	for i := range 40 {
		from, to := int64(1), int64(2)
		if i%2 == 1 {
			from, to = to, from
		}
		wg.Go(func() {
			if err := bank.Transfer(context.Background(), pool, from, to, 1); err != nil {
				t.Errorf("transfer %d to %d: %v", from, to, err)
			}
		})
	}
	wg.Wait()
	if a, b := balance(t, pool, 1), balance(t, pool, 2); a != 1000 || b != 1000 {
		t.Fatalf("balances = %d and %d, want 1000 and 1000", a, b)
	}
}

// Migration 00004 adds a CHECK constraint: the database itself refuses a negative balance.
func TestBalanceCannotGoNegative(t *testing.T) {
	pool := testdb.Pool(t)
	reset(t, pool, map[int64]int64{1: 10})
	err := pgx.BeginFunc(context.Background(), pool, func(tx pgx.Tx) error {
		return bank.SetBalance(context.Background(), tx, 1, -1)
	})
	if code := db.SQLState(err); code != "23514" {
		t.Fatalf("err = %v, want check_violation (SQLSTATE 23514); did you write migration 00004?", err)
	}
}
