---
direction: rtl
lang: he
---

<div dir="rtl" lang="he">

# תרגיל פרק 35: `drills`, תרגילי קוד מעשיים

**מטרה:** לפתור חמישה תרגילי קוד ב־Go בתנאים של ראיון: מסגרת זמן, חשיבה בקול, ובדיקות שהן המפרט. מגביל קצב,
מתזמן משימות, מטמון עם זמן חיים, מקבל webhooks שבטוח לניסיונות חוזרים, והזמנה בלי הזמנה כפולה בזיכרון וב־PostgreSQL.
אחריהם, סקירה של PR שבור בזמן קצוב, לפני שמסתכלים במפתח.

**זמן משוער:** שני ימי לימוד. יום 94: כ־85 דקות (מגביל קצב 25, מתזמן 35, מטמון 25). יום 95: כ־85 דקות (webhooks 30,
הזמנה 35, סקירה 20). הזמנים הם מסגרות של ראיון; אם לא סיימת, עצרי, רשמי, והמשיכי אחר כך מחוץ לשעון.

**ידע מקדים:** פרק 35. מפרקים קודמים: מקביליות, ביטול וגלאי המרוצים (פרק 5), שרת HTTP ובדיקות שלו (פרק 6),
PostgreSQL בקונטיינר ו־pgx (פרק 10), דלי אסימונים (פרק 16), ניסיונות חוזרים ואידמפוטנטיות (פרק 14). Go ו־Docker
מותקנים לפי פרק 2. התרגיל לא מייבא קוד מתרגילים קודמים; המודול תלוי רק ב־`github.com/jackc/pgx/v5`, שכבר נעול
ב־`go.mod`.

## מה בתיקייה הזאת

| קובץ | תפקיד |
| --- | --- |
| `starter/clock/clock.go` | השעון: `Clock`, `Timer`, `Real()` ו־`Fake` עם `Advance` ו־`BlockUntil`; מוכן |
| `starter/ratelimit/ratelimit.go` | תרגיל 1: טיפוסים וחתימות, גופים עם `panic("TODO")` |
| `starter/scheduler/scheduler.go` | תרגיל 2: מצבים, אפשרויות ושגיאות מוכנים; `Permanent`, `Exponential` והמתזמן הם `TODO` |
| `starter/ttlcache/ttlcache.go` | תרגיל 3: החתימות של המטמון הגנרי |
| `starter/webhook/webhook.go` | תרגיל 4: הקבועים, `Event`, `Processor`, והחתימות של ה־handler |
| `starter/booking/booking.go` | תרגיל 5: `Booking`, השגיאות והממשק `Store`; מוכנים |
| `starter/booking/memory.go`, `starter/booking/postgres.go` | תרגיל 5: שני המימושים; בסכמה חסר האילוץ |
| `starter/review.md` | תבנית לסקירה של סעיף 35.7 |
| `starter/go.mod`, `starter/go.sum` | המודול `example.com/ch35-drills` |
| `tests/clock/clock_test.go` | בדיקות של השעון המזויף (עוברות כבר בקוד ההתחלה) |
| `tests/ratelimit/ratelimit_test.go` | 8 בדיקות: תצורה, burst, מילוי, המתנה חלקית, מפתחות, שעון אחורה, מקביליות, גיזום |
| `tests/scheduler/scheduler_test.go` | 14 בדיקות: זמן, ניסיונות חוזרים, שגיאה קבועה, panic, ביטול, השהיה, מגבלה, סדר, כיבוי |
| `tests/ttlcache/ttlcache_test.go` | 9 בדיקות: פג עצל, TTL לערך, LRU, סריקה, janitor, שימוש מקבילי |
| `tests/webhook/webhook_test.go` | 8 בדיקות: בקשות פסולות, כפילויות ברצף ובמקביל, המתנה למוביל, כישלון, panic, חלון, ניתוק |
| `tests/booking/booking_test.go` | 4 בדיקות, כל אחת על שני המימושים: החוזה, ביטול, הזמנה כפולה במקביל, משאבים שונים |
| `data/compose.yaml` | PostgreSQL עם שם הפרויקט `ch35`, בפורט 5435 |
| `data/review-drill-pr.md`, `data/review-drill.diff` | ה־PR לסקירה: התיאור והשינוי |
| `solution/` | הפתרון המלא, עם עותקים זהים של הבדיקות, ו־`review-key.md`, מפתח הסקירה |

