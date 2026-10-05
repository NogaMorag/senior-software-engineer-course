---
direction: rtl
lang: he
---

<div dir="rtl" lang="he">

# תרגיל פרק 27: סקירה של בקשת שינוי, ומודול שהכלים מנקים

**מטרה:** לסקור בקשת שינוי אמיתית עם בעיות שתולות, ולכתוב את ההערות לפני שפותחים את מפתח התשובות. אחר כך
להריץ את כל כלי הניתוח הסטטי של הפרק על מודול Go ועל חבילת Python עם בעיות שתולות, ולתקן כל ממצא.

**זמן משוער:** יום לימוד קל אחד, יום 82: כ־25 דקות לסקירה וכ־20 דקות לכלים.

**ידע מקדים:** פרק 27; פרק 3 (staticcheck); פרק 5 (הדגל `-race`); פרק 6 (handlers ולוגים); פרק 7 (uv, ruff
ו־mypy); govulncheck מותקן לפי פרק 2, סעיף 2.7. התרגיל לא מייבא קוד מתרגילים אחרים.

## מה בתיקייה הזאת

| קובץ | תפקיד |
| --- | --- |
| `data/pr-description.md` | התיאור של בקשת השינוי, כמו שהכותבת כתבה אותו |
| `data/review.diff` | ה־diff לסקירה: handler, מאגר ובדיקה של שירות מסמכים |
| `starter/review.md` | תבנית לסקירה שלך |
| `starter/seeded/` | מודול Go עם בעיות שהכלים מוצאים, כולל `.golangci.yml` |
| `starter/seeded-py/` | חבילת Python קטנה (`tally`) עם בעיות ש־ruff ו־mypy מוצאים |
| `tests/seeded/internal/...` | הבדיקות הקנוניות של מודול Go; עותקים זהים ב־`starter/` וב־`solution/` |
| `tests/seeded-py/tests/test_visits.py` | הבדיקות הקנוניות של חבילת Python |
| `solution/review.md` | מפתח התשובות של הסקירה: ממצאים, חומרה והערות מוצעות |
| `solution/seeded/`, `solution/seeded-py/` | המודולים אחרי התיקון, שקטים בכל הכלים |

> **שימי לב:** קוד ההתחלה של שני המודולים לא עובר `go vet` ו־ruff בכוונה: זו המשימה. בכל תרגיל אחר קוד
> ההתחלה שקט בכלים.

## צעדים

**1. העתקה ושמירה ראשונה.**

<div dir="ltr" align="left">

```bash
mkdir -p ~/senior-course/ch27
cp -r $COURSE/exercises/ch27-review/starter/. ~/senior-course/ch27/
cp -r $COURSE/exercises/ch27-review/data ~/senior-course/ch27/
cd ~/senior-course/ch27
git add . && git commit -m "ch27: starter"
```

</div>

**2. הסקירה (כ־25 דקות, בלי להסתכל בפתרון).** קראי קודם את התיאור, ואז את ה־diff:

<div dir="ltr" align="left">

```bash
less data/pr-description.md
less data/review.diff
```

</div>

מלאי את `review.md`: לכל ממצא קובץ ושורה, תווית חומרה, מה יכול להשתבש, והערה כמו שהיית כותבת לכותבת.
כתבי גם באיזה קובץ התחלת ולמה, ומשפט סיכום עם החלטה. יש בבקשה כשמונה בעיות בחומרות שונות, ודבר אחד ששווה
שבח. שמרי (`git commit`) **לפני** הצעד הבא; זה מה שמוכיח שהסקירה שלך.

**3. השוואה.** פתחי את המפתח, והשלימי בתבנית את שלוש השורות של „אחרי ההשוואה”:

<div dir="ltr" align="left">

```bash
less $COURSE/exercises/ch27-review/solution/review.md
```

</div>

**4. התקנת golangci-lint.** זה הכלי היחיד בפרק שעוד לא הותקן. קובץ ההרצה נכנס ל־`~/.local/bin`, כמו שאר
הכלים מפרק 2 (הגרסה נכונה לאוקטובר 2026; ראי נספח ג):

<div dir="ltr" align="left">

