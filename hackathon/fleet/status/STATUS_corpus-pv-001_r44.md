# STATUS — corpus-pv-001 · round 44

- state: ready-for-review
- executor: codex
- model: gpt-6-sol
- pool: not recorded
- cost: not recorded
- duration: 2 min
- last commit: f23a1a480691883b87f550e59fc830ef524a58ff [corpus-pv-001] tests: expose frame IDs and rerun the case
- updated: 2026-10-24T04:06:44Z

## Summary

Codex (routed model ID not exposed to this leg): duplicate input base=pass, change=pass, reference=fail; the frame IDs now show base=["other","home"], change=["home"], reference=["other"].

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — corpus-pv-001
- state: stop-2-ready
- abstract: Codex (routed model ID not exposed to this leg): duplicate input base=pass, change=pass, reference=fail; the frame IDs now show base=["other","home"], change=["home"], reference=["other"].
- last commit: f23a1a480691883b87f550e59fc830ef524a58ff [corpus-pv-001] tests: expose frame IDs and rerun the case
- done: Paid the round-2 review conditions assigned to this leg.
  - C1 ADDRESSED: Removed answer-revealing comments from both guards; rg control exited 1 with no matches.
  - C2 ADDRESSED: npm ci with an artifact-local cache exited 0; check.sh exited 0 on the tip and printed 13 inputs. Full output is in CASE.md section 8.
  - C3 ADDRESSED: Every line now prints base_ids, change_ids, and reference_ids; duplicate input reads ["other","home"], ["home"], and ["other"].
  - C4 ADDRESSED: Removed the reference reader's break; duplicate input still reads reference_ids=["other"].
  - C5 NOT ADDRESSED: The master owns .gitignore by the order.
  - C6 NOT ADDRESSED: The owner rules on prior art by the order.
- evidence:
  - npm ci --cache ████████████████████████████████████ --prefix test/corpus/cases/PV-001: exit 0, added 4 packages, audited 5 packages, found 0 vulnerabilities.
  - bash test/corpus/cases/PV-001/check.sh: exit 0, 13 input lines; duplicate-attribute base=pass change=pass reference=fail missing=home base_ids=["other","home"] change_ids=["home"] reference_ids=["other"].
  - git diff --check: exit 0.
  - git diff --stat 9303b2d...HEAD: only test/corpus/cases/PV-001 and test/corpus/manifest.json.
- artifacts:
- note: Review 2's OWED evidence gap is addressed by printing the reference's first duplicate value. The package pins remain unchanged. The exact routed model ID and monthly pool shares were not exposed to this leg.
- next / blocked-on: Master verifies this round and routes C5 and C6 to their owners.

CLOSE-OK · corpus-pv-001 · commit f23a1a480691883b87f550e59fc830ef524a58ff · branch fix/corpus-pv-001 · worktree █████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 2 |
