---
direction: rtl
lang: he
title: "פרק 8: Python אסינכרונית ו־FastAPI"
chapter: 8
part: "חלק ב: Python בכירה"
---

<div dir="rtl" lang="he">

# פרק 8: Python אסינכרונית ו־FastAPI

רוב שירותי ה־Python החדשים שתפגשי הם אסינכרוניים, ורוב הבאגים בהם הם באגים של מקביליות: לולאה
שנחסמה, משימה שנשכחה, ביטול שנבלע, זמן קצוב שלא הוגדר. בפרק הזה תלמדי את asyncio כמו שצריך, ואז
תבני עליו שירות FastAPI: מקבל webhooks שבודק חתימה, מסיר כפילויות, סופר, ומתנהג רע לפי
פקודה. השירות הזה יהיה ה־`sink` של פרויקט הגמר.

## מה נלמד בפרק

- לולאת האירועים של asyncio: מה רץ מתי, ומה חוסם אותה.
- משימות, `TaskGroup` ומקביליות מובנית, ואיך ביטול באמת עובד.
- זמנים קצובים עם `asyncio.timeout`, והגבלת מקביליות עם `Semaphore`.
- שירות FastAPI: נתיבים, מודלים, כותרות ותלויות.
- מחזור החיים של השירות (lifespan): משאבים, משימות רקע וכיבוי מסודר.
- בדיקות אסינכרוניות עם httpx, בלי רשת, בלי שעון אמיתי ובלי אקראיות אמיתית.
- ביצועים: GIL, עבודה חוסמת, תהליכים מרובים, ומתי Python היא הכלי הלא נכון.

## ידע מקדים

- פרק 7: פרויקט uv, בודקים סטטיים, pydantic ובדיקות. התרגיל בנוי באותה תבנית.
- פרק 5, סעיפים 5.1 עד 5.4: מקביליות וביטול ב־Go, שאליהם נשווה.
- פרק 6, סעיפים 6.5 ו־6.6: כיבוי מסודר ובדיקות עם `httptest`, שיש להם מקבילות ישירות כאן.
- התרגיל של הפרק יוצר את קובץ דוגמאות החתימה שהצד של Go ישתמש בו בפרק 26 ובאבן הדרך M5 של Hookline.

## 8.1 asyncio: לולאת האירועים

לולאת האירועים (event loop) היא תהליכון אחד שמריץ הרבה משימות בתורות. משימה רצה עד שהיא מגיעה
ל־`await` על משהו שעוד לא מוכן (תשובה מהרשת, שינה, נעילה), ואז מחזירה את השליטה ללולאה. הלולאה
בוחרת משימה אחרת שמוכנה, ממשיכה אותה, וחוזר חלילה. אין החלפה כפויה: משימה שלא מגיעה ל־`await` לא
משחררת את הלולאה לעולם.

מכאן שלושה מונחים שכדאי להחזיק נפרדים:

| מונח | מה הוא | מה קורה כשיוצרים אותו |
| --- | --- | --- |
| coroutine | מה ש־`async def` מחזירה כשקוראים לה | כלום. היא לא רצה עד שמישהו מחכה לה |
| task | coroutine שנמסרה ללולאה | הלולאה מתזמנת אותה, גם אם אף אחד לא מחכה |
| awaitable | כל מה שאפשר לחכות לו | תלוי: coroutine, משימה, או Future |

הטעות הקלאסית היא לקרוא ל־`fetch()` בלי `await` ולחשוב שהקריאה התבצעה. היא לא: נוצר אובייקט, ובסוף
הריצה מופיעה אזהרה שהוא מעולם לא חיכו לו („never awaited”).

הכלל החשוב בפרק: **כל קוד בין שני `await` רץ בלי הפרעה**. זה גם מגבלה וגם מתנה. המגבלה: לולאה
חישובית ארוכה, או קריאה חוסמת כמו `time.sleep` או ספרייה סינכרונית של רשת, עוצרת את כל השירות.
המתנה: אין מרוץ נתונים בין משימות על אותה לולאה כל עוד אין `await` באמצע, ולכן מבנה הנתונים בזיכרון
של התרגיל לא צריך נעילה.

> **שימי לב:** מגרסה 3.14 אפשר להציץ בתהליך חי ולראות אילו משימות קיימות ועל מה כל אחת מחכה, עם
> `python -m asyncio ps` ומספר התהליך. זה הכלי הראשון כששירות „תקוע” ואין שום שגיאה בלוג.

## 8.2 משימות, TaskGroup וביטול

את התרגיל של יום 17 תריצי מתוך הקובץ `warmup.py`. הוא מריץ כמה קריאות איטיות מזויפות, ומדפיס ציר
זמן. זו הקריאה המזויפת:

<div dir="ltr" align="left">

```python
async def fake_call(call: Call, limit: asyncio.Semaphore, timeout_s: float, t: Tracker) -> None:
    async with limit:
        t.in_flight += 1
        t.max_in_flight = max(t.max_in_flight, t.in_flight)
        t.note(f"start {call.name} (in flight: {t.in_flight})")
        try:
            async with asyncio.timeout(timeout_s):
                await asyncio.sleep(call.delay_s)
            if call.fail:
                raise RuntimeError(f"{call.name} failed")
            t.results[call.name] = "ok"
            t.note(f"done  {call.name}")
        except TimeoutError:
            t.results[call.name] = "timeout"  # handled here: the siblings keep running
            t.note(f"timeout {call.name}")
        except RuntimeError:
            t.results[call.name] = "failed"
            t.note(f"failed {call.name}")
            raise  # the TaskGroup sees it and cancels the siblings
        except asyncio.CancelledError:
            t.results[call.name] = "cancelled"
            t.note(f"cancelled {call.name}")
            raise  # never swallow cancellation
        finally:
            t.in_flight -= 1
```

