---
direction: rtl
lang: he
title: "פרק 19: Docker כמו שצריך"
chapter: 19
part: "חלק ה: Cloud-native"
---

<div dir="rtl" lang="he">

# פרק 19: Docker כמו שצריך

כל שירות שתכתבי בעבודה יגיע לייצור בתוך image. image טוב הוא קטן, נבנה מהר כי הוא משתמש במטמון, רץ
כמשתמש שאינו root, ואפשר לומר בדיוק מה יש בתוכו. image רע הוא ג'יגה־בייט של קומפיילר, קוד מקור והרשאות
root, שאף אחד לא יודע מה נכנס אליו. בראיון לתפקיד בכיר ב־backend או בתשתית, השאלה „איך תבני image לשירות
Go?” היא שאלת פתיחה, והפרק הזה נותן לך את התשובה, עם המספרים שתמדדי בעצמך.

## מה נלמד בפרק

- מה זה image: שכבות, תצורה, ואיך סדר ההוראות קובע מה נבנה מחדש.
- לבנות את shorty מפרק 6 ב־Dockerfile רב־שלבי, עם מטמון של BuildKit.
- להריץ על distroless כמשתמש שאינו root, עם מערכת קבצים לקריאה בלבד, ולבדוק את זה כמו שצריך.
- לבנות image ל־receiver מפרק 8 עם uv, בלי uv בתוצאה.
- להרים את שניהם ב־compose, עם בדיקות בריאות והמתנה לתלויות.
- לסרוק את ה־images ב־Trivy וב־Grype, להפיק SBOM ב־Syft, ולהחליט מה עושים עם ממצא.
- לתת ל־image שם נכון: תגית, תקציר (digest) ומזהה commit.

## ידע מקדים

- פרק 2: Docker Desktop פועל, והפקודה `docker run hello-world` מסעיף 2.9.4 עבדה.
- פרק 6: השירות shorty, נקודות הקצה שלו והתצורה מהסביבה. התרגיל מביא עותק משלו, עם שתי תוספות קטנות
  שהפרק מסביר.
- פרק 7 ופרק 8: פרויקט uv, הקובץ `uv.lock`, וה־receiver. גם ממנו התרגיל מביא עותק.
- אין צורך להתקין את הסורקים: הם רצים כקונטיינרים.

## 19.1 מה זה image: שכבות ומטמון

המונח image (תבנית שממנה מריצים קונטיינר) הוא שני דברים יחד. הראשון: ערימה של שכבות (layers), שכל אחת
היא ארכיון של שינויים במערכת הקבצים: קבצים שנוספו, השתנו או נמחקו. השני: מסמך תצורה קטן בפורמט JSON, ובו
פקודת ההפעלה, משתני הסביבה, המשתמש והפורטים. קונטיינר הוא ה־image ועוד שכבה אחת דקה שאפשר לכתוב אליה,
והיא נעלמת כשהקונטיינר נמחק.

כל הוראה ב־Dockerfile שמשנה קבצים יוצרת שכבה, ובעיקר `COPY` ו־`RUN`. הוראות אחרות, כמו `ENV`
ו־`USER`, משנות רק את התצורה. הפקודה הבאה מראה את השכבות של ה־image שתבני בסעיף 19.2, מהעליונה
לתחתונה:

<div dir="ltr" align="left">

```bash
docker history shorty:dev --format '{{.Size}}\t{{.CreatedBy}}'
```

</div>

<div dir="ltr" align="left">

```text
0B	ENTRYPOINT ["/shorty"]
0B	EXPOSE [8080/tcp]
0B	ENV SHORTY_ADDR=:8080
0B	USER 65532:65532
7.35MB	COPY /out/shorty /shorty # buildkit
0B	LABEL org.opencontainers.image.source=https:…
0B	ARG GIT_SHA=unknown
319kB	bazel build //common:cacerts_debian13_arm64_…
16.4kB	bazel build //common:os_release_debian13
```

</div>

השורות של 0 בתים הן תצורה; השורה של הבינארי היא השכבה היחידה שהוספת, ומתחתיה השכבות של בסיס ה־distroless.
המספרים אצלך יהיו אחרים מעט (הם מההרצה שבה נכתב הספר, על Mac עם מעבד arm64).

### 19.1.1 מטמון הבנייה

מנוע הבנייה של Docker, BuildKit (ברירת המחדל בגרסאות של השנים האחרונות), שומר כל שכבה שבנה. לפני כל
הוראה הוא שואל: האם כבר הרצתי בדיוק את ההוראה הזאת, על בדיוק אותה שכבה קודמת, עם בדיוק אותם קבצי קלט?
עבור `COPY` הקלט הוא סכום הבדיקה (checksum) של הקבצים שמועתקים, לא תאריך השינוי שלהם. אם כן, השכבה
נלקחת מהמטמון בלי להריץ דבר.

הכלל החשוב בפרק: **ברגע שהוראה אחת מחמיצה את המטמון, כל ההוראות אחריה נבנות מחדש.** לכן הסדר ב־Dockerfile
הוא החלטה הנדסית. מה שמשתנה לעתים רחוקות בא ראשון, ומה שמשתנה בכל commit בא אחרון:

| סדר | הוראה | משתנה כש... |
| --- | --- | --- |
| 1 | תמונת הבסיס | מעדכנים גרסה, פעם בחודש |
| 2 | העתקת `go.mod` ו־`go.sum`, הורדת מודולים | מוסיפים תלות |
| 3 | העתקת הקוד | כל commit |
| 4 | הקומפילציה | כל commit |

אם מעתיקים את כל הקוד לפני הורדת המודולים, כל שינוי של שורה אחת מוריד מחדש את כל התלויות. זה עדיין עובד,
ולכן הטעות הזאת שורדת שנים במאגרים אמיתיים.

### 19.1.2 הקשר הבנייה ו־`.dockerignore`

