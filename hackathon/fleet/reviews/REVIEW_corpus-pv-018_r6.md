# REVIEW — corpus-pv-018 · round 6

- verdict: accept
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-23T20:50:06Z

## Summary

PV-018 check reproduces all five CASE.md lines at 0039117; the retired path is repaired, and the mutant exposes the shipped-test gap.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — corpus-pv-018

### Ground

`git rev-parse --short=7 HEAD` printed `0039117`; `git rev-parse HEAD` printed `00391174ee86b43da7ef6f7ef0a4e92ea96b705d`. Reviewed `9303b2d...0039117`, especially the repair `b079dc8..0039117`. `go version` printed `go1.26.5 linux/amd64`, matching the toolchain named in `CASE.md:83`. The check ran from the repository root with its scratch and Go cache under the worktree; its own script set `GOTOOLCHAIN=local`, `GOPROXY=off`, and `GOSUMDB=off`.

### Differential

`git diff b079dc8..0039117 --` shows one changed line, `test/corpus/cases/PV-018/CASE.md:85`: `tests/corpus` becomes `test/corpus`, with no other change. `git diff --name-only 9303b2d...0039117` lists only nine files under `test/corpus/cases/PV-018/` and `test/corpus/manifest.json`; `git diff --check b079dc8..0039117` passes. `rg -n 'tests/|fleet/|conformance/' test/corpus/cases/PV-018 test/corpus/manifest.json` finds no remaining retired path. The documented script exists at its new path. The repair stays in the case's public lane and adds no private identifier in its diff.

`bash test/corpus/cases/PV-018/check.sh` exited 0 and printed the five lines recorded in `CASE.md:90-95`, verbatim:

```
input=shipped base=pass change=pass mutant=pass reference=pass
input=planted-citing-row-outside-the-three-pages base=pass change=fail(changelog#0:160>120) mutant=pass reference=fail(changelog#0:160>120)
input=plain-row-between-budgets-on-a-listed-page base=pass change=pass mutant=fail(install#2:160>120) reference=pass
suite change=pass(4/4)
suite mutant=pass(4/4)
```

### Unplug

`test/corpus/cases/PV-018/check.sh:32-59` runs the unchanged four tests with only `change/validator.go` replaced by the three-page mutant. Both suites pass. The shipped tests therefore do not alarm when the content-derived rule is replaced. The planted-row probe distinguishes the two rules: change and reference reject `changelog#0:160>120`, while the mutant passes. A world in which the tests pin the rule would print `suite mutant=fail`, not this output; a world in which the change itself applies the page list would print `change=pass` for the planted row. The check's own expected input lines at `check.sh:15-22` also reject a different input trace. Reference agreement is not independent proof: `check/main.go` encodes the stated rule separately.

### Closure

**Accept.** `expected/EXPECTED.md:3-10`, read after the case and its validators, requires evidence of the surviving mutant and the planted-row difference, not a claim that today's validator is wrong. Both are printed by the independent run. The retired path condition is ADDRESSED by `CASE.md:85`; no other round-2 condition in this brief remains NOT ADDRESSED. No BLOCKING or OWED finding in this review's scope.

### Remainder

The reproduction establishes one test-power gap in this synthetic case; it does not establish a live production defect or reproduce the original private reviews. Full repository tests were outside the ordered differential. The temporary Go cache remains under `██████████████████████████████████████████` (60 MB): cleanup was refused by the command guard; do not treat the final untracked `████████` as pre-existing. No source file was changed. Cost and monthly-pool readings are not available to this leg; none are inferred.

### Profile

Go test-corpus and documentation re-review (R14). The case uses its own offline Go module and a four-test suite, with no declared coverage or lint baseline to compare. Check strength: the mutant differential is stronger than a passing suite, but weaker than an external independent rule oracle; the reference follows the stated rule. No quality threshold or assertion changed in the repair.

### Focus

Round-2 path repair, runnable command, exact §8 reproduction, mutation discrimination, case-local paths, and public diff scope. License questions and other cases are outside this review.

### Conditions

none

### Unrun

none

## The review status as filed

## REVIEW — corpus-pv-018
- state: recommendation-ready
- verdict: accept
- ref: _REVIEW_NOTES.md
- abstract: PV-018 check reproduces all five CASE.md lines at 0039117; the retired path is repaired, and the mutant exposes the shipped-test gap.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-23T20:41:44Z

## REDACTED

| what | times |
|---|---|
| absolute or private path | 2 |
