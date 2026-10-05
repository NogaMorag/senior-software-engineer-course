// Package page implements cursor pagination: an opaque cursor that encodes the sort key of the last
// item a client has seen, and the limit rules shared by every list endpoint.
package page

import (
	"errors"
	"strconv"
	"time"
)

// Limits for the page size. A client that sends no limit gets DefaultLimit.
const (
	DefaultLimit = 20
	MaxLimit     = 100
)

// ErrInvalid is returned for any cursor the server did not produce, or produced for another version.
var ErrInvalid = errors.New("invalid cursor")

// Cursor is the sort key of the last item on a page. Lists are ordered by (CreatedAt, ID), newest
// first; ID breaks ties between items created in the same instant, so the order is total and stable.
type Cursor struct {
	CreatedAt time.Time
	ID        string
}

// Page is one page of a list response. NextCursor is empty on the last page.
type Page[T any] struct {
	Items      []T    `json:"items"`
	NextCursor string `json:"next_cursor,omitempty"`
}

// wire is the cursor's JSON form before base64. V lets a later version change the format and still
// reject old cursors cleanly instead of misreading them.
type wire struct {
	V  int    `json:"v"`
	T  int64  `json:"t"`
	ID string `json:"id"`
}

const version = 1

// Encode turns c into an opaque, URL-safe string. Clients must not parse it; only round-trip it.
func Encode(c Cursor) string {
	// TODO(day 20): marshal wire{V: version, T: unix nanoseconds, ID} and base64url-encode it without padding.
	panic("TODO(day 20): page.Encode")
}

// Decode parses a cursor produced by Encode. Anything else is ErrInvalid.
func Decode(s string) (Cursor, error) {
	// TODO(day 20): reverse Encode. Any failure (bad base64, bad JSON, wrong version, empty id,
	// non-positive time) returns ErrInvalid.
	panic("TODO(day 20): page.Decode")
}

// ParseLimit reads the limit query parameter: empty means DefaultLimit; otherwise 1..MaxLimit.
func ParseLimit(s string) (int, error) {
	if s == "" {
		return DefaultLimit, nil
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 || n > MaxLimit {
		return 0, errors.New("limit must be an integer from 1 to " + strconv.Itoa(MaxLimit))
	}
	return n, nil
}
