package api

import (
	"errors"
	"net/http"

	"example.com/hookline/internal/domain"
	"example.com/hookline/internal/page"
	"example.com/hookline/internal/problem"
)

// eventDeliveries handles GET /v1/events/{id}/deliveries: every delivery of the event with its full
// attempt history (spec F8). The list is bounded by the tenant's endpoints, so it is not paged.
func (s *server) eventDeliveries(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !domain.IsID(id) {
		notFound(w, r, "event", id)
		return
	}
	ds, err := s.store.EventDeliveries(r.Context(), tenantFrom(r.Context()).ID, id)
	switch {
	case errors.Is(err, domain.ErrNotFound):
		notFound(w, r, "event", id)
	case err != nil:
		s.internal(w, r, err)
	default:
		if ds == nil {
			ds = []domain.Delivery{}
		}
		writeJSON(w, http.StatusOK, page.Page[domain.Delivery]{Items: ds})
	}
}

// deadDeliveries handles GET /v1/deliveries?status=dead&limit=&cursor= (spec F7). Only status=dead
// exists in this version; the parameter is required so that adding others later changes nothing.
func (s *server) deadDeliveries(w http.ResponseWriter, r *http.Request) {
	if st := r.URL.Query().Get("status"); st != string(domain.Dead) {
		validation(w, r, []problem.FieldError{{Detail: `must be "dead"`, Pointer: "query:status"}})
		return
	}
	limit, after, ok := pageParams(w, r)
	if !ok {
		return
	}
	items, more, err := s.store.DeadDeliveries(r.Context(), tenantFrom(r.Context()).ID, after, limit)
	if err != nil {
		s.internal(w, r, err)
		return
	}
	writePage(w, items, more, func(d domain.Delivery) page.Cursor { return page.Cursor{CreatedAt: d.CreatedAt, ID: d.ID} })
}

// replayDelivery handles POST /v1/deliveries/{id}/replay: 202 and the delivery, now pending, or 409
// when it is not dead. Replaying twice is therefore harmless: the second call gets the 409.
func (s *server) replayDelivery(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !domain.IsID(id) {
		notFound(w, r, "delivery", id)
		return
	}
	d, err := s.store.ReplayDelivery(r.Context(), tenantFrom(r.Context()).ID, id)
	switch {
	case errors.Is(err, domain.ErrNotFound):
		notFound(w, r, "delivery", id)
	case errors.Is(err, domain.ErrNotDead):
		problem.Write(w, r, problem.NotDead.New("delivery "+id+" is pending, in flight or already delivered"))
	case err != nil:
		s.internal(w, r, err)
	default:
		writeJSON(w, http.StatusAccepted, d)
	}
}
