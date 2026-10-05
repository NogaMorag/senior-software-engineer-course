---
direction: rtl
lang: he
title: "פרק 7: Python עם טיפוסים, אריזה ובדיקות"
chapter: 7
part: "חלק ב: Python בכירה"
---

<div dir="rtl" lang="he">

# פרק 7: Python עם טיפוסים, אריזה ובדיקות

את כבר כותבת Python. מה שמבדיל קוד של בכירה הוא היגיינה: פרויקט שנבנה מקובץ נעילה, בדיקה סטטית
קפדנית, טיפוסים שמתעדים חוזים, בדיקת תקינות בגבולות, בדיקות שרצות בשנייה, ומדידה לפני כל שיפור.
אלה הדברים שסוקרת קוד רואה ראשונים. בפרק תבני כלי שורת פקודה קטן ושלם, `eventcheck`, שבודק קובץ
של אירועי webhook ומדווח על כל שורה פגומה.

## מה נלמד בפרק

- פרויקט uv: מה יש ב־`pyproject.toml`, מה מבטיח קובץ הנעילה, וקבוצות תלויות.
- בודק סגנון ובודק טיפוסים במצב קפדני, כחלק מהיום־יום.
- טיפוסים מתקדמים: ממשקים מבניים, מילונים עם צורה ידועה, generics בתחביר החדש, ועוד כמה כלים קטנים.
- pydantic v2: מודלים, בודקי תקינות, איחוד מובחן ומתאם טיפוסים.
- pytest כמו בצוות: fixtures והטווחים שלהן, פרמטרים, תיקיות זמניות, החלפות זמניות וסימונים.
- מבנה חבילה (src layout), נקודות כניסה, והקובץ שמפרסם את הטיפוסים שלך.
- מדידה עם הכלים של הספרייה התקנית, ומספרים שאת מודדת ורושמת בעצמך.

## ידע מקדים

- פרק 2, סעיף 2.8: uv ו־Python 3.14 מותקנים ועובדים.
- פרק 3, סעיף 3.8, ופרק 4, סעיף 4.2: בדיקות טבלאיות וממשקים ב־Go, שאליהם נשווה את הכלים של Python.
- פרק 6, סעיף 6.8: פרופיל ראשון עם pprof.
- Python ברמה של יום עבודה: מחלקות, מחלקות נתונים, מחוללים (generators), חריגות ומנהלי הקשר.

## 7.1 מה „Python בכירה” אומרת

כשמודעה מבקשת „Python ברמה גבוהה”, היא לא מתכוונת לטריקים. היא מתכוונת לכך שאפשר לתת לך מאגר,
ושבעוד חצי שנה הוא עדיין יהיה תקין. בפועל זה שישה הרגלים:

| הרגל | מה רואים במאגר |
| --- | --- |
| בנייה משוחזרת | קובץ הפרויקט וקובץ הנעילה מחויבים ב־Git |
| ניתוח סטטי | ruff ו־mypy עוברים בלי השתקות |
| חוזים מפורשים | רמזי טיפוס (type hints) בכל גבול של מודול |
| בדיקת תקינות בגבול | קלט חיצוני עובר דרך מודל pydantic |
| בדיקות מהירות | בלי רשת ובלי שעון אמיתי |
| מדידה | פרופיל לפני אופטימיזציה |

הרעיון שמחבר את כולם: בתוך פונקציה Python דינמית ונוחה, אבל בגבול (קובץ, בקשת HTTP, הודעה מתור)
היא חייבת להיות קפדנית, כי שם נכנס קלט שאת לא שולטת בו. התרגיל הוא גבול כזה. קובץ JSONL (שורת JSON
אחת לכל אירוע) מגיע מבחוץ, והכלי צריך להגיד על כל שורה אם היא תקינה, ואם לא, למה. בפרק 8 תבני את
הצד השני: שירות שמקבל את האירועים האלה ב־HTTP.

## 7.2 פרויקט עם uv: pyproject, lock ו־scripts

הפקודה `uv init --package` יוצרת שלד של חבילה: מבנה `src/`, מערכת בנייה, פקודה וקובץ שנועל את גרסת
המפרש (בתרגיל תריצי אותה פעם אחת בתיקייה זמנית, כדי לראות). שימי לב שהשדה `authors` מתמלא מהגדרות
ה־Git שלך; בפרויקט פתוח מחליטים במודע מה נכנס לשם.

### 7.2.1 מה יש ב־pyproject.toml

זה קובץ הפרויקט של התרגיל (`exercises/ch07-eventcheck/solution/pyproject.toml`), בלי הגדרות הכלים:

<div dir="ltr" align="left">

```toml
[project]
name = "eventcheck"
version = "0.1.0"
description = "Validate a JSONL file of webhook events and report every bad line."
requires-python = ">=3.14"
dependencies = [
    "pydantic==2.13.5",
]

[project.scripts]
eventcheck = "eventcheck.cli:main"

[dependency-groups]
dev = [
    "mypy==2.4.0",
    "pytest==9.1.1",
    "ruff==0.16.10",
]

[build-system]
requires = ["uv_build>=0.12.23,<0.13.0"]
build-backend = "uv_build"
```

</div>

