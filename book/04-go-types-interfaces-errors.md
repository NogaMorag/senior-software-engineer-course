---
direction: rtl
lang: he
title: "פרק 4: טיפוסים, ממשקים ושגיאות"
chapter: 4
part: "חלק א: Go"
---

<div dir="rtl" lang="he">

# פרק 4: טיפוסים, ממשקים ושגיאות

בפרק 3 כתבת תוכנית. בפרק הזה תכתבי ספרייה: קוד שאחרים קוראים לו, שמחזיר שגיאות שאפשר לפעול לפיהן,
ושאפשר להחליף את האחסון שלו בלי לגעת בכללים. אלה הנושאים שסוקרת קוד בכירה בודקת קודם: האם הממשק קטן,
האם השגיאה נושאת הקשר בלי לאבד את הסיבה, והאם יש panic במקום שאסור. הכלי שתבני, `ledger`, הוא ספר
חשבונות קטן עם שני מימושי אחסון וחבילת בדיקות אחת שרצה על שניהם.

## מה נלמד בפרק

- הגדרת structs ומתודות, ומתי receiver הוא מצביע.
- ממשקים מרומזים מול traits, ומלכודת ה־`nil` בתוך ממשק.
- ממשקים קטנים: `io.Reader` ו־`io.Writer`.
- שגיאות כערכים: sentinel מול טיפוס שגיאה, עטיפה ושליפה דרך שרשרת העטיפות.
- panic ו־recover, ולמה כמעט אף פעם לא.
- שימוש ב־generics עם אילוצים, כולל מתודות גנריות.
- הטמעה (embedding): מה היא מקדמת, ולמה היא לא ירושה.

## ידע מקדים

- פרק 3: מודולים, slices ובאג הכינוי, maps, בדיקות טבלאיות, ושלושת כלי הבדיקה.
- היכרות עם traits ועם `Result` ב־Rust.

## 4.1 structs, מתודות ו־receivers

מתודה ב־Go היא פונקציה עם receiver: הפרמטר שלפני השם. אפשר להגדיר מתודות על כל טיפוס בעל שם בחבילה
שלך, לא רק על struct. כך הספר מייצג כסף, בקובץ `cents.go` של התרגיל:

<div dir="ltr" align="left">

```go
// Cents is an amount of money in hundredths. Money is never a float.
type Cents int64

// String implements fmt.Stringer: Cents(1250) prints as "12.50".
func (c Cents) String() string {
	sign := ""
	if c < 0 {
		sign, c = "-", -c
	}
	return fmt.Sprintf("%s%d.%02d", sign, c/100, c%100)
}
```

</div>

הטיפוס `Cents` הוא טיפוס חדש שבסיסו `int64`. זו המקבילה של newtype: אי אפשר לערבב אותו בטעות עם
מספר רגיל בלי המרה מפורשת. המתודה `String` הופכת אותו ל־`fmt.Stringer`, ולכן הוא מודפס עם נקודה
עשרונית, 12.50 ולא 1250. אין הצהרת מימוש; מספיקה מתודה בשם הנכון.

ה־receiver יכול להיות ערך או מצביע:

| receiver | משמעות | מתי |
| --- | --- | --- |
| ערך, `(c Cents)` | המתודה מקבלת עותק | טיפוס קטן שהמתודה לא משנה |
| מצביע, `(s *MemStore)` | המתודה רואה את המקור ויכולה לשנות אותו | כל מתודה שמשנה מצב, או struct גדול |

המוסכמה: אם מתודה אחת של טיפוס צריכה מצביע, כולן מקבלות מצביע. המתודות של `MemStore` משנות maps
ומונה, ולכן כולן עם `*MemStore`.

לקבועים ממוספרים אין `enum`. משתמשים ב־`iota`, מונה שגדל באחד בכל שורה של בלוק קבועים. כך בקובץ
`store.go`:

<div dir="ltr" align="left">

```go
const (
	Deposit Kind = iota + 1 // start at 1 so the zero Kind is visibly invalid
	Withdrawal
)
```

</div>

ה־`+ 1` אינו קישוט. ערך האפס של `Kind` הוא 0, ו־struct שלא אותחל במלואו יקבל אותו. אם 0 היה `Deposit`,
טרנזקציה שנשכח בה השדה הייתה הפקדה תקינה. כך האפס פסול ונראה לעין, והמתודה `String` מדפיסה אותו
כ־`Kind(0)`.

## 4.2 ממשקים מול traits

ממשק (interface) הוא רשימת מתודות. כל טיפוס שיש לו את המתודות האלה מממש את הממשק, בלי להצהיר על כך.
זה הממשק המרכזי של התרגיל, בקובץ `store.go`:

