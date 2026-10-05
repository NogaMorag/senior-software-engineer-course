package domain

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// NewID returns a UUID version 7 (RFC 9562) in its canonical text form. The first 48 bits are the Unix
// time in milliseconds, so ids made later sort later, both as strings and as PostgreSQL uuid values,
// and B-tree inserts land at the right edge of the index instead of at random pages.
func NewID(now time.Time) string {
	var b [16]byte
	_, _ = rand.Read(b[:]) // crypto/rand.Read never fails
	ms := uint64(now.UnixMilli())
	for i := 0; i < 6; i++ {
		b[i] = byte(ms >> (40 - 8*i))
	}
	b[6] = b[6]&0x0f | 0x70 // version 7
	b[8] = b[8]&0x3f | 0x80 // variant 10xx
	var s [36]byte
	hex.Encode(s[0:8], b[0:4])
	s[8] = '-'
	hex.Encode(s[9:13], b[4:6])
	s[13] = '-'
	hex.Encode(s[14:18], b[6:8])
	s[18] = '-'
	hex.Encode(s[19:23], b[8:10])
	s[23] = '-'
	hex.Encode(s[24:], b[10:])
	return string(s[:])
}

// Now is the timestamp every resource stores: UTC, truncated to the microsecond that PostgreSQL keeps,
// so a value returned at creation equals the value read back later from either store.
func Now(clock func() time.Time) time.Time {
	return clock().UTC().Truncate(time.Microsecond)
}

// IsID reports whether s is a UUID in the canonical lowercase form NewID produces. Handlers check path
// and body ids with it first, so a malformed id is "not found" and never reaches a uuid column.
func IsID(s string) bool {
	if len(s) != 36 {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch i {
		case 8, 13, 18, 23:
			if c != '-' {
				return false
			}
		default:
			if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
				return false
			}
		}
	}
	return true
}
