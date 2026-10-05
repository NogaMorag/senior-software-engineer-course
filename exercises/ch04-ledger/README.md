---
direction: rtl
lang: he
---

<div dir="rtl" lang="he">

# תרגיל פרק 4: `ledger`

**מטרה:** לבנות ספר חשבונות קטן: ספרייה שבה הכללים תלויים רק בממשק `Store`, שני מימושים מאחוריו
(בזיכרון ובקובץ JSON), שגיאות sentinel וטיפוס שגיאה שעוברים עטיפה, דף גנרי `Page[T]`, וכלי שורת
פקודה מעל הכול. חבילת בדיקות אחת רצה על שני המימושים.

**זמן משוער:** שני ימי לימוד. יום 6: כ־50 דקות בנייה (טיפוסים, שגיאות, מאגר בזיכרון). יום 7: כ־60 דקות
בנייה (מאגר בקובץ, generics, הכלי).

**ידע מקדים:** פרקים 3 ו־4. התרגיל לא מייבא קוד מתרגיל קודם, אבל חוזר על שני רעיונות מפרק 3: ביטוי
slice מלא נגד כינוי, ומיון לפני פלט שמקורו ב־map.

## מה בתיקייה הזאת

| קובץ | תפקיד |
| --- | --- |
| `starter/cents.go` | הטיפוס `Cents` ופירוק סכומים; מוכן |
| `starter/store.go` | הטיפוסים, הממשק `Store` והפונקציה `apply`; ב־`apply` יש `TODO` |
| `starter/errors.go` | שגיאות ה־sentinel והטיפוס `InsufficientFundsError`; בשתי המתודות יש `TODO` |
| `starter/memstore.go` | המאגר בזיכרון; ב־`CreateAccount`, `Record` ו־`Transactions` יש `TODO` |
| `starter/ledger.go` | הכללים מעל `Store`; ב־`post` יש `TODO` של עטיפה |
| `starter/filestore.go` | המאגר בקובץ, עם הטמעה; ב־`save` ובשתי ההגדרות מחדש יש `TODO` |
| `starter/page.go` | `Page[T]`, `Paginate`, `Map` ו־`Sum`; בשלושתן יש `TODO` |
| `starter/cmd/ledger/main.go` | הכלי; ב־`history` וב־`report` יש `TODO` |
| `tests/*_test.go` | בדיקות הספרייה: חבילה משותפת לשני המאגרים, שגיאות, דפים, סכומים, דוגמאות |
| `tests/cmd/ledger/main_test.go` | הכלי כולו: מפגש שלם מול קובץ אחד, וקודי היציאה |
| `solution/` | הפתרון המלא, עם עותקים זהים של הבדיקות |

אין כאן `data/`: הבדיקות יוצרות קבצים בתיקייה זמנית עם `t.TempDir`.

## יום 6: טיפוסים, שגיאות ומאגר בזיכרון

**1. העתיקי את קוד ההתחלה.** הפעם המודול כבר קיים בקוד ההתחלה.

<div dir="ltr" align="left">

```bash
mkdir -p ~/senior-course/ch04
cd ~/senior-course/ch04
cp -r "$COURSE"/exercises/ch04-ledger/starter/. .
printf 'bin/\nledger.json\n' > .gitignore
go build ./... && go vet ./... && echo builds
```

</div>

הקובץ `.gitignore` מונע מהבינארי ומקובץ הנתונים של ההרצות שלך להיכנס ל־Git.

**2. קראי בסדר הזה:** `cents.go`, `store.go`, `errors.go`, `memstore.go`, `ledger.go`. שימי לב לשורת
`var _ Store = (*MemStore)(nil)`, ל־`iota + 1` ב־`Kind`, ולכך שהמתודה `Error` מוגדרת על ערך.

**3. מלאי את `Error` ואת `Shortfall`** בקובץ `errors.go`, ואת המקרה של משיכה ב־`apply`, בקובץ
`store.go`. משיכה גדולה מהיתרה מחזירה `InsufficientFundsError` כערך, לא כמצביע.

**4. מלאי את שלוש המתודות של `MemStore`.** ב־`Record` הסדר חשוב: קודם בודקים ומחשבים, ורק אם הכול
הצליח מקצים מספר וכותבים. כך משיכה שנכשלה לא משאירה עקבות. ב־`Transactions` החזירי עותק.

