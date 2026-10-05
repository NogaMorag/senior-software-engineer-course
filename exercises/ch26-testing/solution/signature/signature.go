// Package signature signs and verifies Hookline webhook deliveries in Go, to the same rules as
// the chapter 8 Python receiver (exercises/ch08-receiver/src/receiver/signature.py):
//
//	Hookline-Signature: t=<unix seconds>,v1=<lowercase hex HMAC-SHA256(secret, "<t>." + raw body)>
//
// testdata/signature-vectors.json is a copy of exercises/ch08-receiver/data/signature-vectors.json;
// both implementations must pass it.
package signature

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// DefaultTolerance is how far the signed timestamp may be from the receiver's clock, either way.
const DefaultTolerance = 300 * time.Second

// maxTimestampDigits matches the receiver's [0-9]{1,12}: enough for unix seconds for 30,000 years.
const maxTimestampDigits = 12

// Reason says why verification failed; the values are the reasons in the test vectors.
type Reason string

const (
	Malformed               Reason = "malformed"
	TimestampOutOfTolerance Reason = "timestamp_out_of_tolerance"
	SignatureMismatch       Reason = "signature_mismatch"
)

// Error is returned by ParseHeader and Verify.
type Error struct {
	Reason Reason
	Detail string
}

func (e *Error) Error() string { return string(e.Reason) + ": " + e.Detail }

func fail(r Reason, format string, args ...any) error {
	return &Error{Reason: r, Detail: fmt.Sprintf(format, args...)}
}

// Digest is the raw 32-byte HMAC over the exact bytes "<t>." followed by the body.
func Digest(secret []byte, timestamp int64, body []byte) []byte {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(strconv.FormatInt(timestamp, 10) + "."))
	mac.Write(body)
	return mac.Sum(nil)
}

// Sign returns the header value for one delivery.
func Sign(secret []byte, timestamp int64, body []byte) string {
	return Header{Timestamp: timestamp, Signatures: [][]byte{Digest(secret, timestamp, body)}}.String()
}

// Header is a parsed Hookline-Signature value: one timestamp and one or more candidate
// signatures (more than one during a secret rotation).
type Header struct {
	Timestamp  int64
	Signatures [][]byte // each exactly sha256.Size bytes
}

// String formats h the way Sign does; ParseHeader(h.String()) returns h again.
func (h Header) String() string {
	var b strings.Builder
	b.WriteString("t=")
	b.WriteString(strconv.FormatInt(h.Timestamp, 10))
	for _, s := range h.Signatures {
		b.WriteString(",v1=")
		b.WriteString(hex.EncodeToString(s))
	}
	return b.String()
}

// ParseHeader splits the header into its timestamp and candidate signatures. Items are
// "key=value" separated by commas, with optional spaces; unknown keys are ignored.
func ParseHeader(header string) (Header, error) {
	if header == "" {
		return Header{}, fail(Malformed, "missing signature header")
	}
	var h Header
	timestamps := 0
	for item := range strings.SplitSeq(header, ",") {
		item = strings.TrimSpace(item)
		key, value, ok := strings.Cut(item, "=")
		if !ok {
			return Header{}, fail(Malformed, "item %q is not key=value", item)
		}
		switch key {
		case "t":
			timestamps++
			t, err := parseTimestamp(value)
			if err != nil {
				return Header{}, err
			}
			h.Timestamp = t
		case "v1":
			if len(value) != hex.EncodedLen(sha256.Size) {
				return Header{}, fail(Malformed, "v1 must be %d hex characters", hex.EncodedLen(sha256.Size))
			}
			sig, err := hex.DecodeString(value)
			if err != nil {
				return Header{}, fail(Malformed, "v1 is not hex")
			}
			h.Signatures = append(h.Signatures, sig)
		}
	}
	if timestamps != 1 {
		return Header{}, fail(Malformed, "exactly one t= is required, got %d", timestamps)
	}
	if len(h.Signatures) == 0 {
		return Header{}, fail(Malformed, "at least one v1= is required")
	}
	return h, nil
}

// parseTimestamp accepts 1 to 12 ASCII digits only. strconv.ParseInt alone would also accept
// "+5" and "-5", which the Python receiver rejects; the two sides must agree.
func parseTimestamp(s string) (int64, error) {
	if s == "" || len(s) > maxTimestampDigits {
		return 0, fail(Malformed, "t must be 1 to %d digits", maxTimestampDigits)
	}
	for i := range len(s) {
		if s[i] < '0' || s[i] > '9' {
			return 0, fail(Malformed, "t must be ASCII digits")
		}
	}
	return strconv.ParseInt(s, 10, 64) // cannot fail: at most 12 digits
}

// Verify returns nil when header carries a valid signature of body under secret, made within
// tolerance of now. Checks run in a fixed order: format, then freshness, then the HMAC itself,
// compared in constant time against every candidate.
func Verify(header string, body, secret []byte, now time.Time, tolerance time.Duration) error {
	h, err := ParseHeader(header)
	if err != nil {
		return err
	}
	age := now.Unix() - h.Timestamp
	if age > int64(tolerance/time.Second) || -age > int64(tolerance/time.Second) {
		return fail(TimestampOutOfTolerance, "t=%d is off by more than %s", h.Timestamp, tolerance)
	}
	want := Digest(secret, h.Timestamp, body)
	matched := false
	for _, candidate := range h.Signatures {
		// No early exit, so timing says nothing about which candidate matched.
		matched = hmac.Equal(want, candidate) || matched
	}
	if !matched {
		return fail(SignatureMismatch, "no v1 signature matches")
	}
	return nil
}
