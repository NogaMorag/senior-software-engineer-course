---
direction: rtl
lang: he
---

<div dir="rtl" lang="he">

# תרגיל פרק 5: `linkcheck`, בודק קישורים מקבילי

**מטרה:** לבנות בודק קישורים שמריץ לכל היותר N בקשות בבת אחת, נותן לכל בקשה זמן קצוב משלה, יודע לעצור את
כל הריצה (בשגיאה הראשונה, או ב־Ctrl-C), ומדפיס דוח לפי סדר הקלט. פעם עם `errgroup` ופעם כ־worker pool של
channels, ובסוף להוכיח עם גלאי המרוצים ועם בדיקת הדליפות שהקוד נקי.

**זמן משוער:** שלושה ימי לימוד, ימים 8 עד 10: כ־45 דקות בנייה ביום 8, כ־55 ביום 9 וכ־40 ביום 10 (יום קל).
**ידע מקדים:** פרק 5 כולו; פרקים 3 ו־4 (מודולים, בדיקות טבלאיות, שגיאות עם `%w` ו־`errors.Is`); Go מותקנת
לפי פרק 2. המודול תלוי בחבילה אחת מחוץ לספרייה הסטנדרטית, `golang.org/x/sync/errgroup`, שכבר נעולה
ב־`go.mod`. התרגיל לא מייבא קוד מתרגילים קודמים.

## מה בתיקייה הזאת

| קובץ | תפקיד |
| --- | --- |
| `starter/linkcheck.go` | `Result`, `Verdict` ו־`Checker` מוכנים; `TODO 1` ב־`checkOne` ו־`TODO 2` ב־`Check` |
| `starter/pool.go` | `TODO 3`: הפונקציה `CheckPool`, שלושה שלבים של channels |
| `starter/report.go` | `WriteReport`, מוכן |
| `starter/cmd/linkcheck/main.go` | התוכנית: דגלים, קריאת כתובות, `signal.NotifyContext`, קודי יציאה; מוכנה |
| `starter/linkcheck_test.go` | עותק של קובץ הבדיקות |
| `starter/go.mod`, `starter/go.sum` | המודול `example.com/ch05-linkcheck`, עם `golang.org/x/sync` נעול |
| `tests/linkcheck_test.go` | הבדיקות הקנוניות: שש פונקציות, כל אחת לשני המימושים, עם בדיקת דליפות |
| `solution/` | הפתרון המלא, באותו מבנה |
| `data/urls.txt` | רשימת כתובות לדוגמה להרצה ידנית |
| `data/racedemo/main.go` | המרוץ השתול של סעיף 5.8, מודול נפרד וקטן |
| `data/planted-leak.txt` | הדליפה השתולה של סעיף 5.9, להדבקה זמנית |

## יום 8: בקשה אחת עם זמן קצוב, וכל הכתובות במקביל

**1. העתיקי את נקודת ההתחלה** לתיקייה של הפרק, ובדקי שהיא מתקמפלת:

<div dir="ltr" align="left">

```bash
mkdir -p ~/senior-course/ch05
cp -r $COURSE/exercises/ch05-linkcheck/starter/. ~/senior-course/ch05/
cd ~/senior-course/ch05
go build ./... && go vet ./...
```

</div>

אין פלט, וזה הסימן שהכול תקין.

**2. ודאי שגלאי המרוצים עובד אצלך.** ב־Linux הוא צריך מהדר C, שהותקן בפרק 2 (סעיף 2.3). הפקודה הראשונה
לא תזיק אם הוא כבר שם; השנייה צריכה להסתיים בלי שגיאה:

<div dir="ltr" align="left">

```bash
sudo apt install -y build-essential
go test -race -run TestWriteReport .
```

</div>

<div dir="ltr" align="left">

```text
ok  	example.com/ch05-linkcheck	1.2s
```

</div>

**3. קראי את `linkcheck.go`** מלמעלה למטה, בסדר הזה:

- שתי השגיאות, `ErrBroken` ו־`ErrSkipped`.
- הטיפוס `Result` והפונקציה `Verdict` (סדר הבדיקות בה חשוב).
- השדות של `Checker`.
- שתי מתודות העזר, `client` ו־`workers`.

