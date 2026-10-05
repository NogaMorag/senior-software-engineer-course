---
direction: rtl
lang: he
title: "פרק 11: PostgreSQL: טרנזקציות, נעילות ומיגרציות"
chapter: 11
part: "חלק ג: יסודות backend"
---

<div dir="rtl" lang="he">

# פרק 11: PostgreSQL: טרנזקציות, נעילות ומיגרציות

באגים של מקביליות במסד נתונים לא נראים בבדיקה של משתמשת אחת. הם מופיעים בייצור, פעם בכמה אלפי בקשות:
יתרה שירדה פעם אחת במקום פעמיים, משמרת שנשארה בלי אף תורן, עבודה שרצה פעמיים. בפרק הזה לא נקרא עליהם
אלא נייצר אותם בכוונה, בבדיקות Go עם שני חיבורים, ואז נתקן כל אחד בכלי המתאים: נעילת שורה, רמת בידוד
גבוהה עם ניסיון חוזר, או סדר נעילות קבוע. בסוף נוסיף את מה שכל שירות אמיתי צריך: מיגרציות עם goose ומאגר
חיבורים בגודל מחושב.

## מה נלמד בפרק

- מה ACID אומר בפועל ב־PostgreSQL, ואיך תמונות מצב (snapshots) מאפשרות לקוראים לא לחסום כותבים.
- שלוש רמות הבידוד ש־PostgreSQL מממש, ומה כל אחת מונעת, לפי התיעוד של גרסה 18.
- לייצר בבדיקה עדכון אבוד (lost update) ו־write skew, ולראות איך המסד עוצר כל אחד מהם.
- נעילות שורה עם `FOR UPDATE`, ותור עבודות עם `FOR UPDATE SKIP LOCKED`.
- לגרום לקיפאון (deadlock) בכוונה, לקרוא את הדוח של השרת, ולמנוע אותו בסדר נעילות.
- מיגרציות SQL עם goose, משורת הפקודה ומתוך Go.
- לקבוע גודל ל־pool של pgx, ומה משתנה מאחורי PgBouncer.
- פונקציית ניסיון חוזר גנרית לכשלי סריאליזציה, שפרויקט הגמר ישתמש בה.

## ידע מקדים

- פרק 10 כולו, ובעיקר סעיפים 10.1 ו־10.7: המסד בקונטיינר, psql, והספרייה pgx.
- פרק 4, סעיפים 4.5 ו־4.7: עטיפת שגיאות ו־generics. פונקציית הניסיון החוזר נשענת על שניהם.
- פרק 5, סעיף 5.2 וסעיפים 5.4 עד 5.8: channels, context, הכלים של החבילה `sync`, ובדיקת מרוצי נתונים.
- ‏`GOBIN` מוגדר לפי פרק 2, סעיף 2.7. הכלי goose מותקן בצעד הראשון של התרגיל, בפקודת `go install` אחת.

## 11.1 טרנזקציות ו־ACID בפועל

טרנזקציה (transaction) היא קבוצת פקודות שמתנהגת כיחידה אחת: `BEGIN`, הפקודות, ואז `COMMIT` או `ROLLBACK`.
בלי `BEGIN` מפורש, כל פקודה ב־PostgreSQL היא טרנזקציה בפני עצמה (autocommit). ראשי התיבות ACID מתארים
ארבע הבטחות, וכדאי לדעת מה כל אחת אומרת ב־PostgreSQL ממש:

| אות | ההבטחה | איך PostgreSQL מקיים אותה |
| --- | --- | --- |
| A, אטומיות | הכול או כלום | שגיאה באמצע מבטלת את כל הטרנזקציה; אחרי שגיאה, כל פקודה נוספת נכשלת עד `ROLLBACK` |
| C, עקביות | מצב חוקי לפני ואחרי | אילוצים: `CHECK`, ‏`UNIQUE`, מפתחות זרים. רק מה שהגדרת כאילוץ נאכף |
| I, בידוד | טרנזקציות במקביל לא מפריעות זו לזו | רמות בידוד; ברירת המחדל לא מבטיחה בידוד מלא, וזה כל הפרק |
| D, עמידות | מה שאושר לא נעלם בקריסה | יומן כתיבה מראש (WAL) נכתב לדיסק לפני שה־`COMMIT` חוזר |

ה־C הוא האות הכי לא מובנת: המסד לא יודע ש„יתרה לא שלילית” או „תמיד יש תורן” הם כללים, אלא אם אמרת לו. את
הראשון אפשר לכתוב כ־`CHECK`, ובתרגיל תעשי את זה במיגרציה. את השני אי אפשר לבטא כאילוץ על שורה אחת, ולכן
הוא נשען על הבידוד, וזה בדיוק המקום שבו הוא נשבר (סעיף 11.3).

### 11.1.1 תמונות מצב וגרסאות שורה

PostgreSQL משתמש ב־MVCC (בקרת מקביליות מרובת גרסאות): עדכון לא דורס שורה, אלא כותב גרסה חדשה ומסמן את
הישנה כמתה. כל פקודה קוראת לפי תמונת מצב (snapshot), כלומר רשימה של אילו טרנזקציות כבר אושרו מבחינתה, ורואה
את גרסת השורה שמתאימה לה. התוצאה המעשית: **קריאה רגילה לא חוסמת כתיבה, וכתיבה לא חוסמת קריאה.** רק שתי
כתיבות לאותה שורה מחכות זו לזו. זה גם מקור האנומליות: שתי טרנזקציות יכולות לקרוא את אותו ערך ישן, כל אחת
בתמונת המצב שלה, ולהחליט על סמכו.

ב־Go, pgx עוטף את הטרנזקציה ב־`pgx.BeginFunc`: הוא פותח, מריץ את הפונקציה שלך, ומאשר אם היא החזירה `nil`
או מבטל אם היא החזירה שגיאה. כך אין `ROLLBACK` שנשכח בענף שגיאה.

## 11.2 רמות בידוד

תקן SQL מגדיר ארבע רמות בידוד (isolation levels) לפי התופעות שכל אחת מתירה. זו הטבלה מתיעוד PostgreSQL 18,
פרק 13.2, כפי שהיא שם:

| רמה | קריאה מלוכלכת | קריאה שאינה חוזרת | שורות פנטום | אנומליית סריאליזציה |
| --- | --- | --- | --- | --- |
| Read uncommitted | מותרת בתקן, לא ב־PostgreSQL | אפשרית | אפשרית | אפשרית |
| Read committed | לא | אפשרית | אפשרית | אפשרית |
| Repeatable read | לא | לא | מותרת בתקן, לא ב־PostgreSQL | אפשרית |
| Serializable | לא | לא | לא | לא |

התיעוד מוסיף ש־PostgreSQL מקבל את ארבעת השמות, אבל מממש רק שלוש רמות: Read uncommitted מתנהגת בדיוק כמו
Read committed. וזה מה שכל רמה עושה בפועל:

