-- +goose Up
CREATE TABLE accounts (
    id      bigint PRIMARY KEY,
    owner   text NOT NULL,
    balance bigint NOT NULL
);

-- +goose Down
DROP TABLE accounts;
