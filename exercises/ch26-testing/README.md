---
direction: rtl
lang: he
---

<div dir="rtl" lang="he">

# תרגיל פרק 26: `ch26-testing`, בדיקות בכל רמה ובדיקת עומס

**מטרה:** לכתוב את הבדיקות שאי אפשר לכתוב עם `testing` לבד: בדיקות אינטגרציה שמרימות לעצמן PostgreSQL,
בדיקת חוזה שרצה משני צידי ה־webhook, בדיקות תכונה ל־backoff ולחתימה, fuzzing שמוצא קריסה אמיתית, ובדיקת
עומס עם ספים מול `shorty`, שאת התוצאות שלה את קוראת ורושמת בעצמך.

**זמן משוער:** שני ימי לימוד. יום 80: כ־65 דקות בנייה (testcontainers, החוזה, בדיקות התכונה). יום 81: כ־65 דקות
בנייה (fuzzing, k6, קריאת התוצאות).

**ידע מקדים:** פרק 26; פרק 3 (staticcheck); פרק 7 (uv, ruff ו־mypy, לצרכן ב־Python). Docker Desktop פועל
(פרק 2, סעיף 2.4), ו־k6 מותקן (פרק 2, סעיף 2.10). התרגיל מעתיק
לתוכו, בלי לייבא, ארבעה דברים מפרקים קודמים, כדי שיעמוד בפני עצמו:

| מה | מאיפה | איפה בתרגיל |
| --- | --- | --- |
| ה־repository של ההזמנות, ב־pgx | פרק 10 (`exercises/ch10-orders-db/solution`) | `orders/` |
| פונקציית ה־backoff עם פיזור מלא | פרק 14 (`exercises/ch14-jobqueue/solution/backoff`) | `backoff/` |
| שירות קיצור הכתובות, בלי הבדיקות שלו | פרק 6 (`exercises/ch06-shorty/solution`) | `shorty/` |
| וקטורי החתימה המשותפים | פרק 8 (`exercises/ch08-receiver/data/signature-vectors.json`) | `signature/testdata/` |

החותם ב־Go, בחבילה `signature`, נכתב לתרגיל הזה לפי אותם כללים בדיוק כמו המקבל ב־Python מפרק 8, ועובר את
אותם וקטורים. הצרכן ב־Python הוא פרויקט uv קטן (פרק 7). הגרסאות של testcontainers-go, ‏rapid, הבודק של JSON
Schema ו־`jsonschema` מוצמדות ב־`go.mod` וב־`consumer/pyproject.toml` (נכון לאוקטובר 2026, ראי נספח ג).

## מה בתיקייה הזאת

| קובץ | תפקיד |
| --- | --- |
| `starter/go.mod`, `starter/go.sum` | המודול `example.com/ch26-testing`, עם כל התלויות כבר מוצמדות |
| `starter/internal/pgtest/pgtest.go` | העוזר של testcontainers: ‏`Main` ו־`NewDatabase` עם `TODO` |
| `starter/orders/` | ה־repository מפרק 10 (שלם), ובדיקות האינטגרציה שלו עם בונה נתונים |
| `starter/backoff/backoff.go`, `backoff_test.go` | ה־backoff מפרק 14 והבדיקות הטבלאיות שלו (שלמים) |
| `starter/backoff/property_test.go` | שלוש בדיקות תכונה, כל אחת `TODO` |
| `starter/signature/signature.go` | ‏`Sign`, ‏`Verify`, ‏`ParseHeader`; עובר את כל הווקטורים, ובכל זאת יש בו באג |
| `starter/signature/vectors_test.go` | הבדיקה מול הווקטורים המשותפים (שלמה) |
| `starter/signature/property_test.go`, `fuzz_test.go` | בדיקות התכונה ויעד ה־fuzzing, ‏`TODO` |
| `starter/contract/event.go` | צד היצרן: ‏`New` ו־`Body` של גוף ה־webhook (שלם) |
| `starter/contract/webhook-event.schema.json` | החוזה. כרגע הוא מקבל הכול; את כותבת אותו |
| `starter/contract/testdata/examples/` | דוגמאות תקינות ושגויות, ששני הצדדים בודקים |
| `starter/contract/contract_test.go` | בדיקת החוזה מצד Go (שלמה) |
| `starter/consumer/` | צד הצרכן: פרויקט uv עם `parse_event` ו־`tests/test_contract.py` (שלמים) |
| `starter/shorty/` | ‏`shorty` מפרק 6, יעד בדיקת העומס |
| `starter/loadtest/shorty.js` | סקריפט k6 שרץ, אבל בלי תרחישים וספים; ‏`TODO` |
| `tests/` | הבדיקות הקנוניות; עותקים זהים ב־`starter/` וב־`solution/` |
| `solution/` | הפתרון המלא, כולל `signature/testdata/fuzz/FuzzParseHeader/`, הקריסה שנשמרה |

