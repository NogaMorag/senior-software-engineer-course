---
direction: rtl
lang: he
title: "פרק 38: אבני הדרך"
chapter: 38
part: "חלק י: פרויקט הגמר"
---

<div dir="rtl" lang="he">

# פרק 38: אבני הדרך

הפרק הזה הוא מדריך הבנייה של Hookline: עשר אבני דרך, כל אחת בסוף החלק בספר שמלמד את מה שהיא צריכה. כל סעיף
הוא תדריך עבודה באותו מבנה: המטרה, המושגים והיכן נלמדו, הצעדים, הקבצים, הבדיקות שצריכות להוריק בפקודות מדויקות,
הפלט הצפוי, מתי נגמר, הטעויות הנפוצות, והרחבה. את הפרק לא קוראים ברצף: חוזרים לסעיף של אבן הדרך ביום שלה.

## מה נלמד בפרק

- להפוך את השרת מפרק 6 לשלד של API אמיתי: אימות, משאבים, עימוד ופרטי בעיה, מאחורי ממשקים.
- להחליף מאגר בזיכרון במאגר PostgreSQL בלי לשנות שורה ב־handlers, ולהוכיח את זה בבדיקות התאמה.
- לקלוט אירועים באופן אידמפוטנטי, כשאילוץ ייחודיות במסד מכריע בין בקשות מקבילות.
- לחבר לפרויקט את מה שנלמד בכל חלק: מטמון, outbox ו־Redpanda, תור עבודות, חוסן, קונטיינרים, תצפיתיות ואבטחה.
- לעבוד מול בדיקות שכבר כתובות, ולדעת מתי אבן דרך באמת נגמרה.

## ידע מקדים

- פרק 37 כולו: המפרט, הארכיטקטורה, מודל הנתונים, וההתקנה של ה־starter בסעיף 37.8.
- לאבן הדרך הראשונה: פרק 6 (השרת, ה־middleware, הלוגים והכיבוי) ופרק 9, סעיפים 9.2 עד 9.5 (OpenAPI, פרטי בעיה,
  עימוד ואידמפוטנטיות).
- לאבן הדרך השנייה: פרק 10, סעיפים 10.5 עד 10.7 (אינדקסים, עימוד מהיר ו־pgx), ופרק 11, סעיפים 11.1, 11.2, 11.6
  ו־11.7 (טרנזקציות, רמות בידוד, goose ומאגר חיבורים).
- כל אבן דרך מאוחרת מציינת בסעיף שלה את הפרקים שהיא נשענת עליהם.

## 38.1 M1: שלד ה־API

### 38.1.1 המטרה

‏API שעובד ונבדק, שתואם לחוזה שלו: דיירות שנוצרות בכלי המפעילה ומקבלות מפתח פעם אחת, אימות ב־Bearer, ‏endpoints
עם סוד חתימה, מנויים, וקליטת אירועים עם מפתח אידמפוטנטיות. הכול מעל מאגר בזיכרון שיושב מאחורי ממשקים, כך שבאבן
הדרך הבאה מחליפים אותו ב־PostgreSQL ולא נוגעים ב־handler אחד.

### 38.1.2 המושגים, והיכן נלמדו

| מושג | איפה בספר | איפה בפרויקט |
| --- | --- | --- |
| שרת עם זמנים קצובים וכיבוי מסודר | פרק 6, סעיף 6.5 | `internal/platform/server.go` (שלם ב־starter) |
| שרשרת middleware ומזהה בקשה בלוג | פרק 6, סעיפים 6.2 ו־6.4 | `internal/platform/middleware.go`, ‏`logging.go` |
| הגדרות מהסביבה, כל הבעיות בבת אחת | פרק 6, סעיף 6.3 | `internal/config/config.go` |
| ‏OpenAPI כמקור אמת | פרק 9, סעיף 9.2 | `api/openapi.yaml`, ‏`internal/api/spec_test.go` |
| פרטי בעיה (RFC 9457) | פרק 9, סעיף 9.3 | `internal/problem` |
| עימוד עם cursor אטום | פרק 9, סעיף 9.4 | `internal/page`, ‏`internal/api/decode.go` |
| מפתח אידמפוטנטיות | פרק 9, סעיף 9.5 | `internal/api/ingest.go`, ‏`internal/store/memory` |
| גבולות קלט לפני פענוח | פרק 9, סעיף 9.9 | `limitBody` ו־`decodeFrom` |
| ממשקים כגבול בין שכבות | פרק 4, סעיפים 4.2 ו־4.3 | `internal/store/store.go` |

### 38.1.3 הקבצים

| קובץ | מצב ב־starter | מה עושים |
| --- | --- | --- |
| `internal/store/store.go` | שלם | לקרוא: ארבעה ממשקים, וההבטחות שבתיעוד שלהם |
| `internal/store/storetest/storetest.go` | שלם | לקרוא: בדיקות ההתאמה שכל מאגר עובר |
| `internal/store/memory/memory.go` | גופים ריקים | לממש את כל המתודות |
| `internal/admin/admin.go` | גוף ריק | ‏`CreateTenant` |
| `internal/api/server.go` | שלם | לקרוא: הנתיבים וסדר ה־middleware |
| `internal/api/decode.go` | גופים ריקים | להעביר מהפתרון שלך לפרק 9 ולהתאים |
| `internal/api/auth.go` | גופים ריקים | ה־middleware של האימות ו־`GET /v1/tenant` |
| `internal/api/endpoints.go`, ‏`subscriptions.go`, ‏`ingest.go` | גופים ריקים | ה־handlers |
| `cmd/api/main.go`, ‏`cmd/hookline-admin/main.go` | שלם | לקרוא: החיווט |

### 38.1.4 הצעדים

**יום 22, אחרי פרק 37.** קראי את החוזה לפני הקוד. פתחי את `api/openapi.yaml` ועברי על כל נתיב: מה הוא מקבל, אילו
קודים הוא מחזיר, ואיזה סוג בעיה מגיע עם כל קוד. אחר כך בדקי שהוא תקין, באותה פקודה מפרק 9:

<div dir="ltr" align="left">

```bash
cd ~/senior-course/hookline
docker run --rm -v "$PWD/api:/spec" -w /spec redocly/cli:2.57.0 lint openapi.yaml
```

</div>

<div dir="ltr" align="left">

```text
validating openapi.yaml...
openapi.yaml: validated in 47ms

Woohoo! Your API description is valid. 🎉
1 problem is explicitly ignored.
```

</div>

הבעיה שמוחרגת היא היחידה שמותרת: לנתיב `/healthz` אין תשובת 4xx לתעד. הקובץ `api/.redocly.lint-ignore.yaml`
אומר את זה במפורש, כך שהכלל ממשיך לשמור על כל שאר הנתיבים.

עכשיו קראי את `internal/store/store.go`. ארבעת הממשקים הם הגבול: ה־handlers מכירים רק אותם, ושום handler לא יודע
אם מאחוריהם יש מפה או מסד. התיעוד של כל מתודה הוא חוזה. במיוחד המתודה של הקליטה:

<div dir="ltr" align="left">

```go
// Events are ingested once per (tenant, idempotency key).
type Events interface {
	// IngestEvent stores e unless the tenant already used e.IdempotencyKey. Then, with the same
	// BodyHash, it returns the stored event and replayed=true; with another BodyHash it returns
	// domain.ErrIdempotencyMismatch. The decision is atomic: concurrent calls with one key store
	// exactly one event and all return it. The PostgreSQL store also writes the outbox row (M4) in the
	// same transaction.
	IngestEvent(ctx context.Context, e domain.Event) (stored domain.Event, replayed bool, err error)
	GetEvent(ctx context.Context, tenantID, id string) (domain.Event, error)
}
```

</div>

התחילי מהמאגר בזיכרון, כי כל בדיקה של ה־API צריכה אותו. ממשי בקובץ `internal/store/memory/memory.go` שלוש מתודות:

- ‏`New`
- ‏`CreateTenant`
- ‏`TenantByKeyHash`

ובדקי:

<div dir="ltr" align="left">

```bash
go test -race -run 'TestConformance/TenantByKeyHash' ./internal/store/memory/
go test -race -run 'TestHealthNeedsNoKey' ./internal/api/
```

</div>

שתיהן צריכות להחזיר `ok`. הבדיקה של הבריאות עוברת כבר עכשיו, כי הנתיב `/healthz` מוגדר בתוך `server.go` עצמו.

**יום 23: המשאבים.** הסדר הזה מוריק בדיקה אחרי בדיקה:

| צעד | מה כותבים | הבדיקה |
| --- | --- | --- |
| 1 | ‏`admin.CreateTenant`: אימות קלט, `domain.NewAPIKey`, שמירה של הגיבוב בלבד | `TestCreateTenant` ב־`./internal/admin/` |
| 2 | ב־`decode.go`: ‏`limitBody`, ‏`isJSON`, ‏`decodeFrom`, ‏`readBody` מפרק 9; ‏`pageParams` ו־`writePage` | ייבדקו דרך הצעדים הבאים |
| 3 | ב־`auth.go`: הכותרת, צורת המפתח, החיפוש לפי גיבוב, והדיירת ב־context | `TestAuthentication`, ‏`TestGetTenant` |
| 4 | ‏endpoints במאגר וב־handlers: יצירה עם סוד, קריאה, רשימה, השבתה | `TestCreateEndpoint`, ‏`TestListEndpoints`, ‏`TestDisableEndpoint`, ‏`TestUnknownAndMalformedIDs` |
| 5 | מנויים במאגר וב־handlers | `TestSubscriptions`, ‏`TestCreateSubscriptionValidation`, ‏`TestListSubscriptions` |
| 6 | קליטה במאגר וב־handler | `TestIngest` |
| 7 | אין קוד חדש: הבדיקה מוכיחה את הבידוד | `TestTenantIsolation` |

כל שורה בטבלה נבדקת בפקודה אחת, עם הדפוס שבעמודה האחרונה. למשל, לצעד 4:

<div dir="ltr" align="left">

```bash
go test -race -run 'TestCreateEndpoint|TestListEndpoints|TestDisableEndpoint|TestUnknownAndMalformedIDs' ./internal/api/
go test -race -run 'TestConformance/Endpoint' ./internal/store/memory/
```

</div>

שלושה פרטים בצעד 3, ה־middleware של האימות, מהקובץ `internal/api/auth.go` בפתרון:

<div dir="ltr" align="left">

```go
		scheme, key, ok := strings.Cut(r.Header.Get("Authorization"), " ")
		if !ok || !strings.EqualFold(scheme, "Bearer") || !domain.LooksLikeAPIKey(key) {
			unauthorized(w, r, "send the header Authorization: Bearer followed by your API key")
			return
		}
		t, err := s.store.TenantByKeyHash(r.Context(), domain.HashAPIKey(key))
		if errors.Is(err, domain.ErrNotFound) {
			unauthorized(w, r, "the API key is not valid")
			return
		}
```

</div>

- **השם של הסכמה לא תלוי רישיות.** כך מגדיר את זה תקן ה־HTTP. לכן `bearer` חייב לעבוד כמו `Bearer`.
- **בדיקת הצורה באה לפני החיפוש.** ערך שאינו נראה כמו מפתח נדחה בלי לגעת במאגר, וזה זול.
- **‏401 תמיד עם `WWW-Authenticate`.** התקן דורש אותה, והבדיקה בודקת אותה.

בצעד 4 שימי לב לשני דברים. הסוד מוחזר רק בתשובת היצירה. בטיפוס `domain.Endpoint` השדה `Secret` מסומן
כך שלא יקודד, ותשובת היצירה עוטפת את הטיפוס בטיפוס שמוסיף אותו. ומזהה בנתיב נבדק ב־`domain.IsID` לפני כל קריאה למאגר: מזהה לא תקין
הוא 404, ובאבן הדרך השנייה הוא לא יגיע לעמודת `uuid` ויהפוך ל־500.

בצעד 6, המאגר בזיכרון צריך להחליט באופן אטומי אם המפתח חדש. מנעול אחד סביב הבדיקה וההכנסה עושה את מה שאילוץ
הייחודיות יעשה במסד. זה הקוד מהקובץ `internal/store/memory/memory.go`:

<div dir="ltr" align="left">

```go
func (s *Store) IngestEvent(_ context.Context, e domain.Event) (domain.Event, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	k := idemKey{e.TenantID, e.IdempotencyKey}
	if id, ok := s.byIdemKey[k]; ok {
		stored := s.events[id]
		if stored.BodyHash != e.BodyHash {
			return domain.Event{}, false, domain.ErrIdempotencyMismatch
		}
		return stored, true, nil
	}
	e.Payload = slices.Clone(e.Payload) // the caller's buffer is not ours to keep
	s.events[e.ID] = e
	s.byIdemKey[k] = e.ID
	return e, false, nil
}
```

</div>

ה־handler מחשב את הגיבוב על הבתים הגולמיים של הגוף, לפני הפענוח. כך „אותו גוף” פירושו אותם בתים בדיוק, כמו בפרק 9.
שימי לב גם שבניגוד ל־middleware של פרק 9, כאן אין מצב „בתהליך” ואין 409: התשובה נגזרת כולה מהאירוע השמור,
והאירוע נוצר באופן אטומי. בקשה מקבילה עם אותו מפתח מחכה ומקבלת את אותו אירוע.

לסיום, הריצי את השרת על המאגר בזיכרון ושלחי אליו בקשות אמיתיות:

<div dir="ltr" align="left">

```bash
make run-dev
```

</div>

השרת מדפיס פעם אחת מפתח של דיירת פיתוח. בחלון שני:

<div dir="ltr" align="left">

```bash
export KEY=hk_...   # the key that make run-dev printed
curl -s localhost:8080/v1/endpoints -H "Authorization: Bearer $KEY" \
  -H 'Content-Type: application/json' -d '{"url":"http://localhost:9000/hooks"}'
curl -s -i localhost:8080/v1/events -H "Authorization: Bearer $KEY" -H 'Content-Type: application/json' \
  -H 'Idempotency-Key: order-42' -d '{"type":"invoice.paid","payload":{"invoice":"inv_1","amount":4200}}'
```

</div>

הריצי את בקשת האירוע פעמיים, ואז פעם שלישית עם `"amount":9900`. הראשונה מחזירה 202, השנייה 202 עם
`Idempotent-Replayed: true` ואותו מזהה, והשלישית 422.

### 38.1.5 הבדיקות שצריכות להוריק

<div dir="ltr" align="left">

```bash
DATABASE_URL= go test -race -skip 'Invalidat|Deliver|Replay|ActiveSubscriptions|RateLimit|OverItsLimit|LimiterFailure|Shed|Trace|PrivateAddress' ./internal/api/ ./internal/store/... ./internal/admin/
```

</div>

ההשמה הריקה של `DATABASE_URL` מבטיחה שבדיקות ה־API רצות על המאגר בזיכרון, גם אם הגדרת את המשתנה בחלון אחר.
הדגל `-skip` מדלג על הבדיקות של אבני דרך מאוחרות שכבר נמצאות באותן חבילות: המטמון של אבן הדרך השלישית
(`Invalidat`, ‏`ActiveSubscriptions`), המסירות של החמישית (`Deliver`, ‏`Replay`), ההגנות של השישית (`RateLimit`, ‏`Shed` ועוד שתיים), והמעקב של השמינית (`Trace`).
כל אבן דרך מורידה ממנו את מה שהיא מממשת.

### 38.1.6 הפלט הצפוי

<div dir="ltr" align="left">

```text
ok  	example.com/hookline/internal/api	2.213s
ok  	example.com/hookline/internal/store	1.848s
ok  	example.com/hookline/internal/store/memory	2.257s
ok  	example.com/hookline/internal/store/postgres	3.498s
ok  	example.com/hookline/internal/store/storetest	2.641s
ok  	example.com/hookline/internal/admin	3.010s
```

</div>

הזמנים אצלך יהיו אחרים. החבילה `postgres` עוברת כי כל הבדיקות שלה מדלגות בלי מסד; ב־`-v` תראי `SKIP` לכל אחת.

### 38.1.7 קריטריון סיום

- כל הבדיקות בפקודה של סעיף 38.1.5 עוברות עם `-race`.
- מפתח לא מוכר, כתובת endpoint לא תקינה, וגוף JSON פגום מחזירים כל אחד את פרטי הבעיה שמתועדים בחוזה.
- ‏`TestTenantIsolation` עוברת: לכל נתיב, המזהים של דיירת א' עם המפתח של דיירת ב' מחזירים 404.
- ‏`TestSpecCoversEveryRoute` עוברת, והחוזה עובר את הבדיקה של redocly.
- קומיט עם ההודעה `M1: API skeleton`.

### 38.1.8 טעויות נפוצות

- **לבדוק בעלות ב־handler במקום במאגר.** אם המאגר מחזיר כל שורה, וה־handler הוא שמשווה דיירת, קל לשכוח
  את ההשוואה בנתיב החדש הבא. מזהה הדיירת הוא פרמטר של כל מתודה במאגר, והמאגר מסנן.
- **להחזיר `null` לרשימה ריקה.** פרוסה שלא אותחלה מקודדת כ־`null`. הבדיקה מצפה ל־`{"items":[]}`.
- **לגבב את הגוף אחרי הפענוח.** אז `{"a":1}` ו־`{ "a": 1 }` נחשבים לאותו גוף במאגר אחד ולא באחר. מגבבים בתים
  גולמיים, לפני הפענוח.
- **להחזיר את הסוד גם ב־`GET`.** הסוד מוצג פעם אחת. הבדיקה בודקת שהוא לא מופיע בקריאה ולא ברשימה.
- **שעון גלובלי בבדיקות.** הבדיקות מזריקות שעון דרך `api.Options.Now`; קוד שקורא ישירות ל־`time.Now` ייתן זמני
  יצירה זהים, ובדיקות העימוד יהיו תלויות במזל.

### 38.1.9 הרחבה

- הוסיפי `POST /v1/endpoints/{id}/enable`, קודם בחוזה ורק אחר כך בקוד, עם בדיקה. ‏`TestSpecCoversEveryRoute` תכשל
  עד ששניהם יתאימו.
- הוסיפי נתיב לסיבוב סוד החתימה, שמחזיר סוד חדש פעם אחת. חשבי מה קורה למסירות שכבר נחתמו בסוד הישן.

## 38.2 M2: PostgreSQL וקליטה אידמפוטנטית

### 38.2.1 המטרה

להעביר את המצב של Hookline מזיכרון ל־PostgreSQL. סכמה מלאה במיגרציות goose, מאגר pgx שמממש את אותם ממשקים
ועובר את אותן בדיקות התאמה, וקליטה שכותבת אירוע ושורת outbox בטרנזקציה אחת, כשאילוץ הייחודיות מכריע בין בקשות
מקבילות. בסוף, כל בדיקות ה־API מאבן הדרך הראשונה עוברות, בלי שינוי, מול מסד אמיתי.

### 38.2.2 המושגים, והיכן נלמדו

| מושג | איפה בספר | איפה בפרויקט |
| --- | --- | --- |
| ‏PostgreSQL בקונטיינר | פרק 10, סעיף 10.1 | `deploy/compose.yaml` |
| אינדקסים מרוכבים וחלקיים | פרק 10, סעיף 10.5 | המיגרציות בתיקייה `migrations/` |
| עימוד keyset ב־SQL | פרק 10, סעיף 10.6 | `ListEndpoints` ו־`ListSubscriptions` |
| ‏pgx ו־pgxpool | פרק 10, סעיף 10.7; פרק 11, סעיף 11.7 | `internal/db/db.go`, ‏`internal/store/postgres` |
| טרנזקציות ורמת Read committed | פרק 11, סעיפים 11.1 ו־11.2 | `IngestEvent` |
| מיגרציות goose מוטמעות | פרק 11, סעיף 11.6 | `migrations/migrations.go`, ‏`hookline-admin migrate` |
| מסד זמני לכל חבילת בדיקות | התרגיל של פרק 11 | `internal/testdb` |
| תבנית ה־outbox | פרק 13 (בהמשך) | הטבלה `outbox`, שנכתבת כבר עכשיו |