ארבע טבלאות. הראשונה: מטא־נתונים ותלויות זמן ריצה (PEP 621). השנייה: פקודה שקוראת לפונקציה `main`.
השלישית: תלויות פיתוח שלא נשלחות עם החבילה (PEP 735). הרביעית: מערכת הבנייה.
הגרסאות מוצמדות במדויק, נכון לאוקטובר 2026; הרשימה המלאה בנספח ג (`book/appendices/c-tool-versions.md`). ספרייה שאחרים מתקינים מצמידה טווח, כדי לא
לכפות גרסה על המשתמשות בה.

### 7.2.2 מה קובץ הנעילה מבטיח

הקובץ `uv.lock` נוצר בפעם הראשונה שמסנכרנים את הפרויקט או מוסיפים לו תלות. הוא רושם את כל עץ
התלויות, כולל העקיפות, עם גרסה מדויקת ותמצית (hash) לכל קובץ. הוא אוניברסלי: אותו קובץ מתאר את
הפתרון לכל מערכת הפעלה וארכיטקטורה, ולא רק למחשב שיצר אותו.

<div dir="ltr" align="left">

```bash
uv sync --locked
uv sync --frozen
```

</div>

| פקודה | מה היא עושה | מתי |
| --- | --- | --- |
| `uv sync --locked` | נכשלת אם הנעילה לא תואמת לקובץ הפרויקט | ב־CI: מוכיחה שאיש לא שכח לעדכן |
| `uv sync --frozen` | מתקינה בדיוק מה שבנעילה, בלי לבדוק התאמה | בבניית image, אחרי שהנעילה נבדקה |

מה הוא **לא** מבטיח: שהגרסאות טובות, בטוחות, או עובדות עם הקוד שלך. הוא מבטיח רק שכולן מקבלות
אותן גרסאות. עדכון הוא פעולה מפורשת, בקומיט משלה:

<div dir="ltr" align="left">

```bash
uv lock --upgrade-package pydantic
uv sync
```

</div>

> **איך Rust עושה את זה:**
> - הקובץ `uv.lock` מקביל ל־`Cargo.lock`.
> - לדגל `--locked` יש אותה משמעות בשני הכלים.
> - קבוצת התלויות `dev` מקבילה ל־`[dev-dependencies]`.
> - ההבדל המהותי: ב־Rust המהדר בודק טיפוסים, וב־Python זה כלי נפרד שאת בוחרת להריץ.

### 7.2.3 uv run וקבוצות תלויות

הפקודה `uv run` מסנכרנת את הסביבה לפי הנעילה ואז מריצה בתוכה, בלי הפעלה ידנית של סביבה וירטואלית.
קבוצת `dev` מותקנת כברירת מחדל; בבניית image לייצור מוותרים עליה. תלות מוסיפים תמיד דרך uv, שמעדכן
את קובץ הפרויקט ואת הנעילה יחד. לא מתקינים עם pip, בדיוק כי הוא לא יודע לעדכן את הנעילה.

<div dir="ltr" align="left">

```bash
uv run eventcheck --help
uv sync --frozen --no-dev
uv add 'pydantic==2.13.5'
uv add --dev 'pytest==9.1.1'
```

</div>

## 7.3 ruff ו־mypy

הכלי ruff משלב בודק סגנון (linter) ומעצב קוד. הוא כתוב ב־Rust, ומהיר מספיק כדי לרוץ בכל שמירה.
הכלי mypy בודק טיפוסים: הוא קורא את רמזי הטיפוס ומוכיח שהקוד עקבי איתם, בלי להריץ אותו. ההגדרות
של שניהם יושבות בקובץ הפרויקט:

<div dir="ltr" align="left">

```toml
[tool.ruff]
line-length = 100
target-version = "py314"

[tool.ruff.lint]
select = ["E", "W", "F", "I", "B", "UP", "SIM", "RUF", "PT", "N"]

[tool.mypy]
python_version = "3.14"
strict = true
plugins = ["pydantic.mypy"]
enable_error_code = ["explicit-override", "ignore-without-code", "redundant-expr", "truthy-bool"]
```

</div>

כל אות ברשימת `select` היא קבוצת כללים: ייבוא שלא בשימוש, סדר ייבוא, באגים נפוצים כמו ברירת מחדל
שהיא רשימה, תחביר ישן, וסגנון בדיקות. הרשימה היא החלטה של צוות, ומרחיבים אותה לאט. בדיקה, ואחריה תיקון:

<div dir="ltr" align="left">

```bash
uv run ruff check . && uv run ruff format --check .
uv run ruff check --fix . && uv run ruff format .
uv run mypy src tests
```

</div>

הדגל `strict` מפעיל קבוצת בדיקות: אין פונקציה בלי טיפוסים, אין ערך בלי טיפוס שמוחזר בשקט, אין
השתקה מיותרת. בפרויקט חדש מתחילים קפדניים, כי קפדנות שמוסיפים אחר כך פירושה מאות שגיאות בבת אחת.
ארבעת הקודים הנוספים אינם חלק מ־`strict`:

- **הקוד `explicit-override`:** מתודה שדורסת מתודה של מחלקת אב חייבת לשאת `@override`.
- **הקוד `ignore-without-code`:** אסור `# type: ignore` בלי לציין איזו שגיאה משתיקים.
- **הקוד `redundant-expr`:** תנאי שלפי הטיפוסים תמיד אמת או תמיד שקר.
- **הקוד `truthy-bool`:** בדיקת אמת על אובייקט שתמיד אמת, למשל פונקציה ששכחו לקרוא לה.