<div dir="ltr" align="left">

```go
// Store is everything the ledger needs from storage. Any type with these
// four methods is a Store; it never has to say so.
type Store interface {
	CreateAccount(a Account) error
	Account(id AccountID) (Account, error)
	// Record applies tx to its account, stores it with a fresh ID, and returns
	// the stored transaction and the updated account.
	Record(tx Transaction) (Transaction, Account, error)
	Transactions(id AccountID) ([]Transaction, error)
}
```

</div>

אין `impl Store for MemStore` בשום מקום. כדי שבכל זאת תקבלי שגיאת קומפילציה אם מתודה חסרה, מקובל
להוסיף שורת בדיקה כזו, מהקובץ `memstore.go`:

<div dir="ltr" align="left">

```go
// Compile-time check that *MemStore implements Store.
var _ Store = (*MemStore)(nil)
```

</div>

שימי לב שהשורה כתובה על `*MemStore` ולא על `MemStore`. המתודות מוגדרות על מצביע, ולכן רק המצביע
מממש את הממשק. זו קבוצת המתודות (method set) של הטיפוס, והמהדר אוכף אותה.

> **איך Rust עושה את זה:** מימוש של trait הוא הצהרה מפורשת, ופונקציה גנרית עם trait bound מקבלת
> קריאה סטטית בלי עלות (monomorphization). ב־Go המימוש מרומז, ולכן הצרכנית יכולה להגדיר ממשק אחרי שהטיפוס כבר קיים, אפילו
> לטיפוס מספרייה אחרת. קריאה דרך ממשק היא תמיד דינמית, כמו `dyn Trait`: ערך הממשק הוא זוג של טיפוס
> ומצביע לנתונים.

מכאן נובעת מוסכמה: **הממשק מוגדר אצל מי שמשתמשת בו**, לא אצל מי שמממשת. החבילה `ledger` מגדירה את
`Store` כי היא צריכה אותו. ומי שמממשת מחזירה טיפוס קונקרטי, כמו `*FileStore`, ולא ממשק.

### 4.2.1 ממשק שהוא `nil`, וממשק שמכיל `nil`

כי ערך ממשק הוא זוג של טיפוס וערך, הוא `nil` רק כששניהם ריקים. מצביע `nil` שהוכנס לממשק יוצר ממשק
שאינו `nil`. מהקובץ `example_test.go`:

<div dir="ltr" align="left">

```go
// lookup returns a nil *InsufficientFundsError inside a non-nil error interface.
// This is the bug: the interface holds a type, so it is not nil.
func lookup() error {
	var e *InsufficientFundsError // nil pointer
	return e                      // wrapped in an interface: (type=*InsufficientFundsError, value=nil)
}

func Example_nilInterfaceTrap() {
	err := lookup()
	fmt.Println(err == nil)
	fmt.Printf("%T\n", err)
	// Output:
	// false
	// *ledger.InsufficientFundsError
}
```

</div>

זה אחד הבאגים הקלאסיים ב־Go: פונקציה מחזירה משתנה שגיאה מטיפוס קונקרטי שנשאר `nil`, והקוראת רואה
שגיאה. הכלל: פונקציה שמחזירה `error` מחזירה `nil` מפורש בהצלחה, ולעולם לא משתנה מטיפוס קונקרטי.
ב־Rust `Option` מבדיל בין „אין” ל„יש, וזה ריק” בטיפוס עצמו; ב־Go ההבחנה חבויה בתוך הממשק.

## 4.3 ממשקים קטנים: `io.Reader` ו־`io.Writer`

הממשקים החשובים בספרייה הסטנדרטית קטנים מאוד. הפקודה `go doc` מציגה אותם:

<div dir="ltr" align="left">

```bash
go doc io.Reader | head -5
```

</div>

<div dir="ltr" align="left">

```text
package io // import "io"

type Reader interface {
	Read(p []byte) (n int, err error)
}
```

</div>

מתודה אחת. קובץ, חיבור רשת, גוף בקשת HTTP, מחרוזת בזיכרון ודחיסה, כולם מממשים אותה, ולכן כל קוד
שמקבל `io.Reader` עובד עם כולם. כך `Collect` מפרק 3 קוראת גם קובץ וגם מחרוזת בבדיקה. ו־`io.Writer`
הוא התמונה ההפוכה: מתודת `Write` אחת.

התוכנית `ledger` כותבת את כל הפלט שלה ל־`io.Writer` שהיא מקבלת, ולא ישירות לטרמינל. מהקובץ
`cmd/ledger/main.go`:

<div dir="ltr" align="left">

```go
func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr, time.Now))
}
```

</div>

