---
direction: rtl
lang: he
---

<div dir="rtl" lang="he">

# תרגיל פרק 11: `isolation-lab`, אנומליות, נעילות, תור, קיפאון ומיגרציות

**מטרה:** לראות בעיניים עדכון אבוד ו־write skew בבדיקות Go עם שני חיבורים, לתקן כל אחד בכלי המתאים, לבנות
תור עבודות עם `SKIP LOCKED`, לגרום לקיפאון ולמנוע אותו, ולנהל את הסכמה במיגרציות goose. הקוד של
הניסיון החוזר, של התור ושל החיבור למסד עובר אחר כך לפרויקט הגמר, ולכן כדאי לכתוב אותו נקי.

**זמן משוער:** שלושה ימי לימוד. יום 26: כ־60 דקות בנייה (האנומליות). יום 27: כ־45 דקות (נעילות, תור, קיפאון).
יום 28: כ־65 דקות (goose, ה־pool והניסיון החוזר).

**ידע מקדים:** פרק 11; פרק 10 (המסד בקונטיינר ו־pgx, סעיפים 10.1 ו־10.7); עטיפת שגיאות ו־generics (פרק 4,
סעיפים 4.5 ו־4.7); context, ‏`sync.WaitGroup` ו־`-race` (פרק 5, סעיפים 5.4, 5.5 ו־5.8). התרגיל לא מייבא קוד
מפרקים קודמים. הספריות הן pgx ו־goose, בגרסאות שמקובעות ב־`go.mod` (נכון לאוקטובר 2026, ראי נספח ג).

## מה בתיקייה הזאת

| קובץ | תפקיד |
| --- | --- |
| `data/compose.yaml` | שרת PostgreSQL 18 בפרויקט `ch11`, בפורט 5411, מסד `lab` |
| `starter/migrations/00001_accounts.sql` עד `00003_jobs.sql` | שלוש מיגרציות שלמות; את 00004 את כותבת ביום 28 |
| `starter/migrations/migrations.go` | ‏`embed.FS` של קובצי ה־SQL |
| `starter/db/db.go` | ‏`Config` עם גבולות pool, ופונקציות החיבור, המיגרציות והגרסה (שלם) |
| `starter/db/retry.go` | פונקציות העזר שלמות; ‏`RetrySerializable` עם `TODO` |
| `starter/bank/bank.go` | חשבונות: הקריאה, הכתיבה, הנעילה וההעברה המסודרת עם `TODO`; השאר שלם |
| `starter/oncall/oncall.go` | תורנויות: שלוש פונקציות עם `TODO` |
| `starter/queue/queue.go` | תור: `Enqueue` שלם; ‏`Work` עם `TODO` |
| `starter/internal/testdb/testdb.go` | מסד זמני לכל חבילת בדיקות, ו־`WaitForLockWaiters` (שלם) |
| `tests/<package>/*_test.go` | הבדיקות; עותק זהה בכל חבילה ב־`starter/` וב־`solution/` |
| `solution/` | הפתרון המלא, כולל `migrations/00004_accounts_balance_nonnegative.sql` |

כל חבילת בדיקות יוצרת לעצמה מסד חדש בשרת (`ch11_test_...`), מריצה עליו את המיגרציות, ומוחקת אותו בסוף.
המסד `lab` עצמו נשאר נקי, והוא לניסויים שלך עם psql ועם goose.

## יום 26: עדכון אבוד ו־write skew

**1. העתקה, goose, והשרת.**

<div dir="ltr" align="left">

```bash
mkdir -p ~/senior-course/ch11
cp -r $COURSE/exercises/ch11-isolation-lab/starter/. ~/senior-course/ch11/
cp -r $COURSE/exercises/ch11-isolation-lab/data ~/senior-course/ch11/data
cd ~/senior-course/ch11
go install github.com/pressly/goose/v3/cmd/goose@v3.28.0
goose --version
docker compose -p ch11 -f data/compose.yaml up -d --wait
export DATABASE_URL=postgres://course:change-me@localhost:5411/lab
```

</div>

הפקודה `goose --version` צריכה להדפיס את הגרסה. אם היא אומרת `command not found`, תיקיית ה־GOBIN לא נמצאת
בנתיב החיפוש (פרק 2, סעיף 2.7).

**2. קראי לפני שאת כותבת.** פתחי את `bank/bank.go` ואת קובץ הבדיקות שלידו, `bank/bank_test.go`. שימי לב
ש־`Withdraw` הנאיבית כבר כתובה: היא הבאג. הבדיקה `TestLostUpdateUnderReadCommitted` מריצה שתי טרנזקציות שלב
אחרי שלב, ומצפה ליתרה 70.

