---
direction: rtl
lang: he
title: "פרק 21: הפעלת Kubernetes ו־Helm"
chapter: 21
part: "חלק ה: Cloud-native"
---

<div dir="rtl" lang="he">

# פרק 21: הפעלת Kubernetes ו־Helm

פרק 20 הרים את shorty על אשכול. הפרק הזה עוסק במה שקורה אחרי שהשירות רץ: כמה זיכרון ומעבד מותר לו, מי
מוסיף עותקים כשהעומס עולה, איך מחליפים גרסה בלי שמשתמש ירגיש ואיך חוזרים אחורה כשהגרסה שבורה, ואיך מוצאים
תקלה בשיטה ולא בניחוש. בסוף הפרק את אורזת את כל זה ב־chart של Helm. אלה בדיוק השאלות שראיון platform או SRE
שואל אחרי „מה זה pod”.

## מה נלמד בפרק

- בקשות ותקרות משאבים, שלוש מחלקות QoS, ומה קורה בחריגה מכל אחת מהן: חניקה במעבד, הריגה בזיכרון.
- הרכיב metrics-server על kind, ו־HPA שמוסיף עותקים תחת עומס אמיתי ומוריד אותם אחריו.
- אסטרטגיית rolling update, גרסה שבורה שנתקעת בלי לפגוע במשתמשים, וחזרה אחורה.
- גישה מבחוץ (ingress) על kind בדרך שהתיעוד של kind ממליץ עליה נכון לאוקטובר 2026, ולמה ingress-nginx כבר לא
  ברירת המחדל.
- סדר בדיקות קבוע לתקלה, מהסטטוס ועד קונטיינר דיבוג זמני, על ארבע תקלות שתולות.
- חבילה ראשונה (chart) של Helm: תבניות, ערכים וסכמה שבודקת אותם, ומחזור החיים של התקנה, שדרוג וחזרה אחורה.

## ידע מקדים

- פרק 20 כולו: pods, Deployments, Services, ConfigMaps ובדיקות. הפרק מניח שאת יודעת לקרוא manifest ולהריץ
  `kubectl apply`.
- פרק 19: ה־image של shorty, משתמש מספרי שאינו root ומערכת קבצים לקריאה בלבד. התרגיל מביא עותק של
  ה־Dockerfile, עם שינוי אחד שמוסבר בסעיף 21.1.
- פרק 6: shorty עצמו, ומחולל העומס `cmd/loadgen` מסעיף 6.8.
- פרק 2, סעיף 2.10: kind, kubectl ו־Helm מותקנים בגרסאות הנעולות.
- כ־2 GB פנויים במכונה של Docker. כל האשכול הוא node אחד.

## 21.1 requests ו־limits

כל קונטיינר ב־pod יכול להצהיר על שני מספרים לכל משאב. ההבדל ביניהם הוא הבסיס של כל השאר בפרק:

| שדה | מי משתמש בו | מה קורה בחריגה |
| --- | --- | --- |
| `requests` (בקשה) | ה־scheduler, שמציב pod רק על node שיש בו מספיק מקום פנוי | אין חריגה: זו הבטחה מינימלית |
| `limits` במעבד (תקרה) | ה־kernel, דרך cgroups | הקונטיינר נחנק (throttling): ממתין, ולא מת |
| `limits` בזיכרון (תקרה) | ה־kernel, דרך cgroups | הקונטיינר נהרג (OOMKilled) ומופעל מחדש |

ההבדל בין מעבד לזיכרון נובע מהטבע של המשאב. זמן מעבד אפשר לחלק: תהליך שחרג ממכסה מחכה לחלון הזמן הבא.
זיכרון שכבר הוקצה אי אפשר לקחת בחזרה בלי לשבור את התהליך, ולכן ה־kernel הורג אותו. יחידת המעבד היא
millicore: `200m` פירושו חמישית ליבה.

ב־Deployment של התרגיל, בקובץ `solution/k8s/deployment.yaml`, המספרים נבחרו קטנים בכוונה, כדי שיהיה קל
לחצות אותם:

<div dir="ltr" align="left">

```yaml
          resources:
            requests:
              cpu: 50m       # what the scheduler reserves, and what the HPA divides by
              memory: 32Mi
            limits:
              cpu: 200m      # above this the container is throttled, not killed
              memory: 64Mi   # above this the container is OOM-killed
```

</div>

### 21.1.1 שלוש מחלקות QoS

מתוך שני המספרים Kubernetes גוזר לכל pod מחלקת איכות שירות (QoS class). כשה־node נגמר לו הזיכרון, ה־kubelet
מפנה (evicts) pods לפי המחלקה, מהנמוכה לגבוהה:

| מחלקה | התנאי | מתי מפונה |
| --- | --- | --- |
| `Guaranteed` | לכל קונטיינר requests שווים ל־limits, גם במעבד וגם בזיכרון | אחרונה |
| `Burstable` | יש לפחות request אחד, אבל לא התנאי הקודם | באמצע |
| `BestEffort` | אין requests ואין limits בכלל | ראשונה |

הקובץ `qos-pods.yaml` בתרגיל מגדיר שלושה pods זהים שנבדלים רק במשאבים. אחרי `kubectl apply` המחלקה מופיעה
בסטטוס של כל pod:

<div dir="ltr" align="left">

```bash
kubectl get pods -n shorty -l demo=qos -o custom-columns=NAME:.metadata.name,QOS:.status.qosClass
```

</div>

<div dir="ltr" align="left">

```text
NAME             QOS
qos-besteffort   BestEffort
qos-burstable    Burstable
qos-guaranteed   Guaranteed
```

</div>

ה־Deployment של shorty עצמו הוא `Burstable`: הוא מבקש מעט ומרשה לעצמו להתפרץ עד התקרה. זו בחירה נפוצה
לשירותי רשת. שירות קריטי, כמו מסד נתונים, מקבל לרוב `Guaranteed`.

### 21.1.2 נקודות קצה לעומס, רק בעותק של הפרק

כדי לראות חניקה והריגה צריך שירות שצורך משאבים לפי בקשה. לכן העותק של shorty בתרגיל, בתיקייה
`data/shorty`, מוסיף שתי נקודות קצה שאינן קיימות בפרק 6 ולא ייכנסו לשום שירות אמיתי. הן נרשמות רק כשמשתנה
הסביבה `SHORTY_STRESS` שווה `on`:

| נקודת קצה | מה היא עושה |
| --- | --- |
| `POST /stress/memory?mb=N` | מקצה N מגה־בייט, כותבת לכל עמוד, ומחזיקה אותם עד שהתהליך מת |
| `GET /stress/cpu?work=N` | מבצעת כמות קבועה של חישוב ומחזירה כמה זמן שעון הוא לקח |

זה הלב של נקודת הקצה למעבד, מהקובץ `internal/app/stress.go`:

<div dir="ltr" align="left">

```go
	start := time.Now()
	x := uint64(1)
	for range work * 1_000_000 {
		x = x*6364136223846793005 + 1442695040888963407
	}
	fmt.Fprintf(w, "work %d, wall %d ms, checksum %d\n", work, time.Since(start).Milliseconds(), x&0xff)
```

</div>

העבודה קבועה, והזמן נמדד. בלי תקרה הזמן הוא זמן החישוב; תחת תקרה החישוב זהה, אבל התהליך מחכה בין חלונות
הזמן שלו, והזמן גדל. השינוי השני בעותק נמצא ב־Dockerfile: הוא בונה גם את `loadgen` לתוך אותו image, כך
שמחולל העומס ירוץ בתוך האשכול, קרוב לשירות.

### 21.1.3 חניקה, נמדדת

