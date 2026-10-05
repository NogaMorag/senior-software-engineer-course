# Mentoring plan: Lior joining Hookline (worked example)

Date: 2026-10-05  Mentor: you  Their manager: the team lead  Length: 4 weeks, reviewed each Thursday.

Lior is invented for this example: a junior engineer with one year of Python and no Go.

## 1. Starting point
| Area | What they already know | What they told me they want to learn |
| --- | --- | --- |
| Language(s) | Python, pytest, a little SQL | Go, and "how to write tests that matter" |
| The domain | what a webhook is | why delivery is "at least once" |
| Operating production | never been on call | how to read a dashboard during an incident |

## 2. The goal at the end of week 4
"Lior can take a small Hookline ticket from the backlog to production without help, and has shipped one change
to the worker with a test that would have caught a real bug."

## 3. Weekly plan
| Week | Goal | A task that proves it | What I do | How we know it is done |
| --- | --- | --- | --- | --- |
| 1 | Run Hookline locally and explain the delivery path | Draw the flow from `POST /v1/events` to the sink, from the code, and present it in 10 minutes | Pair for one hour on the first run; answer questions; no code review yet | The drawing is right, and the questions show they read the code |
| 2 | Write Go that passes review | Fix a small bug in the sink or the API with a table-driven test | Review within 4 hours; blocking comments only, with a link to why | The PR merges after at most two review rounds |
| 3 | Change the worker safely | Add a metric to the worker (for example attempts by status code) and a dashboard panel | Pair on the design for 30 minutes, then hands off | The metric shows up in Grafana on kind |
| 4 | Own a ticket end to end | Pick a ticket, estimate it with a range, ship it behind a flag | Only answer questions; they write the status update | Shipped; the estimate is compared with the actual time in the review |

## 4. Rhythm
- One-on-one: weekly, 30 minutes, their agenda first, then the week's goal.
- Code review: week 1 to 2 blocking comments only; from week 3 also design and naming, labelled as nits.
- Pairing: twice a week, an hour, they drive.
- Help: stuck for 30 minutes means ask, with what they tried; that is a rule, not a weakness.

## 5. What I will not do
Take the keyboard; push commits to their branch; answer a question before asking what they already tried.

## 6. Review notes
- Week 1: ____
- Week 2: ____
- Week 3: ____
- Week 4: ____
