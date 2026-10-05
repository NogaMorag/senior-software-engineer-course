---
direction: rtl
lang: he
title: "פרק 23: GitHub Actions"
chapter: 23
part: "חלק ה: Cloud-native"
---

<div dir="rtl" lang="he">

# פרק 23: GitHub Actions

אינטגרציה רציפה (CI) היא המקום שבו „עובד אצלי” הופך לטענה שכל אחד יכול לבדוק. היא רצה על כל שינוי, באותה
סביבה, עם אותן פקודות, ומחליטה מה מותר למזג ומה מותר לפרסם. בראיון בכיר שואלים עליה פחות על התחביר ויותר על
ההחלטות: מה רץ על כל PR, מי מחזיק טוקן שיכול לכתוב, ואיך יודעים שירוק באמת בדק משהו.

בפרק תכתבי את ה־workflows של שלושה פרויקטים מהספר, במאגר העבודה שלך: שירות קיצור הכתובות מפרק 6, מקבל
ההודעות מפרק 8, ומעבדת הבידוד מפרק 11 עם מסד אמיתי. אחר כך תבני image ותבדקי אותו על אשכול kind בתוך ה־CI.
בסוף תפרסמי גרסה (release) מאחורי אישור. הקבצים האלה עוברים לפרויקט הגמר, Hookline, באבן הדרך M7.

## מה נלמד בפרק

- ארבע היחידות של GitHub Actions, ומה רץ איפה.
- בדיקות ל־Go ול־Python, עם גרסאות שנקראות מהפרויקט ומטמון שחוסך דקות.
- ‏service container ל־PostgreSQL, והדרך לוודא שהבדיקות לא דילגו עליו בשקט.
- בניית image ופרסום ל־GHCR רק מ־`main`, עם ההרשאה המינימלית שזה דורש.
- בדיקת עשן על אשכול kind זמני, לפני שה־image מתפרסם.
- ‏environments, סודות, משתנים, ההרשאות של `GITHUB_TOKEN`, ו־OIDC לענן במקום מפתחות.
- גרסה על תגית, הצמדה של actions לגרסה ראשית או ל־SHA, ובדיקה סטטית עם actionlint.

## ידע מקדים

- פרק 6: ‏`shorty`, ה־Makefile שלו (סעיף 6.7), ו־`/healthz`.
- פרק 8: המקבל כפרויקט uv, עם שלושת כלי הבדיקה שלו. פרק 7, סעיף 7.2: קובץ הנעילה ו־`uv sync`.
- פרק 11: המעבדה, והבדיקות שקוראות את `DATABASE_URL` ומדלגות בלעדיו.
- פרק 19: ה־Dockerfile הרב־שלבי, עם משתמש שאינו root (סעיפים 19.2 ו־19.3), ותגיות של images (סעיף 19.7).
- פרק 20: קובצי הפריסה של Kubernetes, בדיקות המוכנות, וטעינת image לאשכול (סעיפים 20.2 עד 20.7).
- פרק 2, סעיף 2.5: חשבון GitHub ו־Git.
- פרק 2, סעיף 2.7: actionlint. סעיף 2.10: kind ו־kubectl.

## 23.1 workflow, jobs ו־steps

‏GitHub Actions היא מערכת ה־CI של GitHub. היחידות שלה, לפי הגודל:

| יחידה | מה היא | איפה מוגדרת |
| --- | --- | --- |
| workflow | קובץ YAML אחד, עם טריגרים וקבוצת jobs | ‏`.github/workflows/*.yml` בשורש המאגר |
| job | רצף צעדים על מכונה אחת; jobs רצים במקביל אלא אם יש `needs` | המפתח `jobs` |
| step | פקודת shell (‏`run`) או שימוש ב־action (‏`uses`) | הרשימה `steps` של ה־job |
| runner | המכונה שמריצה job; ‏`ubuntu-latest` היא מכונה וירטואלית חדשה לכל job | ‏`runs-on` |
| action | יחידה לשימוש חוזר: קוד JavaScript, קונטיינר, או צעדים מורכבים | מאגר אחר, או תיקייה במאגר שלך |

הדבר הראשון להפנים: **כל job מתחיל ממכונה נקייה.** אין קבצים מה־job הקודם, אין מטמון, ואין את הקוד שלך עד
שמריצים `actions/checkout`. מה שעובר בין jobs עובר במפורש: דרך `outputs`, דרך מטמון, או דרך registry.

### 23.1.1 הראש של workflow

כך מתחיל ה־workflow של `shorty` (הקובץ `.github/workflows/ch06-shorty.yml` בפתרון):

<div dir="ltr" align="left">

```yaml
name: ch06 shorty

on:
  push:
    branches: [main]
    paths:
      - "ch06/**"
      - ".github/actions/go-checks/**"
      - ".github/workflows/ch06-shorty.yml"
  pull_request:
    paths:
      - "ch06/**"
      - ".github/actions/go-checks/**"
      - ".github/workflows/ch06-shorty.yml"

# Every job gets a read-only token unless it asks for more.
permissions:
  contents: read

# One run per branch or PR; a new push to a PR cancels the run it replaces.
concurrency:
  group: ${{ github.workflow }}-${{ github.ref }}
  cancel-in-progress: ${{ github.event_name == 'pull_request' }}

defaults:
  run:
    shell: bash
    working-directory: ch06
```

</div>

ארבעה מפתחות, ולכל אחד סיבה:

| מפתח | מה הוא קובע | למה כך |
| --- | --- | --- |
| `on` | אילו אירועים מריצים את ה־workflow | ‏push ל־`main` ו־PRs, ורק כשהשתנה משהו בתיקייה של הפרויקט |
| `permissions` | מה הטוקן האוטומטי של ההרצה רשאי לעשות | קריאה בלבד, וכל job שצריך יותר מבקש בעצמו (סעיף 23.6) |
| `concurrency` | כמה הרצות של אותה קבוצה רצות יחד | ‏push חדש ל־PR מבטל את ההרצה הישנה; ב־`main` כל הרצה מסיימת |
| `defaults.run` | ה־shell והתיקייה של כל צעד `run` | ‏`bash` במפורש מוסיף `pipefail`, והתיקייה היא `ch06` |

**סינון לפי נתיב.** במאגר אחד עם כמה פרויקטים, `paths` הוא מה שמונע מכל שינוי להריץ הכול. שימי לב שהרשימה
כוללת גם את הקובץ של ה־workflow עצמו ואת הפעולה המשותפת: שינוי בהם חייב להיבדק. יש לסינון הזה מחיר שכדאי
לדעת: workflow שלא רץ בגלל סינון משאיר את הבדיקות שלו במצב „ממתין”, ו־PR שדורש אותן ייחסם. כך כתוב בתיעוד
התחביר של GitHub, ולכן בדיקה חובה (required check) לא מצמידים ל־workflow מסונן.

**ה־shell.** בלי מפתח מפורש, צעד `run` רץ כ־`bash -e`. עם `shell: bash` הוא רץ עם דגלים נוספים, כך:

<div dir="ltr" align="left">

```text
bash --noprofile --norc -eo pipefail {0}
```

</div>

ההבדל הוא `pipefail`: בלעדיו, בדיקות שהפלט שלהן עובר בצינור ל־`tee` מצליחות גם כשהן נכשלות, כי קוד
היציאה הוא של `tee`. הפעולה המורכבת של סעיף 23.2 נשענת על זה.

### 23.1.2 גרף ה־jobs

ל־`shorty` יש שלושה jobs. ‏`needs` הופך אותם לשרשרת:

<div dir="ltr" align="left">

```mermaid
flowchart LR
  T[test] --> K[kind-smoke]
  T --> I[image]
  K --> I
  I -->|push to main only| G[(GHCR)]
```