- **‏Read committed (ברירת המחדל).** כל **פקודה** מקבלת תמונת מצב חדשה, ורואה כל מה שאושר לפני שהתחילה. שתי
  קריאות באותה טרנזקציה יכולות לראות ערכים שונים. כש־`UPDATE` נתקל בשורה שטרנזקציה אחרת עדכנה, הוא מחכה לה,
  ואז, כלשון התיעוד, תנאי ה־`WHERE` מוערך מחדש על הגרסה המעודכנת, והעדכון ממשיך ממנה.
- **‏Repeatable read.** תמונת מצב אחת לכל **הטרנזקציה**, מהפקודה הראשונה. אם הטרנזקציה מנסה לעדכן שורה שמישהי
  אחרת עדכנה ואישרה אחרי תחילתה, היא נכשלת בקוד `40001` ובהודעה
  ‏`could not serialize access due to concurrent update`.
- **‏Serializable.** מבטיחה שכל קבוצת טרנזקציות שאושרו תיתן תוצאה זהה להרצה כלשהי שלהן אחת אחרי השנייה. המימוש,
  בידוד תמונת מצב סריאליזבילי (SSI), עוקב אחרי מה שכל טרנזקציה קראה, ומבטל אחת מהן בקוד `40001` כשהוא מזהה
  תלות מעגלית.

התיעוד אומר במפורש, בשתי הרמות הגבוהות, שיישומים שמשתמשים בהן חייבים להיות מוכנים לנסות שוב טרנזקציות
שנכשלו בגלל כשל סריאליזציה (serialization failure). זה לא באג שמטפלים בו כשהוא מופיע; זה חלק מהחוזה. ב־Go
בוחרים רמה בפתיחת הטרנזקציה:

<div dir="ltr" align="left">

```go
	tx, err := pool.BeginTx(context.Background(), pgx.TxOptions{IsoLevel: level})
```

</div>

השורה הזאת מהקובץ `exercises/ch11-isolation-lab/tests/bank/bank_test.go`, שם היא פותחת כל אחת משתי
הטרנזקציות של הבדיקות.

## 11.3 אנומליות מודגמות: עדכון אבוד ו־write skew

בשתי הבדיקות של הסעיף הזה אין goroutines ואין שינה. הבדיקה פותחת שתי טרנזקציות על שני חיבורים מה־pool,
ומריצה את הפקודות שלהן בסדר שהיא בוחרת, שורה אחר שורה. כך האנומליה קורית בכל הרצה, ולא „לפעמים, תחת עומס”.

### 11.3.1 עדכון אבוד

שתי משיכות של 30 מחשבון עם 100. כל אחת קוראת את היתרה, מחשבת ב־Go, וכותבת את התוצאה. זו הבדיקה מהקובץ
`exercises/ch11-isolation-lab/tests/bank/bank_test.go`:

<div dir="ltr" align="left">

```go
	tx1 := begin(t, pool, pgx.ReadCommitted)
	tx2 := begin(t, pool, pgx.ReadCommitted)
	b1, err := bank.ReadBalance(ctx, tx1, 1)
	must(t, err)
	b2, err := bank.ReadBalance(ctx, tx2, 1)
	must(t, err)
	must(t, bank.SetBalance(ctx, tx1, 1, b1-30))
	must(t, tx1.Commit(ctx))
	must(t, bank.SetBalance(ctx, tx2, 1, b2-30)) // does not wait: tx1 already committed
	must(t, tx2.Commit(ctx))

	if got := balance(t, pool, 1); got != 70 {
		t.Fatalf("balance = %d; the lost update should leave 70 (40 would be correct)", got)
	}
```

</div>

שתי הטרנזקציות קראו 100, שתיהן כתבו 70, ושתיהן אושרו. משיכה אחת נעלמה בלי שגיאה ובלי לוג. שימי לב שהבדיקה
**עוברת** כשהאנומליה קורית: היא מתעדת את ההתנהגות של Read committed, לא את מה שהיינו רוצות.

| שלב | טרנזקציה 1 | טרנזקציה 2 | היתרה בטבלה |
| --- | --- | --- | --- |
| 1 | קוראת 100 | | 100 |
| 2 | | קוראת 100 | 100 |
| 3 | כותבת 70, מאשרת | | 70 |
| 4 | | כותבת 70, מאשרת | 70 |

אותו סדר בדיוק תחת Repeatable read נגמר אחרת: בשלב 4 הכתיבה נכשלת בקוד `40001`, כי השורה עודכנה ואושרה
אחרי תחילת הטרנזקציה השנייה. הבדיקה `TestRepeatableReadRejectsTheLostUpdate` באותו קובץ מראה את זה. כלומר
Repeatable read **מזהה** עדכון אבוד, והמחיר הוא שגיאה שהקוד שלך צריך לטפל בה.

### 11.3.2 write skew

ההגנה של Repeatable read נשענת על כך ששתי הטרנזקציות כותבות **לאותה שורה**. write skew הוא המקרה שבו כל אחת
קוראת את אותן שורות, אבל כותבת לשורה אחרת. הדוגמה הקלאסית: במשמרת שני רופאים תורנים, והכלל הוא שתמיד נשאר
לפחות אחד. זו הפונקציה מהקובץ `exercises/ch11-isolation-lab/solution/oncall/oncall.go`:

<div dir="ltr" align="left">

```go
func GoOffCall(ctx context.Context, tx pgx.Tx, shift int, doctor string) error {
	n, err := CountOnCall(ctx, tx, shift)
	if err != nil {
		return err
	}
	if n < 2 {
		return ErrLastOnCall
	}
	_, err = tx.Exec(ctx, `UPDATE oncall SET on_call = false WHERE shift_id = $1 AND doctor = $2`, shift, doctor)
	if err != nil {
		return fmt.Errorf("take %s off call: %w", doctor, err)
	}
	return nil
}
```

</div>

אליס ובוב מבקשים לצאת באותו רגע. כל אחד סופר, בתמונת המצב שלו, שני תורנים, ולכן מותר לו. אליס מעדכנת את
השורה של אליס, בוב את השורה של בוב. אף שורה לא נכתבה פעמיים, אז ל־Repeatable read אין על מה להתלונן, ושתי
הטרנזקציות מאושרות. הבדיקה `TestWriteSkewUnderRepeatableRead` בקובץ `tests/oncall/oncall_test.go` מסתיימת
באפס תורנים.

אותו סדר תחת Serializable נכשל: PostgreSQL רואה שכל טרנזקציה קראה משהו שהשנייה כתבה, תלות מעגלית שאין לה
סדר סדרתי, ומבטל אחת מהן. ההודעה שתראי בפלט של `TestSerializableCatchesWriteSkew`:

<div dir="ltr" align="left">

```text
ERROR: could not serialize access due to read/write dependencies among transactions (SQLSTATE 40001)
```

</div>

> **שימי לב:** write skew הוא מה שהופך „בדקתי ואז כתבתי” למסוכן בכל מקום: הזמנת חדר שלא תופס, שם משתמשת
> שלא קיים, מכסה שלא נוצלה. אם הבדיקה קוראת שורות שהכתיבה לא נוגעת בהן, Repeatable read לא מגן עלייך.

