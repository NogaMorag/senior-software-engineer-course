package degrade_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"example.com/ch15-resilience-lab/degrade"
)

var errDown = errors.New("down")

func TestStale(t *testing.T) {
	now := time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)
	s := &degrade.Stale[string]{MaxStale: time.Minute, Now: func() time.Time { return now }}
	ok := func(context.Context) (string, error) { return "fresh", nil }
	bad := func(context.Context) (string, error) { return "", errDown }
	ctx := context.Background()

	if _, _, err := s.Get(ctx, bad); !errors.Is(err, degrade.ErrNoValue) || !errors.Is(err, errDown) {
		t.Fatalf("failure with nothing cached: err = %v, want ErrNoValue and the cause", err)
	}
	if v, stale, err := s.Get(ctx, ok); v != "fresh" || stale || err != nil {
		t.Fatalf("success = %q, %v, %v", v, stale, err)
	}
	now = now.Add(59 * time.Second)
	if v, stale, err := s.Get(ctx, bad); v != "fresh" || !stale || err != nil {
		t.Fatalf("failure within MaxStale = %q, stale %v, %v; want the old value marked stale", v, stale, err)
	}
	now = now.Add(2 * time.Second)
	if _, _, err := s.Get(ctx, bad); !errors.Is(err, degrade.ErrNoValue) {
		t.Fatalf("failure after MaxStale: err = %v, want ErrNoValue", err)
	}
}

func TestOptional(t *testing.T) {
	var f degrade.Feature
	calls := 0
	fetch := func(context.Context) ([]string, error) { calls++; return []string{"a", "b"}, nil }
	fallback := []string{}

	if v, used := degrade.Optional(context.Background(), &f, fallback, fetch); !used || len(v) != 2 {
		t.Fatalf("feature on: %v, %v", v, used)
	}
	f.Set(false)
	if v, used := degrade.Optional(context.Background(), &f, fallback, fetch); used || len(v) != 0 || calls != 1 {
		t.Fatalf("feature off: %v, %v, calls %d; want the fallback and no call", v, used, calls)
	}
	f.Set(true)
	failing := func(context.Context) ([]string, error) { return nil, errDown }
	if v, used := degrade.Optional(context.Background(), &f, fallback, failing); used || len(v) != 0 {
		t.Fatalf("fetch failed: %v, %v; want the fallback", v, used)
	}
}
