---
direction: rtl
lang: he
title: "פרק 35: תרגילי קוד מעשיים"
chapter: 35
part: "חלק ט: קריירה"
---

<div dir="rtl" lang="he">

# פרק 35: תרגילי קוד מעשיים

בראיון קוד למשרה בכירה ב־backend מבקשים לרוב לבנות משהו קטן ואמיתי: מגביל קצב, מתזמן, מטמון, מקבל webhooks,
הזמנה של משאב. הפרק נותן שיטה לסבב כזה וחמישה תרגילים ב־Go, שבהם הבדיקות הן המפרט, ואחריהם תרגיל של סקירת
קוד. בכל תרגיל, החשוב הוא לא רק שהבדיקות עוברות, אלא שאת יודעת לומר בקול מה בחרת, ומה זה עולה.

## מה נלמד בפרק

- איך נראה סבב קוד מעשי, לפי מה שהמחקר מצא, ושיטה של שישה צעדים לעבור אותו.
- שעון מוזרק ושעון מזויף: איך בודקים קוד שתלוי בזמן בלי לחכות לזמן.
- מגביל קצב: דלי אסימונים לכל מפתח, וגבול לזיכרון.
- מתזמן משימות עם השהיה, המשך, ניסיונות חוזרים עם המתנה גדלה וביטול.
- מטמון עם זמן חיים: פינוי עצל ופעיל, וגודל חסום לפי LRU.
- מקבל webhooks שבטוח לניסיונות חוזרים, כולל משלוחים כפולים שמגיעים באותו רגע.
- הזמנה בלי הזמנה כפולה, פעם בזיכרון עם נעילה לכל משאב ופעם עם אילוץ החרגה ב־PostgreSQL.
- סקירת קוד בזמן קצוב, עם דירוג לפי חומרה.

## ידע מקדים

- פרק 5: goroutines, ה־context, נעילות וגלאי המרוצים. כל תרגיל רץ עם `-race`.
- פרק 6: `net/http` ו־`httptest`, לתרגיל 4.
- פרק 10: PostgreSQL בקונטיינר ו־pgx, לתרגיל 5.
- פרק 11: טרנזקציות ונעילות; תרגיל 5 מראה חלופה שבה המסד עצמו אוכף.
- פרק 12: מטמון, אי־תוקף והתפרצות מטמון. תרגיל 3 הוא מטמון בזיכרון של תהליך אחד.
- פרק 14: ניסיונות חוזרים, המתנה גדלה ואידמפוטנטיות של handler. תרגילים 2 ו־4 בונים עליהם.
- פרק 16: דלי אסימונים. תרגיל 1 כותב אותו מחדש, מאפס, בזמן קצוב.
- פרק 34, סעיף 34.1: מבנה תהליך הגיוס, וסבבים שבהם מותר או אסור להשתמש ב־AI.

## 35.1 איך נראה ראיון קוד מעשי

### 35.1.1 מה ידוע

לפי [מחקר השוק של הספר](../00-stage1/01-market-research.md), שלב הקוד בתהליך ישראלי הוא מטלת בית, או קוד חי של 45
עד 90 דקות, „מעשי יותר מאלגוריתמי” (מחקר השוק, סעיף 3.3, ביטחון בינוני). שאלות כמו מגביל קצב, מטמון LRU עם זמן
חיים, ומתזמן משימות מצוטטות הרבה כסבבי קוד בטעם backend, אבל המקורות הם אתרי הכנה ומאגרי שאלות, ולא נמצא מקור
ישראלי ראשוני (מחקר השוק, סעיף 3.2, ביטחון נמוך עד בינוני).

שלושה ממצאים גלובליים שמשפיעים על הצורה (מחקר השוק, סעיף 3.2):

- **בעיה בכמה חלקים, במקום חידה.** בסקר של מראיינים, 58% שינו את השאלות לכיוון בעיות מותאמות בכמה חלקים, בגלל
  חשש מרמאות בעזרת AI. ובסבב החדש של Meta עובדים על פרויקט של כמה קבצים, ומרחיבים אותו.
- **אלגוריתמים לא נעלמו.** באותו סקר, אף אחד מ־52 המראיינים בחברות הגדולות לא אמר שהחברה שלו ויתרה על שאלות
  אלגוריתמיות. הספר לא מלמד אותן; אם את מכוונת לחברה כזאת, ראי את ההעשרה בסוף הפרק.
- **סבב סקירת קוד:** עדות אחת, וביטחון נמוך. סעיף 35.7 מתרגל אותו, כי ההכנה זולה.

### 35.1.2 השיטה: שישה צעדים

מה שמבדיל בכירה בסבב כזה הוא סדר העבודה, יותר מהקוד. זו שיטת עבודה, לא ממצא של המחקר:

| צעד | מה עושים | כמה זמן מתוך 45 דקות |
| --- | --- | --- |
| הבהרה | שואלים על הדרישות, על הקנה מידה, על מה שקורה בקצוות; כותבים את התשובות כהערה | 5 |
| ממשק קודם | כותבים את החתימות והטיפוסים, ומראים אותם לפני שכותבים גוף | 3 |
| בדיקות, או רשימה שלהן | הנתיב העיקרי ושני מקרי קצה. אם אין זמן לכתוב, אומרים אותם בקול | 7 |
| גרסה פשוטה ונכונה | נעילה אחת גדולה, מבנה נתונים פשוט. עובד לפני שמהיר | 15 |
| קצוות ומקביליות | גלאי המרוצים, ביטול, שגיאות, זמן שהולך אחורה | 10 |
| המחירים בקול | מה היית משנה בקנה מידה גדול יותר, ולמה לא עשית את זה עכשיו | 5 |

שני עקרונות חוצים את כל הצעדים. ראשית, **חושבים בקול**: מראיינת שלא שומעת אותך לא יכולה לתת לך נקודות על שיקול
דעת, וגם לא יכולה לעזור כשאת תקועה. שנית, **„נכון קודם, ואז מהיר”**: גרסה עם נעילה אחת שעוברת את הבדיקות שווה
יותר מגרסה עם נעילות עדינות שנתקעת.

> **שימי לב:** בסבב שבו מותר AI, השיטה זהה, ורק צעד הגרסה הפשוטה מתקצר. מה שנבדק אז הוא בדיוק הצעדים האחרים:
> ההבהרה, הבדיקות, ומציאת הבעיות בקוד שהכלי כתב (מחקר השוק, סעיף 3.2). אם לא אמרו שמותר, שאלי.

### 35.1.3 שעון מוזרק

ארבעה מחמשת התרגילים תלויים בזמן: אסימונים שמתמלאים, משימה שמחכה דקה, ערך שפג, חלון של כפילויות. בדיקה שמחכה
דקה אמיתית היא בדיקה שלא תריצי. לכן כל תרגיל מקבל את השעון כפרמטר, וזה גם הדבר הראשון שמראיינת תשאל עליו: „איך
תבדקי את זה?”.

שתי צורות. כשהקוד צריך רק את השעה הנוכחית, הוא מקבל פונקציה, `now func() time.Time`. כשהוא צריך גם טיימר, הוא
מקבל את הממשק מהחבילה `clock` של התרגיל (הקובץ `clock/clock.go`):

<div dir="ltr" align="left">

```go
// Clock is what code needs from time: the current instant and a timer.
type Clock interface {
	Now() time.Time
	NewTimer(d time.Duration) Timer
}

// Timer is the part of *time.Timer the drills use.
type Timer interface {
	C() <-chan time.Time
	Stop() bool
}
```

</div>

בייצור מעבירים `clock.Real()`, ובבדיקות `clock.NewFake(t0)`. השעון המזויף זז רק כשהבדיקה קוראת ל־`Advance`,
ואז יורים כל הטיימרים שהגיע זמנם. יש בו עוד מתודה אחת שחשובה לבדיקות של מקביליות, `BlockUntil(n)`: היא ממתינה עד
שיש לפחות `n` טיימרים פעילים. בלעדיה, בדיקה יכולה להזיז את הזמן לפני שהקוד הספיק לכוון את הטיימר, ואז הטיימר
יחושב מהזמן החדש ולא יירה לעולם. תראי אותה בעבודה בסעיף 35.4.