כשמריצים `docker build .`, הנקודה היא הקשר הבנייה (build context): התיקייה שממנה `COPY` מעתיקה. הקובץ
`.dockerignore` מוציא ממנו קבצים, וזה חשוב משתי סיבות. הראשונה, מטמון: קובץ שלא נכנס להקשר לא יכול לשבור
את השכבה של `COPY . .`. השנייה, דליפה: תיקיית `.git`, קובץ `.env` או מפתח שהשארת בטעות לא ייכנסו ל־image.

הקובץ `shorty/.dockerignore` של התרגיל מוציא גם את קובצי הבדיקות, כי הם לא משתתפים בבינארי:

<div dir="ltr" align="left">

```text
# Only what the build needs goes into the build context. Anything listed here
# can change without invalidating the COPY . . layer.
.git
bin/
**/*_test.go
*.out
Dockerfile*
.dockerignore
```

</div>

> **שימי לב:** שכבה שמחקה קובץ לא מקטינה את ה־image. הקובץ עדיין יושב בשכבה שמתחתיה, ומי שמוריד את
> ה־image מוריד גם אותו. סוד שהועתק ונמחק בהוראה הבאה עדיין שם, וכל אחד יכול לחלץ אותו.

## 19.2 Dockerfile רב־שלבי ל־Go

כדי לקמפל Go צריך את כל שרשרת הכלים: מאות מגה־בייט של קומפיילר, ספרייה סטנדרטית ומטמון מודולים. כדי
להריץ את התוצאה צריך קובץ אחד. בנייה רב־שלבית (multi-stage build) מפרידה בין השניים: שלב ראשון עם הכלים,
ושלב שני שמתחיל מבסיס ריק כמעט ומעתיק אליו רק את הבינארי. ה־image הסופי הוא השלב האחרון בלבד.

זה החלק הראשון של הקובץ `shorty/Dockerfile` בתרגיל, שלב הבנייה:

<div dir="ltr" align="left">

```dockerfile
# syntax=docker/dockerfile:1
# Multi-stage build for shorty: a full Go toolchain to compile, a distroless
# base to run. Build from this directory: docker build -t shorty:dev .

# ---- build stage: compiler, module cache, source ----
FROM golang:1.27.1 AS build
WORKDIR /src

# Dependencies first: this layer stays cached until go.mod or go.sum change.
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Then the source. A change to a .go file rebuilds from here down only.
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/shorty ./cmd/shorty
```

</div>

השורה הראשונה בוחרת את גרסת התחביר של ה־Dockerfile, וזה מה שמאפשר את `--mount`. השם `build` אחרי
`AS` מאפשר לשלב הבא להעתיק ממנו. התג של תמונת הבסיס נעול לגרסת Go של הספר (נכון לאוקטובר 2026; ראי
נספח ג).

### 19.2.1 שלושת הדגלים של הקומפילציה

| דגל | מה הוא עושה | למה כאן |
| --- | --- | --- |
| `CGO_ENABLED=0` | בלי קוד C, ולכן בלי תלות בספריית C של המערכת | הבינארי רץ על בסיס שאין בו ספריית C בכלל |
| `-trimpath` | מוחק מהבינארי את הנתיבים המקומיים של הקבצים | בנייה זהה מכל מכונה, ובלי לחשוף את מבנה התיקיות שלך |
| `-ldflags='-s -w'` | בלי טבלת סמלים ובלי מידע ניפוי (DWARF) | בינארי קטן יותר |

המחיר של הדגל האחרון: דיבאגר כמו Delve לא יוכל להציג משתנים בבינארי הזה. עקבות מחסנית (stack traces)
ב־panic עדיין כוללות שמות פונקציות ושורות, כי Go שומרת אותם בטבלה נפרדת שהדגל לא מוחק. ובשירות רגיל
שאין בו `import "C"`, הכיבוי של cgo לא מוותר על כלום. מי שמשתמשת בספרייה של C, כמו מנהל התקן של SQLite,
צריכה בסיס עם ספריית C.

### 19.2.2 cache mounts: מטמון שלא נכנס ל־image

הדגל `--mount=type=cache` מצמיד תיקייה שנשמרת בין בנייה לבנייה, אבל לא נכנסת לשכבה. ההבדל מהמטמון של
סעיף 19.1.1 הוא בגרעיניות. מטמון השכבות הוא הכול או כלום: שינוי של קובץ אחד מריץ את `go build` מאפס. מטמון
התיקייה שומר את התוצרים של כל חבילה בנפרד, ולכן הקומפיילר בונה מחדש רק את החבילות שהשתנו.

שתי התיקיות שהתרגיל מצמיד הן מטמון המודולים של Go, ומטמון הקומפילציה שלו. אצל משתמש root בתמונת
`golang`, אלה הנתיבים:

| תיקייה | מה יש בה |
| --- | --- |
| `/go/pkg/mod` | קוד המקור של התלויות שהורדו |
| `/root/.cache/go-build` | תוצרי קומפילציה לפי חבילה |

### 19.2.3 שלב הריצה

החלק השני של `shorty/Dockerfile` הוא שלב הריצה. סעיף 19.3 מסביר כל שורה בו:

<div dir="ltr" align="left">

```dockerfile
# ---- runtime stage: the binary and almost nothing else ----
FROM gcr.io/distroless/static-debian13:nonroot
ARG GIT_SHA=unknown
LABEL org.opencontainers.image.source="https://example.com/senior-course/ch19" \
      org.opencontainers.image.revision="${GIT_SHA}"
COPY --from=build /out/shorty /shorty
# Numeric, so Kubernetes can check runAsNonRoot (65532 is distroless "nonroot").
USER 65532:65532
ENV SHORTY_ADDR=:8080
EXPOSE 8080
ENTRYPOINT ["/shorty"]
```

</div>

