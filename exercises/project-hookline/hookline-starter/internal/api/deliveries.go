package api

import (
	"net/http"
)

// eventDeliveries handles GET /v1/events/{id}/deliveries: every delivery of the event with its full
// attempt history (spec F8). The list is bounded by the tenant's endpoints, so it is not paged.
func (s *server) eventDeliveries(w http.ResponseWriter, r *http.Request) {
	panic("TODO: see chapter 38")

}

// deadDeliveries handles GET /v1/deliveries?status=dead&limit=&cursor= (spec F7). Only status=dead
// exists in this version; the parameter is required so that adding others later changes nothing.
func (s *server) deadDeliveries(w http.ResponseWriter, r *http.Request) {
	panic("TODO: see chapter 38")

}

// replayDelivery handles POST /v1/deliveries/{id}/replay: 202 and the delivery, now pending, or 409
// when it is not dead. Replaying twice is therefore harmless: the second call gets the 409.
func (s *server) replayDelivery(w http.ResponseWriter, r *http.Request) {
	panic("TODO: see chapter 38")

}
