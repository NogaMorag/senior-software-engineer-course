---
direction: rtl
lang: he
---

<div dir="rtl" lang="he">

# תרגיל פרק 6: `shorty`, ונקודת ביקורת 1

**מטרה:** לבנות שירות קיצור כתובות שאפשר להפעיל בייצור: יצירה, פענוח וסטטיסטיקה, מאגר בזיכרון מאחורי
ממשק, תצורה ממשתני סביבה שנבדקת בעלייה, שלוש שכבות middleware (מזהה בקשה, לוג גישה, התאוששות), כיבוי
מסודר, pprof על מאזין פרטי, בדיקות `httptest` ובנצ'מרק. ואז לסגור את חלק א בנקודת ביקורת 1. החבילה
`internal/platform` שתכתבי כאן היא השלד של Hookline באבן הדרך M1.

**זמן משוער:** ארבעה ימי לימוד, ימים 11 עד 14: כ־55 דקות בנייה ביום 11, כ־65 ביום 12, כ־60 ביום 13,
ויום 14 קל לנקודת הביקורת. **ידע מקדים:** פרק 6 כולו; פרקים 3 עד 5, ובמיוחד context ו־`errgroup` (סעיפים 5.4 ו־5.6);
`make` ומהדר C מהחבילה `build-essential` (פרק 2, סעיף 2.3). המודול תלוי רק
ב־`golang.org/x/sync`, נעול ב־`go.mod`. הקוד לא מייבא תרגילים קודמים; נקודת הביקורת מריצה גם את
`logstat`, `ledger` ו־`linkcheck` שלך.

## מה בתיקייה הזאת

| קובץ | תפקיד |
| --- | --- |
| `starter/cmd/shorty/main.go` | `main`: תצורה, לוגר, `signal.NotifyContext`, `app.Run`; מוכן |
| `starter/cmd/loadgen/main.go` | מחולל עומס קטן לסעיף 6.8; מוכן |
| `starter/internal/platform/env.go` | קריאת משתני סביבה עם צבירת שגיאות; מוכן |
| `starter/internal/platform/logging.go` | לוגר JSON שמוסיף את מזהה הבקשה; מוכן |
| `starter/internal/platform/httpjson.go` | `TODO 3`: `DecodeJSON`; ו־`WriteJSON`, `WriteError`, `RespondError` מוכנות |
| `starter/internal/platform/middleware.go` | `Chain` ו־`statusWriter` מוכנים; `TODO 5` עד `TODO 7`: שלוש השכבות |
| `starter/internal/platform/server.go` | `NewServer` מוכן; `TODO 8`: `Serve` עם כיבוי מסודר |
| `starter/internal/platform/debug.go` | `TODO 9`: `DebugHandler` של pprof |
| `starter/internal/links/store.go` | `Link`, `Store`, `MemoryStore`; `TODO 1` בשלוש המתודות |
| `starter/internal/links/code.go` | יצירת קוד אקראי ובדיקות קלט; מוכן |
| `starter/internal/links/handlers.go` | `Register` מוכן; `TODO 2` בשלושת ה־handlers |
| `starter/internal/app/config.go` | `Config` ו־`LoadConfig`; `TODO 4`: הבדיקות שמערבות כמה ערכים |
| `starter/internal/app/app.go` | `NewHandler` ו־`Run`; מוכן |
| `tests/internal/...` | הבדיקות הקנוניות: 22 פונקציות בדיקה ו־`BenchmarkResolve` בשלוש חבילות; עותקים זהים ב־`starter/` וב־`solution/` |
| `solution/` | הפתרון המלא, כולל `Makefile` |

עד שתמלאי אותם, `TODO 5` עד `TODO 9` מחזירים התנהגות ריקה ולא panic: השכבות מעבירות את הבקשה הלאה,
`Serve` מתעלמת מה־context, ו־`DebugHandler` עונה 404. כך השירות רץ כבר ביום 11, והבדיקות של הימים הבאים
אדומות עד שתגיעי אליהן.

## יום 11: נתיבים, JSON, מאגר ותצורה

