---
direction: rtl
tags: [senior-course, book]
chapter: 32
---
# הנדסה בעזרת AI

- **שלוש קטגוריות:** השלמה, שיחה בעורך, סוכן. הסוכן עורך קבצים ומריץ פקודות, ואני סוקרת.
- **המודעות מבקשות:** שיקול דעת מול הפלט, לא מהירות.
- **האצלה:** לפי סיכון ולא לפי קושי; מה שאני לא יכולה לסקור, אני לא מאצילה.
- **מפרט:** קצוות ממוספרים, וגמור כבדיקות ופקודות.
- **סקירה:** הכלים קודם, ה־diff של הבדיקות קודם, אחר כך לפי סיכון.
- **חמישה כשלים:** API שגוי, דגל או גרסה מומצאים, שינוי שקט, שינוי רחב, בדיקות שנחלשו.
- **הסיכום של הכלי:** רשימת מקומות לבדוק, לא ראיה.
- **כשהוא טועה:** אחרי שני ניסיונות, מקטינים, מוסיפים הקשר, או כותבת בעצמי.
- **ראיונות:** שואלת מה הכללים; בלי AI זה בלי AI; עם AI, שיקול דעת בקול.
- **שאלה פתוחה:** ___

<div dir="ltr" align="left">

```text
tools        completion | editor chat | agent (edits files, runs commands, opens PRs)
instructions CLAUDE.md | .cursor/rules/*.mdc | .github/copilot-instructions.md | AGENTS.md (all three)
delegate     scaffolding, tests from a spec, mechanical migrations, explaining code
not alone    auth/authz/crypto/untrusted input, concurrency without tests, prod data migrations, design
spec         goal + non-goals | numbered rules with edges | interface | constraints | done = commands
done         go test -race -count=1 ./...   staticcheck ./...
review       tools first -> test diff first -> risk order -> each claim in the summary
AI failures  wrong API (go doc) | invented flag/version (--help, go list -m -versions) | silent change
             | over-broad diff (git diff --stat) | deleted/weakened tests
ch32 defects RLock + counter write | break in select (SA4011) | After vs !Before | body, _ := io.ReadAll
interviews   Meta AI round pilot (Oct 2025) | Canva expects AI (Jun 2025) | no Israeli AI round found
my story     AI was wrong when ___, I found it by ___
```

</div>

קישורים: [[31-working-as-senior]] · [[33-rust-edge]]
