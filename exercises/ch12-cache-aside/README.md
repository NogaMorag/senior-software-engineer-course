---
direction: rtl
lang: he
---

<div dir="rtl" lang="he">

# תרגיל פרק 12: `cache-aside`

**מטרה:** לבנות מטמון cache-aside ב־Redis מעל מאגר מוצרים ב־PostgreSQL: מסלול קריאה עם TTL ופיזור אקראי,
ומסלול כתיבה שמוחק אחרי ה־commit. אחר כך למדוד התפרצות מטמון על מפתח קר, לתקן אותה עם singleflight בחבילה
גנרית `internal/cache` שהפרויקט המסכם ישתמש בה, ולהוסיף מונה `INCR` ומגביל קצב קטן ב־Lua. כל הבדיקות רצות
מול קונטיינרים אמיתיים.

**זמן משוער:** שני ימי לימוד. יום 31: כ־55 דקות בנייה (הקונטיינרים, `Jittered`, מסלול הקריאה והכתיבה
הידני). יום 32: כ־55 דקות בנייה (העוזר עם singleflight, מדידת ההתפרצות, המונה והסקריפט, הרצה מלאה).

**ידע מקדים:** פרק 12. מפרק 2: Docker Desktop פועל, Go ו־staticcheck מותקנים, והמשתנה `COURSE` מצביע על
המאגר של הקורס. מפרק 10: pgx ו־`pgxpool`. התרגיל עצמאי: המאגר `store.go` וטבלת `products` משקפים את
התרגיל של פרק 10 (אותה טבלה, אותו `Config` ואותה בדיקה בסכמה פרטית), אבל הקוד כאן, ושום דבר לא מיובא
מתיקיית פרק 10.

## מה בתיקייה הזאת

| קובץ | תפקיד |
| --- | --- |
| `data/compose.yaml` | PostgreSQL ו־Redis, פרויקט `ch12`, פורטים 5412 ו־6312 על localhost בלבד |
| `starter/schema.sql` | טבלת `products` מפרק 10 עם `updated_at`, ו־2,000 שורות |
| `starter/store.go` | המאגר: `GetProduct` שסופר קריאות, `UpdatePrice` בטרנזקציה; מוכן |
| `starter/productcache.go` | המטמון הידני; ב־`Get` וב־`UpdatePrice` יש `TODO` של יום 31 |
| `starter/internal/cache/ttl.go` | הפונקציה `Jittered`, עם `TODO` של יום 31 |
| `starter/internal/cache/cache.go` | העוזר הגנרי: `Store`, `Options`, `Stats`; ב־`Cache.Get` יש `TODO` של יום 32 |
| `starter/internal/cache/redis.go` | המתאם `RedisStore` מעל go-redis; מוכן |
| `starter/products.go` | `CachedProducts`, המוצרים דרך העוזר; מוכן |
| `starter/counters.go` | הסקריפט מוכן; ב־`IncrViews` וב־`Allow` יש `TODO` של יום 32 |
| `starter/cmd/stampede/main.go` | ההדגמה: N קוראות על מפתח קר, בלי singleflight ועם; מוכנה |
| `tests/catalog_test.go` | מול שני השירותים: פספוס ופגיעה, אי־תוקף, לא נמצא, התפרצות, מונה, מגביל |
| `tests/internal/cache/cache_test.go` | העוזר מול Redis: פיזור, TTL, טעינה אחת, המרוץ, Redis שנפל |
| `solution/` | הפתרון המלא, אותו מבנה |

קובצי הבדיקה ב־`starter/` וב־`solution/` זהים בדיוק לאלה שב־`tests/`. בדיקה שלא מוצאת את `DATABASE_URL` או
את `REDIS_ADDR` מדלגת עם הודעה; הבדיקות `TestJittered` ו־`TestStoreDownFallsBackToSource` רצות גם בלעדיהם.

## יום 31: הקונטיינרים, TTL, ומסלול הקריאה והכתיבה

**1. העתיקי את קוד ההתחלה ואת קובץ ה־compose.** כל הפקודות מכאן והלאה רצות מתוך תיקיית הפרק.

<div dir="ltr" align="left">