בייצור אלה הזרמים האמיתיים; בבדיקה הם `bytes.Buffer`, ובודקים מה נכתב. גם השעון מוזרק, כדי שהבדיקות
יראו תמיד אותו זמן. ככל שהממשק קטן יותר, קל יותר לממש, לזייף ולהרכיב אותו.

## 4.4 שגיאות כערכים מול `Result`

ב־Go שגיאה היא ערך רגיל מטיפוס `error`, ממשק עם מתודה אחת, `Error() string`. פונקציה שיכולה להיכשל
מחזירה אותה כערך האחרון, והקוראת בודקת: `if err != nil`. יש שתי צורות של שגיאות:

| צורה | מה היא | שליפה | בתרגיל |
| --- | --- | --- | --- |
| sentinel | ערך קבוע של חבילה | `errors.Is` | `ErrAccountNotFound` |
| טיפוס שגיאה | struct שמממש `error` ונושא נתונים | `errors.As` או `errors.AsType` | `InsufficientFundsError` |

כך הן מוגדרות, בקובץ `errors.go`:

<div dir="ltr" align="left">

```go
// Sentinel errors: fixed values that callers compare with errors.Is.
var (
	ErrAccountNotFound = errors.New("account not found")
	ErrAccountExists   = errors.New("account already exists")
	ErrInvalidAmount   = errors.New("amount must be positive")
)

// InsufficientFundsError carries data the caller may want, so it is a type,
// not a sentinel. Callers extract it with errors.As or errors.AsType.
type InsufficientFundsError struct {
	Account   AccountID
	Balance   Cents
	Requested Cents
}
```

</div>

בוחרים sentinel כשהקוראת צריכה רק לדעת מה קרה, וטיפוס כשהיא צריכה נתונים: כאן, כמה כסף חסר. הכלל
העסקי עצמו נמצא בקובץ `store.go`, ומחזיר את טיפוס השגיאה כערך:

<div dir="ltr" align="left">

```go
	case Withdrawal:
		if tx.Amount > a.Balance {
			return a, InsufficientFundsError{Account: a.ID, Balance: a.Balance, Requested: tx.Amount}
		}
		a.Balance -= tx.Amount
```

</div>

> **איך Rust עושה את זה:** הטיפוס `Result` מכריח אותך לטפל בשגיאה כדי להגיע לערך, והאופרטור `?`
> מעביר אותה הלאה בתו אחד. ב־Go אין כפייה כזו: מותר להתעלם מ־`err`, ורק כלי ניתוח
> סטטי יעירו. אין גם `?`; כל העברה היא `if err != nil { return ..., err }` מפורש. המחיר הוא
> טקסט; התמורה היא שכל נקודת כשל נראית, וזה בדיוק המקום להוסיף הקשר.

## 4.5 עטיפה: `%w`, `errors.Is` ו־`errors.As`

שגיאה שעוברת שכבות צריכה לצבור הקשר: „לא נמצא” לבדו לא אומר איזה חשבון ובאיזו פעולה. מציין הפורמט `%w`
של `fmt.Errorf` יוצר שגיאה חדשה עם הודעה מורחבת, שעוטפת את המקורית ושומרת אותה. מהקובץ `ledger.go`:

<div dir="ltr" align="left">

```go
	_, a, err := l.store.Record(tx)
	if err != nil {
		// Add context, keep the cause: callers can still use errors.Is and errors.As.
		return Account{}, fmt.Errorf("%v of %s on %s: %w", kind, amount, id, err)
	}
```

</div>

ההודעה שמתקבלת היא `withdrawal of 5.00 on ghost: account not found`, והשגיאה המקורית עדיין בפנים. עכשיו
השוואה ישירה, `err == ErrAccountNotFound`, נכשלת, כי `err` הוא העוטף. לכן משווים עם `errors.Is`, שעוברת
על שרשרת העטיפות. ושולפים טיפוס עם `errors.As`, או עם הגרסה הגנרית `errors.AsType` (קיימת מאז Go
1.26). מהקובץ `errors_test.go`:

<div dir="ltr" align="left">

```go
	// The classic form: a target variable and a pointer to it.
	var target InsufficientFundsError
	if !errors.As(err, &target) {
		t.Fatalf("errors.As found no InsufficientFundsError in %v", err)
	}
	// The generic form (Go 1.26 and later): no target variable.
	got, ok := errors.AsType[InsufficientFundsError](err)
```

</div>

שלושה כללים מעשיים לעטיפה:

- **עטיפה עם `%w` כשהקוראת אמורה לפעול לפי הסיבה.** עם `%v` כשהסיבה היא פרט מימוש שאסור להפוך
  לחלק מהחוזה. ברגע שעטפת עם `%w`, מי שבודקת `errors.Is` תלויה בשגיאה הפנימית שלך.
