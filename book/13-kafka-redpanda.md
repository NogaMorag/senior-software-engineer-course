---
direction: rtl
lang: he
title: "פרק 13: Kafka עם Redpanda"
chapter: 13
part: "חלק ג: יסודות backend"
---

<div dir="rtl" lang="he">

# פרק 13: Kafka עם Redpanda

כשמערכת אחת צריכה לספר לאחרות שמשהו קרה, היא כותבת אירוע ליומן משותף, ומי שמעוניין קורא אותו בקצב
שלו. Kafka היא היומן הנפוץ ביותר לתפקיד הזה, ובראיון בכיר ישאלו אותך פחות איך שולחים הודעה ויותר מה
קורה כשמשהו נופל באמצע. בפרק הזה תריצי Redpanda, שמדברת את הפרוטוקול של Kafka. תביני מחיצות, מיקומים
וקבוצות צרכנים, ותבני את תבנית ה־outbox עם בדיקה שמפילה את התהליך בדיוק ברגע הלא נוח.

## מה נלמד בפרק

- ההבדל בין יומן לתור, ולמה הוא קובע את כל השאר.
- להריץ Redpanda בקונטיינר ולעבוד מולה עם `rpk`.
- topics ומחיצות: מה מחיצה מסדרת, ומה היא לא מסדרת.
- producers: מפתחות, אישורי כתיבה (`acks`) ו־producer אידמפוטנטי.
- קבוצות צרכנים, שמירת מיקום, חלוקה מחדש ועיבוד „לפחות פעם אחת” (at-least-once).
- תבנית ה־outbox, ובדיקה שמפילה את ה־relay בין הפרסום לסימון.
- למה „בדיוק פעם אחת” (exactly-once) הוא חוזה בין שני צדדים ולא תכונה של ה־broker.

## ידע מקדים

- פרק 2: Docker ו־compose, Go, והמשתנה של מאגר הקורס.
- פרק 5: goroutines, context וביטול.
- פרק 9, סעיף 9.5: מפתחות אידמפוטנטיות. הצרכן בפרק הזה משתמש באותו רעיון.
- פרק 10: PostgreSQL בקונטיינר, והגישה אליו מ־Go.
- פרק 11, סעיפים 11.1 ו־11.4: טרנזקציה כיחידה אטומית, ונעילת שורות עם `FOR UPDATE`.

## 13.1 יומן מול תור

שתי משפחות של מערכות מעבירות הודעות בין שירותים, והן נראות דומות רק מבחוץ. **תור (queue)**, למשל
RabbitMQ, מחזיק הודעות עד שמישהו מטפל בהן. צרכן לוקח הודעה, מאשר אותה, והיא נמחקת. ה־broker
זוכר כל הודעה בנפרד: מי לקח אותה, אם אושרה, ומתי לשלוח אותה שוב.

**יומן (log)**, כמו Kafka, הוא רצף שרק מוסיפים לסופו. הודעה שנכתבה נשארת שם לפי מדיניות שמירה (זמן או
גודל), בלי קשר לשאלה אם מישהו קרא אותה. כל צרכן זוכר בעצמו מספר אחד: עד איפה הגיע. המספר הזה נקרא
**offset**, המיקום של רשומה ביומן.

| שאלה | תור | יומן |
| --- | --- | --- |
| מה קורה להודעה אחרי טיפול | נמחקת | נשארת עד שתוקף השמירה פג |
| מה ה־broker זוכר על הצרכן | מצב של כל הודעה | מספר אחד לכל מחיצה |
| קוראים בלתי תלויים | כל אחד צריך תור משלו | כל קבוצה קוראת לבד, מאותו יומן |
| קריאה חוזרת של העבר | לא, ההודעה נמחקה | כן, מזיזים את ה־offset אחורה |
| סדר | חלש, כשיש כמה צרכנים | מובטח בתוך מחיצה |
| ניסיון חוזר להודעה אחת, בעיכוב | מובנה ברוב התורים | אין; היומן מתקדם קדימה |

מכאן נובע כמעט כל מה שבפרק. אפשר להריץ שוב את ההיסטוריה, ולכן צרכן חדש יכול לבנות את המצב שלו מאפס.
המצב של צרכן הוא מספר, ולכן הוא זול וקל לשמור. וגם המחיר ברור: אין ניסיון חוזר עם עיכוב להודעה בודדת.
כשצריך דבר כזה, למשל משלוח webhook שנכשל, עדיף תור עבודות של ממש, וזה פרק 14.

> **מונח:** יש שלוש הבטחות מסירה:
>
> - „לכל היותר פעם אחת” (at-most-once): אולי תאבד, לעולם לא תוכפל.
> - „לפחות פעם אחת” (at-least-once): לעולם לא תאבד, אולי תוכפל.
> - „בדיוק פעם אחת” (exactly-once): הפרק מסביר למה זו הסכם ולא קסם.

## 13.2 Redpanda בקונטיינר ו־rpk

Redpanda היא מימוש של הפרוטוקול של Kafka, שמופץ כבינארי אחד. לקוח של Kafka מדבר איתה בלי לדעת
שזה לא אותו שרת, וזה נכון גם לספרייה של Go שנשתמש בה. היא נבחרה לספר כי broker יחיד בקונטיינר אחד עולה מהר על
מחשב נייד. כל הסמנטיקה שנלמד כאן היא של Kafka, ותקפה גם מול אשכול Kafka אמיתי.

קובץ ה־compose של התרגיל בנוי לפי קובץ ההתחלה המהירה שבתיעוד של Redpanda. הוא
מצומצם ל־broker אחד, בלי אימות משתמשים ובלי רכיבים שהפרק לא צריך. הגרסאות מקובעות בקובץ, נכון לאוקטובר
2026 (ראי נספח ג). אלה ההחלטות שכדאי להבין בו:

| שורה | למה |
| --- | --- |
| `--kafka-addr` | שני מאזינים: הפנימי לקונטיינרים, והחיצוני לתוכניות על המחשב שלך |
| `--advertise-kafka-addr` | הכתובת שה־broker מוסר ללקוחות; היא חייבת להיות נגישה מהצד של הלקוח |
| `--mode dev-container` | הגדרות פיתוח: בלי `fsync`, בלי שריון זיכרון, יצירת topic אוטומטית |
| `--smp 1` | ליבת מעבד אחת מספיקה לתרגילים |
| `console` | ממשק דפדפן לצפייה ב־topics, בהודעות ובקבוצות |
| `127.0.0.1:19092` | ה־Kafka API נפתח רק למחשב שלך |
| `db` | ה־PostgreSQL של התרגיל, לתבנית ה־outbox |

השורה החשובה ביותר להבנה היא כתובת הפרסום. הלקוח מתחבר קודם לכתובת אחת, ומקבל ממנה את רשימת
השרתים ואת הכתובות שלהם. משם הוא מתחבר ישירות ל־broker הנכון. אם ה־broker מפרסם כתובת שאין לה משמעות
מבחוץ, החיבור הראשון מצליח וכל השאר נכשלים. זו התקלה הנפוצה ביותר בהרצת Kafka בקונטיינר.

