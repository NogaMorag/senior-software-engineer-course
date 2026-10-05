-- +goose NO TRANSACTION
-- +goose Up
-- Step 2, backfill. Runs after every instance dual-writes, so no new NULLs can appear.
-- One small transaction per batch of ids: row locks are held for one batch only, and a
-- crash leaves finished batches committed (the WHERE clause makes a rerun harmless).
-- +goose StatementBegin
DO $$
DECLARE
    batch   constant bigint := 500;
    last_id bigint;
    from_id bigint := 0;
BEGIN
    SELECT coalesce(max(id), 0) INTO last_id FROM users;
    WHILE from_id < last_id LOOP
        UPDATE users SET display_name = fullname
        WHERE id > from_id AND id <= from_id + batch AND display_name IS NULL;
        COMMIT;
        from_id := from_id + batch;
    END LOOP;
END
$$;
-- +goose StatementEnd

-- +goose Down
-- Nothing to undo: the old column still holds every value.
SELECT 1;