- **הקשר אחד לכל שכבה, בלי חזרות.** השכבה שקוראת מוסיפה את מה שהיא יודעת, לא את מה שהשגיאה כבר אומרת.
- **אותה צורה לאורך השרשרת.** אם השגיאה נוצרה כערך, שולפים ערך; אם כמצביע, שולפים מצביע.

הכלל האחרון מגיע עם עזרה חדשה: החל מ־Go 1.27 הכלי `go vet` מזהיר כשעוטפים מצביע לטיפוס שהערך שלו
כבר מממש את ממשק השגיאה. המתודה `Error` של `InsufficientFundsError` מוגדרת על ערך, ולכן גם הערך וגם
המצביע הם שגיאות. עטיפה של מצביע הייתה גורמת לשליפה לפי ערך להיכשל בשקט. וכי `go test` מריץ את
הבדיקה הזו בעצמו, טעות כזו עוצרת כבר את הבדיקות.

השגיאות הופכות להחלטה במקום אחד בלבד, בגבול של התוכנית. מהקובץ `cmd/ledger/main.go`:

<div dir="ltr" align="left">

```go
// report is the one place that turns errors into messages and exit codes.
func report(err error, stderr io.Writer) int {
	if err == nil {
		return exitOK
	}
	if insufficient, ok := errors.AsType[ledger.InsufficientFundsError](err); ok {
		fmt.Fprintf(stderr, "ledger: %v (short by %s)\n", err, insufficient.Shortfall())
		return exitInsufficient
	}
	fmt.Fprintln(stderr, "ledger:", err)
	switch {
	case errors.Is(err, errUsage):
		fmt.Fprintln(stderr, usage)
		return exitUsage
	case errors.Is(err, ledger.ErrAccountNotFound):
		return exitNotFound
	default:
		return exitError
	}
}
```

</div>

בשכבות הפנימיות לא מדפיסים ולא רושמים ללוג; רק מוסיפים הקשר ומחזירים. שגיאה שגם נרשמת וגם מוחזרת
מופיעה בלוג פעמיים או שלוש, וזו טעות שסוקרות מחפשות.

## 4.6 panic ו־recover, ולמה כמעט אף פעם

ה־panic עוצר את הפונקציה, מריץ את ה־`defer`s בדרך למעלה, ואם איש לא עוצר אותו, מפיל את התהליך עם
מחסנית קריאות. בשימוש הוא דומה ל־`panic!`: לבאג של המתכנתת, לא לתנאי שקורה בזמן ריצה. קלט שגוי מהמשתמשת,
קובץ חסר או רשת שנפלה הם שגיאות, לא panics.

המקום הלגיטימי הנפוץ הוא פונקציית `Must`, לקלט שכתוב בקוד עצמו, מהקובץ `cents.go`:

<div dir="ltr" align="left">

```go
// MustParseCents is for constants in code and tests. It panics on bad input,
// because bad input here is a programmer error, not a runtime condition.
func MustParseCents(s string) Cents {
	c, err := ParseCents(s)
	if err != nil {
		panic(err)
	}
	return c
}
```

</div>

הפונקציה `recover` עוצרת panic, אבל רק כשהיא נקראת ישירות מתוך `defer`. הבדיקה
`TestMustParseCentsPanicsOnBadInput` בקובץ `cents_test.go` משתמשת בה כדי לוודא שה־panic אכן קורה.

למה `recover` שייכת רק לגבול של goroutine או של בקשה:

- **אין תפיסה בין goroutines.** panic שלא נתפס באותה goroutine מפיל את כל התהליך, ותפיסה במקום אחר
  לא תעזור.
- **בגבול יש למי לדווח.** שרת HTTP ממיר panic בבקשה אחת לתשובת 500 ולשורת לוג, והשרת ממשיך לשרת
  אחרות. בפרק 6 תכתבי middleware כזה.
- **באמצע הקוד אין מה לעשות איתו.** recover שם מסתיר באג ומשאיר מצב חצי מעודכן.

## 4.7 generics

ה־generics הגיעו ל־Go מאוחר, ובכוונה מצומצמים. פרמטרי טיפוס נכתבים בסוגריים מרובעים, ולכל אחד
אילוץ (constraint). בקובץ `page.go` מוגדר הטיפוס הגנרי `Page[T]`: רשימת פריטים, סמן לדף הבא ודגל
שאומר אם יש עוד. וזו הפונקציה שמייצרת דף:

<div dir="ltr" align="left">

