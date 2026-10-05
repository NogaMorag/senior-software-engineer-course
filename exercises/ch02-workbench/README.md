---
direction: rtl
lang: he
---

<div dir="rtl" lang="he">

# תרגיל פרק 2: סביבת העבודה, `doctor.sh` ובדיקות העשן

**מטרה:** סביבה מותקנת ומאומתת, וסקריפט שמוכיח את זה בכל רגע; בדיקת עשן ל־Go ול־Python; קונטיינר ראשון;
וההתחייבות הראשונה במאגר שלך. ביום 57 מוסיפים את הכלים של חלק ה ומרחיבים את הסקריפט.

**זמן משוער:** שלושה ימי לימוד. יום 0: כ־70 דקות התקנה. יום 3: כ־55 דקות בנייה. יום 57: כ־45 דקות התקנה.

**ידע מקדים:** פרק 0, ופרק 2 כולו (כל יום קורא את הסעיפים שלו). התרגיל לא מייבא קוד מתרגיל קודם.

## מה בתיקייה הזאת

| קובץ | תפקיד |
| --- | --- |
| `starter/doctor.sh` | הרופא: בודק כל כלי של יום 0 (כולל כלי ה־Go ו־golangci-lint) ואת תיקיות הקורס; בלוק חלק ה מסומן `TODO` ליום 57 |
| `starter/test_doctor.sh` | מריץ את `doctor.sh` מול כלים מזויפים ובודק קודי יציאה (11 בדיקות) |
| `starter/hello/` | מודול Go: `main.go` עם `TODO` אחד, ו־`main_test.go` עם בדיקה טבלאית |
| `starter/smoke/` | פרויקט uv: `src/smoke/__init__.py` עם `TODO` אחד, `tests/test_smoke.py`, קובץ נעילה |
| `solution/` | אותו מבנה, מלא: `doctor.sh` עם בלוק חלק ה, `hello` ו־`smoke` גמורים |
| `tests/` | הבדיקות הקנוניות: `test_doctor.sh`, `hello/main_test.go`, `smoke/tests/test_smoke.py` |
| `data/gitignore` | קובץ ההחרגות של המאגר שלך, `~/senior-course/.gitignore` |

הבדיקות שב־`starter/` וב־`solution/` הן העתקים זהים של הקבצים שב־`tests/`.

## יום 0: התקנות

**1. התקיני לפי פרק 2, סעיפים 2.2 עד 2.8**, בסדר שלהם. סעיף 2.7.1 כולל את כלי ה־Go ואת golangci-lint, עם
בדיקת הטביעה שלו. על Mac, סעיף 2.13 מחליף את ההוראות של Windows.

**2. פתחי לשונית חדשה ובדקי גרסאות.** הסקריפט עוד לא אצלך (הוא מגיע עם הספר ביום 3), ולכן בודקים ידנית:

<div dir="ltr" align="left">

```bash
git --version
docker version --format '{{.Client.Version}} / {{.Server.Version}}'
code --version | head -1
go version
staticcheck -version
buf --version
golangci-lint --version
uv --version
uv python find 3.14
ssh -T git@github.com
```

</div>

השווי כל מספר לנספח ג. הפקודה האחרונה צריכה להדפיס `Hi <your-user>!`.

## יום 3: מאגר הקורס ובדיקת העשן

**3. הספר בתוך WSL, והמשתנה `COURSE`** (פרק 2, סעיף 2.9). ודאי שהפקודה הזאת מדפיסה את רשימת התיקיות של הספר:

<div dir="ltr" align="left">

```bash
ls "$COURSE/exercises/ch02-workbench"
```

</div>

**4. המאגר שלך וקוד ההתחלה.**

<div dir="ltr" align="left">

```bash
mkdir -p ~/senior-course
cd ~/senior-course
git init
cp "$COURSE/exercises/ch02-workbench/data/gitignore" .gitignore
mkdir -p ch02
cp -r "$COURSE/exercises/ch02-workbench/starter/." ch02/
```

</div>

**5. הרופא.** כל שורת `FAIL` אומרת איזה סעיף לפתוח. תקני מלמעלה למטה והריצי שוב, עד קוד יציאה 0:

<div dir="ltr" align="left">

```bash
bash ch02/doctor.sh
echo "exit code: $?"
bash ch02/test_doctor.sh
```

</div>

הבדיקות של הסקריפט מסתיימות ב־`9 passed, 2 failed`. שתי הנכשלות בודקות את בלוק חלק ה, והן יעברו ביום 57.

