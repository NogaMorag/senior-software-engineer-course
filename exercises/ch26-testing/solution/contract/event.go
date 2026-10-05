// Package contract is the producer side of the webhook contract: the body Hookline POSTs to a
// receiver. webhook-event.schema.json is the contract itself. The Go tests check that every body
// this package produces conforms to it; the Python consumer in ../consumer checks that it accepts
// every body the schema allows. Neither side imports the other: they share only the schema file
// and the examples in testdata/examples.
package contract

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"regexp"
	"time"
)

// Schema is the JSON Schema (draft 2020-12) of a webhook body.
//
//go:embed webhook-event.schema.json
var Schema []byte

// SchemaURL is the schema's $id; validators use it as the schema's name.
const SchemaURL = "https://example.com/hookline/webhook-event.schema.json"

// typePattern is the same rule as the schema's "type" pattern: dotted lowercase words,
// at least two of them, like "invoice.paid".
var typePattern = regexp.MustCompile(`^[a-z][a-z0-9_]*(\.[a-z][a-z0-9_]*)+$`)

const maxTypeLen = 100

// Event is one webhook body.
type Event struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"`
	CreatedAt time.Time       `json:"created_at"`
	Data      json.RawMessage `json:"data"`
}

// New builds an event, refusing what the schema would refuse: a malformed type, or data that
// is not a JSON object. It catches the mistake here, at the producer, instead of at a receiver.
func New(id, eventType string, createdAt time.Time, data any) (Event, error) {
	if len(eventType) > maxTypeLen || !typePattern.MatchString(eventType) {
		return Event{}, fmt.Errorf("event type %q: want dotted lowercase words, like invoice.paid", eventType)
	}
	raw, err := json.Marshal(data)
	if err != nil {
		return Event{}, fmt.Errorf("event data: %w", err)
	}
	if len(raw) == 0 || raw[0] != '{' {
		return Event{}, fmt.Errorf("event data must be a JSON object, got %s", raw)
	}
	// RFC 3339 in UTC, to the microsecond: what PostgreSQL stores, and what the schema's
	// "date-time" format accepts.
	return Event{ID: id, Type: eventType, CreatedAt: createdAt.UTC().Truncate(time.Microsecond), Data: raw}, nil
}

// Body is the exact bytes that get signed and sent.
func (e Event) Body() ([]byte, error) {
	return json.Marshal(e)
}
