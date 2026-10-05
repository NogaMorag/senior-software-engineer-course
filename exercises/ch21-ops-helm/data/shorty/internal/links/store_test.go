package links_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"example.com/ch06-shorty/internal/links"
)

func TestMemoryStore(t *testing.T) {
	ctx := context.Background()
	s := links.NewMemoryStore()

	if err := s.Create(ctx, links.Link{Code: "go", URL: "https://go.dev/"}); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := s.Create(ctx, links.Link{Code: "go", URL: "https://example.com/"}); !errors.Is(err, links.ErrExists) {
		t.Errorf("second Create = %v, want ErrExists", err)
	}
	for want := int64(1); want <= 2; want++ {
		l, err := s.Resolve(ctx, "go")
		if err != nil || l.URL != "https://go.dev/" || l.Hits != want {
			t.Errorf("Resolve #%d = %+v, %v; want go.dev with %d hits", want, l, err, want)
		}
	}
	if l, err := s.Get(ctx, "go"); err != nil || l.Hits != 2 {
		t.Errorf("Get = %+v, %v; want 2 hits (Get does not count)", l, err)
	}
	if _, err := s.Resolve(ctx, "nope"); !errors.Is(err, links.ErrNotFound) {
		t.Errorf("Resolve(unknown) = %v, want ErrNotFound", err)
	}
	if _, err := s.Get(ctx, "nope"); !errors.Is(err, links.ErrNotFound) {
		t.Errorf("Get(unknown) = %v, want ErrNotFound", err)
	}
}

// Run with -race: concurrent creates and resolves must not race, and no hit
// may be lost.
func TestMemoryStoreConcurrent(t *testing.T) {
	ctx := context.Background()
	s := links.NewMemoryStore()
	if err := s.Create(ctx, links.Link{Code: "hot", URL: "https://go.dev/"}); err != nil {
		t.Fatal(err)
	}
	const goroutines, each = 8, 500
	var wg sync.WaitGroup
	for g := range goroutines {
		wg.Go(func() {
			for i := range each {
				if _, err := s.Resolve(ctx, "hot"); err != nil {
					t.Error(err)
					return
				}
				if i%50 == 0 { // writers interleaved with the readers
					_ = s.Create(ctx, links.Link{Code: string(rune('a'+g)) + "-" + string(rune('a'+i/50)), URL: "https://example.com/"})
				}
			}
		})
	}
	wg.Wait()
	if l, _ := s.Get(ctx, "hot"); l.Hits != goroutines*each {
		t.Errorf("hits = %d, want %d", l.Hits, goroutines*each)
	}
}
