---
direction: rtl
lang: he
title: "תוכן העניינים הראשי"
---

<div dir="rtl" lang="he">

# מהנדסת תוכנה בכירה בתעשייה: ספר לימוד ותוכנית למידה

ספר לימוד בעברית למפתחת מנוסה שמכוונת לתפקיד backend או תשתיות בכיר בשוק הישראלי של סוף 2026. Go מאפס
ועד הרגלי ייצור, Python ברמה בכירה, יסודות backend ומערכות מבוזרות, Kubernetes, תצפיתיות, אבטחה, אומנות
הבכירות והכנה לראיונות. הספר בנוי סביב קוד שרץ באמת, ומסתיים בפרויקט תיק עבודות: Hookline, פלטפורמה
למסירת webhooks.

איך קוראים: הדרך הנוחה היא גרסת ה־HTML בתיקייה `site/`. אפשר גם לפתוח את התיקייה `book/` ב־Obsidian
עם התוסף RTL Support. לוח הלימוד היומי נמצא בנספח ו, ובאנגלית בקובץ `00-stage1/05-daily-schedule.md`.

## מצב הכתיבה

| סימון | משמעות |
| --- | --- |
| ✅ | הפרק כתוב במלואו, כולל תרגיל ופתק Obsidian |
| 🛠 | בכתיבה |
| ⏳ | טרם נכתב |

## חלק 0: לפני שמתחילים

| פרק | קובץ | מצב |
| --- | --- | --- |
| פרק 0: איך ללמוד מהספר הזה | [`00-how-to-study.md`](00-how-to-study.md) | ⏳ |
| פרק 1: מה זה „בכירה” ב־2026 | [`01-senior-in-2026.md`](01-senior-in-2026.md) | ⏳ |
| פרק 2: סביבת העבודה: WSL2, Docker, Go ו־uv | [`02-workbench-wsl2.md`](02-workbench-wsl2.md) | ⏳ |

## חלק א: Go

| פרק | קובץ | מצב |
| --- | --- | --- |
| פרק 3: צעדים ראשונים ב־Go | [`03-go-first-steps.md`](03-go-first-steps.md) | ⏳ |
| פרק 4: טיפוסים, ממשקים ושגיאות | [`04-go-types-interfaces-errors.md`](04-go-types-interfaces-errors.md) | ⏳ |
| פרק 5: מקביליות ב־Go | [`05-go-concurrency.md`](05-go-concurrency.md) | ⏳ |
| פרק 6: שירות HTTP לייצור | [`06-go-http-service.md`](06-go-http-service.md) | ⏳ |

## חלק ב: Python בכירה

| פרק | קובץ | מצב |
| --- | --- | --- |
| פרק 7: Python עם טיפוסים, אריזה ובדיקות | [`07-python-typed-packaged.md`](07-python-typed-packaged.md) | ⏳ |
| פרק 8: Python אסינכרונית ו־FastAPI | [`08-python-async-fastapi.md`](08-python-async-fastapi.md) | ⏳ |

## חלק ג: יסודות backend

| פרק | קובץ | מצב |
| --- | --- | --- |
| פרק 9: עיצוב API ו־gRPC | [`09-api-design-grpc.md`](09-api-design-grpc.md) | ⏳ |
| פרק 10: PostgreSQL: שאילתות ואינדקסים | [`10-postgres-queries-indexes.md`](10-postgres-queries-indexes.md) | ⏳ |
| פרק 11: PostgreSQL: טרנזקציות, נעילות ומיגרציות | [`11-postgres-transactions-migrations.md`](11-postgres-transactions-migrations.md) | ⏳ |
| פרק 12: Redis ומטמון | [`12-redis-caching.md`](12-redis-caching.md) | ⏳ |
| פרק 13: Kafka עם Redpanda | [`13-kafka-redpanda.md`](13-kafka-redpanda.md) | ⏳ |
| פרק 14: עבודות רקע וניסיונות חוזרים | [`14-background-jobs.md`](14-background-jobs.md) | ⏳ |

## חלק ד: מערכות מבוזרות

| פרק | קובץ | מצב |
| --- | --- | --- |
| פרק 15: זמנים קצובים, ניסיונות חוזרים ומפסקי זרם | [`15-timeouts-retries-breakers.md`](15-timeouts-retries-breakers.md) | ⏳ |
| פרק 16: עומס: הגבלת קצב, לחץ חוזר והשלת עומס | [`16-overload-rate-limiting.md`](16-overload-rate-limiting.md) | ⏳ |
| פרק 17: עקביות, שכפול וחלוקה למחיצות | [`17-consistency-replication-partitioning.md`](17-consistency-replication-partitioning.md) | ⏳ |
| פרק 18: נעילות, בחירת מנהיג וסאגות | [`18-locks-leaders-sagas.md`](18-locks-leaders-sagas.md) | ⏳ |

