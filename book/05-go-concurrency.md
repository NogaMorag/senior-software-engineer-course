---
direction: rtl
lang: he
title: "פרק 5: מקביליות ב־Go"
chapter: 5
part: "חלק א: Go"
---

<div dir="rtl" lang="he">

# פרק 5: מקביליות ב־Go

שירות backend עושה כמה דברים בבת אחת כל הזמן: מטפל בבקשות רבות, מחכה לבסיס נתונים, קורא לשירותים אחרים,
ועוצר הכול כשמגיע אות כיבוי. ב־Go זה נעשה עם goroutines, channels ו־context, וזה מהנושאים הראשונים
שמראיינים בודקים אצל מי שאומרת שכתבה Go בייצור.

הפרק מלמד את הכלים מול מה שאת מכירה מ־Rust. ההבדל המרכזי פשוט לנסח וקשה להפנים: ב־Go כמעט כלום לא נבדק
בזמן קומפילציה. מרוצים, דליפות וביטול שלא מגיע הם באחריותך, ולכן חלק גדול מהפרק עוסק בכלים שתופסים אותם.

## מה נלמד בפרק

- מה זו goroutine, ובמה היא שונה מ־thread ומ־async/await.
- עבודה עם channels ו־`select`: העברת עבודה ותוצאות, סגירה, וזמנים קצובים.
- ה־context: ביטול וגבול זמן כולל שעוברים לאורך השרשרת, עד בקשת ה־HTTP.
- החבילות `sync` ו־`sync/atomic`, ומתי Mutex ומתי channel.
- החבילה `errgroup`: המתנה, שגיאה ראשונה וביטול, כולל הגבלת מקביליות.
- דפוסים קלאסיים של מקביליות, ושמירה על סדר התוצאות.
- גלאי המרוצים, ובדיקה שלא נשארו goroutines תקועות.

## ידע מקדים

- פרק 2: Go מותקנת ב־WSL2, ומאגר הקורס משוכפל אצלך.
- פרק 3: מודולים (סעיף 3.3), slices (סעיף 3.6) ובדיקות טבלאיות (סעיף 3.8).
- פרק 4: ממשקים (סעיף 4.2), שגיאות ועטיפה (סעיפים 4.4 ו־4.5), ו־panic (סעיף 4.6).
- החבילה `net/http` משמשת כאן כלקוח. השרת והחבילה `httptest` מוסברים בפרק 6.

## 5.1 goroutines מול async/await ו־threads

ה־goroutine (פונקציה שרצה במקביל לשאר התוכנית, בניהול ה־runtime של Go) מתחילה במילה `go` לפני קריאה. אין
ערך חזרה ואין handle: הקריאה חוזרת מיד, והפונקציה רצה ברקע.

ה־runtime מריץ הרבה goroutines על מעט threads של מערכת ההפעלה. כל אחת מתחילה עם מחסנית קטנה שגדלה לפי
הצורך, ולכן מאות אלפים בתהליך אחד הם מצב רגיל. מספר ה־threads שמריצים קוד בו־זמנית נקבע
בהגדרה `GOMAXPROCS`. ברירת המחדל היא מספר הליבות הזמינות לתהליך, ובקונטיינר עם מגבלת מעבד גם לפי המגבלה.

ההבדל החשוב מול async ב־Rust: אין ב־Go „צבע” לפונקציות. כל פונקציה רשאית לחסום, וכשהיא מחכה לרשת,
ה־runtime מחנה אותה ומריץ goroutine אחרת על אותו thread. אין `async`, אין `.await`, ואין צורך בגרסה
אסינכרונית של כל ספרייה.

| | thread של מערכת ההפעלה | async ב־Rust (tokio) | goroutine |
| --- | --- | --- | --- |
| מי מתזמן | הקרנל | ה־executor, בנקודות `.await` | ה־runtime, גם באמצע לולאה |
| עלות יצירה | גבוהה, מחסנית קבועה | נמוכה: future הוא מבנה בזיכרון | נמוכה: מחסנית קטנה שגדלה |
| קריאה חוסמת | חוסמת thread אחד | תוקעת worker של ה־executor | בסדר: ה־runtime מחנה אותה |
| ביטול | אין מנגנון מובנה | שחרור ה־future | שיתופי, דרך context (סעיף 5.4) |
| בדיקת מרוצים | `Send` ו־`Sync` בקומפילציה | `Send` ו־`Sync` בקומפילציה | בזמן ריצה, עם `-race` (סעיף 5.8) |

מקביליות (concurrency) היא מבנה שבו כמה משימות מתקדמות בתקופות חופפות; ביצוע מקבילי (parallelism) הוא
ריצה בו־זמנית ממש, על כמה ליבות. תוכנית עם ליבה אחת עדיין יכולה להיות מקבילית במבנה שלה.

> **איך Rust עושה את זה:** ב־tokio, הפונקציה `tokio::spawn` דורשת future שהוא `Send` ו־`'static`. הדרישה
> הזאת מונעת ממך לשתף בטעות וקטור מקומי בין משימות. ב־Go אין דרישה כזאת: `go f()` מקבל סגור (closure) שנוגע בכל משתנה בטווח,
> והקומפיילר שותק. החופש הזה הוא הסיבה לסעיפים 5.8 ו־5.9.

הצורה הבסיסית, מתוך תוכנית ההדגמה של התרגיל (הקובץ `data/racedemo/main.go`): שמונה goroutines, והמתנה
לכולן.

<div dir="ltr" align="left">

