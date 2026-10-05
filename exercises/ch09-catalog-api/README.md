---
direction: rtl
lang: he
---

<div dir="rtl" lang="he">

# תרגיל פרק 9: `catalog-api`, שירות קטלוג מתוך חוזה

**מטרה:** לבנות שירות קטלוג ב־Go שהחוזה שלו כתוב לפני הקוד: כל שגיאה כפרטי בעיה, עימוד עם cursor שעומד
בהוספות, מפתחות אידמפוטנטיות שעומדים בבקשות מקבילות, גבולות קלט, ושירות gRPC קטן שמחולל עם buf. שלוש
החבילות `internal/problem`, `internal/page` ו־`internal/idem` יחזרו ב־Hookline.

**זמן משוער:** שלושה ימי לימוד. יום 19: כ־35 דקות בנייה. יום 20: כ־65 דקות. יום 21: כ־60 דקות.

**ידע מקדים:** פרק 9; פרק 6 (נתב, שכבות ביניים, בדיקות עם `httptest`); פרק 5 (מנעולים והדגל `-race`).
התרגיל לא מייבא קוד מתרגילים אחרים. צריך Docker פועל, ו־buf עם שני התוספים שלו (נבדק ביום 21).

## מה בתיקייה הזאת

| קובץ | תפקיד |
| --- | --- |
| `starter/` | נקודת ההתחלה: מודול Go שמתקמפל, עם `TODO` לפי ימים |
| `starter/api/openapi.yaml` | החוזה, בלי הסכמה `Problem` (יום 19) |
| `starter/internal/problem/problem.go` | סוגי הבעיות; `Write` הוא `TODO` של יום 19 |
| `starter/internal/httpapi/server.go` | נתב ו־handlers; יצירה וקריאה ביום 19, רשימה ביום 20 |
| `starter/internal/page/page.go` | קידוד ופענוח של cursor (יום 20) |
| `starter/internal/catalog/memory.go` | החנות בזיכרון; `List` הוא `TODO` של יום 20 |
| `starter/internal/idem/idem.go` | מפתחות אידמפוטנטיות; `Begin` ו־`Middleware` של יום 20 |
| `starter/internal/httpapi/decode.go` | פענוח ובדיקת תקינות; הקשחה ביום 21 |
| `starter/internal/httpapi/middleware.go` | מזהה בקשה, לוג, recover, הוצאה משימוש; `limitBody` של יום 21 |
| `starter/internal/grpcapi/server.go` | שרת ה־gRPC; `GetItem` של יום 21 |
| `starter/proto/catalog/v1/catalog.proto` | הגדרת שירות ה־gRPC |
| `starter/buf.yaml`, `starter/buf.gen.yaml` | הגדרות buf: בדיקה, תאימות וחילול |
| `starter/gen/catalog/v1/` | הקוד המחולל, שמור במאגר |
| `starter/cmd/catalog/main.go` | הבינארי: HTTP על 8080, gRPC על 9090, כיבוי מסודר |
| `starter/api/redocly.yaml`, `starter/api/.redocly.lint-ignore.yaml` | הגדרות הבדיקה של החוזה |
| `tests/internal/...` | 31 הבדיקות הקנוניות; עותקים זהים יושבים ב־`starter/` וב־`solution/` |
| `solution/` | הפתרון המלא |

## יום 19: חוזה, פרטי בעיה, יצירה וקריאה

**1. העתקה ושמירה ראשונה.** התרגיל יושב בתיקייה `ch09` של מאגר העבודה שלך:

<div dir="ltr" align="left">

```bash
mkdir -p ~/senior-course/ch09
cp -r $COURSE/exercises/ch09-catalog-api/starter/. ~/senior-course/ch09/
cd ~/senior-course/ch09
go build ./... && go vet ./...
git add . && git commit -m "ch09: starter"
```

</div>

השמירה הראשונה חשובה: ביום 21 תשווי אליה את הקוד המחולל. הרצה של כל הבדיקות עכשיו מסתיימת בהרבה
`panic: TODO(day 19)`, וזה צפוי.