> **שימי לב:** המצב `dev-container` מדלג על `fsync` לדיסק, ולכן אסור בייצור. לפי התיעוד של Redpanda הוא
> גם מפעיל יצירה אוטומטית של topics, ומגדיר שלוש מחיצות לנתונים הפנימיים של הקבוצות. בבדיקות שלנו כל topic
> נוצר במפורש, כדי שמספר המחיצות לא יהיה מקרי.

מפעילים מתוך תיקיית העבודה של הפרק, ובודקים שה־broker עונה:

<div dir="ltr" align="left">

```bash
docker compose -p ch13 -f data/compose.yaml up -d --wait
docker compose -p ch13 -f data/compose.yaml exec redpanda rpk cluster info
```

</div>

<div dir="ltr" align="left">

```text
CLUSTER
=======
redpanda.92bf3677-3bef-4f71-bcb0-6163f97e6b0d

BROKERS
=======
ID    HOST      PORT
0*    redpanda  9092
```

</div>

הכלי `rpk` הוא שורת הפקודה של Redpanda, והוא כבר מותקן בתוך הקונטיינר. כשמריצים אותו בתוך הקונטיינר הוא
מתחבר למאזין הפנימי בלי הגדרות. כדי לא להקליד את השורה הארוכה בכל פעם, הגדירי פונקציה במעטפת:

<div dir="ltr" align="left">

```bash
rpk() { docker compose -p ch13 -f data/compose.yaml exec -T redpanda rpk "$@"; }
rpk topic list
```

</div>

הדגל `-T` מכבה את המסוף המדומה, וכך אפשר להזרים קלט לפקודה דרך צינור. ממשק הדפדפן נפתח בכתובת
`http://localhost:8088`. שם רואים את אותם נתונים בצורה גרפית.

## 13.3 topics ומחיצות

**topic** הוא יומן עם שם, למשל `orders.placed`. בפועל הוא מחולק ל־**partitions** (מחיצות): כל מחיצה היא
יומן נפרד, עם offsets משלה שמתחילים מ־0. הסדר מובטח רק בתוך מחיצה. בין שתי מחיצות אין שום יחס סדר, גם לא
לפי זמן.

<div dir="ltr" align="left">

```mermaid
flowchart LR
  P[producer] -->|key alice| P0[partition 0: 0 1 2 3]
  P -->|key carol| P2[partition 2: 0 1]
  P0 --> C1[consumer A]
  P2 --> C2[consumer B]
```

</div>

מספר המחיצות הוא החלטה עם השלכות:

- **מקביליות.** בתוך קבוצת צרכנים, כל מחיצה שייכת לצרכן אחד. שלוש מחיצות פירושן שלושה צרכנים פעילים
  לכל היותר; רביעי יחכה בטל.
- **סדר לפי מפתח.** רשומות עם אותו מפתח נוחתות באותה מחיצה, כל עוד מספר המחיצות לא משתנה.
- **הוספת מחיצות.** אפשר להוסיף, ואי אפשר להפחית. אחרי הוספה חלק מהמפתחות עוברים למחיצה אחרת, ובאותו
  רגע הסדר שלהם נשבר.
- **שכפול.** לכל מחיצה יש כמה עותקים על שרתים שונים (replication factor). אצלנו יש broker אחד, ולכן
  עותק אחד.

שמירה (retention) נקבעת לכל topic לפי זמן או גודל. יש גם מצב שמירה שונה, דחיסה (compaction), שבו
נשמרת רק הרשומה האחרונה לכל מפתח. הוא מתאים ל־topic שמייצג מצב עדכני, כמו „הכתובת האחרונה של כל לקוח”.

נסי בעצמך. צרי topic עם שלוש מחיצות, כתבי אליו הודעות עם מפתח, וקראי אותן עם פורמט שמראה מחיצה ו־offset:

<div dir="ltr" align="left">

```bash
rpk topic create demo -p 3
printf 'alice-1\nalice-2\nalice-3\n' | rpk topic produce demo -k alice
printf 'bob-1\nbob-2\n' | rpk topic produce demo -k bob
rpk topic consume demo -n 5 -f '%p %o %k %v\n'
```

</div>

<div dir="ltr" align="left">

```text
0 0 alice alice-1
0 1 alice alice-2
0 2 alice alice-3
0 3 bob bob-1
0 4 bob bob-2
```

</div>

העמודות הן מחיצה, offset, מפתח וערך. שימי לב שגם alice וגם bob נחתו במחיצה 0. זה לא באג: שני מפתחות
יכולים לקבל את אותה מחיצה, וההבטחה היא רק בכיוון ההפוך, שמפתח אחד לא יתפצל. בהרצה שלך המחיצות עשויות
להיות אחרות אם בחרת מפתחות אחרים; רשמי בפתק איזה מפתח נחת איפה.

## 13.4 producers: מפתחות, סדר ו־acks

**producer** הוא הלקוח שכותב. לכל רשומה הוא בוחר מחיצה: כשיש מפתח, לפי גיבוב של המפתח ושארית החלוקה
במספר המחיצות. הספרייה של Go משתמשת באותו גיבוב (murmur2) כמו הלקוח הרשמי ו־`rpk`. לכן התוכנית
`cmd/produce` של התרגיל שולחת את alice לאותה מחיצה שבה היא נחתה קודם:

<div dir="ltr" align="left">

```bash
export KAFKA_BROKERS=localhost:19092
go run ./cmd/produce -topic demo -n 6 -keys alice,bob,carol
```

</div>

<div dir="ltr" align="left">

```text
key=alice  partition=0 offset=5 value=alice-0
key=bob    partition=0 offset=6 value=bob-1
key=carol  partition=2 offset=0 value=carol-2
key=alice  partition=0 offset=7 value=alice-3
key=bob    partition=0 offset=8 value=bob-4
key=carol  partition=2 offset=1 value=carol-5
```

</div>

רשומה בלי מפתח מתפזרת בין המחיצות, ואז אין לה הבטחת סדר מול רשומות אחרות. בחירת המפתח היא לכן החלטה
עסקית: הוא קובע מה חייב להישאר בסדר. בתרגיל המפתח הוא הלקוחה, כי שני אירועים של אותה לקוחה צריכים להגיע
לפי הסדר, ואירועים של לקוחות שונות לא.

השאלה השנייה היא מתי הכתיבה נחשבת הצלחה. ההגדרה `acks` קובעת על אישור של מי ה־producer מחכה:

| ערך | ה־broker מאשר אחרי | מה עלול להיאבד |
| --- | --- | --- |
| `acks=0` | לא מאשר בכלל | כל רשומה, בלי שתדעי |
| `acks=1` | המוביל (leader) של המחיצה כתב | רשומות שהמוביל כתב ונפל לפני ששוכפלו |
| `acks=all` | כל העותקים המסונכרנים (in-sync replicas) כתבו | רק אם כל העותקים המסונכרנים נופלים |

