---
direction: rtl
lang: he
---

<div dir="rtl" lang="he">

# תרגיל פרק 17: `replica-lag`, פיגור שכפול, לקרוא את מה שכתבת, ומחיצות חמות

**מטרה:** להקים שרת PostgreSQL ראשי ושרת משני עם שכפול זורם, למדוד את פיגור השכפול שלך בשלושה מצבים, לייצר
קריאה שמפספסת את הכתיבה של עצמה ולתקן אותה בהדבקה לסשן ובאסימון LSN, ולהריץ סימולטור שמראה איך מפתח חלוקה
יוצר מחיצה חמה. את המספרים שתמדדי את רושמת בפתק, ולא מעתיקה מהספר.

**זמן משוער:** שני ימי לימוד. יום 50: כ־65 דקות בנייה (השכפול, הבדיקה ו„לקרוא את מה שכתבת”). יום 51: כ־35 דקות
(הסימולטור והגיבוב העקבי).

**ידע מקדים:** פרק 17; פרק 10 (PostgreSQL בקונטיינר, psql ו־pgx, סעיפים 10.1 ו־10.7); פרק 11, סעיף 11.1 (WAL);
goroutines, ‏context ו־`-race` (פרק 5, סעיפים 5.1, 5.4 ו־5.8). התרגיל לא מייבא קוד מפרקים קודמים. הספרייה היחידה
היא pgx, בגרסה שמקובעת ב־`go.mod` (נכון לאוקטובר 2026, ראי נספח ג).

## מה בתיקייה הזאת

| קובץ | תפקיד |
| --- | --- |
| `data/compose.yaml` | שני שרתי PostgreSQL 18 בפרויקט `ch17`: ראשי בפורט 5417, משני בפורט 5427 |
| `data/primary-init.sh` | רץ פעם אחת בראשי: משתמש שכפול, ‏slot בשם `replica1`, ושורה ב־`pg_hba.conf` |
| `data/replica-entrypoint.sh` | נקודת הכניסה של המשני: `pg_basebackup` עם `-R` כשהתיקייה ריקה, ואז השרת |
| `starter/replica/replica.go` | ‏`Connect`, ‏`Setup` שלמים; ‏`Probe` ו־`Summarize` עם `TODO` |
| `starter/replica/status.go` | ‏`Senders`, ‏`ReplayAge`, ‏`SetApplyDelay`, ‏`Load` שלמים: לקרוא ולהריץ |
| `starter/replica/ryw.go` | ‏`CurrentLSN`, ‏`CaughtUp`, ‏`WaitForLSN`, ‏`WriteNote` שלמים; ‏`ReadNote` עם `TODO` |
| `starter/replica/router.go` | הדבקה לסשן: `RecordWrite` שלמה; ‏`ForRead` עם `TODO` |
| `starter/skew/skew.go` | מחוללי מפתחות ומחלקים: `Zipf`, ‏`Hash.Partition`, ‏`Range.Partition`, ‏`HotFactor` עם `TODO` |
| `starter/skew/ring.go` | גיבוב עקבי: `NewRing` ו־`Moved` שלמות; ‏`Ring.Node` עם `TODO` |
| `starter/cmd/probe/main.go` | התוכנית שמודדת פיגור (שלמה) |
| `starter/cmd/skew/main.go` | התוכנית שמריצה את הסימולטור (שלמה) |
| `tests/replica/*_test.go`, ‏`tests/skew/skew_test.go` | הבדיקות; עותק זהה בכל חבילה ב־`starter/` וב־`solution/` |
| `solution/` | הפתרון המלא |

הבדיקות של החבילה `replica` משנות את `recovery_min_apply_delay` במשני ומחזירות אותו לאפס בסוף כל בדיקה. לכן הן
לא רצות במקביל זו לזו, וכדאי לא להריץ את התוכנית `probe` עם `-delay` באותו זמן.

## יום 50: השכפול, הבדיקה ו„לקרוא את מה שכתבת”

**1. העתקה והקמה.**

<div dir="ltr" align="left">

```bash
mkdir -p ~/senior-course/ch17
cp -r $COURSE/exercises/ch17-replica-lag/starter/. ~/senior-course/ch17/
cp -r $COURSE/exercises/ch17-replica-lag/data ~/senior-course/ch17/data
cd ~/senior-course/ch17
docker compose -p ch17 -f data/compose.yaml up -d --wait
export DATABASE_URL=postgres://course:change-me@localhost:5417/lab
export REPLICA_URL=postgres://course:change-me@localhost:5427/lab
```

