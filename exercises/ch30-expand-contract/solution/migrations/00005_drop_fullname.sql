-- +goose Up
-- Step 4, contract. Only after no running instance reads or writes fullname. Dropping
-- a column only hides it in the catalog; the space comes back as rows are rewritten.
ALTER TABLE users DROP COLUMN fullname;

-- +goose Down
-- The values are gone; Down restores the shape for development, not the data.
ALTER TABLE users ADD COLUMN fullname text;
