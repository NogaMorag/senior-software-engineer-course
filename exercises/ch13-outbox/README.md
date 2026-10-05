---
direction: rtl
lang: he
---

<div dir="rtl" lang="he">

# תרגיל פרק 13: `outbox`

**מטרה:** להריץ Redpanda ולראות בעיניים איך מפתח קובע מחיצה. לבנות צרכן בקבוצה ששומר offsets רק אחרי
עיבוד, ולראות מה קורה כשהורגים אותו באמצע. ולבנות את תבנית ה־outbox: הזמנה ושורת הודעה באותה טרנזקציה,
relay שמפרסם ואז מסמן, צרכן שמסנן כפילויות, ובדיקה שמפילה את ה־relay בין הפרסום לסימון.

**זמן משוער:** שלושה ימי לימוד. יום 36: כ־35 דקות (Redpanda ו־`rpk`, בלי קוד). יום 37: כ־60 דקות (לולאת
הצרכן ושמירת offsets). יום 38: כ־70 דקות (outbox, relay, סינון ובדיקת הקריסה).

**ידע מקדים:** פרק 13. מפרקים קודמים: Docker ו־compose (פרק 2), context וביטול (פרק 5), pgx וקובץ compose
של PostgreSQL (פרק 10), וטרנזקציות ונעילת שורות (פרק 11, סעיפים 11.1 ו־11.4). התרגיל לא מייבא קוד מתרגיל
קודם: החבילות שלו עומדות בפני עצמן.

## מה בתיקייה הזאת

| קובץ | תפקיד |
| --- | --- |
| `data/compose.yaml` | Redpanda, ה־console ו־PostgreSQL, עם שם הפרויקט `ch13` |
| `starter/db/schema.sql`, `starter/db/db.go` | הטבלאות: הזמנות, outbox, סינון ומשלוחים; מוכן |
| `starter/bus/bus.go` | בניית producer וצרכן בקבוצה, והמפרסם ל־Kafka; מוכן, לקריאה ביום 37 |
| `starter/consumer/consumer.go` | הלולאה `Run` (ב־`TODO` של יום 37: השמירה) והפונקציה `Process` (ב־`TODO` של יום 38) |
| `starter/orders/orders.go` | הפונקציה `Place`; ב־`TODO` של יום 38 |
| `starter/outbox/outbox.go` | `Enqueue` מוכנה; ב־`RunOnce` יש `TODO` של יום 38: לפרסם ואז לסמן |
| `starter/cmd/produce`, `starter/cmd/consume` | תוכניות הדגמה לימים 36 ו־37; מוכנות |
| `starter/testenv/testenv.go` | עזר לבדיקות: סכמה פרטית ב־PostgreSQL ו־topic פרטי לכל בדיקה |
| `tests/bus/bus_test.go` | מפתח ומחיצה, מסירה חוזרת של מה שלא נשמר, rebalance שיתופי (צריך Redpanda) |
| `tests/consumer/consumer_test.go` | הסינון: כפילות, צרכנים שונים, אפקט שנכשל (צריך PostgreSQL) |
| `tests/outbox/outbox_test.go` | הזמנה ושורת outbox יחד, relay בסדר, כשל פרסום, ונקודת הקריסה עם מפרסם מדומה |
| `tests/e2e/crash_test.go` | נקודת הקריסה מקצה לקצה, מול Redpanda ו־PostgreSQL אמיתיים |
| `solution/` | הפתרון המלא, עם עותקים זהים של הבדיקות |

כל הבדיקות קוראות את `DATABASE_URL` ואת `KAFKA_BROKERS`, ומדלגות עם הודעה כשהם לא מוגדרים. כל בדיקה יוצרת
סכמה ו־topic משלה ומוחקת אותם בסוף, ולכן אפשר להריץ אותן שוב ושוב על אותו stack.

## יום 36: Redpanda, topics ומחיצות

**1. העתיקי את קוד ההתחלה ואת קובץ ה־compose.**

<div dir="ltr" align="left">

```bash
mkdir -p ~/senior-course/ch13
cd ~/senior-course/ch13
cp -r "$COURSE"/exercises/ch13-outbox/starter/. .
cp -r "$COURSE"/exercises/ch13-outbox/data .
printf 'bin/\n' > .gitignore
go build ./... && go vet ./... && echo builds
```

</div>

**2. הפעילי את ה־stack** ובדקי שה־broker עונה. ההפעלה הראשונה מורידה שלושה images, וזה לוקח כמה דקות.

<div dir="ltr" align="left">

