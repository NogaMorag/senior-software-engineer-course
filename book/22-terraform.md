---
direction: rtl
lang: he
title: "פרק 22: Terraform"
chapter: 22
part: "חלק ה: Cloud-native"
---

<div dir="rtl" lang="he">

# פרק 22: Terraform

בפרק 21 יצרת אובייקטים באשכול בפקודות: `kubectl apply`, `helm install`. זה עובד כל עוד את זוכרת מה הרצת.
Terraform מחליף את הזיכרון הזה בקבצים: את כותבת מה צריך להתקיים, הכלי משווה למה שקיים, מראה לך את ההבדל,
ורק אז משנה. המילה Terraform הופיעה ב־12 מתוך 30 המודעות הבכירות שנבדקו ביד לספר, ובשבע מתוך שמונה מודעות
ה־DevOps וה־SRE (סעיף 1.4 בקובץ המחקר `00-stage1/01-market-research.md`). הפרק מריץ אותו מול אשכול kind,
כך שכל מה שתלמדי רץ באמת ובחינם.

## מה נלמד בפרק

- מה זה תשתית כקוד, ומה היא נותנת מעבר לסקריפט של פקודות.
- providers, resources ו־state: שלושת המושגים שכל השאר נשען עליהם.
- `plan` ו־`apply` מול kind: namespace ו־ConfigMap, שינוי אחד, סטייה שנוצרה ביד, ו־`destroy`.
- משתנים, outputs ומודול מקומי שאפשר להשתמש בו שוב.
- state מרוחק ונעילות: למה צוותים לא שומרים state במחשב, ולמה לא ב־Git.
- ה־provider של Helm: להתקין את ה־chart מפרק 21 דרך Terraform, ומתי לא כדאי.
- אופציונלי: תצורה קטנה ל־AWS, עם אזהרות על עלות.

## ידע מקדים

- פרק 21: ה־chart של shorty וה־image `shorty:ch21`. התרגיל מעתיק את ה־chart לתוך מודול, ובונה את ה־image
  מהתיקייה של תרגיל 21.
- פרק 20: namespaces, ConfigMaps וההקשר `kind-…` בקובץ `~/.kube/config`.
- פרק 2, סעיף 2.10: Terraform מותקן בגרסה הנעולה. בהרצה הראשונה של `terraform init` הוא מוריד providers
  מהאינטרנט.
- כ־2 GB פנויים במכונה של Docker. האשכול הוא node אחד.

## 22.1 תשתית כקוד

תשתית כקוד (infrastructure as code) פירושה שהמצב הרצוי של התשתית כתוב בקבצים שנמצאים ב־Git, ושכלי אחד
אחראי להביא את העולם למצב הזה. את המונח „מצב רצוי” פגשת בפרק 20: זה אותו רעיון, מחוץ לאשכול.

| דרך | מה נשמר | מה קורה בהרצה השנייה | מי יודע מה קיים |
| --- | --- | --- | --- |
| פקודות ביד | כלום, או היסטוריה של shell | תלוי בזיכרון שלך | אף אחד |
| סקריפט של פקודות | רצף פעולות | לרוב נכשל, כי המשאב כבר קיים | הסקריפט, אם הצליח |
| תשתית כקוד | מצב רצוי | לא עושה כלום, אם אין הבדל | הקבצים וה־state |

הטור השלישי הוא העיקר. תצורה של Terraform היא הצהרתית (declarative): היא לא אומרת „צור namespace”, היא אומרת
„צריך להתקיים namespace בשם הזה”. לכן הרצה שנייה בלי שינוי היא פעולה ריקה, ושינוי של שורה אחת מתורגם לפעולה
אחת. זו אידמפוטנטיות (פרק 9) של תשתית שלמה.

התמורה בעבודת צוות היא סקירה. שינוי בתשתית הוא שינוי בקובץ, הוא נסקר ב־pull request כמו קוד, ולפני שמישהו
מאשר, `terraform plan` מראה בדיוק מה ייווצר, ישתנה וימחק. זה ההבדל בין „אני חושבת שזה מוסיף רק כלל אחד”
לבין רשימה כתובה.

## 22.2 providers, resources ו־state

שלושה מושגים, וכל אחד מהם הוא שאלת ראיון בפני עצמו:

| מושג | מה הוא | בתרגיל |
| --- | --- | --- |
| provider | תוסף שיודע לדבר עם API אחד ולתרגם resources לקריאות | `hashicorp/kubernetes`, `hashicorp/helm` |
| resource | אובייקט אחד שהתצורה מצהירה עליו | `kubernetes_namespace_v1`, `helm_release` |
| state | קובץ JSON שבו Terraform זוכר מה הוא יצר ואיך זה נראה בפעם האחרונה | `terraform.tfstate` |

### 22.2.1 נעילת גרסאות

כל תצורה מתחילה בהצהרה אילו providers היא צריכה ובאיזו גרסה. זה הקובץ `solution/part1/versions.tf`:

<div dir="ltr" align="left">

