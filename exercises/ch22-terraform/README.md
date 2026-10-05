---
direction: rtl
lang: he
---

<div dir="rtl" lang="he">

# תרגיל פרק 22: `terraform`, תשתית מקבצים מול kind

**מטרה:** לנהל תשתית על אשכול kind מקבצים בלבד. ביום הראשון, namespace ו־ConfigMap עם ה־provider של
Kubernetes: plan, apply, שינוי של ערך אחד, סטייה שיוצרים ביד, ו־destroy. ביום השני, מודול מקומי שיוצר
namespace ומתקין לתוכו את ה־chart של shorty מפרק 21 דרך ה־provider של Helm, עם משתנים ו־outputs. התיקייה
`optional-aws` אופציונלית, ולא הורצה מול AWS על ידי הכותב.

**זמן משוער:** שני ימי לימוד. יום 64: כ־50 דקות בנייה (חלק 1). יום 65: כ־60 דקות (חלק 2, וקריאת סעיף 22.5).

**ידע מקדים:** פרק 22; פרק 21 (ה־chart ותיקיית `data/shorty` שממנה בונים את ה־image); פרק 20 (kind
וההקשר של kubectl). Terraform, kind, kubectl ו־Helm מותקנים מפרק 2, סעיף 2.10. ההרצה הראשונה של
`terraform init` מורידה את ה־providers, ולכן צריך רשת. Docker פועל, עם כ־2 GB פנויים.

## מה בתיקייה הזאת

| קובץ | תפקיד |
| --- | --- |
| `data/kind-config.yaml` | אשכול `ch22` של node אחד, node image נעול ב־digest |
| `starter/part1/versions.tf`, `variables.tf`, `.terraform.lock.hcl` | שלמים: provider נעול, ארבעה משתנים, טביעות לארבע פלטפורמות |
| `starter/part1/main.tf` | ה־provider שלם; `TODO 1` (namespace) ו־`TODO 2` (ConfigMap) |
| `starter/part1/outputs.tf` | `TODO 3` |
| `starter/part2/versions.tf`, `backend.tf`, `variables.tf`, `terraform.tfvars.example` | שלמים |
| `starter/part2/providers.tf` | ה־provider של Kubernetes שלם; `TODO 4` (ה־provider של Helm) |
| `starter/part2/modules/shorty/` | `versions.tf` ו־`variables.tf` שלמים; `TODO 5` ו־`TODO 6` ב־`main.tf` וב־`outputs.tf` |
| `starter/part2/main.tf`, `outputs.tf` | `TODO 7` (קריאה למודול) ו־`TODO 8` |
| `tests/check_static.sh` | `fmt -check` ו־`validate` על עותק זמני של כל חלק, ובדיקה שה־TODO נכתבו |
| `tests/smoke.sh` | הכול מול האשכול, על עותקים זמניים: apply, plan ריק, שינוי אחד, סטייה, destroy, ואז המודול |
| `solution/part1/`, `solution/part2/` | הפתרון; ב־`part2/modules/shorty/chart` עותק של ה־chart מפתרון תרגיל 21 |
| `solution/optional-aws/` | אופציונלי, סעיף 22.7: דלי S3 ודוגמה ל־backend מרוחק. עבר `validate` בלבד |

עותקים זהים של `tests/` יושבים ב־`starter/tests/` וב־`solution/tests/`. הסקריפטים עובדים על עותקים זמניים,
ולכן לא משאירים `.terraform/` או קובץ state בתיקייה שלך.

## יום 64: providers, resources, state, plan ו־apply

**1. העתיקי, בני את ה־image והרימי את האשכול.** ה־image הוא אותו `shorty:ch21` מתרגיל 21; אם הוא עדיין
קיים ב־Docker, הבנייה מהירה.

<div dir="ltr" align="left">

```bash
mkdir -p ~/senior-course/ch22
cp -r $COURSE/exercises/ch22-terraform/starter/. ~/senior-course/ch22/
cp -r $COURSE/exercises/ch22-terraform/data ~/senior-course/ch22/
cd ~/senior-course/ch22
docker build -t shorty:ch21 $COURSE/exercises/ch21-ops-helm/data/shorty
kind create cluster --config data/kind-config.yaml
kind load docker-image shorty:ch21 --name ch22
```

</div>

**2. init ו־validate של ה־starter.** הוא תקין כבר עכשיו; הוא פשוט לא יוצר כלום.

<div dir="ltr" align="left">

```bash
cd part1
terraform init
terraform validate
terraform plan
```

</div>

**3. כתבי את `TODO 1` עד `TODO 3`,** ואז הריצי את הסדר המלא. קראי את ה־plan לפני שאת עונה `yes`:

<div dir="ltr" align="left">

```bash
terraform fmt -check || terraform fmt
terraform validate
terraform plan
terraform apply
terraform state list
terraform output
kubectl -n tf-demo get configmap shorty-settings -o yaml
```

</div>

**4. plan ריק, ושינוי אחד.**

