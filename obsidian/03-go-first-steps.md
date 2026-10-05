---
direction: rtl
tags: [senior-course, book]
chapter: 3
---
# צעדים ראשונים ב־Go

- **הוויתור של Go:** אוסף זבל במקום בעלות, ערך אפס במקום אופציה, שגיאה כערך.
- **מודול וחבילה:** מודול עם גרסאות מינימליות; חבילה היא תיקייה; אות גדולה מייצאת.
- **ערכים ומצביעים:** הכול מועתק; ערך שחי אחרי הפונקציה בורח לערימה.
- **slice:** מצביע, אורך וקיבולת. הוספה כותבת במקום כשיש קיבולת. עותק או ביטוי מלא נגד כינוי.
- **map:** סדר מעבר אקראי ולכן מיון; כתיבה ל־map שערכו nil היא panic; כתיבה מקבילה מפילה את התהליך.
- **מחרוזות:** בתים; לולאת `range` נותנת runes והיסט בבתים; עברית היא שני בתים לאות.
- **בדיקות:** טבלה, תת־בדיקה עם שם, הודעת got ו־want; דוגמה עם הערת פלט נבדקת.
- **שאלה פתוחה:** ___

<div dir="ltr" align="left">

```text
go mod init example.com/NAME && go mod edit -go=1.27
go run .            go build ./...        go test ./...
go test -run 'TestX/case_name' -v .
gofmt -l .          go vet ./...          staticcheck ./...
go build -o /dev/null -gcflags=-m .       # escape analysis
s[low:high:max]     slices.Clone(s)       make([]T, 0, n)
v, ok := m[k]       slices.SortFunc(rows, func(a, b T) int { return cmp.Compare(a.X, b.X) })
len("שלום") = ___   utf8.RuneCountInString("שלום") = ___
sc.Buffer(make([]byte, 0, 64*1024), 1024*1024);  check sc.Err() after the loop
```

</div>

קישורים: [[02-workbench-wsl2]] · [[04-go-types-interfaces-errors]]