### 35.1.4 איך עובדים עם התרגילים

כל תרגיל הוא חבילה במודול אחד, `example.com/ch35-drills`. בכל חבילה, קובץ הבדיקות הוא המפרט, וקובץ הקוד מכיל
טיפוסים, חתימות ו־`panic("TODO")`. בהערה בראש כל קובץ כתובה מסגרת הזמן. לכל תרגיל:

1. הפעילי שעון עצר.
2. קראי את קובץ הבדיקות קודם, וכתבי לעצמך שתי שאלות שהיית שואלת את המראיינת.
3. כתבי בקול, כאילו מישהי מקשיבה. אפשר להקליט.
4. הריצי את הבדיקות של החבילה בלבד, עם גלאי המרוצים.
5. כשהזמן נגמר, עצרי, ורשמי כמה זמן לקח ואיפה נתקעת, גם אם לא סיימת.
6. רק אז השווי לפתרון בתיקייה `solution/`.

הפקודה לתרגיל אחד, למשל הראשון:

<div dir="ltr" align="left">

```bash
cd ~/senior-course/ch35
go test -race ./ratelimit
```

</div>

## 35.2 מגביל קצב

### 35.2.1 הדרישה

מגביל קצב לכל מפתח (לקוח, מפתח API, כתובת): כל מפתח מקבל דלי של `burst` אסימונים שמתמלא בקצב `rate` אסימונים
לשנייה. בקשה לוקחת אסימון אחד; אם אין, היא נדחית, והתשובה אומרת כמה זמן לחכות. הזיכרון צריך להיות חסום, והמגביל
בטוח לשימוש מקבילי. מסגרת הזמן: 25 דקות.

שאלות הבהרה שכדאי לשאול כאן:

- האם המגבלה לכל תהליך, או משותפת לכמה עותקים של השירות? (כאן: לכל תהליך. משותפת, זה פרק 16.)
- מה מחזירים בדחייה? (כאן: `RetryAfter`, כדי שהשכבה מעל תוכל לשלוח את הכותרת `Retry-After`.)
- מה קורה למפתח שלא נראה שעה? (כאן: `Prune`.)

### 35.2.2 הממשק

הממשק, מתוך קוד ההתחלה (הקובץ `ratelimit/ratelimit.go` בתיקייה `starter/`):

<div dir="ltr" align="left">

```go
// Decision is the answer for one request.
type Decision struct {
	Allowed    bool
	RetryAfter time.Duration // when refused: how long until one token has accrued
}

// New returns a limiter. now is the clock: time.Now in production, a fake in tests.
func New(rate float64, burst int, now func() time.Time) (*Limiter, error) {
	panic("TODO: drill 1, see chapter 35, section 35.2")
}

// Allow takes one token from key's bucket, or refuses and says when to come back.
func (l *Limiter) Allow(key string) Decision {
	panic("TODO: drill 1, see chapter 35, section 35.2")
}

// Prune forgets every key whose bucket has refilled completely, and returns how many it
// forgot.
func (l *Limiter) Prune() int {
	panic("TODO: drill 1, see chapter 35, section 35.2")
}
```

</div>

### 35.2.3 הבדיקות שמגדירות אותו

| בדיקה | מה היא דורשת |
| --- | --- |
| `TestNewRejectsBadConfig` | קצב אפס, שלילי, NaN או אינסופי, `burst` אפס ושעון חסר מחזירים `ErrConfig` |
| `TestBurstThenRefuse` | שלוש בקשות עוברות, הרביעית נדחית עם `RetryAfter` של 500 מילי־שניות |
| `TestRefillOverTime` | בדיוק אסימון אחד אחרי חצי שנייה, ולעולם לא יותר מ־`burst` |
| `TestRetryAfterShrinksWithPartialToken` | חצי אסימון שנצבר מקצר את ההמתנה, וההמתנה המדויקת מספיקה |
| `TestKeysAreIndependent` | מפתח אחד לא גוזל ממפתח אחר |
| `TestClockGoingBackwardsEarnsNothing` | שעון שחזר אחורה לא נותן אסימונים, גם כשהוא חוזר קדימה |
| `TestConcurrentCallersShareOneBucket` | אלף בקשות מ־100 goroutines על מפתח אחד: בדיוק 50 עוברות |
| `TestPruneForgetsOnlyFullBuckets` | רק דליים מלאים נשכחים, ומפתח שנשכח מתנהג כמו חדש |

### 35.2.4 הלב של הפתרון

מילוי עצל: אין טיימר ואין goroutine ברקע. כל דלי זוכר כמה אסימונים היו בו ומתי נגע בו מישהו לאחרונה, וכשמגיעה
בקשה הוא מחשב כמה נצברו מאז. מתוך הפתרון (הקובץ `ratelimit/ratelimit.go`):

<div dir="ltr" align="left">

```go
// Allow takes one token from key's bucket, or refuses and says when to come back.
func (l *Limiter) Allow(key string) Decision {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	b, ok := l.buckets[key]
	if !ok {
		// A new key starts full: a client that has never been seen gets its whole burst.
		b = &bucket{tokens: l.burst, last: now}
		l.buckets[key] = b
	}
	l.refill(b, now)
	if b.tokens >= 1 {
		b.tokens--
		return Decision{Allowed: true}
	}
	wait := time.Duration(math.Ceil((1 - b.tokens) / l.rate * float64(time.Second)))
	return Decision{RetryAfter: wait}
}

// refill adds the tokens earned since the bucket was last touched. There is no timer:
// the bucket is brought up to date only when someone asks. A clock that went backwards
// earns nothing and does not move last, so it cannot be used to earn tokens twice.
func (l *Limiter) refill(b *bucket, now time.Time) {
	if elapsed := now.Sub(b.last); elapsed > 0 {
		b.tokens = math.Min(l.burst, b.tokens+elapsed.Seconds()*l.rate)
		b.last = now
	}
}
```

</div>

שלושה פרטים שמראיינת תשים לב אליהם. מפתח חדש מתחיל מלא, כי לקוח שלא נראה מעולם מגיע לו כל ה־burst. הזמן שעבר
נבדק כחיובי, ולכן שעון שחזר אחורה לא נותן כלום ולא מזיז את `last`. וההמתנה מעוגלת למעלה, כדי שמי שמחכה בדיוק
`RetryAfter` לא תידחה שוב בגלל ננו־שנייה.

ו־`Prune` נשען על תובנה אחת: דלי מלא מתנהג בדיוק כמו מפתח שלא נראה מעולם. לכן מותר לשכוח אותו בלי לשנות אף
החלטה, והזיכרון נשאר חסום במספר המפתחות שהיו פעילים לאחרונה.

### 35.2.5 מה מראיינת תשאל

| שאלה | תשובה עם המחיר |
| --- | --- |
| למה דלי אסימונים ולא חלון קבוע? | חלון קבוע מאפשר פי שניים מהמגבלה סביב גבול החלון. דלי מאפשר burst מוגדר ומחליק אחריו. חלון נע עם יומן מדויק, אבל זוכר כל בקשה |
| נעילה אחת לכל המגביל? | פשוט ונכון, וצוואר בקבוק אם יש אלפי בקשות במקביל. הצעד הבא: חלוקה לכמה מפות לפי גיבוב של המפתח, כל אחת עם נעילה |
| למה `float64` לאסימונים? | קצב של חצי אסימון לשנייה צריך שברים. החלופה: לשמור זמן במקום אסימונים, כמו באלגוריתם GCRA, שזוכר רק מתי הבקשה הבאה מותרת |
| ואם יש שלושה עותקים של השירות? | כל אחד נותן את המגבלה, וביחד פי שלושה. מגבלה משותפת צריכה מאגר משותף, למשל סקריפט Lua ב־Redis, כמו בפרק 16 |
| ומה אם Redis למטה? | החלטה מתועדת: לפתוח (לאפשר) עם מדד והתראה, או לסגור. ב־Hookline המגביל נפתח, ראי פרק 37 |

