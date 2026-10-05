package jwtauth_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"example.com/ch28-security/internal/jwtauth"
)

const (
	issuer   = "https://idp.example.test/"
	audience = "projects-api"
)

// now is the frozen clock of every test, so expiry cases are exact.
var now = time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)

// testKeys holds one RSA and one EC key, generated once per test binary.
// The JWKS built from them is the "local static JWKS" of the tests.
var testKeys = sync.OnceValue(func() struct {
	rsa  *rsa.PrivateKey
	ec   *ecdsa.PrivateKey
	jwks []byte
} {
	rk, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	ek, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(err)
	}
	enc := base64.RawURLEncoding.EncodeToString
	point, err := ek.PublicKey.Bytes() // 0x04 || X || Y
	if err != nil {
		panic(err)
	}
	doc := map[string]any{"keys": []map[string]string{
		{"kty": "RSA", "kid": "rsa-1", "use": "sig", "alg": "RS256",
			"n": enc(rk.N.Bytes()), "e": enc(big.NewInt(int64(rk.E)).Bytes())},
		{"kty": "EC", "kid": "ec-1", "use": "sig", "alg": "ES256", "crv": "P-256",
			"x": enc(point[1:33]), "y": enc(point[33:])},
	}}
	data, err := json.Marshal(doc)
	if err != nil {
		panic(err)
	}
	return struct {
		rsa  *rsa.PrivateKey
		ec   *ecdsa.PrivateKey
		jwks []byte
	}{rk, ek, data}
})

func newVerifier(t *testing.T) *jwtauth.Verifier {
	t.Helper()
	keys, err := jwtauth.ParseJWKS(testKeys().jwks)
	if err != nil {
		t.Fatal(err)
	}
	v, err := jwtauth.NewVerifier(jwtauth.Config{
		Issuer:   issuer,
		Audience: audience,
		Methods:  []string{"RS256", "ES256"},
		Leeway:   30 * time.Second,
		Keys:     keys,
		Now:      func() time.Time { return now },
	})
	if err != nil {
		t.Fatal(err)
	}
	return v
}

// goodClaims are valid at `now`; each case changes one thing.
func goodClaims() jwt.MapClaims {
	return jwt.MapClaims{
		"iss":    issuer,
		"aud":    []string{audience},
		"sub":    "user-17",
		"tenant": "acme",
		"iat":    now.Add(-time.Minute).Unix(),
		"nbf":    now.Add(-time.Minute).Unix(),
		"exp":    now.Add(5 * time.Minute).Unix(),
	}
}