**2. בדיקת החוזה, לפני שמתקנים.** הפקודה מושכת את התמונה בפעם הראשונה (שניות ספורות):

<div dir="ltr" align="left">

```bash
docker run --rm -v "$PWD/api:/spec:ro" -w /spec redocly/cli:2.57.0 lint --format=stylish openapi.yaml
```

</div>

שבע שגיאות `no-unresolved-refs`: כל תשובות השגיאה מפנות לסכמה `Problem`, שעדיין לא קיימת.

**3. הוסיפי את הסכמה `Problem`** במקום ההערה `TODO(day 19)` בקובץ החוזה. הדרישות:

- אובייקט עם חמשת החברים שבטבלה, ותיאור קצר לכל אחד.
- שלושת הראשונים חובה, בשורה `required: [type, title, status]`.
- חבר הרחבה `errors`: מערך של אובייקטים עם `detail` ו־`pointer`, שניהם חובה.

| חבר | טיפוס |
| --- | --- |
| `type` | מחרוזת בפורמט `uri` |
| `title` | מחרוזת |
| `status` | מספר שלם מ־400 עד 599 |
| `detail` | מחרוזת |
| `instance` | מחרוזת |

סעיף 9.3 מראה את הסכמה בטבלה דומה.

הריצי שוב את הבדיקה מצעד 2, בלי `--format=stylish`, עד שהיא מסתיימת ב־`Your API description is valid`.

**4. מלאי את `problem.Write`.** ברירת מחדל ל־`instance`, הכותרות, הסטטוס, ואז הקידוד. בקובץ ההתחלה
חסר הייבוא `encoding/json`; הוסיפי אותו (העורך עושה את זה לבד אם `gopls` פעיל).

<div dir="ltr" align="left">

```bash
go test ./internal/problem/
```

</div>

**5. מלאי את `createItem` ואת `getItem`** בקובץ `internal/httpapi/server.go`, לפי ההערות שבהם. שימי לב למיפוי:

| שגיאה מהחנות | מה מחזירים |
| --- | --- |
| מק"ט כפול | `problem.Conflict` |
| פריט חסר | `problem.NotFound` |
| כל שגיאה אחרת | `s.internal` |

ואז הבדיקות של היום:

<div dir="ltr" align="left">

```bash
go test ./internal/httpapi/ -run 'TestCreateAndGet|TestSpec|TestMethodNotAllowed|TestErrorsAreProblems/(unknown_item|unknown_route|duplicate_sku|all_fields_missing|malformed_json|method_not_allowed)$' -v
```

</div>

**6. נסי ביד.** בחלון אחד הפעילי את השירות, ובחלון שני שלחי בקשות:

<div dir="ltr" align="left">

```bash
go run ./cmd/catalog
```

</div>

<div dir="ltr" align="left">

```bash
curl -s -i -X POST localhost:8080/v1/items -H 'Content-Type: application/json' -H 'Idempotency-Key: k1' -d '{"sku":"MUG-001","name":"Blue mug","price_cents":4500}'
curl -s -i localhost:8080/v1/items/itm_99999999
```

</div>

עצרי את השירות ב־Ctrl+C ושמרי (`git commit`).

## יום 20: cursor, רשימה ומפתחות אידמפוטנטיות

**7. מלאי את `page.Encode` ואת `page.Decode`.** החזירי את הייבוא של `encoding/base64` ושל `encoding/json`.

<div dir="ltr" align="left">

```bash
go test ./internal/page/
```

</div>

**8. מלאי את `Memory.List`.** חפשי בחיפוש בינארי את הפריט הראשון שממוין **אחרי** הסמן, בעזרת `newer`.
התנאי בפונקציית ההשוואה הוא העיקר: הפריט של הסמן עצמו חייב להיחשב „לפני”, אחרת הוא יחזור בעמוד הבא.

<div dir="ltr" align="left">

```bash
go test ./internal/catalog/ -v -run 'TestList|TestCursor|TestOffset'
```

</div>

**9. מלאי את `listItems`**, לפי ההערה בקובץ. רשימה ריקה היא מערך ריק, לא `null`.

