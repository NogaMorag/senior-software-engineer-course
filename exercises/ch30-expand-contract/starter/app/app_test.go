package app_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"example.com/ch30-expand-contract/app"
	"example.com/ch30-expand-contract/db"
	"example.com/ch30-expand-contract/internal/testdb"
)

func migrated(t *testing.T) (store app.Store, migPool *pgxpool.Pool) {
	t.Helper()
	url := testdb.New(t)
	ctx := context.Background()
	appCfg, err := db.AppConfig(url)
	if err != nil {
		t.Fatal(err)
	}
	migCfg, err := db.MigrationConfig(url)
	if err != nil {
		t.Fatal(err)
	}
	appPool, err := db.Connect(ctx, appCfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(appPool.Close)
	migPool, err = db.Connect(ctx, migCfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(migPool.Close)
	return app.Store{Pool: appPool}, migPool
}

// roundTrip inserts, renames and reads back one user with the SQL of phase p.
func roundTrip(ctx context.Context, s app.Store, p app.Phase) error {
	id, err := s.Insert(ctx, p, "first")
	if err != nil {
		return fmt.Errorf("insert: %w", err)
	}
	if err := s.Rename(ctx, p, id, "second"); err != nil {
		return fmt.Errorf("rename: %w", err)
	}
	got, err := s.Name(ctx, p, id)
	if err != nil {
		return fmt.Errorf("read: %w", err)
	}
	if got != "second" {
		return fmt.Errorf("read %q, want %q", got, "second")
	}
	return nil
}

// TestCompatibilityMatrix runs every code version against every schema version. The plan
// is safe because each migration keeps working for the versions that run on both sides of it.
func TestCompatibilityMatrix(t *testing.T) {
	store, migPool := migrated(t)
	ctx := context.Background()
	phases := []app.Phase{app.PhaseOld, app.PhaseDualWrite, app.PhaseReadNew, app.PhaseNewOnly}
	works := map[int64][]bool{ // schema version: Old, DualWrite, ReadNew, NewOnly
		1: {true, false, false, false},
		2: {true, true, true, false},
		3: {true, true, true, false},
		4: {false, true, true, true},
		5: {false, false, false, true},
	}
	for version := int64(1); version <= 5; version++ {
		// The old code's rows from schema 3 have a NULL display_name, and 00004 would refuse
		// them. That is why every instance must dual-write before the backfill starts.
		if version > 1 {
			if _, err := store.Pool.Exec(ctx, "DELETE FROM users"); err != nil {
				t.Fatal(err)
			}
		}
		if err := db.MigrateTo(ctx, migPool, version); err != nil {
			t.Fatal(err)
		}
		for i, p := range phases {
			t.Run(fmt.Sprintf("schema %d with %v", version, p), func(t *testing.T) {
				err := roundTrip(ctx, store, p)
				switch {
				case works[version][i] && err != nil:
					t.Errorf("want it to work: %v", err)
				case !works[version][i] && err == nil:
					t.Errorf("want it to fail, but it worked")
				}
			})
		}
	}
}

func TestBackfillFillsRowsWrittenByOldCode(t *testing.T) {
	store, migPool := migrated(t)
	ctx := context.Background()
	if err := db.MigrateTo(ctx, migPool, 2); err != nil {
		t.Fatal(err)
	}
	var ids []int64
	for i := range 1200 { // more than two backfill batches
		id, err := store.Insert(ctx, app.PhaseOld, fmt.Sprintf("old %d", i))
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, id)
	}
	if _, err := store.Name(ctx, app.PhaseReadNew, ids[0]); err == nil {
		t.Fatal("before the backfill the new column should still be NULL for old rows")
	}
	if err := db.MigrateTo(ctx, migPool, 3); err != nil {
		t.Fatal(err)
	}
	for i, id := range ids {
		got, err := store.Name(ctx, app.PhaseReadNew, id)
		if want := fmt.Sprintf("old %d", i); err != nil || got != want {
			t.Fatalf("after the backfill user %d = %q, %v; want %q", id, got, err, want)
		}
	}
}
