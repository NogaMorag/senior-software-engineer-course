---
direction: rtl
tags: [senior-course, book]
chapter: 17
---
# עקביות, שכפול וחלוקה למחיצות

- **מודלים:** ליניאריזביליות, סדרתית, סיבתית, לקרוא את מה שכתבת, מונוטוניות, בסופו של דבר. שתיים מהן הבטחות לסשן.
- **שכפול ב־PostgreSQL:** ‏WAL זורם לשרת משני; אסינכרוני כברירת מחדל; המשני קורא בלבד.
- **הקמה:** משתמש שכפול, שורה למסד `replication`, ‏slot, ואז עותק עם `-R` שיוצר את קובץ הסימון ואת מחרוזת החיבור.
- **מדידה:** בראשי בבתים ובזמנים, במשני לפי גיל, בלקוח לפי זמן עד שהשורה נראית. כל אחד עם מלכודת.
- **הפיגור שלי:** חציון ___, ‏p95 ___ בראשי שקט; חציון ___ עם עומס; ___ עם השהיה של שתי שניות.
- **לקרוא את מה שכתבת:** חלון לפי הפיגור הגרוע, או אסימון LSN עם נפילה לראשי.
- **חלוקה:** hash לפיזור, טווחים לשאילתות טווח. המפתח נבחר לפי השאילתה החמה והסדר.
- **מחיצה חמה:** מקדם חום ___ ב־Zipf, ___ עם מפתח מורכב. עוד מחיצות לא עוזרות למפתח חם.
- **CAP ו־PACELC:** ‏C היא ליניאריזביליות; בלי פיצול, הבחירה היא זמן תגובה מול עקביות.
- **שאלה פתוחה:** ___

<div dir="ltr" align="left">

```text
docker compose -p ch17 -f data/compose.yaml up -d --wait     # primary :5417, replica :5427
pg_basebackup -h primary -U replicator -D "$PGDATA" --wal-method=stream --slot=replica1 --write-recovery-conf
  -> standby.signal + primary_conninfo + primary_slot_name in postgresql.auto.conf
primary:  SELECT application_name, state, sync_state, replay_lsn, replay_lag FROM pg_stat_replication;
primary:  SELECT pg_wal_lsn_diff(pg_current_wal_lsn(), replay_lsn) FROM pg_stat_replication;
replica:  SELECT pg_is_in_recovery(), pg_last_wal_replay_lsn(), now() - pg_last_xact_replay_timestamp();
replica:  ALTER SYSTEM SET recovery_min_apply_delay = 2000; SELECT pg_reload_conf();
RYW:      token := pg_current_wal_lsn() after write; replica ok when pg_last_wal_replay_lsn() >= token::pg_lsn
go run ./cmd/probe -n 50 [-load 4] [-delay 2s]      median ___  p95 ___  missed ___
go run ./cmd/skew -key uniform|zipf|composite [-partitioner range]    hot factor ___ / ___ / ___
rand.NewZipf(r, s>1, v>=1, imax)    ring 4->5 nodes moves ~1/5, modulo moves ~4/5
PACELC: if P then A or C, else L or C
```

</div>

קישורים: [[16-overload-rate-limiting]] · [[18-locks-leaders-sagas]]