</div>

והפונקציה שמריצה את כולן במקביל:

<div dir="ltr" align="left">

```python
async def run_calls(calls: list[Call], *, limit: int, timeout_s: float) -> Tracker:
    tracker = Tracker()
    semaphore = asyncio.Semaphore(limit)
    try:
        async with asyncio.TaskGroup() as group:
            for call in calls:
                group.create_task(fake_call(call, semaphore, timeout_s, tracker), name=call.name)
    except* RuntimeError as failures:
        for error in failures.exceptions:
            tracker.note(f"group raised: {error}")
    return tracker
```

</div>

### 8.2.1 מקביליות מובנית

הבלוק הזה הוא מקביליות מובנית (structured concurrency): שום משימה לא יוצאת מהבלוק
בחיים. היציאה ממנו מחכה לכל המשימות, וכשאחת נכשלת, השאר מבוטלות. זה אותו רעיון כמו
`errgroup` של Go מפרק 5, אבל אכוף על ידי השפה. הנה מה שקרה בהרצה השנייה של התרגיל, שבה קריאה אחת
נכשלת אחרי 0.2 שניות:

<div dir="ltr" align="left">

```text
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
```

</div>

שימי לב לשורה של `c`: היא חיכתה למקום פנוי ב־`Semaphore`, קיבלה אותו כש־`boom` נכשלה, והתחילה
לרגע לפני שהביטול הגיע אליה. הכישלון מגיע החוצה כקבוצת חריגות (`ExceptionGroup`), כי כמה משימות
יכולות להיכשל יחד, ונתפס בתחביר `except*`.

| כלי | מה קורה כשמשימה אחת נכשלת |
| --- | --- |
| `asyncio.TaskGroup` | השאר מבוטלות, היציאה מחכה לכולן, והחריגות מגיעות כקבוצה |
| `asyncio.gather(...)` | החריגה הראשונה מגיעה למחכה, והשאר **ממשיכות לרוץ** בלי בעלים |
| `asyncio.create_task` בלי לשמור הפניה | המשימה עלולה להיאסף כזבל באמצע, והחריגה שלה הולכת לאיבוד |

לכן בקוד חדש, `TaskGroup` הוא ברירת המחדל. `gather` נשאר מתאים כשרוצים את כל התוצאות גם אם חלק
נכשלו, עם `return_exceptions=True`.

### 8.2.2 איך ביטול עובד

ביטול ב־asyncio הוא חריגה. הקריאה `task.cancel()` לא עוצרת את המשימה מיד: היא מבקשת, ובפעם הבאה
שהמשימה עומדת על `await`, נזרקת בתוכה `CancelledError`. משלושה דברים נובעים כללים:

- **החריגה יורשת מ־`BaseException`.** תפיסה של `Exception` לא בולעת אותה בטעות, וטוב שכך.
- **מי שתופס חייב לזרוק שוב.** הענף של הביטול ב־`fake_call` רושם ואז זורק. בליעה שוברת את הקבוצה ואת `asyncio.timeout`.
- **ניקוי שייך ל־`finally`.** הוא רץ גם בביטול, ולכן `in_flight` יורד תמיד.

> **איך Rust עושה את זה:**
> - ב־tokio, ביטול הוא שחרור של ה־future: הוא לא ממשיך יותר, ורק הניקוי של `Drop` רץ.
> - ב־Python, ביטול הוא חריגה שנזרקת בנקודת ה־`await`, ולכן הקוד רואה אותו ויכול לנקות או לטעות.
> - הקבוצה `TaskGroup` מקבילה ל־`JoinSet`.
> - מנהל ההקשר `asyncio.timeout` מקביל ל־`tokio::time::timeout`.

## 8.3 זמנים קצובים ו־Semaphore

### 8.3.1 asyncio.timeout