```bash
docker compose -p ch13 -f data/compose.yaml up -d --wait
docker compose -p ch13 -f data/compose.yaml exec redpanda rpk cluster info
```

</div>

**3. הגדירי את הפונקציה `rpk`** במסוף הנוכחי. היא מריצה את הכלי בתוך הקונטיינר, ועובדת מתוך התיקייה
`~/senior-course/ch13`:

<div dir="ltr" align="left">

```bash
rpk() { docker compose -p ch13 -f data/compose.yaml exec -T redpanda rpk "$@"; }
rpk topic list
```

</div>

**4. topic, מפתחות ומחיצות.** צרי topic עם שלוש מחיצות, כתבי הודעות עם שני מפתחות, וקראי אותן:

<div dir="ltr" align="left">

```bash
rpk topic create demo -p 3
printf 'alice-1\nalice-2\nalice-3\n' | rpk topic produce demo -k alice
printf 'bob-1\nbob-2\n' | rpk topic produce demo -k bob
printf 'carol-1\ncarol-2\n' | rpk topic produce demo -k carol
rpk topic consume demo -n 7 -f '%p %o %k %v\n'
rpk topic describe demo -p
```

</div>

| פקודה | מה היא עשתה |
| --- | --- |
| `topic create demo -p 3` | יצרה topic עם שלוש מחיצות ועותק אחד |
| `topic produce demo -k alice` | כתבה כל שורה מהקלט כרשומה עם המפתח alice |
| `topic consume ... -f '%p %o %k %v\n'` | הדפיסה מחיצה, offset, מפתח וערך לכל רשומה |
| `topic describe demo -p` | הראתה לכל מחיצה את טווח ה־offsets שלה |

**5. פתחי את ה־console** בדפדפן, בכתובת `http://localhost:8088`. מצאי את `demo`, ובדקי שאותן הודעות
מופיעות שם תחת אותן מחיצות.

**6. רשמי בפתק** איזה מפתח נחת באיזו מחיצה. כתבי עוד שלוש הודעות עם `-k alice`, וודאי שהן נחתו באותה
מחיצה, עם offsets שממשיכים מהמקום שבו עצרו.

## יום 37: producer, קבוצה ושמירה אחרי עיבוד

**7. קראי את `bus/bus.go`** ואת ההערות על כל אפשרות. אחר כך הריצי את ה־producer של Go מול אותו topic:

<div dir="ltr" align="left">

```bash
export KAFKA_BROKERS=localhost:19092
go run ./cmd/produce -topic demo -n 6 -keys alice,bob,carol
```

</div>

שימי לב שכל מפתח נחת באותה מחיצה שבה `rpk` שם אותו אתמול: שני הלקוחות מגבבים באותה דרך.

**8. השלימי את השמירה ב־`Run`**, בקובץ `consumer/consumer.go`, לפי ה־`TODO`. רק הרשומות שב־`done`
נשמרות, ואחריהן מגיעה הקריאה ל־`AllowRebalance`. הריצי את בדיקות היום:

<div dir="ltr" align="left">

```bash
go test -race -count=1 -v ./bus
```

</div>

הבדיקה `TestUncommittedRecordsAreRedelivered` היא קריטריון הסיום של היום. היא מדמה חבר שמת אחרי שטיפל
בהודעה ולפני ששמר אותה, ומוודאת שהחבר הבא מתחיל בדיוק ממנה.

**9. רואים את זה קורה.** צרי topic עם ארבע מחיצות ומלאי אותו:

<div dir="ltr" align="left">

```bash
rpk topic create work -p 4
go run ./cmd/produce -topic work -n 40 -keys a,b,c,d,e,f,g,h
```

</div>

פתחי שני מסופים נוספים, ובכל אחד הגדירי שוב את `KAFKA_BROKERS` ועברי לתיקייה. בראשון הפעילי צרכן איטי:

<div dir="ltr" align="left">

```bash
go run ./cmd/consume -topic work -group work-group -name A -sleep 1s
```

</div>

אחרי כמה הודעות, הפעילי בשני מסוף את החבר B עם `-name B` ואותה קבוצה. ב־A תראי שורת `revoked` עם חלק
מהמחיצות בלבד, וב־B שורת `assigned` עם אותן מחיצות. זה rebalance שיתופי.

**10. הריגה באמצע.** חפשי את התהליך של A והרגי אותו בלי שיספיק לשמור:

<div dir="ltr" align="left">

```bash
pkill -9 -f 'consume -topic work -group work-group -name A'
```

</div>