func sign(t *testing.T, method jwt.SigningMethod, kid string, claims jwt.MapClaims, key any) string {
	t.Helper()
	tok := jwt.NewWithClaims(method, claims)
	if kid != "" {
		tok.Header["kid"] = kid
	}
	s, err := tok.SignedString(key)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func with(change func(jwt.MapClaims)) jwt.MapClaims {
	c := goodClaims()
	change(c)
	return c
}

func TestVerify(t *testing.T) {
	k := testKeys()
	pubDER, err := x509.MarshalPKIXPublicKey(&k.rsa.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name  string
		token func(t *testing.T) string
		want  error // nil means the token must be accepted
	}{
		{"valid RS256", func(t *testing.T) string {
			return sign(t, jwt.SigningMethodRS256, "rsa-1", goodClaims(), k.rsa)
		}, nil},
		{"valid ES256", func(t *testing.T) string {
			return sign(t, jwt.SigningMethodES256, "ec-1", goodClaims(), k.ec)
		}, nil},
		{"alg none", func(t *testing.T) string {
			return sign(t, jwt.SigningMethodNone, "rsa-1", goodClaims(), jwt.UnsafeAllowNoneSignatureType)
		}, jwt.ErrTokenSignatureInvalid},
		{"alg confusion: HS256 keyed with the public key", func(t *testing.T) string {
			return sign(t, jwt.SigningMethodHS256, "rsa-1", goodClaims(), pubDER)
		}, jwt.ErrTokenSignatureInvalid},
		{"alg outside the allow-list (RS512)", func(t *testing.T) string {
			return sign(t, jwt.SigningMethodRS512, "rsa-1", goodClaims(), k.rsa)
		}, jwt.ErrTokenSignatureInvalid},
		{"wrong issuer", func(t *testing.T) string {
			return sign(t, jwt.SigningMethodRS256, "rsa-1", with(func(c jwt.MapClaims) { c["iss"] = "https://evil.example.test/" }), k.rsa)
		}, jwt.ErrTokenInvalidIssuer},
		{"token for another audience", func(t *testing.T) string {
			return sign(t, jwt.SigningMethodRS256, "rsa-1", with(func(c jwt.MapClaims) { c["aud"] = []string{"billing-api"} }), k.rsa)
		}, jwt.ErrTokenInvalidAudience},
		{"no audience", func(t *testing.T) string {
			return sign(t, jwt.SigningMethodRS256, "rsa-1", with(func(c jwt.MapClaims) { delete(c, "aud") }), k.rsa)
		}, jwt.ErrTokenRequiredClaimMissing},
		{"expired beyond the leeway", func(t *testing.T) string {
			return sign(t, jwt.SigningMethodRS256, "rsa-1", with(func(c jwt.MapClaims) { c["exp"] = now.Add(-31 * time.Second).Unix() }), k.rsa)
		}, jwt.ErrTokenExpired},
		{"expired within the leeway", func(t *testing.T) string {
			return sign(t, jwt.SigningMethodRS256, "rsa-1", with(func(c jwt.MapClaims) { c["exp"] = now.Add(-10 * time.Second).Unix() }), k.rsa)
		}, nil},
		{"no exp at all", func(t *testing.T) string {
			return sign(t, jwt.SigningMethodRS256, "rsa-1", with(func(c jwt.MapClaims) { delete(c, "exp") }), k.rsa)
		}, jwt.ErrTokenRequiredClaimMissing},
		{"not valid yet beyond the leeway", func(t *testing.T) string {
			return sign(t, jwt.SigningMethodRS256, "rsa-1", with(func(c jwt.MapClaims) { c["nbf"] = now.Add(2 * time.Minute).Unix() }), k.rsa)
		}, jwt.ErrTokenNotValidYet},
		{"issued in the future", func(t *testing.T) string {
			return sign(t, jwt.SigningMethodRS256, "rsa-1", with(func(c jwt.MapClaims) { c["iat"] = now.Add(time.Hour).Unix() }), k.rsa)
		}, jwt.ErrTokenUsedBeforeIssued},
		{"no kid", func(t *testing.T) string {
			return sign(t, jwt.SigningMethodRS256, "", goodClaims(), k.rsa)
		}, jwtauth.ErrNoKID},
		{"unknown kid", func(t *testing.T) string {
			return sign(t, jwt.SigningMethodRS256, "rsa-2", goodClaims(), k.rsa)
		}, jwtauth.ErrUnknownKID},
		{"kid of an EC key on an RS256 token", func(t *testing.T) string {
			return sign(t, jwt.SigningMethodRS256, "ec-1", goodClaims(), k.rsa)
		}, jwtauth.ErrKeyType},
		{"payload changed after signing", func(t *testing.T) string {
			s := sign(t, jwt.SigningMethodRS256, "rsa-1", goodClaims(), k.rsa)
			parts := strings.Split(s, ".")
			evil, _ := json.Marshal(with(func(c jwt.MapClaims) { c["tenant"] = "globex" }))
			parts[1] = base64.RawURLEncoding.EncodeToString(evil)
			return strings.Join(parts, ".")
		}, jwt.ErrTokenSignatureInvalid},
		{"no tenant claim", func(t *testing.T) string {
			return sign(t, jwt.SigningMethodRS256, "rsa-1", with(func(c jwt.MapClaims) { delete(c, "tenant") }), k.rsa)
		}, jwt.ErrTokenInvalidClaims},
		{"not a JWT", func(*testing.T) string { return "hello" }, jwt.ErrTokenMalformed},
	}
	v := newVerifier(t)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			claims, err := v.Verify(tc.token(t))
			if tc.want == nil {
				if err != nil {
					t.Fatalf("Verify: %v; want the token accepted", err)
				}
				if claims.Tenant != "acme" || claims.Subject != "user-17" {
					t.Fatalf("claims = %+v", claims)
				}
				return
			}
			if !errors.Is(err, tc.want) {
				t.Fatalf("Verify error = %v; want %v", err, tc.want)
			}
		})
	}
}

