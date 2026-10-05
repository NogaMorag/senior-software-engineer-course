---
direction: rtl
lang: he
---

<div dir="rtl" lang="he">

# תרגיל פרק 16: הגבלת קצב, לחץ חוזר והשלת עומס

**מטרה:** לכתוב דלי אסימונים עם שעון מוזרק, מגביל קצב מבוזר בסקריפט Lua ב־Redis, תור חסום עם שתי מדיניות,
ושכבה שמשילה עומס לפי מקביליות ועדיפות. ואז למדוד בעצמך, עם כלי עומס קטן, מה קורה לשירות כשמגיע
אליו יותר ממה שהוא מסוגל לשרת, עם השלה ובלעדיה.

**זמן משוער:** שני ימי לימוד. יום 48: כ־65 דקות בנייה (דלי אסימונים, מגביל ב־Redis, סימולציית תורים). יום
49: כ־55 דקות בנייה (תור חסום, השלת עומס, מדידות עם כלי העומס).

**ידע מקדים:** פרק 16 ופרק 15. מפרק 12: Redis בקונטיינר ו־go-redis. מפרק 5: channels, בחירה עם `select`, וביטול.
מפרק 2: Docker ו־compose. התרגיל לא מייבא קוד מתרגילים קודמים.

## מה בתיקייה הזאת

| קובץ | תפקיד |
| --- | --- |
| `starter/ratelimit/bucket.go` | הדלי בזיכרון; `AllowN` עם `TODO(day 48)` |
| `starter/ratelimit/local.go` | דלי לכל מפתח בזיכרון, והממשק `Limiter`; מוכן |
| `starter/ratelimit/tokenbucket.lua` | הסקריפט שרץ ב־Redis; הגוף עם `TODO(day 48)` |
| `starter/ratelimit/redis.go` | המגביל המבוזר; `Allow` עם `TODO(day 48)` |
| `starter/ratelimit/middleware.go` | שכבת HTTP: 429, ‏`Retry-After` ושדות `RateLimit`; מוכנה |
| `starter/bounded/queue.go` | תור חסום; `Put` ו־`TryPut` עם `TODO(day 49)` |
| `starter/shed/shed.go` | השלת עומס עם עדיפויות; `Wrap` עם `TODO(day 49)` |
| `starter/queuesim/sim.go` | סימולציית תור בזמן מדומה; מוכנה |
| `starter/cmd/queuesim/main.go` | טבלת זמן תגובה מול ניצולת |
| `starter/cmd/pipeline/main.go` | יצרן מהיר מצרכן, בשלוש מדיניות תור |
| `starter/cmd/server/main.go` | שירות עם קיבולת קבועה, ודגלים להשלה ולהגבלת קצב |
| `starter/cmd/load/main.go` | כלי עומס במודל פתוח, עם אחוזונים לכל קוד תשובה |
| `data/compose.yaml` | ‏Redis לפרק, בפרויקט compose בשם `ch16`, על הפורט 6316 |
| `tests/` | הבדיקות הקנוניות, באותם נתיבים; עותקים זהים ב־`starter/` וב־`solution/` |
| `solution/` | הפתרון המלא |

בדיקות Redis קוראות את המשתנה `REDIS_ADDR`. כשהוא ריק הן מדלגות עם הודעה, וכל השאר רצות בלי Docker.

## יום 48: דלי אסימונים ומגביל ב־Redis

**1. העתיקי את קוד ההתחלה ואת קובץ ה־compose.**

<div dir="ltr" align="left">

```bash
mkdir -p ~/senior-course/ch16/data
cp -R "$COURSE/exercises/ch16-ratelimit/starter/." ~/senior-course/ch16/
cp "$COURSE/exercises/ch16-ratelimit/data/compose.yaml" ~/senior-course/ch16/data/
cd ~/senior-course/ch16
go build ./... && go vet ./... && echo builds
```

</div>

**2. סימולציית התורים.** היא מוכנה, ואין בה מה למלא. הריצי אותה והעתיקי את הטבלה לפתק:

<div dir="ltr" align="left">

```bash
go run ./cmd/queuesim
go test -race ./queuesim/
```

</div>

