---
direction: rtl
lang: he
---

<div dir="rtl" lang="he">

# פרויקט הגמר: Hookline, פלטפורמה למסירת webhooks

**מטרה:** לבנות, אבן דרך אחר אבן דרך, שירות שמקבל אירועים ממוצר SaaS ומוסר אותם, חתומים ו„לפחות פעם אחת”, לכל
endpoint מנוי, עם בידוד בין מקבלים, ניסיונות חוזרים, הרצה מחדש של כשלים, ותצפיתיות מלאה. בסוף יש לך פרויקט אחד,
שלך, שאת יודעת להסביר בראיון שורה אחר שורה, עם מספרים שמדדת בעצמך.

**זמן משוער:** עשר אבני דרך לאורך הספר, בימים 22 עד 99; כ־55 עד 90 דקות בנייה בכל יום של אבן דרך. הטבלה בצעד 2
מפרטת את הימים של כל אחת.

**ידע מקדים:** פרק 37 (המפרט), ופרק 38 לפי אבן הדרך. הפרויקט מעתיק ומתאים קוד מהתרגילים שלך: החבילה
`internal/platform` מפרק 6 (`ch06-shorty`), החבילות של פרטי הבעיה, העימוד והאידמפוטנטיות מפרק 9
(`ch09-catalog-api`), וחיבור המסד, המיגרציות והמסד הזמני לבדיקות מפרק 11 (`ch11-isolation-lab`). אבני הדרך
הבאות יביאו קוד מפרקים 8, 12, 13 ו־14. הגרסאות של הספריות ושל המסד מקובעות ב־`go.mod` ובקובץ
ה־compose, נכון לאוקטובר 2026 (נספח ג).

## מה בתיקייה הזאת

| קובץ או תיקייה | תפקיד |
| --- | --- |
| `README.md` | הקובץ הזה: ההתקנה, הדרך, ורשימת אבני הדרך |
| `hookline/` | המימוש המלא, מעודכן עד אבן הדרך האחרונה שפורסמה. לקרוא כשנתקעת |
| `hookline-starter/` | העץ שאת מתחילה ממנו: אותו עץ, אותן בדיקות, עם `panic("TODO: see chapter 38")` בגופי הפונקציות שהן עבודה שלך |
| `hookline-starter/STARTER.md` | הסבר קצר באנגלית על איך ה־starter נוצר |

בתוך כל אחת משתי התיקיות, המבנה זהה (פרק 37, סעיף 37.3):

| נתיב | מה יש בו | אבן דרך |
| --- | --- | --- |
| `cmd/api/main.go` | חיווט השרת: הגדרות, מאגר, שני listeners, כיבוי מסודר | M1 |
| `cmd/hookline-admin/main.go` | כלי המפעילה: `migrate`, ‏`create-tenant` | M1, M2 |
| `api/openapi.yaml` | החוזה: OpenAPI 3.1, כל נתיב וכל סוג בעיה | M1 |
| `internal/config` | ההגדרות מהסביבה (שלם) | M1 |
| `internal/platform` | השרת, ה־middleware, הלוגים וה־pprof מפרק 6 (שלם) | M1 |
| `internal/problem`, ‏`internal/page` | פרטי בעיה ועימוד מפרק 9 (שלמים) | M1 |
| `internal/domain` | טיפוסי התחום, מזהים, מפתחות, אימות קלט (שלם) | M1 |
| `internal/store` | ממשקי המאגר, ובדיקות ההתאמה ב־`storetest` (שלמים) | M1 |
| `internal/store/memory` | המאגר בזיכרון | M1 |
| `internal/api` | ה־handlers, האימות והקליטה; ‏`server.go` שלם | M1 |
| `internal/admin` | יצירת דיירת | M1 |
| `migrations/` | ארבע מיגרציות goose לכל מודל הנתונים (שלמות) | M2 |
| `internal/db`, ‏`internal/testdb` | ה־pool, המיגרציות מתוך Go, ומסד זמני לבדיקות, מפרק 11 (שלמים) | M2 |
| `internal/store/postgres` | המאגר ב־PostgreSQL | M2 |
| `deploy/compose.yaml` | ‏PostgreSQL 18 בפרויקט compose בשם `hookline`, בפורט 5438 | M2 |
| `Makefile` | כל הפקודות: `test-unit`, ‏`test`, ‏`lint`, ‏`db-up`, ‏`migrate`, ‏`run-dev` | M1 |
| `docs/` | יקבל runbooks, תחקיר, מודל איומים ותוצאות הערכה | M8 עד M10 |