## 35.3 מתזמן משימות

### 35.3.1 הדרישה

מתזמן בתוך תהליך: מגישים משימה עם מזהה ושעה, והוא מריץ אותה בשעה שלה, לכל היותר `workers` במקביל. משימה שנכשלה
מנוסה שוב אחרי המתנה גדלה, עד מספר ניסיונות; שגיאה קבועה לא מנוסה שוב. אפשר להשהות ולהמשיך את כל המתזמן, ולבטל
משימה אחת, גם באמצע ריצה. מסגרת הזמן: 35 דקות. זה התרגיל הארוך מהחמישה.

מצבי המשימה:

<div dir="ltr" align="left">

```mermaid
stateDiagram-v2
  [*] --> Pending: Submit
  Pending --> Running: due, worker free, not paused
  Running --> Succeeded: nil
  Running --> Pending: error, attempts left
  Running --> Failed: permanent or last attempt
  Pending --> Cancelled: Cancel
  Running --> Cancelled: Cancel, attempt returns
  Running --> Pending: Run stopped
```

</div>

### 35.3.2 הממשק

| מתודה | מה היא עושה |
| --- | --- |
| `New(clk, workers)` | מתזמן עם שעון ומגבלת מקביליות |
| `Submit(id, at, task, opts)` | מוסיפה משימה; `ErrDuplicate` למזהה קיים |
| `Cancel(id)` | ממתינה: לא תרוץ. רצה: ה־context שלה מבוטל, והיא מסומנת `Cancelled` כשהיא חוזרת |
| `Pause()`, `Resume()` | אין התחלות חדשות; ניסיונות שרצים ממשיכים |
| `Status(id)` | תמונת מצב: מצב, מספר ניסיונות, מועד הניסיון הבא, השגיאה האחרונה |
| `Done(id)` | channel שנסגר כשהמשימה מגיעה למצב סופי |
| `Run(ctx)` | הלולאה; חוזרת כשה־context מבוטל, אחרי שכל הניסיונות שרצו חזרו |

והאפשרויות של הניסיונות החוזרים, מתוך קוד ההתחלה:

<div dir="ltr" align="left">

```go
// Options control retries. The zero value means one attempt.
type Options struct {
	MaxAttempts int                             // attempts in total, including the first; < 1 means 1
	Backoff     func(attempt int) time.Duration // delay after failed attempt n (1-based); nil means no delay
}
```

</div>

### 35.3.3 התכנון: לולאה אחת, טיימר אחד

התכנון הנאיבי הוא goroutine לכל משימה, שישנה עד הזמן שלה. הוא עובד, אבל השהיה, ביטול ומגבלת מקביליות הופכים
אותו לסבך של channels. התכנון בפתרון הוא לולאה אחת: בכל התעוררות היא מתחילה את כל מה שהגיע זמנו, ככל שיש
workers פנויים, ואז ישנה עד המשימה הבאה, או עד שמשהו משתנה. מתוך הפתרון (הקובץ `scheduler/scheduler.go`):

<div dir="ltr" align="left">

```go
// startDue starts every attempt that may start now and returns a timer for the earliest
// pending task that is not due yet (nil if there is none, or while paused).
func (s *Scheduler) startDue(ctx context.Context, wg *sync.WaitGroup) clock.Timer {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.paused {
		return nil
	}
	now := s.clk.Now()
	pending := make([]*entry, 0, len(s.tasks))
	for _, e := range s.tasks {
		if e.state == Pending {
			pending = append(pending, e)
		}
	}
	// Earliest first, then first submitted: the order is deterministic. A heap would make
	// this O(log n) per change instead of O(n log n) per wake-up.
	sort.Slice(pending, func(i, j int) bool {
		if !pending[i].next.Equal(pending[j].next) {
			return pending[i].next.Before(pending[j].next)
		}
		return pending[i].seq < pending[j].seq
	})
	for _, e := range pending {
		if e.next.After(now) {
			// Not due: sleep until it is. Workers that free up later send a wake-up.
			return s.clk.NewTimer(e.next.Sub(now))
		}
		if s.running == s.workers {
			return nil // a finishing attempt will wake the loop
		}
		s.start(ctx, wg, e)
	}
	return nil
}
```

</div>

את הלולאה מעירים בשני מקרים: טיימר שירה, או הודעה ב־channel בשם `wake`, שכל שינוי שולח אליו (הגשה, המשך, ניסיון
שהסתיים). ל־channel יש מקום להודעה אחת, והשליחה לא חוסמת לעולם: הודעה אחת ממתינה מספיקה, כי הלולאה קוראת את כל
המצב מחדש כשהיא מתעוררת.

כשניסיון חוזר, מחליטים מה הלאה, תחת אותה נעילה:

<div dir="ltr" align="left">

```go
	var permanent permanentError
	switch {
	case e.stopped:
		s.finish(e, Cancelled)
	case err == nil:
		s.finish(e, Succeeded)
	case runCtx.Err() != nil:
		// Shutdown, not a failure of the task: give the attempt back.
		e.state, e.next = Pending, s.clk.Now()
		e.attempt--
	case errors.As(err, &permanent) || e.attempt >= e.opts.MaxAttempts:
		s.finish(e, Failed)
	default:
		var delay time.Duration
		if e.opts.Backoff != nil {
			delay = e.opts.Backoff(e.attempt)
		}
		e.state, e.next = Pending, s.clk.Now().Add(delay)
	}
	s.notify()
```

</div>

הסדר של ה־`case` הוא ההחלטות. ביטול גובר על הכול, גם על הצלחה. ניסיון שנקטע בגלל כיבוי של `Run` לא נספר, והמשימה
חוזרת להמתנה, כי זו לא אשמתה. וניסיון שזרק panic הופך לשגיאה קבועה (בפונקציה `call`), כך שמשימה אחת לא מפילה
את התהליך ולא „מחזיקה” worker לנצח.

### 35.3.4 איך בודקים את זה בלי לחכות

ההמתנה בין ניסיונות היא שנייה ואז שתיים, והבדיקה לא מחכה לאף אחת מהן. מתוך הבדיקה `TestRetryWaitsForBackoff`
(הקובץ `scheduler/scheduler_test.go`):

<div dir="ltr" align="left">

```go
	eventually(t, "first attempt failed", func() bool { st = status(t, s, "a"); return st.Attempts == 1 && st.State == Pending })
	if !st.NextRun.Equal(t0.Add(time.Second)) || !errors.Is(st.Err, errBoom) {
		t.Fatalf("after attempt 1: NextRun %v, Err %v; want t0+1s, boom", st.NextRun, st.Err)
	}
	clk.Advance(999 * time.Millisecond)
	settle()
	if calls.Load() != 1 {
		t.Fatal("retried before the backoff of 1s had passed")
	}
	clk.Advance(time.Millisecond)
	eventually(t, "second attempt failed", func() bool { st = status(t, s, "a"); return st.Attempts == 2 && st.State == Pending })
	if want := clk.Now().Add(2 * time.Second); !st.NextRun.Equal(want) {
		t.Fatalf("after attempt 2: NextRun %v, want %v (backoff doubles)", st.NextRun, want)
	}
	clk.Advance(2 * time.Second)
	st = waitState(t, s, "a", Succeeded)
```

</div>

שני כלים עובדים כאן יחד. הזמן מזויף, ולכן `Advance` של 999 מילי־שניות הוא מיידי. וה־goroutines אמיתיות, ולכן
הבדיקה ממתינה בזמן אמיתי (עד שתי שניות) למצב שהיא מצפה לו, ב־`eventually`. החשוב הוא הסדר: **מחכים שהניסיון
הראשון יירשם, ורק אז מזיזים את הזמן.** אילו הזמן היה זז קודם, המתזמן היה מחשב את מועד הניסיון הבא מהזמן החדש,
והבדיקה הייתה ממתינה לשווא.

