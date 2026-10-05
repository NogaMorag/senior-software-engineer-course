package domain

import "errors"

// Errors every store returns; the API maps each to one problem type.
var (
	// ErrNotFound: no such resource for this tenant. A resource of another tenant is also "not found",
	// so an id never reveals that it exists (chapter 28, broken object level authorization).
	ErrNotFound = errors.New("not found")
	// ErrConflict: the resource already exists (a subscription for the same endpoint and event type).
	ErrConflict = errors.New("already exists")
	// ErrIdempotencyMismatch: the idempotency key was already used by this tenant with a different body.
	ErrIdempotencyMismatch = errors.New("idempotency key reused with a different body")
)
