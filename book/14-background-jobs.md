---
direction: rtl
lang: he
title: "פרק 14: עבודות רקע וניסיונות חוזרים"
chapter: 14
part: "חלק ג: יסודות backend"
---

<div dir="rtl" lang="he">

# פרק 14: עבודות רקע וניסיונות חוזרים

כמעט כל שירות אמיתי עושה חלק מהעבודה שלו אחרי שהבקשה כבר קיבלה תשובה: שולח מייל, מפיק דוח, קורא לשירות
חיצוני שנופל לפעמים. עבודה כזו צריכה לשרוד קריסות, לנסות שוב בתבונה, ולא להיעשות פעמיים. בפרק הזה תבני
תור עבודות על PostgreSQL עם חכירה, ניסיונות חוזרים עם המתנה גדלה ופיזור אקראי, תור מתים עם הפעלה חוזרת,
מתזמן שלא מריץ פעמיים, ו־workers שנכבים בלי לאבד עבודה. כל אלה הם נושאים שחוזרים בראיונות תכנון מערכות.

## מה נלמד בפרק

- מתי עבודה שייכת לרקע, ומתי זו סתם מורכבות מיותרת.
- תור עבודות על PostgreSQL: מצבים, זמן הרצה, ניסיונות, וחכירה עם `locked_until`.
- ניסיון חוזר (retry) עם המתנה גדלה מעריכית ופיזור אקראי, כפונקציה טהורה ונבדקת.
- אידמפוטנטיות אצל הצרכן, כשהעבודה נוגעת בעולם החיצוני.
- תור מתים (dead-letter queue) והפעלה חוזרת שלא מכפילה עבודה.
- מתזמנים: שעון בתוך התהליך מול CronJob של Kubernetes, וסכנת ההרצה הכפולה.
- כיבוי מסודר של workers על `SIGTERM`.

## ידע מקדים

- פרק 4, סעיף 4.5: עטיפת שגיאות, ושליפת טיפוס שגיאה עם `errors.AsType`.
- פרק 5: goroutines, קבוצת המתנה (`sync.WaitGroup`) וביטול.
- פרק 6, סעיפים 6.4 ו־6.5: לוגים מובנים וכיבוי מסודר.
- פרק 9, סעיף 9.5: מפתחות אידמפוטנטיות.
- פרק 10: PostgreSQL בקונטיינר, והגישה אליו מ־Go.
- פרק 11, סעיפים 11.1, 11.4 ו־11.8: טרנזקציות, `FOR UPDATE SKIP LOCKED`, וניסיון חוזר על כשל סריאליזציה.
  התרגיל כאן כותב תור משלו מאפס, ולא מייבא את הקוד של פרק 11.
- פרק 13, סעיפים 13.7 ו־13.8: טבלת סינון, ולמה „בדיוק פעם אחת” הוא חוזה.

## 14.1 עבודות רקע: מתי ולמה

**עבודת רקע** (background job) היא עבודה שהשירות מבצע מחוץ למחזור של בקשה ותשובה. הבקשה רק רושמת שצריך
לעשות משהו, מחזירה מיד, והעבודה עצמה רצה אחר כך בתהליך אחר. יש ארבע סיבות טובות לזה:

- **זמן.** העבודה ארוכה מהזמן הקצוב של הבקשה, למשל דוח של דקה.
- **אמינות.** העבודה תלויה בשירות חיצוני שנופל, וצריך לנסות שוב גם אחרי שעה.
- **החלקת עומס.** שיא של אלף בקשות בשנייה הופך לתור שנצרך בקצב קבוע.
- **תזמון.** משהו צריך לקרות כל לילה, בלי שמישהו יבקש.

הפיתוי הראשון ב־Go הוא להריץ את העבודה ב־goroutine ולהחזיר תשובה. זה עובד עד הפריסה הראשונה: התהליך מת,
וה־goroutine מתה איתו בלי עקבות. ההבדל בין שלוש הגישות הוא בשאלה מה שורד קריסה:

| גישה | מה שורד קריסה | מתאים ל |
| --- | --- | --- |
| בתוך הבקשה | כלום; הלקוחה מקבלת שגיאה ומנסה שוב | עבודה קצרה שהתוצאה שלה נחוצה עכשיו |
| goroutine אחרי התשובה | כלום, ואף אחד לא יודע | עבודה שמותר לאבד, כמו חימום מטמון |
| תור עבודות עמיד | כל עבודה שנרשמה | כל השאר |

> **איך Rust עושה את זה:** המקבילה היא `tokio::spawn` אחרי שליחת התשובה, ויש לו בדיוק אותה בעיה: המשימה
> חיה בזיכרון של התהליך. העמידות לא באה מהשפה אלא ממקום שנשמר לדיסק לפני שמחזירים תשובה.

למה דווקא PostgreSQL, כשיש Kafka מפרק 13? כי לעבודה בודדת צריך דברים שיומן לא נותן בקלות: ניסיון חוזר עם
עיכוב לעבודה אחת, מצב שאפשר לשאול עליו, והפעלה חוזרת של עבודה ספציפית. ובנוסף, כשהתור באותו מסד של
הנתונים העסקיים, אפשר להוסיף עבודה באותה טרנזקציה של השינוי שגרם לה. זו אותה הגנה מכתיבה כפולה שתבנית
ה־outbox נתנה.

הגבול הוא תפוקה. כל עבודה היא כמה עדכונים של שורה, ו־PostgreSQL צריך לנקות את הגרסאות הישנות שלהן. לתור
בקצב צנוע זה בדרך כלל מספיק; בתפוקה גבוהה מאוד יומן מתאים יותר. איפה עובר הגבול תלוי במסד ובחומרה, ולכן
אם זה חשוב, תמדדי ותרשמי. בייצור יש ספריות שמממשות את התבנית הזאת, ובהן River ל־Go.

## 14.2 תור עבודות על PostgreSQL

לכל עבודה יש מחזור חיים של ארבעה מצבים:

<div dir="ltr" align="left">

```mermaid
stateDiagram-v2
  [*] --> pending: Enqueue
  pending --> running: Claim
  running --> done: Complete
  running --> pending: Fail, attempts left
  running --> dead: Fail, no attempts left
  running --> running: lease expired, reclaimed
  dead --> pending: Replay
```

