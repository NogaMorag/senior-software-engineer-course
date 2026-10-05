---
direction: rtl
lang: he
---

<div dir="rtl" lang="he">

# פרויקט הגמר: Hookline, פלטפורמה למסירת webhooks

**מטרה:** לבנות, אבן דרך אחר אבן דרך, שירות שמקבל אירועים ממוצר SaaS ומוסר אותם, חתומים ו„לפחות פעם אחת”, לכל
endpoint מנוי, עם בידוד בין מקבלים, ניסיונות חוזרים, הרצה מחדש של כשלים, ותצפיתיות מלאה. בסוף יש לך פרויקט אחד,
שלך, שאת יודעת להסביר בראיון שורה אחר שורה, עם מספרים שמדדת בעצמך.

**זמן משוער:** עשר אבני דרך לאורך הספר, בימים 22 עד 99; כ־55 עד 90 דקות בנייה בכל יום של אבן דרך. הטבלה בצעד 2
מפרטת את הימים של כל אחת.

**ידע מקדים:** פרק 37 (המפרט), ופרק 38 לפי אבן הדרך. הפרויקט מעתיק ומתאים קוד מהתרגילים שלך: החבילה
`internal/platform` מפרק 6 (`ch06-shorty`), החבילות של פרטי הבעיה, העימוד והאידמפוטנטיות מפרק 9
(`ch09-catalog-api`), וחיבור המסד, המיגרציות והמסד הזמני לבדיקות מפרק 11 (`ch11-isolation-lab`). מאבן הדרך
השלישית: המקבל מפרק 8 (`ch08-receiver`), המטמון מפרק 12 (`ch12-cache-aside`), ה־outbox והלקוחות של Kafka מפרק 13
(`ch13-outbox`), וההמתנה, התור וה־workers מפרק 14 (`ch14-jobqueue`). מאבן הדרך השישית: המפסק והמחיצה מפרק 15
(`ch15-resilience-lab`), הגבלת הקצב וההשלה מפרק 16 (`ch16-ratelimit`), ובחירת המנהיג מפרק 18 (`ch18-leader`).
מאבן הדרך השביעית: ה־Dockerfiles מפרק 19, ‏kind מפרק 20, ה־chart מפרק 21, ‏Terraform מפרק 22 וה־CI מפרק 23.
מאבן הדרך השמינית: הטלמטריה מפרק 24 (`ch24-telemetry`, כולל ה־sink עם OpenTelemetry), והכללים, התבניות וסקריפט
התקרית מפרק 25 (`ch25-slo`). מאבן הדרך התשיעית: הסכמה, בדיקות התכונה ו־k6 מפרק 26 (`ch26-testing`), והגנת
SSRF, הסתרת הסודות ומודל האיומים מפרק 28 (`ch28-security`). באבן הדרך העשירית: ההערכה, הכאוס וההדגמה של פרק 39.
הגרסאות של הספריות ושל המסד מקובעות ב־`go.mod` ובקובץ ה־compose, נכון לאוקטובר 2026 (נספח ג).

## מה בתיקייה הזאת

| קובץ או תיקייה | תפקיד |
| --- | --- |
| `README.md` | הקובץ הזה: ההתקנה, הדרך, ורשימת אבני הדרך |
| `hookline/` | המימוש המלא, מעודכן עד אבן הדרך האחרונה שפורסמה. לקרוא כשנתקעת |
| `hookline-starter/` | העץ שאת מתחילה ממנו: אותו עץ, אותן בדיקות, עם `panic("TODO: see chapter 38")` בגופי הפונקציות שהן עבודה שלך |
| `hookline-starter/STARTER.md` | הסבר קצר באנגלית על איך ה־starter נוצר |

בתוך כל אחת משתי התיקיות, המבנה זהה (פרק 37, סעיף 37.3):

