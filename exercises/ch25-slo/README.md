---
direction: rtl
lang: he
---

<div dir="rtl" lang="he">

# תרגיל פרק 25: `slo`, חוקי SLO, runbook ותחקיר

**מטרה:** לכתוב כללי הקלטה ל־SLI של הזמינות ושל זמן התגובה של `shorty`, התראות על קצב שריפה בכמה חלונות,
ובדיקות יחידה להן ב־promtool; לכתוב runbook; לביים תקרית אמיתית במחסנית של פרק 24, לראות את ההתראה יורה
ונסגרת, ולכתוב עליה תחקיר ללא האשמה.

**זמן משוער:** שני ימי לימוד. יום 75: כ־55 דקות בנייה (כללי ההקלטה, ההתראות והבדיקות). יום 76, יום קל: כ־35
דקות (runbook, התקרית המבוימת והתחקיר).

**ידע מקדים:** פרק 25; פרק 24 ותרגיל המחסנית שלו (`exercises/ch24-telemetry`), שהתרגיל הזה מריץ עם החוקים
שלו; פרק 13, כי התקרית נוגעת בכתיבה ל־topic. אין קוד חדש להתקין: promtool רץ מתוך ה־image של Prometheus.

## מה בתיקייה הזאת

| קובץ | תפקיד |
| --- | --- |
| `starter/rules/slo-recording.yml` | כלל ההקלטה של חמש דקות מוכן; שאר החלונות וזמן התגובה הם `TODO` של יום 75 |
| `starter/rules/slo-objectives.yml` | היעדים של `shorty`, שלם: 99.5% זמינות, 99% מתחת ל־250 אלפיות |
| `starter/rules/slo-alerts.yml` | ארבע ההתראות, `TODO` של יום 75 |
| `starter/runbooks/TEMPLATE.md` | תבנית ל־runbook |
| `starter/postmortems/TEMPLATE.md` | תבנית לתחקיר |
| `tests/slo_test.yml` | שש בדיקות promtool, ובהן שריפה מהירה שמעירה ורעש קצר שלא |
| `tests/run_promtool.sh` | `promtool check rules` ו־`promtool test rules`, מתוך `prom/prometheus` |
| `data/stage_incident.sh` | מביים את התקרית ומדפיס ציר זמן |
| `solution/rules/` | החוקים המלאים |
| `solution/runbooks/` | runbook לכל אחת מארבע ההתראות |
| `solution/postmortems/` | התבנית ותחקיר מלא לתקרית המבוימת |

ה־runbooks והתחקירים כתובים באנגלית, כמו שהם נכתבים בצוותים, כדי שתוכלי להעתיק אותם לפרויקט ולהראות אותם
בראיון. החוקים לא מזכירים אף שירות בשמו, חוץ מקובץ היעדים; Hookline משתמש בהם כמו שהם באבן הדרך M8.

## יום 75: SLI, התראות ובדיקות

**1. העתקה.**

<div dir="ltr" align="left">

```bash
mkdir -p ~/senior-course/ch25
cp -r $COURSE/exercises/ch25-slo/starter/. ~/senior-course/ch25/
cp -r $COURSE/exercises/ch25-slo/data ~/senior-course/ch25/data
cd ~/senior-course/ch25
bash tests/run_promtool.sh
```

</div>

הבדיקה הראשונה נכשלת: `job:slo_slow_requests:ratio_rate5m` לא קיים עדיין.

**2. כללי ההקלטה.** השלימי ב־`rules/slo-recording.yml` את ארבעת החלונות הנוספים של הזמינות, ואת חמשת
החלונות של זמן התגובה, לפי סעיף 25.1.1. הריצי שוב את `tests/run_promtool.sh`; הבדיקה הראשונה עוברת.

**3. ההתראות.** כתבי ב־`rules/slo-alerts.yml` את ארבע ההתראות שבהערה, לפי סעיף 25.3.1, עם `summary`,
`description` ו־`runbook_url`. הבדיקות משוות גם את ההערות, אז העתיקי את הניסוח מהבדיקות או שני אותו בשני
המקומות.

