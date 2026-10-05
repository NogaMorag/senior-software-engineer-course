---
direction: rtl
tags: [senior-course, book]
chapter: 10
---
# PostgreSQL: שאילתות ואינדקסים

- **קודם מודדים:** ‏`EXPLAIN (ANALYZE, BUFFERS)` על נתונים בגודל ובצורה אמיתיים; `EXPLAIN` לבד הוא רק הערכה.
- **קוראים מבפנים החוצה:** שורות מוערכות מול בפועל, מספר הלולאות, שורות שסוננו ודפים שנקראו.
- **מפתח זר לא מאונדקס אוטומטית.** זה החשוד הראשון בכל חיבור איטי.
- **מרוכב:** שוויון קודם, טווח או מיון אחר כך. הסדר ההפוך כותב `Index Cond` וקורא הכול.
- **חלקי:** כשהשאילתה תמיד מסננת לאותה תת־קבוצה קטנה; התנאי חייב להופיע מילולית בשאילתה.
- **מכסה:** עמודות ב־`INCLUDE` נותנות סריקת אינדקס בלבד, בתנאי שמפת הנראות עדכנית (`Heap Fetches: 0`).
- **מחליפים אינדקס, לא מוסיפים כמעט זהה.**
- **‏N+1:** סופרים שאילתות; מתקנים עם `= ANY($1)`.
- **עימוד keyset:** ‏`(created_at, id) < ($2, $3)` עם אינדקס באותו סדר; מבקשים `limit+1`.
- **שאלה פתוחה:** ___

<div dir="ltr" align="left">

```text
up        docker compose -p ch10 -f data/compose.yaml up -d --wait
psql      docker compose -p ch10 -f data/compose.yaml exec db psql -U course -d orders
seed      time docker compose -p ch10 -f data/compose.yaml exec -T db psql -U course -d orders -f /work/data/seed.sql
url       DATABASE_URL=postgres://course:change-me@localhost:5410/orders
psql      \dt  \d orders  \di+  \timing on  \x auto  \i /work/data/queries/q1-order-items.sql
explain   EXPLAIN (ANALYZE, BUFFERS) <query>;   BEGIN; EXPLAIN ANALYZE <dml>; ROLLBACK;
fk        CREATE INDEX order_items_order_id_idx ON order_items (order_id);
composite CREATE INDEX ... ON orders (customer_id, created_at, id) INCLUDE (total_cents);
partial   CREATE INDEX ... ON orders (created_at) WHERE status = 'pending';
keyset    WHERE customer_id = $1 AND (created_at, id) < ($2, $3) ORDER BY created_at DESC, id DESC LIMIT n+1
nplus1    WHERE order_id = ANY($1)        go run ./cmd/nplus1 -customer 1 -limit 200
seed time on my machine: ___ s     Q1 before/after: ___ / ___ ms
down      docker compose -p ch10 -f data/compose.yaml down -v
```

</div>

קישורים: [[09-api-design-grpc]] · [[11-postgres-transactions-migrations]]
