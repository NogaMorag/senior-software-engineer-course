---
direction: rtl
lang: he
---

<div dir="rtl" lang="he">

# תרגיל פרק 19: `images`, images קטנים ובטוחים ל־shorty ול־receiver

**מטרה:** לכתוב Dockerfile רב־שלבי ל־shorty על distroless ו־Dockerfile ל־receiver עם uv, שניהם רצים כמשתמש
שאינו root ועם מערכת קבצים לקריאה בלבד; להרים אותם ב־compose עם בדיקות בריאות; לסרוק אותם, להפיק SBOM,
ולתייג בגרסה ובמזהה commit. את הגדלים את מודדת בעצמך.

**זמן משוער:** שני ימי לימוד. יום 58: כ־60 דקות בנייה (שני ה־Dockerfiles והמדידה). יום 59: כ־60 דקות
(compose, סריקה ו־SBOM, תגיות).

**ידע מקדים:** פרק 19; פרק 6 (shorty) ופרק 8 (ה־receiver), שמהם התרגיל מביא עותקים; פרק 7 (uv); Docker
פועל מפרק 2. לא מתקינים כלים חדשים: Trivy, Grype ו־Syft רצים כקונטיינרים.

## מה בתיקייה הזאת

| קובץ | תפקיד |
| --- | --- |
| `starter/shorty/` | עותק של shorty מפרק 6, עם `/readyz` ופקודת `healthcheck` (ראי בהמשך) |
| `starter/shorty/Dockerfile` | שלד עם `TODO` של יום 58 |
| `starter/shorty/Dockerfile.naive` | הבנייה הנאיבית, שלמה, רק כדי למדוד מולה |
| `starter/shorty/.dockerignore` | `TODO` של יום 58 |
| `starter/receiver/` | עותק של ה־receiver מפרק 8 (פרויקט uv), כולל הבדיקות שלו |
| `starter/receiver/Dockerfile`, `starter/receiver/.dockerignore` | שלדים עם `TODO` של ימים 58 ו־59 |
| `starter/compose.yaml` | שלושה שירותים; בדיקות הבריאות והתלויות הן `TODO` של יום 59 |
| `tests/check_images.sh` | בונה, מריץ ובודק: משתמש, קריאה בלבד, בקשות, compose; מדפיס גדלים |
| `tests/scan_images.sh` | SBOM עם Syft, סריקה עם Grype ועם Trivy, לתיקייה `reports/` |
| `data/signature-vectors.json` | וקטורי החתימה של פרק 8, שהבדיקות של ה־receiver קוראות |
| `solution/` | הפתרון המלא; עותקים זהים של `tests/` יושבים ב־`starter/tests/` וב־`solution/tests/` |

**מאיפה הקוד.** התיקייה `shorty/` היא העתק של `exercises/ch06-shorty/solution/` (בלי `Makefile`), והמודול
נשאר `example.com/ch06-shorty`, כדי שאפשר יהיה להשוות אותו לפרק 6 בפקודת `diff`. נוספו שלושה דברים, עם
בדיקות: הקובץ `internal/app/ready.go` (הנתיב `GET /readyz`, ומתג `POST /debug/ready` על פורט הניפוי
הפרטי, לתרגיל של פרק 20), הקובץ `cmd/shorty/healthcheck.go` (בדיקת בריאות בלי `curl`), והמילה `readyz`
ברשימת הקודים השמורים. התיקייה `receiver/` היא העתק של `exercises/ch08-receiver/solution/` בלי שינוי.

## יום 58: Dockerfile ל־Go, distroless, ו־image ל־Python עם uv

**1. העתקה.** התרגיל יושב בתיקייה `ch19` של מאגר העבודה שלך, יחד עם וקטורי החתימה:

<div dir="ltr" align="left">