הצורה `["/shorty"]`, עם סוגריים מרובעים, נקראת צורת exec: Docker מריץ את הבינארי ישירות, והוא מקבל את
האות `SIGTERM` כשהקונטיינר נעצר. בצורה בלי סוגריים, הפקודה רצה דרך מעטפת, והמעטפת היא שמקבלת את האות. זה
ההבדל בין הכיבוי המסודר מפרק 6 לבין עשר שניות של המתנה ואז `SIGKILL`. וב־distroless אין מעטפת בכלל, אז
הצורה השנייה פשוט לא תעבוד.

### 19.2.4 הבנייה הנאיבית, כדי למדוד

כדי להבין מה הרווחת, התרגיל כולל גם את הקובץ `shorty/Dockerfile.naive`. אל תשתמשי בו בשום מקום אחר:

<div dir="ltr" align="left">

```dockerfile
# The naive single-stage build, kept only to measure against. Do not ship this:
# it carries the whole toolchain and the source, and it runs as root.
FROM golang:1.27.1
WORKDIR /src
COPY . .
RUN go build -o /usr/local/bin/shorty ./cmd/shorty
ENTRYPOINT ["shorty"]
```

</div>

את הגודל של שני ה־images תמדדי ותרשמי בתרגיל. ההבדל צפוי להיות של שני סדרי גודל, וכדאי שתדעי לומר
את המספרים שלך בראיון.

> **איך Rust עושה את זה:** אותו עיקרון, עם שני הבדלים. בינארי סטטי ב־Rust דורש לבנות מול musl (המטרה
> `x86_64-unknown-linux-musl`), כי ברירת המחדל מקושרת ל־glibc. ומטמון התלויות נשבר כאן קל יותר, ולכן
> משתמשים בכלי כמו cargo-chef שבונה קודם רק את התלויות לשכבה נפרדת. ב־Go, הורדת המודולים ומטמון
> הקומפילציה של `go build` נותנים את זה בלי כלי נוסף.

## 19.3 distroless ומשתמש שאינו root

### 19.3.1 מה יש ב־distroless, ומה אין

הבסיס `gcr.io/distroless/static-debian13` (בסיס שמכיל רק את מה שתוכנית סטטית צריכה כדי לרוץ) מכיל
מעט מאוד:

| יש | אין |
| --- | --- |
| תעודות CA, כדי שבקשת HTTPS יוצאת תעבוד | מעטפת (`sh`, `bash`) |
| נתוני אזורי זמן | מנהל חבילות (`apt`) |
| הקבצים `/etc/passwd` ו־`/etc/group`, עם המשתמשים `root`, `nonroot` ו־`nobody` | `curl`, `ps`, `ls`, `id` |
| תיקיית `/tmp` | ספריית C |

התג `nonroot` של הבסיס כבר מגדיר את המשתמש 65532. התרגיל כותב את `USER` בכל זאת, ובמספרים: כך ה־image לא
תלוי בתג של הבסיס, ו־Kubernetes יכול לוודא שהמשתמש אינו root (פרק 20, סעיף 20.7). משתמש בשם, כמו
`nonroot`, הוא מחרוזת ש־Kubernetes לא יכול לבדוק בלי לפתוח את `/etc/passwd`, ולכן הוא מסרב להריץ אותו
כשמבקשים ממנו לוודא.

המחיר של distroless הוא בניפוי. אין לאן להיכנס עם `docker exec` ואין כלים להריץ. יש שלוש דרכים לעקוף את
זה, לפי סדר העדפה:

- **לוגים ומדדים טובים** (פרקים 6 ו־24), כך שלא צריך להיכנס.
- **קונטיינר ניפוי זמני** שמצטרף לקונטיינר הרץ, עם הכלים שלו. ב־Kubernetes זה `kubectl debug`, בפרק 21.
- **התג `debug`** של distroless, שמוסיף מעטפת busybox. לניפוי מקומי בלבד, לא לייצור.

### 19.3.2 למה לא root

קונטיינר אינו מכונה וירטואלית. התהליך שבתוכו רץ על הליבה של המארח, ובלי מיפוי משתמשים (user namespaces),
המשתמש 0 בקונטיינר הוא המשתמש 0 במארח. אם תוקף מוצא פרצה בשירות וגם דרך לברוח מהקונטיינר, ההבדל בין root
למשתמש רגיל הוא ההבדל בין שליטה במכונה לבין תקלה מקומית. זה מינימום הרשאות (least privilege) ברמת
ה־image: השירות לא צריך root כדי להאזין לפורט 8080, ולכן הוא לא מקבל root.

### 19.3.3 לבדוק שזה באמת לא root

הדרך הרגילה לבדוק היא להריץ את `id` בתוך ה־image. ב־distroless אין `id`, וזה מה שקורה:

<div dir="ltr" align="left">

```bash
docker run --rm --entrypoint id shorty:dev
```

</div>

<div dir="ltr" align="left">

```text
docker: Error response from daemon: failed to create task for container: failed to create shim task: OCI runtime create failed: runc create failed: unable to start container process: error during container init: exec: "id": executable file not found in $PATH
```

</div>

שתי דרכים נכונות. הראשונה קוראת את התצורה של ה־image. השנייה שואלת את Docker איזה משתמש מריץ את התהליך
בקונטיינר חי, מבחוץ, בלי צורך בכלים בפנים:

<div dir="ltr" align="left">

```bash
docker image inspect --format '{{.Config.User}}' shorty:dev
docker run -d --name shorty-check --read-only -p 127.0.0.1:18080:8080 -e SHORTY_BASE_URL=http://localhost:18080 shorty:dev
docker top shorty-check -o pid,uid
docker rm -f shorty-check
```

</div>

<div dir="ltr" align="left">

```text
65532:65532
PID                 UID
812100              65532
```

</div>

ב־image של Python, שיש בו `id`, הבדיקה הרגילה עובדת, והיא חלק מהבדיקות של התרגיל.

### 19.3.4 מערכת קבצים לקריאה בלבד

