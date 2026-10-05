# Estimate: bulk replay of dead deliveries (worked example)

Date: 2026-10-05  Author: the book  Input: `data/feature-brief.md`

The numbers below are the author's judgment for an engineer who built Hookline through M9. They are an example of
the shape, not a measurement; yours come from your own breakdown.

## 1. What I am estimating
A tenant can replay all dead deliveries of one endpoint in a time range with one API call, follow its progress,
and it cannot overload workers or receivers. "Done" means: merged, deployed to kind behind a feature flag, a
dashboard panel and an alert for stuck replays, a runbook entry, and the API documented in `openapi.yaml`.

Out of scope for this estimate: a UI; replay across several endpoints in one call; changing retry policy.

## 2. Breakdown
| # | Task | Low | High | Confidence | Depends on | Unknown it hides |
| --- | --- | --- | --- | --- | --- | --- |
| 1 | Decide sync vs background job, and write a one-page design with the PM's answers to the four open questions | 0.5 | 1 | medium | PM answers | Who may run it (U2) |
| 2 | Migration: a `replays` table (id, tenant, endpoint, range, status, counts), expand only | 0.5 | 1 | high | 1 | |
| 3 | `POST /v1/endpoints/{id}/replays` and `GET .../replays/{replay_id}`, with tenant scoping and idempotency key | 1 | 2 | high | 2 | |
| 4 | Background job that re-queues dead rows in batches of N with a per-tenant pace | 1.5 | 4 | low | 2 | Batch size and pace that do not hurt other tenants (U1) |
| 5 | Load test: replay 200,000 rows while normal traffic runs; check N2 and N3 still hold | 1 | 3 | low | 4 | Same as U1; may send us back to task 4 |
| 6 | Metrics, dashboard panel, alert on a replay with no progress for 10 minutes, runbook entry | 0.5 | 1 | high | 4 | |
| 7 | Feature flag, rollout to one tenant, docs | 0.5 | 1 | high | 3, 6 | |
|   | **Total** | **5.5** | **13** | | | |

## 3. From working days to a date
- Focus factor: about 60% (a guess; my last two weeks were not tracked).
- Calendar range: 5.5 / 0.6 ≈ 9 working days to 13 / 0.6 ≈ 22 working days, so **2 to 4.5 weeks**.
- Not included: the PM's time to answer the open questions; review time from the team.

## 4. Assumptions
| Assumption | If it is wrong |
| --- | --- |
| A replay is a background job, not one long request | A synchronous version is faster to build (minus 2 days) but times out at scale; I would push back |
| Replayed deliveries go through the same workers, breaker and bulkhead | If they need their own pool: plus 2 to 3 days |
| No data migration of existing dead rows is needed | If old rows lack `endpoint_id` indexes on the range: plus 1 day for an index built concurrently |

## 5. Unknowns and how I will shrink them
| Unknown | Spike or question | Who answers | By when | Effect on the range |
| --- | --- | --- | --- | --- |
| U1: safe pace | One-day spike: replay 50,000 rows on kind at three paces, watch delivery lag of healthy tenants | me | end of week 1 | High end drops by about 3 days if the first pace works |
| U2: who may run it | Ask the PM; propose admin keys only | PM | day 2 | Plus 1 day if it needs a new key scope |

## 6. The message
"Bulk replay: 2 to 4.5 weeks, medium confidence. The wide end is one unknown, how fast we can replay without
slowing other tenants; a one-day spike at the end of week 1 will tell us, and I will send a narrower range then.
End of the month is possible if the spike goes well and we agree on admin-only keys by Wednesday."
