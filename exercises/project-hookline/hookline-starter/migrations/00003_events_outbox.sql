-- +goose Up
CREATE TABLE events (
    id              uuid PRIMARY KEY,
    tenant_id       uuid NOT NULL REFERENCES tenants (id) ON DELETE CASCADE,
    type            text NOT NULL CHECK (length(type) BETWEEN 1 AND 128),
    payload         jsonb NOT NULL,
    idempotency_key text NOT NULL CHECK (length(idempotency_key) BETWEEN 1 AND 255),
    body_hash       bytea NOT NULL CHECK (length(body_hash) = 32),
    created_at      timestamptz NOT NULL DEFAULT now(),
    -- The constraint, not application code, decides that a retried request is a duplicate.
    UNIQUE (tenant_id, idempotency_key)
);

-- The transactional outbox (chapter 13): written in the same transaction as the event, read by the relay
-- (milestone M4), which publishes each row to Redpanda and then sets published_at.
CREATE TABLE outbox (
    id           bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    event_id     uuid NOT NULL REFERENCES events (id) ON DELETE CASCADE,
    created_at   timestamptz NOT NULL DEFAULT now(),
    published_at timestamptz
);
-- The relay only ever looks for unpublished rows, in id order; published rows leave the index.
CREATE INDEX outbox_unpublished_idx ON outbox (id) WHERE published_at IS NULL;

-- +goose Down
DROP TABLE outbox;
DROP TABLE events;