התוסף `pydantic.mypy` מלמד את הבודק איך נבנה הבנאי של כל מודל, כך שארגומנט חסר נתפס כבר סטטית. כך
נראית שגיאה אמיתית. בהרצה שבה נכתב הספר הוסר המעטר `@override` מהמתודה `__str__` בתרגיל, והבודק אמר:

<div dir="ltr" align="left">

```text
src/eventcheck/checker.py:36: error: Method "__str__" is not using @override but is overriding a method in class "builtins.object"  [explicit-override]
Found 1 error in 1 file (checked 7 source files)
```

</div>

## 7.4 טיפוסים מתקדמים: Protocol, TypedDict ו־generics

הדוגמאות כאן נבדקו מול התיעוד של Python 3.14. כדאי לדעת מאיזו גרסה כל כלי קיים:

| כלי | מגרסה | מה הוא נותן |
| --- | --- | --- |
| `Protocol`, `TypedDict` | 3.8 | ממשק מבני; מילון עם מפתחות ידועים |
| `typing.Self` | 3.11 | „המחלקה הנוכחית”, גם בתת־מחלקה |
| `class Box[T]` ו־`type X = ...` | 3.12 | תחביר generics חדש (PEP 695) |
| `typing.override` | 3.12 | מסמן דריסה; הבודק מוודא שיש מה לדרוס |
| הערות טיפוס עצלות | 3.14 | בלי `from __future__ import annotations` |

השורה האחרונה חשובה כשקוראים קוד. מגרסה 3.14 הערות הטיפוס מחושבות רק כשמבקשים אותן (PEP 649), ולכן
אפשר להפנות למחלקה שמוגדרת בהמשך הקובץ בלי מרכאות. הייבוא מ־`__future__` בקוד ישן הוא שריד.

### 7.4.1 Protocol: ממשק בלי ירושה

הטיפוס `Protocol` מגדיר צורה ולא שושלת: כל אובייקט עם המתודות הנכונות מקיים אותו, גם אם לא שמע עליו. זה
הממשק של Go מפרק 4. בקובץ `report.py` של התרגיל יש שניים:

<div dir="ltr" align="left">

```python
class Writer(Protocol):
    """Anything with write(str): sys.stdout, io.StringIO, an open text file."""

    def write(self, text: str, /) -> object: ...


class Reporter(Protocol):
    def write(self, report: Report, out: Writer) -> None: ...
```

</div>

הראשון מראה למה Protocol קיים. הפלט הסטנדרטי, מחרוזת בזיכרון וקובץ פתוח לא יורשים ממחלקה שלנו,
ואין דרך לגרום להם לרשת; הם פשוט מקיימים את הצורה. מחלקת בסיס מופשטת (ABC) לא הייתה עובדת כאן.
מחלקה מופשטת עדיפה כשרוצים לשתף מימוש, או שגיאה בזמן ריצה כשמחלקה לא מימשה מתודה. Protocol עדיף כשהחוזה שייך
לצרכנית: הפונקציה שכותבת דוח אומרת מה היא צריכה, וכל מי שמתאים נכנס. זה „ממשק קטן אצל הצרכן” של Go.

> **טעות נפוצה:** להניח ש־`Protocol` נבדק בזמן ריצה. הוא לא. בדיקת `isinstance` עליו נכשלת, אלא אם
> מסמנים אותו `@runtime_checkable`, וגם אז נבדק רק שהמתודות קיימות. האכיפה כולה של הבודק הסטטי.

### 7.4.2 TypedDict: צורה של JSON

כשהפלט הוא JSON, `TypedDict` מתאר את המבנה בלי מחלקה בזמן ריצה: זה מילון רגיל, והבודק מוודא כל
מפתח וכל ערך. מתוך `report.py`:

<div dir="ltr" align="left">

```python
class SummaryDict(TypedDict):
    lines: int
    valid: int
    bad_lines: list[int]
    by_type: dict[str, int]
    issues: list[IssueDict]
```

</div>

אם תכתבי בטעות `"bad_line"` במקום `"bad_lines"`, הבודק יעצור אותך. בלעדיו, הטעות הייתה מתגלה רק
כשכלי אחר קורא את ה־JSON. לנתונים פנימיים מחלקת נתונים עדיפה, כי יש לה מתודות ואי־שינוי.

### 7.4.3 generics בתחביר החדש, ועוד שלושה כלים קטנים

התחביר החדש מצהיר על משתנה הטיפוס ליד השם, בלי `TypeVar`. בקובץ `checker.py`, ניתוח שורה מחזיר
הצלחה עם ערך, או כישלון עם רשימת בעיות:

<div dir="ltr" align="left">

```python
@dataclass(frozen=True, slots=True)
class Ok[T]:
    value: T


@dataclass(frozen=True, slots=True)
class Err:
    issues: list[Issue]


type Outcome[T] = Ok[T] | Err
```

</div>

ההוראה `type` יוצרת כינוי טיפוס (alias) שמחושב בעצלות. הפונקציה `check_lines` מפרקת את התוצאה
עם `match`. בענף `case Ok(event)` הבודק כבר יודע שיש בידיו אירוע אמיתי.

ושלושה כלים קטנים, באותו קובץ:

- **הטיפוס `Literal`:** סוג הבעיה, `IssueKind`, הוא אחת משלוש מחרוזות. ערך כמו `"jsno"` הוא שגיאת טיפוס.
- **הטיפוס `Self`:** המתודה `from_validation_error` מחזירה רשימה של המחלקה הנוכחית, כך שבתת־מחלקה תחזור תת־המחלקה.
- **המעטר `@override`:** על `__str__`, מצהיר שזו דריסה. אם מחלקת האב תשנה שם, הבודק יגיד.

