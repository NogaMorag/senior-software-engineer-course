package api

import (
	"encoding/json"

	"net/http"
)

// Header names of the idempotency contract (chapter 9, section 9.5).
const (
	IdempotencyHeader = "Idempotency-Key"
	ReplayedHeader    = "Idempotent-Replayed"
	maxKeyLen         = 255
)

type ingestRequest struct {
	Type    *string         `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// ingestEvent handles POST /v1/events. The store decides, atomically, whether the (tenant, key) pair is
// new: then the event is stored (and, in PostgreSQL, its outbox row in the same transaction); a retry
// with the same body gets the original event back; another body under the same key is 422.
func (s *server) ingestEvent(w http.ResponseWriter, r *http.Request) {
	panic("TODO: see chapter 38")

}

// getEvent handles GET /v1/events/{id}.
func (s *server) getEvent(w http.ResponseWriter, r *http.Request) {
	panic("TODO: see chapter 38")

}

// parseKey accepts a bare token or the quoted Structured Fields string form the IETF draft specifies.
func parseKey(v string) (string, bool) {
	panic("TODO: see chapter 38")

}