### 38.2.3 הקבצים

| קובץ | מצב ב־starter | מה עושים |
| --- | --- | --- |
| `deploy/compose.yaml` | שלם | להריץ; PostgreSQL 18 בפרויקט `hookline`, בפורט 5438 |
| `migrations/00001_tenants.sql` עד `00004_deliveries_attempts.sql` | שלמים | לקרוא כל אחד, ולהריץ עלייה וירידה |
| `internal/db/db.go`, ‏`internal/testdb/testdb.go` | שלמים | מפרק 11, עם שם המודול של הפרויקט |
| `internal/store/postgres/postgres.go` | גופים ריקים | כל המתודות |
| `internal/api/store_test.go` | שלם | לקרוא: הוא בוחר את המאגר לפי `DATABASE_URL` |

המיגרציות מגיעות שלמות, כי אבני הדרך הבאות נשענות על הסכמה המדויקת שלהן. העבודה שלך עליהן היא לקרוא ולהסביר:
לכל אילוץ, איזה באג הוא מונע. זו גם שאלה שתישאלי עליה.

### 38.2.4 הצעדים

**יום 29: הסכמה והמאגר.** העלי את המסד והריצי את המיגרציות:

<div dir="ltr" align="left">

```bash
cd ~/senior-course/hookline
make db-up
export DATABASE_URL=postgres://hookline:change-me@localhost:5438/hookline
go run ./cmd/hookline-admin migrate up
go run ./cmd/hookline-admin migrate down-to 0
go run ./cmd/hookline-admin migrate up
```

</div>

<div dir="ltr" align="left">

```text
schema version: 4
schema version: 0
schema version: 4
```

</div>

עכשיו קראי את ארבע המיגרציות, ומלאי לעצמך טבלה: אילוץ, והבאג שהוא מונע. שלושה שכדאי במיוחד להבין:

- **הייחודיות של הזוג `(tenant_id, id)` בטבלת ה־endpoints.** היא נראית מיותרת, כי המזהה כבר מפתח ראשי. הוא קיים כדי שהמנויים יוכלו
  להפנות לזוג, וכך המסד דוחה מנוי של דיירת אחת ל־endpoint של אחרת.
- **‏`UNIQUE (tenant_id, idempotency_key)` על `events`.** זה מה שמכריע בקליטה. בלעדיו, שתי בקשות מקבילות יעברו שתיהן
  כל בדיקה שתכתבי בקוד.
- **הייחודיות של המנויים מתחילה במזהה הדיירת.** ‏PostgreSQL בודק את האינדקס הייחודי לפני המפתח הזר. אילו המפתח
  היה רק endpoint וסוג, דיירת ב' שמנסה צירוף שכבר קיים אצל א' הייתה מקבלת 409 במקום 404, ולומדת שהמזהה קיים.
  הבאג הזה נתפס בבדיקות ההתאמה בזמן כתיבת הפתרון, וזו הסיבה שהן קיימות.
- **האינדקס החלקי `outbox_unpublished_idx`.** ה־relay של אבן הדרך הרביעית שואל רק על שורות שלא פורסמו. שורה שפורסמה
  יוצאת מהאינדקס, והוא נשאר קטן.

ממשי את המאגר ב־`internal/store/postgres/postgres.go`, מתודה אחרי מתודה, באותו סדר כמו באבן הדרך הראשונה. כל מתודה
מתרגמת שגיאת מסד לשגיאת תחום, לפי הקוד שלה:

| ‏SQLSTATE | שם | שגיאת התחום | איפה |
| --- | --- | --- | --- |
| `23505` | unique_violation | `domain.ErrConflict` | דיירת עם אותו גיבוב; מנוי כפול |
| `23503` | foreign_key_violation | `domain.ErrNotFound` | מנוי ל־endpoint שאינו של הדיירת |
| אין שורות | `pgx.ErrNoRows` | `domain.ErrNotFound` | כל קריאה של שורה אחת |

שאילתת הרשימה היא keyset מפרק 10, עם פרמטר שיכול להיות `NULL` לעמוד הראשון. זו השאילתה מהקובץ
`internal/store/postgres/postgres.go`:

<div dir="ltr" align="left">

```go
	rows, err := s.pool.Query(ctx, `
		SELECT `+endpointColumns+` FROM endpoints
		WHERE tenant_id = $1 AND ($2::timestamptz IS NULL OR (created_at, id) < ($2, $3::uuid))
		ORDER BY created_at DESC, id DESC
		LIMIT $4`, tenantID, at, id, limit+1)
```

</div>

ההשוואה בין הזוגות `(created_at, id)` תואמת את סדר האינדקס `endpoints_tenant_page_idx`, ולכן המסד קורא רק את שורות
העמוד, עמוק ככל שיהיה. מבקשים שורה אחת יותר מהגבול, כדי לדעת אם יש עמוד נוסף בלי שאילתת ספירה.

בדקי את המאגר מול בדיקות ההתאמה, ואז את כל בדיקות ה־API מול המסד:

<div dir="ltr" align="left">

```bash
go test -race -count=1 -run 'TestConformance/(Tenant|Endpoint|Subscription)' ./internal/store/postgres/
go test -race -count=1 -skip 'Invalidat|Deliver|Replay|ActiveSubscriptions|RateLimit|OverItsLimit|LimiterFailure|Shed|Trace|PrivateAddress' ./internal/api/
```

</div>

הפקודה השנייה לא השתנתה מאבן הדרך הראשונה. מה שהשתנה הוא שהמשתנה `DATABASE_URL` מוגדר, והקובץ
`internal/api/store_test.go` בוחר לכן את המאגר של PostgreSQL. זה כל ההבדל, וזו ההוכחה שהממשקים עשו את העבודה.

**יום 30: הקליטה.** המתודה `IngestEvent` היא לב אבן הדרך. היא רצה בטרנזקציה אחת, ברמת Read committed. זה הקוד
מהקובץ `internal/store/postgres/postgres.go`:

<div dir="ltr" align="left">

```go
	err := pgx.BeginTxFunc(ctx, s.pool, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, `
			INSERT INTO events (id, tenant_id, type, payload, idempotency_key, body_hash, created_at, trace_context)
			VALUES ($1, $2, $3, $4, $5, $6, $7, NULLIF($8, ''))
			ON CONFLICT (tenant_id, idempotency_key) DO NOTHING
			RETURNING `+eventColumns,
			e.ID, e.TenantID, e.Type, []byte(e.Payload), e.IdempotencyKey, e.BodyHash[:], e.CreatedAt, e.TraceContext)
		got, err := scanEvent(row)
		if err == nil {
			stored = got
			_, err = tx.Exec(ctx, `INSERT INTO outbox (event_id, created_at) VALUES ($1, $2)`, got.ID, got.CreatedAt)
			return err
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		stored, err = scanEvent(tx.QueryRow(ctx, `
			SELECT `+eventColumns+` FROM events WHERE tenant_id = $1 AND idempotency_key = $2`,
			e.TenantID, e.IdempotencyKey))
		if err != nil {
			return err
		}
		if stored.BodyHash != e.BodyHash {
			return domain.ErrIdempotencyMismatch
		}
		replayed = true
		return nil
	})
```

</div>

העמודה `trace_context` בקוד הזה נוספת רק באבן הדרך השמינית (סעיף 38.8). באבן הדרך השנייה ההכנסה כוללת שבע עמודות.

כך זה עובד כששתי בקשות עם אותו מפתח מגיעות יחד:

| רגע | בקשה א' | בקשה ב' |
| --- | --- | --- |
| 1 | ‏`INSERT` מצליח ומחזיק את ערך המפתח באינדקס הייחודי | ‏`INSERT` נתקל באותו ערך ומחכה לסוף הטרנזקציה של א' |
| 2 | ‏`INSERT` לשורת ה־outbox, ואז `COMMIT` | עדיין מחכה |
| 3 | מחזירה את האירוע | ‏`ON CONFLICT DO NOTHING`: אין שורה מוחזרת |
| 4 | | ‏`SELECT` חדש, בתמונת מצב חדשה, רואה את השורה של א'; הגיבובים זהים, ולכן החזרה חוזרת |

ב־Read committed כל פקודה מקבלת תמונת מצב משלה (פרק 11, סעיף 11.2), ולכן הקריאה של ב' רואה את מה ש־א' אישרה. אם
א' הייתה מבטלת, ה־`INSERT` של ב' היה ממשיך ומצליח. אין כאן ניסיון חוזר ואין `Serializable`: האילוץ מספיק.

הגוף של האירוע נשמר כ־`jsonb`, והמסד מחזיר אותו בצורה מנורמלת: רווח אחרי נקודתיים, וסדר מפתחות משלו. בגלל זה
הבדיקות משוות JSON לפי ערך ולא לפי בתים. שתי התשובות של בקשות מקבילות עדיין זהות, כי שתיהן נקראות מאותה שורה.

הריצי את בדיקות הקליטה והמקביליות:

<div dir="ltr" align="left">

```bash
go test -race -count=1 -run 'TestConformance/Ingest|TestConcurrentIngest|TestMismatch|TestListEndpointsUsesTheIndex' ./internal/store/postgres/
go test -race -count=1 -run 'TestIngest' ./internal/api/
```

</div>

הבדיקה `TestConcurrentIngestLeavesOneEventAndOneOutboxRow` שולחת 24 קליטות במקביל עם מפתח אחד, וסופרת במסד
עצמו: שורת אירוע אחת ושורת outbox אחת. הבדיקה `TestListEndpointsUsesTheIndex` מריצה `EXPLAIN` ובודקת שהתוכנית
משתמשת באינדקס העימוד.

לבסוף, צרי דיירת אמיתית בכלי המפעילה, והריצי את השרת מול המסד:

<div dir="ltr" align="left">

```bash
go run ./cmd/hookline-admin create-tenant -name acme
make run
```

</div>

<div dir="ltr" align="left">

```text
tenant id: 01a10b63-...
api key:   hk_...
The key is shown once and only its hash is stored. Keep it now.
```

</div>

חזרי על בקשות ה־curl מסעיף 38.1.4 עם המפתח החדש. אחר כך בדקי ב־psql שהשורות נמצאות בטבלת האירועים
ובטבלת ה־outbox.

### 38.2.5 הבדיקות שצריכות להוריק

<div dir="ltr" align="left">

```bash
make db-up
export DATABASE_URL=postgres://hookline:change-me@localhost:5438/hookline
go test -race -count=1 -skip 'Invalidat|Deliver|Replay|ActiveSubscriptions|RateLimit|OverItsLimit|LimiterFailure|Shed|Trace|PrivateAddress' ./internal/api/ ./internal/store/... ./internal/admin/ ./internal/db/ ./migrations/
```

</div>

### 38.2.6 הפלט הצפוי

<div dir="ltr" align="left">

```text
ok  	example.com/hookline/internal/admin	1.534s
ok  	example.com/hookline/internal/api	2.040s
ok  	example.com/hookline/internal/db	2.815s
ok  	example.com/hookline/internal/store	3.395s
ok  	example.com/hookline/internal/store/memory	3.646s
ok  	example.com/hookline/internal/store/postgres	4.074s
ok  	example.com/hookline/internal/store/storetest	3.183s
ok  	example.com/hookline/migrations	3.585s
```

</div>

בהרצה עם `-v`, אף בדיקה של אבן הדרך לא מדלגת: כל מה שדילג באבן הדרך הראשונה רץ עכשיו מול המסד.

### 38.2.7 קריטריון סיום

- המיגרציות עולות על מסד ריק, יורדות עד גרסה 0 ועולות שוב (`TestMigrationsRoundTrip`).
- בדיקות ההתאמה עוברות על שני המאגרים, ובדיקות ה־API עוברות מול PostgreSQL בלי שינוי.
- בדיקת המקביליות מוצאת שורת אירוע אחת ושורת outbox אחת, ותשובות זהות לכל הבקשות.
- את יודעת להסביר למה האילוץ, ולא קוד, מכריע בכפילות, ולמה האירוע וה־outbox באותה טרנזקציה.
- ‏`make db-down` בסוף היום, וקומיט של אבן הדרך:

<div dir="ltr" align="left">

```bash
git commit -m "M2: PostgreSQL and idempotent ingest"
```

</div>

### 38.2.8 טעויות נפוצות

- **קריאה ואז הכנסה.** בדיקה בקוד אם המפתח קיים, ואז הכנסה, נשברת בבקשות מקבילות: שתיהן לא מוצאות, ושתיהן
  מכניסות. אחת נכשלת באילוץ ומקבלת 500. מתחילים בהכנסה עם `ON CONFLICT DO NOTHING`.
- **שורת ה־outbox מחוץ לטרנזקציה.** אם היא נכתבת אחרי ה־`COMMIT`, קריסה ביניהם משאירה אירוע שלא יימסר לעולם.
- **זמנים באזור הזמן המקומי.** ‏pgx מחזיר `timestamptz` בזמן המקומי של המכונה, והתשובה תשתנה בין מכונות. ממירים
  ל־UTC בסריקה. ושומרים מראש זמנים חתוכים למיקרו־שנייה, כי זה הדיוק שהמסד שומר; ‏`domain.Now` עושה את זה.
- **מזהה לא תקין שמגיע לעמודת `uuid`.** הוא נכשל בקוד `22P02` ונהיה 500. ה־handlers בודקים `domain.IsID` קודם.
- **לשכוח `-count=1`.** ‏Go שומר תוצאות בדיקה, ולא יודע שהמסד השתנה. בלי הדגל, „ok” יכול להיות מהרצה קודמת.

### 38.2.9 הרחבה

- הוסיפי זמן חיים למפתחות אידמפוטנטיות: אחרי 24 שעות, אותו מפתח יוצר אירוע חדש. חשבי איך עושים את זה בלי למחוק
  אירועים, ומה זה עושה לאילוץ הייחודיות.
- הריצי `EXPLAIN ANALYZE` על שאילתת הרשימה עם מאה אלף endpoints לדיירת אחת, בעמוד הראשון ובעמוד העמוק, ורשמי את
  הזמנים שמדדת.

## 38.3 M3: מטמון והמקבל המדומה

### 38.3.1 המטרה

להוסיף לפרויקט את שני הרכיבים שהמסירה תצטרך: מטמון של מנויים ב־Redis, שה־dispatcher ישאל דרכו „מי מקבל את
האירוע הזה?”, וה־sink, המקבל המדומה מפרק 8, שרץ ב־compose ומקבל את ה־webhooks. בסוף אבן הדרך כל התשתית
המקומית עולה בפקודה אחת, וה־API מודיע למטמון על כל שינוי במנויים.

### 38.3.2 המושגים, והיכן נלמדו

| מושג | איפה בספר | איפה בפרויקט |
| --- | --- | --- |
| ‏cache-aside | פרק 12, סעיף 12.2 | `internal/cache/cache.go` |
| זמן חיים עם פיזור אקראי, ואי־תוקף אחרי כתיבה | פרק 12, סעיף 12.3 | `internal/cache/ttl.go`, ‏`subcache.Invalidate` |
| התפרצות מטמון ו־singleflight | פרק 12, סעיף 12.4 | `cache.Get`, ‏`TestStampedeLoadsOnce` |
| מתי לא לשים מטמון | פרק 12, סעיף 12.6 | ה־API לא קורא מהמטמון, רק מוחק ממנו |
| המקבל המדומה: חתימה, סינון כפילויות, תקלות לפי פקודה | פרק 8, סעיף 8.8 | `sink/` |
| ‏worker אחד, כי המצב בזיכרון | פרק 8, סעיף 8.7 | `sink/Dockerfile` |

### 38.3.3 הקבצים

| קובץ | מצב ב־starter | מה עושים |
| --- | --- | --- |
| `internal/cache/*.go` | גופים ריקים | להעביר מהפתרון שלך לפרק 12, עם שם המודול של הפרויקט |
| `internal/subcache/subcache.go` | גופים ריקים | ‏`Endpoints`, ‏`Invalidate`, ‏`Key` ו־`Disabled` |
| `internal/store/*` | גופים ריקים | ‏`ActiveSubscriptions` בשני המאגרים |
| `internal/api/server.go` | שלם | לקרוא: האפשרות `Subscriptions` והפונקציה `changed` |
| `internal/wire/wire.go` | שלם | בונה את המטמון מעל Redis, או בלי מטמון כשאין `REDIS_ADDR` |
| `sink/` | גופים ריקים ב־Python | להעביר מהפתרון שלך לפרק 8; ‏`settings.py` שלם |
| `sink/Dockerfile`, ‏`deploy/compose.yaml` | שלמים | לקרוא ולהריץ |

### 38.3.4 הצעדים

**יום 33: המטמון.** התחילי מהחבילה `internal/cache`: העתיקי את ארבעת הקבצים מהתרגיל של פרק 12, והחליפי את
שם המודול:

<div dir="ltr" align="left">

```bash
cd ~/senior-course/hookline
cp ~/senior-course/ch12/internal/cache/*.go internal/cache/
sed -i 's#example.com/ch12-cache-aside#example.com/hookline#g' internal/cache/*.go
DATABASE_URL= REDIS_ADDR= go test -race ./internal/cache/
```

</div>

> **ב־Mac:** ‏`sed -i` דורש שם קובץ גיבוי. כתבי `sed -i ''` במקום `sed -i`.

הבדיקות שצריכות Redis מדלגות עכשיו, ושתי בדיקות היחידה (הפיזור האקראי, ו־Redis שנפל) עוברות.

עכשיו ההחלטה החשובה של אבן הדרך: **מה המפתח של המטמון.** ה־dispatcher שואל על זוג של דיירת וסוג אירוע. הצעה
טבעית היא רשומה לכל זוג כזה. הבעיה היא אי־התוקף: מנוי לכל הסוגים (`*`) משנה את התשובה של כל סוג, ואין דרך לדעת
אילו רשומות למחוק. לכן הפרויקט שומר **רשומה אחת לכל דיירת**, עם כל המנויים הפעילים שלה, ומסנן לפי הסוג ב־Go.
אי־התוקף נעשה פשוט: כל שינוי מוחק רשומה אחת. זו הפונקציה מהקובץ `internal/subcache/subcache.go`:

<div dir="ltr" align="left">

```go
func (c *Cache) Endpoints(ctx context.Context, tenantID, eventType string) ([]string, error) {
	subs, err := c.c.Get(ctx, Key(tenantID), func(ctx context.Context) ([]domain.Subscription, error) {
		return c.loader.ActiveSubscriptions(ctx, tenantID)
	})
	if err != nil {
		return nil, err
	}
	var ids []string
	for _, s := range subs {
		if s.EventType == eventType || s.EventType == domain.AllTypes {
			ids = append(ids, s.EndpointID)
		}
	}
	slices.Sort(ids)
	return slices.Compact(ids), nil
}
```

</div>

שלושה דברים בקוד הזה:

- **רק endpoints פעילים.** ‏`ActiveSubscriptions` מצטרפת לטבלת ה־endpoints ומשאירה רק את אלה שלא הושבתו, ולכן
  השבתה היא גם שינוי שמוחק את הרשומה.
- **כל endpoint פעם אחת.** endpoint שמנוי גם לסוג מדויק וגם ל־`*` צריך מסירה אחת, ולא שתיים. המיון וההסרה של
  כפילויות עושים את זה.
- **מטמון שנפל לא מפיל את ה־dispatcher.** ‏`cache.Get` מפרק 12 חוזר למקור כשהקריאה מהמטמון נכשלת. זו החלטה
  מהמפרט (פרק 37, סעיף 37.6): המטמון חוזר ל־PostgreSQL.

