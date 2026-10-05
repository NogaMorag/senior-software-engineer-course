package signature_test

import (
	"errors"
	"testing"
	"time"

	"pgregory.net/rapid"

	"example.com/ch26-testing/signature"
)

// The largest timestamp the header format allows: 12 digits.
const maxTimestamp = 999_999_999_999

func reasonOf(err error) signature.Reason {
	var serr *signature.Error
	if errors.As(err, &serr) {
		return serr.Reason
	}
	return ""
}

// Round trip: whatever the secret, body and time, what Sign produces Verify accepts, and
// ParseHeader returns the same timestamp and digest.
func TestPropertySignVerifyRoundTrip(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		secret := rapid.SliceOfN(rapid.Byte(), 1, 64).Draw(t, "secret")
		body := rapid.SliceOf(rapid.Byte()).Draw(t, "body")
		ts := rapid.Int64Range(0, maxTimestamp).Draw(t, "timestamp")
		// Any receiver clock within the tolerance, on either side.
		skew := rapid.Int64Range(-300, 300).Draw(t, "skew")

		header := signature.Sign(secret, ts, body)
		if err := signature.Verify(header, body, secret, time.Unix(ts+skew, 0), signature.DefaultTolerance); err != nil {
			t.Fatalf("Verify(Sign(...)) = %v, header %q", err, header)
		}
		h, err := signature.ParseHeader(header)
		if err != nil {
			t.Fatalf("ParseHeader(%q) = %v", header, err)
		}
		if h.Timestamp != ts || len(h.Signatures) != 1 || string(h.Signatures[0]) != string(signature.Digest(secret, ts, body)) {
			t.Fatalf("ParseHeader(%q) = %+v, want timestamp %d and the digest", header, h, ts)
		}
	})
}

// Tampering: changing any one byte of the body, or using another secret, is a mismatch,
// never a pass and never a different reason.
func TestPropertyTamperingIsDetected(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		secret := rapid.SliceOfN(rapid.Byte(), 1, 64).Draw(t, "secret")
		body := rapid.SliceOfN(rapid.Byte(), 1, 256).Draw(t, "body")
		ts := rapid.Int64Range(0, maxTimestamp).Draw(t, "timestamp")
		header := signature.Sign(secret, ts, body)
		now := time.Unix(ts, 0)

		i := rapid.IntRange(0, len(body)-1).Draw(t, "index")
		flip := rapid.ByteRange(1, 255).Draw(t, "xor")
		tampered := append([]byte(nil), body...)
		tampered[i] ^= flip
		if r := reasonOf(signature.Verify(header, tampered, secret, now, signature.DefaultTolerance)); r != signature.SignatureMismatch {
			t.Fatalf("body byte %d changed: reason %q, want %q", i, r, signature.SignatureMismatch)
		}

		other := rapid.SliceOfN(rapid.Byte(), 1, 64).Filter(func(s []byte) bool { return string(s) != string(secret) }).Draw(t, "other secret")
		if r := reasonOf(signature.Verify(header, body, other, now, signature.DefaultTolerance)); r != signature.SignatureMismatch {
			t.Fatalf("other secret: reason %q, want %q", r, signature.SignatureMismatch)
		}
	})
}
