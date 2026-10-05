---
direction: rtl
lang: he
---

<div dir="rtl" lang="he">

# תרגיל פרק 15: מעבדת עמידות

**מטרה:** לבנות מעבדה שבה שרת משתבש בכוונה, ולקוח שכל מנגנון עמידות בו נדלק לחוד: העברת תקציב זמן לאורך
שרשרת, ניסיון חוזר עם המתנה גדלה ופיזור אקראי, תקציב ניסיונות, מפסק זרם ומחיצה אטומה. ואז להריץ את המעבדה עם
כל מנגנון ובלעדיו, ולרשום את המספרים שלך.

**זמן משוער:** שני ימי לימוד. יום 46: כ־35 דקות בנייה (העברת תקציב, ניסיון חוזר, הרצת הסערה). יום 47: כ־65
דקות בנייה (תקציב ניסיונות, מחיצה, מפסק זרם, כל המעבדה).

**ידע מקדים:** פרק 15. מפרק 5: ביטול עם context, ובדיקת מרוצים עם `-race`. מפרק 6: middleware ו־`httptest`. התרגיל לא מייבא
קוד מתרגילים קודמים, ואין לו תלויות חיצוניות.

## מה בתיקייה הזאת

| קובץ | תפקיד |
| --- | --- |
| `starter/flaky/flaky.go` | השרת שמשתבש בכוונה: זמן תגובה, שיעור כשלים, ומונים; מוכן |
| `starter/deadline/deadline.go` | העברת תקציב בכותרת; שלוש פונקציות עם `TODO(day 46)` |
| `starter/retry/retry.go` | המתנה גדלה והלולאה עם `TODO(day 46)`; תקציב הניסיונות עם `TODO(day 47)` |
| `starter/bulkhead/bulkhead.go` | מחיצה אטומה כסמפור; `Acquire` ו־`Release` עם `TODO(day 47)` |
| `starter/breaker/breaker.go` | מפסק זרם; הטיפוסים מוכנים, ארבע מתודות עם `TODO(day 47)` |
| `starter/degrade/degrade.go` | ערך ישן עם גבול, ותכונה שאפשר לכבות; מוכן, לקריאה |
| `starter/hop/hop.go` | קפיצה בשרשרת: הלקוח שמרכיב את כל המנגנונים, וה־handler; מוכן |
| `starter/cmd/upstream/main.go` | השרת המשתבש כתוכנית, עם דגלים |
| `starter/cmd/client/main.go` | לקוח עם דגל לכל מנגנון |
| `starter/cmd/lab/main.go` | חמשת התרחישים של הפרק, כל אחד עם המנגנון ובלעדיו |
| `tests/` | הבדיקות הקנוניות, באותם נתיבים; עותקים זהים נמצאים ב־`starter/` וב־`solution/` |
| `solution/` | הפתרון המלא |

המעבדה מדלגת בעצמה על תרחיש שהמנגנון שלו עוד לא נבנה, ולכן אפשר להריץ אותה כבר ביום 46.

## יום 46: תקציב זמן, ניסיון חוזר וסערה

**1. העתיקי את קוד ההתחלה.** הבדיקות כבר בפנים.

<div dir="ltr" align="left">

```bash
mkdir -p ~/senior-course/ch15
cp -R "$COURSE/exercises/ch15-resilience-lab/starter/." ~/senior-course/ch15/
cd ~/senior-course/ch15
go build ./... && go vet ./... && echo builds
```

</div>

<div dir="ltr" align="left">

```text
builds
```

</div>

**2. קראי את `flaky/flaky.go` ואת `hop/hop.go`.** בראשון, שימי לב איך השרת מזהה בקשה נטושה. בשני, עקבי אחרי
`Call`: לולאת הניסיונות, ובתוכה ניסיון בודד, ובתוכו השליחה.

**3. מלאי את שלוש הפונקציות בקובץ `deadline/deadline.go`**, לפי ההערות שבגוף כל אחת. אחר כך:

<div dir="ltr" align="left">

```bash
go test -race ./deadline/
```

</div>

**4. מלאי את `Backoff` ואת `Do`**, בקובץ `retry/retry.go`. תקציב הניסיונות שייך ליום 47, ולכן מריצים רק את
הבדיקות שלא נוגעות בו:

<div dir="ltr" align="left">

```bash
go test -race -run 'TestBackoff|TestDoStops|TestDoRespects|TestDo$/^(success|attempts|not)' ./retry/
```

</div>

**5. הבדיקה שמראה קריאה שנחתכת כשהתקציב נגמר**, ושתי בדיקות השרשרת שלא צריכות את יום 47:

<div dir="ltr" align="left">

```bash
go test -race -run 'TestBudgetCutsOff|TestStatusCodes' -v ./hop/
go test -race -run 'TestRetryStormAmplifies/^(three_attempts_at|one_attempt)' -v ./hop/
```

</div>

