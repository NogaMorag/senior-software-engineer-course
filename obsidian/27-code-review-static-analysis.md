---
direction: rtl
tags: [senior-course, book]
chapter: 27
---
# סקירת קוד וניתוח סטטי

- **סדר סקירה:** תיאור קודם, ואז לפי סיכון: מה נשבר, רדיוס פגיעה, הפיכות.
- **מה חסר:** בדיקת הרשאה, סגירת משאב, בדיקה למקרה המסוכן.
- **שמונה תחומים:** גבולות, שגיאות, מקביליות, משאבים, אבטחה, בדיקות, שמות, הפעלה.
- **הערה:** תווית חומרה, נימוק, הצעה; שבח ספציפי; סיכום אחד.
- **לקבל סקירה:** לענות לכל הערה; לחלוק עם עובדה; שיחה אחרי שני סבבים.
- **שלוש שכבות ב־Go:** vet, staticcheck, golangci-lint; תצורה בגרסה 2.
- **govulncheck:** חולשה שנקראת מול חולשה שרק קיימת; trace מוכיח נגישות.
- **שאלה פתוחה:** ___

<div dir="ltr" align="left">

```text
labels: blocking: should-fix: nit: question: praise:
go vet ./...                 staticcheck ./...
golangci-lint run ./...      golangci-lint migrate      //nolint:<linter> // reason
.golangci.yml: version: "2" / linters.default: standard / linters.enable / formatters.enable
staticcheck prefixes: SA bugs, S simplify, ST style, U unused
uv run ruff check .          uv run mypy src
govulncheck ./...            govulncheck -show verbose ./...    -scan symbol|package|module    -mode binary
exit code 3 = vulnerabilities found
my review: found ___ of 8, missed ___
first file I opened: ___
```

</div>

קישורים: [[26-testing-strategy-load]] · [[28-security-backend]]
