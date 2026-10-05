---
direction: rtl
lang: he
---

<div dir="rtl" lang="he">

# תרגיל פרק 8: `receiver`, מקבל ה־webhooks

**מטרה:** לבנות שירות שמקבל webhooks חתומים, בודק את החתימה, מסיר כפילויות לפי מזהה האירוע, סופר
קבלות, ומתנהג רע לפי פקודה. לפניו, תרגיל חימום ב־asyncio: קבוצת משימות, זמן קצוב לכל קריאה, ומגבלת
מקביליות. השירות הזה יהפוך ל־`sink` של פרויקט הגמר, ודוגמאות החתימה שלו ישמשו גם את הצד של Go.

**זמן משוער:** שני ימי לימוד. יום 17: כ־40 דקות בנייה. יום 18: כ־70 דקות בנייה.

**ידע מקדים:** פרק 8, ופרק 7: אותה תבנית פרויקט, ואותם בודקים ובדיקות. התרגיל לא מייבא קוד מתרגילים
קודמים.

## מה בתיקייה הזאת

| קובץ | תפקיד |
| --- | --- |
| `starter/src/receiver/warmup.py` | החימום, עם המשימות 1 עד 3 |
| `starter/src/receiver/signature.py` | החתימה: `sign` ו־`digest` מוכנות, המשימות 4 ו־5 |
| `starter/src/receiver/state.py` | מונים, קבלות ומודלי השליטה, עם המשימה 6 |
| `starter/src/receiver/app.py` | האפליקציה, עם המשימות 7 עד 9 |
| `starter/src/receiver/settings.py`, `main.py`, `send.py` | תצורה, הפעלה וכלי שליחה, שלמים |
| `tests/helpers.py`, `tests/conftest.py` | הזיופים של זמן ואקראיות, והלקוח שמחובר לאפליקציה |
| `tests/test_warmup.py` | 4 בדיקות ליום 17 |
| `tests/test_signature.py`, `test_settings.py`, `test_app.py` | 58 בדיקות ליום 18 |
| `data/signature-vectors.json` | דוגמאות החתימה המשותפות: 5 של חתימה ו־25 של בדיקה |
| `solution/` | הפתרון המלא |

## יום 17: חימום ב־asyncio

**1. העתיקי את קוד ההתחלה ואת הנתונים, וסנכרני:**

<div dir="ltr" align="left">

```bash
mkdir -p ~/senior-course/ch08
cp -r "$COURSE/exercises/ch08-receiver/starter/." ~/senior-course/ch08/
cp -r "$COURSE/exercises/ch08-receiver/data" ~/senior-course/ch08/
cd ~/senior-course/ch08
uv sync --locked
```

</div>

הבדיקות מחפשות את הקובץ `data/signature-vectors.json` בתיקיות שמעליהן, ולכן הנתונים חייבים לשבת
ליד `tests/`.

**2. קראי את `warmup.py`.** המחלקה `Call` מתארת קריאה מזויפת (שם, משך, האם להיכשל). המחלקה `Tracker`
סופרת קריאות פעילות ורושמת ציר זמן. בסוף הקובץ, `main` מריצה שלושה תרחישים.

**3. מלאי את המשימה 1** (`fake_call`), לפי הפירוט בהערה. הסדר חשוב: המגבלה בחוץ, הזמן הקצוב בפנים,
והמונה יורד ב־`finally`.

**4. מלאי את המשימה 2** (`run_calls`): משימה לכל קריאה בתוך קבוצת משימות, `Semaphore` אחד לכולן,
ותפיסה של הכישלונות עם `except*`.

**5. מלאי את המשימה 3** (`in_thread`): אותה עבודה חוסמת, דרך `asyncio.to_thread`.

**6. בדיקות היום, ואז ההרצה:**

<div dir="ltr" align="left">

```bash
uv run pytest -q tests/test_warmup.py
uv run python -m receiver.warmup
```

