-- Q5: monthly spend of the heaviest customer in 2026 (an account dashboard).
EXPLAIN (ANALYZE, BUFFERS)
SELECT date_trunc('month', created_at) AS month, count(*), sum(total_cents)
FROM orders
WHERE customer_id = 1
  AND created_at >= TIMESTAMPTZ '2026-01-01 00:00+00'
GROUP BY 1
ORDER BY 1;
