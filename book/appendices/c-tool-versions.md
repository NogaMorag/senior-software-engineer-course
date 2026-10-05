---
direction: rtl
lang: he
title: "נספח ג: גרסאות כלים"
---

<div dir="rtl" lang="he">

# נספח ג: גרסאות כלים

הנספח הזה מרכז את כל הגרסאות שהספר מצמיד, **נכון לאוקטובר 2026**. כל הגרסאות של הכלים, המודולים, החבילות
וה־images נבדקו ב־5 באוקטובר 2026 מול מקור ראשי. הפרקים עצמם לא כותבים גרסאות בטקסט; הם אומרים „נכון לאוקטובר
2026” ומפנים לכאן. הגרסה שקובעת בפועל היא זו שבקובץ, והנספח מראה איפה כל אחת יושבת.

## איפה הגרסאות חיות

| סוג | הקובץ |
| --- | --- |
| כלים שמותקנים בתיקיית הבית שלך | פקודות ההתקנה בפרק 2, והמינימום שהרופא בודק (`exercises/ch02-workbench/solution/doctor.sh`) |
| מודולים של Go | `go.mod` של כל מודול, עם `go.sum` לצידו |
| חבילות Python | `pyproject.toml` עם גרסאות מדויקות, ו־`uv.lock` שנועל את כל העץ |
| images | קובצי compose בתיקייה `data/` של כל תרגיל, `Dockerfile`, קובצי kind, וסקריפטים שמריצים כלי בקונטיינר |
| GitHub Actions | קובצי ה־workflow של תרגיל פרק 23 |
| Terraform | `versions.tf` של כל תצורה, ו־`.terraform.lock.hcl` שנוצר ב־`init` |
| Rust | `rust-toolchain.toml` של תרגיל פרק 33 |

## איך בודקים אם יש גרסה חדשה