## 11.4 נעילות שורה: `FOR UPDATE` ו־`SKIP LOCKED`

### 11.4.1 ‏`FOR UPDATE`

‏`SELECT ... FOR UPDATE` קורא שורה ונועל אותה עד סוף הטרנזקציה. טרנזקציה אחרת שמבקשת את אותה נעילה מחכה,
וכשהיא ממשיכה, היא קוראת את הגרסה החדשה. זה התיקון לעדכון האבוד, והוא שינוי של מילה אחת בשאילתה. מהקובץ
`exercises/ch11-isolation-lab/solution/bank/bank.go`:

<div dir="ltr" align="left">

```go
// WithdrawLocked is Withdraw with the read turned into SELECT ... FOR UPDATE: the second
// caller waits for the first to commit and then reads the new balance.
func WithdrawLocked(ctx context.Context, pool *pgxpool.Pool, id, amount int64) error {
	return pgx.BeginFunc(ctx, pool, func(tx pgx.Tx) error {
		balance, err := ReadBalanceForUpdate(ctx, tx, id)
		if err != nil {
			return err
		}
		if balance < amount {
			return ErrInsufficientFunds
		}
		return SetBalance(ctx, tx, id, balance-amount)
	})
}
```

</div>

איך בודקים ש„הטרנזקציה השנייה מחכה” בלי לישון ולקוות? שואלים את השרת. התצוגה `pg_stat_activity` מראה לכל
חיבור על מה הוא מחכה, והפונקציה `testdb.WaitForLockWaiters` בתרגיל מחכה עד שמופיע חיבור עם
`wait_event_type = 'Lock'`. רק אז הבדיקה ממשיכה לאשר את הטרנזקציה הראשונה. הבדיקה מוכיחה שההמתנה קרתה, לא
מניחה אותה.

יש תיקון פשוט עוד יותר, כשהחישוב פשוט: לתת למסד לעשות אותו. הפונקציה `WithdrawAtomic` באותו קובץ שולחת
פקודת עדכון אחת, שמפחיתה את הסכום רק אם היתרה מספיקה (התנאי `balance >= $2`), ובודקת אם שורה השתנתה. זה
עובד גם ב־Read committed, בזכות ההערכה מחדש של התנאי שתיארנו בסעיף 11.2: העדכון השני מחכה, ואז בודק את
התנאי מול היתרה המעודכנת.

כלל אצבע: אם ההחלטה נכנסת בפקודת `UPDATE` אחת, זו הדרך. אם צריך לקרוא, להחליט בקוד, ואולי לגעת בכמה
טבלאות, זה `FOR UPDATE`. שתיהן נבדקות בתרגיל עם 25 משיכות מקבילות על חשבון של 100: בדיוק 20 מצליחות.

### 11.4.2 ‏`SKIP LOCKED`: תור עבודות

תור עבודות על טבלה הוא דפוס נפוץ: כמה workers, כל אחד לוקח את העבודה הבאה. עם `FOR UPDATE` רגיל, כולם מנסים
לנעול את אותה שורה ראשונה, ומחכים זה לזה. התוספת `SKIP LOCKED` אומרת: שורה שנעולה עכשיו, דלגי עליה. זה החלק
המרכזי של `Work` מהקובץ `exercises/ch11-isolation-lab/solution/queue/queue.go`:

<div dir="ltr" align="left">

```go
		var job Job
		err := tx.QueryRow(ctx, `
			SELECT id, payload FROM jobs
			WHERE status = 'queued'
			ORDER BY id
			LIMIT 1
			FOR UPDATE SKIP LOCKED`).Scan(&job.ID, &job.Payload)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
```

</div>

כל worker נועל את העבודה הוותיקה ביותר שאף אחד לא מחזיק, מריץ את המטפל באותה טרנזקציה, ומסמן אותה כגמורה.
אם המטפל נכשל, הטרנזקציה מתבטלת, הנעילה משתחררת, והעבודה חוזרת לתור. התיעוד של `SELECT` מזהיר שדילוג על
שורות נעולות נותן תמונה לא עקבית של הנתונים, ולכן הוא לא מתאים לעבודה כללית, אבל מתאים בדיוק למצב של כמה
צרכנים על טבלה שמתנהגת כתור. הבדיקה בתרגיל מריצה שמונה workers על 300 עבודות, ודורשת שכל עבודה טופלה פעם
אחת בדיוק.

החיסרון של הגרסה הזאת: הנעילה, והחיבור מה־pool, מוחזקים כל זמן שהמטפל רץ. לעבודה של אלפיות שנייה זה בסדר.
לעבודה שקוראת לשירות חיצוני לעשר שניות, לא. פרק 14, סעיף 14.2, מחליף את זה בחכירה (lease): טרנזקציה קצרה
שמסמנת „שלי עד שעה X”, עבודה מחוץ לטרנזקציה, וטרנזקציה קצרה שנייה לסיום.

> **איך Rust עושה את זה:** ‏`FOR UPDATE` דומה ל־`Mutex::lock`, שהנעילה שלו חיה עד סוף התחום; כאן התחום
> הוא הטרנזקציה. ההבדל: הקומפיילר לא יודע על נעילות במסד, ושום בודק השאלות לא ימנע ממך
> להחזיק נעילת שורה בזמן קריאת רשת. המשמעת היא שלך.

## 11.5 קיפאון (deadlock)

קיפאון הוא מעגל המתנה: טרנזקציה 1 מחזיקה שורה אחת ומחכה לשנייה, וטרנזקציה 2 מחזיקה את השנייה ומחכה לראשונה. אף אחת
לא תתקדם לבד. בבנק זה קורה בהעברות בכיוונים הפוכים, כשכל אחת מחייבת קודם את החשבון שממנו היא מעבירה. הבדיקה
`TestOppositeTransfersDeadlock` בונה את המעגל בכוונה. זה הקטע מהקובץ
`exercises/ch11-isolation-lab/tests/bank/bank_test.go`:

<div dir="ltr" align="left">

```go
	must(t, bank.AddToBalance(ctx, tx1, 1, -10)) // tx1 holds row 1
	must(t, bank.AddToBalance(ctx, tx2, 2, -10)) // tx2 holds row 2

	first := make(chan error, 1)
	go func() { first <- bank.AddToBalance(ctx, tx1, 2, 10) }() // tx1 waits for row 2
	testdb.WaitForLockWaiters(t, pool, 1)
	err2 := bank.AddToBalance(ctx, tx2, 1, 10) // tx2 waits for row 1: a cycle
	err1 := <-first
```

</div>

PostgreSQL לא מחפש מעגלים כל הזמן. חיבור שמחכה לנעילה יותר מ־`deadlock_timeout` (ברירת המחדל שנייה אחת,
ואפשר לראות את הערך ב־psql עם הפקודה `SHOW`) מריץ בדיקה. אם יש מעגל, אחת הטרנזקציות מבוטלת עם SQLSTATE ‏`40P01`,
והשנייה ממשיכה. התיעוד מציין שקשה לחזות איזו מהן תבוטל, ושאין להסתמך על כך. לכן הבדיקה דורשת „בדיוק אחת”,
בלי לקבוע איזו. והשרת כותב ללוג דוח מפורט; בתרגיל תקראי אותו מהלוגים של הקונטיינר, בפקודה שב־README. כך
הוא נראה (מספרי התהליכים והטרנזקציות יהיו אחרים אצלך):

