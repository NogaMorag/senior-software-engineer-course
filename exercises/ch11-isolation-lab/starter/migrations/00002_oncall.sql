-- +goose Up
-- The write-skew example: a shift must always have at least one doctor on call.
CREATE TABLE oncall (
    shift_id int NOT NULL,
    doctor   text NOT NULL,
    on_call  boolean NOT NULL DEFAULT true,
    PRIMARY KEY (shift_id, doctor)
);

-- +goose Down
DROP TABLE oncall;