למה כאן אין צורך ב־`BlockUntil`? כי המתזמן שומר לכל משימה מועד **מוחלט**. גם אם הזמן זז לפני שהלולאה כיוונה
טיימר, היא תחשב `e.next.Sub(now)` שלילי ותתחיל מיד. טיימר יחסי, כמו של המטמון בסעיף הבא, הוא סיפור אחר.

### 35.3.5 מה מראיינת תשאל

| שאלה | תשובה עם המחיר |
| --- | --- |
| למה מיון בכל התעוררות? | ‏O(n log n) לכל שינוי, וזה בסדר לאלפי משימות. ערימה (`container/heap`) הופכת את זה ל־O(log n), במחיר קוד מסובך יותר |
| מה משמעות השהיה? | כאן: אין התחלות חדשות, וניסיונות שרצים ממשיכים. החלופה, לבטל אותם, הופכת השהיה לביטול; מבררים בהבהרה |
| ביטול באמצע ריצה? | ביטול שיתופי: ה־context מבוטל, והמשימה צריכה להקשיב לו. משימה שמתעלמת ממנו תמשיך עד שתסיים |
| פיזור אקראי בהמתנה? | הפונקציה `Backoff` מוזרקת, ולכן אפשר להעביר פיזור מלא כמו בפרק 14. בבדיקות מעבירים פונקציה בלי אקראיות |
| ואם התהליך קורס? | כל המשימות בזיכרון אובדות. מתזמן עמיד שומר אותן במסד, כמו התור של פרק 14, וזה כבר מעבר לתרגיל |

## 35.4 מטמון עם TTL

### 35.4.1 הדרישה

מטמון בזיכרון עם קיבולת חסומה: לכל ערך זמן חיים (TTL), ברירת מחדל או לכל ערך בנפרד. ערך שפג לא מוחזר, ומוסר
בשני מסלולים: בעצלות, כשמישהו ניגש אליו, ובאופן פעיל, בסריקה תקופתית. כשהמטמון מלא, יוצא הערך שהשתמשו בו הכי
מזמן (LRU). בטוח לשימוש מקבילי, וגנרי בטיפוסים. מסגרת הזמן: 25 דקות.

### 35.4.2 הממשק והמבנה

החתימות, מתוך קוד ההתחלה:

| מתודה | מה היא עושה |
| --- | --- |
| `New[K, V](capacity, ttl, now)` | `ErrConfig` לקיבולת או TTL לא תקינים |
| `Get(key)` | ערך וסימן הצלחה; פגיעה מקדמת את הערך, ערך שפג נמחק |
| `Set`, `SetWithTTL` | כתיבה מחדש מחליפה ערך ותפוגה ומקדמת; TTL אפס או שלילי מוחק |
| `Delete`, `Len` | `Len` סופרת גם ערכים שפגו ועוד לא נמחקו |
| `DeleteExpired()` | המסלול הפעיל: מוחק כל מה שפג, ומחזיר כמה |
| `RunJanitor(ctx, clk, every)` | קורא ל־`DeleteExpired` כל פרק זמן, עד שה־context מבוטל |

המבנה הקלאסי הוא שני אינדקסים על אותם ערכים: מפה מהמפתח לאיבר, ורשימה מקושרת כפולה מהשימוש האחרון לראשון.
מפה נותנת חיפוש ב־O(1), והרשימה נותנת הקדמה ופינוי ב־O(1). החבילה `container/list` מהספרייה הסטנדרטית מספיקה.

### 35.4.3 הלב של הפתרון

מתוך הפתרון (הקובץ `ttlcache/ttlcache.go`):

<div dir="ltr" align="left">

```go
// Get returns the value for key if it is present and not expired. A hit makes the entry
// the most recently used; an expired entry is removed on the spot.
func (c *Cache[K, V]) Get(key K) (V, bool) {
	// A full Mutex, not an RWMutex: even a hit writes (it moves the entry to the front).
	c.mu.Lock()
	defer c.mu.Unlock()
	el, ok := c.items[key]
	if !ok {
		var zero V
		return zero, false
	}
	e := el.Value.(*entry[K, V])
	if !c.now().Before(e.expires) {
		c.remove(el)
		var zero V
		return zero, false
	}
	c.order.MoveToFront(el)
	return e.value, true
}
```

</div>

ההערה הראשונה היא תשובה לשאלה שמראיינים אוהבים: „למה לא `RWMutex`, הרי זו קריאה?”. כי ב־LRU גם פגיעה כותבת: היא
מזיזה את האיבר לראש הרשימה. שתי קריאות במקביל תחת נעילת קריאה היו משנות את הרשימה יחד, וזה מרוץ נתונים.

והסריקה הפעילה, שימי לב לשורה שקוראת את האיבר הבא **לפני** המחיקה:

<div dir="ltr" align="left">

```go
// DeleteExpired removes every expired entry and returns how many it removed. It walks the
// whole cache while holding the lock: O(n). That is acceptable for a cache of thousands of
// entries swept every few seconds; for millions, keep a second index ordered by expiry.
func (c *Cache[K, V]) DeleteExpired() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := c.now()
	n := 0
	for el := c.order.Front(); el != nil; {
		next := el.Next() // read before remove, which clears the links
		if !now.Before(el.Value.(*entry[K, V]).expires) {
			c.remove(el)
			n++
		}
		el = next
	}
	return n
}
```

</div>

### 35.4.4 טיימר יחסי בבדיקה

ה־janitor מכוון טיימר יחסי: „עוד עשר שניות מעכשיו”. כאן הבדיקה כן צריכה לחכות שהטיימר קיים לפני שהיא מזיזה את
הזמן, אחרת „עכשיו” יחושב אחרי ההזזה. מתוך `TestJanitorSweepsOnItsInterval` (הקובץ `ttlcache/ttlcache_test.go`):

<div dir="ltr" align="left">

```go
	// The janitor sets a relative timer; wait until it exists before moving time.
	clk.BlockUntil(1)
	clk.Advance(50 * time.Second) // one sweep, nothing expired yet
	clk.BlockUntil(1)
	if c.Len() != 3 {
		t.Fatalf("Len = %d before expiry, want 3", c.Len())
	}
	clk.Advance(10 * time.Second) // now at the TTL: the next sweep removes all three
```

</div>

הקריאה השנייה ל־`BlockUntil` עושה עבודה כפולה: היא מחכה שהסריקה הראשונה הסתיימה (כי הטיימר הבא נוצר רק אחריה),
ולכן הבדיקה של `Len` שאחריה לא מתחרה בסריקה.

### 35.4.5 מה מראיינת תשאל

| שאלה | תשובה עם המחיר |
| --- | --- |
| עצל או פעיל? | עצל לבד משאיר בזיכרון ערכים שאיש לא יקרא, עד שה־LRU ידחוף אותם. פעיל לבד מחזיר ערך שפג בין שתי סריקות. לכן שניהם |
| הסריקה היא O(n) תחת נעילה? | כן. לאלפי ערכים כל כמה שניות זה בסדר. למיליונים: אינדקס שלישי לפי תפוגה, או דגימה אקראית כמו ש־Redis עושה |
| כשמלא, למה לא לפנות קודם ערך שפג? | זה דורש למצוא אותו, כלומר סריקה או אינדקס לפי תפוגה. LRU מפנה ב־O(1); ערך שפג ייצא בסריקה הבאה |
| שתי בקשות לאותו מפתח חסר? | שתיהן ילכו למקור. זו התפרצות מטמון, והפתרון (`singleflight`) בפרק 12 |
| ערכים גדולים במיוחד? | הקיבולת כאן במספר ערכים. חסימה לפי בתים דורשת לדעת את הגודל של כל ערך |