## יום 94: מגביל קצב, מתזמן, מטמון

**1. העתיקי את קוד ההתחלה ואת קבצי הנתונים, ובדקי שהם מתקמפלים:**

<div dir="ltr" align="left">

```bash
mkdir -p ~/senior-course/ch35
cd ~/senior-course/ch35
cp -r "$COURSE"/exercises/ch35-drills/starter/. .
cp -r "$COURSE"/exercises/ch35-drills/data .
go build ./... && go vet ./... && echo builds
go test -race ./clock
```

</div>

**2. קראי את סעיף 35.1.3** ואת `clock/clock.go`. כל תרגיל של היום מקבל את השעון כפרמטר.

**3. תרגיל 1, מגביל קצב (25 דקות).** הפעילי שעון עצר. קראי את `ratelimit/ratelimit_test.go`, כתבי בראש
`ratelimit.go` שתי שאלות הבהרה כהערה, וכתבי את הפתרון בקול. הריצי:

<div dir="ltr" align="left">

```bash
go test -race ./ratelimit
```

</div>

**4. תרגיל 2, מתזמן (35 דקות).** אותו הדבר, עם `scheduler`. התחילי מ־`Exponential`, פונקציה קטנה עם בדיקה משלה,
ואז המתזמן עצמו. הריצי את הבדיקות אחת אחת בזמן העבודה:

<div dir="ltr" align="left">

```bash
go test -race -run TestExponential ./scheduler
go test -race ./scheduler
```

</div>

**5. תרגיל 3, מטמון (25 דקות):**

<div dir="ltr" align="left">

```bash
go test -race ./ttlcache
```

</div>

**6. רשמי בפתק** את הזמן של כל תרגיל, איפה נתקעת, ואת המחיר שהיית אומרת בקול לכל החלטה. אחר כך:

<div dir="ltr" align="left">

```bash
cd ~/senior-course
git add ch35
git commit -m "Chapter 35: rate limiter, scheduler, TTL cache drills"
```

</div>

## יום 95: webhooks, הזמנה, סקירה

**7. תרגיל 4, מקבל webhooks (30 דקות):**

<div dir="ltr" align="left">

```bash
cd ~/senior-course/ch35
go test -race ./webhook
```

</div>

**8. תרגיל 5, הזמנה (35 דקות).** קודם `Memory`, בלי מסד: תת־הבדיקות של `postgres` ידלגו. אחר כך הוסיפי לסכמה
ב־`postgres.go` את האילוץ, כתבי את שלוש המתודות, והריצי מול מסד אמיתי:

<div dir="ltr" align="left">

```bash
go test -race -v ./booking
docker compose -p ch35 -f data/compose.yaml up -d --wait
export DATABASE_URL=postgres://course:change-me@localhost:5435/drills
go test -race -count=1 -v ./booking
docker compose -p ch35 -f data/compose.yaml down -v
unset DATABASE_URL
```

</div>

**9. סקירת קוד (20 דקות).** קראי את התיאור ואת השינוי, שני הקבצים בתיקייה `data/` ששמם מתחיל ב־`review-drill`,
ומלאי את `review.md` לפי סעיף 35.7.2. שמרי ועשי commit **לפני** שאת פותחת את המפתח:

<div dir="ltr" align="left">

```bash
cd ~/senior-course
git add ch35
git commit -m "Chapter 35: webhook and booking drills, review before key"
less "$COURSE"/exercises/ch35-drills/solution/review-key.md
```