הדגל `--read-only` שבפקודה למעלה מצמיד את מערכת הקבצים של הקונטיינר לקריאה בלבד. תוקף שהשיג הרצת קוד לא
יכול להוריד כלי ולשמור אותו, ושירות שכותב לדיסק בלי שהתכוונת יתגלה מיד ולא אחרי שהדיסק התמלא. shorty לא
כותב לדיסק בכלל, ולכן הוא רץ כך בלי שינוי.

שירות שצריך מקום זמני מקבל תיקייה אחת בזיכרון ולא את כל מערכת הקבצים: ב־Docker עם הדגל `--tmpfs /tmp`,
וב־Kubernetes עם volume מסוג `emptyDir`. אותו עיקרון יחזור בפרק 20 כשדה בשם `readOnlyRootFilesystem`.

## 19.4 image ל־Python עם uv

ב־Go הבינארי הוא כל מה שצריך. ב־Python צריך מפרש, ספריות וקוד, ולכן אין גרסה של distroless שנוחה באותה
מידה, והבסיס הוא `python:3.14.8-slim` (Debian מוקטן עם Python). העיקרון של השלבים נשאר: uv מתקין סביבה
וירטואלית בשלב הבנייה, ושלב הריצה מעתיק רק את הסביבה. uv עצמו לא מגיע לייצור.

הדפוס כאן לקוח ממדריך השילוב של uv עם Docker, באתר docs.astral.sh (נבדק באוקטובר 2026). זה שלב הבנייה
בקובץ `receiver/Dockerfile`:

<div dir="ltr" align="left">

```dockerfile
# ---- build stage: uv resolves and installs into /app/.venv ----
FROM python:3.14.8-slim AS build
COPY --from=ghcr.io/astral-sh/uv:0.12.23 /uv /bin/uv
ENV UV_COMPILE_BYTECODE=1 \
    UV_LINK_MODE=copy \
    UV_PYTHON_DOWNLOADS=0
WORKDIR /app

# Dependencies first, from the lock file only: cached until uv.lock changes.
RUN --mount=type=cache,target=/root/.cache/uv \
    --mount=type=bind,source=uv.lock,target=uv.lock \
    --mount=type=bind,source=pyproject.toml,target=pyproject.toml \
    uv sync --locked --no-dev --no-install-project

# Then the project itself, installed as a regular (non-editable) package.
COPY . .
RUN --mount=type=cache,target=/root/.cache/uv \
    uv sync --locked --no-dev --no-editable
```

</div>

שורה אחר שורה:

| שורה | למה |
| --- | --- |
| `COPY --from=ghcr.io/astral-sh/uv:0.12.23` | מעתיקה את uv מה־image הרשמי שלו, בגרסה הנעולה, בלי סקריפט התקנה |
| `UV_COMPILE_BYTECODE=1` | מהדר מראש את קובצי `.pyc`, כך שההפעלה הראשונה מהירה יותר |
| `UV_LINK_MODE=copy` | המטמון בתיקייה מוצמדת אחרת, ושם קישור קשיח לא אפשרי; העתקה שקטה במקום אזהרה |
| `UV_PYTHON_DOWNLOADS=0` | משתמשת ב־Python של תמונת הבסיס, ולא מורידה אחר |
| `--mount=type=bind` לשני הקבצים | השכבה תלויה רק בהם, בלי להעתיק אותם לשכבה |
| `--no-install-project` | מתקינה רק את התלויות; הקוד שלך עוד לא כאן |
| `--no-dev` | בלי pytest, ruff ו־mypy |
| `--no-editable` | מתקינה את החבילה עצמה לתוך הסביבה, כך ששלב הריצה לא צריך את `src/` |

הדגל `--locked` נכשל אם קובץ הנעילה לא תואם ל־`pyproject.toml`. האפשרות האחרת, `--frozen`, מתקינה מקובץ
הנעילה בלי לבדוק את ההתאמה. בבנייה של image עדיף להיכשל: נעילה שלא עודכנה אחרי שינוי תלות היא באג שצריך
לראות.

### 19.4.1 שלב הריצה

<div dir="ltr" align="left">

```dockerfile
# ---- runtime stage: the same python, the virtual environment, no uv ----
FROM python:3.14.8-slim
ARG GIT_SHA=unknown
LABEL org.opencontainers.image.revision="${GIT_SHA}"
# The app never installs packages at runtime, so the base image's pip goes: its own
# package metadata is what the scanners reported. See section 19.6.
RUN rm -rf /usr/local/lib/python3.14/site-packages/pip* /usr/local/bin/pip* \
 && groupadd --system --gid 10001 app \
 && useradd --system --uid 10001 --gid app --no-create-home --shell /usr/sbin/nologin app
# Owned by root and not writable by the app user: the code cannot be changed at runtime.
COPY --from=build /app/.venv /app/.venv
ENV PATH="/app/.venv/bin:${PATH}" \
    PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1 \
    RECEIVER_HOST=0.0.0.0 \
    RECEIVER_PORT=8081
USER 10001:10001
EXPOSE 8081
ENTRYPOINT ["receiver"]
```

</div>

ארבע החלטות כאן:

- **אותה תמונת בסיס בשני השלבים.** בסביבה הווירטואלית יש קישור למפרש בנתיב שלו, ובסיס אחר בשלב הריצה
  יכול לשבור אותו.
- **הסביבה שייכת ל־root.** המשתמש `app` יכול לקרוא ולהריץ, אבל לא לשנות את הקוד. לא צריך `--chown`.
- **`RECEIVER_HOST=0.0.0.0`.** ברירת המחדל של ה־receiver היא `127.0.0.1`, וכתובת כזאת בתוך קונטיינר
  נגישה רק מתוכו. זו הטעות הנפוצה ביותר עם image ראשון של שרת.
- **הסרת pip.** ההחלטה הזאת באה מהסריקה בסעיף 19.6, והיא דוגמה לאיך ממצא הופך לשורה ב־Dockerfile.

## 19.5 compose לפיתוח

ה־compose (כלי שמרים כמה קונטיינרים יחד מקובץ אחד) כבר הופיע בפרקים על PostgreSQL ו־Redis. כאן הוא מריץ
את ה־images שלך, וזה מוסיף שתי שאלות: איך יודעים ששירות מוכן, ואיך שירות מחכה לשירות אחר.