<div dir="ltr" align="left">

```bash
terraform plan -detailed-exitcode; echo "exit $?"
terraform plan -var log_level=debug
```

</div>

**5. סטייה.** פתחי את ה־ConfigMap בעורך, שני את `SHORTY_LOG_LEVEL` ל־`error`, והוסיפי שורה `EXTRA: by-hand`
תחת `data`. שמרי וסגרי, ואז:

<div dir="ltr" align="left">

```bash
kubectl -n tf-demo edit configmap shorty-settings
terraform plan -refresh-only
terraform plan
terraform apply
```

</div>

**6. destroy.**

<div dir="ltr" align="left">

```bash
terraform destroy
kubectl get namespace tf-demo
cd ..
```

</div>

## יום 65: משתנים, outputs, מודול וה־provider של Helm

**7. העתיקי את ה־chart שלך מתרגיל 21 לתוך המודול.** אם לא סיימת אותו, השתמשי בפתרון.

<div dir="ltr" align="left">

```bash
cp -r ~/senior-course/ch21/chart/shorty part2/modules/shorty/chart
# or: cp -r $COURSE/exercises/ch21-ops-helm/solution/chart/shorty part2/modules/shorty/chart
```

</div>

**8. כתבי את `TODO 4` עד `TODO 8`.** הסדר הנוח: ה־provider של Helm, ה־namespace וה־release במודול,
ה־outputs של המודול, הקריאה מהשורש, וה־outputs של השורש. בדקי אחרי כל קובץ:

<div dir="ltr" align="left">

```bash
cd part2
terraform init
terraform fmt -check -recursive || terraform fmt -recursive
terraform validate
```

</div>

**9. apply אחד, ובדיקה שהשירות עונה.**

<div dir="ltr" align="left">

```bash
terraform plan
terraform apply
terraform state list
helm list -n shorty-tf
$(terraform output -raw port_forward) >/dev/null &
sleep 3
curl -s localhost:8080/readyz
kill %1
```

</div>

**10. שינוי דרך משתנה, ומה ה־plan לא רואה.**

<div dir="ltr" align="left">

```bash
terraform plan -var replicas=3
kubectl -n shorty-tf get deploy shorty -o jsonpath='{range .metadata.managedFields[*]}{.manager} {.operation}{"\n"}{end}'
kubectl -n shorty-tf scale deploy shorty --replicas=4
terraform plan
```

</div>

**11. destroy, ובדיקות.**

<div dir="ltr" align="left">

```bash
terraform destroy
helm list -n shorty-tf
kubectl get namespace shorty-tf
cd ..
bash tests/check_static.sh
bash tests/smoke.sh
kind delete cluster --name ch22
```

</div>

**12. קראי את סעיף 22.5,** ואם את בוחרת, את סעיף 22.7. בלי חשבון AWS אפשר רק לבדוק את התצורה:

<div dir="ltr" align="left">

```bash
cp -r $COURSE/exercises/ch22-terraform/solution/optional-aws ~/senior-course/ch22/
cd ~/senior-course/ch22/optional-aws
terraform init -backend=false
terraform validate
```

</div>

**13. ודאי שאין state ב־Git, ושמרי.**

<div dir="ltr" align="left">

```bash
cd ~/senior-course
git status --short ch22 | grep -E 'tfstate|\.terraform/' && echo "remove these before committing"
printf '.terraform/\n*.tfstate\n*.tfstate.backup\nterraform.tfvars\n' >> .gitignore
git add .gitignore ch22 && git commit -m "ch22: terraform against kind, module with the helm provider"
```

</div>

## פלט צפוי

מההרצה שבה נכתב הספר. צעד 3, הסוף של `plan` ושל `apply`:

<div dir="ltr" align="left">

```text
Plan: 2 to add, 0 to change, 0 to destroy.
...
Apply complete! Resources: 2 added, 0 changed, 0 destroyed.

Outputs:

config_map = "tf-demo/shorty-settings"
namespace = "tf-demo"
```

</div>

צעד 4: קוד 0 ל־plan הראשון, ואז שינוי אחד:

<div dir="ltr" align="left">

```text
No changes. Your infrastructure matches the configuration.
...
exit 0
          ~ "SHORTY_LOG_LEVEL" = "info" -> "debug"
Plan: 0 to add, 1 to change, 0 to destroy.
```

</div>

צעד 5, ה־plan אחרי העריכה:

<div dir="ltr" align="left">

```text
          - "EXTRA"            = "by-hand" -> null
          ~ "SHORTY_LOG_LEVEL" = "error" -> "info"
Plan: 0 to add, 1 to change, 0 to destroy.
```

</div>

צעד 9 וצעד 10:

<div dir="ltr" align="left">

```text
module.shorty.helm_release.shorty
module.shorty.kubernetes_namespace_v1.this
ready
terraform-provider-helm_v3.3.0_x5 Update
kube-controller-manager Update
No changes. Your infrastructure matches the configuration.
```

</div>

צעד 11:

<div dir="ltr" align="left">