**4. לשבור בכוונה.** מחקי את החלון הקצר מההתראה של הקריאה, והריצי שוב: הבדיקה „a two-minute blip does not
page” עדיין עוברת. עכשיו מחקי את החלון הארוך במקום הקצר: היא נכשלת. החזירי את שניהם, ורשמי בפתק מה כל חלון
עושה.

**5. החוקים במחסנית.** הרימי את המחסנית של פרק 24 עם החוקים שלך, ובדקי ש־Prometheus טען אותם:

<div dir="ltr" align="left">

```bash
export CH24=~/senior-course/ch24 RULES_DIR=~/senior-course/ch25/rules
docker compose -p ch24 -f $CH24/data/compose.yaml up -d --build --wait
curl -s localhost:9090/api/v1/rules | python3 -c 'import json,sys; print(sum(len(g["rules"]) for g in json.load(sys.stdin)["data"]["groups"]), "rules")'
```

</div>

## יום 76: runbook, תקרית ותחקיר

**6. runbook.** מהתבנית, כתבי `runbooks/slo-availability-budget-burn.md`: מה ההתראה אומרת, חמש הדקות
הראשונות, איך עוצרים את הנזק ואיך יודעים שזה עבד. השווי אחר כך לפתרון.

**7. התקרית.** עם המחסנית למעלה ו־`RULES_DIR` מוגדר באותה מעטפת, הריצי את הסקריפט. הוא לוקח כ־10 עד 25
דקות: חמש דקות של תנועה תקינה, שינוי התצורה, המתנה לקריאה, החזרה לאחור, והמתנה לסגירה:

<div dir="ltr" align="left">

```bash
cd ~/senior-course/ch25
bash data/stage_incident.sh | tee incident-timeline.txt
```

</div>

בזמן שהוא רץ, פתחי את „RED overview” ואת הדף Alerts של Prometheus בכתובת `http://localhost:9090/alerts`,
וראי את ההתראה עוברת מ־inactive ל־firing. חפשי ב־Loki את שורות השגיאה של `shorty` ופתחי מעקב של בקשה שנכשלה.

**8. התחקיר.** העתיקי את `postmortems/TEMPLATE.md` לקובץ חדש, ומלאי אותו מציר הזמן שלך, מהדשבורד
ומ־Prometheus. כמה מהתקציב נשרף, למשל:

<div dir="ltr" align="left">

```bash
curl -s --get localhost:9090/api/v1/query --data-urlencode \
  'query=sum(increase(http_server_request_duration_seconds_count{job="shorty", http_response_status_code=~"5.."}[1h]))'
```

</div>

**9. ניקוי.**

<div dir="ltr" align="left">

```bash
docker compose -p ch24 -f $CH24/data/compose.yaml down -v
```

</div>

## פלט צפוי

`tests/run_promtool.sh` כשהחוקים מלאים:

<div dir="ltr" align="left">

```text
Checking rules/slo-recording.yml
  SUCCESS: 10 rules found

Checking rules/slo-objectives.yml
  SUCCESS: 2 rules found

Checking rules/slo-alerts.yml
  SUCCESS: 4 rules found

  SUCCESS
```

</div>

ציר הזמן של `data/stage_incident.sh` מההרצה שבה נכתב הספר; השעות שלך יהיו אחרות, והמרווחים דומים:

<div dir="ltr" align="left">

```text
11:13:03  traffic starts (20 requests per second, 40 minutes at most)
11:18:05  baseline done, 5m error ratio 0.000
11:18:05  CHANGE: shorty redeployed with SHORTY_KAFKA_TOPIC=link.created
11:18:12  SLOAvailabilityBudgetBurn: none (5m error ratio 0.000)
11:19:22  SLOAvailabilityBudgetBurn: firing (5m error ratio 0.048)
11:19:22  MITIGATION: shorty rolled back to the default topic
11:19:29  rollback done, 5m error ratio 0.048
11:22:21  SLOAvailabilityBudgetBurn: none (5m error ratio 0.052)
11:22:21  latency page: none
11:22:21  traffic stopped; incident staged. Copy these lines into the postmortem timeline.
```

</div>

## רשימת בדיקה

