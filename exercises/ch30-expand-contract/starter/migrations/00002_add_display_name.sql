-- +goose Up
-- Step 1, expand. A nullable column without a default is a catalog change only: no rewrite,
-- no scan. It still needs a brief ACCESS EXCLUSIVE lock, which is why the migration
-- connection sets lock_timeout (see db.MigrationConfig).
ALTER TABLE users ADD COLUMN display_name text;

-- +goose Down
ALTER TABLE users DROP COLUMN display_name;