</div>

ההקמה הראשונה לוקחת כמה שניות יותר מהרגיל: המשני מחכה שהראשי יהיה בריא, ואז מעתיק אותו.

**2. ודאי שהשכפול חי.** הריצי את שלוש הבדיקות של סעיף 17.2.4: הקובץ `postgresql.auto.conf` וקובץ הסימון
`standby.signal` במשני, ו־`pg_stat_replication` בראשי. אחר כך שאלי את המשני אם הוא במצב שחזור, ונסי לכתוב בו:

<div dir="ltr" align="left">

```bash
docker compose -p ch17 -f data/compose.yaml exec replica psql -U course -d lab -c "SELECT pg_is_in_recovery()"
docker compose -p ch17 -f data/compose.yaml exec replica psql -U course -d lab -c "CREATE TABLE nope (id int)"
```

</div>

השגיאה `cannot execute CREATE TABLE in a read-only transaction` היא התשובה הנכונה.

**3. מלאי את הבדיקה ואת הסיכום** בקובץ `replica/replica.go`:

| פונקציה | מה |
| --- | --- |
| `Probe` | ‏`INSERT ... RETURNING id` בראשי, חותמת זמן, ואז לולאה: `SELECT EXISTS (...)` במשני, ואם לא נמצא, המתנה של `interval` או יציאה כש־`ctx` נגמר |
| `Summarize` | מיון של הזמנים, מינימום ומקסימום, חציון ו־p95 בדרגה הקרובה, וספירה של בדיקות עם `Polls` מעל 1 |

בקובץ חסרים הייבואים של `math` ושל `slices`, כי אף פונקציה שלמה בו לא משתמשת בהם; הוסיפי אותם (או תני לעורך להוסיף בשמירה).

<div dir="ltr" align="left">

```bash
go test -race -v -run 'Summarize|Probe|ApplyDelay|Senders|Load' ./replica
```

</div>

**4. מדידה.** הריצי את התוכנית בשלושת המצבים של סעיף 17.3.3, כל אחד פעמיים, ורשמי בפתק את שורת ה־`summary`
ואת שורת ה־`pg_stat_replication` של כל הרצה:

<div dir="ltr" align="left">

```bash
go run ./cmd/probe -n 50
go run ./cmd/probe -n 50 -load 4
go run ./cmd/probe -n 10 -delay 2s
```

</div>

התוכנית מחזירה את ההשהיה לאפס בסוף, גם אחרי Ctrl+C. אם התהליך נהרג באמצע, או שבדיקה נעצרה ב־timeout של
`go test`, ההשהיה נשארת במשני. כך מאפסים אותה ביד:

<div dir="ltr" align="left">

```bash
docker compose -p ch17 -f data/compose.yaml exec replica psql -U course -d lab \
  -c "ALTER SYSTEM RESET recovery_min_apply_delay" -c "SELECT pg_reload_conf()"
```

</div>

**5. ההגנה של הסשן.** מלאי שתי פונקציות:

| קובץ | פונקציה | מה |
| --- | --- | --- |
| `replica/router.go` | `ForRead` | ראשי אם עברו פחות מ־`Window` מאז הכתיבה האחרונה של המשתמשת; אחרת משני, ומחיקה של הרשומה |
| `replica/ryw.go` | `ReadNote` | אסימון ריק: ישר למשני. אחרת `WaitForLSN` עם `context.WithTimeout` של `maxWait`; הצלחה: משני; כשל: ראשי |

<div dir="ltr" align="left">

```bash
go test -race -v -run 'Router|ReadYourWrites' ./replica
```

</div>

**6. בפתק:** איזה חלון היית קובעת ל־`Router` לפי המדידות שלך, ולמה לפי ה־p95 ולא לפי החציון.

## יום 51: מחיצות חמות וגיבוב עקבי

**7. מלאי את הסימולטור** בקובץ `skew/skew.go`, ואת `Ring.Node` בקובץ `skew/ring.go`:

| פונקציה | מה |
| --- | --- |
| `Zipf` | ‏`rand.NewZipf(r, s, 1, uint64(tenants-1))` פעם אחת, ופונקציה שמחזירה `tenantName(int(z.Uint64()))` |
| `Hash.Partition` | ‏`hash64(key)` מודולו `N` |
| `Range.Partition` | ‏`sort.Search` על `Bounds`: האינדקס הראשון שהגבול שלו גדול מהמפתח |
| `Load.HotFactor` | הספירה הגבוהה כפול מספר המחיצות, חלקי הסכום; אפס כשאין אירועים |
| `Ring.Node` | ‏`sort.Search` על `points` לנקודה הראשונה שגדולה או שווה ל־hash, עם חזרה ל־0 בסוף המעגל |

בשני הקבצים חסר הייבוא של `sort`; הוסיפי אותו כשתגיעי לחיפוש הבינארי.

<div dir="ltr" align="left">

```bash
go test -race -v ./skew
```

</div>

**8. שלוש הרצות, ועוד אחת.** שמרי את הפלט של כל אחת בפתק:

<div dir="ltr" align="left">

```bash
go run ./cmd/skew -key uniform
go run ./cmd/skew -key zipf
go run ./cmd/skew -key composite
go run ./cmd/skew -key zipf -partitioner range
```

</div>

שחקי עם `-s` (למשל 1.05 ו־1.5) ועם `-partitions 32`, וראי שמקדם החום עם מפתח חם לא יורד כשמוסיפים מחיצות.

**9. בפתק:** איזה מפתח היית בוחרת לאירועים של Hookline, ולמה, ומה היה משנה את דעתך (סעיף 17.6.3).

**10. כל הבדיקות, התחייבות וכיבוי:**

<div dir="ltr" align="left">

```bash
go test -race ./...
gofmt -l . ; go vet ./... ; staticcheck ./...
cd ~/senior-course
git add ch17
git commit -m "Chapter 17: replica lag probe, read-your-writes, partition skew simulator"
docker compose -p ch17 -f ch17/data/compose.yaml down -v
```

</div>

הדגל `-v` של `down` מוחק את הכרכים. בפעם הבאה שתרימי את הסביבה, המשני ישוכפל מחדש מראשי חדש.

## פלט צפוי

הבדיקות המלאות בסוף יום 51 (הזמנים יהיו שונים אצלך):

<div dir="ltr" align="left">

```text
?   	example.com/ch17-replica-lag/cmd/probe	[no test files]
?   	example.com/ch17-replica-lag/cmd/skew	[no test files]
ok  	example.com/ch17-replica-lag/replica	…s
ok  	example.com/ch17-replica-lag/skew	…s
```

</div>

עם `-v`, הבדיקות מדפיסות את מה שמדדו. בבדיקה עם ההשהיה, הפיגור יהיה קצת מעל שנייה, אחרי עשרות קריאות:

<div dir="ltr" align="left">

```text
replica_test.go:…: observed lag over 20 probes: min …, median …, p95 …, max …; first read missed … times
replica_test.go:…: with recovery_min_apply_delay = 1s: lag 1.0…s after … polls
replica_test.go:…: token 0/…: replica caught up after 1.0…s
skew_test.go:…: keys moved: ring 21.0%, modulo 80.1%
```

</div>

השורה האחרונה זהה אצלך, כי המפתחות קבועים. גם הפלט של `cmd/skew` זהה בכל מחשב עם אותו `-seed`; עם Zipf
ו־hash, מקדם החום הוא 2.14.

## רשימת בדיקה

- [ ] ‏`pg_stat_replication` מראה משני אחד, `streaming` ו־`async`, וכתיבה במשני נכשלת.
- [ ] שלושת המצבים של `probe` רצו פעמיים כל אחד, והשורות רשומות בפתק.
- [ ] ‏`TestReadYourWrites` עוברת: הקריאה הנאיבית מקבלת `ErrNotFound`, והקריאה עם האסימון רואה את ההערה.
- [ ] ‏`TestRouterSticksToPrimaryAfterAWrite` עוברת בלי שום המתנה אמיתית.
- [ ] ארבע הרצות של הסימולטור בפתק, וההחלטה על המפתח של Hookline כתובה.
- [ ] ‏`go test -race ./...` ירוק בלי `SKIP`; `gofmt`, ‏`go vet` ו־`staticcheck` נקיים.
- [ ] ‏`down -v` רץ בסוף.

## טעויות נפוצות ותיקונים

- **המשני לא עולה, ובלוג שלו `no pg_hba.conf entry for replication connection`.** השורה למסד `replication` חסרה,
  כי הכרך של הראשי נוצר לפני שהתסריט היה שם, והתסריט רץ רק על תיקייה חדשה. ‏`down -v` ואז `up` מחדש.
