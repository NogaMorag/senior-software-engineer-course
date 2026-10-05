# Status update template

Rules: fits on one screen; the first line is the state and the risk; every "late" has a new date or a decision
needed; nothing that only you understand (no internal jargon, no ticket numbers without words).

```text
<Feature>: <on track | at risk | off track> for <date or range>.
Risk: <the one thing that could move the date, or "none new">. Decision needed from <name> by <date>: <question>,
      or "none".

Done since <last update date>:
- <outcome, not activity>
- <outcome>

Next, until <next update date>:
- <outcome>
- <outcome>

Changed: <what changed in scope, range or assumptions, and why>, or "nothing".
```

## Variant: incident update (during an incident)

```text
[<SEV level>] <service>: <impact in user terms>, since <time, with time zone>.
Status: <investigating | identified | mitigating | monitoring | resolved>.
What we know: <one or two facts, no guesses presented as facts>.
What we are doing: <the current action and who is on it>.
Next update: <time>, or sooner if the status changes.
```