### 19.5.1 בדיקת בריאות ב־image בלי מעטפת

בדיקת בריאות (healthcheck) ב־Docker היא פקודה שרצה בתוך הקונטיינר כל כמה שניות. קוד יציאה 0 פירושו בריא.
ב־image של Python אפשר להריץ שורת Python שפותחת כתובת. ב־distroless אין מעטפת ואין `curl`, ולכן הבדיקה
חייבת להיות תוכנית שנמצאת ב־image. התרגיל מוסיף ל־shorty פקודת משנה (subcommand) קטנה, בקובץ
`shorty/cmd/shorty/healthcheck.go`:

<div dir="ltr" align="left">

```go
// healthcheck GETs url and returns the process exit code: 0 for a 200 answer,
// 1 for anything else. Errors go to w, so they show up in docker inspect.
func healthcheck(url string, w io.Writer) int {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		fmt.Fprintln(w, "healthcheck:", err)
		return 1
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintln(w, "healthcheck:", url, "answered", resp.Status)
		return 1
	}
	return 0
}
```

</div>

הפונקציה `main` קוראת לה כשהארגומנט הראשון הוא `healthcheck`, כך שאותו בינארי משמש גם כשרת וגם כבודק.
זה לא מוסיף כמעט כלום לגודל, כי חבילת `net/http` כבר נמצאת בבינארי.

### 19.5.2 הקובץ `compose.yaml`

זה הקטע של shorty ושל שירות הבדיקה בקובץ `compose.yaml` של התרגיל:

<div dir="ltr" align="left">

```yaml
  shorty:
    build: ./shorty
    image: shorty:dev
    environment:
      SHORTY_BASE_URL: http://localhost:8080
      SHORTY_LOG_LEVEL: info
    ports:
      - "127.0.0.1:8080:8080"
    read_only: true
    healthcheck:
      # distroless has no shell and no curl: the binary checks itself.
      test: ["CMD", "/shorty", "healthcheck"]
      interval: 5s
      timeout: 3s
      retries: 5
      start_period: 2s

  smoke:
    image: receiver:dev
    depends_on:
      receiver:
        condition: service_healthy
      shorty:
        condition: service_healthy
    # "receiver" resolves through the compose network's DNS to the receiver container.
    entrypoint: ["receiver-send", "http://receiver:8081/hooks/orders", "--secret", "change-me"]
    restart: "no"
```

</div>

מה כל חלק נותן:

| חלק | משמעות |
| --- | --- |
| `127.0.0.1:8080:8080` | הפורט פתוח רק למחשב שלך, לא לרשת שאת מחוברת אליה |
| `read_only: true` | אותה מערכת קבצים לקריאה בלבד מסעיף 19.3.4, גם בפיתוח |
| `start_period` | זמן חסד: כישלונות בזמן ההפעלה לא נספרים |
| `condition: service_healthy` | שירות הבדיקה לא עולה עד ששני השירותים בריאים, לא רק רצים |
| `http://receiver:8081` | ב־compose כל שירות מקבל שם ברשת הפנימית, כמו ש־Kubernetes ייתן בפרק 20 |

בלי התנאי, `depends_on` מחכה רק שהקונטיינר יתחיל, ושירות שמתחבר למסד נתונים שעוד לא מקבל חיבורים נופל.
התנאי `service_healthy` סוגר את הפער הזה בפיתוח. בייצור, שירות חייב לשרוד תלות שלא זמינה בכל רגע, עם
ניסיונות חוזרים (פרק 15), ולכן הוא לא תחליף.

> **שימי לב:** Kubernetes מתעלם מההוראה `HEALTHCHECK` של ה־image ומבדיקות הבריאות של compose. יש לו
> בדיקות משלו (probes), עם שלושה סוגים ומשמעות שונה לכל אחד: פרק 20, סעיף 20.7. לכן התרגיל מגדיר את
> הבדיקה בקובץ compose ולא ב־Dockerfile.

## 19.6 סריקה ו־SBOM

### 19.6.1 מה סורק עושה

סורק פגיעויות (vulnerability scanner) לא מריץ את הקוד שלך. הוא עושה שני דברים: מוציא מה־image רשימה של
כל החבילות שבו (חבילות Debian, חבילות Python, מודולי Go שמוטמעים בבינארי), ומצליב אותה מול מאגרי פגיעויות
ציבוריים. התוצאה טובה כמו שתי הרשימות: חבילה שהסורק לא זיהה לא תיבדק, ופגיעות שלא פורסמה לא תימצא.

הרשימה עצמה נקראת SBOM (רשימת רכיבי תוכנה, software bill of materials), ויש לה פורמטים תקניים: SPDX
ו־CycloneDX. הערך שלה מתגלה אחרי תקרית. כשמתפרסמת פגיעות חדשה בספרייה נפוצה, השאלה „באילו מהשירותים
שלנו היא נמצאת?” נענית בחיפוש ברשימות שכבר שמרת, בלי לבנות ולסרוק מחדש כל image.

### 19.6.2 שלושה כלים, כקונטיינרים

| כלי | מה הוא עושה בתרגיל | image (נכון לאוקטובר 2026) |
| --- | --- | --- |
| Syft | מפיק SBOM בפורמט SPDX JSON | `anchore/syft:v1.54.0` |
| Grype | סורק את ה־SBOM ולא את ה־image | `anchore/grype:v0.120.0` |
| Trivy | סורק את ה־image ישירות, רק HIGH ו־CRITICAL | `aquasec/trivy:0.75.0` |

התגים נבדקו מול Docker Hub; הגרסאות שלהם מרוכזות בנספח ג. שני סורקים שונים לא מיותרים: הם משתמשים
במאגרים ובכללי זיהוי שונים, וההבדלים ביניהם מלמדים משהו.

