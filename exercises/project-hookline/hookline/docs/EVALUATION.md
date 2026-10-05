# Hookline evaluation

Your measurements, with the conditions you measured under. Every number in the "Your run" tables is one you
produced, on your machine, with the command next to it. Where you have no number, write why. Milestone M6
added the first table, M9 the load tests, and M10 (chapter 39) completes the file: one section per target of
spec N1 to N9.

Some sections end with an **Author's run**: what the run on the machine where the book was written produced,
labelled with its date and conditions. They are there so you can compare the shape of your results, not to be
copied. Your numbers will differ, and only yours belong in your copy of this file.

## Machine

| Field | Your run |
| --- | --- |
| Date | ___ |
| CPU, cores, RAM | ___ |
| OS, Docker version, memory given to Docker | ___ |
| Other load on the machine during the runs | ___ |
| Commit | ___ |

Author's run (2026-10-05): Apple M4 Pro, 12 cores, 24 GB RAM; macOS 26.5.2, Docker 29.1.3 with 12 CPUs and
7.65 GiB for its VM. The machine was shared: other workloads used about 4.8 GiB of that VM, and the load
average was between 5 and 12 during every run. Uncommitted milestone M10 tree on top of commit `691bf2c`.

## N1: ingest latency (`make load-ingest`)

Target: 99% of ingest requests return 2xx in under 200 ms in the project's load profile: `RATE` events per
second (default 100) from one tenant for 60 s, after a 10 s warm-up, with relay, dispatcher and workers
running. The numbers come from the `{phase:measure,expected_response:true}` line of k6's summary
(`/tmp/hookline-load/ingest-summary.txt`).

| Run | Command | Rate (/s) | Requests (all phases) | p50 (ms) | p95 (ms) | p99 (ms) | Max (ms) | Failed | Dropped | Thresholds |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| Profile | `make load-ingest` | ___ | ___ | ___ | ___ | ___ | ___ | ___ | ___ | ___ |
| Your choice | `RATE=___ make load-ingest` | ___ | ___ | ___ | ___ | ___ | ___ | ___ | ___ | ___ |

At what rate did p99 cross 200 ms, or did shedding (503) start first? ___

Author's run (2026-10-05 18:04, profile, `RATE=100`): 6252 requests; p50 7.16 ms, p95 15.93 ms, p99 26.65 ms,
max 116.99 ms; 0 failed, 0 dropped; all thresholds passed. An earlier run the same day (16:58), on a busier
machine and with fewer pre-allocated VUs, met p99 (99.97 ms) but failed `dropped_iterations` (3): see chapter 39,
section 39.1.

## N2: delivery lag (`make chaos`, and the dashboard)

