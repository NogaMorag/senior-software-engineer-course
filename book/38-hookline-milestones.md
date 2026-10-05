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
DATABASE_URL= go test -race -skip 'Invalidat|Deliver|Replay|ActiveSubscriptions|RateLimit|OverItsLimit|LimiterFailure|Shed' ./internal/api/ ./internal/store/... ./internal/admin/
```

</div>

ההשמה הריקה של `DATABASE_URL` מבטיחה שבדיקות ה־API רצות על המאגר בזיכרון, גם אם הגדרת את המשתנה בחלון אחר.
הדגל `-skip` מדלג על הבדיקות של אבני דרך מאוחרות שכבר נמצאות באותן חבילות: המטמון של אבן הדרך השלישית
(`Invalidat`, ‏`ActiveSubscriptions`), המסירות של החמישית (`Deliver`, ‏`Replay`), וההגנות של השישית (`RateLimit`, ‏`Shed` ועוד שתיים).
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
go test -race -count=1 ./internal/api/
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
			INSERT INTO events (id, tenant_id, type, payload, idempotency_key, body_hash, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (tenant_id, idempotency_key) DO NOTHING
			RETURNING `+eventColumns,
			e.ID, e.TenantID, e.Type, []byte(e.Payload), e.IdempotencyKey, e.BodyHash[:], e.CreatedAt)
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
go test -race -count=1 -skip 'Invalidat|Deliver|Replay|ActiveSubscriptions|RateLimit|OverItsLimit|LimiterFailure|Shed' ./internal/api/ ./internal/store/... ./internal/admin/ ./internal/db/ ./migrations/
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
go test -race -count=1 -skip 'Deliver|Replay|RateLimit|OverItsLimit|LimiterFailure|Shed' ./internal/cache/ ./internal/subcache/ ./internal/api/ ./internal/store/...
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
go test -race -count=1 -skip 'StaleRelay' ./internal/outbox/
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