```go
func Paginate[T any](items []T, cursor, limit int) (Page[T], error) {
	if limit < 1 {
		return Page[T]{}, fmt.Errorf("limit %d: must be at least 1", limit)
	}
	if cursor < 0 || cursor > len(items) {
		return Page[T]{}, fmt.Errorf("cursor %d of %d items: %w", cursor, len(items), ErrBadCursor)
	}
	end := min(cursor+limit, len(items))
	return Page[T]{
		Items:   items[cursor:end:end], // cap == len: appending to Items cannot touch items
		Next:    end,
		HasMore: end < len(items),
	}, nil
}
```

</div>

האילוץ `any` אומר „כל טיפוס”. שימי לב לאינדקס השלישי ב־`items[cursor:end:end]`: זה ביטוי ה־slice המלא
מפרק 3, והוא מבטיח שמי שמוסיפה לדף לא דורסת את הרשימה המקורית. הבדיקה
`TestPageItemsDoNotAliasTheSource` שומרת עליו.

אילוץ יכול להיות גם קבוצת טיפוסים. מאותו קובץ:

<div dir="ltr" align="left">

```go
// Integer is a constraint: any type whose underlying type is one of these.
// The ~ is what lets Cents (underlying int64) satisfy it.
type Integer interface {
	~int | ~int32 | ~int64
}

// Sum adds up values of any Integer type and returns the same type.
func Sum[N Integer](values []N) N {
	var total N // the zero value of N
	for _, v := range values {
		total += v
	}
	return total
}
```

</div>

בלי הטילדה, `Sum` הייתה מקבלת רק את `int64` עצמו, ודוחה את טיפוס הכסף שלנו. הטילדה פירושה „כל טיפוס שזה
הבסיס שלו”, ובזכותה `Sum([]Cents{...})` מחזירה `Cents`, שמודפס יפה.

החל מ־Go 1.27 גם מתודה יכולה להצהיר על פרמטרי טיפוס משלה. כך `Page` מקבלת מתודה שממירה את הפריטים:

<div dir="ltr" align="left">

```go
// Map converts every item and keeps the paging fields. It is a generic method,
// with its own type parameter U; Go supports these since 1.27.
func (p Page[T]) Map[U any](f func(T) U) Page[U] {
	out := Page[U]{Items: make([]U, 0, len(p.Items)), Next: p.Next, HasMore: p.HasMore}
	for _, item := range p.Items {
		out.Items = append(out.Items, f(item))
	}
	return out
}
```

</div>

מגבלה אחת חשובה, לפי הערות השחרור: מתודה בתוך ממשק לא יכולה להיות גנרית, ומתודה גנרית לא מממשת מתודת
ממשק. בקוד ישן, או בספריות שתומכות בגרסאות קודמות, תמצאי במקומה פונקציה רגילה כמו `MapPage(p, f)`.

> **איך Rust עושה את זה:** אילוץ ב־Go הוא רשימת מתודות או קבוצת טיפוסים, ואילו ב־Rust יש trait
> bounds עשירים עם טיפוסים משויכים ומימושים גנריים. ב־Rust אפשר לכתוב `impl<T: Display> Trait for Vec<T>`;
> ב־Go לא. הכלל המעשי שהקהילה אימצה: generics למבני נתונים ולפונקציות עזר, כמו `Paginate`,
> וממשקים רגילים להתנהגות, כמו `Store`.

## 4.8 הטמעה (embedding) והרכבה

ב־Go אין ירושה. יש **הטמעה**: שדה בלי שם בתוך struct. השדות והמתודות של הטיפוס המוטמע „מקודמים”
(promoted), ואפשר לקרוא להם כאילו הוגדרו על הטיפוס החיצוני. מהקובץ `filestore.go`:

<div dir="ltr" align="left">

```go
// FileStore is a MemStore that saves itself to a JSON file after every change.
// It embeds MemStore: Account and Transactions are promoted unchanged, while
// CreateAccount and Record are redefined here to call the embedded version and then save.
type FileStore struct {
	MemStore
	path string
}
```

</div>

שתי מתודות מקודמות כמו שהן. שתיים מוגדרות מחדש על `*FileStore`, ומסתירות את המקודמות. ההגדרה החדשה
קוראת לגרסה המוטמעת בשמה המלא, ואז שומרת:

<div dir="ltr" align="left">

```go
func (s *FileStore) CreateAccount(a Account) error {
	if err := s.MemStore.CreateAccount(a); err != nil {
		return err
	}
	return s.save()
}
```

</div>

למה זו לא ירושה, ומה זה אומר בפועל:

- **אין קריאה וירטואלית חזרה.** המתודה `MemStore.Record` קוראת ל־`s.Account`. זו תמיד הגרסה של
  המאגר המוטמע, גם כשהיא רצה בתוך המאגר החיצוני. אין `virtual` ואין `super`.