func TestNewVerifierRefusesUnsafeConfig(t *testing.T) {
	keys, err := jwtauth.ParseJWKS(testKeys().jwks)
	if err != nil {
		t.Fatal(err)
	}
	base := jwtauth.Config{Issuer: issuer, Audience: audience, Methods: []string{"RS256"}, Keys: keys}
	for name, change := range map[string]func(*jwtauth.Config){
		"no issuer":   func(c *jwtauth.Config) { c.Issuer = "" },
		"no audience": func(c *jwtauth.Config) { c.Audience = "" },
		"no methods":  func(c *jwtauth.Config) { c.Methods = nil },
		"HS256":       func(c *jwtauth.Config) { c.Methods = []string{"RS256", "HS256"} },
		"none":        func(c *jwtauth.Config) { c.Methods = []string{"none"} },
		"no keys":     func(c *jwtauth.Config) { c.Keys = nil },
	} {
		t.Run(name, func(t *testing.T) {
			cfg := base
			change(&cfg)
			if _, err := jwtauth.NewVerifier(cfg); err == nil {
				t.Fatal("NewVerifier accepted an unsafe config")
			}
		})
	}
}

func TestParseJWKSRejects(t *testing.T) {
	for name, doc := range map[string]string{
		"empty":          `{"keys":[]}`,
		"no kid":         `{"keys":[{"kty":"EC","crv":"P-256","x":"AA","y":"AA"}]}`,
		"encryption key": `{"keys":[{"kty":"RSA","kid":"a","use":"enc","n":"AQAB","e":"AQAB"}]}`,
		"short RSA":      `{"keys":[{"kty":"RSA","kid":"a","n":"AQAB","e":"AQAB"}]}`,
		"other curve":    `{"keys":[{"kty":"EC","kid":"a","crv":"P-384","x":"AA","y":"AA"}]}`,
		"symmetric":      `{"keys":[{"kty":"oct","kid":"a","k":"c2VjcmV0"}]}`,
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := jwtauth.ParseJWKS([]byte(doc)); err == nil {
				t.Fatal("ParseJWKS accepted it")
			}
		})
	}
}

func TestMiddleware(t *testing.T) {
	k := testKeys()
	v := newVerifier(t)
	h := v.Middleware(slog.New(slog.DiscardHandler))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "tenant="+jwtauth.FromContext(r.Context()).Tenant)
	}))
	good := sign(t, jwt.SigningMethodRS256, "rsa-1", goodClaims(), k.rsa)
	expired := sign(t, jwt.SigningMethodRS256, "rsa-1", with(func(c jwt.MapClaims) { c["exp"] = now.Add(-time.Hour).Unix() }), k.rsa)

	tests := []struct {
		name, header string
		status       int
		challenge    string // expected WWW-Authenticate, "" when none
		body         string
	}{
		{"no header", "", 401, `Bearer realm="projects"`, ""},
		{"basic auth", "Basic dXNlcjpwYXNz", 401, `Bearer realm="projects"`, ""},
		{"expired token", "Bearer " + expired, 401, `Bearer realm="projects", error="invalid_token"`, ""},
		{"good token", "Bearer " + good, 200, "", "tenant=acme"},
		{"scheme is case-insensitive", "bearer " + good, 200, "", "tenant=acme"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/v1/projects", nil)
			if tc.header != "" {
				req.Header.Set("Authorization", tc.header)
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if rec.Code != tc.status {
				t.Fatalf("status = %d; want %d", rec.Code, tc.status)
			}
			if got := rec.Header().Get("WWW-Authenticate"); got != tc.challenge {
				t.Errorf("WWW-Authenticate = %q; want %q", got, tc.challenge)
			}
			if tc.body != "" && rec.Body.String() != tc.body {
				t.Errorf("body = %q; want %q", rec.Body.String(), tc.body)
			}
		})
	}
}
