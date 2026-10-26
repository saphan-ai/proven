# REVIEW — corpus-pv-006 · round 3

- verdict: accept-with-conditions
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 0 BLOCKING · 1 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-23T21:31:36Z

## Summary

PV-006 reproduces its recorded evidence at bb594c8; the retired path is fixed. Correct the stale run date and macOS sentence before publication.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — corpus-pv-006

### Ground

Review target: `bb594c8bbcd131cd5822848e2c9140ea970ee285`, measured by `git rev-parse HEAD`; first `git rev-parse --short=7 HEAD` printed `bb594c8`. Scope: `9303b2d...bb594c8`. Local toolchain: `go version go1.26.5 linux/amd64`, matching the version and platform recorded at `test/corpus/cases/PV-006/CASE.md:95`. The case has its own `go 1.26` module. No declared numeric testing bar or baseline was found in the case or order; no bar value is inferred. The check is an author-controlled suite instrument, so its result is read with the independent reference and source, not as a stand-alone proof.

### Differential

`git diff 0b5a194..bb594c8 -- test/corpus/cases/PV-006/CASE.md` shows only two `tests/corpus` to `test/corpus` replacements at lines 98 and 116. `git diff --name-status 9303b2d...bb594c8` names only nine files under `test/corpus/cases/PV-006/` and `test/corpus/manifest.json`; `git diff --check 9303b2d...bb594c8` is clean. `rg -n 'tests/|fleet/|conformance/' test/corpus/cases/PV-006` found no match. The case's relative `../../manifest.json` and `expected/EXPECTED.md` references resolve to files present in the worktree.

### Unplug

`go -C test/corpus/cases/PV-006 test -count=1 -v ./check` exited 0 and printed `negative control alarm: rendered notice does not contain "$wait"` after the test replaced that token in the rendered text. This distinguishes the actual raw-token observation from a mere reference mismatch. The test returns green when the alarm fires as intended; a green exit alone would not show what the notice contained. Source inspection at `check/main.go:32-39` and `check/main.go:57-63` confirms that the same evidence helper is called on the dollar-token input and an absent token aborts the check.

### Closure

`bash test/corpus/cases/PV-006/check.sh` exited 0. Its entire output, compared line for line with CASE.md §8, was:

```
well-formed base=pass change=pass reference=pass
listed-token base=fail({limit}) change=fail({limit}) reference=fail
unlisted-token base=pass change=fail({window}) reference=fail
dollar-token base=pass change=pass reference=fail raw=$wait
duration-0-vs-1 base=pass change=pass reference=fail shown=0 retry-after=1
duration-equal base=pass change=pass reference=pass shown=30 retry-after=30
```

The well-formed and brace-token lines show the change control accepts normal text and catches the earlier brace attacks. The dollar-token line shows a raw token in the rendered notice while the change control passes; the zero-versus-one line shows its duration control passes an unequal value. The hand-written reference fails both. `expected/EXPECTED.md`, read after the implementation and check, names these two outcomes as the established defect. The opposite outcome would require the raw token to be absent or the change control to reject it, and the unequal duration to be rejected; the instrument can print `change=fail` on brace tokens, so those false-world answers are distinguishable. The output does not establish behaviour for every placeholder shape or another language.

### Remainder

OWED — `test/corpus/manifest.json:113` says checks ran on 2026-10-22, while `CASE.md:97` records the check on 2026-10-23. Reconcile the date with the actual run record. `CASE.md:127` says a macOS run was unavailable, while the review order reports a 2026-10-23 darwin/arm64 run by the master; update the sentence only if that run is recorded as evidence. These are provenance wording defects; they do not alter the reproduced result. No blocking finding.

### Profile

Docs and Go test corpus. The changed round-3 lines are commands in CASE.md, both executed locally. The case adds no dependency beyond the standard library. No threshold, test assertion, skip, suppression, or coverage setting was changed. The suite check is more circular than an independent external check; the reference texts and explicit raw-token observation narrow that limitation. No numeric bar was declared, so base/tip bar deltas are unmeasured.

### Focus

Round-2 path condition: ADDRESSED at `CASE.md:98,116`; the documented commands now resolve and run. Scope and privacy: the diff stays in the PV-006 case and its manifest line; no private path or private implementation detail appears in the changed content. Ruling: report the date and macOS sentence as OWED, because both are wording about provenance and do not make the case wrong or unrunnable. Cost if wrong: readers may misattribute which platform and day independently reproduced the check.

### Conditions

Correct the manifest check date against the recorded run; reconcile the macOS availability sentence with a cited macOS run. Neither condition blocks acceptance of the fixture evidence.

### Unrun

macOS reproduction: this seat is linux/amd64; the master's darwin/arm64 run is a claim in the order, not independently rerun here. Whole-repository tests and a base-commit suite: outside this re-review's path-fix scope; no code or test behaviour changed at bb594c8. No other check is claimed.

## The review status as filed

## REVIEW — corpus-pv-006
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: PV-006 reproduces its recorded evidence at bb594c8; the retired path is fixed. Correct the stale run date and macOS sentence before publication.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-23T21:27:49Z

## REDACTED

none