התרגיל שולח את אותה בקשה לשלושה pods: בלי תקרה, עם תקרה של `200m` ועם `100m`. אלה המספרים שנמדדו
במחשב שעליו נכתב הספר; במחשב שלך הם יהיו אחרים, ולכן תמדדי ותרשמי את שלך:

<div dir="ltr" align="left">

```text
work 200, wall 313 ms, checksum 1     <- qos-besteffort, no CPU limit
work 200, wall 1496 ms, checksum 1    <- shorty pod, limit 200m
work 200, wall 3172 ms, checksum 1    <- qos-guaranteed, limit 100m
```

</div>

היחס הוא העיקר, לא המספרים: חמישית ליבה מאריכה פי חמישה בערך, עשירית ליבה פי עשרה. מבחוץ זה נראה כמו
שירות איטי, בלי שגיאה אחת בלוג. זה מה שהופך חניקה לתקלה קשה למציאה: זמן התגובה בזנב (p99) מטפס, והמעבד
של ה־node דווקא פנוי.

> **שימי לב:** חלק מהצוותים לא מגדירים תקרת מעבד בכלל, רק request, בדיוק בגלל החניקה הזאת. הוויכוח אמיתי:
> בלי תקרה שירות אחד יכול לגזול זמן מעבד מהשכנים, ועם תקרה נמוכה מדי הוא נחנק כשה־node פנוי. מה שלא שנוי
> במחלוקת הוא שתקרת זיכרון צריכה להיות.

### 21.1.4 הריגה בזיכרון

התקרה של shorty היא 64 מגה־בייט. התרגיל מבקש ממנו להחזיק 20, ואז עוד 60:

<div dir="ltr" align="left">

```bash
kubectl -n shorty port-forward pod/$POD 18082:8080 &
curl -s -X POST "localhost:18082/stress/memory?mb=20"
curl -s -X POST "localhost:18082/stress/memory?mb=60"
```

</div>

הבקשה הראשונה עונה `holding 20 MiB`. השנייה לא עונה בכלל: החיבור נסגר באמצע. הסיבה רשומה בסטטוס של
ה־pod:

<div dir="ltr" align="left">

```bash
kubectl -n shorty describe pod $POD
```

</div>

<div dir="ltr" align="left">

```text
    Last State:     Terminated
      Reason:       OOMKilled
      Exit Code:    137
    Ready:          True
    Restart Count:  1
```

</div>

קוד היציאה 137 הוא 128 ועוד 9: האות `SIGKILL`. התהליך לא קיבל הזדמנות לכיבוי מסודר. ה־kubelet הפעיל את
הקונטיינר מחדש, וה־pod חזר להיות מוכן. שימי לב גם למה שאבד: shorty שומר קישורים בזיכרון, וכל הקישורים של
העותק הזה נעלמו. בשירות אמיתי, הריגה כזאת היא בקשות שנקטעו באמצע ומטמון שהתרוקן.

> **איך Rust עושה את זה:** ב־Rust הקצאה שנכשלת קוראת ל־`handle_alloc_error`, וברירת המחדל היא abort. אבל
> תחת cgroup ההקצאה עצמה בדרך כלל מצליחה, וה־kernel הורג את התהליך רק כשהוא נוגע בעמודים. לכן גם ב־Rust וגם
> ב־Go אין טיפול בשגיאה שיציל אותך מ־OOMKilled: המגן היחיד הוא תקרה שתואמת את מה שהשירות באמת צורך.

## 21.2 HPA

ה־HPA (Horizontal Pod Autoscaler, בקר שמשנה את מספר העותקים לפי מדד) קורא את צריכת המעבד של כל pod, מחלק
ב־request של אותו pod, ומשווה ליעד. החישוב פשוט, וכדאי לדעת אותו בעל פה:

<div dir="ltr" align="left">

```text
desiredReplicas = ceil( currentReplicas * currentUtilization / targetUtilization )
```

</div>

מכאן נובעת דרישה: בלי request למעבד אין מכנה, וה־HPA מציג `<unknown>`. ומכאן גם פיתוי מסוכן: request נמוך
מדי גורם לכל עומס קטן להיראות כמו מאות אחוזים.

### 21.2.1 metrics-server על kind

ה־HPA לא מודד בעצמו. הוא שואל את ה־API של המדדים, ובאשכול נקי אף אחד לא עונה. הרכיב metrics-server אוסף
צריכת מעבד וזיכרון מה־kubelet של כל node ומגיש אותה. בטבלת התאימות שלו, גרסה 0.9 תומכת ב־Kubernetes 1.34
ומעלה (נכון לאוקטובר 2026; ראי נספח ג).

על kind יש מכשול אחד. ה־kubelet מגיש את המדדים בתעודה שהוא חתם לעצמו, ו־metrics-server מסרב לה. כך זה נראה
כשמתקינים את הקובץ הרשמי כמו שהוא:

<div dir="ltr" align="left">

```text
"Failed to scrape node" err="Get \"https://172.28.0.2:10250/metrics/resource\": tls: failed to verify
certificate: x509: cannot validate certificate for 172.28.0.2 because it doesn't contain any IP SANs"
```

</div>

ה־README של metrics-server מתאר את הדגל `--kubelet-insecure-tls`, שמוותר על בדיקת התעודה, ומסמן אותו
„לבדיקות בלבד”. באשכול מקומי זה מקובל; באשכול אמיתי מתקנים את התעודות. בתרגיל ההוספה נעשית ב־kustomize
(כלי שמשנה קבצי YAML בלי לערוך אותם, ומובנה ב־kubectl), בקובץ `data/metrics-server/kustomization.yaml`:

<div dir="ltr" align="left">

```yaml
resources:
  - https://github.com/kubernetes-sigs/metrics-server/releases/download/v0.9.0/components.yaml
patches:
  - target:
      kind: Deployment
      name: metrics-server
      namespace: kube-system
    patch: |-
      - op: add
        path: /spec/template/spec/containers/0/args/-
        value: --kubelet-insecure-tls
```

</div>

<div dir="ltr" align="left">

```bash
kubectl apply -k data/metrics-server
kubectl -n kube-system rollout status deploy/metrics-server
kubectl top pods -n shorty
```

</div>

היתרון על פני עריכה ביד: הגרסה נעולה בכתובת, והשינוי היחיד כתוב בקובץ. מי שקורא את המאגר רואה בדיוק במה
ההתקנה שלך שונה מהרשמית.

### 21.2.2 HPA בגרסה autoscaling/v2

הגרסה `autoscaling/v2` היא הגרסה היציבה, והיא מאפשרת כמה מדדים והתנהגות מפורשת לגדילה ולהקטנה. זה הקובץ
`solution/k8s/hpa.yaml`:

<div dir="ltr" align="left">

```yaml
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: shorty
  minReplicas: 2
  maxReplicas: 5
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 60   # percent of the CPU request (50m)
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 60  # the default is 300; shorter for the exercise
```

</div>

חלון הייצוב (stabilization window) מונע תנודות: ה־HPA מקטין רק אחרי שההמלצה הנמוכה החזיקה לאורך כל החלון.
ברירת המחדל היא חמש דקות; בתרגיל דקה, כדי שתראי את הירידה באותו יום.

### 21.2.3 עומס, ומה רואים

מחולל העומס רץ כ־pod בתוך האשכול, מאותו image, ופונה ל־Service לפי השם שלו:

<div dir="ltr" align="left">

```bash
kubectl -n shorty run loadgen --image=shorty:ch21 --restart=Never --command -- \
  /loadgen -url "http://shorty/stress/cpu?work=20" -c 4 -d 150s
kubectl -n shorty get hpa shorty --watch
```

</div>

