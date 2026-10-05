---
direction: rtl
lang: he
---

<div dir="rtl" lang="he">

# תרגיל פרק 18: `leader`, בחירת מנהיג, אסימוני גידור וסאגה

**מטרה:** לבנות בחירת מנהיג מעל advisory lock של PostgreSQL ולראות השתלטות אמיתית כשהמנהיג נהרג, להראות
נעילת Redis שמוחזקת בידי שניים, לעצור כתיבה של מנהיג לשעבר באסימון גידור, ולכתוב סאגה מתוזמרת שמפצה בסדר
הפוך. החבילה `leader` עוברת לפרויקט הגמר כמו שהיא (אבן הדרך M6, פרק 38, סעיף 38.6), ולכן כדאי לכתוב אותה נקי.

**זמן משוער:** שני ימי לימוד. יום 52: כ־55 דקות בנייה (advisory lock, ‏Redis, אסימון גידור). יום 53: כ־60 דקות
(הדופק וההשתלטות, הסאגה).

**ידע מקדים:** פרק 18; פרק 11 (נעילות ומאגר החיבורים, סעיפים 11.4 ו־11.7); פרק 12 (Redis ו־go-redis, סעיף 12.1);
goroutines, ‏channels, ‏context ו־`-race` (פרק 5, סעיפים 5.4 עד 5.8); ‏`log/slog` (פרק 6). התרגיל לא מייבא קוד
מפרקים קודמים. הספריות הן pgx ו־go-redis, בגרסאות שמקובעות ב־`go.mod` (נכון לאוקטובר 2026, ראי נספח C).

## מה בתיקייה הזאת

| קובץ | תפקיד |
| --- | --- |
| `data/compose.yaml` | פרויקט `ch18`: PostgreSQL 18 בפורט 5418, מסד `lab`, ו־Redis 8 בפורט 6318 |
| `starter/leader/leader.go` | ‏`Schema`, ‏`EnsureSchema`, ‏`Config`, ‏`New`, ‏`Run`, ‏`HolderPID` שלמים; ‏`campaign` ו־`serve` עם `TODO` |
| `starter/fencing/fencing.go` | ‏`Schema` ו־`Read` שלמים; ‏`Write` עם `TODO` |
| `starter/redislock/redislock.go` | הסקריפטים ו־`Extend` שלמים; ‏`Acquire` ו־`Release` עם `TODO` |
| `starter/saga/saga.go` | ‏`Step` ו־`Error` שלמים; ‏`Run` ו־`compensate` עם `TODO` |
| `starter/internal/testdb/testdb.go` | מסד זמני לכל חבילת בדיקות, ו־`Pool` שפותח מאגר חדש לכל קריאה (שלם) |
| `starter/cmd/elect/main.go` | מועמדת אחת בבחירות, ללוג ולהדגמה עם שני תהליכים (שלמה) |
| `tests/<package>/*_test.go` | הבדיקות; עותק זהה בכל חבילה ב־`starter/` וב־`solution/` |
| `solution/` | הפתרון המלא |

כל חבילה שבודקת מול PostgreSQL יוצרת לעצמה מסד חדש (`ch18_test_...`) ומוחקת אותו בסוף. advisory locks מקומיות
למסד, ולכן חבילות שרצות במקביל לא רואות זו את הנעילות של זו.

## יום 52: advisory lock, ‏Redis ואסימון גידור

**1. העתקה והקמה.**

<div dir="ltr" align="left">

```bash
mkdir -p ~/senior-course/ch18
cp -r $COURSE/exercises/ch18-leader/starter/. ~/senior-course/ch18/
cp -r $COURSE/exercises/ch18-leader/data ~/senior-course/ch18/data
cd ~/senior-course/ch18
docker compose -p ch18 -f data/compose.yaml up -d --wait
export DATABASE_URL=postgres://course:change-me@localhost:5418/lab
export REDIS_ADDR=localhost:6318
```

</div>

**2. נעילה ביד, בשני חלונות.** פתחי שני טרמינלים, ובכל אחד psql:

<div dir="ltr" align="left">

```bash
docker compose -p ch18 -f data/compose.yaml exec db psql -U course -d lab
```

</div>