הדרך הנפוצה ברשת להריץ סורק כקונטיינר היא להצמיד אליו את השקע של Docker
(`/var/run/docker.sock`). אל תעשי את זה מתוך הרגל: מי שמחזיק את השקע יכול להריץ כל קונטיינר עם כל הרשאה,
כלומר הוא root במארח. התרגיל מייצא את ה־image לקובץ, וכל כלי מקבל רק את התיקייה. זה הלב של הקובץ
`tests/scan_images.sh`:

<div dir="ltr" align="left">

```bash
  docker save "$img" -o "reports/$name.tar"

  # SBOM: the list of every package in the image, in SPDX JSON.
  docker run --rm -v "$PWD/reports:/work" "$SYFT" \
    "docker-archive:/work/$name.tar" -o "spdx-json=/work/$name.spdx.json" -q
```

</div>

<div dir="ltr" align="left">

```bash
  # Grype reads the SBOM, not the image: a scan you can repeat later without the image.
  docker run --rm -v "$PWD/reports:/work" -v ch19-grype-db:/root/.cache/grype "$GRYPE" \
    "sbom:/work/$name.spdx.json" -o table --file "/work/$name.grype.txt" -q || true
```

</div>

<div dir="ltr" align="left">

```bash
  # Trivy scans the image archive itself; only HIGH and CRITICAL, fixed or not.
  docker run --rm -v "$PWD/reports:/work" -v ch19-trivy-cache:/root/.cache/trivy "$TRIVY" \
    image --quiet --input "/work/$name.tar" --severity HIGH,CRITICAL \
    --format table --output "/work/$name.trivy.txt"
```

</div>

שני ה־volumes בשם `ch19-grype-db` ו־`ch19-trivy-cache` שומרים את מאגרי הפגיעויות בין הרצות. בלעדיהם כל
הרצה מורידה אותם מחדש.

### 19.6.3 לקרוא את התוצאה ולהחליט

ממצא הוא לא פקודה. לכל שורה בדוח שואלים שלוש שאלות:

1. **האם הקוד הפגיע בכלל רץ?** חבילה שנמצאת ב־image אבל אף תהליך לא טוען אותה היא סיכון קטן יותר,
   אבל גם משקל מת שעדיף להוציא.
2. **האם יש תיקון?** העמודה של הגרסה המתוקנת, או הסטטוס `fixed`, אומרת שיש. סטטוס כמו `will_not_fix`
   או `affected` אומר שההפצה לא תתקן, או עוד לא.
3. **מה המחיר של כל אפשרות?** לעדכן את הבסיס, להסיר את החבילה, להחליף בסיס, או לתעד ולחכות.

בהרצה שבה נכתב הספר (5 באוקטובר 2026), ה־image של shorty יצא נקי בשני הכלים: יש בו בינארי אחד ומעט קבצים של
הבסיס. ב־receiver, Trivy דיווח על שלוש פגיעויות בחבילות Python שהאפליקציה לא משתמשת בהן. החיפוש הראה שכולן
שייכות ל־pip שמגיע עם תמונת הבסיס. השירות לא מתקין חבילות בזמן ריצה, ולכן ההחלטה הייתה להסיר את pip בשלב
הריצה, ובסריקה הבאה נעלמו שלושת הממצאים. נשארו ממצאים בחבילות Debian של הבסיס, ורובם בלי תיקון זמין.
אצלך המספרים יהיו אחרים, כי המאגרים מתעדכנים כל יום; את מה שמצאת ואת מה שהחלטת תרשמי בפתק.

> **שימי לב:** בצנרת CI מקובל להכשיל בנייה רק על ממצאים שיש להם תיקון, עם הדגל `--ignore-unfixed`
> של Trivy, ולטפל בשאר בתהליך מסודר. שער שנכשל על כל ממצא בלי תיקון נכשל כל הזמן, והצוות לומד לעקוף
> אותו. פרק 23 מחבר את הסריקה ל־GitHub Actions.

## 19.7 תגיות וגרסאות

### 19.7.1 תגית מול תקציר

שם של image בנוי משלושה חלקים: registry, שם ותגית. בשם `ghcr.io/astral-sh/uv:0.12.23`, למשל, החלק
הראשון הוא ה־registry, אחריו השם, ואחרי הנקודתיים התגית. התגית (tag) היא מצביע שאפשר להזיז: מי שמחזיק ב־registry יכול לדחוף
תוכן אחר תחת אותה תגית מחר. התקציר (digest) הוא גיבוב SHA-256 של התוכן עצמו, והוא לא יכול להשתנות בלי
שהתוכן ישתנה. אפשר לראות אותו כך:

<div dir="ltr" align="left">

```bash
docker buildx imagetools inspect python:3.14.8-slim
```

</div>

<div dir="ltr" align="left">

```text
Name:      docker.io/library/python:3.14.8-slim
MediaType: application/vnd.oci.image.index.v1+json
Digest:    sha256:c3e521df8b2b498a7a682e7e18676771cb80c6b75b8699af886b2d554ce40151
```

</div>

ה־digest הזה שייך לאינדקס (index): רשימה שמצביעה על image נפרד לכל ארכיטקטורה. לכן אותו שם עובד גם
על מעבד אינטל וגם על arm64.

| כתיבה | מה מקבלים | מתי |
| --- | --- | --- |
| `python:3.14.8-slim` | מה שהתגית מצביעה עליו היום; עדכוני אבטחה של הבסיס נכנסים | פיתוח, ותרגילי הספר |
| `python:3.14.8-slim@sha256:c3e5...` | בדיוק אותם בתים, תמיד | בנייה לייצור, עם כלי שמעדכן את ה־digest בבקשת מיזוג |
| `python:latest` | מה שמישהו דחף אחרון | אף פעם, בשום מקום שחשוב |

