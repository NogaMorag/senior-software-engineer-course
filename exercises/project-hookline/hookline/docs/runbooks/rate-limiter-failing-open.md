# Runbook: RateLimiterFailingOpen

| Field | Value |
| --- | --- |
| Alert | `RateLimiterFailingOpen` (`severity=ticket`) |
| Service | `hookline-api` |
| SLO or signal | rate limiter errors (`hookline_ratelimit_errors_total`) for 10 minutes |
| Rules | `deploy/observability/rules/operations.yml` |
| Owner | you (Hookline, milestone M8) |
| Last reviewed | `<date>` |

## What it means
The API cannot ask Redis for rate-limit decisions and serves every ingest request without a limit. That is
the designed behaviour (failing closed would turn a cache outage into an ingest outage); the protection of
N5 is gone until Redis is back.

## Impact
None for tenants yet. A noisy tenant could now load the database without limit.

## First five minutes
1. the "Hookline API" dashboard (http://localhost:3000, folder Hookline): the limiter stat. 2. The API's warnings: `rate limiter unavailable; serving without a limit`.

## Diagnose
- Redis down or unreachable (`redis-cli ping` in its container).
- `REDIS_ADDR` wrong after a release.

## Mitigate
Restore Redis; nothing to replay. If a tenant floods meanwhile, the shedder (503) still protects the API.

## Escalate
Redis outage longer than an hour: ticket to its owner, and watch the database.

## After
Keep the timeline (alert, acknowledgement, mitigation, recovery) and write the postmortem from
`docs/postmortem-TEMPLATE.md` within a few days.
