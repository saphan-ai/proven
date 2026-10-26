# STATUS — corpus-pv-009 · round 28

- state: ready-for-review
- executor: codex
- model: gpt-6-sol
- pool: not recorded
- cost: not recorded
- duration: 2 min
- last commit: 8640a72db2312a75d62fa18a6eba8cdf91289d1a test/corpus: name the corpus by its path after the ADR-0030 move
- updated: 2026-10-24T06:50:50Z

## Summary

Codex GPT-6 attested 8640a72; the nine-input check matched CASE.md §8 with exit 0. For overflow-int64, base and change passed while the reference failed.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — corpus-pv-009
- state: stop-2-ready
- abstract: Codex GPT-6 attested 8640a72; the nine-input check matched CASE.md §8 with exit 0. For overflow-int64, base and change passed while the reference failed.
- last commit: 8640a72db2312a75d62fa18a6eba8cdf91289d1a test/corpus: name the corpus by its path after the ADR-0030 move
- done: The master's carried tip and path fix are present, and the case check matches §8.
  - `git rev-parse --short=7 HEAD` printed `8640a72`; `git rev-parse HEAD` printed `8640a72db2312a75d62fa18a6eba8cdf91289d1a`.
  - `git log -4 --format='%H %s'` placed the two corpus case commits above base 9303b2d and the path-fix commit at HEAD.
  - `git show --stat --oneline --no-renames HEAD` showed the path-fix substitutions in CASE.md, check.go and check.sh.
  - `bash test/corpus/cases/PV-009/check.sh` exited 0; its nine lines matched CASE.md §8 exactly, excluding timing.
- next / blocked-on: Master to reconcile the review's OWED macOS wording condition with its dated run record before publication.
- artifacts:
- evidence:
  - Command: bash test/corpus/cases/PV-009/check.sh; exit 0; output follows exactly.
  - one-current base=pass change=pass reference=pass
  - one-stale base=fail:LANGUAGES 41 -> 43 change=fail:LANGUAGES 41 -> 43 reference=fail:LANGUAGES 41 -> 43
  - stale-after-current base=pass change=fail:LANGUAGES 41 -> 42; 42 -> 43 reference=fail:LANGUAGES 41 -> 42; 42 -> 43
  - current-after-stale base=fail:LANGUAGES 41 -> 43; 43 -> 42 change=fail:LANGUAGES 41 -> 43; 43 -> 42 reference=fail:LANGUAGES 41 -> 43; 43 -> 42
  - three-pairs base=pass change=fail:LANGUAGES 40 -> 42; 41 -> 42; 42 -> 43 reference=fail:LANGUAGES 40 -> 42; 41 -> 42; 42 -> 43
  - no-arrow base=pass change=pass reference=pass
  - max-int64 base=fail:LANGUAGES 41 -> 9223372036854775807 change=fail:LANGUAGES 41 -> 9223372036854775807 reference=fail:LANGUAGES 41 -> 9223372036854775807
  - overflow-int64 base=pass change=pass reference=fail:LANGUAGES 41 -> 9223372036854775808
  - twenty-digits base=pass change=pass reference=fail:LANGUAGES 41 -> 10000000000000000000
- note: No code or commit was made in round 3, as ordered. The macOS check was not run on this Linux seat. The master's carry and path fix match the order's claims; master-was-wrong: none measured.
- updated: 2026-10-24T06:48:38Z

## REDACTED

none
