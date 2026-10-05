package rollout_test

import (
	"context"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"example.com/ch30-expand-contract/app"
	"example.com/ch30-expand-contract/db"
	"example.com/ch30-expand-contract/internal/testdb"
	"example.com/ch30-expand-contract/rollout"
)

const seedRows = 5000

func pools(t *testing.T) (appPool, migPool *pgxpool.Pool) {
	t.Helper()
	url := testdb.New(t)
	ctx := context.Background()
	for _, p := range []struct {
		cfg  func(string) (*pgxpool.Config, error)
		dest **pgxpool.Pool
	}{{db.AppConfig, &appPool}, {db.MigrationConfig, &migPool}} {
		cfg, err := p.cfg(url)
		if err != nil {
			t.Fatal(err)
		}
		pool, err := db.Connect(ctx, cfg)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(pool.Close)
		*p.dest = pool
	}
	if err := db.MigrateTo(ctx, migPool, 1); err != nil {
		t.Fatal(err)
	}
	return appPool, migPool
}

// startFleet runs n instances in phase p until stop is called or the test ends. stop waits
// until every loop has finished its last request.
func startFleet(t *testing.T, pool *pgxpool.Pool, reg *app.Registry, n int, p app.Phase) (fleet []*app.Instance, stop func()) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	fleet = make([]*app.Instance, n)
	for i := range fleet {
		fleet[i] = app.NewInstance(i, app.Store{Pool: pool}, reg, p)
		wg.Go(func() { fleet[i].Run(ctx) })
	}
	stop = func() { cancel(); wg.Wait() }
	t.Cleanup(stop)
	return fleet, stop
}

func waitForOps(t *testing.T, fleet []*app.Instance, n int64) {
	t.Helper()
	deadline := time.Now().Add(30 * time.Second)
	for _, in := range fleet {
		target := in.Ops() + n
		for in.Ops() < target {
			if time.Now().After(deadline) {
				t.Fatalf("instance %d made no progress", in.ID)
			}
			time.Sleep(5 * time.Millisecond)
		}
	}
}

func failures(fleet []*app.Instance) (n int64, errs []string) {
	for _, in := range fleet {
		n += in.Failures()
		errs = append(errs, in.Errors()...)
	}
	return n, errs
}

type column struct {
	name     string
	nullable string
}

func columns(t *testing.T, pool *pgxpool.Pool) []column {
	t.Helper()
	rows, err := pool.Query(context.Background(), `
		SELECT column_name::text, is_nullable::text FROM information_schema.columns
		WHERE table_schema = 'public' AND table_name = 'users' ORDER BY ordinal_position`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var cols []column
	for rows.Next() {
		var c column
		if err := rows.Scan(&c.name, &c.nullable); err != nil {
			t.Fatal(err)
		}
		cols = append(cols, c)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return cols
}

func TestRenameWithZeroFailedRequests(t *testing.T) {
	appPool, migPool := pools(t)
	ctx := context.Background()
	reg := app.NewRegistry()
	rows, err := appPool.Query(ctx, `
		INSERT INTO users (fullname) SELECT 'seed ' || g FROM generate_series(1, $1) AS g
		RETURNING id, fullname`, seedRows)
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		var id int64
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			t.Fatal(err)
		}
		reg.Set(id, name)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}

	fleet, stop := startFleet(t, appPool, reg, 2, app.PhaseOld)
	waitForOps(t, fleet, 50)
	reports, err := rollout.Run(ctx, migPool, fleet, rollout.Plan, 50)
	stop() // the final checks compare the database with the registry, so the writers must be done
	for _, r := range reports {
		t.Logf("%-45s requests=%-5d failed=%d", r.Step, r.Ops, r.Failures)
	}
	if err != nil {
		t.Fatal(err)
	}
	if n, errs := failures(fleet); n != 0 {
		t.Fatalf("%d failed requests during the rollout; first errors:\n%v", n, errs)
	}
	if len(reports) != 4 {
		t.Fatalf("the plan ran %d steps, want 4 (expand, backfill, swap, contract)", len(reports))
	}
	for _, in := range fleet {
		if in.Phase() != app.PhaseNewOnly {
			t.Errorf("instance %d ends in %v, want %v", in.ID, in.Phase(), app.PhaseNewOnly)
		}
	}

	v, err := db.Version(ctx, migPool)
	if err != nil || v != 5 {
		t.Fatalf("schema version = %d, %v; want 5", v, err)
	}
	want := []column{{"id", "NO"}, {"created_at", "NO"}, {"display_name", "NO"}}
	if got := columns(t, appPool); !slices.Equal(got, want) {
		t.Fatalf("users columns = %v, want %v", got, want)
	}

	// Every write that succeeded is in the new column: nothing was lost between the columns.
	expected := reg.Expected()
	got := map[int64]string{}
	rows, err = appPool.Query(ctx, `SELECT id, display_name FROM users`)
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		var id int64
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			t.Fatal(err)
		}
		got[id] = name
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if len(expected) <= seedRows {
		t.Fatalf("only %d users known; the writers never inserted", len(expected))
	}
	for id, name := range expected {
		if got[id] != name {
			t.Fatalf("user %d: display_name = %q, last write was %q", id, got[id], name)
		}
	}
	t.Logf("users at the end: %d, of them %d written during the rollout", len(got), len(got)-seedRows)
}

func TestNaiveRenameBreaksRunningCode(t *testing.T) {
	appPool, _ := pools(t)
	fleet, _ := startFleet(t, appPool, app.NewRegistry(), 2, app.PhaseOld)
	waitForOps(t, fleet, 50)
	if _, err := appPool.Exec(context.Background(),
		`ALTER TABLE users RENAME COLUMN fullname TO display_name`); err != nil {
		t.Fatal(err)
	}
	waitForOps(t, fleet, 50)
	n, errs := failures(fleet)
	if n == 0 {
		t.Fatal("a rename under running code should break it, but no request failed")
	}
	t.Logf("%d failed requests after a one-step rename; the first: %s", n, errs[0])
}
