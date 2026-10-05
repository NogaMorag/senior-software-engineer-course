-- +goose NO TRANSACTION
-- +goose Up
-- Step 3, make the new column the required one and the old one optional. Each statement
-- commits on its own, so no lock is held longer than one statement.
-- A NOT VALID check is only a catalog change: existing rows are not scanned.
ALTER TABLE users ADD CONSTRAINT users_display_name_not_null CHECK (display_name IS NOT NULL) NOT VALID;
-- The scan happens here, under SHARE UPDATE EXCLUSIVE: reads and writes continue.
ALTER TABLE users VALIDATE CONSTRAINT users_display_name_not_null;
-- A valid CHECK that proves there are no NULLs lets SET NOT NULL skip its own scan.
ALTER TABLE users ALTER COLUMN display_name SET NOT NULL;
ALTER TABLE users DROP CONSTRAINT users_display_name_not_null;
-- The next application version stops writing fullname, so it may no longer be required.
ALTER TABLE users ALTER COLUMN fullname DROP NOT NULL;

-- +goose Down
ALTER TABLE users ALTER COLUMN display_name DROP NOT NULL;