```go
var wg sync.WaitGroup
for range 8 {
	wg.Go(func() {
		for range 1000 {
			add()
		}
	})
}
wg.Wait()
```

</div>

המתודה `WaitGroup.Go` מפעילה goroutine ורושמת אותה; `Wait` חוסמת עד שכולן חוזרות. בקוד ישן תראי את הצורה
הארוכה, `Add(1)` לפני `go` ו־`defer wg.Done()` בפנים. בלי `Wait`, הפונקציה `main` חוזרת, והתהליך נגמר
באמצע העבודה בלי הודעה.

מכאן הכלל של כל הפרק: **לכל goroutine יש בעלים, ודרך ידועה שבה היא מסתיימת.**

## 5.2 channels

ה־channel (צינור עם טיפוס להעברת ערכים בין goroutines) נוצר עם `make`. שליחה היא `ch <- v`, קבלה היא
`v := <-ch`. יש שני סוגים:

- **בלי חוצץ:** שליחה ממתינה למקבל, וקבלה ממתינה לשולח. זו נקודת מפגש.
- **עם חוצץ** בגודל `n`: שליחה מצליחה מיד כל עוד יש מקום, וממתינה כשהתור מלא.

סגירה אומרת „לא יגיעו עוד ערכים”. לולאת `range` על channel נמשכת עד הסגירה, והצורה `v, ok := <-ch` מחזירה
`ok` שקר אחריה. הטבלה הזאת מונעת את רוב התקלות:

| פעולה | channel שהוא `nil` | פתוח | סגור |
| --- | --- | --- | --- |
| שליחה | חוסמת לנצח | נשלחת או ממתינה | panic |
| קבלה | חוסמת לנצח | מקבלת או ממתינה | ערך האפס מיד |
| `close` | panic | נסגר | panic |

מכאן: **רק השולח סוגר, ורק פעם אחת.** כשיש כמה שולחים, אף אחד מהם לא יכול לסגור בבטחה, וצד שלישי מחכה
לכולם וסוגר. כך זה בפונקציה `CheckPool` של התרגיל (הקובץ `pool.go`):

<div dir="ltr" align="left">

```go
// Close out only after every worker has stopped sending.
go func() {
	wg.Wait()
	close(out)
}()
```

</div>

ב־Rust, ה־channel נסגר מעצמו כשה־`Sender` האחרון משתחרר. ב־Go אין שחרור אוטומטי, ולכן הסגירה היא החלטה
מפורשת, במקום אחד. אפשר גם להגביל כיוון בטיפוס: `chan<- T` לשליחה בלבד ו־`<-chan T` לקבלה בלבד,
והקומפיילר אוכף.

## 5.3 select וזמנים קצובים

הפקודה `select` ממתינה על כמה פעולות channel ומבצעת את הראשונה שמוכנה; כשכמה מוכנות יחד, הבחירה אקראית.
ענף `default` הופך אותה ללא חוסמת. השימוש הנפוץ הוא „שלחי, אלא אם בוטלנו”, כמו במחולל שבתחילת `CheckPool`
(הקובץ `pool.go`):

<div dir="ltr" align="left">

```go
go func() {
	defer close(jobs)
	for i, u := range urls {
		select {
		case jobs <- job{i, u}:
		case <-ctx.Done():
			return
		}
	}
}()
```

</div>

בלי ה־`select`, שליחה אחרי שה־workers הפסיקו לקרוא הייתה חוסמת לנצח. איתו, ביטול ה־context מחזיר את
המחולל, וה־`defer` סוגר את `jobs`.

זמן קצוב (timeout) בונים מ־`time.After`, שמחזירה channel שמקבל ערך אחרי זמן נתון. שרת הבדיקות של התרגיל
משתמש בו כדי להיות איטי, אבל לא יותר מהלקוח (הקובץ `linkcheck_test.go`):

<div dir="ltr" align="left">

```go
// /slow answers only after 10 seconds, or never if the client gives up first.
mux.HandleFunc("GET /slow", func(w http.ResponseWriter, r *http.Request) {
	select {
	case <-time.After(10 * time.Second):
	case <-r.Context().Done():
	}
})
```

</div>

שני טריקים: channel שהוא `nil` לא מוכן אף פעם, ולכן הצבת `nil` במשתנה מכבה את הענף שלו ב־`select`.
ובלולאה ארוכה עדיף טיימר אחד שמאפסים (`time.NewTimer`) על פני קריאה חדשה ל־`time.After` בכל סיבוב.

## 5.4 context: ביטול ומועדים

הממשק `context.Context` עובר כפרמטר ראשון לכל פונקציה שעלולה לחכות, ונושא ביטול, גבול זמן כולל וערכים של
הבקשה. זו הדרך של Go לבטל עבודה:

| מתודה | מה היא מחזירה |
| --- | --- |
| `Done()` | channel שנסגר בביטול |
| `Err()` | `nil`, ואחרי ביטול `context.Canceled` או `context.DeadlineExceeded` |
| `Deadline()` | גבול הזמן הכולל, אם הוגדר |
| `Value(key)` | ערך שצורף; סעיף 6.2 שומר כך את מזהה הבקשה |

ערכי context יוצרים עץ: כל פונקציה בטבלה הבאה מקבלת הורה ומחזירה ילד, וביטול הורה מבטל את כל צאצאיו, אבל
לא להפך.