בראשון הריצי `SELECT pg_try_advisory_lock(42);` ובשני את אותה פקודה. הראשון מקבל `t`, השני `f`. בשני, נסי
`SELECT pg_advisory_unlock(42);` וקראי את האזהרה. עכשיו צאי מהראשון (`\q`), והריצי שוב בשני: `t`. רשמי בפתק
מה שחרר את הנעילה.

**3. Redis.** מלאי בקובץ `redislock/redislock.go`:

| פונקציה | מה |
| --- | --- |
| `Acquire` | 16 בתים מ־`crypto/rand` בקידוד hex, ואז `rdb.SetArgs` עם `Mode: "NX"` ו־`TTL`; ‏`redis.Nil` פירושו `ErrNotAcquired` |
| `Release` | ‏`release.Run` עם המפתח והאסימון; התוצאה `Int()` אפס פירושה `ErrNotHeld` |

ייבואים חסרים: `crypto/rand` ו־`encoding/hex`.

<div dir="ltr" align="left">

```bash
go test -race -v ./redislock
```

</div>

‏`TestExpiredHolderAndNewHolderOverlap` עוברת כשהבעיה קורית: שניים מחזיקים, והראשון יודע רק בדיעבד.

**4. אסימון גידור.** מלאי `Write` בקובץ `fencing/fencing.go`: פקודת `INSERT` אחת, עם `ON CONFLICT` ותנאי `WHERE`
שמשווה את האסימון השמור לחדש (סעיף 18.4). כשלא עודכנה אף שורה, מחזירים `ErrStaleToken`. הייבוא של `fmt` חסר.

<div dir="ltr" align="left">

```bash
go test -race -v ./fencing
```

</div>

**5. המועמדות.** מלאי `campaign` בקובץ `leader/leader.go`, לפי הסדר:

| צעד | מה |
| --- | --- |
| חיבור | ‏`e.pool.Acquire(ctx)`, ואז `Hijack()`, ו־`defer conn.Close(context.Background())` |
| נעילה | ‏`SELECT pg_try_advisory_lock($1)`; שקר: להחזיר `false, nil` |
| תקופה | ‏`INSERT ... ON CONFLICT ... RETURNING term` על הטבלה `leader_terms`, שמגדיל את המונה ב־1 |
| לוג ועבודה | לוגר עם `key`, ‏`term` ו־`backend_pid`, הודעה `leader: elected`, ואז `e.serve(...)` ו־`return true, nil` |

הייבוא של `fmt` חסר. את `serve` תכתבי מחר, ולכן בדיקות החבילה `leader` עוד נופלות היום ב־`TODO(day 53)`.

**6. בפתק:** למה `Hijack` ולא רק `Acquire` ו־`Release`, ומה היה קורה עם `pool.Exec` לנעילה ולשחרור.

## יום 53: הדופק, ההשתלטות והסאגה

**7. הקדנציה.** מלאי `serve` בקובץ `leader/leader.go`, לפי סעיף 18.5.1: context לעבודה, goroutine שמריצה את
`lead` ושולחת את התוצאה ל־channel עם מקום אחד, טיקר של `Heartbeat`, ולולאת `select` על שלושה מקרים. אחרי
הלולאה: `cancel()`, המתנה לעבודה אם עוד לא חזרה, לוג של שגיאה שאינה `context.Canceled`, ושחרור מפורש כשהסשן חי.

<div dir="ltr" align="left">

```bash
go test -race -v ./leader
```

</div>

שימי לב לשורה `takeover completed ... after the kill` ולשורות הלוג `leader: elected` עם `backend_pid`.

**8. השתלטות בין שני תהליכים.** בנו את התוכנית פעם אחת, ופתחי שלושה טרמינלים, בכל אחד אותו `DATABASE_URL`:

<div dir="ltr" align="left">

```bash
go build -o /tmp/elect ./cmd/elect
/tmp/elect -name alpha
```

</div>

בטרמינל השני `/tmp/elect -name beta`, ובשלישי מצאי את התהליך המוביל והרגי אותו:

<div dir="ltr" align="left">

```bash
pgrep -f 'elect -name alpha'
kill -9 <the pid it printed>
```

</div>

