-- EXPLAIN ANALYZE really executes the statement. Inspect a write inside a transaction you roll back.
BEGIN;
EXPLAIN (ANALYZE, BUFFERS) UPDATE orders SET status = 'paid' WHERE id = 42;
ROLLBACK;
