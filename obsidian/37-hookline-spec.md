---
direction: rtl
tags: [senior-course, book]
chapter: 37
---
# Hookline: מפרט וארכיטקטורה

- **ההבטחה:** כל אירוע נמסר חתום, „לפחות פעם אחת”, לכל endpoint מנוי ופעיל; המקבל מסנן לפי מזהה האירוע.
- **שחקנים:** דיירת שולחת, מקבל מקבל, מפעילה מפעילה.
- **מסלול:** קליטה עם outbox, פרסום, פיזור למסירות, שליחה חתומה, וספירה ב־sink.
- **דרישה מול יעד:** דרישה נבדקת בבדיקה; יעד נמדד על המחשב שלי ונרשם עם תאריך.
- **כפילויות:** אילוצי ייחודיות במסד מחליטים, לא קוד שבודק לפני הכנסה.
- **בידוד:** הדיירת מגיעה מהמפתח בלבד; משאב של אחרת הוא 404; מפתח זר מורכב במסד.
- **החלטה טובה:** החלופה הפשוטה, למה לא כאן, ומה היה משנה את דעתי.
- **starter:** הבדיקות הן המפרט; קומיט לכל אבן דרך; „תקועה” היא חצי שעה.
- **שאלה פתוחה:** ___

<div dir="ltr" align="left">

```text
setup     cp -r $COURSE/exercises/project-hookline/hookline-starter/. ~/senior-course/hookline/
build     go build ./... && go vet ./...
tests     make test-unit            (red at first; green package by package)
headers   Hookline-Event-Id  Hookline-Delivery-Id  Hookline-Attempt  Hookline-Signature
signature t=<unix>,v1=<hex HMAC-SHA256(secret, "<t>.<raw body>")>   tolerance 5 min
tables    tenants endpoints subscriptions events outbox deliveries attempts
unique    events(tenant_id, idempotency_key)   deliveries(event_id, endpoint_id)
statuses  pending -> in_flight -> delivered | dead     replay: dead -> pending
binaries  cmd/api  cmd/relay  cmd/dispatcher  cmd/worker  cmd/hookline-admin   sink/ (Python)
proof     scripts/verify-delivery.sh -> "0 missing"; duplicates allowed and counted
my path   ingest -> ___ -> ___ -> ___ -> receiver
```

</div>

קישורים: [[36-system-design-interviews]] · [[38-hookline-milestones]]