| פונקציה | מתי הילד מבוטל |
| --- | --- |
| `context.WithCancel` | בקריאה ל־`cancel` |
| `context.WithTimeout` | אחרי משך זמן, או ב־`cancel` |
| `context.WithDeadline` | בשעה נתונה, או ב־`cancel` |
| `context.WithCancelCause` | ב־`cancel(err)`; הסיבה נקראת עם `context.Cause` |

את `cancel` קוראים תמיד, בדרך כלל עם `defer`, גם כשהכול הצליח: היא משחררת טיימר ורישום אצל ההורה.
הכלי `go vet` מתריע כשהיא הולכת לאיבוד.

הביטול **שיתופי**. שום דבר לא נהרג: ביטול רק סוגר channel. קוד שלא בודק את `ctx.Done()`, ולא מעביר את
`ctx` הלאה, ימשיך לרוץ. הספרייה הסטנדרטית מכבדת context בכל מקום שחשוב, ובראשם `net/http`.

> **איך Rust עושה את זה:** ב־tokio, ביטול הוא שחרור: `select!` שבחר ענף משחרר את ה־futures של שאר הענפים,
> והם מבוטלים בלי שיתוף פעולה מצידם. ב־Go, ענף שלא נבחר פשוט לא קרה, וה־goroutine שמאחוריו ממשיכה לרוץ.
> המקבילה הקרובה היא `CancellationToken` של tokio-util: אות שכל משימה בודקת בעצמה.

כך גבול זמן מגיע לבקשת HTTP, בפונקציה `checkOne` של התרגיל (הקובץ `linkcheck.go`):

<div dir="ltr" align="left">

```go
func (c *Checker) checkOne(ctx context.Context, url string) Result {
	ctx, cancel := context.WithTimeout(ctx, c.Timeout)
	defer cancel()

	res := Result{URL: url}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		res.Err = err
		return res
	}
	start := time.Now()
	resp, err := c.client().Do(req)
	res.Elapsed = time.Since(start)
	if err != nil {
		// The client reports context.Cause(ctx), which may be any error.
		// Classify by what happened to ctx itself: timeout or cancellation.
		if ctxErr := ctx.Err(); ctxErr != nil {
			err = ctxErr
		}
		res.Err = err
		return res
	}
```

</div>

לכל בקשה יש זמן קצוב משלה, וביטול של ההורה, כלומר של כל הריצה, מגיע אליה גם כן. הלקוח של `net/http` עוקב
אחרי ה־context: כשהזמן עובר, הוא סוגר את החיבור ו־`Do` חוזרת עם שגיאה. בצד השרת, `r.Context()` מתבטל
כשהלקוח ניתק, ולכן ה־handler האיטי של סעיף 5.3 משתחרר מיד.

ההערה בקוד מסמנת מלכודת אמיתית, שבדיקות התרגיל חשפו. כשה־context בוטל עם סיבה (ו־`errgroup` עושה
בדיוק את זה), הלקוח מחזיר את **הסיבה**, ו־`errors.Is` מול `context.Canceled` נכשלת. לכן שואלים את
`ctx.Err()`:

| שאלה | הפונקציה |
| --- | --- |
| האם בוטל, ובאיזה אופן | `ctx.Err()` |
| למה בוטל | `context.Cause(ctx)` |

ועוד מוסכמות: `ctx` הוא הפרמטר הראשון; לא שומרים אותו בתוך struct; בשורש משתמשים
ב־`context.Background()`; ו־`Value` מיועד לנתוני בקשה כמו מזהה, לא לפרמטרים אופציונליים.

## 5.5 sync: Mutex, WaitGroup, Once

כשכמה goroutines קוראות ומעדכנות מבנה משותף, הכלי הנכון הוא נעילה (lock):

| כלי | מה הוא עושה |
| --- | --- |
| `sync.Mutex` | נעילה בלעדית; ערך האפס מוכן לשימוש |
| `sync.RWMutex` | הרבה קוראים יחד, או כותב אחד |
| `sync.WaitGroup` | המתנה לקבוצת goroutines |
| `sync.Once` | פונקציה שרצה פעם אחת בלבד; וגם `OnceValue` |
| `sync/atomic` | מונה או דגל בלי נעילה, כמו `atomic.Int64` |

בניגוד ל־`Mutex<T>` של Rust, ה־Mutex של Go לא עוטף את הנתון. הוא שדה לצד הנתון, והקשר ביניהם הוא הערה
ומשמעת. התיקון הראשון של המונה המשותף בתוכנית ההדגמה (הקובץ `data/racedemo/main.go`):

<div dir="ltr" align="left">

```go
type progress struct {
	mu      sync.Mutex // guards locked
	locked  int
	racy    int          // plain int, no protection: the planted bug
	counter atomic.Int64 // atomic read-modify-write
}

func (p *progress) addLocked() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.locked++
}
```

</div>

ההערה `guards locked` היא המוסכמה, וה־`defer` משחרר גם ביציאה מוקדמת. struct שמכיל Mutex אסור להעתיק, כי
להעתק יש נעילה נפרדת; לכן מתודות עליו מקבלות מצביע, ו־`go vet` מתריע על העתקה. התיקון השני פשוט יותר
כשמדובר במספר אחד: `p.counter.Add(1)`.

מתי Mutex ומתי channel? זו שאלת ראיון קבועה, והתשובה היא „לפי מה שזורם”:

| המצב | הכלי |
| --- | --- |
| מבנה משותף: מטמון, מונים, map | Mutex או RWMutex |
| מספר אחד או דגל | `sync/atomic` |
| עבודה או תוצאות שעוברות בין goroutines | channel |
| אות „נגמר” לכמה מאזינים | סגירת channel, או context |
| המתנה לקבוצה | `WaitGroup`, או `errgroup` כשיש שגיאות |

## 5.6 errgroup

התבנית „הפעילי כמה משימות, חכי לכולן, קבלי את השגיאה הראשונה ובטלי את השאר” חוזרת כל כך הרבה, שיש לה
חבילה: `golang.org/x/sync/errgroup`. החבילות תחת `golang.org/x` מתוחזקות על ידי צוות Go אבל אינן בספרייה
הסטנדרטית, ולכן מוסיפים אותן למודול. בתרגיל הגרסה נעולה ב־`go.mod` (נכון לאוקטובר 2026; ראי נספח ג):

<div dir="ltr" align="left">

```bash
go get golang.org/x/sync@v0.23.0
```

</div>

| קריאה | מה היא עושה |
| --- | --- |
| `errgroup.WithContext(ctx)` | קבוצה ו־context ילד שמתבטל בשגיאה הראשונה |
| `g.Go(f)` | מריצה את `f` ב־goroutine; עם מגבלה, ממתינה למקום |
| `g.SetLimit(n)` | לכל היותר `n` פונקציות פעילות יחד |
| `g.TryGo(f)` | כמו `Go`, אבל מוותרת כשאין מקום |
| `g.Wait()` | ממתינה לכולן ומחזירה את השגיאה הראשונה |

כך `Check` של התרגיל מגביל מקביליות, מבטל בשגיאה הראשונה ושומר על הסדר (הקובץ `linkcheck.go`):

<div dir="ltr" align="left">

```go
func (c *Checker) Check(ctx context.Context, urls []string) ([]Result, error) {
	results := make([]Result, len(urls))
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(c.workers())

	for i, u := range urls {
		g.Go(func() error {
			if gctx.Err() != nil {
				results[i] = Result{URL: u, Err: ErrSkipped}
				return nil
			}
			results[i] = c.checkOne(gctx, u) // each goroutine writes its own element
			if c.FailFast && results[i].Err != nil {
				return fmt.Errorf("fail fast on %s: %w", u, results[i].Err)
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return results, err
	}
	return results, ctx.Err()
}
```

</div>

- **סדר בלי מרוץ.** כל goroutine כותבת לתא משלה ב־slice שהוקצה מראש. תאים שונים הם זיכרון שונה, ו־`Wait`
  מבטיחה שכל הכתיבות נראות אחריה.
- **משתני לולאה.** מאז Go 1.22 כל סיבוב מקבל עותק משלו של `i` ושל `u`. בקוד ישן תראי `u := u`, התיקון
  לבאג הקלאסי.
- **שני contexts.** ה־goroutines משתמשות ב־`gctx`, שמתבטל בשגיאה הראשונה; הבדיקה בסוף היא של ההורה, למקרה
  שהמשתמשת לחצה Ctrl-C.
- **שגיאה מול תוצאה.** בלי `FailFast`, קישור שבור הוא תוצאה, וה־goroutine מחזירה `nil`. רק כשל שאמור
  לעצור הכול חוזר כשגיאה.

ו־`errgroup` לא תופס panic: panic בתוך goroutine מפיל את כל התהליך. בקוד המקור של החבילה יש הערה שמסבירה
למה.

## 5.7 דפוסים: worker pool, pipeline, fan-out ו־fan-in

השילוב של `errgroup` עם `SetLimit` מכסה את רוב המקרים, אבל הדפוסים הקלאסיים מופיעים בשמם בקוד ובראיונות:

| שם | מה זה |
| --- | --- |
| worker pool | מספר קבוע של goroutines שקוראות עבודה מאותו channel |
| pipeline | שלבים שמחוברים ב־channels; כל שלב קורא מהקודם |
| fan-out | פיזור העבודה של שלב על כמה goroutines |
| fan-in | איסוף התוצאות של כמה goroutines לנקודה אחת |

הפונקציה `CheckPool` עושה את העבודה של `Check` בשלושה שלבים. את המחולל ראית בסעיף 5.3. השלב השני,
ה־workers (הקובץ `pool.go`):

<div dir="ltr" align="left">

```go
// Stage 2: fan-out to a fixed number of workers.
var wg sync.WaitGroup
for range c.workers() {
	wg.Go(func() {
		for j := range jobs {
			r := c.checkOne(ctx, j.url)
			if c.FailFast && r.Err != nil {
				cancel(fmt.Errorf("fail fast on %s: %w", j.url, r.Err))
			}
			out <- done{j.i, r} // the collector reads until out is closed
		}
	})
}
```

</div>

אחריו בא הסוגר מסעיף 5.2, ואז השלב השלישי, שמחזיר כל תוצאה למקומה לפי האינדקס:

<div dir="ltr" align="left">

```go
// Stage 3: fan-in. Results arrive in completion order; the index puts
// each one back in its place.
results := make([]Result, len(urls))
for i, u := range urls {
	results[i] = Result{URL: u, Err: ErrSkipped}
}
for d := range out {
	results[d.i] = d.res
}
```

</div>

שלוש החלטות תכנון:

- **הסדר נשמר באינדקס.** חלופות: מיון בסוף, או חוצץ שמשחרר תוצאות לפי הסדר כשמזרימים פלט.
- **לחץ חוזר (back-pressure).** `jobs` ו־`out` בלי חוצץ, ולכן אף שלב לא רץ קדימה יותר ממה שהבא אחריו
  מעכל, ואין תור שגדל בלי גבול.