```text
all static checks passed
PASS  part1 apply
PASS  namespace and ConfigMap exist
PASS  second plan is empty
PASS  changing one variable plans exactly one change
PASS  drift made with kubectl shows up in the plan
PASS  part1 destroy removes the namespace
PASS  part2 apply installs the release through the module
PASS  the service answers
PASS  outputs
PASS  part2 destroy removes the release and the namespace
smoke test passed
```

</div>

## רשימת בדיקה

- [ ] ה־plan אחרי שינוי של `log_level` הראה שינוי אחד בלבד, ו־`-detailed-exitcode` החזיר 0 בלי שינוי.
- [ ] העריכה ב־`kubectl edit` הופיעה ב־plan, ובפתק כתוב מה ה־apply עשה איתה.
- [ ] destroy של חלק 1 השאיר את `tf-demo` מחוק.
- [ ] apply אחד בחלק 2 יצר namespace ו־release, והשירות ענה ב־`ready`.
- [ ] ה־plan אחרי `kubectl scale` היה ריק, ובפתק למה.
- [ ] destroy של חלק 2 הסיר את ה־release ואת ה־namespace.
- [ ] `check_static.sh` ו־`smoke.sh` עוברים, האשכול נמחק, ואין קובץ state ב־Git.

## טעויות נפוצות ותיקונים

- **`Error: Unsupported block type` על `kubernetes` ב־provider של Helm.** כתבת בלוק, כמו בגרסה 2. בגרסה 3
  זה אובייקט: `kubernetes = { ... }`. שימי לב ש־`validate` מגלה את זה רק כשמשהו בתצורה כבר משתמש
  ב־provider.
- **ה־apply הראשון נכשל לפעמים ב־`namespaces "tf-demo" not found`.** ה־ConfigMap כתב את שם
  ה־namespace כמחרוזת. הפני ל־`kubernetes_namespace_v1.demo.metadata[0].name`, ו־Terraform יסדר את הסדר.
- **שגיאה על ההקשר `kind-ch22` ב־plan.** האשכול לא קיים, או נוצר בשם אחר. בדקי `kind get clusters`, או הריצי
  עם `-var kube_context=...`.
- **ה־release נתקע עד ה־timeout ונכשל.** ה־image לא נטען ל־`ch22`, וה־pods ב־`ErrImagePull`. הריצי את
  `kind load` מצעד 1 ונסי שוב.
- **שגיאה על נתיב ה־chart בחלק 2.** ה־chart לא הועתק לתוך `modules/shorty/chart`, או שהועתק לתיקייה בשם אחר.
  חזרי לצעד 7.
- **`terraform.tfstate` מופיע ב־`git status`.** הוסיפי את השורות של צעד 13 ל־`.gitignore` לפני ה־commit.

## רמזים

- ב־`TODO 1`, ל־`metadata` יש `name` ו־`labels`; התווית היא מפה עם מפתח במרכאות, כי יש בו נקודות ולוכסן.
- ב־`TODO 2`, `data` היא מפה של מחרוזות, כמו ב־YAML של ConfigMap.
- ב־`TODO 3`, מחרוזת עם שני ערכים נבנית ב־`"${a}/${b}"`.
- ב־`TODO 6`, השדה `values` הוא רשימה של מחרוזות YAML; `yamlencode({...})` מייצרת אחת מאובייקט.
- ב־`TODO 6`, outputs של `helm_release` נמצאים תחת `metadata`, למשל `helm_release.shorty.metadata.revision`.
- ב־`TODO 7`, `source` של מודול מקומי מתחיל ב־`./`, וכל משתנה בלי `default` במודול חייב לקבל ערך.

## הפתרון

הפתרון נמצא ב־`solution/part1` וב־`solution/part2`, והתצורה האופציונלית ב־`solution/optional-aws`. **למה הוא
עובד כך:** כל resource מפנה ל־resource שהוא תלוי בו, ולכן Terraform בונה גרף ויוצר ומוחק בסדר הנכון בלי
`depends_on`. המודול מצהיר על providers ולא מגדיר אותם, כך שהחיבור לאשכול נקבע פעם אחת בשורש. ה־release
מקבל את ה־values שלו כמסמך YAML אחד, והסכמה של ה־chart בודקת אותו כאילו הועבר ב־`-f`; `wait = true` מחזיק
את ה־apply עד שה־pods מוכנים, ולכן apply שהצליח הוא שירות שעונה.

## מה התרגיל מדגים

- תצורה הצהרתית: הרצה שנייה בלי שינוי לא עושה כלום, ושינוי אחד הוא שורה אחת ב־plan.
- state כקישור בין קוד לעולם, ולמה הוא לא שייך ל־Git.
- זיהוי סטייה, והמגבלה שלו כשה־resource הוא release של Helm.
- מודול כממשק: משתנים, outputs, providers שמוגדרים בשורש.
- נעילת גרסאות של providers וקובץ נעילה שעובד בכמה פלטפורמות.

</div>