```bash
mkdir -p ~/senior-course/ch12
cd ~/senior-course/ch12
cp -R "$COURSE/exercises/ch12-cache-aside/starter/." .
mkdir -p data
cp "$COURSE/exercises/ch12-cache-aside/data/compose.yaml" data/
printf 'bin/\n' > .gitignore
go build ./... && go vet ./... && echo builds
```

</div>

הפקודה האחרונה צריכה להדפיס `builds`. הגרסאות של go-redis, pgx ו־`golang.org/x/sync` כבר מקובעות בקובץ
`go.mod`, ו־`go build` מוריד אותן בפעם הראשונה.

**2. הפעילי את הקונטיינרים וקבעי את משתני הסביבה.** הדגל `--wait` מחכה לבדיקות הבריאות של שניהם.

<div dir="ltr" align="left">

```bash
docker compose -p ch12 -f data/compose.yaml up -d --wait
export DATABASE_URL=postgres://course:change-me@localhost:5412/catalog
export REDIS_ADDR=localhost:6312
docker compose -p ch12 -f data/compose.yaml exec cache redis-cli PING
```

</div>

התשובה האחרונה היא `PONG`. משתני הסביבה חיים רק בטרמינל הזה; בטרמינל חדש הריצי את שתי שורות ה־`export`
שוב.

**3. נסי את מבני הנתונים בעצמך.** פתחי את `redis-cli` והקלידי את הפקודות מסעיף 12.1, כולל `SET NX PX`
פעמיים ו־`CONFIG GET maxmemory*`. ודאי שהמדיניות היא `allkeys-lru`, כפי שקובץ ה־compose קבע.

<div dir="ltr" align="left">

```bash
docker compose -p ch12 -f data/compose.yaml exec cache redis-cli
```

</div>

**4. קראי את `store.go` ואת `schema.sql`.** שימי לב לשלושה דברים:

- המונה `calls`, שגדל בכל קריאה למסד.
- השדה `Latency`, שמוסיף `pg_sleep` לשאילתה.
- הפונקציה `pgx.BeginFunc` במסלול הכתיבה, שמחייבת את הטרנזקציה לפני שהיא חוזרת.

**5. מלאי את `Jittered`** בקובץ `internal/cache/ttl.go`, והריצי את הבדיקה שלה:

<div dir="ltr" align="left">

```bash
go test -run TestJittered -v ./internal/cache/
```

</div>

**6. מלאי את `ProductCache.Get`** בקובץ `productcache.go`, לפי ההערה שבגוף הפונקציה. הרחבת Go ב־VS Code
מוסיפה ייבואים בשמירה; אלה שתצטרכי:

- `encoding/json`
- `errors`
- `example.com/ch12-cache-aside/internal/cache`

<div dir="ltr" align="left">

```bash
go test -race -run 'TestMissThenHit/by_hand|TestNotFoundIsNotCached/by_hand' -v .
```

</div>

**7. מלאי את `ProductCache.UpdatePrice`:** קודם המאגר, ורק אחרי שהוא החזיר `nil`, מחיקת המפתח.

<div dir="ltr" align="left">

```bash
go test -race -run 'TestWriteInvalidates/by_hand' -v .
```

</div>

הסינון חשוב היום: תת־הבדיקות `with_helper` קוראות לעוזר, שעדיין עוצר ב־panic של TODO.

**8. ראי את המטמון מבפנים.** הבדיקות מוחקות את המפתחות שלהן, ולכן כתבי ערך דרך מוצר אחד ובדקי אותו. התוכנית
`cmd/stampede` עוד לא עובדת, אז השתמשי ב־`redis-cli` ישירות, ואחר כך הסתכלי במוני השרת:

<div dir="ltr" align="left">

```bash
docker compose -p ch12 -f data/compose.yaml exec cache redis-cli SET product:1 test EX 70
docker compose -p ch12 -f data/compose.yaml exec cache redis-cli TTL product:1
docker compose -p ch12 -f data/compose.yaml exec cache redis-cli INFO stats | grep -E 'keyspace_(hits|misses)|evicted_keys'
```

</div>

**9. סוף יום.** עצרי את הקונטיינרים בלי למחוק את הנתונים, ושמרי את העבודה:

<div dir="ltr" align="left">

```bash
docker compose -p ch12 -f data/compose.yaml stop
cd ~/senior-course
git add ch12
git commit -m "ch12 day 31: cache-aside by hand"
```

</div>

## יום 32: singleflight, התפרצות, מונים ומגביל

