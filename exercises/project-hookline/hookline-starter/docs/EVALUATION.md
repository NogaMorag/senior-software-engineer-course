# Hookline evaluation

Your measurements, with the conditions you measured under. Nothing here is filled in by the course: every
number is one you produced, on your machine, with the command next to it. Milestone M10 completes this
file; milestone M6 adds the first table, M9 the load tests.

## Machine

| Field | Value |
| --- | --- |
| Date | ___ |
| CPU, cores, RAM | ___ |
| OS, Docker version | ___ |
| Commit | ___ |

## N3: isolation (milestone M6, `scripts/lab-n3.sh`)

Target (spec N2 and N3): with one receiver timing out on every request, 99% of deliveries to a healthy
receiver arrive within 30 s of ingest.

| Run | Command | Events | Healthy p50 (s) | Healthy p99 (s) | Healthy max (s) | Slow endpoint: attempts spent |
| --- | --- | --- | --- | --- | --- | --- |
| Guards on | `make lab-n3` | ___ | ___ | ___ | ___ | ___ |
| Guards off | `GUARDS=off make lab-n3` | ___ | ___ | ___ | ___ | ___ |

What the difference shows, in your words: ___

## N1: ingest latency (milestone M9, `make load-ingest`)

Target (spec N1): 99% of ingest requests return 2xx in under 200 ms in the project's load profile:
`RATE` events per second (default 100) from one tenant for 60 s, after a 10 s warm-up, with relay,
dispatcher and workers running. The numbers come from k6's summary (`/tmp/hookline-load/ingest-summary.txt`),
the `{phase:measure,expected_response:true}` line.

| Run | Command | Rate (/s) | Requests | p50 (ms) | p95 (ms) | p99 (ms) | Max (ms) | Failed | Dropped iterations | Thresholds |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| Profile | `make load-ingest` | ___ | ___ | ___ | ___ | ___ | ___ | ___ | ___ | ___ |
| Your choice | `RATE=___ make load-ingest` | ___ | ___ | ___ | ___ | ___ | ___ | ___ | ___ | ___ |

At what rate did p99 cross 200 ms, or did shedding (503) start first? ___

## N5: two tenants (milestone M9, `make load-two-tenants`)

Target (spec N5): a tenant over its limit gets 429 with `Retry-After`; the other is not affected. The noisy
tenant sends `NOISY_RATE` per second (default 100) against a limit of a fifth of that; the quiet one sends
`QUIET_RATE` (default 20) inside its own limit, for 40 s.

| Tenant | Sent | 202 | 429 | 429 with valid Retry-After | p99 (ms) | Thresholds |
| --- | --- | --- | --- | --- | --- | --- |
| noisy | ___ | ___ | ___ | ___ | ___ | ___ |
| quiet | ___ | ___ | ___ (want 0) | n/a | ___ | ___ |

Does the noisy tenant's accepted rate match its limit? ___