- **אין יחס „הוא”.** אי אפשר להעביר `*FileStore` לפונקציה שמצפה ל־`*MemStore`. שניהם מממשים את `Store`,
  וזה הקשר היחיד ביניהם.
- **הכול מקודם, גם מה שלא רצית.** אם תטמיעי טיפוס עם המתודות `Lock` ו־`Unlock`, הן יהפכו לחלק
  מהממשק הציבורי שלך. לכן נעילה שומרים בשדה עם שם, לא בהטמעה.

הדוגמה `Example_embeddingPromotes` בקובץ `example_test.go` מראה גם קריאה למתודה המקודמת וגם קריאה
ישירה לגרסה המוטמעת, שלא שומרת לקובץ. ואם תכתבי ליטרל של `FileStore`, החל מ־Go 1.27 מותר לתת בו ערך
גם לשדה מקודם, לא רק לשדות ברמה העליונה.

## 4.9 תוכנית קטנה: `ledger`

הספרייה מחולקת לשלושה חלקים: הטיפוס `Ledger` מחזיק את הכללים ותלוי רק בממשק, ושני מימושים עומדים
מאחוריו.

שני פרטים במימושים ששווה לראות. הראשון: הערך האפס של `MemStore` מוכן לשימוש. maps שערכם `nil`
מותרים לקריאה, ולכן רק כתיבה מאתחלת אותם, דרך `lazyInit`. מי שכותבת `&MemStore{}` לא צריכה בנאי.
השני: `Transactions` מחזירה עותק עם `slices.Clone`, כדי שקוראת שמשנה את הרשימה לא תשנה את האחסון.
זה באג הכינוי מפרק 3, בתפקיד אמיתי.

המאגר `FileStore` שומר את כל המצב אחרי כל שינוי. הוא כותב לקובץ זמני באותה תיקייה ואז מחליף את השם, כך
שקריסה באמצע לא משאירה קובץ חצי כתוב. לפני הכתיבה הוא ממיין, כי סדר המעבר על maps אקראי, וקובץ
שמשתנה בלי שהנתונים השתנו מבלבל את Git.

הבדיקות הן חבילה אחת שרצה על שני המימושים. מהקובץ `store_test.go`:

<div dir="ltr" align="left">

```go
func TestMemStore(t *testing.T) {
	testStore(t, func(t *testing.T) Store { return &MemStore{} })
}

func TestFileStore(t *testing.T) {
	testStore(t, func(t *testing.T) Store {
		s, err := OpenFileStore(filepath.Join(t.TempDir(), "ledger.json"))
		if err != nil {
			t.Fatal(err)
		}
		return s
	})
}
```

</div>

הפונקציה `testStore` מקבלת פונקציה שיוצרת מאגר ריק, ומריצה עליו את כל תתי־הבדיקות. מימוש שלישי,
על מסד נתונים למשל, יקבל את אותה חבילה בשלוש שורות. הפונקציה `t.TempDir` יוצרת תיקייה זמנית
ומוחקת אותה בסוף הבדיקה.

כך נראית הפעלה של הכלי:

<div dir="ltr" align="left">

```bash
go build -o bin/ledger ./cmd/ledger
bin/ledger open alice "Alice Levi"
bin/ledger deposit alice 120 "משכורת"
bin/ledger withdraw alice 45.50 "סופר"
bin/ledger withdraw alice 100
echo "exit=$?"
```

</div>

<div dir="ltr" align="left">

```text
opened alice for Alice Levi
alice balance 120.00
alice balance 74.50
ledger: withdrawal of 100.00 on alice: insufficient funds in alice: balance 74.50, requested 100.00 (short by 25.50)
exit=3
```

</div>

השורה הרביעית מראה את כל השרשרת: ההקשר שהוסיף `Ledger`, ההודעה של טיפוס השגיאה, והחוסר ש־`report`
שלפה ממנו. קוד היציאה 3 מאפשר לסקריפט להבדיל בין „ביקשת לא נכון” לבין „אין מספיק כסף”.

## סיכום

- מתודות מוגדרות על כל טיפוס בעל שם; receiver מצביע כשהמתודה משנה מצב, ואז כל המתודות מצביע.
- ממשקים מרומזים ומוגדרים אצל מי שמשתמשת בהם; שורת `var _ Store` נותנת בדיקת קומפילציה.
- ממשק שמכיל מצביע `nil` אינו `nil`. החזירי `nil` מפורש.
- sentinel לשגיאה שמספיק לזהות, טיפוס לשגיאה שנושאת נתונים; עוטפים עם `%w` ושולפים עם `errors.Is`
  ו־`errors.AsType`.
