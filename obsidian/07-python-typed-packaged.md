---
direction: rtl
tags: [senior-course, book]
chapter: 7
---
# Python עם טיפוסים, אריזה ובדיקות

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

קישורים: [[06-go-http-service]] · [[08-python-async-fastapi]]
