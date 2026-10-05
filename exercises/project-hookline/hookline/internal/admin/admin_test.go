package admin_test

import (
	"context"
	"testing"
	"time"

	"example.com/hookline/internal/admin"
	"example.com/hookline/internal/domain"
	"example.com/hookline/internal/store/memory"
)

func TestCreateTenantStoresOnlyTheHash(t *testing.T) {
	st := memory.New()
	tn, key, err := admin.CreateTenant(context.Background(), st, "Acme", admin.DefaultRatePerSec, admin.DefaultBurst, time.Now)
	if err != nil {
		t.Fatal(err)
	}
	if !domain.LooksLikeAPIKey(key) || tn.APIKeyHash != domain.HashAPIKey(key) {
		t.Fatalf("key %q, hash %x", key, tn.APIKeyHash)
	}
	got, err := st.TenantByKeyHash(context.Background(), domain.HashAPIKey(key))
	if err != nil || got.ID != tn.ID || got.RatePerSec != 50 || got.Burst != 100 {
		t.Fatalf("lookup by the key's hash = %+v, %v", got, err)
	}
}

func TestCreateTenantValidates(t *testing.T) {
	for _, tt := range []struct {
		name        string
		rate, burst int
	}{{"", 50, 100}, {"Acme", 0, 100}, {"Acme", 50, -1}} {
		if _, _, err := admin.CreateTenant(context.Background(), memory.New(), tt.name, tt.rate, tt.burst, time.Now); err == nil {
			t.Errorf("CreateTenant(%q, %d, %d) succeeded", tt.name, tt.rate, tt.burst)
		}
	}
}
