# Hookline demo (10 minutes)

Milestone M10, chapter 39 section 39.5. A script you rehearse until it fits: what to show, the command, and
what to say. Every number you say comes from `docs/EVALUATION.md`, from your own runs, with the date.

## Before the audience arrives (not timed, about 5 minutes)

```bash
cd ~/senior-course/hookline
make images && make obs-up                      # the stack with Grafana, Prometheus, Loki and Tempo
make migrate
make tenant NAME=demo                            # prints an API key once; keep it in this terminal only
export KEY=hk_...                                # the key just printed
```

Open Grafana at `http://localhost:3000`, folder Hookline, with both dashboards in tabs.

Have three terminals ready: one for `make chaos`, one for single commands, one with `docs/EVALUATION.md`
and `docs/threat-model.md` open. Close everything else that uses the CPU: the demo runs a load test.

## The script

| Minute | Show | Command | Say |
| --- | --- | --- | --- |
| 0:00 | The README diagram | none | "A SaaS product emits events; Hookline delivers each one, signed, at least once, to every endpoint subscribed to it, and keeps one bad receiver from hurting the others." |
| 1:00 | One event's path | `scripts/check-telemetry.sh`, then the trace in Grafana (Explore, Tempo) | "The API returns 202 after one transaction writes the event and its outbox row. The trace context is stored with the event, so this one trace runs through the relay, Redpanda, the dispatcher and the worker to the receiver." |
| 2:30 | Start the chaos run | `RATE=50 DURATION=90 SEED=7 make chaos` | "Fifty events a second for ninety seconds. Every few seconds the script SIGKILLs a relay, a dispatcher or a worker and restarts it. The second receiver fails half its requests, then hangs, then fails a fifth." |
| 3:00 | The pipeline dashboard while it runs | none | "Delivery lag for the healthy endpoint stays low; the flaky one's breaker opens and its deliveries wait without spending attempts. Kills show up as short bumps, not as losses." |
| 4:30 | SSRF, while chaos runs | `curl -s localhost:18080/v1/endpoints -H "Authorization: Bearer $KEY" -H 'Content-Type: application/json' -d '{"url":"http://169.254.169.254/latest/meta-data"}'` | "Tenants choose URLs, so the worker would fetch our metadata service. Refused at create with a generic answer, and refused again at connect time, because DNS can change in between." |
| 5:30 | The threat model | `docs/threat-model.md`, table 4 | "Each threat points at the test that proves its mitigation. The highest remaining risk is written down: signing secrets are stored in plain text." |
| 6:30 | The proof | the end of the chaos output | "This is the claim of the whole project: the expected set comes from the database, the received set from the receiver. Zero missing. The duplicates are the price of at-least-once, and receivers dedupe by event id." |
| 7:30 | The numbers | `docs/EVALUATION.md` | "N1 at our load profile, N5 with two tenants, N6 above capacity: these are my runs, on this machine, on this date. Not a production benchmark." |
| 8:30 | Limits | none | "What this does not prove: a real network, a multi-node broker, months of running. What I would change next: rotate secrets without downtime and encrypt them at rest." |
| 9:30 | Questions | none | Chapter 39 section 39.7 has twenty you should be ready for. |

## If something goes wrong

- **Chaos ends with missing pairs.** Say so, and show the delivery status the script prints next to each
  missing pair. That is a finding, not a failed demo; then show the last passing run from `EVALUATION.md`.
- **The trace is not in Tempo yet.** Tempo needs a few seconds; show the Loki lines with the same trace id.
- **The laptop is slow.** Lower `RATE`; say what you lowered it to. The proof does not depend on the rate.

## After

```bash
make obs-down
```
