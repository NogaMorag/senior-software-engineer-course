---
direction: rtl
lang: he
---

<div dir="rtl" lang="he">

# תרגיל פרק 28: `projects`, שירות קטן שמאובטח כמו שצריך

**מטרה:** לבנות את ההגנות של פרק 28 כחבילות Go עם בדיקות: אימות JWT מול JWKS, תיקון של BOLA, לקוח HTTP
מוגן מ־SSRF, מפתחות API מגובבים וסינון סודות מלוגים. ואז להפיק SBOM, לסרוק, ולמלא מודל איומים בשיטת STRIDE.
ארבע החבילות והתבנית של מודל האיומים יחזרו ב־Hookline, באבן הדרך M9.

**זמן משוער:** שני ימי לימוד. יום 83: כ־55 דקות בנייה. יום 84: כ־60 דקות.

**ידע מקדים:** פרק 28; פרק 3 (staticcheck); פרק 5 (הדגל `-race`); פרק 6 (handlers, middleware ו־`slog`);
פרק 10 (שאילתות לפי דיירת); פרק 19, סעיף 19.6 (סריקה ו־SBOM); govulncheck מפרק 27. התרגיל לא מייבא קוד
מתרגילים אחרים, ובסריקה של יום 84 הוא קורא את התיקייה `ch27/seeded` מתרגיל פרק 27 כדי להראות ממצא. צריך
Docker פועל.

## מה בתיקייה הזאת

| קובץ | תפקיד |
| --- | --- |
| `starter/` | נקודת ההתחלה: מודול Go שמתקמפל, עם `TODO` לפי ימים |
| `starter/internal/jwtauth/verifier.go` | אימות הטוקן וה־middleware; האפשרויות של ה־parser ו־`keyFor` של יום 83 |
| `starter/internal/jwtauth/jwks.go` | קריאת JWKS (מפתחות RSA ו־P-256), מלאה |
| `starter/internal/projects/` | המאגר וה־handlers, עם באג BOLA של יום 83 |
| `starter/internal/ssrf/ssrf.go` | מדיניות יעדים ולקוח מוגן; `CheckAddr`, `Control` וההפניות של יום 84 |
| `starter/internal/apikey/apikey.go` | מפתחות API; `Generate` ו־`Matches` של יום 84 |
| `starter/internal/redact/redact.go` | סינון סודות מלוגים; `ReplaceAttr` של יום 84 |
| `starter/cmd/projects/main.go` | הבינארי: תצורה ממשתני סביבה, JWKS מקובץ, שרת עם זמנים קצובים |
| `tests/internal/...` | הבדיקות הקנוניות (5 קבצים); עותקים זהים ב־`starter/` וב־`solution/` |
| `data/stride-template.md` | התבנית של מודל האיומים |
| `solution/` | הפתרון המלא |
| `solution/docs/threat-model.md` | מודל האיומים המלא של השירות |

## יום 83: JWT ו־BOLA

**1. העתקה ושמירה ראשונה.**

<div dir="ltr" align="left">

```bash
mkdir -p ~/senior-course/ch28
cp -r $COURSE/exercises/ch28-security/starter/. ~/senior-course/ch28/
cd ~/senior-course/ch28
go build ./... && go vet ./...
git add . && git commit -m "ch28: starter"
```

</div>

הרצה של כל הבדיקות עכשיו נגמרת ב־`panic: TODO(day 83)` וב־`TODO(day 84)`, וזה צפוי.

**2. קראי את `jwks.go`** ואת הבדיקה `TestParseJWKSRejects`. שימי לב מה נדחה: מפתח בלי `kid`, מפתח הצפנה,
RSA קצר מ־2048 ביט, ומפתח סימטרי.

**3. מלאי את האפשרויות של ה־parser** בפונקציה `NewVerifier`, ואת `keyFor`, לפי ההערות. ואז:

<div dir="ltr" align="left">

```bash
go test ./internal/jwtauth/ -v -run 'TestVerify|TestMiddleware|TestNewVerifier|TestParseJWKS'
```

</div>

כשהכול עובר, נסי ניסוי: מחקי זמנית את `jwt.WithValidMethods(...)` והריצי שוב. איזו שורה בטבלה נכשלת, ולמה
דווקא היא? החזירי את השורה.

**4. הריצי את בדיקות ה־BOLA ותראי אותן נכשלות:**

<div dir="ltr" align="left">

```bash
go test ./internal/projects/ -v
```

</div>

**5. תקני.** בקובץ `store.go`, החליפי את `ByID` ב־`Get(ctx, tenantID, id)` שמחזירה `ErrNotFound` לפרויקט
של דיירת אחרת, וסנני את `List` לפי דיירת. בקובץ `handler.go`, עדכני את הממשק `Store` ואת `get`, כך שהדיירת
תגיע מ־`claims.Tenant`. הריצי שוב את צעד 4, ושמרי (`git commit`).

## יום 84: SSRF, מפתחות, לוגים, SBOM ומודל איומים