ההגדרה `acks=all` שווה מעט בלי ההגדרה `min.insync.replicas` בצד השרת. לפי התיעוד של Kafka, אם נשאר
עותק מסונכרן אחד בלבד, כתיבה עם `acks=all` עדיין מצליחה, ויכולה ללכת לאיבוד אם גם הוא נופל. ההגדרה
`min.insync.replicas=2` גורמת ל־broker לסרב לכתיבה במצב כזה. אצלנו יש עותק אחד, ולכן `acks=all` פירושו
המוביל בלבד. ובנוסף, המצב `dev-container` לא מבצע `fsync`: אישור הוא הבטחה על עותקים, לא על דיסק.

הבעיה השלישית היא ניסיונות חוזרים. לקוח ששלח מנה ולא קיבל תשובה שולח אותה שוב, אבל אולי השרת
כבר כתב אותה והתשובה אבדה. **producer אידמפוטנטי** (idempotent producer) פותר את זה: הוא מקבל מזהה
מהשרת, וממספר כל מנה בכל מחיצה. השרת זורק מנה שכבר כתב, ושומר על הסדר גם כשכמה בקשות בדרך.

> **טעות נפוצה:** לחשוב שה־producer האידמפוטנטי מונע כפילויות של האפליקציה. הוא מונע כפילויות של ניסיון
> חוזר של הלקוח עצמו, בתוך אותו חיבור לוגי. אם התהליך שלך קורס ושולח את אותה הודעה מחדש אחרי הפעלה, זו
> רשומה חדשה מבחינת ה־broker. סעיף 13.7 בונה בדיוק את המקרה הזה.

נכון לאוקטובר 2026, לפי התיעוד של Kafka, שתי ההגדרות האלה הן ברירת המחדל:
`enable.idempotence=true` ו־`acks=all`. גם ב־franz-go שתיהן ברירת מחדל, והקובץ `bus/bus.go` רק כותב אותן במפורש:

<div dir="ltr" align="left">

```go
// NewProducer returns a client for producing. franz-go's defaults are already the safe
// ones; they are spelled out here so that you can see them:
//   - acks=all: the leader answers only after every in-sync replica has the record;
//   - idempotent writes (on by default; kgo.DisableIdempotentWrite would turn them off):
//     the broker drops a retried batch it already wrote, so client retries add no
//     duplicates and keep per-partition order;
//   - the default partitioner hashes the key (murmur2, as the Java client does), so one
//     key always lands on the same partition while the partition count stays the same.
func NewProducer(brokers []string, opts ...kgo.Opt) (*kgo.Client, error) {
	base := []kgo.Opt{
		kgo.SeedBrokers(brokers...),
		kgo.RequiredAcks(kgo.AllISRAcks()),
		kgo.ProducerLinger(5 * time.Millisecond),
	}
	cl, err := kgo.NewClient(append(base, opts...)...)
	if err != nil {
		return nil, fmt.Errorf("new producer: %w", err)
	}
	return cl, nil
}
```

</div>

ההשהיה `ProducerLinger` אוספת רשומות לכמה אלפיות שנייה לפני שליחה, כדי לשלוח batch גדול יותר. זו פשרה
קלאסית בין תפוקה (throughput) לזמן תגובה (latency).

## 13.5 consumer groups ו־offsets

**consumer** הוא לקוח שקורא. **consumer group** (קבוצת צרכנים) היא כמה צרכנים עם אותו שם קבוצה, שמחלקים
ביניהם את המחיצות: כל מחיצה שייכת בכל רגע לחבר אחד בדיוק. כך מקבלים עבודה במקביל עם סדר בתוך מחיצה.
שתי קבוצות שונות קוראות את אותו topic כל אחת לגמרי לבד.

הקבוצה זוכרת את ההתקדמות שלה כ־offset שמור (committed offset) לכל מחיצה. הוא נשמר אצל השרת, במחיצות
פנימיות בשם `__consumer_offsets`. המספר השמור הוא הרשומה הבאה שצריך לקרוא, לא האחרונה שנקראה. אפשר לראות אותו,
ואת הפיגור (lag), עם `rpk`:

<div dir="ltr" align="left">

```bash
rpk group describe demo-group
```

</div>

<div dir="ltr" align="left">

```text
GROUP                  demo-group
COORDINATOR-NODE       0
COORDINATOR-PARTITION  __consumer_offsets/1
STATE                  Empty
BALANCER               
MEMBERS                0
TOTAL-LAG              0

TOPIC  PARTITION  CURRENT-OFFSET  LOG-START-OFFSET  LOG-END-OFFSET  LAG   MEMBER-ID  CLIENT-ID  HOST
demo   0          9               0                 9               0                           
demo   1          -               0                 0               -                           
demo   2          2               0                 2               0
```

</div>

הפלט הזה נלקח אחרי שצרכן אחד קרא את כל ה־topic ויצא. המחיצה 1 ריקה, ולכן אין לה offset שמור. הפיגור
הוא ההפרש בין סוף היומן לבין ה־offset השמור, והוא המדד הראשון שמסתכלים עליו בתקרית: פיגור שגדל אומר
שהצרכנים לא עומדים בקצב.

השאלה המכרעת היא **מתי** לשמור את ה־offset:

| מתי שומרים | מה קורה בקריסה | ההבטחה |
| --- | --- | --- |
| לפני העיבוד | הרשומות שנשמרו ולא עובדו הולכות לאיבוד | לכל היותר פעם אחת |
| אחרי העיבוד | רשומות שעובדו ולא נשמרו מעובדות שוב | לפחות פעם אחת |
| ברקע, כל כמה שניות | תלוי בתזמון; אולי זה ואולי זה | לא ברורה |

שמירה אוטומטית ברקע נוחה, אבל היא שומרת לפי מה שנקרא, לא לפי מה שהסתיים. לכן התרגיל מכבה אותה, ושומר
בעצמו רק את מה שהעיבוד שלו גמר. כך נבנה צרכן הקבוצה בקובץ `bus/bus.go`:

<div dir="ltr" align="left">

```go
// NewGroupConsumer returns a consumer-group member for topic. Offsets are committed by
// the caller after processing (consumer.Run), never automatically:
//   - DisableAutoCommit: no background commit of records you have not finished;
//   - BlockRebalanceOnPoll: a rebalance waits until the caller calls AllowRebalance, so
//     a commit never lands on a partition that has moved to another member;
//   - CooperativeStickyBalancer (the franz-go default, spelled out): in a rebalance only
//     the partitions that move are revoked, the others keep being consumed;
//   - a new group starts at the beginning of the topic, not at its end.
func NewGroupConsumer(brokers []string, group, topic string, opts ...kgo.Opt) (*kgo.Client, error) {
	base := []kgo.Opt{
		kgo.SeedBrokers(brokers...),
		kgo.ConsumerGroup(group),
		kgo.ConsumeTopics(topic),
		kgo.DisableAutoCommit(),
		kgo.BlockRebalanceOnPoll(),
		kgo.Balancers(kgo.CooperativeStickyBalancer()),
		kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()),
	}
	cl, err := kgo.NewClient(append(base, opts...)...)
	if err != nil {
		return nil, fmt.Errorf("new consumer: %w", err)
	}
	return cl, nil
}
```