**10. הפעילי שוב.** מתוך תיקיית הפרק, עם שתי שורות ה־`export` מצעד 2:

<div dir="ltr" align="left">

```bash
cd ~/senior-course/ch12
docker compose -p ch12 -f data/compose.yaml up -d --wait
```

</div>

**11. מלאי את `Cache.Get`** בקובץ `internal/cache/cache.go`, לפי ההערה. זכרי את הסדר מסעיף 12.4.3: המילוי
למטמון קורה בתוך הפונקציה שמועברת ל־`Do`, לפני שהיא חוזרת. הוסיפי את הייבוא `encoding/json`.

<div dir="ltr" align="left">

```bash
go test -race -v ./internal/cache/
```

</div>

**12. מדדי את ההתפרצות.** קודם התוכנית, ואחר כך הבדיקה שמוודאת קריאה אחת. רשמי את שני המספרים בפתק:

<div dir="ltr" align="left">

```bash
go run ./cmd/stampede
go test -race -run TestStampede -v .
```

</div>

**13. מלאי את שתי הפונקציות בקובץ `counters.go`**, והוסיפי את הייבוא `fmt`:

<div dir="ltr" align="left">

```bash
go test -race -run 'TestIncrViewsIsAtomic|TestAllowFixedWindow' -v .
```

</div>

**14. הכול, ואז שלושת הכלים.**

<div dir="ltr" align="left">

```bash
go test -race -count=1 ./...
gofmt -l .
go vet ./...
staticcheck ./...
```

</div>

**15. יחס הפגיעות של השרת.** הריצי את ההדגמה פעמיים ובדקי שוב את המונים מצעד 8. החשבון: פגיעות חלקי
פגיעות ועוד פספוסים. רשמי את התוצאה שלך, ואת ההסבר למה היא שונה מהשורה `hit ratio` שהעוזר מדפיס.

**16. נקי ושמרי.** הדגל `-v` מוחק גם את ה־volume של PostgreSQL:

<div dir="ltr" align="left">

```bash
docker compose -p ch12 -f data/compose.yaml down -v
cd ~/senior-course
git add ch12
git commit -m "ch12 day 32: singleflight, counters, rate limit script"
```

</div>

## פלט צפוי

אחרי צעד 12, ההדגמה. מספרי הקריאות לא אמורים להשתנות; הזמנים אצלך יהיו אחרים:

<div dir="ltr" align="left">

```text
cold key, 50 concurrent readers, 200ms per query
without singleflight:   50 database calls, 1.041s
with singleflight:       1 database calls, 204ms
helper stats: hits 50, misses 50, loads 1, hit ratio 0.50
```

</div>

והשורות החשובות מבדיקת ההתפרצות:

<div dir="ltr" align="left">

```text
=== RUN   TestStampede/by_hand
    catalog_test.go:189: 30 concurrent misses without singleflight: 30 database calls
=== RUN   TestStampede/with_helper
    catalog_test.go:197: 30 concurrent misses with singleflight: 1 database calls
--- PASS: TestStampede (0.98s)
```

</div>

ובצעד 14, כשהקונטיינרים פועלים (הזמנים אצלך יהיו אחרים):

<div dir="ltr" align="left">

```text
ok  	example.com/ch12-cache-aside	4.466s
?   	example.com/ch12-cache-aside/cmd/stampede	[no test files]
ok  	example.com/ch12-cache-aside/internal/cache	3.603s
```

</div>

## רשימת בדיקה

- [ ] שני הקונטיינרים עלו, ו־`CONFIG GET maxmemory-policy` מחזיר `allkeys-lru`.
- [ ] פספוס ואחריו פגיעה עולים קריאה אחת למסד, וה־TTL בין דקה לדקה ועשר שניות.
- [ ] כתיבה מוחקת את המפתח אחרי ה־commit, והקריאה הבאה מחזירה את המחיר החדש.
- [ ] מוצר שלא קיים לא נשמר במטמון.
- [ ] הפלט של `cmd/stampede` מהריצה שלך רשום בפתק, בלי singleflight ועם.
- [ ] הבדיקה של המרוץ עוברת: ערך ישן מיד אחריו, חדש אחרי ה־TTL.
- [ ] כל הבדיקות עוברות עם `-race`, ושלושת הכלים שקטים.
- [ ] הסביבה נמחקה עם `down -v`, והעבודה שמורה ב־Git.