הוסיפי את `ActiveSubscriptions` לשני המאגרים, ובדקי מול בדיקות ההתאמה ובדיקות המטמון:

<div dir="ltr" align="left">

```bash
DATABASE_URL= REDIS_ADDR= go test -race -run 'TestConformance/ActiveSubscriptions' ./internal/store/memory/
DATABASE_URL= REDIS_ADDR= go test -race ./internal/subcache/
```

</div>

אחר כך חברי את ה־API. המטמון לא נקרא ב־API בכלל; ה־API רק מודיע לו, **אחרי** שהשינוי נשמר. זה הקוד מהקובץ
`internal/api/server.go`:

<div dir="ltr" align="left">

```go
func (s *server) changed(r *http.Request, tenantID string) {
	if s.subs == nil {
		return
	}
	if err := s.subs.Invalidate(r.Context(), tenantID); err != nil {
		s.log.WarnContext(r.Context(), "subscription cache invalidation failed; the entry expires by TTL",
			slog.String("tenant_id", tenantID), slog.String("error", err.Error()))
	}
}
```

</div>

הסדר חשוב, וזה המרוץ מפרק 12, סעיף 12.3. אם מוחקים לפני ה־`COMMIT`, קורא שמגיע ביניהם טוען את הערך הישן ומחזיר
אותו למטמון. ואם המחיקה נכשלת, הבקשה עדיין מצליחה, כי השינוי כבר נשמר. זמן החיים, דקה ועוד עד עשר שניות של
פיזור, הוא הגבול לכמה זמן ה־dispatcher עלול לפעול לפי רשימה ישנה. קראי לפונקציה בשלושה מקומות: אחרי מנוי חדש,
אחרי מחיקת מנוי, ואחרי השבתת endpoint. הבדיקה `TestChangesInvalidateTheSubscriptionCache` סופרת אותם.

**יום 34: ה־sink וה־compose.** העתיקי את המקבל מהפתרון שלך לפרק 8 לתיקייה `sink/`, ואת קובץ וקטורי החתימה
לתיקייה `sink/data/`. ה־Go signer של אבן הדרך החמישית יקרא את אותו קובץ:

<div dir="ltr" align="left">

```bash
cp -r ~/senior-course/ch08/src/receiver/. sink/src/receiver/
cp $COURSE/exercises/ch08-receiver/data/signature-vectors.json sink/data/
cd sink && uv sync && uv run pytest -q && cd ..
```

</div>

עכשיו קראי את `sink/Dockerfile` ואת `deploy/compose.yaml`, והעלי את כל הסביבה:

<div dir="ltr" align="left">

```bash
make up
docker compose -p hookline -f deploy/compose.yaml ps
curl -s localhost:8091/stats
```

</div>

הפקודה `make up` בונה את ה־image של ה־sink ומעלה חמישה שירותים:

| שירות | פורט מקומי |
| --- | --- |
| PostgreSQL | 5438 |
| Redis | 6438 |
| Redpanda | 19438 |
| הקונסולה של Redpanda | 8098 |
| ה־sink | 8091 |

 ‏Redpanda נכנס כבר עכשיו כדי שה־compose לא ישתנה באבן הדרך הבאה. שתי החלטות בקובץ:

- **ה־sink מקבל סוד אחד לכל ה־hooks** (`RECEIVER_DEFAULT_SECRET`). בבדיקות, כל endpoint מקבל את הסוד הזה, כדי
  שה־sink יוכל לאמת. מקבל אמיתי מחזיק סוד לכל endpoint.
- **‏Redis הוא מטמון ולא מסד.** אין שמירה לדיסק, יש מגבלת זיכרון ומדיניות פינוי, בדיוק כמו בפרק 12, סעיף 12.1.

### 38.3.5 הבדיקות שצריכות להוריק

<div dir="ltr" align="left">

```bash
make up
go test -race -count=1 -skip 'Deliver|Replay|RateLimit|OverItsLimit|LimiterFailure|Shed|Trace|PrivateAddress' ./internal/cache/ ./internal/subcache/ ./internal/api/ ./internal/store/...
make sink-test
```

</div>

### 38.3.6 הפלט הצפוי

<div dir="ltr" align="left">

```text
ok  	example.com/hookline/internal/api	2.040s
ok  	example.com/hookline/internal/cache	3.612s
ok  	example.com/hookline/internal/store	3.395s
ok  	example.com/hookline/internal/store/memory	3.646s
ok  	example.com/hookline/internal/store/postgres	4.074s
ok  	example.com/hookline/internal/store/storetest	3.183s
ok  	example.com/hookline/internal/subcache	3.270s
...
62 passed in 1.63s
```

</div>

### 38.3.7 קריטריון סיום

- בדיקות המטמון עוברות מול Redis, כולל ההתפרצות: ארבעים קריאות מקבילות טוענות מהמקור פעם אחת.
- ‏`TestChangesInvalidateTheSubscriptionCache` עוברת: שלושה שינויים, שלוש מחיקות, ואף מחיקה על קריאה או על כתיבה שנכשלה.
- ‏`make sink-test` נקי, כולל הבדיקה הסטטית של הטיפוסים.
- ‏`make up` מעלה את כל השירותים במצב בריא, ו־`curl localhost:8091/stats` מחזיר מונים באפס.
- קומיט של אבן הדרך, ו־`make down` בסוף היום.

### 38.3.8 טעויות נפוצות

- **רשומה לכל סוג אירוע.** אז מנוי ל־`*` לא יודע אילו רשומות למחוק, והמסירה ממשיכה לפי רשימה ישנה עד שהזמן נגמר.
- **מחיקה לפני השמירה.** קורא שמגיע בין המחיקה ל־`COMMIT` מחזיר את הערך הישן למטמון, ושום דבר לא מוחק אותו שוב.
- **להחזיר שגיאה כשהמחיקה נכשלת.** הלקוח ינסה שוב שינוי שכבר נשמר, ויקבל 409. השינוי נשמר; המחיקה רק מקצרת את ההמתנה.
- **שני workers ל־sink.** כל תהליך סופר בנפרד, והמונים ב־`/stats` יחצו לשניים. ה־Dockerfile מריץ worker אחד.
- **להשכיח את `-count=1`.** כמו באבן הדרך השנייה: Go לא יודע ש־Redis השתנה, ומחזיר תוצאה שמורה.

### 38.3.9 הרחבה

- הוסיפי מטמון שלילי (פרק 12, סעיף 12.3): לדיירת בלי מנויים בכלל, שמרי רשימה ריקה עם זמן חיים קצר יותר.
- חשבי מה היה קורה עם שני API replicas ושני dispatchers: האם המחיקה מגיעה לכולם? (כן, כי Redis משותף. ומה עם
  מטמון בזיכרון של כל תהליך?)

## 38.4 M4: outbox ו־Redpanda

### 38.4.1 המטרה

לקחת כל אירוע שנקלט אל שורות מסירה, דרך Redpanda. ה־relay מפרסם את שורות ה־outbox, שה־API כותב מאז אבן הדרך
השנייה, ל־topic בשם `hookline.events`. ה־dispatcher צורך אותו, ויוצר שורת מסירה לכל יעד מתאים. קריסה בכל
נקודה לא מאבדת אירוע, וגם לא יוצרת מסירה כפולה.

### 38.4.2 המושגים, והיכן נלמדו

| מושג | איפה בספר | איפה בפרויקט |
| --- | --- | --- |
| ‏Redpanda בקונטיינר | פרק 13, סעיף 13.2 | `deploy/compose.yaml` |
| מחיצות ומפתח לפי דיירת | פרק 13, סעיפים 13.3 ו־13.4 | `bus.KafkaPublisher` |
| ‏consumer group ו־offsets | פרק 13, סעיף 13.5 | `bus.NewGroupConsumer` |
| אישור offset רק אחרי עיבוד | פרק 13, סעיף 13.6 | `dispatch.Run` |
| תבנית ה־outbox ונקודת הקריסה | פרק 13, סעיף 13.7 | `outbox.Relay`, ‏`AfterPublish` |
| למה „בדיוק פעם אחת” הוא חוזה | פרק 13, סעיף 13.8 | האילוץ `UNIQUE (event_id, endpoint_id)` |
| ‏franz-go | פרק 13, סעיף 13.9 | `internal/bus` |

### 38.4.3 הקבצים

| קובץ | מצב ב־starter | מה עושים |
| --- | --- | --- |
| `internal/bus/bus.go` | שלם | מפרק 13: ה־producer, הצרכן ו־`KafkaPublisher`, ועכשיו גם `CreateTopic` |
| `internal/outbox/outbox.go` | גופים ריקים | ‏`Message`, ‏`RunOnce`, ‏`Run`, ‏`Pending` |
| `internal/dispatch/dispatch.go` | גופים ריקים | ‏`Fanout`, ‏`Handle`, ‏`Run` |
| `cmd/relay`, ‏`cmd/dispatcher` | שלמים | לקרוא: החיווט |
| `internal/testenv` | שלם | ‏topic פרטי לכל בדיקה, ונתוני בדיקה |

### 38.4.4 הצעדים

**יום 39: ה־relay.** ה־outbox של Hookline שונה מזה של פרק 13 בפרט אחד: השורה לא מחזיקה את ההודעה, רק את מזהה
האירוע. ה־relay מצרף את טבלת האירועים ומפרסם הודעה קטנה: מזהה, דיירת וסוג. את הגוף ה־worker יקרא מהמסד כשישלח.
כך ה־topic נשאר קטן, ואין שני עותקים של הגוף שעלולים להיות שונים. זו הקריאה מהקובץ `internal/outbox/outbox.go`:

<div dir="ltr" align="left">

```go
		rows, err := tx.Query(ctx, `
			SELECT o.id, e.id, e.tenant_id, e.type
			FROM outbox o JOIN events e ON e.id = o.event_id
			WHERE o.published_at IS NULL
			ORDER BY o.id
			LIMIT $1
			FOR UPDATE OF o`, r.batch())
```

</div>

‏`FOR UPDATE OF o` נועל רק את שורות ה־outbox, ולא את האירועים. שאר `RunOnce` הוא הקוד של פרק 13 כמעט כמו שהוא:
פרסום, הנקודה `AfterPublish` שבה הבדיקות „הורגות” את ה־relay, ואז סימון באותה טרנזקציה. הלולאה `Run` מוסיפה
דבר אחד: כשהפרסום נכשל, כי Redpanda למטה, היא מחכה זמן שמכפיל את עצמו, והקליטה ממשיכה לעבוד. האירועים מחכים
ב־outbox. זו בדיוק השורה האחרונה בטבלת הכשלים בפרק 37, סעיף 37.6.

<div dir="ltr" align="left">

```bash
go test -race -count=1 -skip 'StaleRelay|Trace|PrivateAddress' ./internal/outbox/
```

</div>

**יום 40: ה־dispatcher.** ה־dispatcher עושה שני דברים לכל רשומה: שואל את המטמון של אבן הדרך הקודמת מי מקבל, ומכניס
שורת מסירה לכל אחד. זה הקוד מהקובץ `internal/dispatch/dispatch.go`:

<div dir="ltr" align="left">

```go
	tag, err := db.Exec(ctx, `
		INSERT INTO deliveries (id, tenant_id, event_id, endpoint_id)
		SELECT uuidv7(), $1, $2, endpoint_id FROM unnest($3::uuid[]) AS endpoint_id
		ON CONFLICT (event_id, endpoint_id) DO NOTHING`, m.TenantID, m.EventID, endpointIDs)
```

</div>

- **פקודה אחת לכל האירוע.** כל השורות של אירוע נכנסות יחד, או אף אחת. אין מצב של חצי פיזור.
- **‏`ON CONFLICT DO NOTHING`.** פרסום כפול מה־relay, או רשומה שחוזרת אחרי קריסה לפני אישור ה־offset, לא מוסיפים
  שורה. זו ההגנה של שתי שורות בטבלת הכשלים.
- **הפונקציה `uuidv7()` מובנית ב־PostgreSQL 18.** היא נותנת את אותו סוג מזהים ש־`domain.NewID` נותן.

הלולאה `dispatch.Run` היא `consumer.Run` מפרק 13 עם שינוי אחד: כשהטיפול ברשומה נכשל (המסד למטה), היא מנסה שוב
**במקום**, עם המתנה גדלה, ולא מדלגת. דילוג היה מאשר offset של אירוע שלא נפרש, ומאבד אותו. רשומה שאי אפשר לפענח
היא היוצא מן הכלל: היא נרשמת בלוג ומדולגת, כי ניסיון חוזר לא יתקן אותה, והוא היה חוסם את המחיצה לתמיד.

<div dir="ltr" align="left">

```bash
go test -race -count=1 ./internal/dispatch/ ./internal/bus/
```

</div>

הבדיקה `TestRelayToDispatcherThroughRedpanda` היא לב אבן הדרך. היא קולטת עשרה אירועים, „הורגת” את ה־relay אחרי
הפרסום ולפני הסימון, ומריצה relay שני, כך שכל אירוע נמצא ב־topic פעמיים. אחר כך היא מריצה dispatcher ובודקת
שבסוף יש בדיוק שתי מסירות לכל אירוע: אחת לכל endpoint מתאים.

לסיום, הריצי את שני הבינאריים האמיתיים:

<div dir="ltr" align="left">

```bash
go run ./cmd/hookline-admin create-topic -partitions 6
make relay        # in a second terminal
make dispatcher   # in a third terminal
```

</div>

שלחי אירוע עם curl, ופתחי את הקונסולה של Redpanda בכתובת `http://localhost:8098`. ההודעה מופיעה ב־topic בשם
`hookline.events`, והמפתח שלה הוא מזהה הדיירת. ב־psql תראי שורה חדשה בטבלה `deliveries` במצב `pending`.

### 38.4.5 הבדיקות שצריכות להוריק

<div dir="ltr" align="left">

```bash
make up
go test -race -count=1 -skip 'StaleRelay|Trace|PrivateAddress' ./internal/outbox/ ./internal/dispatch/ ./internal/bus/
```

</div>

### 38.4.6 הפלט הצפוי

<div dir="ltr" align="left">

```text
ok  	example.com/hookline/internal/bus	2.656s
ok  	example.com/hookline/internal/dispatch	4.349s
ok  	example.com/hookline/internal/outbox	4.776s
```

</div>

### 38.4.7 קריטריון סיום

- בדיקת נקודת הקריסה עוברת: אחרי קריסה בין פרסום לסימון, כל אירוע מתפרסם שוב, ובכל זאת יש מסירה אחת לכל endpoint.
- כישלון בפרסום לא מסמן אף שורה, וה־relay ממשיך מעצמו כשה־broker חוזר.
- אירועים של דיירת אחת יוצאים לאותה מחיצה, לפי הסדר.
- ‏endpoint מושבת, או כזה שמנוי לסוג אחר, לא מקבל שורת מסירה.
- קומיט של אבן הדרך.

### 38.4.8 טעויות נפוצות

- **לסמן לפני שה־broker אישר.** אם הסימון קודם לאישור, קריסה ביניהם מאבדת את ההודעה. מפרסמים, מחכים לאישור,
  ורק אז מסמנים.
- **לאשר offset לפני ההכנסה.** קריסה אחרי האישור ולפני ההכנסה מאבדת את האירוע. ההכנסה קודמת, האישור אחריה.
- **לדלג על רשומה כשהמסד למטה.** אז ה־offset שאחריה מאושר, והאירוע נעלם. מנסים שוב במקום.
- **מפתח הודעה ריק.** בלי מפתח, ה־partitioner מפזר, ואירועים של אותה דיירת יוצאים מסדר. המפתח הוא מזהה הדיירת.
- **ה־relay בלי `FOR UPDATE`.** שני relays יפרסמו את אותן שורות. גם עם הנעילה, אבן הדרך השישית משאירה רק אחד פעיל.

### 38.4.9 הרחבה

- הוסיפי מדד של גיל השורה הוותיקה ב־outbox, הבסיס של ההתראה באבן הדרך השמינית.
- נסי להריץ שני dispatchers עם אותה קבוצה, ולראות ב־`rpk group describe` איך המחיצות מתחלקות ביניהם.

## 38.5 M5: מסירה, חתימה וניסיונות חוזרים

### 38.5.1 המטרה

להשלים את ההבטחה המרכזית של Hookline: כל שורת מסירה נשלחת למקבל כבקשת POST חתומה, עם ארבע הכותרות. כשל זמני
מנוסה שוב עם המתנה גדלה. כשל סופי, או סוף הניסיונות, הופכים את המסירה למתה. מסירה מתה אפשר להריץ שוב דרך ה־API,
וכל ניסיון גלוי בהיסטוריה. בסוף, בדיקה מקצה לקצה עוברת מול ה־sink: כל אירוע מגיע לכל מקבל מתאים.

### 38.5.2 המושגים, והיכן נלמדו

| מושג | איפה בספר | איפה בפרויקט |
| --- | --- | --- |
| תור על PostgreSQL עם `SKIP LOCKED` וחכירה | פרק 14, סעיף 14.2 | `delivery.Claim` |
| המתנה גדלה מעריכית עם פיזור אקראי | פרק 14, סעיף 14.3 | `internal/backoff` |
| אידמפוטנטיות אצל הצרכן | פרק 14, סעיף 14.4 | ה־sink מסנן לפי `Hookline-Event-Id` |
| תור מתים והפעלה חוזרת | פרק 14, סעיף 14.5 | `ReplayDelivery`, ‏`POST /v1/deliveries/{id}/replay` |
| כיבוי מסודר של workers | פרק 14, סעיף 14.7 | `delivery.Pool.RunOne` |
| חתימת HMAC ווקטורי בדיקה משותפים | פרק 8, סעיף 8.8 | `delivery.Sign`, ‏`sink/data/signature-vectors.json` |
| מספור ניסיונות אחרי הפעלה חוזרת | פרק 29, סעיף 29.5 | העמודה `attempt_base` |

### 38.5.3 הקבצים

| קובץ | מצב ב־starter | מה עושים |
| --- | --- | --- |
| `internal/backoff/backoff.go` | גופים ריקים | להעביר מהפתרון שלך לפרק 14 |
| `internal/delivery/signer.go` | גוף ריק | ‏`Sign` |
| `internal/delivery/send.go` | גופים ריקים | ‏`Body`, ‏`Send`, ‏`Classify` |
| `internal/delivery/queue.go` | גופים ריקים | ‏`Claim`, ‏`Complete`, ‏`Fail` |
| `internal/delivery/worker.go` | גופים ריקים | ‏`Pool.Run` ו־`RunOne` |
| `migrations/00005_deliveries_attempt_base.sql` | שלם | לקרוא: העמודה היחידה שנוספה מאז אבן הדרך השנייה |
| `internal/api/deliveries.go`, ‏`internal/store/*` | גופים ריקים | שלושת הנתיבים, ומתודות המסירות במאגרים |
| `internal/e2e/e2e_test.go` | שלם | הבדיקה מקצה לקצה |

### 38.5.4 הצעדים

**יום 43: חתימה, שליחה ותור.** התחילי מהחתימה, כי יש לה מפרט מדויק בקובץ: הבדיקה `TestSignMatchesSharedVectors`
קוראת את אותם וקטורים שה־sink ב־Python בודק. זו הפונקציה מהקובץ `internal/delivery/signer.go`:

<div dir="ltr" align="left">

```go
func Sign(secret string, ts int64, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	t := strconv.FormatInt(ts, 10)
	mac.Write([]byte(t))
	mac.Write([]byte{'.'})
	mac.Write(body)
	return "t=" + t + ",v1=" + hex.EncodeToString(mac.Sum(nil))
}
```

</div>

עכשיו השליחה והסיווג. ‏`Classify` היא המדיניות של דרישה F6 בשש שורות:

<div dir="ltr" align="left">

