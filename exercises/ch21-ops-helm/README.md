---
direction: rtl
lang: he
---

<div dir="rtl" lang="he">

# תרגיל פרק 21: `ops-helm`, להפעיל את shorty ולארוז אותו

**מטרה:** להריץ את shorty על kind כמו שירות אמיתי: משאבים ומחלקות QoS, חניקה והריגה בזיכרון שרואים בעיניים,
metrics-server ו־HPA תחת עומס, rollout שבור ו־`rollout undo`, ingress דרך cloud-provider-kind, תרגיל דיבוג
עם ארבע תקלות שתולות, ו־chart של Helm עם סכמת values שעובר lint, template, install, upgrade ו־rollback.

**זמן משוער:** שני ימי לימוד. יום 62: כ־65 דקות בנייה (משאבים, HPA, rollouts). יום 63: כ־75 דקות (ingress,
תרגיל הדיבוג, ה־chart).

**ידע מקדים:** פרק 21; פרק 20 (manifests, kind, `port-forward`); פרק 19 (ה־Dockerfile; התרגיל מביא עותק
משלו); פרק 6 (shorty ו־`cmd/loadgen`). kind, kubectl ו־Helm מותקנים מפרק 2, סעיף 2.10. Docker פועל, עם כ־2 GB
פנויים; עצרי stacks אחרים של compose ואשכולות kind אחרים לפני שמתחילים. הסקריפט `check_static.sh` מריץ את
kubeconform מתוך image, ולכן צריך גם רשת.

## מה בתיקייה הזאת

| קובץ | תפקיד |
| --- | --- |
| `data/shorty/` | הקשר הבנייה: עותק של shorty מתרגיל 19, ועוד `internal/app/stress.go` ובדיקה שלו, ו־`loadgen` בתוך ה־image |
| `data/kind-config.yaml` | אשכול `ch21` של node אחד, node image נעול ב־digest |
| `data/metrics-server/kustomization.yaml` | metrics-server v0.9.0 עם הדגל `--kubelet-insecure-tls` |
| `data/values-kind.yaml` | values להתקנה ידנית של ה־chart ביום 63 |
| `data/debug-nonroot.json` | משתמש שאינו root לקונטיינר זמני של `kubectl debug` |
| `starter/k8s/namespace.yaml`, `configmap.yaml`, `service.yaml`, `qos-pods.yaml` | שלמים |
| `starter/k8s/deployment.yaml` | `TODO 1` (משאבים) ו־`TODO 2` (אסטרטגיית העדכון) |
| `starter/k8s/hpa.yaml`, `starter/k8s/ingress.yaml` | `TODO 3` ו־`TODO 4` |
| `starter/drill/` | ארבעה שירותים שבורים, `app-a` עד `app-d`; לא נוגעים לפני יום 63 |
| `starter/chart/shorty/` | ה־chart: `Chart.yaml`, `values.yaml`, `_helpers.tpl`, Service ו־NOTES שלמים; `TODO 5` עד `TODO 8` |
| `tests/check_static.sh` | kubeconform על `k8s/` ו־`drill/`, helm lint, template, kubeconform על הפלט, וסכמת values |
| `tests/drill_check.sh` | פונה לארבעת שירותי הדיבוג מתוך pod באשכול |
| `tests/smoke.sh` | install, קישור, upgrade, rollback ו־uninstall של ה־chart מול האשכול |
| `solution/` | הפתרון; התיקונים בתרגיל הדיבוג מסומנים `FIX` בהערות. עותקים זהים של `tests/` ב־`starter/tests/` וב־`solution/tests/` |

עד שתמלאי את ה־TODO, `check_static.sh` על התיקייה שלך מדווח `FAIL` בשורות שמפנות אליהם. ה־starter מתקין
בכל זאת: אפשר להתחיל את יום 62 לפני שכתבת שורה.

## יום 62: משאבים, HPA ו־rollouts

**1. העתיקי את נקודת ההתחלה, בני את ה־image והרימי את האשכול.**

<div dir="ltr" align="left">

```bash
mkdir -p ~/senior-course/ch21
cp -r $COURSE/exercises/ch21-ops-helm/starter/. ~/senior-course/ch21/
cp -r $COURSE/exercises/ch21-ops-helm/data ~/senior-course/ch21/
cd ~/senior-course/ch21
docker build -t shorty:ch21 data/shorty
kind create cluster --config data/kind-config.yaml
kind load docker-image shorty:ch21 --name ch21
kubectl get nodes
```