<div dir="ltr" align="left">

```text
ERROR:  deadlock detected
DETAIL:  Process 119 waits for ShareLock on transaction 835; blocked by process 120.
	Process 120 waits for ShareLock on transaction 836; blocked by process 119.
	Process 119: UPDATE accounts SET balance = balance + $2 WHERE id = $1
	Process 120: UPDATE accounts SET balance = balance + $2 WHERE id = $1
HINT:  See server log for query details.
CONTEXT:  while updating tuple (0,1) in relation "accounts"
```

</div>

הביטוי „ShareLock on transaction” הוא הדרך של השרת לומר „מחכה שהטרנזקציה שמחזיקה את השורה תסתיים”. השורות
של `Process` מראות את שתי הפקודות שנתקעו, ובשירות אמיתי זה בדרך כלל כל מה שצריך כדי למצוא את שני מסלולי הקוד.

ההגנה הטובה ביותר, כלשון התיעוד, היא לוודא שכל היישומים נועלים אובייקטים מרובים באותו סדר. זה `Transfer`
מהקובץ `exercises/ch11-isolation-lab/solution/bank/bank.go`:

<div dir="ltr" align="left">

```go
// Transfer locks both rows in ascending id order before changing either, so every
// transfer takes its locks in the same global order and no cycle can form.
func Transfer(ctx context.Context, pool *pgxpool.Pool, from, to, amount int64) error {
	return pgx.BeginFunc(ctx, pool, func(tx pgx.Tx) error {
		first, second := min(from, to), max(from, to)
		if _, err := ReadBalanceForUpdate(ctx, tx, first); err != nil {
			return err
		}
		if _, err := ReadBalanceForUpdate(ctx, tx, second); err != nil {
			return err
		}
		if err := AddToBalance(ctx, tx, from, -amount); err != nil {
			return err
		}
		return AddToBalance(ctx, tx, to, amount)
	})
}
```

</div>

שתי העברות הפוכות נועלות שתיהן קודם את החשבון עם המזהה הנמוך. השנייה מחכה בנעילה הראשונה, לפני שהחזיקה
משהו, ולכן אין מעגל. הבדיקה מריצה 40 העברות מקבילות בשני הכיוונים, וכולן מצליחות.

כשאי אפשר להבטיח סדר (למשל, הסדר תלוי בנתונים שעוד לא קראת), יש שתי הגנות משלימות: לנסות שוב טרנזקציה
שנכשלה ב־`40P01`, כמו שהתיעוד ממליץ ופונקציית הניסיון החוזר בסעיף 11.8 עושה, ולהקטין את הזמן שבו נעילות
מוחזקות. טרנזקציה קצרה היא טרנזקציה שקשה לתפוס אותה במעגל.

## 11.6 מיגרציות עם goose

מיגרציה (migration) היא שינוי סכמה שנשמר כקובץ במאגר הקוד, רץ בסדר קבוע, ונרשם במסד כדי שלא ירוץ פעמיים.
goose הוא כלי מיגרציות ל־Go שעובד עם קובצי SQL פשוטים. כל קובץ מכיל חלק `Up` וחלק `Down`. זה הקובץ
`exercises/ch11-isolation-lab/solution/migrations/00002_oncall.sql`:

<div dir="ltr" align="left">

```sql
-- +goose Up
-- The write-skew example: a shift must always have at least one doctor on call.
CREATE TABLE oncall (
    shift_id int NOT NULL,
    doctor   text NOT NULL,
    on_call  boolean NOT NULL DEFAULT true,
    PRIMARY KEY (shift_id, doctor)
);

-- +goose Down
DROP TABLE oncall;
```

</div>

המספר בתחילת השם הוא הגרסה. goose שומר בטבלה `goose_db_version` אילו גרסאות הוחלו, ומריץ רק את מה שחסר.
כברירת מחדל כל קובץ SQL רץ בתוך טרנזקציה, כך שמיגרציה שנכשלה באמצע לא משאירה חצי טבלה. פקודה שלא יכולה לרוץ
בטרנזקציה, כמו `CREATE INDEX CONCURRENTLY`, דורשת את ההערה `-- +goose NO TRANSACTION` בראש הקובץ.

### 11.6.1 משורת הפקודה

goose קורא את פרטי החיבור ממשתני סביבה, ואז הפקודות קצרות:

<div dir="ltr" align="left">

```bash
export GOOSE_DRIVER=postgres
export GOOSE_DBSTRING=postgres://course:change-me@localhost:5411/lab
export GOOSE_MIGRATION_DIR=migrations
goose status
goose up
goose down
goose down-to 0
goose -s create accounts_balance_nonnegative sql
```

</div>

| פקודה | מה היא עשתה |
| --- | --- |
| `goose status` | רשימת המיגרציות, ולכל אחת מתי הוחלה או `Pending` |
| `goose up` | מחילה את כל מה שחסר, לפי הסדר |
| `goose down` | מבטלת את המיגרציה האחרונה בלבד |
| `goose down-to 0` | מבטלת את כולן, מהאחרונה לראשונה |
| `goose -s create <name> sql` | יוצרת קובץ חדש עם המספר הסידורי הבא (בלי `-s` השם יתחיל בחותמת זמן) |

### 11.6.2 מתוך Go

שירות לא צריך להסתמך על כך שמישהי הריצה `goose up` לפני הפריסה. בתרגיל המיגרציות מוטמעות בקוד עם `embed`,
וגם הבדיקות מריצות אותן בעצמן על מסד ריק. זה הקטע מהקובץ `exercises/ch11-isolation-lab/solution/db/db.go`:

<div dir="ltr" align="left">

```go
func withMigrator(pool *pgxpool.Pool, fn func(*goose.Provider) error) error {
	sqlDB := stdlib.OpenDBFromPool(pool) // closing it does not close the pool
	defer sqlDB.Close()
	locker, err := lock.NewPostgresSessionLocker()
	if err != nil {
		return err
	}
	provider, err := goose.NewProvider(goose.DialectPostgres, sqlDB, migrations.FS, goose.WithSessionLocker(locker))
	if err != nil {
		return fmt.Errorf("goose provider: %w", err)
	}
	return fn(provider)
}
```

</div>

שלושה דברים קורים כאן:

- הכלי goose עובד מעל `database/sql`. הפונקציה `stdlib.OpenDBFromPool` נותנת לו חיבורים מאותו מאגר, בלי מאגר נוסף.
- המשתנה `migrations.FS` מכיל את כל קובצי ה־SQL, כך שהבינארי נושא את הסכמה שלו ולא תלוי בתיקייה שממנה הוא רץ.
- נועל הסשן (session locker) לוקח advisory lock (נעילה על מספר שהיישום בוחר) לכל משך ההרצה. שני עותקים של שירות שעולים
  באותו רגע לא יריצו את אותה מיגרציה פעמיים: השני מחכה לראשון. הנושא הזה מפורט בפרק 18, סעיף 18.2.