**6. Go.** הריצי את הבדיקה, קראי את שלוש השורות שנכשלות, ומלאי את ה־`TODO` ב־`main.go`:

<div dir="ltr" align="left">

```bash
cd ~/senior-course/ch02/hello
go test ./...
go vet ./...
go test -v -run 'TestGreeting/empty' ./...
go run . Noga
```

</div>

הדגל `-run` מריץ רק את השורה שבחרת מהטבלה; זה שימושי כשמתקנים מקרה אחד.

**7. Python.** מלאי את ה־`TODO` ב־`src/smoke/__init__.py`:

<div dir="ltr" align="left">

```bash
cd ~/senior-course/ch02/smoke
uv sync
uv run pytest -q
uv run ruff check .
uv run mypy src
```

</div>

**8. הקונטיינר הראשון, ועצירה וניקוי** (פרק 2, סעיף 2.12). הריצי את הקונטיינר, ואחר כך תרגלי את הפקודות
שמראות ומנקות מקום. כתבי בפתק כמה מקום תפס כל סוג לפני הניקוי ואחריו, כפי שמדדת:

<div dir="ltr" align="left">

```bash
docker run --rm hello-world
docker system df
docker container prune
docker system df
```

</div>

ואז, ב־PowerShell, עצרי את WSL כולו ופתחי טרמינל חדש. ודאי שהרופא עדיין ירוק אחרי ההפעלה מחדש:

<div dir="ltr" align="left">

```powershell
wsl --shutdown
```

</div>

**9. ההתחייבות הראשונה.** היא כוללת גם את `ch01` מתרגיל פרק 1. אם כבר התחייבת עליו, היא תכיל רק את `ch02`.

<div dir="ltr" align="left">

```bash
cd ~/senior-course
git status --short
git add .
git commit -m "ch01-ch02: postings tally, workbench, doctor and smoke tests"
git log --oneline
```

</div>

## יום 57: הכלים של חלק ה

**10. התקיני את חמשת הכלים** לפי פרק 2, סעיף 2.10, כל אחד עם בדיקת הטביעה שלו, ואת האימג' של kind,
מוצמד לפי הטביעה שלו. לא יוצרים אשכול היום.

**11. הרחיבי את `doctor.sh`.** בבלוק של חלק ה החליפי את ה־`TODO` בחמש שורות `check_tool`, אחת לכל
כלי, עם הגרסה המינימלית מנספח ג. הדגל `--client` של kubectl חשוב: בלעדיו הפקודה מנסה להגיע לאשכול.

<div dir="ltr" align="left">

```bash
cd ~/senior-course
bash ch02/doctor.sh --part-e
bash ch02/test_doctor.sh
git add ch02/doctor.sh
git commit -m "ch02: doctor checks the Part E tools"
```

</div>

## פלט צפוי

הרופא ביום 3 (צורה; המספרים אצלך לפי מה שהתקנת):

<div dir="ltr" align="left">

```text
doctor: checking the course workbench on Linux x86_64 (WSL2)
Day 0 tools:
  ok    git                2.53.0       (need >= 2.28)
  ...
  ok    python             3.14.8       (need >= 3.14.0)
Go tools (go install, Day 0):
  ok    staticcheck        2026.2.1     (need >= 2026.2.1)
  ...
  ok    actionlint         1.7.12       (need >= 1.7.12)
  ok    golangci-lint      2.14.0       (need >= 2.14.0)
Course folders:
  ok    COURSE             found        (/home/noga/senior-software-engineer-course)
  ok    senior-course      git-repo     (/home/noga/senior-course)
doctor: 19 ok, 0 failed
doctor: the workbench is ready
```

</div>

הבדיקות, אחרי שני ה־`TODO` של יום 3:

<div dir="ltr" align="left">

```text
ok  	example.com/ch02-workbench/hello	0.002s
Hello, Noga!
...                                                                      [100%]
3 passed in 0.01s
```

</div>

וביום 57, הבדיקות של הסקריפט:

<div dir="ltr" align="left">

```text
PASS healthy Day 0 workbench
PASS healthy workbench with Part E
...
PASS an unknown option is a usage error
test_doctor: 11 passed, 0 failed
```

</div>

## רשימת בדיקה