**6. הריצי את הסערה ואת העברת התקציב**, והעתיקי את שתי הטבלאות לפתק:

<div dir="ltr" align="left">

```bash
go run ./cmd/lab -scenario storm
go run ./cmd/lab -scenario deadline
```

</div>

**7. השרת והלקוח כתוכניות נפרדות.** בטרמינל אחד:

<div dir="ltr" align="left">

```bash
cd ~/senior-course/ch15
go run ./cmd/upstream -error-rate 0.3 -latency 20ms
```

</div>

ובטרמינל שני, פעם בלי ניסיונות ופעם עם שלושה:

<div dir="ltr" align="left">

```bash
cd ~/senior-course/ch15
go run ./cmd/client -n 200 -c 10
go run ./cmd/client -n 200 -c 10 -attempts 3
curl -s localhost:8015/counts
```

</div>

רשמי כמה בקשות הצליחו בכל הרצה, וכמה בקשות השרת קיבל בסך הכול. עצרי את השרת עם Ctrl+C.

## יום 47: תקציב ניסיונות, מחיצה ומפסק זרם

**8. מלאי את `Deposit` ואת `Withdraw`** בקובץ `retry/retry.go`, והריצי את כל בדיקות החבילה:

<div dir="ltr" align="left">

```bash
go test -race ./retry/
```

</div>

**9. מלאי את `Acquire` ואת `Release`** בקובץ `bulkhead/bulkhead.go`:

<div dir="ltr" align="left">

```bash
go test -race ./bulkhead/
```

</div>

**10. מלאי את המפסק** בקובץ `breaker/breaker.go`, בסדר הזה:

- המתודה `setState`, שמחליפה מצב ופותחת דור חדש;
- המתודה `current`, שמעבירה מפתוח לחצי־פתוח;
- המתודה `Allow`, שמחליטה אם קריאה יוצאת;
- המתודה `record`, שרושמת את התוצאה.

הבדיקות משתמשות בשעון מזויף, ולכן הן לא ישנות אף פעם:

<div dir="ltr" align="left">

```bash
go test -race -v ./breaker/
```

</div>

**11. כל הבדיקות, ושלושת הכלים.** הבדיקה של תרחישי המעבדה לוקחת כמה שניות:

<div dir="ltr" align="left">

```bash
go test -race ./...
gofmt -l .
go vet ./...
staticcheck ./...
```

</div>

**12. המעבדה כולה.** ההרצה לוקחת כמה עשרות שניות. העתיקי לפתק את כל הטבלאות:

<div dir="ltr" align="left">

```bash
go run ./cmd/lab | tee lab-$(date +%F).txt
```

</div>

**13. הלקוח עם כל המנגנונים**, מול השרת מצעד 7:

<div dir="ltr" align="left">

```bash
go run ./cmd/client -n 200 -c 10 -attempts 3 -budget 0.1 -breaker -bulkhead 5
```

</div>

**14. שמרי את העבודה.**

<div dir="ltr" align="left">

```bash
cd ~/senior-course
git add ch15
git commit -m "ch15: resilience lab with deadlines, retry budget, breaker and bulkhead"
```

</div>

## פלט צפוי

בסוף יום 46, טבלת הסערה. שלוש השורות הראשונות הן חשבון, ולכן הן אצלך בדיוק כך; הרביעית מחכה ליום 47:

<div dir="ltr" align="left">

```text
== storm: 100 user requests, 3 layers, upstream error rate 1.00
config                             A    B    C    upstream  per request  user ok
no retries                         100  100  100  100       1.0          0
3 attempts at every layer          100  300  900  2700      27.0         0
3 attempts at the top only         100  300  300  300       3.0          0
3 attempts + 10% budget per layer  skipped: retry.Budget not built yet
```

</div>

ביום 47 השורה הרביעית מתמלאת במספרים שלך, ושאר התרחישים מדפיסים טבלאות באותו מבנה. בטבלאות התזמון (העברת
תקציב, מפסק, מחיצה) המספרים משתנים בין הרצות ובין מכונות; את התבנית מתאר הפרק. והבדיקות, בסוף יום 47 (הזמנים
אצלך יהיו אחרים):

<div dir="ltr" align="left">

```text
ok  	example.com/ch15-resilience-lab/breaker	1.308s
ok  	example.com/ch15-resilience-lab/bulkhead	1.673s
?   	example.com/ch15-resilience-lab/cmd/client	[no test files]
ok  	example.com/ch15-resilience-lab/cmd/lab	4.984s
?   	example.com/ch15-resilience-lab/cmd/upstream	[no test files]
ok  	example.com/ch15-resilience-lab/deadline	2.249s
ok  	example.com/ch15-resilience-lab/degrade	1.953s
?   	example.com/ch15-resilience-lab/flaky	[no test files]
ok  	example.com/ch15-resilience-lab/hop	3.130s
ok  	example.com/ch15-resilience-lab/retry	2.880s
```

</div>

## רשימת בדיקה

