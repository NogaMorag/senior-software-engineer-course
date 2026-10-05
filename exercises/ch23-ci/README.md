---
direction: rtl
lang: he
---

<div dir="rtl" lang="he">

# תרגיל פרק 23: `ch23-ci`, אינטגרציה רציפה לשלושה פרויקטים מהספר

**מטרה:** לכתוב את קובצי ה־workflow שבודקים, בונים ופורסים את מה שכבר בנית: `shorty` מפרק 6, מקבל
ה־webhooks מפרק 8 ומעבדת הבידוד מפרק 11 עם PostgreSQL אמיתי. בסוף יש בדיקות על כל שינוי, image
שמתפרסם רק מ־`main`, בדיקת עשן על אשכול kind, וגרסה על תגית שמחכה לאישור. הקבצים האלה עוברים
אחר כך לפרויקט הגמר באבן הדרך M7.

**זמן משוער:** שני ימי לימוד. יום 66: כ־65 דקות בנייה (בדיקות, service container ו־image). יום 67: כ־40 דקות
(kind, environment, release ו־actionlint).

**ידע מקדים:** פרק 23, ושלושת הפרויקטים שלך מפרקים 6, 8 ו־11 (ובפרק 11, הבדיקות שקוראות את `DATABASE_URL`).
ה־Dockerfile מפרק 19 וקובצי הפריסה בסגנון של פרק 20 מועתקים לכאן כמו שהם. מפרק 2 צריך את actionlint (סעיף
2.7), ואת kind עם kubectl (סעיף 2.10). התרגיל מעתיק קבצים לתוך `ch06` שלך, ולא משנה את הקוד של שלושת
הפרויקטים.

> **שימי לב:** את הקבצים אי אפשר להריץ על runners של GitHub מתוך התרגיל. כאן הם נבדקים ב־actionlint, והפקודות
> שלהם רצות על המחשב שלך בסקריפט אחד. ההרצה הראשונה שלהם ב־GitHub תהיה במאגר שלך, אם תבחרי להעלות אותו
> (סוף יום 67).

## מה בתיקייה הזאת

| קובץ | תפקיד |
| --- | --- |
| `starter/.github/actions/go-checks/action.yml` | פעולה מורכבת (composite action): ‏`gofmt`, ‏`go vet`, ‏staticcheck ו־`go test -race` (שלמה) |
| `starter/.github/workflows/ch06-shorty.yml` | הטריגרים, ההרשאות וה־job של הבדיקות שלמים; שני jobs נוספים כ־`TODO` |
| `starter/.github/workflows/ch11-isolation-lab.yml` | ה־job בלי ה־service container (‏`TODO`) |
| `starter/.github/workflows/ch08-receiver.yml` | השלד; צעדי ה־Python הם `TODO` |
| `starter/.github/workflows/release.yml` | השלד של ה־release; שני ה־jobs הם `TODO` |
| `starter/ch06/Dockerfile`, `starter/ch06/.dockerignore` | ה־Dockerfile מפרק 19, בלי תוויות משלו (שלם) |
| `starter/ch06/deploy/shorty.yaml` | ‏Deployment ו־Service ל־kind, עם `imagePullPolicy: Never` (שלם) |
| `starter/ch06/deploy/smoke.sh` | בדיקת העשן: port-forward, יצירת קישור, ציפייה ל־302 (שלם) |
| `tests/lint.sh` | ‏actionlint, ‏shellcheck, ובדיקה שלכל workflow יש `permissions:` |
| `tests/run_ci_locally.sh` | אותם צעדים כמו ב־workflows, על המחשב שלך, עם PostgreSQL בקונטיינר |
| `data/broken-workflow.yml` | ‏workflow עם שש טעויות מכוונות, לסעיף 23.8 |
| `solution/` | המאגר המלא: ‏`.github/`, ‏`tests/`, ועותקים של `ch06`, ‏`ch08` ו־`ch11` מהפתרונות שלהם |