> **איך Rust עושה את זה:**
> - הזוג שלנו הוא בעצם `Result<T, Vec<Issue>>` בלי הסוכר, והפירוק עם `match` זהה.
> - ההבדל: ב־Rust המהדר דורש לטפל בכל הענפים.
> - ב־Python הבודק מתריע על ענף חסר רק אם מוסיפים `assert_never` בסוף.
> - ובזמן ריצה, פירוק בלי ענף מתאים פשוט לא עושה כלום.

## 7.5 pydantic v2

רמזי טיפוס הם הבטחה של המתכנתת, ו־Python לא אוכפת אותם. קלט חיצוני לא חתם על ההבטחה. הספרייה
pydantic סוגרת את הפער: מודל שמוגדר ברמזי טיפוס, ובנייה שבודקת וממירה את הקלט, או זורקת `ValidationError`
עם כל השגיאות. הליבה שלה כתובה ב־Rust, בחבילה נפרדת, ולכן היא מהירה מספיק
לשמש בכל בקשת HTTP.

### 7.5.1 מודלים, אילוצים ובודקי תקינות

מתוך `models.py` של התרגיל:

<div dir="ltr" align="left">

```python
class Strict(BaseModel):
    """Shared configuration: unknown fields are errors, instances are immutable."""

    model_config = ConfigDict(extra="forbid", frozen=True)


class EventBase(Strict):
    id: UUID
    tenant: TenantId
    created_at: AwareDatetime

    @field_validator("created_at")
    @classmethod
    def to_utc(cls, value: datetime) -> datetime:
        return value.astimezone(UTC)


class InvoicePaidData(Strict):
    invoice_id: str = Field(min_length=5)
    amount_cents: int = Field(gt=0, strict=True)
    currency: Currency
```

</div>

| הגדרה | מה היא אוכפת |
| --- | --- |
| `extra="forbid"` | שדה לא מוכר הוא שגיאה, ולא נבלע בשקט |
| `frozen=True` | אי אפשר לשנות אחרי הבנייה |
| `UUID`, `AwareDatetime` | מחרוזת שאינה מזהה תקין, או תאריך בלי אזור זמן, נדחים |
| `Field(gt=0, strict=True)` | מספר שלם חיובי; המחרוזת `"1999"` **לא** מתקבלת |
| `@field_validator` | קוד שלך שרץ אחרי הבדיקה המובנית של השדה |

ברירת המחדל היא מצב מקל (lax), שממיר `"1999"` ל־1999. בשדה של כסף זה מסוכן, כי זה מסתיר שולחת שבנתה
JSON לא נכון. ההחלטה היא לכל שדה, לא לכל הפרויקט. כשהבדיקה תלויה בכמה שדות, משתמשים בבודק ברמת המודל:

<div dir="ltr" align="left">

```python
    @model_validator(mode="after")
    def location_when_in_transit(self) -> Self:
        if self.status == "in_transit" and not self.location:
            raise ValueError("location is required when status is 'in_transit'")
        return self
```

</div>

בודק תקינות מעלה `ValueError` או `AssertionError`, והספרייה הופכת אותו לשגיאה עם מיקום. חריגה מסוג
אחר עוברת הלאה כמו שהיא, וזה טוב: באג בבודק לא מתחפש לקלט שגוי.

### 7.5.2 איחוד מובחן ו־TypeAdapter

השדה `type` של כל אירוע קובע איזה מודל חל עליו. זה איחוד מובחן (discriminated union):

<div dir="ltr" align="left">

```python
type Event = Annotated[InvoicePaid | ShipmentMoved, Field(discriminator="type")]

# The discriminator values; pydantic puts the matching one first in an error's location.
EVENT_TYPES = frozenset({"invoice.paid", "shipment.moved"})

EVENT_ADAPTER: TypeAdapter[Event] = TypeAdapter(Event)
```

</div>

בלי המבחין, הספרייה מנסה כל מודל בתורו ומחזירה שגיאות מכולם, איטי ומבלבל. איתו, היא קוראת את `type`,
בוחרת מודל אחד ומדווחת רק את השגיאות שלו. מתאם הטיפוסים (`TypeAdapter`) נותן את כל יכולות הבדיקה
לטיפוס שאינו מודל, ונוצר **פעם אחת** ברמת המודול, כי הבנייה שלו מקמפלת סכמה. יש לו שתי דרכים לקרוא:

| קריאה | מה היא עושה |
| --- | --- |
| `EVENT_ADAPTER.validate_json(line)` | מפרקת את ה־JSON ובודקת, הכול בתוך הליבה |
| `EVENT_ADAPTER.validate_python(json.loads(line))` | פירוק ב־Python, ואז בדיקה של המבנה |

הפתרון משתמש בראשונה, ובסעיף 7.8 תמדדי את ההבדל. יתרון נוסף: JSON שבור מגיע כשגיאת תקינות רגילה,
מסוג `json_invalid`, באותו ערוץ כמו שאר השגיאות.

### 7.5.3 מה pydantic בודקת שרמזי טיפוס לא

זו שאלת ראיון קלאסית, והתשובה היא „את הקלט, בזמן ריצה”: טיפוסים בפועל, אילוצים שאין להם שפה
ברמזי טיפוס, המרות מבוקרות, צורה, וכל השגיאות בבת אחת, כל אחת עם מיקום. המיקום הוא מה שהופך את הדוח
של התרגיל לשימושי. השורה הזאת אומרת בדיוק מה לתקן ואיפה:

<div dir="ltr" align="left">

```text
line 16: schema at data.amount_cents: Input should be greater than 0
```

</div>

## 7.6 pytest: fixtures, parametrize ו־monkeypatch

שלושה דברים מבדילים חבילת בדיקות של צוות: מצב מוכן דרך fixtures ולא דרך העתקה, מקרים דרך פרמטרים
ולא דרך לולאות, ובידוד מהסביבה.

### 7.6.1 fixtures והטווח שלהן

פונקציית הכנה (fixture) מתבקשת לפי שם הפרמטר. המריצה מוצאת אותה בקובץ `conftest.py`. מתוך הקובץ
הזה בתרגיל:

<div dir="ltr" align="left">

```python
@pytest.fixture(scope="session")
def big_file(tmp_path_factory: pytest.TempPathFactory) -> Path:
    """Built once per test session and shared: 50,000 lines, one bad line in every 1,000."""
    path = tmp_path_factory.mktemp("big") / "big.jsonl"
    with path.open("w", encoding="utf-8") as handle:
        for line in generate_lines(50_000, seed=11, bad_every=1_000):
            handle.write(line + "\n")
    return path
```

</div>

| טווח | כמה פעמים נבנית | מתי |
| --- | --- | --- |
| `function` (ברירת מחדל) | לכל בדיקה | כמעט תמיד: כל בדיקה מקבלת מצב נקי |
| `class`, `module`, `package` | פעם לכל קבוצה | משאב יקר שרק קבוצה צריכה |
| `session` | פעם לכל הריצה | קובץ גדול, קונטיינר, שרת |

הכלל: טווח רחב רק למשאב שאיש לא משנה. אם בדיקה אחת משנה משאב משותף, הבאה תלויה בסדר, ותיכשל רק
כשמריצים אותה עם השכנות שלה. ה־fixture המובנית `tmp_path` נותנת תיקייה ריקה וייחודית לכל בדיקה.
בתרגיל, `write_jsonl` עוטפת אותה כ„מפעל” (factory): היא מחזירה פונקציה, הבדיקה מחליטה מה לכתוב,
והיא מחליטה איפה.

### 7.6.2 parametrize

לולאה בתוך בדיקה עוצרת בכישלון הראשון ולא אומרת איזה מקרה נכשל. במקומה נותנים ל־pytest טבלה. מתוך
`tests/test_models.py`, בקיצור:

<div dir="ltr" align="left">

```python
@pytest.mark.parametrize(
    ("path", "value", "location"),
    [
        (("data", "amount_cents"), 0, ("invoice.paid", "data", "amount_cents")),
        (("data", "amount_cents"), "1999", ("invoice.paid", "data", "amount_cents")),
        (("data", "currency"), "GBP", ("invoice.paid", "data", "currency")),
```

</div>

כל שורה היא בדיקה נפרדת עם שם משלה (הפרמטר `ids` נותן שמות כמו `amount-as-string`). כשאחת נכשלת,
הפלט אומר איזו, ואפשר להריץ רק אותה. זו הבדיקה הטבלאית של Go, בלי לכתוב את הלולאה.

### 7.6.3 monkeypatch, capsys וסימונים

ה־fixture `monkeypatch` משנה משהו לזמן הבדיקה ומחזיר אותו אחריה, גם אם הבדיקה נכשלה. מתוך `tests/test_cli.py`:

<div dir="ltr" align="left">

```python
def test_dash_reads_stdin(
    monkeypatch: pytest.MonkeyPatch, invoice: dict[str, Any], capsys: pytest.CaptureFixture[str]
) -> None:
    monkeypatch.setattr(sys, "stdin", io.StringIO(json.dumps(invoice) + "\n"))
    assert main(["check", "-"]) == EXIT_OK
    assert "invoice.paid=1" in capsys.readouterr().out
```

</div>

בדיקה אחרת בקובץ משתמשת ב־`monkeypatch.setenv` כדי לבחור פורמט פלט, ו־`capsys` לוכדת את מה
שהודפס. הבדיקות קוראות ל־`main` עם רשימת ארגומנטים ומקבלות קוד יציאה כמספר, בלי להריץ תהליך. זה
אפשרי רק כי `main` מחזיר מספר ולא קורא ל־`sys.exit` בעצמו.

סימון (mark) מתייג בדיקה. הבדיקה על הקובץ הגדול מסומנת `slow`, והסימון רשום בקובץ הפרויקט, כדי
שהדגל `--strict-markers` יתפוס שגיאת כתיב:

<div dir="ltr" align="left">

```toml
[tool.pytest]
testpaths = ["tests"]
addopts = ["--strict-markers", "--strict-config", "-ra"]
markers = ["slow: generates a large input file; deselect with -m 'not slow'"]
```

</div>

הטבלה `[tool.pytest]` עם טיפוסי TOML אמיתיים נתמכת מגרסה 9. כדי לדלג על הבדיקות האיטיות:

<div dir="ltr" align="left">

```bash
uv run pytest -q -m "not slow"
```

</div>

## 7.7 מבנה חבילה ו־entry points

המבנה של התרגיל הוא src layout, כלומר הקוד יושב בתיקיית מקור נפרדת:

<div dir="ltr" align="left">

