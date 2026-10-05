// Package redact keeps secrets out of logs written with log/slog.
//
// Three tools, from strongest to weakest:
//   - Secret, a string type whose every printed form is "[REDACTED]";
//   - ReplaceAttr, a slog.HandlerOptions hook that redacts by key name and by
//     value shape (bearer tokens, API keys);
//   - Header, which logs an http.Header with credential headers redacted,
//     because ReplaceAttr does not look inside maps.
package redact

import (
	"log/slog"
	"net/http"
	"regexp"
	"slices"
	"strings"
)

// Redacted replaces every secret value.
const Redacted = "[REDACTED]"

// Secret is a string that never prints itself. Use it for fields that hold
// credentials, so a struct logged by mistake stays safe.
type Secret string

// LogValue implements slog.LogValuer.
func (Secret) LogValue() slog.Value { return slog.StringValue(Redacted) }

// String implements fmt.Stringer, for %v and %s.
func (Secret) String() string { return Redacted }

// GoString implements fmt.GoStringer, for %#v.
func (Secret) GoString() string { return Redacted }

// MarshalText makes encoding/json and friends print the placeholder too.
func (Secret) MarshalText() ([]byte, error) { return []byte(Redacted), nil }

// Reveal returns the value, for the one place that needs it.
func (s Secret) Reveal() string { return string(s) }

// sensitiveKeys are attribute keys whose values are always redacted,
// compared after lowercasing and turning '-' into '_'.
var sensitiveKeys = []string{
	"authorization", "proxy_authorization", "cookie", "set_cookie",
	"password", "passwd", "secret", "client_secret", "signing_secret",
	"token", "access_token", "refresh_token", "id_token",
	"api_key", "apikey", "x_api_key", "private_key", "dsn", "database_url",
}

// secretShapes match values that are secrets whatever the key is called.
var secretShapes = regexp.MustCompile(`(?i)\bbearer\s+\S+|\bhl_[0-9a-f]{12}_[A-Za-z0-9_-]{43}\b|-----BEGIN [A-Z ]*PRIVATE KEY-----`)

func sensitive(key string) bool {
	k := strings.ReplaceAll(strings.ToLower(key), "-", "_")
	return slices.Contains(sensitiveKeys, k) || strings.HasSuffix(k, "_password") || strings.HasSuffix(k, "_secret") || strings.HasSuffix(k, "_token")
}

// ReplaceAttr is a slog.HandlerOptions.ReplaceAttr hook. slog calls it for
// every attribute, including those inside groups, after LogValuer values have
// been resolved; it does not descend into maps or structs.
func ReplaceAttr(_ []string, a slog.Attr) slog.Attr {
	if a.Value.Kind() == slog.KindGroup {
		return a // slog calls ReplaceAttr again for each member
	}
	if sensitive(a.Key) {
		return slog.String(a.Key, Redacted)
	}
	if a.Value.Kind() == slog.KindString && secretShapes.MatchString(a.Value.String()) {
		return slog.String(a.Key, secretShapes.ReplaceAllString(a.Value.String(), Redacted))
	}
	return a
}

// Header returns h as a slog group with credential headers redacted.
func Header(key string, h http.Header) slog.Attr {
	names := make([]string, 0, len(h))
	for name := range h {
		names = append(names, name)
	}
	slices.Sort(names)
	attrs := make([]any, 0, len(names))
	for _, name := range names {
		value := strings.Join(h.Values(name), ", ")
		if sensitive(name) {
			value = Redacted
		}
		attrs = append(attrs, slog.String(name, value))
	}
	return slog.Group(key, attrs...)
}