```bash
mkdir -p ~/senior-course/ch19
cp -r $COURSE/exercises/ch19-images/starter/. ~/senior-course/ch19/
cp -r $COURSE/exercises/ch19-images/data ~/senior-course/ch19/data
cd ~/senior-course/ch19
(cd shorty && go test ./...)
(cd receiver && uv sync --frozen && uv run pytest -q)
git add . && git commit -m "ch19: starter"
```

</div>

שני השירותים ירוקים כבר עכשיו: התרגיל לא משנה קוד, רק את מה שעוטף אותו.

**2. הבנייה הנאיבית, למדידה.** מתוך התיקייה `ch19`:

<div dir="ltr" align="left">

```bash
docker build -t shorty:naive -f shorty/Dockerfile.naive shorty
docker image ls shorty:naive
docker image inspect --format '{{.Config.User}}' shorty:naive
```

</div>

רשמי את הגודל. השורה האחרונה ריקה: אין `USER`, ולכן ה־image רץ כ־root.

**3. מלאי את `shorty/.dockerignore` ואת `shorty/Dockerfile`** לפי ה־`TODO` שבהם וסעיף 19.2. ואז:

<div dir="ltr" align="left">

```bash
docker build -t shorty:dev shorty
docker run --rm -p 127.0.0.1:8080:8080 -e SHORTY_BASE_URL=http://localhost:8080 shorty:dev
```

</div>

בחלון שני:

<div dir="ltr" align="left">

```bash
curl -s localhost:8080/healthz
curl -s localhost:8080/readyz
curl -s -X POST localhost:8080/v1/links -H 'Content-Type: application/json' -d '{"url":"https://go.dev/","code":"gohome"}'
```

</div>

עצרי ב־Ctrl+C. השירות מקבל את האות ונכבה מיד, כי `ENTRYPOINT` בצורת exec.

**4. מדדי את המטמון.** שני ניסויים, ורשמי לכל אחד מה נבנה מחדש (השורות `CACHED` בפלט):

<div dir="ltr" align="left">

```bash
echo "// touch" >> shorty/internal/links/code.go
docker build --progress=plain -t shorty:dev shorty 2>&1 | grep -E '^#[0-9]+ (CACHED|\[build)'
echo "// touch" >> shorty/internal/links/store_test.go
docker build --progress=plain -t shorty:dev shorty 2>&1 | grep -E '^#[0-9]+ (CACHED|\[build)'
git checkout shorty/internal/links/
```

</div>

בניסוי הראשון הורדת המודולים מגיעה מהמטמון והקומפילציה רצה. בשני, שינוי בקובץ בדיקה לא שובר אף שכבה,
בזכות `.dockerignore`.

**5. בדקי שהוא לא root**, בשתי הדרכים של סעיף 19.3.3:

<div dir="ltr" align="left">

```bash
docker image inspect --format '{{.Config.User}}' shorty:dev
docker run -d --name shorty-check --read-only -p 127.0.0.1:18080:8080 -e SHORTY_BASE_URL=http://localhost:18080 shorty:dev
docker top shorty-check -o pid,uid
docker rm -f shorty-check
```

</div>

**6. מלאי את `receiver/.dockerignore` ואת `receiver/Dockerfile`** לפי סעיף 19.4. את ה־`TODO` של הסרת pip
השאירי ליום 59. ואז:

<div dir="ltr" align="left">

```bash
docker build -t receiver:dev receiver
docker run --rm --entrypoint id receiver:dev
docker run --rm --read-only -p 127.0.0.1:8081:8081 -e RECEIVER_DEFAULT_SECRET=change-me receiver:dev
```

</div>

בחלון שני, `curl -s localhost:8081/stats`. ב־image הזה יש `id`, ולכן הבדיקה הרגילה עובדת.

**7. הגדלים.** הריצי `docker image ls` ומלאי את הטבלה בפתק:

| image | גודל שמדדת |
| --- | --- |
| `shorty:naive` | ___ |
| `shorty:dev` | ___ |
| `receiver:dev` | ___ |

## יום 59: compose, סריקה, SBOM ותגיות