```text
ch07/
├── pyproject.toml   uv.lock   .python-version
├── src/eventcheck/
│   ├── __init__.py  __main__.py  py.typed
│   └── models.py  checker.py  report.py  generate.py  cli.py
└── tests/
    └── conftest.py  test_models.py  test_checker.py  test_cli.py
```

</div>

למה `src/`? כי בלעדיה Python מוצאת את החבילה בתיקייה הנוכחית ולא את מה שהותקן, ובדיקות עוברות אצלך
ונכשלות אצל מי שמתקינה, למשל כשקובץ חסר בבנייה. עם `src/`, הדרך היחידה לייבא את החבילה היא דרך
ההתקנה. הפקודה `uv sync` מתקינה אותה במצב עריכה (editable), כך ששינוי בקוד נראה מיד.

### 7.7.1 נקודת כניסה וקודי יציאה

הפקודה מנקודת הכניסה, ההרצה עם `python -m`, והקריאה הישירה מבדיקה מגיעות כולן לאותה פונקציה. השנייה
עוברת דרך הקובץ `__main__.py` של התרגיל:

<div dir="ltr" align="left">

```python
from eventcheck.cli import main

raise SystemExit(main())
```

</div>

גם המעטפת שהמתקין יוצר לנקודת הכניסה קוראת ל־`main()` ומעבירה את הערך הלאה כקוד יציאה. קודי היציאה
הם חלק מהחוזה: 0 תקין, 1 נמצאו שורות פגומות, 2 שגיאת שימוש. צעד ב־CI יכול לסמוך עליהם.

### 7.7.2 הקובץ py.typed

הקובץ הריק `py.typed` מסמן לבודקי טיפוסים שהחבילה מתעדת טיפוסים (PEP 561). בלעדיו, קוד חיצוני רואה
הכול כערך בלי טיפוס. זה נבדק בהרצה שבה נכתב הספר: לפני שהקובץ נוסף, הבודק דילג על החבילה עם שגיאת
`import-untyped`. אחרי שנוסף, אותה בדיקה הראתה `InvoicePaid | ShipmentMoved`.

## 7.8 ביצועים: מודדים לפני שמשפרים

בפרק 6 ראית את pprof. ב־Python יש שני כלים בספרייה התקנית, ולכל אחד שאלה אחרת:

| כלי | השאלה |
| --- | --- |
| `cProfile` | איפה התוכנית מבלה את הזמן, לפי פונקציה |
| `timeit` | כמה זמן לוקח קטע קוד קטן, בחזרות רבות |

הסדר חשוב: קודם פרופיל על עבודה אמיתית, כדי לדעת **מה** איטי, ורק אז השוואה על החשוד.

### 7.8.1 פרופיל עם cProfile

התרגיל כולל מחולל שמשתמש בזרע אקראי (seed) קבוע, ולכן מייצר בדיוק אותו קובץ בכל ריצה:

<div dir="ltr" align="left">

```bash
mkdir -p /tmp/ch07
uv run eventcheck generate 200000 /tmp/ch07/big.jsonl --bad-every 1000
uv run python -m cProfile -o /tmp/ch07/check.prof -m eventcheck check /tmp/ch07/big.jsonl > /dev/null
uv run python -c "import pstats; pstats.Stats('/tmp/ch07/check.prof').sort_stats('cumulative').print_stats(12)"
```

</div>

הפקודות יוצרות מאתיים אלף שורות, מריצות את החבילה תחת פרופיל ושומרות את התוצאה, ומדפיסות 12 שורות
מובילות לפי זמן מצטבר. בפלט יש שלוש עמודות:

- **העמודה `ncalls`:** כמה פעמים הפונקציה נקראה.
- **העמודה `tottime`:** הזמן בתוך הפונקציה עצמה.
- **העמודה `cumtime`:** הזמן כולל כל מה שהיא קראה לו.

בהרצה שבה נכתב הספר, השורות המובילות אחרי מנגנון ההרצה היו הפונקציות של הכלי עצמו, מ־`main` ועד
`validate_json`. הזמנים שלך יהיו אחרים, ולכן הם לא מופיעים כאן: תמדדי ותרשמי בפתק את חמש השורות
המובילות שלך. ואז שאלי: כמה מהזמן הוא הבדיקה של pydantic, וכמה הלולאה שלך? אם רוב הזמן בספרייה, שכתוב
הלולאה לא ישנה הרבה.

> **שימי לב:** תחת `cProfile` קוד היציאה אובד. על הקובץ הזה הכלי מחזיר 1, אבל הפקודה עם הפרופיל
> מחזירה 0. אל תבדקי קודי יציאה בהרצה עם פרופיל.

### 7.8.2 השוואה עם timeit

השאלה מסעיף 7.5.2: האם פירוק בתוך הליבה מהיר מפירוק ב־Python ובדיקה אחריו? על אותה שורה בדיוק:

<div dir="ltr" align="left">

```bash
head -1 /tmp/ch07/big.jsonl > /tmp/ch07/one.jsonl
uv run python -m timeit -s "from pathlib import Path; from eventcheck.models import EVENT_ADAPTER; line = Path('/tmp/ch07/one.jsonl').read_text()" "EVENT_ADAPTER.validate_json(line)"
uv run python -m timeit -s "import json; from pathlib import Path; from eventcheck.models import EVENT_ADAPTER; line = Path('/tmp/ch07/one.jsonl').read_text()" "EVENT_ADAPTER.validate_python(json.loads(line))"
```

