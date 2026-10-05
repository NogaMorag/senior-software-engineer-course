package jwtauth

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
)

// KeySet maps a key id (the JWT header "kid") to a public key.
type KeySet map[string]crypto.PublicKey

// jwk is the subset of RFC 7517 that this package accepts: RSA and P-256 signing keys.
type jwk struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
	Crv string `json:"crv"`
	X   string `json:"x"`
	Y   string `json:"y"`
}

// ParseJWKS reads a JSON Web Key Set ({"keys": [...]}) into a KeySet.
// Every key needs a unique kid; keys marked for encryption are refused.
func ParseJWKS(data []byte) (KeySet, error) {
	var doc struct {
		Keys []jwk `json:"keys"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("jwks: %w", err)
	}
	if len(doc.Keys) == 0 {
		return nil, errors.New("jwks: no keys")
	}
	set := make(KeySet, len(doc.Keys))
	for i, k := range doc.Keys {
		if k.Kid == "" {
			return nil, fmt.Errorf("jwks: key %d has no kid", i)
		}
		if _, dup := set[k.Kid]; dup {
			return nil, fmt.Errorf("jwks: duplicate kid %q", k.Kid)
		}
		if k.Use != "" && k.Use != "sig" {
			return nil, fmt.Errorf("jwks: key %q is not a signing key (use %q)", k.Kid, k.Use)
		}
		pub, err := k.publicKey()
		if err != nil {
			return nil, fmt.Errorf("jwks: key %q: %w", k.Kid, err)
		}
		set[k.Kid] = pub
	}
	return set, nil
}

func (k jwk) publicKey() (crypto.PublicKey, error) {
	switch k.Kty {
	case "RSA":
		n, err := b64(k.N)
		if err != nil {
			return nil, fmt.Errorf("n: %w", err)
		}
		e, err := b64(k.E)
		if err != nil {
			return nil, fmt.Errorf("e: %w", err)
		}
		if len(n)*8 < 2048 {
			return nil, fmt.Errorf("RSA key is %d bits; at least 2048 required", len(n)*8)
		}
		if len(e) == 0 || len(e) > 4 {
			return nil, errors.New("bad RSA exponent")
		}
		return &rsa.PublicKey{N: new(big.Int).SetBytes(n), E: int(new(big.Int).SetBytes(e).Int64())}, nil
	case "EC":
		if k.Crv != "P-256" {
			return nil, fmt.Errorf("curve %q not supported", k.Crv)
		}
		x, err := b64(k.X)
		if err != nil {
			return nil, fmt.Errorf("x: %w", err)
		}
		y, err := b64(k.Y)
		if err != nil {
			return nil, fmt.Errorf("y: %w", err)
		}
		if len(x) != 32 || len(y) != 32 {
			return nil, errors.New("P-256 coordinates must be 32 bytes each")
		}
		// Uncompressed point encoding: 0x04 || X || Y. The parser checks the point is on the curve.
		return ecdsa.ParseUncompressedPublicKey(elliptic.P256(), append(append([]byte{4}, x...), y...))
	default:
		return nil, fmt.Errorf("key type %q not supported", k.Kty)
	}
}

func b64(s string) ([]byte, error) { return base64.RawURLEncoding.DecodeString(s) }