**6. SSRF.** מלאי את `CheckAddr`, את `Control` ואת הבדיקה של ההפניות ב־`Client`:

<div dir="ltr" align="left">

```bash
go test -race ./internal/ssrf/ -v
```

</div>

הבדיקות מריצות שרת מקומי על 127.0.0.1 ובודקות שאף בקשה לא הגיעה אליו. אם `TestClientRefusesOddIPSpelling`
מסתיימת אצלך בשגיאת DNS ולא ב־`ErrBlocked`, זה בסדר: ה־resolver שלך לא מפרש את הכתיב הזה, והבדיקה דורשת רק
שגיאה ואפס בקשות.

**7. מפתחות API.** מלאי את `Generate` ואת `Matches`, והוסיפי לייבוא של הקובץ את שלוש החבילות האלה:

<div dir="ltr" align="left">

```text
crypto/rand
crypto/sha256
crypto/subtle
```

</div>

ואז:

<div dir="ltr" align="left">

```bash
go test -race ./internal/apikey/ -v
```

</div>

**8. לוגים.** מלאי את `ReplaceAttr`:

<div dir="ltr" align="left">

```bash
go test ./internal/redact/ -v
```

</div>

**9. הסוללה המלאה:**

<div dir="ltr" align="left">

```bash
gofmt -l .
go vet ./...
staticcheck ./...
go test -race -count=1 ./...
govulncheck ./...
```

</div>

**10. SBOM וסריקה.** הפקודות מושכות את התמונות ואת מאגרי החולשות בפעם הראשונה, וזה לוקח כמה דקות (הגרסאות
נכונות לאוקטובר 2026; ראי נספח ג):

<div dir="ltr" align="left">

```bash
cd ~/senior-course/ch28
docker run --rm -v "$PWD:/src:ro" anchore/syft:v1.54.0 scan dir:/src -o table
docker run --rm -v "$PWD:/src:ro" anchore/syft:v1.54.0 scan dir:/src -o cyclonedx-json > sbom.cdx.json
docker run --rm -v "$PWD:/out:ro" anchore/grype:v0.120.0 sbom:/out/sbom.cdx.json
docker run --rm -v "$HOME/senior-course/ch27/seeded:/src:ro" anchore/grype:v0.120.0 dir:/src
docker run --rm -v "$HOME/senior-course/ch27/seeded:/src:ro" aquasec/trivy:0.75.0 fs --scanners vuln /src
```

</div>

הסריקה של `ch27/seeded` מוצאת ממצאים רק אם עוד לא תיקנת שם את הגרסה של `x/text`. אם כבר תיקנת, סרקי את
העותק המקורי, `$COURSE/exercises/ch27-review/starter/seeded`. שמרי את הפלט של הסריקות, עם התאריך, בקובץ
`scan-results.txt`.

**11. מודל איומים.** העתיקי את התבנית ומלאי אותה לשירות `projects`: תרשים עם גבולות אמון, לפחות שלושה נכסים,
ושורה לכל אחת משש האותיות של STRIDE, עם ההגנה, הבדיקה שמוכיחה אותה והסיכון שנשאר:

<div dir="ltr" align="left">

```bash
mkdir -p docs
cp $COURSE/exercises/ch28-security/data/stride-template.md docs/threat-model.md
```

</div>

רק אחרי שמילאת, השווי עם `solution/docs/threat-model.md`. שמרי (`git commit`).

## פלט צפוי

סוף הסוללה המלאה (הזמנים אצלך יהיו אחרים):

<div dir="ltr" align="left">

```text
?   	example.com/ch28-security/cmd/projects	[no test files]
ok  	example.com/ch28-security/internal/apikey	1.888s
ok  	example.com/ch28-security/internal/jwtauth	1.862s
ok  	example.com/ch28-security/internal/projects	2.002s
ok  	example.com/ch28-security/internal/redact	2.297s
ok  	example.com/ch28-security/internal/ssrf	2.618s
No vulnerabilities found.
```

</div>

בצעד 4, לפני התיקון:

<div dir="ltr" align="left">

```text
--- FAIL: TestTenantCannotReadAnotherTenantsProject (0.00s)
    handler_test.go:51: globex reading acme's project: status 200, body "{\"id\":\"prj_a1\",\"name\":\"Acme roadmap\"}\n"; want 404
--- FAIL: TestListIsScopedToTenant (0.00s)
```

</div>

ובצעד 10, טבלת Syft כוללת את `github.com/golang-jwt/jwt/v5` בגרסה שב־`go.mod`, וב־Grype על `ch27/seeded`
מופיעות שתי חולשות על `golang.org/x/text`, בהרצה שבה נכתב הספר `GHSA-69ch-w2m2-3vjp` ו־`GO-2026-5970`; ב־Trivy אותן שתיים
כ־`CVE-2022-32149` ו־`CVE-2026-56852`. הסריקה של ה־SBOM של התרגיל מסתיימת ב־`No vulnerabilities found`. המספרים
והמזהים תלויים במאגר ביום הסריקה.