</div>

ה־job בשם `test` מריץ את הבדיקות. ‏`kind-smoke` בונה image ומריץ אותו על Kubernetes. ‏`image` מפרסם, רק
מ־`main` ורק אחרי ששניהם עברו. בכל PR רצים שני הראשונים; ה־image לא מתפרסם מ־PR אף פעם.

### 23.1.3 ביטויים והקשרים

הצורה `${{ ... }}` היא ביטוי (expression) שמוערך לפני שהצעד רץ. הוא קורא מהקשרים (contexts): האירוע והענף,
הסודות והמשתנים, הקלטים, והתוצאות של צעדים ו־jobs קודמים. התוצאה מודבקת לטקסט של הקובץ. זה נוח, וזה גם מקור לבעיית אבטחה אמיתית, כי ביטוי בתוך `run` הופך לחלק מהסקריפט עצמו.
נחזור לזה בסעיף 23.6.

## 23.2 בנייה ובדיקה ל־Go ול־Python

### 23.2.1 פעולה מורכבת לבדיקות של Go

לשני פרויקטי ה־Go יש אותם ארבעה צעדים: עיצוב, ניתוח סטטי בשני כלים, ובדיקות עם גלאי המרוצים. במקום להעתיק
אותם, הם כתובים פעם אחת כפעולה מורכבת (composite action), בתיקייה `.github/actions/go-checks`:

<div dir="ltr" align="left">

```yaml
name: Go checks
description: gofmt, go vet, staticcheck and go test -race for one Go module.

inputs:
  working-directory:
    description: Directory that holds go.mod, relative to the repository root.
    required: true
  fail-on-skip:
    description: Fail when any test was skipped (use it where a database is expected).
    required: false
    default: "false"

runs:
  using: composite
  steps:
    - name: Set up Go from go.mod
      uses: actions/setup-go@v7
      with:
        go-version-file: ${{ inputs.working-directory }}/go.mod
        cache-dependency-path: ${{ inputs.working-directory }}/go.sum

    - name: gofmt
      shell: bash
      working-directory: ${{ inputs.working-directory }}
      run: |
        unformatted="$(gofmt -l .)"
        if [ -n "$unformatted" ]; then
          echo "::error::gofmt would change these files:"
          echo "$unformatted"
          exit 1
        fi

    - name: go vet
      shell: bash
      working-directory: ${{ inputs.working-directory }}
      run: go vet ./...

    - name: staticcheck
      shell: bash
      working-directory: ${{ inputs.working-directory }}
      run: go run honnef.co/go/tools/cmd/staticcheck@2026.2.1 ./...
```

</div>

כמה החלטות קטנות שכדאי לשים לב אליהן:

- **הגרסה של Go מגיעה מ־`go.mod`.** הקלט `go-version-file` קורא אותה מהקובץ, כך שהפרויקט, המחשב שלך
  וה־CI מסכימים בלי לכתוב אותה פעמיים.
- **המטמון כבר דלוק.** ‏`setup-go` שומר את מטמון המודולים ואת מטמון הבנייה, במפתח שנגזר מ־`go.sum`. בפרויקט
  שבו `go.sum` לא בשורש, צריך `cache-dependency-path`; אחרת המטמון לא יימצא.
- **‏staticcheck בגרסה מוצמדת, דרך `go run`.** הצורה `module@version` בונה את הכלי בלי לגעת בקובץ המודול של
  הפרויקט, והבנייה עצמה נכנסת למטמון. נכון לאוקטובר 2026, זו הגרסה של נספח ג.
- **‏`::error::`** היא פקודת workflow: שורה שמתחילה בה מופיעה כהערה אדומה בדף ההרצה, ולא רק בלוג.
- **בפעולה מורכבת, כל `run` חייב `shell`.** אין לה `defaults`, ולכן כתוב `shell: bash` בכל צעד.

והצעד האחרון, הבדיקות. הדגל `-race` מפעיל את גלאי מרוצי הנתונים מפרק 5. הדגל `-count=1` מבטל את מטמון
התוצאות, כי ב־CI רוצים הרצה אמיתית:

<div dir="ltr" align="left">

```yaml
    - name: go test -race
      shell: bash
      working-directory: ${{ inputs.working-directory }}
      env:
        FAIL_ON_SKIP: ${{ inputs.fail-on-skip }}
      run: |
        go test -race -count=1 -v ./... | tee "$RUNNER_TEMP/go-test.log"
        if [ "$FAIL_ON_SKIP" = "true" ] && grep -q -- '--- SKIP' "$RUNNER_TEMP/go-test.log"; then
          echo "::error::tests were skipped; is DATABASE_URL set for this job?"
          grep -- '--- SKIP' "$RUNNER_TEMP/go-test.log"
          exit 1
        fi
```

</div>

הקלט `fail-on-skip` עובר לסקריפט דרך משתנה סביבה ולא דרך `${{ }}` בתוך הטקסט. למה זה חשוב יופיע בסעיף 23.6,
ולמה הקלט קיים בכלל יופיע בסעיף 23.3. ה־job ב־workflow של `shorty` רק קורא לפעולה:

<div dir="ltr" align="left">

```yaml
jobs:
  test:
    runs-on: ubuntu-latest
    timeout-minutes: 15
    steps:
      - uses: actions/checkout@v7
        with:
          persist-credentials: false
      - uses: ./.github/actions/go-checks
        with:
          working-directory: ch06
```

</div>

שני פרטים. ‏`timeout-minutes` קיים בכל job, כי ברירת המחדל של GitHub ארוכה בהרבה ממה שבדיקה תקועה שווה.
ו־`persist-credentials: false` מונע מ־checkout להשאיר את הטוקן בהגדרות Git של המכונה, שם כל צעד אחר היה יכול
לקרוא אותו.

**פעולה מורכבת או workflow לשימוש חוזר?** ל־GitHub יש שני כלים לשימוש חוזר, והם עונים על שאלות שונות:

| | פעולה מורכבת | workflow לשימוש חוזר (reusable workflow) |
| --- | --- | --- |
| מה היא | רצף צעדים בתוך job של מי שקורא לה | ‏workflow שלם, עם jobs משלו, שנקרא דרך `workflow_call` |
| איפה רצה | על ה־runner של ה־job הקורא | על runners משלה |
| ‏services, ‏`permissions`, ‏environment | של ה־job הקורא | מגדירה בעצמה |
| מתי לבחור | „אותם צעדים בכמה jobs” | „אותו תהליך שלם בכמה מאגרים”, למשל release אחיד לארגון |

כאן הבחירה ברורה: המעבדה של פרק 11 צריכה מסד לצד ה־job, והפעולה המורכבת רצה בתוך אותו job ומשתמשת בו.

### 23.2.2 ה־job של Python

ל־Python, ‏uv עושה את רוב העבודה (הקובץ `.github/workflows/ch08-receiver.yml`):

<div dir="ltr" align="left">

```yaml
    steps:
      - uses: actions/checkout@v7
        with:
          persist-credentials: false
      - name: Set up uv
        uses: astral-sh/setup-uv@v10
        with:
          version: "0.12.23"
          working-directory: ch08
          enable-cache: true
          cache-dependency-glob: uv.lock
      - name: Lockfile matches pyproject.toml
        run: uv lock --check
      - name: Install exactly what uv.lock says
        run: uv sync --frozen
      - name: ruff check
        run: uv run --no-sync ruff check .
      - name: ruff format
        run: uv run --no-sync ruff format --check .
      - name: mypy
        run: uv run --no-sync mypy src
      - name: pytest
        run: uv run --no-sync pytest -q
```

</div>

שלושה דגלים של uv, ושלושה תפקידים שונים:

