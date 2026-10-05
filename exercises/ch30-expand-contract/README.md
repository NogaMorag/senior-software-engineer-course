---
direction: rtl
lang: he
---

<div dir="rtl" lang="he">

# תרגיל פרק 30: `expand-contract`, שינוי שם של עמודה בלי בקשה אחת שנכשלת

**מטרה:** לשנות את שם העמודה `fullname` בטבלת המשתמשים ל־`display_name`, בארבע מיגרציות, בזמן ששני עותקים של שירות
כותבים וקוראים בלי הפסקה, ולהוכיח בבדיקה שאף בקשה לא נכשלה בשום שלב. בדרך: מטריצת תאימות בין גרסאות קוד לגרסאות
סכמה, הדגמה של תור הנעילות ו־`lock_timeout`, ואותו שינוי שם באירוע.

**זמן משוער:** יום לימוד אחד, יום 89: כ־65 דקות בנייה.

**ידע מקדים:** פרק 30; פרק 11, סעיף 11.6 (מיגרציות goose מתוך Go, ומיגרציה בלי טרנזקציה); פרק 10 (PostgreSQL
בקונטיינר ו־pgx); פרק 5, סעיפים 5.4 עד 5.8 (context, מונים אטומיים, המתנה לסיום, ובדיקת מרוצים). החבילות `db`
ו־`internal/testdb` בנויות כמו בתרגיל של פרק 11, אבל לא מיובאות ממנו. גרסאות הספריות מקובעות ב־`go.mod`, ותמונת
המסד בקובץ ה־compose (נכון לאוקטובר 2026, ראי נספח ג).

## מה בתיקייה הזאת

| קובץ | תפקיד |
| --- | --- |
| `data/compose.yaml` | שרת PostgreSQL 18 בפרויקט `ch30`, בפורט 5430, מסד `lab` |
| `starter/migrations/00001_users.sql` | הטבלה בגרסה 1, עם `fullname` (שלם) |
| `starter/migrations/00002_add_display_name.sql` | ההרחבה (שלם, כדוגמה); את 00003 עד 00005 את כותבת |
| `starter/db/db.go` | שני מאגרי חיבורים, אחד לאפליקציה ואחד למיגרציות עם `lock_timeout`, ו־`MigrateTo` (שלם) |
| `starter/app/app.go` | השירות: ה־SQL של כל גרסה, ולולאות הכתיבה והקריאה; ה־SQL של v2 עד v4 עם `TODO` |
| `starter/rollout/rollout.go` | ‏`Plan` עם הצעד הראשון בלבד, ו־`Run` שמבצעת אותה (שלמה) |
| `starter/events/events.go` | האירוע `user.renamed` בשתי גרסאות; ‏`Name` עם `TODO` |
| `starter/internal/testdb/testdb.go` | מסד זמני לכל בדיקה, ו־`WaitForLockWaiters` (שלם) |
| `tests/<package>/*_test.go` | הבדיקות; עותק זהה בכל חבילה ב־`starter/` וב־`solution/` |
| `solution/` | הפתרון המלא |

כל בדיקה יוצרת לעצמה מסד חדש בשרת (`ch30_test_...`) ומוחקת אותו בסוף. המסד `lab` נשאר לניסויים שלך.

## צעדים

**1. העתקה והשרת.**

<div dir="ltr" align="left">

```bash
mkdir -p ~/senior-course/ch30
cp -r $COURSE/exercises/ch30-expand-contract/starter/. ~/senior-course/ch30/
cp -r $COURSE/exercises/ch30-expand-contract/data ~/senior-course/ch30/data
cd ~/senior-course/ch30
docker compose -p ch30 -f data/compose.yaml up -d --wait
export DATABASE_URL=postgres://course:change-me@localhost:5430/lab
go build ./... && go vet ./...
```

</div>

**2. קודם לראות את הבעיה.** הבדיקה הזאת עוברת כבר עכשיו: היא מוכיחה ששינוי שם בצעד אחד שובר קוד שרץ.

<div dir="ltr" align="left">

```bash
go test -race -v -run TestNaiveRenameBreaksRunningCode ./rollout
```

</div>

העתיקי לפתק את השורה עם `one-step rename`: כמה בקשות נכשלו, ומה השגיאה של PostgreSQL.

**3. תור הנעילות.** גם זו עוברת כבר עכשיו. קראי אותה לפני שאת מריצה, ונסי לנחש מה יקרה:

<div dir="ltr" align="left">

```bash
go test -race -v ./db
```

</div>

העתיקי לפתק את השורה `ALTER gave up after`. כמה זמן היא חיכתה, ולמה זה בדיוק `LockTimeout`?

