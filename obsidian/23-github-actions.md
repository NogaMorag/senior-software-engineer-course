---
direction: rtl
tags: [senior-course, book]
chapter: 23
---
# GitHub Actions

- **מבנה:** כל job רץ על מכונה נקייה; צעד הוא פקודה או action.
- **בין jobs:** רק דרך `needs` ו־`outputs`, מטמון, או registry.
- **סינון לפי נתיב:** התיקייה, הקובץ של ה־workflow והקוד המשותף; לא לבדיקה חובה.
- **‏shell:** ‏`shell: bash` מוסיף `pipefail`; בפעולה מורכבת חובה בכל צעד.
- **‏service container:** תמיד עם בדיקת בריאות; הכתובת `localhost` כשה־job רץ על המכונה עצמה.
- **ירוק שקרי:** דילוג על בדיקות מסד ב־CI הוא כישלון.
- **הרשאות:** ‏`contents: read` למעלה; ‏`packages: write` ו־`contents: write` רק ב־job שצריך.
- **סודות ומשתנים:** סוד מוסתר בלוג, משתנה לא; סודות של environment רק אחרי אישור.
- **‏OIDC:** אסימון קצר מועד במקום מפתח קבוע; תנאי על המאגר וה־environment.
- **הצמדה:** גרסה ראשית לבדיקות, ‏SHA עם הערה ל־workflow שכותב.
- **‏actionlint:** לפני כל push; לא מכיר כל גרסה של כל action.
- **שאלה פתוחה:** ___

<div dir="ltr" align="left">

```text
on: push: branches: [main] paths: ["ch06/**", ".github/workflows/ch06-shorty.yml"]   pull_request: paths: [...]
permissions: contents: read      job: permissions: packages: write | contents: write | id-token: write
concurrency: group: ${{ github.workflow }}-${{ github.ref }}  cancel-in-progress: ${{ github.event_name == 'pull_request' }}
defaults: run: shell: bash   (bash --noprofile --norc -eo pipefail)
actions/setup-go@v7 go-version-file cache-dependency-path   go run honnef.co/go/tools/cmd/staticcheck@2026.2.1 ./...
astral-sh/setup-uv@v10   uv lock --check   uv sync --frozen   uv run --no-sync pytest -q
services: postgres: image options: --health-cmd "pg_isready ..."   ports: 5432:5432   DATABASE_URL=...@localhost:5432/...
docker/metadata-action@v6  docker/login-action@v4 (GITHUB_TOKEN)  docker/build-push-action@v7 push load cache-from type=gha
helm/kind-action@v1.15.0   kind load docker-image shorty:ci --name ch23   kubectl rollout status   if: failure()
environment: name: production   secrets.X   vars.X   gh release create "$TAG" --verify-tag --generate-notes
git ls-remote https://github.com/<owner>/<action> 'refs/tags/vX.Y.Z*'   uses: owner/action@<40-char-sha> # vX.Y.Z
actionlint   tests/lint.sh   tests/run_ci_locally.sh --kind
my first run on GitHub: date ___  total minutes ___  slowest job ___
```

</div>

קישורים: [[22-terraform]] · [[24-telemetry]]
