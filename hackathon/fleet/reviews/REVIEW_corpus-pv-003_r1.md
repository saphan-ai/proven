# REVIEW — corpus-pv-003 · round 1

- verdict: accept-with-conditions
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-22T18:02:40Z

## Summary

PV-003 reproduces both move failures and its fixed-placement controls on Linux; require the macOS run before treating portability as verified.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — corpus-pv-003

### Ground

Reviewed `fe146e0...93ce778`; measured HEAD: `93ce778f1c7e7a8bde67d6496bfbe5fb348ae24b`. The seven changed paths are confined to `tests/corpus/cases/PV-003/` outside `expected/`, plus the PV-003 status in `tests/corpus/manifest.json:58-63`. No dependency or lockfile changed. `go version` printed `go version go1.26.5 linux/amd64` against `go.mod:3` (`go 1.26`). `git diff --check fe146e0...93ce778`, `go vet ./tests/corpus/cases/PV-003/...`, and `gofmt -l` on the four added `.go` files returned no output and exit 0. The author's prior claims were not used as proof.

`bash tests/corpus/cases/PV-003/check.sh` returned exit 0 on Linux and printed exactly:

```
own-case-cooled-slot | base=allowed | change=refused(A) | reference=refused(A)
own-case-other-slot | base=allowed | change=allowed | reference=allowed
move-moved-tenant | base=refused(X) | change=allowed | reference=refused(X)
vacated-slot | base=refused(X) | change=refused(X) | reference=allowed
control-no-cooldown | base=allowed | change=allowed | reference=allowed
```

### Differential

`tests/corpus/cases/PV-003/check/main.go:25-58` runs all five §2 inputs with a fresh gate and fixed clock for each version. The fixed-placement inputs differ at the cooled slot as intended and agree with the reference on both slots. On the move, base refuses X, change allows X, and reference refuses X. On replacement, both base and change refuse X, while reference allows Y. The contract does not require base/change to differ on the replacement; it requires change/reference to differ there. The no-cooldown control allows in all three. `tests/corpus/cases/PV-003/base/gate.go:35-50` keys on node; `tests/corpus/cases/PV-003/change/gate.go:36-52` keys on node and slot and never uses the tenant to select that key; `tests/corpus/cases/PV-003/check/reference.go:23-45` keys the cooldown on the run's tenant and looks up the current occupant. Thus the results follow from the implementations, not a hard-coded result line.

### Unplug

The false-world answer is distinguishable in the printed rows: a corrected change keyed by tenant would print `change=refused(X)` for the move and `change=allowed` for the replacement; reverting to the node key would make the first fixed-placement cooled-slot row `change=allowed`. However, `tests/corpus/cases/PV-003/check/main.go:68-77` only prints results; it never asserts them. The command's exit code would remain zero if those rows inverted. Treat the displayed rows and source as the evidence, never the exit code alone. No temporary mutant was run against this review-only worktree. Check strength ranking: the stated rule is an independent contract; `go vet` and formatting are project/toolchain checks; this printer is a suite-level observational check and cannot gate a mutant by its exit status.

### Closure

Recommendation: accept with the portability condition below. `tests/corpus/cases/PV-003/expected/EXPECTED.md:3-9` calls the change defective in exactly the two printed move outcomes and recognizes that the fixed-slot re-keying works. The contrary interpretation, that location-keyed cooldown is acceptable after a legal move, would require changing §1's rule that the tenant remains cooled; it does not fit that rule and the two printed change/reference mismatches. The reference is a coded rendering of the stated rule, not an external independent oracle. No condition silently reclassifies the defect; the finding remains an established defect in the synthetic change.

### Remainder

No blocking defect in the fixture found. One limitation to carry forward: `tests/corpus/cases/PV-003/check/main.go:76` supplies no machine-enforced expected-output assertion; a later altered fixture could exit zero while printing different evidence. There was no pre-existing assertion to weaken and no skip, suppression, threshold, or baseline change. `tests/corpus/cases/PV-003/CASE.md:91` mentions a seat permission class without naming a host or seat; the changed files contain no internal hostnames, private paths, private repository names, real tokens, or email addresses. Scope excluded licenses and the separate PV-004 case. No numeric quality baseline was declared for this slice, so no base/tip metric delta is claimed.

### Profile

Go test-corpus fixture and documentation. Single-process, fake-clock, in-memory only; no network or external dependency. `check.sh:1-5` invokes Bash (`set -euo pipefail`) and Go from the repository root. Bash `pipefail` is not POSIX `sh`, so invoking it as `sh check.sh` is non-portable; the documented `bash check.sh` form uses features present in macOS Bash and Linux Bash. The installed Go 1.26 toolchain and repository root are prerequisites on either OS. No GNU-only utility or platform-specific shell flag appears; actual macOS execution was not available on this Linux seat.

### Focus

The assignment move and replacement in `tests/corpus/cases/PV-003/check/main.go:40-52` preserve X's run identity while changing the occupied slot, exposing the difference between tenant identity and reused location. `tests/corpus/cases/PV-003/CASE.md:69-80` distinguishes the two historical reviewers' judgment without claiming the accepting reviewer missed the transition. Diff confinement, path scan, and output were checked against the exact reviewed SHA.

### Conditions

- Before a cross-platform guarantee is recorded, run `bash tests/corpus/cases/PV-003/check.sh` from the repository root on macOS with Go 1.26 and compare all five rows above. This is an outstanding verification condition, not evidence of a known macOS defect.

### Unrun

- Native macOS execution: no macOS runner was available in this Linux review seat. A Linux success and shell-source inspection do not prove the macOS outcome.
- Temporary source mutation: not run because this review must not alter the code under review. The source-level false-world rows above identify what a mutation must change, but no mutation score is claimed.

## The review status as filed

## REVIEW — corpus-pv-003
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: PV-003 reproduces both move failures and its fixed-placement controls on Linux; require the macOS run before treating portability as verified.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md

## REDACTED

none
