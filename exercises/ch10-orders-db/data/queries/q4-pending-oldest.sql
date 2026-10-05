-- Q4: the 50 oldest pending orders (what a fulfilment worker picks up next).
EXPLAIN (ANALYZE, BUFFERS)
SELECT id, customer_id, created_at
FROM orders
WHERE status = 'pending'
ORDER BY created_at
LIMIT 50;
