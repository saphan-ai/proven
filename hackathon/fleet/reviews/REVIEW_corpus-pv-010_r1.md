# REVIEW — corpus-pv-010 · round 1

- verdict: accept-with-conditions
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 0 BLOCKING · 1 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-22T18:08:53Z

## Summary

PV-010 reproduces both defects on Linux; correct the manifest's false "checks: not run" before landing. macOS execution was not available.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — corpus-pv-010

### Ground
Reviewed `fe146e0...4570315`; `git rev-parse --short=7 HEAD` printed `4570315`. I ran `bash tests/corpus/cases/PV-010/check.sh` on Go 1.26.5 linux/amd64; exit 0. Complete stdout:

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

Three modules (`base`, `change`, `reference`) pin `go 1.26` and use only the standard library. `check.sh` builds them and runs `go test -count=1 ./...` on change and change-edited; both captured outputs were `ok pv010/change 0.002s`. The case's recorded macOS 1.26.8 run is the author's claim, not my measurement.

### Differential
The six note inputs and description input in `CASE.md` §2 occur in `check.sh:44-65`. Base and change share the digit-only matcher (`base/guard.go:17`, `change/guard.go:23-24`); base skips parse errors (`base/guard.go:28-31`), whereas change marks them stale (`change/guard.go:37-40`). The twenty-digit repair is real. `-999`, `+41`, and Arabic-Indic digits find no pair and pass in both, as §2 requires. The reference reads the token after the arrow without using the matcher (`reference/main.go:30-43`); its stale results are independent of the guard's digit filter. Its width probe makes 257 separate guard calls (`reference/main.go:49-65`) and finds only 0..48. No `expected/` file changed. The diff adds only PV-010 files outside `expected/` and changes PV-010's own manifest entry.

### Unplug
The check copies change into change-edited and replaces only `statedMaxGap = 48` with 256 (`check.sh:22-28`); it checks the source changed. The matcher still uses `matcherMaxGap = 48`, and `TestDescriptionFollowsMatcher` changes only `gapClass` (`change/guard_test.go:34-45`). The edited description says 0..256 while all change tests still pass. This negative control distinguishes the suite's circular width assertion from the separate reference probe. The false-world answer for a synchronized description would be that changing its stated width changes observed accepted widths or fails a test; neither happens. The reference is an independently specified rule, not proof that a real published note carried these strings or that the only remedy is widening the guard.

### Closure
The outcome in `expected/EXPECTED.md` follows from the signed/other-script disagreements, the real overflow repair, and the surviving edited-width tests. The case intentionally depicts an established defect in *change*; it is not itself the defect. No assertion was weakened to manufacture a passing line: the replacement is checked, the reference does not call the guard for note classification, and the change tests genuinely pass. Script exit 0 reports completion, not correctness of each comparison; unvalidated command substitutions (`check.sh:32-40,61-66`) could print empty fields on a failed binary, so exit code alone is insufficient. `git diff --check fe146e0...4570315` passed.

### Remainder
**OWED — `tests/corpus/manifest.json:188`:** `checks: "not run"` contradicts the check recorded at `tests/corpus/cases/PV-010/CASE.md:109-128` and my independent run. Correct the manifest before landing. **NOTE — `tests/corpus/cases/PV-010/check.sh:6,32-40,61-66`:** missing fail-fast checks on command substitutions could yield empty fields with exit 0; future hardening can validate output shape and propagate errors without changing comparison semantics. No private host, seat, absolute internal path, repository name, credential, or email was found in changed files. The relative `█████████████` scratch path is not an internal absolute path. No dependencies, install scripts, threshold changes, skips, or suppressions were added. No declared numeric quality baseline exists for this case; ranking: independent reference (project-owned but separately implemented), Go suite (author-owned, demonstrably survives width mutation). Declined to judge: licensing (owner scope), unreconstructed third historical finding, and whether the published page should shrink or the guard expand.

### Profile
Fixture review (Go and shell): inspected paired base/change, separate reference, tests, and manifest. Linux check passed as measured above; `git diff --name-status fe146e0...4570315` confined the change to the case and its own manifest line. `bash`, `sed -e`, `cmp -s`, `grep -o`, `head -1`, `tr`, `cp`, `rm`, `mkdir`, and Go commands have Linux and macOS implementations; no GNU-only option was found. `printf '\xd9\xa4\xd9\xa1'` relies on Bash's hexadecimal escapes rather than POSIX `printf` and is compatible with the explicit Bash interpreter. Judges require Go 1.26+ and Bash. macOS runtime was not measured.

### Focus
The author's prior claim that the check exited 0 and touched no expected file was remeasured, not assumed. The manifest discrepancy is metadata drift, not a defect in the fixture's mechanism. Ruling: OWED with a manifest-correction condition; if wrong, public corpus readiness metadata remains misleading. Under §2's reference rule, the opposite outcome (no defect in change) cannot produce both a stale reference note with change=pass and a width claim of 256 against observed maximum 48.

### Conditions
Before landing, set `tests/corpus/manifest.json:188` to an accurate checked state consistent with the measured and recorded run. The case itself requires no repair. No test-bar exception requested.

### Unrun
macOS reproduction was not run: this leg has a Linux host only. Standalone race, vet, whole-repository suite and numeric metric comparisons were not run: the order confines this review to this diff and its check, there is no concurrency path or declared metric baseline, and the base commit has no PV-010 modules to compare. No live endpoint or network activity was run; the contract needs neither.
Cleanup of the check's generated `███████████████████████████` was refused by the command guard (`rm -f style commands are not permitted`). I did not try another spelling of the denied deletion. The generated files remain untracked; the receiving master must dispose of them through an authorized act before requiring a clean worktree. This does not affect the measured review verdict.

## The review status as filed

## REVIEW — corpus-pv-010
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: PV-010 reproduces both defects on Linux; correct the manifest's false "checks: not run" before landing. macOS execution was not available.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-22T18:04:56Z

## REDACTED

| what | times |
|---|---|
| absolute or private path | 2 |