| נתיב | מה יש בו | אבן דרך |
| --- | --- | --- |
| `cmd/api/main.go` | חיווט השרת: הגדרות, מאגר, שני listeners, כיבוי מסודר | M1 |
| `cmd/hookline-admin/main.go` | כלי המפעילה: `migrate`, ‏`create-tenant`, ‏`create-topic` | M1, M2, M4 |
| `api/openapi.yaml` | החוזה: OpenAPI 3.1, כל נתיב וכל סוג בעיה | M1 |
| `internal/config` | ההגדרות מהסביבה (שלם) | M1 |
| `internal/platform` | השרת, ה־middleware, הלוגים וה־pprof מפרק 6 (שלם) | M1 |
| `internal/problem`, ‏`internal/page` | פרטי בעיה ועימוד מפרק 9 (שלמים) | M1 |
| `internal/domain` | טיפוסי התחום, מזהים, מפתחות, אימות קלט (שלם) | M1 |
| `internal/store` | ממשקי המאגר, ובדיקות ההתאמה ב־`storetest` (שלמים) | M1 |
| `internal/store/memory` | המאגר בזיכרון | M1 |
| `internal/api` | ה־handlers, האימות והקליטה; ‏`server.go` שלם | M1 |
| `internal/admin` | יצירת דיירת | M1 |
| `migrations/` | ארבע מיגרציות goose לכל מודל הנתונים, ועמודה אחת שנוספה באבן הדרך החמישית (שלמות) | M2, M5 |
| `internal/db`, ‏`internal/testdb` | ה־pool, המיגרציות מתוך Go, ומסד זמני לבדיקות, מפרק 11 (שלמים) | M2 |
| `internal/store/postgres` | המאגר ב־PostgreSQL | M2 |
| `deploy/compose.yaml` | הסביבה בפרויקט compose בשם `hookline`: PostgreSQL, ‏Redis, ‏Redpanda, הקונסולה וה־sink | M2 עד M4 |
| `internal/cache`, ‏`internal/subcache` | ‏cache-aside מפרק 12, ומטמון המנויים לכל דיירת | M3 |
| `internal/wire` | החיבורים המשותפים לבינאריים: מסד, מטמון, אותות (שלם) | M3 |
| `sink/` | המקבל המדומה מפרק 8, עם `Dockerfile` ווקטורי החתימה ב־`sink/data/` | M3 |
| `internal/bus` | לקוחות franz-go מפרק 13 (שלם) | M4 |
| `internal/outbox`, ‏`internal/dispatch` | ה־relay וה־dispatcher | M4 |
| `cmd/relay`, ‏`cmd/dispatcher`, ‏`cmd/worker` | שלושת הבינאריים החדשים (שלמים) | M4, M5 |
| `internal/backoff`, ‏`internal/delivery` | ההמתנה מפרק 14; החתימה, השליחה, התור וה־workers | M5 |
| `internal/testenv` | ‏topics פרטיים ונתוני בדיקה (שלם) | M4, M5 |
| `internal/e2e` | הבדיקה מקצה לקצה מול ה־compose | M5 |
| `internal/breaker`, ‏`internal/bulkhead` | מפסק ומחיצה מפרק 15 | M6 |
| `internal/ratelimit`, ‏`internal/shed` | הגבלת קצב והשלת עומס מפרק 16 | M6 |
| `internal/leader` | בחירת מנהיג מפרק 18, ל־relay | M6 |
| `scripts/lab-n3.sh` | ניסוי הבידוד של N3, עם sink שני ב־compose | M6 |
| `Dockerfile`, ‏`sink/Dockerfile` | שני ה־images, לפי פרק 19 | M7 |
| `deploy/kind/` | הגדרת האשכול, והתשתית לאשכול kind | M7 |
| `deploy/helm/hookline/` | ה־chart של Helm, עם סכמה לערכים ו־Job של מיגרציות | M7 |
| `deploy/terraform/` | ‏Terraform שמתקין את ה־chart, עם קובץ נעילה לארבע פלטפורמות | M7 |
| `.github/` | ה־workflow ו־action הבדיקות של פרק 23; מקומם בשורש המאגר שלך | M7 |
| `scripts/check-chart.sh`, ‏`scripts/kind-smoke.sh` | בדיקות ה־chart, ובדיקת העשן על kind | M7 |
| `internal/telemetry`, ‏`internal/metrics` | ‏OpenTelemetry מפרק 24, והמדדים של Hookline | M8 |
| `deploy/compose.observability.yaml`, ‏`deploy/observability/` | ה־collector, ‏Prometheus, ‏Loki, ‏Tempo, ‏Grafana, הכללים והבדיקות שלהם | M8 |
| `docs/runbooks/`, ‏`docs/postmortem-*.md` | ‏runbook לכל התראה, ותבנית ודוגמה של תחקיר | M8 |
| `scripts/promtool.sh`, ‏`check-telemetry.sh`, ‏`stage-incident.sh` | בדיקות הכללים, בדיקת הטלמטריה, והתקרית המבוימת | M8 |
| `internal/ssrf`, ‏`internal/redact` | הגנת SSRF והסתרת סודות מפרק 28, עם ההתאמות של Hookline (שלמות) | M9 |
| `contract/`, ‏`sink/tests/test_contract.py` | הסכמה של גוף ה־webhook, הדוגמאות, והבדיקות משני הצדדים | M9 |
| `loadtest/`, ‏`scripts/load.sh` | תרחישי k6 של N1 ו־N5, והסקריפט שמכין את הדיירות | M9 |
| `scripts/supply-chain.sh` | ‏govulncheck, ‏SBOM לכל image, וסריקה עם Grype | M9 |
| `docs/threat-model.md` | מודל האיומים של Hookline, לפי התבנית של פרק 28 | M9 |
| `scripts/chaos.sh`, ‏`scripts/verify-delivery.sh`, ‏`scripts/stage-lag.sh` | ניסוי הכאוס, הוכחת המסירה, ופירוק הפיגור לשלבים | M10 |
| `loadtest/overload.js` | התרחיש של N6, דרך `scripts/load.sh overload` | M10 |
| `docs/EVALUATION.md`, ‏`docs/DEMO.md` | התוצאות שלך ל־N1 עד N9, ותסריט ההדגמה | M10 |
| `Makefile` | כל הפקודות: `test-unit`, ‏`test`, ‏`lint`, ‏`up`, ‏`down`, ‏`migrate`, ‏`run`, ‏`relay`, ‏`dispatcher`, ‏`worker`, ‏`sink-test` | M1 |
| `docs/` | ‏runbooks, תחקיר, מודל איומים, תוצאות הערכה ותסריט הדגמה | M8 עד M10 |