**10. מלאי את `Store.Begin` ואת `Store.Middleware`** בקובץ `internal/idem/idem.go`. ב־`Begin` ההחלטה
כולה, כולל השמירה של מפתח חדש, קורית תחת נעילה אחת. ב־`Middleware` חמשת השלבים מתוארים בהערה, והעזרים
שלהם כבר כתובים בקובץ:

- הפונקציה `parseKey`, שבודקת את הכותרת.
- הפונקציה `fingerprint`, שמחשבת את טביעת האצבע.
- הפונקציות `writeStored` ו־`copyHeader`, שמשחזרות תשובה שמורה.
- הטיפוס `recorder`, שמקליט את התשובה לפני שהיא נשלחת.

הריצי עם גלאי המרוצים:

<div dir="ltr" align="left">

```bash
go test -race ./internal/idem/ -v
```

</div>

**11. הבדיקות של היום דרך כל השרת:**

<div dir="ltr" align="left">

```bash
go test ./internal/httpapi/ -run 'TestCreateAndGet|TestSpec|TestMethodNotAllowed|TestListPagesThroughEverything|TestEmptyListIsArray|TestIdempotentCreateThroughTheStack|TestDeprecatedAlias|TestErrorsAreProblems/(unknown_item|unknown_route|duplicate_sku|all_fields_missing|malformed_json|method_not_allowed|no_idempotency_key|limit_too_large|forged_cursor)$'
```

</div>

**12. שחזור ביד.** הפעילי שוב את השירות, ושלחי את אותה בקשה פעמיים עם אותו מפתח, ואז פעם שלישית עם גוף
אחר. הראשונה 201, השנייה 201 עם `Idempotent-Replayed: true`, והשלישית 422:

<div dir="ltr" align="left">

```bash
curl -s -i -X POST localhost:8080/v1/items -H 'Content-Type: application/json' -H 'Idempotency-Key: demo-1' -d '{"sku":"MUG-001","name":"Blue mug","price_cents":4500}'
curl -s -i -X POST localhost:8080/v1/items -H 'Content-Type: application/json' -H 'Idempotency-Key: demo-1' -d '{"sku":"MUG-001","name":"Blue mug","price_cents":4500}'
curl -s -i -X POST localhost:8080/v1/items -H 'Content-Type: application/json' -H 'Idempotency-Key: demo-1' -d '{"sku":"MUG-002","name":"Red mug","price_cents":4500}'
curl -s -i localhost:8080/v1/products/itm_00000001
```

</div>

הבקשה האחרונה הולכת לנתיב שהוצא משימוש: שימי לב לשלוש כותרות ההוצאה משימוש מסעיף 9.6.
שמרי (`git commit`).

## יום 21: gRPC, גבולות ופענוח קפדני

**13. בדיקת הכלים.** בודקים שהשלושה מותקנים ובגרסאות הנכונות:

<div dir="ltr" align="left">

```bash
buf --version
protoc-gen-go --version
protoc-gen-go-grpc --version
```

</div>

<div dir="ltr" align="left">

```text
1.73.0
protoc-gen-go v1.36.12
protoc-gen-go-grpc 1.6.2
```

</div>

אם אחד חסר, הפקודות האלה מתקינות את הגרסאות הנעולות (נכון לאוקטובר 2026; ראי נספח ג):

<div dir="ltr" align="left">

```bash
go install github.com/bufbuild/buf/cmd/buf@v1.73.0
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.12
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.6.2
```

</div>

**14. חילול מחדש, והוכחה שהקוד השמור עדכני:**

<div dir="ltr" align="left">

```bash
buf lint
rm -rf gen
buf generate
git status --short gen
```

</div>

פלט ריק בפקודה האחרונה פירושו שהקבצים שחוללו עכשיו זהים בדיוק לאלה שנשמרו ביום 19.

**15. מלאי את `GetItem`** בקובץ `internal/grpcapi/server.go`. החזירי את הייבוא של `errors` ושל `strings`,
ואז הריצי:

<div dir="ltr" align="left">

```bash
go test ./internal/grpcapi/ -v
```

</div>

