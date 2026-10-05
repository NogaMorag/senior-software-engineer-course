package api

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"

	"example.com/hookline/internal/domain"
	"example.com/hookline/internal/page"
	"example.com/hookline/internal/problem"
)

// limitBody caps the request body before anything reads it (chapter 9). Reading past the cap fails with
// *http.MaxBytesError, which decodeJSON and readBody turn into 413.
func limitBody(maxBytes int64, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ContentLength > maxBytes {
			problem.Write(w, r, problem.TooLarge.New("the request body exceeds the limit of this endpoint"))
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
		next.ServeHTTP(w, r)
	})
}

// isJSON checks the Content-Type; on failure it has written the 415 and returns false.
func isJSON(w http.ResponseWriter, r *http.Request) bool {
	mt, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mt != "application/json" {
		problem.Write(w, r, problem.UnsupportedMedia.New("send the body as application/json"))
		return false
	}
	return true
}

// decodeJSON reads exactly one JSON object from the request body into dst.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	return isJSON(w, r) && decodeFrom(w, r, r.Body, dst)
}

// decodeFrom decodes one JSON object from body into dst, rejecting unknown fields and trailing data.
// On failure it has already written the problem response and returns false.
func decodeFrom(w http.ResponseWriter, r *http.Request, body io.Reader, dst any) bool {
	dec := json.NewDecoder(body)
	dec.DisallowUnknownFields()
	err := dec.Decode(dst)
	if err == nil && dec.Decode(&struct{}{}) != io.EOF {
		err = errors.New("unexpected data after the JSON object")
	}
	if err == nil {
		return true
	}
	if _, tooLarge := errors.AsType[*http.MaxBytesError](err); tooLarge {
		problem.Write(w, r, problem.TooLarge.New("the request body exceeds the limit of this endpoint"))
		return false
	}
	if field, ok := strings.CutPrefix(err.Error(), "json: unknown field "); ok {
		validation(w, r, []problem.FieldError{{Detail: "unknown field", Pointer: "#/" + strings.Trim(field, `"`)}})
		return false
	}
	if te, ok := errors.AsType[*json.UnmarshalTypeError](err); ok {
		validation(w, r, []problem.FieldError{{Detail: "must be " + te.Type.String(), Pointer: "#/" + strings.ReplaceAll(te.Field, ".", "/")}})
		return false
	}
	problem.Write(w, r, problem.MalformedJSON.New(err.Error()))
	return false
}

// readBody reads the whole (limited) body; on failure it has written the problem and returns false.
func readBody(w http.ResponseWriter, r *http.Request) ([]byte, bool) {
	b, err := io.ReadAll(r.Body)
	if err == nil {
		return b, true
	}
	if _, tooLarge := errors.AsType[*http.MaxBytesError](err); tooLarge {
		problem.Write(w, r, problem.TooLarge.New("the request body exceeds the limit of this endpoint"))
	} else {
		problem.Write(w, r, problem.MalformedJSON.New("the request body could not be read"))
	}
	return nil, false
}

// pageParams reads ?limit= and ?cursor= for a list route. On failure it has written the problem.
func pageParams(w http.ResponseWriter, r *http.Request) (limit int, after *page.Cursor, ok bool) {
	q := r.URL.Query()
	limit, err := page.ParseLimit(q.Get("limit"))
	if err != nil {
		validation(w, r, []problem.FieldError{{Detail: err.Error(), Pointer: "query:limit"}})
		return 0, nil, false
	}
	if raw := q.Get("cursor"); raw != "" {
		c, err := page.Decode(raw)
		if err != nil || !domain.IsID(c.ID) {
			problem.Write(w, r, problem.InvalidCursor.New("pass back the next_cursor value of a previous page unchanged"))
			return 0, nil, false
		}
		after = &c
	}
	return limit, after, true
}

// writePage sends one page of items; key gives the cursor of an item.
func writePage[T any](w http.ResponseWriter, items []T, more bool, key func(T) page.Cursor) {
	if items == nil {
		items = []T{} // an empty page is [], never null
	}
	out := page.Page[T]{Items: items}
	if more {
		out.NextCursor = page.Encode(key(items[len(items)-1]))
	}
	writeJSON(w, http.StatusOK, out)
}
