-- +goose Up
-- A tenant is a customer of Hookline. Only the SHA-256 of its API key is stored; the key itself is
-- printed once by hookline-admin and never seen by the server again.
CREATE TABLE tenants (
    id           uuid PRIMARY KEY,
    name         text NOT NULL CHECK (length(name) BETWEEN 1 AND 100),
    api_key_hash bytea NOT NULL UNIQUE CHECK (length(api_key_hash) = 32),
    rate_per_sec integer NOT NULL DEFAULT 50 CHECK (rate_per_sec > 0),
    burst        integer NOT NULL DEFAULT 100 CHECK (burst > 0),
    created_at   timestamptz NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE tenants;
