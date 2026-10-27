# STATUS — corpus-pv-010 · round 34

- state: ready-for-review
- executor: codex
- model: gpt-6-sol
- pool: not recorded
- cost: not recorded
- duration: 3 min
- last commit: 394b2804dc2fc0a400ffd3afd16a6870abf056dd [corpus-pv-010] tests: reconcile PV-010 checks with Linux run
- updated: 2026-10-23T09:59:49Z

## Summary

Codex (model ID not exposed by this run) paid both manifest conditions; for minus-999 base=pass and change=pass, while reference=stale.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — corpus-pv-010
- state: stop-2-ready
- abstract: Codex (model ID not exposed by this run) paid both manifest conditions; for minus-999 base=pass and change=pass, while reference=stale.
- last commit: 394b2804dc2fc0a400ffd3afd16a6870abf056dd [corpus-pv-010] tests: reconcile PV-010 checks with Linux run
- done: PV-010 round 2 conditions addressed and committed.
  - The Linux run at 457031534a09d291251572cb9e4f87577070fcb0 matched the recorded macOS lines except the toolchain line; bash tests/corpus/cases/PV-010/check.sh exited 0.
  - The PV-010 manifest status is an object with fixture built 2026-10-22 and checks run 2026-10-22; git diff 4570315..HEAD shows only its status and PV-010's CASE.md and check.sh.
  - GOTOOLCHAIN=local prevents a missing Go 1.26 toolchain from causing a network download; bash tests/corpus/cases/PV-010/check.sh exited 0 after the edit.
- next / blocked-on: Master independently reruns the Linux check for the review's gate-side condition.
- artifacts:
- evidence:
  - bash tests/corpus/cases/PV-010/check.sh before edits: exit 0, nine output lines, go1.26.5 linux/amd64; full output in CASE.md section 8.
  - bash tests/corpus/cases/PV-010/check.sh after edits: exit 0, same nine output lines.
  - python3 -m json.tool tests/corpus/manifest.json: exit 0.
  - git diff --check: exit 0.
  - git diff --stat fe146e0..HEAD: only PV-010 case files outside expected and tests/corpus/manifest.json.
- note: Master was wrong in round 1: ordering manifest status as the string built broke the status object shape; this round restored that shape. Review NOTE-1 and NOTE-3 remain judgement items, with no check result changed. No independent check was run by this author.

## REDACTED

none
