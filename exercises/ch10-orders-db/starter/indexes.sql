-- Chapter 10: your indexes, one block per slow query. Run each block in psql, then re-run the
-- matching file in data/queries/ and record the new plan and timing in the README table.

-- Q1 (data/queries/q1-order-items.sql): TODO

-- Q2 (data/queries/q2-orders-one-day.sql): TODO

-- Q3 (data/queries/q3-customer-latest.sql): TODO

-- Q4 (data/queries/q4-pending-oldest.sql): TODO

-- Q5 (data/queries/q5-customer-monthly.sql): TODO

-- Index Only Scan needs an up-to-date visibility map.
VACUUM (ANALYZE) orders, order_items;
