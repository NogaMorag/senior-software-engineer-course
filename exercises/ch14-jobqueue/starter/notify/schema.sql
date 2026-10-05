-- What the notify handler has sent, keyed by the idempotency key it gave the provider.
CREATE TABLE notifications (
    idempotency_key  text PRIMARY KEY,
    user_name        text        NOT NULL,
    message          text        NOT NULL,
    sent_at          timestamptz NOT NULL DEFAULT now()
);
