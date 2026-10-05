---
direction: rtl
tags: [senior-course, book]
chapter: 22
---
# Terraform

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

קישורים: [[21-kubernetes-operations-helm]] · [[23-github-actions]]