- panic לבאג של המתכנתת; `recover` רק בגבול. generics למבני נתונים ולעזרים, עם `~` לטיפוסים בעלי שם.
- הטמעה מקדמת שדות ומתודות, אבל אין בה קריאה וירטואלית ואין יחס „הוא”.

## מונחים לזכור

| מונח בעברית | באנגלית | במשפט אחד |
| --- | --- | --- |
| מתודה | method | פונקציה עם receiver, מוגדרת על טיפוס בעל שם |
| קבוצת מתודות | method set | המתודות שטיפוס מסוים מממש; קובעת אילו ממשקים הוא מספק |
| ממשק | interface | רשימת מתודות; כל טיפוס שיש לו אותן מממש אותו בלי הצהרה |
| שגיאת sentinel | sentinel error | ערך שגיאה קבוע של חבילה, שמזהים עם `errors.Is` |
| טיפוס שגיאה | error type | struct שמממש `error` ונושא נתונים, ששולפים עם `errors.As` |
| עטיפה | wrapping | הוספת הקשר לשגיאה עם `%w` בלי לאבד את המקורית |
| אילוץ | constraint | הממשק שמגביל פרמטר טיפוס בפונקציה גנרית |
| קבוצת טיפוסים | type set | אילוץ כמו `~int64`, שמונה טיפוסים מותרים |
| הטמעה | embedding | שדה בלי שם, שהשדות והמתודות שלו מקודמים לטיפוס החיצוני |

## טעויות נפוצות

- **להחזיר משתנה שגיאה מטיפוס קונקרטי.** כשהוא `nil`, הקוראת מקבלת ממשק שאינו `nil`. החזירי `nil`
  מפורש.
- **להשוות שגיאות עם `==`.** אחרי עטיפה ההשוואה נכשלת. השתמשי ב־`errors.Is`.
- **לעטוף עם `%v` כשהקוראת צריכה את הסיבה.** ההודעה נשמרת, אבל השגיאה עצמה אובדת ו־`errors.Is` לא
  ימצא אותה.
- **לרשום ללוג וגם להחזיר.** השגיאה תופיע פעמיים. החליטי במקום אחד, בגבול.
- **לממש ממשק על ערך ולהעביר מצביע, או להפך.** קבוצת המתודות של `T` אינה של `*T`. הוסיפי שורת
  `var _ Iface = (*T)(nil)`.
- **להטמיע כדי לחסוך הקלדה.** כל מתודה של המוטמע הופכת ל־API שלך. שדה עם שם הוא ברירת המחדל.

## שאלות ראיון

1. **מה ההבדל בין ממשק ב־Go ל־trait ב־Rust, ומה כל גישה מאפשרת?** ממשק מתממש מעצמו כשיש לטיפוס
   המתודות, ולכן הצרכנית יכולה להגדיר ממשק קטן אחרי מעשה, גם לטיפוס מספרייה זרה. trait ב־Rust מצריך
   הצהרה, אבל נותן קריאה סטטית בלי עלות, מימושים גנריים וטיפוסים משויכים. ב־Go כל קריאה דרך ממשק
   היא שיגור דינמי (dynamic dispatch). הבחירה ב־Go מעדיפה צימוד חלש ופשטות על פני כוח ביטוי.

2. **מתי תגדירי sentinel ומתי טיפוס שגיאה?** sentinel כשהקוראת צריכה רק לזהות את המקרה, כמו „לא נמצא”.
   טיפוס כשהיא צריכה נתונים כדי לפעול, כמו כמה כסף חסר או אחרי כמה זמן לנסות שוב. בשני המקרים זה הופך
   לחלק מה־API: אחרי שקוראות בודקות אותם, שינוי שלהם שובר אותן. לכן לא חושפים כל שגיאה פנימית, ועוטפים
   עם `%v` את מה שאינו חוזה.

   **המשך:** למה `errors.Is` ולא `==`? כי אחרי עטיפה השגיאה היא עוטף שמכיל את ה־sentinel. הפונקציה `errors.Is`
   עוברת על כל השרשרת, כולל עצים של שגיאות מרובות.

3. **למה ממשק שמכיל מצביע `nil` אינו `nil`?** ערך ממשק הוא זוג: טיפוס דינמי וערך. הוא `nil` רק אם שניהם
   ריקים. מצביע `nil` מטיפוס ידוע ממלא את חלק הטיפוס, ולכן ההשוואה ל־`nil` שקרית. הבאג מופיע כשפונקציה
   מחזירה משתנה מטיפוס שגיאה קונקרטי; התיקון הוא להחזיר `nil` מפורש.