```bash
cd /tmp
curl -sSfLO https://github.com/golangci/golangci-lint/releases/download/v2.14.0/golangci-lint-2.14.0-linux-amd64.tar.gz
tar xzf golangci-lint-2.14.0-linux-amd64.tar.gz
mv golangci-lint-2.14.0-linux-amd64/golangci-lint ~/.local/bin/
rm -r golangci-lint-2.14.0-linux-amd64 golangci-lint-2.14.0-linux-amd64.tar.gz
golangci-lint --version
```

</div>

> **ב־Mac:** בשם הקובץ, `darwin-arm64` במקום `linux-amd64` (בשלוש השורות שמזכירות אותו).

**5. מודול Go: להריץ ולקרוא.** הריצי כל כלי בנפרד, ורשמי לעצמך כמה ממצאים כל אחד מצא:

<div dir="ltr" align="left">

```bash
cd ~/senior-course/ch27/seeded
go vet ./...
staticcheck ./...
golangci-lint run ./...
govulncheck ./...
go test -race ./...
```

</div>

**6. מודול Go: לתקן.** תקני ממצא אחד בכל פעם והריצי שוב. סדר מומלץ: קודם מה שמשנה התנהגות (הנעילה,
ה־printf, השגיאה שנדרסת, ההשוואה עם `!=`, גוף התשובה שלא נסגר, ה־cancel), אחר כך הסגנון, ובסוף
התלות. לחולשה:

<div dir="ltr" align="left">

```bash
go get golang.org/x/text@v0.40.0
go mod tidy
```

</div>

ואת `strings.Title` מחליפים ב־`cases.Title(language.Und)` מהחבילה `golang.org/x/text/cases`, שהמודול כבר
תלוי בה. סיימת כשכל חמש הפקודות של צעד 5 שקטות או ירוקות. שמרי.

**7. חבילת Python.**

<div dir="ltr" align="left">

```bash
cd ~/senior-course/ch27/seeded-py
uv sync
uv run ruff check .
uv run mypy src
uv run pytest -q
```

</div>

תקני את ארבעת הממצאים של ruff ואת זה של mypy ביד, לא עם `--fix`, כדי לראות מה כל אחד אומר. ודאי שגם
`uv run ruff format --check .` שקט, ושמרי.

## פלט צפוי

בצעד 5, `go vet` על קוד ההתחלה:

<div dir="ltr" align="left">

```text
internal/visits/visits.go:30:9: Lookup passes lock by value: example.com/ch27-review/internal/visits.Counter contains sync.Mutex
internal/visits/visits.go:52:22: fmt.Sprintf format %s has arg n of wrong type int
internal/holiday/holiday.go:30:7: the cancel function returned by context.WithTimeout should be called, not discarded, to avoid a context leak
```

</div>

הסיכום של golangci-lint הוא `12 issues`, ו־govulncheck מדווח על `GO-2022-1059` עם ה־trace
`greet.Word calls language.ParseAcceptLanguage` ויוצא בקוד 3. הבדיקות לפני התיקון:

<div dir="ltr" align="left">

```text
--- FAIL: TestTodayRejectsBadAnswers (0.01s)
    --- FAIL: TestTodayRejectsBadAnswers/days_of_the_wrong_type (0.00s)
        holiday_test.go:45: Today = {Name:Sukkot Date:2026-10-05 Days:0}; want an error
FAIL	example.com/ch27-review/internal/visits [build failed]
```

</div>

אחרי התיקון (הזמנים אצלך יהיו אחרים):

<div dir="ltr" align="left">

```text
ok  	example.com/ch27-review/internal/greet	1.270s
ok  	example.com/ch27-review/internal/holiday	2.005s
ok  	example.com/ch27-review/internal/visits	1.802s
0 issues.
No vulnerabilities found.
```

</div>

ובחבילת Python, לפני התיקון `Found 4 errors.` מ־ruff ושגיאת `return-value` אחת מ־mypy; אחרי התיקון:

<div dir="ltr" align="left">

```text
All checks passed!
Success: no issues found in 2 source files
10 passed
```

</div>

## רשימת בדיקה

- [ ] הסקירה שמורה בשמירה נפרדת לפני שפתחת את המפתח, עם תווית חומרה לכל הערה.
- [ ] כתבת באיזה קובץ התחלת ולמה, ומשפט סיכום עם החלטה.
- [ ] רשימת המפוספסים כתובה, עם הרגל אחד לסקירה הבאה.
- [ ] במודול Go, `go vet`, staticcheck ו־golangci-lint שקטים, ו־`go test -race ./...` ירוק.
- [ ] הכלי govulncheck מסיים ב־`No vulnerabilities found.`
- [ ] בחבילת Python, ruff (בדיקה ועיצוב) ו־mypy שקטים, ו־pytest עובר.
- [ ] לא השתקת אף ממצא עם `//nolint` או `# noqa`.