</div>

האפשרות האחרונה חשובה בהרצה הראשונה: קבוצה חדשה, בלי offset שמור, מתחילה מתחילת ה־topic. ברירת המחדל
היא דווקא הסוף, וצרכן חדש שמתחיל מהסוף לא יראה הודעות שנכתבו לפני שעלה.

## 13.6 rebalance ועיבוד „לפחות פעם אחת”

**rebalance** הוא חלוקה מחדש של המחיצות בין חברי הקבוצה. הוא קורה כשחבר מצטרף, כשחבר עוזב, כשחבר מפסיק
לשלוח פעימות לב (heartbeats) בזמן, וכשמספר המחיצות משתנה. ב־franz-go, חבר שלא נשמע ממנו 45 שניות מוצא
מהקבוצה כברירת מחדל. עד אז, המחיצות שלו עומדות.

יש שתי משפחות של חלוקה. בחלוקה הישנה (eager), כל חבר מוותר על כל המחיצות שלו בכל rebalance, וכולם
עוצרים עד שהחלוקה החדשה מוכנה. בחלוקה השיתופית (cooperative-sticky), רק המחיצות שעוברות בפועל נלקחות,
והשאר ממשיכות לעבוד. זו ברירת המחדל של franz-go, והבדיקה של התרגיל מראה אותה: צרכן שני מצטרף לקבוצה עם
ארבע מחיצות, והראשון מוותר על שתיים בלבד.

ה־rebalance יוצר שני סיכונים לעבודה שבאמצע:

- **שמירה על מחיצה שכבר לא שלך.** חבר ששומר מיקום אחרי שהמחיצה עברה לחבר אחר עלול להחזיר אחורה את
  מה שהבעלים החדש כבר שמר. האפשרות `BlockRebalanceOnPoll` מעכבת את החלוקה מחדש עד שהלולאה קוראת
  ל־`AllowRebalance`, כלומר עד שהיא סיימה וגם שמרה.
- **עבודה שעובדה ולא נשמרה.** רשומה שטופלה ממש לפני שהמחיצה עברה, ועוד לא נשמרה, תגיע שוב לבעלים החדש.
  זו כפילות, וזה המחיר של „לפחות פעם אחת”.

הלולאה בקובץ `consumer/consumer.go` מחברת את הכול. היא מטפלת ברשומות לפי הסדר, נעצרת בשגיאה הראשונה,
שומרת רק את מה שהסתיים, ורק אז משחררת את ה־rebalance:

<div dir="ltr" align="left">

```go
func Run(ctx context.Context, cl *kgo.Client, handle func(context.Context, *kgo.Record) error) error {
	for {
		fetches := cl.PollRecords(ctx, 100)
		if ctx.Err() != nil {
			cl.AllowRebalance()
			return ctx.Err()
		}
		if fetches.IsClientClosed() {
			return nil
		}
		if errs := fetches.Errors(); len(errs) > 0 {
			cl.AllowRebalance()
			return fmt.Errorf("fetch %s/%d: %w", errs[0].Topic, errs[0].Partition, errs[0].Err)
		}
		var done []*kgo.Record
		var handleErr error
		iter := fetches.RecordIter()
		for !iter.Done() {
			r := iter.Next()
			if err := handle(ctx, r); err != nil {
				handleErr = fmt.Errorf("handle %s/%d@%d: %w", r.Topic, r.Partition, r.Offset, err)
				break
			}
			done = append(done, r)
		}
		// Commit after processing: only what handle finished. The commit gets its own
		// short deadline instead of ctx, so that work finished just before a shutdown
		// is still committed.
		if len(done) > 0 {
			cctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
			err := cl.CommitRecords(cctx, done...)
			cancel()
			if err != nil {
				cl.AllowRebalance()
				return fmt.Errorf("commit: %w", err)
			}
		}
		cl.AllowRebalance()
		if handleErr != nil {
			return handleErr
		}
	}
}
```

</div>

שימי לב לשמירה עם `context.WithoutCancel`. כשמגיע אות כיבוי, ה־context של הלולאה מבוטל, אבל עבודה
שהסתיימה רגע לפני צריכה עדיין להישמר. זה הרעיון של כיבוי מסודר (graceful shutdown) מפרק 6, סעיף 6.5:
מפסיקים לקחת עבודה חדשה, ומסיימים בצורה נקייה את מה שכבר נלקח.

הבדיקה `TestUncommittedRecordsAreRedelivered` מדמה חבר שמת אחרי שטיפל ב־`m-4` ולפני ששמר אותה. החבר השני
מתחיל בדיוק משם:

<div dir="ltr" align="left">

```text
first member: m-0 m-1 m-2 m-3 m-4 | second member: m-4 m-5 m-6 m-7 m-8 m-9
```

</div>

שום הודעה לא אבדה, והודעה אחת עובדה פעמיים. זה מה שהבטחנו. כל מה שנשאר הוא לדאוג שהעיבוד הכפול לא יזיק,
וזה הנושא של שני הסעיפים הבאים.

> **טעות נפוצה:** הלולאה נעצרת על שגיאה ולא מדלגת. אם הודעה אחת שבורה לתמיד (poison message), הקבוצה
> תיתקע עליה ותנסה שוב לנצח. צרכן אמיתי מבחין בין שגיאה זמנית לשגיאה קבועה, ומעביר הודעה שבורה לתור מתים
> (dead-letter queue). פרק 14, סעיף 14.5, בונה את המנגנון הזה.

## 13.7 תבנית ה־outbox

נניח שירות הזמנות: הוא שומר הזמנה ב־PostgreSQL, וצריך לפרסם את האירוע „הזמנה בוצעה” כדי ששירות המשלוחים
יכין משלוח. הגרסה התמימה כותבת למסד ואז שולחת ל־Kafka. זו **כתיבה כפולה** (dual write): שתי מערכות
נפרדות, בלי טרנזקציה משותפת, ולכן כל סדר נכשל באמצע:

| סדר | נפילה באמצע | התוצאה |
| --- | --- | --- |
| קודם המסד, אחר כך Kafka | אחרי ה־commit, לפני הפרסום | הזמנה בלי אירוע; המשלוח לא יוצא לעולם |
| קודם Kafka, אחר כך המסד | אחרי הפרסום, לפני ה־commit | אירוע על הזמנה שלא קיימת |

ניסיון חוזר לא פותר את זה, כי התהליך עצמו יכול למות בין שתי השורות. **תבנית ה־outbox**
מבטלת את הכתיבה הכפולה. השירות כותב שתי שורות לאותו מסד, באותה טרנזקציה: את ההזמנה, ושורה בטבלה `outbox`
שמתארת את ההודעה. אטומיות של טרנזקציה (פרק 11, סעיף 11.1) מבטיחה ששתיהן קיימות או ששתיהן לא. תהליך נפרד,
ה־**relay**, קורא שורות שלא פורסמו, מפרסם אותן ומסמן אותן.

<div dir="ltr" align="left">