**איפה כל דבר יושב במאגר שלך.** ‏GitHub קורא workflows רק מהתיקייה `.github/workflows` בשורש המאגר, ולכן
העבודה של הפרק הזה לא נכנסת ל־`ch23`, אלא לשורש של `~/senior-course`:

<div dir="ltr" align="left">

```text
~/senior-course/
  .github/workflows/ch06-shorty.yml        paths: ch06/**
  .github/workflows/ch08-receiver.yml      paths: ch08/**
  .github/workflows/ch11-isolation-lab.yml paths: ch11/**
  .github/workflows/release.yml            tags: v*.*.*
  .github/actions/go-checks/action.yml     used by ch06 and ch11
  tests/lint.sh  tests/run_ci_locally.sh
  ch06/  (your shorty + Dockerfile, .dockerignore, deploy/)
  ch08/  (your receiver, with data/)
  ch11/  (your isolation lab, with data/)
```

</div>

כל workflow מסנן לפי התיקייה שלו, ולכן שינוי ב־`ch11` לא מריץ את הבדיקות של `ch06`. התיקייה `solution/` היא
בדיוק העץ הזה, עם הקוד של הפתרונות במקום הקוד שלך.

## יום 66: בדיקות, service container ו־image

**1. העתקה.** מהשורש של המאגר שלך:

<div dir="ltr" align="left">

```bash
cd ~/senior-course
cp -r "$COURSE/exercises/ch23-ci/starter/.github" .
cp -r "$COURSE/exercises/ch23-ci/starter/tests" .
cp "$COURSE/exercises/ch23-ci/starter/ch06/Dockerfile" "$COURSE/exercises/ch23-ci/starter/ch06/.dockerignore" ch06/
cp -r "$COURSE/exercises/ch23-ci/starter/ch06/deploy" ch06/
sudo apt install -y shellcheck
ls ch06/go.mod ch08/uv.lock ch11/go.mod ch08/data ch11/data
```

</div>

הפקודה האחרונה מוודאת ששלושת הפרויקטים נמצאים איפה שה־workflows מחפשים אותם. אם אחד חסר, או שהבדיקות שלו
לא ירוקות אצלך, העתיקי את הפתרון שלו מהתיקייה `solution/` של התרגיל הזה, למשל:

<div dir="ltr" align="left">

```bash
cp -r "$COURSE/exercises/ch23-ci/solution/ch11/." ch11/
```

</div>

**2. קראי את מה שכבר כתוב.** פתחי את `.github/actions/go-checks/action.yml` ואת ה־job בשם `test`
ב־`ch06-shorty.yml`. שימי לב לשלושה דברים: הגרסה של Go נקראת מ־`go.mod`, ‏staticcheck רץ דרך `go run` בגרסה
מוצמדת, והקלט `fail-on-skip` הופך דילוג על בדיקה לכישלון. הריצי את הבדיקה הסטטית:

<div dir="ltr" align="left">

```bash
tests/lint.sh
```

</div>

**3. ה־job של Python.** ב־`ch08-receiver.yml`, החליפי את הצעד `TODO` בצעדים של סעיף 23.2:

| צעד | פקודה או action |
| --- | --- |
| התקנת uv | ‏`astral-sh/setup-uv@v10`, עם `version`, ‏`working-directory`, ‏`enable-cache` ו־`cache-dependency-glob` |
| קובץ הנעילה עדכני | `uv lock --check` |
| התקנה מדויקת | `uv sync --frozen` |
| ארבע בדיקות | ‏`ruff check .`, ‏`ruff format --check .`, ‏`mypy src`, ‏`pytest -q`, כל אחת אחרי `uv run --no-sync` |

**4. ה־service container.** ב־`ch11-isolation-lab.yml`, הוסיפי ל־job שירות בשם `postgres` ומשתנה סביבה
בשם `DATABASE_URL`, לפי סעיף 23.3. את ההערה `TODO` החליפי בקלט שמכשיל דילוגים:

<div dir="ltr" align="left">

```yaml
          fail-on-skip: "true"
```

</div>

 לפני שאת כותבת, ראי בעיניים למה הקלט הזה
קיים:

<div dir="ltr" align="left">

```bash
cd ~/senior-course/ch11
unset DATABASE_URL
go test ./...
go test -v ./... | grep -c -- '--- SKIP'
```

</div>

הפקודה הראשונה מדפיסה `ok` לכל חבילה. השנייה סופרת את הבדיקות שדילגו, כי אין מסד. זה ה־CI הירוק שלא בדק
כלום.

**5. ה־image.** ב־`ch06-shorty.yml`, כתבי את ה־job בשם `image` לפי סעיף 23.4. היום הוא תלוי רק ב־`test`
(‏`needs: test`); מחר יתווסף לו `kind-smoke`. בדקי את התנאי: רק `push` ל־`main`, ורק שם `packages: write`.

**6. הריצי את הכול מקומית.** הסקריפט מריץ את אותן פקודות כמו ה־workflows, ומקים PostgreSQL בקונטיינר במקום
ה־service container. הוא מוחק את הקונטיינר בסוף, גם כשמשהו נכשל:

<div dir="ltr" align="left">

```bash
cd ~/senior-course
tests/lint.sh
tests/run_ci_locally.sh --image
git add .github tests ch06/Dockerfile ch06/.dockerignore ch06/deploy
git commit -m "Chapter 23: CI for ch06, ch08 and ch11, image build"
```

</div>

## יום 67: kind, environment, release ו־actionlint

**1. בדיקת עשן על kind.** ב־`ch06-shorty.yml`, כתבי את `kind-smoke` לפי סעיף 23.5, ועדכני את `needs`
של `image` כך שיחכה גם לו. כך image מתפרסם רק אחרי שעלה באשכול וענה. הריצי את כל השרשרת מקומית:

<div dir="ltr" align="left">

```bash
tests/run_ci_locally.sh --kind
kind get clusters
```

</div>

הסקריפט מקים אשכול בשם `ch23-local` עם קובץ kubeconfig פרטי, כך ש־`~/.kube/config` שלך לא משתנה, ומוחק אותו
בסוף. הפקודה השנייה צריכה להדפיס `No kind clusters found.`

**2. ה־release.** כתבי את `release.yml` לפי סעיפים 23.6 ו־23.7. כאן כל action מוצמד ל־SHA מלא. כך מוצאים את
ה־SHA של תגית:

<div dir="ltr" align="left">

```bash
git ls-remote https://github.com/docker/build-push-action 'refs/tags/v7.4.0*'
```

</div>

השורה שחוזרת היא ה־SHA ואחריו שם התגית. אם מופיעה גם שורה שנגמרת ב־`^{}`, התגית מוערת (annotated), וה־SHA
שבשורה הזאת הוא של ה־commit; אותו מצמידים. כתבי את התגית כהערה אחרי ה־SHA, כדי שבני אדם ו־Dependabot יבינו
מה מוצמד.

**3. תרגיל actionlint.** העתיקי את הקובץ השבור למאגר זמני, והריצי עליו actionlint:

<div dir="ltr" align="left">

```bash
mkdir -p /tmp/ch23-broken/.github/workflows && cd /tmp/ch23-broken && git init -q
cp "$COURSE/exercises/ch23-ci/data/broken-workflow.yml" .github/workflows/
actionlint
```

</div>

‏actionlint צריך למצוא שש בעיות. לכל אחת, כתבי בפתק מה היא ואיך מתקנים, ותקני עד שהפלט ריק.

**4. בדיקה אחרונה והתחייבות.**

<div dir="ltr" align="left">

```bash
cd ~/senior-course
tests/lint.sh
git add .github && git commit -m "Chapter 23: kind smoke test, release on tag"
```

</div>

