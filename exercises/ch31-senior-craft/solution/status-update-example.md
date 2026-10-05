# Status updates: worked examples

## A weak status update (do not send this)

```text
Hi all, quick update on replay. This week I worked on the replay job and had a lot of meetings.
I also looked into the batching thing we talked about and did some testing on kind, which took
longer than expected because of the cluster. Still working on the load test, there were some
issues with lag but I think it's mostly fine. Should be done soon hopefully. Let me know if you
have questions!
```

What is wrong with it:
- The state and the risk are buried, and "mostly fine" and "soon hopefully" carry no information.
- It reports activity ("worked on", "looked into", "meetings"), not outcomes.
- "Some issues with lag" is the one important fact, and it has no number, no impact and no next step.
- No date, no decision asked for, so the reader cannot act on it.

## A good status update

```text
Bulk replay: at risk for Oct 30; the range is now Oct 30 to Nov 6.
Risk: replaying at 500 rows/s raised delivery lag for other tenants above our 30 s SLO in the
kind load test. Decision needed from Dana by Thursday: is a slower replay (about 4 hours for
200,000 rows) acceptable, or do we spend 3 more days on a separate worker pool?

Done since Oct 12:
- API for create and status merged behind the flag replay_bulk.
- Background job re-queues in batches of 1,000; idempotent on restart (test added).

Next, until Oct 21:
- Load test at 100 and 200 rows/s, with lag numbers for each.
- Alert and runbook entry for a replay with no progress.

Changed: the range moved by a week because of the lag result above.
```

Why it works: the first line says the state and the date; the risk has a number and an effect; the decision has
an owner, a deadline and two priced options; done and next are outcomes; the change is explained.

## An incident update

```text
[SEV2] Hookline delivery: about 30% of webhooks to all receivers are delayed by 5 to 15 minutes,
since 14:05 IDT. No deliveries lost so far.
Status: identified.
What we know: the 14:00 release changed the worker lease from 30 s to 3 s; workers are
re-claiming deliveries that are still in flight.
What we are doing: rolling back the release (Noa); watching delivery lag and duplicate count (me).
Next update: 14:45, or sooner if the status changes.
```

What it does not contain: guesses as facts ("probably the database"), blame, internal channel names, or a
promise of a fix time nobody can keep.