```mermaid
sequenceDiagram
  participant S as orders service
  participant DB as PostgreSQL
  participant R as relay
  participant K as Redpanda
  S->>DB: BEGIN; INSERT order; INSERT outbox; COMMIT
  R->>DB: SELECT unpublished FOR UPDATE
  R->>K: produce, wait for acks
  Note over R: crash here means a duplicate later
  R->>DB: UPDATE published_at; COMMIT
```

</div>

הטבלה עצמה, מהקובץ `db/schema.sql`. לכל שורה יש שני מזהים: `id` קובע את סדר הפרסום, ו־`message_id` הוא
המזהה שהצרכנים יסננו לפיו כפילויות:

<div dir="ltr" align="left">

```sql
CREATE TABLE outbox (
    id            bigserial PRIMARY KEY,           -- publish order
    message_id    uuid        NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    topic         text        NOT NULL,
    key           text        NOT NULL,           -- the Kafka record key (ordering scope)
    payload       jsonb       NOT NULL,
    created_at    timestamptz NOT NULL DEFAULT now(),
    published_at  timestamptz                     -- NULL until the relay marks it
);

-- The relay only ever looks for unpublished rows, so index exactly those.
CREATE INDEX outbox_unpublished ON outbox (id) WHERE published_at IS NULL;
```

</div>

הצד של השירות קצר. הפונקציה `Place` לא נוגעת בשרת ההודעות בכלל. מהקובץ `orders/orders.go`:

<div dir="ltr" align="left">

```go
func Place(ctx context.Context, db *pgxpool.Pool, o Order) (int64, error) {
	var orderID int64
	err := pgx.BeginFunc(ctx, db, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, `
			INSERT INTO orders (customer, sku, quantity) VALUES ($1, $2, $3)
			RETURNING id`, o.Customer, o.SKU, o.Quantity).Scan(&orderID); err != nil {
			return fmt.Errorf("insert order: %w", err)
		}
		payload, err := json.Marshal(Placed{OrderID: orderID, Customer: o.Customer, SKU: o.SKU, Quantity: o.Quantity})
		if err != nil {
			return err
		}
		_, err = outbox.Enqueue(ctx, tx, Topic, o.Customer, payload)
		return err
	})
	if err != nil {
		return 0, fmt.Errorf("place order: %w", err)
	}
	return orderID, nil
}
```

</div>

וה־relay, מהקובץ `outbox/outbox.go`. הוא קורא עד מנה אחת של שורות לפי הסדר ונועל אותן, מפרסם את כולן
ומחכה לאישור, ורק אז מסמן אותן באותה טרנזקציה שקראה:

<div dir="ltr" align="left">

```go
func (r *Relay) RunOnce(ctx context.Context) (int, error) {
	limit := r.BatchSize
	if limit <= 0 {
		limit = 100
	}
	marked := 0
	err := pgx.BeginFunc(ctx, r.DB, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, `
			SELECT id, message_id::text, topic, key, payload::text FROM outbox
			WHERE published_at IS NULL
			ORDER BY id
			LIMIT $1
			FOR UPDATE`, limit)
		if err != nil {
			return fmt.Errorf("read outbox: %w", err)
		}
		msgs, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (Message, error) {
			var m Message
			var payload string
			err := row.Scan(&m.Seq, &m.ID, &m.Topic, &m.Key, &payload)
			m.Payload = []byte(payload)
			return m, err
		})
		if err != nil {
			return fmt.Errorf("read outbox: %w", err)
		}
		if len(msgs) == 0 {
			return nil
		}
		// 1. Publish. If this fails, nothing is marked and the rows are retried later.
		if err := r.Publisher.Publish(ctx, msgs); err != nil {
			return fmt.Errorf("publish %d messages: %w", len(msgs), err)
		}
		// The dangerous gap: the broker has the messages, the database does not know yet.
		if r.AfterPublish != nil {
			if err := r.AfterPublish(ctx, msgs); err != nil {
				return err
			}
		}
		// 2. Mark. Only after the broker acknowledged every message.
		seqs := make([]int64, len(msgs))
		for i, m := range msgs {
			seqs[i] = m.Seq
		}
		if _, err := tx.Exec(ctx, `UPDATE outbox SET published_at = now() WHERE id = ANY($1)`, seqs); err != nil {
			return fmt.Errorf("mark published: %w", err)
		}
		marked = len(msgs)
		return nil
	})
	if err != nil {
		return 0, err
	}
	return marked, nil
}
```

</div>

השדה `AfterPublish` הוא תפר לבדיקה (test seam): בקוד אמיתי הוא ריק, ובבדיקה הוא מחזיר שגיאה בדיוק
ברגע המסוכן. השגיאה מגלגלת אחורה את הטרנזקציה, כאילו התהליך מת: ההודעות כבר ב־broker, והסימון לא קרה.
ה־relay הבא יפרסם אותן שוב. **כפילות, ולא אובדן**: זה הכיוון הנכון לטעות בו, כי אובדן אי אפשר לתקן אחר כך.

למה `FOR UPDATE` ולא `FOR UPDATE SKIP LOCKED`, שהכרת בפרק 11, סעיף 11.4? עם דילוג על שורות נעולות, שני
relays היו לוקחים מנות שונות ומפרסמים אותן במקביל, והסדר בין המנות היה נשבר. עם נעילה רגילה, ה־relay השני
מחכה; כשהראשון מסיים, השורות כבר מסומנות, והשני מדלג עליהן. בפרויקט הגמר רץ relay פעיל אחד, שנבחר בנעילה
של PostgreSQL (פרק 18).

> **שימי לב:** הטבלה גדלה לנצח אם לא מנקים. מחיקה תקופתית של שורות שפורסמו לפני כמה ימים היא חלק מהתבנית.
> החלופה הנפוצה ל־relay שמתשאל את הטבלה היא לכידת שינויים (change data capture): כלי כמו Debezium קורא
> את יומן הכתיבה של המסד ומפרסם ממנו.

נשאר הצד של הצרכן. ההודעה עלולה להגיע פעמיים: פעם בגלל ה־relay, ופעם בגלל rebalance. לכן הצרכן שומר
טבלת סינון (dedupe table) של מזהי הודעות שכבר טופלו, וכותב אליה **באותה טרנזקציה** של התוצאה העסקית.
מהקובץ `consumer/consumer.go`:

<div dir="ltr" align="left">

```go
func Process(ctx context.Context, db *pgxpool.Pool, consumer string, m Message, effect Effect) (bool, error) {
	applied := false
	err := pgx.BeginFunc(ctx, db, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, `
			INSERT INTO processed_messages (consumer, message_id) VALUES ($1, $2)
			ON CONFLICT DO NOTHING`, consumer, m.ID)
		if err != nil {
			return fmt.Errorf("dedupe insert: %w", err)
		}
		if tag.RowsAffected() == 0 {
			return nil // seen before: commit nothing new
		}
		if err := effect(ctx, tx, m); err != nil {
			return err
		}
		applied = true
		return nil
	})
	if err != nil {
		return false, fmt.Errorf("process message %s: %w", m.ID, err)
	}
	return applied, nil
}
```