</div>

הטבלה, מהקובץ `queue/schema.sql`:

<div dir="ltr" align="left">

```sql
CREATE TABLE jobs (
    id            bigserial PRIMARY KEY,
    kind          text        NOT NULL,                 -- which handler runs it
    payload       jsonb       NOT NULL DEFAULT '{}',
    unique_key    text        UNIQUE,                   -- optional: enqueue at most once per key
    state         text        NOT NULL DEFAULT 'pending'
                  CHECK (state IN ('pending', 'running', 'done', 'dead')),
    run_at        timestamptz NOT NULL DEFAULT now(),   -- not before this time
    attempts      int         NOT NULL DEFAULT 0,       -- claims so far; also the fencing token
    max_attempts  int         NOT NULL DEFAULT 5 CHECK (max_attempts > 0),
    locked_by     text,                                 -- worker holding the lease
    locked_at     timestamptz,
    locked_until  timestamptz,                          -- lease end; after it, any worker may reclaim
    last_error    text,
    created_at    timestamptz NOT NULL DEFAULT now(),
    finished_at   timestamptz
);

-- Workers look for due pending jobs, and for running jobs whose lease ran out.
CREATE INDEX jobs_due    ON jobs (run_at, id)    WHERE state = 'pending';
CREATE INDEX jobs_leased ON jobs (locked_until)  WHERE state = 'running';
```

</div>

ארבע עמודות נושאות את כל הלוגיקה:

| עמודה | התפקיד |
| --- | --- |
| `state` | איפה העבודה במחזור החיים |
| `run_at` | לא להריץ לפני הזמן הזה; ניסיון חוזר פשוט מזיז אותו קדימה |
| `attempts` | כמה פעמים העבודה נלקחה, וגם אסימון הגידור (fencing token) של החכירה הנוכחית |
| `locked_until` | סוף החכירה: אחריו כל worker רשאי לקחת את העבודה |

### 14.2.1 חכירה במקום נעילה

התור של פרק 11, סעיף 11.4, החזיק את נעילת השורה כל זמן שהעבודה רצה. זה פשוט ונכון לעבודות קצרות, אבל
לעבודה של דקה זה אומר טרנזקציה פתוחה של דקה: חיבור תפוס מהמאגר, וגרסאות ישנות שאי אפשר לנקות. כאן הנעילה
מחזיקה רק את רגע הלקיחה. העבודה מקבלת **חכירה** (lease): השורה מסומנת `running` עם זמן סיום, הטרנזקציה
נסגרת, וה־handler רץ בלי שום טרנזקציה פתוחה.

אם ה־worker מת, אף אחד לא צריך לנקות אחריו. החכירה פשוט פגה, והשאילתה של הלקיחה רואה את העבודה שוב.
מהקובץ `queue/queue.go`:

<div dir="ltr" align="left">

```go
func Claim(ctx context.Context, db *pgxpool.Pool, worker string, lease time.Duration) (*Job, error) {
	var claimed *Job
	err := pgx.BeginFunc(ctx, db, func(tx pgx.Tx) error {
		for {
			var id int64
			var state State
			var attempts, maxAttempts int
			err := tx.QueryRow(ctx, `
				SELECT id, state, attempts, max_attempts FROM jobs
				WHERE (state = 'pending' AND run_at <= now())
				   OR (state = 'running' AND locked_until < now())
				ORDER BY run_at, id
				LIMIT 1
				FOR UPDATE SKIP LOCKED`).Scan(&id, &state, &attempts, &maxAttempts)
			if errors.Is(err, pgx.ErrNoRows) {
				return nil
			}
			if err != nil {
				return fmt.Errorf("claim: %w", err)
			}
			if state == Running {
				// The previous holder's lease ran out: that attempt failed without a word.
				if _, err := tx.Exec(ctx, `
					INSERT INTO job_attempts (job_id, attempt, worker, started_at, error)
					SELECT id, attempts, locked_by, locked_at, 'lease expired' FROM jobs WHERE id = $1`, id); err != nil {
					return fmt.Errorf("claim: record expired attempt: %w", err)
				}
				if attempts >= maxAttempts {
					if _, err := tx.Exec(ctx, `
						UPDATE jobs SET state = 'dead', last_error = 'lease expired on the last attempt',
						       locked_by = NULL, locked_until = NULL, finished_at = now()
						WHERE id = $1`, id); err != nil {
						return fmt.Errorf("claim: bury job %d: %w", id, err)
					}
					continue // look for another job
				}
			}
			j, err := scanJob(tx.QueryRow(ctx, `
				UPDATE jobs SET state = 'running', attempts = attempts + 1, locked_by = $2,
				       locked_at = now(), locked_until = now() + $3::interval
				WHERE id = $1
				RETURNING `+jobColumns, id, worker, lease))
			if err != nil {
				return fmt.Errorf("claim job %d: %w", id, err)
			}
			claimed = &j
			return nil
		}
	})
	if err != nil {
		return nil, err
	}
	return claimed, nil
}
```

</div>

מה לשים לב אליו:

- **דילוג על נעולות.** הצירוף `FOR UPDATE SKIP LOCKED` לוקח את השורה הראשונה שאף טרנזקציה אחרת לא מחזיקה.
  כמה workers לוקחים במקביל בלי לחכות זה לזה, ושום עבודה לא נלקחת פעמיים.
- **חכירה שפגה.** אם השורה במצב `running` וזמנה עבר, ה־worker הקודם מת או נתקע. הניסיון שלו נרשם כ־`lease expired`.
  אם זה היה הניסיון האחרון, העבודה הולכת ישר לתור המתים, ולא רצה עוד פעם.
- **ספירה בלקיחה.** המונה `attempts` עולה ברגע הלקיחה, לא בסיום. עבודה שמפילה את ה־worker בכל פעם נספרת
  גם היא, ובסוף מגיעה לתור המתים במקום להפיל workers לנצח.

הבדיקה `TestConcurrentClaimsNeverShareAJob` מריצה שמונה goroutines שלוקחות 60 עבודות במקביל, ובודקת שכל
עבודה נלקחה פעם אחת בדיוק.

### 14.2.2 גידור: מי עוד מחזיק בחכירה