זה מה שנרשם בהרצה שבה נכתב הספר, שורה כל 20 שניות בערך (העמודות: יעד, מינימום, מקסימום, עותקים):

<div dir="ltr" align="left">

```text
shorty   Deployment/shorty   cpu: 252%/60%   2     5     2     65s
shorty   Deployment/shorty   cpu: 268%/60%   2     5     5     86s
shorty   Deployment/shorty   cpu: 161%/60%   2     5     5     106s
shorty   Deployment/shorty   cpu: 13%/60%    2     5     5     3m46s
shorty   Deployment/shorty   cpu: 2%/60%     2     5     2     5m6s
```

</div>

שלושה דברים כדאי לראות כאן. הקפיצה מ־2 ל־5 היא צעד אחד, כי הנוסחה ביקשה יותר מהמקסימום. גם ב־5 עותקים
הניצול נשאר מעל היעד, כי התקרה היא 5 והעומס גדול ממה שהם יכולים לשאת. והירידה הגיעה רק אחרי שהעומס נגמר
וחלון הייצוב עבר. הסיבות כתובות באירועים של ה־HPA:

<div dir="ltr" align="left">

```text
Normal  SuccessfulRescale  horizontal-pod-autoscaler  New size: 5; reason: cpu resource utilization (percentage of request) above target
Normal  SuccessfulRescale  horizontal-pod-autoscaler  New size: 2; reason: All metrics below target
```

</div>

> **טעות נפוצה:** להשאיר `replicas: 2` ב־manifest של Deployment שיש לו HPA. כל `kubectl apply` מחזיר את
> המספר ל־2 באמצע עומס, וה־HPA מגדיל שוב דקה אחר כך. כשיש HPA, הוא הבעלים של מספר העותקים; ה־chart בסעיף
> 21.6 משמיט את השדה כשה־HPA פעיל.

## 21.3 rollouts ו־rollbacks

כשמשנים את התבנית של Deployment, למשל את התג של ה־image, הבקר יוצר ReplicaSet חדש ומעביר אליו את העותקים
בהדרגה. שני מספרים קובעים את הקצב:

| שדה | משמעות | בתרגיל |
| --- | --- | --- |
| `maxSurge` | כמה pods מעבר למספר הרצוי מותר להוסיף בזמן העדכון | 1 |
| `maxUnavailable` | בכמה pods מותר לרדת מתחת למספר הרצוי | 0 |

עם `maxUnavailable: 0`, pod ישן נמחק רק אחרי ש־pod חדש עבר את בדיקת המוכנות. זה החיבור שכדאי לומר בראיון:
**בדיקת המוכנות היא מה שהופך rolling update לבטוח.** בלעדיה „מוכן” פירושו „התהליך עלה”, ו־pod שבור מקבל
תעבודה ומחליף את הישנים. השדה השלישי, `progressDeadlineSeconds`, קובע אחרי כמה זמן בלי התקדמות הבקר מכריז
על כישלון; ברירת המחדל עשר דקות, ובתרגיל דקה.

### 21.3.1 גרסה שבורה

התרגיל מחליף את התג בתג שלא קיים, ומתעד את הסיבה בהערה שה־history מציג:

<div dir="ltr" align="left">

```bash
kubectl -n shorty annotate deploy/shorty kubernetes.io/change-cause="first release, image shorty:ch21" --overwrite
kubectl -n shorty set image deploy/shorty shorty=shorty:ch21-typo
kubectl -n shorty annotate deploy/shorty kubernetes.io/change-cause="image shorty:ch21-typo" --overwrite
kubectl -n shorty rollout status deploy/shorty --timeout=90s
```

</div>

<div dir="ltr" align="left">

```text
Waiting for deployment "shorty" rollout to finish: 1 out of 2 new replicas have been updated...
error: deployment "shorty" exceeded its progress deadline
```

</div>

עכשיו המצב באשכול:

<div dir="ltr" align="left">

```text
NAME                      READY   STATUS         RESTARTS        AGE
shorty-57fcb8b776-szlfp   0/1     ErrImagePull   0               61s
shorty-6fd59fcc59-cbt4h   1/1     Running        1 (7m48s ago)   8m25s
shorty-6fd59fcc59-g9v6w   1/1     Running        0               8m25s
```

</div>

pod חדש אחד תקוע, ושני הישנים ממשיכים לשרת. המשתמשים לא הרגישו כלום, כי `maxUnavailable: 0` לא הרשה לבקר
לגעת בישנים לפני שהחדש מוכן. את הכישלון רואים בתנאים של ה־Deployment:

<div dir="ltr" align="left">

```text
Available=True MinimumReplicasAvailable
Progressing=False ProgressDeadlineExceeded
```

</div>

שני התנאים יחד הם התמונה המלאה: השירות זמין, והעדכון נכשל. מערכת פריסה טובה בודקת את התנאי השני ונכשלת
בקול, במקום להכריז „נפרס” כי `kubectl apply` הצליח.

### 21.3.2 חזרה אחורה

<div dir="ltr" align="left">

```bash
kubectl -n shorty rollout history deploy/shorty
kubectl -n shorty rollout undo deploy/shorty
kubectl -n shorty rollout status deploy/shorty
```

</div>

<div dir="ltr" align="left">

```text
Warning: resource deployments/shorty was previously managed with 'kubectl apply'. Rolling back will not update
the kubectl.kubernetes.io/last-applied-configuration annotation, which may cause unexpected behavior on future
'kubectl apply' operations. Consider using 'kubectl apply' with your previous configuration file instead.
deployment.apps/shorty rolled back
deployment "shorty" successfully rolled out
```

</div>

הפקודה `undo` לא מוחקת את ה־ReplicaSet השבור; היא מעתיקה את התבנית הקודמת לתבנית הנוכחית, וזו נרשמת
כגרסה חדשה (מספר 3 בהיסטוריה). האזהרה חשובה יותר מהפקודה. הקובץ במאגר שלך עדיין אומר מה שאמר, וה־`apply`
הבא יחזיר את מה שכתוב בו. בצוות שעובד מ־Git, `undo` הוא כפתור חירום; התיקון האמיתי הוא commit שמחזיר את
הקובץ.

## 21.4 ingress

ה־Service מפרק 20 הוא כתובת בתוך האשכול. Ingress (אובייקט שמתאר ניתוב HTTP מבחוץ לפי שם ונתיב) הוא החוזה
לגישה מבחוץ, אבל האובייקט לבדו לא עושה כלום: צריך בקר שמממש אותו. במשך שנים הבקר הנפוץ היה ingress-nginx.

### 21.4.1 מה השתנה ב־2026

בבלוג של Kubernetes הוכרז בנובמבר 2025 שהפרויקט ingress-nginx יפרוש במרץ 2026. אחרי הפרישה אין עוד גרסאות,
גם לא תיקוני אבטחה, ו־SIG Network ממליצה לעבור ל־Gateway API (ממשק הניתוב החדש של Kubernetes) או לבקר
ingress אחר. מותקנים קיימים ממשיכים לעבוד, וזו בדיוק הסכנה: הם לא מקבלים תיקונים.

גם התיעוד של kind השתנה. נכון לאוקטובר 2026, העמוד Ingress באתר של kind אומר שהרכיב cloud-provider-kind,
מגרסה 0.9.0 ומעלה, תומך ב־Ingress בעצמו, ושאין צורך בבקר נוסף. אותו רכיב מממש גם Gateway API ושירותים מסוג
`LoadBalancer`. לכן התרגיל לא מתקין ingress-nginx בכלל. הגרסה הנעולה של cloud-provider-kind רשומה בנספח ג.

### 21.4.2 להפעיל את cloud-provider-kind

