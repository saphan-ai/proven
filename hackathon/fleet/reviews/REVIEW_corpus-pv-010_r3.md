# REVIEW — corpus-pv-010 · round 3

- verdict: reject
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-23T15:14:40Z

## Summary

BLOCKING: the shipped checker still targets tests/corpus after the move to test/corpus; the recorded Linux command is not runnable here. An isolated copy reproduces the intended evidence, not the shipped invocation. Fix both paths and rerun at this layout.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — corpus-pv-010

### Ground

RED HEADER — BLOCKING: `test/corpus/cases/PV-010/check.sh:10` still sets `case_dir` to `tests/corpus/cases/PV-010`, but the case now lives under `test/corpus/cases/PV-010`. `test -d tests/corpus/cases/PV-010` returned 1 at HEAD `a9927e7` (from `git rev-parse --short=7 HEAD`). Its build at line 20 cannot find its source. `CASE.md:113,132` and the comment at `check.sh:4` record the same absent path. Repair both paths and rerun at the committed layout; do not infer that the recorded exit-0 Linux run occurred at this layout. `git diff --name-status ae34d44...a9927e7` lists only PV-010 files outside `expected/` and PV-010's manifest status object. `git diff --check ae34d44...a9927e7` exited 0. No lockfile or new dependency. Local toolchain: `go version go1.26.5 linux/amd64`, fixture modules require Go 1.26. A targeted case scan showed no actual credential, email address, internal machine name, or private absolute path.

### Differential

The committed entry point was not run because it writes to `███████████████████████████`, outside this leg's assigned artifact area. I copied the unchanged script and `base/`, `change/`, and `reference/` into `█████████████████████████████████████████████████████████████████████` and ran `bash █████████████████████████████████████████████████████████████████████████████`; exit 0:

```text
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

`change/guard.go:32-39` searches for unsigned ASCII digits before parsing; signed and other-script tokens find no pair and pass both versions. The overflow finds a pair and fails only on change, so the three-way parse change is real. `reference/main.go:27-43` reads each closing token as stale unless it exactly equals ASCII `42`. `CASE.md:12-17` says every closing number is compared. This evidence establishes the intended difference, not that `-999` produces a pair. In the false world where change treated signed and other-script tokens as stale, the three change cells would read `fail`; they read `pass`.

### Unplug

`change/guard.go:17` holds `statedMaxGap` independently of the matcher width at line 13. The copied script changed only the former from 48 to 256: reference still measured 0..48, while `change/guard_test.go:37-45` passed. That test changes `gapClass`, not width, so it cannot detect this disconnect. In the false world of a width-agreement test, `change tests on change-edited` would say `fail`, not the observed `pass`. The reference gap probe calls the change binary but the test suite only tests its own matcher, so the suite check is the more circular instrument.

### Closure

Verdict: `reject`. The case's expected established-defect outcome follows from the isolated reproduction, but the shipped entry point and both documented commands no longer run in the moved layout. This new blocking finding is outside round-two's stated repair. No product code or answer key changed.

### Remainder

Round two added `GOTOOLCHAIN=local`, provided a Linux output claim, and restored the dated PV-010 manifest status object. Toolchain non-download behavior is addressed; runnable recorded command is NOT ADDRESSED after the move. Exact earlier review-condition text was not provisioned, so their full closure is unverified. `sed -e`, `cmp -s`, `grep -o`, `head -1`, Bash `printf`, and standard Go have Linux and macOS implementations; no macOS command was run here. An attempted `rm -rf` cleanup was denied by the command guard, so confined scratch files remain; no alternate spelling of that denied act was tried.

### Profile

Corpus fixture, test-only; no auth, persistence, or new dependency. No declared coverage or mutation metric baseline was available. Declined to judge: licenses, unpublished private source, and unrelated corpus cases. Ruling: the isolated copy is a diagnostic, not validation of the committed entry point; treating it otherwise would hide a broken judge command.

### Focus

File · line · control: `test/corpus/cases/PV-010/check.sh:10` / absent source / committed entry point; `test/corpus/cases/PV-010/CASE.md:132` / stale Linux command / path existence; `test/corpus/cases/PV-010/change/guard_test.go:37` / width-test negative control / edited-constant reproduction.

### Conditions

Repair the checker source path and both recorded invocations to `test/corpus`, then run the shipped script at this layout on Linux and reconcile its recorded output. The signed-number and width defects are the intended evidence; do not fix them in the fixture or modify `expected/`.

### Unrun

The exact committed checker was not run because its scratch output is outside this leg's confinement and its input path is absent. macOS was not available on this seat. The original round-one review orders were not provided. Cleanup was denied by the command guard; confined scratch files remain.

## The review status as filed

## REVIEW — corpus-pv-010
- state: recommendation-ready
- verdict: reject
- ref: _REVIEW_NOTES.md
- abstract: BLOCKING: the shipped checker still targets tests/corpus after the move to test/corpus; the recorded Linux command is not runnable here. An isolated copy reproduces the intended evidence, not the shipped invocation. Fix both paths and rerun at this layout.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-23T15:06:06Z

## REDACTED

| what | times |
|---|---|
| absolute or private path | 3 |