## צעדים

**1. ההתקנה (יום 22).** העתיקי את ה־starter ובדקי שהוא מתקמפל:

<div dir="ltr" align="left">

```bash
mkdir -p ~/senior-course/hookline
cp -r $COURSE/exercises/project-hookline/hookline-starter/. ~/senior-course/hookline/
cd ~/senior-course/hookline
go build ./...
go vet ./...
make test-unit
```

</div>

הבנייה וה־vet עוברים. ‏`make test-unit` נכשל בחבילות של אבני הדרך, וזה המצב שממנו מתחילים (פרק 37, סעיף 37.8).

**2. אבן דרך אחת בכל פעם.** כל אבן דרך מתוארת בסעיף שלה בפרק 38: המטרה, הקבצים, הצעדים, והפקודה שצריכה להוריק.

| אבן דרך | ימים | סעיף | הפקודה שמסיימת אותה |
| --- | --- | --- | --- |
| ‏M1: שלד ה־API | 22 ו־23 | 38.1 | `DATABASE_URL= go test -race ./internal/api/ ./internal/store/... ./internal/admin/` |
| ‏M2: PostgreSQL וקליטה אידמפוטנטית | 29 ו־30 | 38.2 | `go test -race -count=1 ./...` עם `DATABASE_URL` מוגדר |
| ‏M3 עד M10 | לפי הטבלה בפרק 37 | 38.3 עד 38.10 | יפורסמו עם אבני הדרך |

**3. קומיט לכל אבן דרך.** כשהפקודה ירוקה:

<div dir="ltr" align="left">

```bash
cd ~/senior-course
git add hookline
git commit -m "M1: API skeleton"
```

</div>

**4. המסד, מאבן הדרך השנייה.** כל עבודה עם PostgreSQL מתחילה ונגמרת כך:

<div dir="ltr" align="left">

```bash
cd ~/senior-course/hookline
make db-up
export DATABASE_URL=postgres://hookline:change-me@localhost:5438/hookline
go run ./cmd/hookline-admin migrate up
make test
make db-down
```

</div>

הסיסמה `change-me` היא סיסמת תרגול מקומית, ולא סוד. הפקודה `make db-down` מוחקת גם את ה־volume, כך שכל יום מתחיל
ממסד נקי.

## פלט צפוי

אחרי אבן הדרך הראשונה, בלי מסד:

<div dir="ltr" align="left">

```text
ok  	example.com/hookline/internal/api	2.213s
ok  	example.com/hookline/internal/store	1.848s
ok  	example.com/hookline/internal/store/memory	2.257s
ok  	example.com/hookline/internal/store/postgres	3.498s
ok  	example.com/hookline/internal/store/storetest	2.641s
ok  	example.com/hookline/internal/admin	3.010s
```

</div>

אחרי אבן הדרך השנייה, הרצה מלאה מול המסד מסתיימת כך (הזמנים אצלך יהיו אחרים):

<div dir="ltr" align="left">

```text
ok  	example.com/hookline/internal/admin	1.352s
ok  	example.com/hookline/internal/api	2.177s
ok  	example.com/hookline/internal/config	2.713s
ok  	example.com/hookline/internal/db	3.995s
ok  	example.com/hookline/internal/domain	1.727s
ok  	example.com/hookline/internal/page	2.376s
ok  	example.com/hookline/internal/platform	3.285s
ok  	example.com/hookline/internal/problem	4.484s
ok  	example.com/hookline/internal/store	4.788s
ok  	example.com/hookline/internal/store/memory	2.058s
ok  	example.com/hookline/internal/store/postgres	3.889s
ok  	example.com/hookline/internal/store/storetest	4.147s
ok  	example.com/hookline/migrations	4.469s
```

</div>

## רשימת בדיקה

- [ ] ה־starter מועתק לתיקייה `~/senior-course/hookline`, והבנייה והבדיקה הסטטית עוברות.
- [ ] ‏M1: הפקודה של סעיף 38.1.5 ירוקה עם `-race`.
- [ ] ‏M1: ‏`make run-dev`, ואז קליטה של אותו אירוע פעמיים מחזירה את אותו מזהה, וגוף אחר מחזיר 422.
- [ ] ‏M1: החוזה עובר את redocly, ו־`TestSpecCoversEveryRoute` ירוקה.
- [ ] ‏M2: המיגרציות עולות, יורדות עד 0 ועולות שוב.
- [ ] ‏M2: ‏`go test -race -count=1 ./...` ירוק מול המסד, ואף בדיקה לא מדלגת.
- [ ] ‏M2: ‏`make lint` נקי, כולל staticcheck.
- [ ] קומיט לכל אבן דרך שהסתיימה, ו־`make db-down` בסוף כל יום עם מסד.