```go
func Classify(status int) (retryable bool, err error) {
	switch {
	case status >= 200 && status < 300:
		return false, nil
	case status == http.StatusRequestTimeout || status == http.StatusTooManyRequests || status >= 500:
		return true, fmt.Errorf("receiver answered %d", status)
	default:
		return false, fmt.Errorf("receiver answered %d; not retried", status)
	}
}
```

</div>

- **‏4xx הוא תשובה סופית**, חוץ מ־408 ו־429. מקבל שאומר 400 או 404 לא ישנה את דעתו בניסיון הבא.
- **‏3xx הוא גם תשובה סופית.** הלקוח של `NewClient` לא עוקב אחרי הפניות, כי הפניה הייתה שולחת את הגוף החתום
  לכתובת שהדיירת לא רשמה.
- **שגיאת רשת וזמן קצוב הם זמניים.** המקבל עלול לחזור.

<div dir="ltr" align="left">

```bash
DATABASE_URL= go test -race -run 'Vectors|Classify|Send' -skip 'PrivateAddress' ./internal/delivery/
go test -race ./internal/backoff/
```

</div>

הבדיקה `TestPolicyProperties` היא בדיקת תכונה עם הספרייה rapid, כמו שדרישה F6 מבקשת: לכל מדיניות ולכל
מספר ניסיון, ההמתנה בטווח, התקרה לא עוברת את הגבול, והיא לא קטנה כשהניסיונות גדלים.

התור הוא החבילה `queue` מפרק 14, ממופה לטבלה `deliveries`:

| בפרק 14 | ב־Hookline |
| --- | --- |
| `running` | `in_flight` |
| `run_at` | `next_attempt_at` |
| `locked_until` | `lease_until` |
| `max_attempts` בשורה | `HOOKLINE_MAX_ATTEMPTS`, לסבב |

הפונקציה `Claim` מחזירה גם את מה שהשליחה צריכה
(כתובת, סוד, האם ה־endpoint פעיל, והאירוע), בשאילתה אחת שמצטרפת לשתי הטבלאות. ‏`attempts` הוא אסימון הגידור:
‏`record` מעדכן רק שורה שעדיין במצב `in_flight` עם אותו מספר ניסיון.

<div dir="ltr" align="left">

```bash
go test -race -count=1 -run 'TestClaim|TestRetry|TestPermanent|TestExpiredLease|TestConcurrentClaims' ./internal/delivery/
```

</div>

**יום 44: הפעלה חוזרת, ה־pool והבדיקה מקצה לקצה.** ההפעלה החוזרת נראית פשוטה: להחזיר מסירה מתה ל־`pending`.
הפרט שמסבך אותה הוא המספור. בטבלה `attempts` יש מפתח ייחודי על הזוג מסירה ומספר. אם מאפסים את `attempts` לאפס,
הניסיון הבא נרשם עם מספר 1 ומתנגש בניסיון 1 הישן. אם לא מאפסים, המסירה כבר הגיעה למקסימום ומתה אחרי כישלון אחד.

הפתרון הוא העמודה `attempt_base` מהמיגרציה `00005`, שמסמך התכנון בפרק 29, סעיף 29.5 מציע. ההפעלה החוזרת שומרת בה
את מספר הניסיונות עד עכשיו, וה־worker משווה למקסימום את ההפרש. זה הקוד מהקובץ `internal/store/postgres/postgres.go`:

<div dir="ltr" align="left">

```go
		UPDATE deliveries SET status = 'pending', attempt_base = attempts, next_attempt_at = now(),
		       lease_until = NULL, last_error = NULL
		WHERE tenant_id = $1 AND id = $2 AND status = 'dead'
		RETURNING `+deliveryColumns, tenantID, id))
```

</div>

- **היומן נשאר רציף.** המספר בטבלת הניסיונות ממשיך לעלות, וכך גם הכותרת `Hookline-Attempt`. המקבלת רואה שזה ניסיון חוזר.
- **התנאי `status = 'dead'` בתוך הפקודה.** שתי לחיצות במקביל מזיזות את המסירה פעם אחת. השנייה מקבלת 409, והבדיקה
  `TestConcurrentReplaysMoveTheDeliveryOnce` שולחת שמונה כאלה במקביל.
- **רק הפעלה של מסירה אחת.** הפעלה המונית אחרי תקלה ארוכה, עם קצב ובקרה, היא ההרחבה שמסמך התכנון בפרק 29 מתאר.

ה־pool הוא ה־worker של פרק 14, עם שני שינויים. ‏endpoint שהושבת אחרי שהמסירה נוצרה לא נשלח: המסירה מתה עם הסיבה
„endpoint disabled”. ומקבל ששולח `Retry-After` מקבל את ההמתנה שביקש, עד התקרה של המדיניות. זה הקוד מהקובץ
`internal/delivery/worker.go`:

<div dir="ltr" align="left">

```go
	delay := p.Backoff.Delay(j.Round, p.rand())
	if res.RetryAfter > delay { // the receiver asked for more time; honour it up to the cap
		delay = min(res.RetryAfter, p.Backoff.Cap)
	}
	status, err := Fail(rec, p.DB, j, res, p.MaxAttempts, delay)
```

</div>

הוסיפי את שלושת הנתיבים לחוזה ול־handlers, ואז הריצי את הכול, כולל הבדיקה מקצה לקצה:

<div dir="ltr" align="left">

```bash
make up
go test -race -count=1 -skip 'RateLimit|OverItsLimit|LimiterFailure|Shed|Breaker|Bulkhead|Trace|PrivateAddress' ./internal/backoff/ ./internal/delivery/ ./internal/api/ ./internal/store/... ./internal/e2e/
```

</div>

הבדיקה `TestPipelineEndToEnd` מריצה בתוך תהליך הבדיקה את כל הרכיבים. ה־API עונה מעל HTTP, וכל השאר רצים
לצידו: ה־relay, ה־dispatcher עם המטמון, וארבעה workers. היא שולחת אירועים ל־sink שב־compose, ובודקת שלושה תרחישים:

| תרחיש | מה הבדיקה עושה | מה היא מצפה |
| --- | --- | --- |
| מסירה רגילה | 25 אירועים, שלושה endpoints, ושלושה ניסיונות חוזרים של לקוח | ‏20, 25 ו־5 אירועים ייחודיים, ואף חתימה לא נדחתה |
| מקבל שנכשל ומתאושש | ה־sink מחזיר 503, ואחרי שני ניסיונות חוזר לעבוד | המסירה נמסרת, עם שלושה ניסיונות או יותר |
| כשל סופי והפעלה חוזרת | ה־sink מחזיר 422 | שתי מסירות מתות אחרי ניסיון אחד; אחרי ההפעלה, נמסרות בניסיון 2 |

לבסוף, הריצי את ארבעת הבינאריים, כל אחד בחלון משלו:

<div dir="ltr" align="left">

```bash
make run          # terminal 1: the API
make relay        # terminal 2
make dispatcher   # terminal 3
make worker       # terminal 4
```

</div>

צרי endpoint שמצביע על הכתובת `http://localhost:8091/hooks/demo`, ושלחי אירועים. כדי שה־sink יאמת את החתימה,
ה־endpoint צריך את הסוד שה־sink מכיר. ה־API יצר לו סוד אחר, ולכן בתרגול מחליפים אותו ישירות במסד:

<div dir="ltr" align="left">

```bash
docker compose -p hookline -f deploy/compose.yaml exec postgres psql -U hookline -d hookline \
  -c "UPDATE endpoints SET secret = 'whsec_change-me' WHERE url LIKE '%/hooks/demo'"
```

</div>

זה הפלט של הרצה כזאת, עם חמישה אירועים:

<div dir="ltr" align="left">

```text
$ curl -s localhost:8091/stats
{"received":5,"unique":5,"duplicates":0,"rejected":0,"failed":0,"hooks":{"smoke":{"received":5,"unique":5,"duplicates":0,"rejected":0,"failed":0}}}
```

</div>

### 38.5.5 הבדיקות שצריכות להוריק

<div dir="ltr" align="left">

```bash
make up
go test -race -count=1 -skip 'RateLimit|OverItsLimit|LimiterFailure|Shed|Breaker|Bulkhead|Trace|PrivateAddress' ./internal/backoff/ ./internal/delivery/ ./internal/api/ ./internal/store/... ./internal/e2e/
go test -race -count=1 -skip 'StaleRelay|Trace|PrivateAddress' ./internal/outbox/ ./internal/dispatch/ ./internal/bus/ ./internal/cache/ ./internal/subcache/
make sink-test
make down
```

</div>

### 38.5.6 הפלט הצפוי

<div dir="ltr" align="left">

```text
ok  	example.com/hookline/internal/backoff	1.817s
ok  	example.com/hookline/internal/delivery	4.267s
ok  	example.com/hookline/internal/api	2.040s
ok  	example.com/hookline/internal/store	3.395s
ok  	example.com/hookline/internal/store/memory	3.646s
ok  	example.com/hookline/internal/store/postgres	4.074s
ok  	example.com/hookline/internal/store/storetest	3.183s
ok  	example.com/hookline/internal/e2e	5.467s
ok  	example.com/hookline/internal/outbox	4.776s
ok  	example.com/hookline/internal/dispatch	4.349s
ok  	example.com/hookline/internal/bus	2.656s
ok  	example.com/hookline/internal/cache	3.612s
ok  	example.com/hookline/internal/subcache	3.270s
```

</div>

### 38.5.7 קריטריון סיום

- ‏`TestPipelineEndToEnd` עוברת: כל אירוע מגיע חתום לכל מקבל מתאים, מסירה מתה מופעלת שוב ונמסרת.
- ‏`TestSignMatchesSharedVectors` בגו ובדיקות החתימה ב־Python עוברות על אותו קובץ.
- בדיקת החכירה עוברת: ‏worker שנתקע מאבד את המסירה, והתוצאה המאוחרת שלו נדחית.
- ‏`TestPoolDrainsOnShutdown` עוברת: ביטול באמצע בקשה מחכה לה ורושם את התוצאה.
- כל המודול נקי ב־`make lint`, וכל הבדיקות ירוקות עם `-race` מול ה־compose.
- קומיט של אבן הדרך, ו־`make down`.

### 38.5.8 טעויות נפוצות

- **לחתום על גוף אחד ולשלוח אחר.** אם הגוף נבנה פעמיים, או עובר קידוד מחדש, החתימה לא תתאים. חותמים על הבתים
  שנשלחים בדיוק.
- **לאפס את `attempts` בהפעלה חוזרת.** הניסיון הבא מתנגש במספר ניסיון קיים, וההכנסה לטבלת הניסיונות נכשלת.
- **לנסות שוב 400.** המקבל לא ישנה את התשובה, וה־worker מבזבז את הניסיונות ואת התור על מסירה שלא תצליח.
- **חכירה קצרה מהזמן הקצוב.** ‏worker שני תופס את המסירה בזמן שהראשון עוד מחכה לתשובה, ושניהם שולחים. ההגדרות
  דוחות חכירה שאינה ארוכה מהזמן הקצוב.
- **להקשיב ל־`ctx` של הכיבוי באמצע בקשה.** הבקשה נקטעת, התוצאה לא נרשמת, והמסירה נשלחת שוב אחרי שהחכירה פגה.
  הבקשה רצה עם context מנותק, מוגבל בחכירה.

### 38.5.9 הרחבה

- הוסיפי `POST /v1/endpoints/{id}/replays`, ההפעלה ההמונית של פרק 29, סעיף 29.5, עם הטבלה `replays` והקצב שהמסמך מציע.
- רשמי בהיסטוריה גם את הכותרת `Retry-After` שהמקבל שלח, ובדקי שה־API מחזיר אותה.

## 38.6 M6: חוסן ועומס

### 38.6.1 המטרה

לגרום ל־Hookline להתנהג טוב כשמשהו מסביבו מתנהג רע. מקבל שנתקע לא יתפוס את כל ה־workers, ומקבל שנכשל שוב
ושוב לא יקבל עוד בקשות. דיירת ששולחת יותר מדי תקבל 429, ושאר הדיירות לא ירגישו. ‏API עמוס יענה 503 מיד, ולא יצבור
תור בלי גבול. וה־relay ירוץ בכמה עותקים, עם אחד פעיל. בסוף אבן הדרך את מריצה את ניסוי הבידוד של דרישה N3 ורושמת
את המספרים שלך.

### 38.6.2 המושגים, והיכן נלמדו

| מושג | איפה בספר | איפה בפרויקט |
| --- | --- | --- |
| מפסק זרם | פרק 15, סעיף 15.5 | `internal/breaker`, ‏`delivery.Guards` |
| מחיצה אטומה | פרק 15, סעיף 15.6 | `internal/bulkhead`, ‏`delivery.Guards` |
| ‏token bucket, ו־Redis משותף | פרק 16, סעיפים 16.2 ו־16.3 | `internal/ratelimit`, ‏`api/limit.go` |
| השלת עומס לפני כל עבודה | פרק 16, סעיף 16.5 | `internal/shed`, ‏`shedV1` |
| ‏advisory lock כמנהיג | פרק 18, סעיפים 18.2 ו־18.5 | `internal/leader`, ‏`cmd/relay` |
| ‏fencing token | פרק 18, סעיף 18.4 | `Relay.Term` |

### 38.6.3 הקבצים

| קובץ | מצב ב־starter | מה עושים |
| --- | --- | --- |
| `internal/breaker`, ‏`internal/bulkhead` | גופים ריקים | להעביר מהפתרון שלך לפרק 15 |
| `internal/ratelimit` | גופים ריקים | מפרק 16, עם קצב ופרץ לכל קריאה במקום לכל מגביל |
| `internal/shed` | גופים ריקים | מפרק 16, עם תשובה בפרטי בעיה |
| `internal/leader` | גופים ריקים | מפרק 18 |
| `internal/delivery/guards.go` | גופים ריקים | מפסק ומחיצה לכל endpoint, ו־`Postpone` |
| `internal/api/limit.go` | גוף ריק | הגבלת הקצב לפי הדיירת |
| `migrations/00006_leader_terms.sql` | שלם | הטבלה של הקדנציות |
| `deploy/compose.yaml`, ‏`scripts/lab-n3.sh` | שלמים | ‏sink שני, וניסוי הבידוד |

### 38.6.4 הצעדים

**יום 54: הגנות לכל endpoint.** העתיקי את שתי החבילות, `breaker` ו־`bulkhead`, מהתרגיל של פרק 15, והחליפי את שם
המודול. הבדיקות
שלהן עוברות מיד, כי הן לא תלויות בשום דבר בפרויקט.

השאלה המעניינת היא איפה הן יושבות. כל worker מחזיק מפסק ומחיצה **לכל endpoint**, בזיכרון של התהליך. זו ההחלטה
מטבלת ההחלטות בפרק 37, סעיף 37.5: מצב מקומי פשוט יותר ונכשל בבטחה. המחיר הוא שכל תהליך לומד לבד שה־endpoint מת.
זה הקוד מהקובץ `internal/delivery/worker.go`:

<div dir="ltr" align="left">

```go
	g := p.Guards.get(j.EndpointID)
	if g.bulkhead != nil {
		if err := g.bulkhead.Acquire(ctx); err != nil {
			p.postpone(ctx, j, busyWait(), "bulkhead", err)
			return Result{}, false
		}
		defer g.bulkhead.Release()
	}
	done, err := g.breaker.Allow()
	if err != nil {
		p.postpone(ctx, j, p.Guards.openWait(), "breaker", err)
		return Result{}, false
	}
	res = attempt()
	done(breakerError(res))
	return res, true
```

</div>

שלוש החלטות בקוד:

- **מחיצה בלי המתנה.** כשה־endpoint כבר מחזיק את כל המקומות שלו, ה־worker לא מחכה. הוא מחזיר את המסירה לתור
  ותופס אחרת. אחרת, ה־worker עצמו היה נתקע, וזה בדיוק מה שהמחיצה נועדה למנוע.
- **דחייה בלי ניסיון.** ‏`Postpone` מחזירה את `attempts` לערך שלפני התפיסה, ולא רושמת כלום בהיסטוריה. אף בקשה לא
  נשלחה, ולכן אסור שזה יקצר את הניסיונות של המסירה.
- **רק כשל זמני פוגע במפסק.** תשובת 400 היא תשובה של מקבל שעובד. ‏`breakerError` מחזירה שגיאה רק למה שמנוסה שוב:
  זמן קצוב, חיבור, 408, 429 ו־5xx.

<div dir="ltr" align="left">

```bash
go test -race ./internal/breaker/ ./internal/bulkhead/
go test -race -count=1 -run 'Breaker|Bulkhead' ./internal/delivery/
```

</div>

הבדיקה `TestBulkheadKeepsASlowEndpointFromStarvingOthers` היא דרישה N3 בתוך תהליך אחד: מקבל אחד לא עונה לעולם,
והמסירות של מקבל בריא יוצאות בכל זאת, כשהמקבל התקוע לא מחזיק יותר ממקום אחד.

**יום 54, המשך: הגבלת קצב והשלת עומס.** הגבלת הקצב של פרק 16 קבעה קצב אחד לכל המגביל. ב־Hookline לכל דיירת יש קצב
משלה, בעמודות `rate_per_sec` ו־`burst`. לכן `Allow` מקבלת אותם בכל קריאה, והסקריפט ב־Lua לא משתנה בכלל: הוא כבר קיבל
אותם כפרמטרים. זה הקוד מהקובץ `internal/api/limit.go`:

<div dir="ltr" align="left">

```go
		t := tenantFrom(r.Context())
		d, err := s.limiter.Allow(r.Context(), t.ID, float64(t.RatePerSec), t.Burst)
		if err != nil {
			if s.onLimit != nil {
				s.onLimit(err)
			}
			s.log.WarnContext(r.Context(), "rate limiter unavailable; serving without a limit",
				slog.String("tenant_id", t.ID), slog.String("error", err.Error()))
			next.ServeHTTP(w, r)
			return
		}
```

</div>

- **אחרי האימות.** המגביל צריך את הדיירת, ולכן הוא יושב בתוך ה־middleware של האימות, על נתיב הקליטה בלבד.
- **נכשל פתוח.** כש־Redis למטה, הבקשה עוברת ונרשמת בלוג. זו השורה של Redis בטבלת הכשלים בפרק 37, סעיף 37.6. באבן
  הדרך השמינית הקריאה ל־`OnLimiterError` תהפוך למדד.

השלת העומס יושבת בצד השני: **לפני** האימות ולפני המגביל, כך שבקשה שנזרקת לא עולה קריאה למסד או ל־Redis (פרק 16, סעיף
16.5). היא עוטפת את כל הנתיבים שמתחילים ב־`/v1/`, ולא את `/healthz`: בדיקת חיים חייבת לענות, ובמיוחד כשהשירות עמוס.

<div dir="ltr" align="left">

```bash
go test -race ./internal/ratelimit/ ./internal/shed/
DATABASE_URL= go test -race -run 'RateLimited|OverItsLimit|OnlyIngest|FailsOpen|Sheds' ./internal/api/
```

</div>

**יום 55: relay אחד פעיל, וניסוי הבידוד.** העתיקי את `leader` מפרק 18. הטבלה `leader_terms` נוצרת עכשיו במיגרציה
`00006`, ולא ב־`EnsureSchema`. ה־relay עוטף את הלולאה שלו בבחירה: מי שמחזיק את ה־advisory lock מפרסם, והשאר מחכים.

הקדנציה של המנהיג היא גם fencing token, כמו בפרק 18, סעיף 18.4. ‏relay שאיבד את המנהיגות ולא יודע עדיין, למשל אחרי
עצירה ארוכה, יכול עוד לפרסם. זו כפילות, וה־dispatcher סופג אותה. אבל הסימון שלו נדחה, כי הוא בודק את הקדנציה. זה
הקוד מהקובץ `internal/outbox/outbox.go`:

<div dir="ltr" align="left">

