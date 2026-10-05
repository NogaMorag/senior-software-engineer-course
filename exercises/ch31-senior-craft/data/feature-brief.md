# Feature brief: bulk replay of dead deliveries (Hookline)

Written by the product manager. This is the input for the estimate, the status update and the mentoring plan.

## The problem
When a receiver is down for hours, its deliveries end up `dead` (spec F7). Today a tenant can replay them one at
a time with `POST /v1/deliveries/{id}/replay`. After an outage that is thousands of calls. Two tenants asked for
"replay everything that failed for this endpoint yesterday" in the last month.

## What the PM asked for
- An API call that replays every dead delivery of one endpoint inside a time range.
- The tenant sees how many were queued and can follow progress.
- "It must not take Hookline down when someone replays 200,000 deliveries at once."
- A date: "Can we have it by the end of the month?"

## What is already there (from chapters 37 and 38)
- `deliveries` rows with `status`, `next_attempt_at`, `dead` state; single replay endpoint (M5).
- Per-tenant rate limiting (M6) and load shedding on ingest (M6).
- Workers claim due deliveries with `FOR UPDATE SKIP LOCKED` (M5).
- Metrics, dashboards and burn-rate alerts (M8).

## What nobody has decided yet
- Is the replay synchronous (one request does it all) or a background job?
- Does a replay of 200,000 rows count against the tenant's rate limit, and how?
- Who may run it: any API key of the tenant, or only an admin key?
- What happens if the endpoint is still down when the replay starts?