- **‏`pg_basebackup: error: replication slot "replica1" does not exist`.** אותה סיבה, מהצד של ה־slot. אותו תיקון.
- **‏`TestProbeSeesItsWriteOnTheReplica` נכשלת ב־`context deadline exceeded` אחרי כמה בדיקות.** השהיה נשארה
  במשני מהרצה קודמת שנקטעה (תהליך שנהרג באמצע, או בדיקה שנעצרה ב־timeout). הריצי את פקודת האיפוס שבסוף צעד 4.
- **‏`ReadNote` מחזירה תמיד `primary`.** ההשוואה נעשית כמחרוזות ב־Go, או שה־LSN נלקח לפני ה־`INSERT`. ההשוואה
  שייכת לשרת, ב־`CaughtUp`, והאסימון נלקח אחרי שהכתיבה חזרה.
- **כל הבדיקות של `replica` מסומנות `SKIP`.** אחד משני המשתנים, `DATABASE_URL` או `REPLICA_URL`, לא מוגדר בחלון
  הזה. מריצים שוב את שורות ה־`export` מצעד 1.
- **‏`TestRingMovesFewKeysWhenANodeJoins` מראה שכמעט כל המפתחות זזו.** ‏`Ring.Node` מחפשת ב־hash של המפתח במקום
  בנקודות, או מחזירה את הנקודה שלפני ה־hash. הנקודה הנכונה היא הראשונה שגדולה או שווה לו.

## רמזים

- ב־`Probe`, ‏`select` עם `ctx.Done()` ועם `time.After(interval)` נותן גם המתנה וגם יציאה מסודרת. עטפי את
  השגיאה של ה־context כך שיופיע בה מזהה השורה.
- אחוזון בדרגה הקרובה: האינדקס הוא `ceil(q*n) - 1` במערך הממוין. בדקי את המקרה של איבר אחד.
- ב־`ReadNote`, קראי ל־`cancel` של ההמתנה מיד אחריה, ולא ב־`defer` בסוף הפונקציה. ואם ה־`ctx` של הקוראת עצמה
  בוטל, החזירי את השגיאה שלו במקום לעבור לראשי.
- ב־`Zipf`, צרי את המחולל מחוץ לפונקציה שמוחזרת. מחולל חדש בכל קריאה הוא איטי, וגם מחזיר תמיד את אותם מספרים.
- ‏`sort.Search(n, f)` מחזירה את ה־i הקטן ביותר שעבורו `f(i)` אמת, ו־n אם אין כזה.

## הפתרון

‏`solution/`, בעיקר `replica/replica.go`, ‏`replica/ryw.go`, ‏`replica/router.go` ו־`skew/`. **למה זה עובד:** הבדיקה
מודדת מה שמשתמשת חווה, ולא מה שהשרת מדווח: הזמן מסוף הכתיבה ועד שהשורה נראית בקריאה, כגבול עליון. ההשהיה
במשני הופכת את הפיגור לגדול ודטרמיניסטי, ולכן הבדיקות של „לקרוא את מה שכתבת” יכולות לצפות מראש לכישלון של
הקריאה הנאיבית. אסימון ה־LSN עובד כי מיקום ה־WAL שנלקח אחרי ה־`COMMIT` נמצא אחרי רשומת ה־commit, וההשוואה
נעשית בטיפוס `pg_lsn` בשרת. הסימולטור דטרמיניסטי בזכות זרע קבוע, ולכן הבדיקות שלו יכולות לחסום את מקדם החום.

## מה התרגיל מדגים

- שכפול זורם אמיתי של PostgreSQL 18 בשני קונטיינרים, עם `pg_basebackup`, ‏slot ו־`standby.signal`.
- פיגור שכפול שנמדד מהצד של הלקוח, ונעשה גלוי בעומס או בהשהיה מכוונת.
- שני מימושים של הבטחת סשן: חלון זמן עם שעון מוזרק, ואסימון LSN עם נפילה לראשי.
- מחיצה חמה שנוצרת מהתפלגות Zipf, ולמה מפתח מורכב מפזר אותה במחיר הסדר.
- גיבוב עקבי שמזיז חמישית מהמפתחות כשמוסיפים שרת חמישי, מול ארבע חמישיות במודולו.

</div>
