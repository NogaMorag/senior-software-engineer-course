---
direction: rtl
tags: [senior-course, book]
chapter: 11
---
# PostgreSQL: טרנזקציות, נעילות ומיגרציות

- **‏Read committed:** תמונת מצב לכל פקודה; קריאה ואז כתיבה מאבדת עדכונים.
- **‏Repeatable read:** תמונה אחת לטרנזקציה; עדכון של שורה שהשתנתה נכשל ב־`40001`; write skew עובר.
- **‏Serializable:** עוצר גם write skew; חובה לנסות שוב את כל הטרנזקציה.
- **עדכון אבוד:** ‏`UPDATE` אטומי עם תנאי, או `SELECT ... FOR UPDATE`.
- **תור:** ‏`FOR UPDATE SKIP LOCKED LIMIT 1`; הנעילה מוחזקת כל זמן העיבוד.
- **קיפאון:** ‏`40P01` אחרי `deadlock_timeout`; נועלים לפי סדר מזהים קבוע.
- **goose:** חלק עלייה וחלק ירידה בכל קובץ SQL; לא עורכים מיגרציה שהוחלה.
- **goose מתוך Go:** המיגרציות מוטמעות עם `embed` ורצות תחת advisory lock של סשן.
- **pool:** גודל ה־pool כפול מספר העותקים נשאר מתחת ל־`max_connections`.
- **‏PgBouncer:** במצב טרנזקציה הוא שובר כל מצב של סשן.
- **ניסיון חוזר:** הפונקציה `fn` בלי השפעות חיצוניות; רק כשל סריאליזציה וקיפאון; חמישה ניסיונות עם פיזור אקראי.
- **שאלה פתוחה:** ___

<div dir="ltr" align="left">

```text
up        docker compose -p ch11 -f data/compose.yaml up -d --wait
url       DATABASE_URL=postgres://course:change-me@localhost:5411/lab
tests     go test -race ./...        go test -race -run 'LostUpdate|WriteSkew' ./bank ./oncall
tx        pgx.BeginTxFunc(ctx, pool, pgx.TxOptions{IsoLevel: pgx.Serializable}, fn)
sqlstate  40001 serialization_failure | 40P01 deadlock_detected | 23514 check_violation
lock      SELECT ... FOR UPDATE        SELECT ... FOR UPDATE SKIP LOCKED LIMIT 1
atomic    UPDATE accounts SET balance = balance - $2 WHERE id = $1 AND balance >= $2
waiting   SELECT pid, wait_event_type, query FROM pg_stat_activity WHERE wait_event_type = 'Lock';
deadlock  SHOW deadlock_timeout;   docker compose -p ch11 -f data/compose.yaml logs db | grep -A6 "deadlock detected"
goose     GOOSE_DRIVER=postgres GOOSE_DBSTRING=$DATABASE_URL GOOSE_MIGRATION_DIR=migrations goose status|up|down|down-to 0
create    goose -s create <name> sql
pool      SHOW max_connections;   MaxConns x replicas < max_connections - reserve = ___
retry     db.RetrySerializable(ctx, pool, func(ctx context.Context, tx pgx.Tx) (T, error) { ... })
down      docker compose -p ch11 -f data/compose.yaml down -v
```

</div>

קישורים: [[10-postgres-queries-indexes]] · [[12-redis-caching]]
