---
direction: rtl
tags: [senior-course, book]
chapter: 38
---
# אבני הדרך

- **‏M1:** ממשקי מאגר; מימוש בזיכרון; אימות לפי גיבוב; 404 לכל משאב של דיירת אחרת.
- **‏M2:** מיגרציות לכל המודל; ‏`ON CONFLICT DO NOTHING` ואז `SELECT`; אירוע ו־outbox בטרנזקציה אחת.
- **שאלה פתוחה:** ___

<div dir="ltr" align="left">

```text
M1 tests  DATABASE_URL= go test -race ./internal/api/ ./internal/store/... ./internal/admin/
M2 db     make db-up && export DATABASE_URL=postgres://hookline:change-me@localhost:5438/hookline
M2 tests  go test -race -count=1 ./...
admin     go run ./cmd/hookline-admin migrate up | create-tenant -name <name>
```

</div>

קישורים: [[37-hookline-spec]] · [[39-hookline-evaluation]]
