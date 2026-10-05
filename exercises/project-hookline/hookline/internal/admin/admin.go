// Package admin holds the operator commands of hookline-admin, as functions the tests can call.
package admin

import (
	"context"
	"errors"
	"fmt"
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
	if reason := domain.ValidateTenantName(name); reason != "" {
		return domain.Tenant{}, "", fmt.Errorf("tenant name %s", reason)
	}
	if ratePerSec < 1 || burst < 1 {
		return domain.Tenant{}, "", errors.New("rate and burst must be positive")
	}
	key, hash := domain.NewAPIKey()
	now := domain.Now(clock)
	t := domain.Tenant{ID: domain.NewID(now), Name: name, APIKeyHash: hash, RatePerSec: ratePerSec, Burst: burst, CreatedAt: now}
	if err := tenants.CreateTenant(ctx, t); err != nil {
		return domain.Tenant{}, "", fmt.Errorf("create tenant: %w", err)
	}
	return t, key, nil
}