הרכיב רץ מחוץ לאשכול, כקונטיינר שמדבר עם Docker, ומקים לכל gateway קונטיינר envoy (פרוקסי HTTP) ברשת של
kind:

<div dir="ltr" align="left">

```bash
docker run -d --name ch21-cloud-provider --network kind \
  -v /var/run/docker.sock:/var/run/docker.sock \
  registry.k8s.io/cloud-provider-kind/cloud-controller-manager:v0.12.0 --enable-lb-port-mapping
```

</div>

הדגל האחרון ממפה את הפורט של ה־gateway לפורט במחשב שלך. ה־README של cloud-provider-kind מסביר למה: ב־Mac
וב־Windows הקונטיינרים רצים במכונה וירטואלית, וכתובות ה־IP של kind לא נגישות מהמחשב. על WSL2 עם רשת NAT,
לפי אותו README, הכתובת החיצונית אמורה להיות נגישה ישירות. בהרצה שבה נכתב הספר נבדקה רק הדרך של המיפוי;
אם הכתובת לא עונה אצלך, השתמשי במיפוי.

> **שימי לב:** הקונטיינר רץ בתוך Linux, ולכן הוא לא יכול לזהות לבד שהמארח הוא Mac או Windows. בהרצה שבה נכתב
> הספר, על Mac, הרצה ראשונה בלי הדגל מיפתה רק את פורט הניהול של envoy, ו־`curl` קיבל את דף הניהול שלו במקום את shorty. אם את רואה רשימה של פקודות
> ניהול, זה הסימן.

### 21.4.3 ה־Ingress של shorty

<div dir="ltr" align="left">

```yaml
spec:
  rules:
    - http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: shorty
                port:
                  name: http
```

</div>

אחרי `kubectl apply` מופיעה כתובת, ו־IngressClass חדש שהרכיב יצר וסימן כברירת מחדל:

<div dir="ltr" align="left">

```text
NAME     CLASS                 HOSTS   ADDRESS                            PORTS   AGE
shorty   cloud-provider-kind   *       172.28.0.4,fc00:f853:ccd:e793::4   80      30s
```

</div>

את הפורט במחשב שלך מוצאים בקונטיינר של ה־gateway, שהשם שלו מתחיל ב־`kindccm-`:

<div dir="ltr" align="left">

```bash
GW=$(docker ps --format '{{.Names}}' | grep '^kindccm-' | head -1)
PORT=$(docker port "$GW" 80/tcp | head -1 | cut -d: -f2)
curl -s localhost:$PORT/healthz
```

</div>

ועכשיו ניסוי קטן שמלמד משהו על המערכת ולא רק על ingress. יוצרים קישור דרך ה־ingress, ואז מבקשים אותו עשר
פעמים:

<div dir="ltr" align="left">

```text
302 302 404 302 302 404 302 404 302 404
```

</div>

חצי מהבקשות מקבלות 404. ה־ingress מאזן בין שני העותקים, וכל עותק של shorty שומר קישורים בזיכרון שלו.
הקישור קיים רק בעותק שקיבל את ה־POST. שירות עם מצב בזיכרון לא מתרחב אופקית; זו הסיבה ש־Hookline שומר מצב
ב־PostgreSQL, וזו שאלת המשך קלאסית לכל תשובה על HPA.

## 21.5 דיבוג: describe, logs, events ו־exec

כש־pod לא עובד, הפיתוי הוא לנחש ולשנות. הדרך המקצועית היא סדר קבוע, מהכללי לפרטני, שבו כל צעד מצמצם את
מרחב האפשרויות:

| צעד | הפקודה | השאלה שהיא עונה עליה |
| --- | --- | --- |
| 1 | `kubectl get pods` | האם ה־pod רץ, מוכן, ומופעל מחדש? |
| 2 | `kubectl describe pod` | מה ה־kubelet ניסה, ומה נכשל (בתחתית: Events)? |
| 3 | `kubectl logs --previous` | מה התהליך כתב לפני שמת בפעם האחרונה? |
| 4 | `kubectl get endpointslices` | האם ה־Service מצא pods, ובאיזה פורט? |
| 5 | `kubectl get events --sort-by=.lastTimestamp` | מה קרה ב־namespace כולו, לפי הסדר? |
| 6 | `kubectl debug` או `exec` | מה רואים מבפנים: האם השירות עונה על הפורט שלו? |

הסטטוס בצעד הראשון כבר ממיין את רוב התקלות לשלוש משפחות:

| מה רואים | המשמעות | איפה הראיה |
| --- | --- | --- |
| `CrashLoopBackOff`, מונה הפעלות עולה | התהליך מת מיד אחרי שעלה | לוג של ההרצה הקודמת |
| `Running` אבל `0/1` | התהליך חי, בדיקת המוכנות נכשלת | אירועים מסוג `Unhealthy` |
| `Running` ו־`1/1`, והשירות לא עונה | הבעיה בין ה־Service ל־pod | ה־endpoints והפורטים |

### 21.5.1 התרגיל: ארבע תקלות שתולות

בתיקייה `starter/drill` ארבעה שירותים קטנים, `app-a` עד `app-d`, כל אחד Deployment ו־Service, וכל אחד שבור
בדרך אחרת. הסקריפט `tests/drill_check.sh` פונה לכל אחד מתוך pod באשכול ומדווח:

<div dir="ltr" align="left">

```text
FAIL  app-a
FAIL  app-b
FAIL  app-c
FAIL  app-d
4 service(s) still broken
```

</div>

הסעיף לא מגלה מה שבור באיזה שירות; זה התרגיל. הוא כן מראה איך נראה כל אחד מארבעת סוגי הראיות, כדי שתזהי
אותן כשתגיעי אליהן. לוג של התהליך הקודם, כשהתצורה שגויה, נראה כך:

<div dir="ltr" align="left">

```text
shorty: invalid configuration:
SHORTY_SHUTDOWN_TIMEOUT: "30" is not a duration (example: 10s)
```

</div>

זה ההחזר על ההשקעה של פרק 6: שירות שבודק את התצורה בעלייה ומדווח על כל הבעיות יחד, נכשל עם הודעה אחת
ברורה במקום להתנהג מוזר שעה אחר כך. בדיקה שנכשלת משאירה אירוע ב־describe:

<div dir="ltr" align="left">

```text
Warning  Unhealthy  1s (x13 over 61s)  kubelet  Readiness probe failed: HTTP probe failed with statuscode: 404
```

</div>

ושני שירותים שה־pods שלהם רצים ומוכנים נבדלים ב־EndpointSlices. בראשון אין כתובות בכלל; בשני יש כתובת, אבל
בפורט שגוי:

<div dir="ltr" align="left">

```text
NAME          ADDRESSTYPE   PORTS     ENDPOINTS     AGE
app-a-lbwzt   IPv4          <unset>   <unset>       40s
app-c-ct8xm   IPv4          8081      10.244.0.18   40s
```

</div>

לשורה הראשונה יש שאלת המשך אחת: מה ה־selector של ה־Service, ואילו תוויות יש ל־pods? שתי הפקודות האלה
מניחות אותם זה ליד זה:

<div dir="ltr" align="left">

```bash
kubectl -n drill get svc app-a -o jsonpath='{.spec.selector}{"\n"}'
kubectl -n drill get pods -l app=app-a --show-labels
```

</div>

### 21.5.2 exec, ולמה הוא לא עובד כאן

הצעד האחרון הוא להסתכל מבפנים. אבל ה־image של shorty הוא distroless (פרק 19): אין בו shell, ולכן `exec`
נכשל:

<div dir="ltr" align="left">

