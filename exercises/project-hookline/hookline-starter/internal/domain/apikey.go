package domain

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"strings"
)

// Prefixes make a leaked credential recognisable (secret scanners match them) and tell the two kinds apart.
const (
	APIKeyPrefix = "hk_"
	SecretPrefix = "whsec_"
)

// APIKeyHash is the SHA-256 of an API key, the only form in which a key is stored.
type APIKeyHash [32]byte

// NewAPIKey returns a fresh API key and its hash. The key is shown to the operator once; the hash goes to
// the database. 32 random bytes are 256 bits of entropy, so a fast hash is enough: nobody can guess the
// preimage, which is what a slow hash such as argon2 protects for low-entropy passwords.
func NewAPIKey() (string, APIKeyHash) {
	key := APIKeyPrefix + random32()
	return key, HashAPIKey(key)
}

// HashAPIKey returns the stored form of key.
func HashAPIKey(key string) APIKeyHash {
	return sha256.Sum256([]byte(key))
}

// LooksLikeAPIKey is a cheap shape check that lets the auth middleware reject garbage without a lookup.
func LooksLikeAPIKey(s string) bool {
	rest, ok := strings.CutPrefix(s, APIKeyPrefix)
	if !ok || len(rest) != 43 {
		return false
	}
	_, err := base64.RawURLEncoding.DecodeString(rest)
	return err == nil
}

// NewSigningSecret returns a fresh endpoint signing secret (milestone M5 signs with it).
func NewSigningSecret() string {
	return SecretPrefix + random32()
}

func random32() string {
	var b [32]byte
	_, _ = rand.Read(b[:]) // crypto/rand.Read never fails
	return base64.RawURLEncoding.EncodeToString(b[:])
}
