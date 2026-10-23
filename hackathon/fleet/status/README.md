# fleet/status/

A status report is what the executor writes back: what it built, what it ran, and what happened. It answers the order's acceptance item by item, with real output, not with a summary of it.

**Rules every status report follows**

- **The state is one of three:** `in-progress`, `blocked` or `ready-for-review`.
- **Blocked is a normal answer.** An executor that cannot proceed says what blocks it and whose decision it is. It does not guess, and a stop is not counted as a failure.
- **Every acceptance item gets the command that was run and the result it gave.**
- **Numbers, not adjectives.** Test counts, exit codes, timings.
- **Every run states its model and its cost, to the cent.** The cost is the one Saphan Studio recorded for the run, not an estimate.

**Template**

```markdown
# STATUS — <slug> · round <n>

- state: <in-progress | blocked | ready-for-review>
- executor: <which Cursor agent>
- model: <the model the agent ran on; with Auto, the model Auto chose>
- pool: <share of the monthly model pool, before → after this run>
- cost: <$0.00, to the cent, for this run>
- duration: <minutes, for this run>
- last commit: <commit> "<commit message>"
- updated: <date and time, UTC>

## Summary
<at most five lines: what was built and what proves it — or, when blocked, what stops the work>

## Acceptance, item by item
1. <acceptance 1: the command run, and the result it gave>
2. <acceptance 2: …>

## Evidence
<test counts, command output, numbers>

## Blocked on / next
<the exact blocker and whose decision it is — or what comes next>
```