כל קריאה שיוצאת מהתהליך צריכה זמן קצוב (timeout). מנהל ההקשר `asyncio.timeout` מבטל את מה שבתוכו
כשהזמן נגמר, וממיר את הביטול ל־`TimeoutError` ביציאה מהבלוק. בהרצה הראשונה של התרגיל, הקריאה `slow`
אמורה לישון חמש שניות, עם זמן קצוב של שנייה ומגבלה של שתי קריאות בו־זמנית:

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
```

</div>

שני דברים לראות כאן. הזמן הקצוב טופל **בתוך** הקריאה, ולכן הוא לא הפיל את הקבוצה; האחיות סיימו
כרגיל. והוא נמדד מהרגע שהקריאה קיבלה מקום, לא מתחילת הריצה. אם רוצים גבול לכל העבודה יחד, עוטפים גם
את הקבוצה כולה בזמן קצוב: זה הגבול הכולל (deadline) מהספר הקודם, באותו מבנה.

הצורה הישנה `asyncio.wait_for(coro, timeout)` עדיין קיימת, ומתאימה לקריאה בודדת. מנהל ההקשר עדיף כי
הוא עוטף כמה פעולות יחד ונראה כמו כל בלוק אחר.

### 8.3.2 Semaphore: הגבלת מקביליות

`asyncio.Semaphore(2)` מאפשר רק לשתי משימות להיות בתוך `async with limit` בבת אחת. בציר הזמן,
`max in flight` לא עבר את 2, ו־`c` התחילה רק כשמקום התפנה. בשירות אמיתי הוא שומר על התלות שלך: חמישים
בקשות במקביל לשירות שמחזיק עשר הן דרך מהירה לקבל שגיאות 503. זה דפוס ה־worker pool מפרק 5, בלי
ערוצים, ובמחיר של שורה אחת. הזמן הקצוב והמגבלה מורכבים יחד: הזמן הקצוב בתוך המגבלה, כדי שהמתנה בתור
לא תיספר כזמן של הקריאה.

## 8.4 FastAPI: נתיבים, מודלים ותלויות

הספרייה FastAPI בנויה על שתי ספריות: Starlette לשכבת הרשת, ו־pydantic למודלים. פונקציה עם מעטר נתיב
היא נקודת קצה, והחתימה שלה היא החוזה: FastAPI קוראת אותה, ומפיקה ממנה את פירוק הבקשה, את בדיקת
התקינות ואת מסמך ה־OpenAPI. מתוך `app.py` של התרגיל, נקודת הקצה שמקבלת webhooks:

<div dir="ltr" align="left">

```python
    @app.post("/hooks/{name}", response_model=Accepted)
    async def receive(
        name: str,
        request: Request,
        state: StateDep,
        signature: Annotated[str | None, Header(alias="Hookline-Signature")] = None,
        event_id: Annotated[str | None, Header(alias="Hookline-Event-Id", max_length=200)] = None,
        delivery_id: Annotated[str | None, Header(alias="Hookline-Delivery-Id")] = None,
        attempt: Annotated[int | None, Header(alias="Hookline-Attempt", ge=1)] = None,
    ) -> Accepted | JSONResponse:
```

</div>

| פרמטר | מאיפה FastAPI לוקחת אותו |
| --- | --- |
| `name: str` | מהנתיב, כי `{name}` מופיע בו |
| `request: Request` | הבקשה עצמה, כדי לקרוא את הגוף הגולמי |
| `state: StateDep` | תלות (dependency), ראי בהמשך |
| `Header(alias=...)` | מכותרת, עם אילוצים של pydantic; ערך לא תקין מחזיר 422 בלי שורת קוד שלך |

### 8.4.1 מודלים בכניסה וביציאה

גוף של בקשה שמוגדר כמודל נבדק אוטומטית. מתוך `state.py`, ההגדרות של השליטה בהתנהגות:

<div dir="ltr" align="left">

```python
FailureRate = Annotated[float, Field(ge=0.0, le=1.0)]
LatencyMs = Annotated[int, Field(ge=0, le=60_000)]
FailureStatus = Annotated[int, Field(ge=400, le=599)]


class Control(BaseModel):
    """How the receiver misbehaves. Defaults: a perfectly healthy receiver."""

    model_config = ConfigDict(extra="forbid")
    failure_rate: FailureRate = 0.0
    latency_ms: LatencyMs = 0
    failure_status: FailureStatus = 503
    timeout_mode: bool = False
```

</div>

האילוצים מוגדרים פעם אחת ככינויי טיפוס, ומשמשים גם את `Control` וגם את `ControlPatch`, שבו כל שדה
אופציונלי. כך גוף של `POST /control` משנה רק את מה שנשלח, ושיעור כישלון 1.5 נדחה ב־422 לפני שהקוד שלך
רץ. טיפוס ההחזרה של פונקציה (`-> StatsOut`) הוא מודל הפלט: FastAPI מסננת ובודקת לפיו, ושדה פנימי
לא דולף החוצה בטעות.

### 8.4.2 תלויות

תלות היא פונקציה ש־FastAPI מריצה לפני נקודת הקצה ומעבירה את התוצאה כפרמטר. בתרגיל היא מוציאה את
מצב השירות מהאפליקציה:

<div dir="ltr" align="left">

```python
def get_state(request: Request) -> AppState:
    state: AppState = request.app.state.receiver
    return state


StateDep = Annotated[AppState, Depends(get_state)]
```

</div>

הכינוי `StateDep` חוסך לחזור על `Depends` בכל נקודת קצה, והבודק רואה טיפוס אמיתי. תלויות הן גם
המקום הטבעי לאימות (authentication), לחיבור למסד נתונים מתוך מאגר, ולכל מה שבדיקה רוצה להחליף. FastAPI
מאפשרת להחליף תלות בבדיקה (`app.dependency_overrides`), אבל בתרגיל לא צריך: הזמן, האקראיות והשינה
מוזרקים כבר ביצירת האפליקציה, בסעיף הבא.

> **טעות נפוצה:** לכתוב `async def` לנקודת קצה שקוראת לספרייה חוסמת. FastAPI מריצה פונקציה רגילה
> (`def`) במאגר תהליכונים, ופונקציה אסינכרונית ישירות על הלולאה. קריאה חוסמת בתוך `async def` עוצרת
> את כל השרת; בתוך `def` היא תופסת תהליכון אחד בלבד.

## 8.5 lifespan ומשאבים

למשאבים של שירות יש מחזור חיים: חיבורים, מאגרים, משימות רקע. הם נפתחים כשהשירות עולה, ונסגרים
כשהוא יורד. הספרייה מקבלת מנהל הקשר אסינכרוני אחד, ה־lifespan. מה שלפני ה־`yield` שלו רץ בעלייה, ומה
שאחריו רץ בכיבוי. מתוך `app.py`:

<div dir="ltr" align="left">

```python
    @contextlib.asynccontextmanager
    async def lifespan(app: FastAPI) -> AsyncIterator[None]:
        state = AppState(settings, clock, rng or random.Random(), sleep)
        app.state.receiver = state
        if not settings.secrets and not settings.default_secret:
            log.warning("no secrets configured: every hook will answer 404")
        reporter = None
        if settings.stats_interval_s:
            reporter = asyncio.create_task(report_stats(state, settings.stats_interval_s))
        try:
            yield
        finally:
            if reporter is not None:
                reporter.cancel()
                with contextlib.suppress(asyncio.CancelledError):
                    await reporter
            log.info("shutting down, final stats %s", state.store.total.as_dict())
