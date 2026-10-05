---
direction: rtl
lang: he
---

<div dir="rtl" lang="he">

# תרגיל פרק 20: `k8s-basics`, shorty על אשכול kind

**מטרה:** להרים אשכול kind מקובץ תצורה עם node image נעול ב־digest, לטעון אליו את images של shorty ושל
ה־receiver, ולכתוב ל־shorty Deployment, ConfigMap, Service ושלוש בדיקות. לראות ReplicaSet מחזיר pod שנמחק,
לקרוא לשירות דרך `port-forward`, להוציא pod מהתעבורה בבדיקת מוכנות בלי שיופעל מחדש, ולשלוח webhook
ל־receiver לפי שם DNS.

**זמן משוער:** שני ימי לימוד. יום 60: כ־50 דקות בנייה (אשכול, טעינה, Deployment). יום 61: כ־45 דקות
(ConfigMap, Secret, Service, בדיקות ותרגיל המוכנות).

**ידע מקדים:** פרק 20; פרק 19 (ה־Dockerfiles; התרגיל מביא עותק משלו, כך שלא צריך את תרגיל 19); פרק 6
(התצורה של shorty והפורט הפרטי של הניפוי). kind ו־kubectl מותקנים מפרק 2, סעיף 2.10. האשכול צריך כ־2 GB
פנויים במכונה של Docker; עצרי stacks אחרים של compose לפני שמתחילים.

## מה בתיקייה הזאת

| קובץ | תפקיד |
| --- | --- |
| `data/shorty/` | הקשר הבנייה של shorty: העתק של `exercises/ch19-images/solution/shorty/` (בלי הבנייה הנאיבית) |
| `data/receiver/` | הקשר הבנייה של ה־receiver: העתק של `exercises/ch19-images/solution/receiver/` בלי הבדיקות |
| `starter/kind-config.yaml` | אשכול של node אחד; שורת ה־image היא `TODO` של יום 60 |
| `starter/k8s/00-namespace.yaml` | ה־namespace `ch20`, שלם |
| `starter/k8s/shorty-deployment.yaml` | שלד; `TODO` של יום 60 (Deployment) ושל יום 61 (תצורה ובדיקות) |
| `starter/k8s/shorty-configmap.yaml`, `starter/k8s/shorty-service.yaml` | `TODO` של יום 61 |
| `starter/k8s/receiver-deployment.yaml`, `starter/k8s/receiver-service.yaml` | ה־receiver, שלם: דוגמה לקרוא |
| `starter/pods/sender.yaml` | pod בודד ששולח webhook אחד ל־receiver לפי שם, שלם |
| `tests/check_manifests.sh` | בדיקה מול ה־API server (dry run בצד השרת) ושבעה כללים של הפרק |
| `tests/smoke.sh` | rollout, קריאות דרך `port-forward`, pod שנמחק, תרגיל המוכנות, DNS |
| `tests/run_all.sh` | הכול מהתחלה: אשכול, בנייה, טעינה, Secret, שתי הבדיקות, ומחיקת האשכול |
| `solution/` | הפתרון; עותקים זהים של `tests/` יושבים ב־`starter/tests/` וב־`solution/tests/` |

## יום 60: אשכול, טעינה, Deployment ו־ReplicaSet

**1. העתקה.** הקבצים שלך והקשרי הבנייה יושבים יחד בתיקייה `ch20`:

<div dir="ltr" align="left">

```bash
mkdir -p ~/senior-course/ch20
cp -r $COURSE/exercises/ch20-k8s-basics/starter/. ~/senior-course/ch20/
cp -r $COURSE/exercises/ch20-k8s-basics/data ~/senior-course/ch20/data
cd ~/senior-course/ch20
git add . && git commit -m "ch20: starter"
```

</div>

**2. נעלי את ה־node image.** מצאי את ה־digest בהוראות הפרסום של kind v0.33.0, והשלימי את שורת ה־`image`
בקובץ `kind-config.yaml`:

<div dir="ltr" align="left">

