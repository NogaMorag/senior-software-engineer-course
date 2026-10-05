---
direction: rtl
tags: [senior-course, book]
chapter: 24
---
# טלמטריה: OpenTelemetry, Prometheus, Grafana ו־Loki

- **שלושה אותות:** מדד אומר שיש בעיה, מעקב איפה, לוג למה.
- **הקשר:** מזהה המעקב בכל שורת לוג, ודוגמאות על ה־histogram.
- **collector:** כתובת אחת לשירותים, עיבוד במקום אחד, שלושה יעדים.
- **Go:** ספריות תלויות ב־API, ו־`main` מתקין SDK; הפצת הקשר של W3C; כיבוי שמרוקן.
- **otelhttp:** מחוץ ללוג ולהתאוששות; התבנית צריכה להגיע אליו.
- **Python:** ספקים במפורש; המוסכמות היציבות של HTTP בבקשה מפורשת.
- **PromQL:** `rate` ואז `sum`; ב־`histogram_quantile` משאירים `le`.
- **RED ו־USE:** שירות מול משאב; ניצולת כ־counter של שניות עבודה.
- **Loki:** labels מעטים; מזהה המעקב במטא־נתונים מובנים.
- **Kafka:** `traceparent` ב־headers; הורה מול קישור.
- **שאלה פתוחה:** ___

<div dir="ltr" align="left">

```text
docker compose -p ch24 -f data/compose.yaml up -d --build --wait
docker compose -p ch24 -f data/compose.yaml run --rm loadgen -d 2m
bash tests/check_stack.sh
docker compose -p ch24 -f data/compose.yaml down -v
OTEL_EXPORTER_OTLP_ENDPOINT=http://collector:4318   OTEL_METRIC_EXPORT_INTERVAL=5000
sum by (http_route) (rate(http_server_request_duration_seconds_count{job="shorty"}[1m]))
histogram_quantile(0.99, sum by (le) (rate(http_server_request_duration_seconds_bucket{job="shorty"}[5m])))
rate(linkworker_busy_seconds_total[1m])     linkworker_wait_seconds_bucket
{service_name="shorty"} | trace_id="..."     { resource.service.name = "shorty" && status = error }
traceparent: 00-<32 hex trace id>-<16 hex span id>-01
grafana http://localhost:3000  prometheus :9090  loki :3100  tempo :3200
my p99 for shorty: ___ ms at ___ req/s   my trace id: ___
```

</div>

קישורים: [[23-github-actions]] · [[25-slo-alerting-incidents]]
