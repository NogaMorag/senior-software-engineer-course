# Postmortem: <short title of what users experienced>

> Blameless: this document describes what the system and the people in it did, and why each step made
> sense with the information available at the time. It names roles, not culprits. The goal is to change
> the system so the same mistake is harder to make, easier to see, or cheaper when it happens.

| Field | Value |
| --- | --- |
| Date | `<YYYY-MM-DD>` |
| Authors | `<names>` |
| Status | draft / reviewed / action items done |
| Severity | `<SEV1, SEV2 or SEV3>` (see the chapter's table) |
| Duration | `<from first impact to full recovery, minutes>` |
| Detection | `<alert name>`, `<minutes from impact to page>` |

## Summary
Three to five sentences a reader outside the team understands: what users saw, for how long, why, and
how it was fixed.

## Impact
- Users affected: `<who, how many, which operations>`.
- SLO: `<share of requests failed>`; error budget spent: `<percent of the 30-day budget>`.
- Data: `<lost, delayed, inconsistent, or none>`.

## Timeline (UTC)
One line per event, with the time and the source (alert, log, chat, deploy record). Include what people
thought at the time, not only what turned out to be true.

| Time | Event |
| --- | --- |
| `hh:mm` | ... |

## Root cause and contributing factors
The technical cause, then the conditions that let it reach users (no validation, no canary, an alert
that took too long). There is rarely one cause; list each factor on its own line.

## What went well
## What went badly
## Where we were lucky

## Action items
Each item: one specific change, an owner, a due date, and the kind (prevent, detect, mitigate). "Be
more careful" is not an action item; a check that makes the mistake impossible is.

| Item | Kind | Owner | Due | Ticket |
| --- | --- | --- | --- | --- |
| ... | prevent / detect / mitigate | ... | ... | ... |

## Lessons
What other teams should know.