```bash
gh api repos/kubernetes-sigs/kind/releases/tags/v0.33.0 --jq .body | grep 'v1.37.0'
```

</div>

הפקודה `gh` (GitHub CLI) אינה חלק מההתקנות של פרק 2. אם היא לא מותקנת ומחוברת אצלך, אותו טקסט נמצא
בעמוד הגרסה של kind ב־GitHub, במאגר `kubernetes-sigs/kind`.

**3. צרי את האשכול**, ובדקי שה־node מוכן ושהגרסה היא 1.37.0:

<div dir="ltr" align="left">

```bash
kind create cluster --name ch20 --config kind-config.yaml
kubectl get nodes
kubectl config current-context
```

</div>

**4. בני וטעני את שני ה־images.** בלי הטעינה, ה־pods ייתקעו ב־`ErrImagePull`:

<div dir="ltr" align="left">

```bash
docker build -t shorty:ch20 data/shorty
docker build -t receiver:ch20 data/receiver
kind load docker-image shorty:ch20 receiver:ch20 --name ch20
docker exec ch20-control-plane crictl images | grep -E 'shorty|receiver'
```

</div>

הפקודה האחרונה מסתכלת לתוך ה־node: `crictl` הוא הלקוח של containerd, וה־images שטענת ברשימה שלו.

**5. כתבי את ה־Deployment** של shorty לפי ה־`TODO` של יום 60 וסעיף 20.4. בשלב הזה `SHORTY_BASE_URL` מגיע
מ־`env` ישירות. ואז:

<div dir="ltr" align="left">

```bash
kubectl apply -f k8s/00-namespace.yaml -f k8s/shorty-deployment.yaml
kubectl -n ch20 rollout status deployment/shorty
kubectl -n ch20 get deployments,replicasets,pods -l app=shorty
```

</div>

**6. מחקי pod, וראי את ה־ReplicaSet מחזיר אותו.** בחלון אחד עקבי, בחלון שני מחקי:

<div dir="ltr" align="left">

```bash
kubectl -n ch20 get pods -l app=shorty --watch
```

</div>

<div dir="ltr" align="left">

```bash
kubectl -n ch20 delete pod "$(kubectl -n ch20 get pods -l app=shorty -o jsonpath='{.items[0].metadata.name}')"
```

</div>

רשמי בפתק כמה זמן עבר עד שה־pod החדש היה `1/1`. עצרי את המעקב ב־Ctrl+C, ושמרי (`git commit`). את
האשכול אפשר להשאיר ליום 61, או למחוק עכשיו ולהרים מחדש בצעדים 3 ו־4.

## יום 61: ConfigMap, Secret, Service, בדיקות ו־DNS

**7. ConfigMap ו־Service.** מלאי את `k8s/shorty-configmap.yaml` ואת `k8s/shorty-service.yaml`, והחליפי
ב־Deployment את `env` ב־`envFrom` (סעיפים 20.5 ו־20.6).

**8. שלוש הבדיקות.** הוסיפי ל־Deployment את `startupProbe`, `livenessProbe` ו־`readinessProbe` לפי
ה־`TODO` וסעיף 20.7.

**9. ה־Secret, מהטרמינל ולא מקובץ.** ואז בדיקת ה־manifests ופריסה של הכול:

<div dir="ltr" align="left">

```bash
kubectl apply -f k8s/00-namespace.yaml
kubectl -n ch20 create secret generic receiver-secret --from-literal=signing-secret=change-me
kubectl -n ch20 get secret receiver-secret -o jsonpath='{.data.signing-secret}' | base64 -d; echo
bash tests/check_manifests.sh
kubectl apply -f k8s/
kubectl -n ch20 rollout status deployment/shorty
```

</div>

**10. קריאות דרך `port-forward`.** בחלון אחד:

<div dir="ltr" align="left">

```bash
kubectl -n ch20 port-forward svc/shorty 8080:80
```

</div>

ובחלון שני:

<div dir="ltr" align="left">

