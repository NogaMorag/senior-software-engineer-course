---
direction: rtl
lang: he
---

<div dir="rtl" lang="he">

# תרגיל פרק 24: `telemetry`, שלושה אותות ומעקב אחד דרך Kafka

**מטרה:** לכשר את `shorty`, את העובד `linkworker` ואת המקבל ב־OpenTelemetry; להריץ מחסנית של collector,
Prometheus, Loki, Tempo ו־Grafana עם מקורות נתונים ודשבורד RED שנטענים מקבצים; ולראות מעקב אחד שעובר
מ־`shorty` דרך Redpanda אל `linkworker` ומשם למקבל ב־Python.

**זמן משוער:** שלושה ימי לימוד. יום 72: כ־60 דקות בנייה (הגדרת OpenTelemetry בשני השירותים והמחסנית). יום
73: כ־60 דקות (מדדים, PromQL, הדשבורד). יום 74: כ־65 דקות (לוגים, מעקבים והקפיצה דרך Kafka).

**ידע מקדים:** פרק 24; פרק 6 (`shorty`) ופרק 8 (המקבל), שמהם התרגיל מביא עותקים; פרק 13 (Redpanda
ו־franz-go); פרק 19 (ה־Dockerfiles, שהועתקו משם). Docker Desktop עם כ־4 GB פנויים, ו־`python3` ב־WSL2.

## מה בתיקייה הזאת

| קובץ | תפקיד |
| --- | --- |
| `starter/shorty/` | מודול Go (`example.com/ch24-telemetry`): `shorty`, `linkworker` ו־`loadgen`, עם `TODO` |
| `starter/shorty/internal/telemetry/telemetry.go` | הגדרת שלושת הספקים, `TODO` של יום 72 |
| `starter/shorty/internal/platform/logging.go`, `route.go` | מזהה המעקב בלוג, והתבנית לכל השכבות: `TODO` של יום 72 |
| `starter/shorty/internal/app/app.go` | שרשרת ה־middleware, בלי `otelhttp`: `TODO` של יום 72 |
| `starter/shorty/internal/events/events.go` | הכתיבה ל־Kafka; ה־hooks של `kotel` הם `TODO` של יום 74 |
| `starter/shorty/internal/linkworker/worker.go` | העובד; ה־span של הטיפול ומדדי USE הם `TODO` של ימים 73 ו־74 |
| `starter/shorty/cmd/loadgen/` | מחולל תנועה: יצירות, פענוחים וקודים חסרים |
| `starter/receiver/` | המקבל, פרויקט uv; `src/receiver/telemetry.py` עם `TODO` של יום 72 |
| `data/compose.yaml` | המחסנית כולה, הפרויקט `ch24`, עם מגבלות זיכרון ובדיקות בריאות |
| `data/otel-collector.yaml`, `prometheus.yml`, `loki.yaml`, `tempo.yaml` | התצורות של ה־collector ושל שלושת האחסונים |
| `data/grafana/provisioning/`, `data/grafana/dashboards/red.json` | מקורות הנתונים והדשבורד „RED overview” |
| `data/rules/` | תיקייה ריקה לחוקים; פרק 25 מחליף אותה בשלו |
| `data/signature-vectors.json` | וקטורי החתימה של פרק 8, שהבדיקות קוראות |
| `tests/` | הבדיקות הקנוניות: של Go תחת `tests/shorty/`, של Python תחת `tests/receiver/tests/`, ו־`tests/check_stack.sh` |
| `solution/` | הפתרון המלא; עותקים זהים של הבדיקות יושבים במקומות המקבילים ב־`starter/` וב־`solution/` |

**מאיפה הקוד.** התיקייה `shorty/` היא העתק של `exercises/ch06-shorty/solution/`, עם שם מודול חדש, עם פקודת
`healthcheck` מפרק 19 (שעברה לחבילה `platform` כדי ששתי הפקודות ישתמשו בה), ועם מה שהפרק מוסיף: החבילות
`telemetry`, `events` ו־`linkworker`, הפקודה `linkworker`, ו־`loadgen` חדש. ל־`shorty` נוספה כתיבה
ל־topic בשם `links.created` לכל קישור שנוצר. התיקייה `receiver/` היא העתק של
`exercises/ch08-receiver/solution/`, עם המודול `telemetry.py`, שורת לוג לכל משלוח שהתקבל, ו־`main.py` חדש.
ה־Dockerfiles הם אלה של פרק 19; של Go בונה את שלוש התוכנות לאותו image.