> **איך Rust עושה את זה:** רשימה מקושרת כפולה היא המקום שבו ה־borrow checker הכי מורגש: לכל צומת שני בעלים,
> ולכן משתמשים בדרך כלל בספרייה מוכנה, או באינדקסים לתוך `Vec` במקום מצביעים. ב־Go אוסף הזבל פותר את
> שאלת הבעלות, ו־`container/list` מספיקה; מה שנשאר באחריותך הוא הנעילה.

## 35.5 מקבל webhooks שבטוח לניסיונות חוזרים

### 35.5.1 הדרישה

השולח מבטיח „לפחות פעם אחת”: הוא שולח שוב כשהוא לא קיבל תשובה, וזה כולל מקרים שבהם הבקשה הראשונה דווקא הצליחה.
המקבל צריך להפוך את זה ל„פעם אחת” מבחינת ההשפעה: משלוחים של אותו מזהה אירוע, בתוך חלון זמן, מעובדים פעם אחת,
**כולל משלוחים כפולים שמגיעים באותו רגע**. זה בדיוק הצד המקבל של Hookline. מסגרת הזמן: 30 דקות.

החוזה, כפי שהבדיקות מגדירות אותו:

| מקרה | תשובה |
| --- | --- |
| המשלוח שעיבד | 200 עם הסטטוס `processed` |
| משלוח של מזהה שכבר עובד בתוך החלון | 200 עם הסטטוס `duplicate` |
| משלוח של מזהה שעובד כרגע | ממתין לראשון, ואז אותה תוצאה |
| העיבוד נכשל או זרק panic | 500, והמזהה **לא** נזכר, כדי שהניסיון הבא יעבד |
| שיטה שאינה POST, בלי מזהה, גוף שאינו JSON, גוף גדול מדי | 405, 400, 400, 413 |

### 35.5.2 שלושה תפקידים למשלוח

הנקודה העדינה היא המשלוחים שמגיעים יחד. בדיקה ואז כתיבה, בשני צעדים נפרדים, תיתן לשניים לעבור יחד. לכן ההחלטה
מה משלוח עושה נעשית בצעד אחד, תחת נעילה, ויש לה שלוש תוצאות: כפילות של מזהה שעובד, „מוביל” שמעבד, או „עוקב”
שממתין למוביל. מתוך הפתרון (הקובץ `webhook/webhook.go`):

<div dir="ltr" align="left">

```go
// claim decides, atomically, what this delivery does: it is a duplicate of a processed id,
// or the leader that processes the id, or a follower that waits for the leader.
func (h *Handler) claim(id string) (c *call, leader, dup bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.prune()
	if _, ok := h.done[id]; ok {
		return nil, false, true
	}
	if c, ok := h.inflight[id]; ok {
		return c, false, false
	}
	c = &call{finished: make(chan struct{})}
	h.inflight[id] = c
	return c, true, false
}

// complete records the outcome and releases the followers. A failure is not remembered,
// so the sender's next retry is processed again.
func (h *Handler) complete(id string, c *call) {
	h.mu.Lock()
	delete(h.inflight, id)
	if c.err == nil {
		h.done[id] = h.now()
		h.order = append(h.order, id)
	}
	h.mu.Unlock()
	close(c.finished)
}
```

</div>

הצורה מוכרת: זה מה שהחבילה `singleflight` עושה, עם שני הבדלים. התוצאה נזכרת אחרי שהעבודה הסתיימה, למשך החלון.
וכישלון לא נזכר, כי „עובד” צריך לסמן הצלחה, לא הגעה.

### 35.5.3 שתי החלטות קטנות שחשובות

- **העיבוד מנותק מהבקשה.** המוביל מעביר למעבד `context.WithoutCancel(r.Context())`. אם השולח מתנתק באמצע, העבודה
  נגמרת ונרשמת, והעוקבים מקבלים את התשובה שלה. בלי זה, ניתוק של השולח היה מבטל עבודה שכמעט הסתיימה.
- **החלון נגזם בלי סריקה.** המזהים נכנסים לתור לפי סדר הסיום, ולכן אלה שיצאו מהחלון הם תמיד תחילת התור. הגיזום
  עולה כמספר המזהים שהוסרו, לא כמספר המזהים שבזיכרון. זו גם התשובה לשאלה „והזיכרון?”.

### 35.5.4 הבדיקה עם goroutines

חמישים משלוחים של אותו אירוע משתחררים יחד משער, והמעבד ישן קצת כדי שהם יחפפו. מתוך
`TestConcurrentDuplicatesAreProcessedOnce` (הקובץ `webhook/webhook_test.go`):

<div dir="ltr" align="left">

```go
	close(gate)
	wg.Wait()
	if c := calls.Load(); c != 1 {
		t.Fatalf("processor called %d times for %d concurrent deliveries of one event, want 1", c, n)
	}
```

</div>

הבדיקה הזאת תופסת מימוש שבודק ואז כותב, כמעט תמיד. כדי לתפוס אותו **תמיד**, יש בדיקה שנייה,
`TestDuplicateWaitsForTheDeliveryInFlight`, שעוצרת את המוביל בתוך המעבד, שולחת כפילות, ומוודאת שהיא לא חוזרת ולא
נכנסת למעבד עד שהמוביל משתחרר.

### 35.5.5 מה מראיינת תשאל

| שאלה | תשובה עם המחיר |
| --- | --- |
| ואם יש שלושה עותקים של המקבל? | הזיכרון של כל תהליך לא רואה את האחרים. צריך מאגר משותף: מפתח ייחודי ב־Redis עם זמן חיים, או, טוב יותר, אילוץ ייחודי במסד שבו נכתבת ההשפעה |
| למה „טוב יותר” במסד? | כי אז רישום המזהה וההשפעה הם טרנזקציה אחת. בזיכרון או ב־Redis, קריסה בין ההשפעה לרישום תגרום לעיבוד כפול |
| כמה ארוך החלון? | לפחות כמו טווח הניסיונות של השולח. ב־Hookline, ניסיונות עם המתנה גדלה ותקרה, ושליחה מחדש ידנית שיכולה להגיע מאוחר מאוד |
| למה 200 לכפילות ולא 409? | מבחינת השולח, כפילות היא הצלחה. 409 נראה כמו כישלון, ושולחים רבים לא מנסים שוב קוד 4xx (Hookline לא מנסה שוב 4xx פרט ל־408 ו־429) |
| ובדיקת חתימה? | חסרה בתרגיל בכוונה. במקבל אמיתי היא באה לפני הכול, וב־Hookline היא בפרק 8 ובאבן הדרך M5 |

## 35.6 הזמנה בלי הזמנה כפולה

### 35.6.1 הדרישה

הזמנה של משאב (חדר, שולחן, מכשיר בדיקה) לטווח זמן, כך ששתי הזמנות של אותו משאב לעולם לא חופפות. הטווח חצי־פתוח,
`[Start, End)`: הזמנה שמסתיימת ב־10:00 ואחת שמתחילה ב־10:00 לא חופפות. שני מימושים לאותו חוזה: בזיכרון, ועם
PostgreSQL. מסגרת הזמן: 35 דקות לשניהם.

החוזה, מתוך קוד ההתחלה (הקובץ `booking/booking.go`):

<div dir="ltr" align="left">

```go
// Store is the contract both implementations meet.
type Store interface {
	// Book reserves resource for [start, end) or returns ErrConflict or ErrInvalid.
	Book(ctx context.Context, resource, who string, start, end time.Time) (Booking, error)
	// Cancel removes a booking, freeing its range.
	Cancel(ctx context.Context, id int64) error
	// List returns the bookings of resource ordered by start.
	List(ctx context.Context, resource string) ([]Booking, error)
}
```

</div>

אותן בדיקות רצות על שני המימושים. הבדיקה המרכזית שולחת 40 בקשות במקביל לאותו חדר, בטווחים חופפים שונים, ודורשת
הצלחה אחת בדיוק ו־39 התנגשויות.

