// Package jwtauth validates OIDC access tokens (JWTs) issued by one identity provider
// and puts the caller's identity into the request context.
package jwtauth

import (
	"context"
	"crypto/ecdsa"
	"crypto/rsa"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims are the token claims this service relies on.
type Claims struct {
	jwt.RegisteredClaims
	Tenant string `json:"tenant"`
	Scope  string `json:"scope,omitempty"`
}

// Validate is called by the parser after the registered claims pass,
// so a token without a tenant is rejected like an expired one.
func (c *Claims) Validate() error {
	if c.Tenant == "" {
		return errors.New("tenant claim is missing")
	}
	if c.Subject == "" {
		return errors.New("sub claim is missing")
	}
	return nil
}

// Config fixes everything a token must match. There are no defaults for
// Issuer, Audience or Methods: a verifier that accepts "any" is a bug.
type Config struct {
	Issuer   string        // exact "iss"
	Audience string        // must appear in "aud"
	Methods  []string      // allow-list of "alg" values, e.g. RS256, ES256
	Leeway   time.Duration // clock skew tolerated on exp, nbf and iat
	Keys     KeySet        // kid -> public key
	Now      func() time.Time
}

// Verifier checks tokens against a Config.
type Verifier struct {
	cfg    Config
	parser *jwt.Parser
}

var (
	// ErrNoKID is returned for a token whose header has no "kid".
	ErrNoKID = errors.New("token has no kid")
	// ErrUnknownKID is returned when the kid is not in the key set.
	ErrUnknownKID = errors.New("unknown kid")
	// ErrKeyType is returned when the key for the kid does not fit the algorithm.
	ErrKeyType = errors.New("key type does not match alg")
)

// NewVerifier checks the configuration and builds the parser once.
func NewVerifier(cfg Config) (*Verifier, error) {
	switch {
	case cfg.Issuer == "":
		return nil, errors.New("jwtauth: Issuer is required")
	case cfg.Audience == "":
		return nil, errors.New("jwtauth: Audience is required")
	case len(cfg.Methods) == 0:
		return nil, errors.New("jwtauth: Methods is required")
	case len(cfg.Keys) == 0:
		return nil, errors.New("jwtauth: Keys is required")
	}
	for _, m := range cfg.Methods {
		if m == "none" || strings.HasPrefix(m, "HS") {
			return nil, fmt.Errorf("jwtauth: method %q is not allowed for tokens from an identity provider", m)
		}
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	parser := jwt.NewParser(
		jwt.WithValidMethods(cfg.Methods),
		jwt.WithIssuer(cfg.Issuer),
		jwt.WithAudience(cfg.Audience),
		jwt.WithLeeway(cfg.Leeway),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
		jwt.WithTimeFunc(cfg.Now),
	)
	return &Verifier{cfg: cfg, parser: parser}, nil
}

// Verify parses raw, checks its signature and claims, and returns the claims.
func (v *Verifier) Verify(raw string) (*Claims, error) {
	claims := &Claims{}
	if _, err := v.parser.ParseWithClaims(raw, claims, v.keyFor); err != nil {
		return nil, err
	}
	return claims, nil
}

// keyFor picks the verification key by kid and refuses a key that does not
// belong to the algorithm family in the header. The parser has already
// rejected algorithms outside the allow-list before calling it.
func (v *Verifier) keyFor(t *jwt.Token) (any, error) {
	kid, _ := t.Header["kid"].(string)
	if kid == "" {
		return nil, ErrNoKID
	}
	key, ok := v.cfg.Keys[kid]
	if !ok {
		return nil, fmt.Errorf("%w %q", ErrUnknownKID, kid)
	}
	switch t.Method.(type) {
	case *jwt.SigningMethodRSA, *jwt.SigningMethodRSAPSS:
		if _, ok := key.(*rsa.PublicKey); !ok {
			return nil, ErrKeyType
		}
	case *jwt.SigningMethodECDSA:
		if _, ok := key.(*ecdsa.PublicKey); !ok {
			return nil, ErrKeyType
		}
	default:
		return nil, ErrKeyType
	}
	return key, nil
}

type ctxKey struct{}

// WithClaims returns a context that carries c. Middleware uses it; tests use it
// to call handlers as a given tenant.
func WithClaims(ctx context.Context, c *Claims) context.Context {
	return context.WithValue(ctx, ctxKey{}, c)
}

// FromContext returns the claims put there by Middleware, or nil.
func FromContext(ctx context.Context) *Claims {
	c, _ := ctx.Value(ctxKey{}).(*Claims)
	return c
}

// Middleware rejects requests without a valid bearer token with 401.
// The reason goes to the log, never to the client, and the token is never logged.
func (v *Verifier) Middleware(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw, ok := bearer(r.Header.Get("Authorization"))
			if !ok {
				unauthorized(w, "")
				return
			}
			claims, err := v.Verify(raw)
			if err != nil {
				log.InfoContext(r.Context(), "token rejected", slog.String("reason", err.Error()))
				unauthorized(w, "invalid_token")
				return
			}
			next.ServeHTTP(w, r.WithContext(WithClaims(r.Context(), claims)))
		})
	}
}

func bearer(h string) (string, bool) {
	scheme, token, ok := strings.Cut(h, " ")
	if !ok || !strings.EqualFold(scheme, "Bearer") || token == "" {
		return "", false
	}
	return token, true
}

// unauthorized follows RFC 6750: no error code when credentials are missing.
func unauthorized(w http.ResponseWriter, code string) {
	value := `Bearer realm="projects"`
	if code != "" {
		value += `, error="` + code + `"`
	}
	w.Header().Set("WWW-Authenticate", value)
	http.Error(w, "unauthorized", http.StatusUnauthorized)
}
