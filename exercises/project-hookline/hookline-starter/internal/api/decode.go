package api

import (
	"io"

	"net/http"

	"example.com/hookline/internal/page"
)

// limitBody caps the request body before anything reads it (chapter 9). Reading past the cap fails with
// *http.MaxBytesError, which decodeJSON and readBody turn into 413.
func limitBody(maxBytes int64, next http.Handler) http.Handler {
	panic("TODO: see chapter 38")

}

// isJSON checks the Content-Type; on failure it has written the 415 and returns false.
func isJSON(w http.ResponseWriter, r *http.Request) bool {
	panic("TODO: see chapter 38")

}

// decodeJSON reads exactly one JSON object from the request body into dst.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	panic("TODO: see chapter 38")

}

// decodeFrom decodes one JSON object from body into dst, rejecting unknown fields and trailing data.
// On failure it has already written the problem response and returns false.
func decodeFrom(w http.ResponseWriter, r *http.Request, body io.Reader, dst any) bool {
	panic("TODO: see chapter 38")

}

// readBody reads the whole (limited) body; on failure it has written the problem and returns false.
func readBody(w http.ResponseWriter, r *http.Request) ([]byte, bool) {
	panic("TODO: see chapter 38")

}

// pageParams reads ?limit= and ?cursor= for a list route. On failure it has written the problem.
func pageParams(w http.ResponseWriter, r *http.Request) (limit int, after *page.Cursor, ok bool) {
	panic("TODO: see chapter 38")

}

// writePage sends one page of items; key gives the cursor of an item.
func writePage[T any](w http.ResponseWriter, items []T, more bool, key func(T) page.Cursor) {
	panic("TODO: see chapter 38")

}