## טעויות נפוצות ותיקונים

- **בדיקות ה־API עוברות בזיכרון ונכשלות מול המסד.** בדרך כלל מזהה שאינו UUID הגיע לשאילתה (שגיאה `22P02` בלוג),
  או זמן שלא הומר ל־UTC. בדקי ש־`domain.IsID` נקרא בכל נתיב, ושכל סריקה של זמן ממירה ל־UTC.
- **‏`go test` ירוק מיד אחרי שינוי במסד.** זו תוצאה שמורה מהרצה קודמת. הוסיפי `-count=1`, או השתמשי ב־`make test`.
- **‏`connection refused` על פורט 5438.** המסד לא למעלה, או שעלה בפרויקט compose אחר. הריצי `make db-up` מתוך
  `~/senior-course/hookline`, ובדקי ב־`docker compose -p hookline ps`.
- **‏`relation "events" does not exist` בהרצת השרת.** לא הרצת מיגרציות על המסד של השרת. הבדיקות יוצרות מסד זמני
  משלהן ומריצות עליו מיגרציות, אבל `make run` משתמש במסד `hookline` עצמו: `go run ./cmd/hookline-admin migrate up`.
- **‏panic עם `TODO: see chapter 38` בחבילה שחשבת שגמרת.** הבדיקות של חבילה אחת קוראות לקוד של אחרת: בדיקות ה־API
  קוראות למאגר בזיכרון. ה־panic אומר בדיוק איזו פונקציה עוד ריקה, בשורה הראשונה של ה־stack.

## רמזים

- התחילי כל אבן דרך מקריאת הבדיקות שלה. שם הבדיקה ושורת ה־`t.Fatalf` אומרים מה בדיוק מצופה.
- במאגר בזיכרון, מנעול אחד לכל המאגר מספיק. מהירות לא נמדדת כאן, נכונות כן.
- הפונקציות `pageParams` ו־`writePage` הן אותו עימוד מפרק 9, עם בדיקה אחת נוספת: מזהה בתוך ה־cursor חייב להיות UUID.
- ב־PostgreSQL, אל תבדקי אם מפתח אידמפוטנטיות קיים לפני ההכנסה. התחילי בהכנסה עם `ON CONFLICT DO NOTHING`,
  ורק כשלא חזרה שורה, קראי את השורה הקיימת.
- כשמנוי ל־endpoint של דיירת אחרת מחזיר 500 במקום 404, חפשי את הקוד `23503` בשגיאה. המפתח הזר המורכב עובד,
  ונשאר רק לתרגם אותו.

## הפתרון

המימוש המלא נמצא בתיקייה [`hookline/`](hookline/), עם הסבר קצר באנגלית בקובץ ה־README שלה. הוא עובד כי כל החלטה
על כפילות או בעלות יושבת במקום אחד שאי אפשר לעקוף. הדיירת נקבעת רק מהמפתח, וכל מתודת מאגר מסננת לפי מזהה
הדיירת שלה. הכפילות בקליטה מוכרעת באילוץ ייחודיות במסד, או במנעול אחד במאגר בזיכרון, ולא בבדיקה שקודמת להכנסה.
ובדיקות ההתאמה ב־`storetest` מבטיחות ששני המאגרים מתנהגים אותו דבר, ולכן ה־handlers לא צריכים לדעת איזה מהם
מאחוריהם.

## מה התרגיל מדגים

- איך מפרט עם דרישות ובדיקות הופך לעבודה של שבוע בכל פעם, שתמיד נגמרת בבדיקות ירוקות.
- ממשקים כגבול: אותן בדיקות API רצות על מאגר בזיכרון ועל PostgreSQL, בלי שינוי.
- אידמפוטנטיות אמיתית תחת מקביליות: אילוץ במסד, טרנזקציה אחת לאירוע ול־outbox, ובדיקה שסופרת שורות.
- הגנה מפני BOLA בשלוש שכבות: המפתח, המאגר והמסד.
- שימוש חוזר בקוד משלך מפרקים קודמים, במקום לכתוב הכול מחדש בכל פעם.

</div>
