---
direction: rtl
lang: he
---

<div dir="rtl" lang="he">

# תרגיל פרק 14: `jobqueue`

**מטרה:** לבנות תור עבודות עמיד על PostgreSQL. לקיחה עם `FOR UPDATE SKIP LOCKED` וחכירה, פונקציית המתנה
גדלה עם גבול ופיזור מלא ובדיקות לגבולות שלה, handler אידמפוטנטי, תור מתים עם הפעלה חוזרת, מתזמן שלא יוצר
כפילויות, ו־workers שנכבים על `SIGTERM` בלי לאבד עבודה.

**זמן משוער:** שני ימי לימוד. יום 41: כ־65 דקות (המתנה, לקיחה, ניסיון חוזר, handler אידמפוטנטי). יום 42:
כ־60 דקות (תור מתים, הפעלה חוזרת, מתזמן, כיבוי מסודר).

**ידע מקדים:** פרק 14. מפרקים קודמים: goroutines וביטול (פרק 5), PostgreSQL בקונטיינר ו־pgx (פרק 10),
טרנזקציות ו־`SKIP LOCKED` (פרק 11, סעיפים 11.1 ו־11.4), וטבלת סינון (פרק 13, סעיף 13.7). התרגיל לא מייבא
קוד מתרגיל קודם; התור כאן נכתב מאפס, עם חכירה במקום נעילה ארוכה.

## מה בתיקייה הזאת

| קובץ | תפקיד |
| --- | --- |
| `data/compose.yaml` | PostgreSQL עם שם הפרויקט `ch14`, בפורט 5414 |
| `data/cronjob-preview.yaml` | CronJob של Kubernetes לדוגמה; לקריאה בלבד, ירוץ בפרק 20 |
| `starter/backoff/backoff.go` | `Policy`, `Ceiling` ו־`Delay`; בשתי האחרונות יש `TODO` של יום 41 |
| `starter/queue/schema.sql` | הטבלאות `jobs` ו־`job_attempts`; מוכן |
| `starter/queue/queue.go` | התור. ב־`Claim` (השאילתה) וב־`Fail` יש `TODO` של יום 41; ב־`Fail` וב־`Replay` של יום 42 |
| `starter/queue/permanent.go` | `Permanent` ו־`IsPermanent`; מוכן |
| `starter/notify/notify.go` | handler אידמפוטנטי; חסרה בו ההגנה הראשונה, `TODO` של יום 41 |
| `starter/worker/worker.go` | מאגר ה־workers; הכיבוי המסודר הוא `TODO` של יום 42 |
| `starter/scheduler/scheduler.go` | מתזמן עם מפתח לכל חלון זמן; מוכן, לקריאה ביום 42 |
| `starter/cmd/worker/main.go` | תוכנית הדגמה לכיבוי מסודר; מוכנה |
| `starter/testenv/testenv.go` | עזר לבדיקות: סכמה פרטית לכל בדיקה |
| `tests/backoff/backoff_test.go` | תקרות, קצוות הטווח, ואלפי הגרלות עם זרע קבוע (בלי מסד) |
| `tests/queue/queue_test.go` | לקיחה, `run_at`, מפתח ייחודי, לקיחה מקבילית, ניסיון חוזר, מוות, הפעלה חוזרת, חכירה וגידור |
| `tests/notify/notify_test.go` | הרצה כפולה, ספק שנכשל, קריסה בין שליחה לרישום |
| `tests/worker/worker_test.go` | ניסיונות עד הצלחה, עד מוות, panic, וכיבוי שמחכה לעבודה |
| `tests/scheduler/scheduler_test.go` | חלונות זמן, ושלושה מתזמנים במקביל |
| `solution/` | הפתרון המלא, עם עותקים זהים של הבדיקות |

הבדיקות של `backoff` לא צריכות מסד. כל השאר קוראות את `DATABASE_URL`, מדלגות כשהוא לא מוגדר, ויוצרות סכמה
פרטית שנמחקת בסוף הבדיקה.

## יום 41: המתנה, לקיחה, ניסיון חוזר ואידמפוטנטיות

**1. העתיקי את קוד ההתחלה ואת קבצי הנתונים, והפעילי את המסד.**

<div dir="ltr" align="left">