**8. מלאי את `compose.yaml`** לפי סעיף 19.5: בדיקת בריאות ל־receiver (שורת Python), בדיקת בריאות ל־shorty
(`/shorty healthcheck`), והתלויות של `smoke`. ואז:

<div dir="ltr" align="left">

```bash
docker compose -p ch19 up --build --detach --wait receiver shorty
docker compose -p ch19 ps
docker compose -p ch19 run --rm smoke
```

</div>

**9. הבדיקה המלאה.** הסקריפט בונה הכול מחדש, ומנקה אחריו (גם את compose):

<div dir="ltr" align="left">

```bash
bash tests/check_images.sh
```

</div>

**10. סריקה ו־SBOM.** בהרצה הראשונה הכלים מורידים את מאגרי הפגיעויות, וזה לוקח כמה דקות:

<div dir="ltr" align="left">

```bash
bash tests/scan_images.sh
less reports/receiver.trivy.txt
```

</div>

**11. החליטי על ממצא.** חפשי בדוח של Trivy את הקטע `Python (python-pkg)`. בדקי איפה הספריות האלה יושבות:

<div dir="ltr" align="left">

```bash
docker run --rm --entrypoint python receiver:dev -c "import pip, sys; print(pip.__file__)"
```

</div>

אם הן בתוך pip של הבסיס, מלאי את ה־`TODO` של יום 59 בקובץ `receiver/Dockerfile`, בני מחדש, וסרקי שוב. כתבי
בפתק ממצא אחד שנשאר, ומה החלטת לגביו ולמה.

**12. תגיות.** שמרי את העבודה, ואז תייגי בגרסה ובמזהה ה־commit:

<div dir="ltr" align="left">

```bash
git add . && git commit -m "ch19: images, compose, scans"
SHA=$(git rev-parse --short=12 HEAD)
docker build -t shorty:0.1.0 -t "shorty:$SHA" --build-arg GIT_SHA="$SHA" shorty
docker build -t receiver:0.1.0 -t "receiver:$SHA" --build-arg GIT_SHA="$SHA" receiver
docker image inspect --format '{{index .Config.Labels "org.opencontainers.image.revision"}}' shorty:0.1.0
docker image ls --filter reference='shorty:*' --filter reference='receiver:*'
```

</div>

את התיקייה `reports/` אל תכניסי ל־Git (היא גדולה ומשתנה); הוסיפי אותה ל־`.gitignore`, ושמרי את מה שחשוב
בפתק.

## פלט צפוי

סוף `check_images.sh` (הגדלים מושמטים: אותם את רושמת בעצמך):

<div dir="ltr" align="left">

```text
== compose
ok   compose receiver is healthy
ok   compose shorty is healthy
ok   smoke delivery through compose DNS: 200 {"status":"ok","duplicate":false}
== sizes (record them in your notes)
receiver:dev	___
shorty:dev	___
shorty:naive	___
PASS: 15 checks
```

</div>

סוף `scan_images.sh` (מספר הרכיבים והממצאים אצלך יהיה אחר):

<div dir="ltr" align="left">

```text
== receiver:dev
sbom   reports/receiver.spdx.json (___ SPDX elements)
grype  reports/receiver.grype.txt (___ findings, all severities)
trivy  reports/receiver.trivy.txt
```

</div>

## רשימת בדיקה

- [ ] הקובץ `shorty/Dockerfile` רב־שלבי, מוריד מודולים לפני העתקת הקוד, ובונה עם `CGO_ENABLED=0`, `-trimpath` ו־`-s -w`.
- [ ] שינוי בקובץ בדיקה לא שובר אף שכבה של shorty.
- [ ] שני ה־images מוגדרים עם משתמש מספרי שאינו root, ושניהם עונים עם `--read-only`.
- [ ] ב־image של ה־receiver אין uv ואין את תיקיית `src/`, והוא מאזין לכל הכתובות.
- [ ] compose מעלה שני שירותים בריאים, ושירות הבדיקה מגיע ל־receiver בשמו.
- [ ] דוחות הסריקה וקובצי ה־SBOM שמורים, וממצא אחד והחלטה כתובים בפתק.
- [ ] הסקריפט `check_images.sh` מסתיים ב־`PASS: 15 checks`, והגדלים בטבלה.

