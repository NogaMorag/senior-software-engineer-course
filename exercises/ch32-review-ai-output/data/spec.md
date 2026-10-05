# Task given to the AI assistant

This is the specification the module in `starter/ttlcache/` was generated from. Read it before the code: a
review checks the code against what was asked, not against what the code says about itself.

## Goal
An in-memory TTL cache for a single Go service, with a small HTTP API in front of it, in package `ttlcache`.
Standard library only.

## Behaviour
1. `New(ttl, now)` creates a cache. `now` is an injected clock (`nil` means `time.Now`).
2. `Set(key, value)` stores `value` until `now() + ttl`. An entry set at time `t` is visible for times in
   `[t, t+ttl)`: at exactly `t+ttl` it is expired.
3. `Get(key)` returns the value and `true`, or `nil, false` when the key is missing or expired. Every call
   counts as one hit or one miss.
4. `Delete(key)` removes a key and reports whether it was there.
5. `Sweep()` removes expired entries (same definition of "expired" as `Get`) and returns how many.
6. `StartJanitor(ctx, every)` runs `Sweep` every `every` in a background goroutine until `ctx` is cancelled, and
   returns a channel that is closed once the goroutine has returned.
7. `Stats()` returns hits, misses and the number of stored entries.
8. Every method is safe for concurrent use; the service calls `Get` from many goroutines at once.

## HTTP API (`NewHandler(cache, maxBody)`)
| Method and path | Result |
| --- | --- |
| `GET /cache/{key}` | 200 with the raw value, or 404 |
| `PUT /cache/{key}` | the body is the value; 204 when stored; **413 when the body is larger than `maxBody`, and nothing is stored**; 400 when the body cannot be read |
| `DELETE /cache/{key}` | 204, or 404 |
| `GET /stats` | 200 with `{"hits":..,"misses":..,"entries":..}` |

## Tests
Write table-driven tests with the injected clock; no sleeping where the clock can be moved instead.

## Assistant's summary when it finished (verbatim)
> Implemented `ttlcache` with an RWMutex for high read concurrency, a background janitor with clean context
> handling, and an HTTP handler with request-size protection via `http.MaxBytesReader`. Added 5 tests covering
> set/get, expiry, sweeping, stats and the HTTP round trip. All tests pass (`go test ./...`).
