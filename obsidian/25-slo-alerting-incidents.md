---
direction: rtl
tags: [senior-course, book]
chapter: 25
---
# SLO, התראות, תורנות ותחקיר

- **SLI:** טובים חלקי תקפים, מבוסס בקשות; סף זמן הוא גבול של דלי.
- **SLO:** יעד לאורך 30 ימים; לא 100%; ה־SLA רופף ממנו.
- **תקציב:** אחד פחות היעד; ב־99.5%, 3.6 שעות השבתה בחודש.
- **קצב שריפה:** שיעור שגיאות חלקי תקציב; 1 מוציא בדיוק בסוף החלון.
- **התראות:** 14.4 בשעה וב־5 דקות, 6 בשש שעות וב־30 דקות, קריאה; 1 בשלושה ימים ובשש שעות, כרטיס.
- **חלונות:** הארוך מוכיח גודל, הקצר מוכיח שזה עדיין קורה.
- **promtool:** בדיקה שרעש קצר לא מעיר.
- **תורנות:** כל קריאה דורשת פעולה; סקירה שבועית; העברת משמרת כתובה.
- **תקרית:** להכריז מוקדם, תפקידים, להחזיר לאחור לפני שמוכיחים.
- **תחקיר:** ללא האשמה; פעולות עם בעלים, תאריך וסוג.
- **שאלה פתוחה:** ___

<div dir="ltr" align="left">

```text
SLI = good / valid        budget = 1 - SLO        burn rate = error ratio / budget
page:   1h & 5m > 14.4 x budget    or   6h & 30m > 6 x budget
ticket: 3d & 6h > 1 x budget
job:slo_errors_per_request:ratio_rate5m   job:slo_availability:objective
a > on (job) (14.4 * (1 - b))  and on (job)  ...
bash tests/run_promtool.sh    (promtool check rules / promtool test rules)
export CH24=~/senior-course/ch24 RULES_DIR=~/senior-course/ch25/rules
bash data/stage_incident.sh
curl -s localhost:9090/api/v1/alerts
roles: incident commander, communications lead, operations lead, scribe
my run: change at ___  page at ___  rollback at ___  resolved at ___
```

</div>

קישורים: [[24-telemetry]] · [[26-testing-strategy-load]]
