-- Q3: a customer's 20 latest orders (the "my orders" page). Customer 4242 is typical;
-- try customer 1 too, the heaviest buyer in the data.
EXPLAIN (ANALYZE, BUFFERS)
SELECT id, status, total_cents, created_at
FROM orders
WHERE customer_id = 4242
ORDER BY created_at DESC
LIMIT 20;
