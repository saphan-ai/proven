# REVIEW — corpus-pv-010 · round 4

- verdict: reject
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 2 BLOCKING · 1 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-23T16:51:44Z

## Summary

The checked-in check fails before exercising its fixture because its case path predates the directory move. A path-corrected in-memory diagnostic reproduces the expected defect, but cannot validate the shipped command. Fix the path and recorded invocation, then rerun.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — corpus-pv-010

### Ground

`git rev-parse --short=7 HEAD` printed `44c261e`. Review scope is `9303b2d...44c261e`. `git diff --name-only 9303b2d...HEAD` contains only PV-010's case (outside `expected/`) and its own manifest line. `git diff --check 9303b2d...HEAD` passed. Go 1.26.5 linux/amd64 is installed. The case uses three standard-library Go modules. The manifest preserves a `status` object with dated fixture/check claims. The entire checked-in case path changed from `tests/corpus` to `test/corpus`, but its executable path and two documented commands did not.

### Differential

**BLOCKING — `test/corpus/cases/PV-010/check.sh:10`**. Running the precise ordered command `bash test/corpus/cases/PV-010/check.sh` at HEAD printed:

```
toolchain: go version go1.26.5 linux/amd64
test/corpus/cases/PV-010/check.sh: line 20: cd: ███████████████████████████████████████████████████████████████████████████████████████████ No such file or directory
build failed: base
check_exit=1
```

The local absolute path above is diagnostic only and must never be copied into the public repository. `case_dir` points at absent `tests/corpus` while the actual case is under `test/corpus`. Remedy: correct that reference and the invocation in `test/corpus/cases/PV-010/check.sh:4` plus the commands in `test/corpus/cases/PV-010/CASE.md:113` and `test/corpus/cases/PV-010/CASE.md:132`, and repeat the recorded check on this layout. The latter document claims exit 0 from an earlier SHA, not from this checkout. Its case is otherwise in the proper lane.

To separate the path defect from the fixture's behavior, I ran the check body **in memory**, changing only its root, case directory, and scratch location to this checkout and its reviewer scratch directory; no tracked file was altered. Full output:

```
toolchain: go version go1.26.5 linux/amd64
pinned count: 42
current | base=pass | change=pass | reference=current
stale | base=fail | change=fail | reference=stale
twenty-digits | base=pass | change=fail | reference=stale
minus-999 | base=pass | change=pass | reference=stale
plus-41 | base=pass | change=pass | reference=stale
arabic-indic-41 | base=pass | change=pass | reference=stale
description | stated widths: base=none | change=0..48 | change-edited=0..256 | reference=0..48 (49 widths) | change tests on change=pass | change tests on change-edited=pass
```

The corrected diagnostic exited 0; exit 0 means completion, not that the case has no defect. `test/corpus/cases/PV-010/base/guard.go:33` skips conversion failures; `test/corpus/cases/PV-010/change/guard.go:38` refuses them. Both skip signed and other-script closing tokens before conversion because their digit-only matcher finds no pair. `test/corpus/cases/PV-010/reference/main.go:29` reads the token independently and calls these cases stale. Overflow differs between base and change as §2 requires. `test/corpus/cases/PV-010/change/guard_test.go:36` mutates the derived character class, not the stated width; the width-only change-edited mutant survives. This supports the expected outcome but is not the shipped check.

### Unplug

False-world answer: if change rejected the signed and other-script tokens, the three `change=pass` rows would instead say `change=fail` while the independent token reader still reports stale. If its test detected a false width, the change-edited test would say `fail`, not `pass`. At HEAD the checked-in script instead says `build failed: base` in either world; therefore that instrument alone cannot prove or disprove the expected outcome. The reference width probe invokes the change binary and measures only its matcher, not a separately specified policy. The script does not compare expected results or exit nonzero on a semantically wrong table.

### Closure

**Reject** because the primary runnable reproduction fails after relocation. The round-two `GOTOOLCHAIN=local` and status-object repair are present. The master's claim of exit 0 is refuted at this HEAD. The path-corrected diagnostic does support the substantive expected defect without pretending that `-999` produced a pair or that a real published note contained it. Full texts of the two round-one reviews are not in this worktree; no assertion that every individual condition was paid follows from the brief alone.

### Remainder

The failed original check created untracked `███████████████████████████`; the diagnostic created `███████████████████████████████████████████`. A scoped cleanup command was refused by the shell tool with `rm -f style commands are not permitted`; I did not rephrase a denied deletion. These generated artifacts remain and require an authorized actor for cleanup. This does not satisfy the requested pristine return. `GOTOOLCHAIN=local` prevents toolchain auto-download. Bash, Go, `sed`, `grep`, `tr`, `head`, `cmp`, `cp`, and `printf` are used in Linux/macOS-compatible forms, but macOS execution was not measured. The diff scan for local absolute path prefixes, sample email strings and named credential variables returned no matches; this is not a proof of the absence of all private data.

### Profile

Test-corpus and documentation profile: read both guards, the reference caller, the changed test, `check.sh`, and manifest. New `go.mod` files contain no third-party dependencies. Check ranking: project-owned independent reference token reader (strongest for token semantics); project-owned width probe calling the change executable (measures the matcher, not the prose); change's own suite (weakest, surviving the width mutant); prior recorded output (not a current run). No quantitative test-bar baseline, threshold or accepted exception is defined here; no threshold was edited. Ruling: classify the obsolete execution path as blocking because a judge cannot run the published reproduction. If that ruling were wrong, the published command would still produce no discrimination between defect and no defect.

### Focus

1. Reproduction: `test/corpus/cases/PV-010/check.sh:10` fails before cases; the in-memory path-only correction runs all §2 inputs.
2. Expected outcome: corrected diagnostic shows signed/other-script `change=pass` versus `reference=stale`; overflowing number now fails; width-only description edit survives tests while matcher still accepts only 0..48.
3. Proof quality: `test/corpus/cases/PV-010/change/guard_test.go:36` only checks the derived character class; it never checks the separately stated width.
4. Privacy: limited diff scan found no private material; the absolute path in this local diagnostic must be redacted from any public copy.
5. Scope: only the PV-010 tree outside `expected/` and its manifest entry changed.
6. Portability: the stale path fails regardless of Linux/macOS; macOS itself was not run.

### Conditions

- BLOCKING: correct `test/corpus/cases/PV-010/check.sh:10`, its usage comment and `CASE.md`'s two recorded commands; rerun the checked-in check from the current repository layout and record its actual output and exit status.
- OWED: authorized cleanup of the generated untracked `█████████████` scratch tree, without disguising the tool refusal as successful cleanup.

### Unrun

- The unmodified checked-in check through completion: refused at its first build directory.
- macOS: this seat is Linux.
- Comparison against individual round-one review texts: those texts were not provisioned; only conditions described by the brief could be checked.
- Numeric base-versus-tip metrics: none declared for this fixture.

## The review status as filed

## REVIEW — corpus-pv-010
- state: recommendation-ready
- verdict: reject
- ref: _REVIEW_NOTES.md
- abstract: The checked-in check fails before exercising its fixture because its case path predates the directory move. A path-corrected in-memory diagnostic reproduces the expected defect, but cannot validate the shipped command. Fix the path and recorded invocation, then rerun.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-23T16:47:58Z

## REDACTED

| what | times |
|---|---|
| absolute or private path | 4 |