```hcl
terraform {
  required_version = ">= 1.16.0"

  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "3.3.0" # exact pin; .terraform.lock.hcl records the checksums
    }
  }
}
```

</div>

הפקודה `terraform init` מורידה את ה־provider וכותבת את הקובץ `.terraform.lock.hcl` עם הטביעות שלו. את קובץ
הנעילה מכניסים ל־Git, בדיוק כמו `go.sum`. בקובץ שבתרגיל יש טביעות לארבע פלטפורמות, כדי שאותו קובץ יעבוד גם
ב־WSL2 וגם ב־Mac; כך מפיקים אותו:

<div dir="ltr" align="left">

```bash
terraform providers lock -platform=linux_amd64 -platform=linux_arm64 -platform=darwin_arm64 -platform=darwin_amd64
```

</div>

### 22.2.2 provider ו־resources

זה הקובץ `solution/part1/main.tf` כולו:

<div dir="ltr" align="left">

```hcl
# Day 64: the kubernetes provider against the kind cluster "ch22".

provider "kubernetes" {
  config_path    = var.kubeconfig
  config_context = var.kube_context
}

resource "kubernetes_namespace_v1" "demo" {
  metadata {
    name = var.namespace
    labels = {
      "app.kubernetes.io/managed-by" = "terraform"
    }
  }
}

resource "kubernetes_config_map_v1" "settings" {
  metadata {
    name      = "shorty-settings"
    namespace = kubernetes_namespace_v1.demo.metadata[0].name # a reference: Terraform creates the namespace first
  }

  data = {
    SHORTY_BASE_URL  = "http://localhost:8080"
    SHORTY_LOG_LEVEL = var.log_level
  }
}
```

</div>

לכל resource שני שמות: הסוג (`kubernetes_config_map_v1`), שקובע ה־provider, והשם המקומי (`settings`), שקובעת
את. יחד הם הכתובת שלו בתצורה ובקובץ ה־state.

השורה עם ההערה היא הרעיון החשוב בקובץ. ה־ConfigMap לא כותב את שם ה־namespace; הוא מפנה ל־resource של
ה־namespace. מההפניות Terraform בונה גרף תלויות, ומהגרף הוא גוזר את הסדר: קודם namespace, אחר כך ConfigMap,
וב־destroy בסדר ההפוך. אילו כתבת את המחרוזת, שני ה־resources היו נוצרים במקביל, וה־ConfigMap היה נכשל לפעמים.

> **שימי לב:** בגרסה 3 של ה־provider של Kubernetes, השמות בלי הסיומת `_v1` הוצאו משימוש, למשל
> `kubernetes_namespace` ו־`kubernetes_config_map`. דוגמאות ישנות ברשת משתמשות בהם. השתמשי בשמות עם `_v1`
> (נכון לאוקטובר 2026; ראי נספח ג).

### 22.2.3 מה יש ב־state

אחרי `apply` נוצר בתיקייה הקובץ `terraform.tfstate`. ההתחלה שלו, מההרצה שבה נכתב הספר:

<div dir="ltr" align="left">

```text
{
  "version": 4,
  "terraform_version": "1.16.5",
  "serial": 3,
  "lineage": "fa78b475-aa68-b7ca-786b-47aab616e6c1",
  "outputs": {
    "config_map": {
      "value": "tf-demo/shorty-settings",
```

</div>

ה־state מקשר כל כתובת בתצורה לאובייקט אמיתי: הכתובת `kubernetes_namespace_v1.demo` היא ה־namespace
`tf-demo`, עם ה־uid שלו. בלי הקישור הזה Terraform לא יודע שהוא יצר את ה־namespace. הוא ינסה ליצור אותו שוב
ויקבל „already exists”, או גרוע מזה, ישכח שהוא קיים ולא ימחק אותו לעולם. **לאבד state פירושו לאבד את הבעלות
על התשתית.**

<div dir="ltr" align="left">

```bash
terraform state list
```

</div>

<div dir="ltr" align="left">

```text
kubernetes_config_map_v1.settings
kubernetes_namespace_v1.demo
```

</div>

> **טעות נפוצה:** להכניס את `terraform.tfstate` ל־Git. הקובץ מכיל כל ערך של כל resource, כולל סודות בטקסט
> גלוי (סיסמה של מסד נתונים, למשל), ושתי מפתחות שעובדות במקביל על עותקים שונים שלו דורסות זו את זו. הקובץ
> `.gitignore` של מאגר הקורס מתעלם ממנו; סעיף 22.5 מראה איפה כן שומרים אותו.

## 22.3 plan ו־apply מול kind

הסדר תמיד זהה, וכדאי להכיר כל צעד:

| פקודה | מה היא עושה |
| --- | --- |
| `terraform init` | מורידה providers, מאתחלת את ה־backend של ה־state |
| `terraform fmt -check` | בודקת עיצוב אחיד, בלי לשנות; בלי `-check` היא מתקנת |
| `terraform validate` | בודקת תחביר והפניות, בלי לפנות לשום API |
| `terraform plan` | קוראת את המצב האמיתי, משווה לתצורה ול־state, ומדפיסה את ההבדל |
| `terraform apply` | מחשבת plan, מחכה לאישור, ומבצעת |
| `terraform destroy` | מוחקת את כל מה שב־state |

