// Package apikey issues API keys and verifies them against a stored hash.
//
// A key looks like "<prefix>_<id>_<secret>":
//
//	hl_4f9c2a7d1e03_mJ4Q...43 base64url characters...
//
// The prefix says what the string is (scanners and humans recognise it), the
// id is public and finds the stored row, and the secret is 32 random bytes.
// Only SHA-256 of the whole key is stored; the key itself is shown once.
package apikey

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"regexp"
	"strings"
	"sync"
)

const (
	idBytes     = 6  // 12 hex characters, public
	secretBytes = 32 // 256 bits from crypto/rand
)

var (
	// ErrMalformed means the string cannot be a key of this format.
	ErrMalformed = errors.New("apikey: malformed key")
	// ErrInvalid means the key is well formed but not valid.
	ErrInvalid = errors.New("apikey: invalid key")

	prefixRE = regexp.MustCompile(`^[a-z]{2,8}$`)
)

// Stored is what the database keeps for a key: never the key itself.
type Stored struct {
	ID    string   // public id, unique; the lookup key
	Hash  [32]byte // SHA-256 of the full key
	Owner string   // tenant the key authenticates
}

// Generate makes a new key for owner. The plaintext is returned once, to be
// shown to the user; only the Stored value may be persisted.
func Generate(prefix, owner string) (plaintext string, stored Stored, err error) {
	if !prefixRE.MatchString(prefix) {
		return "", Stored{}, errors.New("apikey: prefix must be 2 to 8 lowercase letters")
	}
	buf := make([]byte, idBytes+secretBytes)
	if _, err := rand.Read(buf); err != nil { // never fails on supported platforms; checked anyway
		return "", Stored{}, err
	}
	id := hex.EncodeToString(buf[:idBytes])
	secret := base64.RawURLEncoding.EncodeToString(buf[idBytes:])
	plaintext = prefix + "_" + id + "_" + secret
	return plaintext, Stored{ID: id, Hash: sha256.Sum256([]byte(plaintext)), Owner: owner}, nil
}

// Parse checks the format and returns the public id.
func Parse(prefix, key string) (id string, err error) {
	rest, ok := strings.CutPrefix(key, prefix+"_")
	if !ok {
		return "", ErrMalformed
	}
	id, secret, ok := strings.Cut(rest, "_")
	if !ok || len(id) != 2*idBytes || len(secret) != base64.RawURLEncoding.EncodedLen(secretBytes) {
		return "", ErrMalformed
	}
	if _, err := hex.DecodeString(id); err != nil {
		return "", ErrMalformed
	}
	if _, err := base64.RawURLEncoding.DecodeString(secret); err != nil {
		return "", ErrMalformed
	}
	return id, nil
}

// Matches reports whether key hashes to s.Hash, in constant time.
func (s Stored) Matches(key string) bool {
	sum := sha256.Sum256([]byte(key))
	return subtle.ConstantTimeCompare(sum[:], s.Hash[:]) == 1
}

// Store is an in-memory table of stored keys, indexed by public id,
// the way a database table would be (unique index on id).
type Store struct {
	prefix string
	mu     sync.RWMutex
	byID   map[string]Stored
}

// NewStore returns an empty store for keys with the given prefix.
func NewStore(prefix string) *Store {
	return &Store{prefix: prefix, byID: map[string]Stored{}}
}

// Add saves a stored key.
func (s *Store) Add(k Stored) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.byID[k.ID] = k
}

// Revoke removes a key by id.
func (s *Store) Revoke(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.byID, id)
}

// Authenticate returns the owner of key. Every failure is ErrInvalid or
// ErrMalformed, so a caller cannot learn whether an id exists.
func (s *Store) Authenticate(key string) (owner string, err error) {
	id, err := Parse(s.prefix, key)
	if err != nil {
		return "", err
	}
	s.mu.RLock()
	k, found := s.byID[id]
	s.mu.RUnlock()
	if !found || !k.Matches(key) {
		return "", ErrInvalid
	}
	return k.Owner, nil
}