</div>

השווי את הפלט לסעיף „פלט צפוי”. הזמנים אצלך יכולים לזוז בכמה מאיות, אבל הסדר והתוצאות זהים.

**7. קומיט:**

<div dir="ltr" align="left">

```bash
cd ~/senior-course
git add ch08
git commit -m "Chapter 8: asyncio warm-up with TaskGroup, timeouts and a semaphore"
```

</div>

## יום 18: המקבל

**8. החתימה: המשימות 4 ו־5** בקובץ `signature.py`. הפירוק (`parse`) מקבל את הכותרת ומחזיר חותמת זמן
ורשימת חתימות כבתים. הבדיקה (`verify`) עובדת בסדר קבוע: צורה, זמן, ורק אז HMAC בהשוואה בזמן קבוע.
כל כלל של הצורה מתועד בשדה `reasons` של קובץ הדוגמאות.

<div dir="ltr" align="left">

```bash
cd ~/senior-course/ch08
uv run pytest -q tests/test_signature.py
```

</div>

כל דוגמה היא בדיקה נפרדת עם שם, למשל `body re-serialized with spaces`. כשאחת נכשלת, השם אומר מה שברת.

**9. הרישום: המשימה 6** בקובץ `state.py`. המונים מתעדכנים גם בסיכום וגם במונים של נקודת הקצה. הפונקציה
מחזירה אמת כשזו כפילות.

**10. האפליקציה: המשימות 7 עד 9** בקובץ `app.py`. הוסיפי קודם את הייבוא שבהערה בראש הקובץ.

| משימה | מה עושים |
| --- | --- |
| 7 | ה־lifespan: יוצרים את המצב, שומרים אותו על האפליקציה, ומפעילים משימת רקע לפי ההגדרות. בכיבוי מבטלים אותה ומחכים לה |
| 8 | נקודת הקצה: נקודה לא מוכרת, גוף גולמי, בדיקת חתימה, מזהה חסר, ואז התנהגות רעה ורישום, בסדר שבהערה |
| 9 | השליטה: רק השדות שנשלחו משתנים |

<div dir="ltr" align="left">

```bash
uv run pytest -q
uv run ruff check . && uv run ruff format --check . && uv run mypy src tests
```

</div>

**11. הריצי את השירות** בחלון אחד:

<div dir="ltr" align="left">

```bash
cd ~/senior-course/ch08
RECEIVER_DEFAULT_SECRET=change-me RECEIVER_STATS_INTERVAL_S=10 uv run receiver
```

</div>

**12. ושלחי אליו בקשות** מחלון שני. כלי השליחה חותם בדיוק כמו העובדים של Hookline:

<div dir="ltr" align="left">

```bash
cd ~/senior-course/ch08
uv run receiver-send http://127.0.0.1:8081/hooks/orders --secret change-me --event-id evt_42
uv run receiver-send http://127.0.0.1:8081/hooks/orders --secret change-me --event-id evt_42
uv run receiver-send http://127.0.0.1:8081/hooks/orders --secret wrong
uv run receiver-send http://127.0.0.1:8081/hooks/orders --secret change-me --skew -400
curl -s -X POST http://127.0.0.1:8081/control -H 'Content-Type: application/json' -d '{"failure_rate": 1.0}'
uv run receiver-send http://127.0.0.1:8081/hooks/orders --secret change-me
curl -s http://127.0.0.1:8081/stats
```

</div>

| פקודה | מה היא בדקה |
| --- | --- |
| שליחה ראשונה של `evt_42` | בקשה חתומה מתקבלת |
| שליחה שנייה של `evt_42` | כפילות: 200, אבל נרשמת פעם אחת |
| סוד שגוי | דחייה עם הסיבה `signature_mismatch` |
| הזזה של 400 שניות אחורה | דחייה על חותמת זמן מחוץ לסבולת |
| שיעור כישלון 1.0 ואז שליחה | כישלון מוזרק, שלא נרשם כקבלה |