</div>

**10. השווי למפתח,** ומלאי את החלק האחרון של `review.md`, זה שאחרי הבדיקה מול המפתח. אחר כך הסוללה המלאה, כמה פעמים ברצף:

<div dir="ltr" align="left">

```bash
cd ~/senior-course/ch35
gofmt -l .
go vet ./...
go test -race -count=3 ./...
```

</div>

## פלט צפוי

הסוללה של צעד 10, בלי `DATABASE_URL` (הזמנים שלך יהיו אחרים):

<div dir="ltr" align="left">

```text
ok  	example.com/ch35-drills/booking	1.1s
ok  	example.com/ch35-drills/clock	1.0s
ok  	example.com/ch35-drills/ratelimit	1.1s
ok  	example.com/ch35-drills/scheduler	1.9s
ok  	example.com/ch35-drills/ttlcache	1.2s
ok  	example.com/ch35-drills/webhook	1.4s
```

</div>

עם המסד, בצעד 8, כל בדיקה של `booking` רצה פעמיים:

<div dir="ltr" align="left">

```text
--- PASS: TestBookingContract (0.06s)
    --- PASS: TestBookingContract/memory (0.00s)
    --- PASS: TestBookingContract/postgres (0.06s)
--- PASS: TestCancelFreesTheRange (0.03s)
    --- PASS: TestCancelFreesTheRange/memory (0.00s)
    --- PASS: TestCancelFreesTheRange/postgres (0.03s)
--- PASS: TestNoDoubleBookingUnderConcurrency (0.08s)
    --- PASS: TestNoDoubleBookingUnderConcurrency/memory (0.00s)
    --- PASS: TestNoDoubleBookingUnderConcurrency/postgres (0.08s)
--- PASS: TestDifferentResourcesDoNotBlockEachOther (0.05s)
    --- PASS: TestDifferentResourcesDoNotBlockEachOther/memory (0.00s)
    --- PASS: TestDifferentResourcesDoNotBlockEachOther/postgres (0.05s)
PASS
```

</div>

בלי המסד, במקום שורות ה־`postgres` תראי שורות דילוג, עם הסבר שמבקש להגדיר את `DATABASE_URL`.

## רשימת בדיקה

- [ ] כל חמשת התרגילים ירוקים עם `go test -race -count=3 ./...`, ו־`gofmt -l .` לא מדפיסה דבר.
- [ ] הבדיקות של ההזמנה עברו גם מול PostgreSQL: תת־הבדיקות `postgres` עברו ולא דילגו.
- [ ] בסכמה שלך יש אילוץ החרגה, וההרחבה `btree_gist` מוזכרת בה.
- [ ] בקוד שלך אין `time.Now()` ואין `time.Sleep` מחוץ לבדיקות; הזמן מגיע מהשעון שהוזרק.
- [ ] בפתק רשומים הזמן של כל תרגיל, איפה נתקעת, ומחיר אחד לכל תרגיל.
- [ ] הסקירה שלך שמורה ב־commit לפני שפתחת את המפתח, וההשוואה רשומה בסופה.
- [ ] הורדת את המסד עם `down -v`.

## טעויות נפוצות ותיקונים

- **בדיקות המתזמן נתקעות עד `timed out waiting for`.** כמעט תמיד הלולאה לא מתעוררת אחרי שינוי: `Submit`, `Resume`
  או ניסיון שהסתיים לא שולחים ל־channel ההתעוררות, או שהשליחה חוסמת. שליחה עם `select` ו־`default` לערוץ בגודל 1.
- **`TestJanitorSweepsOnItsInterval` נתקעת.** ה־janitor יוצר טיימר אחד ואז לא יוצר את הבא, או שהוא משתמש
  ב־`time.NewTimer` במקום בשעון שקיבל. כל סבב: `clk.NewTimer(every)`, ואחרי שהוא יורה, סריקה ואז טיימר חדש.
