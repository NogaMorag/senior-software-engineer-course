package api

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"

	"example.com/hookline/internal/domain"
	"example.com/hookline/internal/problem"
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
	key, ok := parseKey(r.Header.Get(IdempotencyHeader))
	if !ok {
		problem.Write(w, r, problem.IdempotencyMissing.New("send a unique Idempotency-Key header (1 to 255 visible ASCII characters) with every event"))
		return
	}
	if !isJSON(w, r) {
		return
	}
	body, ok := readBody(w, r)
	if !ok {
		return
	}
	var req ingestRequest
	if !decodeFrom(w, r, bytes.NewReader(body), &req) {
		return
	}
	var errs []problem.FieldError
	typ := deref(req.Type)
	if reason := domain.ValidateEventType(typ, false); reason != "" {
		errs = append(errs, problem.FieldError{Detail: reason, Pointer: "#/type"})
	}
	if len(req.Payload) == 0 || req.Payload[0] != '{' { // the decoder already trimmed whitespace
		errs = append(errs, problem.FieldError{Detail: "must be a JSON object", Pointer: "#/payload"})
	}
	if len(errs) > 0 {
		validation(w, r, errs)
		return
	}
	now := domain.Now(s.now)
	e := domain.Event{
		ID: domain.NewID(now), TenantID: tenantFrom(r.Context()).ID, Type: typ, Payload: req.Payload,
		IdempotencyKey: key, BodyHash: sha256.Sum256(body), CreatedAt: now,
		TraceContext: traceparent(r.Context()),
	}
	stored, replayed, err := s.store.IngestEvent(r.Context(), e)
	switch {
	case errors.Is(err, domain.ErrIdempotencyMismatch):
		problem.Write(w, r, problem.KeyReused.New("this Idempotency-Key was already used with a different body; use a new key for a new event"))
		return
	case err != nil:
		s.internal(w, r, err)
		return
	}
	s.metrics.Ingested(r.Context(), replayed)
	if replayed {
		w.Header().Set(ReplayedHeader, "true")
	}
	w.Header().Set("Location", "/v1/events/"+stored.ID)
	writeJSON(w, http.StatusAccepted, stored)
}

// getEvent handles GET /v1/events/{id}.
func (s *server) getEvent(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !domain.IsID(id) {
		notFound(w, r, "event", id)
		return
	}
	e, err := s.store.GetEvent(r.Context(), tenantFrom(r.Context()).ID, id)
	s.respond(w, r, "event", id, e, err)
}

// traceparent is the W3C trace context of ctx as a header value, "" when ctx has no sampled span. It is
// stored with the event, so the processes that handle the event later continue this trace.
func traceparent(ctx context.Context) string {
	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	return carrier.Get("traceparent")
}

// parseKey accepts a bare token or the quoted Structured Fields string form the IETF draft specifies.
func parseKey(v string) (string, bool) {
	if len(v) >= 2 && v[0] == '"' && v[len(v)-1] == '"' {
		v = v[1 : len(v)-1]
	}
	if v == "" || len(v) > maxKeyLen {
		return "", false
	}
	for i := 0; i < len(v); i++ {
		if v[i] < 0x21 || v[i] > 0x7e || v[i] == '"' {
			return "", false
		}
	}
	return v, true
}