```bash
mkdir -p ~/senior-course/ch14
cd ~/senior-course/ch14
cp -r "$COURSE"/exercises/ch14-jobqueue/starter/. .
cp -r "$COURSE"/exercises/ch14-jobqueue/data .
printf 'bin/\n' > .gitignore
go build ./... && go vet ./... && echo builds
docker compose -p ch14 -f data/compose.yaml up -d --wait
export DATABASE_URL=postgres://course:change-me@localhost:5414/jobs
```

</div>

**2. המתנה גדלה.** מלאי את `Ceiling` ואת `Delay`, בקובץ `backoff/backoff.go`, לפי ההערות. הבדיקות לא צריכות
מסד:

<div dir="ltr" align="left">

```bash
go test -v ./backoff
```

</div>

שימי לב למקרה `{64, 2 * time.Second}` בבדיקה: הזזה בביטים של 100 אלפיות שנייה ב־63 גולשת, ולכן הלולאה
מכפילה צעד אחר צעד ועוצרת בגבול.

**3. לקיחה.** בקובץ `queue/queue.go`, החליפי את `WHERE false` בפונקציה `Claim` בשאילתה האמיתית: עבודה ממתינה
שהגיע זמנה, או עבודה רצה שהחכירה שלה פגה; הישנה ביותר קודם; `FOR UPDATE SKIP LOCKED`. ואז:

<div dir="ltr" align="left">

```bash
go test -race -run 'TestEnqueueClaimComplete|TestClaimRespectsRunAt|TestUniqueKey|TestConcurrentClaims|TestExpiredLease' -v ./queue
```

</div>

**4. ניסיון חוזר.** השלימי את הענף של ניסיון חוזר בפונקציה `Fail`. הענף של המוות נשאר ליום 42.

<div dir="ltr" align="left">

```bash
go test -race -run 'TestFailReschedulesLater' -v ./queue
go test -race -run 'TestRetriesUntilSuccess' -v ./worker
```

</div>

**5. handler אידמפוטנטי.** הוסיפי את ההגנה הראשונה בקובץ `notify/notify.go`: אם כבר יש רשומה למפתח, לא
שולחים.

<div dir="ltr" align="left">

```bash
go test -race -v ./notify
```

</div>

בבדיקה `TestCrashAfterSendBeforeRecordIsSafe` שימי לב מי מגן במקרה הזה: לא הרשומה שלך, כי היא לא נכתבה,
אלא הסינון של הספק לפי המפתח.

## יום 42: תור מתים, הפעלה חוזרת, מתזמן וכיבוי

**6. מוות.** השלימי את הענף השני בפונקציה `Fail`: כישלון קבוע, או ניסיון אחרון, שולחים לתור המתים.

**7. הפעלה חוזרת.** מלאי את `Replay`: עדכון אחד, מותנה במצב מת, שמחזיר `ErrNotDead` כשאין מה לעדכן.

<div dir="ltr" align="left">

```bash
go test -race -v ./queue
```

</div>

**8. כיבוי מסודר.** בקובץ `worker/worker.go`, שני ה־contexts בפונקציה `runOne` נבנים עדיין מ־`ctx` עצמו,
ולכן `SIGTERM` מבטל את העבודה ואת הרישום שלה. נתקי את שניהם עם `context.WithoutCancel`:

<div dir="ltr" align="left">

```bash
go test -race -v ./worker
```

</div>

**9. המתזמן.** קראי את `scheduler/scheduler.go` ואת `data/cronjob-preview.yaml`, והריצי את הבדיקות שלו:

<div dir="ltr" align="left">

```bash
go test -race -v ./scheduler
```

</div>

**10. ראי את הכיבוי בעיניים.** הדגל `-init` יוצר את הטבלאות בסכמה הראשית; תני אותו רק בפעם הראשונה.
התוכנית מוסיפה שלוש עבודות של שלוש שניות ומריצה שני workers:

<div dir="ltr" align="left">

```bash
go run ./cmd/worker -init -enqueue 3 -ms 3000
```

</div>

אחרי שתראי שתי שורות `enqueued` ושנייה בערך, לחצי Ctrl-C. התוכנית מחכה לשתי העבודות שרצות, ורק אז יוצאת.
הריצי אותה שוב בלי `-init`, הפעם עם עבודה שנכשלת פעמיים, ועצרי אחרי כמה שניות:

