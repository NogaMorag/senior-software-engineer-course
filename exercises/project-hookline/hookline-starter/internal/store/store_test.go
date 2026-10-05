package store_test

import (
	"context"
	"errors"
	"testing"

	"example.com/hookline/internal/domain"
	"example.com/hookline/internal/store"
	"example.com/hookline/internal/store/memory"
)

// The interfaces are the contract of milestone M1: code that holds only a store.Store works.
func TestUseThroughTheInterface(t *testing.T) {
	var s store.Store = memory.New()
	_, err := s.TenantByKeyHash(context.Background(), domain.HashAPIKey("hk_nobody"))
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("unknown key: %v, want domain.ErrNotFound", err)
	}
}
