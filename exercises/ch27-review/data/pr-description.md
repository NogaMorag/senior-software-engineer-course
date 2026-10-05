# Add read endpoints for documents

## What
- `GET /v1/documents?sort=<column>&limit=<n>` lists the caller's documents.
- `GET /v1/documents/{id}` returns one document.
- `internal/docs/repo.go`: PostgreSQL repository (`List`, `Get`).
- A small in-process cache for single documents, because the dashboard re-reads the same documents a lot.

## Why
The web dashboard needs to show documents. Writes come in a follow-up PR.

## How I tested
`go test ./...` passes. Tried both endpoints locally against my dev database.

## Rollout
Behind no flag; the endpoints are new, so nothing existing changes.
