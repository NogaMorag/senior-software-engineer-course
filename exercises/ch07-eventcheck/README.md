---
direction: rtl
lang: he
---

<div dir="rtl" lang="he">

# תרגיל פרק 7: `eventcheck`, כלי שבודק קובץ אירועים

**מטרה:** לבנות כלי שורת פקודה ארוז, עם טיפוסים ובדיקות, שבודק קובץ JSONL של אירועי webhook. הכלי
מדווח על כל שורה פגומה עם מספר השורה, הסיבה והמיקום בתוך האירוע. בדרך: פרויקט uv נעול, בודקים סטטיים
במצב קפדני, ממשקים מבניים, מודלים של pydantic, בדיקות עם פרמטרים, ומדידת פרופיל אחת.

**זמן משוער:** שני ימי לימוד. יום 15: כ־45 דקות בנייה. יום 16: כ־65 דקות בנייה.

**ידע מקדים:** פרק 7, ופרק 2: הכלים, המשתנה `COURSE` והמאגר `~/senior-course`. התרגיל לא מייבא
קוד מתרגילים קודמים.

## מה בתיקייה הזאת

| קובץ | תפקיד |
| --- | --- |
| `starter/pyproject.toml`, `starter/uv.lock` | הפרויקט והנעילה, מוכנים; קוראים אותם ביום 15 |
| `starter/src/eventcheck/report.py` | הדוחות, עם `TODO 1` עד `TODO 3` |
| `starter/src/eventcheck/models.py` | המודלים, עם `TODO 4` עד `TODO 6` (שלושה בודקי תקינות) |
| `starter/src/eventcheck/checker.py` | בדיקת השורות, עם `TODO 7` עד `TODO 9` |
| `starter/src/eventcheck/cli.py`, `generate.py`, `__main__.py` | שורת הפקודה והמחולל, שלמים |
| `tests/conftest.py` | ה־fixtures: אירועים תקינים, כתיבת קובץ זמני, קובץ גדול בטווח `session` |
| `tests/test_report.py`, `test_models.py`, `test_checker.py`, `test_cli.py` | 37 בדיקות; עותקים זהים בתוך `starter/` ו־`solution/` |
| `data/events.jsonl` | 25 אירועים, שישה מהם פגומים בכוונה |
| `solution/` | הפתרון המלא |

## יום 15: הפרויקט, הבודקים והדוחות

**1. ראי מה uv מייצר בעצמו.** בתיקייה זמנית, כדי לא ללכלך את המאגר:

<div dir="ltr" align="left">

```bash
cd /tmp
uv init --package --python 3.14 eventcheck-scratch
find eventcheck-scratch -type f | sort
cat eventcheck-scratch/pyproject.toml
```

</div>

השווי לקובץ הפרויקט של ההתחלה: מה נוסף (`[dependency-groups]`, הגדרות הכלים), ומה הוסר (`authors`).
אחר כך מחקי את התיקייה הזמנית.

**2. העתיקי את קוד ההתחלה ואת הנתונים, וסנכרני מהנעילה:**

<div dir="ltr" align="left">

```bash
mkdir -p ~/senior-course/ch07
cp -r "$COURSE/exercises/ch07-eventcheck/starter/." ~/senior-course/ch07/
cp -r "$COURSE/exercises/ch07-eventcheck/data" ~/senior-course/ch07/
cd ~/senior-course/ch07
uv sync --locked
uv run eventcheck --help
```

</div>

הדגל `--locked` נכשל אם הנעילה לא תואמת לקובץ הפרויקט. כאן הם תואמים, ולכן הוא עובר: זו ההוכחה
שהפרויקט נבנה מהנעילה בעותק נקי.

**3. קראי את הקוד מלמעלה למטה,** לפי הסדר הזה:

| קובץ | מה לחפש בו |
| --- | --- |
| `pyproject.toml` | ארבע טבלאות, ואחריהן ההגדרות של הכלים |
| `checker.py` | המחלקות של בעיה, הצלחה, כישלון ודוח |
| `report.py` | שני ממשקים מבניים ושני מילונים עם צורה |

**4. הריצי את הבודקים.** ruff עובר כבר עכשיו. mypy נכשל על שגיאה אחת, וזה מכוון: הבדיקה ב־`test_report.py`
מבקשת מפתח שעוד לא הוספת ל־`SummaryDict`.

<div dir="ltr" align="left">

```bash
uv run ruff check . && uv run ruff format --check .
uv run mypy src tests
```

</div>

**5. מלאי את המשימות 1 עד 3** בקובץ `report.py`:

- הוסיפי ל־`SummaryDict` את שלושת המפתחות החסרים, בטיפוסים של פלט ה־JSON בסעיף „פלט צפוי”.
- בני ב־`summary_dict` את המילון. אם תשכחי מפתח או תטעי בטיפוס, mypy יגיד לפני כל בדיקה.
- כתבי ב־`TextReporter.write` שורה לכל בעיה ואז שורת סיכום. אם אין אף אירוע תקין, בסוגריים כתוב `no valid events`.

