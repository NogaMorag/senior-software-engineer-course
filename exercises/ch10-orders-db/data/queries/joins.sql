-- Two joins, two strategies. Not one of the five slow queries: read the plans.

-- Few outer rows, an index on the inner side: Nested Loop (one index probe per outer row).
EXPLAIN (ANALYZE, BUFFERS)
SELECT o.id, i.product_id, i.quantity
FROM orders o
JOIN order_items i ON i.order_id = o.id
WHERE o.customer_id = 4242;

-- Many rows on both sides: Hash Join (build a hash table from one side, probe it with the other).
EXPLAIN (ANALYZE, BUFFERS)
SELECT i.product_id, sum(i.quantity) AS units
FROM orders o
JOIN order_items i ON i.order_id = o.id
WHERE o.created_at >= TIMESTAMPTZ '2026-09-01 00:00+00'
GROUP BY i.product_id
ORDER BY units DESC
LIMIT 10;
