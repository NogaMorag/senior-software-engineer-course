---
direction: rtl
tags: [senior-course, book]
chapter: 13
---
# Kafka עם Redpanda

- **יומן מול תור:** יומן שומר ומתקדם לפי offset; תור מוחק אחרי אישור.
- **מחיצה:** הסדר מובטח רק בתוכה. המפתח קובע מחיצה, ולכן קובע מה נשאר בסדר.
- **אישורי כתיבה:** אישור מכל העותקים ועוד מינימום מסונכרנים מגן מאובדן; עם עותק אחד זה רק המוביל.
- **producer אידמפוטנטי:** סופג ניסיונות חוזרים של הלקוח, לא שליחה חוזרת של האפליקציה.
- **שמירת offset:** אחרי עיבוד נותנת „לפחות פעם אחת”; לפני עיבוד נותנת „לכל היותר פעם אחת”.
- **rebalance:** שיתופי לוקח רק מה שזז; לחסום עד שמירה, ואז לשחרר.
- **outbox:** שורת הודעה באותה טרנזקציה; relay מפרסם ואז מסמן; קריסה ביניהם יוצרת כפילות, לא אובדן.
- **חוזה „בדיוק פעם אחת”:** מזהה יציב בשולח, סינון באותה טרנזקציה בצרכן; לא לפי offset.
- **שאלה פתוחה:** ___

<div dir="ltr" align="left">

```text
docker compose -p ch13 -f data/compose.yaml up -d --wait
rpk() { docker compose -p ch13 -f data/compose.yaml exec -T redpanda rpk "$@"; }
rpk cluster info      rpk topic create demo -p 3      rpk topic describe demo -p
printf 'a\nb\n' | rpk topic produce demo -k alice
rpk topic consume demo -n 5 -f '%p %o %k %v\n'
rpk group describe demo-group
export KAFKA_BROKERS=localhost:19092
export DATABASE_URL=postgres://course:change-me@localhost:5413/shop
kgo.RequiredAcks(kgo.AllISRAcks())   kgo.DisableAutoCommit()   kgo.BlockRebalanceOnPoll()
cl.CommitRecords(ctx, done...); cl.AllowRebalance()
my key for alice landed on partition ___
docker compose -p ch13 -f data/compose.yaml down -v
```

</div>

קישורים: [[12-redis-caching]] · [[14-background-jobs]]
