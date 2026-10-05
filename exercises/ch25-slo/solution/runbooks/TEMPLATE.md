# Runbook: <AlertName>

> One runbook per alert. The alert's `runbook_url` annotation points here, so the person who was paged
> lands on this page with the alert in front of them. Write it for someone tired, at 03:00, who did not
> build the service. Keep commands copy-pasteable; keep explanations to one line each.

| Field | Value |
| --- | --- |
| Alert | `<AlertName>` (`severity=<page or ticket>`) |
| Service | `<job>` |
| SLO | `<what is promised, over how long>` |
| Owner | `<team or person>` |
| Last reviewed | `<date>` |

## What it means
One paragraph: what the user is experiencing, and how much error budget is going. Name the expression
that fired and its windows.

## Impact
Who is affected and how badly. What happens if nobody acts for an hour (for a page) or a week (for a
ticket).

## First five minutes
Numbered steps, the cheapest and most informative first. Each step says what to look at and what a
normal value looks like.

1. Acknowledge the page; say in the incident channel that you are looking.
2. Open the dashboard: `<link>`. Compare with `<normal value>`.
3. ...

## Diagnose
The questions to answer, each with the query or command that answers it:
- Did something change? (deploys, config, feature flags, traffic)
- Is it one route, one instance, one dependency?
- What do the traces of failing requests say?

## Mitigate
Stop the bleeding before finding the root cause. The safest first: roll back, turn off a flag, shed
load. Each with the exact command and how to tell it worked.

## Escalate
When and to whom: no improvement after N minutes, data loss suspected, more than one service.

## After
Open the incident record, keep the timeline, schedule the postmortem (see `postmortems/TEMPLATE.md`).