**5. הוסיפי הקשר ב־`post`**, בקובץ `ledger.go`: עטיפה עם `%w`, באותה צורה של השגיאה שמעליה.

**6. הבדיקות של היום.** שתי פקודות, כי השנייה בוחרת תתי־בדיקות של `TestMemStore` בלבד:

<div dir="ltr" align="left">

```bash
go test -run 'Wrapping|InvalidAmount|Cents' -v .
go test -run 'TestMemStore/(open|deposit|missing|returned)' -v .
```

</div>

שתי תתי־בדיקות של `TestMemStore` עוברות דרך ההיסטוריה ותלויות ב־`Paginate`, ולכן הן נשארות ליום 7.
הבדיקות `TestSentinelThroughWrapping` ו־`TestTypedErrorThroughWrapping` הן קריטריון הסיום של היום.

## יום 7: מאגר בקובץ, generics והכלי

**7. מלאי את `Paginate`, `Map` ו־`Sum`** בקובץ `page.go`, והריצי את הבדיקות שלהן:

<div dir="ltr" align="left">

```bash
go test -run 'Paginate|Page|Sum' -v .
```

</div>

**8. מלאי את `save`** בקובץ `filestore.go`, והוסיפי קריאה אליה בשתי המתודות שמוגדרות מחדש. כל שאר
המתודות מגיעות מ־`MemStore` דרך ההטמעה, בלי שורת קוד אחת.

**9. כל בדיקות הספרייה.** חבילת הבדיקות המשותפת רצה עכשיו על שני המאגרים:

<div dir="ltr" align="left">

```bash
go test -v -run 'TestMemStore|TestFileStore' . | grep -E '^(--- |ok)'
```

</div>

**10. השלימי את הכלי:** `history` עם `page.Map`, ו־`report` עם `errors.AsType` ו־`errors.Is`, לפי
ההערות. ואז הכול:

<div dir="ltr" align="left">

```bash
go test ./...
gofmt -l .
go vet ./...
staticcheck ./...
```

</div>

**11. מפגש אמיתי עם הכלי.** הוא שומר ל־`ledger.json` בתיקייה הנוכחית:

<div dir="ltr" align="left">

```bash
go build -o bin/ledger ./cmd/ledger
bin/ledger open alice "Alice Levi"
bin/ledger deposit alice 120 "משכורת"
bin/ledger withdraw alice 45.50 "סופר"
bin/ledger withdraw alice 100
echo "exit=$?"
bin/ledger deposit alice 1
bin/ledger deposit alice 2
bin/ledger history -limit 3 alice
bin/ledger history -cursor 3 -limit 3 alice
bin/ledger balance bob
echo "exit=$?"
```

</div>

פתחי את `ledger.json` וראי את המבנה: חשבונות וטרנזקציות ממוינים, וההערות בעברית כמו שהן.

**12. שמרי את העבודה.**

<div dir="ltr" align="left">

```bash
cd ~/senior-course
git add ch04
git commit -m "ch04: ledger with two stores, wrapped errors and Page[T]"
```

</div>

## פלט צפוי

המפגש מצעד 11. השעות בהיסטוריה יהיו השעה שבה הרצת, ב־UTC:

<div dir="ltr" align="left">

```text
opened alice for Alice Levi
alice balance 120.00
alice balance 74.50
ledger: withdrawal of 100.00 on alice: insufficient funds in alice: balance 74.50, requested 100.00 (short by 25.50)
exit=3
alice balance 75.50
alice balance 77.50
#1   2026-10-05 08:20  deposit        120.00  משכורת
#2   2026-10-05 08:20  withdrawal      45.50  סופר
#3   2026-10-05 08:20  deposit          1.00
more: ledger history -cursor 3 -limit 3 alice
#4   2026-10-05 08:20  deposit          2.00
ledger: balance of bob: account not found
exit=4
```

</div>

והבדיקות (הזמנים אצלך יהיו אחרים):

<div dir="ltr" align="left">

```text
ok  	example.com/ch04-ledger	0.950s
ok  	example.com/ch04-ledger/cmd/ledger	1.039s
```

</div>

## רשימת בדיקה

