package page_test

import (
	"testing"
	"time"

	"pgregory.net/rapid"

	"example.com/hookline/internal/page"
)

// Round trip (milestone M9): every cursor the API hands out decodes to the same position, whatever the
// time and id, and is safe to put in a URL without escaping. Times start one second after the epoch:
// Decode reads t <= 0 as a missing time, and no item is created in 1970 (rapid found that edge first).
func TestPropertyCursorRoundTrip(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		c := page.Cursor{
			CreatedAt: time.Unix(rapid.Int64Range(1, 4102444800).Draw(t, "unix"), rapid.Int64Range(0, 999_999_999).Draw(t, "nanos")).UTC(),
			ID:        rapid.StringN(1, 40, 80).Draw(t, "id"),
		}
		s := page.Encode(c)
		for _, ch := range s {
			if ch == '+' || ch == '/' || ch == '=' {
				t.Fatalf("cursor %q is not URL-safe", s)
			}
		}
		got, err := page.Decode(s)
		if err != nil {
			t.Fatalf("Decode(Encode(%+v)) error: %v", c, err)
		}
		if !got.CreatedAt.Equal(c.CreatedAt) || got.ID != c.ID {
			t.Fatalf("round trip = %+v, want %+v", got, c)
		}
	})
}

// A client can send anything as a cursor: Decode never panics, and either refuses it or returns a
// position that encodes again.
func TestPropertyDecodeNeverPanics(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		s := rapid.OneOf(rapid.String(), rapid.StringMatching(`[A-Za-z0-9_-]{0,60}`)).Draw(t, "cursor")
		c, err := page.Decode(s)
		if err != nil {
			if err != page.ErrInvalid {
				t.Fatalf("Decode(%q) error = %v, want ErrInvalid", s, err)
			}
			return
		}
		if _, err := page.Decode(page.Encode(c)); err != nil {
			t.Fatalf("Decode(%q) = %+v, which does not encode again: %v", s, c, err)
		}
	})
}