```bash
curl -s -X POST localhost:8080/v1/links -H 'Content-Type: application/json' -d '{"url":"https://kubernetes.io/","code":"k8s"}'
curl -s -o /dev/null -w '%{http_code} %{redirect_url}\n' localhost:8080/k8s
```

</div>

השדה `short_url` בתשובה הראשונה מגיע מה־ConfigMap.

**11. תרגיל המוכנות.** בחלון שלישי, `port-forward` לפורט הניפוי של pod אחד; ובחלון השני, כבי והדליקי:

<div dir="ltr" align="left">

```bash
POD=$(kubectl -n ch20 get pods -l app=shorty -o jsonpath='{.items[0].metadata.name}')
kubectl -n ch20 port-forward "pod/$POD" 6060:6060
```

</div>

<div dir="ltr" align="left">

```bash
curl -s -X POST 'localhost:6060/debug/ready?state=off'
kubectl -n ch20 get pods -l app=shorty
kubectl -n ch20 get endpointslices -l kubernetes.io/service-name=shorty -o jsonpath='{range .items[*].endpoints[*]}{.targetRef.name}{" ready="}{.conditions.ready}{"\n"}{end}'
curl -s -X POST 'localhost:6060/debug/ready?state=on'
```

</div>

רשמי: כמה שניות עד שה־pod הראה `0/1`, ומה היה בעמודה `RESTARTS`.

**12. DNS בתוך האשכול:**

<div dir="ltr" align="left">

```bash
kubectl apply -f pods/sender.yaml
kubectl -n ch20 logs -f sender
kubectl -n ch20 delete pod sender
```

</div>

**13. הבדיקה המלאה, ומחיקה.** עצרי את חלונות ה־`port-forward` ואת האשכול, והריצי את הכול מאפס. הסקריפט
מוחק את האשכול בסוף, גם אם משהו נכשל:

<div dir="ltr" align="left">

```bash
kind delete cluster --name ch20
bash tests/run_all.sh
kind get clusters
```

</div>

## פלט צפוי

הבדיקה של צעד 9 (שורת ה־images מקוצרת):

<div dir="ltr" align="left">

```text
ok   API server accepts every manifest (server-side dry run)
ok   every image has an explicit, non-latest tag: ...
ok   no Secret object is stored in the manifests
ok   shorty probes: startup, liveness, readiness = /healthz /healthz /readyz
ok   k8s/shorty-deployment.yaml: runAsNonRoot, read-only root, no privilege escalation
ok   k8s/receiver-deployment.yaml: runAsNonRoot, read-only root, no privilege escalation
ok   Service shorty selects the pods of Deployment shorty (app=shorty)
ok   Service receiver selects the pods of Deployment receiver (app=receiver)
ok   shorty reads its environment from ConfigMap shorty-config
PASS: 9 checks
```

</div>

סוף `run_all.sh` (שמות ה־pods אצלך יהיו אחרים):

<div dir="ltr" align="left">

```text
ok   deleted shorty-...; the ReplicaSet replaced it and 2 endpoints are ready again
ok   readiness off: shorty-... left the endpoints, restarts=0
ok   readiness on: back to 2 ready endpoints
ok   sender reached receiver.ch20.svc.cluster.local: 200 {"status":"ok","duplicate":false}
PASS: 9 checks
Deleting cluster "ch20" ...
Deleted nodes: ["ch20-control-plane"]
```

</div>

## רשימת בדיקה

- [ ] שורת ה־image בקובץ `kind-config.yaml` נעולה ב־digest מהוראות הפרסום, ו־`kubectl get nodes` מראה 1.37.0.
- [ ] שני ה־images נטענו ל־node, ואף pod לא היה ב־`ErrImagePull`.
- [ ] ה־Deployment מדווח `2/2`, ו־pod שנמחק הוחלף ב־pod עם שם אחר.
- [ ] יצירת קישור ומעקב אחריו דרך `port-forward` מצליחים, וה־`short_url` מגיע מה־ConfigMap.
- [ ] כיבוי המוכנות הוריד את ה־pod ל־`0/1` ול־`ready=false` בלי אף הפעלה מחדש, והדלקה החזירה אותו.
- [ ] אין קובץ Secret במאגר, וה־sender קיבל 200 מה־receiver לפי שם.
- [ ] הסקריפט `run_all.sh` מסתיים ב־`PASS` בשתי הבדיקות, ו־`kind get clusters` ריק אחריו.