**4. ה־SQL של שלוש הגרסאות.** בקובץ `app/app.go` יש שלוש פונקציות שמחזירות SQL לפי גרסה. מלאי בכל אחת את
הגרסאות החדשות, לפי ההערה על `Phase`. הכתיבה הכפולה היא פקודה אחת שמעדכנת את שתי העמודות.

| פונקציה | הפקודה |
| --- | --- |
| `insertSQL` | הוספת משתמש, עם `RETURNING id` |
| `renameSQL` | עדכון השם לפי `id` |
| `readSQL` | קריאת השם לפי `id` |

**5. שלוש המיגרציות.** צרי אותן בתיקייה `migrations/`, עם המספרים האלה בדיוק:

| קובץ | מה | שימי לב |
| --- | --- | --- |
| `00003_backfill_display_name.sql` | מילוי של `display_name` מ־`fullname`, במנות של 500 לפי `id` | ‏`NO TRANSACTION`, ו־`COMMIT` אחרי כל מנה |
| `00004_swap_not_null.sql` | ‏`display_name` הופכת לחובה, ו־`fullname` מפסיקה להיות חובה | ‏`CHECK ... NOT VALID`, ואז `VALIDATE`, ואז `SET NOT NULL` |
| `00005_drop_fullname.sql` | מחיקת `fullname` | ‏`Down` שמחזיר את העמודה, בלי הנתונים |

**6. מטריצת התאימות:**

<div dir="ltr" align="left">

```bash
go test -race -v -run 'TestCompatibilityMatrix|TestBackfill' ./app
```

</div>

עשרים תת־בדיקות, אחת לכל צירוף של גרסת קוד וגרסת סכמה. אם צירוף שאמור להיכשל עובר, ה־SQL שלך נוגע בעמודה הלא נכונה.

**7. התוכנית.** ב־`rollout/rollout.go` הוסיפי ל־`Plan` את שלושת הצעדים החסרים: לכל מיגרציה, איזו גרסה מותר לפרוס
אחריה. המטריצה מצעד 6 היא התשובה. ואז הבדיקה של הפרק:

<div dir="ltr" align="left">

```bash
go test -race -v -run TestRenameWithZeroFailedRequests ./rollout
```

</div>

**8. האירוע.** מלאי את `Name` ב־`events/events.go`, והריצי:

<div dir="ltr" align="left">

```bash
go test -race -v ./events
```

</div>

**9. ניסוי: לשבור בכוונה (עשר דקות).** החליפי בתוכנית את הסדר של שני הצעדים האמצעיים, כך שהמיגרציה 00004 רצה
כשעוד יש עותק ב־v2, או הסירי את השורה `DROP NOT NULL` מ־00004. הריצי שוב את הבדיקה של צעד 7, והעתיקי לפתק את
השגיאה הראשונה. החזירי את המצב.

**10. הכול, ואז התחייבות וכיבוי:**

<div dir="ltr" align="left">

```bash
go test -race ./...
gofmt -l . ; go vet ./... ; staticcheck ./...
cd ~/senior-course
git add ch30
git commit -m "Chapter 30: rename a column with expand and contract, zero failed requests"
docker compose -p ch30 -f ch30/data/compose.yaml down -v
```

</div>

## פלט צפוי

הבדיקה של צעד 7, עם `-v`. מספרי הבקשות יהיו אחרים אצלך; עמודת הכשלונות היא אפס בכל שורה:

<div dir="ltr" align="left">

```text
=== RUN   TestRenameWithZeroFailedRequests
    rollout_test.go:138: expand: add display_name                      requests=322   failed=0
    rollout_test.go:138: backfill in batches, then switch reads        requests=339   failed=0
    rollout_test.go:138: swap NOT NULL, then stop writing fullname     requests=320   failed=0
    rollout_test.go:138: contract: drop fullname                       requests=110   failed=0
    rollout_test.go:190: users at the end: 5316, of them 316 written during the rollout
--- PASS: TestRenameWithZeroFailedRequests (1.06s)
```

</div>

והבדיקות המלאות בסוף צעד 10 (הזמנים יהיו שונים):

<div dir="ltr" align="left">

```text
ok  	example.com/ch30-expand-contract/app	…s
ok  	example.com/ch30-expand-contract/db	…s
ok  	example.com/ch30-expand-contract/events	…s
?   	example.com/ch30-expand-contract/internal/testdb	[no test files]
?   	example.com/ch30-expand-contract/migrations	[no test files]
ok  	example.com/ch30-expand-contract/rollout	…s
```

</div>

## רשימת בדיקה