</div>

**2. כתבי את `TODO 1` ו־`TODO 2`** בקובץ `k8s/deployment.yaml`: requests ו־limits, ואסטרטגיית rolling update
עם `progressDeadlineSeconds: 60`. ואז:

<div dir="ltr" align="left">

```bash
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/configmap.yaml -f k8s/deployment.yaml -f k8s/service.yaml
kubectl -n shorty rollout status deploy/shorty
```

</div>

**3. שלוש מחלקות QoS.**

<div dir="ltr" align="left">

```bash
kubectl apply -f k8s/qos-pods.yaml
kubectl -n shorty wait --for=condition=Ready pod -l demo=qos --timeout=60s
kubectl get pods -n shorty -l demo=qos -o custom-columns=NAME:.metadata.name,QOS:.status.qosClass
kubectl get pods -n shorty -l app=shorty -o custom-columns=NAME:.metadata.name,QOS:.status.qosClass
```

</div>

**4. חניקה.** שלושה `port-forward`, ואותה בקשה לשלושה pods: בלי תקרה, `200m` ו־`100m`. הריצי את הלולאה
פעמיים ורשמי בפתק את הזמנים:

<div dir="ltr" align="left">

```bash
kubectl -n shorty port-forward pod/qos-besteffort 18081:8080 >/dev/null &
kubectl -n shorty port-forward deploy/shorty 18082:8080 >/dev/null &
kubectl -n shorty port-forward pod/qos-guaranteed 18083:8080 >/dev/null &
sleep 3
for p in 18081 18082 18083; do curl -s "localhost:$p/stress/cpu?work=200"; done
kill %1 %2 %3
```

</div>

**5. הריגה בזיכרון.** בחרי pod אחד של shorty, ובקשי ממנו להחזיק יותר מהתקרה:

<div dir="ltr" align="left">

```bash
POD=$(kubectl -n shorty get pod -l app=shorty -o jsonpath='{.items[0].metadata.name}')
kubectl -n shorty port-forward pod/$POD 18082:8080 >/dev/null &
sleep 2
curl -s -X POST "localhost:18082/stress/memory?mb=20"
curl -s -X POST "localhost:18082/stress/memory?mb=60"
kill %1
kubectl -n shorty describe pod $POD | sed -n '/Last State/,/Restart Count/p'
```

</div>

**6. metrics-server, ו־`TODO 3`.** התקיני את metrics-server, חכי דקה עד שיש מדדים, וכתבי את ה־HPA:

<div dir="ltr" align="left">

```bash
kubectl apply -k data/metrics-server
kubectl -n kube-system rollout status deploy/metrics-server --timeout=180s
kubectl delete -f k8s/qos-pods.yaml
kubectl top pods -n shorty
kubectl apply -f k8s/hpa.yaml
kubectl -n shorty get hpa shorty
```

</div>

**7. עומס.** הריצי את מחולל העומס בתוך האשכול, ובחלון שני עקבי אחרי ה־HPA. רשמי כמה זמן עבר עד 5 עותקים,
ומתי חזר ל־2:

<div dir="ltr" align="left">

```bash
kubectl -n shorty run loadgen --image=shorty:ch21 --restart=Never --command -- \
  /loadgen -url "http://shorty/stress/cpu?work=20" -c 4 -d 150s
kubectl -n shorty get hpa shorty --watch
```

</div>

<div dir="ltr" align="left">

```bash
kubectl -n shorty logs loadgen
kubectl -n shorty describe hpa shorty | sed -n '/Events/,$p'
kubectl -n shorty delete pod loadgen
```

</div>

**8. rollout שבור, ו־undo.**

<div dir="ltr" align="left">