## טעויות נפוצות ותיקונים

- **`SHORTY_BASE_URL: required, not set` והקונטיינר יוצא מיד.** shorty מסרב לעלות בלי כתובת בסיס (פרק 6).
  מעבירים אותה עם `-e` ב־`docker run`, או ב־`environment` ב־compose.
- **`curl: (52) Empty reply from server` מול ה־receiver.** הוא מאזין ל־`127.0.0.1` בתוך הקונטיינר. בודקים
  ש־`RECEIVER_HOST=0.0.0.0` מוגדר בשלב הריצה, לא בשלב הבנייה.
- **`exec: "id": executable file not found` על shorty.** זה צפוי: ב־distroless אין `id`. בודקים עם
  `docker image inspect` ועם `docker top`, כמו בצעד 5.
- **`exec /shorty: no such file or directory` כשהקובץ קיים.** הבינארי נבנה עם cgo ומחפש ספריית C שאין
  בבסיס. בודקים ש־`CGO_ENABLED=0` כתוב באותה פקודת `RUN` של `go build`.
- **`uv sync` נכשל ב־`The lockfile at uv.lock needs to be updated`.** הדגל `--locked` עושה את עבודתו: מישהו
  שינה את `pyproject.toml` בלי לנעול מחדש. מריצים `uv lock` מחוץ ל־Docker ושומרים את הנעילה.
- **compose מחכה ואז `dependency failed to start: container ch19-shorty-1 is unhealthy`.** בדיקת הבריאות
  נכתבה בצורת מעטפת (`CMD-SHELL` או מחרוזת), ואין מעטפת ב־distroless. כותבים אותה כרשימה שמתחילה ב־`CMD`.

## רמזים

- ב־`shorty/Dockerfile`, הורדת המודולים היא `go mod download`, והקומפילציה בונה את `./cmd/shorty`.
- הצורה של מטמון מוצמד היא `RUN --mount=type=cache,target=<dir>`, ואפשר כמה `--mount` באותה הוראה.
- ב־`receiver/Dockerfile`, שני הצעדים של `uv sync` נבדלים רק בדגלים, והראשון מקבל את שני הקבצים כ־bind mount.
- יצירת המשתמש: `groupadd --system --gid 10001 app` ואחריה `useradd` עם אותו מספר ובלי תיקיית בית.
- בדיקת הבריאות של ה־receiver: `python -c` עם `urllib.request.urlopen` ו־`timeout=2`.

## הפתרון

הפתרון נמצא ב־`solution/`. **למה הוא עובד כך:** בשני ה־Dockerfiles, השכבה של התלויות תלויה רק בקובצי
התלויות, ולכן שינוי בקוד בונה מחדש רק את הקומפילציה או את התקנת הפרויקט. שלב הריצה מעתיק תוצר אחד, בינארי
סטטי או סביבה וירטואלית, ולא את הכלים שבנו אותו. המשתמש מספרי ושאינו root, ושום דבר לא כותב לדיסק, כך
ש־`--read-only` עובד בלי חריגות. ובדיקת הבריאות של shorty היא אותו בינארי, כי ב־distroless אין דבר אחר
להריץ.

## מה התרגיל מדגים

- סדר שכבות שנבחר לפי קצב השינוי, ומדידה של מה המטמון חוסך.
- בנייה רב־שלבית ל־Go ול־Python, והגודל שנחסך מול בנייה נאיבית.
- משתמש שאינו root ומערכת קבצים לקריאה בלבד, ובדיקה שלהם בלי כלים בתוך ה־image.
- compose עם בדיקות בריאות ותלויות שמחכות לבריאות, לא רק להפעלה.
- SBOM וסריקה בלי גישה לשקע של Docker, וממצא שהפך לשינוי ב־Dockerfile.

</div>
