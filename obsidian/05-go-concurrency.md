---
direction: rtl
tags: [senior-course, book]
chapter: 5
---
# מקביליות ב־Go

- **goroutine:** זולה, וקוד חוסם בה בסדר. לכל אחת בעלים ודרך סיום.
- **channel:** בלי חוצץ הוא נקודת מפגש, עם חוצץ הוא תור. רק השולח סוגר, פעם אחת.
- **כמה שולחים:** goroutine נפרדת מחכה ל־`Wait` ואז סוגרת.
- **select:** `ctx.Done()` כענף בכל המתנה ארוכה.
- **context:** ביטול שיתופי; `defer cancel()` תמיד; פרמטר ראשון.
- **Err מול Cause:** `Err` אומרת מה קרה, `Cause` אומרת למה. לקוח HTTP מחזיר את הסיבה.
- **Mutex מול channel:** מבנה משותף בנעילה, מספר אחד ב־`atomic`, זרימה ב־channel.
- **errgroup:** `SetLimit` למקביליות חסומה; השגיאה הראשונה מבטלת; panic לא נתפס.
- **-race:** מוצא מרוצים שקרו; שתיקה אינה הוכחה. ב־WSL2 צריך מהדר C.
- **דליפות:** שולח בלי מקבל הוא הנפוץ; `goroutineleak` תופס חסומות לנצח.
- **שאלה פתוחה:** ___

<div dir="ltr" align="left">

```text
go f()                                  start a goroutine; it ends when f returns
wg.Go(func(){...}); wg.Wait()           wait for a group
ch := make(chan T) / make(chan T, n)    unbuffered (rendezvous) / buffered
v, ok := <-ch                           ok == false after close
close(ch)                               sender only, once
select { case ch <- v: case <-ctx.Done(): return }
ctx, cancel := context.WithTimeout(ctx, d); defer cancel()
ctx.Err()  Canceled | DeadlineExceeded       context.Cause(ctx)  the reason
g, gctx := errgroup.WithContext(ctx); g.SetLimit(n); g.Go(func() error {...}); err := g.Wait()
go get golang.org/x/sync@v0.23.0
go test -race ./...                     always; needs a C compiler on Linux
pprof.Lookup("goroutineleak").WriteTo(&buf, 1)   want "goroutineleak profile: total 0"
my linkcheck: workers ___  timeout ___  leak seen on day 10 at pool.go line ___
```

</div>

קישורים: [[04-go-types-interfaces-errors]] · [[06-go-http-service]]
