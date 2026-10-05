---
direction: rtl
tags: [senior-course, book]
chapter: 26
---
# אסטרטגיית בדיקות ובדיקות עומס

- **רמה:** לפי הבאג שהבדיקה תתפוס ומחירה; פירמידה ללוגיקה, גביע לדבק סביב תלויות.
- **‏flaky:** זמן, סדר, מקביליות, אקראיות, סביבה, שירות חיצוני; להוציא מהבדיקה מה שהיא לא שולטת בו.
- **בונה נתונים:** ברירות מחדל תקינות, מתודה לכל שדה שבדיקה משנה.
- **‏testcontainers:** קונטיינר אחד לחבילה ב־`TestMain`, מסד לכל בדיקה, ‏`defer` לפני בדיקת השגיאה, ‏Ryuk לגיבוי.
- **חוזה:** קובץ סכמה ודוגמאות; היצרן מוכיח שהוא שולח בתוך החוזה, הצרכן שהוא מקבל כל מה שהחוזה מתיר.
- **‏format:** הערה בלבד עד `AssertFormat` או `format_checker`; ‏`date-time` ב־Python צריך `rfc3339-validator`.
- **‏`$`:** ב־Python מתאים לפני שורה חדשה בסוף, ב־Go לא.
- **תכונה:** טענה לכל קלט, הקטנה, זרע לשחזור; מקור אקראיות מוזרק נותן ל־rapid לשלוט בו.
- **‏fuzz:** זרעים עם `f.Add`; קריסה נשמרת ב־`testdata/fuzz` ורצה בכל `go test`.
- **‏k6:** ‏`constant-arrival-rate`, ‏`maxRedirects: 0`, תגית לחימום, ספים על p95, ‏p99, שגיאות, ‏`dropped_iterations`.
- **תוצאות:** אחוזונים, השמטה מתואמת, חימום; מספרי מחשב נייד להשוואה ולצורה, לא לייצור.
- **שאלה פתוחה:** ___

<div dir="ltr" align="left">

```text
func TestMain(m *testing.M) { os.Exit(pgtest.Main(m)) }
postgres.Run(ctx, "postgres:18.6", postgres.WithDatabase(..), WithUsername(..), WithPassword(..), postgres.BasicWaitStrategies())
defer testcontainers.TerminateContainer(ctr)   ctr.ConnectionString(ctx, "sslmode=disable")   go test -short
docker ps -a --filter label=org.testcontainers=true      TESTCONTAINERS_RYUK_DISABLED (leave unset)
jsonschema.NewCompiler(); c.AssertFormat(); c.AddResource(url, doc); c.Compile(url)
Draft202012Validator(schema, format_checker=Draft202012Validator.FORMAT_CHECKER)   rfc3339-validator
rapid.Check(t, func(t *rapid.T) { x := rapid.IntRange(a, b).Draw(t, "x") })   -rapid.checks=2000 -rapid.seed=N
go test -run '^$' -fuzz '^FuzzX$' -fuzztime 60s ./pkg/      testdata/fuzz/FuzzX/<hash>   $(go env GOCACHE)/fuzz
k6 run -e RATE=200 --summary-export=results/rate-200.json loadtest/shorty.js
executor: 'constant-arrival-rate' rate timeUnit duration startTime preAllocatedVUs maxVUs tags exec
thresholds: 'http_req_duration{phase:measure,name:resolve}': ['p(95)<25','p(99)<100']  'dropped_iterations{scenario:resolve}': ['count==0']
my runs: rate 200 p99 ___  rate 1000 p99 ___  rate 3000 p99 ___  first dropped iterations at rate ___
```

</div>

קישורים: [[25-slo-alerting-incidents]] · [[27-code-review-static-analysis]]