```bash
kubectl -n shorty annotate deploy/shorty kubernetes.io/change-cause="first release, image shorty:ch21" --overwrite
kubectl -n shorty set image deploy/shorty shorty=shorty:ch21-typo
kubectl -n shorty annotate deploy/shorty kubernetes.io/change-cause="image shorty:ch21-typo" --overwrite
kubectl -n shorty rollout status deploy/shorty --timeout=90s
kubectl -n shorty get pods
kubectl -n shorty get deploy shorty -o jsonpath='{range .status.conditions[*]}{.type}={.status} {.reason}{"\n"}{end}'
kubectl -n shorty rollout history deploy/shorty
kubectl -n shorty rollout undo deploy/shorty
kubectl -n shorty rollout status deploy/shorty
```

</div>

את האשכול משאירים ליום 63. אם את עוצרת כאן ליותר מיום, מחקי אותו, ומחר צעדים 1, 2 ו־6 יחזירו אותו:

<div dir="ltr" align="left">

```bash
kind delete cluster --name ch21
```

</div>

## יום 63: ingress, תרגיל הדיבוג וה־chart

**9. `TODO 4` ו־cloud-provider-kind.** כתבי את ה־Ingress, הפעילי את הרכיב, ומצאי את הפורט במחשב שלך:

<div dir="ltr" align="left">

```bash
docker run -d --name ch21-cloud-provider --network kind \
  -v /var/run/docker.sock:/var/run/docker.sock \
  registry.k8s.io/cloud-provider-kind/cloud-controller-manager:v0.12.0 --enable-lb-port-mapping
kubectl apply -f k8s/ingress.yaml
kubectl -n shorty get ingress --watch
```

</div>

<div dir="ltr" align="left">

```bash
GW=$(docker ps --format '{{.Names}}' | grep '^kindccm-' | head -1)
PORT=$(docker port "$GW" 80/tcp | head -1 | cut -d: -f2)
curl -s localhost:$PORT/healthz
curl -s -X POST localhost:$PORT/v1/links -H 'Content-Type: application/json' -d '{"url":"https://go.dev/","code":"viaing"}'
for i in $(seq 1 10); do curl -s -o /dev/null -w "%{http_code} " localhost:$PORT/viaing; done; echo
```

</div>

על WSL2, לפי ה־README של cloud-provider-kind, אפשר לנסות גם את הכתובת שבעמודה `ADDRESS` ישירות, עם `curl`
לכתובת עצמה. אם היא לא עונה, הפורט הממופה עובד.

**10. תרגיל הדיבוג.** פרסי את ארבעת השירותים השבורים, והריצי את הבדיקה:

<div dir="ltr" align="left">

```bash
kubectl apply -f drill/namespace.yaml
kubectl apply -f drill/
bash tests/drill_check.sh
```

</div>

לכל שירות, לפי סדר הבדיקות של סעיף 21.5, אספי ראיה ורשמי אותה בפתק לפני שאת משנה משהו. ארגז הכלים:

<div dir="ltr" align="left">

```bash
kubectl -n drill get pods
kubectl -n drill describe pod -l app=app-a
kubectl -n drill logs deploy/app-a --previous
kubectl -n drill get endpointslices
kubectl -n drill get svc app-a -o jsonpath='{.spec.selector}{"\n"}'
kubectl -n drill get pods -l app=app-a --show-labels
kubectl -n drill get events --sort-by=.lastTimestamp
```

</div>

ומבפנים, עם קונטיינר זמני (החליפי את שם השירות):

<div dir="ltr" align="left">

```bash
POD=$(kubectl -n drill get pod -l app=app-a -o jsonpath='{.items[0].metadata.name}')
kubectl -n drill debug -i --quiet $POD --image=busybox:1.37.0 --target=shorty \
  --custom=data/debug-nonroot.json -- sh -c 'wget -qO- -T 2 localhost:8080/healthz'
```

</div>

תקני את הקבצים בתיקייה `drill/`, החילי שוב, וחזרי על הבדיקה עד שכולם עוברים:

<div dir="ltr" align="left">

```bash
kubectl apply -f drill/
bash tests/drill_check.sh
```

</div>

**11. ה־chart: `TODO 5` עד `TODO 8`.** כתבי את התבנית של ה־ConfigMap, השלימי את ה־Deployment, כתבי את
ה־HPA המותנה ואת הסכמה. אחרי כל שינוי:

<div dir="ltr" align="left">

```bash
helm lint --strict chart/shorty
helm template ch21 chart/shorty --set autoscaling.enabled=true | less
bash tests/check_static.sh
```

</div>

**12. התקנה, שדרוג, שדרוג שבור, rollback והסרה.**