| פקודה | מה היא מבטיחה |
| --- | --- |
| `uv lock --check` | שקובץ הנעילה תואם ל־`pyproject.toml`; תלות שנוספה בלי `uv lock` נכשלת כאן |
| `uv sync --frozen` | התקנה מ־`uv.lock` כמו שהוא, בלי לפתור תלויות מחדש |
| `uv run --no-sync` | הרצה בסביבה שכבר הותקנה, בלי ש־`uv run` יסנכרן בעצמו |

את Python עצמו לא מתקינים בצעד נפרד: `uv sync` קורא את `.python-version` ומוריד את הגרסה המנוהלת. הגרסה של
uv מוצמדת, כמו כל כלי אחר בספר. ולכל בדיקה צעד משלה, כדי שבדף ההרצה יופיע בדיוק מה נכשל: ruff או mypy, ולא
„הבדיקות”.

## 23.3 service containers

מעבדת הבידוד מפרק 11 צריכה PostgreSQL אמיתי, כי כל מה שהיא בודקת קורה בתוך המסד. מסד כזה נקרא service
container: קונטיינר שמופעל לפני הצעדים של ה־job ונכבה אחריהם. הנה ההגדרה, מהקובץ
`.github/workflows/ch11-isolation-lab.yml`:

<div dir="ltr" align="left">

```yaml
jobs:
  test:
    runs-on: ubuntu-latest
    timeout-minutes: 15
    services:
      postgres:
        image: postgres:18.6
        env:
          POSTGRES_USER: course
          POSTGRES_PASSWORD: change-me # a throw-away CI database, not a secret
          POSTGRES_DB: lab
        ports:
          - 5432:5432
        # The job's steps start only after this health check passes.
        options: >-
          --health-cmd "pg_isready -U course -d lab"
          --health-interval 2s
          --health-timeout 3s
          --health-retries 30
    env:
      # The job runs on the runner itself, so the service is on localhost.
      DATABASE_URL: postgres://course:change-me@localhost:5432/lab
    steps:
      - uses: actions/checkout@v7
        with:
          persist-credentials: false
      - uses: ./.github/actions/go-checks
        with:
          working-directory: ch11
          fail-on-skip: "true"
```

</div>

**הבריאות.** ה־runner מחכה עד שבדיקת הבריאות של הקונטיינר עוברת, ורק אז מתחיל את ה־steps. בלי `--health-cmd`
הוא מתחיל מיד, והבדיקה הראשונה פוגשת מסד שעוד מאתחל: `connection refused`, לפעמים. „לפעמים” הוא הסוג הגרוע
ביותר של כשל ב־CI. הערכים זהים לאלה של קובץ ה־compose של פרק 11, כדי שהמסד ב־CI יתנהג כמו המסד אצלך.

**הכתובת.** איך מגיעים לשירות תלוי באיפה ה־job רץ:

| ה־job רץ | הכתובת של השירות | מה צריך |
| --- | --- | --- |
| על ה־runner עצמו (כמו כאן) | `localhost` והפורט שמופה | `ports` |
| בתוך קונטיינר (`container:`) | שם השירות, כאן `postgres`, והפורט המקורי | רשת שה־runner יוצר; בלי `ports` |

**הסיסמה בקובץ.** ‏`change-me` כתובה בגלוי, וזה בסדר: זה מסד שחי כמה דקות, על מכונה שנמחקת, ואין בו כלום. סוד
אמיתי לעולם לא נכתב בקובץ, וגם לא „רק לבדיקה”; סעיף 23.6.

### 23.3.1 ירוק שלא בדק כלום

הבדיקות של פרק 11 מדלגות כש־`DATABASE_URL` לא מוגדר. זה נכון על המחשב שלך, וזה מסוכן ב־CI. כך זה נראה, מהרצה
אמיתית על הפתרון בלי המשתנה:

<div dir="ltr" align="left">

```bash
cd ~/senior-course/ch11
unset DATABASE_URL
go test -count=1 ./...
go test -count=1 -v ./... | grep -c -- '--- SKIP'
```

</div>

<div dir="ltr" align="left">

```text
ok  	example.com/ch11-isolation-lab/bank	0.693s
ok  	example.com/ch11-isolation-lab/db	0.378s
?   	example.com/ch11-isolation-lab/internal/testdb	[no test files]
ok  	example.com/ch11-isolation-lab/migrations	1.009s
ok  	example.com/ch11-isolation-lab/oncall	1.343s
ok  	example.com/ch11-isolation-lab/queue	1.646s
20
```

</div>

כל החבילות `ok`, ועשרים בדיקות דילגו. מספיק ששם המשתנה ישתנה, או שמישהו יזיז את `env` מה־job לצעד אחר,
וה־CI ימשיך להיות ירוק בלי לבדוק את המסד. בשביל זה הקלט `fail-on-skip`: ב־job שבו המסד אמור להיות, דילוג
הוא כישלון. הכלל הכללי: **‏CI צריך להיכשל כשהתנאים לבדיקה לא מתקיימים**, ולא לדווח הצלחה על מה שלא רץ.

## 23.4 בניית image ופרסום ל־GHCR

‏GHCR, ה־registry של GitHub, מחובר לחשבון ולמאגר שלך. הכתובת שלו היא `ghcr.io`, ואפשר לדחוף אליו עם הטוקן
האוטומטי של ההרצה, בלי סיסמה נוספת. זה ה־job שמפרסם את `shorty`:

<div dir="ltr" align="left">

```yaml
  image:
    # Publish only what passed the tests and the smoke test, and only from main.
    if: github.event_name == 'push' && github.ref == 'refs/heads/main'
    needs: [test, kind-smoke]
    runs-on: ubuntu-latest
    timeout-minutes: 20
    permissions:
      contents: read
      packages: write
    steps:
      - uses: actions/checkout@v7
        with:
          persist-credentials: false
      - uses: docker/setup-buildx-action@v4
      - name: Image names and labels
        id: meta
        uses: docker/metadata-action@v6
        with:
          images: ghcr.io/${{ github.repository_owner }}/shorty
          tags: |
            type=ref,event=branch
            type=sha
      - name: Log in to GHCR
        uses: docker/login-action@v4
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      - name: Build and push
        uses: docker/build-push-action@v7
        with:
          context: ch06
          push: true
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
          cache-from: type=gha,scope=shorty
          cache-to: type=gha,scope=shorty,mode=max
```

</div>

ארבעה actions, וכל אחד עושה דבר אחד:

| action | התפקיד |
| --- | --- |
| `docker/setup-buildx-action` | מקים builder של BuildKit, שתומך במטמון של GitHub Actions |
| `docker/metadata-action` | מחשב תגיות ותוויות OCI מהאירוע: כאן `main` ו־`sha-<7 תווים>` |
| `docker/login-action` | מתחבר ל־`ghcr.io` עם `GITHUB_TOKEN` |
| `docker/build-push-action` | בונה מה־Dockerfile של פרק 19 ודוחף |

**התגיות.** כל image מקבל שתי תגיות: `main`, שזזה עם הענף, ו־`sha-...`, שלא זזה לעולם. זה הכלל מסעיף 19.7:
תגית שזזה נוחה לבני אדם, ותגית קבועה היא מה שפורסים. התוויות שהפעולה מוסיפה כוללות את כתובת המאגר ואת
ה־commit, ולכן ה־Dockerfile שבתרגיל כבר לא מגדיר תוויות משלו. הפעולה גם ממירה את שם ה־image לאותיות
קטנות: שם של image לא יכול להכיל אותיות גדולות, ושם משתמש ב־GitHub יכול.

**ההרשאה.** ‏`packages: write` מוגדר רק ב־job הזה. כל שאר ה־jobs מקבלים `contents: read` מהרמה העליונה, ולא
יכולים לדחוף. וה־job כולו לא רץ ב־PR, ולכן קוד מ־PR לא מגיע לטוקן שכותב.

