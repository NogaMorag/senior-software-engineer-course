---
direction: rtl
tags: [senior-course, book]
chapter: 14
---
# עבודות רקע וניסיונות חוזרים

- **מתי רקע:** זמן, אמינות, החלקת עומס, תזמון. goroutine אחרי התשובה אינה רקע עמיד.
- **ארבעה מצבים:** ממתינה, רצה, הסתיימה, מתה. ניסיון חוזר מזיז את זמן ההרצה.
- **חכירה:** נעילה רק לרגע הלקיחה; worker שמת משחרר כשהחכירה פגה.
- **גידור:** מונה הניסיונות נבדק בכל סיום; מחזיק ישן לא כותב.
- **המתנה:** תקרה מכפילה עד גבול, ופיזור מלא בין אפס לתקרה; האקראיות מוזרקת.
- **אידמפוטנטיות:** רשומה באותה טרנזקציה, ומפתח לכל מערכת חיצונית.
- **תור מתים:** שומר ראיות; הפעלה חוזרת רק ממצב מת, במנות, אחרי תיקון.
- **מתזמן:** רק מוסיף עבודה, עם מפתח לחלון; CronJob עלול להריץ פעמיים.
- **כיבוי:** להפסיק לקחת, לסיים, לרשום עם context מנותק, ואז לצאת.
- **שאלה פתוחה:** ___

<div dir="ltr" align="left">

```text
docker compose -p ch14 -f data/compose.yaml up -d --wait
export DATABASE_URL=postgres://course:change-me@localhost:5414/jobs
SELECT ... FOR UPDATE SKIP LOCKED      locked_until = now() + $lease
WHERE id = $1 AND state = 'running' AND attempts = $2   -- fencing
ceiling = min(Cap, Base * 2^(attempt-1))     delay = rand in [0, ceiling]
Policy{Base: 100ms, Cap: 2s}: attempt 6 ceiling = ___
UPDATE jobs SET state='pending', attempts=0 WHERE id=$1 AND state='dead'   -- replay
queue.Options{UniqueKey: "name@" + slot}
context.WithTimeout(context.WithoutCancel(ctx), lease)
go run ./cmd/worker -init -enqueue 3 -ms 3000     then: kill -TERM <pid>
docker compose -p ch14 -f data/compose.yaml down -v
```

</div>

קישורים: [[13-kafka-redpanda]] · [[15-timeouts-retries-breakers]]
