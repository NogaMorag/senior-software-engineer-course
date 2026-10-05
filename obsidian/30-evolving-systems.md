---
direction: rtl
tags: [senior-course, book]
chapter: 30
---
# מערכות שמשתנות

- **הרחבה ואז כיווץ:** כל שלב תואם לקוד שרץ לפניו ואחריו; פריסה מלאה בין שלבים.
- **שינוי שם:** להוסיף, לכתוב כפול, למלא במנות, לקרוא מהחדשה, להפסיק לכתוב לישנה, למחוק.
- **כיווץ:** קודם מרפים את העמודה הישנה, ורק בסוף מוחקים אותה.
- **נעילות:** ברירת מחדל קבועה בלי שכתוב; אינדקס במקביל; אילוץ לא תקף ואז אימות.
- **תור נעילות:** פקודה שמחכה חוסמת את כל מי שאחריה; גבול זמן לנעילה בחיבור המיגרציות.
- **אירועים:** תאימות לאחור ולפנים; שינוי מוסיף; קורא סובלני; גרסה באירוע.
- **דגלים:** שחרור, תפעול, ניסוי, הרשאה; כל דגל נולד עם תאריך הסרה.
- **תאנה חונקת:** נקודת ניתוב, חלק אחרי חלק, השוואה, ומחיקה בתוכנית.
- **חוב טכני:** ריבית נמדדת באזורים שמשתנים; משכנעים במחיר, בצעד הפיך ובמדד.
- **שאלה פתוחה:** ___

<div dir="ltr" align="left">

```text
plan      00002 add column -> deploy v2 dual write -> 00003 backfill -> deploy v3 read new
          -> 00004 CHECK NOT VALID, VALIDATE, SET NOT NULL, fullname DROP NOT NULL
          -> deploy v4 new only -> 00005 DROP COLUMN
goose     -- +goose NO TRANSACTION   (CONCURRENTLY, batched backfill, one commit per statement)
locks     ADD COLUMN constant default: catalog only (since PG 11)
          CREATE INDEX CONCURRENTLY: SHARE UPDATE EXCLUSIVE, not in a transaction, invalid on failure
          VALIDATE CONSTRAINT: SHARE UPDATE EXCLUSIVE
          lock_timeout = 2s on the migration connection; SQLSTATE 55P03
long tx   SELECT pid, now() - xact_start, state, query FROM pg_stat_activity ORDER BY 2 DESC NULLS LAST;
events    additive only; tolerant reader; version in the envelope; unknown version -> dead letter
flags     release | ops | experiment | permissioning ; owner + expiry date + removal ticket
churn     git log --since="6 months ago" --format= --name-only | sort | uniq -c | sort -rn | head
test      go test -race -v -run TestRenameWithZeroFailedRequests ./rollout
my run    requests per step ___ ___ ___ ___   failed 0
```

</div>

קישורים: [[29-system-design-method]] · [[31-working-as-senior]]
