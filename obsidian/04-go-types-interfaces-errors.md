---
direction: rtl
tags: [senior-course, book]
chapter: 4
---
# טיפוסים, ממשקים ושגיאות

- **receivers:** מצביע כשמשנים מצב, ואז כל המתודות מצביע; קבוצת המתודות קובעת את הממשקים.
- **ממשקים:** מרומזים, קטנים, מוגדרים אצל הצרכנית; שורת `var _` לבדיקת קומפילציה.
- **nil בתוך ממשק:** טיפוס עם ערך ריק אינו `nil`; החזירי `nil` מפורש.
- **שגיאות:** sentinel לזיהוי, טיפוס לנתונים; `%w` להקשר; `errors.Is` ו־`errors.AsType` לשליפה.
- **גבול אחד:** בשכבות הפנימיות מוסיפים הקשר ומחזירים; בגבול מחליטים, מדפיסים ובוחרים קוד יציאה.
- **panic:** לבאג בלבד; `recover` רק בגבול של goroutine או בקשה.
- **generics:** `any`, קבוצות טיפוסים עם `~`; מאז 1.27 מתודות גנריות, אבל לא בממשקים.
- **הטמעה:** קידום בלי ירושה; אין קריאה וירטואלית ואין יחס „הוא”.
- **שאלה פתוחה:** ___

<div dir="ltr" align="left">

```text
var _ Store = (*MemStore)(nil)               // compile-time interface check
var ErrX = errors.New("x")                    // sentinel
type XError struct{ ... }; func (e XError) Error() string
return fmt.Errorf("op %s: %w", id, err)       // wrap, keep cause
errors.Is(err, ErrX)
var t XError; errors.As(err, &t)
t, ok := errors.AsType[XError](err)           // Go 1.26+
defer func() { if r := recover(); r != nil { ... } }()
func F[T any](xs []T) ...    type Integer interface{ ~int | ~int64 }
func (p Page[T]) Map[U any](f func(T) U) Page[U]   // Go 1.27+
type FileStore struct { MemStore; path string }    // embedding
exit codes: 0 ok / 1 error / 2 usage / 3 insufficient / 4 not found
```

</div>

קישורים: [[03-go-first-steps]] · [[05-go-concurrency]]