</div>

אם האפקט נכשל, גם שורת הסינון מתגלגלת אחורה, והמסירה הבאה תנסה שוב. אם שורת הסינון כבר קיימת, האפקט לא
רץ בכלל. אלה שני הכיוונים שהבדיקה `TestFailedEffectLeavesNoDedupeRow` והבדיקה `TestProcess` מכסות.

עכשיו בדיקת הקריסה עצמה, מהקובץ `e2e/crash_test.go`. היא רצה מול Redpanda ו־PostgreSQL אמיתיים: שלוש
הזמנות, relay שמת אחרי הפרסום, relay שני שמפרסם שוב, וקבוצת צרכנים שקוראת הכול:

<div dir="ltr" align="left">

```go
	// Run 1: publish succeeds, then the relay "dies" before the mark commits.
	crash := errors.New("relay killed between publish and mark")
	first := &outbox.Relay{DB: db, Publisher: pub, AfterPublish: func(context.Context, []outbox.Message) error { return crash }}
	if _, err := first.RunOnce(ctx); !errors.Is(err, crash) {
		t.Fatalf("first relay: %v, want the simulated crash", err)
	}
	if p, _ := outbox.Pending(ctx, db); p != 3 {
		t.Fatalf("pending after crash = %d, want 3 (no business row lost its message)", p)
	}
	// Run 2: a restarted relay publishes the same rows again and marks them.
	second := &outbox.Relay{DB: db, Publisher: pub}
	if n, err := second.RunOnce(ctx); err != nil || n != 3 {
		t.Fatalf("restarted relay = %d, %v; want 3, nil", n, err)
	}
```

</div>

בסוף הבדיקה יש שש רשומות ב־topic, ושלושה מזהים שונים, כל אחד פעמיים. בטבלת המשלוחים יש שלוש שורות,
ובטבלת הסינון שלוש. אף הזמנה לא נשארה בלי הודעה, ואף הודעה לא עובדה פעמיים.

## 13.8 למה „בדיוק פעם אחת” הוא חוזה

„בדיוק פעם אחת” נשמע כמו תכונה שמפעילים. Kafka אכן מציעה טרנזקציות: producer טרנזקציוני כותב לכמה מחיצות
ושומר את המיקום של צרכן כפעולה אטומית אחת. צרכן במצב `read_committed` רואה רק טרנזקציות שהושלמו. התיעוד
של Kafka מגדיר את ההיקף במדויק: בדיוק פעם אחת כש„קוראים, מעבדים וכותבים נתונים על topics של Kafka”.
לגבי מערכות יעד אחרות, הוא אומר שהן „בדרך כלל דורשות שיתוף פעולה מצד המערכות האלה”.

גם ב־Redpanda הטרנזקציות פעילות כברירת מחדל, לפי התיעוד שלה. אבל המשלוח בתרגיל הוא שורה ב־PostgreSQL. בעולם
אמיתי הוא קריאת HTTP, מייל או חיוב כרטיס אשראי. שום טרנזקציה של Kafka לא מגלגלת אחורה מייל שנשלח. לכן
מה שמקבלים בפועל הוא **מסירה „לפחות פעם אחת” ועוד אפקט אידמפוטנטי**, ויש שמכנים את זה „בפועל פעם אחת”
(effectively once).

החוזה הוא בין שני צדדים, ולכל צד יש חובה משלו:

| צד | החובה | בתרגיל |
| --- | --- | --- |
| השולח | מזהה יציב לכל הודעה לוגית, זהה בכל פרסום חוזר | העמודה `message_id`, בכותרת `message-id` |
| הצרכן | לסנן לפי המזהה, באותה טרנזקציה של האפקט | `processed_messages` באותו `BEGIN` של המשלוח |

שימי לב מה **לא** משמש מזהה: ה־offset. פרסום חוזר של אותה הודעה מקבל offset חדש, ולכן סינון לפי offset
לא יתפוס אותו. שלושה מקורות שונים של כפילות פוגשים בתרגיל את אותה הגנה:

| מקור הכפילות | איפה נוצרת | מי סופג אותה |
| --- | --- | --- |
| ניסיון חוזר של ה־producer | בין ה־producer ל־broker | ה־producer האידמפוטנטי |
| קריסת relay בין פרסום לסימון | ב־relay | טבלת הסינון של הצרכן |
| rebalance או קריסת צרכן לפני שמירה | בצד הקריאה | טבלת הסינון של הצרכן |

לפי התיעוד של Kafka, יש גם דרך אחרת: לשמור את ה־offset באותו מקום של התוצאה, כאן באותה טרנזקציה של
PostgreSQL, ולהתחיל לקרוא ממנו בהפעלה. זה עובד לצרכן אחד שכותב למסד אחד. טבלת הסינון כללית יותר: היא
מגינה גם מכפילויות שנוצרו לפני ה־broker, כמו ה־relay שלנו.

## 13.9 franz-go מ־Go

לקוחות Kafka ל־Go מתחלקים לשניים. יש עטיפות של ספריית C בשם librdkafka, ויש ספריות שכתובות כולן
ב־Go. הספרייה franz-go, שהספר משתמש בה, שייכת לסוג השני. אין צורך במהדר של C, הבנייה פשוטה, וגלאי
המרוצים (`-race`) רואה את כל הקוד.

> **איך Rust עושה את זה:** ב־Rust הלקוח הנפוץ הוא `rdkafka`, עטיפה של ספריית ה־C, ולכן הבנייה צריכה
> אותה. הרעיונות זהים: אישורי כתיבה, צרכן בקבוצה, ושמירת מיקום ידנית אחרי עיבוד. ההבדל המעשי הוא
> בסגנון: שם מחכים ל־future של כל שליחה, וכאן קוראים ל־`ProduceSync`.

ב־franz-go יש טיפוס לקוח אחד, `kgo.Client`, שמשמש גם לכתיבה וגם לקריאה. ההתנהגות נקבעת באפשרויות שמעבירים
לבנאי, כמו שראית ב־`NewProducer` וב־`NewGroupConsumer`. אלה הקריאות שהתרגיל משתמש בהן:

| קריאה | מה היא עושה |
| --- | --- |
| `kgo.NewClient(opts...)` | בונה לקוח; החיבורים נפתחים בעצלות |
| `ProduceSync(ctx, recs...)` | שולחת ומחכה לאישור של כולן; `FirstErr()` מחזירה את השגיאה הראשונה |
| `Produce(ctx, rec, callback)` | שולחת בלי לחכות; התוצאה מגיעה ל־callback |
| `PollRecords(ctx, n)` | מחזירה עד n רשומות מהמחיצות שהוקצו לחבר |
| `CommitRecords(ctx, recs...)` | שומרת לכל מחיצה את ה־offset שאחרי הרשומה הגבוהה ביותר |
| `AllowRebalance()` | משחררת rebalance שמחכה, כשמשתמשים ב־`BlockRebalanceOnPoll` |
| `Close()` | עוזבת את הקבוצה וסוגרת חיבורים; השאר מקבלים rebalance מיד |

