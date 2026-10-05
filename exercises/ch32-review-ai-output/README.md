---
direction: rtl
lang: he
---

<div dir="rtl" lang="he">

# תרגיל פרק 32: לסקור קוד שכתב כלי AI

**מטרה:** לסקור מודול Go קטן שכלי AI כתב לפי מפרט: מטמון עם זמן חיים (TTL) ושכבת HTTP מעליו. הקוד מתקמפל, נראה
סביר, והבדיקות שהכלי כתב עוברות. יש בו ארבעה פגמים שתולים. את מוצאת אותם קודם בקריאה, כותבת לכל אחד בדיקה שנכשלת,
מתקנת, ורושמת את הסקירה כהערות לפי חומרה, כמו שהיית מפרסמת אותן על PR.

**זמן משוער:** יום לימוד אחד, יום 91: כ־55 דקות בנייה (כ־20 לקריאה ולסקירה, כ־20 לבדיקות שלך, כ־15 לתיקונים
ולהשוואה).

**ידע מקדים:** פרק 32, ופרק 27 (סקירה לפי סיכון, וחומרות ההערות). מפרק 5: נעילות, `sync.RWMutex`, ערכים אטומיים,
`select` ובדיקת מרוצים עם `-race`. מפרק 6: `net/http`, ניתוב לפי תבנית ו־`httptest`. ‏staticcheck מותקן לפי פרק 2.
התרגיל לא מייבא קוד מתרגילים קודמים.

## מה בתיקייה הזאת

| קובץ | תפקיד |
| --- | --- |
| `data/spec.md` | המפרט שניתן לכלי, והסיכום שהכלי כתב כשסיים |
| `starter/ttlcache/cache.go` | המטמון, כפי שהכלי כתב אותו |
| `starter/ttlcache/handler.go` | שכבת ה־HTTP, כפי שהכלי כתב אותה |
| `starter/ttlcache/ai_test.go` | חמש הבדיקות שהכלי כתב; כולן עוברות |
| `starter/ttlcache/defects_test.go` | ארבע בדיקות, אחת לכל פגם; אל תפתחי אותו לפני צעד 5 |
| `starter/REVIEW.md` | התבנית לסקירה שלך |
| `tests/ttlcache/` | שני קובצי הבדיקות הקנוניים; עותקים זהים ב־`starter/` וב־`solution/` |
| `solution/` | המודול המתוקן, ו־`REVIEW.md` עם הסקירה המלאה |

## צעדים

**1. העתיקי את קוד ההתחלה.**

<div dir="ltr" align="left">

```bash
mkdir -p ~/senior-course/ch32
cp -R "$COURSE/exercises/ch32-review-ai-output/starter/." ~/senior-course/ch32/
cp "$COURSE/exercises/ch32-review-ai-output/data/spec.md" ~/senior-course/ch32/
cd ~/senior-course/ch32
go build ./... && go vet ./... && echo builds
```

</div>

**2. קראי את המפרט, ואז את הקוד.** קודם `spec.md` מההתחלה ועד הסוף, כולל הסיכום של הכלי. אחר כך את שני קובצי
הקוד, לפי סדר הסיכון של פרק 27: נעילות, ה־goroutine שברקע, השוואות הזמן, ומה שמגיע מהלקוח. מלאי את החלק הראשון
של `REVIEW.md` לפני שאת מריצה משהו. רשמי גם הערות שאת לא בטוחה בהן, כשאלות.

**3. הריצי את הבדיקות של הכלי, ואת הכלים.** הבדיקות של הכלי עוברות. שימי לב מה כל אחת בודקת, ומה לא:

<div dir="ltr" align="left">

```bash
go test -v -run 'TestSetGet|TestExpiry|TestSweep|TestStats|TestHandler' ./ttlcache/
go vet ./...
staticcheck ./...
```

</div>