## יום 72: OpenTelemetry ב־Go וב־Python, והמחסנית

**1. העתקה.** התרגיל יושב בתיקייה `ch24`, עם התצורות והבדיקות:

<div dir="ltr" align="left">

```bash
mkdir -p ~/senior-course/ch24
cp -r "$COURSE/exercises/ch24-telemetry/starter/." ~/senior-course/ch24/
cp -r "$COURSE/exercises/ch24-telemetry/data" ~/senior-course/ch24/data
cd ~/senior-course/ch24
(cd shorty && go build ./... && go test ./... 2>&1 | tail -3)
(cd receiver && uv sync --frozen && uv run pytest -q 2>&1 | tail -2)
git add . && git commit -m "ch24: starter"
```

</div>

הבנייה עוברת, ובדיקות הטלמטריה נכשלות עם `TODO`. הבדיקות הישנות של פרקים 6 ו־8 ירוקות.

**2. `Setup` ב־Go.** מלאי את `internal/telemetry/telemetry.go` לפי השלבים שבהערה: שלושה מייצאים ב־HTTP,
שלושה ספקים עם ה־resource, התקנה גלובלית עם ה־propagator, `runtime.Start`, ופונקציית כיבוי:

<div dir="ltr" align="left">

```bash
cd ~/senior-course/ch24/shorty
go test -race -run TestSetupExportsAllThreeSignals ./internal/telemetry/
```

</div>

**3. לוגים ותבנית.** ב־`internal/platform/logging.go` הוסיפי `trace_id` ו־`span_id` ופיצול
ל־`also`; ב־`route.go` את `RoutePattern`; ב־`internal/app/app.go` את `otelhttp.NewMiddleware` עם המסנן
של `/healthz`:

<div dir="ltr" align="left">

```bash
go test -race ./internal/platform/ ./internal/app/
```

</div>

**4. Python.** מלאי את `setup_telemetry`, את `TraceContextFilter` ואת `instrument`
ב־`receiver/src/receiver/telemetry.py`:

<div dir="ltr" align="left">

```bash
cd ~/senior-course/ch24/receiver
uv run pytest -q tests/test_telemetry.py
```

</div>

**5. המחסנית.** הבנייה הראשונה מורידה כמה images ולוקחת כמה דקות:

<div dir="ltr" align="left">

```bash
cd ~/senior-course/ch24
docker compose -p ch24 -f data/compose.yaml up -d --build --wait
docker compose -p ch24 -f data/compose.yaml ps --format 'table {{.Service}}\t{{.Status}}'
```

</div>

| פקודה | מה היא עשתה |
| --- | --- |
| `up -d --build --wait` | בנתה את שני ה־images, הרימה את כל השירותים, וחיכתה לבדיקות הבריאות |
| `ps` | מראה כל שירות ואת המצב שלו; `topics` יצא אחרי שיצר את ה־topic |

**6. מעקב ראשון.** צרי קישור, ופתחי את Grafana בכתובת `http://localhost:3000`. ב־Explore בחרי Tempo,
Search, שירות `shorty`, ופתחי את המעקב האחרון:

<div dir="ltr" align="left">

```bash
curl -s -X POST localhost:8080/v1/links -H 'Content-Type: application/json' \
  -d '{"url": "https://go.dev/", "code": "gohome"}'
```

</div>

## יום 73: מדדים, PromQL ודשבורד

**7. תנועה.** אם הורדת את המחסנית בסוף יום 72, הרימי אותה שוב (הפקודה הראשונה). אחר כך הריצי את מחולל
התנועה לחמש דקות ברקע; הוא בא באותו image:

<div dir="ltr" align="left">

```bash
cd ~/senior-course/ch24
docker compose -p ch24 -f data/compose.yaml up -d --build --wait
docker compose -p ch24 -f data/compose.yaml run -d --rm loadgen -rps 20 -d 5m
```

</div>

**8. PromQL.** הריצי את ארבע השאילתות של סעיף 24.4.3 דרך `curl`, וגם ב־Grafana, Explore, Prometheus.
רשמי בפתק את קצב הבקשות ואת ה־p99 שקיבלת. נסי גם את השגיאה המכוונת: `histogram_quantile` בלי `le`
ב־`by`, וראי שהתוצאה ריקה.

