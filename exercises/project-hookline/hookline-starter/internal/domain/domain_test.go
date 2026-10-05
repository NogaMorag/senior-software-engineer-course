package domain_test

import (
	"regexp"
	"strings"
	"testing"
	"time"

	"example.com/hookline/internal/domain"
)

var uuidV7 = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func TestNewIDIsAUUIDv7ThatSortsByTime(t *testing.T) {
	t0 := time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)
	a, b := domain.NewID(t0), domain.NewID(t0.Add(time.Millisecond))
	for _, id := range []string{a, b} {
		if !uuidV7.MatchString(id) {
			t.Fatalf("NewID = %q, want a canonical UUIDv7", id)
		}
	}
	if a >= b {
		t.Errorf("an id made 1ms later must sort later: %s >= %s", a, b)
	}
	if !strings.HasPrefix(a, "01a10bef-6200-7") { // 2026-10-05T12:00:00Z in Unix milliseconds, hex
		t.Errorf("NewID(%v) = %s; the first 48 bits must be the Unix time in milliseconds", t0, a)
	}
	if domain.NewID(t0) == a {
		t.Error("two ids in the same millisecond must differ (74 random bits)")
	}
}

func TestNowTruncatesToMicroseconds(t *testing.T) {
	in := time.Date(2026, 10, 5, 12, 0, 0, 123456789, time.FixedZone("IDT", 3*3600))
	got := domain.Now(func() time.Time { return in })
	if got.Nanosecond() != 123456000 || got.Location() != time.UTC {
		t.Errorf("Now = %v, want UTC truncated to the microsecond", got)
	}
}

func TestAPIKeys(t *testing.T) {
	key, hash := domain.NewAPIKey()
	if !domain.LooksLikeAPIKey(key) || !strings.HasPrefix(key, "hk_") || len(key) != 46 {
		t.Fatalf("NewAPIKey = %q, want hk_ and 43 base64url characters", key)
	}
	if domain.HashAPIKey(key) != hash {
		t.Error("the returned hash must be HashAPIKey(key)")
	}
	other, otherHash := domain.NewAPIKey()
	if other == key || otherHash == hash {
		t.Error("two keys must differ")
	}
	for _, bad := range []string{"", "hk_", "sk_" + key[3:], key + "x", key[:45] + "!"} {
		if domain.LooksLikeAPIKey(bad) {
			t.Errorf("LooksLikeAPIKey(%q) = true", bad)
		}
	}
	if s := domain.NewSigningSecret(); !strings.HasPrefix(s, "whsec_") || len(s) != 49 {
		t.Errorf("NewSigningSecret = %q, want whsec_ and 43 base64url characters", s)
	}
}

func TestValidateEndpointURL(t *testing.T) {
	tests := []struct {
		url     string
		dev     bool
		wantErr string // "" means valid; otherwise a substring of the reason
	}{
		{"https://example.com/webhooks", false, ""},
		{"https://example.com:8443/hooks?tenant=1", false, ""},
		{"http://localhost:8000/hooks", true, ""},
		{"http://sink:8000/hooks", true, ""},
		{"http://localhost:8000/hooks", false, "https"},
		{"", false, "required"},
		{"example.com/webhooks", false, "absolute"},
		{"https://", false, "absolute"},
		{"ftp://example.com/x", true, "https"},
		{"https://user:pass@example.com/", false, "credentials"},
		{"https://example.com/#frag", false, "fragment"},
		{"https://example.com/" + strings.Repeat("a", 2048), false, "2048"},
	}
	for _, tt := range tests {
		got := domain.ValidateEndpointURL(tt.url, tt.dev)
		if (tt.wantErr == "") != (got == "") || !strings.Contains(got, tt.wantErr) {
			t.Errorf("ValidateEndpointURL(%q, dev=%v) = %q, want %q", tt.url, tt.dev, got, tt.wantErr)
		}
	}
}

func TestValidateEventType(t *testing.T) {
	tests := []struct {
		typ      string
		wildcard bool
		valid    bool
	}{
		{"invoice.paid", false, true},
		{"shipment.status_changed", false, true},
		{"ping", false, true},
		{"*", true, true},
		{"*", false, false},
		{"", true, false},
		{"Invoice.Paid", false, false},
		{"invoice..paid", false, false},
		{".paid", false, false},
		{"invoice.*", true, false}, // only the whole-type wildcard exists
		{strings.Repeat("a", 129), false, false},
	}
	for _, tt := range tests {
		if got := domain.ValidateEventType(tt.typ, tt.wildcard); (got == "") != tt.valid {
			t.Errorf("ValidateEventType(%q, %v) = %q, want valid=%v", tt.typ, tt.wildcard, got, tt.valid)
		}
	}
}

func TestValidateTenantName(t *testing.T) {
	for name, valid := range map[string]bool{"Acme": true, "  ": false, "": false, strings.Repeat("א", 100): true, strings.Repeat("a", 101): false} {
		if got := domain.ValidateTenantName(name); (got == "") != valid {
			t.Errorf("ValidateTenantName(%q) = %q, want valid=%v", name, got, valid)
		}
	}
}

func TestIsID(t *testing.T) {
	good := domain.NewID(time.Now())
	tests := map[string]bool{
		good:                                   true,
		"00000000-0000-0000-0000-000000000000": true,
		strings.ToUpper(good):                  false, // canonical form only
		good[:35]:                              false,
		good + "0":                             false,
		strings.Replace(good, "-", "x", 1):     false,
		"ep_1":                                 false,
		"":                                     false,
	}
	for in, want := range tests {
		if got := domain.IsID(in); got != want {
			t.Errorf("IsID(%q) = %v, want %v", in, got, want)
		}
	}
}