הנעילה ב־digest קונה שחזור מלא, ומשלמת בזה שתיקוני אבטחה לא נכנסים מעצמם. בצוות שעובד כך, כלי כמו
Renovate או Dependabot פותח בקשת מיזוג כשה־digest משתנה, וכך העדכון עובר סקירה ובדיקות. פרק 20 מראה את
אותו רעיון בתצורה של kind, שבה ה־digest הוא חובה.

### 19.7.2 למה `latest` אינה תגית פריסה

התגית `latest` היא רק ברירת המחדל כשלא כותבים תגית; היא לא אומרת „האחרון”. היא לא מספרת מה רץ: אם שני
עותקים של שירות רצים עם `latest` שנמשכה בזמנים שונים, הם יכולים להריץ קוד שונה. היא לא מאפשרת חזרה
לאחור, כי „הגרסה הקודמת” של `latest` לא קיימת בשם. ו־Kubernetes מתייחס אליה אחרת: עם `latest` הוא מנסה
למשוך את ה־image בכל הפעלה (פרק 20, סעיף 20.2).

### 19.7.3 שתי תגיות לכל בנייה

הנוהג שהספר ממליץ עליו הוא שתי תגיות לאותו image: גרסה שבני אדם קוראים, ומזהה ה־commit שהמכונה קוראת. שניהם
לא זזים אף פעם:

<div dir="ltr" align="left">

```bash
SHA=$(git rev-parse --short=12 HEAD)
docker build -t shorty:0.1.0 -t "shorty:$SHA" --build-arg GIT_SHA="$SHA" shorty
docker image inspect --format '{{index .Config.Labels "org.opencontainers.image.revision"}}' shorty:0.1.0
```

</div>

התווית `org.opencontainers.image.revision` (תווית תקנית של OCI) נכנסת לתצורה של ה־image, כך שגם מי
שמחזיקה רק את ה־image יודעת מאיזה commit הוא נבנה. ב־registry אמיתי מפעילים גם תגיות שאי אפשר לדרוס
(immutable tags), אפשרות שקיימת ברוב שירותי ה־registry המנוהלים: ניסיון לדחוף שוב את `0.1.0` נכשל, במקום
להחליף בשקט את מה שרץ בייצור.

## סיכום

- image הוא שכבות ומסמך תצורה; כל `COPY` ו־`RUN` הוא שכבה, והחמצה אחת במטמון בונה מחדש את כל מה שאחריה.
- סדר: בסיס, קובצי תלויות, הורדת תלויות, קוד, קומפילציה; ו־`.dockerignore` שומר על המטמון ועל הסודות.
- בנייה רב־שלבית ל־Go: קומפיילר בשלב אחד, בינארי סטטי על distroless בשלב השני.
- משתמש מספרי שאינו root, מערכת קבצים לקריאה בלבד, ובדיקה מבחוץ עם `docker image inspect` ו־`docker top`.
- ב־Python: uv מתקין בשלב הבנייה סביבה שלמה ולא ניתנת לעריכה, ושלב הריצה מעתיק רק אותה.
- compose לפיתוח: בדיקת בריאות שרצה בתוך ה־image, ו־`depends_on` עם `service_healthy`.
- SBOM נשמר לכל image; סורקים מצליבים אותו מול מאגרים, וכל ממצא מקבל החלטה כתובה.
- תגיות לא זזות: גרסה ומזהה commit, ו־digest כשצריך את אותם בתים בדיוק. לא `latest`.

## מונחים לזכור

| מונח בעברית | באנגלית | במשפט אחד |
| --- | --- | --- |
| image | image | שכבות של מערכת קבצים ותצורה, שמהן מריצים קונטיינר |
| שכבה | layer | ארכיון של שינויים בקבצים שהוראה אחת ב־Dockerfile יצרה |
| מטמון בנייה | build cache | שכבות שמורות שנלקחות מחדש כשההוראה והקלט זהים |
| הקשר בנייה | build context | התיקייה שממנה `COPY` מעתיקה, בסינון של `.dockerignore` |
| בנייה רב־שלבית | multi-stage build | שלב עם כלים לבנייה, ושלב נפרד ונקי לריצה |
| מטמון מוצמד | cache mount | תיקייה שנשמרת בין בניות ולא נכנסת ל־image |
| distroless | distroless | בסיס בלי מעטפת ובלי מנהל חבילות, רק מה שהתוכנית צריכה |
| בדיקת בריאות | healthcheck | פקודה שרצה בקונטיינר וקובעת אם הוא בריא |
| סורק פגיעויות | vulnerability scanner | כלי שמצליב את חבילות ה־image מול מאגרי פגיעויות |
| רשימת רכיבי תוכנה | SBOM | רשימה תקנית של כל החבילות ב־image |
| תגית | tag | שם שמצביע על image, ואפשר להזיז אותו |
| תקציר | digest | גיבוב של התוכן, שמזהה בדיוק אותם בתים |
| תגית שאי אפשר לדרוס | immutable tag | הגדרה ב־registry שמונעת דחיפה חוזרת לאותה תגית |

## טעויות נפוצות

- **להעתיק את כל הקוד לפני הורדת התלויות.** כל שינוי בשורה אחת מוריד מחדש את כל המודולים. מעתיקים קודם
  את קובצי התלויות בלבד.
- **למחוק סוד בשכבה שאחרי זו שהעתיקה אותו.** הוא נשאר בשכבה הקודמת. סודות לא נכנסים ל־image בכלל.
- **צורת מעטפת ב־`ENTRYPOINT`.** האות `SIGTERM` מגיע למעטפת ולא לשירות, ואין כיבוי מסודר. וב־distroless
  זה פשוט לא רץ, כי אין מעטפת.
- **שרת שמאזין ל־`127.0.0.1` בתוך הקונטיינר.** הוא עולה, אבל אף אחד מבחוץ לא מגיע אליו. מאזינים לכל
  הכתובות בקונטיינר, ומגבילים בצד של המארח.
- **`USER nonroot` בשם ולא במספר.** Kubernetes עם `runAsNonRoot` לא יכול לוודא שם, ומסרב להפעיל.
- **להצמיד את השקע של Docker לכל כלי.** זו הרשאת root במארח. מייצאים את ה־image לקובץ.
- **לפרוס עם `latest`.** אין דרך לדעת מה רץ, ואין לאן לחזור.