## יום 80: testcontainers, חוזה, בדיקות תכונה

**1. העתקה לתיקיית העבודה.**

<div dir="ltr" align="left">

```bash
mkdir -p ~/senior-course/ch26
cp -r $COURSE/exercises/ch26-testing/starter/. ~/senior-course/ch26/
cd ~/senior-course/ch26
go build ./... && go vet ./...
```

</div>

שתי הפקודות שקטות. הבדיקות עוד לא: כל `TODO` נכשל בהודעה שמפנה לצעד שלו כאן.

**2. ‏`pgtest.Main`.** פתחי את `internal/pgtest/pgtest.go`. ‏`Main` נקראת מ־`TestMain` של כל חבילה שצריכה מסד, ומרימה
קונטיינר אחד לכל קובץ הבדיקות המהודר. הסדר בהערה של ה־`TODO` חשוב: ה־`defer` שמוחק את הקונטיינר נכתב לפני
בדיקת השגיאה של `postgres.Run`, כי קונטיינר שעלה ולא הגיע למוכנות עדיין קיים. מלאי את הייבוא לפי מה שאת
משתמשת בו; ‏`go vet` יגיד אם שכחת משהו.

**3. ‏`pgtest.NewDatabase`.** מסד ריק לכל בדיקה, בשם ייחודי, שנמחק ב־`t.Cleanup`. כך בדיקות שקוראות לו יכולות
לרוץ במקביל בלי לראות אחת את הנתונים של השנייה. הריצי:

<div dir="ltr" align="left">

```bash
go test -race -count=1 -v ./orders/ 2>&1 | grep -E '^(--- |ok|FAIL)'
docker ps --filter label=org.testcontainers=true
```

</div>

הפעם הראשונה מורידה את ה־image של PostgreSQL ואת זה של Ryuk, ולוקחת דקה או שתיים. אחרי שהבדיקות נגמרות,
הפקודה השנייה צריכה להראות רשימה ריקה, או רק את Ryuk בשניות שלפני שהוא נסגר.

**4. בונה הנתונים.** קראי את `orders/builder_test.go` והשווי לפונקציה `setup` של פרק 10. שם כל הנתונים הם
מחרוזת SQL אחת. כאן, הבדיקה `TestRecentOrdersOnlyThatCustomer` אומרת רק את מה שחשוב לה: של מי ההזמנה. רשמי
בפתק מה היה צריך להשתנות בשני הקבצים אם הייתה נוספת לטבלה `orders` עמודה חובה חדשה.

**5. החוזה.** ‏`contract/webhook-event.schema.json` מקבל כרגע כל אובייקט. כתבי אותו, בטיוטה 2020-12:

| מאפיין | חובה | הכלל |
| --- | --- | --- |
| `id` | כן | מחרוזת, `format: uuid` |
| `type` | כן | מחרוזת עד 100 תווים, מילים קטנות מופרדות בנקודות, לפחות שתיים (`invoice.paid`) |
| `created_at` | כן | מחרוזת, `format: date-time` |
| `data` | כן | אובייקט |
| כל מאפיין אחר | לא | מותר: הצרכן חייב להתעלם ממה שהוא לא מכיר |

הריצי את שני הצדדים:

<div dir="ltr" align="left">

```bash
go test -count=1 ./contract/
cd consumer && uv sync --frozen && uv run pytest -q; cd ..
```

</div>

**6. הפער בין השפות.** אם הבדיקה של Go עוברת והבדיקה של Python נכשלת על `type-trailing-newline.json`, מצאת
בדיוק את מה שבדיקת חוזה קיימת בשבילו. קראי את סעיף 26.3.3 ותקני את הסכמה, לא את הבודק.

**7. בדיקות תכונה ל־backoff.** פתחי את `backoff/property_test.go` וכתבי את שלוש התכונות של סעיף 26.4.2: גבולות,
הכפלה עד התקרה והישארות בה, ופיזור סביר. הריצי עם יותר בדיקות מברירת המחדל:

<div dir="ltr" align="left">

```bash
go test -count=1 -run Property ./backoff/ -rapid.checks=2000
```

