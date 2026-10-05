# Runbook: SLOAvailabilityBudgetDrain

| Field | Value |
| --- | --- |
| Alert | `SLOAvailabilityBudgetDrain` (`severity=ticket`) |
| Service | the `job` label of the alert (for chapter 25: `shorty`) |
| SLO | `job:slo_availability:objective` of the requests are not answered 5xx, over 30 days (shorty: 99.5%) |
| Owner | you (chapter 25) |
| Last reviewed | `<date>` |

## What it means
Over the last 3 days, and still over the last 6 hours, more requests failed than the SLO allows (burn
rate above 1). Nothing is on fire, but at this rate the 30-day budget runs out before the 30 days do.
This is a ticket: nobody is woken up; it is handled in working hours, this week.

## Impact
A small, steady share of users get errors. Each one is minor; together they spend the budget that
would otherwise pay for risky releases.

## First steps (working hours)
1. Open the "RED overview" dashboard with a 7-day time range; find when the error ratio rose above its
   usual level.
2. Break the errors down by route and status code (the query in `slo-availability-budget-burn.md`,
   with `[1d]` instead of `[5m]`).
3. Look for a release, a dependency change or a traffic change near the start.

## Mitigate
Usually a code or configuration fix through the normal release process, not an emergency change. If
the budget is nearly spent, agree with the team to pause risky releases until it recovers (this is the
point of an error budget policy).

## Escalate
If the cause is another team's dependency, open a ticket with them, link this alert and the numbers.

## After
Record the cause in the ticket. If the leak lasted weeks before this alert, ask whether the SLO or the
alert thresholds need review.