**3. מלאי את `AllowN`** בקובץ `ratelimit/bucket.go`, לפי ההערה שבגוף הפונקציה. בדיקות הדלי, ההשוואה
ל־`x/time/rate` ושכבת ה־HTTP רצות מיד, ובדיקות Redis מדלגות:

<div dir="ltr" align="left">

```bash
go test -race -v -run 'TestBucket|TestClock|TestLocal|TestSameAnswers|TestMiddleware|TestRetryAfter' ./ratelimit/
```

</div>

**4. הפעילי את Redis.** הדגל `--wait` מחכה לבדיקת הבריאות:

<div dir="ltr" align="left">

```bash
docker compose -p ch16 -f data/compose.yaml up -d --wait
docker compose -p ch16 -f data/compose.yaml exec redis redis-cli ping
```

</div>

<div dir="ltr" align="left">

```text
PONG
```

</div>

**5. הריצי את הגרסה הנאיבית** לפני שאת כותבת את הסקריפט, ורשמי את המספר:

<div dir="ltr" align="left">

```bash
REDIS_ADDR=127.0.0.1:6316 go test -race -v -run Naive ./ratelimit/
```

</div>

**6. מלאי את הסקריפט** בקובץ `ratelimit/tokenbucket.lua`, לפי ששת הצעדים שבהערה. אחר כך מלאי את `Allow`
בקובץ `ratelimit/redis.go`. הסקריפט נטען לתוך הבינארי בזמן הבנייה, ולכן כל שינוי בו דורש הרצה מחדש של הבדיקות, לא
יותר:

<div dir="ltr" align="left">

```bash
REDIS_ADDR=127.0.0.1:6316 go test -race -v -run 'Redis' ./ratelimit/
```

</div>

**7. העומס הראשון, בלי שום הגנה.** בטרמינל אחד:

<div dir="ltr" align="left">

```bash
cd ~/senior-course/ch16
go run ./cmd/server
```

</div>

ובטרמינל שני. העתיקי לפתק את הטבלה של כל הרצה:

<div dir="ltr" align="left">

```bash
cd ~/senior-course/ch16
go run ./cmd/load -rate 150 -for 5s
go run ./cmd/load -rate 300 -for 5s
```

</div>

עצרי את השרת עם Ctrl+C, וחכי שהשורה האחרונה שלו תודפס: כיבוי מסודר מחכה לבקשות שבדרך.

**8. הגבלת קצב לכל שוכרת, ב־Redis.** הפעילי את השרת עם מגביל, ושלחי עומס משתי שוכרות במקביל:

<div dir="ltr" align="left">

```bash
go run ./cmd/server -rate 50 -burst 20 -redis 127.0.0.1:6316
```

</div>

<div dir="ltr" align="left">

```bash
go run ./cmd/load -rate 100 -for 5s -H "X-Tenant: a" &
go run ./cmd/load -rate 20 -for 5s -H "X-Tenant: b"
wait
```

</div>

רשמי כמה 429 קיבלה כל שוכרת. אחר כך הריצי פקודה אחת עם הכותרות, וראי את השדות שהשרת מחזיר:

<div dir="ltr" align="left">

```bash
curl -si -H "X-Tenant: c" localhost:8016/work | grep -iE '^(HTTP|ratelimit|retry-after)'
```

</div>

## יום 49: לחץ חוזר, השלת עומס ותורים חסומים

**9. מלאי את `Put` ואת `TryPut`** בקובץ `bounded/queue.go`, והריצי:

<div dir="ltr" align="left">

```bash
go test -race ./bounded/
go run ./cmd/pipeline
```

</div>

העתיקי את טבלת היצרן והצרכן לפתק.

**10. מלאי את `Wrap`** בקובץ `shed/shed.go`. הבדיקה הראשונה מראה בקשה מעל המגבלה שנדחית מיד:

<div dir="ltr" align="left">

```bash
go test -race -v ./shed/
```

</div>

**11. אותו עומס, עם השלה.** הפעילי מחדש את השרת עם מגבלה של שמונה, ושלחי את אותו עומס כמו בצעד 7:

<div dir="ltr" align="left">

```bash
go run ./cmd/server -shed 8
```

</div>

<div dir="ltr" align="left">

```bash
go run ./cmd/load -rate 300 -for 5s
curl -s localhost:8016/stats
```

</div>

