package catalog_test

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"testing"
	"time"

	"example.com/ch09-catalog-api/internal/catalog"
	"example.com/ch09-catalog-api/internal/page"
)

// fakeClock advances by step on every call; step 0 gives every item the same timestamp.
type fakeClock struct {
	t    time.Time
	step time.Duration
}

func (c *fakeClock) now() time.Time { c.t = c.t.Add(c.step); return c.t }

func seed(t *testing.T, m *catalog.Memory, n int, prefix string) []catalog.Item {
	t.Helper()
	out := make([]catalog.Item, 0, n)
	for i := range n {
		it, err := m.Create(context.Background(), fmt.Sprintf("%s-%04d", prefix, i), "item", 100)
		if err != nil {
			t.Fatal(err)
		}
		out = append(out, it)
	}
	return out
}

// listAll pages through the whole list with the cursor, calling between after every page.
func listAll(t *testing.T, m *catalog.Memory, limit int, between func()) []catalog.Item {
	t.Helper()
	var all []catalog.Item
	var after *page.Cursor
	for range 1000 {
		items, more, err := m.List(context.Background(), after, limit)
		if err != nil {
			t.Fatal(err)
		}
		all = append(all, items...)
		if !more {
			return all
		}
		last := items[len(items)-1]
		c, err := page.Decode(page.Encode(page.Cursor{CreatedAt: last.CreatedAt, ID: last.ID}))
		if err != nil {
			t.Fatal(err)
		}
		after = &c
		between()
	}
	t.Fatal("pagination did not terminate")
	return nil
}

func TestCreateGetDuplicate(t *testing.T) {
	m := catalog.NewMemory(nil)
	ctx := context.Background()
	it, err := m.Create(ctx, "MUG-1", "Blue mug", 4500)
	if err != nil {
		t.Fatal(err)
	}
	got, err := m.Get(ctx, it.ID)
	if err != nil || got != it {
		t.Fatalf("Get = %+v, %v; want %+v", got, err, it)
	}
	if _, err := m.Get(ctx, "itm_99999999"); !errors.Is(err, catalog.ErrNotFound) {
		t.Fatalf("Get(missing) error = %v, want ErrNotFound", err)
	}
	if _, err := m.Create(ctx, "MUG-1", "Other", 1); !errors.Is(err, catalog.ErrDuplicateSKU) {
		t.Fatalf("duplicate SKU error = %v, want ErrDuplicateSKU", err)
	}
}

func TestListNewestFirstWithTies(t *testing.T) {
	for _, step := range []time.Duration{time.Millisecond, 0} {
		t.Run(fmt.Sprintf("step=%s", step), func(t *testing.T) {
			clk := &fakeClock{t: time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC), step: step}
			m := catalog.NewMemory(clk.now)
			created := seed(t, m, 7, "A")
			got := listAll(t, m, 3, func() {})
			if len(got) != len(created) {
				t.Fatalf("listed %d items, want %d", len(got), len(created))
			}
			for i := range got {
				if want := created[len(created)-1-i]; got[i].ID != want.ID {
					t.Fatalf("position %d = %s, want %s (newest first)", i, got[i].ID, want.ID)
				}
			}
		})
	}
}

// The core property: paging with a cursor while other clients insert never shows an item twice and
// never skips an item that existed when paging started.
func TestCursorNoDuplicatesNoGapsUnderInserts(t *testing.T) {
	for _, step := range []time.Duration{time.Millisecond, 0} {
		for _, limit := range []int{1, 3, 10} {
			t.Run(fmt.Sprintf("step=%s/limit=%d", step, limit), func(t *testing.T) {
				clk := &fakeClock{t: time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC), step: step}
				m := catalog.NewMemory(clk.now)
				original := seed(t, m, 25, "A")
				rng := rand.New(rand.NewPCG(1, uint64(limit)))
				inserted := 0
				got := listAll(t, m, limit, func() {
					for range rng.IntN(4) {
						seed(t, m, 1, fmt.Sprintf("B%d", inserted))
						inserted++
					}
				})
				seen := map[string]int{}
				for _, it := range got {
					seen[it.ID]++
					if seen[it.ID] > 1 {
						t.Fatalf("item %s listed twice", it.ID)
					}
				}
				for _, it := range original {
					if seen[it.ID] != 1 {
						t.Fatalf("item %s that existed before paging was skipped", it.ID)
					}
				}
				if inserted == 0 {
					t.Fatal("test did not insert anything; it proves nothing")
				}
			})
		}
	}
}

// Offset pagination over a newest-first list repeats items when new ones arrive between pages.
func TestOffsetDuplicatesUnderInserts(t *testing.T) {
	m := catalog.NewMemory(nil)
	ctx := context.Background()
	seed(t, m, 10, "A")
	first := m.ListOffset(ctx, 0, 5)
	seed(t, m, 2, "B") // two new items push everything down by two
	second := m.ListOffset(ctx, 5, 5)
	seen := map[string]bool{}
	dups := 0
	for _, it := range append(first, second...) {
		if seen[it.ID] {
			dups++
		}
		seen[it.ID] = true
	}
	if dups != 2 {
		t.Fatalf("offset pagination showed %d duplicates, expected 2: this test documents the flaw", dups)
	}
}
