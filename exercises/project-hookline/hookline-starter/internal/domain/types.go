// Package domain holds Hookline's resource types, the errors stores return, and the small pure functions
// every layer shares: identifiers, API keys, signing secrets and input validation. It imports nothing
// from the rest of the module, so every other package may import it.
package domain

import (
	"encoding/json"
	"time"
)

// Tenant is a customer of Hookline: the SaaS product that sends events. It authenticates with an API key;
// only the SHA-256 of the key is stored.
type Tenant struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	APIKeyHash APIKeyHash `json:"-"`
	RatePerSec int        `json:"rate_per_sec"` // ingest rate limit, used from milestone M6
	Burst      int        `json:"burst"`
	CreatedAt  time.Time  `json:"created_at"`
}

// Endpoint is a receiver URL registered by a tenant. Secret signs every delivery (milestone M5); it is
// returned once, in the response that creates the endpoint, and never again.
type Endpoint struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"-"`
	URL       string    `json:"url"`
	Secret    string    `json:"-"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
}

// AllTypes is the event type of a subscription that matches every event.
const AllTypes = "*"

// Subscription says that an endpoint wants events of one type, or of every type (AllTypes).
type Subscription struct {
	ID         string    `json:"id"`
	TenantID   string    `json:"-"`
	EndpointID string    `json:"endpoint_id"`
	EventType  string    `json:"event_type"`
	CreatedAt  time.Time `json:"created_at"`
}

// Event is one thing that happened in a tenant's product. IdempotencyKey and BodyHash identify the
// ingest request that created it: the same key with the same body returns this event again.
type Event struct {
	ID             string          `json:"id"`
	TenantID       string          `json:"-"`
	Type           string          `json:"type"`
	Payload        json.RawMessage `json:"payload"`
	IdempotencyKey string          `json:"-"`
	BodyHash       [32]byte        `json:"-"`
	CreatedAt      time.Time       `json:"created_at"`
}