**4. מלאי את `TODO 1`** (`checkOne`), לפי סעיף 5.4. ההערה בקוד מפרטת את הצעדים. הנקודה העדינה: כש־`Do`
נכשלת וה־context כבר לא תקין, שומרים את `ctx.Err()`, ולא את השגיאה של הלקוח.

**5. מלאי את `TODO 2`** (`Check`), בגרסה של היום: `results` בגודל הקלט, goroutine לכל כתובת עם
`sync.WaitGroup`, וכל goroutine כותבת לתא שלה. בלי מגבלה ובלי ביטול; את אלה תוסיפי מחר.

**6. הריצי את הבדיקה של היום**, רק למימוש `Check` (`CheckPool` עדיין זורקת panic):

<div dir="ltr" align="left">

```bash
go test -race -run 'TestCheckClassifies/^Check$' -v .
```

</div>

הבדיקה שולחת חמש כתובות: תקינה, 404, איטית, לא חוקית, ושרת סגור. היא דורשת את חמשת הסיווגים האלה, לפי
הסדר:

<div dir="ltr" align="left">

```text
ok  broken  timeout  error  error
```

</div>

והיא מסתיימת תוך פחות משתי שניות: הנתיב האיטי לא מחכה לעשר השניות שלו.

**7. שמרי את העבודה:**

<div dir="ltr" align="left">

```bash
cd ~/senior-course
git add ch05
git commit -m "Chapter 5: checkOne with per-request timeout"
```

</div>

## יום 9: מגבלה, ביטול, וה־worker pool

**8. כתבי מחדש את `Check` עם `errgroup`**, לפי סעיף 5.6:

- קבוצה עם context משותף: `errgroup.WithContext`.
- מגבלה מיד אחריה: `g.SetLimit(c.workers())`.
- בתחילת כל goroutine, בדיקה של `gctx.Err()`; אם הוא בוטל, התוצאה היא `ErrSkipped`.
- החזרת שגיאה רק כשהשדה `FailFast` דולק.
- בסוף, השגיאה של `Wait`, ואם אין, `ctx.Err()`.

הוסיפי את הייבוא של `golang.org/x/sync/errgroup`; הוא כבר ב־`go.mod`.

<div dir="ltr" align="left">

```bash
cd ~/senior-course/ch05
go test -race -run 'TestCheck.*/^Check$' -v .
```

</div>

ארבע הבדיקות של היום: המקביליות לא עוברת את N (השרת סופר), הסדר נשמר למרות שהתשובות מגיעות הפוך,
fail-fast מבטל את הבקשות האיטיות, וביטול של ההורה עוצר הכול מהר.

**9. מלאי את `TODO 3`** (`CheckPool`), לפי סעיף 5.7. ההערה בקוד היא רשימת הצעדים. שני כללים שאסור לוותר
עליהם: המחולל תמיד סוגר את `jobs` (עם `defer`), והאוסף קורא מ־`out` עד שהוא נסגר.

**10. כל הבדיקות, לשני המימושים:**

<div dir="ltr" align="left">

```bash
go test -race -v .
```

</div>

**11. רשמי בפתק** במשפט אחד מתי היית בוחרת `Mutex` ומתי channel, ואיך `errgroup` משלב המתנה, שגיאה ראשונה
וביטול. אחר כך `git commit` כמו בצעד 7.

## יום 10: מרוץ, דליפה, וסיום

**12. המרוץ השתול.** הריצי את תוכנית ההדגמה בלי הגלאי ואיתו, ואז עם שני התיקונים. היא רצה ישירות מתיקיית
הקורס:

<div dir="ltr" align="left">

```bash
cd $COURSE/exercises/ch05-linkcheck/data/racedemo
go run .
go run -race .
go run -race . -fix mutex
go run -race . -fix atomic
```

</div>

הריצה הראשונה כנראה תדפיס פחות מ־8000. השנייה תדפיס `WARNING: DATA RACE` עם שתי מחסניות, ותסתיים
ב־`exit status 66`. שתי האחרונות ידפיסו `checked: 8000 of 8000` בלי אזהרה. רשמי בפתק את המספר שקיבלת
בריצה הראשונה.