- [ ] כל הבדיקות עוברות עם `-race`, ושלושת הכלים שקטים.
- [ ] הבדיקה `TestBudgetCutsOffTheCallWhenItRunsOut` עוברת, ואת יודעת להסביר למה הלקוחה קיבלה 504 או גבול זמן.
- [ ] טבלת הסערה שמורה בפתק, כולל שורת התקציב מההרצה שלך.
- [ ] טבלאות העברת התקציב, המפסק, המחיצה וההידרדרות שמורות בפתק, עם התאריך.
- [ ] בדיקות המפסק עוברות בפחות משנייה, למרות שהן „מחכות” עשר שניות.
- [ ] העבודה שמורה ב־Git.

## טעויות נפוצות ותיקונים

- **המפסק לא נפתח לעולם, ו־`TestBreakerStopsCallingADeadUpstream` נכשלת עם 100 פניות.** בשינוי של `attempt`
  כתבת `:=` בתוך בלוק התנאי, ומשתנה השגיאה החדש הסתיר את הערך המוחזר בעל השם. השתמשי בשם אחר לשגיאה של
  `Allow`, כמו בקוד המקורי.
- **הבדיקות של יום 46 נופלות עם `TODO(day 47): Budget.Deposit`.** הרצת בדיקה שמשתמשת בתקציב. הריצי את
  הפקודות של צעדים 4 ו־5 כמו שהן, עם הסינון.
- **`TestDoStopsBeforeTheDeadline` מראה שתי שינות במקום אחת.** בדקת את גבול הזמן מול `time.Now` ולא מול
  `p.Now()`, ולכן השעון המזויף לא נלקח בחשבון. השתמשי בשעון שבמדיניות.
- **`TestStaleResultIsIgnored` נכשלת: המפסק נפתח מתוצאה ישנה.** המתודה `record` לא משווה את הדור, או ש־`setState` לא
  מעלה אותו. כל מעבר מצב חייב להעלות את המונה.
- **`go test` נתקע ב־`TestFullFailsFast`.** המתודה `Acquire` מחכה בלי טיימר, או בלי `ctx.Done()`. ודאי שיש שלושה ענפים
  ב־`select` השני.
- **שגיאה `address already in use` בצעד 7.** שרת קודם עוד רץ. עצרי אותו עם Ctrl+C, או בחרי פורט אחר עם הדגל
  `-addr`. את הכתובת החדשה מעבירים ללקוח עם `-url`.

## רמזים

- הכותרת מגיעה כמחרוזת. הפונקציה `strconv.ParseInt` מחזירה שגיאה גם על מחרוזת ריקה, וזה בדיוק המקרה של
  כותרת חסרה.
- בפונקציה `Downstream` הזמן שנשאר הוא `time.Until(dl)`. את הגבול החדש בני מ־`dl` עצמו, לא מהשעה הנוכחית.
- הזזה שמאלה של `time.Duration` היא כפל בחזקת שתיים. ערך שלילי אחרי הזזה פירושו גלישה.
- במפסק, `current` רץ תמיד כשהנעילה תפוסה, ולכן הוא לא נועל בעצמו. המתודה `record` כן נועלת, כי קוראים לה מבחוץ.
- ההמתנה במחיצה דומה ל־`checkOne` מפרק 5: בחירה בין הערוץ, טיימר וביטול ה־context.

## הפתרון

הפתרון המלא נמצא בתיקייה `solution/`. התקציב עובר כנקודת זמן מוחלטת בתוך התהליך וכמשך בכותרת בין תהליכים,
ולכן זמן שעבר בכל מקום נספר בלי חישוב נוסף, וכל קפיצה שומרת רזרבה לעצמה. הניסיון החוזר בודק לפני כל המתנה
שהיא לא חוצה את הגבול, ומושך מתקציב משותף לתלות, ולכן גם שלוש שכבות שמנסות שוב לא מכפילות את העומס. המפסק
מתעלם מתוצאות של דור קודם, כך שקריאה איטית לא משנה מצב שכבר השתנה, והשעון המוזרק הופך כל מעבר מצב לבדיקה
מיידית.

## מה התרגיל מדגים

- כשל חלקי מדוד: בקשות נטושות לעומת עבודה שהסתיימה בשביל אף אחד.
- העברת גבול זמן כולל בין שירותים, עם רזרבה בכל קפיצה.
- המתנה גדלה מעריכית עם פיזור אקראי מלא, ולולאת ניסיונות שמכבדת את גבול הזמן.
- תקציב ניסיונות משותף, וסערת ניסיונות חוזרים שנמדדת שכבה אחר שכבה.
- מפסק זרם עם דורות ושעון מוזרק, ובדיקות מצב שלא ישנות.
- מחיצה אטומה כסמפור, והידרדרות מכובדת עם ערך ישן ותכונה כבויה.
- חבילות המפסק והמחיצה נכנסות כמו שהן לאבן הדרך M6 של Hookline.

</div>
