# REVIEW — corpus-pv-010 · round 5

- verdict: accept
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-23T21:51:22Z

## Summary

The four path references are fixed; the check reproduces the recorded case and establishes the expected defect. Generated check outputs remain untracked because cleanup was refused; authorized cleanup is needed before teardown.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — corpus-pv-010

### Ground

Review target `6b101a9` (measured with `git rev-parse --short=7 HEAD`); base `9303b2d`, repair parent `44c261e`. `CASE.md` §2 is the contract; §8 is the claimed reproduction. I read `expected/EXPECTED.md` only after inspecting the case, changed paths, check and output. Verdict: accept. The path repair is ADDRESSED; the case still demonstrates the established defect and misleading green tests.

### Differential

- ADDRESSED — `test/corpus/cases/PV-010/CASE.md:113`, `:132` and `test/corpus/cases/PV-010/check.sh:4`, `:11`: `git diff 44c261e..6b101a9` changes four occurrences of `tests/corpus` to `test/corpus` and nothing else (two files, four insertions, four deletions). The executable path and documented command now exist. `rg -n 'tests/|fleet/|conformance/' test/corpus/cases/PV-010` returned no matches.
- `git diff --name-only 9303b2d...6b101a9` lists eleven case files and `test/corpus/manifest.json`; `git diff --check 9303b2d...6b101a9` passes. No other package, dependency or test bar was changed by the repair. The case diff and its inputs contain no identified private credential, host address or private path; provenance mentions a private record but exposes no internal path or implementation.

### Unplug

The negative control built by `check.sh:20-26` changes only `statedMaxGap` from 48 to 256; `change/guard.go:23-24` still matches at most 48, while `change/guard_test.go:34-45` mutates characters, not widths. The check reports `change-edited=0..256`, independently probed `reference=0..48 (49 widths)`, and tests passing for both. If the description test watched the width, its false-world answer would be FAIL on change-edited. The observed PASS therefore shows the missing alarm; it is not evidence the width claim is true. `change/guard.go:33-36` explains the signed and other-script escape: no pair is found, so the note passes.

### Closure

I ran `bash test/corpus/cases/PV-010/check.sh` from the repository root; exit code 0. Full stdout:

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

All lines after the toolchain line reproduce both §8 recorded runs byte-for-byte. The Go version differs from the macOS run but matches the recorded Linux run. The twenty-digit result proves the cannot-convert repair is real; the three signed/other-script lines and width mutant demonstrate the remaining defects. `expected/EXPECTED.md` classifies this as an established defect and requires these comparisons. A case with the opposite outcome would have to make signed and other-script notes agree with the reference or make the width mutant fail its tests; these measured lines distinguish that world. The reference's choice of ASCII closing token and the help page's wider promise are explicit parts of this fixture, not a claim about every possible policy.

### Remainder

No BLOCKING or OWED findings in the reviewed repair. No claim that a published note actually contained a signed or other-script number. Deletion of the check's generated `███████████████████████████` directory was refused by the command tool (`rm -f style commands are not permitted. Use a safer approach`); I did not bypass the denial. Consequently the worktree has generated untracked files and cannot be represented as clean. An authorized cleanup role must remove these outputs before teardown; this is an operational return impediment, not evidence against the case.

### Profile

Fixture/test review. The local suite plus reference probe is not an external oracle: the reference measures widths by invoking the changed matcher. A green test alone would be circular: the changed-width mutant passes despite a false description. No declared quantitative coverage, mutation or lint baseline appears in this case; no bar is claimed met. No dependency manifest or lockfile changes in the path-only repair.

### Focus

Only `9303b2d...6b101a9`, especially the round-2 path condition and `44c261e..6b101a9`; `test/corpus/cases/PV-010/CASE.md` §2 and §8, `check.sh`, implementation, reference, manifest entry and answer key. Licenses and unrelated repository behavior were not judged. The master's run is corroborated on Linux by my independent run, not accepted on authority.

### Conditions

none

### Unrun

No independent macOS run or full repository suite; those are outside this path-repair review. Generated work directory cleanup was attempted and refused as stated under Remainder; no replacement deletion command was used.

## The review status as filed

## REVIEW — corpus-pv-010
- state: recommendation-ready
- verdict: accept
- ref: _REVIEW_NOTES.md
- abstract: The four path references are fixed; the check reproduces the recorded case and establishes the expected defect. Generated check outputs remain untracked because cleanup was refused; authorized cleanup is needed before teardown.
- artifacts:
    - _REVIEW_NOTES.md
    - _REVIEW_STATUS.md
- updated: 2026-10-23T21:47:46Z

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
