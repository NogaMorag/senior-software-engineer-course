---
direction: rtl
tags: [senior-course, book]
chapter: 2
---
# סביבת העבודה: WSL2, Docker, Go ו־uv

- **מערכת:** Ubuntu בתוך WSL2.
- **קונטיינרים:** Docker Desktop, על אותו מנוע.
- **כלים:** Git עם מפתח SSH, ועורך שעובד בתוך המערכת.
- **שפות:** Go ו־uv, בתיקיית הבית.
- **איפה עובדים:** רק תחת `~`. הכונן של Windows הוא לייבוא חד־פעמי.
- **איפה הכלים:** בתיקיית הבית, בלי `sudo`. מה שמוקדם ב־`PATH` מנצח.
- **הורדה בטוחה:** טביעה מאותו פרסום, בדיקה, ורק אז התקנה. טביעה מוכיחה שלמות, לא מקור.
- **הרופא:** `doctor.sh` בודק הכול; קוד היציאה הוא מספר הכשלים. ביום 57 הוספתי את חלק ה.
- **ניקוי:** מסתכלים לפני שמוחקים. נפחים נמחקים רק במפורש.
- **שאלה פתוחה:** ___

<div dir="ltr" align="left">

```text
wsl       wsl --install | wsl --list --verbose | wsl --shutdown | wsl --export Ubuntu <file.tar>
path      export PATH="$HOME/.local/go/bin:$HOME/.local/bin:$PATH"   (~/.bashrc; ~/.zshrc on a Mac)
course    export COURSE="$HOME/senior-software-engineer-course"     my work: ~/senior-course/chNN
verify    sha256sum -c <file>.sha256sum | sha256sum -c --ignore-missing <SHA256SUMS>
go        ~/.local/go  GOBIN=~/.local/bin  go version  -> go1.27.1
go tools  go install <module>@<pin>: staticcheck goose buf protoc-gen-go protoc-gen-go-grpc govulncheck actionlint
lint      golangci-lint 2.14.0: release tarball + golangci-lint-2.14.0-checksums.txt -> ~/.local/bin
uv        curl -LsSf https://astral.sh/uv/0.12.23/install.sh | sh ; uv python install 3.14
doctor    bash ch02/doctor.sh [--part-e] ; echo $?   (exit = number of failures)
docker    docker version | docker run --rm hello-world | docker system df | docker system prune
kind img  kindest/node:v1.37.0@sha256:a1ed56cf...   (pinned by digest)
my pins   go ____  uv ____  python ____  docker ____  kind ____  kubectl ____
```

</div>

קישורים: [[01-senior-in-2026]] · [[03-go-first-steps]]