**13. עכשיו בקוד שלך.** הוסיפי ל־`Checker` שדה `checked int`. בתוך ה־goroutine של `Check`, הגדילי אותו
עם `c.checked++`. הריצי את הבדיקות עם `-race`, ראי את הדוח מצביע על השורה שלך, ומחקי את השדה (או
הפכי אותו ל־`atomic.Int64`).

**14. הדליפה השתולה.** פתחי את `data/planted-leak.txt` של הקורס, והחליפי לרגע את לולאת האיסוף ב־`pool.go`
שלך בלולאה שבקובץ. הריצי:

<div dir="ltr" align="left">

```bash
cd ~/senior-course/ch05
go test -race -run 'TestCheckFailFast' -v .
```

</div>

הבדיקה תיכשל עם `leaked goroutines:` ושתי מחסניות: השליחה ל־`out` בתוך ה־worker, ו־`wg.Wait()`
ב־goroutine שסוגרת. החזירי את הלולאה המקורית.

**15. הסוללה המלאה,** כמה פעמים ברצף, כדי לתת לבעיות תזמון הזדמנות להופיע:

<div dir="ltr" align="left">

```bash
gofmt -l .
go vet ./...
go test -race -count=3 ./...
```

</div>

**16. הרצה אמיתית** על הקובץ לדוגמה, ואחר כך `git commit`:

<div dir="ltr" align="left">

```bash
go run ./cmd/linkcheck -workers 3 -timeout 2s $COURSE/exercises/ch05-linkcheck/data/urls.txt
echo "exit: $?"
```

</div>

## פלט צפוי

הסוללה של צעד 15 (הזמנים שלך יהיו אחרים):

<div dir="ltr" align="left">

```text
ok  	example.com/ch05-linkcheck	14.2s
?   	example.com/ch05-linkcheck/cmd/linkcheck	[no test files]
```

</div>

עם `-v`, שש פונקציות בדיקה, וחמש מהן עם שתי תת־בדיקות (`Check` ו־`CheckPool`):

<div dir="ltr" align="left">

```text
--- PASS: TestCheckClassifies (0.84s)
--- PASS: TestCheckBoundsConcurrency (0.75s)
--- PASS: TestCheckKeepsInputOrder (0.65s)
--- PASS: TestCheckFailFast (0.44s)
--- PASS: TestCheckParentCancel (0.63s)
--- PASS: TestWriteReport (0.00s)
```

</div>

וההרצה של צעד 16, בצורה הזאת. קוד היציאה הוא 1, כי יש קישורים שאינם תקינים; ברשת שלך `10.255.255.1` יכולה
להיות `error` במקום `timeout`:

<div dir="ltr" align="left">

```text
ok        200  https://go.dev/
broken    404  https://go.dev/this-page-does-not-exist
ok        200  https://example.com/
error       -  http://127.0.0.1:9/
timeout     -  https://10.255.255.1/
error       -  not a url
checked 6, ok 2, not ok 4
exit: 1
```

</div>

## רשימת בדיקה

- [ ] הפונקציה `checkOne` גוזרת context עם `context.WithTimeout`, ובונה את הבקשה עם
  `http.NewRequestWithContext`.
- [ ] כשה־context נגמר, `checkOne` שומרת את `ctx.Err()`.
- [ ] `Check` משתמשת ב־`errgroup.WithContext` וב־`SetLimit`, וכל goroutine כותבת רק לתא שלה.
- [ ] הפונקציה `CheckPool` סוגרת את `jobs` במחולל, ואת `out` רק אחרי `wg.Wait()`.
- [ ] האוסף ב־`CheckPool` קורא עד הסגירה.
- [ ] ראית את דוח המרוץ של `racedemo` ואת הדוח על השדה ששתלת בקוד שלך.
- [ ] ראית את דוח הדליפה השתולה, והחזרת את הלולאה המקורית.
- [ ] הפקודה `go test -race -count=3 ./...` ירוקה, ו־`gofmt -l .` לא מדפיסה דבר.

## טעויות נפוצות ותיקונים

- **השגיאה `-race requires cgo; enable cgo by setting CGO_ENABLED=1`.** חסר מהדר C ב־WSL2, ולכן cgo כבוי.
  התקיני את `build-essential` (צעד 2) ופתחי טרמינל חדש.
