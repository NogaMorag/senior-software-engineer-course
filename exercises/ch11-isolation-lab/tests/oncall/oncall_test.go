package oncall_test

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"example.com/ch11-isolation-lab/db"
	"example.com/ch11-isolation-lab/internal/testdb"
	"example.com/ch11-isolation-lab/oncall"
)

func TestMain(m *testing.M) { os.Exit(testdb.Main(m)) }

// reset puts alice and bob on call for shift 1.
func reset(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	_, err := pool.Exec(context.Background(), `
		TRUNCATE oncall;
		INSERT INTO oncall (shift_id, doctor) VALUES (1, 'alice'), (1, 'bob');`)
	if err != nil {
		t.Fatal(err)
	}
}

func onCall(t *testing.T, pool *pgxpool.Pool) int {
	t.Helper()
	var n int
	err := pool.QueryRow(context.Background(), `SELECT count(*) FROM oncall WHERE shift_id = 1 AND on_call`).Scan(&n)
	if err != nil {
		t.Fatal(err)
	}
	return n
}

func begin(t *testing.T, pool *pgxpool.Pool, level pgx.TxIsoLevel) pgx.Tx {
	t.Helper()
	tx, err := pool.BeginTx(context.Background(), pgx.TxOptions{IsoLevel: level})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = tx.Rollback(context.Background()) })
	return tx
}

// Alice and Bob each check "is someone else on call?" in their own snapshot, see two,
// and each update only their own row. No row is written twice, so REPEATABLE READ has
// nothing to refuse, and the shift ends up with nobody on call.
func TestWriteSkewUnderRepeatableRead(t *testing.T) {
	pool := testdb.Pool(t)
	reset(t, pool)
	ctx := context.Background()

	tx1 := begin(t, pool, pgx.RepeatableRead)
	tx2 := begin(t, pool, pgx.RepeatableRead)
	if err := oncall.GoOffCall(ctx, tx1, 1, "alice"); err != nil {
		t.Fatal(err)
	}
	if err := oncall.GoOffCall(ctx, tx2, 1, "bob"); err != nil {
		t.Fatal(err)
	}
	if err := tx1.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	if err := tx2.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	if got := onCall(t, pool); got != 0 {
		t.Fatalf("%d doctors on call; write skew should leave 0", got)
	}
}

// The same interleaving under SERIALIZABLE: PostgreSQL sees the read/write dependency
// cycle and aborts one of the two with 40001. At least one doctor stays on call.
func TestSerializableCatchesWriteSkew(t *testing.T) {
	pool := testdb.Pool(t)
	reset(t, pool)
	ctx := context.Background()

	tx1 := begin(t, pool, pgx.Serializable)
	tx2 := begin(t, pool, pgx.Serializable)
	var errs []error
	errs = append(errs, oncall.GoOffCall(ctx, tx1, 1, "alice"))
	errs = append(errs, oncall.GoOffCall(ctx, tx2, 1, "bob"))
	errs = append(errs, tx1.Commit(ctx))
	errs = append(errs, tx2.Commit(ctx))

	failures := 0
	for _, err := range errs {
		if db.SQLState(err) == db.SerializationFailure {
			failures++
		} else if err != nil && !errors.Is(err, pgx.ErrTxClosed) && !errors.Is(err, pgx.ErrTxCommitRollback) {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if failures == 0 {
		t.Fatalf("no serialization failure; errors were %v", errs)
	}
	if got := onCall(t, pool); got != 1 {
		t.Fatalf("%d doctors on call, want 1", got)
	}
	t.Logf("PostgreSQL said: %v", errors.Join(errs...))
}

// The retry helper makes the conflict invisible to callers: of two doctors leaving at the
// same moment, one succeeds, the other is told she is the last one on call.
func TestGoOffCallSerializableKeepsOneOnCall(t *testing.T) {
	pool := testdb.Pool(t)
	for range 20 {
		reset(t, pool)
		var wg sync.WaitGroup
		errs := make([]error, 2)
		for i, doctor := range []string{"alice", "bob"} {
			wg.Go(func() { errs[i] = oncall.GoOffCallSerializable(context.Background(), pool, 1, doctor) })
		}
		wg.Wait()
		ok, last := 0, 0
		for _, err := range errs {
			switch {
			case err == nil:
				ok++
			case errors.Is(err, oncall.ErrLastOnCall):
				last++
			default:
				t.Fatalf("unexpected error: %v", err)
			}
		}
		if ok != 1 || last != 1 || onCall(t, pool) != 1 {
			t.Fatalf("ok=%d last=%d on call=%d, want 1, 1 and 1", ok, last, onCall(t, pool))
		}
	}
}

// A retry, made visible: on the first attempt another transaction takes alice off call
// between our read and our write. PostgreSQL aborts us, the helper runs the function
// again, and the second attempt sees one doctor left and refuses.
func TestRetrySerializableRunsTheWholeTransactionAgain(t *testing.T) {
	pool := testdb.Pool(t)
	reset(t, pool)
	attempts := 0
	_, err := db.RetrySerializable(context.Background(), pool, func(ctx context.Context, tx pgx.Tx) (struct{}, error) {
		attempts++
		if _, err := oncall.CountOnCall(ctx, tx, 1); err != nil { // our read
			return struct{}{}, err
		}
		if attempts == 1 { // the interfering transaction commits in between
			if err := oncall.GoOffCallSerializable(ctx, pool, 1, "alice"); err != nil {
				return struct{}{}, err
			}
		}
		return struct{}{}, oncall.GoOffCall(ctx, tx, 1, "bob") // our write
	})
	if !errors.Is(err, oncall.ErrLastOnCall) {
		t.Fatalf("err = %v, want ErrLastOnCall from the second attempt", err)
	}
	if attempts != 2 {
		t.Fatalf("fn ran %d times, want 2", attempts)
	}
	if got := onCall(t, pool); got != 1 {
		t.Fatalf("%d doctors on call, want 1", got)
	}
}
