# Postmortem: creating short links failed after a topic rename

> The incident was staged on purpose (chapter 25, `data/stage_incident.sh`) on the chapter 24 stack.
> Numbers in `<angle brackets>` are placeholders: fill them from **your** run. The section "Author's run"
> at the end holds the only measured numbers in this file, from the book author's machine; they are
> there to show what a filled timeline looks like, not as a result to match.

| Field | Value |
| --- | --- |
| Date | `<YYYY-MM-DD>` |
| Authors | `<you>` |
| Status | reviewed |
| Severity | SEV2: one operation (creating links) failed for every user; resolving links kept working |
| Duration | `<minutes from the config change to the error ratio back at 0>` |
| Detection | `SLOAvailabilityBudgetBurn` (page), `<minutes from the change to firing>` after the change |

## Summary
A configuration change redeployed shorty with `SHORTY_KAFKA_TOPIC=link.created`; the real topic is
`links.created`. shorty never asks the broker to create topics, so every publish failed, and because
creating a link waits for its publish, every `POST /v1/links` answered 503. Resolving existing links was
not affected. The availability burn-rate page fired `<n>` minutes after the change; rolling the change back
restored service within a minute, and the alert resolved `<n>` minutes later.

## Impact
- Users affected: everyone who tried to create a link, `<n>` requests over `<n>` minutes. Redirects and
  stats kept working.
- SLO: `<share>` of all shorty requests failed during the incident (about one request in five, since
  creates were about 20% of the traffic).
- Error budget: `<n>` failed requests, or `<percent>` of the 30-day budget at the measured traffic
  (budget = 0.5% of the month's requests).
- Data: links created during the incident were stored, but their `links.created` events were not
  published, so linkworker never announced them (the dual write that chapter 13's outbox removes). The
  rollback restarted shorty, and the in-memory store lost every link created before it, so links that
  existed before the incident started answering 404.

## Timeline (UTC)
Source of each line: the output of `data/stage_incident.sh` (`incident-timeline.txt`), the Alerts page of
Prometheus, and the dashboard.

| Time | Event |
| --- | --- |
| `<hh:mm:ss>` | Steady traffic, 20 requests per second; 5m error ratio 0.000 |
| `<hh:mm:ss>` | CHANGE: shorty redeployed with `SHORTY_KAFKA_TOPIC=link.created` |
| `<hh:mm:ss>` | First `publish failed` log lines in Loki, with `exception_message="publish to link.created: context deadline exceeded"` |
| `<hh:mm:ss>` | `SLOAvailabilityBudgetBurn` pending, then firing; 5m error ratio `<value>` |
| `<hh:mm:ss>` | On-call opens the RED dashboard: errors only on `POST /v1/links`, status 503 |
| `<hh:mm:ss>` | A trace of a failed create shows the `link.created publish` span in error; the name differs from the usual `links.created publish` |
| `<hh:mm:ss>` | MITIGATION: shorty rolled back to the default topic |
| `<hh:mm:ss>` | 5m error ratio back near 0; the alert resolves |

## Root cause and contributing factors
- **Cause:** the topic name in the configuration did not exist. The client kept asking the broker for the
  topic's metadata until the 2-second publish timeout ran out, so the error in the logs was
  `context deadline exceeded`, not "unknown topic": the first clue pointed at a slow broker.
- **No startup check:** shorty validates its configuration at startup (chapter 6), but only the format of
  values, not that the topic exists. It started, passed its health check, and failed on the first create.
- **Health check does not cover the dependency:** `/healthz` answers `ok` without touching Kafka, so the
  deploy looked healthy.
- **Creating a link depends on a synchronous publish:** a broker-side problem turns directly into a user
  error.
- **No post-deploy smoke test:** nothing created a link right after the change.

## What went well
- The burn-rate alert fired within minutes and pointed at the right SLO.
- The dashboard isolated the failing route at once; the error log line named the topic.
- The rollback was one command and took effect in seconds.

## What went badly
- The deploy was reported healthy while every create failed.
- Links created during the incident were never announced, and nothing records which ones.

## Where we were lucky
- Only one setting changed, so the rollback target was obvious.
- Resolving links, most of the traffic, did not depend on Kafka.

## Action items

| Item | Kind | Owner | Due | Ticket |
| --- | --- | --- | --- | --- |
| At startup, shorty checks that `SHORTY_KAFKA_TOPIC` exists (a metadata request) and refuses to start if not | prevent | `<owner>` | `<date>` | `<id>` |
| Post-deploy smoke test: create one link and resolve it; roll back automatically on failure | detect, mitigate | `<owner>` | `<date>` | `<id>` |
| Decouple link creation from the publish with a transactional outbox (chapter 13) | mitigate | `<owner>` | `<date>` | `<id>` |
| Add the topic name to the deploy diff that reviewers see | detect | `<owner>` | `<date>` | `<id>` |
| Log the topic name in the publish error and in the startup line, and update `runbooks/slo-availability-budget-burn.md` with this signature | detect | `<owner>` | `<date>` | `<id>` |

## Lessons
A health check that does not exercise a dependency proves only that the process is running. Configuration
that names a resource (a topic, a bucket, a table) is only valid if the resource exists; check it at
startup, where a failure stops the deploy instead of the users.

## Author's run (measured on the author's machine, for illustration only)
Measured on 2026-10-05 with `data/stage_incident.sh` on the chapter 24 stack, on a Mac (Apple silicon) with
Docker Desktop shared with other workloads. Times are UTC, as the script prints them:

```text
11:13:03  traffic starts (20 requests per second, 40 minutes at most)
11:18:05  baseline done, 5m error ratio 0.000
11:18:05  CHANGE: shorty redeployed with SHORTY_KAFKA_TOPIC=link.created
11:18:12  SLOAvailabilityBudgetBurn: none (5m error ratio 0.000)
11:19:22  SLOAvailabilityBudgetBurn: firing (5m error ratio 0.048)
11:19:22  MITIGATION: shorty rolled back to the default topic
11:19:29  rollback done, 5m error ratio 0.048
11:22:21  SLOAvailabilityBudgetBurn: none (5m error ratio 0.052)
11:22:21  latency page: none
11:22:21  traffic stopped; incident staged. Copy these lines into the postmortem timeline.
```

What the author's run showed, beyond the timeline:
- Failed requests: about 310, all `POST /v1/links` answered 503, out of about 7,900 shorty requests in
  the 15 minutes around the incident (a Prometheus `increase(...[15m])` query right after the run).
- The error recorded in Loki (structured metadata of the OTLP log record):
  `exception_message="publish to link.created: context deadline exceeded"`.
- The page fired through the 6h/30m pair, 77 seconds after the change: on a stack with only minutes of
  history the long windows react almost as fast as the short ones. With a full 6 hours of history the
  same error ratio (about 5% over 5 minutes) would not have crossed the 1h threshold (7.2%) at all.
- The alert resolved about 3 minutes after the rollback, when the 30-minute ratio fell below 3%.
- The rollback restarted shorty, whose links live in memory, so links created before it answered 404
  afterwards: a side effect of the fix.
- The latency page did not fire: the slow creates (each waited 2 seconds) lasted about a minute.