### 35.6.2 בזיכרון: נעילה לכל משאב

הבדיקה וההכנסה חייבות להיות צעד אחד. נעילה אחת לכל המאגר תעבוד, אבל תגרום להזמנות של חדרים שונים לחכות זו לזו.
לכן לכל משאב נעילה משלו, ולמפה של המשאבים נעילה נפרדת שמוחזקת רק כדי למצוא או ליצור משאב. מתוך הפתרון (הקובץ
`booking/memory.go`):

<div dir="ltr" align="left">

```go
// Book checks for an overlap and inserts while holding the resource's lock: the check and
// the insert are one step, so two callers cannot both see a free slot.
func (m *Memory) Book(_ context.Context, name, who string, start, end time.Time) (Booking, error) {
	if err := validate(name, who, start, end); err != nil {
		return Booking{}, err
	}
	r := m.resource(name)
	r.mu.Lock()
	defer r.mu.Unlock()
	// Bookings are sorted and never overlap, so their ends are sorted too: the first
	// booking that ends after start is the only one that can collide.
	i, _ := slices.BinarySearchFunc(r.bookings, start, func(b Booking, t time.Time) int {
		if b.End.After(t) {
			return 1
		}
		return -1
	})
	if i < len(r.bookings) && overlaps(r.bookings[i].Start, r.bookings[i].End, start, end) {
		return Booking{}, ErrConflict
	}
	m.mu.Lock() // lock order: resource, then store. Cancel takes them in the same order.
	m.nextID++
	b := Booking{ID: m.nextID, Resource: name, Who: who, Start: start, End: end}
	m.byID[b.ID] = name
	m.mu.Unlock()
	r.bookings = slices.Insert(r.bookings, i, b)
	return b, nil
}
```

</div>

שני פרטים. ההזמנות של משאב שמורות ממוינות, ומאחר שהן לא חופפות, גם זמני הסיום שלהן ממוינים, ולכן חיפוש בינארי
מוצא את ההזמנה היחידה שיכולה להתנגש. ובתוך `Book` נלקחת גם הנעילה של המאגר, כדי לתת מזהה; לכן יש סדר נעילות קבוע,
משאב ואז מאגר, ו־`Cancel` שומרת עליו. סדר הפוך בשתי פונקציות הוא מתכון לקיפאון.

### 35.6.3 ב־PostgreSQL: המסד אוכף

כאן אין בדיקה באפליקציה בכלל. הסכמה אומרת למסד לסרב לשורה שהמשאב שלה שווה למשאב של שורה קיימת **והטווח שלה חופף**
לטווח שלה. זה אילוץ החרגה (exclusion constraint). מתוך הפתרון (הקובץ `booking/postgres.go`):

<div dir="ltr" align="left">

```go
const Schema = `
CREATE EXTENSION IF NOT EXISTS btree_gist;

CREATE TABLE IF NOT EXISTS bookings (
    id       bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    resource text NOT NULL,
    who      text NOT NULL,
    during   tstzrange NOT NULL CHECK (NOT isempty(during)),
    CONSTRAINT bookings_no_overlap EXCLUDE USING gist (resource WITH =, during WITH &&)
);
`
```

</div>

למה צריך את `btree_gist`? אילוץ החרגה נבנה על אינדקס GiST. החפיפה (`&&`) של טווחים נתמכת ב־GiST מובנה, אבל
שוויון על עמודת `text` הוא פעולה של B-tree. ההרחבה `btree_gist` מוסיפה ל־GiST את אותן פעולות על
טיפוסים רגילים, וכך אפשר לשלב את שניהם באילוץ אחד. כך בדיוק מתואר הצירוף בתיעוד של PostgreSQL 18, בפרק על טיפוסי
טווח, בדוגמה של הזמנת חדרים (נבדק באוקטובר 2026).

ההכנסה עצמה פשוטה, והמיפוי של השגיאה הוא כל העבודה:

<div dir="ltr" align="left">

```go
	b := Booking{Resource: resource, Who: who, Start: start, End: end}
	err := p.DB.QueryRow(ctx,
		`INSERT INTO bookings (resource, who, during) VALUES ($1, $2, tstzrange($3, $4, '[)')) RETURNING id`,
		resource, who, start, end).Scan(&b.ID)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == exclusionViolation {
		return Booking{}, ErrConflict
	}
	if err != nil {
		return Booking{}, err
	}
	return b, nil
```

</div>

קוד השגיאה `23P01` הוא `exclusion_violation`, ממחלקה 23 של הפרות אילוצים. אין `SELECT` לפני ה־`INSERT`, ולכן
אין חלון בין הבדיקה לכתיבה, ואין חשיבות לכמה תהליכים מזמינים יחד.

### 35.6.4 החלופות ומחיריהן

| דרך | איך | מחיר |
| --- | --- | --- |
| נעילה בזיכרון | נעילה לכל משאב, בדיקה והכנסה | עובדת רק בתהליך אחד; קריסה מוחקת הכול |
| `SELECT ... FOR UPDATE` על שורת המשאב | נועלים את שורת החדר, בודקים חפיפה, מכניסים | נכון, אבל כל קוד שכותב הזמנה חייב לזכור לנעול (פרק 11) |
| רמת בידוד `SERIALIZABLE` | בודקים ומכניסים בטרנזקציה, והמסד מזהה התנגשות | נכון, אבל צריך לנסות שוב טרנזקציות שנכשלו, והשגיאות פחות ספציפיות (פרק 11) |
| נעילה מייעצת | `pg_advisory_xact_lock` על גיבוב של המשאב | עובדת בין תהליכים, ושוב תלויה בכך שכל קוד זוכר (פרק 18) |
| אילוץ החרגה | המסד מסרב לחפיפה | הכי חזק: אי אפשר לשכוח אותו. דורש `btree_gist` והרשאה ליצור הרחבה, וקושר את הלוגיקה ל־PostgreSQL |

> **איך Rust עושה את זה:** הצורה של המאגר בזיכרון הייתה `Mutex<HashMap<String, Arc<Mutex<Vec<Booking>>>>>`,
> והקומפיילר היה מכריח אותך לשחרר את הנעילה החיצונית לפני שאת מחזיקה את הפנימית לאורך זמן. ב־Go אותו מבנה כתוב
> בשני שדות `sync.Mutex`, וסדר הנעילות הוא משמעת שלך, לא של הקומפיילר.

### 35.6.5 להריץ את הגרסה של PostgreSQL

הבדיקות של PostgreSQL קוראות את `DATABASE_URL`, ומדלגות כשהוא לא מוגדר. כדי להריץ אותן, מפעילים את המסד של הפרק
(קובץ compose בתיקייה `data/`, עם שם פרויקט קבוע), מריצים, ומורידים:

<div dir="ltr" align="left">

```bash
cd ~/senior-course/ch35
docker compose -p ch35 -f data/compose.yaml up -d --wait
export DATABASE_URL=postgres://course:change-me@localhost:5435/drills
go test -race -v -run 'Booking|Cancel|Double|Different' ./booking
docker compose -p ch35 -f data/compose.yaml down -v
```

</div>

עם המסד, כל בדיקה רצה פעמיים, כתת־בדיקה `memory` ותת־בדיקה `postgres`. בלעדיו, תת־הבדיקות של `postgres` מסומנות
`SKIP` עם ההסבר.

## 35.7 תרגיל סקירת קוד

### 35.7.1 הסבב

בסבב סקירה מקבלים pull request ובו בעיות שתולות, וזמן קצוב. לפי העדות היחידה שהמחקר מצא, כחצי שעה, והציון ניתן על
דירוג הבעיות לפי חומרה, לא על מספרן (מחקר השוק, סעיף 3.2, ביטחון נמוך). כאן מסגרת הזמן 20 דקות, כי זה התרגיל
השלישי של היום.

### 35.7.2 התרגיל