### 22.3.1 ה־plan הראשון

<div dir="ltr" align="left">

```bash
kind create cluster --config data/kind-config.yaml
cd ~/senior-course/ch22/part1
terraform init
terraform plan
```

</div>

<div dir="ltr" align="left">

```text
  # kubernetes_config_map_v1.settings will be created
  + resource "kubernetes_config_map_v1" "settings" {
      + data = {
          + "SHORTY_BASE_URL"  = "http://localhost:8080"
          + "SHORTY_LOG_LEVEL" = "info"
        }
      + id   = (known after apply)
...
Plan: 2 to add, 0 to change, 0 to destroy.

Changes to Outputs:
  + config_map = "tf-demo/shorty-settings"
  + namespace  = "tf-demo"
```

</div>

הסימן `+` הוא יצירה, `~` שינוי במקום, `-` מחיקה, ו־`-/+` מחיקה ויצירה מחדש. האחרון הוא זה שצריך לחפש בכל
סקירה: שינוי בשדה שה־API לא מרשה לשנות גורם להחלפה, ולמסד נתונים החלפה היא אובדן הנתונים. הערך
`(known after apply)` אומר שהשרת יקבע אותו, למשל `uid`.

### 22.3.2 apply, ושינוי אחד

<div dir="ltr" align="left">

```bash
terraform apply
terraform plan -var log_level=debug
```

</div>

<div dir="ltr" align="left">

```text
  # kubernetes_config_map_v1.settings will be updated in-place
  ~ resource "kubernetes_config_map_v1" "settings" {
      ~ data        = {
          ~ "SHORTY_LOG_LEVEL" = "info" -> "debug"
            # (1 unchanged element hidden)
        }
        id          = "tf-demo/shorty-settings"
...
Plan: 0 to add, 1 to change, 0 to destroy.
```

</div>

שינית ערך אחד, וה־plan מראה בדיוק שינוי אחד, במקום. זה מה שהופך את ה־plan לשימושי בסקירה: הוא לא מראה את
כל הקובץ, רק את ההבדל. ובלי שינוי בכלל, `terraform plan -detailed-exitcode` מסתיים בקוד 0, ובקוד 2 כשיש
הבדל. כך CI בודק שאין שינויים מחוץ לקוד.

> **שימי לב:** ה־plan שמודפס ב־`plan` הוא תצוגה. ה־`apply` מחשב plan חדש, וזה יכול להיות שונה אם מישהו שינה
> משהו בינתיים. בצוות שומרים את ה־plan לקובץ, `terraform plan -out=tfplan`, ומבצעים בדיוק אותו עם
> `terraform apply tfplan`.

### 22.3.3 סטייה

סטייה (drift) היא הבדל בין העולם לתצורה שנוצר מחוץ ל־Terraform. התרגיל יוצר אחת: הוא משנה את ה־ConfigMap
ב־`kubectl edit` ומוסיף מפתח. עכשיו ה־plan, בלי שום שינוי בקבצים:

<div dir="ltr" align="left">

```text
  # kubernetes_config_map_v1.settings will be updated in-place
  ~ resource "kubernetes_config_map_v1" "settings" {
      ~ data        = {
          - "EXTRA"            = "by-hand" -> null
          ~ "SHORTY_LOG_LEVEL" = "error" -> "info"
            # (1 unchanged element hidden)
        }
...
Plan: 0 to add, 1 to change, 0 to destroy.
```

</div>

Terraform מציע להחזיר את המצב לתצורה: למחוק את המפתח שנוסף ולהחזיר את הערך. זו ההגנה: שינוי ידני לא
שורד את ה־apply הבא. ואם השינוי הידני היה נכון, למשל תיקון בזמן תקרית, מעבירים אותו לקובץ לפני ה־apply,
אחרת הוא יימחק. אם רוצים רק לראות מה השתנה בלי להציע תיקון, יש מצב נפרד:

<div dir="ltr" align="left">

```bash
terraform plan -refresh-only
```

</div>

<div dir="ltr" align="left">

```text
Note: Objects have changed outside of Terraform

Terraform detected the following changes made outside of Terraform since the
last "terraform apply" which may have affected this plan:
```

</div>

### 22.3.4 destroy

<div dir="ltr" align="left">

```bash
terraform destroy
kubectl get namespace tf-demo
```

</div>

<div dir="ltr" align="left">

```text
Destroy complete! Resources: 2 destroyed.
Error from server (NotFound): namespaces "tf-demo" not found
```

</div>

הסדר היה הפוך ליצירה, לפי אותו גרף. וה־state נשאר, ריק: Terraform זוכר שהוא כבר לא אחראי לכלום.

## 22.4 משתנים, outputs ומודולים

### 22.4.1 משתנים

משתנה (variable) הוא קלט של התצורה. יש לו סוג, תיאור, ברירת מחדל אם רוצים, וכלל אימות. מהקובץ
`solution/part1/variables.tf`:

<div dir="ltr" align="left">