### 11.6.3 כללים למיגרציות

- **לא עורכים מיגרציה שכבר הוחלה** באף סביבה. goose לא יריץ אותה שוב, והסביבות יתפצלו בשקט. מתקנים במיגרציה חדשה.
- **‏`Down` הוא כלי פיתוח.** בייצור, ביטול של `DROP COLUMN` לא מחזיר את הנתונים. בודקים שהוא עובד (בתרגיל יש
  בדיקה שעולה, יורדת לאפס ועולה שוב), אבל מתכננים קדימה.
- **שינוי לא תואם נעשה בשלבים.** שינוי שם של עמודה בזמן שגרסה ישנה של הקוד עדיין רצה ישבור אותה. הדפוס של
  הרחבה ואז כיווץ (expand and contract) נמצא בפרק 30, סעיף 30.1.

## 11.7 מאגר חיבורים

כל חיבור ל־PostgreSQL הוא תהליך מערכת הפעלה בשרת, עם זיכרון משלו. לכן יש לשרת גבול, `max_connections`,
שברירת המחדל שלו 100 (גם אותו רואים ב־psql עם `SHOW`). פתיחת חיבור היא עבודה לא זניחה לשרת (תהליך
חדש, אימות, טעינת קטלוג), ולכן שירות לא פותח חיבור לכל בקשה, אלא מחזיק pool. אלה הגבולות המפורשים מהקובץ
`exercises/ch11-isolation-lab/solution/db/db.go`:

<div dir="ltr" align="left">

```go
	// Upper bound on connections this process opens. Every replica of a service has
	// its own pool, so replicas times MaxConns must stay below the server's
	// max_connections, with room left for migrations, psql and monitoring.
	cfg.MaxConns = 10
	cfg.MinConns = 0
	cfg.MaxConnLifetime = time.Hour       // recycle, so server-side memory does not grow forever
	cfg.MaxConnIdleTime = 5 * time.Minute // give idle connections back to the server
	cfg.HealthCheckPeriod = 30 * time.Second
	cfg.ConnConfig.ConnectTimeout = 5 * time.Second
```

</div>

בלי השורה הראשונה, pgx קובע כברירת מחדל את הגדול מבין 4 למספר המעבדים במכונה. בקונטיינר על שרת גדול זה
יכול להיות מספר גבוה מאוד, והוא מוכפל במספר העותקים של השירות. חשבון הגודל נראה כך:

| גורם | שאלה |
| --- | --- |
| גבול השרת | מה `max_connections`, ומה שמור למנהלת, לניטור ולמיגרציות? |
| עותקים | כמה replicas של השירות רצים בשיא, כולל בזמן פריסה, כשישנים וחדשים חיים יחד? |
| לכל עותק | ‏`MaxConns` כפול מספר העותקים חייב להישאר מתחת לשארית |
| בפועל | המסד לא עובד מהר יותר עם יותר חיבורים פעילים מאשר יש לו ליבות ודיסק; מעבר לזה השאילתות רק מחכות בתור בשרת במקום ב־pool |

ומה קורה כשה־pool מלא? הבקשה הבאה מחכה לחיבור שיתפנה, עד שה־context שלה פג. הבדיקה `TestPoolWaitsAtMaxConns`
בתרגיל מראה את זה עם pool של שניים: החיבור השלישי מחזיר `context deadline exceeded`, וכשמשחררים אחד, הוא
מתקבל מיד. זו הסיבה שטרנזקציה ארוכה, או `rows` שלא נסגרו (פרק 10, סעיף 10.7.2), מורגשות כ„המסד איטי” אצל
בקשות אחרות לגמרי.

### 11.7.1 ‏PgBouncer ומצב טרנזקציה

כשיש הרבה שירותים, או מאות עותקים, מציבים לפני PostgreSQL מתווך חיבורים כמו PgBouncer. במצב טרנזקציה
(transaction pooling) הוא נותן ללקוח חיבור אמיתי לשרת רק לזמן טרנזקציה אחת, ואז מעביר אותו ללקוח אחר. כך אלף
חיבורי לקוח חולקים כמה עשרות חיבורי שרת. המחיר: כל מה שקשור **לסשן** ולא לטרנזקציה מפסיק לעבוד, כי
הטרנזקציה הבאה שלך עשויה לרוץ על חיבור אחר. לפי טבלת התאימות בתיעוד של PgBouncer, אלה לא נתמכים במצב הזה:

- הפקודות `SET` ו־`RESET` ברמת סשן, למשל לשינוי נתיב החיפוש של סכמות. הגרסה המקומית לטרנזקציה, `SET LOCAL`, עובדת.
- האזנה להתראות עם `LISTEN`.
- נעילות מייעצות ברמת סשן, כמו זו ש־goose לוקח (מריצים מיגרציות ישירות מול השרת, לא דרך PgBouncer).
- הפקודה `PREPARE` כפקודת SQL, טבלאות זמניות שחיות לאורך הסשן, וסמנים שנשמרים אחרי הטרנזקציה.

ויש עוד נקודה שנוגעת ל־pgx ישירות: כברירת מחדל pgx מכין כל שאילתה ושומר אותה במטמון על החיבור (prepared
statements ברמת הפרוטוקול). PgBouncer תומך בזה במצב טרנזקציה רק אם הגדירו לו `max_prepared_statements` גדול
מאפס. אחרת מגדירים ב־pgx מצב הרצה שלא משאיר הכנות על החיבור (`QueryExecModeExec` או הפרוטוקול הפשוט). אין
PgBouncer בתרגיל; מספיק שתדעי לשאול את השאלות האלה כשתפגשי אותו.

## 11.8 ניסיון חוזר על כשל סריאליזציה

ראינו ש־Serializable עוצר write skew על ידי ביטול אחת הטרנזקציות. כדי שזה יהיה שימושי, הקוד צריך לנסות שוב.
התיעוד של PostgreSQL אומר שני דברים חשובים על זה:

- חשוב לנסות שוב את **כל** הטרנזקציה, כולל כל הלוגיקה שמחליטה איזה SQL לשלוח ואילו ערכים, ולכן PostgreSQL
  לא מציע ניסיון חוזר אוטומטי.
- ‏`40001` מנסים שוב בלי תנאי; כדאי לנסות שוב גם קיפאון (`40P01`); שגיאות כמו הפרת ייחודיות דורשות זהירות, כי
  הן עלולות להיות קבועות ולא זמניות.

מזה נגזרת הצורה של הפונקציה: היא לא מקבלת שאילתה, אלא פונקציה שמריצה את כל העבודה בתוך טרנזקציה, ומריצה
אותה מההתחלה בכל ניסיון. זה הקובץ `exercises/ch11-isolation-lab/solution/db/retry.go`, בחלק המרכזי:

<div dir="ltr" align="left">