אם `beta` היא המנהיגה, החליפי את השם בפקודה. אחרי שהשנייה השתלטה, נסי גם את הדרך השנייה: הפעילי שוב את
`alpha`, ומ־psql סיימי את הסשן של המנהיגה הנוכחית, לפי ה־`backend_pid` שבלוג שלה:

<div dir="ltr" align="left">

```bash
docker compose -p ch18 -f data/compose.yaml exec db psql -U course -d lab -c \
  "SELECT pid, classid, objid, objsubid, granted FROM pg_locks WHERE locktype = 'advisory'"
docker compose -p ch18 -f data/compose.yaml exec db psql -U course -d lab -c "SELECT pg_terminate_backend(<pid>)"
```

</div>

העתיקי לפתק את שורות הלוג של שתי ההשתלטויות, וחשבי מהחותמות כמה זמן עבר בכל אחת. עצרי את התהליכים עם Ctrl+C.

**9. הסאגה.** מלאי `Run` ו־`compensate` בקובץ `saga/saga.go`, לפי סעיף 18.7.3:

<div dir="ltr" align="left">

```bash
go test -race -v ./saga
```

</div>

**10. כל הבדיקות, התחייבות וכיבוי:**

<div dir="ltr" align="left">

```bash
go test -race ./...
gofmt -l . ; go vet ./... ; staticcheck ./...
cd ~/senior-course
git add ch18
git commit -m "Chapter 18: advisory-lock leader election, fencing tokens, Redis lock, saga"
docker compose -p ch18 -f ch18/data/compose.yaml down -v
```

</div>

## פלט צפוי

הבדיקות המלאות בסוף יום 53 (הזמנים יהיו שונים אצלך):

<div dir="ltr" align="left">

```text
?   	example.com/ch18-leader/cmd/elect	[no test files]
ok  	example.com/ch18-leader/fencing	…s
?   	example.com/ch18-leader/internal/testdb	[no test files]
ok  	example.com/ch18-leader/leader	…s
ok  	example.com/ch18-leader/redislock	…s
ok  	example.com/ch18-leader/saga	…s
```

</div>

עם `-v` על החבילה `leader`, בבדיקת ההשתלטות, הסדר בין השורות של A ו־B יכול להשתנות מהרצה להרצה:

<div dir="ltr" align="left">

```text
leader_test.go:…: … msg="leader: elected" who=A key=18001 term=1 backend_pid=…
leader_test.go:…: … msg="leader: elected" who=B key=18001 term=2 backend_pid=…
leader_test.go:…: … msg="leader: heartbeat failed, leadership lost" who=A key=18001 term=1 … err="FATAL: terminating connection due to administrator command (SQLSTATE 57P01)"
leader_test.go:…: terminated the leader's backend, pid …
leader_test.go:…: takeover completed …ms after the kill
```

</div>

והלוג של `beta` אחרי ש־`alpha` נהרגה, מקוצר:

<div dir="ltr" align="left">

```text
level=INFO msg=campaigning name=beta key=1818
level=INFO msg="leader: elected" name=beta key=1818 term=2 backend_pid=…
level=INFO msg="working as leader" name=beta term=2 tick=1
```

</div>

## רשימת בדיקה

- [ ] בניסוי ב־psql, יציאה מהסשן שחררה את הנעילה, ושחרור מסשן אחר החזיר `f` עם אזהרה.
- [ ] ‏`TestExpiredHolderAndNewHolderOverlap` ו־`TestOnlyTheHolderCanRelease` עוברות.
- [ ] ‏`TestStaleTokenIsRejected` ו־`TestPausedLeaderIsFenced` עוברות: הכתיבה עם 33, ואחר כך עם 1, נדחות.
- [ ] שתי השתלטויות, אחרי `kill -9` ואחרי `pg_terminate_backend`, רשומות בפתק עם הזמנים שחישבת.
- [ ] בדיקות הסאגה עוברות לכשל בכל צעד, וגם לפיצוי שנכשל ול־context שבוטל.
- [ ] ‏`go test -race ./...` ירוק בלי `SKIP`, גם עם `-count=3`; `gofmt`, ‏`go vet` ו־`staticcheck` נקיים.

## טעויות נפוצות ותיקונים

