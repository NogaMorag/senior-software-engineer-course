---
direction: rtl
tags: [senior-course, book]
chapter: 18
---
# נעילות, בחירת מנהיג וסאגות

- **יעילות מול נכונות:** לנעילת יעילות מספיק Redis אחד; לנעילת נכונות צריך בדיקה בצד האחסון.
- **למה זה נשבר:** עצירת תהליך, שעון שקופץ, רשת שמאחרת. חכירה משתחררת גם כשהמחזיק חי.
- **advisory lock:** סשן או טרנזקציה; נעילת סשן שייכת לחיבור אחד, ולכן חיבור ייעודי ולא המאגר.
- **Redis:** ערך אקראי, תפוגה, ושחרור רק אם הערך שלך. ‏Redlock לא נותן אסימונים.
- **אסימון גידור:** מספר עולה; האחסון דוחה נמוך מהגבוה שראה, בפקודה אחת עם הכתיבה.
- **בחירת מנהיג:** `try`, תקופה חדשה, עבודה ב־goroutine, דופק על אותו חיבור. השתלטות אצלי אחרי ___.
- **קונצנזוס:** רוב, תקופות, שכפול יומן. etcd ו־ZooKeeper, לא מימוש עצמי.
- **סאגה:** צעדים מקומיים, פיצויים בסדר הפוך, אידמפוטנטיים; מה שלא מתבטל אחרון.
- **שאלה פתוחה:** ___

<div dir="ltr" align="left">

```text
SELECT pg_try_advisory_lock(k)        -- session lock, no wait; true/false
SELECT pg_advisory_unlock(k)          -- same session only
SELECT pg_advisory_xact_lock(k)       -- released at COMMIT/ROLLBACK
pg_locks: (classid::bigint << 32) | objid::bigint = k, objsubid = 1
SELECT pg_terminate_backend(pid, 5000)
pgx: pc, _ := pool.Acquire(ctx); conn := pc.Hijack(); defer conn.Close(ctx)
Redis: SET key token NX PX ttl ; release = Lua GET==token then DEL  (8.4+: DELEX key IFEQ token)
fence: INSERT ... ON CONFLICT (name) DO UPDATE ... WHERE t.token <= EXCLUDED.token ; 0 rows = stale
go run ./cmd/elect -name alpha | -name beta ; kill -9 <leader pid> ; takeover after ___
saga: do 1..n ; on failure at i: undo i-1..1 with context.WithoutCancel(ctx)
quorum of 5 = 3, tolerates ___ failures
```

</div>

קישורים: [[17-consistency-replication-partitioning]] · [[19-docker-done-right]]