**המטמון.** ‏`type=gha` שומר שכבות של BuildKit במטמון של GitHub Actions. ‏`scope=shorty` מפריד אותו מ־images
אחרים במאגר, ו־`mode=max` שומר גם את שכבות שלב הבנייה, ולא רק את התוצאה. ה־job של סעיף 23.5 כותב לאותו
מטמון, ולכן כשה־job הזה רץ, רוב השכבות כבר שם.

**בנייה שנייה, לא אותם בתים.** שימי לב שה־image נבנה פעמיים: פעם ב־`kind-smoke`, ופעם כאן. עם אותו Dockerfile
ואותו מטמון, התוצאה כמעט תמיד זהה, אבל זו לא הבטחה. אם הדרישה היא „פורסים בדיוק את מה שנבדק”, בונים פעם אחת,
דוחפים לפי digest, ובודקים את ה־digest הזה. כאן הבחירה היא פשטות, ובפרויקט הגמר תחליטי בעצמך.

## 23.5 פריסה ל־kind ב־CI

בדיקות יחידה לא יודעות אם ה־image עולה בכלל. משתמש בלי הרשאה לקובץ, משתנה חובה שחסר בקובץ הפריסה, בדיקת
מוכנות לנתיב לא נכון: כל אלה עוברות אותן. אשכול kind שקם בתוך ה־job עונה על זה בכמה דקות, ונמחק עם המכונה:

<div dir="ltr" align="left">

```yaml
  kind-smoke:
    needs: test
    runs-on: ubuntu-latest
    timeout-minutes: 20
    steps:
      - uses: actions/checkout@v7
        with:
          persist-credentials: false
      - uses: docker/setup-buildx-action@v4
      - name: Build the image into the local Docker
        uses: docker/build-push-action@v7
        with:
          context: ch06
          tags: shorty:ci
          load: true
          cache-from: type=gha,scope=shorty
          cache-to: type=gha,scope=shorty,mode=max
      - name: Create a kind cluster
        uses: helm/kind-action@v1.15.0
        with:
          version: v0.33.0
          node_image: kindest/node:v1.37.0@sha256:a1ed56cfb0e7b93589bdf97c8cd566405a265939e3620fc4f5de89adff580ae5
          kubectl_version: v1.37.1
          cluster_name: ch23
      - name: Load the image into the cluster
        run: kind load docker-image shorty:ci --name ch23
      - name: Deploy
        run: |
          kubectl apply -f deploy/shorty.yaml
          kubectl rollout status deployment/shorty --timeout=120s
      - name: Smoke test
        run: ./deploy/smoke.sh
      - name: Show what the cluster saw
        if: failure()
        run: |
          kubectl get pods -o wide
          kubectl describe deployment/shorty
          kubectl logs deployment/shorty --tail=100 || true
          kubectl get events --sort-by=.lastTimestamp
```

</div>

הצעדים, לפי הסדר:

| צעד | מה קורה | הפרט שחשוב |
| --- | --- | --- |
| בנייה | ה־image נבנה ונטען ל־Docker של ה־runner | ‏`load: true`; בלעדיו הוא נשאר בתוך ה־builder |
| אשכול | ‏kind מותקן ואשכול בשם `ch23` קם | גרסאות kind, ה־node ו־kubectl מוצמדות, ה־node לפי digest |
| טעינה | ה־image מועתק לתוך ה־node | בלי registry; ה־manifest אומר `imagePullPolicy: Never` |
| פריסה | ‏Deployment ו־Service מפרק 20 | ‏`rollout status` מחכה ל־readiness probe |
| עשן | ‏port-forward, יצירת קישור, ‏302 | הסקריפט `deploy/smoke.sh` |
| אבחון | רק אם משהו נכשל | ‏`if: failure()`; בלי זה, אשכול שנמחק לא משאיר עקבות |

ה־action של kind מוצמד כאן לתגית מלאה, ולא רק לגרסה הראשית, כי ברירות המחדל שלו יכולות להשתנות בין גרסאות
משנה, ובדיקת עשן צריכה סביבה קבועה. מאותה סיבה שלוש הגרסאות כתובות גם במפורש: kind, ה־node ו־kubectl. התגית
ושמות הקלטים נבדקו מול קובץ ה־`action.yml` של התגית, ב־5 באוקטובר 2026.

בדיקת העשן עצמה היא סקריפט bash, כדי שאותו קוד ירוץ ב־CI ועל המחשב שלך (הקובץ `ch06/deploy/smoke.sh`):

<div dir="ltr" align="left">

```bash
kubectl port-forward service/shorty "${port}:80" >"${log}" 2>&1 &
forward_pid=$!
trap 'kill "${forward_pid}" 2>/dev/null || true; rm -f "${log}"' EXIT

# port-forward needs a moment before it accepts connections.
for _ in $(seq 1 30); do
  if curl -fsS "${base}/healthz" >/dev/null 2>&1; then
    break
  fi
  sleep 1
done
curl -fsS "${base}/healthz" >/dev/null || { cat "${log}"; exit 1; }
echo "healthz: ok"

created="$(curl -fsS -X POST "${base}/v1/links" \
  -H 'Content-Type: application/json' \
  -d '{"url": "https://example.com/from-ci", "code": "ci-smoke"}')"
echo "created: ${created}"

status="$(curl -s -o /dev/null -w '%{http_code}' "${base}/ci-smoke")"
if [ "${status}" != "302" ]; then
  echo "expected 302 from /ci-smoke, got ${status}" >&2
  exit 1
fi
```

</div>

הבדיקה יוצרת קישור עם קוד קבוע, `ci-smoke`, כדי שלא תצטרך לפענח JSON, ובודקת את מה שמשתמשת הייתה רואה: הפניה.
היא לא בודקת את כל ה־API; זה תפקיד הבדיקות של פרק 6. בדיקת עשן עונה על שאלה אחת: „השירות עלה ועונה?”. ה־`trap`
סוגר את ה־port-forward גם כשהבדיקה נכשלת באמצע, אחרת התהליך נשאר ברקע.

## 23.6 environments, סודות והרשאות

### 23.6.1 ‏`GITHUB_TOKEN` ומינימום הרשאות

לכל הרצה GitHub מנפיק טוקן זמני, `GITHUB_TOKEN`, שתקף רק למאגר הזה ורק עד סוף ה־job. המפתח `permissions` קובע
מה מותר לו. הכלל מהתיעוד של GitHub פשוט ומחמיר: **ברגע שכותבים הרשאה אחת, כל מה שלא נכתב הוא `none`.** לכן
`permissions: contents: read` ברמה העליונה הוא נקודת התחלה טובה, וכל job מוסיף רק את מה שהוא צריך:

| job | ההרשאות | למה |
| --- | --- | --- |
| `test`, ‏`kind-smoke`, הבדיקות של Python | `contents: read` | ‏checkout בלבד |
| `image` | `contents: read`, ‏`packages: write` | דחיפה ל־GHCR |
| `publish` של ה־release | `contents: write` | יצירת release ותגית |
| ‏job שמתחבר לענן ב־OIDC | `id-token: write` | בקשת אסימון זהות (סעיף 23.6.4) |

למה לא להשאיר את ברירת המחדל של המאגר? כי היא הגדרה שמישהו יכול לשנות בדף ההגדרות, ואז כל workflow במאגר
מקבל פתאום יותר. הרשאה שכתובה בקובץ נבדקת בסקירת קוד, ונשארת כמו שהיא. ב־PR ממאגר מפוצל (fork), הטוקן מקבל
קריאה בלבד והסודות לא מועברים בכלל, ולכן קוד זר שרץ בבדיקות לא יכול לכתוב.