```go
func RetrySerializable[T any](ctx context.Context, db TxBeginner, fn func(ctx context.Context, tx pgx.Tx) (T, error)) (T, error) {
	var zero T
	for attempt := 1; ; attempt++ {
		var out T
		err := pgx.BeginTxFunc(ctx, db, pgx.TxOptions{IsoLevel: pgx.Serializable}, func(tx pgx.Tx) error {
			var err error
			out, err = fn(ctx, tx)
			return err
		})
		if err == nil {
			return out, nil
		}
		if !IsRetryable(err) {
			return zero, err
		}
		if attempt == MaxAttempts {
			return zero, fmt.Errorf("serializable transaction failed after %d attempts: %w", attempt, err)
		}
		if err := sleep(ctx, backoff(attempt)); err != nil {
			return zero, err
		}
	}
}
```

</div>

ההחלטות בקוד הזה, אחת אחת:

| החלטה | למה |
| --- | --- |
| הפרמטר הגנרי `T` | הפונקציה מחזירה ערך מתוך הטרנזקציה (מזהה שנוצר, יתרה), בלי משתנים חיצוניים |
| ‏`pgx.BeginTxFunc` | פותח ב־Serializable, מאשר בהצלחה, מבטל בשגיאה; גם כשל ב־`COMMIT` עצמו חוזר כשגיאה |
| ‏`IsRetryable` | רק שני הקודים, כשל סריאליזציה וקיפאון, שמזוהים עם `errors.As` מול `*pgconn.PgError`; כל שגיאה אחרת חוזרת מיד |
| ‏`MaxAttempts` | חמישה ניסיונות; אחריהם השגיאה עטופה עם `%w`, כך שהקוראת עדיין יכולה לזהות `40001` |
| ‏`backoff` | המתנה גדלה מעריכית עם פיזור אקראי (jitter) מלא, כדי ששתי טרנזקציות מתחרות לא יתנגשו שוב באותו רגע |
| ‏`sleep(ctx, ...)` | ההמתנה נקטעת אם ה־context בוטל; בקשה שהלקוח עזב לא ממשיכה לנסות |

והחוזה שהקוראת חייבת לכבד, כתוב בהערה מעל הפונקציה: `fn` עלולה לרוץ כמה פעמים, ולכן אסור לה לעשות שום דבר
מחוץ לטרנזקציה. לא קריאת HTTP, לא שליחת הודעה, לא הוספה ל־slice משותף. כל מה שהיא עושה במסד מתבטל בניסיון
שנכשל; כל מה שהיא עושה בחוץ יקרה שוב.

> **איך Rust עושה את זה:** ב־Rust היית כותבת את הפרמטר עם החסם `FnMut`, והחתימה עצמה הייתה מבהירה
> שהסגירה נקראת כמה פעמים. ב־Go אין הבחנה כזאת בין סוגי סגירות, ולכן החוזה „מותר לקרוא לי שוב” חי בהערה
> ובבדיקות, לא בטיפוס.

### 11.8.1 לראות ניסיון חוזר קורה

בדיקה שרק מריצה שתי טרנזקציות במקביל ובודקת את התוצאה לא מוכיחה שהיה ניסיון חוזר: אולי פשוט רצו אחת אחרי
השנייה. הבדיקה הזאת מכריחה התנגשות בניסיון הראשון ומונה ניסיונות. מהקובץ
`exercises/ch11-isolation-lab/tests/oncall/oncall_test.go`:

<div dir="ltr" align="left">

```go
	attempts := 0
	_, err := db.RetrySerializable(context.Background(), pool, func(ctx context.Context, tx pgx.Tx) (struct{}, error) {
		attempts++
		if _, err := oncall.CountOnCall(ctx, tx, 1); err != nil { // our read
			return struct{}{}, err
		}
		if attempts == 1 { // the interfering transaction commits in between
			if err := oncall.GoOffCallSerializable(ctx, pool, 1, "alice"); err != nil {
				return struct{}{}, err
			}
		}
		return struct{}{}, oncall.GoOffCall(ctx, tx, 1, "bob") // our write
	})
```

</div>

בניסיון הראשון, אחרי שבוב קרא, טרנזקציה אחרת מוציאה את אליס ומאשרת. כשבוב כותב, PostgreSQL מזהה את המעגל
ומבטל אותו. בניסיון השני הטרנזקציה של בוב מתחילה מחדש, הספירה כבר רואה תורן אחד, והתשובה היא
‏`ErrLastOnCall`. הבדיקה דורשת שני ניסיונות בדיוק ותורן אחד בסוף. שימי לב שהבדיקה עצמה עושה בתוך `fn` משהו
מחוץ לטרנזקציה, בדיוק מה שאסור בקוד אמיתי, וזו הדרך היחידה להכריח התנגשות בנקודה מדויקת.

### 11.8.2 מתי Serializable ומתי `FOR UPDATE`

| מצב | בחירה טבעית |
| --- | --- |
| שורה אחת, קריאה ואז כתיבה | ‏`UPDATE` אטומי, או `FOR UPDATE` |
| כלל שנשען על כמה שורות או על היעדר שורה (write skew) | Serializable עם ניסיון חוזר, או נעילה של שורת „אב” משותפת |
| הרבה טרנזקציות שונות שקשה לנתח את התלויות ביניהן | Serializable לכולן, עם פונקציה אחת לניסיון חוזר |
| טרנזקציות ארוכות עם הרבה קריאות | זהירות: ב־Serializable הן צוברות תלויות ונכשלות יותר |

הבחירה ב־Serializable היא בחירה ברמת המערכת: הבטחת הבידוד חלה רק בין טרנזקציות שרצות ב־Serializable. לכן
בפרויקט הגמר כל טרנזקציית כתיבה שנשענת על כלל עוברת דרך אותה פונקציה.

## סיכום

- ‏ACID ב־PostgreSQL: אטומיות לכל טרנזקציה, עקביות רק לכללים שהוגדרו כאילוצים, עמידות דרך WAL, ובידוד שתלוי
  ברמה שבחרת.
- ‏MVCC: קריאה לא חוסמת כתיבה; כל פקודה או טרנזקציה רואה תמונת מצב, ושתי טרנזקציות יכולות להחליט על אותו ערך ישן.
- רמת Read committed מתירה עדכון אבוד. רמת Repeatable read מזהה אותו בקוד `40001`, אבל מתירה write
  skew. רמת Serializable עוצרת את שניהם, בתנאי שמנסים שוב.
- נעילת שורה או עדכון אטומי מתקנים עדכון אבוד. התוספת `SKIP LOCKED` הופכת טבלה לתור שכמה workers חולקים.
- קיפאון מתגלה אחרי `deadlock_timeout` ומבטל טרנזקציה אחת (`40P01`); נעילה בסדר קבוע מונעת אותו.
- מיגרציות goose הן קובצי SQL עם חלק עלייה וחלק ירידה. הן מוטמעות בבינארי ורצות מהקוד תחת advisory lock, ולא
  עורכים מיגרציה שהוחלה.