</div>

הדגל `-s` הוא הכנה שלא נמדדת; הכלי בוחר מספר חזרות ומדווח את הטובה מחמש סדרות. רשמי בפתק:

<div dir="ltr" align="left">

```text
validate_json:                    ____ usec per loop
validate_python(json.loads(...)): ____ usec per loop
machine / date:                   ____
```

</div>

שני כללים למדידה הוגנת: נתונים מייצגים (שורה אמיתית, לא `{}`), והשוואה רק בין מדידות מאותה מכונה
באותו מצב. מחשב נייד על סוללה, או עם תהליכים כבדים ברקע, נותן מספרים שאי אפשר להשוות.

## סיכום

- פרויקט uv הוא קובץ פרויקט וקובץ נעילה, שניהם מחויבים. ב־CI, הדגל `--locked` מוכיח שהם תואמים.
- הבודקים הסטטיים רצים כל הזמן, במצב קפדני, וההגדרות שלהם בקובץ הפרויקט.
- הטיפוס `Protocol` מתאים כשהחוזה שייך לצרכנית; מחלקה מופשטת כשמשתפים מימוש.
- הטיפוס `TypedDict` מתאר צורות שעוברות גבול.
- התחביר החדש של generics והמעטר `override` הופכים כוונה לבדיקה.
- המודלים של pydantic בודקים קלט בזמן ריצה, עם אילוצים, איחוד מובחן ושגיאות עם מיקום.
- בבדיקות: fixtures בטווח מתאים, פרמטרים במקום לולאות, החלפות זמניות לבידוד, וסימונים רשומים.
- מבנה src, נקודת כניסה, וקובץ `py.typed` כדי שהטיפוסים יגיעו למשתמשות.
- מודדים קודם, משווים אחר כך, ורושמים רק מספרים שמדדת בעצמך.

## מונחים לזכור

| מונח בעברית | באנגלית | במשפט אחד |
| --- | --- | --- |
| קובץ נעילה | lock file | כל התלויות בגרסה ובתמצית מדויקות, כדי שכולם יתקינו אותו דבר |
| קבוצת תלויות | dependency group | תלויות פיתוח שלא נשלחות עם החבילה |
| נקודת כניסה | entry point | פקודה שהמתקין יוצר ושקוראת לפונקציה בחבילה |
| מבנה src | src layout | הקוד בתוך `src/`, כך שאפשר לייבא אותו רק דרך ההתקנה |
| רמזי טיפוס | type hints | הערות טיפוס שבודק סטטי קורא ושהמפרש לא אוכף |
| טיפוסיות מבנית | structural typing | התאמה לפי צורה ולא לפי ירושה |
| בדיקת תקינות | validation | בדיקה בזמן ריצה שקלט מתאים למודל |
| איחוד מובחן | discriminated union | איחוד שבו שדה אחד קובע איזה מודל חל |
| מדידת פרופיל | profiling | מדידה של היכן תוכנית מבלה את זמנה |
| בנצ'מרק | benchmark | מדידה חוזרת של קטע קוד קטן כדי להשוות חלופות |

## טעויות נפוצות

- **לא לחייב את קובץ הנעילה.** כל מכונה פותרת תלויות מחדש, ו„אצלי זה עובד” חוזר.
- **להתקין עם pip לתוך סביבת uv.** הנעילה לא מתעדכנת, וה־CI מקבל סביבה אחרת.
- **להשתיק את הבודק בלי קוד שגיאה.** ההשתקה בולעת גם את השגיאה הבאה. תמיד עם קוד, ועדיף לתקן.
- **להניח שרמזי טיפוס בודקים קלט.** הם לא. קלט חיצוני עובר דרך pydantic, בגבול.
- **ליצור מתאם טיפוסים בתוך לולאה.** כל יצירה מקמפלת סכמה מחדש.
- **לשנות בבדיקה משאב בטווח רחב.** הבדיקות נהיות תלויות סדר.
- **לשפר לפני שמודדים.** מה שנראה איטי הוא לעתים רחוקות מה שאיטי.

## שאלות ראיון

1. **מה קובץ נעילה מבטיח, ומה לא?** שחזור: כולן מקבלות אותו עץ תלויות, כולל העקיפות, עם תמציות
   שמזהות קובץ שהוחלף. הוא לא מבטיח שהגרסאות בטוחות או נכונות, ולא מתעדכן לבד; עדכון הוא קומיט נפרד
   שעובר את הבדיקות. בספרייה מצמידים טווחים, כדי לא לכפות גרסאות על אחרים.

2. **מתי `Protocol` עדיף על מחלקת בסיס מופשטת?** כשהחוזה שייך לצרכנית, ובמיוחד כשאין שליטה במחלקות
   שצריכות לקיים אותו; הפלט הסטנדרטי לא יירש לעולם ממחלקה שלי. מחלקה מופשטת עדיפה כשמשתפים מימוש,
   או כשרוצים שגיאה בזמן ריצה על מתודה חסרה. המחיר של Protocol: האכיפה כולה סטטית.

   **המשך:** ואם צריך בדיקת `isinstance`? מסמנים `@runtime_checkable`, וזוכרים שנבדק רק קיום המתודות.
   לרוב עדיף לבדוק בגבול עם pydantic.