## צעדים

**1. ההתקנה (יום 22).** העתיקי את ה־starter ובדקי שהוא מתקמפל:

<div dir="ltr" align="left">

```bash
mkdir -p ~/senior-course/hookline
cp -r $COURSE/exercises/project-hookline/hookline-starter/. ~/senior-course/hookline/
cd ~/senior-course/hookline
go build ./...
go vet ./...
make test-unit
```

</div>

הבנייה וה־vet עוברים. ‏`make test-unit` נכשל בחבילות של אבני הדרך, וזה המצב שממנו מתחילים (פרק 37, סעיף 37.8).

**2. אבן דרך אחת בכל פעם.** כל אבן דרך מתוארת בסעיף שלה בפרק 38: המטרה, הקבצים, הצעדים, והפקודה שצריכה להוריק.

| אבן דרך | ימים | סעיף | הפקודה שמסיימת אותה |
| --- | --- | --- | --- |
| ‏M1: שלד ה־API | 22 ו־23 | 38.1 | הפקודה בסעיף 38.1.5 |
| ‏M2: PostgreSQL וקליטה אידמפוטנטית | 29 ו־30 | 38.2 | הפקודה בסעיף 38.2.5, מול המסד |
| ‏M3: מטמון והמקבל המדומה | 33 ו־34 | 38.3 | הפקודה בסעיף 38.3.5, ו־`make sink-test` |
| ‏M4: ‏outbox ו־Redpanda | 39 ו־40 | 38.4 | הפקודה בסעיף 38.4.5 |
| ‏M5: מסירה, חתימה וניסיונות חוזרים | 43 ו־44 | 38.5 | הפקודה בסעיף 38.5.5, כולל הבדיקה מקצה לקצה |
| ‏M6: חוסן ועומס | 54 ו־55 | 38.6 | הפקודה בסעיף 38.6.5, ו־`make lab-n3` בשתי הגרסאות |
| ‏M7: קונטיינרים, Kubernetes ו־CI | 68 עד 70 | 38.7 | הפקודות בסעיף 38.7.5, כולל `make kind-smoke` |
| ‏M8: תצפיתיות ו־SLO | 77 עד 79 | 38.8 | הפקודות בסעיף 38.8.5, כולל התקרית המבוימת |
| ‏M9: בדיקות ואבטחה | 85 ו־86 | 38.9 | הפקודות בסעיף 38.9.5, כולל שני תרחישי k6 |
| ‏M10: הערכה והדגמה | 98 ו־99 | 38.10, ופרק 39 | הפקודות בסעיף 38.10.5, ו־„PASS: 0 missing” אחרי `make chaos` |

