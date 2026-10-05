-- Q1: the lines of one order, with product names (an order page).
EXPLAIN (ANALYZE, BUFFERS)
SELECT i.product_id, p.name, i.quantity, i.price_cents
FROM order_items i
JOIN products p ON p.id = i.product_id
WHERE i.order_id = 777777;
