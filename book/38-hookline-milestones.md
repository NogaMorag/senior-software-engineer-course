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
DATABASE_URL= go test -race ./internal/api/ ./internal/store/... ./internal/admin/
```

</div>

ההשמה הריקה של `DATABASE_URL` מבטיחה שבדיקות ה־API רצות על המאגר בזיכרון, גם אם הגדרת את המשתנה בחלון אחר.

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
go test -race -count=1 ./...
```

</div>

### 38.2.6 הפלט הצפוי

<div dir="ltr" align="left">

```text
ok  	example.com/hookline/internal/admin	1.352s
ok  	example.com/hookline/internal/api	2.177s
ok  	example.com/hookline/internal/config	2.713s
ok  	example.com/hookline/internal/db	3.995s
ok  	example.com/hookline/internal/domain	1.727s
ok  	example.com/hookline/internal/page	2.376s
ok  	example.com/hookline/internal/platform	3.285s
ok  	example.com/hookline/internal/problem	4.484s
ok  	example.com/hookline/internal/store	4.788s
ok  	example.com/hookline/internal/store/memory	2.058s
ok  	example.com/hookline/internal/store/postgres	3.889s
ok  	example.com/hookline/internal/store/storetest	4.147s
ok  	example.com/hookline/migrations	4.469s
```

</div>

בהרצה עם `-v` אף בדיקה לא מדלגת: כל הבדיקות שדילגו באבן הדרך הראשונה רצות עכשיו מול המסד.

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

יושלם עם אבן הדרך.

## 38.4 M4: outbox ו־Redpanda

יושלם עם אבן הדרך.

## 38.5 M5: מסירה, חתימה וניסיונות חוזרים

יושלם עם אבן הדרך.

## 38.6 M6: חוסן ועומס

יושלם עם אבן הדרך.

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
- הסיכום יושלם עם אבני הדרך הבאות.

## מונחים לזכור

| מונח בעברית | באנגלית | במשפט אחד |
| --- | --- | --- |
| אבן דרך | milestone | שלב בפרויקט שנגמר כשקבוצת בדיקות מוגדרת ירוקה |
| בדיקות התאמה | conformance tests | סוויטה אחת שכל מימוש של ממשק חייב לעבור |
| ‏keyset | keyset pagination | עימוד לפי ערכי המיון של הפריט האחרון במקום לפי היסט |

הטבלה תושלם עם אבני הדרך הבאות.

## טעויות נפוצות

- **לסמן אבן דרך כגמורה כשרק חלק מהבדיקות שלה ירוקות.** הקריטריון הוא כל הפקודה של הסעיף, עם `-race`.
- הרשימה תושלם עם אבני הדרך הבאות.

## שאלות ראיון

1. **למה המאגר יושב מאחורי ממשק כבר באבן הדרך הראשונה, כשהוא עוד בזיכרון?** יושלם עם אבני הדרך הבאות.

## פתק Obsidian

הפתק המוכן נמצא גם כקובץ: `obsidian/38-hookline-milestones.md`. זה תוכנו:

- **‏M1:** ממשקי מאגר; מימוש בזיכרון; אימות לפי גיבוב; 404 לכל משאב של דיירת אחרת.
- **‏M2:** מיגרציות לכל המודל; ‏`ON CONFLICT DO NOTHING` ואז `SELECT`; אירוע ו־outbox בטרנזקציה אחת.
- **שאלה פתוחה:** ___

<div dir="ltr" align="left">

```text
M1 tests  DATABASE_URL= go test -race ./internal/api/ ./internal/store/... ./internal/admin/
M2 db     make db-up && export DATABASE_URL=postgres://hookline:change-me@localhost:5438/hookline
M2 tests  go test -race -count=1 ./...
admin     go run ./cmd/hookline-admin migrate up | create-tenant -name <name>
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
- הרשימה המלאה תושלם עם אבני הדרך הבאות.

## העשרה אופציונלית

- יושלם עם אבני הדרך הבאות.

</div>