```text
OCI runtime exec failed: exec failed: unable to start container process: exec: "sh": executable file not found in $PATH
```

</div>

זו לא תקלה, זו הצלחה של פרק 19: לתוקף שנכנס לקונטיינר אין כלים. הפקודה `kubectl debug` פותרת את זה בלי לוותר
על ה־image הקטן. היא מוסיפה ל־pod הרץ קונטיינר זמני (ephemeral container) עם image משלך, והדגל `--target`
משתף איתו את מרחב התהליכים של הקונטיינר המקורי. ברשת הם תמיד משתפים, כי זה אותו pod.

הניסיון הראשון, בהרצה שבה נכתב הספר, נתקע. הקונטיינר הזמני לא עלה, והסיבה הופיעה רק בסטטוס שלו:

<div dir="ltr" align="left">

```text
container has runAsNonRoot and image will run as root
```

</div>

ה־pod דורש `runAsNonRoot`, הדרישה חלה גם על הקונטיינר הזמני, ו־busybox רץ כ־root. הקובץ
`data/debug-nonroot.json` מוסיף לקונטיינר הזמני משתמש שאינו root, והדגל `--custom` מחיל אותו:

<div dir="ltr" align="left">

```bash
kubectl -n drill debug -i --quiet $POD --image=busybox:1.37.0 --target=shorty \
  --custom=data/debug-nonroot.json -- sh -c 'wget -qO- -T 2 localhost:8080/healthz'
```

</div>

תשובה `ok` מבפנים, יחד עם כישלון מבחוץ, מוכיחה שהתהליך בסדר והבעיה בדרך אליו. זה החצי השני של כל אבחנה:
לא רק למצוא מה שבור, אלא להוכיח מה לא.

## 21.6 Helm: chart ראשון

עד עכשיו התרגיל הוא שבעה קבצי YAML עם ערכים קבועים בתוכם. כדי להתקין את shorty פעמיים, בשני namespaces עם
שתי רמות לוג, צריך להעתיק ולערוך. Helm (מנהל חבילות ל־Kubernetes) פותר את זה בשלוש מילים:

| מונח | מה הוא |
| --- | --- |
| chart | תיקייה עם תבניות של manifests, ערכי ברירת מחדל ומטא־דאטה |
| values | הערכים שממלאים את התבניות; ברירות מחדל ב־chart, דריסות בהתקנה |
| release | התקנה אחת של chart באשכול, עם שם והיסטוריית גרסאות |

### 21.6.1 המבנה

<div dir="ltr" align="left">

```text
chart/shorty/
  Chart.yaml            name, chart version, appVersion
  values.yaml           defaults
  values.schema.json    what values are allowed
  .helmignore           files left out of the package
  templates/
    _helpers.tpl        named templates: names and labels
    configmap.yaml
    deployment.yaml
    service.yaml
    hpa.yaml            rendered only when autoscaling.enabled
    NOTES.txt           printed after install
```

</div>

בקובץ `Chart.yaml` יש שתי גרסאות, ומבלבלים ביניהן כל הזמן:

<div dir="ltr" align="left">

```yaml
apiVersion: v2
name: shorty
description: The shorty URL shortener from chapter 6, packaged in chapter 21.
type: application
version: 0.1.0        # the chart's own version; bump it on every chart change
appVersion: "ch21"    # the application version; the default image tag
```

</div>

השדה `version` שייך ל־chart ומשתנה בכל שינוי בתבניות; השדה `appVersion` שייך לתוכנה שבפנים. אפשר לשחרר
chart חדש לאותה גרסת תוכנה, ולהפך.

### 21.6.2 מה השתנה ב־Helm 4

הספר נועל את Helm 4 (נכון לאוקטובר 2026; נספח ג). רוב מה שתקראי ברשת נכתב על Helm 3, ולכן כדאי להכיר את
ההבדלים. הרשימה כוללת רק מה שמופיע בעמוד הסקירה של Helm 4 באתר helm.sh, או בעזרה של הקובץ הבינארי עצמו:

| נושא | ב־Helm 4 |
| --- | --- |
| החלה | התקנה חדשה משתמשת ב־server-side apply כברירת מחדל; release ישן שומר על השיטה שלו (`--server-side` ברירת מחדל `auto` ב־upgrade) |
| המתנה | הדגל `--wait` לבדו פירושו האסטרטגיה `watcher`, מבוססת kstatus; בלי הדגל, ברירת המחדל היא `hookOnly` |
| שמות דגלים | `--atomic` הפך ל־`--rollback-on-failure`, ו־`--force` הפך ל־`--force-replace` |
| post-renderers | מומשו כתוספים (plugins); אי אפשר עוד להעביר קובץ הרצה ישירות |
| תוספים | מערכת תוספים חדשה, עם סביבת הרצה אופציונלית ב־WebAssembly |
| charts | הפורמט `apiVersion: v2` ממשיך לעבוד; v3 ניסיוני בלבד |

ההבדל הראשון נראה באשכול. אחרי התקנה, השדה `managedFields` של ה־Deployment מראה מי כתב אותו ואיך:

<div dir="ltr" align="left">

```bash
kubectl -n web get deploy shorty -o jsonpath='{range .metadata.managedFields[*]}{.manager} {.operation}{"\n"}{end}'
```

</div>

<div dir="ltr" align="left">

```text
helm Apply
kube-controller-manager Update
```

</div>

הפעולה `Apply` היא server-side apply: השרת זוכר אילו שדות שייכים ל־Helm, ומתריע על התנגשות כשמישהו אחר שינה
אותם. ב־Helm 3 הייתה שם `Update`.

### 21.6.3 install, upgrade, rollback, uninstall

<div dir="ltr" align="left">

```bash
helm install shorty chart/shorty -n web --create-namespace -f data/values-kind.yaml --wait --timeout 120s
helm list -n web
helm upgrade shorty chart/shorty -n web -f data/values-kind.yaml --set config.logLevel=info --wait
helm history shorty -n web
```

</div>

כל `install` ו־`upgrade` יוצר revision (גרסה של ה־release), ו־Helm שומר את כולן כ־Secrets ב־namespace של
ה־release. זה החלק החשוב: Helm יודע מה הוא התקין בכל גרסה, ולכן `rollback` מחזיר בדיוק אותה.

עכשיו שדרוג שבור, הפעם עם רשת ביטחון. הדגל `--rollback-on-failure` מחכה שהמשאבים יהיו מוכנים, ואם לא,
מחזיר את הגרסה הקודמת בעצמו:

<div dir="ltr" align="left">

```bash
helm upgrade shorty chart/shorty -n web -f data/values-kind.yaml \
  --set image.tag=ch21-typo --rollback-on-failure --timeout 60s
```

</div>

<div dir="ltr" align="left">

```text
Error: UPGRADE FAILED: release shorty failed, and has been rolled back due to rollback-on-failure being set:
resource Deployment/web/shorty not ready. status: InProgress, message: Updated: 1/2
context deadline exceeded
```

</div>

<div dir="ltr" align="left">

```text
REVISION  STATUS      DESCRIPTION
1         superseded  Install complete
2         superseded  Upgrade complete
3         failed      Upgrade "shorty" failed: resource Deployment/web/shorty not ready. ...
4         deployed    Rollback to 2
```

</div>

שימי לב להבדל מסעיף 21.3: שם ה־Deployment נשאר תקוע עד שמישהו הסתכל. כאן כלי הפריסה ראה את הכישלון, החזיר
לבד, והחזיר קוד יציאה שאינו אפס, ש־CI יכול לעצור עליו (פרק 23). חזרה ידנית לגרסה מסוימת וניקוי:

<div dir="ltr" align="left">