**6. בדיקות היום, והבודקים:**

<div dir="ltr" align="left">

```bash
uv run pytest -q tests/test_report.py
uv run ruff check . && uv run ruff format --check . && uv run mypy src tests
```

</div>

**7. קומיט:**

<div dir="ltr" align="left">

```bash
cd ~/senior-course
git add ch07
git commit -m "Chapter 7: eventcheck project, reporters, ruff and mypy strict"
```

</div>

## יום 16: מודלים, בדיקת שורות ופרופיל

**8. מלאי את המשימות 4 עד 6** בקובץ `models.py`:

- `to_utc`: אותו רגע בזמן, באזור UTC.
- `invoice_prefix`: `ValueError` עם ההודעה שבהערה, אם המזהה לא מתחיל ב־`inv_`.
- `location_when_in_transit`: משלוח בדרך בלי מיקום הוא שגיאה; אחרת מחזירים את `self`.

<div dir="ltr" align="left">

```bash
cd ~/senior-course/ch07
uv run pytest -q tests/test_models.py
```

</div>

**9. מלאי את המשימות 7 עד 9** בקובץ `checker.py`. הוסיפי קודם את הייבוא שבהערה בראש הקובץ.

| TODO | מה עושים |
| --- | --- |
| 7 | בעיה אחת לכל שגיאה ב־`error.errors(include_url=False)`. הסוג הוא `json` רק לשגיאת JSON שבור. המיקום הוא הרכיבים של `loc` מחוברים בנקודות, בלי התג של האיחוד בהתחלה |
| 8 | הצלחה עם האירוע מ־`validate_json`, או כישלון עם הבעיות |
| 9 | מספור מ־1; שורות ריקות נספרות אבל לא נבדקות; מסירים את סוף השורה לפני הבדיקה; מזהה שחוזר הוא בעיה מסוג `duplicate` במיקום `id` |

**10. כל הבדיקות, ואז הכלי על קובץ הדוגמה:**

<div dir="ltr" align="left">

```bash
uv run pytest -q
uv run eventcheck check data/events.jsonl
echo "exit code: $?"
uv run eventcheck check data/events.jsonl --format json
```

</div>

**11. פרופיל אחד ושתי מדידות.** אלה הפקודות מסעיפים 7.8.1 ו־7.8.2:

<div dir="ltr" align="left">

```bash
mkdir -p /tmp/ch07
uv run eventcheck generate 200000 /tmp/ch07/big.jsonl --bad-every 1000
uv run python -m cProfile -o /tmp/ch07/check.prof -m eventcheck check /tmp/ch07/big.jsonl > /dev/null
uv run python -c "import pstats; pstats.Stats('/tmp/ch07/check.prof').sort_stats('cumulative').print_stats(12)" > profile.txt
head -1 /tmp/ch07/big.jsonl > /tmp/ch07/one.jsonl
uv run python -m timeit -s "from pathlib import Path; from eventcheck.models import EVENT_ADAPTER; line = Path('/tmp/ch07/one.jsonl').read_text()" "EVENT_ADAPTER.validate_json(line)"
uv run python -m timeit -s "import json; from pathlib import Path; from eventcheck.models import EVENT_ADAPTER; line = Path('/tmp/ch07/one.jsonl').read_text()" "EVENT_ADAPTER.validate_python(json.loads(line))"
```

</div>

הפלט של הפרופיל נשמר בקובץ `profile.txt`. בפתק שלך רשמי את חמש השורות המובילות אחרי מנגנון ההרצה,
את שתי תוצאות ההשוואה, ואת התאריך ותיאור המחשב. אל תעתיקי מספרים ממקום אחר: רק מה שמדדת.

**12. קומיט:**

<div dir="ltr" align="left">

```bash
cd ~/senior-course
git add ch07
git commit -m "Chapter 7: pydantic models, line checks, tests and a measured profile"
```

</div>

## פלט צפוי

הבדיקות, בסוף יום 16 (הזמן אצלך יהיה אחר):

<div dir="ltr" align="left">

```text
37 passed
```

</div>

הכלי על קובץ הדוגמה, עם קוד יציאה 1:

<div dir="ltr" align="left">

```text
line 4: schema: Input tag 'invoice.refunded' found using 'type' does not match any of the expected tags: 'invoice.paid', 'shipment.moved'
line 8: duplicate at id: id 1a61dbe2-2e44-458b-ae97-ba94d0eda82f first seen on line 7
line 12: schema at tenant: Field required
line 16: schema at data.amount_cents: Input should be greater than 0
line 20: json: Invalid JSON: EOF while parsing a value at line 1 column 205
line 24: schema: Input tag 'invoice.refunded' found using 'type' does not match any of the expected tags: 'invoice.paid', 'shipment.moved'
checked 25 lines: 19 valid, 6 bad (invoice.paid=9, shipment.moved=10)
exit code: 1
```

