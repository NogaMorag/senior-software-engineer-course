-- Chapter 10 solution: one index per slow query. Run inside psql, one block at a time, and
-- re-run the matching query file after each block to see the new plan.

-- Q1: foreign keys are not indexed automatically. Every "items of this order" lookup,
-- and every join from orders to order_items, needs this one.
CREATE INDEX order_items_order_id_idx ON order_items (order_id);

-- Q2: a range on one column: a plain B-tree on that column.
CREATE INDEX orders_created_at_idx ON orders (created_at);

-- Q3: equality column first, then the sort column. One index answers "this customer,
-- newest first, stop after 20" without sorting. id is the tie-breaker the keyset
-- pagination in orders.go relies on. A B-tree can be read backwards, so ASC works for DESC.
CREATE INDEX orders_customer_created_idx ON orders (customer_id, created_at, id);

-- Q4: only about 1% of the rows are pending, and only those are ever queried this way.
-- A partial index holds just those rows: small, cheap to maintain, sorted by created_at.
CREATE INDEX orders_pending_created_idx ON orders (created_at) WHERE status = 'pending';

-- Q5: the Q3 index already finds the rows; adding total_cents as a payload column lets
-- PostgreSQL answer from the index alone (Index Only Scan). Replace, do not add a second
-- index with the same key columns.
DROP INDEX orders_customer_created_idx;
CREATE INDEX orders_customer_created_idx ON orders (customer_id, created_at, id) INCLUDE (total_cents);

-- Index Only Scan needs an up-to-date visibility map.
VACUUM (ANALYZE) orders, order_items;
