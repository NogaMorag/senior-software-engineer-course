-- Chapter 13 schema. The business table, the outbox, and the consumer's tables live in
-- one database here for convenience; in a real system the consumer owns its own database.

-- The business row.
CREATE TABLE orders (
    id          bigserial PRIMARY KEY,
    customer    text        NOT NULL,
    sku         text        NOT NULL,
    quantity    int         NOT NULL CHECK (quantity > 0),
    created_at  timestamptz NOT NULL DEFAULT now()
);

-- The outbox: one row per message that must reach the log. It is written in the same
-- transaction as the business row, so either both exist or neither does.
CREATE TABLE outbox (
    id            bigserial PRIMARY KEY,           -- publish order
    message_id    uuid        NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    topic         text        NOT NULL,
    key           text        NOT NULL,           -- the Kafka record key (ordering scope)
    payload       jsonb       NOT NULL,
    created_at    timestamptz NOT NULL DEFAULT now(),
    published_at  timestamptz                     -- NULL until the relay marks it
);

-- The relay only ever looks for unpublished rows, so index exactly those.
CREATE INDEX outbox_unpublished ON outbox (id) WHERE published_at IS NULL;

-- The consumer's dedupe table: one row per (consumer, message) it has applied.
CREATE TABLE processed_messages (
    consumer      text        NOT NULL,
    message_id    uuid        NOT NULL,
    processed_at  timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (consumer, message_id)
);

-- The consumer's effect: one shipment per order. No UNIQUE (order_id) on purpose, so
-- that a missing dedupe would show up as two shipments.
CREATE TABLE shipments (
    id        bigserial PRIMARY KEY,
    order_id  bigint      NOT NULL REFERENCES orders (id),
    created_at timestamptz NOT NULL DEFAULT now()
);