</div>

ובפורמט JSON, ההתחלה:

<div dir="ltr" align="left">

```text
{
  "lines": 25,
  "valid": 19,
  "bad_lines": [
    4,
    8,
    12,
    16,
    20,
    24
  ],
  "by_type": {
```

</div>

## רשימת בדיקה

- [ ] הפקודה `uv sync --locked` עוברת בעותק נקי.
- [ ] הבודקים הסטטיים ובדיקת העיצוב עוברים על הקוד ועל הבדיקות.
- [ ] כל 37 הבדיקות עוברות, כולל זו המסומנת `slow`.
- [ ] הכלי מדווח על שש השורות הפגומות בקובץ הדוגמה ומחזיר 1; על קובץ תקין הוא מחזיר 0.
- [ ] הקובץ `profile.txt` נשמר, והפתק כולל את חמש השורות המובילות ואת שתי מדידות ההשוואה.
- [ ] שני קומיטים, אחד לכל יום.

## טעויות נפוצות ותיקונים

- **להריץ את `uv sync` בלי `--locked` אחרי שינוי ידני בקובץ הפרויקט.** הנעילה מתעדכנת בשקט, והעותק שלך
  כבר לא זהה לשל ההתחלה. אם שינית תלות בכוונה, זה בסדר; אם לא, החזירי את הקובץ והריצי עם `--locked`.
- **להשאיר את סוף השורה בקלט של `validate_json`.** בשורה 20 ההודעה מדברת על `line 2 column 0`, שורה שלא
  קיימת בקובץ. בשורות שבורות אחרות מופיעה הודעה על תו בקרה. מסירים את סוף השורה לפני הבדיקה.
- **לדווח מיקום כמו `invoice.paid.data.amount_cents`.** הבדיקה על המיקום נכשלת. התג של האיחוד המובחן
  הוא הרכיב הראשון של `loc`; מסירים אותו כשהוא אחד מ־`EVENT_TYPES`.
- **לזרוק `TypeError` מתוך בודק תקינות.** הבדיקה מקבלת חריגה רגילה במקום שגיאת תקינות. הספרייה
  ממירה רק `ValueError` ו־`AssertionError`, ולכן בודק מעלה תמיד אחת מהן.
- **לבדוק את קוד היציאה בהרצה תחת `cProfile`.** שם יוצא 0 גם כשהכלי מחזיר 1. בודקים בהרצה רגילה, כמו
  בצעד 10.
- **לסמן בדיקה `@pytest.mark.slwo`.** עם `--strict-markers` הריצה נעצרת על סימון לא רשום. זו ההגנה
  שעובדת; תקני את השם או רשמי את הסימון ב־`[tool.pytest]`.

## רמזים

- ב־`TODO 9`, מילון ממזהה למספר השורה שבה הופיע לראשונה נותן גם את הבדיקה וגם את ההודעה.
- בפירוק עם `match`, ענף עם תנאי (`case Ok(event) if ...`) לפני הענף הכללי חוסך תנאי מקונן.
- כדי לראות מה הספרייה מחזירה, הריצי בדיקה אחת עם הדגל `--pdb`, והדפיסי את `error.errors()`.
- הפורמט של שורת הסיכום מופיע במלואו ב־`tests/test_report.py`; קראי את הבדיקה לפני שאת כותבת.

## הפתרון

`solution/`. **למה הוא עובד כך:** הבדיקה כולה היא פונקציות טהורות על מחרוזות, ולכן הבדיקות לא צריכות
קבצים, ושורת הפקודה היא שכבה דקה שפותחת קובץ ובוחרת דוח. הקריאה `validate_json` מחזירה JSON שבור
ושגיאת מודל באותו ערוץ, כך שיש רק טיפול אחד בחריגה. הדוחות מקיימים ממשק מבני בלי ירושה, וצורת ה־JSON
מוגדרת במילון עם צורה, כך שהבודק מוודא כל מפתח. הקובץ `py.typed` מבטיח שקוד אחר יקבל את הטיפוסים.

## מה התרגיל מדגים

- פרויקט uv משוחזר: קובץ פרויקט, נעילה, קבוצת תלויות לפיתוח ונקודת כניסה.
- בדיקה סטטית קפדנית שתופסת טעויות במבנה הפלט עוד לפני הבדיקות.
- בדיקת תקינות בגבול עם pydantic: אילוצים, קפדנות לפי שדה, איחוד מובחן ושגיאות עם מיקום.
- בדיקות של צוות: fixtures בטווחים שונים, פרמטרים, תיקיות זמניות, החלפות זמניות וסימון רשום.
- מדידה לפני שיפור, עם מספרים שנרשמים ולא מנוחשים.

</div>
