package api

import (
	"errors"
	"net/http"

	"example.com/hookline/internal/domain"
	"example.com/hookline/internal/page"
	"example.com/hookline/internal/problem"
)

type createEndpointRequest struct {
	URL *string `json:"url"`
}

// createdEndpoint is the one response that carries the signing secret.
type createdEndpoint struct {
	domain.Endpoint
	Secret string `json:"secret"`
}

// createEndpoint handles POST /v1/endpoints: 201 with the endpoint and its new signing secret.
func (s *server) createEndpoint(w http.ResponseWriter, r *http.Request) {
	var req createEndpointRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	url := ""
	if req.URL != nil {
		url = *req.URL
	}
	if reason := domain.ValidateEndpointURL(url, s.devMode); reason != "" {
		validation(w, r, []problem.FieldError{{Detail: reason, Pointer: "#/url"}})
		return
	}
	now := domain.Now(s.now)
	e := domain.Endpoint{
		ID: domain.NewID(now), TenantID: tenantFrom(r.Context()).ID, URL: url,
		Secret: domain.NewSigningSecret(), Enabled: true, CreatedAt: now,
	}
	if err := s.store.CreateEndpoint(r.Context(), e); err != nil {
		s.internal(w, r, err)
		return
	}
	w.Header().Set("Location", "/v1/endpoints/"+e.ID)
	writeJSON(w, http.StatusCreated, createdEndpoint{Endpoint: e, Secret: e.Secret})
}

// getEndpoint handles GET /v1/endpoints/{id}.
func (s *server) getEndpoint(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !domain.IsID(id) {
		notFound(w, r, "endpoint", id)
		return
	}
	e, err := s.store.GetEndpoint(r.Context(), tenantFrom(r.Context()).ID, id)
	s.respond(w, r, "endpoint", id, e, err)
}

// listEndpoints handles GET /v1/endpoints?limit=&cursor=.
func (s *server) listEndpoints(w http.ResponseWriter, r *http.Request) {
	limit, after, ok := pageParams(w, r)
	if !ok {
		return
	}
	items, more, err := s.store.ListEndpoints(r.Context(), tenantFrom(r.Context()).ID, after, limit)
	if err != nil {
		s.internal(w, r, err)
		return
	}
	writePage(w, items, more, func(e domain.Endpoint) page.Cursor { return page.Cursor{CreatedAt: e.CreatedAt, ID: e.ID} })
}

// disableEndpoint handles POST /v1/endpoints/{id}/disable: no new deliveries go to it (from M4 on).
func (s *server) disableEndpoint(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !domain.IsID(id) {
		notFound(w, r, "endpoint", id)
		return
	}
	e, err := s.store.DisableEndpoint(r.Context(), tenantFrom(r.Context()).ID, id)
	if err == nil {
		s.changed(r, e.TenantID)
	}
	s.respond(w, r, "endpoint", id, e, err)
}

// respond sends v with 200, or the problem for err: domain.ErrNotFound is 404, anything else 500.
func (s *server) respond(w http.ResponseWriter, r *http.Request, kind, id string, v any, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		notFound(w, r, kind, id)
	case err != nil:
		s.internal(w, r, err)
	default:
		writeJSON(w, http.StatusOK, v)
	}
}