עצרי את השירות עם Ctrl+C, וראי בלוג את שורת הסטטיסטיקה האחרונה: זה ה־lifespan שנסגר.

**13. אופציונלי: שני תהליכי שרת.** הריצי שוב, הפעם עם שני תהליכים, שלחי כמה בקשות, ושאלי את
`/stats` כמה פעמים. התשובות יתחלפו, כי לכל תהליך זיכרון משלו (סעיף 8.7.3):

<div dir="ltr" align="left">

```bash
RECEIVER_DEFAULT_SECRET=change-me uv run uvicorn --factory receiver.app:create_app --port 8081 --workers 2
```

</div>

**14. קומיט:**

<div dir="ltr" align="left">

```bash
cd ~/senior-course
git add ch08
git commit -m "Chapter 8: webhook receiver with signatures, dedupe, controls and async tests"
```

</div>

## פלט צפוי

החימום (הזמנים בערך):

<div dir="ltr" align="left">

```text
== timeout handled per call (max in flight: 2)
  0.00s  start a (in flight: 1)
  0.00s  start b (in flight: 2)
  0.30s  done  a
  0.30s  done  b
  0.30s  start c (in flight: 1)
  0.30s  start slow (in flight: 2)
  0.60s  done  c
  0.60s  start d (in flight: 2)
  0.90s  done  d
  1.30s  timeout slow
results: {'a': 'ok', 'b': 'ok', 'c': 'ok', 'd': 'ok', 'slow': 'timeout'}
== one failure cancels the siblings (max in flight: 3)
  0.00s  start a (in flight: 1)
  0.00s  start boom (in flight: 2)
  0.00s  start b (in flight: 3)
  0.20s  failed boom
  0.20s  start c (in flight: 3)
  0.20s  cancelled c
  0.20s  cancelled b
  0.20s  cancelled a
  0.20s  group raised: boom failed
results: {'a': 'cancelled', 'b': 'cancelled', 'boom': 'failed', 'c': 'cancelled'}
== a 0.5 s blocking call, heartbeat every 0.05 s
ticks while blocking the loop: 0
ticks while in a thread:       9
```

</div>

מספר התקתוקים בתהליכון יכול להיות 8 או 9; העיקר שבחסימה הוא 0.

כל הבדיקות, בסוף יום 18:

<div dir="ltr" align="left">

```text
62 passed
```

</div>

הבקשות מצעד 12, בסדר שלהן:

<div dir="ltr" align="left">

```text
200 {"status":"ok","duplicate":false}
200 {"status":"ok","duplicate":true}
401 {"detail":"signature_mismatch"}
401 {"detail":"timestamp_out_of_tolerance"}
{"failure_rate":1.0,"latency_ms":0,"failure_status":503,"timeout_mode":false}
503 {"detail":"injected failure"}
{"received":2,"unique":1,"duplicates":1,"rejected":2,"failed":1,"hooks":{"orders":{"received":2,"unique":1,"duplicates":1,"rejected":2,"failed":1}}}
```

</div>

## רשימת בדיקה

- [ ] החימום מראה מגבלה של שתיים, זמן קצוב שלא פוגע באחיות, וכישלון שמבטל אותן.
- [ ] בחסימה של הלולאה אין תקתוקים, ובתהליכון יש.
- [ ] כל 30 דוגמאות החתימה מהקובץ המשותף עוברות.
- [ ] הבדיקות מראות חתימה שגויה נדחית, כפילות שנרשמת פעם אחת, ושיעור כישלון שמייצר שגיאות.
- [ ] השירות רץ דרך uv, ועונה על הבקשות מצעד 12 כמו בפלט הצפוי.
- [ ] הבודקים הסטטיים עוברים על הקוד ועל הבדיקות.
- [ ] שני קומיטים, אחד לכל יום.

## טעויות נפוצות ותיקונים