הפקודות של כל אבן דרך מונות את החבילות שלה בלבד, ובדגל `-skip` מדלגות על בדיקות של אבני דרך מאוחרות שכבר נמצאות
באותן חבילות. ה־starter מכיל את כל העץ, ולכן `go test ./...` ירוק רק בסוף.

**3. קומיט לכל אבן דרך.** כשהפקודה ירוקה:

<div dir="ltr" align="left">

```bash
cd ~/senior-course
git add hookline
git commit -m "M1: API skeleton"
```

</div>

**4. הסביבה, מאבן הדרך השנייה.** באבן הדרך השנייה מספיק המסד (`make db-up`); מהשלישית, כל ה־compose. כל יום עבודה
מתחיל ונגמר כך:

<div dir="ltr" align="left">

```bash
cd ~/senior-course/hookline
make up
export DATABASE_URL=postgres://hookline:change-me@localhost:5438/hookline
export REDIS_ADDR=localhost:6438 KAFKA_BROKERS=localhost:19438 SINK_URL=http://localhost:8091
go run ./cmd/hookline-admin migrate up
make down
```

</div>

בין שתי הפקודות האחרונות מריצים את הבדיקות של אבן הדרך. הסיסמה `change-me` והסוד `whsec_change-me` הם ערכי
תרגול מקומיים, ולא סודות. ‏`make down` מוחקת גם את ה־volumes, כך שכל יום מתחיל ממצב נקי.

## פלט צפוי

אחרי אבן הדרך הראשונה, בלי מסד:

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

אחרי אבן הדרך החמישית, כל המודול מול ה־compose (`go test -race -count=1 ./...`) מסתיים כך, והזמנים אצלך יהיו אחרים:

<div dir="ltr" align="left">

```text
ok  	example.com/hookline/internal/admin	1.534s
ok  	example.com/hookline/internal/api	2.040s
ok  	example.com/hookline/internal/backoff	1.817s
ok  	example.com/hookline/internal/bus	2.656s
ok  	example.com/hookline/internal/cache	3.612s
ok  	example.com/hookline/internal/config	2.497s
ok  	example.com/hookline/internal/db	2.815s
ok  	example.com/hookline/internal/delivery	4.267s
ok  	example.com/hookline/internal/dispatch	4.349s
ok  	example.com/hookline/internal/domain	3.762s
ok  	example.com/hookline/internal/e2e	5.467s
ok  	example.com/hookline/internal/outbox	4.776s
ok  	example.com/hookline/internal/page	4.339s
ok  	example.com/hookline/internal/platform	4.002s
ok  	example.com/hookline/internal/problem	3.882s
ok  	example.com/hookline/internal/store	3.395s
ok  	example.com/hookline/internal/store/memory	3.646s
ok  	example.com/hookline/internal/store/postgres	4.074s
ok  	example.com/hookline/internal/store/storetest	3.183s
ok  	example.com/hookline/internal/subcache	3.270s
ok  	example.com/hookline/internal/wire	3.076s
ok  	example.com/hookline/migrations	3.585s
```