- **ביום 8, `go test -race .` נופלת על `panic: TODO 3: CheckPool`.** זה צפוי: `CheckPool` נכתבת ביום 9.
  עד אז מריצים עם המסנן `'TestCheckClassifies/^Check$'`.
- **הבקשה האיטית מסווגת כשבורה או כשגיאה, ולא כזמן קצוב או ביטול.** שמרת את השגיאה של הלקוח, שהיא הסיבה
  של הביטול עטופה. בדקי קודם את `ctx.Err()`, כמו בהערה של `TODO 1`.
- **הבדיקות נתקעות עד שהן נכשלות על timeout אחרי עשר דקות.** כמעט תמיד `jobs` לא נסגר: ה־workers מחכים
  ב־`range` לנצח, ו־`wg.Wait()` איתם. ודאי שבמחולל יש `defer close(jobs)`, ושהוא חוזר כש־`ctx.Done()`
  נסגר.
- **`panic: send on closed channel`.** סגרת את `out` מתוך worker, או לפני ש־`wg.Wait()` חזרה. רק
  ה־goroutine שמחכה לכולם סוגרת אותו.
- **`max in flight = 20, want at most 3`.** ה־errgroup נוצר בלי `SetLimit`. הקריאה ל־`SetLimit` באה מיד
  אחרי `WithContext`, לפני ה־`Go` הראשון (שינוי מגבלה בזמן שיש goroutines פעילות זורק panic).

## רמזים

- ב־`Verdict`, סדר הענפים חשוב: `ErrSkipped` לפני הבדיקות של context, ו־`ErrBroken` אחרונה.
- השגיאות שאת יוצרת עוטפות עם `%w`, כדי ש־`errors.Is` יעבוד.
- את `Elapsed` מודדים סביב `Do` בלבד: לפני השורה `start := time.Now()` ואחריה `time.Since(start)`.
- ב־`CheckPool`, הטיפוסים `job` ו־`done` יכולים להיות מוגדרים בתוך הפונקציה. האינדקס עובר יחד עם הכתובת
  ועם התוצאה, וזה כל הסוד של שמירת הסדר.
- אם בדיקת המקביליות נכשלת עם `max in flight = 1`, הבקשות רצות אחת אחרי השנייה: אולי `Wait` נקראת בתוך
  הלולאה, או ש־`checkOne` נקראת לפני ה־`go`.
- כדי לראות את הדליפה מתרחשת בזמן אמת, הוסיפי לבדיקה שורה זמנית עם `runtime.NumGoroutine()` לפני ואחרי
  הקריאה, והדפיסי את שניהם עם `t.Log`.

## הפתרון

הפתרון נמצא בתיקייה `solution/`, באותו מבנה כמו נקודת ההתחלה; הקבצים שמילאת הם `linkcheck.go`
ו־`pool.go`. **למה הוא עובד:** הסדר נשמר בלי שום נעילה, כי כל goroutine כותבת לתא אחר ב־slice שהוקצה מראש, ו־`Wait`
(של `errgroup` או של `WaitGroup`) היא נקודת הסנכרון שאחריה כל הכתיבות נראות. הביטול עובר בשני כיוונים דרך
אותו עץ contexts: ההורה (Ctrl-C) מבטל הכול, ו־fail-fast מבטל את הילד המשותף. כל goroutine ב־`CheckPool`
שייכת למשהו שממתין לה, ולכן בדיקת הדליפות שקטה.

## מה התרגיל מדגים

- מקביליות חסומה בשתי צורות: `errgroup.SetLimit` מול worker pool, עם אותן בדיקות לשתיהן.
- זמן קצוב לכל בקשה וגבול לכל הריצה, דרך context שמגיע עד הלקוח של `net/http`.
- ההבדל בין `ctx.Err()` ל־`context.Cause`, כפי שהוא מופיע בקוד אמיתי.
- בדיקות מקביליות דטרמיניסטיות מול שרת `httptest` מקומי, בלי אינטרנט ובלי `time.Sleep` ארוכים.
- גלאי המרוצים ובדיקת דליפות כחלק קבוע מהבדיקות, לא ככלי חירום.

</div>
