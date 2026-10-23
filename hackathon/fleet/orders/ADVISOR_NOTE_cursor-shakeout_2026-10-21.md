# ADVISOR NOTE — the test corpus as the Cursor shake-out, 21 October 2026

- to: the master of `proven-test-corpus`
- from: the advisor of the hackathon track "The Agent That Can Explain Why"
- kind: advice, not an order. The order is yours to amend; take from this what you can use.

## The owner, verbatim (written in Polish; given in the advisor's translation)

> *"Exactly so: every run of ours has a fresh context. But since tomorrow they write the competition work, we have to warm
> them up a bit and find the bugs, because there certainly are some. It's a new adapter, so I'd rather catch them today
> than tomorrow and lose time."*

So the test corpus has two jobs today: build the case code, and find what breaks in running Cursor agents before the build
window opens on 22 October 00:00 UTC. **A failure today is a result, not a setback.** Record each one where the Cursor
backend train's master can act on it, the same hour.

**Separation, settled by the owner:** a test-author run and a later implementation run are separate because each run has
a fresh context. Each case records the seat, the model that really ran, and the run id, so the separation can be checked.

## Before the first leg

1. The engine on macm5 and runner-cursor includes both far-Cursor fixes (the sandbox on far Linux; the hook and CLI
   settings for far work folders). Measure it with `merge-base --is-ancestor` on each host; do not take it from a message.
2. The W6 smoke turn of `proven-cursor-infra`, again: it must start **without** `--egress-open`, and Shell must answer.
   If either fails, stop here and report. A test author that cannot run its own test is yesterday's failure again.

## What to hit on purpose

Each line is a question the first legs can answer cheaply. The order you take them in is yours.

| # | what to exercise | what a bug looks like |
|---|---|---|
| S1 | one Go case and one TypeScript case | `go test` or `npm test` refused, or a toolchain missing in the leg's environment though present on the host |
| S2 | the module and package downloads a case needs (Go proxy, npm registry) | a download refused by the network policy, or allowed though not on it |
| S3 | a leg that tries to reach a host not on its policy | it gets through: the sandbox is not holding |
| S4 | the same case kind on macm5 (local) and on runner-cursor (far) | different results, or different files in the return |
| S5 | two legs at once, one per seat | one seat's leg blocks or reuses the other's state |
| S6 | the return: `_EXEC_STATUS.md`, the commit, the files the case declares | a leg reported `ok` with nothing committed, or files outside its case |
| S7 | the model pin: `composer-2.5`, never Auto | the run records another model than the one pinned |
| S8 | cost and pool: tokens recorded; the owner's pool reading before and after | cost "unknown" without the pool reading next to it |
| S9 | a leg that runs out of time on purpose (a short limit on one cheap case) | `killed-timeout` that reads as a clean finish |
| S10 | commit messages | any AI attribution line: no `Co-Authored-By`, no tool name, no session link |
| S11 | a case whose expected outcome is a **failing** test before the fix | the leg "fixes" the test so it passes, instead of leaving the failure the contract asks for |

S11 is the one most likely to be wrong in a way nobody notices: a test author eager to see green can make the case prove
nothing.

## Reporting

- One line per finding in your progress record: what was run, what happened, what was expected, the run id.
- Engine defects go to the master of the Cursor backend train, high salience, with the run id. Defects in the contracts go
  back to the scribe.
- At the end of the day, one sentence for the owner: is Cursor ready for the build, yes or no, and what is left.
