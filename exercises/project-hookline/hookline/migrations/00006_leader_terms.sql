-- +goose Up
-- Leader election for the relay (milestone M6, chapter 18): every time a relay takes the advisory lock it
-- increments its role's term here. The term is the fencing token the relay's mark carries, so a relay
-- that lost leadership without noticing cannot mark rows the new leader owns.
CREATE TABLE leader_terms (
    lock_key bigint PRIMARY KEY,
    term     bigint NOT NULL
);

-- +goose Down
DROP TABLE leader_terms;
