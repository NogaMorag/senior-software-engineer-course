-- +goose Up
CREATE TABLE endpoints (
    id         uuid PRIMARY KEY,
    tenant_id  uuid NOT NULL REFERENCES tenants (id) ON DELETE CASCADE,
    url        text NOT NULL CHECK (length(url) BETWEEN 1 AND 2048),
    secret     text NOT NULL, -- signing secret; encrypting it at rest is an extension, not milestone scope
    enabled    boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    -- Redundant with the primary key, but it lets subscriptions reference (tenant_id, id) and so makes
    -- "a subscription's endpoint belongs to the same tenant" a rule the database enforces.
    UNIQUE (tenant_id, id)
);
-- GET /v1/endpoints: one tenant's rows, newest first, keyset pagination on (created_at, id).
CREATE INDEX endpoints_tenant_page_idx ON endpoints (tenant_id, created_at DESC, id DESC);

CREATE TABLE subscriptions (
    id          uuid PRIMARY KEY,
    tenant_id   uuid NOT NULL,
    endpoint_id uuid NOT NULL,
    event_type  text NOT NULL CHECK (length(event_type) BETWEEN 1 AND 128),
    created_at  timestamptz NOT NULL DEFAULT now(),
    FOREIGN KEY (tenant_id, endpoint_id) REFERENCES endpoints (tenant_id, id) ON DELETE CASCADE,
    -- One subscription per endpoint and type. tenant_id leads the key on purpose: PostgreSQL checks the
    -- unique index before the foreign key, so with (endpoint_id, event_type) alone another tenant probing
    -- an existing pair would get "already exists" (409) instead of "not found" and learn that it exists.
    UNIQUE (tenant_id, endpoint_id, event_type)
);
CREATE INDEX subscriptions_tenant_page_idx ON subscriptions (tenant_id, created_at DESC, id DESC);
-- The dispatcher (milestone M4) asks: which subscriptions of this tenant match this type or '*'?
CREATE INDEX subscriptions_match_idx ON subscriptions (tenant_id, event_type);

-- +goose Down
DROP TABLE subscriptions;
DROP TABLE endpoints;