**5. אופציונלי: ההרצה הראשונה ב־GitHub.** אם תרצי לראות את ה־workflows רצים, צרי ב־GitHub מאגר חדש (למשל
`senior-course`), חברי אליו את המאגר שלך ודחפי:

<div dir="ltr" align="left">

```bash
git remote add origin git@github.com:<your-user>/senior-course.git
git push -u origin main
```

</div>

בלשונית Actions תראי שלושה workflows רצים, כי הדחיפה הראשונה נוגעת בכל התיקיות. ל־release צריך קודם להגדיר
את ה־environment:

1. ב־Settings, תחת Environments, צרי environment בשם `production`.
2. הוסיפי את עצמך ב־Required reviewers, ושמרי.
3. אופציונלי: תחת Environment secrets הוסיפי `RELEASE_WEBHOOK_URL`, ותחת Environment variables את
   `ANNOUNCE_RELEASES` עם הערך `true`.

ואז תגית:

<div dir="ltr" align="left">

```bash
git tag v0.1.0 && git push origin v0.1.0
```

</div>

ה־job בשם `publish` יחכה לאישור שלך בדף ההרצה. שימי לב לתוכנית של החשבון: נכון לאוקטובר 2026, בחשבון GitHub
Free, אישור של מבקרים, סודות של environment ומשתנים של environment זמינים רק במאגר ציבורי (כך כתוב בדף התיעוד
של environments). אם המאגר שלך פרטי וחינמי, העלי את התרגיל למאגר ציבורי נפרד, או קראי את הצעד הזה בלי להריץ
אותו.

## פלט צפוי

‏`tests/lint.sh` על הפתרון:

<div dir="ltr" align="left">

```text
==> actionlint
actionlint: no findings
==> shellcheck
shellcheck: no findings
==> permissions
every workflow declares its permissions
```

</div>

הסיכום של ההרצה המקומית עם `--kind`, על הפתרון, אצל המחברת. השורות של Go סופרות שורות `--- PASS`, כולל
תתי־בדיקות; את המספרים שלך תרשמי בפתק:

<div dir="ltr" align="left">

```text
==> summary
ch06: 51 passed, 0 skipped
ch11: 27 passed, 0 skipped
ch08: 62 passed in …s
image: shorty:ci built
kind: smoke test passed
```

</div>

ובאמצע, הפלט של בדיקת העשן (התאריך יהיה שונה):

<div dir="ltr" align="left">

```text
healthz: ok
created: {"code":"ci-smoke","url":"https://example.com/from-ci","short_url":"http://localhost:18080/ci-smoke","hits":0,"created_at":"…"}
resolve: 302
smoke test passed
```

</div>

## רשימת בדיקה

- [ ] ‏`tests/lint.sh` נקי: actionlint ו־shellcheck בלי ממצאים, ולכל workflow יש `permissions:` ברמה העליונה.
- [ ] כל workflow מסנן לפי התיקייה שלו ולפי הקובץ של עצמו, ושני ה־workflows של Go מסננים גם לפי הפעולה המורכבת.
- [ ] ההרצה המקומית עם `--kind` מסתיימת בסיכום, בלי דילוגים ב־`ch11`, ובלי קונטיינר או אשכול שנשארו.
- [ ] הרשאות הכתיבה מופיעות רק ב־jobs שצריכים אותן: דחיפת image ויצירת גרסה.
- [ ] ה־job בשם `image` רץ רק על `push` ל־`main`, ותלוי ב־`kind-smoke`.
- [ ] כל action ב־`release.yml` מוצמד ל־SHA מלא, עם התגית כהערה.
- [ ] שש הבעיות של הקובץ השבור רשומות בפתק, עם התיקון של כל אחת.

## טעויות נפוצות ותיקונים

- **‏actionlint אומר `no project was found in any parent directories`.** הרצת אותו מחוץ למאגר Git. הוא מחפש
  את שורש המאגר כדי למצוא את `.github/workflows`; עברי ל־`~/senior-course` (או `git init` בתיקייה זמנית).
