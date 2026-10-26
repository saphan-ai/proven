# REVIEW — held-out-set-1 · round 2

- verdict: reject
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-24T16:43:22Z

## Summary

BLOCKING: HO-002's frozen key demands failure on a state with a valid current bot marker; the check correctly passes. Publish a post-freeze grading erratum without changing frozen material. Five other cases match their contracts.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — held-out-set-1

### Ground

RED — BLOCKING: The frozen HO-002 answer key says the older-run-marker row should `fail` (`test/corpus/held-out/HO-002/expected/EXPECTED.md:11-13`), yet the frozen contract expressly requires that same state to include a valid current-run bot marker (`test/corpus/held-out/HO-002/CASE.md:39`). The check builds both markers (`test/corpus/held-out/HO-002/check.sh:39-43`), and the gate correctly prints `pass`. A grader following the key can penalize a correct reviewer. Remedy: publish a post-freeze grading erratum excluding this one impossible demand; leave the frozen materials unchanged. If an isolated stale-marker failure is needed, freeze a new case rather than changing this one.

Reviewed `9b11366...589541d` at SHA `589541d` (`git rev-parse --short=7 HEAD` printed `589541d`). With §8 removed, each of the six CASE.md files has a byte-identical SHA-256 digest at `b4bab25` and tip; `git diff --quiet b4bab25..HEAD -- 'test/corpus/held-out/*/expected/*'` exited 0. The six cases address distinct stale evidence, marker authorship, umask, ambiguous API errors, model substitution, and dirty-tree leftovers. No identifiable real-world product, vendor, host or person appears in the case code; names such as alice are synthetic fixture inputs, and generic language/OS names remain.

### Differential

Ran all six checks from an isolated copy under `████████████████████████████████████████████████████`, preserving each script's four-level relative path to its scratch root. Every check exited 0; printed 4, 4, 10, 3, 4 and 3 lines respectively. Decisive lines, copied verbatim:

- HO-001: `success-then-failure base fail change pass reference should fail`; `one-run-success base pass change pass reference should pass`. The change reads any bot success comment, not the latest run (`change/gate.sh:11`).
- HO-002: `person-marker-current base pass change pass reference should fail`; `person-marker-older base pass change pass reference should pass`. The change matches run and commit but not author (`change/gate.sh:14`). The latter row's current bot marker also matches.
- HO-003: `. umask 022 base 0755 change 0755 reference 0755`; `. umask 002 base 0755 change 0775 reference 0755`; `change test umask 022 ok`; `change test umask 002 FAIL (4 entries with a wrong mode)`. The change ORs read bits into creation modes (`change/snapshot/snapshot.go:42`).
- HO-004: `user-exists base create change ok last step team reference should succeed`; `new-user-refused-email base create change team last step team reference should fail at create`. Both 422 causes are treated alike (`change/setup.sh:15`).
- HO-005: `base job-2 asked model-z exit 2 logged - served none reference should not be served`; `change job-2 asked model-z exit 0 logged model-z served model-a reference should not be served`. The change silently substitutes a default; the same wrapper logs the requested name.
- HO-006: `base base ok change - reference should pass`; `change-dirty base - change ok reference should fail`; `change-clean base - change fail reference should fail`. `change/verify.sh:19` reads the obsolete path; `check.sh:42-48` confirms the leftover is untracked and missing in the clean clone.

### Unplug

False-world answers: HO-001 single success passes both gates; HO-002 no-marker fails both and authentic current bot marker passes both; HO-003 umask 022 agrees but 002 differs; HO-004 existing user succeeds in change but refused email fails at team; HO-005 supported model is correctly served and logged; HO-006 clean clone fails where dirty copy passes. These controls distinguish each defect from a trivially broken fixture. A check's exit 0 means it completed the comparison, **not** that the change passed its reference. No source files or frozen answers were mutated for testing.

### Closure

Verdict: reject the six-case set as currently gradable. Five implementations and answer keys follow their frozen contracts and checks; HO-002's main forged-author row does too. But the HO-002 answer key's extra required `fail` for the older marker contradicts a valid current bot marker in that row. A truthful reviewer could be judged wrong. This is one blocker, not a reason to change frozen texts.

### Remainder

No root dependency manifests or CI rules changed; HO-003 adds two isolated Go 1.26 modules without external dependencies. On Go 1.26.5 Linux, both Go builds and the two umask test runs completed within HO-003's check. `TMPDIR` was set inside the confined scratch path for `sh scripts/check-links.sh`, which exited 0. No numeric baseline, command and tolerance were declared for a testing-bar metric; no target was invented. Ranking: frozen §2 contracts are prior independent project checks; case checks and Go tests are author-owned suite checks, so their exit 0 is not an external quality bar. The prewritten key is independent of case code but inconsistent for HO-002. No bar suppression or exception observed.

### Profile

Review covered paired base/change paths, reference data, source and check logic before opening `expected/` last. No network was needed (HO-006 clones a local repository). The master's check counts were re-measured rather than accepted. Ruling: grade contradiction is BLOCKING because it makes a case unfair; cost if wrong is delaying the five sound cases until an erratum. Declined to judge actual Proven results: the held-out reviewer under evaluation must not receive checks or answer keys before its run.

### Focus

File · line · control: `test/corpus/held-out/HO-002/expected/EXPECTED.md:12` demands `fail`, but `test/corpus/held-out/HO-002/check.sh:40-43` writes an exact marker from `ci-bot` for the current run AND an older marker from a person. `test/corpus/held-out/HO-002/change/gate.sh:14` must accept the valid marker. The observed `person-marker-older ... change pass reference should pass` proves that the key's additional requirement does not follow from its check. `test/corpus/held-out/HO-002/CASE.md:39` confirms the valid current marker is intentional.

### Conditions

Publish a grading erratum outside frozen `CASE.md` and `expected/`: the `person-marker-older` row correctly passes while a valid current bot marker exists, so exclude the key sentence requiring failure. Keep the `person-marker-current` forged-author row as the proof of this case's defect. Do not edit frozen material after the freeze.

### Unrun

macOS/Bash 3.2 was unavailable on this Linux seat. The full product suite and Proven's actual evaluation were outside this isolated corpus-contract review; no result for them is claimed. Destructive mutations were not run against the source because this reviewer was ordered to leave it unchanged.
The isolated sandbox copy remains inside `████████████████████████████████████████████`: the cleanup command was refused by the command guard, and the denied deletion was not retried under another spelling.

## The review status as filed

## REVIEW — held-out-set-1
- state: recommendation-ready
- verdict: reject
- ref: _REVIEW_NOTES.md
- abstract: BLOCKING: HO-002's frozen key demands failure on a state with a valid current bot marker; the check correctly passes. Publish a post-freeze grading erratum without changing frozen material. Five other cases match their contracts.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-24T16:41:00Z

## REDACTED

| what | times |
|---|---|
| absolute or private path | 2 |