**1. העתיקי את נקודת ההתחלה** ובדקי שהיא מתקמפלת:

<div dir="ltr" align="left">

```bash
mkdir -p ~/senior-course/ch06
cp -r $COURSE/exercises/ch06-shorty/starter/. ~/senior-course/ch06/
cd ~/senior-course/ch06
go build ./... && go vet ./...
```

</div>

**2. קראי את המבנה** לפי הטבלה של סעיף 6.7: `cmd/shorty/main.go` ואז `internal/app/app.go`, כדי לראות איך
החלקים מתחברים, ורק אחר כך את החבילות.

**3. `TODO 1`: המאגר** (`internal/links/store.go`). ההערות בקוד מפרטות כל מתודה. הבדיקה המקבילית רצה עם
`-race` ומוודאת שאף hit לא הולך לאיבוד:

<div dir="ltr" align="left">

```bash
go test -race -run 'TestMemoryStore' ./internal/links/
```

</div>

**4. `TODO 3`: `DecodeJSON`** (`internal/platform/httpjson.go`), לפי סעיף 6.2:

<div dir="ltr" align="left">

```bash
go test -race -run 'TestDecodeJSON|TestRespondError' ./internal/platform/
```

</div>

**5. `TODO 2`: שלושת ה־handlers** (`internal/links/handlers.go`). ואז כל הבדיקות של החבילה:

<div dir="ltr" align="left">

```bash
go test -race ./internal/links/
```

</div>

**6. `TODO 4`: הבדיקות של התצורה** (`internal/app/config.go`), לפי סעיף 6.3:

<div dir="ltr" align="left">

```bash
go test -race -run 'TestLoadConfig' ./internal/app/
```

</div>

**7. הפעלה ידנית.** קודם בלי תצורה, כדי לראות את הסירוב, ואז עם ערך שגוי:

<div dir="ltr" align="left">

```bash
go run ./cmd/shorty
SHORTY_BASE_URL=localhost SHORTY_SHUTDOWN_TIMEOUT=ten go run ./cmd/shorty
```

</div>

שתי ההרצות יוצאות עם `exit status 1` והודעה שמונה כל בעיה. עכשיו כמו שצריך, בטרמינל אחד:

<div dir="ltr" align="left">

```bash
SHORTY_BASE_URL=http://localhost:8080 go run ./cmd/shorty
```

</div>

ובטרמינל שני, יצירה, פענוח, סטטיסטיקה, ומתודה שגויה:

<div dir="ltr" align="left">

```bash
curl -s -X POST localhost:8080/v1/links -d '{"url": "https://go.dev/", "code": "gohome"}'
curl -s -i localhost:8080/gohome | head -3
curl -s localhost:8080/v1/links/gohome
curl -s -i localhost:8080/v1/links | head -2
```

</div>

הפלט המלא נמצא בסעיף „פלט צפוי”. עכשיו לחצי Ctrl-C בטרמינל של השירות: הלוג מדפיס `shutting down`, אבל
התהליך ממשיך לרוץ, כי `Serve` עוד מתעלמת מה־context. לחיצה שנייה הורגת אותו, בזכות השורה עם
`context.AfterFunc` ב־`main`. מחר זה יתוקן. רשמי בפתק במשפט אחד למה התצורה נבדקת בעלייה, ושמרי:

<div dir="ltr" align="left">

```bash
cd ~/senior-course && git add ch06 && git commit -m "Chapter 6: shorty routes, JSON, store, config"
```

</div>

## יום 12: לוגים, כיבוי מסודר ובדיקות

**8. `TODO 5` עד `TODO 7`: שלוש השכבות** (`internal/platform/middleware.go`), לפי סעיפים 6.2 ו־6.4:

<div dir="ltr" align="left">

```bash
cd ~/senior-course/ch06
go test -race -run 'TestRequestID|TestEveryLogLine|TestRecover' -v ./internal/platform/
```

</div>

**9. `TODO 8`: `Serve`** (`internal/platform/server.go`), לפי סעיף 6.5:

<div dir="ltr" align="left">