ה־PR מוסיף סינון כפילויות למקבל webhooks. הוא קצר, כ־70 שורות, וקרוב מאוד לתרגיל 4, ולכן תזהי בו דברים. החומר:

| קובץ | מה בו |
| --- | --- |
| `data/review-drill-pr.md` | התיאור של ה־PR, כפי שהכותב שלח אותו |
| `data/review-drill.diff` | השינוי עצמו: קובץ חדש, בדיקה חדשה, ושינוי ב־handler |
| `starter/review.md` | התבנית לסקירה שלך |
| `solution/review-key.md` | מפתח התשובות. לא פותחים לפני שהסקירה שמורה |

השיטה:

1. קראי את התיאור. רשמי מה ה־PR טוען שהוא עושה, ומה צריך להיות נכון כדי שזה יהיה נכון.
2. סדרי את הקבצים לפי סיכון לפני שאת קוראת שורה: מה משנה מתי השולח חושב שסיימנו? מה מצב משותף?
3. קראי לפי הסדר הזה. לכל בעיה: קובץ ושורה, חומרה, ומה היית כותבת לכותב, כולל כיוון לתיקון.
4. סמני חומרה בחמש דרגות: חוסם, כדאי לתקן, שאלה, הערה קטנה, מחמאה. לפחות מחמאה אחת אמיתית.
5. כתבי החלטה ומשפט סיכום.
6. שמרי את הקובץ ועשי commit. רק אז פתחי את המפתח.

### 35.7.3 איך משווים למפתח

המפתח מסדר את הממצאים לפי חומרה, ומראה לכל אחד הערה מוצעת. כשאת משווה, השאלות החשובות הן לא „כמה מצאתי”, אלא:

- האם שלושת הממצאים הראשונים שלך הם שלושת החוסמים שבמפתח?
- האם כל הערה חוסמת שלך אומרת למה זה חשוב, ומציעה כיוון?
- מה החמצת, ומה היה צריך לגרום לך להסתכל שם?

בסוף המפתח יש טבלה שמראה מה כלים היו תופסים לבד. התשובה, כשהרצנו אותם על השינוי: כמעט כלום. זה בדיוק הטיעון
בעד סבב כזה.

## סיכום

- סבב קוד בישראל, נכון לאוקטובר 2026: מטלת בית או קוד חי של 45 עד 90 דקות, מעשי יותר מאלגוריתמי, בביטחון בינוני.
- השיטה: הבהרה, ממשק, בדיקות, גרסה פשוטה ונכונה, קצוות ומקביליות, ומחירים בקול.
- קוד שתלוי בזמן מקבל את השעון כפרמטר. בבדיקה מזיזים זמן מזויף, ומחכים בזמן אמיתי למצב, לא לזמן.
- מגביל קצב: מילוי עצל, אסימונים חלקיים, ושכחה של דליים מלאים כגבול לזיכרון.
- מתזמן: לולאה אחת עם טיימר אחד, ערוץ התעוררות שלא חוסם, וסדר `case` שהוא סדר ההחלטות.
- מטמון: מפה ורשימה, נעילה מלאה כי גם פגיעה כותבת, ופינוי עצל ופעיל יחד.
- מקבל webhooks: ההחלטה מי מעבד נעשית בצעד אחד; הצלחה נזכרת, כישלון לא; במספר עותקים, המזהה עובר למסד.
- הזמנה: נעילה לכל משאב בתהליך אחד, ואילוץ החרגה עם `btree_gist` כשהמסד צריך לאכוף.
- סקירת קוד: לפי סיכון, עם חומרה, עם כיוון. החוסמים קודם.

## מונחים לזכור

| מונח בעברית | באנגלית | במשפט אחד |
| --- | --- | --- |
| קוד חי | Live coding | פתרון בעיה מעשית מול מראיינת, בזמן קצוב, בקול |
| שעון מוזרק | Injected clock | שעון שהקוד מקבל כפרמטר, כדי שבדיקה תוכל להחליף אותו |
| שעון מזויף | Fake clock | שעון שזז רק כשהבדיקה מזיזה אותו, ויורה טיימרים שהגיע זמנם |
| דלי אסימונים | Token bucket | מגבלה שמאפשרת פרץ מוגדר ואחריו קצב קבוע |
| מילוי עצל | Lazy refill | חישוב האסימונים שנצברו רק כשמגיעה בקשה, בלי טיימר |
| ביטול שיתופי | Cooperative cancellation | ביטול שמבקש מהקוד לעצור דרך ה־context, ותלוי בכך שהקוד מקשיב |
| שגיאה קבועה | Permanent error | שגיאה שניסיון חוזר לא יתקן, ולכן לא מנסים שוב |
| זמן חיים (TTL) | Time to live | כמה זמן ערך במטמון נחשב תקף |
| פינוי | Eviction | הוצאה של ערך מהמטמון כדי לפנות מקום |
| LRU | Least recently used | מדיניות פינוי שמוציאה את הערך שהשתמשו בו הכי מזמן |
| חלון כפילויות | Dedupe window | הזמן שבו מזהה שעובד נזכר, ומשלוח חוזר שלו נחשב כפילות |
| מוביל ועוקב | Leader and follower | המשלוח שמעבד, ומשלוחים כפולים שממתינים לתוצאה שלו |
| טווח חצי־פתוח | Half-open range | טווח שכולל את ההתחלה ולא את הסוף, כך שטווחים צמודים לא חופפים |
| אילוץ החרגה | Exclusion constraint | אילוץ במסד שמסרב לשורה שמתנגשת בשורה קיימת לפי אופרטורים שמוגדרים בו |
| סדר נעילות | Lock ordering | כלל שלפיו כל הקוד לוקח נעילות באותו סדר, כדי למנוע קיפאון |

## טעויות נפוצות

- **לקפוץ לקוד לפני ההבהרה.** עשר דקות אחר כך מתגלה שהמגבלה הייתה משותפת לכמה עותקים. חמש דקות של שאלות קודם.
- **`time.Now()` ו־`time.Sleep` בתוך הקוד.** אי אפשר לבדוק בלי לחכות, והבדיקות נעשות איטיות ומהבהבות. מזריקים
  שעון.
- **בדיקה ואז כתיבה בשתי נעילות נפרדות.** בין השתיים נכנס משלוח שני. ההחלטה והרישום, תחת נעילה אחת.
- **`RWMutex` למטמון LRU.** פגיעה מזיזה את האיבר ברשימה, כלומר כותבת. גלאי המרוצים יראה את זה בבדיקה המקבילית.
- **לזכור מזהה לפני שהעיבוד הצליח.** כישלון אחד והאירוע אבוד, כי כל ניסיון חוזר ייענה „כפילות”.
- **לבטל עבודה כשהשולח מתנתק.** ה־context של הבקשה מבוטל כשהיא נגמרת. עבודה שצריכה להסתיים לא רצה עליו.
- **לשתוק כשנתקעים.** המראיינת יכולה לעזור רק אם היא שומעת איפה את. אומרים מה ניסית ומה את חושדת.

## שאלות ראיון

1. **מגביל קצב: דלי אסימונים או חלון נע? מתי כל אחד?** דלי אסימונים זוכר שני מספרים לכל מפתח, מאפשר פרץ מוגדר,
   ומחליק אחריו; הוא מתאים כשמותר ללקוח להתפרץ מעט. חלון נע עם יומן מדויק זוכר כל בקשה בחלון, ולכן מדויק יותר
   ויקר יותר בזיכרון; גרסת המונה המשוקלל שלו זולה ומקורבת. חלון קבוע הכי פשוט, ומאפשר פי שניים מהמגבלה סביב גבול
   החלון. הבחירה תלויה במה שהלקוחות צריכים, לא בטעם.