- **מספר goroutines קבוע.** `Check` יוצרת אחת לכל כתובת, ו־`CheckPool` בדיוק `Workers`. לאלף כתובות זה
  זניח; למיליון פריטים, או כשלכל worker יש משאב יקר כמו חיבור לבסיס נתונים, ה־pool מתאים יותר.

דרך שלישית שתראי בקוד קיים: channel עם חוצץ בגודל `n` כסמפור (semaphore). שולחים אליו לפני העבודה ומקבלים
ממנו אחריה; כך `errgroup` ממומש מבפנים.

## 5.8 מרוצי נתונים ו־`-race`

מרוץ נתונים (data race) קורה כששתי goroutines ניגשות לאותו זיכרון בלי סנכרון, ולפחות אחת כותבת. „בלי
סנכרון” פירושו שאין ביניהן יחס קדימות (happens-before) של מודל הזיכרון: לא Mutex, לא channel, לא `Wait`.
התוצאה יכולה להיות שגויה, ובמבנים של כמה מילים, כמו slice או interface, אפילו זיכרון מושחת.

תוכנית ההדגמה מכילה מרוץ שתול: שמונה goroutines מגדילות את אותו `int`.

<div dir="ltr" align="left">

```go
func (p *progress) addRacy() {
	p.racy++ // DATA RACE: read, add one, write back, while others do the same
}
```

</div>

השורה `p.racy++` היא שלוש פעולות: קריאה, הוספה וכתיבה. שתי goroutines קוראות אותו ערך, וכתיבה אחת הולכת
לאיבוד. הריצי את התוכנית בלי גלאי המרוצים (race detector) ואיתו:

<div dir="ltr" align="left">

```bash
cd $COURSE/exercises/ch05-linkcheck/data/racedemo
go run .
go run -race .
```

</div>

הריצה הראשונה כנראה תדפיס מספר קטן מ־8000, בלי אזהרה. השנייה תדפיס דוח בצורה הזאת (הכתובות והמספרים
משתנים):

<div dir="ltr" align="left">

```text
==================
WARNING: DATA RACE
Read at 0x00c0001c2010 by goroutine 11:
  main.(*progress).addRacy()
      .../racedemo/main.go:34 +0x38

Previous write at 0x00c0001c2010 by goroutine 7:
  main.(*progress).addRacy()
      .../racedemo/main.go:34 +0x4c
...
Found 2 data race(s)
exit status 66
```

</div>

הדוח מראה את שתי הגישות המתנגשות ואת המחסנית של כל אחת. הדגל עובד עם `go run`, `go build` ו־`go test`.
ב־Linux, ולכן גם ב־WSL2, הוא דורש מהדר C; שגיאה שמזכירה cgo פירושה שחסרה החבילה `build-essential` (פרק 2,
סעיף 2.3).

> **איך Rust עושה את זה:** ב־Rust בטוחה, מרוץ נתונים הוא שגיאת קומפילציה: את התוכנית הזאת לא היית מצליחה
> לקמפל. ב־Go הגלאי בודק בזמן ריצה, ורק נתיבים שבאמת רצו. הוא לא מתריע סתם, אבל שתיקה שלו אינה הוכחה. לכן
> כל הבדיקות רצות עם `-race`, תמיד, גם ב־CI.

שני דברים שהגלאי לא מוצא. מרוץ שלא קרה בריצה הזאת: בדיקה שבאמת מפעילה את הקוד המקבילי היא מה שנותן לו
הזדמנות. ומרוץ לוגי (race condition) שכל גישה בו מוגנת לחוד: קוד שבודק „האם הקוד תפוס?” תחת נעילה אחת
וכותב תחת נעילה שנייה נקי מבחינת הגלאי, ועדיין שגוי. לכן בפרק 6 הבדיקה והכתיבה ב־`MemoryStore.Create`
נעשות תחת אותה נעילה. הגלאי מאט ומנפח זיכרון, ולכן לא רץ בייצור.

## 5.9 דליפות goroutines

דליפת goroutine היא goroutine שלא תסתיים לעולם: היא חוסמת על channel שאף אחד לא ישתמש בו, או מחכה
ל־context שלא יבוטל. אחת לא מורגשת; אחת לכל בקשה היא שירות שהזיכרון שלו עולה עד שהוא נופל.

| הסיבה | איך זה נראה | התיקון |
| --- | --- | --- |
| שולח בלי מקבל | האוסף חזר מוקדם, וה־workers חוסמים על שליחה | לקרוא עד הסגירה |
| מקבל בלי סגירה | `range` על channel שלא נסגר אף פעם | השולח היחיד סוגר, עם `defer` |
| context שלא מבוטל | המתנה ל־`Done()` של `context.Background()` | להעביר context שבאמת יבוטל |
| goroutine בלי בעלים | `go f()` ושוכחים | `WaitGroup` או `errgroup` שמישהו ממתין להם |

כך נראית הסיבה הראשונה. נניח שמישהו „משפר” את האוסף ב־`CheckPool` כך שיפסיק לקרוא מיד כשהריצה בוטלה. הקטע
שמור בתרגיל, בקובץ `data/planted-leak.txt`:

<div dir="ltr" align="left">

```go
for d := range out {
	results[d.i] = d.res
	if ctx.Err() != nil {
		break // BUG: stop reading early; workers block on out forever
	}
}
```

</div>