</div>

ואחרי הרצה של ארבעת הבינאריים וחמישה אירועים (סעיף 38.5.4), ה־sink מדווח:

<div dir="ltr" align="left">

```text
{"received":5,"unique":5,"duplicates":0,"rejected":0,"failed":0,"hooks":{"smoke":{"received":5,"unique":5,"duplicates":0,"rejected":0,"failed":0}}}
```

</div>

## רשימת בדיקה

- [ ] ה־starter מועתק לתיקייה `~/senior-course/hookline`, והבנייה והבדיקה הסטטית עוברות.
- [ ] ‏M1: הפקודה של סעיף 38.1.5 ירוקה עם `-race`.
- [ ] ‏M1: ‏`make run-dev`, ואז קליטה של אותו אירוע פעמיים מחזירה את אותו מזהה, וגוף אחר מחזיר 422.
- [ ] ‏M1: החוזה עובר את redocly, ו־`TestSpecCoversEveryRoute` ירוקה.
- [ ] ‏M2: המיגרציות עולות, יורדות עד 0 ועולות שוב.
- [ ] ‏M2: הפקודה של סעיף 38.2.5 ירוקה מול המסד, ואף בדיקה שלה לא מדלגת.
- [ ] ‏M3: בדיקות המטמון וההתפרצות עוברות מול Redis, ו־`make sink-test` נקי.
- [ ] ‏M4: בדיקת נקודת הקריסה עוברת: כל אירוע מתפרסם פעמיים, ויש מסירה אחת לכל endpoint.
- [ ] ‏M5: ‏`TestPipelineEndToEnd` ירוקה, ו־`go test -race -count=1 ./...` ירוק מול ה־compose.
- [ ] ‏M6: המפסק, המחיצה, הגבלת הקצב וההשלה עוברים; שני relays, אחד מוביל.
- [ ] ‏M6: ‏`make lab-n3` ו־`GUARDS=off make lab-n3` רצו, והמספרים שלך רשומים ב־`docs/EVALUATION.md`.
- [ ] ‏M7: שלוש הבדיקות של סעיף 38.7.5 עוברות, ו־Terraform התקין והסיר את ה־release על kind.
- [ ] ‏M7: ה־workflow בשורש המאגר, ‏actionlint נקי, והוא ירוק ב־GitHub.
- [ ] ‏M8: ‏`make check-telemetry` ו־`make promtool` עוברים, והתחקיר שלך כתוב עם המספרים שמדדת.
- [ ] ‏M9: סריקות הלוגים, החוזה ובדיקות ה־SSRF ירוקות, ושני תרחישי k6 רשומים ב־`docs/EVALUATION.md`.
- [ ] ‏M9: ‏`make supply-chain` רץ, ומודל האיומים כולל שורה משלך.
- [ ] ‏M10: ‏`make chaos` מסתיים ב־„PASS: 0 missing”, ו־`docs/EVALUATION.md` מכסה את N1 עד N9 עם המספרים שלך.
- [ ] ‏M10: ההדגמה תורגלה עם שעון, והמאגר פומבי עם CI ירוק (פרק 39, סעיף 39.6).
- [ ] ‏M2: ‏`make lint` נקי, כולל staticcheck.
- [ ] קומיט לכל אבן דרך שהסתיימה, ו־`make down` בסוף כל יום.

## טעויות נפוצות ותיקונים