<div dir="ltr" align="left">

```bash
go run ./cmd/worker -enqueue 1 -ms 200 -fail 2
```

</div>

**11. ההיסטוריה.** הסתכלי על המצב ועל הניסיונות במסד:

<div dir="ltr" align="left">

```bash
docker compose -p ch14 -f data/compose.yaml exec db psql -U course -d jobs \
  -c "SELECT id, state, attempts FROM jobs ORDER BY id" \
  -c "SELECT job_id, attempt, worker, error FROM job_attempts ORDER BY id"
```

</div>

**12. הכול, הכלים, כיבוי ושמירה.**

<div dir="ltr" align="left">

```bash
go test -race -count=1 ./...
gofmt -l .
go vet ./...
staticcheck ./...
docker compose -p ch14 -f data/compose.yaml down -v
cd ~/senior-course
git add ch14
git commit -m "ch14: Postgres job queue with leases, backoff, dead letters and drain"
```

</div>

## פלט צפוי

בדיקות `backoff` ביום 41:

<div dir="ltr" align="left">

```text
--- PASS: TestCeiling (0.00s)
--- PASS: TestCeilingEdgePolicies (0.00s)
--- PASS: TestDelayEndsOfTheRange (0.00s)
--- PASS: TestDelayStaysInBoundsAndSpreads (0.00s)
--- PASS: TestZeroPolicyNeverCallsRand (0.00s)
PASS
```

</div>

הכיבוי בצעד 10, בלי חותמת הזמן שבתחילת כל שורה. אצלך אולי ירוצו עבודות אחרות מתוך השלוש:

<div dir="ltr" align="left">

```text
level=INFO msg=enqueued job=1
level=INFO msg=enqueued job=2
level=INFO msg=enqueued job=3
level=INFO msg="shutdown requested: no new claims, waiting for running jobs"
level=INFO msg="job done" worker=w-0 job=2 kind=demo attempt=1
level=INFO msg="job done" worker=w-1 job=1 kind=demo attempt=1
level=INFO msg="drained, exiting"
```

</div>

וההיסטוריה בצעד 11, אחרי שתי ההרצות:

<div dir="ltr" align="left">

```text
 id |  state  | attempts 
----+---------+----------
  1 | done    |        1
  2 | done    |        1
  3 | done    |        1
  4 | done    |        3
(4 rows)

 job_id | attempt | worker |       error       
--------+---------+--------+-------------------
      1 |       1 | w-1    | 
      2 |       1 | w-0    | 
      4 |       1 | w-0    | simulated failure
      4 |       2 | w-0    | simulated failure
      4 |       3 | w-0    | 
      3 |       1 | w-1    | 
(6 rows)
```

</div>

והכול, ביום 42:

<div dir="ltr" align="left">

```text
ok  	example.com/ch14-jobqueue/backoff	1.317s
?   	example.com/ch14-jobqueue/cmd/worker	[no test files]
ok  	example.com/ch14-jobqueue/notify	1.787s
ok  	example.com/ch14-jobqueue/queue	3.313s
ok  	example.com/ch14-jobqueue/scheduler	3.383s
?   	example.com/ch14-jobqueue/testenv	[no test files]
ok  	example.com/ch14-jobqueue/worker	2.825s
```

</div>

## רשימת בדיקה

- [ ] בדיקות ההמתנה עוברות, כולל ניסיון 64 וגבול בגודל המקסימלי.
- [ ] שמונה goroutines לוקחות 60 עבודות, וכל עבודה נלקחת פעם אחת.
- [ ] עבודה שנכשלה חוזרת עם `run_at` מאוחר יותר, ואחרי הניסיון האחרון היא מתה.
- [ ] worker שהחכירה שלו פגה מקבל `ErrLeaseLost`, וההיסטוריה מראה את שני הניסיונות.
- [ ] הפעלה חוזרת עובדת פעם אחת, והשנייה מחזירה `ErrNotDead`.
- [ ] Ctrl-C באמצע עבודה מחכה לה, והיא רשומה `done`.
- [ ] `go test -race ./...` עובר, שלושת הכלים שקטים, והמסד כבוי עם `down -v`.

## טעויות נפוצות ותיקונים

