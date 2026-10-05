-- OFFSET versus keyset pagination over the orders of customer 1, page size 20.
-- Run after the Q3 and Q5 indexes exist.

-- Page 1000 with OFFSET: PostgreSQL reads and throws away 19,980 rows first.
EXPLAIN (ANALYZE, BUFFERS)
SELECT id, status, total_cents, created_at
FROM orders
WHERE customer_id = 1
ORDER BY created_at DESC, id DESC
OFFSET 19980 LIMIT 20;

-- The same page with a keyset cursor: "after the last row I saw". Copy created_at and id of the
-- last row of the previous page into the comparison; the values below are an example.
EXPLAIN (ANALYZE, BUFFERS)
SELECT id, status, total_cents, created_at
FROM orders
WHERE customer_id = 1
  AND (created_at, id) < (TIMESTAMPTZ '2025-03-01 00:00+00', 0)
ORDER BY created_at DESC, id DESC
LIMIT 20;
