# Design: <short title>

| Field | Value |
| --- | --- |
| Author | <you> |
| Reviewers | <names: who must say yes, who should read> |
| Status | Draft / In review / Approved / Superseded by <link> |
| Last updated | <YYYY-MM-DD> |

## 1. Context
<What exists today, what problem we have, and why now. Link to the incident, ticket or metric.
A reader who has never seen the system should understand the problem from this section alone.>

## 2. Goals and non-goals
Goals (each one checkable):
- G1: <...> (checked by: <test, metric, SLO>)

Non-goals (things a reasonable reader might expect, that we deliberately do not do):
- NG1: <...> (why not now: <...>)

## 3. Requirements and estimates
- Functional: <who does what, through which API>
- Non-functional: <latency, availability, durability, cost targets>
- Estimate (assumptions file: `<path>.toml`, output pasted below; every number is an assumption):

```text
<output of: uv run sizing <path>.toml>
```

## 4. Proposal
### 4.1 API
<Endpoints or messages, with one example request and response.>

### 4.2 Data model
<Tables, columns, keys, indexes; what is new and what changes.>

### 4.3 High-level design
<One diagram and one paragraph per component: what it owns, what it calls.>

### 4.4 Deep dives
<The two or three parts that are hard: concurrency, failure handling, scale limits.>

### 4.5 Failure modes
| Failure | Effect | Handling | How we prove it |
| --- | --- | --- | --- |

## 5. Alternatives considered
| Option | Why it is attractive | Why not (here, now) | What would change my mind |
| --- | --- | --- | --- |
| Do nothing | | | |

## 6. Risks
| Risk | Likelihood | Impact | Mitigation | Owner |
| --- | --- | --- | --- | --- |

## 7. Rollout and rollback
<Steps in order, each one reversible or explicitly not; migrations (expand, then contract);
feature flags and who flips them; what we watch (dashboards, alerts) and the rollback trigger.>

## 8. Open questions
- Q1: <question> (who decides, by when)
