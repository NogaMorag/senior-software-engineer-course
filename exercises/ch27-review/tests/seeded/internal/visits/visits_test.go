package visits_test

import (
	"errors"
	"sync"
	"testing"

	"example.com/ch27-review/internal/visits"
)

func TestReport(t *testing.T) {
	c := visits.New()
	c.Add("noga")
	c.Add("noga")
	if got, want := c.Report("noga"), "noga visited 2 times"; got != want {
		t.Fatalf("Report = %q; want %q", got, want)
	}
}

func TestSeenAndLookup(t *testing.T) {
	c := visits.New()
	c.Add("noga")
	if !c.Seen("noga") {
		t.Error("Seen(noga) = false after Add")
	}
	if c.Seen("dana") {
		t.Error("Seen(dana) = true, but dana never visited")
	}
	if _, err := c.Lookup("dana"); !errors.Is(err, visits.ErrUnknown) {
		t.Errorf("Lookup(dana) error = %v; want ErrUnknown", err)
	}
}

// TestConcurrentUse is meant for go test -race.
func TestConcurrentUse(t *testing.T) {
	c := visits.New()
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			for range 100 {
				c.Add("noga")
				_ = c.Report("noga")
			}
		})
	}
	wg.Wait()
	if n, err := c.Lookup("noga"); err != nil || n != 800 {
		t.Fatalf("Lookup = %d, %v; want 800", n, err)
	}
}