- **בדיקות ה־API עוברות בזיכרון ונכשלות מול המסד.** בדרך כלל מזהה שאינו UUID הגיע לשאילתה (שגיאה `22P02` בלוג),
  או זמן שלא הומר ל־UTC. בדקי ש־`domain.IsID` נקרא בכל נתיב, ושכל סריקה של זמן ממירה ל־UTC.
- **‏`go test` ירוק מיד אחרי שינוי במסד.** זו תוצאה שמורה מהרצה קודמת. הוסיפי `-count=1`, או השתמשי ב־`make test`.
- **‏`connection refused` על פורט 5438.** המסד לא למעלה, או שעלה בפרויקט compose אחר. הריצי `make db-up` מתוך
  `~/senior-course/hookline`, ובדקי ב־`docker compose -p hookline ps`.
- **‏`relation "events" does not exist` בהרצת השרת.** לא הרצת מיגרציות על המסד של השרת. הבדיקות יוצרות מסד זמני
  משלהן ומריצות עליו מיגרציות, אבל `make run` משתמש במסד `hookline` עצמו: `go run ./cmd/hookline-admin migrate up`.
- **‏panic עם `TODO: see chapter 38` בחבילה שחשבת שגמרת.** הבדיקות של חבילה אחת קוראות לקוד של אחרת: בדיקות ה־API
  קוראות למאגר בזיכרון. ה־panic אומר בדיוק איזו פונקציה עוד ריקה, בשורה הראשונה של ה־stack.

## רמזים

- התחילי כל אבן דרך מקריאת הבדיקות שלה. שם הבדיקה ושורת ה־`t.Fatalf` אומרים מה בדיוק מצופה.
- במאגר בזיכרון, מנעול אחד לכל המאגר מספיק. מהירות לא נמדדת כאן, נכונות כן.
- הפונקציות `pageParams` ו־`writePage` הן אותו עימוד מפרק 9, עם בדיקה אחת נוספת: מזהה בתוך ה־cursor חייב להיות UUID.
- ב־PostgreSQL, אל תבדקי אם מפתח אידמפוטנטיות קיים לפני ההכנסה. התחילי בהכנסה עם `ON CONFLICT DO NOTHING`,
  ורק כשלא חזרה שורה, קראי את השורה הקיימת.
- כשמנוי ל־endpoint של דיירת אחרת מחזיר 500 במקום 404, חפשי את הקוד `23503` בשגיאה. המפתח הזר המורכב עובד,
  ונשאר רק לתרגם אותו.

## הפתרון

המימוש המלא נמצא בתיקייה [`hookline/`](hookline/), עם הסבר קצר באנגלית בקובץ ה־README שלה. הוא עובד כי כל החלטה
על כפילות או בעלות יושבת במקום אחד שאי אפשר לעקוף. הדיירת נקבעת רק מהמפתח, וכל מתודת מאגר מסננת לפי מזהה
הדיירת שלה. הכפילות בקליטה מוכרעת באילוץ ייחודיות במסד, או במנעול אחד במאגר בזיכרון, ולא בבדיקה שקודמת להכנסה.
ובדיקות ההתאמה ב־`storetest` מבטיחות ששני המאגרים מתנהגים אותו דבר, ולכן ה־handlers לא צריכים לדעת איזה מהם
מאחוריהם.

## מה התרגיל מדגים

- איך מפרט עם דרישות ובדיקות הופך לעבודה של שבוע בכל פעם, שתמיד נגמרת בבדיקות ירוקות.
- ממשקים כגבול: אותן בדיקות API רצות על מאגר בזיכרון ועל PostgreSQL, בלי שינוי.
- אידמפוטנטיות אמיתית תחת מקביליות: אילוץ במסד, טרנזקציה אחת לאירוע ול־outbox, ובדיקה שסופרת שורות.
- הגנה מפני BOLA בשלוש שכבות: המפתח, המאגר והמסד.
- שימוש חוזר בקוד משלך מפרקים קודמים, במקום לכתוב הכול מחדש בכל פעם.

</div>