אחרי כמה שניות B יקבל את המחיצות של A. בדקי שההודעות הראשונות ש־B מדפיס מהמחיצות האלה הן הודעות ש־A
כבר הדפיס: אלה הודעות שעובדו ולא נשמרו. עצרי את B עם Ctrl-C, והסתכלי על המצב השמור:

<div dir="ltr" align="left">

```bash
rpk group describe work-group
```

</div>

## יום 38: outbox, relay, סינון והקריסה

**11. השלימי את `Place`**, בקובץ `orders/orders.go`: טרנזקציה אחת עם `pgx.BeginFunc`, הוספת ההזמנה, ואז
`outbox.Enqueue` עם המפתח של הלקוחה.

**12. השלימי את `RunOnce`**, בקובץ `outbox/outbox.go`. הקריאה של השורות מוכנה; נשארו שלושה צעדים, בסדר
הזה בדיוק: לפרסם, לקרוא ל־`AfterPublish` אם הוא קיים, ולסמן.

**13. השלימי את `Process`**, בקובץ `consumer/consumer.go`: שורת סינון ואפקט באותה טרנזקציה.

**14. הבדיקות של היום**, מול ה־stack:

<div dir="ltr" align="left">

```bash
export DATABASE_URL=postgres://course:change-me@localhost:5413/shop
go test -race -count=1 -v ./outbox ./consumer ./e2e
```

</div>

הבדיקה `TestRelayCrashAfterPublishBeforeMark` היא קריטריון הסיום של היום ושל הפרק.

**15. הכול, והכלים.**

<div dir="ltr" align="left">

```bash
go test -race -count=1 ./...
gofmt -l .
go vet ./...
staticcheck ./...
```

</div>

**16. כבי את ה־stack ושמרי את העבודה.** הדגל `-v` מוחק גם את הנפחים, כלומר את כל הנתונים.

<div dir="ltr" align="left">

```bash
docker compose -p ch13 -f data/compose.yaml down -v
cd ~/senior-course
git add ch13
git commit -m "ch13: consumer group, transactional outbox, crash-point test"
```

</div>

## פלט צפוי

בדיקת `rpk` מיום 36. המחיצות של כל מפתח יכולות להיות אחרות אצלך; מה שחייב להתקיים הוא שכל ההודעות של
מפתח אחד באותה מחיצה, ב־offsets עוקבים. גם סדר המחיצות בפלט יכול להשתנות, כי אין סדר בין מחיצות:

<div dir="ltr" align="left">

```text
0 0 alice alice-1
0 1 alice alice-2
0 2 alice alice-3
0 3 bob bob-1
0 4 bob bob-2
2 0 carol carol-1
2 1 carol carol-2
```

</div>

בדיקות יום 37 (הזמנים אצלך יהיו אחרים):

<div dir="ltr" align="left">

```text
    bus_test.go:81: key to partition: map[alice:0 bob:0 carol:2 dave:1 erin:2 frank:1]
--- PASS: TestSameKeySamePartitionInOrder (0.63s)
    bus_test.go:156: first member: m-0 m-1 m-2 m-3 m-4 | second member: m-4 m-5 m-6 m-7 m-8 m-9
--- PASS: TestUncommittedRecordsAreRedelivered (0.42s)
--- PASS: TestCooperativeRebalanceRevokesOnlyMovedPartitions (1.28s)
PASS
ok  	example.com/ch13-outbox/bus	4.230s
```

</div>

והכול, ביום 38:

<div dir="ltr" align="left">

```text
ok  	example.com/ch13-outbox/bus	4.230s
?   	example.com/ch13-outbox/cmd/consume	[no test files]
?   	example.com/ch13-outbox/cmd/produce	[no test files]
ok  	example.com/ch13-outbox/consumer	2.754s
?   	example.com/ch13-outbox/db	[no test files]
ok  	example.com/ch13-outbox/e2e	3.223s
?   	example.com/ch13-outbox/orders	[no test files]
ok  	example.com/ch13-outbox/outbox	3.338s
?   	example.com/ch13-outbox/testenv	[no test files]
```

</div>

## רשימת בדיקה

- [ ] הודעות עם אותו מפתח באותה מחיצה ובסדר, בהרצה שלך עם `rpk`, ורשמת איזה מפתח נחת איפה.
- [ ] `rpk group describe` מראה offsets שמורים ופיגור לכל מחיצה.
- [ ] ראית rebalance שבו A ויתר רק על חלק מהמחיצות.
- [ ] אחרי `kill -9`, החבר הבא עיבד שוב את מה שלא נשמר, ושום הודעה לא חסרה.
- [ ] בדיקת הקריסה עוברת: כל הודעה פעמיים ב־topic, משלוח אחד לכל הזמנה.
- [ ] `go test -race ./...` עובר מול ה־stack, ושלושת הכלים לא מדפיסים דבר.
- [ ] ה־stack כבוי עם `down -v`, והעבודה שמורה ב־Git.

