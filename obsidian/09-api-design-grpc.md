---
direction: rtl
tags: [senior-course, book]
chapter: 9
---
# עיצוב API ו־gRPC

- **חוזה:** משאבים, פעלים אחידים, קודי סטטוס; 4xx הלקוח טעה, 5xx אנחנו.
- **מקור אמת:** הקובץ `openapi.yaml` קודם לקוד; כלי בודק אותו, בדיקות בודקות שהשרת מקיים אותו.
- **שגיאות:** פרטי בעיה לפי RFC 9457; מחליטים לפי `type`, לא לפי הטקסט; גם שגיאות הנתב.
- **עימוד:** היסט מכפיל ומדלג תחת שינויים; cursor אטום על זמן ומזהה לא.
- **אידמפוטנטיות:** אותה בקשה מקבלת את התשובה השמורה, אחרת 422, מקבילה 409; 5xx משחרר; החלטה אטומית.
- **טיוטה:** הכותרת `Idempotency-Key` היא טיוטה שפג תוקפה, לא RFC (נכון לאוקטובר 2026).
- **תאימות:** מוסיפים ולא משנים; לקוח סובלני, שרת קפדן; הוצאה משימוש עם כותרות ומדידה.
- **gRPC:** המספרים הם החוזה; הקוד המחולל במאגר. לא לדפדפנים, לא ל־API ציבורי.
- **עוד נגד:** ניפוי באגים קשה; מאזני עומסים ברמת החיבור.
- **גבולות:** גודל לפני פענוח; שדות לא מוכרים נדחים; כל הבעיות מדווחות יחד.
- **שאלה פתוחה:** ___

<div dir="ltr" align="left">

```text
lint spec     docker run --rm -v "$PWD/api:/spec:ro" -w /spec redocly/cli:2.57.0 lint openapi.yaml
problem       application/problem+json  {type, title, status, detail, instance, errors[{detail, pointer}]}
cursor        base64url(json{v, t, id}); ORDER BY created_at DESC, id DESC; WHERE (created_at, id) < ($1, $2); LIMIT n+1
idempotency   same key + same body -> replay (Idempotent-Replayed: true); other body -> 422; running -> 409 + Retry-After
deprecation   Deprecation: @<unix>   Sunset: <HTTP-date>   Link: <...>; rel="successor-version"
buf           buf lint | buf generate | git status --short gen | buf breaking --against '../.git#branch=main,subdir=ch09'
grpc codes    400 InvalidArgument | 404 NotFound | 409 AlreadyExists | 504 DeadlineExceeded | 500 Internal
limits        http.MaxBytesReader before decoding; dec.DisallowUnknownFields(); second Decode must return io.EOF
mine          body limit ___ KiB   key TTL ___ h   tests passing ___ / 31
```

</div>

קישורים: [[08-python-async-fastapi]] · [[10-postgres-queries-indexes]]