חכירה פותחת בעיה שנעילה לא הייתה בה. worker איטי, שהחכירה שלו פגה, עלול להתעורר ולסמן הצלחה על עבודה
שכבר נלקחה על ידי worker אחר. ההגנה היא **אסימון גידור** (fencing token): מספר שעולה בכל לקיחה. כאן זה
המונה `attempts`. כל סיום בודק שהמספר עוד שלו:

<div dir="ltr" align="left">

```go
func finish(ctx context.Context, db *pgxpool.Pool, j *Job, set string, args []any, cause string) error {
	return pgx.BeginFunc(ctx, db, func(tx pgx.Tx) error {
		var lockedBy string
		var lockedAt time.Time
		err := tx.QueryRow(ctx, `
			SELECT locked_by, locked_at FROM jobs
			WHERE id = $1 AND state = 'running' AND attempts = $2
			FOR UPDATE`, j.ID, j.Attempts).Scan(&lockedBy, &lockedAt)
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("job %d attempt %d: %w", j.ID, j.Attempts, ErrLeaseLost)
		}
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `UPDATE jobs SET `+set+`, locked_by = NULL, locked_until = NULL WHERE id = $1`,
			append([]any{j.ID}, args...)...); err != nil {
			return err
		}
		var errText any
		if cause != "" {
			errText = cause
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO job_attempts (job_id, attempt, worker, started_at, error)
			VALUES ($1, $2, $3, $4, $5)`, j.ID, j.Attempts, lockedBy, lockedAt, errText)
		return err
	})
}
```

</div>

ה־worker האיטי מקבל `ErrLeaseLost`, והתוצאה שלו נזרקת. הבדיקה `TestExpiredLeaseIsReclaimedAndFenced` מראה
את זה, ואת ההיסטוריה שנשארת: ניסיון ראשון של `w1`, שהסתיים בחכירה שפגה, וניסיון שני של `w2` שהצליח.
שימי לב שהגידור מגן על המצב במסד בלבד. אם ה־worker האיטי כבר שלח מייל, המייל יצא. לזה מיועד סעיף 14.4.

### 14.2.3 הוספה באותה טרנזקציה

הפונקציה `Enqueue` מקבלת ממשק קטן במקום מאגר חיבורים, וכך אפשר להעביר לה גם טרנזקציה פתוחה:

<div dir="ltr" align="left">

```go
// Querier is satisfied by *pgxpool.Pool and by pgx.Tx, so a job can be enqueued inside
// the same transaction as the business write that caused it.
type Querier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}
```

</div>

המפתח הייחודי (`UniqueKey`) הופך את ההוספה עצמה לאידמפוטנטית: הוספה שנייה עם אותו מפתח מחזירה את העבודה
הקיימת. סעיף 14.6 משתמש בזה כדי שמתזמן לא ייצור עבודה כפולה.

## 14.3 ניסיון חוזר, המתנה גדלה ו־jitter

לא כל כישלון שווה ניסיון נוסף. **כישלון זמני** (transient), כמו זמן קצוב שפג, תשובת 503 או 429, עשוי לעבור
לבד. **כישלון קבוע** (permanent), כמו קלט לא תקין, תשובת 404 או חשבון שנמחק, יחזור על עצמו בכל ניסיון.
ה־handler מסמן כישלון קבוע בעטיפה, והתור שולח עבודה כזו ישר לתור המתים. מהקובץ `queue/permanent.go`:

<div dir="ltr" align="left">

```go
// Permanent wraps err so that Fail moves the job straight to dead.
func Permanent(err error) error { return permanentError{err} }