**9. USE.** מלאי ב־`internal/linkworker/worker.go` את חלק המדדים של `Handle` (צעד 2 בהערה), והריצי
את הבדיקה. אחר כך בני מחדש:

<div dir="ltr" align="left">

```bash
(cd shorty && go test -race -run TestUSEMetrics ./internal/linkworker/)
docker compose -p ch24 -f data/compose.yaml up -d --build --wait shorty linkworker
```

</div>

**10. הדשבורד.** בתיקייה Course ב־Grafana פתחי את „RED overview”. החליפי את השירות ל־`receiver`. הוסיפי
פאנל אחד משלך (למשל קצב לפי `http_request_method`), ייצאי את ה־JSON, החליפי את `data/grafana/dashboards/red.json`,
והפעילי מחדש את Grafana כדי לראות שהוא נטען מהקובץ.

## יום 74: לוגים, מעקבים ו־Kafka

**11. Kafka.** מלאי את `Hooks` ב־`internal/events/events.go` ואת שאר `Handle` (ה־span, השגיאה, התוצאה).
הבדיקה השנייה צריכה את Redpanda של המחסנית, ולכן הפקודה הראשונה מרימה אותו אם הורדת את המחסנית אתמול:

<div dir="ltr" align="left">

```bash
cd ~/senior-course/ch24
docker compose -p ch24 -f data/compose.yaml up -d --wait redpanda
(cd shorty && go test -race ./...)
(cd shorty && KAFKA_BROKERS=localhost:19092 go test -race -run TestTraceContextThroughRedpanda -v ./internal/events/)
docker compose -p ch24 -f data/compose.yaml up -d --build --wait
```

</div>

**12. מלוג למעקב ובחזרה.** ב־Explore, Loki, הריצי `{service_name="linkworker"}`, פתחי שורה, ולחצי על
„Open trace”. במעקב, מה־span של המקבל, לחצי על כפתור הלוגים.

**13. הבדיקה המלאה.** הסקריפט שולח בקשה עם מזהה מעקב שהוא בחר, ושואל כל אחסון דרך ה־API שלו:

<div dir="ltr" align="left">

```bash
bash tests/check_stack.sh
```

</div>

**14. ניקוי.** בסוף כל יום, כשסיימת:

<div dir="ltr" align="left">

```bash
docker compose -p ch24 -f data/compose.yaml down -v
```

</div>

## פלט צפוי

הסוף של `tests/check_stack.sh`, מההרצה שבה נכתב הספר (המספרים והמזהים שלך יהיו אחרים):

<div dir="ltr" align="left">

```text
== 3. Prometheus has the series
PASS  sum(http_server_request_duration_seconds_count{job="shorty"}) = 41
PASS  sum(http_server_request_duration_seconds_count{job="receiver"}) = 1
PASS  sum(linkworker_deliveries_total{outcome="ok"}) = 1
PASS  sum(go_goroutine_count{job="shorty"}) = 61
PASS  exemplars (trace IDs on histogram samples) are stored
== 4. Loki has the log lines of that trace
PASS  Loki: 5 lines carry trace_id=4ab7ac96aa315eff58409ade877a62ee
== 5. Tempo has the trace, across the Kafka hop
PASS  Tempo: trace 4ab7ac96aa315eff58409ade877a62ee has 6 spans:
           1 linkworker:HTTP POST
           1 linkworker:links.created process
           1 linkworker:links.created receive
           1 receiver:POST /hooks/{name}
           1 shorty:POST /v1/links
           1 shorty:links.created publish
== 6. Grafana loaded the provisioned files
PASS  datasource prometheus
PASS  datasource loki
PASS  datasource tempo
PASS  dashboard red-overview

ALL CHECKS PASSED
```

</div>

והבדיקות של Go ושל Python, כשהכול מלא:

<div dir="ltr" align="left">

```text
ok      example.com/ch24-telemetry/internal/app
ok      example.com/ch24-telemetry/internal/events
ok      example.com/ch24-telemetry/internal/links
ok      example.com/ch24-telemetry/internal/linkworker
ok      example.com/ch24-telemetry/internal/platform
ok      example.com/ch24-telemetry/internal/telemetry
66 passed
```

</div>

## רשימת בדיקה

