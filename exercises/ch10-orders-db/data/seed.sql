-- Chapter 10 seed: 50,000 customers, 2,000 products, 1,000,000 orders, about 2,000,000 order items.
-- Deterministic (setseed) and skewed on purpose:
--   * customers: a few heavy buyers. customer_id = 1 + floor(50000 * random()^3), so customer 1 has
--     tens of thousands of orders while a typical customer has about ten.
--   * products: popular products sell more (random()^2).
--   * time: orders cluster near 2026-10-01 (random()^2 over two years).
--   * status: about 1% pending, 2% cancelled, 3% paid, 6% shipped, the rest delivered.
-- Run it from psql (see the README) and measure how long it takes on your machine.
\set ON_ERROR_STOP on
SELECT setseed(0.42);

INSERT INTO customers (id, email, name, country, created_at)
SELECT i,
       'customer' || i || '@example.com',
       'Customer ' || i,
       (ARRAY['IL', 'US', 'DE', 'GB', 'FR', 'IN', 'BR', 'JP'])[1 + floor(power(random(), 2) * 8)::int],
       TIMESTAMPTZ '2024-01-01 00:00+00' + random() * INTERVAL '640 days'
FROM generate_series(1, 50000) AS i;

INSERT INTO products (id, name, price_cents)
SELECT i, 'Product ' || i, 500 + floor(random() * 20000)::bigint
FROM generate_series(1, 2000) AS i;

-- Items first (1 to 3 per order, chosen by a hash of the order id), so order totals can be exact.
CREATE TEMP TABLE staged_items AS
SELECT x.order_id, x.product_id, x.quantity, p.price_cents
FROM (
    SELECT o AS order_id,
           1 + floor(2000 * power(random(), 2))::bigint AS product_id,
           1 + floor(random() * 3)::int AS quantity
    FROM generate_series(1, 1000000) AS o
    CROSS JOIN LATERAL generate_series(1, 1 + abs(hashint8(o)) % 3) AS n
) AS x
JOIN products p ON p.id = x.product_id;

INSERT INTO orders (id, customer_id, status, total_cents, created_at)
SELECT s.i,
       1 + floor(50000 * power(random(), 3))::bigint,
       CASE WHEN s.r < 0.01 THEN 'pending'
            WHEN s.r < 0.03 THEN 'cancelled'
            WHEN s.r < 0.06 THEN 'paid'
            WHEN s.r < 0.12 THEN 'shipped'
            ELSE 'delivered' END,
       t.total,
       TIMESTAMPTZ '2026-10-01 00:00+00' - power(random(), 2) * INTERVAL '730 days'
FROM (SELECT i, random() AS r FROM generate_series(1, 1000000) AS i) AS s
JOIN (SELECT order_id, sum(quantity * price_cents) AS total FROM staged_items GROUP BY order_id) AS t
  ON t.order_id = s.i;

INSERT INTO order_items (order_id, product_id, quantity, price_cents)
SELECT order_id, product_id, quantity, price_cents FROM staged_items ORDER BY order_id;

DROP TABLE staged_items;

-- Identity columns continue after the explicit ids we inserted.
SELECT setval(pg_get_serial_sequence('customers', 'id'), (SELECT max(id) FROM customers));
SELECT setval(pg_get_serial_sequence('products', 'id'), (SELECT max(id) FROM products));
SELECT setval(pg_get_serial_sequence('orders', 'id'), (SELECT max(id) FROM orders));

-- Fresh statistics for the planner and a fresh visibility map (needed for Index Only Scan).
VACUUM (ANALYZE) customers, products, orders, order_items;

SELECT (SELECT count(*) FROM customers) AS customers,
       (SELECT count(*) FROM orders) AS orders,
       (SELECT count(*) FROM order_items) AS order_items;