// IsPermanent reports whether err, or anything it wraps, was made by Permanent.
func IsPermanent(err error) bool {
	_, ok := errors.AsType[permanentError](err)
	return ok
}
```

</div>

לכישלון זמני מחכים לפני הניסיון הבא, וההמתנה גדלה. **המתנה גדלה מעריכית** (exponential backoff) מכפילה את
התקרה בכל ניסיון, עד גבול עליון (cap). בלי גבול, הניסיון העשירי יחכה שעות. בלי הכפלה, שירות שנפל יקבל
ניסיונות בקצב קבוע בדיוק כשהוא מנסה להתאושש.

אבל הכפלה לבדה לא מספיקה. כשאלף עבודות נכשלות באותה שנייה, כי השירות החיצוני נפל, כולן מחכות 100 אלפיות
שנייה, אחר כך 200, אחר כך 400, וכולן חוזרות יחד, בגלים מסונכרנים. **פיזור אקראי (jitter)** שובר את הסנכרון.
בגרסה שנקראת **פיזור מלא** (full jitter), ההמתנה היא מספר אקראי אחיד בין אפס לתקרה. הגלים נמרחים על פני
כל הטווח, והשירות שמתאושש מקבל זרם רציף במקום מכות.

הפונקציה בקובץ `backoff/backoff.go` טהורה: היא לא קוראת לשעון ולא לאקראיות גלובלית. את מקור האקראיות
מזריקים, ולכן אפשר לבדוק כל ערך:

<div dir="ltr" align="left">

```go
func (p Policy) Ceiling(attempt int) time.Duration {
	if p.Base <= 0 || p.Cap <= 0 {
		return 0
	}
	c := min(p.Base, p.Cap)
	for i := 1; i < attempt && c < p.Cap; i++ {
		if c > p.Cap/2 {
			return p.Cap
		}
		c *= 2
	}
	return c
}
```

</div>

הלולאה מכפילה צעד אחר צעד, ונעצרת כשהיא מגיעה לגבול. הזזה בביטים, כמו בפרק 11, סעיף 11.8, פשוטה יותר אבל
גולשת כשמספר הניסיון גדול, וכאן הגבול מגיע מהגדרות ולא קבוע בקוד. ההמתנה עצמה:

<div dir="ltr" align="left">

```go
func (p Policy) Delay(attempt int, rnd Rand) time.Duration {
	c := p.Ceiling(attempt)
	if c <= 0 {
		return 0
	}
	n := int64(c)
	if n < math.MaxInt64 {
		n++ // include the ceiling itself
	}
	return time.Duration(rnd.Int64N(n))
}
```

</div>

עם בסיס של 100 אלפיות שנייה וגבול של 2 שניות, אלה התקרות שהבדיקה `TestCeiling` מצפה להן:

| ניסיון | 1 | 2 | 3 | 4 | 5 | 6 ומעלה |
| --- | --- | --- | --- | --- | --- | --- |
| התקרה | 100ms | 200ms | 400ms | 800ms | 1.6s | 2s |

איך בודקים פונקציה אקראית? בשתי דרכים. הראשונה היא מקור מדומה שמחזיר תמיד את הקצה התחתון או העליון, וכך
בודקים את שני קצות הטווח בדיוק. השנייה היא בדיקה שדומה לבדיקת תכונה (property-based test): אלפי הגרלות
עם זרע קבוע, ולכל אחת בודקים שהיא בטווח. מהקובץ `backoff/backoff_test.go`:

<div dir="ltr" align="left">

```go
func TestDelayStaysInBoundsAndSpreads(t *testing.T) {
	rnd := rand.New(rand.NewPCG(1, 2)) // fixed seed: the test is deterministic
	for attempt := 1; attempt <= 12; attempt++ {
		ceiling := policy.Ceiling(attempt)
		var lowHalf, highHalf int
		for range 2000 {
			d := policy.Delay(attempt, rnd)
			if d < 0 || d > ceiling || d > policy.Cap {
				t.Fatalf("attempt %d: Delay = %v outside [0, %v]", attempt, d, ceiling)
			}
			if d < ceiling/2 {
				lowHalf++
			} else {
				highHalf++
			}
		}
		// Uniform draws land in each half about 1000 times; 800 is a loose floor.
		if lowHalf < 800 || highHalf < 800 {
			t.Errorf("attempt %d: %d draws in the low half and %d in the high half; not uniform", attempt, lowHalf, highHalf)
		}
	}
}
```

</div>

בדיקת „החצי התחתון והחצי העליון” לא מוכיחה התפלגות אחידה; היא תופסת את הטעות הנפוצה, המתנה קבועה בגובה
התקרה. בפרק 26 תכתבי בדיקת תכונה אמיתית לאותה פונקציה, עם ספרייה שמחפשת דוגמה נגדית.

> **שימי לב:** הטיפוס `*rand.Rand` אינו בטוח לשימוש מכמה goroutines. לכן החבילה מספקת את `Global`,
> שקוראת לפונקציות של `math/rand/v2` ברמת החבילה, ואלה בטוחות. הבדיקות מעבירות מקור עם זרע משלהן.

## 14.4 אידמפוטנטיות אצל הצרכן

התור הזה מבטיח „לפחות פעם אחת”, בדיוק כמו קבוצת צרכנים בפרק 13. עבודה תרוץ יותר מפעם אחת בשלושה מקרים:

| מקרה | מה קרה |
| --- | --- |
| קריסה אחרי האפקט ולפני הסיום | ה־handler סיים, התהליך מת לפני `Complete`, והחכירה פגה |
| חכירה שפגה בזמן ריצה | ה־handler איטי מהחכירה, ו־worker אחר לקח את העבודה במקביל |
| הפעלה חוזרת | מפעילה החזירה לתור עבודה שחלק ממנה כבר בוצע |

לכן כל handler חייב להיות **אידמפוטנטי** (idempotent): הרצה שנייה של אותה עבודה לא משנה את התוצאה. יש שלוש
דרכים, מהקלה לקשה:

- **אידמפוטנטי מטבעו.** „קבעי את הסטטוס ל־shipped” בטוח להריץ פעמיים; „הוסיפי 10 ליתרה” לא.
- **רשומה באותה טרנזקציה.** כשהאפקט הוא כתיבה למסד, כותבים באותה טרנזקציה שורה עם מזהה העבודה. זו טבלת
  הסינון מפרק 13, סעיף 13.7.
- **מפתח אידמפוטנטיות למערכת החיצונית.** כשהאפקט הוא קריאה לשירות אחר, שולחים מפתח שנגזר מהעבודה, והשירות
  מסנן. זה הצד המקבל של פרק 9, סעיף 9.5.

ה־handler בקובץ `notify/notify.go` משתמש בשתי האחרונות יחד, כי קריסה יכולה ליפול בין כל שתי שורות:

<div dir="ltr" align="left">

```go
func Handler(db *pgxpool.Pool, s Sender) func(context.Context, queue.Job) error {
	return func(ctx context.Context, j queue.Job) error {
		var p Payload
		if err := json.Unmarshal(j.Payload, &p); err != nil || p.User == "" {
			return queue.Permanent(fmt.Errorf("bad notify payload %s", j.Payload))
		}
		key := fmt.Sprintf("notify-job-%d", j.ID) // stable across attempts and replays

		var done bool
		if err := db.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM notifications WHERE idempotency_key = $1)`, key).Scan(&done); err != nil {
			return err
		}
		if done {
			return nil
		}
		if err := s.Send(ctx, key, p.User, p.Message); err != nil {
			return fmt.Errorf("send to %s: %w", p.User, err)
		}
		_, err := db.Exec(ctx, `
			INSERT INTO notifications (idempotency_key, user_name, message) VALUES ($1, $2, $3)
			ON CONFLICT (idempotency_key) DO NOTHING`, key, p.User, p.Message)
		return err
	}
}
```

</div>

המפתח נגזר ממזהה העבודה, ולכן הוא זהה בכל ניסיון ובכל הפעלה חוזרת. כך נראה כל אחד ממקרי הקריסה:

| הקריסה קרתה | בהרצה הבאה | התוצאה |
| --- | --- | --- |
| לפני `Send` | שום רשומה; שולחים | הודעה אחת |
| אחרי `Send`, לפני ההוספה | שום רשומה; שולחים שוב עם אותו מפתח | הספק מסנן; הודעה אחת |
| אחרי ההוספה, לפני `Complete` | הרשומה קיימת; מדלגים | הודעה אחת |

השורה השנייה תלויה לגמרי בהבטחה של הספק החיצוני. אם הוא לא מקבל מפתח אידמפוטנטיות, החלון בין השליחה
לרישום נשאר פתוח, ואת זה צריך לכתוב בחוזה ובתיעוד: „ייתכן שתקבלו הודעה פעמיים, והמזהה בכותרת יעזור לכם
לסנן”. זה בדיוק מה ש־Hookline, פרויקט הגמר, מבטיח ללקוחות שלו.

## 14.5 תור מתים והפעלה חוזרת

עבודה שנגמרו לה הניסיונות, או שנכשלה בכישלון קבוע, עוברת למצב `dead`. זה **תור מתים** (dead-letter queue),
והוא לא פח אשפה. הוא שומר שלושה דברים: את הראיות (השגיאה האחרונה וכל ההיסטוריה), את העבודה עצמה כדי
להריץ אותה שוב, ואת העובדה שמשהו נכשל, כדי שמישהו יידע. ההחלטה מתקבלת בפונקציה `Fail`:

<div dir="ltr" align="left">

```go
func Fail(ctx context.Context, db *pgxpool.Pool, j *Job, cause error, retryIn time.Duration) (State, error) {
	next := Pending
	set, args := `state = 'pending', run_at = now() + $2::interval, last_error = $3`, []any{retryIn, cause.Error()}
	if IsPermanent(cause) || j.Attempts >= j.MaxAttempts {
		next = Dead
		set, args = `state = 'dead', finished_at = now(), last_error = $2`, []any{cause.Error()}
	}
	if err := finish(ctx, db, j, set, args, cause.Error()); err != nil {
		return "", fmt.Errorf("fail: %w", err)
	}
	return next, nil
}
```

</div>

ההפעלה החוזרת היא עדכון אחד עם תנאי:

<div dir="ltr" align="left">

```go
func Replay(ctx context.Context, db *pgxpool.Pool, id int64) error {
	tag, err := db.Exec(ctx, `
		UPDATE jobs SET state = 'pending', attempts = 0, run_at = now(), finished_at = NULL
		WHERE id = $1 AND state = 'dead'`, id)
	if err != nil {
		return fmt.Errorf("replay job %d: %w", id, err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("replay job %d: %w", id, ErrNotDead)
	}
	return nil
}
```

</div>

התנאי `state = 'dead'` חשוב יותר ממה שנראה. שתי מפעילות שלוחצות „הפעל שוב” באותו רגע, או כפתור שנלחץ פעמיים,
מגיעות לאותה שורה; רק הראשונה מוצאת אותה מתה. השנייה מקבלת `ErrNotDead`, ואף עבודה לא רצה פעמיים בגלל
ההפעלה. המונה `attempts` מתאפס, כי זו הזדמנות חדשה, אבל ההיסטוריה ב־`job_attempts` נשארת. כך נראית
ההיסטוריה של עבודה שנכשלה פעמיים והצליחה בשלישית, בהרצה שבה נכתב הספר:

<div dir="ltr" align="left">

```text
 job_id | attempt | worker |       error       
--------+---------+--------+-------------------
      4 |       1 | w-0    | simulated failure
      4 |       2 | w-0    | simulated failure
      4 |       3 | w-0    | 
```

</div>

בתפעול, תור מתים דורש שלושה הרגלים:

- **התראה על גודל.** מספר העבודות המתות הוא מדד. עלייה פתאומית אומרת שמשהו נשבר, וזה נושא לפרק 25.
- **לתקן לפני שמפעילים שוב.** הפעלה חוזרת של אלף עבודות שנכשלו על באג שעוד לא תוקן מחזירה אלף עבודות מתות.
- **הפעלה חוזרת במנות.** אלף עבודות שחוזרות יחד הן עומס פתאומי על השירות שרק התאושש. מפעילים בהדרגה.

## 14.6 מתזמנים ו־cron

**מתזמן** (scheduler) מריץ משהו לפי לוח זמנים: כל לילה בשתיים, כל חמש דקות. הכלל הראשון: המתזמן רק מוסיף
עבודה לתור, ולא עושה אותה בעצמו. כך עבודה איטית לא מעכבת את התזמון הבא, ועבודה שנכשלה מקבלת ניסיונות חוזרים
ותור מתים כמו כל עבודה אחרת.

הדרך הפשוטה היא שעון בתוך התהליך: `time.Ticker` שמוסיף עבודה בכל תקתוק. הבעיה מופיעה כשמריצים שלושה
עותקים של השירות, כמו שעושים ב־Kubernetes: שלושה שעונים, שלוש עבודות. יש שלוש תשובות:

| פתרון | איך | המחיר |
| --- | --- | --- |
| עותק אחד מתזמן | בחירת מנהיג (leader election) | מנגנון נוסף; פרק 18 |
| כל העותקים, מפתח לכל חלון זמן | הוספה עם `UniqueKey`; רק הראשונה נוצרת | צריך תור עם מפתח ייחודי |
| מתזמן חיצוני | CronJob של Kubernetes | עוד משאב להפעיל; גם הוא לא מושלם |

התרגיל מממש את השני. כל עותק מחשב את תחילת חלון הזמן הנוכחי, ומוסיף עבודה עם מפתח שנגזר ממנו. מהקובץ
`scheduler/scheduler.go`:

<div dir="ltr" align="left">

```go
func Slot(t time.Time, every time.Duration) time.Time {
	return t.UTC().Truncate(every)
}
```

</div>

הבדיקה `TestConcurrentSchedulersCreateOneJobPerSlot` מריצה שלושה מתזמנים במקביל, כל אחד עם שעון משלו,
ובודקת שנוצרה עבודה אחת לכל חלון.

שעון בתוך תהליך לא משלים תקתוקים שהוחמצו. אם השירות היה למטה בשתיים בלילה, הדוח של הלילה לא ייווצר. אם
זה חשוב, המתזמן צריך לבדוק בעלייה אילו חלונות עברו בלי עבודה. ההחלטה, להשלים או לדלג, היא עסקית, וצריך לקבל
אותה במודע.

### 14.6.1 הצצה: CronJob של Kubernetes

ב־Kubernetes יש משאב בשם `CronJob`. הוא יוצר pod, היחידה הקטנה ביותר שרצה באשכול, לפי לוח זמנים
בתחביר של cron. תריצי דברים כאלה ב־kind בפרק 20; כאן רק קוראים קובץ לדוגמה, `data/cronjob-preview.yaml`:

<div dir="ltr" align="left">

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: nightly-report
spec:
  schedule: "0 2 * * *"            # 02:00 every day
  timeZone: "Asia/Jerusalem"
  concurrencyPolicy: Forbid        # skip a run while the previous one is still going
  startingDeadlineSeconds: 600     # missed by more than 10 minutes: skip it, do not catch up
  successfulJobsHistoryLimit: 3
  failedJobsHistoryLimit: 3
  jobTemplate:
    spec:
      backoffLimit: 2
      template:
        spec:
          restartPolicy: Never
          containers:
            - name: enqueue
              image: ghcr.io/YOUR-USER/jobqueue-tools:1.0.0   # placeholder image
              args: ["enqueue-due", "-name", "nightly-report", "-every", "24h"]
              env:
                - name: DATABASE_URL
                  valueFrom:
                    secretKeyRef:
                      name: jobqueue-db
                      key: url
```

</div>

התיעוד של Kubernetes מזהיר, נכון לאוקטובר 2026, שבנסיבות מסוימות CronJob אחד יכול ליצור כמה הרצות לאותו
מועד, או לא ליצור אף אחת, ולכן העבודות צריכות להיות אידמפוטנטיות. שלוש הגדרות בקובץ עוסקות בזה:

| הגדרה | מה היא קובעת |
| --- | --- |
| `concurrencyPolicy: Forbid` | אם ההרצה הקודמת עוד רצה, מדלגים על החדשה |
| `startingDeadlineSeconds` | הרצה שהתאחרה יותר מזה מדולגת, ולא מושלמת בדיעבד |
| `timeZone` | אזור הזמן של לוח הזמנים; בלעדיו, לפי אזור הזמן של kube-controller-manager |

שימי לב שגם כאן הפקודה רק מוסיפה עבודה לתור עם מפתח לחלון. ההגנה מכפילות נשארת בתור, לא במתזמן.

## 14.7 כיבוי מסודר של workers

כש־Kubernetes מחליף גרסה, הוא שולח לכל pod אות כיבוי. האות הוא `SIGTERM`, ואחרי זמן חסד (30 שניות
כברירת מחדל, נכון לאוקטובר 2026) מגיע `SIGKILL`. ב־worker יש באותו רגע עבודות באמצע. **כיבוי מסודר** (graceful shutdown) של worker אומר ארבעה
דברים, בסדר הזה:

- להפסיק לקחת עבודות חדשות.
- לתת לכל עבודה שנלקחה לרוץ עד הסוף, בלי לבטל אותה.
- לרשום את התוצאה שלה במסד, למרות שה־context הראשי כבר בוטל.
- רק אז לצאת.

הלולאה בודקת את ה־context לפני כל לקיחה. ה־handler מקבל context מנותק. מהקובץ `worker/worker.go`:

<div dir="ltr" align="left">

```go
func (p *Pool) loop(ctx context.Context, name string) {
	for ctx.Err() == nil { // stop claiming as soon as shutdown starts
		j, err := queue.Claim(ctx, p.DB, name, p.Lease)
		if err != nil && ctx.Err() == nil {
			p.log().Error("claim failed", "worker", name, "err", err)
		}
		if j == nil {
			select { // nothing to do (or an error): wait, unless shutting down
			case <-ctx.Done():
			case <-time.After(p.Poll):
			}
			continue
		}
		p.runOne(ctx, name, j)
	}
}

// runOne runs a claimed job to the end even if shutdown starts meanwhile. The handler's
// context is detached from ctx (it is not cancelled by SIGTERM) but bounded by the lease:
// after the lease another worker may take the job, so running longer is pointless.
func (p *Pool) runOne(ctx context.Context, name string, j *queue.Job) {
	jobCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), p.Lease)
	defer cancel()
	err := p.call(jobCtx, *j)

	// Record the result even during shutdown: that is the point of draining.
	rec, cancelRec := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancelRec()
	log := p.log().With("worker", name, "job", j.ID, "kind", j.Kind, "attempt", j.Attempts)
	if err == nil {
		if err := queue.Complete(rec, p.DB, j); err != nil {
			log.Error("complete failed", "err", err)
			return
		}
		log.Info("job done")
		return
	}
	delay := p.Backoff.Delay(j.Attempts, p.rand())
	state, ferr := queue.Fail(rec, p.DB, j, err, delay)
	if ferr != nil {
		log.Error("recording failure failed", "err", ferr, "cause", err)
		return
	}
	log.Warn("job failed", "err", err, "state", state, "retry_in", delay)
}
```

</div>

הנקודה העדינה היא שני ה־contexts שנגזרים מ־`context.WithoutCancel`. הראשון נותן ל־handler לרוץ גם אחרי
`SIGTERM`, אבל לא יותר מהחכירה: אחרי החכירה worker אחר רשאי לקחת את העבודה, ואין טעם להמשיך. השני נותן
לרישום התוצאה עשר שניות, גם כשה־context הראשי מבוטל. בלעדיו, עבודה שהסתיימה בזמן הכיבוי תישאר `running`,
החכירה תפוג, והעבודה תרוץ שוב.

הבדיקה `TestGracefulDrainWaitsForTheClaimedJob` מבטלת את ה־context באמצע עבודה, ובודקת שלושה דברים: ש־`Run`
לא חוזרת כל עוד העבודה רצה, שהעבודה נרשמה `done`, ושעבודה שנייה שחיכתה בתור לא נלקחה. וכך זה נראה
בתוכנית `cmd/worker`, בהרצה שבה נכתב הספר. התוכנית קיבלה `SIGTERM` שנייה וחצי אחרי שהתחילה שתי עבודות
של שלוש שניות:

<div dir="ltr" align="left">

```text
time=2026-10-05T12:00:06.814+03:00 level=INFO msg="shutdown requested: no new claims, waiting for running jobs"
time=2026-10-05T12:00:09.525+03:00 level=INFO msg="job done" worker=w-0 job=2 kind=demo attempt=1
time=2026-10-05T12:00:09.525+03:00 level=INFO msg="job done" worker=w-1 job=1 kind=demo attempt=1
time=2026-10-05T12:00:09.525+03:00 level=INFO msg="drained, exiting"
```

</div>

ומה אם העבודה ארוכה מזמן החסד? אז `SIGKILL` מגיע באמצע, החכירה פגה, ו־worker אחר מריץ אותה מההתחלה. זה עוד
מקרה של „לפחות פעם אחת”, ולכן ה־handler חייב להיות אידמפוטנטי גם כאן. עבודות ארוכות באמת מחלקים לשלבים קטנים,
או מאריכים את החכירה בזמן ריצה עם פעימת לב.

## סיכום

- עבודת רקע עמידה נרשמת למקום שנשמר לפני התשובה; goroutine אחרי התשובה מתה עם התהליך.
- תור על PostgreSQL נותן ניסיון חוזר לעבודה בודדת, מצב גלוי, והוספה באותה טרנזקציה של השינוי.
- לקיחה עם `SKIP LOCKED` וחכירה עם `locked_until` מחזיקות נעילה רק לרגע; תהליך שמת משחרר את העבודה כשהחכירה פגה.
- מונה הניסיונות הוא גם אסימון גידור: worker שהחכירה שלו פגה לא יכול לסמן את התוצאה.
- המתנה גדלה מעריכית עם גבול ופיזור מלא מונעת גלים מסונכרנים; כפונקציה טהורה היא נבדקת עד הסוף.
- כל handler אידמפוטנטי: רשומה באותה טרנזקציה, ומפתח אידמפוטנטיות לכל מערכת חיצונית.
- תור מתים שומר ראיות ומאפשר הפעלה חוזרת; התנאי `state = 'dead'` מונע הפעלה כפולה.
- מתזמן רק מוסיף עבודות, עם מפתח לכל חלון זמן; worker נכבה רק אחרי שרשם את מה שלקח.

## מונחים לזכור

| מונח בעברית | באנגלית | במשפט אחד |
| --- | --- | --- |
| עבודת רקע | background job | עבודה שרצה מחוץ למחזור הבקשה, בתהליך אחר |
| תור עבודות | job queue | טבלה או שירות שמחזיקים עבודות עד שמישהו מבצע אותן |
| חכירה | lease | זכות זמנית לעבודה; כשהיא פגה, אחרים רשאים לקחת אותה |
| אסימון גידור | fencing token | מספר שעולה בכל לקיחה, ומונע ממחזיק ישן לכתוב |
| ניסיון חוזר | retry | הרצה נוספת של עבודה שנכשלה |
| כישלון זמני | transient failure | כישלון שעשוי לעבור לבד, כמו זמן קצוב או 503 |
| כישלון קבוע | permanent failure | כישלון שיחזור בכל ניסיון, כמו קלט לא תקין |
| המתנה גדלה | backoff | המתנה בין ניסיונות שגדלה עם מספר הניסיון |
| פיזור אקראי | jitter | אקראיות בהמתנה, כדי שלקוחות רבים לא יחזרו יחד |
| פיזור מלא | full jitter | המתנה אקראית אחידה בין אפס לתקרה |
| אידמפוטנטיות | idempotency | הרצה נוספת של אותה פעולה לא משנה את התוצאה |
| תור מתים | dead-letter queue | עבודות שהתייאשו מהן, שמורות לבדיקה ולהפעלה חוזרת |
| הפעלה חוזרת | replay | החזרה של עבודה מתה לתור, עם ניסיונות חדשים |
| מתזמן | scheduler | רכיב שמוסיף עבודות לפי לוח זמנים |
| כיבוי מסודר | graceful shutdown | עצירה שמסיימת ורושמת את מה שנלקח לפני שהתהליך יוצא |

## טעויות נפוצות

- **להריץ עבודה ב־goroutine אחרי התשובה ולקרוא לזה רקע.** הפריסה הבאה מוחקת אותה בלי עקבות.
- **להחזיק טרנזקציה פתוחה כל זמן ש־handler רץ.** חיבור תפוס, וגרסאות ישנות שלא מתנקות. השתמשי בחכירה.
- **להמתין זמן קבוע בין ניסיונות.** כל מי שנכשל יחד חוזר יחד. הוסיפי פיזור מלא וגבול.
- **לנסות שוב כישלון קבוע.** קלט לא תקין לא יתוקן בניסיון החמישי. עטפי ב־`Permanent`.
- **לבטל את ה־handler על `SIGTERM`.** העבודה נקטעת באמצע ותרוץ שוב. נתקי את ה־context, והגבילי בחכירה.
- **מתזמן בכל עותק בלי מפתח ייחודי.** שלושה עותקים, שלוש הרצות. מפתח לכל חלון זמן, או מנהיג אחד.

## שאלות ראיון

1. **תכנני מערכת עבודות רקע על PostgreSQL. מה בטבלה, ואיך worker לוקח עבודה?** טבלה עם מצב, זמן הרצה, מונה
   ניסיונות וחכירה. כל worker לוקח עם `FOR UPDATE SKIP LOCKED`, כך שהם לא מחכים זה לזה, ומסמן חכירה
   עם זמן סיום בטרנזקציה קצרה. חכירה במקום נעילה ארוכה משחררת את החיבור, ו־worker שמת משחרר את העבודה מעצמו.
   הפשרה היא בתפוקה: לעומס גבוה מאוד יומן כמו Kafka מתאים יותר, במחיר ניסיונות חוזרים מסובכים יותר.

   **המשך:** מה מונע מ־worker איטי לדרוס תוצאה של מי שלקח אחריו? אסימון גידור: מונה הלקיחות נבדק בכל סיום,
   ומחזיק ישן מקבל שגיאה. זה מגן על המסד, לא על מייל שכבר יצא.

2. **למה פיזור אקראי, ולמה דווקא מלא?** כשהרבה לקוחות נכשלים יחד, המתנה מעריכית בלבד מחזירה את כולם יחד,
   בגלים. פיזור מלא בוחר המתנה אחידה בין אפס לתקרה, ומורח את הגלים. היא גם מקצרת חלק מההמתנות, וזה בסדר:
   התקרה עדיין גדלה. מה שחשוב יותר מהנוסחה הוא הגבול, וזיהוי של כישלונות קבועים שלא כדאי לנסות שוב בכלל.

3. **ה־worker קרס באמצע עבודה ששולחת חיוב לספק תשלומים. מה יקרה?** החכירה תפוג, worker אחר ייקח את העבודה,
   והיא תרוץ שוב. לכן ה־handler שולח לספק מפתח אידמפוטנטיות שנגזר מהעבודה, ורושם במסד את מה שהושלם. אם
   הספק לא תומך במפתח, החלון בין החיוב לרישום פתוח, וצריך להחליט במודע: לשאול את הספק לפני חיוב חוזר, או
   לתעד שכפילות אפשרית ולבנות לה תהליך החזר.

4. **שירות עם שלושה עותקים מריץ דוח לילי שלוש פעמים. איך מתקנים?** אפשר לבחור עותק אחד שמתזמן, עם בחירת
   מנהיג. אפשר להעביר את התזמון ל־CronJob, שגם הוא לפי התיעוד עלול להריץ פעמיים. והפשוט ביותר: כל עותק
   מוסיף עבודה עם מפתח ייחודי לחלון הזמן, והמסד יוצר רק אחת. בכל מקרה העבודה עצמה צריכה להיות אידמפוטנטית.

5. **מה עושים עם עבודות שנכשלו סופית?** מעבירים לתור מתים, ולא מוחקים: שומרים את השגיאה ואת ההיסטוריה,
   מתריעים על הגודל, ומאפשרים הפעלה חוזרת. ההפעלה מותנית במצב מת, כדי שלחיצה כפולה לא תריץ פעמיים, ונעשית
   במנות ורק אחרי שהסיבה תוקנה. הפשרה היא בעבודה התפעולית: מישהו צריך להסתכל על התור הזה.

## פתק Obsidian

הפתק המוכן נמצא גם כקובץ: `obsidian/14-background-jobs.md`. זה תוכנו:

- **מתי רקע:** זמן, אמינות, החלקת עומס, תזמון. goroutine אחרי התשובה אינה רקע עמיד.
- **ארבעה מצבים:** ממתינה, רצה, הסתיימה, מתה. ניסיון חוזר מזיז את זמן ההרצה.
- **חכירה:** נעילה רק לרגע הלקיחה; worker שמת משחרר כשהחכירה פגה.
- **גידור:** מונה הניסיונות נבדק בכל סיום; מחזיק ישן לא כותב.
- **המתנה:** תקרה מכפילה עד גבול, ופיזור מלא בין אפס לתקרה; האקראיות מוזרקת.
- **אידמפוטנטיות:** רשומה באותה טרנזקציה, ומפתח לכל מערכת חיצונית.
- **תור מתים:** שומר ראיות; הפעלה חוזרת רק ממצב מת, במנות, אחרי תיקון.
- **מתזמן:** רק מוסיף עבודה, עם מפתח לחלון; CronJob עלול להריץ פעמיים.
- **כיבוי:** להפסיק לקחת, לסיים, לרשום עם context מנותק, ואז לצאת.
- **שאלה פתוחה:** ___

<div dir="ltr" align="left">

```text
docker compose -p ch14 -f data/compose.yaml up -d --wait
export DATABASE_URL=postgres://course:change-me@localhost:5414/jobs
SELECT ... FOR UPDATE SKIP LOCKED      locked_until = now() + $lease
WHERE id = $1 AND state = 'running' AND attempts = $2   -- fencing
ceiling = min(Cap, Base * 2^(attempt-1))     delay = rand in [0, ceiling]
Policy{Base: 100ms, Cap: 2s}: attempt 6 ceiling = ___
UPDATE jobs SET state='pending', attempts=0 WHERE id=$1 AND state='dead'   -- replay
queue.Options{UniqueKey: "name@" + slot}
context.WithTimeout(context.WithoutCancel(ctx), lease)
go run ./cmd/worker -init -enqueue 3 -ms 3000     then: kill -TERM <pid>
docker compose -p ch14 -f data/compose.yaml down -v
```

</div>

## תרגיל

**מטרה:** לבנות תור עבודות עמיד על PostgreSQL: לקיחה עם `SKIP LOCKED` וחכירה, פונקציית המתנה עם גבול ופיזור
מלא ובדיקות לגבולות שלה, handler אידמפוטנטי, תור מתים עם הפעלה חוזרת, מתזמן בלי כפילויות, ו־workers שנכבים
בלי לאבד עבודה. **זמן:** שני ימי לימוד, ימים 41 ו־42, כ־65 דקות בנייה ביום 41 וכ־60 ביום 42. **ידע מקדים:** סעיפים 14.1 עד 14.4 ליום 41; 14.5 עד 14.7 ליום 42; פרקים 4, 5, 6, 9, 10, 11 ו־13
כמפורט בידע המקדים של הפרק.

הצעדים, קוד ההתחלה, הרמזים והפתרון נמצאים ב־
[`exercises/ch14-jobqueue/README.md`](../exercises/ch14-jobqueue/README.md).

**קריטריון סיום:**

- בדיקות ההמתנה בודקות את הטווח של כל ניסיון ואת הגבול, ועוברות.
- עבודה שנכשלה חוזרת לתור עם `run_at` מאוחר יותר, ואחרי הניסיון האחרון עוברת למצב מת.
- הפעלה חוזרת מחזירה עבודה מתה לתור, ופעם שנייה נכשלת עם `ErrNotDead`.
- הכיבוי מחכה לעבודה שנלקחה, רושם אותה `done`, ולא לוקח עבודה חדשה.
- כל הבדיקות עוברות עם `-race` מול PostgreSQL אמיתי, וכלי הבדיקה של התרגיל לא מדפיסים דבר.

## העשרה אופציונלית

- הוסיפי פעימת לב: goroutine שמאריכה את `locked_until` כל שליש חכירה, עם בדיקת גידור. מה קורה אם המסד לא
  זמין לרגע?
- קראי בבלוג הארכיטקטורה של AWS את המאמר על המתנה גדלה ופיזור (`Exponential Backoff And Jitter`). הוא
  משווה כמה גרסאות של פיזור. אל תעתיקי את המספרים שלו; הריצי סימולציה קטנה משלך, ותרשמי.
- כתבי פקודה `enqueue-due` שמשתמשת ב־`scheduler.EnqueueDue`, כמו זו שקובץ התזמון לדוגמה מתאר.
- הוסיפי עדיפות: עמודה `priority`, ומיון לפיה לפני `run_at`. מה קורה לעבודות בעדיפות נמוכה תחת עומס?

</div>