## טעויות נפוצות ותיקונים

- **כל הבדיקות מדלגות עם `set DATABASE_URL and REDIS_ADDR`.** הטרמינל חדש ומשתני הסביבה חסרים. הריצי את
  שתי שורות ה־`export` מצעד 2, ובדקי עם `echo $REDIS_ADDR`.
- **`TestMissThenHit/by_hand` נכשלת עם 2 קריאות למסד במקום 1.** הערך לא נשמר: בדקי שה־`SET` קורה אחרי
  טעינה מוצלחת, ושהמפתח נבנה עם `ProductKey` בדיוק כמו ב־`GET`.
- **`TestWriteInvalidates` נכשלת עם `key exists = 1`.** המחיקה קרתה לפני העדכון, ואז קריאה מילאה את הישן,
  או שהמחיקה חסרה. מחקי רק אחרי ש־`c.store.UpdatePrice` החזירה `nil`.
- **`TestStampede/with_helper` נכשלת עם 2 קריאות.** המילוי למטמון קורה אחרי `group.Do` ולא בתוכה, ולכן
  קוראת מאוחרת פותחת טיסה שנייה. העבירי את ה־`Set` לתוך הפונקציה.
- **`TestNotFoundIsNotCached` נכשלת עם `a missing product was cached`.** השגיאה של המאגר לא עצרה את המילוי.
  החזירי מיד כשהטעינה נכשלה, בלי `SET`.
- **panic עם `TODO(day 32)` כבר ביום 31.** הרצת את כל הבדיקות בלי סינון, ותת־בדיקה `with_helper` הגיעה
  ל־`Cache.Get`. השתמשי בסינון `by_hand` מצעדים 6 ו־7.

## רמזים

- הפקודה `c.rdb.Get(ctx, key).Bytes()` מחזירה ערך ושגיאה; `errors.Is(err, redis.Nil)` הוא פספוס, וכל שגיאה
  אחרת היא Redis לא תקין.
- התוצאה של `group.Do` היא שלושה ערכים: `any`, שגיאה, ודגל שאומר אם התוצאה שותפה. את הדגל אפשר להתעלם ממנו.
- המונה `c.misses` גדל לפני הקריאה ל־`Do`. המונה `c.loads` גדל בתוכה, לפני הטעינה.
- בשביל `Allow`: `fixedWindow.Run(ctx, rdb, []string{key}, window.Milliseconds()).Int64()`.
- כשבדיקה אחת נכשלת, הריצי אותה לבד עם `-run` ו־`-v`. כדי לראות את המפתחות שלה לפני הניקוי, הוסיפי
  זמנית `time.Sleep` בסוף הבדיקה, וחפשי אותם בלקוח עם `KEYS 'ch12_test*'`.

## הפתרון

הפתרון המלא נמצא בתיקייה `solution/`. מסלול הקריאה מבדיל בין פספוס לבין Redis שלא עונה, ולכן המטמון אף
פעם לא הסיבה שקריאה נכשלת. מסלול הכתיבה מוחק רק אחרי שהטרנזקציה חוזרת בהצלחה, כך שקוראת שמפספסת אחרי
המחיקה רואה תמיד את השורה החדשה; את המרוץ שנשאר חוסם ה־TTL, והבדיקה של המרוץ מראה זאת. העוזר ממלא את
המטמון בתוך הטיסה של singleflight, ולכן כל הקוראות שהגיעו בזמן הטעינה, וגם אלה שמגיעות אחריה, עולות
קריאה אחת למסד.

## מה התרגיל מדגים

- cache-aside ביד ובעוזר גנרי, עם ממשק `Store` קטן שקל להחליף.
- אי־תוקף אחרי ה־commit, ובדיקה דטרמיניסטית של המרוץ שה־TTL חוסם.
- פיזור אקראי ב־TTL, ומדיניות פינוי שנקבעת במודע בקובץ ה־compose.
- התפרצות מטמון שנמדדת במונה של המאגר, והתיקון שלה עם singleflight.
- מונה אטומי עם `INCR`, וסקריפט Lua שהופך שתי פקודות לאחת.
- נפילה חזרה למסד כש־Redis לא עונה, ובדיקות שרצות מול קונטיינרים אמיתיים או מדלגות עם הסבר.

</div>