- **`processor called 2 times` בבדיקת הכפילויות המקבילות.** הבדיקה אם המזהה קיים והרישום שלו נעשים בשתי נעילות
  נפרדות. ההחלטה, מוביל, עוקב או כפילות, תחת נעילה אחת.
- **אזהרת מרוץ בבדיקה המקבילית של המטמון.** כנראה נעילת קריאה, `RLock`, בתוך `Get`. פגיעה מזיזה את האיבר
  ברשימה, ולכן צריכה נעילה מלאה.
- **`btree_gist` או `data type text has no default operator class for access method "gist"`.** האילוץ משלב שוויון על
  `text` עם חפיפת טווחים, ולשוויון בתוך GiST צריך את ההרחבה. בסכמה: `CREATE EXTENSION IF NOT EXISTS btree_gist;`
  לפני הטבלה.
- **הבדיקות של `postgres` נכשלות על `connection refused`.** המסד לא עלה, או שהכתובת מצביעה על פורט אחר.
  הריצי שוב את פקודת ה־`up` עם `--wait`, ובדקי שהפורט 5435.

## רמזים

- **מגביל קצב:** כל דלי צריך רק שני שדות: כמה אסימונים יש בו, ומתי נגעו בו לאחרונה. `math.Min` ו־`math.Ceil`.
- **מתזמן:** אל תפעילי goroutine לכל משימה שישנה עד הזמן שלה. לולאה אחת, שבכל סבב מתחילה את מה שהגיע זמנו ומכוונת
  טיימר אחד למשימה הבאה. הפונקציה `context.WithCancel` לכל ניסיון נותנת לך את `Cancel` של משימה שרצה.
- **מטמון:** `container/list` ומפה מהמפתח ל־`*list.Element`. בסריקה, קראי את `el.Next()` לפני המחיקה.
- **מקבל webhooks:** במפה של המשלוחים שבעבודה, שמרי channel שנסגר בסיום ואת השגיאה. מי שמוצא שם רשומה, ממתין עליו.
  לגיזום החלון, slice של מזהים לפי סדר הסיום מספיק.
- **הזמנה:** את קוד השגיאה בודקים עם `errors.As` לטיפוס `*pgconn.PgError`. בהכנסה, הטווח נבנה כך:

<div dir="ltr" align="left">

```sql
tstzrange($3, $4, '[)')
```

</div>

## הפתרון

הפתרון בתיקייה `solution/`, חבילה לכל תרגיל, ומפתח הסקירה ב־`solution/review-key.md`. **למה הוא עובד:** בכל
תרגיל, ההחלטה שמגינה על המצב המשותף נעשית בצעד אחד תחת נעילה אחת: לקיחת אסימון, התחלת ניסיון, הקדמה ברשימה, בחירת
מוביל, בדיקת חפיפה והכנסה. הזמן מגיע תמיד מהשעון המוזרק, ולכן הבדיקות מזיזות שעות בלי לחכות להן. ובגרסת
PostgreSQL של ההזמנה, אין בדיקה באפליקציה בכלל: האילוץ במסד מחליט, גם כשעשרות בקשות מגיעות יחד.

## מה התרגיל מדגים

- סבב קוד מעשי כמו בראיון: הבדיקות כמפרט, מסגרת זמן, ומחירים שאומרים בקול.
- בדיקות דטרמיניסטיות לקוד שתלוי בזמן, עם שעון מזויף והמתנה למצב ולא לזמן.
- בטיחות מקבילית בחמש צורות, כולן נבדקות עם goroutines וגלאי המרוצים.
- אידמפוטנטיות בצד המקבל של webhooks, כולל משלוחים כפולים בו־זמניים.
- שתי דרכים למנוע הזמנה כפולה: נעילה לכל משאב, ואילוץ החרגה עם `btree_gist` ב־PostgreSQL.
- סקירת קוד לפי סיכון וחומרה, מול מפתח.

</div>