> **טעות נפוצה:** להחליף את `pull_request` ב־`pull_request_target` „כדי שיהיו סודות”. הטריגר הזה רץ בהקשר של
> מאגר היעד, עם טוקן שכותב ועם הסודות, גם על PR ממאגר מפוצל. אם הוא עושה checkout לקוד של ה־PR ומריץ אותו,
> כל אחד שפותח PR מריץ קוד עם הסודות שלך. הוא נועד לפעולות על ה־PR (תוויות, תגובות), לא להרצת הקוד שלו.

### 23.6.2 סודות ומשתנים

| | סוד (secret) | משתנה (variable) |
| --- | --- | --- |
| מה שומרים בו | טוקן, סיסמה, כתובת webhook שמכילה מפתח | הגדרה לא רגישה: דגל, שם, כתובת ציבורית |
| איך קוראים | `${{ secrets.NAME }}` | `${{ vars.NAME }}` |
| בלוג | מוסתר כ־`***` | מודפס כמו שהוא |
| רמות | ארגון, מאגר, environment | ארגון, מאגר, environment |

ההסתרה בלוג היא רשת ביטחון, לא הגנה: סוד שעבר קידוד (base64, למשל) כבר לא מזוהה ומודפס. הכלל הוא לא להדפיס
סודות בכלל, ולהעביר אותם לפקודה דרך משתנה סביבה.

### 23.6.3 environments ומבקרים חובה

‏environment הוא שם, כמו `production`, שאליו מצמידים כללי הגנה וסודות. job שמפנה אליו מקבל את הסודות שלו רק
אחרי שהכללים מתקיימים. הכלל החשוב ביותר הוא מבקרים חובה (required reviewers): ה־job נעצר ומחכה שאחד מהם
יאשר בדף ההרצה. כך נראה ה־job של ה־release (הקובץ `.github/workflows/release.yml`):

<div dir="ltr" align="left">

```yaml
  publish:
    needs: image
    runs-on: ubuntu-latest
    timeout-minutes: 10
    # The job waits here until a required reviewer of "production" approves.
    environment:
      name: production
      url: ${{ github.server_url }}/${{ github.repository }}/releases/tag/${{ github.ref_name }}
    permissions:
      contents: write
    env:
      GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
      TAG: ${{ github.ref_name }}
      DIGEST: ${{ needs.image.outputs.digest }}
      # An environment secret: readable only by jobs that reference "production".
      RELEASE_WEBHOOK_URL: ${{ secrets.RELEASE_WEBHOOK_URL }}
```

</div>

המבקרים עצמם לא מוגדרים ב־YAML, אלא בהגדרות המאגר, תחת Environments. זה בכוונה: מי שכותב workflow לא יכול
לאשר לעצמו גישה בשינוי של הקובץ. שלושה דברים שכדאי לדעת לפני שמסתמכים על זה:

- **הסודות מחכים לאישור.** לפי התיעוד, job לא מקבל את הסודות של environment עד שמבקר מאשר. כך `RELEASE_WEBHOOK_URL`
  לא נחשף ל־job שאף אחד לא אישר.
- **אפשר למנוע אישור עצמי.** הגדרה ב־environment אוסרת על מי שהפעילה את ההרצה לאשר אותה, כך שתמיד יש שני
  אנשים.
- **זה תלוי בתוכנית.** נכון לאוקטובר 2026, בחשבון החינמי של GitHub, מבקרים חובה, סודות ומשתנים של
  environment זמינים רק במאגרים ציבוריים. התרגיל מסביר מה לעשות אם המאגר שלך פרטי.

### 23.6.4 OIDC: בלי מפתחות ענן בסודות

הדרך הישנה לפרוס לענן מ־CI הייתה לשמור מפתח גישה קבוע כסוד. למפתח כזה אין תאריך תפוגה, הוא עובד מכל מקום, ואם
הוא דולף מלוג או מ־action זדוני, הוא עובד גם אצל מי שמצא אותו. ‏OIDC מחליף אותו באסימון קצר מועד:

1. ‏job עם `permissions: id-token: write` מבקש מ־GitHub אסימון זהות (JWT) חתום.
2. באסימון יש טענות (claims) על ההרצה: המאגר, הענף או התגית, ה־environment. הטענה `sub` נראית, למשל, כמו
   `repo:<owner>/senior-course:environment:production`.
3. בענן מגדירים מראש אמון במנפיק של GitHub, עם תנאי על הטענות: רק המאגר הזה, ורק ה־environment הזה.
4. ה־action של הענף מחליף את האסימון בהרשאות זמניות, לדקות, ומהן הוא פורס.

אין סוד לשמור, ואין מה לסובב. ואם התנאי בענן כולל את ה־environment, הרשאות הייצור זמינות רק ל־job שעבר את
המבקרים. בספר אין פריסה לענן (החלק של AWS בפרק 22 אופציונלי), ולכן ה־OIDC נשאר כאן הסבר.

### 23.6.5 הזרקת סקריפט

ביטוי `${{ }}` בתוך `run` מודבק לסקריפט לפני ש־bash קורא אותו. אם הערך מגיע מבחוץ, כמו כותרת של PR, מי שכתב
את הכותרת כותב חלק מהסקריפט שלך. כותרת שמכילה מירכאות ואחריהן פקודה, למשל הורדה והרצה של סקריפט, הופכת
לפקודה בתוך ה־job. התיקון שהתיעוד של GitHub ממליץ עליו הוא לעבור דרך משתנה סביבה, כמו שכל הסקריפטים בפרק עושים. הערך נשמר בזיכרון של התהליך,
ו־bash לא מפרש אותו כקוד. ‏actionlint מזהה את המקרה הזה, וסעיף 23.8 מראה איך.

## 23.7 releases

### 23.7.1 release על תגית

קובץ ה־release רץ על תגית בצורה `v1.2.3`. ה־job הראשון בונה ודוחף image עם תגיות של גרסה. השני מחכה למבקר,
ואז יוצר release. אלה התגיות שהוא מבקש מ־metadata-action:

<div dir="ltr" align="left">

```yaml
          tags: |
            type=semver,pattern={{version}}
            type=semver,pattern={{major}}.{{minor}}
```

</div>

התגית של Git הופכת לשתי תגיות של image, הגרסה המלאה וגרסת המשנה, ובברירת המחדל גם ל־`latest`, אם זו לא
גרסה מקדימה. את ה־release יוצר הכלי `gh`, שמותקן מראש על המכונות של GitHub:

<div dir="ltr" align="left">

```yaml
    steps:
      - name: Create the GitHub release
        run: |
          gh release create "$TAG" \
            --repo "$GITHUB_REPOSITORY" \
            --verify-tag \
            --generate-notes \
            --notes "Image: ghcr.io/${GITHUB_REPOSITORY_OWNER,,}/shorty@${DIGEST}"
      - name: Announce the release
        # A variable (not secret) switches it on; the secret holds the URL.
        if: vars.ANNOUNCE_RELEASES == 'true' && env.RELEASE_WEBHOOK_URL != ''
        run: |
          curl -fsS -X POST "$RELEASE_WEBHOOK_URL" \
            -H 'Content-Type: application/json' \
            -d "{\"text\": \"shorty ${TAG} released\"}"
```

</div>

| דגל או ביטוי | מה הוא עושה |
| --- | --- |
| `--verify-tag` | נכשל אם התגית לא קיימת במאגר, במקום ליצור תגית חדשה מ־`main` |
| `--generate-notes` | הערות שחרור אוטומטיות מה־PRs שמוזגו מאז ה־release הקודם |
| `--notes` | טקסט שנוסף לפני ההערות האוטומטיות: כאן ה־digest של ה־image |
| `${GITHUB_REPOSITORY_OWNER,,}` | הרחבה של bash לאותיות קטנות, כמו ששם ה־image נכתב |