<div dir="ltr" align="left">

```bash
helm install shorty chart/shorty -n web --create-namespace -f data/values-kind.yaml --wait --timeout 120s
helm list -n web
kubectl -n web get deploy shorty -o jsonpath='{range .metadata.managedFields[*]}{.manager} {.operation}{"\n"}{end}'
helm upgrade shorty chart/shorty -n web -f data/values-kind.yaml --set config.logLevel=info --wait
helm upgrade shorty chart/shorty -n web -f data/values-kind.yaml \
  --set image.tag=ch21-typo --rollback-on-failure --timeout 60s
helm history shorty -n web
helm rollback shorty 1 -n web --wait
helm get values shorty -n web
helm install bad chart/shorty -n web --set replicaCount=two   # expected to fail: the schema rejects it
helm uninstall shorty -n web --wait
```

</div>

**13. בדיקת העשן, וניקוי.**

<div dir="ltr" align="left">

```bash
bash tests/smoke.sh
docker rm -f ch21-cloud-provider
docker ps -a --format '{{.Names}}' | grep '^kindccm-' | xargs -r docker rm -f
kind delete cluster --name ch21
git -C ~/senior-course add ch21 && git -C ~/senior-course commit -m "ch21: resources, HPA, rollouts, ingress, drill, chart"
```

</div>

## פלט צפוי

הזמנים והשמות שונים אצלך; אלה מההרצה שבה נכתב הספר. בצעד 4 חשוב היחס בין שלושת הזמנים, לא המספרים עצמם:

<div dir="ltr" align="left">

```text
work 200, wall 313 ms, checksum 1
work 200, wall 1496 ms, checksum 1
work 200, wall 3172 ms, checksum 1
```

</div>

צעד 5:

<div dir="ltr" align="left">

```text
holding 20 MiB
    Last State:     Terminated
      Reason:       OOMKilled
      Exit Code:    137
    Ready:          True
    Restart Count:  1
```

</div>

צעד 7, הסוף של `logs` ושל האירועים:

<div dir="ltr" align="left">

```text
requests 1864, failed 0, 12 req/s over 2m30s with 4 workers
Normal  SuccessfulRescale  4m6s  horizontal-pod-autoscaler  New size: 5; reason: cpu resource utilization (percentage of request) above target
Normal  SuccessfulRescale  36s   horizontal-pod-autoscaler  New size: 2; reason: All metrics below target
```

</div>

צעד 10 לפני התיקונים, ואחריהם:

<div dir="ltr" align="left">

```text
FAIL  app-a
FAIL  app-b
FAIL  app-c
FAIL  app-d
4 service(s) still broken
...
PASS  app-d answers /healthz through its Service
drill solved
```

</div>

צעד 11, הסוף של `check_static.sh`, וצעד 13:

<div dir="ltr" align="left">

```text
PASS  schema rejects replicas=3

all static checks passed
PASS  install, create and resolve
PASS  upgrade to revision 2
PASS  rollback to revision 1 (as revision 3)
smoke test passed
```

</div>

## רשימת בדיקה

- [ ] שלוש מחלקות QoS הופיעו, ו־shorty עצמו `Burstable`.
- [ ] שלושה זמני חניקה רשומים בפתק, והיחס ביניהם מוסבר.
- [ ] `OOMKilled` וקוד 137 בסטטוס של pod, ובפתק מה אבד בהריגה.
- [ ] ה־HPA הציג ניצול, הגדיל ל־5 והקטין ל־2.
- [ ] ה־rollout השבור נתקע עם `ProgressDeadlineExceeded`, שני pods ישנים שירתו, ו־undo החזיר.
- [ ] shorty ענה דרך ה־ingress, ובפתק ההסבר ל־404 בחצי מהבקשות.
- [ ] ארבע תקלות נמצאו ותוקנו, והראיה לכל אחת רשומה.
- [ ] `check_static.sh` ו־`smoke.sh` עוברים, האשכול וקונטיינרי cloud-provider-kind נמחקו.

## טעויות נפוצות ותיקונים

- **ה־HPA מציג `cpu: <unknown>/60%`.** או ש־metrics-server עוד לא אסף (חכי דקה, ובדקי `kubectl top pods`), או
  שחסר request למעבד ב־Deployment (`TODO 1`).
