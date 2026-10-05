package migrations_test

import (
	"context"
	"io/fs"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"example.com/hookline/internal/db"
	"example.com/hookline/internal/testdb"
	"example.com/hookline/migrations"
)

// The whole data model of chapter 37, section 37.4.
var wantTables = []string{"attempts", "deliveries", "endpoints", "events", "leader_terms", "outbox", "subscriptions", "tenants"}

func files(t *testing.T) []string {
	t.Helper()
	names, err := fs.Glob(migrations.FS, "*.sql")
	if err != nil {
		t.Fatal(err)
	}
	return names
}

// Every file has both halves, so every migration can be rolled back (runs without a database).
func TestEveryMigrationHasUpAndDown(t *testing.T) {
	names := files(t)
	if len(names) != 6 {
		t.Fatalf("migrations = %v, want 6 files", names)
	}
	for _, name := range names {
		b, _ := fs.ReadFile(migrations.FS, name)
		if !strings.Contains(string(b), "-- +goose Up") || !strings.Contains(string(b), "-- +goose Down") {
			t.Errorf("%s lacks a -- +goose Up or -- +goose Down section", name)
		}
	}
}

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

// On an empty database: up, all the way down, and up again leave the expected tables each time.
func TestMigrationsRoundTrip(t *testing.T) {
	admin := os.Getenv("DATABASE_URL")
	if admin == "" {
		t.Skip("integration test: set DATABASE_URL to the Hookline database (make db-up)")
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

	latest := int64(len(files(t)))
	for round := 1; round <= 2; round++ {
		if err := db.Migrate(ctx, pool); err != nil {
			t.Fatalf("round %d up: %v", round, err)
		}
		if v, err := db.Version(ctx, pool); err != nil || v != latest {
			t.Fatalf("round %d: version %d (%v), want %d", round, v, err, latest)
		}
		if got := tables(t, pool); !slices.Equal(got, wantTables) {
			t.Fatalf("round %d after up: tables %v, want %v", round, got, wantTables)
		}
		if err := db.MigrateDownTo(ctx, pool, 0); err != nil {
			t.Fatalf("round %d down: %v", round, err)
		}
		if got := tables(t, pool); len(got) != 0 {
			t.Fatalf("round %d after down: tables %v left behind", round, got)
		}
	}
}