```

</div>

שלוש החלטות בקטע הזה:

- **המצב נוצר בעלייה, לא ברמת המודול.** משתנה גלובלי היה משותף לכל הבדיקות; כאן כל אפליקציה, וכל בדיקה, מקבלת מצב משלה.
- **למשימת הרקע יש בעלים.** היא נשמרת במשתנה, מבוטלת בכיבוי, ומחכים לה. בלי זה היא הייתה נעלמת באמצע כתיבה.
- **הכיבוי כותב את הסטטיסטיקה האחרונה.** הסדר הזה הוא הכיבוי המסודר של פרק 6.

כל זה יושב בתוך פונקציית יצירה, `create_app`, שמקבלת את ההגדרות ואת התלויות בזמן ובאקראיות:

<div dir="ltr" align="left">

```python
def create_app(
    settings: Settings | None = None,
    *,
    clock: Clock = time.time,
    rng: RandomSource | None = None,
    sleep: Sleep = asyncio.sleep,
) -> FastAPI:
```

</div>

בייצור כל ברירות המחדל אמיתיות. בבדיקות מעבירים שעון מזויף, מקור אקראיות עם תסריט, ושינה שרק רושמת.
הטיפוס `RandomSource` הוא ממשק מבני עם מתודה אחת. המחלקה `random.Random` מתאימה לו בלי לדעת עליו,
בדיוק כמו בפרק 7.

## 8.6 בדיקות אסינכרוניות עם httpx

בפרק 6 בדקת שרת Go עם `httptest`, בלי רשת. המקבילה כאן היא `httpx.ASGITransport`: לקוח httpx שמדבר
ישירות עם האפליקציה, באותו תהליך. יש מלכודת אחת: התעבורה הזו **לא מריצה את ה־lifespan**, ולכן מצב
השירות לא קיים. ה־fixture ב־`tests/conftest.py` נכנסת אליו בעצמה:

<div dir="ltr" align="left">

```python
@pytest.fixture
async def client(app: FastAPI) -> AsyncIterator[httpx.AsyncClient]:
    """ASGITransport does not run the lifespan, so the fixture enters it explicitly."""
    async with app.router.lifespan_context(app):
        transport = httpx.ASGITransport(app=app)
        async with httpx.AsyncClient(transport=transport, base_url="http://test") as client:
            yield client
```

</div>

כדי ש־pytest יריץ בדיקה אסינכרונית צריך תוסף. התרגיל משתמש בתוסף של anyio, שמגיע ממילא עם httpx,
במקום להוסיף תלות: הקובץ מסמן את כל הבדיקות בסימון של anyio. ה־fixture הקטנה
`anyio_backend` קובעת על איזו לולאה לרוץ.

### 8.6.1 זמן ואקראיות מזויפים

בדיקה שתלויה בשעון או במזל היא בדיקה שתיכשל ביום שלישי בלילה. מתוך `tests/helpers.py`, שני הזיופים:

<div dir="ltr" align="left">

```python
@dataclass
class FakeRandom:
    """Returns the scripted values in order, then 0.99 (no failure at any rate below 0.99)."""

    values: list[float] = field(default_factory=list)

    def random(self) -> float:
        return self.values.pop(0) if self.values else 0.99


@dataclass
class FakeSleep:
    """Records every requested delay and returns at once: latency tests take no time."""

    calls: list[float] = field(default_factory=list)

    async def __call__(self, delay: float) -> None:
        self.calls.append(delay)
```

</div>

עם התסריט, הבדיקה של שיעור הכישלון אומרת בדיוק מה יקרה: הגרלה ראשונה מתחת ל־0.5 נכשלת, השנייה
מעליו עוברת. מתוך `tests/test_app.py`:

<div dir="ltr" align="left">

```python
    rng.values = [0.2, 0.7]  # first draw is below 0.5 (fail), second above (succeed)
    first = await client.post("/hooks/orders", content=BODY, headers=signed(BODY))
    second = await client.post("/hooks/orders", content=BODY, headers=signed(BODY))
    assert (first.status_code, second.status_code) == (500, 200)
    assert second.json()["duplicate"] is False  # the failed attempt was not recorded
```

</div>

גם הזמן מזויף: בדיקה אחת חותמת עם חותמת זמן ישנה ב־301 שניות, ומקבלת דחייה בלי לחכות חמש דקות.
בדיקה אחת בכוונה משתמשת בשינה האמיתית: היא שמה את השירות במצב „לא עונה”, ועוטפת את הבקשה
ב־`asyncio.timeout(0.2)`. הבקשה נקטעת, והביטול עובר דרך התעבורה לתוך נקודת הקצה. כך נבדק שהצד השולח
באמת יראה זמן קצוב.

## 8.7 ביצועים: GIL, עבודה חוסמת ו־workers

### 8.7.1 מה חוסם את הלולאה

ההרצה השלישית של `warmup.py` מריצה עבודה חוסמת של חצי שנייה ליד „דופק” שמתקתק כל 0.05 שניות. פעם
אחת ישירות בתוך coroutine, ופעם דרך `asyncio.to_thread`:

<div dir="ltr" align="left">

```python
async def inline_block() -> None:
    blocking_job(0.5)  # the mistake on purpose; ruff ASYNC251 only sees a direct time.sleep