## חלק ה: Cloud-native

| פרק | קובץ | מצב |
| --- | --- | --- |
| פרק 19: Docker כמו שצריך | [`19-docker-done-right.md`](19-docker-done-right.md) | ⏳ |
| פרק 20: Kubernetes על kind | [`20-kubernetes-on-kind.md`](20-kubernetes-on-kind.md) | ⏳ |
| פרק 21: הפעלת Kubernetes ו־Helm | [`21-kubernetes-operations-helm.md`](21-kubernetes-operations-helm.md) | ⏳ |
| פרק 22: Terraform | [`22-terraform.md`](22-terraform.md) | ⏳ |
| פרק 23: GitHub Actions | [`23-github-actions.md`](23-github-actions.md) | ⏳ |

## חלק ו: תצפיתיות ותפעול

| פרק | קובץ | מצב |
| --- | --- | --- |
| פרק 24: טלמטריה: OpenTelemetry, Prometheus, Grafana ו־Loki | [`24-telemetry.md`](24-telemetry.md) | ⏳ |
| פרק 25: SLO, התראות, תורנות ותחקיר | [`25-slo-alerting-incidents.md`](25-slo-alerting-incidents.md) | ⏳ |

## חלק ז: איכות ואבטחה

| פרק | קובץ | מצב |
| --- | --- | --- |
| פרק 26: אסטרטגיית בדיקות ובדיקות עומס | [`26-testing-strategy-load.md`](26-testing-strategy-load.md) | ⏳ |
| פרק 27: סקירת קוד וניתוח סטטי | [`27-code-review-static-analysis.md`](27-code-review-static-analysis.md) | ⏳ |
| פרק 28: אבטחה לשירותי backend | [`28-security-backend.md`](28-security-backend.md) | ⏳ |

## חלק ח: אומנות הבכירות

| פרק | קובץ | מצב |
| --- | --- | --- |
| פרק 29: שיטת תכנון ומסמכי תכנון | [`29-system-design-method.md`](29-system-design-method.md) | ⏳ |
| פרק 30: מערכות שמשתנות | [`30-evolving-systems.md`](30-evolving-systems.md) | ⏳ |
| פרק 31: לעבוד כבכירה | [`31-working-as-senior.md`](31-working-as-senior.md) | ⏳ |
| פרק 32: הנדסה בעזרת AI | [`32-ai-assisted-engineering.md`](32-ai-assisted-engineering.md) | ⏳ |
| פרק 33: היתרון של Rust | [`33-rust-edge.md`](33-rust-edge.md) | ⏳ |

## חלק ט: קריירה

| פרק | קובץ | מצב |
| --- | --- | --- |
| פרק 34: תהליך הגיוס הבכיר | [`34-senior-interview-loop.md`](34-senior-interview-loop.md) | ⏳ |
| פרק 35: תרגילי קוד מעשיים | [`35-coding-drills.md`](35-coding-drills.md) | ⏳ |
| פרק 36: תכנון מערכות בראיון ומאגר השאלות | [`36-system-design-interviews.md`](36-system-design-interviews.md) | ⏳ |

## חלק י: פרויקט הגמר

| פרק | קובץ | מצב |
| --- | --- | --- |
| פרק 37: Hookline: מפרט וארכיטקטורה | [`37-hookline-spec.md`](37-hookline-spec.md) | ⏳ |
| פרק 38: אבני הדרך | [`38-hookline-milestones.md`](38-hookline-milestones.md) | ⏳ |
| פרק 39: הערכה, הדגמה וראיון על הפרויקט | [`39-hookline-evaluation.md`](39-hookline-evaluation.md) | ⏳ |

## נספחים

| נספח | קובץ | מצב |
| --- | --- | --- |
| א: מילון מונחים | [`appendices/a-glossary.md`](appendices/a-glossary.md) | ⏳ |
| ב: פתרון תקלות | [`appendices/b-troubleshooting.md`](appendices/b-troubleshooting.md) | ⏳ |
| ג: גרסאות כלים | [`appendices/c-tool-versions.md`](appendices/c-tool-versions.md) | ⏳ |
| ד: תבניות Obsidian | [`appendices/d-obsidian-templates.md`](appendices/d-obsidian-templates.md) | ⏳ |
| ה: רשימות בדיקה | [`appendices/e-checklists.md`](appendices/e-checklists.md) | ⏳ |
| ו: לוח הלימוד היומי | [`appendices/f-daily-schedule.md`](appendices/f-daily-schedule.md) | ⏳ |

</div>
