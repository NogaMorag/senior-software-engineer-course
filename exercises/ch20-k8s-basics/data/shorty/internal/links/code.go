package links

import (
	"crypto/rand"
	"fmt"
	"net/url"
)

const (
	alphabet   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	codeLength = 7
	maxURLLen  = 2048
)

// reserved codes would be shadowed by other routes on the same mux.
var reserved = map[string]bool{"healthz": true, "readyz": true, "v1": true, "debug": true}

// newCode returns a random code of codeLength characters from alphabet.
func newCode() string {
	out := make([]byte, 0, codeLength)
	buf := make([]byte, 16)
	for len(out) < codeLength {
		_, _ = rand.Read(buf) // never fails; see crypto/rand.Read
		for _, b := range buf {
			// 62*4 = 248: rejecting bytes >= 248 keeps every character equally likely.
			if b < 248 && len(out) < codeLength {
				out = append(out, alphabet[b%62])
			}
		}
	}
	return string(out)
}

// checkCode accepts 3 to 32 characters from [A-Za-z0-9_-], not reserved.
func checkCode(code string) error {
	if len(code) < 3 || len(code) > 32 {
		return fmt.Errorf("code must be 3 to 32 characters")
	}
	for _, c := range code {
		ok := c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_' || c == '-'
		if !ok {
			return fmt.Errorf("code may contain only letters, digits, '_' and '-'")
		}
	}
	if reserved[code] {
		return fmt.Errorf("code %q is reserved", code)
	}
	return nil
}

// checkURL accepts absolute http and https URLs up to maxURLLen bytes.
func checkURL(raw string) error {
	if raw == "" {
		return fmt.Errorf("url is required")
	}
	if len(raw) > maxURLLen {
		return fmt.Errorf("url is longer than %d bytes", maxURLLen)
	}
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return fmt.Errorf("url must be an absolute http or https URL")
	}
	return nil
}