```hcl
variable "log_level" {
  description = "Value of SHORTY_LOG_LEVEL in the ConfigMap."
  type        = string
  default     = "info"

  validation {
    condition     = contains(["debug", "info", "warn", "error"], var.log_level)
    error_message = "The log_level must be one of debug, info, warn, error."
  }
}
```

</div>

כלל האימות הוא אותו רעיון כמו סכמת ה־values בסעיף 21.7: ערך שגוי נעצר ב־plan, לפני שמשהו משתנה. יש כמה דרכים
לתת ערך, מהחלשה לחזקה:

| מקור | דוגמה |
| --- | --- |
| ברירת המחדל בהגדרה | `default = "info"` |
| קובץ `terraform.tfvars`, שנטען אוטומטית | `log_level = "debug"` |
| משתנה סביבה | `TF_VAR_log_level=debug` |
| דגל בשורת הפקודה | `-var log_level=debug` |

### 22.4.2 outputs

output הוא פלט של התצורה: ערך ש־Terraform מדפיס אחרי apply, ושמודול אחר או סקריפט יכולים לקרוא. מהקובץ
`solution/part1/outputs.tf`:

<div dir="ltr" align="left">

```hcl
output "config_map" {
  description = "Namespace and name of the ConfigMap, as kubectl expects them."
  value       = "${kubernetes_config_map_v1.settings.metadata[0].namespace}/${kubernetes_config_map_v1.settings.metadata[0].name}"
}
```

</div>

<div dir="ltr" align="left">

```bash
terraform output -raw config_map
```

</div>

### 22.4.3 מודול מקומי

מודול (module) הוא תיקייה של קבצי Terraform. כל תצורה היא מודול, שנקרא מודול השורש; מודול שהשורש קורא לו
מקבל משתנים, יוצר resources ומחזיר outputs, כמו פונקציה. בחלק השני של התרגיל המודול `modules/shorty` יוצר
namespace ומתקין לתוכו את ה־chart. כך השורש קורא לו, בקובץ `solution/part2/main.tf`:

<div dir="ltr" align="left">

```hcl
module "shorty" {
  source = "./modules/shorty"

  namespace = var.namespace
  replicas  = var.replicas
  log_level = var.log_level
  image_tag = var.image_tag
}
```

</div>

המודול מצהיר אילו providers הוא צריך, אבל לא מגדיר אותם; את החיבור לאשכול קובע השורש. כך אותו מודול עובד מול
kind, מול אשכול בדיקות ומול ייצור, בלי שינוי. מהקובץ `modules/shorty/versions.tf`:

<div dir="ltr" align="left">

```hcl
# A module declares the providers it needs; the root module configures them.
terraform {
  required_providers {
    kubernetes = {
      source = "hashicorp/kubernetes"
    }
    helm = {
      source = "hashicorp/helm"
    }
  }
}
```

</div>

ב־state הכתובות של resources בתוך מודול מקבלות קידומת:

<div dir="ltr" align="left">

```text
module.shorty.helm_release.shorty
module.shorty.kubernetes_namespace_v1.this
```

</div>

> **איך Rust עושה את זה:** מודול של Terraform דומה ל־crate עם ממשק ציבורי צר: המשתנים הם הפרמטרים,
> ה־outputs הם ערכי ההחזרה, וה־resources בפנים הם פרטי מימוש. כמו ב־Rust, שינוי בשם של משתנה הוא שינוי
> שובר לכל מי שקורא למודול, ולכן מודולים משותפים מקבלים גרסאות.

מתי לא לכתוב מודול? כשיש קורא אחד. מודול עם קורא יחיד הוא שכבת עקיפה בלי תמורה, אותו כלל כמו בקוד. בתרגיל
המודול מוצדק כי הוא אורז החלטה אחת שחוזרת: „shorty הוא namespace ו־release יחד”.

## 22.5 state מרוחק ונעילות

ה־state בתרגיל הוא קובץ מקומי. הקובץ `solution/part2/backend.tf` כותב את זה במפורש, כדי שהמיקום יהיה
החלטה שרואים:

<div dir="ltr" align="left">

```hcl
terraform {
  backend "local" {
    path = "state/terraform.tfstate"
  }
}
```

</div>

ה־backend הוא המקום שבו ה־state נשמר. במחשב אחד, local מספיק. בצוות הוא נשבר בשלוש דרכים:

| בעיה | מה קורה | מה פותר |
| --- | --- | --- |
| כל אחד עם עותק משלו | שתיים מריצות apply על states שונים, ושתיהן חושבות שהן צודקות | state אחד, במקום משותף |
| שתי הרצות באותו רגע | שתיהן קוראות, שתיהן כותבות, והשנייה דורסת | נעילה (lock) לפני כל כתיבה |
| המחשב נגנב או נמחק | אין יותר קישור בין הקוד לתשתית | אחסון עמיד עם גרסאות |

