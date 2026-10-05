package apikey_test

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"
	"testing"

	"example.com/ch28-security/internal/apikey"
)

func TestGenerateFormat(t *testing.T) {
	key, stored, err := apikey.Generate("hl", "acme")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(key, "hl_"+stored.ID+"_") {
		t.Fatalf("key %q does not start with hl_<id>_", key)
	}
	if want := len("hl_") + 12 + 1 + 43; len(key) != want {
		t.Fatalf("len(key) = %d; want %d", len(key), want)
	}
	if id, err := apikey.Parse("hl", key); err != nil || id != stored.ID {
		t.Fatalf("Parse = %q, %v; want %q", id, err, stored.ID)
	}
	if stored.Owner != "acme" {
		t.Fatalf("owner = %q", stored.Owner)
	}
}

func TestStoredValueHoldsNoSecret(t *testing.T) {
	key, stored, err := apikey.Generate("hl", "acme")
	if err != nil {
		t.Fatal(err)
	}
	if stored.Hash != sha256.Sum256([]byte(key)) {
		t.Fatal("Hash is not SHA-256 of the full key")
	}
	secret := key[len("hl_")+12+1:]
	if dump := fmt.Sprintf("%+v %x", stored, stored.Hash); strings.Contains(dump, secret) {
		t.Fatalf("stored value contains the secret: %s", dump)
	}
}

func TestKeysAreUnique(t *testing.T) {
	seen := map[string]bool{}
	for range 1000 {
		key, _, err := apikey.Generate("hl", "acme")
		if err != nil {
			t.Fatal(err)
		}
		if seen[key] {
			t.Fatalf("duplicate key %q", key)
		}
		seen[key] = true
	}
}

func TestGenerateRejectsBadPrefix(t *testing.T) {
	for _, p := range []string{"", "h", "HL", "hl_x", "toolongprefix"} {
		if _, _, err := apikey.Generate(p, "acme"); err == nil {
			t.Errorf("Generate(%q) accepted the prefix", p)
		}
	}
}

func TestAuthenticate(t *testing.T) {
	store := apikey.NewStore("hl")
	key, stored, err := apikey.Generate("hl", "acme")
	if err != nil {
		t.Fatal(err)
	}
	store.Add(stored)
	other, _, err := apikey.Generate("hl", "globex") // generated, never stored
	if err != nil {
		t.Fatal(err)
	}
	secretAt := len("hl_") + 12 + 1 // the secret may itself contain '_'
	changed := []byte(key)
	if changed[secretAt] == 'A' {
		changed[secretAt] = 'B'
	} else {
		changed[secretAt] = 'A'
	}

	tests := []struct {
		name, key string
		owner     string
		err       error
	}{
		{"the issued key", key, "acme", nil},
		{"one secret character changed", string(changed), "", apikey.ErrInvalid},
		{"right id, other secret", key[:secretAt] + other[secretAt:], "", apikey.ErrInvalid},
		{"unknown id", other, "", apikey.ErrInvalid},
		{"wrong prefix", "sk" + key[2:], "", apikey.ErrMalformed},
		{"truncated", key[:len(key)-5], "", apikey.ErrMalformed},
		{"empty", "", "", apikey.ErrMalformed},
		{"the hash instead of the key", fmt.Sprintf("%x", stored.Hash), "", apikey.ErrMalformed},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			owner, err := store.Authenticate(tc.key)
			if owner != tc.owner || !errors.Is(err, tc.err) {
				t.Fatalf("Authenticate = %q, %v; want %q, %v", owner, err, tc.owner, tc.err)
			}
		})
	}

	store.Revoke(stored.ID)
	if _, err := store.Authenticate(key); !errors.Is(err, apikey.ErrInvalid) {
		t.Fatalf("revoked key: err = %v; want ErrInvalid", err)
	}
}
