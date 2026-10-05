# Answer key: review of "Dedupe webhook deliveries in the receiver"

Line numbers refer to the new files in `data/review-drill.diff`
(`dedupe.go`, `dedupe_test.go`, and `handler.go` after the change).

## Where to look first (risk order)
1. `handler.go:31-40`: the PR changes *when* the sender is told "done". That decides whether
   events can be lost, which is worse than processing one twice.
2. `dedupe.go:22-34`: the new shared state. Every concurrent request goes through it.
3. `dedupe_test.go`: does anything prove the two points above? (It does not.)
4. Logging, limits, naming last.

## Findings
| # | File:line | Severity | Issue | Suggested comment |
| --- | --- | --- | --- | --- |
| 1 | `handler.go:35-40` | **blocking** | Events can be lost. The handler answers 202 before processing, so the sender stops retrying; if `Process` fails, the failure is only logged. Worse, the goroutine uses `r.Context()`, which is cancelled as soon as `ServeHTTP` returns, so `Process` will often fail with `context canceled`. The sender's "at least once" becomes "at most once". | "blocking: once we answer 202 the sender considers the event delivered, and the background `Process` runs on a context that is cancelled when this handler returns. A failure (or that cancel) loses the event for good. Could we process synchronously and answer 2xx only after success (the old behaviour), or persist the event first and process from our own durable queue?" |
| 2 | `handler.go:31`, `dedupe.go:30-32` | **blocking** | The id is recorded as seen *before* processing and never removed on failure, so the sender's retry of a failed event gets "duplicate" and is dropped. Dedupe must record success, not arrival. | "blocking: `Seen` marks the id on the first arrival. If processing then fails, every retry is answered as a duplicate and the event is never applied. Record the id only after `Process` succeeds, and while it is in flight make duplicates wait for the outcome (or get a 5xx)." |
| 3 | `dedupe.go:23-33` | **blocking** | Check-then-act race: the lock is released between the read (line 24) and the write (line 31), so two concurrent deliveries of one id can both see "not seen" and both process it, which is exactly what the PR is meant to prevent. And `d.Hits++` (line 27) runs outside the lock: a data race that `go test -race` reports as soon as a test calls `Seen` from two goroutines. | "blocking: the read and the write are separate critical sections, so two deliveries arriving together can both pass. Hold the lock across the check and the insert. `Hits++` is also outside the lock; a test with 50 goroutines calling `Seen` makes `-race` report it." |
| 4 | `dedupe.go:12`, `:31` | **should-fix** | Unbounded memory: ids are added forever and never pruned; the window is only consulted when the same id comes back. | "should-fix: nothing ever deletes from `seen`, so memory grows with every event we ever receive. Entries arrive in time order, so a FIFO of ids lets you drop the expired prefix cheaply on each call." |
| 5 | `handler.go:26` | **should-fix** | Regression: the PR removed `http.MaxBytesReader`, so a client can make us read an arbitrarily large body into memory. | "should-fix: the size limit was removed in this change (old line 24). Was that intended? Without it one request can exhaust memory; please keep the 1 MiB cap and answer 413." |
| 6 | `handler.go:32` | **should-fix** | Wrong status for a duplicate: 409 tells an at-least-once sender that the delivery failed. Many senders (Hookline among them) treat a 4xx other than 408 and 429 as permanent, so the event shows up as failed or dead on their side and in their alerts. A duplicate was handled; answer 2xx. | "should-fix: from the sender's side a duplicate is a success: we already have the event. 409 makes it record a failed delivery. Could we answer 200 (with a body that says duplicate, if you want it visible)?" |
| 7 | `handler.go:24`, `:31` | **should-fix** | A missing `Hookline-Event-Id` is the empty string, which is deduped like any other id: after the first such request, every request without the header is silently dropped as a duplicate. | "should-fix: no header gives `id == \"\"`, and `\"\"` becomes one remembered id, so every later delivery without the header is dropped. Please reject it with 400." |
| 8 | `handler.go:25` | **question** | Logs the signature header of every request at info. It is not the signing secret, but together with a logged body it lets anyone with log access replay a request within the tolerance window. Also: nothing in this handler verifies the signature. | "question: do we need the signature in the logs? On its own it is harmless, but with the body it allows a replay inside the tolerance window. Where is the signature verified: is there middleware I'm not seeing?" |
| 9 | `dedupe_test.go:62-65`, `dedupe.go:26,31` | **should-fix** | The test cannot fail where it matters: `t.Log` instead of `t.Fatal`, a real `time.Sleep`, no concurrent case, no failure case, no window expiry. The code reads the wall clock (`time.Since`, `time.Now`), so expiry cannot be tested without sleeping for the whole window. | "should-fix: line 64 logs instead of failing, so this test passes whatever happens. Could you inject a clock (`now func() time.Time`) and add: N goroutines with the same id process once; a failed processing is retried; an id is forgotten after the window? The concurrent one would have caught #3." |
| 10 | `dedupe.go:14`, `:17`, `handler.go:20` | **nit** | `Hits` is an exported mutable field read without the lock; `NewDedupe` has no doc comment; a `Handler` built without `Dedupe` panics on a nil pointer at the first request. | "nit: I'd unexport `Hits` behind a method (or a metric), add a doc comment to `NewDedupe`, and make the zero `Handler` safe or provide a constructor." |
| 11 | `dedupe.go:8-21` | **praise** | Dedupe is its own small type with its own test file, separate from HTTP: that is what makes #3 and #9 easy to fix and test. | "praise: nice separation; keeping dedupe out of the handler makes it straightforward to test on its own." |

## Verdict
**Request changes.** Summary comment: "Thanks, the problem is real and isolating dedupe in its
own type is the right shape. Before merge, three blocking issues: answering 202 before processing
(with a cancelled context) can lose events; marking an id seen before processing drops the retries
of failed events; and the check and insert are not atomic, so concurrent duplicates still both
process (plus a data race on `Hits`). Then: memory bound, the body limit that was removed, 200 for
duplicates, 400 without an id, and tests that fail when the behaviour is wrong. One more thought for
later: with a second replica an in-memory dedupe stops working; a unique key in the database we
write to would make dedupe and the write one transaction."

## What tools would have caught
Run on the PR's code (Go 1.27.1, staticcheck 2026.2.1):

| Finding | Caught by |
| --- | --- |
| 3 (race on `Hits`) | `go test -race`, but only with a test that calls `Seen` from several goroutines; the PR's own test does not |
| 3 (double processing) | the same concurrent test, sometimes: the window between the two critical sections is narrow |
| all others | nothing automatic: `go vet ./...` and `staticcheck ./...` are silent on this diff; a reviewer who asks "when does the sender think we are done?" finds 1, 2 and 6 |

## Scoring yourself
- Ranking matters more than count: #1, #2 and #3 first. A review that lists ten nits and misses #1
  would approve a change that loses data.
- Every blocking comment should say why it matters and offer a direction, not just "wrong".