3. **מה pydantic בודקת שרמזי טיפוס לא?** את הקלט, בזמן ריצה: טיפוסים בפועל, אילוצים, צורה והמרות
   מבוקרות, עם כל השגיאות ומיקומן. ההחלטה המעניינת היא מקל מול קפדני: מקל נוח לקלט אנושי, קפדני מתאים
   לשדות כמו כסף, שבהם המרה שקטה מסתירה באג אצל השולחת.

4. **למה בדיקה שעוברת אצלך נכשלת אצל אחרת?** בלי מבנה src, הבדיקות מייבאות מתיקיית העבודה ולא מהחבילה
   המותקנת, וקובץ שחסר בבנייה לא נתפס. סיבות נוספות: תלויות לא נעולות, משאבים משותפים שמשתנים, ותלות
   בשעון או ברשת. הפתרונות: מבנה src, נעילה, טווח צר כברירת מחדל, והזרקה של זמן ואקראיות.

5. **מציעים לשכתב את לולאת הבדיקה כדי שתהיה מהירה יותר. מה את עושה?** פרופיל על קלט מייצג. אם רוב
   הזמן בספרייה, השכתוב לא ישנה הרבה, והשאלה האמיתית היא אם הבדיקה נחוצה. אם יש נקודה חמה בקוד שלנו,
   משווים חלופות, משנים ומודדים שוב באותה מכונה. המספרים נכנסים להודעת הקומיט, כדי שאפשר יהיה לשחזר.

## פתק Obsidian

הפתק המוכן נמצא גם כקובץ: `obsidian/07-python-typed-packaged.md`. זה תוכנו:

- **שני קבצים:** קובץ הפרויקט מתאר, קובץ הנעילה מצמיד, ושניהם ב־Git.
- **שני דגלים לסנכרון:** `--locked` נכשל כשהנעילה לא תואמת, `--frozen` מתקין בלי לבדוק.
- **בודקים סטטיים:** ההגדרות בקובץ הפרויקט, מצב קפדני, השתקה רק עם קוד שגיאה.
- **הטיפוס `Protocol`:** חוזה של הצרכנית, טיפוסיות מבנית, נאכף רק סטטית.
- **הטיפוס `TypedDict`:** צורה של JSON שעובר גבול.
- **תחביר חדש:** `class Ok[T]` ו־`type Outcome[T]`; מגרסה 3.14 הערות טיפוס עצלות.
- **בדיקת תקינות:** pydantic בודקת קלט בזמן ריצה, עם קפדנות לפי שדה ואיחוד מובחן.
- **בדיקות:** טווח צר כברירת מחדל, פרמטרים במקום לולאות, החלפה זמנית שמוחזרת בסוף.
- **אריזה:** מבנה src, נקודת כניסה, וקובץ `py.typed` שמפרסם את הטיפוסים.
- **ביצועים:** `cProfile` כדי לדעת מה איטי, `timeit` כדי להשוות.
- **שאלה פתוחה:** ___

<div dir="ltr" align="left">

```text
uv init --package --python 3.14 NAME
uv add 'pkg==X.Y.Z'      uv add --dev 'pytest==9.1.1'
uv sync --locked (CI)    uv sync --frozen --no-dev (image)
uv lock --upgrade-package pkg && uv sync
uv run ruff check . && uv run ruff format --check . && uv run mypy src tests
uv run pytest -q -m "not slow"
uv run python -m cProfile -o check.prof -m eventcheck check big.jsonl > /dev/null
python -m timeit -s "SETUP" "STATEMENT"
class Ok[T]: ...   type Outcome[T] = Ok[T] | Err   @override   -> list[Self]
my profile, top 5 (cumtime): ____
validate_json vs validate_python(json.loads): ____ / ____ usec
```

</div>

## תרגיל

**מטרה:** לבנות את `eventcheck`, כלי שבודק קובץ אירועים ומדווח על כל שורה פגומה. הוא פרויקט נעול
עם נקודת כניסה, בודקים סטטיים במצב קפדני, ממשקים מבניים, מודלים של pydantic, בדיקות עם פרמטרים,
ומדידת פרופיל אחת.
**זמן:** שני ימי לימוד, ימים 15 ו־16, כ־45 דקות בנייה ביום הראשון וכ־65 ביום השני.
**ידע מקדים:** הפרק הזה, ופרק 2 (uv ו־Python 3.14).

הצעדים, קוד ההתחלה, הרמזים והפתרון נמצאים ב־
[`exercises/ch07-eventcheck/README.md`](../exercises/ch07-eventcheck/README.md).

**קריטריון סיום:**

- הפרויקט מסתנכרן מקובץ הנעילה בעותק נקי, והבודקים הסטטיים עוברים.
- הכלי מדווח על כל השורות הפגומות בקובץ הדוגמה, עם מספר השורה, ומחזיר קוד יציאה 1.
- כל הבדיקות עוברות, כולל הבדיקה המסומנת `slow`.
- פלט הפרופיל נשמר, וחמש השורות המובילות ושתי מדידות ההשוואה רשומות בפתק שלך.

## העשרה אופציונלית

- המסמכים PEP 695 ו־PEP 649 קצרים יחסית, ומסבירים למה התחביר החדש עצל.
- בתיעוד של uv, הפרק על נעילה וסנכרון מפרט מתי `uv run` מסנכרן ומתי לא.
- בתיעוד של pydantic, הפרקים על מצב קפדני ועל איחודים מראים את כל ההבדלים.
- הוסיפי לכלי דגל שמגביל את מספר הבעיות המודפסות, ובדיקה עם פרמטרים עליו.

</div>