- גודל pool מפורש (`MaxConns`), כפול מספר העותקים, נשאר מתחת ל־`max_connections`.
- ‏PgBouncer במצב טרנזקציה שובר כל מה שתלוי בסשן.
- הפונקציה `RetrySerializable` מריצה את כל הטרנזקציה מחדש על כשל סריאליזציה או קיפאון, עם המתנה גדלה
  ופיזור אקראי. הפונקציה שהיא מריצה לא עושה דבר מחוץ לטרנזקציה.

## מונחים לזכור

| מונח בעברית | באנגלית | במשפט אחד |
| --- | --- | --- |
| טרנזקציה | transaction | קבוצת פקודות שמאושרת או מתבטלת כיחידה אחת |
| רמת בידוד | isolation level | אילו השפעות של טרנזקציות מקבילות טרנזקציה רשאית לראות |
| תמונת מצב | snapshot | רשימת הטרנזקציות שאושרו מבחינת פקודה או טרנזקציה, שקובעת איזו גרסת שורה היא רואה |
| בקרת מקביליות מרובת גרסאות | MVCC | עדכון כותב גרסה חדשה של השורה, כך שקוראים וכותבים לא חוסמים זה את זה |
| עדכון אבוד | lost update | שתי טרנזקציות קוראות ערך, שתיהן כותבות, ואחת הכתיבות נדרסת |
| הטיית כתיבה, write skew | write skew | שתי טרנזקציות בודקות כלל על אותן שורות וכותבות לשורות שונות, והכלל נשבר |
| כשל סריאליזציה | serialization failure | ביטול טרנזקציה בקוד `40001` כדי לשמור על בידוד; מנסים שוב |
| נעילת שורה | row lock | נעילה ש־`FOR UPDATE` לוקח, עד סוף הטרנזקציה |
| קיפאון | deadlock | מעגל של טרנזקציות שכל אחת מחכה לנעילה של אחרת; קוד `40P01` |
| מיגרציה | migration | שינוי סכמה כקובץ ממוספר, שרץ פעם אחת ונרשם במסד |
| advisory lock | advisory lock | נעילה על מספר שהיישום בוחר, בלי קשר לשורה; goose משתמש בה |
| מאגר חיבורים | connection pool | קבוצה מוגבלת של חיבורים פתוחים שבקשות שואלות ומחזירות |
| ניסיון חוזר | retry | הרצה מחדש של כל הטרנזקציה אחרי כשל זמני |

## טעויות נפוצות

- **לחשוב ש„יש טרנזקציה” פירושו „אין מרוצים”.** ב־Read committed, קריאה ואז כתיבה בתוך טרנזקציה עדיין מאבדת
  עדכונים. צריך נעילה, עדכון אטומי, או רמה גבוהה יותר.
- **לעבור ל־Repeatable read ולא לטפל ב־`40001`.** הרמה מזהה את הבעיה ומחזירה שגיאה; בלי ניסיון חוזר, המשתמשת
  מקבלת 500.
- **לנסות שוב רק את הפקודה שנכשלה.** הטרנזקציה כבר מבוטלת, וההחלטות שלה נשענו על קריאות ישנות. מריצים את
  כולה מההתחלה.
- **קריאת HTTP בתוך פונקציה שמנוסה שוב.** כל ניסיון שולח אותה שוב. השפעות חיצוניות יוצאות מהטרנזקציה, או עוברות
  לתבנית ה־outbox של פרק 13.
- **להחזיק נעילה בזמן עבודה איטית.** `SKIP LOCKED` עם מטפל של עשר שניות מחזיק חיבור ונעילה כל הזמן הזה.
  עבודה ארוכה עוברת לחכירה.
- **לנעול שורות בסדר שתלוי בקלט.** העברות הפוכות ייתקעו זו בזו. ממיינים מזהים לפני שנועלים.
- **לערוך קובץ מיגרציה שכבר רץ.** goose לא יריץ אותו שוב, וסביבות שונות יחזיקו סכמות שונות. כותבים מיגרציה חדשה.
- **‏`MaxConns` גבוה „ליתר ביטחון”.** כפול מספר העותקים זה חוצה את `max_connections`, והשירות נופל דווקא בפריסה.

## שאלות ראיון

1. **מה ההבדל בין רמות הבידוד ב־PostgreSQL, ובמה תבחרי?** ‏Read committed נותנת תמונת מצב לכל פקודה, ולכן
   קריאה ואז כתיבה יכולות לאבד עדכון. Repeatable read נותנת תמונה אחת לכל הטרנזקציה ומחזירה `40001` כשמנסים
   לעדכן שורה שהשתנתה, אבל לא מגינה מ־write skew. Serializable מבטיחה תוצאה של הרצה סדרתית כלשהי, במחיר של
   ביטולים שצריך לנסות שוב ושל מעקב אחרי קריאות. ברירת המחדל הסבירה היא Read committed עם נעילות מפורשות
   במקומות הקריטיים; Serializable כשהכללים נשענים על כמה שורות וקשה לנתח את כל המסלולים, עם פונקציית ניסיון
   חוזר אחת.
   **המשך:** ומה עם Read uncommitted? ב־PostgreSQL היא מתנהגת כמו Read committed; אין קריאות מלוכלכות בכלל.

2. **איך תמנעי עדכון אבוד במשיכה מחשבון?** הכי פשוט: פקודת עדכון אחת שמפחיתה את הסכום רק אם התנאי
   ‏`balance >= $2` מתקיים, ובדיקת מספר השורות שהשתנו. ב־Read committed העדכון השני מחכה ומעריך את התנאי מחדש על
   הגרסה החדשה. כשההחלטה מורכבת יותר מביטוי אחד, `SELECT ... FOR UPDATE` ואז חישוב וכתיבה. אפשרות שלישית היא
   נעילה אופטימית: עמודת גרסה ו־`WHERE version = $n`, שמתאימה כשהקריאה והכתיבה מופרדות בזמן (למשל טופס
   שנפתח ונשמר אחרי דקה). כל אחת מתאימה למשך ולמורכבות אחרים של ההחלטה.

3. **מה זה write skew ואיך מתקנים?** שתי טרנזקציות בודקות כלל על אותה קבוצת שורות, כל אחת רואה שהכלל מתקיים,
   וכל אחת כותבת לשורה אחרת, כך שאחרי שתיהן הכלל נשבר. Repeatable read לא מזהה את זה כי אף שורה לא נכתבה
   פעמיים. תיקונים: Serializable עם ניסיון חוזר; נעילה של שורה משותפת שמייצגת את הקבוצה (למשל `FOR UPDATE`
   על שורת המשמרת); או אילוץ במסד כשאפשר לבטא את הכלל כך. הבחירה תלויה בכמה מקומות בקוד נוגעים בכלל.

4. **איך בונים תור עבודות על PostgreSQL, ומתי לא?** בטרנזקציה אחת: בחירה של שורה אחת עם `FOR UPDATE SKIP LOCKED`,
   עיבוד וסימון, כך שכל worker לוקח שורה שאף אחד אחר לא מחזיק. זה פשוט, טרנזקציוני עם שאר הנתונים, וקל לראות בו כל
   עבודה. לעבודות ארוכות עוברים לחכירה עם זמן תפוגה, כדי לא להחזיק נעילה. לא מתאים כשהנפח גדל עד שהטבלה
   והניקוי שלה (`VACUUM`) נעשים צוואר בקבוק, או כשצריך שידור לצרכנים רבים; שם יומן כמו Kafka (פרק 13) מתאים יותר.