ה־digest מגיע מה־job הקודם, ולכן הגרסה מצביעה על image מדויק, ולא על תגית שיכולה לזוז. ההודעה על השחרור
מראה את ההבדל בין משתנה לסוד: הדגל `ANNOUNCE_RELEASES` לא רגיש, והוא משתנה. הכתובת מכילה מפתח, והיא סוד
של environment.

חלופה נפוצה ל־`gh` היא הפעולה `softprops/action-gh-release`, שהגרסה הראשית שלה נכון לאוקטובר 2026 היא v3. כאן
`gh` עדיף, כי הוא לא מוסיף קוד של צד שלישי ל־job שמחזיק טוקן שכותב.

### 23.7.2 להצמיד לגרסה ראשית או ל־SHA

כל `uses:` מריץ קוד ממאגר של מישהו אחר, עם הטוקן ועם הסודות של ה־job. השאלה היא איזה קוד:

| צורה | דוגמה | מה מקבלים | הסיכון |
| --- | --- | --- | --- |
| גרסה ראשית | `docker/build-push-action@v7` | תיקונים אוטומטית | התגית זזה; מי שמשתלט על המאגר שלהם מריץ קוד אצלך |
| תגית מלאה | `helm/kind-action@v1.15.0` | גרסה ידועה | גם תגית מלאה אפשר להזיז |
| ‏SHA מלא | `docker/build-push-action@c3c9e263...` | בדיוק הקוד שנבדק | עדכונים רק כשמישהו מעדכן |

המדריך של GitHub לשימוש מאובטח ב־Actions אומר זאת במפורש: הצמדה ל־SHA מלא היא כרגע הדרך היחידה להשתמש
ב־action כגרסה שלא משתנה. שם המדריך הוא „Secure use reference”, והכתובת שלו:
<https://docs.github.com/en/actions/reference/security/secure-use>. הכתובת של שמו הקודם, „Security hardening for
GitHub Actions”, מפנה אליו נכון לאוקטובר 2026.

בספר הבחירה מחולקת לפי סיכון. ה־workflows של הבדיקות משתמשים בגרסאות ראשיות, כמו בנספח ג, כי הם רצים עם טוקן
שקורא בלבד. קובץ ה־release כותב, ולכן כל action בו מוצמד ל־SHA, עם התגית כהערה:

<div dir="ltr" align="left">

```yaml
      - uses: actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1 # v7.0.1
        with:
          persist-credentials: false
      - uses: docker/setup-buildx-action@f87e5991a6d7451dcb8d9637bfbc97413f497069 # v4.4.1
```

</div>

ההערה היא לא קישוט. בלעדיה אף אחד לא יודע מה מוצמד. ‏Dependabot, שיודע לעדכן גם actions, נשען עליה כשהוא
פותח PR עם הגרסה החדשה. כך מקבלים את שני הדברים: קוד קבוע, ועדכונים שעוברים סקירה.

## 23.8 actionlint

קובץ workflow שגוי מתגלה רק אחרי הדחיפה, בהרצה שנכשלת אחרי דקה. או גרוע מזה, לא נכשל: תנאי שתמיד שקר, או
job שאף פעם לא רץ. ‏actionlint בודק את הקבצים על המחשב שלך, לפני ה־push. הוא מבין את התחביר, את הביטויים
ואת ההקשרים, ומריץ shellcheck על כל `run` כש־shellcheck מותקן.

**הוורקפלואים נבדקו ב־actionlint; ההרצה הראשונה שלהם תהיה במאגר שלך.** את ה־runners של GitHub אי אפשר להריץ
מכאן. מה שכן נבדק: actionlint ו־shellcheck נקיים על כל הקבצים. והפקודות של כל job הורצו על מחשב, בסקריפט
המקומי של התרגיל, עם מסד בקונטיינר ועם אשכול kind אמיתי.

### 23.8.1 מה הוא מוצא

בתרגיל יש קובץ עם שש טעויות מכוונות. זה הפלט של actionlint עליו, כפי שהתקבל (רשימת התוויות של ה־runners קוצרה
כאן):

<div dir="ltr" align="left">

```text
.github/workflows/broken-workflow.yml:10:3: unknown permission scope "package". all available permission scopes are "actions", ... "packages", ... [permissions]
.github/workflows/broken-workflow.yml:14:14: label "ubuntu-lates" is unknown. available labels are ... [runner-label]
.github/workflows/broken-workflow.yml:21:32: "github.event.pull_request.title" is potentially untrusted. avoid using it directly in inline scripts. instead, pass it through an environment variable. see https://docs.github.com/en/actions/reference/security/secure-use#good-practices-for-mitigating-script-injection-attacks for more details [expression]
.github/workflows/broken-workflow.yml:25:9: shellcheck reported issue in this script: SC2086:info:1:4: Double quote to prevent globbing and word splitting [shellcheck]
.github/workflows/broken-workflow.yml:27:31: property "test" is not defined in object type {} [expression]
.github/workflows/broken-workflow.yml:29:3: job "image" needs job "build" which does not exist in this workflow [job-needs]
```

</div>

| הממצא | מה היה קורה ב־GitHub |
| --- | --- |
| הרשאה `package` | ה־workflow נדחה כלא תקין; ההרשאה הנכונה היא `packages` |
| התווית `ubuntu-lates` | ה־job מחכה לנצח ל־runner עם תווית שלא קיימת |
| כותרת PR בתוך `run` | הזרקת סקריפט (סעיף 23.6.5) |
| ‏SC2086 | משתנה בלי מירכאות; נשבר ברווח הראשון בערך |
| ‏`steps.test` לא קיים | הביטוי מוערך למחרוזת ריקה, בשקט |
| ‏`needs: build` | ה־workflow נדחה כלא תקין |

שימי לב לשורה החמישית: בלי actionlint, הביטוי הזה לא נכשל. הוא פשוט ריק, והשלב מדפיס „tested packages”. זה
הסוג של באג שעולה בתחקיר אחרי שבועות.

### 23.8.2 מה הוא לא מוצא

כדאי לדעת את הגבולות של כלי שסומכים עליו:

- **קלטים של actions חיצוניים, רק בגרסאות שהוא מכיר.** ‏actionlint מגיע עם רשימה של קלטים ל־actions נפוצים. בבדיקה
  ב־5 באוקטובר 2026, גרסה 1.7.12 זיהתה קלט שגוי של `actions/setup-go@v6`, אבל לא של `@v7`. לכן כל שם קלט
  בפרק נבדק מול ה־`action.yml` של התגית עצמה, ב־GitHub.
- **קלטים של הפעולה המקומית, כן.** הוא קורא את `.github/actions/go-checks/action.yml`, ושם קלט שגוי בקריאה לה
  נתפס. בגלל זה הוא צריך לרוץ מתוך מאגר Git.
- **הסקריפטים בתוך הפעולה המורכבת, לא.** הוא בודק workflows, ולא את ה־`run` של קובצי action. אותם צעדים רצים
  בסקריפט המקומי, וזה מה שבודק אותם.
- **התנהגות בזמן ריצה, לא.** הוא לא יודע שהבדיקות ידלגו, שהמסד לא יעלה, או שה־image לא נטען. בשביל זה יש
  `fail-on-skip`, בדיקת בריאות ובדיקת עשן.

### 23.8.3 הבדיקה המלאה

הסקריפט `tests/lint.sh` מריץ את שני הכלים, ובודק שלכל workflow יש `permissions:` ברמה העליונה. על הפתרון
שלושת החלקים שלו נקיים; הפלט המלא נמצא בקובץ ההוראות של התרגיל.

