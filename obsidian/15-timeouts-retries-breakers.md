---
direction: rtl
tags: [senior-course, book]
chapter: 15
---
# זמנים קצובים, ניסיונות חוזרים ומפסקי זרם

- **כשל חלקי:** אין תשובה זה לא נכשל; ניסיון חוזר רק לפעולה אידמפוטנטית.
- **תקציב זמן:** גבול מוחלט אחד לבקשה, בכותרת בין שירותים, רזרבה בכל קפיצה.
- **ניסיון חוזר:** רק לכשל זמני, לא לסירוב שלי, ולא מעבר לגבול הזמן.
- **פיזור אקראי מלא:** הגרלה בין אפס לתקרה שמכפילה את עצמה.
- **תקציב ניסיונות:** כל בקשה מפקידה עשירית, כל ניסיון חוזר מושך אחד, משותף לתלות.
- **סערה:** ניסיונות בחזקת מספר השכבות; שכבה אחת בעלת הניסיונות.
- **מפסק זרם:** סגור, פתוח, חצי־פתוח; דור מתעלם מתוצאות ישנות; ביטול לא נספר.
- **מחיצה אטומה:** סמפור לכל תלות, סירוב מהיר או המתנה קצרה בתוך ה־context.
- **הידרדרות:** ערך ישן עם גבול, ברירת מחדל, תכונה כבויה; החלטת מוצר מראש.
- **שאלה פתוחה:** ___

<div dir="ltr" align="left">

```text
deadline.Middleware(max, h)    deadline.Downstream(ctx, reserve)    deadline.Inject(req)
header X-Timeout-Ms (gRPC: grpc-timeout)
sleep = random(0, min(cap, base * 2^attempt))      # full jitter, Brooker 2015
retry.NewBudget(0.1, 10): Deposit per call, Withdraw per retry
storm: 3 layers x 3 attempts = 27 upstream calls per request
breaker: closed -> open (N failures) -> half-open (timeout) -> closed (K ok) | open (1 fail)
bulkhead: slots := make(chan struct{}, n); acquire = send, release = receive
go run ./cmd/lab -scenario storm|deadline|breaker|bulkhead|degrade
my storm budget row: A ___ B ___ C ___ upstream ___
my breaker run: upstream received without ___ with ___
my bulkhead run: fast ok without ___ with ___
```

</div>

קישורים: [[14-background-jobs]] · [[16-overload-rate-limiting]]