- **‏`TestCleanStepDownHandsOverQuickly` נכשלת: B לא נבחרת תוך חמש שניות.** ‏`serve` לא משחררת את הנעילה בירידה
  מסודרת, או ש־`campaign` מחזירה את החיבור למאגר במקום לסגור אותו. החיבור ממשיך לחיות עם הנעילה. ‏`Hijack` ו־`Close`.
- **הבדיקות של `leader` נתקעות עד ה־timeout של `go test`.** הלולאה ב־`serve` לא יוצאת כשהעבודה חוזרת, או שהיא
  ממתינה ל־`done` פעמיים. אחרי שקיבלת מה־channel, סמני את זה (בפתרון: `done = nil`).
- **‏`go test -race` מדווח על מרוץ ב־`serve`.** משתנה משותף בין הלולאה לבין ה־goroutine של העבודה. התוצאה
  עוברת ב־channel בלבד, וה־goroutine לא נוגעת בשום דבר אחר.
- **‏`TestStaleTokenIsRejected` נכשלת בצעד „term 33 again”.** התנאי ב־`WHERE` הוא `<` במקום `<=`, ולכן מנהיג לא יכול
  לכתוב פעמיים באותה תקופה.
- **‏`elect` השנייה הופכת למנהיגה מיד, בזמן שהראשונה חיה.** שתיהן רצות מול מסדים שונים, או עם `-key` שונה. ודאי
  שה־`DATABASE_URL` זהה בשני הטרמינלים.
- **‏`TestCancelledCallerStillGetsCompensation` נכשלת עם `compensation got a cancelled context`.** הפיצויים מקבלים את
  ה־`ctx` המקורי. העבירי להם `context.WithoutCancel(ctx)`.

## רמזים

- ב־`campaign`, אם לקיחת התקופה נכשלה, אין צורך לשחרר ידנית: ה־`defer` שסוגר את החיבור משחרר גם את הנעילה.
- ב־`serve`, ‏`for running := true; running; { select { ... } }` עם `running = false` בכל מקרה של יציאה קריא
  יותר מ־`break` עם תווית.
- ה־context של הדופק נגזר מ־`ctx` עם `context.WithTimeout(ctx, e.cfg.Heartbeat)`. אל תשכחי לבטל אותו בכל סבב.
- ב־`Run` של הסאגה, ‏`steps[:i]` הם בדיוק הצעדים שהצליחו כשהצעד ה־i נכשל.
- ב־`compensate`, לולאה מ־`len(done)-1` עד 0, ודילוג על צעד ש־`Compensate` שלו `nil`.

## הפתרון

‏`solution/`, בעיקר `leader/leader.go`, ‏`fencing/fencing.go`, ‏`redislock/redislock.go` ו־`saga/saga.go`. **למה זה
עובד:** נעילת הסשן קשורה לחיבור אחד שהוצא מהמאגר, ולכן רק המנהיגה מחזיקה בו, והמוות של התהליך או של הסשן משחרר
אותה בלי קוד. הדופק רץ על אותו חיבור, ולכן הוא נכשל בדיוק כשהנעילה אבדה; החלון שבין שני הרגעים מכוסה במספר
התקופה, שהאחסון בודק בתוך פקודת הכתיבה. הסאגה מפצה רק את מה שהצליח, בסדר הפוך, עם context שלא מתבטל, ואוספת
כשלים בפיצוי במקום לעצור עליהם.

## מה התרגיל מדגים

- בחירת מנהיג מעל advisory lock ברמת סשן, עם חיבור ייעודי ודופק על אותו חיבור.
- השתלטות אמיתית אחרי `kill -9` ואחרי `pg_terminate_backend`, בבדיקה ובשני תהליכים.
- החלון שבו שני תהליכים מאמינים שהם מחזיקים, ב־Redis ובמנהיג שנעצר, ואסימון גידור שסוגר אותו בצד האחסון.
- סאגה מתוזמרת עם פיצויים בסדר הפוך, כשל בפיצוי שמדווח, ו־context שבוטל שלא עוצר את הפיצוי.
- החבילה `leader` שהפרויקט מייבא ב־M6, כדי שרק relay אחד יפרסם מה־outbox.

</div>