- [ ] הבדיקה של צעד 7 עוברת עם `-race`, ובכל ארבע השורות `failed=0`.
- [ ] מטריצת התאימות עוברת בכל עשרים הצירופים.
- [ ] ‏`TestBackfillFillsRowsWrittenByOldCode` עוברת: המילוי עובד על יותר משתי מנות.
- [ ] השורות מצעדים 2, 3 ו־9 בפתק.
- [ ] בדיקות האירוע עוברות, ובפתק: מתי מותר להפסיק לשלוח `fullname` באירוע.
- [ ] כל הבדיקות ירוקות בלי דילוגים, וכלי הניתוח מצעד 10 נקיים.

## טעויות נפוצות ותיקונים

- **כל הבדיקות של המסד מסומנות `SKIP`.** המשתנה `DATABASE_URL` לא מוגדר בחלון הזה. מריצים שוב את שורת ה־`export`
  מצעד 1.
- **המיגרציה 00004 נכשלת ב־`VALIDATE`, עם `is violated by some row`.** יש שורות שבהן העמודה החדשה ריקה: המילוי
  לא רץ, או שעותק כתב רק לעמודה הישנה אחרי המילוי. בדקי שהתוכנית פורסת את v2 לפני המיגרציה 00003.
- **הוספה של v4 נכשלת עם `violates not-null constraint`.** העמודה הישנה עדיין חובה, כי ב־00004 חסרה השורה
  `DROP NOT NULL`.
- **המילוי נכשל עם `invalid transaction termination`.** הקובץ רץ בתוך טרנזקציה של goose, ושם אסור לאשר באמצע.
  ההערה `-- +goose NO TRANSACTION` צריכה להיות בשורה הראשונה, לפני `-- +goose Up`.
- **המילוי נכשל עם `unterminated dollar-quoted string`.** בלי שתי ההערות שתוחמות פקודה ארוכה סביב הבלוק,
  goose חותך אותו בנקודה־פסיק הראשונה. עטפי אותו בשתיהן, כמו בקובץ של הפתרון.
- **הבדיקות קורסות עם `panic: TODO`.** זה ה־SQL של גרסה שעוד לא מילאת ב־`app.go`; ההודעה מציינת איזו. עד צעד 4,
  הריצי רק את הבדיקות של צעדים 2 ו־3.

## רמזים

- המילוי: בלוק `DO` עם לולאה על טווחים של `id`, מ־0 ועד המזהה הגבוה ביותר. בכל טווח, עדכון של השורות שבהן
  העמודה החדשה ריקה, ו־`COMMIT`. ההערות בקובץ 00002 מראות את הסגנון.
- ב־00004, שמרי על הסדר: הוספת האילוץ, אימות, `SET NOT NULL`, מחיקת האילוץ, ואז `fullname DROP NOT NULL`.
- בתוכנית: v3 צריכה עמודה מלאה, ו־v4 צריכה עמודה ישנה שאינה חובה. הצעד האחרון לא פורס כלום, ולכן `Deploy` שלו הוא 0.
- ב־`Name`: בחירה לפי מספר הגרסה עם שני מקרים, ובסוף שגיאה שעוטפת את `ErrUnknownVersion`.

## הפתרון

התיקייה `solution/`, בעיקר שלוש המיגרציות החדשות, השירות והתוכנית. **למה זה עובד:** כל גרסת קוד עובדת על רצף של גרסאות סכמה, והרצפים של כל שתי גרסאות סמוכות
חופפים, ולכן בכל רגע של הפריסה המתגלגלת שני העותקים מצליחים. ‏`Run` מחכה אחרי כל מיגרציה ואחרי כל שדרוג עד שכל
עותק סיים עוד בקשות, כך שכל צירוף שקורה בייצור קורה גם בבדיקה. והבדיקה בסוף משווה כל שורה לכתיבה המוצלחת האחרונה
שלה, כך ש„אפס כשלונות” לא יכול להסתיר כתיבה שהלכה לאיבוד בין העמודות.

## מה התרגיל מדגים

- הרחבה ואז כיווץ על מסד אמיתי, תחת תנועה, עם הוכחה שאף בקשה לא נכשלה.
- למה בין שני צעדים צריכה להסתיים פריסה: מטריצת תאימות שמראה בדיוק אילו צירופים נשברים.
- נעילות ב־PostgreSQL: אילוץ לא תקף ואימות בנפרד, מילוי במנות, ותור נעילות שעוצר גם קריאות.
- ‏`lock_timeout` כפרמטר של חיבור המיגרציות, כך שהוא חל גם על פקודות שרצות בלי טרנזקציה.
- אותה תבנית באירועים: שדה כפול בתקופת המעבר, קורא סובלני, וגרסה בתוך האירוע.

</div>