2. **איך בודקים קוד שתלוי בזמן?** מזריקים את השעון: פונקציה שמחזירה את הזמן, או ממשק עם טיימר. בבדיקה, שעון
   מזויף שזז רק כשמזיזים אותו. בקוד מקבילי מוסיפים כלל: מחכים בזמן אמיתי למצב שמצפים לו, ורק אז מזיזים את הזמן
   המזויף. המחיר: עוד פרמטר בכל בנאי, ועוד חבילה קטנה. התמורה: בדיקות של דקות ושעות שרצות במילי־שניות.

   **המשך:** ומה עם טיימר יחסי, כמו ticker? אז צריך לחכות שהטיימר נוצר לפני ההזזה, למשל עם `BlockUntil`, אחרת
   „עוד עשר שניות” יחושב מהזמן החדש והטיימר לא יירה.

3. **המקבל שלך רץ בשלושה עותקים. מה קורה לסינון הכפילויות?** בזיכרון, כל עותק רואה רק את המשלוחים שהגיעו אליו,
   ומשלוח חוזר שהגיע לעותק אחר יעובד שוב. הפתרון הוא מאגר משותף. מפתח ב־Redis עם `SET NX` וזמן חיים עובד, אבל
   קריסה בין ההשפעה לרישום עדיין גורמת לכפילות. הדרך החזקה: אילוץ ייחודי על מזהה האירוע במסד שבו נכתבת ההשפעה,
   באותה טרנזקציה. המחיר: כתיבה למסד בכל משלוח.

4. **שתי דרכים למנוע הזמנה כפולה, ומה כל אחת עולה.** נעילה: לכל משאב בזיכרון, או נעילת שורת המשאב במסד עם
   `FOR UPDATE`. היא נכונה כל עוד כל קוד שכותב הזמנות זוכר לנעול, והגרסה שבזיכרון לא עובדת בין תהליכים. אילוץ
   החרגה: המסד מסרב לחפיפה, ואי אפשר לשכוח אותו, במחיר הרחבה, הרשאות, ותלות ב־PostgreSQL. בדרך כלל עדיף שהמסד
   יאכוף את מה שאסור שיקרה לעולם.

   **המשך:** ומה אם הזמנה אחת תופסת שני חדרים יחד? באילוץ החרגה מכניסים שתי שורות באותה טרנזקציה, וכישלון של אחת
   מבטל את שתיהן. בנעילות, נועלים את שני החדרים בסדר קבוע, למשל לפי שם, כדי למנוע קיפאון.

5. **למה `Get` במטמון LRU צריך נעילת כתיבה?** כי פגיעה מזיזה את האיבר לראש הרשימה, ושתי פגיעות במקביל תחת נעילת
   קריאה היו משנות את אותה רשימה יחד. החלופות: לוותר על הקדמה בכל פגיעה (קירוב, כמו בחלק מהמטמונים המהירים), או
   לחלק את המטמון לכמה חלקים עם נעילה לכל חלק. כל אחת מוותרת על משהו מהדיוק של ה־LRU או מהפשטות.

6. **נתקעת באמצע ראיון קוד. מה עושים?** אומרים את זה בקול, עם מה שניסית ומה את חושדת. חוזרים לבדיקה הקטנה ביותר
   שנכשלת, ומדפיסים מצב. אם הבאג לא נמצא בכמה דקות, מציעים לפשט: נעילה אחת גדולה במקום עדינות, גרסה בלי
   מקביליות שעוברת, ואז שיפור. מראיינת מעדיפה לראות דרך שיטתית החוצה מאשר שתיקה וקוד מושלם.

## פתק Obsidian

הפתק המוכן נמצא גם כקובץ: `obsidian/35-coding-drills.md`. זה תוכנו:

- **הסבב:** קוד חי של 45 עד 90 דקות, או מטלת בית. מעשי יותר מאלגוריתמי.
- **השיטה:** הבהרה, ממשק, בדיקות, גרסה פשוטה ונכונה, קצוות ומקביליות, מחירים בקול.
- **זמן:** שעון מוזרק. בבדיקה, מחכים למצב ואז מזיזים זמן; לטיימר יחסי, קודם מחכים שהוא קיים.
- **מגביל קצב:** מילוי עצל; דלי מלא שקול למפתח שלא נראה, ולכן מותר לשכוח אותו.
- **מתזמן:** לולאה אחת, טיימר אחד, ערוץ התעוררות שלא חוסם. ביטול גובר על הכול.
- **מטמון:** מפה ורשימה; גם פגיעה כותבת; עצל ופעיל יחד.
- **מקבל webhooks:** ההחלטה בצעד אחד; זוכרים הצלחה, לא הגעה; כמה עותקים, המזהה עובר למסד.
- **הזמנה:** נעילה לכל משאב, או אילוץ החרגה עם `btree_gist`. המסד אוכף מה שאסור שיקרה לעולם.
- **שאלה פתוחה:** ___

<div dir="ltr" align="left">

```text
method      clarify (5) -> API (3) -> tests (7) -> simple + correct (15) -> edges + concurrency (10) -> tradeoffs aloud (5)
clock       now func() time.Time | clock.Clock{Now, NewTimer}; tests: clock.NewFake(t0), Advance, BlockUntil(n)
run         go test -race ./ratelimit   (./scheduler ./ttlcache ./webhook ./booking)
pg          docker compose -p ch35 -f data/compose.yaml up -d --wait
            export DATABASE_URL=postgres://course:change-me@localhost:5435/drills
            docker compose -p ch35 -f data/compose.yaml down -v
limiter     tokens = min(burst, tokens + elapsed*rate); RetryAfter = ceil((1-tokens)/rate)
dedupe      claim under one lock -> leader | follower (wait) | duplicate; remember success only
booking     EXCLUDE USING gist (resource WITH =, during WITH &&)  needs btree_gist; SQLSTATE 23P01
review      risk order -> blocking | should-fix | question | nit | praise -> verdict
my times    limiter ___ scheduler ___ cache ___ webhook ___ booking ___ review ___
```

</div>

## תרגיל

**מטרה:** לפתור חמישה תרגילי קוד ב־Go, שבהם הבדיקות הן המפרט, כל אחד במסגרת הזמן שלו ובקול, כולל הגרסה של
PostgreSQL בתרגיל ההזמנה; ולכתוב סקירה של PR שבור לפני שמסתכלים במפתח. **זמן:** שני ימי לימוד, ימים 94 ו־95, כ־85
דקות בנייה בכל אחד. **ידע מקדים:** פרקים 5, 6, 10, 11, 12, 14 ו־16, ו־Docker לגרסה של PostgreSQL.

הצעדים המלאים, קוד ההתחלה, רשימת הבדיקה, הרמזים והפתרון נמצאים ב־
[`exercises/ch35-drills/README.md`](../exercises/ch35-drills/README.md).

**קריטריון סיום:**

- הבדיקות של מגביל הקצב, המתזמן והמטמון עוברות עם `-race`, והזמן של כל תרגיל רשום בפתק.
- הבדיקות של מקבל ה־webhooks ושל ההזמנה עוברות עם גלאי המרוצים, כולל בדיקת ההזמנה הכפולה מול PostgreSQL.
- הסקירה שלך שמורה ב־commit לפני שפתחת את המפתח, וההשוואה למפתח רשומה בסופה.

## העשרה אופציונלית

- התיעוד של PostgreSQL 18 על טיפוסי טווח, ובו הסעיף על אילוצים על טווחים, ועל ההרחבה `btree_gist`.
- החבילה `golang.org/x/time/rate`: מימוש דלי אסימונים של צוות Go, עם המתנה והזמנה מראש של אסימונים. השווי לתרגיל 1.
- החבילה `golang.org/x/sync/singleflight`: הצורה של מוביל ועוקבים מתרגיל 4, בלי זיכרון אחרי הסיום.
- אם את מכוונת לחברה גדולה שעדיין שואלת שאלות אלגוריתמיות, הספר לא מכסה אותן. מאגר השאלות של interviewing.io,
  שמופיע ברשימת המקורות של מחקר השוק, הוא נקודת פתיחה.

</div>