## רשימת בדיקה

- [ ] כל 19 השורות של `TestVerify` עוברות, וגם בדיקות ה־middleware, התצורה וה־JWKS.
- [ ] דיירת אחרת מקבלת 404 זהה לחסר, והרשימה מחזירה רק את הפרויקטים של הדיירת.
- [ ] הלקוח המוגן חוסם loopback, פרטי, link-local ומטא־דאטה, כולל שם שנפתר ל־loopback והפניה.
- [ ] המפתח מורכב מקידומת, מזהה וסוד, ורק הגיבוב שלו נשמר.
- [ ] אף סוד לא מופיע בפלט של בדיקת הלוגים, והערכים הרגילים נשארו.
- [ ] `gofmt`, `go vet`, staticcheck ו־govulncheck שקטים, ו־`go test -race` ירוק.
- [ ] ה־SBOM הופק ותוצאות הסריקה שמורות עם תאריך; מודל האיומים מלא.

## טעויות נפוצות ותיקונים

- **ב־`keyFor`, לבדוק את `t.Header["alg"]` במקום את `t.Method`.** הבדיקה של `kid of an EC key` נכשלת או
  עוברת במקרה. ה־parser כבר פענח את האלגוריתם לטיפוס; משווים מולו עם `switch t.Method.(type)`.
- **לשכוח את `jwt.WithExpirationRequired()`.** השורה `no exp at all` נכשלת: בלי האפשרות, טוקן בלי `exp`
  מתקבל.
- **לתקן את ה־BOLA ב־handler ולא במאגר.** הבדיקות עוברות, אבל `List` נשאר פתוח, וה־handler הבא ישכח. הדיירת
  שייכת לחתימה של `Get` ולסינון של `List`.
- **ב־`CheckAddr`, לבדוק את הרשימה לפני `Unmap`.** השורה `::ffff:127.0.0.1` עוברת את הבדיקה, כי היא לא
  בטווח של IPv4. קוראים ל־`ip.Unmap()` ראשון.
- **להחזיר ב־`CheckRedirect` שגיאה שלא עוטפת את `ErrBlocked`.** הבדיקה של ההפניות מדווחת `want ErrBlocked`.
  עוטפים עם `%w`.
- **הסריקה ב־Docker נכשלת עם `permission denied` על הקובץ שנכתב.** ה־SBOM נכתב דרך הפניה של ה־shell, ולא
  מתוך הקונטיינר; ודאי שאת בתיקייה `~/senior-course/ch28` ושהיא שלך.

## רמזים

- ב־`NewVerifier`, האפשרויות הן `WithValidMethods`, `WithIssuer`, `WithAudience`, `WithLeeway`,
  `WithExpirationRequired`, `WithIssuedAt` ו־`WithTimeFunc`, כל אחת עם השדה המתאים מ־`cfg`.
- ב־`keyFor`, ה־`kid` הוא `t.Header["kid"].(string)`, ושגיאה עטופה נראית כמו
  `fmt.Errorf("%w %q", ErrUnknownKID, kid)`.
- ב־`Control`, הכתובת מגיעה כ־`"ip:port"`, ו־`netip.ParseAddrPort` מפרק אותה, גם ל־IPv6 בסוגריים.
- ב־`Generate`, `base64.RawURLEncoding` נותן 43 תווים ל־32 בתים, בלי ריפוד.
- ב־`ReplaceAttr`, `a.Value.Kind()` מבחין בין קבוצה (`slog.KindGroup`) למחרוזת (`slog.KindString`).

## הפתרון

הפתרון נמצא ב־`solution/`. **למה הוא עובד כך:** ה־parser של `jwt/v5` אוכף את כל הדרישות מהתצורה, והטוקן לא
בוחר דבר, גם לא את האלגוריתם. ב־BOLA, הדיירת היא חלק מהחתימה של המאגר, ולכן הקומפיילר מוצא כל קורא שצריך לעדכן.
ב־SSRF, הבדיקה יושבת במקום היחיד שרואה את הכתובת האמיתית של כל חיבור, אחרי ה־DNS ואחרי כל הפניה. ומפתח API
עם 256 ביטים אקראיים מוגן היטב בגיבוב מהיר, כי אין מה לנחש.

## מה התרגיל מדגים

- אימות JWT עם בדיקה טבלאית שבה לכל דרישה יש טוקן שמפר רק אותה.
- באג BOLA, הבדיקה שתופסת אותו, ותיקון מבני שהקומפיילר אוכף.
- הגנה מ־SSRF בשתי שכבות, שנבדקת מול שרת מקומי אמיתי, כולל DNS והפניות.
- מפתחות API שנשמרים כגיבוב בלבד, וסינון סודות מלוגים עם בדיקה לשני הכיוונים.
- SBOM, סריקת תלויות ומודל איומים, כחלק מהעבודה ולא כנספח.

</div>