- [ ] ביום 0, כל כלי מדפיס גרסה שמתאימה לנספח ג, והחיבור ב־SSH ל־GitHub מברך אותך בשמך.
- [ ] העבודה שלך נמצאת תחת `~/senior-course`, לא תחת `/mnt/c`.
- [ ] הסקריפט `doctor.sh` מסתיים בקוד יציאה 0.
- [ ] שש השורות של הבדיקה הטבלאית עוברות, ו־`go vet` שקט.
- [ ] שלוש בדיקות ה־pytest עוברות, וכלי הניתוח הסטטי נקיים.
- [ ] הקונטיינר מדפיס `Hello from Docker!`, וכתבת בפתק את מה שמדדת ב־`docker system df`.
- [ ] בהתחייבות הראשונה אין `.venv` ואין קבצי בנייה.
- [ ] ביום 57, הרופא עם הדגל `--part-e` מסתיים ב־0, והבדיקות שלו מדווחות על 11 הצלחות.

## טעויות נפוצות ותיקונים

- **להעתיק את קוד ההתחלה לתוך `/mnt/c`.** הרופא מדווח `FAIL` עם `on-/mnt` בשורת התיקייה. מעבירים את התיקייה
  אל `~/senior-course` וממשיכים משם.
- **להריץ את הרופא בלשונית שנפתחה לפני ההתקנה.** כלי שהותקן עכשיו מופיע כ־`missing`. פותחים לשונית חדשה, או
  מריצים `source ~/.bashrc`, ומריצים שוב.
- **שכחת את `COURSE`, או הגדרת אותו רק בלשונית אחת.** השורה `COURSE` מראה `COURSE is not set`. השורה צריכה
  להיות ב־`~/.bashrc` (סעיף 2.9), לא רק בפקודת `export` חד־פעמית.
- **להריץ `pytest` בלי `uv run`.** בדיקת הסביבה הווירטואלית נכשלת, או ש־pytest בכלל לא נמצא. בתיקיית `smoke`
  תמיד `uv run pytest`.
- **ביום 57, להוסיף את חמש השורות מחוץ לבלוק ה־`if`.** הרופא בודק את כלי חלק ה גם בלי `--part-e`, והבדיקה
  של יום 0 עם כלי חלק ה חסרים נכשלת עם קוד 5. מעבירים את השורות לתוך הבלוק.
- **להשוות גרסאות כמחרוזות בשורה שהוספת.** אם כתבת השוואה משלך במקום `check_tool`, בדיקת השוואת הגרסאות
  כמספרים נכשלת. משתמשים ב־`check_tool`, שמשווה דרך `version_ge`.

## רמזים

- ב־`greeting`, הפונקציה `strings.TrimSpace` מסירה רווחים משני הצדדים; את ברירת המחדל בודקים אחרי הקיצוץ.
- ב־`describe`: שלושת השדות הראשונים של `sys.version_info`, והפונקציה `platform.system()`.
- הסקריפט `test_doctor.sh` מדפיס את הפלט המלא של הרופא מתחת לכל בדיקה שנכשלה. קראי אותו לפני שאת משנה קוד.
- לבדיקה ידנית של שורת חלק ה, הריצי את פקודת הגרסה של הכלי לבד, וראי מה המספר הראשון בפלט.

## הפתרון

הקבצים בתיקייה `solution/`. **למה זה עובד כך:** הרופא לא יודע דבר על אף כלי מעבר לשורה אחת בטבלה: שם, מינימום,
הפניה, ופקודת גרסה. לכן הוספת כלי ביום 57 היא חמש שורות, וכל הלוגיקה (קיים, נשלף מספר, משווה כמספר) נבדקת פעם
אחת, ב־`test_doctor.sh`. קוד היציאה הוא מספר הכשלים, כך שכל סקריפט אחר יכול לסמוך עליו. בדיקות העשן קטנות
בכוונה: כל אחת מוכיחה חוליה אחרת בשרשרת (מהדר ובדיקות, סביבה וירטואלית, מנוע קונטיינרים), וכשאחת נופלת ברור
איזו חוליה שבורה.

## מה התרגיל מדגים

- סביבת עבודה כקוד: סקריפט שמתעד מה נדרש, ובודק את זה בכל רגע, במקום רשימה בזיכרון.
- קוד יציאה כחוזה בין תוכניות, ובדיקה של סקריפט מעטפת מול כלים מזויפים.
- השוואת גרסאות נכונה, כמספרים.
- בדיקה טבלאית ב־Go, ובדיקה שמוכיחה ש־pytest רץ בסביבה הנכונה.
- יום התקנה נפרד מיום שימוש, וטביעה לכל הורדה.

</div>