5. **שירות עם עשרה עותקים מקבל `too many connections`. מה קרה ומה עושים?** כנראה כל עותק פותח pool בגודל ברירת
   המחדל או גדול מדי, והמכפלה, אולי בזמן פריסה כשיש עותקים ישנים וחדשים, חוצה את `max_connections`. קודם מחשבים:
   ‏`MaxConns` לעותק כפול עותקים בשיא, ועוד שוליים לניהול. אחר כך בודקים אם יש חיבורים שמוחזקים זמן רב
   (טרנזקציות ארוכות, `rows` לא סגורים). אם באמת צריך יותר לקוחות מחיבורים, מציבים PgBouncer במצב טרנזקציה,
   ובודקים שהקוד לא תלוי ב־`SET` של סשן, בנעילות מייעצות של סשן או ב־prepared statements בלי תמיכה.
   **המשך:** למה לא פשוט להגדיל את `max_connections`? כל חיבור הוא תהליך עם זיכרון, ומעבר למספר הליבות
   חיבורים פעילים נוספים רק מחכים בתוך השרת; זה מזיז את התור, לא מקצר אותו.

## פתק Obsidian

הפתק המוכן נמצא גם כקובץ: `obsidian/11-postgres-transactions-migrations.md`. זה תוכנו:

- **‏Read committed:** תמונת מצב לכל פקודה; קריאה ואז כתיבה מאבדת עדכונים.
- **‏Repeatable read:** תמונה אחת לטרנזקציה; עדכון של שורה שהשתנתה נכשל ב־`40001`; write skew עובר.
- **‏Serializable:** עוצר גם write skew; חובה לנסות שוב את כל הטרנזקציה.
- **עדכון אבוד:** ‏`UPDATE` אטומי עם תנאי, או `SELECT ... FOR UPDATE`.
- **תור:** ‏`FOR UPDATE SKIP LOCKED LIMIT 1`; הנעילה מוחזקת כל זמן העיבוד.
- **קיפאון:** ‏`40P01` אחרי `deadlock_timeout`; נועלים לפי סדר מזהים קבוע.
- **goose:** חלק עלייה וחלק ירידה בכל קובץ SQL; לא עורכים מיגרציה שהוחלה.
- **goose מתוך Go:** המיגרציות מוטמעות עם `embed` ורצות תחת advisory lock של סשן.
- **pool:** גודל ה־pool כפול מספר העותקים נשאר מתחת ל־`max_connections`.
- **‏PgBouncer:** במצב טרנזקציה הוא שובר כל מצב של סשן.
- **ניסיון חוזר:** הפונקציה `fn` בלי השפעות חיצוניות; רק כשל סריאליזציה וקיפאון; חמישה ניסיונות עם פיזור אקראי.
- **שאלה פתוחה:** ___

<div dir="ltr" align="left">

```text
up        docker compose -p ch11 -f data/compose.yaml up -d --wait
url       DATABASE_URL=postgres://course:change-me@localhost:5411/lab
tests     go test -race ./...        go test -race -run 'LostUpdate|WriteSkew' ./bank ./oncall
tx        pgx.BeginTxFunc(ctx, pool, pgx.TxOptions{IsoLevel: pgx.Serializable}, fn)
sqlstate  40001 serialization_failure | 40P01 deadlock_detected | 23514 check_violation
lock      SELECT ... FOR UPDATE        SELECT ... FOR UPDATE SKIP LOCKED LIMIT 1
atomic    UPDATE accounts SET balance = balance - $2 WHERE id = $1 AND balance >= $2
waiting   SELECT pid, wait_event_type, query FROM pg_stat_activity WHERE wait_event_type = 'Lock';
deadlock  SHOW deadlock_timeout;   docker compose -p ch11 -f data/compose.yaml logs db | grep -A6 "deadlock detected"
goose     GOOSE_DRIVER=postgres GOOSE_DBSTRING=$DATABASE_URL GOOSE_MIGRATION_DIR=migrations goose status|up|down|down-to 0
create    goose -s create <name> sql
pool      SHOW max_connections;   MaxConns x replicas < max_connections - reserve = ___
retry     db.RetrySerializable(ctx, pool, func(ctx context.Context, tx pgx.Tx) (T, error) { ... })
down      docker compose -p ch11 -f data/compose.yaml down -v
```

</div>

## תרגיל

**מטרה:** לכתוב בדיקות Go שמייצרות עדכון אבוד ו־write skew על שני חיבורים, לתקן אותם עם `FOR UPDATE` ועם
Serializable וניסיון חוזר, לבנות תור עבודות עם `SKIP LOCKED`, לגרום לקיפאון ולמנוע אותו, ולהוסיף מיגרציה עם goose.
**זמן:** שלושה ימי לימוד, ימים 26 עד 28, כ־60 דקות בנייה ביום הראשון, כ־45 בשני וכ־65 בשלישי.
**ידע מקדים:** הפרק, ופרק 10 (המסד בקונטיינר ו־pgx). את goose מתקינים בצעד הראשון.

הצעדים, קוד ההתחלה, הרמזים והפתרון נמצאים ב־
[`exercises/ch11-isolation-lab/README.md`](../exercises/ch11-isolation-lab/README.md).

**קריטריון סיום:**

- בדיקות האנומליות מייצרות עדכון אבוד ו־write skew בכל הרצה, ומאשרות אותם.
- בדיקות התיקון עוברות: `FOR UPDATE` גורם להמתנה מוכחת, ו־Serializable עם ניסיון חוזר משאיר תורן אחד.
- שמונה workers מרוקנים 300 עבודות בלי שאף עבודה טופלה פעמיים.
- דוח הקיפאון מהלוג של השרת שמור בפתק, וההעברות המסודרות לא נתקעות.
- המיגרציה החדשה עולה ויורדת עם goose, ובדיקת הסבב עוברת על מסד ריק.
- כל הבדיקות עוברות עם `-race` מול הקונטיינר.

## העשרה אופציונלית

- בתיעוד של PostgreSQL 18, פרק 13 („Concurrency Control”) כולו, ובמיוחד 13.2 על רמות הבידוד ו־13.6 על טיפול
  בכשלי סריאליזציה.
- המאמר „A Critique of ANSI SQL Isolation Levels” (Berenson ואחרים, 1995) הוא המקור של write skew ושל ההבחנה
  שהטבלה של התקן לא מספיקה.
- הפרויקט Jepsen בדק רמות בידוד של מסדים רבים, כולל PostgreSQL; הניתוחים שלו מראים איך מוצאים אנומליות בבדיקות.
- נסי להוסיף `NOWAIT` במקום `SKIP LOCKED` בתור, ולראות איזו שגיאה מקבל ה־worker השני.

</div>