**16. מלאי את שלוש הפונקציות של הגבולות**, לפי ההערות שבהן: `limitBody` בקובץ
`middleware.go`, ואת `decodeJSON` ו־`validate` בקובץ `decode.go`. ב־`decodeJSON` הסדר הוא: סוג
תוכן, פענוח עם `DisallowUnknownFields`, בדיקה שאין נתונים אחרי האובייקט, ואז מיפוי השגיאות. כשכל
הבדיקות ירוקות, הריצי את הסוללה המלאה:

<div dir="ltr" align="left">

```bash
gofmt -l .
go vet ./...
staticcheck ./...
go test -race -count=1 ./...
```

</div>

**17. ניסוי בשינוי שובר.** שני שם של שדה בקובץ ה־proto, הריצי את בדיקת התאימות מול השמירה בענף הראשי,
והחזירי את הקובץ:

<div dir="ltr" align="left">

```bash
sed -i 's/  string name = 3;/  string title = 3;/' proto/catalog/v1/catalog.proto
buf breaking --against '../.git#branch=main,subdir=ch09'
git checkout proto/catalog/v1/catalog.proto
```

</div>

> **ב־Mac:** הדגל של `sed` לעריכה במקום דורש ארגומנט ריק, ולכן השורה הראשונה היא `sed -i ''` ואחריה
> אותו ביטוי.

הביטוי `branch=main` מניח שהענף הראשי של מאגר העבודה נקרא `main`; אם לא, החליפי את השם.

**18. בדיקת החוזה ושמירה אחרונה.** הריצי שוב את בדיקת החוזה מצעד 2 ושמרי (`git commit`).

## פלט צפוי

סוף הסוללה המלאה (הזמנים אצלך יהיו אחרים):

<div dir="ltr" align="left">

```text
ok  	example.com/ch09-catalog-api/internal/catalog	1.325s
ok  	example.com/ch09-catalog-api/internal/config	1.547s
ok  	example.com/ch09-catalog-api/internal/grpcapi	1.633s
ok  	example.com/ch09-catalog-api/internal/httpapi	2.237s
ok  	example.com/ch09-catalog-api/internal/idem	2.592s
ok  	example.com/ch09-catalog-api/internal/page	2.828s
ok  	example.com/ch09-catalog-api/internal/problem	1.888s
```

</div>

הבקשה השנייה בצעד 12, מקוצרת (המזהה והזמן אצלך יהיו אחרים):

<div dir="ltr" align="left">

```text
HTTP/1.1 201 Created
Content-Type: application/json
Idempotent-Replayed: true
Location: /v1/items/itm_00000001

{"id":"itm_00000001","sku":"MUG-001","name":"Blue mug","price_cents":4500,"created_at":"2026-10-05T08:30:39.9905Z"}
```

</div>

ובצעד 17, בדיקת התאימות נכשלת בקוד יציאה 100:

<div dir="ltr" align="left">

```text
proto/catalog/v1/catalog.proto:33:3:Field "3" with name "title" on message "Item" changed option "json_name" from "name" to "title".
proto/catalog/v1/catalog.proto:33:10:Field "3" on message "Item" changed name from "name" to "title".
```

</div>

## רשימת בדיקה

- [ ] החוזה עובר את Redocly, וכולל את הסכמה `Problem`.
- [ ] כל תשובת שגיאה היא `application/problem+json`, גם 404 של נתיב לא מוכר ו־405 עם `Allow`.
- [ ] בדיקת ה־cursor תחת הוספות עוברת, גם בריצה עם שעון קפוא.
- [ ] מפתח חוזר משחזר את התשובה, גוף אחר מקבל 422, בקשה מקבילה מקבלת 409, ו־5xx משחרר את המפתח.
- [ ] גוף מעל 16 KiB מקבל 413, שדה לא מוכר מקבל 400 עם מצביע, ושני אובייקטים רצופים נדחים.
- [ ] חילול מחדש עם buf לא משנה את הקוד השמור, ו־`buf breaking` עוצר שינוי שם של שדה.
- [ ] כל 31 הבדיקות עוברות עם `-race`.
- [ ] שלושת כלי הבדיקה הסטטית של צעד 16 שקטים.