**3. מלאי ארבע פונקציות קטנות:**

| קובץ | פונקציה | מה |
| --- | --- | --- |
| `bank/bank.go` | `ReadBalance` | ‏`SELECT balance` רגיל, בלי נעילה |
| `bank/bank.go` | `SetBalance` | ‏`UPDATE` שכותב ערך שחושב ב־Go |
| `oncall/oncall.go` | `CountOnCall` | כמה רופאים תורנים במשמרת |
| `oncall/oncall.go` | `GoOffCall` | ספירה; פחות משניים זה `ErrLastOnCall`; אחרת עדכון של השורה של הרופא בלבד |

ב־`oncall/oncall.go` חסר הייבוא של `fmt`, כי אף פונקציה שלמה בקובץ לא משתמשת בו עדיין; הוסיפי אותו (או תני לעורך להוסיף בשמירה).

**4. הריצי את בדיקות האנומליות.** הן עוברות כשהאנומליה קורית:

<div dir="ltr" align="left">

```bash
go test -race -v -run 'LostUpdate|RepeatableRead|WriteSkew|SerializableCatches' ./bank ./oncall
```

</div>

**5. בפתק:** טבלת ארבעת השלבים של העדכון האבוד (סעיף 11.3.1), ואותה טבלה ל־write skew. ליד כל אחת: איזו רמה
עוצרת אותה, ובאיזו שגיאה. העתיקי את ההודעה ש־`TestSerializableCatchesWriteSkew` מדפיסה בשורת `PostgreSQL said`.

## יום 27: `FOR UPDATE`, ‏`SKIP LOCKED` וקיפאון

**6. מלאי:**

| קובץ | פונקציה | מה |
| --- | --- | --- |
| `bank/bank.go` | `ReadBalanceForUpdate` | כמו `ReadBalance`, עם `FOR UPDATE` |
| `bank/bank.go` | `WithdrawLocked` | ‏`Withdraw`, עם הקריאה הנועלת |
| `bank/bank.go` | `Transfer` | נועלת את המזהה הנמוך ואז את הגבוה, ורק אז מחייבת ומזכה |
| `queue/queue.go` | `Work` | טרנזקציה אחת: העבודה הוותיקה הפנויה עם `FOR UPDATE SKIP LOCKED`, המטפל, סימון `done` |

ב־`queue/queue.go` חסר הייבוא של `errors`; הוסיפי אותו כשתבדקי את `pgx.ErrNoRows`.

**7. הבדיקות של היום:**

<div dir="ltr" align="left">

```bash
go test -race -v -run 'ForUpdate|Concurrent|Deadlock|Ordered' ./bank
go test -race -v ./queue
```

</div>

**8. דוח הקיפאון.** הבדיקה `TestOppositeTransfersDeadlock` גורמת לקיפאון בכוונה. השרת כתב עליו דוח ללוג:

<div dir="ltr" align="left">

```bash
docker compose -p ch11 -f data/compose.yaml logs db | grep -A6 "deadlock detected" | tail -7
```

</div>

העתיקי את הדוח לפתק, וסמני בו: מי חיכה למי, ואיזו פקודה כל אחד הריץ. הוסיפי את התוצאה של הפקודה הזאת:

<div dir="ltr" align="left">

```bash
docker compose -p ch11 -f data/compose.yaml exec db psql -U course -d lab -c "SHOW deadlock_timeout"
```

</div>

**9. ניסוי ב־psql (רשות, עשר דקות).** פתחי שני חלונות טרמינל, ובכל אחד psql מול `lab`. הריצי ב־`lab` את
המיגרציות (`goose up`, ראי צעד 10), הכניסי חשבון, ונסי להריץ ידנית את העדכון האבוד: `BEGIN;` בשניהם, קריאה,
כתיבה, `COMMIT;`. ואז אותו דבר עם `FOR UPDATE`, ותראי את החלון השני נתקע עד שהראשון מאשר.

## יום 28: goose, ה־pool והניסיון החוזר

**10. goose משורת הפקודה, מול `lab`:**

<div dir="ltr" align="left">

```bash
export GOOSE_DRIVER=postgres GOOSE_DBSTRING=$DATABASE_URL GOOSE_MIGRATION_DIR=migrations
goose status
goose up
goose status
goose down
goose status
```

</div>

**11. מיגרציה חדשה.** צרי אותה עם מספר סידורי, וכתבי בה `CHECK` שמונע יתרה שלילית, עם `Down` שמסיר אותו:

<div dir="ltr" align="left">

