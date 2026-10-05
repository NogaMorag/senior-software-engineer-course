---
direction: rtl
tags: [senior-course, book]
chapter: 6
---
# שירות HTTP לייצור

- **Handler:** כל מה שעונה לבקשה; כל בקשה ב־goroutine משלה, ולכן מצב משותף מוגן.
- **נתב:** תבנית עם מתודה ו־wildcards; הספציפית מנצחת; 404 ו־405 עם `Allow` מגיעים ממנו.
- **middleware:** מזהה בחוץ, לוג באמצע, התאוששות בפנים.
- **JSON נכנס:** גבול גודל, בלי שדות לא ידועים, בלי ערך שני.
- **שגיאות:** `HTTPError` יוצאת כמו שהיא; כל השאר 500 כללי, והפרטים ללוג.
- **תצורה:** נבדקת כולה בעלייה, כל הבעיות יחד, `lookup` מוזרק לבדיקות.
- **לוגים:** JSON עם `slog`; מזהה הבקשה מה־context; רק הגרסאות עם context.
- **כיבוי:** אות הופך ל־context, `Shutdown` עם context חדש לזמן החסד, ו־`Close` אחריו.
- **pprof:** מאזין פרטי בלבד; `flat` מול `cum`.
- **בנצ'מרק:** `b.Loop()`, שלושה מספרים, ותמיד עם שם המכונה.
- **שאלה פתוחה:** ___

<div dir="ltr" align="left">

```text
mux.HandleFunc("GET /{code}", h)        r.PathValue("code")   r.Pattern
type Middleware func(http.Handler) http.Handler
r.Body = http.MaxBytesReader(w, r.Body, n); dec.DisallowUnknownFields()
slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{Level: lvl}))   log.InfoContext(ctx, ...)
ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM); context.AfterFunc(ctx, stop)
srv.Shutdown(context.WithTimeout(context.WithoutCancel(ctx), grace))  then srv.Close()
ReadHeaderTimeout ReadTimeout WriteTimeout IdleTimeout
httptest.NewRecorder()  httptest.NewServer(h)  slog.New(slog.DiscardHandler)
go tool pprof -top 'http://127.0.0.1:6060/debug/pprof/profile?seconds=10'
go test -run '^$' -bench Resolve -benchmem ./internal/app/
my bench: ___ ns/op  ___ B/op  ___ allocs/op  on cpu ___  date ___
my profile top function: ___
```

</div>

קישורים: [[05-go-concurrency]] · [[07-python-typed-packaged]]