- [ ] `tests/run_promtool.sh` מסתיים ב־`SUCCESS`, כולל שש הבדיקות.
- [ ] שברת בכוונה חלון אחד ראית איזו בדיקה נכשלת, והחזרת.
- [ ] Prometheus במחסנית טען 16 חוקים מ־`RULES_DIR`.
- [ ] ה־runbook שלך אומר מה לבדוק קודם, איך לעצור את הנזק ואיך יודעים שעבד.
- [ ] התקרית המבוימת ירתה את `SLOAvailabilityBudgetBurn`, וההתראה נסגרה אחרי ההחזרה לאחור.
- [ ] התחקיר כולל סיכום, השפעה, ציר זמן, גורמים תורמים ופעולות עם בעלים ותאריך, בלי האשמה.

## טעויות נפוצות ותיקונים

- **`yaml: did not find expected key` מ־`promtool check rules`.** בבלוק `expr: |`, שורה עם פחות רווחים מהשורה
  הראשונה שלו סוגרת אותו. התחילי את הביטוי בסוגריים, באותה הזחה לכל השורות.
- **הבדיקה נכשלת עם `got: nil` לכלל הקלטה.** בדרך כלל `interval` של 5 דקות: החלון של `rate` פתוח משמאל,
  ובחלון של 5 דקות נשארת דגימה אחת. דגימה כל דקה.
- **הבדיקה נכשלת רק על ההערות (Annotations).** promtool משווה את הטקסט שלהן במדויק. אל תדפיסי `$value`
  בהערות: כללי ההקלטה וההתראות בקבוצות נפרדות, ו־Prometheus מעריך אותן במקביל, אז הערך משתנה בין הרצות.
- **אין התראות ב־Prometheus בזמן התקרית.** המחסנית עלתה בלי `RULES_DIR`, והחוקים לא נטענו. ב־`/api/v1/rules`
  רואים כמה חוקים יש; הרימי מחדש עם המשתנה מוגדר.
- **הסקריפט מחליף את `shorty` ו־Prometheus מאבד את החוקים.** פקודת `up` בלי `--no-deps` יכולה ליצור מחדש
  שירותים תלויים בלי `RULES_DIR`. הסקריפט משתמש ב־`--no-deps`; בפקודות ידניות, גם את.

## רמזים

- ההתאמה בין SLI ליעד: `a > on (job) (14.4 * (1 - b))`. בלי `on (job)` ה־label `__name__` ושאר ה־labels מונעים
  התאמה.
- זמן הגילוי של חלון של שעה הוא בערך שעה כפול הסף חלקי קצב השריפה בפועל: ב־20% שגיאות וסף 7.2%, כ־22 דקות.
- במחסנית חדשה ההתראה יורה מהר יותר, כי לחלון של שעה יש רק כמה דקות של היסטוריה.
- בתחקיר, כתבי את ציר הזמן לפני כל דבר אחר; שאר החלקים נשענים עליו.

## הפתרון

החוקים, ה־runbooks והתחקיר המלא בתיקייה `solution/`. החוקים עובדים כי כל SLI מוקלט לכל `job` בכל אחד מחמשת
החלונות, והיעד הוא סדרה קבועה עם אותו `job`; ההתראה משווה ביניהם ב־`on (job)`, ולכן שירות בלי יעד לא מתריע
ושירות חדש מצטרף בשתי שורות. החלון הארוך מסנן רעש קצר, והחלון הקצר סוגר את ההתראה תוך דקות אחרי תיקון, וזה
בדיוק מה שהבדיקות 2 ו־3 מוכיחות.

## מה התרגיל מדגים

- SLI מבוסס בקשות ויעד שלו, כקוד שאפשר לבדוק.
- התראות על קצב שריפה לפי ספר העבודה של SRE, עם בדיקות יחידה שמוכיחות גם מה לא מעיר.
- runbook שמחובר להתראה, ותהליך תקרית ותחקיר על תקרית אמיתית שביימת.
- את דפוס החוקים ש־Hookline משתמש בו באבן הדרך M8: קובץ SLI כללי וקובץ יעדים לכל שירות.

</div>