backend מרוחק, כמו S3, פותר את שלושתן: הקובץ במקום משותף, לכל ריצה נלקחת נעילה, והאחסון שומר גרסאות קודמות.
נכון לאוקטובר 2026, בתיעוד של Terraform ה־backend של S3 תומך בנעילה בתוך S3 עצמו, עם הארגומנט
`use_lockfile`. הנעילה הישנה דרך DynamoDB מסומנת שם כמיושנת ותוסר בגרסה עתידית. דוגמה ל־backend כזה נמצאת
בסעיף 22.7, כי היא דורשת חשבון ענן.

נעילה היא לא תחליף לתהליך. גם עם נעילה, שני אנשים שמריצים apply מענפים שונים, זה אחר זה, דורסים זה את
השינויים של זה בתורות. לכן בצוותים apply רץ ממקום אחד, בדרך כלל CI אחרי מיזוג, ו־plan רץ על כל pull request
(פרק 23).

> **שימי לב:** הנעילה נלקחת בכל פקודה שכותבת, וגם ב־plan. ריצה שנקטעה באמצע משאירה נעילה תקועה, והפקודה
> `terraform force-unlock` משחררת אותה. מריצים אותה רק אחרי שווידאת שאף אחד אחר לא באמצע ריצה.

## 22.6 ה־provider של Helm

הסעיף מחבר את שני הפרקים: ה־chart מפרק 21 מותקן כ־resource של Terraform. קודם ההגדרה של ה־provider, בקובץ
`solution/part2/providers.tf`:

<div dir="ltr" align="left">

```hcl
provider "kubernetes" {
  config_path    = var.kubeconfig
  config_context = var.kube_context
}

# Helm provider v3: "kubernetes" is an attribute that takes an object (= { }),
# not a nested block as in provider v2.
provider "helm" {
  kubernetes = {
    config_path    = var.kubeconfig
    config_context = var.kube_context
  }
}
```

</div>

סימן השוויון לפני הסוגריים הוא כל ההבדל, והוא שובר את רוב הדוגמאות ברשת. מדריך השדרוג לגרסה 3 של
ה־provider, במאגר שלו, מסביר שהמעבר ל־Plugin Framework הפך בלוקים מקוננים לאובייקטים: `kubernetes` הוא עכשיו
אובייקט, `registry` הפך לרשימה בשם `registries`, ו־`set` ב־`helm_release` הוא רשימה של אובייקטים.

ה־resource עצמו, בקובץ `modules/shorty/main.tf`:

<div dir="ltr" align="left">

```hcl
resource "helm_release" "shorty" {
  name      = var.release_name
  namespace = kubernetes_namespace_v1.this.metadata[0].name
  chart     = "${path.module}/chart" # a local path, not a repository

  # Values as one YAML document, exactly like a -f file for helm install.
  values = [yamlencode({
    replicaCount = var.replicas
    image        = { tag = var.image_tag }
    config       = { logLevel = var.log_level }
  })]

  wait    = true # apply returns only when the pods are ready
  timeout = 120  # seconds
}
```

</div>

הביטוי `path.module` הוא התיקייה של המודול, כך שה־chart נמצא בלי קשר למקום שממנו מריצים. הפונקציה `yamlencode`
הופכת אובייקט של Terraform לקובץ values, והסכמה מסעיף 21.7 בודקת אותו כרגיל. אחרי apply אחד:

<div dir="ltr" align="left">

```text
Apply complete! Resources: 2 added, 0 changed, 0 destroyed.

Outputs:

namespace = "shorty-tf"
port_forward = "kubectl -n shorty-tf port-forward svc/shorty 8080:80"
release = {
  "chart" = "shorty-0.1.0"
  "name" = "shorty"
  "revision" = 1
}
```

</div>

ו־`helm list -n shorty-tf` מראה release רגיל. Terraform לא עוקף את Helm; הוא משתמש בו כספרייה.

### 22.6.1 שתי עובדות שכדאי לבדוק בעצמך

הראשונה: ה־provider מכיל ספרייה של Helm, לא את הקובץ שהתקנת. בקובץ `go.mod` של ה־provider בגרסה 3.3.0 התלות
היא `helm.sh/helm/v3`, כלומר Helm 3. ההבדל נראה באשכול: הפעם השדה `managedFields` מראה `Update` ולא `Apply`,
כלומר החלה בצד הלקוח, ולא server-side apply כמו ב־`helm install` של Helm 4 בסעיף 21.6.

<div dir="ltr" align="left">

```text
terraform-provider-helm_v3.3.0_x5 Update
kube-controller-manager Update
```

</div>

השנייה חשובה יותר. התרגיל משנה את מספר העותקים ביד, ואז מריץ plan:

<div dir="ltr" align="left">

```bash
kubectl -n shorty-tf scale deploy shorty --replicas=4
terraform plan
```

</div>

<div dir="ltr" align="left">

```text
No changes. Your infrastructure matches the configuration.
```

</div>

ה־resource הוא ה־release, ומבחינת Helm ה־release לא השתנה. מה שקורה בתוך האובייקטים שה־chart יצר בלתי נראה
ל־plan. זה ההבדל מה־ConfigMap של סעיף 22.3, שבו Terraform ניהל את האובייקט עצמו וראה כל סטייה.