</div>

**8. בדיקות תכונה לחתימה.** ב־`signature/property_test.go`: הלוך וחזור (מה ש־`Sign` מייצר, `Verify` מקבל), ושינוי
של בית אחד בגוף או של הסוד נותן תמיד `signature_mismatch`.

**9. ניסוי: לשבור בכוונה.** בקובץ `backoff/backoff.go`, החליפי זמנית את `c > p.Cap/2` ב־`c > p.Cap`, והריצי את
צעד 7. רשמי בפתק את הקלט המוקטן ש־rapid מדפיס, ואת הדגל לשחזור. החזירי את השורה.

**10. התחייבות.**

<div dir="ltr" align="left">

```bash
cd ~/senior-course
git add ch26
git commit -m "Chapter 26 day 1: testcontainers helper, webhook contract, property tests"
```

</div>

## יום 81: fuzzing, k6, ותוצאות

**1. יעד ה־fuzzing.** ב־`signature/fuzz_test.go`, הוסיפי כזרעים את כל הכותרות מקובץ הווקטורים, וכתבי ב־`f.Fuzz`
את מה שחייב להתקיים לכל קלט (ההערה בקובץ מפרטת). הריצי קודם בלי fuzzing, ואז עם:

<div dir="ltr" align="left">

```bash
go test -count=1 -run FuzzParseHeader -v ./signature/
go test -run '^$' -fuzz '^FuzzParseHeader$' -fuzztime 60s ./signature/
```

</div>

הראשונה מריצה רק את הזרעים, ועוברת. השנייה מייצרת קלטים חדשים, ואמורה למצוא קריסה בתוך שניות.

**2. לשמור את הקריסה.** ה־fuzzer כתב קובץ תחת `signature/testdata/fuzz/FuzzParseHeader/`. הריצי עכשיו
`go test ./signature/` רגיל: הוא נכשל, כי הקובץ הזה הוא כבר חלק מהזרעים. תקני את `ParseHeader` (רמז: מה קורה
לפריט באורך תו אחד?), ודאי שהבדיקה הרגילה עוברת, והריצי שוב fuzzing של 60 שניות. הוסיפי את הקובץ ל־Git.

**3. ‏`shorty` מקומי.** בחלון נפרד, מתוך `~/senior-course/ch26`:

<div dir="ltr" align="left">

```bash
go build -o bin/shorty ./shorty/cmd/shorty
SHORTY_BASE_URL=http://127.0.0.1:8080 SHORTY_LOG_LEVEL=warn ./bin/shorty
```

</div>

‏`SHORTY_LOG_LEVEL=warn` מכבה את לוג הגישה. אחרת כל בקשה כותבת שורה לטרמינל, והטרמינל הופך לחלק ממה שנמדד.

**4. הסקריפט של ההתחלה.** הריצי את הסקריפט כמו שהוא, והסתכלי על `http_reqs` ועל השורה של הבדיקות:

<div dir="ltr" align="left">

```bash
k6 run loadtest/shorty.js
```

</div>

שתי שאלות לפתק: כמה בקשות נשלחו בפועל, ולאיזה שרת הלכה חצי מהן. (רמז: אחרי ה־302.)

**5. תרחישים וספים.** כתבי את ה־`TODO` של `options` לפי סעיף 26.6: שלושה תרחישים בקצב קבוע של הגעה, תגית
`phase`, ‏`name` לכל בקשה, וספים. הריצי עם ייצוא של הסיכום:

<div dir="ltr" align="left">

```bash
mkdir -p results
k6 run -e RATE=200 --summary-export=results/rate-200.json loadtest/shorty.js | tee results/rate-200.txt
echo "k6 exit code: ${PIPESTATUS[0]}"
```

</div>

**6. מדרגות.** הריצי שוב עם `RATE=1000` ועם `RATE=3000`, לקבצים משלהם. בכל הרצה רשמי בטבלה. ‏`shorty` רץ על אותו
מחשב כמו k6, ולכן המספרים מתארים את המחשב שלך באותו רגע, לא שירות בייצור (סעיף 26.7.4):

| ‏RATE | ‏`http_reqs` לשנייה | ‏avg ‏(ms) | ‏p95 ‏(ms) | ‏p99 ‏(ms) | ‏max ‏(ms) | ‏`dropped_iterations` | ספים עברו? |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 200 | ___ | ___ | ___ | ___ | ___ | ___ | ___ |
| 1000 | ___ | ___ | ___ | ___ | ___ | ___ | ___ |
| 3000 | ___ | ___ | ___ | ___ | ___ | ___ | ___ |