```go
		tag, err := tx.Exec(ctx, `
			UPDATE outbox SET published_at = now()
			WHERE id = ANY($1)
			  AND ($2::bigint = 0 OR (SELECT term FROM leader_terms WHERE lock_key = $3) = $2)`,
			seqs, r.Term, RelayLockKey)
```

</div>

כשהסימון נדחה, `Run` מחזירה `ErrNotLeader`, וה־relay יורד מהמנהיגות ומתמודד מחדש.

<div dir="ltr" align="left">

```bash
go test -race -count=1 ./internal/leader/ ./internal/outbox/
```

</div>

עכשיו ניסוי הבידוד. ל־compose נוסף מקבל מדומה שני, `sink-slow`, בפורט 8092. הסקריפט `scripts/lab-n3.sh` עושה
ארבעה דברים:

- מעלה את הסביבה ואת ארבעת הבינאריים;
- יוצר שני endpoints, אחד לכל מקבל, ומנה את שניהם לכל האירועים;
- מעביר את המקבל השני למצב `timeout_mode`, שבו כל בקשה נתקעת עד שה־worker מוותר;
- שולח 300 אירועים, ומדפיס את פיגור המסירה של המקבל הבריא, מהקליטה ועד המסירה.

כך מריצים אותו, עם ההגנות ובלעדיהן:

<div dir="ltr" align="left">

```bash
make lab-n3
GUARDS=off make lab-n3
```

</div>

ההרצה השנייה מכבה את ההגנות: מחיצה ברוחב כל ה־pool, ומפסק שלא נפתח לעולם. ההשוואה בין שתי ההרצות היא כל העניין.
הפלט נראה כך; את המספרים את מודדת ורושמת:

<div dir="ltr" align="left">

```text
== delivery lag of the healthy endpoint, seconds from ingest to delivery
delivered=300  p50=___  p95=___  p99=___  max=___
== the slow endpoint: deliveries by status, attempts spent
pending=___ (attempts ___)
== breaker transitions in the worker log
___
```

</div>

> **שימי לב:** אל תכתבי ב־`docs/EVALUATION.md` מספר שלא מדדת, וכתבי ליד כל מספר את התאריך, המחשב ואת הפקודה. היעד
> של N2 הוא 30 שניות ל־p99. האם הוא מתקיים עם ההגנות, והאם הוא מתקיים בלעדיהן? זו השאלה שהניסוי עונה עליה, על
> המחשב שלך.

### 38.6.5 הבדיקות שצריכות להוריק

<div dir="ltr" align="left">

```bash
make up
go test -race -count=1 -skip 'Trace' ./internal/breaker/ ./internal/bulkhead/ ./internal/ratelimit/ ./internal/shed/ \
  ./internal/leader/ ./internal/outbox/ ./internal/delivery/ ./internal/api/ ./internal/e2e/ ./migrations/
make lab-n3
```

</div>

### 38.6.6 הפלט הצפוי

<div dir="ltr" align="left">

```text
ok  	example.com/hookline/internal/breaker	2.193s
ok  	example.com/hookline/internal/bulkhead	2.577s
ok  	example.com/hookline/internal/ratelimit	4.968s
ok  	example.com/hookline/internal/shed	4.894s
ok  	example.com/hookline/internal/leader	6.315s
ok  	example.com/hookline/internal/outbox	6.058s
ok  	example.com/hookline/internal/delivery	8.614s
ok  	example.com/hookline/internal/api	4.493s
ok  	example.com/hookline/internal/e2e	7.583s
ok  	example.com/hookline/migrations	2.740s
```

</div>

### 38.6.7 קריטריון סיום

- המפסק נפתח אחרי מספר הכשלים שהוגדר, ובזמן שהוא פתוח אף בקשה לא מגיעה ל־endpoint, ואף ניסיון לא מתבזבז.
- מקבל תקוע לא מחזיק יותר מקומות מהמחיצה שלו, ומקבל בריא ממשיך לקבל.
- דיירת מעל המגבלה מקבלת 429 עם `Retry-After`, ודיירת אחרת באותו רגע מקבלת 202.
- ‏API עמוס עונה 503 עם `Retry-After` על נתיבי `/v1`, ו־`/healthz` ממשיך לענות.
- שני relays רצים: אחד מוביל, והשני לוקח את המנהיגות כשהראשון נהרג. ‏relay עם קדנציה ישנה לא מסמן.
- ‏`make lab-n3` רץ בשתי הגרסאות, והמספרים שלך רשומים ב־`docs/EVALUATION.md`, עם תאריך ומחשב.
- קומיט של אבן הדרך, ו־`make down`.

### 38.6.8 טעויות נפוצות

- **מחיצה שמחכה.** ‏worker שמחכה למקום פנוי ב־endpoint תקוע נתקע בעצמו, ואחרי כמה כאלה כל ה־pool מחכה. בלי המתנה,
  ודחייה של המסירה.
- **דחייה שמבזבזת ניסיון.** אם `Postpone` לא מחזירה את `attempts`, מפסק פתוח שורף את הניסיונות של כל המסירות של
  ה־endpoint, והן מתות בלי שנשלחה בקשה.
- **לספור 4xx ככשל של המפסק.** ‏endpoint שמחזיר 400 על גוף מסוים ייסגר לגמרי, גם לאירועים תקינים.
- **מגביל בזיכרון עם כמה עותקים.** שלושה עותקים של ה־API נותנים לכל דיירת פי שלושה מהמגבלה. בלי Redis, המגביל
  המקומי הוא רק למצב פיתוח.
- **השלת עומס אחרי האימות.** אז כל בקשה שנזרקת עולה קריאה למסד, והעומס שנחסך חוזר דרך הדלת האחורית.
- **לפרסם מספרים שלא מדדת.** פלט הדוגמה כאן ריק בכוונה. המספרים שלך תלויים במחשב שלך.

### 38.6.9 הרחבה

- הוסיפי ל־sink שליטה לכל hook, במקום sink שני, ובדקי שהבדיקות שלו נשארות ירוקות.
- נסי מפסק משותף ב־Redis לכל ה־workers (פרק 15, סעיף 15.5), ומדדי אם המקבל התקוע מקבל פחות בקשות מאשר עם מפסקים מקומיים.
- הוסיפי עדיפות גבוהה להשלת העומס לנתיב `GET /v1/events/{id}`, כדי שהדיירות יוכלו לבדוק מצב גם בזמן עומס.

## 38.7 M7: קונטיינרים, Kubernetes ו־CI

### 38.7.1 המטרה

להוציא את Hookline מהמחשב שלך אל אשכול. בסוף אבן הדרך יש לך ארבעה דברים:

- שני images קטנים, שרצים כמשתמש שאינו root;
- ‏chart של Helm שמתקין את ארבעת הבינאריים, עם מיגרציות לפני כל התקנה;
- תצורת Terraform שמתקינה את אותו chart;
- ‏workflow שבודק את הכול, כולל התקנה אמיתית על kind. בדיקת העשן היא אירוע אחד שנשלח ל־API באשכול ומגיע
ל־sink באשכול, חתום.

### 38.7.2 המושגים, והיכן נלמדו

| מושג | איפה בספר | איפה בפרויקט |
| --- | --- | --- |
| ‏Dockerfile רב־שלבי ל־Go | פרק 19, סעיף 19.2 | `Dockerfile` |
| ‏distroless ומשתמש מספרי | פרק 19, סעיף 19.3 | `USER 65532:65532`, ‏`api healthcheck` |
| ‏image ל־Python עם uv | פרק 19, סעיף 19.4 | `sink/Dockerfile` |
| ‏kind, ‏Deployments ו־Services | פרק 20, סעיפים 20.2 עד 20.5 | `deploy/kind/` |
| ‏Secrets ו־probes | פרק 20, סעיפים 20.6 ו־20.7 | ‏`hookline-database`, ה־probes של ה־API |
| ‏requests, ‏limits ו־HPA | פרק 21, סעיפים 21.1 ו־21.2 | `values.yaml`, ‏`hpa.yaml` |
| ‏chart, ‏values ו־templates | פרק 21, סעיפים 21.6 ו־21.7 | `deploy/helm/hookline` |
| ‏Terraform עם ה־provider של Helm | פרק 22, סעיפים 22.3, 22.4 ו־22.6 | `deploy/terraform` |
| ‏workflow, ‏images ופריסה ל־kind ב־CI | פרק 23, סעיפים 23.2 עד 23.5 | `.github/workflows/hookline.yml` |
| ‏actionlint | פרק 23, סעיף 23.8 | אותו קובץ |

### 38.7.3 הקבצים

כל הקבצים של אבן הדרך הזאת שלמים ב־starter: אין בהם קוד Go לכתוב. העבודה היא לקרוא כל קובץ, להבין כל שורה, להריץ,
ולתקן את מה שלא עובד אצלך.

| קובץ | תפקיד |
| --- | --- |
| `Dockerfile`, ‏`.dockerignore` | ‏image אחד עם חמשת הבינאריים |
| `cmd/api/healthcheck.go` | ‏`api healthcheck`, כי ב־distroless אין curl |
| `sink/Dockerfile`, ‏`sink/.dockerignore` | ה־sink כמו המקבל של פרק 19 |
| `deploy/kind/kind-config.yaml` | אשכול של צומת אחד, image מוצמד לפי digest |
| `deploy/kind/infra.yaml` | ‏PostgreSQL, ‏Redis, ‏Redpanda וה־sink, עותק אחד לכל אחד |
| `deploy/helm/hookline/` | ה־chart: ‏Deployments, ‏Service, ‏ConfigMap, ‏HPA, וה־Job של המיגרציות |
| `deploy/terraform/` | ‏namespace, ‏Secret ו־`helm_release`, עם מודול מקומי |
| `.github/workflows/hookline.yml` | שישה jobs, מבדיקות היחידה ועד פרסום ל־GHCR |
| `scripts/check-chart.sh`, ‏`scripts/kind-smoke.sh` | הבדיקות הסטטיות, ובדיקת העשן מקצה לקצה |

### 38.7.4 הצעדים

**יום 68: ה־images.** קראי את `Dockerfile`. זה ה־Dockerfile של פרק 19, סעיף 19.2, עם הבדל אחד: הוא בונה את כל
חמשת הבינאריים לאותו image. כל Deployment ב־chart מריץ אחד מהם כפקודה. זה הקטע של הבנייה:

<div dir="ltr" align="left">

```dockerfile
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/ \
      ./cmd/api ./cmd/relay ./cmd/dispatcher ./cmd/worker ./cmd/hookline-admin
```

</div>

‏image אחד לכל הבינאריים הוא בחירה. החלופה, image לכל בינארי, נותנת images קטנים יותר ופריסה נפרדת לכל רכיב. כאן
הבינאריים חולקים כמעט את כל הקוד, משתחררים יחד, וכך יש image אחד לסרוק, לתייג ולטעון ל־kind.

<div dir="ltr" align="left">

```bash
make images
docker image ls | grep -E '^hookline'
docker run --rm --entrypoint id hookline-sink:dev
```

</div>

הפקודה האחרונה צריכה להראות `uid=10001`. ‏image של Go רץ כ־65532, המשתמש `nonroot` של distroless.

**יום 69: kind, ה־chart והמיגרציות.** קראי את `deploy/kind/infra.yaml`. התשתית לא נכנסה ל־chart בכוונה: באשכול
אמיתי היא שירות מנוהל, וה־chart רק מקבל את הכתובות שלה. כאן היא קבצי YAML פשוטים, עותק אחד לכל רכיב, והנתונים
ב־`emptyDir`, שנעלם עם ה־pod. זה אשכול תרגול.

עכשיו ה־chart. שלוש החלטות שכדאי לזהות בו:

- **הסוד לא ב־values.** כתובת המסד מגיעה מ־Secret שקיים לפני ההתקנה, כי ערכי ה־chart נשמרים בהיסטוריה של
  ה־release. ה־chart מכיר רק את שם ה־Secret ואת המפתח.
- **מיגרציות ב־hook.** ‏Job שרץ לפני כל התקנה ועדכון מריץ את המיגרציות ואז יוצר את ה־topic, לפני שאף Deployment
  משתנה. שתי הפקודות אידמפוטנטיות, כך שהרצה חוזרת לא מזיקה.
- **‏`enableServiceLinks: false` בכל pod.** אחרת ה־kubelet מוסיף משתני סביבה לכל Service, למשל `HOOKLINE_API_PORT`,
  וזה המלכודת של פרק 20 עם המשתנה של הפורט של המקבל.

זה הקטע של ה־hook, מהקובץ `deploy/helm/hookline/templates/job-migrate.yaml`:

<div dir="ltr" align="left">

```yaml
  annotations:
    "helm.sh/hook": pre-install,pre-upgrade
    "helm.sh/hook-weight": "0"
    "helm.sh/hook-delete-policy": before-hook-creation,hook-succeeded
```

</div>

ה־hook לא רואה את ה־ConfigMap של ה־chart, כי בזמן `pre-install` הוא עוד לא קיים. לכן ה־Job מחזיק את ההגדרות שלו
בעצמו. הריצי את הבדיקות הסטטיות, ואחר כך את בדיקת העשן:

<div dir="ltr" align="left">

```bash
make chart-check
KEEP=1 make kind-smoke
```

</div>

‏`scripts/kind-smoke.sh` עובד עם קובץ kubeconfig משלו, ולא נוגע בשום הקשר אחר שלך. הוא יוצר את האשכול, בונה וטוען
את שני ה־images, מחיל את התשתית, יוצר את ה־Secret ומתקין את ה־chart. אחר כך הוא יוצר דיירת דרך
`kubectl exec`, ‏endpoint שמצביע על ה־sink באשכול, ושולח אירוע אחד. עם `KEEP=1` האשכול נשאר, למחר.

**יום 70: Terraform ו־CI.** ‏Terraform מתקין את אותו chart, מהנתיב המקומי שלו ולא מעותק, למרחב שמות משלו. הוא
יוצר את ה־namespace ואת ה־Secret, ומתקין את ה־release ליד התשתית שכבר רצה. זה המשאב העיקרי, מהקובץ
`deploy/terraform/modules/hookline/main.tf`:

<div dir="ltr" align="left">

```hcl
resource "helm_release" "hookline" {
  name      = var.release_name
  namespace = kubernetes_namespace_v1.this.metadata[0].name
  chart     = "${path.module}/../../../helm/hookline"
```

</div>

<div dir="ltr" align="left">

```bash
cd deploy/terraform
cp terraform.tfvars.example terraform.tfvars
terraform init
terraform plan -out plan.out
terraform apply plan.out
terraform plan -var worker_replicas=2
terraform destroy
cd ../..
make kind-down
```

</div>

‏`terraform.tfvars` מוחזק מחוץ ל־Git, כי בו כתובת המסד. גם קובץ ה־state יכול להחזיק אותה, ולכן גם הוא מחוץ ל־Git.
הקובץ `.terraform.lock.hcl` כן נכנס, עם ארבע פלטפורמות, כדי שה־CI ומחשבים אחרים יקבלו בדיוק אותם providers.

לסיום, ה־workflow. הוא שייך לשורש המאגר שלך, ליד ה־action של פרק 23:

<div dir="ltr" align="left">

```bash
cp -r ~/senior-course/hookline/.github/. ~/senior-course/.github/
cd ~/senior-course && actionlint
```

</div>

שישה jobs:

| job | מה הוא עושה |
| --- | --- |
| `unit` | ה־action של פרק 23: ‏gofmt, ‏vet, ‏staticcheck ו־`go test -race` |
| `integration` | ה־compose כולו, ואז אותו action עם `fail-on-skip`: כל בדיקת אינטגרציה חייבת לרוץ |
| `sink` | ‏uv, ‏ruff, ‏mypy ו־pytest |
| `deploy-config` | בדיקות ה־chart, ו־Terraform בלי אשכול |
| `kind-smoke` | אשכול kind ב־CI, ואותו סקריפט עשן |
| `images` | שני images ל־GHCR, רק מ־main ורק אחרי שהכול עבר |

‏Redpanda צריך שורת פקודה, ושירות צמוד של GitHub Actions לא מקבל כזאת. לכן ה־job של האינטגרציה מריץ את
קובץ ה־compose, ולא services (פרק 23, סעיף 23.3). הצד הטוב: ה־CI והמחשב שלך מריצים בדיוק אותה סביבה.

> **שימי לב:** ה־workflow נבדק כאן ב־actionlint בלבד. הוא ירוץ באמת רק במאגר שלך ב־GitHub, וההרצה הראשונה שלו
> היא חלק מאבן הדרך: אם job נכשל שם, קראי את הלוג שלו ותקני.

### 38.7.5 הבדיקות שצריכות להוריק

<div dir="ltr" align="left">

```bash
make chart-check
make kind-smoke
make tf-check
cd ~/senior-course && actionlint
```

</div>

### 38.7.6 הפלט הצפוי

<div dir="ltr" align="left">

```text
$ make chart-check
...
PASS  "helm.sh/hook": pre-install,pre-upgrade
...
all chart checks passed
$ make kind-smoke
...
== smoke: one event to the in-cluster sink
event 01a10c1d-c2ea-7846-b280-8f3878a0be35
sink: {"received":1,"unique":1,"duplicates":0,"rejected":0,"failed":0,"hooks":{"smoke":{"received":1,"unique":1,"duplicates":0,"rejected":0,"failed":0}}}
deliveries: [('delivered', 1)]
PASS: the event reached the in-cluster sink, signed
$ terraform apply plan.out
Apply complete! Resources: 3 added, 0 changed, 0 destroyed.
$ terraform plan -var worker_replicas=2
  # module.hookline.helm_release.hookline will be updated in-place
$ terraform destroy
Destroy complete! Resources: 3 destroyed.
```

</div>

### 38.7.7 קריטריון סיום

- שני ה־images נבנים ורצים כמשתמש שאינו root, וב־image של Go אין shell.
- ‏`make chart-check` עובר: ‏kubeconform, ‏`helm lint --strict`, ורשימת הבדיקות על הפלט.
- ‏`make kind-smoke` מסתיים ב־PASS: האירוע הגיע ל־sink באשכול, חתום.
- ‏`terraform apply` מתקין, ‏`plan` אחרי שינוי מראה שינוי אחד, ו־`destroy` מנקה.
- ‏actionlint נקי על ה־workflow, וה־workflow רץ במאגר שלך ב־GitHub.
- קומיט של אבן הדרך, ו־`make kind-down`.

### 38.7.8 טעויות נפוצות

- **‏`localhost` בתוך האשכול.** כתובת כזאת מצביעה על ה־pod של ה־worker עצמו, ולא על ה־sink. באשכול מצביעים על
  ה־Service: ‏`sink.hookline.svc.cluster.local`.
- **‏Job של מיגרציות שמחכה ל־ConfigMap של ה־chart.** לפני ההתקנה הוא עוד לא קיים, וה־Job נתקע. ב־`kubectl describe`
  רואים `CreateContainerConfigError`.
- **‏image בלי תג, או עם `latest`.** עם התג הזה Kubernetes מנסה למשוך מה־registry, ולא מוצא את ה־image שטענת ל־kind.
  תג מפורש, כמו `dev`, פותר את זה.
- **‏`DATABASE_URL` ב־values.** הוא נשמר בהיסטוריה של ה־release, ונראה ב־`helm get values`. רק דרך Secret.
- **הקשר kubectl לא נכון.** פקודה שרצה על האשכול הלא נכון יכולה למחוק דברים אחרים. הסקריפט משתמש בקובץ kubeconfig משלו.

### 38.7.9 הרחבה

- הוסיפי NetworkPolicy שמגבילה את ה־workers ל־DNS, למסד וליציאה החוצה. חשבי מה זה עושה להגנת ה־SSRF של אבן
  הדרך התשיעית.
- הריצי את Trivy או את Grype על שני ה־images (פרק 19, סעיף 19.6), ורשמי מה מצאו.
- הוסיפי workflow של שחרור על תג, עם actions מוצמדים לפי SHA, כמו בקובץ `release.yml` של פרק 23.