async def in_thread() -> None:
    await asyncio.to_thread(blocking_job, 0.5)
```

</div>

<div dir="ltr" align="left">

```text
== a 0.5 s blocking call, heartbeat every 0.05 s
ticks while blocking the loop: 0
ticks while in a thread:       9
```

</div>

אפס תקתוקים: במשך חצי שנייה, שום דבר אחר לא רץ. בשרת, זה חצי שנייה שבה אף בקשה לא נענית. הכלל של
ruff שמזהה שינה חוסמת בתוך פונקציה אסינכרונית לא תפס את זה, כי הקריאה החוסמת עטופה בפונקציה רגילה.
בודקים סטטיים עוזרים, אבל לא מחליפים הבנה של מה כל ספרייה עושה.

### 8.7.2 תהליכונים, GIL ותהליכים

`asyncio.to_thread` (או `loop.run_in_executor`) מעבירה עבודה לתהליכון. זה עובד יפה לקריאה חוסמת של
קלט ופלט, כי בזמן ההמתנה התהליכון משחרר את נעילת המפרש הגלובלית (GIL). לעבודה **חישובית** זה לא
עוזר: בגרסה הרגילה של Python רק תהליכון אחד מריץ קוד Python בכל רגע, ולכן ארבעה תהליכונים חישוביים
לא מהירים מאחד. הלולאה תמשיך להגיב, כי המפרש מחליף תהליכונים כל כמה אלפיות, אבל התפוקה לא תגדל.

| סוג העבודה | הכלי |
| --- | --- |
| קלט ופלט אסינכרוני (httpx, מסד נתונים אסינכרוני) | `await` ישיר, על הלולאה |
| ספרייה חוסמת של קלט ופלט | `asyncio.to_thread` |
| חישוב כבד | `ProcessPoolExecutor` דרך `run_in_executor`, או שירות נפרד |
| הרבה בקשות על כמה ליבות | כמה תהליכי שרת (workers) |

מגרסה 3.14 קיימת גם גרסה רשמית בלי GIL, נכון לאוקטובר 2026, שנקראת free-threaded. היא מותקנת כגרסה נפרדת, ולא כל ספרייה עם הרחבות ב־C תומכת בה. היא שווה ניסוי, לא הנחה.

### 8.7.3 כמה תהליכי שרת, ומחיר המצב בזיכרון

השרת uvicorn יודע להריץ כמה תהליכים שמקשיבים לאותו פורט. כך משתמשים בכמה ליבות בלי GIL משותף:

<div dir="ltr" align="left">

```bash
RECEIVER_DEFAULT_SECRET=change-me uv run uvicorn --factory receiver.app:create_app --port 8081 --workers 2
```

</div>

אבל לכל תהליך יש זיכרון משלו. כשהרצתי את זה, שלחתי שש בקשות, ושאלתי את `/stats` ארבע פעמים, קיבלתי
שתי תשובות שונות לסירוגין: תהליך אחד ראה חמש בקשות, והשני אחת. המקבל של התרגיל שומר מצב בזיכרון,
ולכן הוא חייב לרוץ כתהליך אחד, ובפרויקט הגמר תריצי כמה עותקים שלו כשירותים נפרדים. בשירות אמיתי,
מצב שצריך לשרוד ולהיות משותף עובר ל־PostgreSQL או ל־Redis (פרקים 10 עד 12).

### 8.7.4 מתי Python היא הכלי הלא נכון

Python אסינכרונית טובה לשירותים שרוב הזמן מחכים: קריאות HTTP, מסד נתונים, תורים. היא פחות מתאימה
כשהשירות עצמו מחשב הרבה בכל בקשה, כשזמן התגובה בזנב (p99) צריך להיות יציב מאוד, או כשהשירות חייב
מקביליות אמיתית בתוך תהליך אחד. שם Go, שרצה על כל הליבות, או Rust, הן בחירה טבעית
יותר. בראיון, התשובה הטובה מתחילה במדידה: איפה הזמן הולך, ומה צוואר הבקבוק בפועל.

> **איך Rust עושה את זה:**
> - הפונקציה `spawn_blocking` מקבילה ל־`asyncio.to_thread`, מאותה סיבה בדיוק.
> - ההבדל: משימות של tokio רצות על כמה ליבות, ולכן חישוב בתוך משימה „רק” מאט את הליבה שלו.
> - ב־Python עם GIL, אותו חישוב עוצר את כל הלולאה, וזה הופך את הכלל מהמלצה לחובה.

## 8.8 תוכנית: מקבל webhooks

כל הפרק מתכנס לשירות אחד, `receiver`. הוא מקבל `POST /hooks/{name}`, בודק את החתימה, מסיר כפילויות
לפי מזהה האירוע, סופר, ומתנהג רע לפי פקודה. בפרויקט הגמר הוא יהיה ה־`sink`: היעד שהעובדים של Hookline
שולחים אליו, ושמספרים לפיו כמה משלוחים הגיעו.

### 8.8.1 החתימה

הכותרת `Hookline-Signature` נראית כך: `t=<unix seconds>,v1=<hex>`. הערך אחרי `v1` הוא HMAC-SHA256
עם הסוד של נקודת הקצה. הקלט ל־HMAC הוא חותמת הזמן, נקודה, והגוף הגולמי. מתוך `signature.py`:

<div dir="ltr" align="left">

```python
def digest(secret: bytes, timestamp: int, body: bytes) -> bytes:
    """The raw 32-byte HMAC over the exact bytes "<t>." followed by the body."""
    return hmac.new(secret, f"{timestamp}.".encode() + body, hashlib.sha256).digest()