```bash
go test -race -run 'TestShutdown|TestServe' -v ./internal/platform/
```

</div>

**10. ראי את זה קורה.** הפעילי את השירות כמו בצעד 7, שלחי בקשה אחת, ולחצי Ctrl-C. שלוש השורות האחרונות
בלוג צריכות להיות שורת הגישה של הבקשה, `shutting down` עם הסיבה, ו־`stopped`. בדקי גם שהערך של
`request_id` בשורות של הבקשה זהה לכותרת `X-Request-Id` בתשובה.

**11. כל הבדיקות,** ואז `git commit` כמו בצעד 7:

<div dir="ltr" align="left">

```bash
go test -race ./...
```

</div>

רק `TestDebugHandlerServesPprof` אמורה להיכשל עדיין; היא של מחר.

## יום 13: Makefile, pprof ובנצ'מרק

**12. כתבי `Makefile`** בשורש של `ch06`, עם היעדים `build`, `test`, `lint`, `run`, `bench`, `profile`
ו־`clean`, לפי סעיף 6.7. זכרי: טאב בתחילת כל שורת פקודה. אחר כך השווי לגרסה ב־`solution/Makefile`, ובדקי:

<div dir="ltr" align="left">

```bash
make lint test build
```

</div>

**13. `TODO 9`: `DebugHandler`** (`internal/platform/debug.go`), לפי סעיף 6.8. אחרי זה כל החבילה ירוקה:

<div dir="ltr" align="left">

```bash
go test -race ./...
```

</div>

**14. פרופיל CPU ראשון.** לפי סעיף 6.8, בשלושה טרמינלים: `make run` בראשון; יצירת הקוד `gohome`
ו־`loadgen` למשך עשרים שניות בשני; ובשלישי, בזמן שהעומס רץ:

<div dir="ltr" align="left">

```bash
cd ~/senior-course/ch06
make profile
```

</div>

רשמי בפתק את שורת `Duration` ואת חמש השורות העליונות, ובמשפט אחד: האם בראש יש קוד של `shorty`, של הספרייה
הסטנדרטית, או קריאות מערכת? השורה `requests ..., failed 0` שמדפיס `loadgen` היא המדידה שלך לעומס; רשמי גם
אותה, עם שם המכונה.

**15. הבנצ'מרק:**

<div dir="ltr" align="left">

```bash
make bench
```

</div>

רשמי את `ns/op`, `B/op` ו־`allocs/op`, את השורה `cpu:` ואת התאריך. אחר כך `git commit`.

## יום 14: נקודת ביקורת 1

**16. שכפול נקי.** בדיקה מהמאגר שלך, לא מתיקיות העבודה, מגלה קבצים שלא נשמרו:

<div dir="ltr" align="left">

```bash
rm -rf /tmp/checkpoint1
git clone ~/senior-course /tmp/checkpoint1
cd /tmp/checkpoint1
(cd ch03 && go test ./...)
(cd ch04 && go test ./...)
(cd ch05 && go test -race ./...)
(cd ch06 && make lint test build bench)
```

</div>

**17. הסיור הידני** על השירות מהשכפול הנקי, לפי הרשימה של סעיף 6.10: תצורה חסרה ושגויה (צעד 7), יצירה,
פענוח, סטטיסטיקה ו־405, מזהה בקשה בכל שורה, וכיבוי (צעד 10).

**18. handler שזורק panic.** הוסיפי לרגע ל־`NewHandler` (בקובץ `internal/app/app.go`) נתיב שמפיל את עצמו,
הפעילי, ובקשי אותו פעמיים, ואחר כך בקשה רגילה:

<div dir="ltr" align="left">

```go
mux.HandleFunc("GET /boom", func(w http.ResponseWriter, r *http.Request) { panic("checkpoint") })
```

</div>

<div dir="ltr" align="left">

```bash
curl -s -i localhost:8080/boom | head -1
curl -s localhost:8080/healthz
```

</div>