4. **מתי `recover` מוצדק?** בגבול שבו יש למי לדווח ואפשר להמשיך: בקשת HTTP בודדת, משימה בודדת בתור
   עבודות. שם ממירים panic לשגיאה, רושמים ללוג עם מחסנית, וממשיכים לשרת אחרים. באמצע הקוד זה מסתיר
   באג ומשאיר מצב חלקי. ותפיסה לא עובדת בין goroutines, ולכן כל goroutine ארוכת חיים צריכה גבול
   משלה.

   **המשך:** ולמה לא להשתמש ב־panic כמו בחריגה? כי זה הופך כל קריאה לנקודת יציאה נסתרת, והקוראת לא
   רואה אותה בחתימה. ב־Go שגיאה צפויה היא ערך שחוזר.

## פתק Obsidian

הפתק המוכן נמצא גם כקובץ: `obsidian/04-go-types-interfaces-errors.md`. זה תוכנו:

- **receivers:** מצביע כשמשנים מצב, ואז כל המתודות מצביע; קבוצת המתודות קובעת את הממשקים.
- **ממשקים:** מרומזים, קטנים, מוגדרים אצל הצרכנית; שורת `var _` לבדיקת קומפילציה.
- **nil בתוך ממשק:** טיפוס עם ערך ריק אינו `nil`; החזירי `nil` מפורש.
- **שגיאות:** sentinel לזיהוי, טיפוס לנתונים; `%w` להקשר; `errors.Is` ו־`errors.AsType` לשליפה.
- **גבול אחד:** בשכבות הפנימיות מוסיפים הקשר ומחזירים; בגבול מחליטים, מדפיסים ובוחרים קוד יציאה.
- **panic:** לבאג בלבד; `recover` רק בגבול של goroutine או בקשה.
- **generics:** `any`, קבוצות טיפוסים עם `~`; מאז 1.27 מתודות גנריות, אבל לא בממשקים.
- **הטמעה:** קידום בלי ירושה; אין קריאה וירטואלית ואין יחס „הוא”.
- **שאלה פתוחה:** ___

<div dir="ltr" align="left">

```text
var _ Store = (*MemStore)(nil)               // compile-time interface check
var ErrX = errors.New("x")                    // sentinel
type XError struct{ ... }; func (e XError) Error() string
return fmt.Errorf("op %s: %w", id, err)       // wrap, keep cause
errors.Is(err, ErrX)
var t XError; errors.As(err, &t)
t, ok := errors.AsType[XError](err)           // Go 1.26+
defer func() { if r := recover(); r != nil { ... } }()
func F[T any](xs []T) ...    type Integer interface{ ~int | ~int64 }
func (p Page[T]) Map[U any](f func(T) U) Page[U]   // Go 1.27+
type FileStore struct { MemStore; path string }    // embedding
exit codes: 0 ok / 1 error / 2 usage / 3 insufficient / 4 not found
```

</div>

## תרגיל

**מטרה:** לבנות את `ledger`: ספרייה עם ממשק `Store` ושני מימושים, sentinels וטיפוס שגיאה שעוברים עטיפה,
`Page[T]` גנרי, וכלי שורת פקודה מעליה, עם חבילת בדיקות אחת לשני המימושים. **זמן:** שני ימי לימוד, ימים 6 ו־7, כ־50 דקות בנייה ביום 6 וכ־60 ביום 7. **ידע מקדים:** פרק 3 כולו; סעיפים 4.1 עד 4.5 ליום 6
ו־4.6 עד 4.9 ליום 7.

הצעדים, קוד ההתחלה, הרמזים והפתרון נמצאים ב־
[`exercises/ch04-ledger/README.md`](../exercises/ch04-ledger/README.md).

**קריטריון סיום:**

- הבדיקות מראות ש־`errors.Is` מוצאת את ה־sentinel ו־`errors.As` שולפת את טיפוס השגיאה דרך עטיפה.
- חבילת הבדיקות המשותפת עוברת גם על המאגר בזיכרון וגם על המאגר בקובץ.
- הכלי פותח חשבון, רושם טרנזקציות ומדפדף בהיסטוריה, עם קודי יציאה שונים לכל סוג כשל.
- שלושת הכלים מפרק 3 שקטים.

## העשרה אופציונלית

- בבלוג של Go יש מאמר מ־2019 על השגיאות בגרסה 1.13, שבה נוספה העטיפה. הוא מסביר היטב את ההחלטה.
- הערות השחרור של Go 1.26 ו־1.27 מתארות את `errors.AsType` ואת המתודות הגנריות; קראי את הסעיפים האלה
  במקור.
- נסי לשנות את `apply` כך שתחזיר `&InsufficientFundsError{...}`, והריצי את הבדיקות. מה נכשל, ולמה?
  אחר כך עטפי מצביע כזה עם `%w` במקום אחר, והריצי `go vet`.

</div>
