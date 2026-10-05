// Package admin holds the operator commands of hookline-admin, as functions the tests can call.
package admin

import (
	"context"

	"time"

	"example.com/hookline/internal/domain"
	"example.com/hookline/internal/store"
)

// Default ingest limits of a new tenant; milestone M6 enforces them.
const (
	DefaultRatePerSec = 50
	DefaultBurst      = 100
)

// CreateTenant creates a tenant with a fresh API key and returns both. The key exists only in the return
// value: the store keeps its hash, so the caller must show the key to the operator now or never.
func CreateTenant(ctx context.Context, tenants store.Tenants, name string, ratePerSec, burst int, clock func() time.Time) (domain.Tenant, string, error) {
	panic("TODO: see chapter 38")

}