**12. עדיפויות.** הפעילי את השרת עם שני מקומות שמורים, ושלחי עומס בשתי עדיפויות במקביל:

<div dir="ltr" align="left">

```bash
go run ./cmd/server -shed 8 -reserved-high 2
```

</div>

<div dir="ltr" align="left">

```bash
go run ./cmd/load -rate 200 -for 5s > low.txt &
go run ./cmd/load -rate 100 -for 5s -H "X-Priority: high"
wait
cat low.txt
curl -s localhost:8016/stats
```

</div>

רשמי כמה נדחו מכל עדיפות, והשווי את זמני התגובה לצעד 7.

**13. כל הבדיקות, עם Redis, ושלושת הכלים. ואז כבי את Redis ומחקי את הנתונים שלו.**

<div dir="ltr" align="left">

```bash
REDIS_ADDR=127.0.0.1:6316 go test -race ./...
gofmt -l .
go vet ./...
staticcheck ./...
docker compose -p ch16 -f data/compose.yaml down -v
```

</div>

**14. שמרי את העבודה.**

<div dir="ltr" align="left">

```bash
cd ~/senior-course
git add ch16
git commit -m "ch16: token bucket, Redis limiter, bounded queue and load shedding"
```

</div>

## פלט צפוי

בדיקות Redis בצעד 6, אחרי שהסקריפט עובד (הזמנים אצלך יהיו אחרים):

<div dir="ltr" align="left">

```text
=== RUN   TestRedisTwoInstancesShareOneLimit
--- PASS: TestRedisTwoInstancesShareOneLimit (0.01s)
=== RUN   TestRedisRefills
--- PASS: TestRedisRefills (0.09s)
=== RUN   TestRedisIsAtomicUnderConcurrency
--- PASS: TestRedisIsAtomicUnderConcurrency (0.03s)
=== RUN   TestRedisSurvivesScriptFlush
--- PASS: TestRedisSurvivesScriptFlush (0.01s)
PASS
```

</div>

השורה של הגרסה הנאיבית בצעד 5 נראית כך, עם המספר שלך במקום הקו:

<div dir="ltr" align="left">

```text
    redis_test.go:148: naive read-then-write allowed ___ of 100 with a limit of 10
```

</div>

בצעד 8, בקשה של שוכרת חדשה מחזירה את שדות המגבלה:

<div dir="ltr" align="left">

```text
HTTP/1.1 200 OK
Ratelimit: "default";r=19
Ratelimit-Policy: "default";q=20;w=1
```

</div>

ובצעד 13, כל החבילות:

<div dir="ltr" align="left">

```text
ok  	example.com/ch16-ratelimit/bounded	1.343s
ok  	example.com/ch16-ratelimit/cmd/load	1.967s
?   	example.com/ch16-ratelimit/cmd/pipeline	[no test files]
?   	example.com/ch16-ratelimit/cmd/queuesim	[no test files]
?   	example.com/ch16-ratelimit/cmd/server	[no test files]
ok  	example.com/ch16-ratelimit/queuesim	3.564s
ok  	example.com/ch16-ratelimit/ratelimit	2.499s
ok  	example.com/ch16-ratelimit/shed	1.766s
```

</div>

את הטבלאות של הסימולציה, של היצרן והצרכן ושל כלי העומס הפרק מציג עם מקומות ריקים: הן שלך.

## רשימת בדיקה

- [ ] בדיקות הדלי עם השעון המזויף עוברות בלי לישון, כולל מילוי, פרץ ודלי ריק.
- [ ] ההשוואה ל־`x/time/rate` עוברת על חמשת אלפים אירועים.
- [ ] כל בדיקות Redis רצו מול Redis אמיתי (לא דילגו), כולל שני לקוחות שחולקים מגבלה.
- [ ] בקשה מעל מגבלת המקביליות נדחית מיד עם 503 ו־`Retry-After`.
- [ ] בפתק: טבלת הסימולציה, טבלת היצרן והצרכן, מספר החריגה של הגרסה הנאיבית, ותוצאות כלי העומס עם השלה ובלעדיה.
- [ ] ‏Redis כבוי והנתונים שלו נמחקו.
- [ ] העבודה שמורה ב־Git.

## טעויות נפוצות ותיקונים

