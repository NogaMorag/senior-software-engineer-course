---
direction: rtl
tags: [senior-course, book]
chapter: 39
---
# הערכה, הדגמה וראיון על הפרויקט

- **מדידה:** פרופיל, חימום, אחוזון, מחשב ותאריך; סף שנכשל על מחשב עמוס בודקים שוב.
- **‏N2 משני מקורות:** ההיסטוגרמה רואה רק מה שהגיע; הבדיקה בסוף הניסוי רואה הכול.
- **כאוס:** השערה לפני, ‏SIGKILL, ‏`SEED`, ופירוק הפיגור לשלבים.
- **הממצא:** זמן סשן של 45 שניות עיכב פיזור ביותר מדקה; 10 שניות תיקנו.
- **ההוכחה:** סט צפוי מהמסד בלי שורות המסירה, מול הקבלות; כפילויות נספרות.
- **מה היא לא מוכיחה:** רשת אמיתית, דיסק שאבד, מקבלים שאני לא רואה.
- **עלות:** משאבים שמדדתי, ומחיר מדף עם תאריך.
- **פרסום:** פרטי, בדיקת סודות, ‏CI ירוק, ואז פומבי.
- **שאלה פתוחה:** ___

<div dir="ltr" align="left">

```text
n1        make load-ingest          grep 'phase:measure,expected_response:true' /tmp/hookline-load/ingest-summary.txt
n5 n6     make load-two-tenants     make load-overload   (MAX_IN_FLIGHT=4 RATE=800)
chaos     SEED=4242 make chaos      (RATE=50 DURATION=120 KILL_EVERY=10 SETTLE=300)
verify    SINCE='<ts>' WAIT=300 make verify-delivery     -> "PASS: 0 missing"
stages    SINCE='<ts>' ENDPOINT=chaos-healthy scripts/stage-lag.sh   relay_s dispatch_s deliver_s
missing   none = fan-out lost   pending/in_flight = wait   dead = see last_error   delivered = sink restarted?
session   kgo.SessionTimeout(10s) kgo.HeartbeatInterval(1s)   default 45s
cost      docker stats --no-stream      /tmp/hookline-chaos/resources.txt     per 1M = monthly / (rate*86400*30) * 1e6
secrets   git grep -nE 'hk_[A-Za-z0-9_-]{43}|whsec_[A-Za-z0-9_-]{43}'   no .env tfstate tfvars kubeconfig
publish   gh repo create hookline --private --source . --push   gh run watch   then --visibility public
base      docker buildx imagetools inspect python:3.14.8-slim  -> new digest -> both FROM lines -> build --pull
my n1 p99 ___ ms   my chaos: expected ___ missing ___ duplicates ___   healthy p99 ___ s
```

</div>

קישורים: [[38-hookline-milestones]]
