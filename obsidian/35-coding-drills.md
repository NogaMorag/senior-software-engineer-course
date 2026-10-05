---
direction: rtl
tags: [senior-course, book]
chapter: 35
---
# תרגילי קוד מעשיים

- **הסבב:** קוד חי של 45 עד 90 דקות, או מטלת בית. מעשי יותר מאלגוריתמי.
- **השיטה:** הבהרה, ממשק, בדיקות, גרסה פשוטה ונכונה, קצוות ומקביליות, מחירים בקול.
- **זמן:** שעון מוזרק. בבדיקה, מחכים למצב ואז מזיזים זמן; לטיימר יחסי, קודם מחכים שהוא קיים.
- **מגביל קצב:** מילוי עצל; דלי מלא שקול למפתח שלא נראה, ולכן מותר לשכוח אותו.
- **מתזמן:** לולאה אחת, טיימר אחד, ערוץ התעוררות שלא חוסם. ביטול גובר על הכול.
- **מטמון:** מפה ורשימה; גם פגיעה כותבת; עצל ופעיל יחד.
- **מקבל webhooks:** ההחלטה בצעד אחד; זוכרים הצלחה, לא הגעה; כמה עותקים, המזהה עובר למסד.
- **הזמנה:** נעילה לכל משאב, או אילוץ החרגה עם `btree_gist`. המסד אוכף מה שאסור שיקרה לעולם.
- **שאלה פתוחה:** ___

<div dir="ltr" align="left">

```text
method      clarify (5) -> API (3) -> tests (7) -> simple + correct (15) -> edges + concurrency (10) -> tradeoffs aloud (5)
clock       now func() time.Time | clock.Clock{Now, NewTimer}; tests: clock.NewFake(t0), Advance, BlockUntil(n)
run         go test -race ./ratelimit   (./scheduler ./ttlcache ./webhook ./booking)
pg          docker compose -p ch35 -f data/compose.yaml up -d --wait
            export DATABASE_URL=postgres://course:change-me@localhost:5435/drills
            docker compose -p ch35 -f data/compose.yaml down -v
limiter     tokens = min(burst, tokens + elapsed*rate); RetryAfter = ceil((1-tokens)/rate)
dedupe      claim under one lock -> leader | follower (wait) | duplicate; remember success only
booking     EXCLUDE USING gist (resource WITH =, during WITH &&)  needs btree_gist; SQLSTATE 23P01
review      risk order -> blocking | should-fix | question | nit | praise -> verdict
my times    limiter ___ scheduler ___ cache ___ webhook ___ booking ___ review ___
```

</div>

קישורים: [[34-senior-interview-loop]] · [[36-system-design-interviews]]
