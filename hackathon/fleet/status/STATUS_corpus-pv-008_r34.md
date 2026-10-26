# STATUS — corpus-pv-008 · round 34

- state: ready-for-review
- executor: codex
- model: gpt-6-sol
- pool: not recorded
- cost: not recorded
- duration: 1 min
- last commit: d27b0d722163eab2c0ffc8f0b64e5441aa2ebd4a test/corpus: name the corpus by its path after the ADR-0030 move
- updated: 2026-10-24T06:45:29Z

## Summary

Carried tip d27b0d7 is present; the repository-root check exited 0 and matched CASE.md §8. For the wedged stop, base and change both printed no-frame for the late subscriber. No code changed in this round.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — corpus-pv-008
- state: stop-2-ready
- abstract: Carried tip d27b0d7 is present; the repository-root check exited 0 and matched CASE.md §8. For the wedged stop, base and change both printed no-frame for the late subscriber. No code changed in this round.
- last commit: d27b0d722163eab2c0ffc8f0b64e5441aa2ebd4a test/corpus: name the corpus by its path after the ADR-0030 move
- done: Attested the master's carry and path repair, and ran the four-input check on this tip.
  - `git rev-parse --short=7 HEAD` printed `d27b0d7`; `git rev-parse HEAD` printed `d27b0d722163eab2c0ffc8f0b64e5441aa2ebd4a`.
  - `git log --oneline 9303b2d..HEAD` showed the two carried PV-008 commits followed by the path repair at d27b0d7.
  - `git diff 1970cfc..HEAD -- test/corpus/cases/PV-008/CASE.md test/corpus/cases/PV-008/check.sh` showed only retired `tests/corpus` paths changed to `test/corpus`.
  - `bash test/corpus/cases/PV-008/check.sh` exited 0 and matched the four lines in CASE.md §8, apart from unprinted run timing.
- next / blocked-on: Master verification and owner gate.
- artifacts:
- evidence:
  - `bash test/corpus/cases/PV-008/check.sh` exit code 0; output follows in order.
  - normal-close-then-late-subscriber | no-frame | frame:ended | frame:ended
  - wedged-stop-then-late-subscriber | no-frame | no-frame | frame:ended+alarm:worker-not-released
  - wedged-stop-then-read-state | state:closed | state:closed | state:closed
  - wedged-stop-then-second-close | returned | blocked | returned
- note: The provisioned review verdict is accept with no conditions. This round did not run an independent macOS check or tests outside PV-008. No model identity, cost, or pool usage was supplied in the leg's measurement record, so none is inferred. No master claim was refuted by this round's measurements.

## REDACTED

none
