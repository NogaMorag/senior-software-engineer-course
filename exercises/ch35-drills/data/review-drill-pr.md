# PR: Dedupe webhook deliveries in the receiver

Our sender delivers at least once, so we sometimes apply the same event twice. This PR
adds an in-memory `Dedupe` that remembers event ids for a window and drops repeats, and
makes the handler respond faster by processing in the background.

- New: `internal/inbox/dedupe.go`, `internal/inbox/dedupe_test.go`
- Changed: `internal/inbox/handler.go`
- Tests: `go test ./...` passes.

The service runs as one replica today.