## טעויות נפוצות ותיקונים

- **`ErrImagePull` או `ImagePullBackOff`.** ה־image לא נטען ל־node, או שהתגית היא `latest`. מריצים את צעד
  4 שוב, ובודקים ש־`imagePullPolicy: IfNotPresent` כתוב.
- **`CreateContainerConfigError` על ה־receiver.** ה־Secret `receiver-secret` לא קיים ב־namespace `ch20`.
  יוצרים אותו בפקודה של צעד 9, וה־pod מתחיל לבד.
- **ה־receiver ב־`CrashLoopBackOff`, והלוג נגמר ב־`ValueError` על `tcp://...`.** ה־Service בשם `receiver`
  הזריק את `RECEIVER_PORT` ל־pod. בודקים ש־`enableServiceLinks: false` נשאר בקובץ (סעיף 20.5.1).
- **`port-forward` נשבר אחרי צעד 6.** הוא היה מחובר ל־pod שמחקת. מריצים אותו מחדש; הוא יתחבר ל־pod חי.
- **ה־pod עובר ל־`0/1` ואז `RESTARTS` עולה.** כתבת את `/readyz` ב־`livenessProbe` במקום ב־`readinessProbe`.
  בדיקת חיות שנכשלת מפעילה מחדש, והמתג חוזר למצבו ההתחלתי.
- **`curl: (7) Failed to connect to localhost port 6060`.** ה־`port-forward` של צעד 11 היה ל־Service, לא
  ל־pod. הפורט 6060 לא חשוף ב־Service; מפנים ל־`pod/<name>`.

## רמזים

- ב־Deployment, התוויות ב־`selector.matchLabels` חייבות להיות זהות לאלה שב־`template.metadata.labels`.
- ב־Service, `targetPort: http` מפנה לשם של פורט הקונטיינר, כך ששינוי מספר הפורט נוגע במקום אחד.
- ב־ConfigMap, כל הערכים הם מחרוזות. הערך `10s` נשאר מחרוזת, ו־shorty מפרש אותה כמשך זמן.
- בבדיקות, `port: http` עובד גם כאן. השדה `periodSeconds` קובע כל כמה זמן שואלים.
- כדי לראות למה pod לא מוכן, הריצי עליו `kubectl describe` וקראי את האירועים בסוף הפלט.

## הפתרון

הפתרון נמצא ב־`solution/`. **למה הוא עובד כך:** כל אובייקט מוצא את האחרים לפי תוויות, ולכן ה־ReplicaSet
יודע אילו pods שלו וה־Service יודע לאן לשלוח; pod חדש עם אותה תווית מצטרף לשניהם בלי שאף אחד יגיד לו.
התצורה יושבת ב־ConfigMap והסוד נוצר מבחוץ, כך שה־image זהה בכל סביבה. בדיקת החיות שואלת רק אם התהליך עונה,
ובדיקת המוכנות היא שקובעת אם הוא ב־endpoints; לכן כיבוי המוכנות מוציא pod מהתעבורה ולא הורג אותו.

## מה התרגיל מדגים

- לולאת התאמה בפעולה: pod שנמחק מוחלף בלי שאף אחד ביקש.
- נעילה ב־digest כתנאי של הכלי, וטעינת image ל־node כצעד שאין לו מקבילה באשכול עם registry.
- ההפרדה בין תצורה, סוד ו־image, ולמה base64 אינו הגנה.
- ההבדל בין חיות למוכנות, כפי שהוא נראה ב־EndpointSlice וב־`RESTARTS`.
- DNS של האשכול: שירות אחד מוצא שירות אחר לפי שם, בלי כתובות IP.

</div>
