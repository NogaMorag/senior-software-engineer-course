-- +goose Up
-- One delivery per (event, endpoint): the dispatcher (milestone M4) inserts them, the workers
-- (milestone M5) claim due rows with FOR UPDATE SKIP LOCKED and a lease.
CREATE TABLE deliveries (
    id              uuid PRIMARY KEY,
    tenant_id       uuid NOT NULL REFERENCES tenants (id) ON DELETE CASCADE,
    event_id        uuid NOT NULL REFERENCES events (id) ON DELETE CASCADE,
    endpoint_id     uuid NOT NULL REFERENCES endpoints (id) ON DELETE CASCADE,
    status          text NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending', 'in_flight', 'delivered', 'dead')),
    attempts        integer NOT NULL DEFAULT 0 CHECK (attempts >= 0),
    next_attempt_at timestamptz NOT NULL DEFAULT now(),
    lease_until     timestamptz,
    last_error      text,
    delivered_at    timestamptz,
    created_at      timestamptz NOT NULL DEFAULT now(),
    -- Reprocessing an event (relay or dispatcher crash) cannot create a second delivery.
    UNIQUE (event_id, endpoint_id)
);
-- Workers: the oldest due pending rows. Delivered and dead rows never enter this index.
CREATE INDEX deliveries_due_idx ON deliveries (next_attempt_at) WHERE status = 'pending';
-- Lease expiry: in-flight rows whose worker died.
CREATE INDEX deliveries_lease_idx ON deliveries (lease_until) WHERE status = 'in_flight';
-- GET /v1/deliveries?status=dead (milestone M5): one tenant's dead rows, newest first.
CREATE INDEX deliveries_dead_page_idx ON deliveries (tenant_id, created_at DESC, id DESC) WHERE status = 'dead';
-- GET /v1/events/{id}/deliveries.
CREATE INDEX deliveries_event_idx ON deliveries (event_id);

-- Append-only record of every HTTP attempt.
CREATE TABLE attempts (
    id          bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    delivery_id uuid NOT NULL REFERENCES deliveries (id) ON DELETE CASCADE,
    n           integer NOT NULL CHECK (n >= 1),
    started_at  timestamptz NOT NULL,
    duration_ms integer NOT NULL CHECK (duration_ms >= 0),
    status_code integer CHECK (status_code BETWEEN 100 AND 599),
    error       text,
    UNIQUE (delivery_id, n)
);

-- +goose Down
DROP TABLE attempts;
DROP TABLE deliveries;