## טעויות נפוצות ותיקונים

- **להריץ את כל הבדיקות ביום 19 ולהיבהל.** רוב החבילות נופלות על `panic: TODO(day 20)` או `TODO(day 21)`.
  מריצים את המסנן של היום מהצעדים, ואת השאר כשמגיע תורן.
- **לכתוב כותרת אחרי `WriteHeader` בתוך `problem.Write`.** הבדיקה מדווחת על `Content-Type` שגוי, כי
  כותרות שנקבעות אחרי שורת הסטטוס נעלמות. קובעים את כל הכותרות קודם.
- **הפונקציה `List` מחזירה שוב את הפריט של הסמן.** הבדיקה כותבת `listed twice`. בפונקציית ההשוואה, פריט ששווה
  לסמן צריך להיחשב „לפני”, כך שהחיפוש הבינארי ייעצר אחריו.
- **לפצל את `Begin` לבדיקה ולשמירה.** בדיקת הבקשות המקבילות עוברת לפעמים ונכשלת לפעמים, ו־`-race` עשוי
  לצעוק. ההחלטה כולה, כולל `s.entries[key] = ...`, יושבת בתוך אותה נעילה.
- **שגיאות `undefined: errors` או `undefined: strings` אחרי מילוי `TODO`.** קבצי ההתחלה בלי הייבוא
  שהפתרון צריך. מוסיפים אותו, או נותנים ל־`gopls` להוסיף.
- **להריץ את Redocly מתיקייה אחרת.** הכלי מדווח שהקובץ לא נמצא, כי `$PWD/api` לא קיים שם. מריצים מתוך
  התיקייה `~/senior-course/ch09`.

## רמזים

- ב־`Encode`, הקידוד הוא `base64.RawURLEncoding`: בלי ריפוד ובלי `+` או `/`, ולכן בטוח בכתובת.
- ב־`List`, הפונקציה `slices.BinarySearchFunc` מחזירה את המקום הראשון שבו ההשוואה לא מחזירה ערך שלילי.
  אחרי שמצאת אותו, `min(start+limit, len(m.items))` נותן את הסוף, ו־`slices.Clone` את ההעתק.
- ב־`Middleware`, הקריאה `errors.AsType[*http.MaxBytesError](err)` מזהה גוף גדול מדי.
- ב־`decodeJSON`, השגיאה על שדה לא מוכר מתחילה ב־`json: unknown field`, ואחריה שם השדה במירכאות.
- ב־`GetItem`, מספיק לבדוק קידומת ואורך; החנות כבר מחזירה `catalog.ErrNotFound`.

## הפתרון

הפתרון נמצא ב־`solution/`. **למה הוא עובד כך:** כל שגיאה, גם של הנתב וגם מ־panic, עוברת דרך פונקציה
אחת, ולכן אין נתיב שעונה בצורה אחרת. העימוד יציב כי מפתח המיון שלם, זמן ואחריו מזהה, והסמן מכיל את
המפתח ולא מקום. החלטת האידמפוטנטיות אטומית כי בדיקת המפתח ושמירתו הן פעולה אחת תחת נעילה אחת; במסד
נתונים אילוץ ייחודיות ימלא את התפקיד. וגבול הגודל עוטף את הגוף לפני שמישהו קורא ממנו, כך שגם שכבת
האידמפוטנטיות, שקוראת ראשונה, נעצרת בזמן.

## מה התרגיל מדגים

- חוזה כתוב שנבדק בכלי, ובדיקות שמוודאות שהשרת מקיים אותו.
- חוזה שגיאות אחד לפי RFC 9457, כולל השגיאות שהנתב עצמו מייצר.
- עימוד עם cursor שמוכח בבדיקה תחת הוספות, ובדיקה שמתעדת למה היסט נכשל.
- מפתחות אידמפוטנטיות עם כל ארבעת המצבים, כולל בקשות מקבילות, ובדיקה עם גלאי המרוצים.
- הוצאה משימוש בתוך הפרוטוקול, ו־gRPC עם קוד מחולל שמור ובדיקה בלי רשת.

</div>