### 22.6.2 מתי Terraform צריך לנהל releases של Helm

| מתאים | פחות מתאים |
| --- | --- |
| רכיבי תשתית שנפרסים יחד עם האשכול: metrics-server, בקר ingress, operator | שירותים שמשתנים כמה פעמים ביום |
| כשה־release תלוי בערכים שנוצרים באותה ריצה, כמו כתובת של מסד נתונים | כשצריך זיהוי סטייה בתוך האשכול |
| סביבה שלמה שקמה ונמחקת בפקודה אחת, כמו בתרגיל | כשצוות האפליקציה פורס בעצמו, בקצב שלו |

הכלל המעשי: Terraform מנהל את הבמה, וכלי פריסה של Kubernetes, Helm מ־CI או כלי GitOps, מנהלים את מה שרץ
עליה. שני כלים שמנהלים אותו אובייקט הם מתכון לסטייה שאף אחד מהם לא רואה.

## 22.7 אופציונלי: AWS free tier

> **שימי לב:** הסעיף הזה אופציונלי לגמרי, ואפשר לדלג עליו בלי לפספס שום דבר בהמשך הספר. **הכותב לא הריץ את
> התצורה הזאת מול AWS.** היא עברה `terraform fmt -check` ו־`terraform validate` בלבד. חשבון ענן יכול לעלות
> כסף, וזו הפעם היחידה בספר שהוא מופיע.

לפני כל דבר אחר, אם את בוחרת לנסות:

- **תקציב והתראה קודם.** במסוף של AWS, תחת Billing, צרי תקציב (budget) חודשי קטן עם התראה בדואר. עשי את זה
  לפני ה־apply הראשון, לא אחריו.
- **בדקי את תנאי ה־free tier בעמוד של AWS ביום שאת מתחילה.** התנאים משתנים; הספר לא מצטט אותם כעובדה.
- **משתמש עם הרשאות מינימליות**, לא משתמש ה־root של החשבון, ומפתחות שלא נכנסים ל־Git.
- **`terraform destroy` באותו יום.** משאב ששכחת הוא המקור הנפוץ לחשבונית מפתיעה.

התצורה בתיקייה `solution/optional-aws` יוצרת דלי (bucket) אחד ב־S3: פרטי, עם גרסאות, מוצפן, וחסום לגישה
ציבורית. זה בדיוק הדלי שצוות היה שומר בו state. ההגנה על הגישה הציבורית, מהקובץ `main.tf`:

<div dir="ltr" align="left">

```hcl
resource "aws_s3_bucket_public_access_block" "state" {
  bucket                  = aws_s3_bucket.state.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}
```

</div>

והקובץ `backend-s3.tf.example` מראה איך תצורה אחרת שומרת את ה־state שלה בדלי, עם נעילה:

<div dir="ltr" align="left">

```hcl
terraform {
  backend "s3" {
    bucket       = "senior-course-tfstate-change-me"
    key          = "ch22/terraform.tfstate"
    region       = "eu-central-1"
    use_lockfile = true # a lock object next to the state; DynamoDB locking is deprecated
    encrypt      = true
  }
}
```

</div>

הקובץ נקרא `.example` כדי ש־Terraform לא יטען אותו. כדי לבדוק את התצורה בלי חשבון בכלל:

<div dir="ltr" align="left">

```bash
cd ~/senior-course/ch22/optional-aws
terraform init -backend=false
terraform validate
```

</div>

שימי לב לבעיית הביצה והתרנגולת: הדלי של ה־state נוצר בעצמו בתצורה עם state מקומי. רוב הצוותים יוצרים אותו
פעם אחת כך, ומשם כל שאר התצורות משתמשות בו.

## סיכום

- תשתית כקוד: מצב רצוי בקבצים, שינוי נסקר כמו קוד, והרצה שנייה בלי שינוי לא עושה כלום.
- provider מדבר עם API; resource הוא אובייקט מוצהר; state הוא הקישור בין הכתובות בתצורה לאובייקטים
  אמיתיים.
- הפניות בין resources בונות גרף, והגרף קובע את הסדר ביצירה ובמחיקה.
- `plan` מראה בדיוק את ההבדל; `-/+` הוא החלפה, וזה מה שמחפשים בסקירה.
- שינוי ידני מתגלה כסטייה ב־plan הבא, וה־apply מחזיר אותו לתצורה.
- משתנים עם סוג ואימות, outputs, ומודול שמצהיר על providers אבל לא מגדיר אותם.
- state לא נכנס ל־Git: יש בו סודות, ואין בו נעילה. בצוות, backend מרוחק עם נעילה ו־apply ממקום אחד.
- ה־provider של Helm בגרסה 3: `kubernetes = { }`; הוא מכיל את Helm 3 כספרייה, ולא רואה סטייה בתוך release.

## מונחים לזכור