## טעויות נפוצות ותיקונים

- **כל הבדיקות מדלגות עם `set DATABASE_URL` או `set KAFKA_BROKERS`.** המשתנים לא מוגדרים במסוף הזה.
  הגדירי את שניהם כמו בצעדים 7 ו־14, ובדקי שה־stack למעלה עם `docker compose -p ch13 -f data/compose.yaml ps`.
- **`TestUncommittedRecordsAreRedelivered` נכשלת עם `second member started at [m-0 m-1 ...]`.** הלולאה לא
  שומרת offsets בכלל, ולכן החבר השני מתחיל מההתחלה. השלימי את `CommitRecords` בצעד 8.
- **אותה בדיקה נכשלת עם `second member started at [m-5 ...]`.** השמירה כוללת גם את הרשומה שנכשלה, ולכן
  היא אבדה. הרשומות נכנסות ל־`done` רק אחרי שה־handler הצליח.
- **`TestCrashBetweenPublishAndMarkDuplicatesButAppliesOnce` נכשלת עם `pending = 0`.** הסימון רץ לפני
  `AfterPublish`, או בטרנזקציה נפרדת. הסדר הוא פרסום, `AfterPublish`, סימון, הכול בתוך אותה פונקציה.
- **`TestFailedEffectLeavesNoDedupeRow` נכשלת עם `dedupe rows = 1`.** שורת הסינון נכתבה מחוץ לטרנזקציה של
  האפקט, למשל עם `db.Exec` במקום `tx.Exec`. כל הכתיבות בתוך `Process` עוברות דרך `tx`.
- **`rpk` מתלונן על `the input device is not a TTY`.** הגדרת את הפונקציה בלי `-T`, והזרמת קלט עם `printf`.
  הגדירי אותה שוב בדיוק כמו בצעד 3.

## רמזים

- ב־`Run`: בני `cctx` עם `context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)`, קראי
  ל־`cl.CommitRecords(cctx, done...)`, ואז `cancel()`. כשאין רשומות ב־`done`, אין מה לשמור.
- ב־`Place`: בתוך `pgx.BeginFunc` משתמשים ב־`tx.QueryRow` ולא ב־`db`. פונקציה שמחזירה שגיאה גורמת
  ל־`BeginFunc` לגלגל אחורה.
- ב־`RunOnce`: אספי את `m.Seq` של כל ההודעות ל־slice, וסמני את כולן בשאילתה אחת עם `id = ANY($1)`.
- ב־`Process`: `tag.RowsAffected() == 0` אחרי `INSERT ... ON CONFLICT DO NOTHING` פירושו שהמזהה כבר שם.
  במקרה הזה החזירי `nil` מתוך הטרנזקציה, ו־`applied` נשאר `false`.

## הפתרון

הפתרון המלא נמצא בתיקייה `solution/`. ההזמנה ושורת ה־outbox נכתבות באותה טרנזקציה, ולכן אין מצב שבו קיימת
אחת בלי השנייה. ה־relay מסמן רק אחרי שהשרת אישר את כל המנה, ובאותה טרנזקציה שנעלה את השורות, כך שקריסה בכל
נקודה משאירה את השורות לפרסום חוזר: כפילות, לעולם לא אובדן. הצרכן כותב את שורת הסינון ואת המשלוח באותה
טרנזקציה, ושומר offsets רק אחרי שהיא הסתיימה. כך כל כפילות, מה־relay או מ־rebalance, נבלעת בטבלת הסינון.

## מה התרגיל מדגים

- מפתח שקובע מחיצה, וסדר שמובטח רק בתוכה.
- producer עם `acks=all` ואידמפוטנטיות, וכתיבה סינכרונית עם `ProduceSync`.
- צרכן בקבוצה בלי שמירה אוטומטית, ששומר רק מה שהסתיים, ומשחרר rebalance רק אחרי השמירה.
- rebalance שיתופי שמעביר רק חלק מהמחיצות, נבדק מול broker אמיתי.
- תבנית ה־outbox: שתי שורות בטרנזקציה אחת, ו־relay שמפרסם ואז מסמן.
- תפר לבדיקה שמדמה קריסה בנקודה מדויקת, ובדיקה שמוכיחה כפילות בלי אובדן.
- טבלת סינון באותה טרנזקציה של האפקט: „בדיוק פעם אחת” כחוזה.

</div>