```

</div>

ובדיקת החתימה, אחרי הפירוק של הכותרת:

<div dir="ltr" align="left">

```python
    if abs(now - timestamp) > tolerance_s:
        raise SignatureError(
            "timestamp_out_of_tolerance", f"t={timestamp} is off by more than {tolerance_s}s"
        )
    expected = digest(secret, timestamp, body)
    # Check every candidate, without stopping early, so timing says nothing about which matched.
    matched = False
    for candidate in candidates:
        matched |= hmac.compare_digest(expected, candidate)
    if not matched:
        raise SignatureError("signature_mismatch", "no v1 signature matches")
    return timestamp
```

</div>

| החלטה | למה |
| --- | --- |
| חתימה על הגוף **הגולמי** | פירוק ובנייה מחדש של JSON משנים רווחים וסדר מפתחות, והחתימה נשברת |
| חותמת זמן בתוך הקלט החתום | תוקפת לא יכולה לשנות אותה בלי לשבור את החתימה |
| סבולת של חמש דקות | בקשה שנלכדה ונשלחה שוב מאוחר יותר נדחית |
| השוואה בזמן קבוע | השוואה רגילה עוצרת בתו הראשון השונה, וזמן התגובה מסגיר כמה תווים נכונים |
| כמה ערכי `v1` | בזמן החלפת סוד, השולחת חותמת בשני הסודות, והמקבל מקבל כל אחד מהם |

הסדר של הבדיקות קבוע: קודם צורה, אחר כך זמן, ורק בסוף HMAC. הבדיקה היקרה רצה רק על בקשה שעברה את
הבדיקות הזולות. כל כישלון מחזיר 401 עם הסיבה, ונספר כ־`rejected`.

בתיקיית הנתונים של התרגיל יש קובץ של דוגמאות בדיקה (test vectors): קלט ותוצאה מצופה לכל
מקרה. הן נוצרו מקוד הפתרון ונבדקו מול `openssl dgst -hmac`. בפרק 26 ובאבן הדרך M5, החותם שתכתבי
ב־Go ייבדק מול אותו קובץ, וכך שני הצדדים מסכימים על החוזה ולא רק כל אחד עם עצמו.

### 8.8.2 כפילויות, ספירה ושליטה

Hookline מבטיחה מסירה „לפחות פעם אחת” (at-least-once), ולכן המקבל **יראה** כפילויות: ניסיון חוזר
אחרי זמן קצוב, למשל, כשהתשובה הראשונה דווקא הגיעה. המקבל מזהה אירוע לפי הזוג של שם נקודת הקצה ומזהה
האירוע (`Hookline-Event-Id`). אותו אירוע לשתי נקודות קצה הוא שני משלוחים שונים, ואותו אירוע פעמיים
לאותה נקודה הוא כפילות:

| מונה | מה הוא סופר |
| --- | --- |
| `received` | בקשות חתומות שהתקבלו, כולל כפילויות |
| `unique` | זוגות שונים של נקודת קצה ומזהה אירוע |
| `duplicates` | בקשות שהזוג שלהן כבר נראה |
| `rejected` | חתימה שגויה או חסרה, או מזהה אירוע חסר |
| `failed` | כישלונות שהוזרקו דרך `/control`; הם **לא** נרשמים כקבלה |

המונה האחרון חשוב לפרויקט הגמר. אם כישלון מוזרק היה נרשם כקבלה, משלוח שבסוף נכשל סופית היה נראה
כאילו הגיע, והבדיקה „אפס משלוחים חסרים” הייתה משקרת. הסדר בנקודת הקצה הוא לכן: חתימה, אחר כך התנהגות
רעה מבוקשת (מצב „לא עונה”, השהיה, הגרלת כישלון), ורק בסוף רישום.

השליטה, `POST /control`, מקבלת חלק מהשדות: שיעור כישלון, השהיה באלפיות שנייה, קוד הסטטוס של
הכישלון, ומצב „לא עונה” שמחזיק את הבקשה שעה. `GET /receipts` מחזיר את רשימת הקבלות, ו־`POST /reset`
מאפס הכול בין ריצות.

## סיכום

- הלולאה מריצה קוד עד ה־`await` הבא; קריאה חוסמת או חישוב ארוך עוצרים את כל השירות.
- הקבוצה `TaskGroup` היא מקביליות מובנית: כישלון אחד מבטל את האחיות, והחריגות מגיעות כקבוצה.
- ביטול הוא חריגה שנזרקת ב־`await`: לא בולעים אותה, ומנקים ב־`finally`.
- זמן קצוב עם `asyncio.timeout`, ומגבלת מקביליות עם `Semaphore`; הזמן הקצוב בתוך המגבלה.
- ב־FastAPI, החתימה של הפונקציה היא החוזה: נתיב, כותרות, מודלים ותלויות, עם 422 אוטומטי.
- משאבים ומשימות רקע שייכים ל־lifespan, עם בעלים ועם כיבוי מסודר.
- בבדיקות: `ASGITransport`, כניסה ידנית ל־lifespan, ושעון, אקראיות ושינה מוזרקים.
- תהליכונים לקלט ופלט חוסם, תהליכים לחישוב, ומצב בזיכרון מחייב תהליך שרת אחד.
- המקבל בודק חתימה על הגוף הגולמי בזמן קבוע, מסיר כפילויות לפי נקודת קצה ומזהה, ולא סופר כישלון מוזרק.

## מונחים לזכור

| מונח בעברית | באנגלית | במשפט אחד |
| --- | --- | --- |
| לולאת האירועים | event loop | תהליכון אחד שמריץ משימות בתורות, ומחליף ביניהן רק ב־`await` |
| coroutine | coroutine | מה ש־`async def` מחזירה; לא רצה עד שמחכים לה |
| משימה | task | coroutine שנמסרה ללולאה ומתוזמנת |
| מקביליות מובנית | structured concurrency | משימות שלא חיות יותר מהבלוק שיצר אותן |
| קבוצת חריגות | exception group | כמה חריגות שמגיעות יחד, ונתפסות עם `except*` |
| ביטול | cancellation | בקשה לעצור משימה, שמגיעה אליה כחריגה ב־`await` |
| זמן קצוב | timeout | כמה זמן מחכים לפעולה לפני שמבטלים אותה |
| נעילת המפרש הגלובלית | GIL | הנעילה שמאפשרת רק לתהליכון אחד להריץ קוד Python בכל רגע |
| מחזור חיים | lifespan | הקוד שרץ בעליית השירות ובכיבוי שלו |
| תלות | dependency | פונקציה ש־FastAPI מריצה ומעבירה את התוצאה לנקודת הקצה |
| מקבל webhooks | webhook receiver | שירות שמקבל אירועים חתומים בבקשות HTTP |
| השוואה בזמן קבוע | constant-time comparison | השוואה שזמנה לא תלוי במקום ההבדל הראשון |
| הסרת כפילויות | deduplication | זיהוי של אותו משלוח שהגיע שוב, ורישום אחד בלבד |

## טעויות נפוצות

- **לקרוא ל־coroutine בלי `await`.** שום דבר לא קורה, ומופיעה אזהרה בסוף.
- **ליצור משימה בלי לשמור הפניה ובלי בעלים.** היא עלולה להיעלם באמצע, והחריגה שלה הולכת לאיבוד.
- **לבלוע `CancelledError`.** הקבוצה והזמן הקצוב מפסיקים לעבוד. תופסים רק כדי לרשום, וזורקים שוב.
- **ספרייה חוסמת בתוך `async def`.** כל השרת נעצר; משתמשים ב־`asyncio.to_thread` או בפונקציה רגילה.
- **לבדוק חתימה על JSON שפורק ונבנה מחדש.** החתימה נשברת; בודקים על הבתים הגולמיים.
- **להשוות חתימות עם `==`.** השוואה רגילה מדליפה מידע בזמן התגובה; משתמשים ב־`hmac.compare_digest`.
- **להריץ שירות עם מצב בזיכרון בכמה תהליכי שרת.** כל תהליך רואה חלק מהבקשות.

## שאלות ראיון

1. **מה חוסם את לולאת האירועים, ואיך מזהים את זה?** כל קוד שלא מגיע ל־`await`: חישוב ארוך, שינה
   חוסמת, ספרייה סינכרונית של רשת או קבצים. הסימפטום הוא זמני תגובה שקופצים לכל הבקשות יחד, לא רק
   לאיטית. מזהים עם מצב הדיבאג של asyncio, שמדווח על צעדים ארוכים, עם `python -m asyncio ps` על תהליך
   חי, או עם „דופק” כמו בתרגיל. מתקנים עם תהליכון לקלט ופלט ועם תהליך לחישוב.

2. **מה ההבדל בין `TaskGroup` ל־`gather` כשמשימה נכשלת?** עם `TaskGroup` האחיות מבוטלות, היציאה מחכה
   לכולן, והחריגות מגיעות כקבוצה: אין משימות יתומות. עם `gather`, החריגה הראשונה מגיעה למחכה והשאר
   ממשיכות לרוץ בלי שאיש מחכה להן. `gather` עם `return_exceptions=True` מתאים כשרוצים את כל התוצאות
   למרות כישלונות.

   **המשך:** ואם אחת המשימות לא אמורה להפיל את האחרות? מטפלים בכישלון שלה בתוכה, כמו הזמן הקצוב ב־`fake_call`,
   ומחזירים תוצאה; רק מה שאמור לעצור את הכול נזרק החוצה.

3. **למה בודקים חתימה על הגוף הגולמי, ולמה בזמן קבוע?** כי החתימה היא על בתים, ופירוק ובנייה מחדש של
   JSON לא משמרים אותם; גם רווח אחד שובר אותה. זמן קבוע, כי השוואה שעוצרת בהבדל הראשון מאפשרת לנחש
   חתימה תו אחר תו לפי זמני התגובה. חותמת הזמן החתומה, עם סבולת, מונעת שליחה חוזרת של בקשה שנלכדה.

4. **למה מקבל webhooks חייב להסיר כפילויות?** כי מסירה דרך HTTP לצד שלישי היא לכל היותר „לפחות פעם
   אחת”: תשובה שהלכה לאיבוד אחרי שהבקשה עובדה מובילה לניסיון חוזר. המקבל מזהה לפי מזהה האירוע (וכאן גם
   לפי נקודת הקצה), ומחזיר הצלחה גם לכפילות, כדי שהשולחת תפסיק לנסות. השאלה הבאה בדרך כלל: כמה זמן
   שומרים מזהים, והתשובה תלויה בחלון הניסיונות של השולחת.

5. **השירות שלך ב־Python איטי תחת עומס. מה האפשרויות?** קודם מודדים: האם הלולאה חסומה, האם זה חישוב,
   או המתנה לתלות. אחר כך לפי המקרה: להוציא קריאות חוסמות לתהליכון, להעביר חישוב לתהליך או לשירות נפרד,
   להגביל מקביליות מול תלות איטית, ולהריץ כמה תהליכי שרת (אחרי שמוציאים מצב מהזיכרון). אם החישוב הוא
   לב השירות וזמן התגובה בזנב קריטי, לשקול שפה אחרת לרכיב הזה.

## פתק Obsidian

הפתק המוכן נמצא גם כקובץ: `obsidian/08-python-async-fastapi.md`. זה תוכנו:

- **הלולאה:** תהליכון אחד; מחליפים משימה רק ב־`await`; בין שני `await` אין הפרעה.
- **מה חוסם:** חישוב ארוך, שינה חוסמת, ספרייה סינכרונית. הפתרון הוא תהליכון לקלט ופלט ותהליך לחישוב.
- **הקבוצה `TaskGroup`:** כישלון מבטל את האחיות, היציאה מחכה לכולן, החריגות כקבוצה.
- **ביטול:** חריגה שנזרקת ב־`await`; לא בולעים; מנקים ב־`finally`.
- **זמן קצוב ומגבלה:** `asyncio.timeout` בתוך `Semaphore`, כדי שהתור לא ייספר כזמן הקריאה.
- **ב־FastAPI:** חתימת הפונקציה היא החוזה; תלויות דרך `Annotated`; 422 אוטומטי.
- **ה־lifespan:** משאבים ומשימות רקע עם בעלים, וכיבוי מסודר.
- **בדיקות:** `ASGITransport` לא מריץ lifespan; שעון, אקראיות ושינה מוזרקים.
- **חתימה:** הגוף הגולמי, חותמת זמן חתומה, סבולת של חמש דקות, השוואה בזמן קבוע.
- **כפילויות:** לפי נקודת קצה ומזהה אירוע; כישלון מוזרק לא נרשם כקבלה.
- **שאלה פתוחה:** ___

<div dir="ltr" align="left">

```text
async with asyncio.TaskGroup() as tg: tg.create_task(coro(), name="x")
except* RuntimeError as eg: eg.exceptions
async with asyncio.timeout(1.0): ...          # raises TimeoutError on exit
sem = asyncio.Semaphore(2); async with sem: ...
await asyncio.to_thread(blocking_fn, arg)     # blocking I/O off the loop
python -m asyncio ps PID                      # 3.14: live task listing

