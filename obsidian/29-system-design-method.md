---
direction: rtl
tags: [senior-course, book]
chapter: 29
---
# שיטת תכנון ומסמכי תכנון

- **שיטה:** דרישות ולא־מטרות, הערכה, API, נתונים, תכנון כללי, צלילות; מותר לחזור אחורה.
- **דרישה:** משהו שאפשר לבדוק, עם הבדיקה לידו.
- **לא־מטרה:** ציפייה סבירה שאני מנטרלת מראש.
- **הערכה:** סדר גודל כדי להחליט; ההנחות בקובץ, עם סיבה.
- **מספרים:** יום הוא כ־10 בחזקת 5 שניות; מיליון ביום הם כ־12 בשנייה.
- **יחידות:** אחסון בחזקות של שתיים, רשת בביטים ובחזקות של עשר.
- **זמני תגובה:** זיכרון, מרכז נתונים, יבשת; כל אחד בסדר גודל של פי מאה ויותר מהקודם.
- **מסמך תכנון:** הקשר, מטרות, הצעה, חלופות כולל לא לעשות כלום, סיכונים, פריסה, שאלות פתוחות.
- **רשומת החלטה:** החלטה אחת, ממוספרת, לא נמחקת אלא מוחלפת.
- **פשרה:** החלופה, למה לא כאן, ומה היה משנה את דעתי.
- **שאלה פתוחה:** ___

<div dir="ltr" align="left">

```text
steps     requirements+non-goals | estimates | API | data model | high level | deep dives
estimate  QPS = per_day / 86_400 ; peak = QPS * peak_factor
storage   bytes * per_request * per_day * retention_days * (1 + index_overhead) * replicas
units     KiB=2^10 MiB=2^20 GiB=2^30 TiB=2^40 ; 1 Gbit/s = 125,000,000 bytes/s
latency   datacenter round trip ~0.5 ms ; CA-Netherlands-CA ~150 ms (Dean 2009, orders of magnitude)
run       uv run sizing ../data/hookline-baseline.toml
design    context | goals/non-goals | estimates | proposal | alternatives | risks | rollout | open questions
adr       Title | Status | Context | Decision ("We will ...") | Consequences   (Nygard 2011)
regions   RTO = time down ; RPO = data lost ; fencing before promotion
my doc    topic ___  sizing file ___  decision recorded in ADR-___
```

</div>

קישורים: [[28-security-backend]] · [[30-evolving-systems]]