## שאלות ראיון

1. **איך תבני image לשירות Go, ולמה ככה?** בנייה רב־שלבית: שלב עם תמונת `golang` שמעתיק קודם את קובצי
   התלויות ומוריד אותם, ורק אחר כך את הקוד, כדי שהמטמון יחזיק. קומפילציה סטטית עם cgo כבוי, `-trimpath`
   ובלי סמלים, ומטמון מוצמד לקומפילציה. שלב ריצה על distroless עם משתמש מספרי שאינו root. התמורה היא
   image קטן עם מעט מאוד לסרוק ולתקוף, והמחיר הוא ניפוי בלי מעטפת, שמתוכנן מראש עם לוגים וקונטיינר ניפוי.

2. **מה distroless חוסך, ומה הוא עולה לך?** הוא חוסך מעטפת, מנהל חבילות וכלים: פחות חבילות לסרוק, פחות
   כלים לתוקף שהשיג הרצת קוד, ו־image קטן. הוא עולה בניפוי: אין `exec` לתוך הקונטיינר, ובדיקת בריאות חייבת
   להיות בינארי שב־image. מתכננים לזה מראש: לוגים ומדדים טובים, `kubectl debug` עם קונטיינר זמני, ותג
   `debug` רק מקומית.

   **המשך:** ואם השירות צריך ספריית C? אז הבסיס הוא `distroless/base` או `cc`, שיש בהם ספריית C, או
   Debian מוקטן. הכלל לא משתנה: הבסיס הקטן ביותר שהבינארי באמת צריך.

3. **הסורק מצא 40 פגיעויות ב־image. מה עושים?** קודם ממיינים: מה בחומרה גבוהה ויש לו תיקון, ומה בכלל
   נמצא בנתיב שרץ. בדרך כלל רוב הממצאים בחבילות של הבסיס שאין להן תיקון. אפשרויות: לעדכן את הבסיס, להסיר
   חבילה שלא צריך (כמו pip בשלב ריצה), לעבור לבסיס קטן יותר, או לתעד החרגה עם תאריך. שער CI נכשל על
   ממצאים שיש להם תיקון, אחרת הוא נכשל תמיד וכולם לומדים לעקוף אותו.

4. **למה לשמור SBOM לכל image אם כבר סורקים?** הסריקה היא תמונת מצב של היום; פגיעויות חדשות מתפרסמות
   מחר. עם SBOM שמור, כשמתפרסמת פגיעות בספרייה נפוצה, אפשר לענות בדקות באילו שירותים וגרסאות היא נמצאת,
   ולסרוק מחדש את הרשימה בלי ה־image. זה גם מה שלקוחות ורגולציה מתחילים לבקש.

5. **מה ההבדל בין תגית ל־digest, ומה תפרסי לייצור?** תגית היא מצביע שאפשר להזיז; digest מזהה תוכן. לייצור
   הייתי בונה עם שתי תגיות שלא זזות, גרסה ומזהה commit, ומפעילה תגיות שאי אפשר לדרוס ב־registry. בצנרת
   הפריסה הייתי מעבירה את ה־digest שהבנייה החזירה, כך שמה שנבדק הוא מה שרץ. ואת תמונות הבסיס הייתי נועלת
   ב־digest עם כלי שמעדכן אותן בבקשות מיזוג.

   **המשך:** מה הבעיה עם `latest` ב־Kubernetes דווקא? מדיניות המשיכה המשתמעת היא „תמיד”, ולכן כל
   הפעלה של pod מושכת מחדש, ושני pods יכולים להריץ גרסאות שונות בלי שאף אחד יודע.

## פתק Obsidian

הפתק המוכן נמצא גם כקובץ: `obsidian/19-docker-done-right.md`. זה תוכנו:

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

## תרגיל

**מטרה:** לבנות images קטנים שרצים כמשתמש שאינו root ל־shorty מפרק 6 ול־receiver מפרק 8, להרים אותם ב־compose
עם בדיקות בריאות, לסרוק אותם, להפיק SBOM ולתייג לפי גרסה ו־commit.
**זמן:** שני ימי לימוד, ימים 58 ו־59, כ־60 דקות בנייה בכל אחד.
**ידע מקדים:** הפרק; פרק 6 ופרק 8 (התרגיל מביא עותקים של שני השירותים); Docker פועל.

הצעדים, קבצי ההתחלה, הרמזים והפתרון נמצאים ב־
[`exercises/ch19-images/README.md`](../exercises/ch19-images/README.md).

**קריטריון סיום:**

- שני ה־images נבנים, רצים כמשתמש שאינו root, ועונים לבקשות גם עם מערכת קבצים לקריאה בלבד.
- בטבלה בפתק: הגודל שמדדת ל־shorty, לבנייה הנאיבית ול־receiver.
- הסקריפט `check_images.sh` מסתיים ב־`PASS`, כולל compose עם שני שירותים בריאים ושליחה דרך שם השירות.
- קובצי SBOM ודוחות סריקה שמורים, ובפתק כתוב ממצא אחד ומה החלטת לגביו.
- שני ה־images מתויגים בגרסה ובמזהה commit, והתווית `revision` מראה את ה־commit.

## העשרה אופציונלית

- לקרוא את מדריך השילוב של uv עם Docker באתר docs.astral.sh, ובמיוחד את הסעיף על סביבת פיתוח עם
  `compose watch`.
- לבנות את shorty לשתי ארכיטקטורות בבת אחת עם `docker buildx build --platform linux/amd64,linux/arm64`,
  ולהשוות את ה־digest של האינדקס לשל כל image בנפרד.
- להריץ את Trivy עם `--ignore-unfixed` ועם `--format cyclonedx`, ולהשוות את מספר הממצאים ואת הפורמט
  ל־SPDX של Syft.

</div>
