# STATUS — corpus-pv-012 · round 3

- state: ready-for-review
- executor: codex
- model: gpt-6-sol
- pool: not recorded
- cost: not recorded
- duration: 1 min
- last commit: 479cfb80ad895c4bad3e5f04a7e6c15b83263193 test/corpus: name the corpus by its path after the ADR-0030 move
- updated: 2026-10-24T07:01:51Z

## Summary

Attested 479cfb8; the check matched CASE.md §8, including control base 5 and change 0. No code changed.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — corpus-pv-012
- state: stop-2-ready
- abstract: Attested 479cfb8; the check matched CASE.md §8, including control base 5 and change 0. No code changed.
- last commit: 479cfb80ad895c4bad3e5f04a7e6c15b83263193 test/corpus: name the corpus by its path after the ADR-0030 move
- done: The master-carried tip and path fix are present; the case check matches the recorded output.
  - `git rev-parse --short=7 HEAD` printed `479cfb8`.
  - `git rev-parse HEAD` printed `479cfb80ad895c4bad3e5f04a7e6c15b83263193`; `git branch --show-current` printed `fix/corpus-pv-012`.
  - `git diff --name-only 99d1b6c..HEAD` listed only `test/corpus/cases/PV-012/CASE.md` and `test/corpus/cases/PV-012/check.sh`, confirming the path fix.
  - `bash test/corpus/cases/PV-012/check.sh` exited 0 and printed the eight lines below, matching CASE.md §8 exactly apart from the documented `exit=0` annotation.
- next / blocked-on: Master verification and gate.
- artifacts:
- evidence:
  - `control | base: 5 (Activity for all teams; control=not-present) | change: 0 (Activity scoped to ours; control=PASS) | reference: 0`
  - `vacuity | base: 5 (Activity for all teams; control=not-present) | change: 5 (Activity scoped to theirs; control=PASS) | reference: 5`
  - `application | base: 5 (Activity for all teams; control=not-present) | change: 5 (Activity for all teams; control=not-applicable) | reference: 0`
  - `scope-no-team | base: 5 (Activity for all teams; control=not-present) | change: 5 (Activity for all teams; control=not-applicable) | reference: 0`
  - `scope-with-team | base: 5 (Activity for all teams; control=not-present) | change: 5 (Activity for all teams (request team: ours); control=not-applicable) | reference: 0`
  - `extra-sync | base: 5 (Activity for all teams; control=not-present) | change: 4 (Activity scoped to ours; control=not-applicable) | reference: 0`
  - `control-no-registry | base: 5 (Activity for all teams; control=not-present) | change: 4 (Activity scoped to ours; control=FAIL) | reference: 0`
  - `control-wrong-ownership | base: 5 (Activity for all teams; control=not-present) | change: 4 (Activity scoped to ours; control=FAIL) | reference: 0`
- note: No master-was-wrong measurement for the tip, check, or path fix. The provisioned accept review has a non-empty Unrun section; law #53 requires the master to resolve that verdict inconsistency before relying on acceptance.

## REDACTED

none