- [ ] `errors.Is` מוצאת את `ErrAccountNotFound` דרך עטיפה, וההשוואה הישירה נכשלת.
- [ ] `errors.As` ו־`errors.AsType` שולפות את `InsufficientFundsError` גם אחרי עטיפה כפולה.
- [ ] משיכה שנכשלה לא משנה יתרה ולא נרשמת בהיסטוריה.
- [ ] חבילת הבדיקות המשותפת עוברת על שני המאגרים, וקובץ שנפתח מחדש שומר הכול.
- [ ] הכלי מחזיר 2, 3 ו־4 לשגיאת שימוש, חוסר יתרה וחשבון חסר.
- [ ] שלושת הכלים לא מדפיסים דבר.
- [ ] העבודה שמורה ב־Git, בלי הבינארי ובלי קובץ הנתונים.

## טעויות נפוצות ותיקונים

- **`TestTypedErrorThroughWrapping` נכשלת עם `errors.As found no InsufficientFundsError`.** `apply`
  מחזירה מצביע, `&InsufficientFundsError{...}`, והבדיקה מחפשת ערך. החזירי ערך.
- **`TestSentinelThroughWrapping` נכשלת עם `should add context`.** `post` מחזירה את השגיאה כמו שהיא.
  עטפי עם `fmt.Errorf` ו־`%w`, לא עם `%v`.
- **„overdraft ... changes nothing” נכשלת כי ההיסטוריה ארוכה באחד.** `Record` כותבת לפני ש־`apply`
  הצליחה. הזיזי את הכתיבה אחרי בדיקת השגיאה.
- **`TestFileStoreSurvivesReopen` מוצאת יתרה 0 אחרי פתיחה מחדש.** המתודות שמוגדרות מחדש לא קוראות
  ל־`save`, או ש־`save` עדיין מחזירה את ה־`TODO`. בדקי את שתיהן.
- **panic מסוג `assignment to entry in nil map`.** מתודה כותבת ל־map לפני `lazyInit`. קראי לה בתחילת
  כל מתודה שכותבת.
- **`TestPageItemsDoNotAliasTheSource` נכשלת.** החיתוך הוא `items[cursor:end]` בלי האינדקס השלישי,
  והוספה לדף דורסת את הרשימה. כתבי `items[cursor:end:end]`.

## רמזים

- ב־`Record`: `a, err := s.Account(tx.Account)`, ואז `a, err = apply(a, tx)`, ורק אחר כך `s.lastID++`.
- ב־`Paginate`: חשבי את `end` עם `min`, והחזירי שגיאה שעוטפת את `ErrBadCursor` עם `%w`.
- ב־`Map`: הקצי מראש עם `make([]U, 0, len(p.Items))`, והעתיקי את `Next` ואת `HasMore`.
- ב־`save`: `os.CreateTemp` באותה תיקייה של הקובץ, ואז `os.Rename`. שינוי שם באותה מערכת קבצים
  מחליף את הקובץ בפעולה אחת.
- ב־`report`: בדקי קודם את טיפוס השגיאה, כי הוא צריך הודעה משלו, ורק אחר כך את ה־sentinels.

## הפתרון

הפתרון המלא נמצא בתיקייה `solution/`. הכלל העסקי היחיד, `apply`, משותף לשני המאגרים, ולכן הם לא
יכולים להתנהג אחרת. כל שכבה מוסיפה הקשר עם `%w` ומחזירה, והגבול היחיד שמחליט הוא `report`, שבוחר
הודעה וקוד יציאה לפי `errors.AsType` ו־`errors.Is`. `FileStore` מטמיע את `MemStore` ומגדיר מחדש רק את
שתי המתודות שמשנות מצב, כך שהשמירה לקובץ נוספת בלי לשכפל לוגיקה. והחיתוך בדף מגביל קיבולת, כך שאף
קוראת לא יכולה לדרוס את ההיסטוריה.

## מה התרגיל מדגים

- מתודות על טיפוס בעל שם (`Cents`), ו־receivers של מצביע לטיפוס שמשנה מצב.
- ממשק קטן שמוגדר אצל הצרכנית, ובדיקת מימוש בזמן קומפילציה.
- ערך אפס שימושי: `MemStore` שעובד בלי בנאי.
- sentinel מול טיפוס שגיאה, עטיפה עם `%w`, ושליפה דרך השרשרת.
- panic רק בפונקציית `Must`, ו־`recover` רק בבדיקה.
- generics עם `any` ועם קבוצת טיפוסים, ומתודה גנרית.
- הטמעה שמקדמת מתודות ומגדירה מחדש שתיים, וחבילת בדיקות אחת לשני מימושים.

</div>