**4. כתבי בדיקה לכל ממצא.** צרי קובץ חדש, `ttlcache/mine_test.go`, בחבילה `ttlcache_test`. השעון המזויף `newClock`
כבר מוגדר ב־`ai_test.go`, ואפשר להשתמש בו. לכל ממצא, בדיקה אחת שנכשלת על הקוד הנוכחי:

| סוג ממצא | איך בודקים אותו |
| --- | --- |
| מרוץ | כמה goroutines שקוראות במקביל, והרצה עם `-race` |
| goroutine שלא נגמרת | ביטול ה־context, והמתנה עם זמן קצוב לערוץ שאמור להיסגר |
| קצה של זמן | שעון מזויף שמתקדם בדיוק עד הקצה, ונקודה לפניו ואחריו |
| שגיאה שנבלעת | קלט שגורם לשגיאה, ובדיקה של קוד התשובה ושל מה שנשמר |

<div dir="ltr" align="left">

```bash
go test -race -v -run 'Mine' ./ttlcache/
```

</div>

(תני לבדיקות שלך שמות שמתחילים ב־`TestMine`, כדי שהפקודה תריץ רק אותן.)

**5. הריצי את הבדיקות הקנוניות.** עכשיו פתחי את `defects_test.go`, והריצי את הכול. ארבע בדיקות צריכות להיכשל.
השווי לרשימה שלך: מה מצאת בקריאה, מה רק בבדיקה, ומה לא מצאת.

<div dir="ltr" align="left">

```bash
go test -race ./ttlcache/
```

</div>

**6. תקני את ארבעת הפגמים**, אחד אחד, והריצי אחרי כל תיקון. בסוף הכול ירוק, וגם הכלים שקטים:

<div dir="ltr" align="left">

```bash
go test -race -count=1 ./...
gofmt -l .
go vet ./...
staticcheck ./...
```

</div>

**7. השלימי את `REVIEW.md`**: ההערות לפי חומרה, הבדיקה של טענות הסיכום, ומה היית כותבת בהנחיה לכלי בפעם הבאה. רק
אז פתחי את `solution/REVIEW.md` והשווי.

**8. שמרי את העבודה.**

<div dir="ltr" align="left">

```bash
cd ~/senior-course
git add ch32
git commit -m "ch32: review of AI-written ttlcache, tests for four defects, fixes"
```

</div>

## פלט צפוי

בצעד 3, הבדיקות של הכלי עוברות, ו־staticcheck מוצא דבר אחד:

<div dir="ltr" align="left">

```text
--- PASS: TestSetGet (0.00s)
--- PASS: TestExpiry (0.00s)
--- PASS: TestSweepRemovesExpired (0.00s)
--- PASS: TestStatsCount (0.00s)
--- PASS: TestHandlerRoundTrip (0.00s)
ttlcache/cache.go:110:5: ineffective break statement. Did you mean to break out of the outer loop? (SA4011)
```

</div>

בצעד 5, ארבעת הכישלונות (המספרים במרוץ ישתנו מהרצה להרצה, ומעליהם יופיעו דוחות `WARNING: DATA RACE`):

<div dir="ltr" align="left">

```text
--- FAIL: TestConcurrentGetsDoNotRace (0.01s)
    defects_test.go:35: stats = {Hits:1112 Misses:1135 Entries:1}, want 1600 hits and 1600 misses
--- FAIL: TestJanitorStopsOnCancel (1.01s)
    defects_test.go:50: janitor still running one second after its context was cancelled
--- FAIL: TestEntryExpiresExactlyAtTTL (0.00s)
    --- FAIL: TestEntryExpiresExactlyAtTTL/exactly_at_the_TTL (0.00s)
        defects_test.go:73: Get after 1m0s: found = true, want false
--- FAIL: TestPutTooLargeIs413AndStoresNothing (0.50s)
    defects_test.go:98: PUT of 16 bytes with a limit of 8: status = 204, want 413
```

</div>

ובצעד 6, אחרי התיקונים:

<div dir="ltr" align="left">

```text
ok  	example.com/ch32-review-ai-output/ttlcache	2.463s
```

</div>

## רשימת בדיקה