## 38.8 M8: תצפיתיות ו־SLO

### 38.8.1 המטרה

לראות את Hookline מבפנים, ולדעת מתי הוא שובר הבטחה. בסוף אבן הדרך, מעקב אחד עובר מבקשת הקליטה, דרך ה־relay,
‏Redpanda, ה־dispatcher וה־worker, ועד ה־sink. יש מדדים שאומרים מה שה־HTTP לא יודע להגיד, כמו פיגור המסירה וגיל
ה־outbox. יש יעד רמת שירות (SLO) למסירה עם התראות קצב שריפה, ו־runbook לכל התראה. ויש תחקיר על גרסה ששברת בכוונה,
עם המספרים שמדדת.

### 38.8.2 המושגים, והיכן נלמדו

| מושג | איפה בספר | איפה בפרויקט |
| --- | --- | --- |
| שלושת האותות ו־OTLP דרך collector | פרק 24 | `internal/telemetry`, ‏`deploy/compose.observability.yaml` |
| ‏otelhttp ו־`RoutePattern` | פרק 24 | `internal/api/server.go` |
| מעקב דרך Kafka עם kotel | פרק 24 | `bus.Tracing`, ‏`dispatch.Handle` |
| לוגים עם מזהה המעקב, ל־Loki | פרק 24 | `wire.Telemetry`, ‏`platform.NewLogger` |
| ‏RED ו־USE | פרק 24 | שני לוחות הבקרה |
| ‏SLI, ‏SLO ותקציב שגיאות | פרק 25 | `deploy/observability/rules/` |
| התראות קצב שריפה בכמה חלונות | פרק 25 | `slo-alerts.yml`, ‏`delivery-slo.yml` |
| בדיקות לכללים עם promtool | פרק 25 | `deploy/observability/tests/rules_test.yml` |
| ‏runbook ותחקיר ללא האשמה | פרק 25 | `docs/runbooks/`, ‏`docs/postmortem-*.md` |

### 38.8.3 הקבצים

| קובץ | מצב ב־starter | מה עושים |
| --- | --- | --- |
| `internal/telemetry/telemetry.go`, ‏`start.go` | שלמים | ה־`Setup` של פרק 24, ו־`Start` שלא עושה כלום בלי collector |
| `internal/metrics/metrics.go` | גופים ריקים | המכשירים של Hookline |
| `migrations/00007_events_trace_context.sql` | שלם | העמודה שמחזיקה את המעקב של הקליטה |
| `internal/api`, ‏`internal/bus`, ‏`internal/dispatch`, ‏`internal/delivery` | גופים ריקים | העברת ההקשר של המעקב בכל קפיצה |
| `deploy/compose.observability.yaml`, ‏`deploy/observability/` | שלמים | ה־collector, ‏Prometheus, ‏Loki, ‏Tempo ו־Grafana |
| `deploy/observability/rules/`, ‏`tests/` | שלמים | הכללים והבדיקות שלהם |
| `docs/runbooks/` | שלמים | ‏runbook לכל התראה; לקרוא, ולעדכן אחרי התקרית שלך |
| `scripts/check-telemetry.sh`, ‏`scripts/stage-incident.sh` | שלמים | הבדיקה מקצה לקצה, והתקרית המבוימת |

### 38.8.4 הצעדים

**יום 77: מעקב אחד לכל אירוע.** הקושי של Hookline הוא שהאירוע עובר בין תהליכים **לא באותו זמן**. ה־API מחזיר
202 ומסיים, וה־relay מפרסם רק אחר כך, בטרנזקציה אחרת. ההקשר של המעקב לא יכול לעבור בזיכרון, ולכן הוא נשמר עם
האירוע: עמודה חדשה, `events.trace_context`, מחזיקה את הכותרת `traceparent` של בקשת הקליטה. זו הפונקציה מהקובץ
`internal/api/ingest.go`:

<div dir="ltr" align="left">

```go
func traceparent(ctx context.Context) string {
	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	return carrier.Get("traceparent")
}
```

</div>

מכאן כל קפיצה ממשיכה את אותו מעקב:

| קפיצה | איך ההקשר עובר | ה־span |
| --- | --- | --- |
| ‏API | ‏otelhttp, ו־`RoutePattern` לפניו כדי שהנתיב יופיע | `POST /v1/events` |
| ‏relay | ‏`traceparent` מהשורה, ואז kotel כותב אותו לכותרות של הרשומה | `hookline.events publish` |
| ‏dispatcher | ‏kotel קורא את הכותרות; ‏`WithProcessSpan` | `hookline.events receive`, ‏`process` |
| ‏worker | ‏`traceparent` מהשורה של האירוע, ו־otelhttp בלקוח שולח אותו הלאה | `hookline.deliver`, ‏`HTTP POST` |
| ‏sink | ה־instrumentation של FastAPI מפרק 24 | `POST /hooks/{name}` |

ה־span של ה־worker הוא ילד של ה־span של הקליטה, ולא של ה־dispatcher. הסיבה: ה־worker קורא את ההקשר מהאירוע. זו
פשרה: ניסיון חוזר אחרי עשר דקות עדיין שייך לאותו מעקב, ונראה בו בדיוק מתי ניסו. זה הקוד מהקובץ
`internal/delivery/worker.go`:

<div dir="ltr" align="left">

```go
	ctx, span := otel.Tracer(metrics.Scope).Start(withTraceparent(ctx, j.TraceContext), "hookline.deliver",
		trace.WithAttributes(attribute.String("hookline.delivery.id", j.DeliveryID),
			attribute.String("hookline.event.id", j.EventID), attribute.String("hookline.endpoint.id", j.EndpointID),
			attribute.Int("hookline.attempt", j.Attempt)))
	defer span.End()
```

</div>

בלי collector, `telemetry.Start` לא עושה כלום, וכל המכשירים הם no-op. כך הבדיקות, ה־compose של אבני הדרך הקודמות
והניסויים רצים כמו קודם, בלי שגיאות ייצוא בלוג.

<div dir="ltr" align="left">

```bash
DATABASE_URL= go test -race -run 'Trace|Health' ./internal/api/ ./internal/telemetry/
go test -race -count=1 -run 'TestDeliveryContinuesTheIngestTrace' ./internal/delivery/
```

</div>

**יום 78: מדדים, כללים ולוחות.** המדדים של HTTP אומרים אם ה־API עונה. הם לא אומרים אם webhooks מגיעים. אלה
המכשירים שהחבילה `internal/metrics` מוסיפה:

| מדד ב־Prometheus | סוג | מה הוא אומר |
| --- | --- | --- |
| `hookline_delivery_lag_seconds` | histogram | מקליטה ועד 2xx, לכל מסירה שנמסרה; ה־SLI של N2 |
| `hookline_delivery_attempts_total` | counter, לפי תוצאה | ‏delivered, ‏retry או dead |
| `hookline_delivery_postponed_total` | counter, לפי סיבה | דחיות של מפסק פתוח או מחיצה מלאה |
| `hookline_breaker_transitions_total` | counter | מעברי מצב של מפסקים, בלי endpoint כ־label |
| `hookline_ingest_events_total` | counter, לפי תוצאה | אירועים חדשים, וחזרות עם אותו מפתח |
| `hookline_ratelimit_errors_total` | counter | בקשות שעברו בלי מגבלה כי Redis לא ענה |
| `hookline_dispatch_deliveries_total` | counter | שורות מסירה שה־dispatcher יצר |
| `hookline_outbox_pending`, ‏`hookline_outbox_oldest_age_seconds` | gauge | כמה מחכים ב־outbox, וכמה זמן הוותיק מחכה |

שתי החלטות בטבלה. **‏30 הוא גבול של דלי** בהיסטוגרמה של הפיגור, כי ה־SLI סופר מסירות בתוך 30 שניות, ואת זה אפשר
לחשב בדיוק רק על גבול (פרק 25). ו**אין endpoint כ־label** במעברי המפסק, כי סדרה לכל endpoint גדלה בלי גבול. המזהה
נמצא בשורת הלוג.

המדדים נדחפים אל ה־collector, ומשם ל־Prometheus, כמו בפרק 24, ואין נקודת `/metrics` לגירוד. לשלושת הבינאריים
שאינם ה־API יש עכשיו מאזין קטן עם `/healthz`, בשביל ה־probes של ה־chart.

עכשיו ה־SLO של המסירה, מהקובץ `deploy/observability/rules/delivery-slo.yml`:

<div dir="ltr" align="left">

```yaml
      - record: job:slo_late_deliveries:ratio_rate5m
        expr: |
          1 - (
            sum by (job) (rate(hookline_delivery_lag_seconds_bucket{le="30"}[5m]))
            / sum by (job) (rate(hookline_delivery_lag_seconds_count[5m]))
          )
```

</div>

שימי לב למה שה־SLI הזה **לא** רואה: הוא סופר רק מסירות שהגיעו. כשאף מסירה לא מגיעה, אין לו נתונים, וההתראה שלו
שותקת בדיוק כשהמצב הכי גרוע. לכן יש התראה שנייה, `DeliveryAttemptsFailing`: יותר מחצי מהניסיונות נכשלים במשך
שלוש דקות. מקבל אחד לא יכול לגרום לזה. כשרובם נכשלים יחד, כנראה הבעיה אצלנו. ויש `OutboxStalled`: אירועים
מתקבלים, אבל לא יוצאים מה־outbox.

<div dir="ltr" align="left">

```bash
make promtool
```

</div>

הבדיקות ב־`deploy/observability/tests/rules_test.yml` מריצות שישה תרחישים, וכל אחד מוכיח שהתראה נדלקת בזמן הנכון
ושותקת לפניו. התרחיש הרביעי הוא הלב: כל הניסיונות נכשלים, ה־SLO של הפיגור שותק, ו־`DeliveryAttemptsFailing` נדלקת.
אחרי הבדיקות, הסקריפט בודק שלכל `runbook_url` יש קובץ.

לבסוף, העלי את הסביבה עם הטלמטריה, והריצי את הבדיקה מקצה לקצה:

<div dir="ltr" align="left">

```bash
make obs-up
make check-telemetry
```

</div>

‏`scripts/check-telemetry.sh` מריץ את ארבעת הבינאריים עם ייצוא ל־collector, שולח אירוע עם מזהה מעקב שהוא בחר, ושואל
כל מערכת את השאלה שלה:

| מערכת | השאלה |
| --- | --- |
| Tempo | האם המעקב עובר דרך כל הקפיצות? |
| Prometheus | האם כל הסדרות של הטבלה קיימות, וגם כללי ה־SLO? |
| Loki | האם יש שורות לוג עם מזהה המעקב? |
| Grafana | האם שני הלוחות נטענו? | פתחי אחר כך את Grafana
בכתובת `http://localhost:3000`, בתיקייה Hookline.

**יום 79: הגרסה השבורה והתחקיר.** הריצי את התקרית המבוימת:

<div dir="ltr" align="left">

```bash
make stage-incident
```

</div>

הסקריפט שולח בערך חמישה אירועים בשנייה, ואחרי שתי דקות „משחרר” worker שבו `HOOKLINE_ATTEMPT_TIMEOUT=1ms`, טעות של
יחידה במקום `1s`. כל ניסיון נגמר בזמן קצוב. הקליטה נשארת ירוקה, וה־API עונה 202 לכל אירוע. הסקריפט מחכה ש־`DeliveryAttemptsFailing`
תידלק, מחזיר את הגרסה הקודמת, ומדפיס ציר זמן ומספרים: כמה מסירות, כמה הגיעו מאוחר מ־30 שניות, ומה הפיגור המרבי.

עם הפלט הזה, כתבי את התחקיר בקובץ `docs/postmortem-<date>.md`, לפי התבנית `docs/postmortem-TEMPLATE.md`. בקובץ
שבמימוש המלא יש סעיף מסומן „Author's run”, עם הריצה של כותב הפרויקט, ומקומות ריקים לריצה שלך. המספרים שלך יהיו אחרים,
ורק הם נכנסים לתחקיר שלך.

### 38.8.5 הבדיקות שצריכות להוריק

<div dir="ltr" align="left">

```bash
make promtool
make up
go test -race -count=1 -skip 'PrivateAddress' ./internal/metrics/ ./internal/telemetry/ ./internal/api/ ./internal/delivery/ ./migrations/
make down
make obs-up
make check-telemetry
make stage-incident
make obs-down
```

</div>

### 38.8.6 הפלט הצפוי

<div dir="ltr" align="left">

```text
$ make promtool
Checking rules/delivery-slo.yml
  SUCCESS: 8 rules found
...
  SUCCESS
runbook ok: docs/runbooks/delivery-attempts-failing.md
...
$ make check-telemetry
...
PASS  Tempo: trace 0cfe9befa3c7a3acd4b02b3ac7e97adb has 7 spans:
        1 hookline-api:POST /v1/events
        1 hookline-dispatcher:hookline.events process
        1 hookline-dispatcher:hookline.events receive
        1 hookline-relay:hookline.events publish
        1 hookline-sink:POST /hooks/{name}
        1 hookline-worker:HTTP POST
        1 hookline-worker:hookline.deliver
...
PASS  Loki: 3 lines carry trace_id=0cfe9befa3c7a3acd4b02b3ac7e97adb
PASS  dashboard hookline-api
PASS  dashboard hookline-pipeline

ALL CHECKS PASSED
$ make stage-incident
hh:mm:ss  CHANGE: worker redeployed with HOOKLINE_ATTEMPT_TIMEOUT=1ms
hh:mm:ss  DeliveryAttemptsFailing=pending ...
hh:mm:ss  PAGE: DeliveryAttemptsFailing is firing
hh:mm:ss  MITIGATION: worker rolled back to the default attempt timeout
...
deliveries=___ delivered=___ dead=___ late_over_30s=___ max_lag_s=___ attempts=___
```

</div>

### 38.8.7 קריטריון סיום

- המעקב של אירוע אחד עובר ב־Tempo דרך כל הקפיצות, מהקליטה ועד המקבל.
- ‏Prometheus מקבל את כל המדדים של הטבלה, ו־Loki את השורות עם מזהה המעקב.
- שני הלוחות נטענים ב־Grafana, ומראים את הקליטה ואת המסירה.
- ‏`make promtool` ירוק: כל התרחישים, וכל ה־runbooks קיימים.
- התקרית המבוימת רצה, ההתראה נדלקה, והתחקיר שלך כתוב עם המספרים שמדדת.
- קומיט של אבן הדרך, ו־`make obs-down`.

### 38.8.8 טעויות נפוצות

- **להעביר את ההקשר רק בזיכרון.** ‏`context` לא שורד את ה־outbox. בלי `events.trace_context`, ה־relay מתחיל מעקב חדש,
  והמסלול נשבר לשניים.
- **‏SLI שרואה רק הצלחות.** כשאף מסירה לא מגיעה, אין לו נתונים, ושום SLO לא נדלק. צריך גם אות על כשלים.
- **‏label עם מזהה של endpoint או דיירת.** כל ערך הוא סדרה חדשה, ו־Prometheus נחנק.
- **לגבול של SLI שאינו גבול של דלי.** אי אפשר לחשב „בתוך 200 מילישניות” מהיסטוגרמה שהגבולות שלה 0.1 ו־0.25.
- **להעתיק לתחקיר את מספרי הדוגמה.** התחקיר מתעד את מה שקרה אצלך, עם השעות והמספרים שלך.

### 38.8.9 הרחבה

- הוסיפי מדד של גיל המסירה הוותיקה שעוד מחכה, והגדירי ממנו SLI שרואה גם מסירות שלא הגיעו.
- העבירי את המדדים גם ל־exemplars, וקשרי מנקודה בגרף הפיגור אל המעקב שלה ב־Tempo.
- כתבי בדיקת promtool לתרחיש שבו מקבל אחד בלבד נכשל, והראי ש־`DeliveryAttemptsFailing` לא נדלקת.

## 38.9 M9: בדיקות ואבטחה

### 38.9.1 המטרה

להוכיח את מה שהמפרט מבטיח, ולסגור את הדלתות שדיירת עוינת יכולה לפתוח. בסוף אבן הדרך, ה־worker לא מתחבר לכתובת
פרטית, גם כששם המארח מחליף כתובת בין הרישום לשליחה. מפתחות API וסודות חתימה לא מופיעים באף שורת לוג. גוף
ה־webhook מתואר בסכמה, ושני הצדדים נבדקים מולה. בדיקות עומס מודדות את N1 ואת N5. סורקי חולשות בודקים את הקוד ואת
ה־images, ומודל איומים מחבר כל איום לבדיקה שמוכיחה את ההגנה עליו.

### 38.9.2 המושגים, והיכן נלמדו

| מושג | איפה בספר | איפה בפרויקט |
| --- | --- | --- |
| ‏SSRF, ובדיקה של הכתובת בזמן החיבור | פרק 28 | `internal/ssrf`, ‏`delivery.NewClient` |
| ‏DNS rebinding | פרק 28 | `Policy.Control` |
| הסתרת סודות בלוג, והטיפוס `Secret` | פרק 28 | `internal/redact`, ‏`platform.NewLogger` |
| מפתח API שנשמר כגיבוב | פרק 28 | `internal/domain/apikey.go`, מאבן הדרך הראשונה |
| ‏STRIDE ומודל איומים | פרק 28 | `docs/threat-model.md` |
| בדיקות חוזה עם JSON Schema | פרק 26 | `contract/`, ‏`sink/tests/test_contract.py` |
| בדיקות תכונה עם rapid | פרק 26 | `internal/page`, ‏`contract` |
| בדיקות עומס עם k6, במודל פתוח | פרק 26 | `loadtest/`, ‏`scripts/load.sh` |
| ‏govulncheck | פרק 27 | `scripts/supply-chain.sh`, ה־workflow |
| ‏SBOM עם Syft, וסריקה עם Grype | פרקים 19 ו־28 | `scripts/supply-chain.sh`, ה־workflow |

### 38.9.3 הקבצים

| קובץ | מצב ב־starter | מה עושים |
| --- | --- | --- |
| `internal/ssrf/ssrf.go` | שלם | המדיניות של פרק 28, ועוד `Transport` |
| `internal/ssrf/hookline.go` | שלם | המדיניות של Hookline לפי מצב הפיתוח, והבדיקה אחרי פענוח השם |
| `internal/api/endpoints.go` | המימוש שלך מאבן הדרך הראשונה | בדיקת הכתובת ביצירת endpoint |
| `internal/delivery/send.go` | המימוש שלך מאבן הדרך החמישית | לקוח עם החיבור המוגן, וחסימה שהיא תשובה סופית |
| `internal/redact/redact.go` | שלם | הכלים של פרק 28, ועוד `Handler` לפני כל פלט |
| `contract/` | שלם | הסכמה, הדוגמאות והבדיקות בצד של Go |
| `sink/tests/test_contract.py` | שלם | אותן דוגמאות, בצד של המקבל |
| `loadtest/*.js`, ‏`scripts/load.sh` | שלמים | שני תרחישי k6, והסקריפט שמכין דיירות |
| `scripts/supply-chain.sh` | שלם | ‏govulncheck, ‏SBOM לכל image וסריקה |
| `docs/threat-model.md` | מלא | לקרוא, להוסיף שורה משלך, ולמלא את הכותרת |

ה־logger מאבן הדרך הראשונה כבר עובר דרך `redact.Handler`. לכן החבילה שלמה ב־starter, ואת ההוכחה שלה תריצי היום.

### 38.9.4 הצעדים

**יום 85: כתובות, סודות ואיומים.** דיירת רושמת URL, ו־Hookline שולח אליו בקשה מתוך הרשת שלנו. זו בדיוק ההגדרה
של SSRF מפרק 28. הכתובת `https://169.254.169.254/` מגיעה לשירות המטא־דאטה של הענן, ושם שמצביע על 10.0.0.5 מגיע
למסד.
ההגנה של Hookline בודקת פעמיים, ומכל פעם מצפה לדבר אחר:

| מתי | מה נבדק | למה |
| --- | --- | --- |
| ביצירת endpoint | ה־URL, וכל כתובת שהשם מתפענח אליה | תשובה ברורה מוקדם, לטעות תמימה ולהתקפה גלויה |
| בכל חיבור של ה־worker | הכתובת שמתחברים אליה בפועל | השם יכול להתפענח אחרת בזמן השליחה, וזה ה־DNS rebinding |

הבדיקה הראשונה היא `CheckResolved`, מהקובץ `internal/ssrf/hookline.go`:

<div dir="ltr" align="left">

```go
	addrs, err := lookup(ctx, u.Hostname())
	if err != nil || len(addrs) == 0 {
		return fmt.Errorf("%w: %s", ErrUnresolvable, u.Hostname())
	}
	for _, a := range addrs {
		if err := p.CheckAddr(a); err != nil {
			return err
		}
	}
	return nil
```

</div>

כל כתובת צריכה לעבור, ולא רק הראשונה. שם עם שתי כתובות, אחת ציבורית ואחת פרטית, נדחה. ה־handler מחזיר לדיירת
תשובה כללית, ואת הסיבה כותב רק ללוג שלנו. זה הקוד מהקובץ `internal/api/endpoints.go`:

<div dir="ltr" align="left">

```go
	if s.vetURL != nil {
		if err := s.vetURL(r.Context(), url); err != nil {
			// The reason stays in our log: telling the tenant which internal address a name resolves to
			// would make this endpoint a scanner of our network.
			s.log.InfoContext(r.Context(), "endpoint URL refused", slog.String("error", err.Error()))
			validation(w, r, []problem.FieldError{{Detail: "must resolve to a public address", Pointer: "#/url"}})
			return
		}
	}
```

</div>

הבדיקה השנייה יושבת ב־dialer. הפונקציה `Control` רצה אחרי הפענוח ולפני החיבור, עם הכתובת של החיבור הזה. לכן אין
רווח בין הבדיקה לשימוש. כך בונה ה־worker את הלקוח שלו:

<div dir="ltr" align="left">

```go
func NewClient(policy *ssrf.Policy) *http.Client {
	var base http.RoundTripper = http.DefaultTransport
	if policy != nil {
		base = policy.Transport()
	}
	return &http.Client{
		Transport:     otelhttp.NewTransport(base),
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
}
```

</div>

ב־`Transport` אין proxy. עם שרת ביניים, הבדיקה הייתה רואה את הכתובת שלו ולא את היעד. חיבור שנחסם הוא תשובה
סופית, בלי ניסיון חוזר: ניסיון נוסף לא יהפוך כתובת פרטית לציבורית.

ומה עם הפיתוח? ה־sink רץ על המחשב שלך, וזו בדיוק כתובת פרטית. במצב פיתוח, `ForMode` מתיר גם http, ורק את
הטווחים שבאמת צריך:

- כתובות ה־loopback, בשביל המקבלים שעל המחשב שלך.
- הטווחים של Docker, בשביל שירותי ה־compose.
- שני הטווחים של kind, בשביל ה־sink שבתוך האשכול.

בייצור לא מותר אף טווח פרטי, ורק https. כל הסקריפטים המקומיים מגדירים `HOOKLINE_DEV_MODE=true`. באשכול של kind,
הערך `devMode` עושה אותו דבר.

<div dir="ltr" align="left">

```bash
go test -race ./internal/ssrf/
DATABASE_URL= go test -race -run 'PrivateAddress' ./internal/api/ ./internal/delivery/
```

</div>

עכשיו הסודות. הפונקציה `ReplaceAttr` של פרק 28 היא אפשרות של ה־handlers המובנים. הגשר ל־Loki מאבן הדרך
השמינית לא מכיר אותה. לכן `redact.Handler` עוטף את כל הפלטים יחד, לפני שהם מתפצלים:

<div dir="ltr" align="left">

```go
func (h handler) Handle(ctx context.Context, r slog.Record) error {
	out := slog.NewRecord(r.Time, r.Level, secretShapes.ReplaceAllString(r.Message, Redacted), r.PC)
	r.Attrs(func(a slog.Attr) bool {
		out.AddAttrs(redactAttr(a))
		return true
	})
	return h.next.Handle(ctx, out)
}
```

</div>

הוא מסתיר לפי שלושה סימנים:

- שם המפתח, למשל `secret` או `api_key`.
- צורת הערך: ‏`hk_` או `whsec_` עם 43 תווים אחריהם. גם אסימון אחרי המילה Bearer נתפס. הצורה תופסת גם הודעה
  שמישהו הדביק לתוכה מפתח.
- הצורה המודפסת של map או struct, כי הפונקציה של פרק 28 לא נכנסת לתוכם.

ההוכחה היא סריקת לוגים, הבדיקה של N7. הבדיקה `TestLogsNeverContainKeysOrSecrets` מריצה את ה־API עם רישום
הלוגים האמיתי, ברמה הכי מפורטת. היא יוצרת endpoint, קוראת עם המפתח, קולטת, ונכשלת בכוונה כמה פעמים. בסוף היא
כותבת שורה רשלנית שמכילה את המפתח ואת הסוד. אף אחד מהם לא מופיע בפלט. בדיקה תאומה עושה אותו דבר ל־worker, עם
ניסיון שנכשל וניסיון שמצליח.

<div dir="ltr" align="left">

```bash
DATABASE_URL= go test -race -run 'Redact|LogsNever' ./internal/redact/ ./internal/api/
```

</div>

שני דברים לא השתנו, וזו החלטה. המפתחות נשארים בפורמט של אבן הדרך הראשונה: ‏`hk_` ועוד 256 ביטים אקראיים, שמורים
כגיבוב ונמצאים לפי הגיבוב. החבילה `apikey` של פרק 28 מוסיפה מזהה מפתח והשוואה בזמן קבוע. כאן לא משווים בקוד שום
דבר סודי, ולכן אין צורך בה. אין מיגרציה, וכל דיירת קיימת ממשיכה לעבוד. גם החבילה `jwtauth` לא נכנסה, כי אין
ממשק של מנהלים. המפעילים עובדים עם `hookline-admin`, ישר מול המסד.

לבסוף, מודל האיומים. הקובץ `docs/threat-model.md` בנוי מהתבנית של פרק 28. יש בו תרשים זרימה עם גבולות האמון, טבלת נכסים,
ואחד־עשר איומים לפי STRIDE. לכל איום יש הגנה, הבדיקה שמוכיחה אותה, והסיכון שנשאר:

| # | האיום | ההגנה | ההוכחה |
| --- | --- | --- | --- |
| T1 | ‏SSRF | בדיקה ביצירה ובחיבור, ושגיאה כללית | ארבע בדיקות `ssrf`, ‏`api` ו־`delivery` |
| T2 | ‏BOLA | הדיירת בכל שאילתה, ו־404 אחד | `TestTenantIsolation`, ‏`TestReplayIsTenantScoped` |
| T3 | שליחה חוזרת של webhook חתום | חותמת זמן, 300 שניות, וסינון לפי מזהה האירוע | וקטורי החתימה, ובדיקות ה־sink |
| T4 | סודות בלוג | `redact.Handler` לפני כל פלט | שתי סריקות הלוגים |
| T5 | שכנה רועשת | הגבלת קצב, השלת עומס, מחיצה לכל endpoint | `TestTenantOverItsLimitGets429OthersDoNot`, ו־k6 |
| T6 | הרעלת ה־broker | רשומה שלא מתפענחת נרשמת ומדולגת | `TestPoisonRecordIsSkipped` |

השאר הם גניבת מפתח, שרשרת האספקה, הכחשה, גופים ענקיים, וגיבוי של המסד. האחרון הוא הסיכון הגבוה שנשאר: סודות
החתימה שמורים בטקסט גלוי, כי ה־worker צריך אותם לחישוב ה־HMAC. הוא רשום עם ההחלטות, ובמשימות ההמשך. קראי את
המודל, ומלאי את הכותרת. אחר כך הוסיפי איום אחד שלא חשבנו עליו, עם הבדיקה שלו.

**יום 86: חוזה, תכונות, עומס ושרשרת אספקה.** הגוף של ה־webhook הוא חוזה עם קוד שלא כתבת. הקובץ
`contract/webhook-event.schema.json` הוא הסכמה של פרק 26, מותאמת לגוף של Hookline:

| שדה | הכלל | השינוי מפרק 26 |
| --- | --- | --- |
| `id` | ‏uuid | אותו דבר; ה־uuidv7 של Hookline הוא uuid רגיל |
| `type` | מילים קטנות עם נקודות, עד 128 תווים | מילה אחת מותרת, כמו בקליטה |
| `created_at` | ‏RFC 3339 | אותו דבר |
| `payload` | אובייקט | השם `data` הפך ל־`payload`, כמו ב־API |

שני הצדדים בודקים את אותן דוגמאות, ואף צד לא מייבא את השני:

| צד | מה נבדק | הפרט שחשוב |
| --- | --- | --- |
| ‏Go | כל הדוגמאות, וגם הגוף ש־`Job.Body` מייצר | בלי `AssertFormat`, ‏`evt_123` עובר כ־uuid |
| ‏Python | אותן דוגמאות, וה־sink מקבל כל דוגמה תקינה כשהיא חתומה | בלי `format_checker`, אותה בעיה |

הקבצים הם כל ההסכם.

שתי בדיקות התכונה החדשות הולכות בדרך של פרק 26:

- **הקליטה והסכמה מסכימות.** לכל מחרוזת סוג, הקליטה מקבלת אותה בדיוק כשהסכמה מקבלת גוף איתה. חצי מההגרלות תואמות
  את התבנית, וחצי הן מחרוזות שרירותיות.
- **סמן העימוד חוזר לאותו מקום.** כל סמן שה־API מחזיר מתפענח לאותו פריט. כאן מצאה הספרייה rapid, בהרצה הראשונה,
  את תחילת שנת 1970: הפענוח קורא זמן אפס כזמן חסר. אף פריט לא נוצר אז, ולכן התחום מתחיל שנייה אחרי, עם הערה.

<div dir="ltr" align="left">

```bash
go test -race ./contract/ ./internal/page/
make sink-test
```

</div>

עכשיו העומס. שני התרחישים בנויים כמו `shorty.js` של פרק 26: ‏constant-arrival-rate, חימום מתויג שאף סף לא קורא,
וספים שהופכים את ההרצה לבדיקה:

| תרחיש | פרופיל | הספים |
| --- | --- | --- |
| `loadtest/ingest.js` | ‏`RATE` אירועים בשנייה (ברירת מחדל 100), דקה אחרי עשר שניות חימום | ‏p99 מתחת ל־200 מילישניות, פחות מ־1% כשלים, אפס איטרציות שנפלו |
| `loadtest/two-tenants.js` | רועשת שולחת פי חמישה מהמגבלה שלה, ושקטה שולחת 20 בשנייה | לרועשת יש 429 עם `Retry-After`; לשקטה אין אף 429, וה־p99 שלה מתחת ל־200 |

זה גוש הספים של התרחיש השני, מהקובץ `loadtest/two-tenants.js`:

<div dir="ltr" align="left">

```js
  thresholds: {
    // The noisy tenant is limited, and every limited answer says when to come back.
    'ingest_429{tenant:noisy}': ['count>0'],
    'checks{tenant:noisy}': ['rate==1'],
    // The quiet tenant is not affected: never limited, every event accepted, and as fast as N1 asks.
    'ingest_429{tenant:quiet}': ['count==0'],
    'checks{tenant:quiet}': ['rate==1'],
    'http_req_duration{tenant:quiet}': ['p(99)<200'],
    'dropped_iterations': ['count==0'],
  },
```

</div>

כל דגימה מתויגת בשם הדיירת שלה, ולכן כל סף שופט דיירת אחת. האפשרות `responseCallback` מסמנת 429 כתשובה צפויה,
ולא ככשל. הסקריפט `scripts/load.sh` מעלה את כל הסביבה, כולל הטלמטריה. הוא יוצר דיירות עם מגבלות ידועות, ומריץ את
התרחיש מולן. כך אפשר לראות את לוח ה־RED בזמן ההרצה.

<div dir="ltr" align="left">

```bash
make load-ingest
make load-two-tenants
```

</div>

שימי לב ל־`preAllocatedVUs`. מספר הבקשות באוויר הוא הקצב כפול זמן התגובה: 100 בשנייה ב־200 מילישניות הן 20.
אם הקצאת בדיוק 20, עיכוב קצר אחד גורם ל־k6 להפיל איטרציות, ואת זה ה־API בכלל לא ראה. התרחישים מקצים חצי שנייה של
בקשות. ועוד דבר: על מחשב עמוס, עיכובים כאלה הם של המחשב ולא של Hookline. אם סף נכשל, בדקי את העומס על המחשב,
והריצי שוב כשהוא שקט. את המספרים, עם התאריך והמחשב, רשמי ב־`docs/EVALUATION.md`.

ולסיום, שרשרת האספקה:

<div dir="ltr" align="left">

```bash
make images
make supply-chain
```

</div>

הסקריפט `scripts/supply-chain.sh` עושה שלושה דברים:

1. מריץ את govulncheck על המודול, שנכשל רק על חולשה שהקוד באמת קורא לה.
2. שומר כל image לקובץ, ומפיק ממנו רשימת חבילות עם Syft.
3. סורק את ה־SBOM עם Grype, ונכשל על ממצא בחומרה גבוהה ומעלה שיש לו תיקון.

ממצאים בלי תיקון מופיעים בטבלה, ולא מכשילים. זו המדיניות של פרק 19: שער שנכשל על מה שאי אפשר לתקן נכשל כל הזמן,
והצוות לומד לעקוף אותו. ממצא עם תיקון בחבילה של תמונת הבסיס אומר שצריך בסיס חדש יותר, ולא סף נמוך יותר.

הסורקים רצים בקונטיינרים, בלי גישה לשקע של Docker. גם ה־workflow השתנה:

| ‏job | הצעד החדש |
| --- | --- |
| ‏unit | ‏govulncheck על המודול |
| ‏images | אחרי הדחיפה, ‏Syft קורא את ה־image לפי ה־digest, וה־SBOM נשמר כ־artifact |

<div dir="ltr" align="left">

```bash
cp -r hookline/.github/. .github/ && actionlint
```

</div>

שתי החלטות נוספות, שגם אותן כדאי לנמק. הספרייה testcontainers לא נכנסה. החבילה `testdb` מתחברת למסד של
ה־compose. גם ה־CI מריץ את אותו קובץ, כמו המחשב שלך. כך יש סביבה אחת ולא שתיים. גם Trivy לא נכנס לסקריפט, כי
סורק כללי אחד מספיק כאן. כדעה שנייה, הוא הרחבה טובה.

### 38.9.5 הבדיקות שצריכות להוריק

<div dir="ltr" align="left">

```bash
DATABASE_URL= go test -race ./internal/ssrf/ ./internal/redact/ ./internal/page/ ./contract/ ./internal/api/
make up
go test -race -count=1 ./...
make sink-test
make load-ingest
make load-two-tenants
make down
make images
make supply-chain
```

</div>

### 38.9.6 הפלט הצפוי

<div dir="ltr" align="left">

```text
$ go test -race ./internal/ssrf/ ./internal/redact/ ./internal/page/ ./contract/
ok  	example.com/hookline/internal/ssrf	...
ok  	example.com/hookline/internal/redact	...
ok  	example.com/hookline/internal/page	...
ok  	example.com/hookline/contract	...
$ make load-two-tenants
...
  █ THRESHOLDS
    checks{tenant:noisy}
    ✓ 'rate==1' rate=100.00%
    checks{tenant:quiet}
    ✓ 'rate==1' rate=100.00%
    dropped_iterations
    ✓ 'count==0' count=___
    http_req_duration{tenant:quiet}
    ✓ 'p(99)<200' p(99)=___
    ingest_429{tenant:noisy}
    ✓ 'count>0' count=___
    ingest_429{tenant:quiet}
    ✓ 'count==0' count=0
...
$ make supply-chain
== govulncheck (the module's code, only what it calls)
...
No vulnerabilities found.
== hookline:dev
sbom   /tmp/hookline-supply-chain/hookline.spdx.json (___ SPDX elements)
grype  /tmp/hookline-supply-chain/hookline.grype.txt
== hookline-sink:dev
...
```

</div>

המספרים במקום הקווים הם שלך. השורה `No vulnerabilities found` נכונה ליום שבו הספר נכתב. המאגר מתעדכן, ומחר אולי תהיה
שורה. אז מעדכנים את התלות, ולא את הסף.

בהרצה שבה נכתב הספר, ה־image של Hookline יצא נקי. בתמונת הבסיס של ה־sink נמצא ממצא אחד בחומרה גבוהה שיש לו
תיקון ב־Debian, והשער נכשל עליו. התיקון עוד לא הגיע לתמונת הבסיס. הצעד הנכון הוא לחכות לבנייה החדשה שלה, ואז לבנות
מחדש עם `docker build --pull`. רשמי מה מצאת אצלך, ומה החלטת.

### 38.9.7 קריטריון סיום

- endpoint שהשם שלו מתפענח לכתובת פרטית נדחה ביצירה, וחיבור לכתובת כזו נחסם גם בזמן השליחה.
- שתי סריקות הלוגים ירוקות, ואין מפתח או סוד באף פלט.
- הסכמה נבדקת משני הצדדים, ובדיקות התכונה ירוקות.
- שני תרחישי k6 רצו, והמספרים שלך רשומים ב־`docs/EVALUATION.md`, עם תאריך ומחשב.
- הסריקה של govulncheck נקייה, ולכל image יש רשימת חבילות וסריקה.
- מודל האיומים כולל את השורה שלך, וה־workflow עובר actionlint. קומיט של אבן הדרך, ו־`make down`.

### 38.9.8 טעויות נפוצות

- **לבדוק את הכתובת רק ביצירה.** השם יכול להתפענח לכתובת אחרת בזמן השליחה. רק הבדיקה ב־dialer רואה את הכתובת
  שמתחברים אליה.
- **לספר לדיירת למה הכתובת נדחתה.** תשובה מפורטת הופכת את ה־endpoint לסורק של הרשת שלנו.
- **לקבל proxy מהסביבה.** עם `HTTPS_PROXY`, הבדיקה רואה את שרת הביניים, והיעד עובר בלי בדיקה.
- **מצב פיתוח רחב מדי.** להתיר את כל 10.0.0.0/8 נוח, ופותח בייצור את המסד. מתירים רק את הטווחים שצריך.
- **הסתרה רק בפלט של JSON.** פלט שני, כמו הגשר ל־Loki, עוקף אותה. ההסתרה עוטפת את כולם.
- **סכמה בלי בדיקת פורמט.** בלי `AssertFormat` בצד של Go, כל מחרוזת עוברת. ב־Python זה `format_checker`.
- **להוריד את הסף כשהסריקה נכשלת.** ממצא עם תיקון דורש עדכון של התלות או של תמונת הבסיס.
- **מעט מדי VUs מוקצים.** הכלי מפיל איטרציות, וההרצה מחמיאה ל־API שלא ראה אותן.

### 38.9.9 הרחבה

- החלפת מפתח API בלי השבתה: שני גיבובים תקפים בחלון זמן, ופקודה ב־`hookline-admin`.
- החלפת סוד חתימה: חתימה בשני הסודות בחלון זמן, עם שני ערכי `v1=` בכותרת.
- תרחיש k6 מעל הקיבולת, שמראה את N6: תשובות 503 מההשלה, בלי תור שגדל בלי גבול.
- סריקה שנייה עם Trivy, והשוואה בין הממצאים של שני הסורקים.

## 38.10 M10: הערכה והדגמה

