// Package problem writes errors as RFC 9457 problem details (application/problem+json).
// It is the single error contract of the API: every handler and middleware reports failure through it.
package problem

import (
	"encoding/json"
	"net/http"
)

// ContentType is the media type RFC 9457 registers for JSON problem details.
const ContentType = "application/problem+json"

// TypeBase prefixes every problem type URI of this API. Each type is documented in api/openapi.yaml.
const TypeBase = "https://hookline.example.com/problems/"

// FieldError locates one invalid input: Pointer is a JSON Pointer into the request body ("#/name"),
// or a query parameter written as "query:limit".
type FieldError struct {
	Detail  string `json:"detail"`
	Pointer string `json:"pointer"`
}

// Problem is one problem details object. Errors is an extension member used by the validation type.
type Problem struct {
	Type     string       `json:"type"`
	Title    string       `json:"title"`
	Status   int          `json:"status"`
	Detail   string       `json:"detail,omitempty"`
	Instance string       `json:"instance,omitempty"`
	Errors   []FieldError `json:"errors,omitempty"`
}

// The problem types of this API. The title is fixed per type; the detail explains this occurrence.
var (
	Validation         = Kind{http.StatusBadRequest, "validation", "Request is not valid"}
	MalformedJSON      = Kind{http.StatusBadRequest, "malformed-json", "Request body is not valid JSON"}
	InvalidCursor      = Kind{http.StatusBadRequest, "invalid-cursor", "Cursor is not valid"}
	IdempotencyMissing = Kind{http.StatusBadRequest, "idempotency-key-missing", "Idempotency-Key header is required"}
	Unauthorized       = Kind{http.StatusUnauthorized, "unauthorized", "API key is missing or not valid"}
	NotFound           = Kind{http.StatusNotFound, "not-found", "Resource not found"}
	MethodNotAllowed   = Kind{http.StatusMethodNotAllowed, "method-not-allowed", "Method not allowed"}
	Conflict           = Kind{http.StatusConflict, "conflict", "Resource already exists"}
	TooLarge           = Kind{http.StatusRequestEntityTooLarge, "body-too-large", "Request body is too large"}
	UnsupportedMedia   = Kind{http.StatusUnsupportedMediaType, "unsupported-media-type", "Content-Type must be application/json"}
	KeyReused          = Kind{http.StatusUnprocessableEntity, "idempotency-key-reused", "Idempotency-Key was used with a different request"}
	Internal           = Kind{http.StatusInternalServerError, "internal", "Internal error"}
)

// Kind is one problem type: its status code, URI slug and fixed title.
type Kind struct {
	status int
	slug   string
	title  string
}

// New builds a problem of kind k with an occurrence-specific detail.
func (k Kind) New(detail string) *Problem {
	return &Problem{Type: TypeBase + k.slug, Title: k.title, Status: k.status, Detail: detail}
}

// Write sends p as the response. Instance defaults to the request path, so a log line and the body agree.
func Write(w http.ResponseWriter, r *http.Request, p *Problem) {
	if p.Instance == "" && r != nil {
		p.Instance = r.URL.Path
	}
	w.Header().Set("Content-Type", ContentType)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(p.Status)
	_ = json.NewEncoder(w).Encode(p) // the status line is already sent; nothing useful to do on error
}