Hookline-Signature: t=<unix>,v1=<hex HMAC-SHA256(secret, "<t>." + raw body)>
check order: malformed -> timestamp_out_of_tolerance (300 s) -> signature_mismatch
hmac.compare_digest(expected_digest, candidate_digest)

async with app.router.lifespan_context(app):
    httpx.AsyncClient(transport=httpx.ASGITransport(app=app), base_url="http://test")

uv run receiver            uv run receiver-send URL --secret change-me
GET /stats  GET /receipts  POST /control {"failure_rate": 0.5}  POST /reset
```

</div>

## תרגיל

**מטרה:** לבנות את `receiver`, מקבל webhooks שבודק חתימה, מסיר כפילויות לפי מזהה אירוע, סופר
קבלות, ומתנהג רע לפי פקודה (שיעור כישלון, השהיה, קוד סטטוס ומצב „לא עונה”). הבדיקות אסינכרוניות
ודטרמיניסטיות. לפניו, תרגיל חימום עם `TaskGroup`, זמן קצוב ו־`Semaphore`.
**זמן:** שני ימי לימוד, ימים 17 ו־18, כ־40 דקות בנייה ביום הראשון וכ־70 ביום השני.
**ידע מקדים:** הפרק הזה, ופרק 7 (התרגיל בנוי באותה תבנית).

הצעדים, קוד ההתחלה, הרמזים והפתרון נמצאים ב־
[`exercises/ch08-receiver/README.md`](../exercises/ch08-receiver/README.md).

**קריטריון סיום:**

- תרגיל החימום מראה מגבלה של שתי קריאות, זמן קצוב שלא פוגע באחיות, וכישלון שמבטל אותן.
- הבדיקות מראות שחתימה שגויה נדחית, שכפילות נרשמת פעם אחת, ושההגדרה של שיעור כישלון מייצרת שגיאות.
- כל דוגמאות החתימה מהקובץ המשותף עוברות.
- השירות רץ מקומית דרך uv, ומגיב לבקשה חתומה שנשלחת מכלי השליחה.

## העשרה אופציונלית

- בתיעוד של asyncio, הפרקים על משימות ועל ביטול כתובים היטב וקצרים יחסית.
- המסמך PEP 779 מסביר מה „תמיכה רשמית” בגרסה בלי GIL אומרת, ומה עוד חסר.
- הוסיפי למקבל שליטה נפרדת לכל נקודת קצה, כך שנקודה אחת תהיה במצב „לא עונה” והאחרות בריאות.
- כתבי בדיקת תכונה שמוודאת שכל חתימה שנוצרה עוברת בדיקה, ושכל שינוי של בית אחד בגוף נדחה.

</div>