- [ ] החלק הראשון של `REVIEW.md` נכתב לפני שהרצת בדיקה כלשהי.
- [ ] לכל אחד מארבעת הפגמים יש בדיקה שלך שנכשלה לפני התיקון ועוברת אחריו.
- [ ] כל הבדיקות עוברות עם `-race`, וארבעת הכלים מצעד 6 שקטים.
- [ ] בסקירה יש לכל פגם חומרה, מה עלול לקרות, הצעה לתיקון, ואיך מצאת אותו.
- [ ] בדקת כל אחת מארבע הטענות בסיכום של הכלי, עם ראיה.
- [ ] כתבת מה היית מוסיפה להנחיה כדי שהכלי לא יחזור על הטעויות, והעבודה שמורה ב־Git.

## טעויות נפוצות ותיקונים

- **בדיקת המרוץ עוברת אצלך.** הרצת בלי `-race`. בלי הדגל, מרוץ נתונים לרוב לא נראה בכלל; הוסיפי אותו לכל הרצה.
- **התיקון למרוץ הוא `Lock` במקום `RLock`, והכול עובר.** זה נכון, ומבטל את המקביליות שהמפרט ביקש לקריאות. מונים
  מטיפוס `atomic.Uint64` שומרים את הנעילה לקריאה בלבד.
- **אחרי התיקון למקרה הקצה, `Get` ו־`Sweep` עדיין לא מסכימות.** תיקנת השוואה אחת מתוך שתיים. פונקציה אחת שמגדירה
  „פג תוקף”, ששתיהן קוראות לה, מונעת את זה מעכשיו והלאה.
- **בדיקת ה־413 נכשלת עם 400.** בדקת שהקריאה נכשלה, אבל לא איזו שגיאה. בגוף גדול מדי, השגיאה היא
  `*http.MaxBytesError`, ו־`errors.As` מזהה אותה.
- **בדיקת ה־janitor נתקעת לנצח.** חיכית לערוץ בלי זמן קצוב. המתיני עם `select` ו־`time.After`, כמו בפרק 5.

## רמזים

- נעילה לקריאה מאפשרת לכמה קוראות להחזיק אותה יחד. כל כתיבה תחתיה, גם למונה, היא מרוץ.
- בתוך `select`, הפקודה `break` יוצאת מה־`select` בלבד, לא מהלולאה שמסביבו.
- המתודה `After` של זמן היא קפדנית: ברגע שבו שני הזמנים שווים, היא מחזירה שקר.
- קו תחתון במקום שגיאה הוא תמיד מקום לעצור ולשאול מה קורה כשהיא לא ריקה.

## הפתרון

המודול המתוקן נמצא בתיקייה `solution/ttlcache/`, והסקירה המלאה ב־`solution/REVIEW.md`. המונים הפכו לאטומיים, כך
שהקריאות ממשיכות להחזיק רק את הנעילה לקריאה. ה־janitor חוזר כשה־context מבוטל, וה־`defer` סוגר את הערוץ. ההגדרה
של „פג תוקף” נמצאת בפונקציה אחת שמשמשת את שתי המתודות, ושכבת ה־HTTP בודקת את שגיאת הקריאה, מחזירה 413 או 400, ולא
שומרת חלק מערך.

## מה התרגיל מדגים

- סקירה של קוד שכתב כלי AI לפי סיכון, מול המפרט ולא מול מה שהקוד אומר על עצמו.
- ארבעה פגמים טיפוסיים לקוד סביר למראה: מרוץ תחת נעילה לקריאה, goroutine שלא נגמרת, קצה זמן, ושגיאה שנבלעת.
- בדיקה שנכשלת לפני התיקון לכל פגם, ואיזה כלי היה תופס כל אחד: `-race`, ‏staticcheck, או רק בדיקה בקצה.
- בדיקה של סיכום בטוח בעצמו מול הקוד: כל טענה בו מסמנת מקום לבדוק.
- הערות סקירה לפי חומרה, והנחיה טובה יותר לפעם הבאה.

</div>
