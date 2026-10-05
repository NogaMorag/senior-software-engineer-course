---
direction: rtl
tags: [senior-course, book]
chapter: 8
---
# Python אסינכרונית ו־FastAPI

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

קישורים: [[07-python-typed-packaged]] · [[09-api-design-grpc]]