| מונח בעברית | באנגלית | במשפט אחד |
| --- | --- | --- |
| תשתית כקוד | infrastructure as code | מצב רצוי של תשתית בקבצים, שכלי אחד מממש |
| הצהרתי | declarative | מתאר מה צריך להתקיים, לא אילו פקודות להריץ |
| provider | provider | תוסף שמתרגם resources לקריאות של API אחד |
| resource | resource | אובייקט אחד שהתצורה מצהירה עליו |
| state | state | הקישור בין כתובות בתצורה לאובייקטים אמיתיים |
| backend | backend | המקום שבו ה־state נשמר ונעול |
| plan | plan | ההבדל בין התצורה למצב האמיתי, לפני שמשהו משתנה |
| apply | apply | ביצוע ה־plan |
| סטייה | drift | הבדל בין העולם לתצורה שנוצר מחוץ לכלי |
| גרף תלויות | dependency graph | הסדר ש־Terraform גוזר מההפניות בין resources |
| משתנה | variable | קלט של תצורה או מודול |
| output | output | פלט של תצורה או מודול |
| module | module | תיקייה של קבצי Terraform עם קלט ופלט |
| נעילה | lock | מניעה של שתי ריצות כותבות על אותו state |
| קובץ נעילת גרסאות | dependency lock file | הקובץ `.terraform.lock.hcl` עם הטביעות של ה־providers |

## טעויות נפוצות

- **state ב־Git.** סודות בטקסט גלוי ודריסות בין עותקים. משתמשים ב־backend מרוחק עם נעילה.
- **לכתוב שם של resource כמחרוזת במקום הפניה.** Terraform לא יודע על התלות, והסדר נעשה אקראי.
- **apply בלי לקרוא את ה־plan.** שורת `-/+` אחת על מסד נתונים היא מחיקה של הנתונים.
- **תיקון ידני בזמן תקרית שלא עובר לקובץ.** ה־apply הבא מבטל אותו בשקט.
- **בלוק `kubernetes { }` ב־provider של Helm בגרסה 3.** השגיאה היא „Unsupported block type”. הכתיב הוא
  `kubernetes = { }`.
- **להניח ש־Terraform רואה כל שינוי באשכול.** שינוי בתוך release של Helm לא מופיע ב־plan.

## שאלות ראיון

1. **מה ה־state של Terraform, ולמה לאבד אותו מסוכן?** זה הקישור בין כל כתובת בתצורה לאובייקט אמיתי, עם
   המאפיינים שלו מהקריאה האחרונה. בלעדיו Terraform לא יודע מה הוא יצר: הוא ינסה ליצור שוב ויתנגש, או שלא
   ימחק משאבים שעדיין עולים כסף. אפשר לשחזר בעזרת `terraform import` משאב אחרי משאב, אבל זה ידני וארוך. לכן
   ה־state נשמר באחסון עמיד עם גרסאות, וגישה אליו מוגבלת, כי יש בו גם סודות.

2. **למה צוותים שומרים state מרוחק עם נעילה?** שלוש סיבות: אמת אחת במקום עותק לכל אחד, נעילה שמונעת משתי
   ריצות לכתוב יחד, ואחסון שלא תלוי במחשב של מישהו. נעילה פותרת התנגשות בזמן, לא התנגשות בתוכן: שני ענפים
   שמוחלים זה אחר זה עדיין דורסים. לכן apply רץ ממקום אחד אחרי מיזוג, ו־plan רץ על כל pull request.

   **המשך:** מה קורה כשריצה נקטעת ומשאירה נעילה? הנעילה נשארת, וכל ריצה אחרת נכשלת. משחררים עם
   `terraform force-unlock` עם מזהה הנעילה, ורק אחרי שבדקת שאין ריצה אחרת חיה.

3. **איך תקראי plan לפני שאת מאשרת אותו?** קודם השורה התחתונה: כמה נוספים, משתנים ונמחקים, והאם זה תואם
   לשינוי שציפיתי לו. אחר כך כל `-/+` וכל `-`: למה הוחלף, ואם יש בו נתונים. אחר כך ערכים רגישים ושינויים
   בהרשאות. ולבסוף שינויים שלא ביקשתי, שמסמנים סטייה: מישהו שינה ביד, וה־apply יבטל את זה. plan ששמור לקובץ
   מבטיח שמה שאושר הוא מה שיבוצע.

4. **מתי Terraform צריך לנהל release של Helm, ומתי לא?** כן, לרכיבי תשתית שנפרסים עם האשכול ותלויים
   בערכים מאותה ריצה, כמו בקר ingress או operator. לא, לשירותים שמשתנים כמה פעמים ביום, שם הקצב של Terraform
   והסקירה שלו מאטים. ויש מגבלה טכנית: ה־plan רואה את ה־release, לא את האובייקטים שבתוכו, ולכן הוא לא יגלה
   שמישהו שינה Deployment ביד. החלוקה הנפוצה: Terraform לבמה, כלי פריסה של Kubernetes לאפליקציות.

5. **מתי כדאי לכתוב מודול?** כשיש החלטה שחוזרת יותר מפעם אחת, וכשאפשר לתת לה ממשק צר: מעט משתנים, outputs
   ברורים, ובלי הגדרת providers בפנים. לא, כשיש קורא אחד, או כשהמודול רק עוטף resource אחד ומעביר את כל
   המאפיינים שלו. מודול משותף הוא API: שינוי בשם של משתנה שובר את כל הקוראים, ולכן הוא מקבל גרסאות.

