---
direction: rtl
tags: [senior-course, book]
chapter: 21
---
# הפעלת Kubernetes ו־Helm

- **requests:** הבטחה ל־scheduler ומכנה ל־HPA; בלי request למעבד אין autoscaling.
- **limits:** מעבד נחנק ואיטי; זיכרון נהרג, `OOMKilled`, קוד 137, ומה שבזיכרון אבד.
- **QoS:** Guaranteed, Burstable, BestEffort; קובע מי מפונה ראשון.
- **HPA:** צריכה חלקי request מול יעד; חלון ייצוב להקטנה; הוא הבעלים של מספר העותקים.
- **metrics-server על kind:** תעודת kubelet חתומה עצמית; הדגל לבדיקות בלבד.
- **rollout:** `maxUnavailable: 0` ובדיקת מוכנות משאירים גרסה שבורה תקועה; `ProgressDeadlineExceeded`.
- **undo:** כפתור חירום; התיקון האמיתי הוא commit.
- **ingress:** ingress-nginx פרש במרץ 2026; על kind, cloud-provider-kind; ב־Mac מיפוי פורטים.
- **מצב בזיכרון:** שני עותקים, שני מאגרים; חצי מהבקשות 404.
- **דיבוג:** get, describe, logs `--previous`, endpointslices, events, debug.
- **Helm:** values משתנים, templates זהים; revisions ו־rollback; בגרסה 4 server-side apply.
- **בדיקות:** סכמת values תופסת טעויות כתיב; kubeconform חוקיות; smoke נכונות.
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

קישורים: [[20-kubernetes-on-kind]] · [[22-terraform]]
