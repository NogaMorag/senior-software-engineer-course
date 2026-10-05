# Answer key: review of the AI-generated `ttlcache`

Line numbers refer to the starter (`starter/ttlcache/`). The review is written the way it would be posted on a
pull request: severity first, then what goes wrong, why it matters and a fix. The author happens to be an AI
assistant; the bar is the same as for any author.

## Where to look first (risk order)
1. `cache.go`, the locking: every request goes through `Get`, from many goroutines at once (spec item 8).
2. `cache.go`, the janitor: a background goroutine with a context is where leaks live.
3. `cache.go`, the time comparisons: the spec defines an exact edge (`[t, t+ttl)`), and there are two places
   that compare times.
4. `handler.go`, the trust boundary: what happens with input the client controls (body size, read errors).
5. `ai_test.go`: what do the tests actually prove? Each test checks the happy path, far from every edge above.

## Comments, by severity
| # | File:line | Severity | Defect | Comment as posted | Found by |
| --- | --- | --- | --- | --- | --- |
| 1 | `cache.go:55-62` | **blocking** | Data race: `Get` holds only the read lock, which many goroutines hold at once, and writes `c.hits` / `c.misses` under it. | "blocking: `RLock` allows many readers together, but `c.hits++` is a write, so concurrent `Get`s race on the counters (`go test -race` with parallel `Get`s reports it, and the counts come out low). Either make the counters `atomic.Uint64` and keep the read lock for the map, or take the write lock. I'd go atomic: it keeps the read concurrency the summary promises." | reading (a write under `RLock` is a pattern to look for), confirmed by `-race` |
| 2 | `cache.go:105-111` | **blocking** | Goroutine leak: `break` inside `select` leaves the `select`, not the `for`, so the janitor never returns, `done` is never closed, and the cache it references is never freed. | "blocking: this `break` only exits the `select`; the loop runs forever after cancel, so every cache leaks a goroutine and a ticker. `return` here (the deferred `close(done)` then runs). staticcheck flags it as SA4011, so it would be worth adding staticcheck to CI." | staticcheck (SA4011), or a test that waits on `done` |
| 3 | `cache.go:58` vs `cache.go:89` | **should-fix** | Off-by-one at the expiry edge: `Get` uses `now.After(expires)`, so an entry is still served at exactly `t+ttl`; `Sweep` uses `!now.Before(expires)` and removes it at that instant. The two disagree, and `Get` violates spec item 2. | "should-fix: at exactly `t+ttl` the spec says expired, `Sweep` agrees, `Get` doesn't (`After` is strict). One helper, `expired(e, now) = !now.Before(e.expires)`, used by both, and a table test at `ttl - 1ns`, `ttl`, `ttl + 1ns`. The current expiry test advances by `2*ttl`, which can't see this." | reading the two comparisons side by side; a boundary test |
| 4 | `handler.go:27-28` | **blocking** | Swallowed error: the `io.ReadAll` error is discarded, so a body over `maxBody` is silently truncated, stored, and answered with 204. The client believes the full value was saved. | "blocking: `MaxBytesReader` does its job, but its error is thrown away, so a 16-byte PUT with an 8-byte limit stores the first 8 bytes and returns 204. That is silent data corruption. Check the error: `errors.As(err, &*http.MaxBytesError)` gives 413, anything else 400, and in both cases don't call `Set`." | reading (`_, _ =` and `x, _ :=` on an error are always worth a look); a test with an oversized body |
| 5 | `ai_test.go` | **should-fix** | The tests prove only the happy path: no concurrency, no cancel, no boundary, no oversized body. The summary's "All tests pass" is true and says nothing about defects 1 to 4. | "should-fix: please add one test per spec rule that has an edge: parallel `Get`s under `-race`, janitor stops on cancel, expiry at exactly the TTL, PUT over the limit. Each of those fails today." | reading the tests against the spec |
| 6 | `cache.go:55` | **nit** | The comment justifies `RLock` with "better concurrency", which is right for the map and is exactly why the counter writes are wrong. | "nit: worth a line saying the counters are atomic because readers share the lock, so the next person doesn't 'fix' them back to plain fields." | reading |
| 7 | `handler.go`, `cache.go:39-44` | **praise** | Injected clock, small API, `PathValue` routing from the standard library, `Stats` as a snapshot. | "praise: the injected clock makes the boundary test trivial, and the handler stays thin. The fixes above are all local." | |

## Verdict
**Request changes.** Summary comment: "Good shape and a testable design (injected clock, thin handler). Three
blocking issues before merge: a data race on the hit/miss counters, a janitor that never stops (`break` inside
`select`), and an oversized PUT that is truncated and reported as success. Also the expiry edge, where `Get` and
`Sweep` disagree with each other and with the spec. Each needs a test that fails today; the current tests only
cover the happy path."

## How each defect was found, and what would have caught it
| Defect | Reading | `go vet` | `staticcheck` | `go test -race` | A test at the edge |
| --- | --- | --- | --- | --- | --- |
| 1 race | yes, if you look for writes under `RLock` | no | no | yes, with parallel `Get`s | yes (with `-race`) |
| 2 leak | yes, if you know the `break` rule | no | **yes, SA4011** | no | yes (wait on `done`) |
| 3 off-by-one | yes, by comparing the two time checks | no | no | no | yes (at exactly `ttl`) |
| 4 swallowed error | yes, `_` on an error | no | no (golangci-lint's `errcheck` only with `check-blank: true`, as one of ten reports) | no | yes (oversized body) |

## AI-specific notes
- The summary's three claims ("high read concurrency", "clean context handling", "request-size protection") each
  name the exact area that is broken. Treat a confident summary as a list of places to check, not as evidence.
- The code is idiomatic and compiles cleanly with `go vet`; the defects are in semantics, not syntax. That is the
  typical shape of an AI-written defect: plausible code that does something slightly different from the spec.
- Prompt fix for next time: put the edge cases into the spec as tests to pass ("a parallel `Get` test under
  `-race`; the janitor's `done` closes within a second of cancel; expiry at exactly the TTL; 413 over the
  limit"), and ask the assistant to run `go test -race ./...` and `staticcheck ./...` before it reports.
