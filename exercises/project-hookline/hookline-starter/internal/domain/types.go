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
	// TraceContext is the W3C traceparent of the ingest request (milestone M8), "" without telemetry.
	TraceContext string `json:"-"`
}

// DeliveryStatus is where a delivery is in its life (chapter 37, section 37.4).
type DeliveryStatus string

// The four statuses. pending: waiting for next_attempt_at. in_flight: claimed by a worker under a lease.
// delivered: a 2xx arrived. dead: out of attempts, or failed permanently; waits for a replay.
const (
	Pending   DeliveryStatus = "pending"
	InFlight  DeliveryStatus = "in_flight"
	Delivered DeliveryStatus = "delivered"
	Dead      DeliveryStatus = "dead"
)

// Delivery is one event on its way to one endpoint. Attempts counts every attempt ever made and never
// goes down: it numbers the history and the Hookline-Attempt header. A replay sets AttemptBase to
// Attempts, so the round in progress has Attempts-AttemptBase attempts (chapter 29, section 29.5).
type Delivery struct {
	ID            string         `json:"id"`
	TenantID      string         `json:"-"`
	EventID       string         `json:"event_id"`
	EndpointID    string         `json:"endpoint_id"`
	Status        DeliveryStatus `json:"status"`
	Attempts      int            `json:"attempts"`
	AttemptBase   int            `json:"attempt_base"`
	NextAttemptAt time.Time      `json:"next_attempt_at"`
	LastError     string         `json:"last_error,omitempty"`
	DeliveredAt   *time.Time     `json:"delivered_at,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	History       []Attempt      `json:"history,omitempty"`
}

// Attempt is one recorded HTTP attempt of a delivery. N is the delivery's Attempts value at the time,
// so it keeps rising across replays; StatusCode is 0 when no response arrived (timeout, connection error), and then Error says why.
type Attempt struct {
	N          int       `json:"n"`
	StartedAt  time.Time `json:"started_at"`
	DurationMS int       `json:"duration_ms"`
	StatusCode int       `json:"status_code,omitempty"`
	Error      string    `json:"error,omitempty"`
}