יושלם עם אבן הדרך.

## סיכום

- אבן הדרך הראשונה: ‏API מלא מעל מאגר בזיכרון, מאחורי ממשקים, עם אימות, עימוד, פרטי בעיה ובדיקות בידוד.
- אבן הדרך השנייה: אותם ממשקים מעל PostgreSQL, עם מיגרציות לכל מודל הנתונים וקליטה שהאילוץ מכריע בה.
- אבן הדרך השלישית: מטמון של מנויים, רשומה אחת לכל דיירת, שה־API מוחק אחרי כל שינוי; וה־sink ב־compose.
- אבן הדרך הרביעית: ‏relay שמפרסם את ה־outbox לפי הסדר, ו־dispatcher שמפזר כל אירוע לשורות מסירה בלי כפילויות.
- אבן הדרך החמישית: ‏workers שתופסים מסירות בחכירה, שולחים חתום, מנסים שוב רק כשזה הגיוני, ומתים כשצריך.
- הפעלה חוזרת לא מאפסת את מספר הניסיונות; העמודה `attempt_base` מסמנת איפה הסבב החדש מתחיל.
- אבן הדרך השישית: מפסק ומחיצה לכל endpoint שדוחים מסירה בלי לבזבז ניסיון, הגבלת קצב לכל דיירת שנכשלת פתוח,
  השלת עומס לפני האימות, ו־relay יחיד פעיל שהקדנציה שלו מגדרת את הסימון.
- אבן הדרך השביעית: ‏image אחד לחמשת הבינאריים ו־image ל־sink, ‏chart עם מיגרציות ב־hook ו־Secret מבחוץ, ‏Terraform
  שמתקין את אותו chart, ו־workflow שמריץ את הכול, כולל kind.
- אבן הדרך השמינית: מעקב אחד מהקליטה ועד המקבל דרך עמודה במסד, מדדים של מסירה, ‏SLO של פיגור עם התראה שנייה
  על כשלים, ‏runbook לכל התראה, ותחקיר על גרסה ששברת בכוונה.
- אבן הדרך התשיעית: כתובת שנבדקת ביצירה ושוב בכל חיבור, הסתרת סודות לפני כל פלט עם סריקת לוגים שמוכיחה אותה, חוזה
  לגוף ה־webhook שנבדק משני הצדדים, בדיקות תכונה ועומס, סריקה של שרשרת האספקה, ומודל איומים.
- הסיכום יושלם עם אבן הדרך האחרונה.

## מונחים לזכור

| מונח בעברית | באנגלית | במשפט אחד |
| --- | --- | --- |
| אבן דרך | milestone | שלב בפרויקט שנגמר כשקבוצת בדיקות מוגדרת ירוקה |
| בדיקות התאמה | conformance tests | סוויטה אחת שכל מימוש של ממשק חייב לעבור |
| ‏keyset | keyset pagination | עימוד לפי ערכי המיון של הפריט האחרון במקום לפי היסט |
| אי־תוקף אחרי שמירה | invalidate after commit | מחיקת רשומת מטמון רק אחרי שהשינוי נשמר, כדי שקורא לא יחזיר ערך ישן |
| פיזור | fan-out | הפיכת אירוע אחד לשורת מסירה לכל יעד מתאים, בפקודה אחת |
| הודעה רעילה | poison message | רשומה שאי אפשר לעבד לעולם; מדלגים עליה כדי לא לחסום את המחיצה |
| אסימון גידור | fencing token | מספר שעולה בכל תפיסה, כך שתוצאה של מחזיק ישן נדחית |
| סבב ניסיונות | attempt round | הניסיונות מאז ההפעלה החוזרת האחרונה, שמושווים למקסימום |
| דחייה | postpone | החזרת מסירה לתור בלי לשלוח ובלי לספור ניסיון, כשהמפסק פתוח או המחיצה מלאה |
| כשל פתוח | fail open | כשהמגביל עצמו לא זמין, הבקשה עוברת; ההחלטה ההפוכה הייתה מפילה את הקליטה |
| קדנציה | term | מספר שעולה בכל בחירת מנהיג, ומשמש כ־fencing token |
| פיגור מסירה | delivery lag | הזמן מקליטת האירוע ועד שהמקבל אישר אותו; היעד של N2 |
| ‏hook של Helm | Helm hook | משאב שרץ בנקודה במחזור של release, למשל Job לפני כל התקנה |
| בדיקת עשן | smoke test | בדיקה קצרה מקצה לקצה שמראה שהמערכת בכלל עובדת אחרי פריסה |
| הקשר מעקב | trace context | מזהה המעקב וה־span שעוברים בין תהליכים, כאן בכותרת `traceparent` |
| התראה על סיבה | cause-based alert | התראה על אות פנימי, למשל כשלים או outbox תקוע, שה־SLO רואה מאוחר או בכלל לא |
| בדיקה בזמן החיבור | dial-time check | בדיקה של הכתובת שמתחברים אליה בפועל, אחרי הפענוח; רק היא עומדת מול DNS rebinding |
| סריקת לוגים | log scan | בדיקה שמריצה את הקוד עם ה־logger האמיתי, ומוודאת שסוד ידוע לא מופיע בפלט |
| בדיקת חוזה | contract test | בדיקה של שני צדדים מול אותה סכמה ואותן דוגמאות, בלי שאחד מייבא את השני |
| ‏SBOM | software bill of materials | רשימת כל החבילות ב־image, שסורקים אותה גם אחרי שה־image כבר נדחף |

הטבלה תושלם עם אבן הדרך האחרונה.

## טעויות נפוצות

- **לסמן אבן דרך כגמורה כשרק חלק מהבדיקות שלה ירוקות.** הקריטריון הוא כל הפקודה של הסעיף, עם `-race`.
- **לבדוק את הצינור רק ברכיבים בודדים.** כל רכיב יכול לעבור לבד, והחיבור ביניהם עדיין שבור. הבדיקה מקצה לקצה של
  אבן הדרך החמישית היא שמוכיחה את ההבטחה.
- **לסמוך על Kafka לסינון כפילויות.** ה־offset של פרסום כפול שונה. הכפילות נעצרת באילוץ במסד, ובמקבל לפי מזהה האירוע.
- **לנסות שוב כל כישלון.** תשובת 4xx סופית; ניסיון חוזר שלה רק מאריך את התור של כולם.
- **‏SLO שבנוי רק מהצלחות.** כשאף דבר לא מצליח אין לו נתונים, והוא שותק בדיוק בזמן התקלה.
- **סוד ב־values של chart.** הוא נשמר בהיסטוריה של ה־release; ה־chart מקבל שם של Secret, לא את הערך.
- **הגנה שעולה יותר ממה שהיא חוסכת.** מחיצה שמחכה, או השלת עומס אחרי קריאה למסד, מחזירות את העומס מהדלת האחורית.
- **הגנה בלי בדיקה שמוכיחה אותה.** שורה במודל האיומים בלי שם של בדיקה היא תקווה, ולא הגנה.
- הרשימה תושלם עם אבן הדרך האחרונה.

## שאלות ראיון

1. **למה המאגר יושב מאחורי ממשק כבר באבן הדרך הראשונה, כשהוא עוד בזיכרון?** כי ההחלפה ל־PostgreSQL באבן הדרך
   השנייה נעשית בלי לגעת ב־handler, ואותן בדיקות API רצות על שני המאגרים. הממשק גם מכריח להחליט מוקדם מה
   החוזה: מי מסנן לפי דיירת, מה השגיאות. המחיר הוא שכבה אחת נוספת, וזה זול לעומת שכתוב של כל ה־handlers.

2. **איך מבטיחים שקריסה של ה־relay לא תיצור מסירה כפולה?** לא מבטיחים שהיא לא תפרסם פעמיים; היא כן תפרסם, כי
   קריסה בין הפרסום לסימון משאירה את השורה לא מסומנת. מבטיחים שהפרסום הכפול לא משנה כלום: ה־dispatcher מכניס
   עם `ON CONFLICT DO NOTHING` על הזוג אירוע ו־endpoint. ומול המקבל, שבו אין אילוץ משותף, החוזה הוא „לפחות פעם
   אחת” וסינון לפי `Hookline-Event-Id`.
   **המשך:** ולמה לא לסמן לפני הפרסום? כי אז קריסה בין הסימון לפרסום מאבדת את האירוע לגמרי; עדיף כפילות שמסננים
   על פני אובדן שאי אפשר לגלות.

3. **מה קורה כש־worker נתקע באמצע מסירה?** החכירה שלו פגה, ו־worker אחר תופס את המסירה עם מספר ניסיון גבוה יותר.
   הניסיון התקוע נרשם כ„חכירה פגה”. אם הראשון מתעורר ומנסה לרשום תוצאה, הרישום בודק את מספר הניסיון ונדחה. המחיר
   האפשרי הוא שהמקבל קיבל שתי בקשות, וזה בסדר בחוזה „לפחות פעם אחת”. לכן החכירה חייבת להיות ארוכה מהזמן הקצוב של
   ניסיון.

4. **למה לא לאפס את מונה הניסיונות בהפעלה חוזרת?** כי ההיסטוריה ממוספרת לפי אותו מונה, עם מפתח ייחודי, והמקבל רואה
   את המספר בכותרת. איפוס יתנגש בהיסטוריה ויבלבל את המקבל. במקום זה שומרים את נקודת ההתחלה של הסבב בעמודה נפרדת,
   ומשווים למקסימום את ההפרש. זו גם ההצעה במסמך התכנון של פרק 29.

5. **מקבל אחד של דיירת אחת תקוע. איך שאר המקבלים לא מרגישים?** שתי הגנות בכל worker. מחיצה אטומה מגבילה כמה workers
   endpoint אחד יכול להחזיק, וכשהיא מלאה המסירה נדחית בלי לחכות. מפסק זרם מפסיק לשלוח אחרי כמה כשלים רצופים, ובודק
   שוב אחרי זמן. בשני המקרים הדחייה לא סופרת ניסיון. ההגנות מקומיות לכל תהליך: פשוט יותר ונכשל בבטחה, במחיר
   שכל worker לומד לבד. את המחיר הזה מודדים בניסוי N3, עם ההגנות ובלעדיהן.
   **המשך:** למה לא מפסק משותף ב־Redis? כי אז Redis נעשה תלות של כל שליחה, ותקלה בו משפיעה על כל המקבלים; הייתי
   עוברת אליו רק אם המדידה מראה שהרבה workers מכים endpoint מת לפני שהמפסקים המקומיים נפתחים.

6. **איך את מריצה מיגרציות בפריסה ל־Kubernetes, ומה יכול להשתבש?** ‏Job של Helm ב־`pre-install` ו־`pre-upgrade`, כך
   שהסכמה מוכנה לפני שאף pod חדש עולה, עם פקודה אידמפוטנטית. שני דברים לשים לב אליהם: ה־Job לא רואה משאבים
   שה־chart עוד לא התקין, והמיגרציה חייבת להתאים גם לגרסה הישנה שעוד רצה, כי בעדכון מתגלגל שתיהן חיות יחד
   (פרק 30). לכן מיגרציה שמוחקת עמודה נכנסת רק אחרי שאף גרסה לא קוראת אותה.

7. **איך עוקבים אחרי אירוע שעובר דרך תור, כשכל תהליך מטפל בו בזמן אחר?** ההקשר לא יכול לעבור בזיכרון, ולכן הוא
   עובר עם הנתונים: ב־Hookline הכותרת `traceparent` של בקשת הקליטה נשמרת בשורת האירוע, ה־relay ממשיך ממנה, ‏kotel
   כותב אותה לכותרות של הרשומה ב־Kafka, וה־worker קורא אותה שוב מהאירוע. המחיר הוא עמודה ועוד קוד בכל קפיצה.
   היתרון: ניסיון חוזר אחרי עשר דקות עדיין באותו מעקב, וכשמשהו נכשל רואים את כל הדרך.

8. **השירות שלך שולח בקשות לכתובות שהלקוחות בוחרים. איך את מגינה עליו?** בודקת פעמיים. ביצירה, את ה־URL ואת כל
   הכתובות שהשם מתפענח אליהן, עם תשובה כללית ללקוח. ובכל חיבור, ב־`Control` של ה־dialer, את הכתובת שמתחברים אליה
   בפועל. השנייה היא שעומדת מול DNS rebinding, כי בין הרישום לשליחה השם יכול להצביע למקום אחר. בנוסף, אין proxy
   מהסביבה, אין הפניות, ומצב הפיתוח מתיר רק טווחים מוגדרים.
   **המשך:** ולמה לא רשימה של כתובות מותרות? כי היעדים הם של הלקוחות, ואי אפשר לדעת אותם מראש. לכן חוסמים את מה
   שמסוכן, ומשאירים את הבדיקה בחיבור עצמו.

## פתק Obsidian

הפתק המוכן נמצא גם כקובץ: `obsidian/38-hookline-milestones.md`. זה תוכנו:

- **‏M1:** ממשקי מאגר; מימוש בזיכרון; אימות לפי גיבוב; 404 לכל משאב של דיירת אחרת.
- **‏M2:** מיגרציות לכל המודל; ‏`ON CONFLICT DO NOTHING` ואז `SELECT`; אירוע ו־outbox בטרנזקציה אחת.
- **‏M3:** מטמון לכל דיירת; מוחקים אחרי השמירה; זמן החיים הוא הגבול לרשימה ישנה; ה־sink עם worker אחד.
- **‏M4:** פרסום, אישור, ואז סימון; ‏offset אחרי ההכנסה; ‏`ON CONFLICT DO NOTHING` על אירוע ו־endpoint.
- **‏M5:** ‏`SKIP LOCKED` וחכירה; 408, 429 ו־5xx מנוסים שוב, שאר 4xx לא; ‏`attempt_base` בהפעלה חוזרת.
- **‏M6:** מחיצה בלי המתנה ומפסק לכל endpoint; דחייה בלי ניסיון; מגבלה לכל דיירת, נכשלת פתוח; השלה לפני האימות.
- **‏M7:** ‏image אחד עם כל הבינאריים; משתמש מספרי; מיגרציות ב־hook; ‏Secret מבחוץ; ‏compose גם ב־CI.
- **‏M8:** ‏`traceparent` בשורת האירוע; ‏SLI של פיגור על גבול 30; התראה על כשלים כי ה־SLI שותק; תחקיר עם מספרים שלי.
- **‏M9:** ‏SSRF נבדק ביצירה ובחיבור; הסתרה לפני כל פלט, וסריקה; סכמה משני הצדדים; ‏VUs לפי קצב כפול זמן תגובה.
- **שאלה פתוחה:** ___

<div dir="ltr" align="left">

```text
M1 tests  DATABASE_URL= go test -race -skip 'Invalidat|Deliver|Replay|ActiveSubscriptions|RateLimit|OverItsLimit|LimiterFailure|Shed|Trace|PrivateAddress' ./internal/api/ ./internal/store/... ./internal/admin/
M2 db     make db-up && export DATABASE_URL=postgres://hookline:change-me@localhost:5438/hookline
M2 tests  go test -race -count=1 -skip 'Invalidat|Deliver|Replay|ActiveSubscriptions|RateLimit|OverItsLimit|LimiterFailure|Shed|Trace|PrivateAddress' ./internal/api/ ./internal/store/... ./internal/admin/ ./internal/db/ ./migrations/
admin     go run ./cmd/hookline-admin migrate up | create-tenant -name <name> | create-topic -partitions 6
stack     make up   (postgres, redis, redpanda, console :8098, sink :8091)   make down
env       REDIS_ADDR=localhost:6438 KAFKA_BROKERS=localhost:19438 SINK_URL=http://localhost:8091
run       make run | make relay | make dispatcher | make worker
sink      curl -s localhost:8091/stats    curl -s -X POST localhost:8091/control -d '{"failure_rate":1}'
retry     408 429 5xx timeout refused -> retry      other 4xx 3xx -> dead      replay: dead -> pending
m7        make images | make chart-check | KEEP=1 make kind-smoke | make tf-check | make kind-down
kube      export KUBECONFIG=/tmp/hookline-kind.kubeconfig; kubectl -n hookline get pods
ci        cp -r hookline/.github/. .github/ && actionlint
m8        make obs-up | make promtool | make check-telemetry | make stage-incident | make obs-down
grafana   http://localhost:3000 (folder Hookline)   prometheus :9090   tempo :3200   loki :3100
m9        make load-ingest | make load-two-tenants | make images && make supply-chain   (RATE= NOISY_RATE= QUIET_RATE=)
m9 tests  DATABASE_URL= go test -race ./internal/ssrf/ ./internal/redact/ ./internal/page/ ./contract/ ./internal/api/
lab       make lab-n3      GUARDS=off make lab-n3      (record your p50/p99, date, machine)
guards    HOOKLINE_ENDPOINT_CONCURRENCY=2 HOOKLINE_BREAKER_FAILURES=5 HOOKLINE_BREAKER_OPEN=30s HOOKLINE_MAX_IN_FLIGHT=256
m5 tests  go test -race -count=1 -skip 'RateLimit|OverItsLimit|LimiterFailure|Shed|Breaker|Bulkhead|Trace|PrivateAddress' ./internal/backoff/ ./internal/delivery/ ./internal/api/ ./internal/store/... ./internal/e2e/
skip      -skip drops later milestones' tests: Invalidat ActiveSubscriptions (M3), Deliver Replay (M5), StaleRelay RateLimit OverItsLimit LimiterFailure Shed Breaker Bulkhead (M6), Trace (M8), PrivateAddress (M9)
```

</div>

## תרגיל

**מטרה:** לבנות את Hookline אבן דרך אחר אבן דרך, מה־starter ועד הדגמה, כשכל אבן דרך נגמרת בבדיקות ירוקות ובקומיט.
**זמן:** עשר אבני דרך לאורך הספר, לפי הסדר: ימים 22 ו־23, ימים 29 ו־30, ימים 33 ו־34, ימים 39 ו־40, ימים 43 ו־44, ימים 54 ו־55, ימים 68 עד 70, ימים 77 עד 79, ימים 85 ו־86, ימים 98 ו־99; כ־55 עד 90 דקות בנייה בכל יום.
**ידע מקדים:** פרק 37, וכל סעיף כאן מפרט את הפרקים שאבן הדרך שלו נשענת עליהם.

ההוראות של הפרויקט ורשימת אבני הדרך נמצאות ב־
[`exercises/project-hookline/README.md`](../exercises/project-hookline/README.md).

**קריטריון סיום:**

- כל אבן דרך: הפקודה של סעיף הבדיקות שלה ירוקה, וקומיט עם שם אבן הדרך.
- אחרי אבן הדרך החמישית: ‏`TestPipelineEndToEnd` עוברת מול ה־compose, וארבעת הבינאריים מוסרים אירועים ל־sink.
- אחרי אבן הדרך השישית: ניסוי N3 רץ בשתי הגרסאות, והמספרים שמדדת רשומים, עם תאריך ומחשב.
- אחרי אבן הדרך השביעית: ‏`make kind-smoke` מסתיים ב־PASS, וה־workflow ירוק במאגר שלך ב־GitHub.
- אחרי אבן הדרך השמינית: מעקב שלם ב־Tempo, ‏promtool ירוק, והתחקיר שלך כתוב עם המספרים שמדדת.
- אחרי אבן הדרך התשיעית: סריקות הלוגים והחוזה ירוקות, שני תרחישי k6 רשומים, ו־govulncheck נקי.
- הרשימה המלאה תושלם עם אבן הדרך האחרונה.

## העשרה אופציונלית

- המסמך של Standard Webhooks על ניסיונות חוזרים ועל קודי תשובה: השווי את המדיניות שלו ל־`Classify`.
- המאמר של Marc Brooker מ־2015 על המתנה גדלה ופיזור אקראי: המקור של „full jitter” שה־backoff מממש.
- יושלם עם אבני הדרך הבאות.

</div>