הפונקציה חוזרת מהר יותר, והתוצאות נראות תקינות. אבל כל worker שמחזיק תוצאה חוסם על שליחה ל־`out`,
וה־goroutine שמחכה להם ב־`wg.Wait()` חוסמת גם היא, לנצח.

הדרך הוותיקה לתפוס את זה בבדיקה היא לספור את `runtime.NumGoroutine()` לפני ואחרי. היא רועשת, כי
goroutines של חיבורי HTTP פתוחים נספרות גם הן. נכון לאוקטובר 2026, ב־Go 1.27 (ראי נספח ג) יש דרך מדויקת יותר: הפרופיל
`goroutineleak` של `runtime/pprof`. כשכותבים אותו, ה־runtime מריץ איסוף זבל (GC) שמוצא goroutines שחוסמות
על channel או נעילה שאף goroutine חיה לא מגיעה אליהם. הן, בהגדרה, לא יתעוררו. כך זה בבדיקות של התרגיל
(הקובץ `linkcheck_test.go`):

<div dir="ltr" align="left">

```go
// checkNoLeaks fails the test if the runtime finds goroutines blocked forever
// on channels or locks that nothing else can reach (section 5.9). A goroutine
// that is about to block needs a moment to get there, so it looks several
// times over a short window.
func checkNoLeaks(t *testing.T) {
	t.Helper()
	for range 10 {
		var buf bytes.Buffer
		if err := pprof.Lookup("goroutineleak").WriteTo(&buf, 1); err != nil {
			t.Fatalf("goroutineleak profile: %v", err)
		}
		if !strings.HasPrefix(buf.String(), "goroutineleak profile: total 0\n") {
			t.Errorf("leaked goroutines:\n%s", buf.String())
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
}
```

</div>

כל בדיקה של `linkcheck` קוראת לה בסוף. עם הקטע השתול, בדיקת fail-fast נכשלת ומראה איפה הן תקועות. כך זה נראה בהרצה שבה נכתב הספר (מקוצר; הכתובות
אצלך יהיו אחרות):

<div dir="ltr" align="left">

```text
--- FAIL: TestCheckFailFast/CheckPool (0.00s)
    linkcheck_test.go:226: leaked goroutines:
        goroutineleak profile: total 2
        #	0x10133c393	example.com/ch05-linkcheck.(*Checker).CheckPool.func2+0x93	.../pool.go:49
        #	0x10133c2d3	example.com/ch05-linkcheck.(*Checker).CheckPool.func3+0x33	.../pool.go:55
```

</div>

שורה 49 היא השליחה ל־`out`, ושורה 55 היא `wg.Wait()`. הגבולות של הפרופיל: הוא לא רואה goroutine שחוסמת על
channel שמשתנה גלובלי מחזיק, על קריאה מהרשת, או בלולאה אינסופית; והוא בודק את כל התהליך, כך שדליפה מבדיקה
אחת תכשיל גם את הבאות. בשירות שרץ, הפרופיל `goroutine` של סעיף 6.8 מראה את כל ה־goroutines, ומספר שעולה
בהתמדה הוא הסימן הראשון לדליפה.

## 5.10 תוכנית קטנה: `linkcheck`

התרגיל הוא בודק קישורים: רשימת כתובות, מקביליות חסומה, זמן קצוב לכל בקשה, עצירה של הכול, ודוח לפי סדר
הקלט. בקטן, זה כל מה שהפרק לימד, ובגדול זה מה שעושה כל שירות שמדבר עם הרבה שירותים אחרים.

| חלק | מה הוא | יום |
| --- | --- | --- |
| `checkOne` | בקשה אחת עם זמן קצוב, וסיווג התוצאה | 8 |
| `Check` | קודם goroutine לכל כתובת עם `WaitGroup`, אחר כך `errgroup` | 8 ו־9 |
| `CheckPool` | אותו דבר, בשלושה שלבים של channels | 9 |
| מרוץ ודליפה שתולים | לראות את הדוחות, לתקן, והכול ירוק עם `-race` | 10 |

הדוח, הדגלים וקודי היציאה (בתוכנית `cmd/linkcheck`) נתונים מוכנים. כל תוצאה מסווגת במילה אחת:

| מילה | פירוש |
| --- | --- |
| `ok` | תשובה עם סטטוס מתחת ל־400 |
| `broken` | תשובה עם סטטוס 400 ומעלה |
| `timeout` | הזמן הקצוב עבר, של הבקשה או של כל הריצה |
| `cancelled` | הבקשה התחילה ובוטלה |
| `skipped` | הריצה בוטלה לפני שהבקשה התחילה |
| `error` | כל השאר: כתובת לא חוקית, חיבור שנדחה |

הבדיקות לא נוגעות באינטרנט. הן מרימות שרת מקומי עם `httptest.NewServer`, עם ארבעה נתיבים בהתנהגות ידועה:

- הנתיב `/ok` עונה מיד.
- הנתיב `/missing` עונה 404.
- הנתיב `/slow` משתהה עשר שניות (סעיף 5.3).
- הנתיב `/work` סופר כמה בקשות בטיפול בבת אחת.

 כל בדיקה רצה פעם לכל מימוש, ומסתיימת
בבדיקת הדליפות. אחרי שהכול ירוק, הריצי את התוכנית על הקובץ לדוגמה:

<div dir="ltr" align="left">

```bash
cd ~/senior-course/ch05
go run ./cmd/linkcheck -workers 3 -timeout 2s $COURSE/exercises/ch05-linkcheck/data/urls.txt
```