```bash
helm rollback shorty 1 -n web --wait
helm get values shorty -n web
helm uninstall shorty -n web --wait
```

</div>

## 21.7 values ו־templates

תבנית של Helm היא YAML עם ביטויים של Go templates בתוך סוגריים מסולסלים כפולים. ההפרדה פשוטה: **values הם
מה ששונה בין התקנות, templates הם מה שזהה בכולן.** הנה הקובץ `values.yaml` של ה־chart:

<div dir="ltr" align="left">

```yaml
replicaCount: 2

image:
  repository: shorty
  tag: ""               # empty means .Chart.AppVersion
  pullPolicy: IfNotPresent

# Becomes the ConfigMap that the container reads as SHORTY_* variables.
config:
  baseURL: http://localhost:8080
  logLevel: info        # debug, info, warn or error
  stress: false         # the chapter 21 /stress routes; keep false outside the exercise

service:
  port: 80

resources:
  requests:
    cpu: 50m
    memory: 32Mi
  limits:
    cpu: 200m
    memory: 64Mi

autoscaling:
  enabled: false        # when true, the HPA owns the replica count
  minReplicas: 2
  maxReplicas: 5
  targetCPUUtilizationPercentage: 60
```

</div>

### 21.7.1 `_helpers.tpl`: שמות ותוויות במקום אחד

קובץ שמתחיל בקו תחתון לא מופק כ־manifest. הוא מגדיר תבניות בעלות שם, שכל קובץ אחר קורא להן. התווית החשובה
ביותר היא זו שה־selector בוחר לפיה:

<div dir="ltr" align="left">

```yaml
{{/* Labels that select the pods. They must never change after the first install. */}}
{{- define "shorty.selectorLabels" -}}
app.kubernetes.io/name: {{ .Chart.Name }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}
```

</div>

למה „לעולם לא משתנה”? כי ה־selector של Deployment הוא שדה שאי אפשר לשנות אחרי היצירה. chart שמכניס לתוויות
של ה־selector את גרסת ה־chart שובר את ה־upgrade הראשון. לכן הגרסה נמצאת בתבנית השנייה, `shorty.labels`, ולא
כאן. הקידומת `shorty.` בשם התבנית מונעת התנגשות עם תבניות של charts אחרים שנכללים כתלויות.

### 21.7.2 ה־Deployment, ושלושה דפוסים שכדאי להכיר

מהקובץ `templates/deployment.yaml`, החלקים שבהם התבנית עושה יותר מהעתקה:

<div dir="ltr" align="left">

```yaml
spec:
  {{- if not .Values.autoscaling.enabled }}
  replicas: {{ .Values.replicaCount }}
  {{- end }}
```

</div>

זה התיקון לטעות מסעיף 21.2: כשה־HPA פעיל, השדה לא נכתב בכלל, וה־upgrade לא מתווכח עם ה־HPA.

<div dir="ltr" align="left">

```yaml
      annotations:
        # A changed ConfigMap changes this hash, and so rolls the pods.
        checksum/config: {{ include (print $.Template.BasePath "/configmap.yaml") . | sha256sum }}
```

</div>

shorty קורא את משתני הסביבה רק בעלייה (פרק 20, סעיף 20.6). שינוי ב־ConfigMap לבדו לא מגיע ל־pods שכבר רצים.
הגיבוב של ה־ConfigMap בהערה של התבנית הופך כל שינוי בתצורה לשינוי בתבנית, ולכן ל־rolling update.

<div dir="ltr" align="left">

```yaml
          resources:
            {{- toYaml .Values.resources | nindent 12 }}
```

</div>

הפונקציה `toYaml` הופכת עץ של values ל־YAML, ו־`nindent 12` מוסיפה שורה חדשה והזחה של 12 רווחים. טעות
בהזחה היא הבאג הנפוץ ביותר בתבניות, ולכן אחרי כל שינוי מריצים `helm template` וקוראים את התוצאה. בקובץ הזה גם
ה־securityContext מפרק 19: משתמש שאינו root, מערכת קבצים לקריאה בלבד, בלי הסלמת הרשאות ובלי capabilities.

### 21.7.3 סכמה ל־values

בלי סכמה, `--set replicaCount=two` מגיע לתבנית, מופק כמחרוזת, והשרת דוחה אותו באמצע ההתקנה. או גרוע מזה:
`--set replicas=3`, עם שם שגוי, עובר בשקט ולא עושה כלום. הקובץ `values.schema.json` הוא JSON Schema, ו־Helm
בודק מולו את ה־values לפני שהוא מפיק משהו. קטע ממנו:

<div dir="ltr" align="left">

```json
  "additionalProperties": false,
  "required": ["replicaCount", "image", "config", "service", "resources", "autoscaling"],
  "properties": {
    "replicaCount": {"type": "integer", "minimum": 1, "maximum": 10},
```

</div>

השורה `additionalProperties: false` היא זו שתופסת שגיאות כתיב בשמות. והתוצאה:

<div dir="ltr" align="left">

```text
Error: INSTALLATION FAILED: values don't meet the specifications of the schema(s) in the following chart(s):
shorty:
- at '/replicaCount': got string, want integer
```

</div>

> **שימי לב:** הסכמה בתרגיל מקבלת זיכרון רק ביחידות בינאריות, כמו `64Mi`. Kubernetes מקבל גם `64M`, שהוא
> 64 מיליון בתים, אבל כמעט תמיד מי שכותב אותו התכוון ל־`Mi`. סכמה יכולה להיות מחמירה יותר מהשרת, וזו החלטה
> שכותבים בה.

### 21.7.4 lint, template ו־kubeconform

שלושה כלים, שלוש שאלות שונות:

| כלי | מה הוא בודק | מה הוא לא בודק |
| --- | --- | --- |
| `helm lint --strict` | מבנה ה־chart, תבניות שמופקות, סכמת values | שה־manifests חוקיים ל־Kubernetes |
| `helm template` | מה בדיוק יישלח לשרת, בלי אשכול | שום דבר; זה הפלט שקוראים |
| kubeconform | שכל manifest תואם לסכמה של גרסת Kubernetes מסוימת | שה־selector תואם לתוויות, שהפורט נכון |

הסקריפט `tests/check_static.sh` מריץ את שלושתם, ואת kubeconform מתוך ה־image שלו, בתג נעול. אלה השורות
שמגדירות את kubeconform ומעבירות אליו את פלט התבניות:

<div dir="ltr" align="left">

```bash
KUBECONFORM_IMAGE="ghcr.io/yannh/kubeconform:v0.8.0"
K8S_VERSION="1.37.0"
kubeconform() { docker run --rm -i -v "$ROOT:/work:ro" "$KUBECONFORM_IMAGE" -strict -summary -kubernetes-version "$K8S_VERSION" "$@"; }
rendered="$(helm template ch21 "$CHART" --set autoscaling.enabled=true)" || rendered=""
if [ -n "$rendered" ] && printf '%s\n' "$rendered" | kubeconform -; then pass "helm template | kubeconform"; else fail "helm template | kubeconform"; fi
```

</div>

<div dir="ltr" align="left">

```text
Summary: 4 resources found parsing stdin - Valid: 4, Invalid: 0, Errors: 0, Skipped: 0
```

</div>

הטור השלישי בטבלה הוא הלקח של סעיף 21.5: ארבעת השירותים השבורים בתרגיל הדיבוג עוברים את kubeconform בלי
הערה. סכמה מוכיחה שהקובץ חוקי, לא שהוא נכון. את הפער סוגרים בבדיקת עשן מול אשכול אמיתי: `tests/smoke.sh`
מתקין את ה־chart, יוצר קישור ופותר אותו, משדרג, חוזר אחורה ומסיר.