| מה | הפקודה או המקום |
| --- | --- |
| מודולים של Go | `go list -m -u all` בתיקיית המודול: כל שורה עם סוגריים מרובעים היא גרסה חדשה יותר |
| חבילת Python אחת | `uv lock --upgrade-package <pkg>`, ואז `uv sync` |
| image | דף התגיות ב־Docker Hub (או ב־GHCR, לאימג'ים שמתפרסמים שם) |
| כלי שורת פקודה ו־actions | דף ה־Releases של הפרויקט ב־GitHub |
| כלי Go שמותקן ב־`go install` | `go list -m -versions` עם שם המודול של הכלי |

הפקודות, כשהן נכתבות במלואן:

<div dir="ltr" align="left">

```bash
go list -m -u all
uv lock --upgrade-package fastapi && uv sync
go list -m -versions honnef.co/go/tools
```

</div>

הכלל: **משדרגים בכוונה, אחד בכל פעם, ומריצים את הבדיקות מחדש.** שדרוג של כמה דברים יחד הופך כישלון לחידה, כי
אי אפשר לדעת איזה מהם שבר. גרסה ראשית חדשה (כמו המעבר ל־Helm 4 או ל־k6 2) כמעט תמיד שוברת משהו,
ושם קוראים קודם את הערות השחרור. ואחרי שדרוג, מעדכנים את הקובץ שבו הגרסה מוצמדת ואת השורה כאן.

העמודה „נבדק ב” אומרת מתי ומול איזה מקור נבדקה הגרסה:

| קיצור | המקור |
| --- | --- |
| GH | דף ה־Releases ב־GitHub, דרך ה־API |
| proxy | ‏`proxy.golang.org`, המקור הקובע למודולים של Go |
| PyPI | ה־API של PyPI |
| Hub | ה־API של התגיות ב־Docker Hub |
| רשמי | דף ההורדות או הערות השחרור של היצרן |
| קובץ | הגרסה כפי שנבחרה בקובץ התרגיל; אין לה שורה ברשומת הבדיקה של 5 באוקטובר |

## שרשרת הכלים

| כלי | גרסה | איפה היא מוצמדת | נבדק ב |
| --- | --- | --- | --- |
| WSL | 3.0.1 | פרק 2, סעיף 2.2 | 2026-10-05, GH |
| Ubuntu ב־WSL | 26.04 LTS (26.04.1) | ברירת המחדל של `wsl --install`; פרק 2, סעיף 2.2 | 2026-10-05, רשמי |
| Docker Desktop | 4.93.0, עם Engine 29.8.1 | פרק 2, סעיף 2.4; הרופא דורש Engine 29.0 ומעלה | 2026-10-05, רשמי |
| Docker Compose | v5.6.0 (הגרסה האחרונה); הרופא דורש 2.20 ומעלה | `doctor.sh` | 2026-10-05, GH |
| Git | גרסת ההפצה; לפחות 2.28 | `doctor.sh` | |
| VS Code והרחבת WSL | הגרסה הנוכחית | `doctor.sh` בודק רק שהפקודה קיימת | |
| Go | 1.27.1 | פרק 2, סעיף 2.7; `doctor.sh`; השורה `go 1.27` בכל `go.mod`; ‏`golang:1.27.1` בכל `Dockerfile` | 2026-10-05, רשמי |
| uv | 0.12.23 | `doctor.sh`; ‏`uv_build>=0.12.23,<0.13` בכל `pyproject.toml`; ‏`ch08-receiver.yml` בתרגיל פרק 23 | 2026-10-05, PyPI |
| Python | 3.14 (3.14.8) | `requires-python` בכל `pyproject.toml`; ‏`python:3.14.8-slim` בכל `Dockerfile` | 2026-10-05, רשמי |
| staticcheck | 2026.2.1 (המודול v0.8.1) | פרק 2, סעיף 2.7.1; `doctor.sh`; ‏`go-checks/action.yml` בתרגיל פרק 23 | 2026-10-05, GH |
| golangci-lint | v2.14.0 | README של תרגיל פרק 27 (כתובת ההורדה) | 2026-10-05, GH |
| govulncheck | v1.8.0 | פרק 2, סעיף 2.7.1; `doctor.sh` | 2026-10-05, proxy |
| goose (הכלי) | v3.28.0 | פרק 2, סעיף 2.7.1; `doctor.sh` | 2026-10-05, proxy |
| buf | v1.73.0 | פרק 2, סעיף 2.7.1; `doctor.sh` | 2026-10-05, GH |
| protoc-gen-go | v1.36.12 | פרק 2, סעיף 2.7.1; `doctor.sh` | 2026-10-05, proxy |
| protoc-gen-go-grpc | v1.6.2 | פרק 2, סעיף 2.7.1; `doctor.sh` | 2026-10-05, proxy |
| protoc | v36.2 | לא בשימוש בתרגילים: buf מחליף אותו | 2026-10-05, GH |
| actionlint | v1.7.12 | פרק 2, סעיף 2.7.1; `doctor.sh` | 2026-10-05, GH |

## הכלים של חלק ה (יום 57)

| כלי | גרסה | איפה היא מוצמדת | נבדק ב |
| --- | --- | --- | --- |
| kind | v0.33.0 | פרק 2, סעיף 2.10; `doctor.sh --part-e`; ‏`ch06-shorty.yml` בתרגיל פרק 23 | 2026-10-05, GH |
| ‏node image של kind | `kindest/node:v1.37.0`, מוצמד לפי digest | `kind-config.yaml` בתרגילים 20, 21 ו־22; ‏`ch06-shorty.yml` בתרגיל פרק 23 | 2026-10-05, GH |
| kubectl | v1.37.1 | פרק 2, סעיף 2.10; `doctor.sh --part-e`; ‏`ch06-shorty.yml` בתרגיל פרק 23 | 2026-10-05, רשמי |
| Helm | v4.3.0 | פרק 2, סעיף 2.10; `doctor.sh --part-e` | 2026-10-05, GH |
| Terraform | 1.16.5 | פרק 2, סעיף 2.10; `doctor.sh --part-e`; ‏`versions.tf` דורש 1.16.0 ומעלה | 2026-10-05, רשמי |
| ‏provider ‏`hashicorp/kubernetes` | 3.3.0 | `versions.tf` בתרגיל פרק 22 | 2026-10-05, GH |
| ‏provider ‏`hashicorp/helm` | 3.3.0 | `versions.tf` בתרגיל פרק 22 | 2026-10-05, GH |
| ‏provider ‏`hashicorp/aws` | 6.67.0 | `optional-aws/versions.tf` בתרגיל פרק 22 (אופציונלי, לא הוחל) | 2026-10-05, קובץ |
| k6 | v2.3.0 | פרק 2, סעיף 2.10; `doctor.sh --part-e` | 2026-10-05, GH |
| Syft | v1.54.0 | `tests/scan_images.sh` בתרגיל פרק 19; README של תרגיל פרק 28 | 2026-10-05, GH |
| Grype | v0.120.0 | `tests/scan_images.sh` בתרגיל פרק 19; README של תרגיל פרק 28 | 2026-10-05, GH |
| Trivy | v0.75.0 (התגית של ה־image היא `0.75.0`, בלי `v`) | `tests/scan_images.sh` בתרגיל פרק 19; README של תרגיל פרק 28 | 2026-10-05, GH |

## מודולים של Go

| מודול | גרסה | איפה היא מוצמדת | נבדק ב |
| --- | --- | --- | --- |
| `github.com/jackc/pgx/v5` | v5.11.0 | תרגילים 10 עד 14, 17, 18, 26, 30, 35, ‏`ch23-ci/solution/ch11`, ‏Hookline | 2026-10-05, proxy |
| `github.com/redis/go-redis/v9` | v9.22.0 | תרגילים 12, 16, 18, ‏Hookline | 2026-10-05, proxy |
| `github.com/twmb/franz-go` | v1.22.1 | תרגילים 13, 24 (`shorty`), ‏Hookline | 2026-10-05, proxy |
| `github.com/twmb/franz-go/pkg/kmsg` | v1.14.0 | תרגיל 13, ‏Hookline | קובץ |
| `github.com/twmb/franz-go/plugin/kotel` | v1.7.1 | תרגיל 24 (`shorty`) | קובץ |
| `github.com/pressly/goose/v3` | v3.28.0 | תרגילים 11, 30, ‏`ch23-ci/solution/ch11`, ‏Hookline | 2026-10-05, proxy |
| `google.golang.org/grpc` | v1.84.0 | תרגיל 9 | 2026-10-05, proxy |
| `google.golang.org/protobuf` | v1.36.12 | תרגיל 9 | 2026-10-05, proxy |
| `go.opentelemetry.io/otel` וחבילות ה־SDK והמייצאים | v1.47.0 | תרגיל 24 (`shorty`) | 2026-10-05, proxy |
| ‏semconv של OpenTelemetry | החבילה `semconv/v1.43.0`, בתוך המודול `otel` | הייבוא בקוד של תרגיל 24 | 2026-10-05, GH |
| `go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp` | v0.23.0 | תרגיל 24 (`shorty`) | קובץ |
| ‏`otelhttp` (‏contrib) | v0.72.0 | תרגיל 24 (`shorty`) | 2026-10-05, proxy |
| `go.opentelemetry.io/contrib/instrumentation/runtime` | v0.72.0 | תרגיל 24 (`shorty`) | קובץ |
| `go.opentelemetry.io/contrib/bridges/otelslog` | v0.21.0 | תרגיל 24 (`shorty`) | קובץ |
| `github.com/testcontainers/testcontainers-go` | v0.44.0 | תרגיל 26 | 2026-10-05, proxy |
| `github.com/testcontainers/testcontainers-go/modules/postgres` | v0.44.0 | תרגיל 26 | קובץ (אותה גרסה כמו המודול הראשי) |
| `pgregory.net/rapid` | v1.3.0 | תרגיל 26, ‏Hookline | 2026-10-05, proxy |
| `github.com/santhosh-tekuri/jsonschema/v6` | v6.0.3 | תרגיל 26 | קובץ |
| `golang.org/x/time` | v0.16.0 | תרגיל 16 | 2026-10-05, proxy |
| `golang.org/x/sync` | v0.23.0 | תרגילים 5, 6, 12, 19, 20, 21, 24, ‏`ch23-ci/solution/ch06`, ‏Hookline | 2026-10-05, קובץ (נבדק בכתיבת התרגיל) |
| `golang.org/x/text` | v0.40.0 (התיקון), ו־v0.3.7 הפגיע בכוונה | `seeded/go.mod` בתרגיל פרק 27 | קובץ |
| `github.com/golang-jwt/jwt/v5` | v5.3.1 | תרגיל 28 | 2026-10-05, proxy |
| `github.com/coreos/go-oidc/v3` | v3.21.0 | לא מוצמד בקובץ: פרק 28 מזכיר אותו בלבד | 2026-10-05, proxy |
| `github.com/prometheus/client_golang` | v1.24.1 | לא מוצמד בקובץ: התרגילים מייצאים דרך OpenTelemetry | 2026-10-05, proxy |
| `github.com/sony/gobreaker/v2` | v2.4.0 | לא מוצמד בקובץ: פרק 15 מזכיר אותו להשוואה בלבד | 2026-10-05, proxy |
| `github.com/golang-migrate/migrate/v4` | v4.20.1 | לא בשימוש: הספר משתמש ב־goose | 2026-10-05, proxy |
| `github.com/stretchr/testify` | v1.12.1 | לא בשימוש: הבדיקות בספרייה הסטנדרטית | 2026-10-05, proxy |

## חבילות Python

| חבילה | גרסה | איפה היא מוצמדת | נבדק ב |
| --- | --- | --- | --- |
| fastapi | 0.142.2 | תרגילים 8, 19, 20, 23 (`ch08`), 24, ‏Hookline (`sink`) | 2026-10-05, PyPI |
| uvicorn | 0.54.0 | אותם קבצים כמו fastapi | 2026-10-05, PyPI |
| pydantic | 2.13.5 | תרגילים 7, 8 ושאר הפרויקטים עם FastAPI | 2026-10-05, PyPI |
| httpx | 0.28.1 | קבוצת `dev` בפרויקטים עם FastAPI | 2026-10-05, PyPI |
| pytest | 9.1.1 | קבוצת `dev` בכל `pyproject.toml` | 2026-10-05, PyPI |
| ruff | 0.16.10 | קבוצת `dev` בכל `pyproject.toml` | 2026-10-05, PyPI |
| mypy | 2.4.0 | קבוצת `dev` בכל `pyproject.toml` | 2026-10-05, PyPI |
| ‏`uv_build` (‏build backend) | 0.12.23 ומעלה, מתחת ל־0.13 | `[build-system]` בכל `pyproject.toml` | קובץ |
| opentelemetry-sdk, opentelemetry-api | 1.45.0 | `receiver/pyproject.toml` בתרגיל פרק 24 | 2026-10-05, PyPI |
| opentelemetry-exporter-otlp-proto-http | 1.45.0 | `receiver/pyproject.toml` בתרגיל פרק 24 | קובץ |
| opentelemetry-instrumentation-fastapi, opentelemetry-instrumentation-logging | 0.66b0 | `receiver/pyproject.toml` בתרגיל פרק 24 | קובץ |
| jsonschema | 4.26.0 | `consumer/pyproject.toml` בתרגיל פרק 26 | קובץ |
| rfc3339-validator | 0.1.4 | `consumer/pyproject.toml` בתרגיל פרק 26 | קובץ |
| types-jsonschema | 4.26.0.20260518 | `consumer/pyproject.toml` בתרגיל פרק 26 | קובץ |
| hypothesis | 6.168.4 | לא מוצמד בקובץ: מוזכר בהעשרה של פרק 26 | 2026-10-05, PyPI |
| psycopg | 3.3.6 | לא מוצמד בקובץ | 2026-10-05, PyPI |
| redis (‏redis-py) | 8.1.0 | לא מוצמד בקובץ | 2026-10-05, PyPI |
| asyncpg, sqlalchemy, alembic, aiokafka, confluent-kafka | 0.31.0, 2.1.3, 1.20.0, 0.14.0, 2.15.1 | לא בשימוש; נבדקו כחלופות | 2026-10-05, PyPI |

## images

| image | תגית | איפה היא מוצמדת | נבדק ב |
| --- | --- | --- | --- |
| `postgres` | 18.6 | compose בתרגילים 10 עד 14, 17, 18, 30, 35 ו־Hookline; ‏`pgtest.go` בתרגיל 26; ה־workflow של `ch11` בתרגיל 23 | 2026-10-05, Hub |
| `redis` | 8.10.2 | compose בתרגילים 12, 16, 18 ו־Hookline | 2026-10-05, Hub |
| `docker.redpanda.com/redpandadata/redpanda` | v26.2.3 | compose בתרגילים 13, 24 ו־Hookline | 2026-10-05, Hub |
| `docker.redpanda.com/redpandadata/console` | v3.12.0 | compose בתרגיל 13 וב־Hookline | 2026-10-05, Hub |
| `prom/prometheus` | v3.15.0 | compose בתרגיל 24; ‏`tests/run_promtool.sh` בתרגיל 25 | 2026-10-05, Hub |
| `grafana/grafana` | 13.2.3 | compose בתרגיל 24 | 2026-10-05, Hub |
| `grafana/loki` | 3.7.8 | compose בתרגיל 24 | 2026-10-05, GH |
| `grafana/tempo` | 3.1.0 | compose בתרגיל 24 | 2026-10-05, Hub |
| `grafana/alloy` | v1.20.1 | לא בשימוש: ה־collector של OpenTelemetry במקומו | 2026-10-05, Hub |
| ‏collector-contrib של OpenTelemetry | 0.162.0, מ־GHCR | compose בתרגיל 24 | 2026-10-05, GH |
| `gcr.io/distroless/static-debian13` | `nonroot` | שלב הריצה בכל `Dockerfile` של Go | 2026-10-05, רשמי |
| `golang` | 1.27.1 | שלב הבנייה בכל `Dockerfile` של Go | 2026-10-05, Hub |
| `python` | 3.14.8-slim | כל `Dockerfile` של Python | 2026-10-05, Hub |
| `redocly/cli` | 2.57.0 | README של תרגיל פרק 9 | 2026-10-05, Hub |
| `anchore/syft`, ‏`anchore/grype`, ‏`aquasec/trivy` | v1.54.0, ‏v0.120.0, ‏0.75.0 | `tests/scan_images.sh` בתרגיל 19; README של תרגיל 28 | 2026-10-05, GH |
| `ghcr.io/yannh/kubeconform` | v0.8.0 | `tests/check_static.sh` בתרגיל 21 | 2026-10-05, קובץ (נבדק בכתיבת התרגיל) |
| `busybox` | 1.37.0 | `tests/drill_check.sh` בתרגיל 21; פרק 21, סעיף 21.5.2 | 2026-10-05, קובץ (נבדק בכתיבת התרגיל) |
| ‏cloud-provider-kind (‏`cloud-controller-manager`) | v0.12.0 | README של תרגיל פרק 21 | 2026-10-05, קובץ (נבדק בכתיבת התרגיל) |

## GitHub Actions

קובצי הבדיקות מצמידים גרסה ראשית. קובץ הפרסום (`release.yml`) מצמיד SHA מלא, עם התגית כהערה
(פרק 23, סעיף 23.6).

| action | גרסה | איפה היא מוצמדת | נבדק ב |
| --- | --- | --- | --- |
| `actions/checkout` | v7; ב־`release.yml` ה־SHA של v7.0.1 | כל ה־workflows בתרגיל פרק 23 | 2026-10-05, GH |
| `actions/setup-go` | v7 | `go-checks/action.yml` בתרגיל פרק 23 | 2026-10-05, GH |
| `astral-sh/setup-uv` | v10 | `ch08-receiver.yml` | 2026-10-05, GH |
| `docker/setup-buildx-action` | v4; ב־`release.yml` ה־SHA של v4.4.1 | `ch06-shorty.yml`, ‏`release.yml` | 2026-10-05, GH |
| `docker/build-push-action` | v7; ב־`release.yml` ה־SHA של v7.4.0 | `ch06-shorty.yml`, ‏`release.yml` | 2026-10-05, GH |
| `docker/login-action` | v4; ב־`release.yml` ה־SHA של v4.6.0 | `ch06-shorty.yml`, ‏`release.yml` | 2026-10-05, GH |
| `docker/metadata-action` | v6; ב־`release.yml` ה־SHA של v6.2.0 | `ch06-shorty.yml`, ‏`release.yml` | 2026-10-05, GH |
| `helm/kind-action` | v1.15.0 | `ch06-shorty.yml` | 2026-10-05, GH |
| `actions/setup-python`, ‏`actions/cache`, ‏`actions/upload-artifact` | v7, ‏v6, ‏v7 | לא בשימוש בקובצי התרגיל | 2026-10-05, GH |
| `azure/setup-helm`, ‏`hashicorp/setup-terraform` | v5, ‏v4 | לא בשימוש בקובצי התרגיל | 2026-10-05, GH |
| `golangci/golangci-lint-action`, ‏`grafana/setup-k6-action` | v9, ‏v1 | לא בשימוש בקובצי התרגיל | 2026-10-05, GH |
| `softprops/action-gh-release` | v3 | מוזכר בפרק 23; התרגיל משתמש ב־`gh release create` | 2026-10-05, קובץ (נבדק בכתיבת התרגיל) |

## שונות

| כלי | גרסה | איפה היא מוצמדת | נבדק ב |
| --- | --- | --- | --- |
| metrics-server | v0.9.0 | `data/metrics-server/kustomization.yaml` בתרגיל פרק 21 | 2026-10-05, קובץ (נבדק בכתיבת התרגיל) |
| shellcheck | 0.11.0 (הגרסה היציבה) | README של תרגיל פרק 23 מתקין אותו מ־`apt`, בגרסת ההפצה | 2026-10-05, קובץ (נבדק בכתיבת התרגיל) |
| Rust | 1.99.0, עם rustfmt ו־clippy | `rust-toolchain.toml` בתרגיל פרק 33 (אופציונלי) | 2026-10-05, רשמי |

## אי־התאמות ידועות

נמצאו בסריקה של קובצי התרגילים מול הטבלאות למעלה, באוקטובר 2026. הן לא שוברות אף בדיקה, אבל כדאי לדעת עליהן
לפני שמשדרגים:

- **‏`golang.org/x/sync` בשלוש גרסאות.** כל המודולים שמצמידים אותו ישירות, כולל Hookline, מצמידים v0.23.0. אחרים
  מקבלים אותו כתלות עקיפה: v0.22.0 בתרגילים 11, 26, 30 וב־`ch23-ci/solution/ch11`, ו־v0.17.0 בתרגילים 10, 13, 14,
  17, 18 ו־35.
- **‏`golang.org/x/text` כתלות עקיפה** מופיע בארבע גרסאות שונות בין התרגילים (v0.29.0 עד v0.42.0). הוא לא
  מוצמד במפורש, ו־Go בוחר אותו לפי התלויות. הגרסה הפגיעה v0.3.7 בתרגיל פרק 27 היא בכוונה.
- **גבול עליון ל־Python.** בתרגילים 1 ו־2 כתוב `>=3.14,<3.15`, ובשאר `>=3.14` בלבד.
- **‏shellcheck** מוצמד בגרסה 0.11.0 ברשימת הכלים, אבל התרגיל מתקין את גרסת ההפצה, שיכולה להיות אחרת.
- **‏`actions/setup-go`** קורא את הגרסה מהשורה `go 1.27` שב־`go.mod`. ב־CI הגרסה נקבעת לפי השורה הזאת, ולא
  לפי 1.27.1 של שרשרת הכלים המקומית.

</div>