הסקריפט המקומי, עם הדגל `--kind`, מריץ את כל הצעדים של שלושת ה־workflows, ומסכם. המספרים כאן הם מההרצה של
המחברת על הפתרון; שורות Go סופרות `--- PASS`, כולל תתי־בדיקות:

<div dir="ltr" align="left">

```text
==> summary
ch06: 51 passed, 0 skipped
ch11: 27 passed, 0 skipped
ch08: 62 passed in 1.72s
image: shorty:ci built
kind: smoke test passed
```

</div>

## סיכום

- כל job רץ על מכונה נקייה, ומה שעובר בין jobs עובר במפורש.
- סינון לפי נתיב מריץ רק את מה שהשתנה, אבל workflow מסונן לא מתאים לבדיקה חובה.
- המפתח `shell: bash` מוסיף `pipefail`.
- קבוצת מקביליות מבטלת הרצות מיושנות של PR, ולכל job יש זמן קצוב.
- פעולה מורכבת מאחדת צעדים בתוך job; ‏workflow לשימוש חוזר מאחד תהליך שלם עם jobs משלו.
- ‏service container עם בדיקת בריאות נותן מסד אמיתי, ו־`fail-on-skip` מונע ירוק שלא בדק כלום.
- ה־image מתפרסם רק מ־`main` ורק אחרי בדיקת עשן; ההרשאה לדחוף ניתנת רק ל־job שדוחף.
- ‏`permissions` מתחיל מקריאה; סודות של environment זמינים רק אחרי אישור; ‏OIDC מחליף מפתחות ענן קבועים.
- גרסה ראשית נוחה, ‏SHA מלא קבוע; ב־workflow שכותב, מצמידים ל־SHA עם התגית כהערה.
- ‏actionlint תופס טעויות לפני push, אבל לא מחליף הרצה של הפקודות עצמן.

## מונחים לזכור

| מונח בעברית | באנגלית | במשפט אחד |
| --- | --- | --- |
| אינטגרציה רציפה | continuous integration (CI) | בדיקה אוטומטית של כל שינוי, באותה סביבה ובאותן פקודות |
| workflow | workflow | קובץ YAML עם טריגרים ו־jobs, ב־`.github/workflows` |
| job | job | רצף צעדים על runner אחד; ‏jobs מקבילים אלא אם יש `needs` |
| step | step | פקודת shell או שימוש ב־action, בתוך job |
| runner | runner | המכונה שמריצה job; ב־GitHub, מכונה חדשה לכל job |
| action | action | יחידה לשימוש חוזר שנקראת ב־`uses` |
| פעולה מורכבת | composite action | ‏action שהוא רשימת צעדים, ורץ בתוך ה־job הקורא |
| workflow לשימוש חוזר | reusable workflow | ‏workflow שלם שנקרא מ־workflow אחר דרך `workflow_call` |
| service container | service container | קונטיינר שה־runner מפעיל לצד job, כמו מסד לבדיקות |
| סינון לפי נתיב | path filter | הרצה רק כששונה קובץ שמתאים לתבנית |
| קבוצת מקביליות | concurrency group | מגבלה של הרצה אחת בכל פעם לכל ענף או PR |
| ‏environment | environment | שם עם כללי הגנה, סודות ומשתנים, שאליו job מפנה |
| מבקר חובה | required reviewer | מי שחייב לאשר job לפני שהוא רץ ומקבל את הסודות |
| ‏OIDC | OpenID Connect | אסימון זהות קצר מועד שהענן מחליף בהרשאות זמניות |
| הזרקת סקריפט | script injection | ערך מבחוץ שהופך לקוד, כי ביטוי הודבק לתוך `run` |
| הצמדה ל־SHA | SHA pinning | ‏`uses` עם commit מלא במקום תגית, כדי שהקוד לא ישתנה |
| בדיקת עשן | smoke test | בדיקה קצרה שהמערכת עלתה ועונה, לא בדיקה של כל ההתנהגות |

## טעויות נפוצות

- **בלי `permissions` בקובץ.** ה־workflow מקבל את ברירת המחדל של המאגר, שמישהו יכול להרחיב בהגדרות. כותבים
  `contents: read` ברמה העליונה, ומוסיפים לכל job רק את מה שהוא צריך.
- **ביטוי של קלט חיצוני בתוך `run`.** כותרת של PR, שם ענף או גוף של תגובה הופכים לקוד. מעבירים אותם דרך
  `env` ומשתמשים במשתנה במירכאות.
- **‏CI שמדלג בשקט.** בדיקות שמדלגות בלי מסד מדווחות `ok`. ב־job שבו המסד אמור להיות, דילוג הוא כישלון.
- **‏service container בלי בדיקת בריאות.** ה־steps מתחילים לפני שהמסד מוכן, והבדיקות נכשלות לפעמים. מוסיפים
  `--health-cmd` ב־`options`.
- **‏`packages: write` ברמת ה־workflow.** כל job, כולל הבדיקות שמריצות קוד מ־PR, מחזיק טוקן שכותב. ההרשאה
  עוברת ל־job שדוחף.
- **בדיקה חובה על workflow מסונן לפי נתיב.** PR שלא נגע בתיקייה נשאר חסום עם בדיקה „ממתינה” לנצח.
- **צינור ל־`tee` בלי `pipefail`.** הבדיקות נכשלות וה־step ירוק, כי קוד היציאה הוא של `tee`. כותבים
  `shell: bash`.

## שאלות ראיון

1. **איך תבני CI למאגר אחד עם כמה שירותים?**
   ‏workflow לכל שירות, עם סינון לפי התיקייה שלו, לפי הקובץ של עצמו ולפי כל קוד משותף שהוא תלוי בו. צעדים זהים
   עוברים לפעולה מורכבת או ל־workflow לשימוש חוזר, כדי שתיקון אחד יגיע לכולם. המחיר של הסינון: ‏workflow שלא
   רץ משאיר בדיקה ממתינה, ולכן בדיקה חובה צריכה להיות job שרץ תמיד ומסכם, או כלי שמחשב מה השתנה בתוך ה־job.
   השיקול המרכזי הוא זמן משוב מול כיסוי: לא להריץ הכול על כל שינוי, ולא לפספס שינוי בקוד משותף.

   **המשך:** ומה אם שירות A תלוי בספרייה בתיקייה של שירות B? מוסיפים את התיקייה ל־`paths` של A, או מעבירים
   את הספרייה לתיקייה משותפת שכל התלויים מסננים לפיה. בשני המקרים התלות כתובה בקובץ, ולא בזיכרון של מישהו.

2. **למה להגדיר `permissions` בכל workflow, אם יש ברירת מחדל?**
   כי ברירת המחדל היא הגדרה של המאגר או הארגון, שמשתנה מחוץ לקוד ובלי סקירה. כשהיא רחבה, כל action של צד
   שלישי וכל קוד בבדיקות מחזיקים טוקן שיכול לכתוב לקוד, ל־releases ול־packages. הגדרה בקובץ מתחילה מקריאה,
   מוסיפה לכל job רק את מה שהוא צריך, ונבדקת בסקירה כמו כל שינוי אחר. ההשפעה של פריצה ל־action אחד מצטמצמת
   לאותו job.

3. **איך ה־CI צריך להתחבר לענן?**
   ב־OIDC: ה־job מבקש מ־GitHub אסימון זהות חתום, והענן, שהוגדר לסמוך על המנפיק עם תנאים על המאגר, הענף
   וה־environment, מחליף אותו בהרשאות לדקות. אין מפתח קבוע בסודות שיכול לדלוף ולעבוד מכל מקום. התנאים הם
   העיקר: תנאי על המאגר בלבד נותן הרשאות ייצור לכל ענף, ולכן מצמידים גם את ה־environment, שמוגן במבקרים.
   החיסרון: הגדרה ראשונית בצד הענן, ותלות בזמינות של המנפיק.

