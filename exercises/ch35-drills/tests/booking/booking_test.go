package booking

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// at returns 09:00 UTC on the drill day plus h hours.
func at(h float64) time.Time {
	return time.Date(2026, 10, 5, 9, 0, 0, 0, time.UTC).Add(time.Duration(h * float64(time.Hour)))
}

// stores returns every implementation to run the contract against: always Memory, and
// Postgres when DATABASE_URL is set.
func stores() map[string]func(t *testing.T) Store {
	return map[string]func(t *testing.T) Store{
		"memory":   func(*testing.T) Store { return NewMemory() },
		"postgres": func(t *testing.T) Store { return &Postgres{DB: postgres(t)} },
	}
}

var schemaSeq atomic.Int64

// postgres gives the test a private schema with the bookings table, dropped at the end.
func postgres(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("integration test: set DATABASE_URL to the chapter 35 database; see README")
	}
	ctx := context.Background()
	admin, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(admin.Close)
	schema := fmt.Sprintf("ch35_test_%d_%d", os.Getpid(), schemaSeq.Add(1))
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := admin.Exec(context.Background(), "DROP SCHEMA "+schema+" CASCADE"); err != nil {
			t.Errorf("drop schema: %v", err)
		}
	})
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		t.Fatal(err)
	}
	// Schema (yours) creates btree_gist if the database lacks it; it then lives in this
	// test's schema and is dropped with it. Tests in this package run one at a time.
	cfg.ConnConfig.RuntimeParams["search_path"] = schema + ", public"
	cfg.MaxConns = 20
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	if _, err := pool.Exec(ctx, Schema); err != nil {
		t.Fatalf("apply schema: %v", err)
	}
	return pool
}

func TestBookingContract(t *testing.T) {
	ctx := context.Background()
	for name, open := range stores() {
		t.Run(name, func(t *testing.T) {
			s := open(t)
			// An existing booking of room-1 from 10:00 to 12:00.
			if _, err := s.Book(ctx, "room-1", "noga", at(1), at(3)); err != nil {
				t.Fatalf("first booking: %v", err)
			}
			cases := []struct {
				name       string
				resource   string
				start, end float64
				want       error
			}{
				{"identical range", "room-1", 1, 3, ErrConflict},
				{"inside", "room-1", 1.5, 2, ErrConflict},
				{"around", "room-1", 0, 4, ErrConflict},
				{"overlaps the start", "room-1", 0.5, 1.5, ErrConflict},
				{"overlaps the end", "room-1", 2.5, 3.5, ErrConflict},
				{"ends exactly at its start", "room-1", 0, 1, nil},
				{"starts exactly at its end", "room-1", 3, 4, nil},
				{"same time, other room", "room-2", 1, 3, nil},
				{"end before start", "room-1", 6, 5, ErrInvalid},
				{"empty range", "room-1", 6, 6, ErrInvalid},
				{"no resource", "", 6, 7, ErrInvalid},
			}
			for _, tc := range cases {
				_, err := s.Book(ctx, tc.resource, "dana", at(tc.start), at(tc.end))
				if tc.want == nil && err != nil {
					t.Errorf("%s: unexpected error %v", tc.name, err)
				}
				if tc.want != nil && !errors.Is(err, tc.want) {
					t.Errorf("%s: err = %v, want %v", tc.name, err, tc.want)
				}
			}
			got, err := s.List(ctx, "room-1")
			if err != nil {
				t.Fatal(err)
			}
			wantStarts := []time.Time{at(0), at(1), at(3)}
			if len(got) != len(wantStarts) {
				t.Fatalf("List: %d bookings, want %d: %+v", len(got), len(wantStarts), got)
			}
			for i, b := range got {
				if !b.Start.Equal(wantStarts[i]) || b.Resource != "room-1" || b.ID == 0 {
					t.Errorf("List[%d] = %+v, want room-1 starting at %v with an id", i, b, wantStarts[i])
				}
			}
		})
	}
}

func TestCancelFreesTheRange(t *testing.T) {
	ctx := context.Background()
	for name, open := range stores() {
		t.Run(name, func(t *testing.T) {
			s := open(t)
			b, err := s.Book(ctx, "rig", "noga", at(0), at(1))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := s.Book(ctx, "rig", "dana", at(0), at(1)); !errors.Is(err, ErrConflict) {
				t.Fatalf("before cancel: %v, want ErrConflict", err)
			}
			if err := s.Cancel(ctx, b.ID); err != nil {
				t.Fatal(err)
			}
			if err := s.Cancel(ctx, b.ID); !errors.Is(err, ErrNotFound) {
				t.Fatalf("second cancel: %v, want ErrNotFound", err)
			}
			if _, err := s.Book(ctx, "rig", "dana", at(0), at(1)); err != nil {
				t.Fatalf("after cancel: %v, want the range to be free", err)
			}
		})
	}
}

func TestNoDoubleBookingUnderConcurrency(t *testing.T) {
	ctx := context.Background()
	for name, open := range stores() {
		t.Run(name, func(t *testing.T) {
			s := open(t)
			const n = 40
			var wins, conflicts atomic.Int32
			gate := make(chan struct{})
			var wg sync.WaitGroup
			for i := range n {
				wg.Go(func() {
					<-gate
					// Every caller wants a different but overlapping slice of the same hour.
					start := at(float64(i%4) * 0.25)
					_, err := s.Book(ctx, "room-1", fmt.Sprintf("user-%d", i), start, start.Add(time.Hour))
					switch {
					case err == nil:
						wins.Add(1)
					case errors.Is(err, ErrConflict):
						conflicts.Add(1)
					default:
						t.Errorf("caller %d: %v", i, err)
					}
				})
			}
			close(gate)
			wg.Wait()
			if wins.Load() != 1 || conflicts.Load() != n-1 {
				t.Fatalf("%d wins and %d conflicts, want 1 and %d", wins.Load(), conflicts.Load(), n-1)
			}
			got, err := s.List(ctx, "room-1")
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != 1 {
				t.Fatalf("List has %d bookings, want 1", len(got))
			}
		})
	}
}

func TestDifferentResourcesDoNotBlockEachOther(t *testing.T) {
	ctx := context.Background()
	for name, open := range stores() {
		t.Run(name, func(t *testing.T) {
			s := open(t)
			var wg sync.WaitGroup
			var failures atomic.Int32
			for i := range 20 {
				wg.Go(func() {
					if _, err := s.Book(ctx, fmt.Sprintf("desk-%d", i), "noga", at(0), at(8)); err != nil {
						failures.Add(1)
					}
				})
			}
			wg.Wait()
			if failures.Load() != 0 {
				t.Fatalf("%d of 20 bookings of different desks failed, want 0", failures.Load())
			}
		})
	}
}