- **פקודה אחת לכל האירוע.** כל השורות של אירוע נכנסות יחד, או אף אחת. אין מצב של חצי פריסה.
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
go test -race -count=1 -skip 'StaleRelay' ./internal/outbox/ ./internal/dispatch/ ./internal/bus/
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
DATABASE_URL= go test -race -run 'Vectors|Classify|Send' ./internal/delivery/
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
go test -race -count=1 -skip 'RateLimit|OverItsLimit|LimiterFailure|Shed|Breaker|Bulkhead' ./internal/backoff/ ./internal/delivery/ ./internal/api/ ./internal/store/... ./internal/e2e/
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
go test -race -count=1 -skip 'RateLimit|OverItsLimit|LimiterFailure|Shed|Breaker|Bulkhead' ./internal/backoff/ ./internal/delivery/ ./internal/api/ ./internal/store/... ./internal/e2e/
go test -race -count=1 -skip 'StaleRelay' ./internal/outbox/ ./internal/dispatch/ ./internal/bus/ ./internal/cache/ ./internal/subcache/
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
			p.postpone(ctx, j, busyWait(), err)
			return Result{}, false
		}
		defer g.bulkhead.Release()
	}
	done, err := g.breaker.Allow()
	if err != nil {
		p.postpone(ctx, j, p.Guards.openWait(), err)
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
`00006`, ולא ב־`EnsureSchema`. ה־relay עוטף את הלולאה שלו בבחירה: מי שמחזיק את הנעילה המייעצת מפרסם, והשאר מחכים.

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
go test -race -count=1 ./internal/breaker/ ./internal/bulkhead/ ./internal/ratelimit/ ./internal/shed/ \
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

יושלם עם אבן הדרך.

## 38.8 M8: תצפיתיות ו־SLO

יושלם עם אבן הדרך.

## 38.9 M9: בדיקות ואבטחה

יושלם עם אבן הדרך.

## 38.10 M10: הערכה והדגמה

יושלם עם אבן הדרך.

## סיכום

- אבן הדרך הראשונה: ‏API מלא מעל מאגר בזיכרון, מאחורי ממשקים, עם אימות, עימוד, פרטי בעיה ובדיקות בידוד.
- אבן הדרך השנייה: אותם ממשקים מעל PostgreSQL, עם מיגרציות לכל מודל הנתונים וקליטה שהאילוץ מכריע בה.
- אבן הדרך השלישית: מטמון של מנויים, רשומה אחת לכל דיירת, שה־API מוחק אחרי כל שינוי; וה־sink ב־compose.
- אבן הדרך הרביעית: ‏relay שמפרסם את ה־outbox לפי הסדר, ו־dispatcher שפורש כל אירוע לשורות מסירה בלי כפילויות.
- אבן הדרך החמישית: ‏workers שתופסים מסירות בחכירה, שולחים חתום, מנסים שוב רק כשזה הגיוני, ומתים כשצריך.
- הפעלה חוזרת לא מאפסת את מספר הניסיונות; העמודה `attempt_base` מסמנת איפה הסבב החדש מתחיל.
- אבן הדרך השישית: מפסק ומחיצה לכל endpoint שדוחים מסירה בלי לבזבז ניסיון, הגבלת קצב לכל דיירת שנכשלת פתוח,
  השלת עומס לפני האימות, ו־relay יחיד פעיל שהקדנציה שלו מגדרת את הסימון.
- הסיכום יושלם עם אבני הדרך הבאות.

## מונחים לזכור

| מונח בעברית | באנגלית | במשפט אחד |
| --- | --- | --- |
| אבן דרך | milestone | שלב בפרויקט שנגמר כשקבוצת בדיקות מוגדרת ירוקה |
| בדיקות התאמה | conformance tests | סוויטה אחת שכל מימוש של ממשק חייב לעבור |
| ‏keyset | keyset pagination | עימוד לפי ערכי המיון של הפריט האחרון במקום לפי היסט |
| אי־תוקף אחרי שמירה | invalidate after commit | מחיקת רשומת מטמון רק אחרי שהשינוי נשמר, כדי שקורא לא יחזיר ערך ישן |
| פריסה לשורות מסירה | fan-out | הפיכת אירוע אחד לשורה לכל יעד מתאים, בפקודה אחת |
| הודעה רעילה | poison message | רשומה שאי אפשר לעבד לעולם; מדלגים עליה כדי לא לחסום את המחיצה |
| אסימון גידור | fencing token | מספר שעולה בכל תפיסה, כך שתוצאה של מחזיק ישן נדחית |
| סבב ניסיונות | attempt round | הניסיונות מאז ההפעלה החוזרת האחרונה, שמושווים למקסימום |
| דחייה | postpone | החזרת מסירה לתור בלי לשלוח ובלי לספור ניסיון, כשהמפסק פתוח או המחיצה מלאה |
| כשל פתוח | fail open | כשהמגביל עצמו לא זמין, הבקשה עוברת; ההחלטה ההפוכה הייתה מפילה את הקליטה |
| קדנציה | term | מספר שעולה בכל בחירת מנהיג, ומשמש כ־fencing token |
| פיגור מסירה | delivery lag | הזמן מקליטת האירוע ועד שהמקבל אישר אותו; היעד של N2 |

הטבלה תושלם עם אבני הדרך הבאות.

## טעויות נפוצות

- **לסמן אבן דרך כגמורה כשרק חלק מהבדיקות שלה ירוקות.** הקריטריון הוא כל הפקודה של הסעיף, עם `-race`.
- **לבדוק את הצינור רק ברכיבים בודדים.** כל רכיב יכול לעבור לבד, והחיבור ביניהם עדיין שבור. הבדיקה מקצה לקצה של
  אבן הדרך החמישית היא שמוכיחה את ההבטחה.
- **לסמוך על Kafka לסינון כפילויות.** ה־offset של פרסום כפול שונה. הכפילות נעצרת באילוץ במסד, ובמקבל לפי מזהה האירוע.
- **לנסות שוב כל כישלון.** תשובת 4xx סופית; ניסיון חוזר שלה רק מאריך את התור של כולם.
- **הגנה שעולה יותר ממה שהיא חוסכת.** מחיצה שמחכה, או השלת עומס אחרי קריאה למסד, מחזירות את העומס מהדלת האחורית.
- הרשימה תושלם עם אבני הדרך הבאות.

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

## פתק Obsidian

הפתק המוכן נמצא גם כקובץ: `obsidian/38-hookline-milestones.md`. זה תוכנו:

- **‏M1:** ממשקי מאגר; מימוש בזיכרון; אימות לפי גיבוב; 404 לכל משאב של דיירת אחרת.
- **‏M2:** מיגרציות לכל המודל; ‏`ON CONFLICT DO NOTHING` ואז `SELECT`; אירוע ו־outbox בטרנזקציה אחת.
- **‏M3:** מטמון לכל דיירת; מוחקים אחרי השמירה; זמן החיים הוא הגבול לרשימה ישנה; ה־sink עם worker אחד.
- **‏M4:** פרסום, אישור, ואז סימון; ‏offset אחרי ההכנסה; ‏`ON CONFLICT DO NOTHING` על אירוע ו־endpoint.
- **‏M5:** ‏`SKIP LOCKED` וחכירה; 408, 429 ו־5xx מנוסים שוב, שאר 4xx לא; ‏`attempt_base` בהפעלה חוזרת.
- **‏M6:** מחיצה בלי המתנה ומפסק לכל endpoint; דחייה בלי ניסיון; מגבלה לכל דיירת, נכשלת פתוח; השלה לפני האימות.
- **שאלה פתוחה:** ___

<div dir="ltr" align="left">

```text
M1 tests  DATABASE_URL= go test -race -skip 'Invalidat|Deliver|Replay|ActiveSubscriptions|RateLimit|OverItsLimit|LimiterFailure|Shed' ./internal/api/ ./internal/store/... ./internal/admin/
M2 db     make db-up && export DATABASE_URL=postgres://hookline:change-me@localhost:5438/hookline
M2 tests  go test -race -count=1 -skip 'Invalidat|Deliver|Replay|ActiveSubscriptions|RateLimit|OverItsLimit|LimiterFailure|Shed' ./internal/api/ ./internal/store/... ./internal/admin/ ./internal/db/ ./migrations/
admin     go run ./cmd/hookline-admin migrate up | create-tenant -name <name> | create-topic -partitions 6
stack     make up   (postgres, redis, redpanda, console :8098, sink :8091)   make down
env       REDIS_ADDR=localhost:6438 KAFKA_BROKERS=localhost:19438 SINK_URL=http://localhost:8091
run       make run | make relay | make dispatcher | make worker
sink      curl -s localhost:8091/stats    curl -s -X POST localhost:8091/control -d '{"failure_rate":1}'
retry     408 429 5xx timeout refused -> retry      other 4xx 3xx -> dead      replay: dead -> pending
lab       make lab-n3      GUARDS=off make lab-n3      (record your p50/p99, date, machine)
guards    HOOKLINE_ENDPOINT_CONCURRENCY=2 HOOKLINE_BREAKER_FAILURES=5 HOOKLINE_BREAKER_OPEN=30s HOOKLINE_MAX_IN_FLIGHT=256
m5 tests  go test -race -count=1 -skip 'RateLimit|OverItsLimit|LimiterFailure|Shed|Breaker|Bulkhead' ./internal/backoff/ ./internal/delivery/ ./internal/api/ ./internal/store/... ./internal/e2e/
skip      -skip drops later milestones' tests: Invalidat ActiveSubscriptions (M3), StaleRelay RateLimit OverItsLimit LimiterFailure Shed Breaker Bulkhead (M6), Deliver Replay (M5)
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
- הרשימה המלאה תושלם עם אבני הדרך הבאות.

## העשרה אופציונלית

- המסמך של Standard Webhooks על ניסיונות חוזרים ועל קודי תשובה: השווי את המדיניות שלו ל־`Classify`.
- המאמר של Marc Brooker מ־2015 על המתנה גדלה ופיזור אקראי: המקור של „full jitter” שה־backoff מממש.
- יושלם עם אבני הדרך הבאות.

</div>