- **לבלוע את `CancelledError` בתוך `fake_call`.** הבדיקה של ביטול האחיות נתקעת או נכשלת, כי הקבוצה לא
  יכולה לסיים. רושמים ואז זורקים שוב, בשורה נפרדת.
- **לשים את `asyncio.timeout` מחוץ ל־`async with limit`.** בפלט של החימום, `timeout slow` מופיע אחרי
  שנייה אחת במקום אחרי 1.30, כי הזמן בתור נספר כזמן הקריאה. המגבלה בחוץ, הזמן הקצוב בפנים.
- **לבדוק חתימה על JSON שפורק ונבנה מחדש.** בדיקות האפליקציה מקבלות 401 על בקשות תקינות, כי רווחים
  נוספו לגוף. החתימה היא על `await request.body()`, הבתים כפי שהגיעו.
- **להשוות את ה־hex כמחרוזת עם `==`.** הבדיקה שמחפשת קריאה ל־`hmac.compare_digest` נכשלת. ממירים את
  שני הצדדים לבתים ומשווים בזמן קבוע.
- **לבדוק ספרות עם `str.isdigit()`.** הדוגמה `t with non-ASCII digits` נכשלת עם סיבה אחרת, ותו כמו
  „²” מפיל את ההמרה למספר. משתמשים בתבנית `_UNIX_SECONDS`, שמקבלת רק 0 עד 9.
- **לרשום קבלה לפני הגרלת הכישלון.** בבדיקה של שיעור הכישלון, הבקשה השנייה מסומנת ככפילות. בפרויקט
  הגמר, משלוח שנכשל ייראה כאילו הגיע. רישום הוא תמיד הצעד האחרון.

## רמזים

- ב־`parse`, הפעולה `item.strip().partition("=")` מחזירה שלושה חלקים; אם החלק האמצעי ריק, אין סימן שוויון.
- ב־`verify`, השתמשי ב־`abs(now - timestamp)`: הסבולת חלה גם על שעון של שולחת שממהר.
- ב־lifespan, משימת רקע שבוטלה זורקת `CancelledError` כשמחכים לה; עטפי את ההמתנה ב־`contextlib.suppress`.
- שגיאה על `app.state` בבדיקות האפליקציה אומרת שהמצב לא נשמר בשם `receiver` בתחילת מחזור החיים.

## הפתרון

`solution/`. **למה הוא עובד כך:** כל מה שמשתנה בין ייצור לבדיקה (שעון, אקראיות, שינה) מוזרק
ל־`create_app`, ולכן הבדיקות מדויקות ומהירות בלי קוד מיוחד לבדיקות. הבדיקה של החתימה היא פונקציה טהורה
על בתים, בסדר קבוע, וקובץ הדוגמאות הוא החוזה שהצד של Go יבדוק את עצמו מולו. כל העדכונים של המצב קורים
בין שני `await`, ולכן אין צורך בנעילה. והמונה `failed` נפרד מהקבלות, כך שהספירה של משלוחים ייחודיים
נשארת אמינה.

## מה התרגיל מדגים

- מקביליות מובנית עם `TaskGroup`, ביטול שעובר נכון, זמן קצוב לכל קריאה ומגבלת מקביליות.
- למה עבודה חוסמת עוצרת את הלולאה, ואיך מוציאים אותה לתהליכון.
- שירות FastAPI עם נתיבים, כותרות, מודלים, תלויות ו־lifespan שמנהל משאב ומשימת רקע.
- בדיקות אסינכרוניות בתוך התהליך, עם זמן ואקראיות מוזרקים.
- חוזה חתימה משותף לשתי שפות, עם דוגמאות בדיקה, השוואה בזמן קבוע וסבולת זמן.
- הסרת כפילויות לפי נקודת קצה ומזהה, כבסיס לבדיקת „אפס משלוחים חסרים” בפרויקט הגמר.

</div>