- [ ] `go test -race ./...` ירוק ב־`shorty`, ו־`uv run pytest -q` ירוק ב־`receiver`.
- [ ] הבדיקה מול Redpanda רצה עם `KAFKA_BROKERS` ועברה, ולא דולגה.
- [ ] בקשה ל־`shorty` נותנת מעקב ב־Tempo, עם ה־span בשם `POST /v1/links`.
- [ ] הדשבורד „RED overview” נטען מהקובץ והפאנלים זזים עם התנועה.
- [ ] משורת לוג ב־Loki הגעת למעקב, ומה־span חזרת ללוגים.
- [ ] מעקב אחד מכיל `links.created publish`, `links.created process` ו־`POST /hooks/{name}`.
- [ ] `tests/check_stack.sh` הסתיים ב־`ALL CHECKS PASSED`, והמחסנית הורדה עם `down -v`.

## טעויות נפוצות ותיקונים

- **ה־span נקרא `GET` ולמדד אין `http_route`.** `otelhttp` לא רואה את התבנית, כי `RequestID` מעביר עותק של
  הבקשה. בדקי ש־`RoutePattern` היא השכבה החיצונית ושהיא מעתיקה לפני שהיא כותבת.
- **אין מדדים של המקבל ב־Prometheus, אבל יש של `shorty`.** כנראה `http_server_duration_milliseconds` במקום
  השם היציב: `OTEL_SEMCONV_STABILITY_OPT_IN` לא הוגדר לפני `instrument_app`.
- **ארבעה מעקבים קצרים במקום אחד ארוך.** ה־propagator לא הותקן, או ש־`events.Hooks` נקראה לפני `Setup`
  ולקחה את ברירות המחדל. בעלייה, `Setup` ראשונה.
- **`docker compose up` נכשל ב־collector עם `unknown type: "otlphttp"`.** תצורה מגרסה ישנה: בגרסה המוצמדת
  המייצאים הם `otlp_http` ו־`otlp_grpc`.
- **כל `POST /v1/links` מחזיר 503.** ה־topic לא קיים: השירות `topics` לא רץ, או ש־`SHORTY_KAFKA_TOPIC` מוגדר
  בסביבה של המעטפת שלך. הריצי `unset SHORTY_KAFKA_TOPIC` והרימי שוב (פרק 25 עושה את זה בכוונה).
- **בדיקה נכשלת רק כשהמחסנית רצה.** המחשב עמוס; הריצי שוב את הבדיקה לבד לפני שאת מחפשת באג.

## רמזים

- ב־`Setup`, כל שגיאה אחרי שספק כבר נבנה צריכה לכבות אותו: `errors.Join(err, tp.Shutdown(ctx))`.
- הבדיקה של `Setup` מפנה את המייצאים לשרת `httptest` עם `t.Setenv`; אם היא תקועה, כנראה שהכיבוי לא נקרא.
- ב־Python, `format(context.trace_id, "032x")`; בלי ה־`032` מזהים שמתחילים באפס לא נמצאים.
- ב־`Handle`, `defer span.End()` מיד אחרי `WithProcessSpan`, והקריאה ל־`deliver` מקבלת את ה־`ctx` שחזר ממנה.
- שאילתה שמחזירה ריק ב־Prometheus: נסי קודם רק את שם המדד, בלי פונקציות, ובדקי את שמות ה־labels.

## הפתרון

הפתרון המלא בתיקייה `solution/`. הוא עובד כי כל שלושת השירותים חולקים propagator אחד של W3C ושם שירות
במשאב: `otelhttp` ממשיך מעקב מהכותרת `traceparent`, `kotel` כותב אותה ל־headers של הרשומה וקורא אותה בצד
השני, ו־`otelhttp.NewTransport` שולח אותה למקבל, שהכישור של FastAPI ממשיך ממנה. ה־collector הוא כתובת
אחת לכולם, ושלושת האחסונים מקבלים OTLP ישירות; Grafana קושרת ביניהם דרך מזהה המעקב.

## מה התרגיל מדגים

- הגדרה של OpenTelemetry בשתי שפות, עם מוסכמות אחידות, כך שדשבורד אחד מתאים לשתיהן.
- מחסנית תצפיתיות שלמה שעולה בפקודה אחת ונבדקת אוטומטית דרך ה־APIs שלה.
- RED לשירות ו־USE לעובד, ודשבורד שנשמר כקוד.
- הפצת הקשר מעקב דרך Kafka, ובדיקה שלה גם בלי broker וגם מולו.
- את התשתית ש־Hookline יקבל באבן הדרך M8: `internal/telemetry`, הדשבורד, וחוקים בתיקייה `data/rules/`.

</div>
