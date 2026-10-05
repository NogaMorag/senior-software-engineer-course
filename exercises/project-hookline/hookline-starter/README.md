# Hookline (reference implementation)

A webhook delivery platform: tenants send each event once; Hookline delivers it, signed, at least once,
to every subscribed endpoint. The Hebrew project brief is `../README.md`; the specification and the
milestones are chapters 37 and 38 of the book.

<div dir="ltr" align="left">

```bash
make test-unit          # unit tests; the API tests run on the memory store
make run-dev            # API on :8080 with the memory store; prints a development API key
make db-up migrate      # PostgreSQL 18 in compose (project "hookline"), schema migrated
make tenant NAME=acme   # prints an API key once
make test               # everything, including the integration tests, against the compose database
make db-down            # stop and delete the database volume
```

</div>

State at this point: milestones M1 (API skeleton) and M2 (PostgreSQL and idempotent ingest).
