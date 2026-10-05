---
direction: rtl
tags: [senior-course, book]
chapter: 36
---
# תכנון מערכות בראיון ומאגר השאלות

- **הסבב:** אחד או שניים בתהליך, לעיתים על תחום המוצר; אין תשובה נכונה, יש פשרות.
- **השעון:** 5 דרישות, 5 מספרים, 10 ‏API ונתונים, 10 תרשים, 12 צלילות, 3 סיכום.
- **הפתיחה:** לחזור על הבעיה, לשאול מי ומה, כמה, ואיזו איכות; לכתוב לא־מטרות; לומר את התוכנית.
- **הלוח:** דרישות בפינה, מספרים מתחת, ‏API בצד, תרשים במרכז, שאלות פתוחות למטה.
- **רמז:** מידע. מה הבנתי, מה משתנה, עדכון הלוח; מחלוקת עם נימוק ותנאי.
- **הסיום:** המסלול הראשי בחצי דקה, שני סיכונים, מה הייתי מודדת, מה עם עוד זמן.
- **המספר שמכריע:** מבצע בהתראות, ראש השעה במתזמן, תחרות בכרטיסים, קרדינליות במדדים.
- **המנגנונים:** outbox, תור על המסד עם חכירה, מפתח ייחודי, מחיצה אטומה, מגבלה לכל דיירת.
- **המאגר:** שאלות מגורלות עם `quiz`, תשובה בקול לפני Enter, ורשימת חזרה בסוף.
- **שאלה פתוחה:** ___

<div dir="ltr" align="left">

```text
clock      45 min: reqs 5 | numbers 5 | API+data 10 | diagram 10 | deep dives 12 | wrap-up 3
open       restate -> who/what -> how many -> latency/availability/consistency -> non-goals -> plan aloud
board      top-left reqs | numbers | API+tables | diagram center (write path, numbered) | risks bottom
hint       redirect | challenge | changed assumption -> restate, what changes, update board
cases      hookline | rate limiter | notifications | scheduler | tickets | metrics ingestion
numbers    1e7/day ~ 116/s ; 1M users x 5/day ~ 1,157/s ; 10M series / 15 s ~ 666,667 samples/s
quiz       uv run quiz check ../data/questions.jsonl
           uv run quiz ask ../data/questions.jsonl -n 10 -c distributed -c backend
mock 1     case ___  score ___/16  fixes ___ ___ ___
mock 2     case ___  score ___/16  fixes ___ ___
weak       chapters ___
```

</div>

קישורים: [[35-coding-drills]] · [[37-hookline-spec]]
