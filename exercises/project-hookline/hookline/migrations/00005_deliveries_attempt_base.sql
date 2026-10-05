-- +goose Up
-- A replay (milestone M5) must not reset deliveries.attempts: attempts.n is unique per delivery, and
-- receivers should see Hookline-Attempt keep rising. Instead a replay stores the attempts so far here,
-- and a worker compares attempts - attempt_base with the maximum. This is the column chapter 29,
-- section 29.5 proposes; with the default 0 every existing row behaves as before.
ALTER TABLE deliveries
    ADD COLUMN attempt_base integer NOT NULL DEFAULT 0 CHECK (attempt_base >= 0),
    ADD CONSTRAINT deliveries_attempt_base_le_attempts CHECK (attempt_base <= attempts);

-- +goose Down
ALTER TABLE deliveries DROP COLUMN attempt_base;
