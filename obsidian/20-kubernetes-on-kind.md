---
direction: rtl
tags: [senior-course, book]
chapter: 20
---
# Kubernetes על kind

- **מצב רצוי:** כותבים מה צריך לרוץ; בקרים בלולאה סוגרים כל פער.
- **רכיבים:** API server הוא הדלת היחידה; etcd שומר; scheduler מציב; kubelet מריץ ובודק.
- **kind:** nodes כקונטיינרים; node image נעול ב־digest; image מקומי נטען עם `kind load`.
- **תגית:** `latest` או בלי תגית פירושו משיכה תמיד; ב־kind זה `ErrImagePull`.
- **pod:** בן חלוף, IP משלו; pod בודד אף אחד לא מחזיר.
- **Deployment:** מנהל ReplicaSets לפי גרסאות; ReplicaSet מחזיק את מספר העותקים.
- **תוויות:** הקשר בין Deployment, ReplicaSet, pods ו־Service.
- **Service:** שם ו־IP יציבים, רק ל־pods מוכנים; DNS לפי שם, namespace ו־`svc.cluster.local`.
- **port-forward:** pod אחד, בלי איזון, לפיתוח.
- **תצורה:** ConfigMap נקרא בעלייה; Secret הוא base64, לא הצפנה; לא ב־Git.
- **בדיקות:** הפעלה מגינה על עלייה; חיות מפעילה מחדש, בלי תלויות; מוכנות מוציאה מהתעבורה בלי הפעלה מחדש.
- **שאלה פתוחה:** ___

<div dir="ltr" align="left">

```text
cluster     kind create cluster --name ch20 --config kind-config.yaml ; kind delete cluster --name ch20
node image  kindest/node:v1.37.0@sha256:a1ed56cf...   (from: github.com/kubernetes-sigs/kind/releases/tag/v0.33.0)
load        docker build -t shorty:ch20 data/shorty ; kind load docker-image shorty:ch20 --name ch20
apply       kubectl apply -f k8s/ ; kubectl -n ch20 rollout status deployment/shorty
look        kubectl -n ch20 get deploy,rs,pods -l app=shorty ; kubectl -n ch20 describe pod POD
endpoints   kubectl -n ch20 get endpointslices -l kubernetes.io/service-name=shorty
dns         <service>.<namespace>.svc.cluster.local
forward     kubectl -n ch20 port-forward svc/shorty 8080:80   (one pod, no balancing)
secret      kubectl -n ch20 create secret generic receiver-secret --from-literal=signing-secret=change-me
decode      kubectl -n ch20 get secret receiver-secret -o jsonpath='{.data.signing-secret}' | base64 -d
config      kubectl -n ch20 rollout restart deployment/shorty   (env is read at start)
probes      startup /healthz -> liveness /healthz (restart) ; readiness /readyz (out of endpoints)
drill       kubectl -n ch20 port-forward pod/POD 6060:6060 ; curl -X POST 'localhost:6060/debug/ready?state=off'
mine        replacement pod appeared after ___ s   readiness out after ___ s   restarts during drill ___
```

</div>

קישורים: [[19-docker-done-right]] · [[21-kubernetes-operations-helm]]