התשובה הראשונה היא `HTTP/1.1 500 Internal Server Error`, השנייה `ok`, ובלוג יש שורת `panic in handler` עם
המחסנית ועם `request_id`. מחקי את הנתיב.

**19. בעל פה.** עני בקול על שאלות הראיון של פרקים 3 עד 6, ותארי בשתי דקות את חייה של בקשה אחת ב־`shorty`,
לפי התרשים של סעיף 6.10. כל דבר שנכשל היום נרשם ביומן התקיעות, עם התיקון.

## פלט צפוי

הבדיקות של צעד 13:

<div dir="ltr" align="left">

```text
?   	example.com/ch06-shorty/cmd/loadgen	[no test files]
?   	example.com/ch06-shorty/cmd/shorty	[no test files]
ok  	example.com/ch06-shorty/internal/app	1.4s
ok  	example.com/ch06-shorty/internal/links	1.6s
ok  	example.com/ch06-shorty/internal/platform	2.2s
```

</div>

תצורה שגויה (צעד 7):

<div dir="ltr" align="left">

```text
shorty: invalid configuration:
SHORTY_SHUTDOWN_TIMEOUT: "ten" is not a duration (example: 10s)
SHORTY_BASE_URL: "localhost" is not an absolute http or https URL
exit status 1
```

</div>

ארבע הבקשות של צעד 7, אחרי יום 12 (ביום 11 אין עדיין כותרת `X-Request-Id`; הזמן והמזהים אצלך יהיו אחרים):

<div dir="ltr" align="left">

```text
{"code":"gohome","url":"https://go.dev/","short_url":"http://localhost:8080/gohome","hits":0,"created_at":"2026-10-05T08:26:39.243276Z"}
HTTP/1.1 302 Found
Content-Type: text/html; charset=utf-8
Location: https://go.dev/
{"code":"gohome","url":"https://go.dev/","short_url":"http://localhost:8080/gohome","hits":1,"created_at":"2026-10-05T08:26:39.243276Z"}
HTTP/1.1 405 Method Not Allowed
Allow: POST
```

</div>

והלוג של צעד 10, בצורה הזאת:

<div dir="ltr" align="left">

```text
{"time":"...","level":"INFO","msg":"listening","addr":"[::]:8080","base_url":"http://localhost:8080"}
{"time":"...","level":"INFO","msg":"debug listener","addr":"127.0.0.1:6060"}
{"time":"...","level":"INFO","msg":"request","method":"GET","path":"/gohome","pattern":"GET /{code}","status":302,"bytes":38,"duration":47292,"request_id":"e4d4cc2e3e5c6fa0"}
{"time":"...","level":"INFO","msg":"shutting down","cause":"interrupt signal received"}
{"time":"...","level":"INFO","msg":"stopped","error":null}
```

</div>

## רשימת בדיקה

- [ ] בלי `SHORTY_BASE_URL`, או עם ערך שגוי, השירות מסרב לעלות ומדווח על כל הבעיות יחד.
-
  [ ] יצירה מחזירה 201 עם `short_url` וכותרת `Location`; פענוח מחזיר 302; הסטטיסטיקה סופרת; `GET /v1/links` מחזיר 405 עם `Allow`.
- [ ] גוף גדול מ־`SHORTY_MAX_BODY_BYTES` מקבל 413, ושדה לא מוכר מקבל 400.
- [ ] לכל שורות הלוג של בקשה אותו `request_id`, וזה הערך בכותרת התשובה.
- [ ] Ctrl-C מסיים בשורות `shutting down` ו־`stopped`, ובדיקת הכיבוי ירוקה.
- [ ] `/debug/pprof/` זמין רק על `127.0.0.1:6060`, ופרופיל CPU נשמר.
- [ ] `make lint test build bench` ירוק בשכפול נקי, ותוצאות הבנצ'מרק רשומות עם המכונה.

## טעויות נפוצות ותיקונים

- **`code must be 3 to 32 characters`.** קוד מותאם אישית קצר משלושה תווים, כמו `go`. בחרי קוד ארוך יותר,
  או השמיטי את השדה `code` כדי לקבל קוד אקראי.