- **metrics-server תקוע ב־`0/1` עם `x509: cannot validate certificate`.** התקנת את הקובץ הרשמי ישירות. החילי
  את `data/metrics-server` עם `-k`; הוא מוסיף את `--kubelet-insecure-tls`.
- **`curl` לפורט של ה־ingress מחזיר רשימה של פקודות ניהול.** cloud-provider-kind רץ בלי
  `--enable-lb-port-mapping`, ומופה רק פורט הניהול של envoy. מחקי את ה־Ingress, את `ch21-cloud-provider` ואת
  קונטיינרי `kindccm-`, והפעילי מחדש עם הדגל.
- **`kubectl debug` לא מחזיר כלום ונתקע.** הקונטיינר הזמני נכשל ב־`container has runAsNonRoot and image will
  run as root`. הוסיפי `--custom=data/debug-nonroot.json`. לקונטיינר זמני שנכשל אין מחיקה; הוא נשאר בסטטוס עד
  שה־pod מתחלף.
- **`drill_check.sh` נכשל רגע אחרי תיקון של Service.** השינוי לוקח כמה שניות להגיע ל־kube-proxy. הסקריפט מנסה
  שלוש פעמים; אם עדיין נכשל, הריצי שוב אחרי עשר שניות לפני שאת מחפשת באג.
- **`helm upgrade` נכשל ב־`field is immutable` על ה־selector.** הכנסת תווית משתנה, כמו גרסת ה־chart,
  ל־`shorty.selectorLabels`. היא שייכת ל־`shorty.labels` בלבד; הסירי את ה־release והתקיני מחדש.

## רמזים

- ב־`TODO 2`, שני המספרים יושבים תחת `strategy.rollingUpdate`, ו־`progressDeadlineSeconds` ישירות תחת `spec`.
- ב־`TODO 3`, הצורה של `metrics` היא רשימה; כל פריט עם `type: Resource` ובתוכו `resource.target`.
- ב־`TODO 5`, `ternary "on" "off" .Values.config.stress` מחזיר את הראשון כשהערך אמת; ואז `quote`.
- ב־`TODO 6`, `toYaml .Values.resources | nindent 12`: ההזחה היא של השורה שמתחת ל־`resources:`.
- ב־`TODO 8`, התחילי מ־`additionalProperties: false` בשורש, והוסיפי מפתח אחרי מפתח עד שכל הבדיקות עוברות.
- בתרגיל הדיבוג: שירות אחד קורס, אחד לא מוכן, ושניים רצים ומוכנים ולא עונים. ה־EndpointSlices מבדילים
  בין השניים האחרונים.

## הפתרון

הפתרון נמצא ב־`solution/`: ה־manifests ב־`k8s/`, השירותים המתוקנים ב־`drill/` עם הערת `FIX` שמסבירה כל תקלה
והראיה שמגלה אותה, וה־chart ב־`chart/shorty/`. **למה הוא עובד כך:** ה־requests נותנים ל־HPA מכנה ול־scheduler
הבטחה, וה־limits הופכים חריגה לתופעה שרואים; `maxUnavailable: 0` יחד עם בדיקת המוכנות משאיר גרסה שבורה
בצד בלי לגעת בעותקים שעובדים. ב־chart, תוויות ה־selector קבועות, ה־HPA מקבל את מספר העותקים כשהוא פעיל, והגיבוב של
ה־ConfigMap מגלגל pods כשהתצורה משתנה; הסכמה עוצרת ערכים שגויים לפני שהם מגיעים לשרת.

## מה התרגיל מדגים

- ההבדל בין חניקה להריגה, כמספרים שמדדת ולא כהגדרה.
- autoscaling שמחובר ל־requests, ולמה מצב בזיכרון לא מתרחב.
- rollout בטוח כתוצאה של בדיקת מוכנות, ו־undo כפתרון חירום ולא כתיקון.
- ingress בשנת 2026: בקר שמגיע עם הסביבה, ו־Gateway API כיורש.
- דיבוג כשיטה: ראיה לפני שינוי, ו־`kubectl debug` כשאין shell.
- Helm כחבילה עם היסטוריה, וסכמה ובדיקות סטטיות שתופסות טעויות לפני האשכול.

</div>
