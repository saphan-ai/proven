# STATUS — corpus-pv-003 · round 50

- state: ready-for-review
- executor: codex
- model: gpt-6-sol
- pool: not recorded
- cost: not recorded
- duration: 1 min
- last commit: ceba195a6d2cc3dfaf63b502d63d56fbedb70555 test/corpus: name the corpus by its path after the ADR-0030 move
- updated: 2026-10-24T06:03:40Z

## Summary

GPT-6 attested ceba195; moved-tenant input printed base=refused(X), change=allowed; the five-line check matched CASE.md §8.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — corpus-pv-003
- state: stop-2-ready
- abstract: GPT-6 attested ceba195; moved-tenant input printed base=refused(X), change=allowed; the five-line check matched CASE.md §8.
- last commit: ceba195a6d2cc3dfaf63b502d63d56fbedb70555 test/corpus: name the corpus by its path after the ADR-0030 move
- done: The master's carried tip and path repair are present; the PV-003 check matches the recorded result.
  - `git rev-parse --short=7 HEAD` printed `ceba195` (exit 0); `git rev-parse HEAD` printed `ceba195a6d2cc3dfaf63b502d63d56fbedb70555` (exit 0). This attests the carried tip in this worktree.
  - `git show --format=fuller --stat --no-renames HEAD` (exit 0) shows the path repair in CASE.md, check.sh, and check/main.go; `bash test/corpus/cases/PV-003/check.sh` completed from the repository root (exit 0).
  - The five output lines below match CASE.md §8, excluding timing, by direct comparison.
- next / blocked-on: Return the status for master verification; independent native macOS confirmation remains owed at the gate.
- artifacts:
- evidence:
  - `bash test/corpus/cases/PV-003/check.sh` exit 0; full output follows.
  - own-case-cooled-slot | base=allowed | change=refused(A) | reference=refused(A)
  - own-case-other-slot | base=allowed | change=allowed | reference=allowed
  - move-moved-tenant | base=refused(X) | change=allowed | reference=refused(X)
  - vacated-slot | base=refused(X) | change=refused(X) | reference=allowed
  - control-no-cooldown | base=allowed | change=allowed | reference=allowed
- note: No code or commit was made in round 3. Native macOS execution was not run on this Linux seat. No master claim was refuted by the measured HEAD, path repair, or check output.

## REDACTED

none
