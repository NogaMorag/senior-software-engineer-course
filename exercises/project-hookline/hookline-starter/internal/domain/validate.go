package domain

import (
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"
)

// Input limits. They are part of the contract: api/openapi.yaml states the same numbers.
const (
	MaxURLLength       = 2048
	MaxEventTypeLength = 128
	MaxTenantName      = 100
)

// eventTypePattern: dot-separated lowercase segments, such as invoice.paid or shipment.status_changed.
var eventTypePattern = regexp.MustCompile(`^[a-z][a-z0-9_]*(\.[a-z][a-z0-9_]*)*$`)

// ValidateEndpointURL returns "" for an acceptable receiver URL, or why it is not. Production accepts only
// absolute https URLs; dev mode also accepts http (localhost, or a compose service such as the sink).
// Resolving the host and rejecting private addresses (SSRF) is milestone M9's job, at delivery time.
func ValidateEndpointURL(raw string, devMode bool) string {
	if raw == "" {
		return "is required"
	}
	if len(raw) > MaxURLLength {
		return "must be at most 2048 bytes"
	}
	u, err := url.Parse(raw)
	if err != nil || !u.IsAbs() || u.Host == "" || u.Hostname() == "" {
		return "must be an absolute URL such as https://example.com/webhooks"
	}
	switch {
	case u.Scheme == "https":
	case u.Scheme == "http" && devMode:
	case u.Scheme == "http":
		return "must use https (http is accepted only in dev mode)"
	default:
		return "must use https"
	}
	if u.User != nil {
		return "must not contain credentials"
	}
	if u.Fragment != "" || strings.Contains(raw, "#") {
		return "must not contain a fragment"
	}
	return ""
}

// ValidateEventType returns "" for a valid event type, or why it is not. A subscription may also use
// AllTypes; an ingested event may not (allowWildcard false).
func ValidateEventType(t string, allowWildcard bool) string {
	switch {
	case t == AllTypes && allowWildcard:
		return ""
	case t == "":
		return "is required"
	case len(t) > MaxEventTypeLength:
		return "must be at most 128 characters"
	case !eventTypePattern.MatchString(t):
		return "must be dot-separated lowercase words, such as invoice.paid"
	}
	return ""
}

// ValidateTenantName returns "" for a valid tenant name, or why it is not.
func ValidateTenantName(name string) string {
	switch {
	case strings.TrimSpace(name) == "":
		return "is required"
	case !utf8.ValidString(name) || utf8.RuneCountInString(name) > MaxTenantName:
		return "must be valid UTF-8 of at most 100 characters"
	}
	return ""
}