```bash
goose -s create accounts_balance_nonnegative sql
goose up
goose down
goose up
```

</div>

הקובץ שנוצר הוא `migrations/00004_accounts_balance_nonnegative.sql`. שם האילוץ שלך יכול להיות כל שם; הבדיקה
`TestBalanceCannotGoNegative` בודקת רק שהמסד מחזיר `23514`.

**12. הניסיון החוזר.** מלאי שתי פונקציות:

- הפונקציה `RetrySerializable` בקובץ `db/retry.go`. פונקציות ההמתנה כבר שם; את השורה שמתחילה ב־`_, _ =`
  מוחקים כשמשתמשים בהן, ואת הייבוא של `fmt` מוסיפים.
- הפונקציה `GoOffCallSerializable` בקובץ `oncall/oncall.go`. הוסיפי את הייבוא של החבילה הזאת:

<div dir="ltr" align="left">

```text
example.com/ch11-isolation-lab/db
```

</div>

**13. ה־pool.** קראי את `Config` בקובץ `db/db.go`, ואת הבדיקה `TestPoolWaitsAtMaxConns` שלידו. בדקי את הגבול
של השרת:

<div dir="ltr" align="left">

```bash
docker compose -p ch11 -f data/compose.yaml exec db psql -U course -d lab -c "SHOW max_connections"
```

</div>

רשמי בפתק את הערך, וכמה עותקים של שירות עם `MaxConns = 10` היו נכנסים בו עם שוליים של עשרה חיבורים: ___.

**14. כל הבדיקות, ואז התחייבות וכיבוי:**

<div dir="ltr" align="left">

```bash
go test -race ./...
gofmt -l . ; go vet ./... ; staticcheck ./...
cd ~/senior-course
git add ch11
git commit -m "Chapter 11: isolation anomalies, locks, SKIP LOCKED queue, goose migrations"
docker compose -p ch11 -f ch11/data/compose.yaml down -v
```

</div>

## פלט צפוי

הבדיקות המלאות, בסוף יום 28 (הזמנים יהיו שונים אצלך):

<div dir="ltr" align="left">

```text
ok  	example.com/ch11-isolation-lab/bank	…s
ok  	example.com/ch11-isolation-lab/db	…s
?   	example.com/ch11-isolation-lab/internal/testdb	[no test files]
ok  	example.com/ch11-isolation-lab/migrations	…s
ok  	example.com/ch11-isolation-lab/oncall	…s
ok  	example.com/ch11-isolation-lab/queue	…s
```

</div>

בדיקת הקיפאון לוקחת כשנייה, כי זה הערך של `deadlock_timeout`. עם הדגל `-v` תראי שתי שורות שמתחילות
ב־`PostgreSQL said`, ובהן ההודעות האלה:

<div dir="ltr" align="left">

```text
deadlock detected (SQLSTATE 40P01)
could not serialize access due to read/write dependencies among transactions (SQLSTATE 40001)
```

</div>

הפלט של `goose up` על מסד `lab` ריק, אחרי שיצרת את 00004, נראה כך. הכלי מתחיל כל שורה בתאריך ובשעה, וכאן
הם הוחלפו בשלוש נקודות:

<div dir="ltr" align="left">

```text
… OK   00001_accounts.sql (…)
… OK   00002_oncall.sql (…)
… OK   00003_jobs.sql (…)
… OK   00004_accounts_balance_nonnegative.sql (…)
… goose: successfully migrated database to version: 4
```

</div>

## רשימת בדיקה

- [ ] בדיקות האנומליות עוברות בכל הרצה (נסי `-count=5`), והטבלאות של שתי האנומליות בפתק.
- [ ] ‏`TestForUpdateMakesTheSecondReaderWait` עוברת: ההמתנה מוכחת דרך `pg_stat_activity`, לא דרך שינה.
- [ ] ‏`TestWorkersNeverShareAJob` עוברת עם `-race`.
- [ ] דוח הקיפאון וערך `deadlock_timeout` בפתק.
- [ ] המיגרציה 00004 עולה, יורדת ועולה שוב עם goose, ו־`TestMigrationsRoundTrip` עוברת.
- [ ] ‏`RetrySerializable` לא מנסה שוב שגיאה רגילה, נעצרת אחרי `MaxAttempts`, ומכבדת context שבוטל.
- [ ] כל הבדיקות עם `-race` ירוקות בלי דילוגים.
- [ ] שלושת הכלים של צעד 14 (‏gofmt, ‏go vet ו־staticcheck) לא מדפיסים דבר.

## טעויות נפוצות ותיקונים