</div>

<div dir="ltr" align="left">

```text
ok        200  https://go.dev/
broken    404  https://go.dev/this-page-does-not-exist
ok        200  https://example.com/
error       -  http://127.0.0.1:9/
timeout     -  https://10.255.255.1/
error       -  not a url
checked 6, ok 2, not ok 4
```

</div>

הצורה היא מה שחשוב: ברשת שלך הכתובת `10.255.255.1` יכולה להחזיר `error` במקום `timeout`. ו־Ctrl-C באמצע
ריצה מבטל את כל הבקשות שבדרך, דרך `signal.NotifyContext`, שתכירי בסעיף 6.5.

## סיכום

- ה־goroutine זולה, מתוזמנת על מעט threads, וקוד חוסם בה בסדר. לכל אחת בעלים ודרך סיום.
- רק השולח סוגר channel, ופעם אחת; עם כמה שולחים, צד שלישי סוגר אחרי `Wait`.
- הפקודה `select`, עם `ctx.Done()` כענף, בכל המתנה שעלולה להימשך.
- ה־context הוא ביטול שיתופי וגבול זמן כולל לאורך השרשרת. `Err` אומרת מה קרה, `Cause` אומרת למה.
- נעילה (Mutex) למבנה משותף, `atomic` למספר אחד, channel לזרימה.
- החבילה `errgroup` עם `SetLimit` היא הדרך הקצרה למקביליות חסומה עם ביטול; worker pool נותן מספר
  goroutines קבוע.
- הדגל `-race` מוצא מרוצים שקרו; `goroutineleak` מוצא goroutines תקועות לנצח. שניהם בכל הבדיקות.

## מונחים לזכור

| מונח בעברית | באנגלית | במשפט אחד |
| --- | --- | --- |
| goroutine | goroutine | פונקציה שרצה במקביל בניהול ה־runtime, ומתחילה במילה `go` |
| מקביליות | concurrency | מבנה שבו כמה משימות מתקדמות בתקופות חופפות |
| ביצוע מקבילי | parallelism | ריצה בו־זמנית ממש, על כמה ליבות |
| channel | channel | צינור עם טיפוס להעברת ערכים בין goroutines |
| context | context | ממשק שנושא ביטול, גבול זמן כולל וערכי בקשה |
| ביטול שיתופי | cooperative cancellation | ביטול שמגיע רק לקוד שבודק אותו |
| זמן קצוב | timeout | כמה זמן מחכים לפעולה אחת לפני שמוותרים |
| גבול זמן כולל | deadline | הרגע שבו כל העבודה חייבת להסתיים |
| נעילה | lock | מבטיחה שרק goroutine אחת ניגשת לנתון בכל רגע |
| worker pool | worker pool | מספר קבוע של goroutines שקוראות עבודה מאותו channel |
| fan-out ו־fan-in | fan-out, fan-in | פיזור עבודה על כמה goroutines, ואיסוף התוצאות לנקודה אחת |
| לחץ חוזר | back-pressure | שלב מהיר מואט כי השלב הבא לא מקבל עוד |
| מרוץ נתונים | data race | שתי גישות לא מסונכרנות לאותו זיכרון, ולפחות אחת כתיבה |
| מרוץ | race condition | תוצאה שתלויה בסדר האירועים, גם בלי מרוץ נתונים |
| גלאי המרוצים | race detector | הדגל `-race`, שמוצא מרוצי נתונים שקרו בריצה |
| דליפת goroutines | goroutine leak | goroutine שלא תסתיים לעולם ומחזיקה זיכרון |

## טעויות נפוצות

- **לשכוח לחכות.** `main` או בדיקה חוזרות לפני שה־goroutines סיימו, והעבודה נעלמת בשקט.
- **לסגור channel מהצד המקבל, או מכמה שולחים.** התוצאה היא panic. השולח היחיד סוגר, או goroutine שמחכה
  לכולם.
- **להפסיק לקרוא מוקדם.** השולחים נשארים חסומים לנצח. קוראים עד הסגירה.
- **לבדוק `errors.Is(err, context.Canceled)` על שגיאה של לקוח HTTP.** אם בוטל עם סיבה, חוזרת הסיבה.
  שואלים את `ctx.Err()`.
- **לקרוא ל־`wg.Add(1)` בתוך ה־goroutine.** `Wait` עלולה לרוץ לפניו ולחזור מוקדם. `Add` לפני `go`, או
  `wg.Go`.
- **בדיקות בלי `-race`.** מרוץ שלא נבדק יופיע בייצור, בעומס, בלי עקבות.

## שאלות ראיון

1. **מה ההבדל בין goroutine ל־thread ול־async/await, ולמה אפשר להריץ מאות אלפי goroutines?**
   מחסנית קטנה שגדלה ותזמון של הרבה goroutines על מעט threads הופכים יצירה והחלפה לזולות. בניגוד ל־async,
   אין צבע לפונקציות: קוד חוסם רגיל עובד, וה־runtime מחנה את ה־goroutine בזמן קלט ופלט. המחיר: אין בדיקה
   בקומפילציה של שיתוף נתונים, והביטול לא אוטומטי. וזולות אינן חינמיות: בלי גבול מקביליות אפשר עדיין
   להפיל שירות.

