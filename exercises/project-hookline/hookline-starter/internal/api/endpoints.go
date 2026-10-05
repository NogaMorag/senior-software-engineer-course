package api

import (
	"net/http"

	"example.com/hookline/internal/domain"
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
	panic("TODO: see chapter 38")

}

// getEndpoint handles GET /v1/endpoints/{id}.
func (s *server) getEndpoint(w http.ResponseWriter, r *http.Request) {
	panic("TODO: see chapter 38")

}

// listEndpoints handles GET /v1/endpoints?limit=&cursor=.
func (s *server) listEndpoints(w http.ResponseWriter, r *http.Request) {
	panic("TODO: see chapter 38")

}

// disableEndpoint handles POST /v1/endpoints/{id}/disable: no new deliveries go to it (from M4 on).
func (s *server) disableEndpoint(w http.ResponseWriter, r *http.Request) {
	panic("TODO: see chapter 38")

}

// respond sends v with 200, or the problem for err: domain.ErrNotFound is 404, anything else 500.
func (s *server) respond(w http.ResponseWriter, r *http.Request, kind, id string, v any, err error) {
	panic("TODO: see chapter 38")

}