**7. הקריאה שלך.** מתחת לטבלה, בשלוש עד חמש שורות: איפה הממוצע הסתיר את הזנב, האם היו איטרציות שנזרקו ומה
זה אומר על הקצב שהשרת באמת קיבל, ואיזה סף היית קובעת ל־`shorty` אחרי ההרצות האלה, ולמה.

**8. התחייבות.**

<div dir="ltr" align="left">

```bash
cd ~/senior-course
git add ch26
git commit -m "Chapter 26 day 2: fuzz crasher kept as a test, k6 thresholds and results"
```

</div>

## פלט צפוי

בדיקות ה־Go ירוקות, עם `-race`. הזמנים שלך יהיו אחרים:

<div dir="ltr" align="left">

```text
ok  	example.com/ch26-testing/backoff	…s
ok  	example.com/ch26-testing/contract	…s
?   	example.com/ch26-testing/internal/pgtest	[no test files]
ok  	example.com/ch26-testing/orders	…s
ok  	example.com/ch26-testing/signature	…s
?   	example.com/ch26-testing/shorty/cmd/shorty	[no test files]
?   	example.com/ch26-testing/shorty/internal/app	[no test files]
?   	example.com/ch26-testing/shorty/internal/links	[no test files]
?   	example.com/ch26-testing/shorty/internal/platform	[no test files]
```

</div>

בצד ה־Python, 18 בדיקות עוברות (`18 passed`). ה־fuzzer, על ה־`ParseHeader` של ההתחלה, מסיים בשורה כזו (השם
של הקובץ הוא גיבוב של התוכן, ולכן יכול להיות אחר אם הוא מצא קלט אחר):

<div dir="ltr" align="left">

```text
    Failing input written to testdata/fuzz/FuzzParseHeader/b074d9d373d06c31
    To re-run:
    go test -run=FuzzParseHeader/b074d9d373d06c31
```

</div>

ב־k6, שורת הספים בראש הסיכום נראית כך, עם המספרים שלך; ‏✓ לכל סף שעבר, ‏✗ לכל סף שנכשל, וקוד יציאה 99 אם
אחד מהם נכשל:

<div dir="ltr" align="left">

```text
  █ THRESHOLDS

    checks{phase:measure}
    ✓ 'rate>0.99' rate=___

    http_req_duration{phase:measure,name:resolve}
    ✓ 'p(95)<25' p(95)=___
    ✓ 'p(99)<100' p(99)=___
```

</div>

## רשימת בדיקה

- [ ] ‏`go test -race -count=1 ./...` ירוק, בלי דילוגים, וקונטיינר אחד בלבד עלה לחבילת `orders`.
- [ ] אחרי הבדיקות, אין קונטיינרים עם התווית `org.testcontainers=true`.
- [ ] הסכמה דוחה את כל הדוגמאות השגויות בשני הצדדים, כולל השורה החדשה בסוף ה־`type`.
- [ ] שלוש התכונות של ה־backoff עוברות עם `-rapid.checks=2000`, ובפתק רשום הקלט המוקטן מצעד 9.
- [ ] קובץ הקריסה של ה־fuzzer נמצא ב־Git, ו־`go test ./signature/` מריץ אותו.
- [ ] ‏`loadtest/shorty.js` משתמש ב־`constant-arrival-rate`, ‏`maxRedirects: 0`, ספים על p95, ‏p99, שיעור שגיאות ובדיקות.
- [ ] טבלת התוצאות מלאה בשלוש שורות, עם הקריאה שלך מתחתיה ועם קובצי `results/`.
- [ ] ‏`gofmt -l .` ריק, ‏`go vet ./...` ו־`staticcheck ./...` נקיים; ב־`consumer`, ‏ruff ו־mypy נקיים.

## טעויות נפוצות ותיקונים

- **‏`pgtest: start PostgreSQL ... Cannot connect to the Docker daemon`.** Docker Desktop לא רץ, או שהשילוב
  עם ההפצה של WSL כבוי. מפעילים אותו ובודקים עם `docker ps`. כדי להריץ רק את הבדיקות המהירות בינתיים,
  מוסיפים `-short`.
- **קונטיינרים של `postgres:18.6` נשארים אחרי `Ctrl-C`.** הבדיקות נעצרו לפני ה־`defer`, ו־Ryuk כבוי
  (`TESTCONTAINERS_RYUK_DISABLED=true` בסביבה או ב־`~/.testcontainers.properties`). מסירים את ההגדרה, ומוחקים
  את מה שנשאר עם `docker rm -f` על הרשימה של `docker ps -q --filter label=org.testcontainers=true`.
