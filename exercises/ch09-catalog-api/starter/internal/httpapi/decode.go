package httpapi

import (
	"encoding/json"
	"net/http"
	"regexp"

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
	// TODO(day 21): harden this decoder (section 9.9): require Content-Type application/json (415),
	// call DisallowUnknownFields (an unknown field is a validation problem with its pointer), reject
	// data after the object, map *http.MaxBytesError to 413 and *json.UnmarshalTypeError to a
	// validation problem. Today it only rejects malformed JSON.
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		problem.Write(w, r, problem.MalformedJSON.New(err.Error()))
		return false
	}
	return true
}

// validate checks every field and reports all problems at once, so a client fixes them in one round.
func (c createItemRequest) validate() []problem.FieldError {
	// TODO(day 21): add the format and range rules: sku matches skuPattern, name is not blank and
	// at most maxNameRunes characters of valid UTF-8, price_cents is from 0 to maxPriceCents.
	// Report every problem at once, as today's required-field checks do.
	var errs []problem.FieldError
	if c.SKU == nil {
		errs = append(errs, problem.FieldError{Detail: "is required", Pointer: "#/sku"})
	}
	if c.Name == nil {
		errs = append(errs, problem.FieldError{Detail: "is required", Pointer: "#/name"})
	}
	if c.PriceCents == nil {
		errs = append(errs, problem.FieldError{Detail: "is required", Pointer: "#/price_cents"})
	}
	return errs
}