- **בדיקות Redis מדווחות `SKIP`.** המשתנה `REDIS_ADDR` לא הוגדר באותה פקודה. כתבי אותו בתחילת השורה, כמו
  בצעדים 5 ו־6, או ייצאי אותו לפני כן.
- **בדיקות Redis נכשלות עם `TODO(day 48): tokenbucket.lua` או עם `Error compiling script`.** הסקריפט עדיין
  מחזיר את שגיאת ה־TODO, או שיש בו שגיאת תחביר. ההודעה של Redis כוללת את מספר השורה, למשל `user_script:1`.
- **`TestRedisRefills` נכשלת: הבקשה אחרי 80 מילישניות נדחית.** הזמן נשמר במיקרו־שניות, או הומר למחרוזת בלי
  `tostring`, ואיבד דיוק. עבדי באלפיות שנייה, כמו בהערה.
- **המקרה `more than burst is never allowed` נכשל עם זמן חזרה שאינו אפס.** בקשה גדולה מהדלי לא תצליח לעולם,
  ולכן אין לה זמן חזרה. בדקי אותה לפני חישוב ההמתנה, והחזירי רק את מספר האסימונים שנשארו.
- **`TestNeverMoreThanTheLimitInFlight` נכשלת לפעמים, עם שיא גבוה מ־5.** הבדיקה קוראת את המונה ואחר כך מוסיפה. הוסיפי קודם
  עם `Add(1)`, ובדקי את הערך שהוחזר.
- **השרת החדש נכשל עם `address already in use`.** הקודם עוד מסיים כיבוי מסודר. חכי לשורה האחרונה שלו, או בדקי
  מי מחזיק את הפורט עם `lsof -i :8016`.

## רמזים

- בדלי, `elapsed.Seconds()` מחזירה מספר עשרוני, ו־`math.Min` מגבילה לתקרה.
- בסקריפט, `redis.call('TIME')` מחזירה טבלה של שתי מחרוזות: שניות ומיקרו־שניות. ‏`tonumber` הופכת כל אחת למספר.
- מערך שהסקריפט מחזיר מגיע ל־Go כרשימה של מספרים שלמים, ולכן `Int64Slice` מתאימה לו.
- בשכבת ההשלה, העדיפות הנמוכה רואה מגבלה קטנה יותר: `Limit` פחות `ReservedForHigh`.
- בחירה עם ענף `default` לא מחכה אף פעם. בלי ענף כזה, היא מחכה לאחד הענפים האחרים.

## הפתרון

הפתרון המלא נמצא בתיקייה `solution/`. הדלי מחשב את המילוי רק כשמגיעה בקשה, לפי השעון שהוזרק, ולכן אין לו טיימר
ובדיקה של דקה שלמה לוקחת אפס זמן. הסקריפט ב־Redis עושה קריאה, מילוי, לקיחה וכתיבה בצעד אטומי אחד עם השעון של
Redis, כך שעותקים רבים חולקים דלי אחד בלי מרוץ ובלי בעיית שעונים, והספרייה מטפלת במטמון הסקריפטים הנדיף.
שכבת ההשלה מוסיפה למונה לפני הבדיקה, ולכן אף פעם לא עוברת את המגבלה, והמקום השמור לעדיפות גבוהה הוא פשוט
מגבלה קטנה יותר לעדיפות הנמוכה.

## מה התרגיל מדגים

- תורת תורים בזמן מדומה: חוק ליטל, ופיצוץ זמן התגובה ליד הרוויה.
- דלי אסימונים עם שעון מוזרק, ובדיקה שמשווה אותו לספרייה מוכנה על אותו ציר זמן.
- הגבלת קצב מבוזרת בסקריפט Lua אטומי, עם השעון של Redis ותפוגה לכל מפתח, מול גרסה נאיבית שחורגת.
- תשובת 429 עם `Retry-After` ושדות המגבלה מהטיוטה, ובחירה בין פתוח לסגור בכשל.
- לחץ חוזר בתור חסום: חסימה מול סירוב, מול תור שגדל בלי סוף.
- השלת עומס לפי מקביליות עם עדיפויות, נמדדת עם כלי עומס במודל פתוח.
- חבילות הגבלת הקצב והשלת העומס נכנסות לאבן הדרך M6 של Hookline.

</div>
