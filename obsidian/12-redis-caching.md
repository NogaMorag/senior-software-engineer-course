---
direction: rtl
tags: [senior-course, book]
chapter: 12
---
# Redis ומטמון

- **Redis:** פקודה אחת רצה עד סופה; לכן כל פקודה אטומית, ו־Lua הופך כמה לאחת.
- **cache-aside:** קריאה מהמטמון; בפספוס טעינה מהמסד ומילוי עם TTL; בכתיבה מחיקה.
- **למחוק, לא לעדכן:** מחיקה אינה תלויה בסדר בין כותבים.
- **אחרי ה־commit:** מחיקה לפניו ממלאת את הישן. גם אחריו נשאר מרוץ, וה־TTL חוסם אותו.
- **TTL:** רשת ביטחון וחוזה התיישנות; תמיד עם פיזור אקראי.
- **פינוי:** ברירת המחדל `noeviction` שגויה למטמון; `allkeys-lru` כנקודת פתיחה.
- **התפרצות:** המספרים שלי מ־`cmd/stampede`: בלי singleflight ___, עם ___.
- **singleflight:** מאחד בתוך תהליך בלבד; context ושגיאה משותפים.
- **מתי לא:** כתיבות רבות, עקביות חזקה, יחס פגיעות נמוך, מקור זול.
- **שאלה פתוחה:** ___

<div dir="ltr" align="left">

```text
docker compose -p ch12 -f data/compose.yaml up -d --wait
docker compose -p ch12 -f data/compose.yaml exec cache redis-cli
SET k v EX 60      GET k      TTL k / PTTL k      DEL k      INCR k
SET lock:k id NX PX 5000      HSET h f v      ZINCRBY z 1 m      ZREVRANGE z 0 -1 WITHSCORES
CONFIG GET maxmemory*         INFO stats   (keyspace_hits, keyspace_misses, evicted_keys)
err == redis.Nil  ->  miss
read:  GET -> miss -> load -> SET ttl+jitter
write: COMMIT -> DEL
singleflight.Group.Do(key, fn) -> (v any, err, shared bool)
redis.NewScript(src).Run(ctx, rdb, keys, args...)   // EVALSHA, then EVAL
my stampede run: without ___ calls, with ___ calls; hit ratio ___
docker compose -p ch12 -f data/compose.yaml down -v
```

</div>

קישורים: [[11-postgres-transactions-migrations]] · [[13-kafka-redpanda]]