2. **איך מבטלים עבודה ב־Go, ומה קורה ל־goroutine שמתעלמת מהביטול?**
   מעבירים `context.Context` לכל פונקציה שעלולה לחכות, ומבטלים מלמעלה: ידנית, בזמן קצוב או בגבול זמן
   כולל. הביטול שיתופי, ולכן goroutine שמתעלמת ממשיכה לרוץ, ואם היא חוסמת על משהו שלא יגיע, היא דולפת. זה
   ההבדל מ־tokio, שבו שחרור ה־future הוא הביטול.

   **המשך:** מה ההבדל בין `ctx.Err()` ל־`context.Cause(ctx)`? הראשונה אומרת אם ואיך בוטל; השנייה מחזירה
   את הסיבה שנמסרה ל־`cancel(err)`, למשל השגיאה הראשונה של `errgroup`. מסווגים לפי הראשונה, ומדווחים לפי
   השנייה.

3. **מתי Mutex ומתי channel?**
   לפי מה שקורה לנתון. מבנה משותף שכולם קוראים ומעדכנים, כמו מטמון, מתאים ל־Mutex, ול־RWMutex כשהקריאות
   רבות בהרבה. נתון שעובר מ־goroutine לאחרת מתאים ל־channel, שמבטא העברת בעלות ונותן לחץ חוזר. למספר
   בודד, `atomic`. הטעויות הנפוצות: channel כ„נעילה מסובכת” סביב map, ונעילה שמוחזקת בזמן קריאה לרשת.

4. **איך מגבילים מקביליות ל־N ועוצרים הכול בשגיאה הראשונה?**
   הדרך הקצרה היא `errgroup.WithContext` עם `SetLimit(N)`: השגיאה הראשונה מבטלת את ה־context המשותף,
   ו־`Wait` מחזירה אותה. חלופות: worker pool של N goroutines, או channel עם חוצץ N כסמפור. השיקולים: מספר
   goroutines קבוע או אחת לפריט, משאב לכל worker, ושמירת סדר, שמשיגים בכתיבה לתא לפי אינדקס.

5. **מה גלאי המרוצים מוכיח, ומה לא?**
   הוא מוכיח שהיה מרוץ נתונים בריצה הזאת. הוא לא מוכיח שאין מרוצים: הוא רואה רק נתיבים שרצו, ולא מזהה
   מרוץ לוגי שכל גישה בו מוגנת לחוד. לכן מריצים איתו את כל הבדיקות בכל שינוי, וכותבים בדיקות שבאמת
   מפעילות את הקוד המקבילי.

   **המשך:** ואם בדיקה נכשלת עם `-race` רק לפעמים? זה מרוץ אמיתי שתלוי בתזמון. משחזרים עם `-count` גבוה
   ומתקנים את הסנכרון; לא מוחקים את הבדיקה.

6. **איך goroutines דולפות, ואיך תבדקי שאין דליפה?**
   הקלאסית היא שולח בלי מקבל: אוסף שחזר מוקדם, ו־workers שחוסמים לנצח. אחרות: `range` על channel שלא
   נסגר, והמתנה ל־context שלא יבוטל. בבדיקות, הפרופיל `goroutineleak` או ספירה לפני ואחרי; בייצור, מעקב
   אחרי מספר ה־goroutines לאורך זמן והפרופיל `goroutine` כשהוא עולה.

## פתק Obsidian

הפתק המוכן נמצא גם כקובץ: `obsidian/05-go-concurrency.md`. זה תוכנו:

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

## תרגיל

**מטרה:** לבנות את `linkcheck`, בודק קישורים עם מקביליות חסומה, זמן קצוב לכל בקשה, ביטול של כל הריצה ודוח
לפי סדר הקלט, בשני מימושים, ולהוכיח עם `-race` ובדיקת הדליפות שהוא נקי.
**זמן:** שלושה ימי לימוד, ימים 8 עד 10, כ־45, 55 ו־40 דקות בנייה בהתאמה.
**ידע מקדים:** הפרק, פרקים 3 ו־4, ו־Go מותקנת לפי פרק 2.

הצעדים, קוד ההתחלה, הרמזים והפתרון נמצאים
ב־[`exercises/ch05-linkcheck/README.md`](../exercises/ch05-linkcheck/README.md).

**קריטריון סיום:**

- הבדיקה של הנתיב האיטי מדווחת `timeout` לכתובת שלו ו־`ok` לאחרות, ומסתיימת בלי לחכות לשרת האיטי.
- בדיקת המקביליות מראה שלא היו יותר מ־N בקשות בטיפול, והדוח שומר על סדר הקלט.
- מצב fail-fast וביטול של ההורה עוצרים את הריצה מהר, בשני המימושים.
- ראית את דוח המרוץ ואת דוח הדליפה השתולה, ותיקנת את שניהם.
- כל הבדיקות עוברות עם `go test -race ./...`, ובדיקת הדליפות לא מוצאת דבר.

## העשרה אופציונלית

- המאמר „Go Concurrency Patterns: Pipelines and cancellation” בבלוג של Go הוא המקור הקלאסי לדפוסי
  סעיף 5.7.
- „The Go Memory Model” באתר go.dev מגדיר מתי כתיבה נראית לקריאה. קריאה קשה, ואחרי סעיף 5.8 מובנת.
- החבילה `testing/synctest` בודקת קוד עם טיימרים בלי לחכות באמת: הזמן מתקדם רק כשכל ה־goroutines חסומות.
- הספרייה `go.uber.org/goleak` היא חלופה נפוצה לבדיקת דליפות, מבוססת ספירה. מה כל אחת תופסת שהשנייה לא?

</div>