- **כל הבדיקות מסומנות `SKIP`.** המשתנה `DATABASE_URL` לא מוגדר בחלון הנוכחי. מריצים את שורת ה־`export`
  מצעד 1 שוב, ובודקים עם `go test -v` שאין `SKIP`.
- **‏`TestLostUpdateUnderReadCommitted` נתקעת ולא מסתיימת.** ‏`ReadBalance` נכתבה עם `FOR UPDATE`, ולכן הקריאה
  של הטרנזקציה השנייה מחכה לראשונה, שרצה באותה goroutine ולא תתקדם. ביום 26 הקריאה רגילה; עוצרים עם Ctrl+C ומתקנים.
- **‏`TestForUpdateMakesTheSecondReaderWait` נכשלת ב־`no session started waiting for a lock`.** ‏`ReadBalanceForUpdate`
  לא נועלת, ולכן הקורא השני לא מחכה. מוסיפים `FOR UPDATE` בסוף השאילתה.
- **‏`TestWorkersNeverShareAJob` מוצאת עבודה שטופלה פעמיים.** הבחירה ב־`Work` רצה מחוץ לטרנזקציה של העדכון,
  או בלי נעילה. הבחירה, המטפל והסימון חייבים להיות באותה טרנזקציה.
- **‏`goose up` אומר `no migrations to run` אחרי שערכת את 00003.** goose כבר רשם את גרסה 3 כמוחלת ולא יריץ
  אותה שוב. מבטלים את העריכה וכותבים מיגרציה חדשה.
- **אחרי בדיקה שקרסה נשארים מסדים בשם `ch11_test_...`.** ‏panic בבדיקה עוצר את התהליך לפני הניקוי. זה לא
  מזיק לבדיקות הבאות, ו־`down -v` מוחק הכול.

## רמזים

- הפונקציה `pgx.BeginFunc` מאשרת כשהפונקציה שלך מחזירה `nil`, ומבטלת אחרת.
- הפונקציה `Withdraw` הנאיבית היא תבנית טובה ל־`WithdrawLocked`.
- ‏`min` ו־`max` מובנות ב־Go, ועובדות על `int64`.
- ב־`Work`, משתנה `claimed` מחוץ לסגירה, שמקבל `true` רק אחרי הסימון, פותר את השאלה מה להחזיר.
- ב־`RetrySerializable`, הפונקציה `pgx.BeginTxFunc` מקבלת את רמת הבידוד באפשרויות הטרנזקציה.
- את הערך שהפונקציה של הקוראת מחזירה שומרים במשתנה מחוץ לסגירה.
- סדר הבדיקות בלולאה: הצלחה, שגיאה שאינה ניתנת לניסיון, הניסיון האחרון, ואז המתנה.
- כשהשגיאה האחרונה נעטפת, השתמשי ב־`%w`, כדי ש־`IsRetryable` עדיין יזהה אותה.

## הפתרון

התיקייה `solution/`. הקבצים העיקריים:

- ‏`db/retry.go`
- ‏`bank/bank.go`
- ‏`queue/queue.go`
- ‏`migrations/00004_accounts_balance_nonnegative.sql`

**למה זה עובד:** הבדיקות שולטות בסדר הפקודות של שתי הטרנזקציות, ולכן האנומליה היא תוצאה דטרמיניסטית של רמת
הבידוד ולא מקרה של תזמון. התיקונים פועלים בתוך המסד (נעילת שורה, `UPDATE` עם תנאי, `SKIP LOCKED`, סדר נעילות),
ולכן הם נכונים בלי קשר למספר העותקים של השירות. ‏`RetrySerializable` מקבלת את כל הטרנזקציה כפונקציה ומריצה אותה
מההתחלה, כך שבניסיון השני ההחלטה נשענת על תמונת מצב חדשה, וזה בדיוק מה שהבדיקה עם ההתנגשות המכוונת מוכיחה.

## מה התרגיל מדגים

- אנומליות מקביליות כבדיקות דטרמיניסטיות: שני חיבורים, סדר מפורש, והוכחת המתנה מתוך השרת.
- ההתאמה בין בעיה לכלי: עדכון אטומי או `FOR UPDATE` לשורה אחת, Serializable וניסיון חוזר לכלל על כמה שורות,
  ‏`SKIP LOCKED` לתור, וסדר נעילות לקיפאון.
- מיגרציות כקוד: מוטמעות בבינארי, מורצות מ־Go תחת advisory lock, ונבדקות בסבב מלא על מסד ריק.
- שלושה רכיבים שפרויקט הגמר לוקח כמו שהם: `RetrySerializable`, תבנית ה־`SKIP LOCKED`, והמיגרציות עם goose.

</div>