המתאם בין ה־relay ל־broker הוא מתודה אחת. הוא בונה רשומה לכל שורת outbox, עם המפתח ועם כותרת שנושאת את
מזהה ההודעה. מהקובץ `bus/bus.go`:

<div dir="ltr" align="left">

```go
func (p KafkaPublisher) Publish(ctx context.Context, msgs []outbox.Message) error {
	recs := make([]*kgo.Record, len(msgs))
	for i, m := range msgs {
		recs[i] = &kgo.Record{
			Topic:   m.Topic,
			Key:     []byte(m.Key),
			Value:   m.Payload,
			Headers: []kgo.RecordHeader{{Key: outbox.HeaderMessageID, Value: []byte(m.ID)}},
		}
	}
	return p.Client.ProduceSync(ctx, recs...).FirstErr()
}
```

</div>

הממשק `outbox.Publisher` מוגדר אצל הצרכנית שלו, כמו בפרק 4. לכן בדיקות ה־relay רצות עם מפרסם מדומה
בלי שרת הודעות, ורק בדיקת הקצה לקצה צריכה Redpanda.

התוכנית `cmd/consume` מראה את הלולאה מבחוץ. היא מצטרפת לקבוצה, מדפיסה כל רשומה וכל שינוי בחלוקה, ומפסיקה
בצורה מסודרת על אות כיבוי. כשהיא מקבלת `SIGTERM`, ה־context מבוטל והעבודה שהסתיימה נשמרת. אחר כך הקריאה
ל־`Close` עוזבת את הקבוצה, כך שהחברים האחרים מקבלים את המחיצות שלה מיד ולא אחרי 45 שניות. זה קטע מהפלט
בהרצה של המחבר:

<div dir="ltr" align="left">

```text
[A] assigned map[demo:[0 1 2]]
[A] partition=0 offset=0 key=alice value=alice-1
[A] partition=2 offset=1 key=carol value=carol-5
[A] revoked  map[demo:[0 1 2]]
```

</div>

## סיכום

- יומן שומר הודעות לפי זמן או גודל, וכל קבוצה זוכרת offset משלה. תור מוחק הודעה שטופלה.
- הסדר מובטח רק בתוך מחיצה. המפתח קובע מחיצה, ולכן הוא קובע מה חייב להישאר בסדר.
- אישור מכל העותקים, עם מינימום עותקים מסונכרנים, מגן מאובדן. ה־producer האידמפוטנטי מגן מכפילויות של
  ניסיון חוזר.
- שמירת offset אחרי עיבוד נותנת „לפחות פעם אחת”: כלום לא הולך לאיבוד, ויש כפילויות.
- rebalance שיתופי מעביר רק מחיצות שזזות; `BlockRebalanceOnPoll` מונע שמירה על מחיצה שעברה.
- תבנית ה־outbox מחליפה כתיבה כפולה בטרנזקציה אחת; ה־relay מפרסם ואז מסמן, וקריסה באמצע יוצרת כפילות.
- „בדיוק פעם אחת” מחוץ ל־Kafka הוא חוזה: מזהה יציב בשולח, וסינון באותה טרנזקציה בצרכן.

## מונחים לזכור

| מונח בעברית | באנגלית | במשפט אחד |
| --- | --- | --- |
| יומן | log | רצף שרק מוסיפים לסופו, נשמר בלי קשר לקריאה |
| תור | queue | מחזיק הודעה עד שצרכן מאשר אותה, ואז מוחק |
| topic | topic | יומן עם שם, מחולק למחיצות |
| מחיצה | partition | יומן נפרד בתוך topic; הסדר מובטח רק בתוכו |
| offset | offset | המיקום של רשומה במחיצה; ה־offset השמור הוא הבא לקריאה |
| קבוצת צרכנים | consumer group | צרכנים עם שם משותף שמחלקים ביניהם את המחיצות |
| פיגור | lag | כמה רשומות יש בין סוף המחיצה לבין ה־offset השמור |
| rebalance | rebalance | חלוקה מחדש של המחיצות כשההרכב של הקבוצה משתנה |
| חלוקה שיתופית | cooperative-sticky | rebalance שלוקח רק את המחיצות שעוברות |
| producer אידמפוטנטי | idempotent producer | ה־broker זורק batch שכבר כתב, לפי מזהה ומספר רץ |
| כתיבה כפולה | dual write | כתיבה לשתי מערכות בלי טרנזקציה משותפת |
| תבנית ה־outbox | outbox pattern | ההודעה נכתבת כשורה באותה טרנזקציה של השינוי, ותהליך אחר מפרסם |
| relay | relay | התהליך שמפרסם שורות outbox ומסמן אותן |
| טבלת סינון | dedupe table | מזהי הודעות שטופלו, נכתבים באותה טרנזקציה של האפקט |
| „לפחות פעם אחת” | at-least-once | שום הודעה לא אובדת, וייתכנו כפילויות |
| „בדיוק פעם אחת” | exactly-once | כל הודעה משפיעה פעם אחת; מחוץ ל־Kafka זה חוזה בין שני הצדדים |

## טעויות נפוצות

- **לפרסם ל־Kafka מתוך הטרנזקציה של המסד.** זו עדיין כתיבה כפולה, רק מוסתרת. כתבי שורת outbox.
- **לסנן כפילויות לפי offset.** פרסום חוזר מקבל offset חדש. סנני לפי מזהה ההודעה.
- **לכתוב לטבלת הסינון בטרנזקציה נפרדת מהאפקט.** קריסה ביניהן משאירה הודעה „מטופלת” שלא טופלה.
- **לסמוך על שמירה אוטומטית של offsets.** היא שומרת מה שנקרא, לא מה שהסתיים.
- **להוסיף מחיצות ל־topic קיים בלי לחשוב על מפתחות.** חלק מהמפתחות עוברים מחיצה, והסדר שלהם נשבר.
- **כתובת פרסום שגויה ב־broker בקונטיינר.** החיבור הראשון מצליח, והכתיבה נתקעת. בדקי את `--advertise-kafka-addr`.

## שאלות ראיון

1. **מה ההבדל בין Kafka ל־RabbitMQ, ומתי תבחרי כל אחד?** הראשונה היא יומן: הודעות נשמרות, כל קבוצה זוכרת
   offset, ואפשר לקרוא את העבר שוב. זה מתאים לזרם אירועים שכמה מערכות צורכות, ולבנייה מחדש של מצב. תור
   כמו RabbitMQ זוכר כל הודעה, תומך באישור לכל הודעה ובמסירה חוזרת עם עיכוב, ומתאים לעבודות בודדות עם
   ניסיונות חוזרים. הפשרה: ביומן מקבלים סדר במחיצה ושידור חוזר, ומוותרים על טיפול גמיש בהודעה בודדת.