Target: 99% of deliveries to healthy receivers arrive within 30 s of ingest. Two sources: the lag column of
`scripts/verify-delivery.sh` (ingest to the sink's first receipt, every pair of the run), and the delivery SLO
panel of the pipeline dashboard (`job:slo_late_deliveries:ratio_rate5m`, from the M8 histogram).

| Run | Source | Deliveries | p50 (s) | p99 (s) | Max (s) | Within 30 s (%) |
| --- | --- | --- | --- | --- | --- | --- |
| Quiet (`make load-ingest`, then `make verify-delivery`) | verify-delivery | ___ | ___ | ___ | ___ | ___ |
| Chaos, healthy endpoint | verify-delivery | ___ | ___ | ___ | ___ | ___ |
| Chaos, healthy endpoint | dashboard | n/a | n/a | n/a | n/a | ___ |

Author's run: see N3 and N4 below; the healthy endpoint's lag under chaos is the N2 number of that run.

## N3: isolation (`make lab-n3`, and `make chaos`)

Target: a receiver whose every request hangs does not push the healthy receivers' lag above N2.

| Run | Command | Events | Healthy p50 (s) | Healthy p99 (s) | Healthy max (s) | Slow endpoint: attempts spent |
| --- | --- | --- | --- | --- | --- | --- |
| Guards on | `make lab-n3` | ___ | ___ | ___ | ___ | ___ |
| Guards off | `GUARDS=off make lab-n3` | ___ | ___ | ___ | ___ | ___ |
| Chaos run | `make chaos` | ___ | ___ | ___ | ___ | n/a |

What the difference shows, in your words: ___

Author's run: `make lab-n3` was not run for this file. The chaos runs are below.

## N4: no delivery lost when a process dies (`make chaos`)

Target: killing a relay, a dispatcher or a worker at any moment loses no delivery. `scripts/chaos.sh` sends
`RATE` events per second for `DURATION` seconds, SIGKILLs one of two instances of each pipeline process every 3
to 12 s and restarts it, makes the second sink fail, hang, then fail again, and ends with
`scripts/verify-delivery.sh`. The result that matters is **0 missing**.

| Run | Seed | Rate, duration | Kills | Expected pairs | Missing | Duplicates | Healthy lag p50 / p99 / max (s) | Flaky lag p50 / p99 / max (s) |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| 1 | ___ | ___ | ___ | ___ | ___ | ___ | ___ | ___ |
| 2 | ___ | ___ | ___ | ___ | ___ | ___ | ___ | ___ |

Where did the lag come from (`scripts/stage-lag.sh`)? ___

Author's run (2026-10-05, `SEED=4242`, 50 events/s for 120 s, 11 kills, the same kill schedule both times):

| Run | Consumer session | Expected pairs | Missing | Duplicates | Healthy lag p50 / p99 / max (s) | Flaky lag p50 / p99 / max (s) |
| --- | --- | --- | --- | --- | --- | --- |
| 1 (18:00) | franz-go default, 45 s | 12244 | 0 | 0 | 71.61 / 118.54 / 121.85 | 74.31 / 118.53 / 123.26 |
| 2 (18:07) | 10 s (`bus.NewGroupConsumer`) | 12242 | 0 | 1 | 1.40 / 19.80 / 28.43 | 18.86 / 87.12 / 102.25 |

Run 1 kept every delivery and broke N2: `scripts/stage-lag.sh` put the whole delay in the dispatch stage (a
15-second bucket averaged 102.5 s from publish to delivery row), starting at the first dispatcher kill. A
SIGKILLed consumer keeps its partitions until its session times out, and a restarted member's rebalance waits
for it. Run 2, after the session was shortened to 10 s, kept the dispatch average below 9 s per bucket. The flaky
endpoint's lag is expected: it was failing or hanging for most of the run.

## N5: two tenants (`make load-two-tenants`)

Target: a tenant over its limit gets 429 with `Retry-After`; the other is not affected. The noisy tenant sends
`NOISY_RATE` per second (default 100) against a limit of a fifth of that; the quiet one sends `QUIET_RATE`
(default 20) inside its own limit, for 40 s.

| Tenant | Sent | 202 | 429 | 429 with valid Retry-After | p99 (ms) | Thresholds |
| --- | --- | --- | --- | --- | --- | --- |
| noisy | ___ | ___ | ___ | ___ | ___ | ___ |
| quiet | ___ | ___ | ___ (want 0) | n/a | ___ | ___ |

Does the noisy tenant's accepted rate match its limit? ___

Author's run (2026-10-05 18:10): 4802 requests; the noisy tenant got 3182 answers of 429, every one with a valid
`Retry-After`; the quiet tenant got 0, with p99 15.54 ms; all thresholds passed. An earlier run the same day, on
a busier machine, failed the quiet tenant's p99 threshold (277.51 ms) and dropped 18 iterations while the
functional checks passed: see chapter 39, section 39.1.

## N6: overload (`make load-overload`)

Target: above capacity, ingest sheds load with 503 instead of queueing without bound. The API runs with
`HOOKLINE_MAX_IN_FLIGHT=4` (`MAX_IN_FLIGHT`), k6 sends `RATE` (default 800) per second for 30 s, and a probe hits
`/healthz` five times a second.

| Run | Max in flight | Rate (/s) | Requests | 503 | Every 503 with Retry-After | Ingest p99 (ms) | Health always 200 | Thresholds |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| Default | 4 | ___ | ___ | ___ | ___ | ___ | ___ | ___ |

Author's run (2026-10-05 18:03): 24152 requests in total, 9385 ingest answers of 503, every one with a valid
`Retry-After`; ingest p99 17.13 ms, max 163.41 ms; every health probe answered 200; all thresholds passed.

## N7: keys hashed, secrets not logged, private addresses refused

Target: proven by tests, not measured. Record that they ran.

| Check | Command | Result |
| --- | --- | --- |
| Log scans | `go test -race -run 'LogsNever' ./internal/api/ ./internal/delivery/` | ___ |
| SSRF | `go test -race ./internal/ssrf/ && go test -race -run 'PrivateAddress' ./internal/api/ ./internal/delivery/` | ___ |
| Keys stored as hashes | `go test -race -run 'StoresOnlyTheHash' ./internal/admin/` | ___ |

## N8: every request and delivery traced; RED dashboards; a runbook per alert

| Check | Command | Result |
| --- | --- | --- |
| One trace from ingest to the sink | `make check-telemetry` | ___ |
| Rules and runbooks | `make promtool` | ___ |
| Dashboards looked at during `make chaos` | Grafana, folder Hookline | ___ |

## N9: everything runs on one laptop

Resource usage while `make chaos` runs (`scripts/chaos.sh` samples it halfway, to `$OUT/resources.txt`).

| Part | CPU (%) | Memory |
| --- | --- | --- |
| PostgreSQL | ___ | ___ |
| Redpanda | ___ | ___ |
| Redis | ___ | ___ |
| Observability (collector, Prometheus, Loki, Tempo, Grafana) | ___ | ___ |
| Two sinks | ___ | ___ |
| Seven Go processes | ___ | ___ |
| Total | ___ | ___ |

Author's run (2026-10-05 18:08, during chaos run 2): containers 1504 MiB in all, of which Redpanda 563 MiB,
PostgreSQL 97 MiB at 79% of one core, the observability stack 676 MiB, the two sinks 115 MiB; the seven Go
processes 231 MiB of RSS together, none above 16% of one core. One sample, at 50 events per second; not a
capacity measurement.