- **ה־job של `ch11` נכשל ב־`tests were skipped; is DATABASE_URL set for this job?`.** ‏`DATABASE_URL` לא מוגדר
  ברמת ה־job, או שהוגדר רק בצעד אחד. מעבירים אותו ל־`env` של ה־job, ליד `services`.
- **ה־job של `ch11` נכשל ב־`connection refused`.** חסר מיפוי פורטים לשירות, או שבדיקת הבריאות חסרה והצעדים
  התחילו לפני שהמסד מוכן. שניהם הכרחיים כשה־job רץ על המכונה עצמה.
- **ה־pod של `shorty` נשאר ב־`ErrImageNeverPull`.** ה־image לא נטען לאשכול, או שנטען לאשכול בשם אחר.
  ‏`kind load docker-image` צריך את אותו `--name` כמו `cluster_name`, ואת אותה תגית כמו ב־`shorty.yaml`.
- **‏`uv run` בצעד הבדיקות מתקין מחדש, או מעדכן את `uv.lock`.** חסר `--no-sync`. בלעדיו `uv run` מסנכרן בעצמו,
  והצעד כבר לא בודק את מה ש־`uv sync --frozen` התקין.
- **הדחיפה ל־GHCR נכשלת ב־`denied` או ב־`403`.** ל־job אין `packages: write`, או שה־package כבר קיים ומקושר
  למאגר אחר. בודקים את `permissions` של ה־job, ובדף ה־package את ההגדרה Manage Actions access.

## רמזים

- ב־service container, ‏`options` הוא מחרוזת אחת של דגלים ל־`docker create`. הצורה `>-` של YAML מחברת את השורות
  לשורה אחת.
- הקלט `fail-on-skip` של הפעולה המורכבת הוא מחרוזת, לא בוליאני: כותבים `"true"` במירכאות.
- בבנייה שבתוך `kind-smoke` חייבים `load: true`. בלעדיו ה־image נשאר בתוך ה־builder, והטעינה לאשכול לא
  תמצא אותו.
- ב־PR אין צורך להריץ את `image` בכלל, כי `kind-smoke` כבר בונה את ה־image. התנאי בודק שני דברים: שסוג האירוע
  הוא `push`, ושהענף הוא `refs/heads/main`.
- את ה־digest מוציאים מהפלט של צעד הבנייה, ומעבירים ל־job הבא דרך `outputs` של ה־job.

## הפתרון

‏`solution/.github/`, בעיקר `workflows/ch06-shorty.yml` ו־`workflows/release.yml`, ו־`solution/tests/`.
**למה זה עובד:** כל workflow מתחיל מ־`contents: read` ומוסיף הרשאה רק ל־job שצריך אותה, כך שקוד שרץ בבדיקות
לא מחזיק טוקן שיכול לכתוב. הבדיקות שתלויות במסד נכשלות כשהן מדלגות, ולכן ירוק אומר שהמסד באמת היה שם. ה־image
מתפרסם רק אחרי שאותו Dockerfile, עם אותו מטמון, עלה ב־Kubernetes וענה לבקשה אמיתית. וה־release מחכה לאדם, עם
actions מוצמדים ל־SHA, כי שם הטוקן כותב.

## מה התרגיל מדגים

- ‏workflow לכל פרויקט במאגר אחד, עם סינון לפי נתיב, ‏`concurrency` שמבטל הרצות מיושנות, ומינימום הרשאות.
- בדיקות עם מסד אמיתי כ־service container, והגנה מפני CI ירוק שדילג על הבדיקות.
- פעולה מורכבת אחת לשני פרויקטי Go, במקום אותם ארבעה צעדים פעמיים.
- שרשרת אספקה של image: בנייה עם מטמון, בדיקת עשן על kind, ופרסום ל־GHCR רק מ־`main`.
- ‏release על תגית מאחורי environment מוגן, עם הצמדה ל־SHA, ובדיקה סטטית של הכול ב־actionlint.

</div>
