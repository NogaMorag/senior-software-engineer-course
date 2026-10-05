package api

import (
	"net/http"
)

type createSubscriptionRequest struct {
	EndpointID *string `json:"endpoint_id"`
	EventType  *string `json:"event_type"`
}

// createSubscription handles POST /v1/subscriptions. An endpoint of another tenant is "not found".
func (s *server) createSubscription(w http.ResponseWriter, r *http.Request) {
	panic("TODO: see chapter 38")

}

// getSubscription handles GET /v1/subscriptions/{id}.
func (s *server) getSubscription(w http.ResponseWriter, r *http.Request) {
	panic("TODO: see chapter 38")

}

// listSubscriptions handles GET /v1/subscriptions?endpoint_id=&limit=&cursor=.
func (s *server) listSubscriptions(w http.ResponseWriter, r *http.Request) {
	panic("TODO: see chapter 38")

}

// deleteSubscription handles DELETE /v1/subscriptions/{id}: 204, or 404 when it is not there (any more).
func (s *server) deleteSubscription(w http.ResponseWriter, r *http.Request) {
	panic("TODO: see chapter 38")

}

func deref(p *string) string {
	panic("TODO: see chapter 38")

}
