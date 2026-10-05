// Package webhook is drill 4: an HTTP handler that receives webhooks from a sender that
// delivers "at least once", and makes processing safe to retry. Deliveries with the same
// event id within a window are processed once, including deliveries that arrive at the
// same moment.
// Time box: 30 minutes. The tests in webhook_test.go are the specification.
package webhook

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// EventIDHeader carries the sender's id for the event: the same on every retry.
const EventIDHeader = "Hookline-Event-Id"

// MaxBody is the largest request body the handler reads.
const MaxBody = 1 << 20

// Event is what the processor receives.
type Event struct {
	ID      string
	Payload json.RawMessage
}

// Processor does the real work for one event. An error means "not done": the id is
// forgotten, the sender gets a 500, and its retry will be processed again.
type Processor func(ctx context.Context, ev Event) error

// Handler dedupes deliveries by event id. It is an http.Handler and must be safe for
// concurrent use.
type Handler struct {
	// TODO: your fields.
}

// New returns a handler that remembers a processed id for window. now is the clock.
func New(p Processor, window time.Duration, now func() time.Time) *Handler {
	panic("TODO: drill 4, see chapter 35, section 35.5")
}

// ServeHTTP answers 200 {"status":"processed"} for the delivery that did the work,
// 200 {"status":"duplicate"} for every other delivery of a processed id, and 500 when
// processing failed (for the delivery that tried and for duplicates that waited on it).
// Bad requests: 405 for a method other than POST, 400 without the id header or with a body
// that is not JSON, 413 for a body above MaxBody.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	panic("TODO: drill 4, see chapter 35, section 35.5")
}

// Remembered reports how many processed ids are inside the window right now.
func (h *Handler) Remembered() int {
	panic("TODO: drill 4, see chapter 35, section 35.5")
}