- **‏`invalid/id-not-uuid.json` עובר בצד אחד.** הבודק לא אוכף `format`. ב־Go חסר `c.AssertFormat()`, וב־Python
  חסר `format_checker`. בלי החבילה `rfc3339-validator`, ‏`jsonschema` מדלג על `date-time` בשקט, ולכן היא ב־`pyproject.toml`.
- **בדיקת התכונה נכשלת פעם אחת ועוברת בהרצה הבאה.** rapid מגריל זרע חדש בכל הרצה. מעתיקים את הדגל
  `-rapid.seed=...` מההודעה ומריצים איתו עד שהבאג מתוקן; קובץ הכשל ב־`testdata/rapid/` נשמר רק מקומית.
- **ב־k6 כל הבדיקות עוברות, אבל `http_req_duration` גבוה פי כמה.** k6 עוקב אחרי ה־302 ל־`example.com`, והמדידה
  כוללת שרת זר ברשת. ‏`maxRedirects: 0` ב־`options`.
- **‏`dropped_iterations` גדול מאפס, והספים עוברים.** אין מספיק VUs לקצב, ו־k6 לא התחיל חלק מהבקשות; השרת
  ראה פחות עומס ממה שביקשת. מגדילים את `maxVUs`, ומשאירים את הסף על `dropped_iterations` כדי שזה ייכשל.

## רמזים

- ‏`postgres.Run` מחזיר `*PostgresContainer`, ו־`ConnectionString(ctx, "sslmode=disable")` נותן את הכתובת.
  ‏`net/url` מחליף בה את שם המסד בלי לפרק מחרוזות ביד.
- בתכונה של הגבולות, ‏`backoff.Rand` הוא ממשק. מימוש שמושך כל מספר מ־`rapid.Int64Range(0, n-1)` נותן ל־rapid
  לחפש בעצמו את הקצוות, 0 ו־`n-1`.
- ב־`rapid.OneOf` אפשר לשלב טווח קטן של ניסיונות, שבו ההכפלה קורית, עם טווח ענק, שבו רק התקרה חשובה.
- בבדיקת הפיזור, הממוצע של n הגרלות אחידות על `[0, c]` סוטה ב־`c/sqrt(12n)` בערך. סף של עשר סטיות כאלה לא
  נכשל במקרה.
- ב־k6, הסף `'dropped_iterations{scenario:resolve}': ['count==0']` מתייחס רק לתרחיש המדוד, בלי החימום.

## הפתרון

‏`solution/`, ובמיוחד `internal/pgtest/pgtest.go`, ‏`contract/webhook-event.schema.json`, קובצי ה־`property_test.go`,
‏`signature/fuzz_test.go` עם הקריסה ב־`testdata/fuzz/`, ו־`loadtest/shorty.js`. **למה זה עובד:** קונטיינר אחד
לחבילה ומסד ריק לכל בדיקה נותנים בידוד מלא בעלות של שנייה או שתיים לחבילה, לא לבדיקה. החוזה הוא קובץ אחד
עם דוגמאות, ששני הצדדים מריצים בלי לייבא זה את זה, ולכן הוא תופס גם הבדלים בין הבודקים עצמם. בדיקות התכונה
מתארות את מה שנכון לכל קלט במקום לרשום קלטים, ו־rapid מקטין כל כשל לדוגמה הקטנה ביותר. וסקריפט ה־k6 בנוי
במודל פתוח עם ספים, ולכן שרת איטי מופיע כזמן תגובה גבוה וככישלון, ולא כקצב נמוך יותר בשקט.

## מה התרגיל מדגים

- בדיקות אינטגרציה מול PostgreSQL אמיתי, בלי הגדרה מחוץ ל־Go, ובלי שנשארים קונטיינרים.
- בונה נתונים לבדיקות, שבו הבדיקה אומרת רק את מה שחשוב לה.
- בדיקת חוזה משני צידי השפה, ופער אמיתי בין Go ל־Python שרק היא מוצאת.
- בדיקות תכונה עם מחולל, הקטנה ושחזור לפי זרע, ו־fuzzing שהקריסה שלו הופכת לבדיקה קבועה.
- בדיקת עומס במודל פתוח, עם חימום, תגיות וספים, וקריאה של תוצאות לפי אחוזונים.

</div>