## פתק Obsidian

הפתק המוכן נמצא גם כקובץ: `obsidian/22-terraform.md`. זה תוכנו:

- **תשתית כקוד:** מצב רצוי בקבצים; נסקר כמו קוד; הרצה שנייה בלי שינוי לא עושה כלום.
- **provider:** תוסף ל־API אחד, נעול בגרסה מדויקת; קובץ הנעילה ב־Git.
- **resource:** סוג ושם מקומי; הפניות בונות גרף שקובע סדר.
- **state:** הקישור לאובייקטים אמיתיים; מכיל סודות; לא ב־Git.
- **plan:** `+` יצירה, `~` שינוי, `-` מחיקה, `-/+` החלפה; שומרים לקובץ ומבצעים אותו.
- **סטייה:** שינוי ידני מתגלה ב־plan; `-refresh-only` רק מראה.
- **מודול:** משתנים פנימה, outputs החוצה; מצהיר על providers ולא מגדיר אותם.
- **backend מרוחק:** state אחד, נעילה, גרסאות; ב־S3 `use_lockfile`, DynamoDB מיושן.
- **Helm provider v3:** `kubernetes = { }`; מכיל Helm 3; לא רואה סטייה בתוך release.
- **AWS:** אופציונלי; תקציב והתראה קודם; destroy באותו יום.
- **שאלה פתוחה:** ___

<div dir="ltr" align="left">

```text
workflow     terraform init | fmt -check -recursive | validate | plan | apply | destroy
ci checks    terraform init -backend=false && terraform validate ; terraform plan -detailed-exitcode (0 none, 2 changes)
saved plan   terraform plan -out=tfplan ; terraform apply tfplan
one change   terraform plan -var log_level=debug  ->  Plan: 0 to add, 1 to change, 0 to destroy.
drift        kubectl -n tf-demo edit configmap shorty-settings ; terraform plan ; terraform plan -refresh-only
state        terraform state list ; terraform output -raw config_map ; never commit *.tfstate
lock file    terraform providers lock -platform=linux_amd64 -platform=darwin_arm64 ...
helm v3      provider "helm" { kubernetes = { config_path = ..., config_context = "kind-ch22" } }
release      helm_release { chart = "${path.module}/chart"  values = [yamlencode({...})]  wait = true }
remote       backend "s3" { bucket key region use_lockfile = true encrypt = true } ; terraform force-unlock ID
mine         first apply ___ s ; resources in part2 state ___ ; drift found by plan: yes / no (helm scale: ___)
```

</div>

## תרגיל

**מטרה:** לנהל תשתית על kind מקבצים: namespace ו־ConfigMap עם ה־provider של Kubernetes, plan שמראה שינוי אחד,
סטייה שנוצרה ביד ו־destroy; ואז מודול מקומי שמתקין את ה־chart של shorty דרך ה־provider של Helm, עם משתנים
ו־outputs. הסעיף על AWS אופציונלי ולא הורץ על ידי הכותב.
**זמן:** שני ימי לימוד, ימים 64 ו־65, כ־50 ו־60 דקות בנייה בהתאמה.
**ידע מקדים:** הפרק; פרק 21 (ה־chart וה־image); פרק 20 (kind והקשר של kubectl); Terraform מפרק 2.

הצעדים, קבצי ההתחלה, הרמזים והפתרון נמצאים ב־
[`exercises/ch22-terraform/README.md`](../exercises/ch22-terraform/README.md).

**קריטריון סיום:**

- ה־plan השני, אחרי שינוי של ערך אחד, מראה בדיוק שינוי אחד, ו־destroy משאיר את ה־namespace מחוק.
- ה־plan מגלה את השינוי שעשית ב־`kubectl edit`, ובפתק כתוב מה ה־apply היה עושה איתו.
- apply אחד יוצר דרך המודול את ה־namespace ואת ה־release של shorty, ו־destroy מסיר את שניהם.
- הסקריפטים `check_static.sh` ו־`smoke.sh` עוברים, ואין קובץ state בתיקייה שב־Git.
- בפתק: למה state מרוחק עם נעילה, ומתי Terraform צריך לנהל release של Helm.

## העשרה אופציונלית

- לקרוא את מדריך השדרוג לגרסה 3 של ה־provider של Helm במאגר שלו (`docs/guides/v3-upgrade-guide.md`), ולתרגם
  דוגמה ישנה מהרשת לכתיב החדש.
- להוסיף `moved` block שמשנה את השם המקומי של resource בלי למחוק וליצור אותו מחדש, ולראות ב־plan שאין החלפה.
- לנסות `terraform import` על namespace שיצרת ב־kubectl, ולראות איך Terraform מתחיל לנהל אותו.
- להפעיל את ה־experiment `manifest` של ה־provider של Helm, ולבדוק אם עכשיו ה־plan רואה את השינוי ב־`kubectl scale`.

</div>
