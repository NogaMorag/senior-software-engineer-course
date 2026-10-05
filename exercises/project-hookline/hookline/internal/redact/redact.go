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
	"context"
	"fmt"
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

// secretShapes match values that are secrets whatever the key is called: bearer tokens, Hookline's
// API keys (hk_ and 43 base64url characters) and signing secrets (whsec_ and 43), and private keys.
var secretShapes = regexp.MustCompile(`(?i)\bbearer\s+\S+|\bhk_[A-Za-z0-9_-]{43}|\bwhsec_[A-Za-z0-9_-]{43}|-----BEGIN [A-Z ]*PRIVATE KEY-----`)

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

// Handler wraps next so that every record passes the same rules as ReplaceAttr before next sees it:
// the attributes by key and by shape, and the message by shape (milestone M9). Hookline's logger puts
// it in front of every output, the JSON lines and the OpenTelemetry bridge to Loki alike, because
// ReplaceAttr is an option of slog's own handlers that the bridge does not have.
func Handler(next slog.Handler) slog.Handler { return handler{next} }

type handler struct{ next slog.Handler }

func (h handler) Enabled(ctx context.Context, l slog.Level) bool { return h.next.Enabled(ctx, l) }

func (h handler) Handle(ctx context.Context, r slog.Record) error {
	out := slog.NewRecord(r.Time, r.Level, secretShapes.ReplaceAllString(r.Message, Redacted), r.PC)
	r.Attrs(func(a slog.Attr) bool {
		out.AddAttrs(redactAttr(a))
		return true
	})
	return h.next.Handle(ctx, out)
}

func (h handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	clean := make([]slog.Attr, len(attrs))
	for i, a := range attrs {
		clean[i] = redactAttr(a)
	}
	return handler{h.next.WithAttrs(clean)}
}

func (h handler) WithGroup(name string) slog.Handler { return handler{h.next.WithGroup(name)} }

// redactAttr is ReplaceAttr applied recursively, after resolving LogValuers (so a Secret becomes its
// placeholder before anything else looks at it). A map or struct value, which ReplaceAttr cannot see
// into, is checked in its printed form: if a secret shape shows there, the value is logged as that
// printed form with the secret replaced.
func redactAttr(a slog.Attr) slog.Attr {
	a.Value = a.Value.Resolve()
	if a.Value.Kind() == slog.KindAny {
		if printed := fmt.Sprint(a.Value.Any()); secretShapes.MatchString(printed) {
			return slog.String(a.Key, secretShapes.ReplaceAllString(printed, Redacted))
		}
	}
	if a.Value.Kind() != slog.KindGroup {
		return ReplaceAttr(nil, a)
	}
	members := a.Value.Group()
	clean := make([]slog.Attr, len(members))
	for i, m := range members {
		clean[i] = redactAttr(m)
	}
	return slog.Attr{Key: a.Key, Value: slog.GroupValue(clean...)}
}
