package migrations_test

import (
	"context"
	"os"
	"path"
	"slices"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"example.com/ch11-isolation-lab/db"
	"example.com/ch11-isolation-lab/internal/testdb"
	"example.com/ch11-isolation-lab/migrations"
)

func tables(t *testing.T, pool *pgxpool.Pool) []string {
	t.Helper()
	rows, err := pool.Query(context.Background(), `
		SELECT table_name::text FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name <> 'goose_db_version'
		ORDER BY 1`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var names []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			t.Fatal(err)
		}
		names = append(names, n)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return names
}

// Every migration has a Down that really undoes its Up: on an empty database, up, all
// the way down, and up again leave the expected tables each time.
func TestMigrationsRoundTrip(t *testing.T) {
	admin := os.Getenv("DATABASE_URL")
	if admin == "" {
		t.Skip("integration test: set DATABASE_URL to the chapter 11 database; see README")
	}
	ctx := context.Background()
	url, drop, err := testdb.CreateDatabase(ctx, admin)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := drop(); err != nil {
			t.Errorf("drop: %v", err)
		}
	})
	cfg, err := db.Config(url)
	if err != nil {
		t.Fatal(err)
	}
	pool, err := db.Connect(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	files, err := migrations.FS.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	latest := int64(0)
	for _, f := range files {
		if path.Ext(f.Name()) == ".sql" {
			latest++
		}
	}
	want := []string{"accounts", "jobs", "oncall"}

	for round := 1; round <= 2; round++ {
		if err := db.Migrate(ctx, pool); err != nil {
			t.Fatalf("round %d up: %v", round, err)
		}
		if v, err := db.Version(ctx, pool); err != nil || v != latest {
			t.Fatalf("round %d: version %d (%v), want %d", round, v, err, latest)
		}
		if got := tables(t, pool); !slices.Equal(got, want) {
			t.Fatalf("round %d after up: tables %v, want %v", round, got, want)
		}
		if err := db.MigrateDownTo(ctx, pool, 0); err != nil {
			t.Fatalf("round %d down: %v", round, err)
		}
		if got := tables(t, pool); len(got) != 0 {
			t.Fatalf("round %d after down: tables %v left behind", round, got)
		}
	}
}