- **`listen tcp :8080: bind: address already in use`.** שירות קודם עדיין רץ, אולי כי נעצר עם Ctrl-Z ולא
  עם Ctrl-C. הריצי `fg` ואז Ctrl-C, או הפעילי עם `SHORTY_ADDR=:8081` ו־`SHORTY_BASE_URL` מתאים.
- **שורות לוג בלי `request_id`.** ב־handler נכתב `Info` במקום `InfoContext`, או ש־`RequestID` לא החיצונית
  בשרשרת. הבדיקה `TestEveryLogLineCarriesTheRequestID` מראה איזו שורה חסרה.
- **בדיקת הכיבוי נכשלת אחרי חמש שניות עם `is ctx ignored?`.** `Serve` עדיין מחזירה ישר את
  `srv.Serve(ln)`. היא צריכה לחכות ל־`ctx.Done()` ולקרוא ל־`Shutdown`.
- **`panic: pattern "GET /{id}" ... conflicts with pattern "GET /{code}"`.** רשמת נתיב שמתאים בדיוק לאותן
  בקשות כמו נתיב קיים. שני wildcards באותו מקום הם אותה תבנית; תני לנתיב החדש קידומת משלו.
- **`Makefile:13: *** missing separator.  Stop.`** שורת פקודה מתחילה ברווחים. ב־Makefile היא חייבת להתחיל
  בטאב; ב־VS Code בדקי שהקובץ לא מוגדר להמיר טאבים לרווחים.

## רמזים

- ב־`Resolve`, שחררי את `RLock` לפני `e.hits.Add(1)`: המונה אטומי, וה־map לא משתנה. הבדיקה המקבילית תוודא
  שלא איבדת hits.
- ב־`create`, לולאת הניסיונות לקוד אקראי נעצרת ברגע ש־`Create` מחזירה משהו שאינו `ErrExists`, גם `nil`.
  בדקי את `errors.Is` אחרי הלולאה, לא בתוכה.
- ב־`AccessLog`, `r.Pattern` נקרא **אחרי** `next.ServeHTTP`: הנתב ממלא אותו בזמן ההתאמה.
- ב־`Recover`, בדיקת `sw.status == 0` חשובה: אם ה־handler כבר כתב כותרות, ניסיון לכתוב 500 רק ירשום אזהרה
  מ־`net/http`.
- ב־`Serve`, אם הבדיקה `TestShutdownGivesUpAfterGrace` לא רואה `context.DeadlineExceeded`, בדקי שאת עוטפת
  את השגיאה של `Shutdown` עם `%w`.

## הפתרון

`solution/` כולו, ובמיוחד `internal/platform/server.go`, `middleware.go` ו־`httpjson.go`. **למה הוא
עובד:** כל שכבה עושה דבר אחד ותלויה רק בספרייה הסטנדרטית, ולכן אפשר לבדוק כל אחת לבד עם
`httptest.NewRecorder`, ולהעביר את כל `internal/platform` לפרויקט אחר כמו שהיא. הכיבוי בנוי מאותם חלקים
של פרק 5: אות שהופך ל־context, `errgroup` לשני השרתים, channel עם חוצץ שמונע דליפה, ו־context חדש לזמן
החסד. והתצורה נבדקת בפונקציה שמקבלת `lookup`, כך שכל מקרה שגוי הוא שורה בטבלת בדיקות.

## מה התרגיל מדגים

- שירות HTTP עם הספרייה הסטנדרטית בלבד: נתב Go 1.22 ומעלה, middleware כפונקציות, JSON עם גבולות.
- תצורה שנכשלת מוקדם וברור, ולוגים מובנים שקושרים בקשה אחת לכל השורות שלה.
- כיבוי מסודר שנבדק בבדיקה אוטומטית, ולא רק „נראה שעובד”.
- מבנה `cmd` ו־`internal` עם כיוון תלויות שמאפשר לקחת את `platform` לפרויקט הגמר.
- מדידה כנה: פרופיל ובנצ'מרק שאת מריצה, רושמת, ומצמידה למכונה.

</div>