## טעויות נפוצות ותיקונים

- **לפתוח את המפתח „רק להציץ” באמצע הסקירה.** אז אי אפשר לדעת מה מצאת לבד. סגרי את הסקירה בשמירה, ורק אז
  השווי.
- **`unsupported version of the configuration` מ־golangci-lint.** הכלי מצא קובץ תצורה אחר, ישן, בתיקייה מעל
  או בתיקיית הבית. הריצי מתוך `seeded`, או מחקי את הקובץ הישן.
- **לתקן את הנעילה ולשכוח את שאר המתודות.** אחרי שינוי של `Lookup` ל־`*Counter`, ודאי שלכל המתודות של
  `Counter` יש receiver מצביע; `go vet` שקט רק כשכולן כאלה.
- **לבדוק את השגיאה של `json.Unmarshal` ועדיין להחזיר את הערך.** הבדיקה `days_of_the_wrong_type` נכשלת עד
  שהשגיאה של הפענוח מוחזרת לפני `validate`.
- **לעדכן את `x/text` ל־`v0.3.8` בלבד.** החולשה הראשונה נעלמת, אבל `govulncheck -show verbose` עדיין מראה את
  `GO-2026-5970`. משדרגים לגרסה שמופיעה בצעד 6.
- **להריץ את `uv run ruff check --fix .` ולהמשיך.** התיקון האוטומטי לא נוגע ב־`B006` וב־`E722`, והבדיקה
  `test_tally_calls_are_independent` ממשיכה להיכשל. מתקנים ביד.

## רמזים

- בסקירה, חפשי שלוש בעיות אבטחה, שתיים של משאבים ומקביליות, שתיים של שגיאות וקלט, ובדיקה אחת שלא יכולה
  להיכשל. ושאלי לגבי ה־context ב־`Get`.
- את `Sprintf` לתוך `ORDER BY` לא פותרים עם placeholder: שמות עמודות לא נקשרים. רשימה סגורה של ערכים מותרים.
- ב־`holiday.go`, `defer func() { _ = resp.Body.Close() }()` מיד אחרי בדיקת השגיאה של `Do`, ו־`defer cancel()`
  מיד אחרי `WithTimeout`.
- ב־`main.go`, `&http.Server{...}` עם `ReadHeaderTimeout` ושאר הזמנים של פרק 6, סעיף 6.5, ואז `srv.ListenAndServe()`.
- ב־Python, הצורה המקובלת לברירת מחדל ניתנת לשינוי היא `counts: dict[str, int] | None = None`, ובתוך הפונקציה
  `counts = {} if counts is None else counts`.

## הפתרון

הפתרון נמצא ב־`solution/`. **למה הוא עובד כך:** מפתח הסקירה מסודר לפי הסדר שבו סוקרת לפי סיכון קוראת את
הקוד, ולכל ממצא יש הערה מוכנה עם נימוק והצעה, וטבלה בסוף שמראה מה מהם כלי היה תופס. במודול Go כל תיקון
משנה התנהגות שבדיקה מוכיחה (נעילה על מצביע, שגיאת פענוח שמוחזרת, `errors.Is` על שגיאה עטופה), ושדרוג אחד של
`x/text` סוגר גם את החולשה שנקראת וגם את זו שרק קיימת. ובחבילת Python, ברירת המחדל `None` היא מה שמבדיל
בין שתי קריאות.

## מה התרגיל מדגים

- סקירה לפי סיכון, עם תוויות חומרה, ושמירה של הסקירה לפני ההשוואה כדי למדוד את עצמך ביושר.
- שלוש שכבות ניתוח סטטי ב־Go, ומה כל אחת תופסת שהאחרות לא.
- קובץ תצורה של golangci-lint בגרסה 2.
- ההבדל בין חולשה שנקראת לבין חולשה שרק קיימת בגרסה, כפי ש־govulncheck מציג אותו.
- ruff ו־mypy כשני מסננים שונים: דפוסים מסוכנים מול טיפוסים שלא מתיישבים.

</div>
