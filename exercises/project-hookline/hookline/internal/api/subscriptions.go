package api

import (
	"errors"
	"net/http"

	"example.com/hookline/internal/domain"
	"example.com/hookline/internal/page"
	"example.com/hookline/internal/problem"
)

type createSubscriptionRequest struct {
	EndpointID *string `json:"endpoint_id"`
	EventType  *string `json:"event_type"`
}

// createSubscription handles POST /v1/subscriptions. An endpoint of another tenant is "not found".
func (s *server) createSubscription(w http.ResponseWriter, r *http.Request) {
	var req createSubscriptionRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	var errs []problem.FieldError
	endpointID, eventType := deref(req.EndpointID), deref(req.EventType)
	if !domain.IsID(endpointID) {
		errs = append(errs, problem.FieldError{Detail: "must be the id of one of your endpoints", Pointer: "#/endpoint_id"})
	}
	if reason := domain.ValidateEventType(eventType, true); reason != "" {
		errs = append(errs, problem.FieldError{Detail: reason + ` (or "*" for every type)`, Pointer: "#/event_type"})
	}
	if len(errs) > 0 {
		validation(w, r, errs)
		return
	}
	now := domain.Now(s.now)
	sub := domain.Subscription{ID: domain.NewID(now), TenantID: tenantFrom(r.Context()).ID, EndpointID: endpointID, EventType: eventType, CreatedAt: now}
	switch err := s.store.CreateSubscription(r.Context(), sub); {
	case errors.Is(err, domain.ErrNotFound):
		notFound(w, r, "endpoint", endpointID)
	case errors.Is(err, domain.ErrConflict):
		problem.Write(w, r, problem.Conflict.New("endpoint "+endpointID+" is already subscribed to "+eventType))
	case err != nil:
		s.internal(w, r, err)
	default:
		w.Header().Set("Location", "/v1/subscriptions/"+sub.ID)
		writeJSON(w, http.StatusCreated, sub)
	}
}

// getSubscription handles GET /v1/subscriptions/{id}.
func (s *server) getSubscription(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !domain.IsID(id) {
		notFound(w, r, "subscription", id)
		return
	}
	sub, err := s.store.GetSubscription(r.Context(), tenantFrom(r.Context()).ID, id)
	s.respond(w, r, "subscription", id, sub, err)
}

// listSubscriptions handles GET /v1/subscriptions?endpoint_id=&limit=&cursor=.
func (s *server) listSubscriptions(w http.ResponseWriter, r *http.Request) {
	limit, after, ok := pageParams(w, r)
	if !ok {
		return
	}
	endpointID := r.URL.Query().Get("endpoint_id")
	if endpointID != "" && !domain.IsID(endpointID) {
		validation(w, r, []problem.FieldError{{Detail: "must be an endpoint id", Pointer: "query:endpoint_id"}})
		return
	}
	items, more, err := s.store.ListSubscriptions(r.Context(), tenantFrom(r.Context()).ID, endpointID, after, limit)
	if err != nil {
		s.internal(w, r, err)
		return
	}
	writePage(w, items, more, func(s domain.Subscription) page.Cursor { return page.Cursor{CreatedAt: s.CreatedAt, ID: s.ID} })
}

// deleteSubscription handles DELETE /v1/subscriptions/{id}: 204, or 404 when it is not there (any more).
func (s *server) deleteSubscription(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !domain.IsID(id) {
		notFound(w, r, "subscription", id)
		return
	}
	switch err := s.store.DeleteSubscription(r.Context(), tenantFrom(r.Context()).ID, id); {
	case errors.Is(err, domain.ErrNotFound):
		notFound(w, r, "subscription", id)
	case err != nil:
		s.internal(w, r, err)
	default:
		w.WriteHeader(http.StatusNoContent)
	}
}

func deref(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