## סיכום

- הבקשות (requests) הן הבטחה ל־scheduler ומכנה ל־HPA; התקרות (limits) הן גבול עליון. חריגה במעבד היא חניקה ואיטיות, חריגה בזיכרון
  היא `OOMKilled` וקוד 137.
- מחלקת QoS נגזרת מהמספרים וקובעת מי מפונה ראשון; shorty הוא `Burstable`.
- ה־HPA צריך את metrics-server ובקשת מעבד; על kind צריך את הדגל `--kubelet-insecure-tls`, לבדיקות בלבד.
- כשיש HPA, הוא הבעלים של מספר העותקים, וה־manifest לא כותב את השדה `replicas`.
- עדכון מסוג rolling update, עם `maxUnavailable: 0` ובדיקת מוכנות, משאיר גרסה שבורה תקועה בלי לפגוע
  במשתמשים. התנאי `ProgressDeadlineExceeded` מודיע על הכישלון, והפקודה `rollout undo` חוזרת אחורה.
- הפרויקט ingress-nginx פרש במרץ 2026. על kind, הרכיב cloud-provider-kind מממש בעצמו גם Ingress וגם Gateway API.
- דיבוג הוא סדר קבוע: סטטוס, תיאור, לוג, endpoints ואירועים, ורק בסוף קונטיינר זמני מבפנים.
- הכלי Helm מפריד ערכים מתבניות ושומר היסטוריה של כל release; בגרסה 4 הוא מתקין ב־server-side apply.
- סכמה לערכים תופסת טעויות לפני ההתקנה; הכלי kubeconform בודק חוקיות, ורק אשכול אמיתי בודק נכונות.

## מונחים לזכור

| מונח בעברית | באנגלית | במשפט אחד |
| --- | --- | --- |
| בקשת משאבים | resource request | הכמות שה־scheduler שומר ל־pod, והמכנה של ה־HPA |
| תקרת משאבים | resource limit | הכמות שה־kernel לא מרשה לקונטיינר לעבור |
| חניקה | CPU throttling | קונטיינר שהגיע לתקרת המעבד ממתין לחלון הזמן הבא |
| הריגה בזיכרון | OOMKilled | ה־kernel הורג קונטיינר שעבר את תקרת הזיכרון |
| מחלקת איכות שירות | QoS class | Guaranteed, Burstable או BestEffort; קובעת סדר פינוי |
| פינוי | eviction | ה־kubelet מוציא pods מ־node שנגמר לו משאב |
| HPA | Horizontal Pod Autoscaler | בקר שמשנה מספר עותקים לפי מדד ויעד |
| חלון ייצוב | stabilization window | זמן שהמלצה נמוכה צריכה להחזיק לפני הקטנה |
| rolling update | rolling update | החלפה הדרגתית של pods לפי `maxSurge` ו־`maxUnavailable` |
| rollback | rollback | חזרה לתבנית או ל־revision קודמים |
| Ingress | Ingress | אובייקט שמתאר ניתוב HTTP מבחוץ לשירותים |
| Gateway API | Gateway API | ממשק הניתוב החדש של Kubernetes, יורש של Ingress |
| קונטיינר זמני | ephemeral container | קונטיינר שמוסיפים ל־pod רץ לצורך דיבוג |
| chart | chart | חבילה של Helm: תבניות, values ומטא־דאטה |
| release | release | התקנה אחת של chart, עם היסטוריית revisions |
| values | values | הערכים שממלאים את התבניות בהתקנה |
| server-side apply | server-side apply | החלה שבה השרת מנהל בעלות על שדות |

## טעויות נפוצות

- **HPA בלי request למעבד.** היעד מוצג `<unknown>` ולא קורה כלום. ה־HPA מחלק ב־request.
- **`replicas` ב־manifest של Deployment עם HPA.** כל apply מחזיר את המספר אחורה באמצע עומס.
- **rolling update בלי בדיקת מוכנות.** „מוכן” הופך ל־„התהליך עלה”, ו־pod שבור מחליף את הטובים.
- **`rollout undo` כתיקון קבוע.** הקובץ ב־Git לא השתנה, וה־apply הבא מחזיר את הגרסה השבורה.
- **לחפש באג בקוד כשהשירות רק איטי.** בדקי חניקה קודם: תקרת מעבד נמוכה נראית בדיוק כמו קוד איטי.
- **תוויות גרסה ב־selector של chart.** ה־selector לא ניתן לשינוי, וה־upgrade הראשון נכשל.
- **להסתפק ב־kubeconform.** manifest חוקי עם selector שגוי עובר את הבדיקה ולא עובד.

## שאלות ראיון

1. **מה ההבדל בין חריגה מתקרת מעבד לחריגה מתקרת זיכרון?** מעבד אפשר לחלק בזמן, ולכן קונטיינר שחרג ממתין
   לחלון הבא: הוא איטי, לא מת, ובלוג אין כלום. זיכרון שהוקצה אי אפשר לקחת בחזרה, ולכן ה־kernel הורג את
   התהליך, `OOMKilled` עם קוד 137, וה־kubelet מפעיל אותו מחדש. המשמעות לתפעול: חניקה מתגלה במדדי זמן תגובה
   ובמונה החניקה של cgroup, הריגה במונה ההפעלות ובסיבה של המצב האחרון. יש צוותים שמוותרים על תקרת מעבד
   כדי לא לשלם בזמן תגובה, במחיר של שכן רועש; על תקרת זיכרון כמעט אף אחד לא מוותר.

2. **איך HPA מחליט כמה עותקים להריץ, ומה יכול להשתבש?** הוא מחלק את הצריכה ב־request, משווה ליעד, ומכפיל
   את מספר העותקים ביחס, עם מינימום, מקסימום וחלון ייצוב להקטנה. משתבש: request נמוך מדי מנפח את האחוזים,
   חסר request מבטל הכול, `replicas` ב־manifest נלחם בו, ושירות עם מצב בזיכרון נשבר כשמוסיפים עותקים. ובעיקר,
   מעבד הוא לא תמיד צוואר הבקבוק: שירות שמחכה למסד נתונים לא יתרחב לפי מעבד.

   **המשך:** ומה אם צוואר הבקבוק הוא אורך תור? autoscaling/v2 תומך במדדים מותאמים וחיצוניים, למשל פיגור
   של consumer group ב־Kafka דרך מתאם מדדים. ההיגיון זהה; המדד משתנה.

3. **למה rolling update עם בדיקת מוכנות מגן על משתמשים?** עם `maxUnavailable: 0` הבקר מוחק pod ישן רק אחרי
   ש־pod חדש מוכן, ו־Service שולח תעבודה רק ל־pods מוכנים. גרסה שלא עוברת את הבדיקה פשוט נתקעת, הישנים
   ממשיכים לשרת, ואחרי `progressDeadlineSeconds` התנאי `Progressing` הופך ל־`False`. בלי בדיקה, או עם בדיקה
   שבודקת רק שהתהליך חי, המנגנון מחליף pods טובים בשבורים באותו קצב. המחיר: עדכון איטי יותר וצורך במקום
   ל־pod נוסף.

4. **ה־pod לא עובד. מה סדר הבדיקות שלך?** קודם `get` כדי למיין: קורס, לא מוכן, או רץ ולא עונה. אם הוא קורס,
   לוג של ההרצה הקודמת; אם הוא לא מוכן, האירועים ב־describe; אם הוא רץ ולא עונה, ה־endpoints של ה־Service
   והפורטים. אחר כך אירועי ה־namespace לפי זמן, ורק בסוף מבפנים, עם `kubectl debug` כשה־image לא מכיל shell.
   הנימוק: כל צעד זול מהבא ומצמצם את האפשרויות, ובכל שלב אני מחפשת ראיה, לא משנה דברים כדי לראות מה קורה.

   **המשך:** ואם ה־image הוא distroless? הפקודה `exec` לא תעבוד, ו־`kubectl debug` מוסיף קונטיינר זמני עם כלים.
   אם ה־pod דורש משתמש שאינו root, הקונטיינר הזמני צריך לקבל משתמש כזה, אחרת הוא לא יעלה.

