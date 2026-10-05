# Answer key: review of "Add read endpoints for documents"

Line numbers refer to the new files in `data/review.diff`.

## Where to look first (risk order)
1. `repo.go`: it talks to the database and builds SQL. Highest blast radius: every tenant's data.
2. `handler.go`: the trust boundary. Who may see what, what goes to logs and to clients, shared state.
3. `handler_test.go`: does anything prove the risky parts? (It does not.)
4. Naming and style last.

## Findings
| # | File:line | Severity | Issue | Suggested comment |
| --- | --- | --- | --- | --- |
| 1 | `repo.go:17-18`, `handler.go:45-48` | **blocking** | SQL injection: `sort` comes straight from the query string and is pasted into `ORDER BY` with `fmt.Sprintf`. `?sort=(SELECT ...)` runs attacker SQL; placeholders cannot bind identifiers. | "blocking: `sort` reaches SQL unescaped (`?sort=created_at;--` or a subquery). Placeholders can't bind a column name, so map an allow-list instead, e.g. `map[string]string{"created_at": "created_at DESC, id", "title": "title, id"}`, and return 400 for anything else. Could you add a test with a hostile `sort`?" |
| 2 | `repo.go:35-41`, `handler.go:17,74` | **blocking** | BOLA (broken object level authorization): `Get` looks up by `id` only; the handler ignores the caller's tenant. Any authenticated tenant can read any document by id. | "blocking: tenant B can read tenant A's document by guessing or seeing its id. Please make the tenant part of the query (`WHERE tenant_id = $1 AND id = $2`) and of the `Store.Get` signature, so a caller cannot forget it, and add a test where another tenant gets 404." |
| 3 | `handler.go:21,69,83` | **blocking** | Data race and a second BOLA path: a package-level `map` written from concurrent handlers without a lock (crashes with `concurrent map writes`), unbounded growth, never invalidated, and keyed by id only, so it serves one tenant's document to another. | "blocking: handlers run concurrently, so this map write races (`go test -race` with two parallel requests shows it), and the key has no tenant, so it bypasses any fix for #2. I'd drop the cache from this PR and measure first; if it is needed, a bounded cache keyed by (tenant, id) with a TTL, owned by the `Handler`, not a global." |
| 4 | `repo.go:19-31` | **blocking** | Resource leak: `rows` is never closed, so every list call holds a pooled connection until GC; under load the pool is exhausted and all requests hang. `rows.Err()` is not checked, so a broken read returns a truncated list as success. | "blocking: missing `defer rows.Close()` right after the error check; each call leaks a connection from the pool. Also check `rows.Err()` after the loop, otherwise a network error mid-read returns a partial list with 200." |
| 5 | `handler.go:55,75,80` | **should-fix** | Error handling: `err == sql.ErrNoRows` never matches because `Get` wraps the error with `%w`, so a missing document is a 500. And `err.Error()` is sent to the client, leaking SQL, table names and ids. | "should-fix: `Get` wraps with `%w`, so use `errors.Is(err, sql.ErrNoRows)` (better: have the repo return `ErrNotFound`, which already exists, so the handler does not know about `database/sql`). And please don't send `err.Error()` to clients: log it, return a generic message." |
| 6 | `handler.go:43` | **should-fix** | A secret in logs: the full `Authorization` header (a bearer token) is logged at info level for every list request. Logs are copied, retained and readable by more people than the API. | "should-fix: this logs the caller's bearer token. Anyone with log access can replay it until it expires. Log the subject from `auth.From` instead." |
| 7 | `handler.go:49-52` | **should-fix** | Input at the boundary: the `Atoi` error is ignored (`limit=abc` becomes the default silently), and there is no upper bound: `limit=1000000` reads a whole tenant into memory; a negative limit is sent to PostgreSQL. | "should-fix: please validate `limit`: 400 for non-numbers, and clamp to 1..100 (or 400 above 100). Unbounded limits are API4 in the OWASP API list: one request can make us read everything." |
| 8 | `handler_test.go:41-43`, `:14-16` | **should-fix** | Tests that do not test: `TestListDocuments` only `t.Log`s a wrong status, so it can never fail; it checks no body. The fake ignores `tenantID` and `sort`, so nothing would catch #1, #2 or #3. | "should-fix: `t.Log` here means the test passes whatever happens; please `t.Fatalf` and check the body. The risky behaviour has no test: another tenant's id must be 404, a bad `sort` 400, `limit` above the max 400. Making the fake honour `tenantID` would have caught #2." |
| 9 | `repo.go:36` | **question** | `context.Background()` replaces the request context, so a cancelled request (client gone, server shutting down) keeps the query running for up to 30 s. `staticcheck` reports it as SA4009. | "question: is dropping the caller's context deliberate? With `context.Background()` a cancelled request keeps the query alive for 30 s. If you want a cap, `context.WithTimeout(ctx, ...)` keeps both." |
| 10 | `handler.go:34,64,68` | **nit** | Naming: exported `ListDocumentsHandler`/`GetDocumentHandler` only need to be unexported methods `list`/`get`; Go spells `docID`, not `docId`. Unchecked `Encode` errors are fine for a response but worth a `_ =`. | "nit: these don't need to be exported, and `docId` should be `docID` (Go initialisms). Not blocking." |
| 11 | `handler.go:15-18` | **praise** | The handler depends on a small `Store` interface and reads the identity from `auth.From` rather than from a header; both make the fixes above easy. | "praise: nice that the handler depends on a two-method interface and takes identity from the verified context. That makes #2 a one-line signature change." |

## Verdict
**Request changes.** Summary comment: "Thanks, the shape is good (small interface, identity from context). Before merge: four blocking issues: SQL injection through `sort`, cross-tenant reads in `Get` and in the cache, a data race in the cache, and a leaked `rows` per list call. Then the error mapping, the token in logs, `limit` bounds and tests that actually fail. Happy to pair on the tenant-scoped query if useful."

## What tools would have caught
Run on the PR's code with golangci-lint v2.14.0 (default linters, then with `--enable sqlclosecheck,rowserrcheck,errorlint,gosec`):

| Finding | Caught by |
| --- | --- |
| 9 (`ctx` overwritten) | `staticcheck` SA4009, also inside golangci-lint's default set |
| 10 (unchecked `Encode`) | `errcheck` (default set), three reports |
| 1 (SQL built with `Sprintf`) | `gosec` G201 and G701, only when enabled |
| 5 (`==` on a wrapped error) | `errorlint`, only when enabled |
| 4 (`rows` not closed) | nothing in this run: `sqlclosecheck` and `rowserrcheck` stayed silent on this code |
| 3 (race) | `go test -race`, only with a test that sends parallel requests |
| 2, 6, 7, 8 | a person who thinks about what the code is for |