2. **איך תבטיחי שאירוע יפורסם אם ורק אם הטרנזקציה של המסד הצליחה?** בתבנית ה־outbox: שורת ההודעה נכתבת
   באותה טרנזקציה של השינוי, ו־relay מפרסם ומסמן. כך אין אירוע בלי שינוי ואין שינוי בלי אירוע. המחיר הוא
   כפילות אפשרית כשה־relay נופל בין פרסום לסימון, ולכן הצרכן חייב לסנן. חלופה היא קריאה מה־WAL, שחוסכת
   את ה־polling אבל מוסיפה רכיב להפעיל.

   **המשך:** למה לא פשוט לנסות שוב את הפרסום עד שיצליח? כי התהליך עצמו יכול למות בין ה־commit לפרסום, ואז
   אין מי שינסה שוב. רק מה שנשמר במסד שורד קריסה.

3. **צרכן קרס באמצע batch. מה יקרה להודעות?** הוא שמר offsets רק אחרי עיבוד, ולכן מה שעובד ולא נשמר יגיע
   שוב לחבר שיקבל את המחיצה. מה שלא עובד בכלל יגיע גם הוא, ושום דבר לא יאבד. העיבוד הכפול צריך להיות
   בטוח, בטבלת סינון או באפקט אידמפוטנטי מטבעו. אם שמירה הייתה קורית לפני העיבוד, היו הולכות לאיבוד הודעות.

4. **האם Kafka מספקת „בדיוק פעם אחת”?** בתוך Kafka, כן: טרנזקציות מחברות כתיבה לכמה מחיצות ושמירת offsets,
   וצרכן `read_committed` רואה רק מה שהושלם. מחוץ ל־Kafka, כמו מסד, HTTP או מייל, התיעוד עצמו אומר שנדרש
   שיתוף פעולה מהמערכת השנייה. בפועל בונים „לפחות פעם אחת” ועוד אפקט אידמפוטנטי, וקוראים לזה חוזה: מזהה
   יציב אצל השולח, וסינון באותה טרנזקציה אצל הצרכן.

   **המשך:** למה לא לסנן לפי offset? כי פרסום חוזר של ה־relay יוצר רשומה חדשה עם offset חדש. רק מזהה
   שנקבע לפני ה־broker תופס את המקרה הזה.

5. **כמה מחיצות תיתני ל־topic חדש?** לפחות כמספר הצרכנים שתרצי להריץ במקביל, עם מרווח, כי הוספה מאוחרת
   משנה את המיפוי של מפתחות. מצד שני, כל מחיצה עולה משאבים ב־broker ומאריכה rebalance. הייתי מתחילה ממספר
   שמכסה את המקביליות הצפויה לשנה, ומודדת פיגור לפני שאני מגדילה.

## פתק Obsidian

הפתק המוכן נמצא גם כקובץ: `obsidian/13-kafka-redpanda.md`. זה תוכנו:

- **יומן מול תור:** יומן שומר ומתקדם לפי offset; תור מוחק אחרי אישור.
- **מחיצה:** הסדר מובטח רק בתוכה. המפתח קובע מחיצה, ולכן קובע מה נשאר בסדר.
- **אישורי כתיבה:** אישור מכל העותקים ועוד מינימום מסונכרנים מגן מאובדן; עם עותק אחד זה רק המוביל.
- **producer אידמפוטנטי:** סופג ניסיונות חוזרים של הלקוח, לא שליחה חוזרת של האפליקציה.
- **שמירת offset:** אחרי עיבוד נותנת „לפחות פעם אחת”; לפני עיבוד נותנת „לכל היותר פעם אחת”.
- **rebalance:** שיתופי לוקח רק מה שזז; לחסום עד שמירה, ואז לשחרר.
- **outbox:** שורת הודעה באותה טרנזקציה; relay מפרסם ואז מסמן; קריסה ביניהם יוצרת כפילות, לא אובדן.
- **חוזה „בדיוק פעם אחת”:** מזהה יציב בשולח, סינון באותה טרנזקציה בצרכן; לא לפי offset.
- **שאלה פתוחה:** ___

<div dir="ltr" align="left">

```text
docker compose -p ch13 -f data/compose.yaml up -d --wait
rpk() { docker compose -p ch13 -f data/compose.yaml exec -T redpanda rpk "$@"; }
rpk cluster info      rpk topic create demo -p 3      rpk topic describe demo -p
printf 'a\nb\n' | rpk topic produce demo -k alice
rpk topic consume demo -n 5 -f '%p %o %k %v\n'
rpk group describe demo-group
export KAFKA_BROKERS=localhost:19092
export DATABASE_URL=postgres://course:change-me@localhost:5413/shop
kgo.RequiredAcks(kgo.AllISRAcks())   kgo.DisableAutoCommit()   kgo.BlockRebalanceOnPoll()
cl.CommitRecords(ctx, done...); cl.AllowRebalance()
my key for alice landed on partition ___
docker compose -p ch13 -f data/compose.yaml down -v
```

</div>

## תרגיל

**מטרה:** להריץ Redpanda ולראות מחיצות ומפתחות בעיניים. לבנות צרכן בקבוצה ששומר מיקום רק אחרי עיבוד.
לבנות outbox עם מפרסם, צרכן שמסנן כפילויות, ובדיקה שמפילה את המפרסם בין הפרסום לסימון. **זמן:** שלושה ימי לימוד, ימים 36 עד 38: כ־35 דקות בנייה ביום 36, כ־60 ביום 37 וכ־70 ביום 38. **ידע מקדים:** סעיפים 13.1 עד 13.3 ליום 36; 13.4 עד 13.6 ליום 37; 13.7 עד 13.9 ליום 38;
פרקים 5, 10 ו־11 כמפורט בידע המקדים של הפרק.

הצעדים, קוד ההתחלה, הרמזים והפתרון נמצאים ב־
[`exercises/ch13-outbox/README.md`](../exercises/ch13-outbox/README.md).

**קריטריון סיום:**

- הודעות עם אותו מפתח נראות אצלך באותה מחיצה ובסדר, בהרצה שלך עם `rpk`.
- אחרי הריגת צרכן באמצע batch, הקבוצה מעבדת שוב את מה שלא נשמר, ושום הודעה לא אובדת.
- בדיקת הקריסה מראה כל הודעה מפורסמת פעמיים, משלוח אחד לכל הזמנה, ואף הזמנה בלי הודעה.
- כל הבדיקות עוברות עם `-race` מול ה־stack האמיתי, ושלושת הכלים שקטים.

## העשרה אופציונלית

- הוסיפי ל־relay לולאה שרצה כל שנייה. אחר כך העירי אותה מוקדם יותר עם ההתראות של PostgreSQL
  (`LISTEN` ו־`NOTIFY`).
- כתבי מחיקה תקופתית של שורות outbox שפורסמו לפני יותר משבוע, ובדקי שהאינדקס החלקי נשאר קטן.
- קראי בתיעוד של Kafka על share groups (KIP-932), שמוסיפות ליומן התנהגות של תור. בספרייה שלנו יש
  להן אפשרות בשם `ShareGroup`. בדקי בתיעוד של Redpanda אם הגרסה שלך תומכת בהן.
- הריצי את `cmd/consume` בשני מסופים עם `-sleep 500ms`, והוסיפי שלישי. ספרי כמה מחיצות זזו בכל פעם.

</div>