- **כל בדיקות התור מדלגות עם `set DATABASE_URL`.** המשתנה לא מוגדר במסוף הזה. הגדירי אותו כמו בצעד 1,
  ובדקי שהמסד למעלה עם `docker compose -p ch14 -f data/compose.yaml ps`.
- **`TestCeiling` נכשלת רק במקרים של 64 ו־1073741824, עם תקרה של 0 או שלילית.** החישוב הוא הזזה בביטים או
  חזקה, והוא גלש. הכפילי בלולאה, ועצרי ברגע שהערך מגיע לגבול.
- **`TestConcurrentClaimsNeverShareAJob` נכשלת עם `claimed by a and b`.** חסר `SKIP LOCKED`, או שהשאילתה
  נועלת בלי `FOR UPDATE`. בלי נעילה, שני workers רואים את אותה שורה ממתינה.
- **`TestGracefulDrainWaitsForTheClaimedJob` נכשלת עם `drained job state = running`.** ה־context של הרישום
  עדיין נגזר מ־`ctx` המבוטל, ולכן `Complete` נכשל. נתקי גם אותו, לא רק את ה־handler.
- **`TestRunningTheSameJobTwiceSendsOnce` נכשלת עם `provider called 2 times`.** ההגנה הראשונה חסרה, ולכן
  כל הרצה שולחת. בדקי את הטבלה `notifications` לפני `Send`.
- **`go run ./cmd/worker -init` נכשל עם `relation "jobs" already exists`.** הטבלאות כבר נוצרו בהרצה קודמת.
  הריצי בלי `-init`, או התחילי מחדש עם `down -v`.

## רמזים

- ב־`Ceiling`: התחילי מ־`min(p.Base, p.Cap)`, ובכל צעד בדקי `c > p.Cap/2` לפני ההכפלה. כך ההכפלה לעולם לא
  עוברת את הגבול ולא גולשת.
- ב־`Delay`: הגבול העליון של `Int64N` אינו כלול, ולכן מעבירים את התקרה ועוד אחד, חוץ מהמקרה של המספר הגדול
  ביותר.
- ב־`Claim`: התנאי הוא `(state = 'pending' AND run_at <= now()) OR (state = 'running' AND locked_until < now())`.
- ב־`Fail`: הפונקציה `finish` כבר מוסיפה את מזהה העבודה כפרמטר הראשון, ולכן הפרמטרים שלך מתחילים מ־`$2`.
- ב־`Replay`: `tag.RowsAffected() == 0` פירושו שהעבודה לא הייתה מתה. עטפי את `ErrNotDead` עם `%w`.

## הפתרון

הפתרון המלא נמצא בתיקייה `solution/`. הלקיחה נועלת שורה רק לרגע, מסמנת חכירה ומונה ניסיונות, ונסגרת; כך
worker שמת לא משאיר נעילה, והמונה משמש גם כאסימון גידור לכל סיום. ההמתנה היא פונקציה טהורה עם אקראיות
מוזרקת, ולכן הבדיקות מכסות את הקצוות בדיוק ואת הטווח באלפי הגרלות. ה־handler אידמפוטנטי בשתי הגנות, רשומה
משלו ומפתח לספק, והכיבוי מנתק את העבודה ואת הרישום שלה מאות הכיבוי, כך שמה שנלקח מסתיים ונרשם.

## מה התרגיל מדגים

- תור עבודות על PostgreSQL עם `SKIP LOCKED`, `run_at`, מונה ניסיונות וחכירה.
- גידור לפי מונה הלקיחות, וחכירה שפגה שמחזירה עבודה לתור או קוברת אותה.
- המתנה גדלה עם גבול ופיזור מלא, כפונקציה טהורה, עם בדיקות טבלאיות ובדיקה דמוית תכונה.
- הבחנה בין כישלון זמני לכישלון קבוע.
- handler אידמפוטנטי מול מערכת חיצונית: רשומה משלו, ומפתח אידמפוטנטיות לספק.
- תור מתים, היסטוריית ניסיונות, והפעלה חוזרת שמותנית במצב.
- מתזמן שבטוח להריץ בכל עותק, בזכות מפתח ייחודי לכל חלון זמן.
- כיבוי מסודר עם `context.WithoutCancel`: מפסיקים לקחת, מסיימים, רושמים, יוצאים.

</div>
