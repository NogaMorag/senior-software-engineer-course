-- Q2: how many orders, and how much money, on one day a year ago (a daily report).
EXPLAIN (ANALYZE, BUFFERS)
SELECT count(*), sum(total_cents)
FROM orders
WHERE created_at >= TIMESTAMPTZ '2025-10-01 00:00+00'
  AND created_at <  TIMESTAMPTZ '2025-10-02 00:00+00';
