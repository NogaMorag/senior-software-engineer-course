# Hookline evaluation

Your measurements, with the conditions you measured under. Nothing here is filled in by the course: every
number is one you produced, on your machine, with the command next to it. Milestone M10 completes this
file; milestone M6 adds the first table.

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
