---
direction: rtl
tags: [senior-course, book]
chapter: 19
---
# Docker כמו שצריך

- **image:** שכבות של קבצים ומסמך תצורה; קונטיינר מוסיף שכבה אחת לכתיבה.
- **מטמון:** החמצה אחת בונה מחדש את כל מה שאחריה; מה שמשתנה לעתים רחוקות בא ראשון.
- **הקשר:** הקובץ `.dockerignore` שומר על המטמון ומונע דליפה; מחיקה בשכבה מאוחרת לא מוחקת.
- **Go:** בנייה רב־שלבית, בינארי סטטי, `-trimpath`, בלי סמלים, מטמון מוצמד לקומפילציה.
- **distroless:** בלי מעטפת ובלי כלים; ניפוי עם לוגים וקונטיינר ניפוי; בדיקת בריאות כבינארי.
- **root:** משתמש מספרי שאינו root; בודקים מבחוץ, עם תצורת ה־image ועם `docker top`.
- **Python:** uv בשלב הבנייה בלבד; תלויות לפני קוד; סביבה לא ניתנת לעריכה; לאזין לכל הכתובות.
- **compose:** בדיקת בריאות ו־`service_healthy`; Kubernetes מתעלם משתיהן ומשתמש ב־probes.
- **סריקה:** SBOM נשמר; שני סורקים; כל ממצא מקבל החלטה; בלי להצמיד את השקע של Docker.
- **שמות:** גרסה ומזהה commit שלא זזים; digest לשחזור; לא `latest`.
- **שאלה פתוחה:** ___

<div dir="ltr" align="left">

```text
build        docker build -t shorty:dev --build-arg GIT_SHA=$(git rev-parse --short=12 HEAD) shorty
layers       docker history shorty:dev --format '{{.Size}}\t{{.CreatedBy}}'
go flags     CGO_ENABLED=0 go build -trimpath -ldflags='-s -w'
cache mount  RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build
uv           COPY --from=ghcr.io/astral-sh/uv:<pin> /uv /bin/uv ; uv sync --locked --no-dev --no-install-project
user         USER 65532:65532 ; docker image inspect --format '{{.Config.User}}' IMG ; docker top CTR -o pid,uid
read-only    docker run --read-only [--tmpfs /tmp] ...   compose: read_only: true
compose      healthcheck.test ["CMD", ...] ; depends_on.X.condition: service_healthy ; up --build --wait
scan         docker save IMG -o x.tar ; syft docker-archive:x.tar ; grype sbom:x.spdx.json ; trivy image --input x.tar
digest       docker buildx imagetools inspect IMG
mine         shorty ___ MB  naive ___ MB  receiver ___ MB  findings kept ___  decision ___
```

</div>

קישורים: [[18-locks-leaders-sagas]] · [[20-kubernetes-on-kind]]