4. **האם להצמיד actions לגרסה ראשית או ל־SHA?**
   גרסה ראשית נותנת תיקוני אבטחה בלי עבודה, אבל התגית יכולה לזוז, ומי שמשתלט על מאגר ה־action מריץ קוד עם
   הסודות שלך. ‏SHA מלא קבוע, ולפי המדריך של GitHub זו הדרך היחידה לגרסה שלא משתנה, אבל עדכונים מגיעים רק
   כשמישהו מעדכן. הפשרה המקובלת: ‏SHA עם התגית כהערה, ו־Dependabot שפותח PRs לעדכון. ואם צריך לבחור איפה
   להשקיע, מתחילים ב־workflows שמחזיקים טוקן שכותב או סודות של ייצור.

   **המשך:** ו־actions של GitHub עצמה, כמו `actions/checkout`? הסיכון קטן יותר, אבל הוא לא אפס, והמדיניות
   פשוטה יותר כשהיא אחידה. ארגונים רבים מצמידים הכול, ומפעילים מדיניות שדורשת SHA.

5. **ה־CI ירוק, ובייצור השירות לא עולה. מה חסר?**
   בדיקה של מה שנפרס, ולא רק של הקוד. בדיקות יחידה לא יודעות על הרשאות ב־image, על משתנה חובה שחסר בקובץ
   הפריסה או על בדיקת מוכנות לנתיב לא נכון. אשכול kind בתוך ה־CI, עם ה־image שנבנה וקובצי הפריסה האמיתיים,
   ובדיקת עשן אחת, תופסים את רוב אלה בכמה דקות. עוד דבר לבדוק: האם ה־CI דילג על משהו בשקט, כמו בדיקות מסד בלי מסד.
   ובסוף, אם ה־image שנבדק הוא לא ה־image שנפרס, צריך לפרוס לפי digest.

## פתק Obsidian

הפתק המוכן נמצא גם כקובץ: `obsidian/23-github-actions.md`. זה תוכנו:

- **מבנה:** כל job רץ על מכונה נקייה; צעד הוא פקודה או action.
- **בין jobs:** רק דרך `needs` ו־`outputs`, מטמון, או registry.
- **סינון לפי נתיב:** התיקייה, הקובץ של ה־workflow והקוד המשותף; לא לבדיקה חובה.
- **‏shell:** ‏`shell: bash` מוסיף `pipefail`; בפעולה מורכבת חובה בכל צעד.
- **‏service container:** תמיד עם בדיקת בריאות; הכתובת `localhost` כשה־job רץ על המכונה עצמה.
- **ירוק שקרי:** דילוג על בדיקות מסד ב־CI הוא כישלון.
- **הרשאות:** ‏`contents: read` למעלה; ‏`packages: write` ו־`contents: write` רק ב־job שצריך.
- **סודות ומשתנים:** סוד מוסתר בלוג, משתנה לא; סודות של environment רק אחרי אישור.
- **‏OIDC:** אסימון קצר מועד במקום מפתח קבוע; תנאי על המאגר וה־environment.
- **הצמדה:** גרסה ראשית לבדיקות, ‏SHA עם הערה ל־workflow שכותב.
- **‏actionlint:** לפני כל push; לא מכיר כל גרסה של כל action.
- **שאלה פתוחה:** ___

<div dir="ltr" align="left">

```text
on: push: branches: [main] paths: ["ch06/**", ".github/workflows/ch06-shorty.yml"]   pull_request: paths: [...]
permissions: contents: read      job: permissions: packages: write | contents: write | id-token: write
concurrency: group: ${{ github.workflow }}-${{ github.ref }}  cancel-in-progress: ${{ github.event_name == 'pull_request' }}
defaults: run: shell: bash   (bash --noprofile --norc -eo pipefail)
actions/setup-go@v7 go-version-file cache-dependency-path   go run honnef.co/go/tools/cmd/staticcheck@2026.2.1 ./...
astral-sh/setup-uv@v10   uv lock --check   uv sync --frozen   uv run --no-sync pytest -q
services: postgres: image options: --health-cmd "pg_isready ..."   ports: 5432:5432   DATABASE_URL=...@localhost:5432/...
docker/metadata-action@v6  docker/login-action@v4 (GITHUB_TOKEN)  docker/build-push-action@v7 push load cache-from type=gha
helm/kind-action@v1.15.0   kind load docker-image shorty:ci --name ch23   kubectl rollout status   if: failure()
environment: name: production   secrets.X   vars.X   gh release create "$TAG" --verify-tag --generate-notes
git ls-remote https://github.com/<owner>/<action> 'refs/tags/vX.Y.Z*'   uses: owner/action@<40-char-sha> # vX.Y.Z
actionlint   tests/lint.sh   tests/run_ci_locally.sh --kind
my first run on GitHub: date ___  total minutes ___  slowest job ___
```

</div>

## תרגיל

**מטרה:** לכתוב את ה־workflows של שלושה פרויקטים מהספר, מהבדיקות עם מסד אמיתי ועד גרסה שמתפרסמת מאחורי
אישור, ולוודא שכולם נקיים ב־actionlint.
**זמן:** שני ימי לימוד, ימים 66 ו־67, כ־65 דקות בנייה ביום הראשון וכ־40 ביום השני.
**ידע מקדים:** הפרק; פרקים 6, 8 ו־11, ששלושת הפרויקטים שלהם במאגר שלך; פרקים 19 ו־20; והכלים של פרק 2.

הצעדים, קוד ההתחלה, הרמזים והפתרון נמצאים
ב־[`exercises/ch23-ci/README.md`](../exercises/ch23-ci/README.md).

**קריטריון סיום:**

- הבדיקה הסטטית נקייה, ולכל workflow יש `permissions:` משלו.
- ההרצה המקומית המלאה מסתיימת בסיכום, בלי בדיקות שדילגו ב־`ch11`.
- ה־image מתפרסם רק מ־`main` ואחרי בדיקת העשן, וההרשאה לדחוף רק ב־job שדוחף.
- כל action בקובץ ה־release מוצמד ל־SHA מלא, ויצירת הגרסה עוברת דרך `production`.
- שש הבעיות של הקובץ השבור, והתיקון של כל אחת, רשומות בפתק.

## העשרה אופציונלית

- המדריך „Secure use reference” של GitHub
  (<https://docs.github.com/en/actions/reference/security/secure-use>) מכסה עוד נושאים: ‏runners בהרצה עצמית,
  הרשאות של `GITHUB_TOKEN`, ו־`pull_request_target`. קראי את החלק על הזרקת סקריפט עם הקובץ השבור פתוח.
- הוסיפי קובץ `.github/dependabot.yml` עם `package-ecosystem: github-actions`, וראי איך Dependabot מציע עדכון
  ל־SHA המוצמדים ב־`release.yml`.
- ‏job שבונה image פעם אחת, מייצא אותו כקובץ, ומעביר אותו ל־`kind-smoke` ול־`image` דרך artifact. מדדי
  ותרשמי את ההבדל בזמן מול הגרסה שבונה פעמיים, ובדקי אם ה־digest שנבדק זהה לזה שנדחף.
- החליפי את ה־manifests ב־`kind-smoke` ב־chart של Helm מפרק 21, עם `helm install --wait`, וחשבי מה נבדק
  עכשיו שלא נבדק קודם.
- ‏golangci-lint (הגרסה בנספח ג) מריץ staticcheck ועוד עשרות בודקים. נסי את
  `golangci/golangci-lint-action@v9` במקום הצעד של staticcheck, והשווי את משך הצעד.

</div>