5. **מה Helm נותן שאין ב־manifests רגילים, ומה המחיר?** חבילה עם פרמטרים, release עם היסטוריה ו־rollback
   לגרסה מדויקת, ובגרסה 4 המתנה למוכנות וחזרה אוטומטית בכישלון. המחיר: שפת תבניות של מחרוזות, שבה טעות
   הזחה שוברת YAML, ומצב נוסף באשכול (ה־revisions) שצריך להבין. לכן כותבים סכמה ל־values, מריצים
   `helm template` ו־kubeconform ב־CI, ושומרים את התבניות פשוטות. לשירות פנימי אחד עם התקנה אחת, kustomize
   יכול להספיק.

## פתק Obsidian

הפתק המוכן נמצא גם כקובץ: `obsidian/21-kubernetes-operations-helm.md`. זה תוכנו:

- **requests:** הבטחה ל־scheduler ומכנה ל־HPA; בלי request למעבד אין autoscaling.
- **limits:** מעבד נחנק ואיטי; זיכרון נהרג, `OOMKilled`, קוד 137, ומה שבזיכרון אבד.
- **QoS:** Guaranteed, Burstable, BestEffort; קובע מי מפונה ראשון.
- **HPA:** צריכה חלקי request מול יעד; חלון ייצוב להקטנה; הוא הבעלים של מספר העותקים.
- **metrics-server על kind:** תעודת kubelet חתומה עצמית; הדגל לבדיקות בלבד.
- **rollout:** `maxUnavailable: 0` ובדיקת מוכנות משאירים גרסה שבורה תקועה; `ProgressDeadlineExceeded`.
- **undo:** כפתור חירום; התיקון האמיתי הוא commit.
- **ingress:** הפרויקט ingress-nginx פרש במרץ 2026; על kind משתמשים ב־cloud-provider-kind, עם מיפוי פורטים.
- **מצב בזיכרון:** שני עותקים, שני מאגרים; חצי מהבקשות 404.
- **דיבוג:** סדר קבוע מהכללי לפרטני; הפקודות בטבלה שלמטה.
- **Helm:** ערכים משתנים בין התקנות, תבניות זהות בכולן; היסטוריה וחזרה לגרסה מדויקת.
- **בדיקות:** סכמה לערכים תופסת טעויות כתיב; בדיקת סכמה מוכיחה חוקיות; רק בדיקת עשן מוכיחה נכונות.
- **שאלה פתוחה:** ___

<div dir="ltr" align="left">

```text
cluster      kind create cluster --config data/kind-config.yaml ; kind load docker-image shorty:ch21 --name ch21
qos          kubectl get pods -n shorty -o custom-columns=NAME:.metadata.name,QOS:.status.qosClass
oom          curl -X POST 'localhost:18082/stress/memory?mb=60' ; describe -> Reason: OOMKilled, Exit Code: 137
throttle     curl 'localhost:PORT/stress/cpu?work=200'   no limit ___ ms   200m ___ ms   100m ___ ms
metrics      kubectl apply -k data/metrics-server  (adds --kubelet-insecure-tls) ; kubectl top pods -n shorty
hpa          desired = ceil(current * utilization / target) ; kubectl -n shorty get hpa --watch
load         kubectl -n shorty run loadgen --image=shorty:ch21 --restart=Never --command -- /loadgen -url ... -c 4 -d 150s
rollout      maxSurge 1, maxUnavailable 0, progressDeadlineSeconds 60 ; rollout status | history | undo
ingress      docker run ... cloud-controller-manager:v0.12.0 --enable-lb-port-mapping ; docker port kindccm-... 80/tcp
debug order  get -> describe -> logs --previous -> get endpointslices -> get events --sort-by=.lastTimestamp -> debug
debug        kubectl debug -i POD --image=busybox:1.37.0 --target=shorty --custom=data/debug-nonroot.json -- sh -c '...'
helm         lint --strict | template | install --wait | upgrade --rollback-on-failure | history | rollback N | uninstall
schema       values.schema.json, additionalProperties: false ; kubeconform -strict -kubernetes-version 1.37.0
mine         HPA max replicas reached after ___ s ; drill faults: a ___  b ___  c ___  d ___
```

</div>

## תרגיל

**מטרה:** להפעיל את shorty על kind כמו שירות אמיתי: משאבים ומחלקות QoS, חניקה והריגה שרואים בעיניים,
metrics-server ו־HPA תחת עומס, rollout שבור ו־undo, ingress דרך cloud-provider-kind, תרגיל דיבוג עם ארבע תקלות
שתולות, ו־chart של Helm עם סכמת values שעובר lint, template, install, upgrade ו־rollback.
**זמן:** שני ימי לימוד, ימים 62 ו־63, כ־65 ו־75 דקות בנייה בהתאמה.
**ידע מקדים:** הפרק; פרק 20 (manifests ו־kind); פרק 19 (ה־Dockerfile, שהתרגיל מביא עותק שלו); kind, kubectl
ו־Helm מפרק 2; Docker פועל.

הצעדים, קבצי ההתחלה, הרמזים והפתרון נמצאים ב־
[`exercises/ch21-ops-helm/README.md`](../exercises/ch21-ops-helm/README.md).

**קריטריון סיום:**

- כל אחת משלוש מחלקות QoS מופיעה על pod, ורשמת בפתק את שלושת זמני החניקה שמדדת.
- ראית `OOMKilled` עם קוד 137 בסטטוס של pod של shorty.
- ה־HPA מציג ניצול נוכחי, מגדיל תחת העומס ומקטין אחריו.
- ה־rollout השבור נתקע עם `ProgressDeadlineExceeded` בזמן שה־pods הישנים משרתים, ו־undo מחזיר אותו.
- shorty עונה דרך ה־ingress, ובפתק כתוב למה חצי מהבקשות קיבלו 404.
- הסקריפט `drill_check.sh` מדווח שארבעת השירותים עונים, ובפתק הראיה שהובילה לכל תקלה.
- הסקריפטים `check_static.sh` ו־`smoke.sh` עוברים על ה־chart שלך, והאשכול נמחק בסוף.

## העשרה אופציונלית

- לקרוא את ההכרזה על פרישת ingress-nginx בבלוג של Kubernetes (נובמבר 2025), ואת המדריך למעבר ל־Gateway API
  באתר gateway-api.sigs.k8s.io; ואז לכתוב `Gateway` ו־`HTTPRoute` ל־shorty במקום ה־Ingress.
- לקרוא את מונה החניקה מתוך ה־node, שהוא הקונטיינר `ch21-control-plane`: להיכנס אליו ב־`docker exec`, ולמצוא
  את הקובץ `cpu.stat` של הקונטיינר תחת `/sys/fs/cgroup`. השדה `nr_throttled` גדל בזמן שאת שולחת בקשות מעבד.
- להוסיף ל־chart תבנית Ingress אופציונלית, עם `ingress.enabled` בסכמה, ולבדוק אותה ב־`helm template`.
- לקרוא את עמוד הסקירה של Helm 4 באתר helm.sh, ולנסות `helm install` עם `--server-side=false` כדי לראות
  את ההבדל ב־`managedFields`.

</div>
