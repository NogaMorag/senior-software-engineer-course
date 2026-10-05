package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"regexp"
	"strings"
	"unicode/utf8"

	"example.com/ch09-catalog-api/internal/problem"
)

// Field limits. They are part of the contract: api/openapi.yaml states the same numbers.
const (
	maxNameRunes  = 200
	maxPriceCents = 100_000_000 // one million in the currency's main unit
)

var skuPattern = regexp.MustCompile(`^[A-Z0-9-]{3,32}$`)

// createItemRequest is the body of POST /v1/items. Pointers tell "missing" apart from "zero".
type createItemRequest struct {
	SKU        *string `json:"sku"`
	Name       *string `json:"name"`
	PriceCents *int64  `json:"price_cents"`
}

// decodeJSON reads exactly one JSON object into dst. It rejects a wrong Content-Type, unknown fields,
// trailing data and bodies over the limit set by limitBody. On failure it has already written the
// problem response and returns false.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	mt, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mt != "application/json" {
		problem.Write(w, r, problem.UnsupportedMedia.New("send the body as application/json"))
		return false
	}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err = dec.Decode(dst)
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
		p := problem.Validation.New("the body has a field this API does not define")
		p.Errors = []problem.FieldError{{Detail: "unknown field", Pointer: "#/" + strings.Trim(field, `"`)}}
		problem.Write(w, r, p)
		return false
	}
	if te, ok := errors.AsType[*json.UnmarshalTypeError](err); ok {
		p := problem.Validation.New("a field has the wrong JSON type")
		p.Errors = []problem.FieldError{{Detail: "must be " + te.Type.String(), Pointer: "#/" + strings.ReplaceAll(te.Field, ".", "/")}}
		problem.Write(w, r, p)
		return false
	}
	problem.Write(w, r, problem.MalformedJSON.New(err.Error()))
	return false
}

// validate checks every field and reports all problems at once, so a client fixes them in one round.
func (c createItemRequest) validate() []problem.FieldError {
	var errs []problem.FieldError
	add := func(ptr, format string, args ...any) {
		errs = append(errs, problem.FieldError{Detail: fmt.Sprintf(format, args...), Pointer: ptr})
	}
	switch {
	case c.SKU == nil:
		add("#/sku", "is required")
	case !skuPattern.MatchString(*c.SKU):
		add("#/sku", "must be 3 to 32 characters of A-Z, 0-9 and '-'")
	}
	switch {
	case c.Name == nil:
		add("#/name", "is required")
	case strings.TrimSpace(*c.Name) == "":
		add("#/name", "must not be blank")
	case !utf8.ValidString(*c.Name) || utf8.RuneCountInString(*c.Name) > maxNameRunes:
		add("#/name", "must be valid UTF-8 of at most %d characters", maxNameRunes)
	}
	switch {
	case c.PriceCents == nil:
		add("#/price_cents", "is required")
	case *c.PriceCents < 0 || *c.PriceCents > maxPriceCents:
		add("#/price_cents", "must be from 0 to %d", maxPriceCents)
	}
	return errs
}
