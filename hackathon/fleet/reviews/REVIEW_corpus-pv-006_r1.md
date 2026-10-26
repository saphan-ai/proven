# REVIEW — corpus-pv-006 · round 1

- verdict: reject
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 1 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-22T18:25:45Z

## Summary

The fixture recreates the named conditions, but the printed check cannot prove that a raw dollar token reaches the visitor. The same verdicts fit a different non-raw text mismatch. Make the escaped notice observable before accepting the established-defect label.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — corpus-pv-006

### Ground

`git rev-parse --short=7 HEAD` printed `7a0e7ab`. Reviewed `fe146e0...7a0e7ab`. `go version` printed `go version go1.26.5 linux/amd64`, matching the module's Go 1.26 declaration and CASE.md §8. `git diff --name-only fe146e0...7a0e7ab` listed only PV-006 files outside `expected/` and its own manifest entry. `git diff --check fe146e0...7a0e7ab` exited 0. No private host, seat, absolute path, private repository name, real token, or email address appears in the diff.

`bash tests/corpus/cases/PV-006/check.sh` exited 0 and printed:

```
well-formed base=pass change=pass reference=pass
listed-token base=fail({limit}) change=fail({limit}) reference=fail
unlisted-token base=pass change=fail({window}) reference=fail
dollar-token base=pass change=pass reference=fail
duration-0-vs-1 base=pass change=pass reference=fail
duration-equal base=pass change=pass reference=pass
```

`go -C tests/corpus/cases/PV-006 test ./...` exited 0 with five `[no test files]` packages; it is not an independent check. `go -C tests/corpus/cases/PV-006 vet ./...` exited 0. No coverage or mutation threshold/baseline is declared. Ranking: hand-written reference is a project check, the executable is a suite check, and neither is an external auditor. No exception or weakening of an existing assertion was found.

### Differential

`tests/corpus/cases/PV-006/product/product.go:15-26` fills `{wait}` only and copies `$wait` unchanged. `tests/corpus/cases/PV-006/base/base.go:9-24` recognizes listed brace tokens and always passes duration; `tests/corpus/cases/PV-006/change/change.go:10-22` recognizes all brace tokens and tolerates a duration up to two seconds short. `tests/corpus/cases/PV-006/check/main.go:39-48` supplies all six §2 inputs. `tests/corpus/cases/PV-006/reference/reference.go:8-23` holds independent hand-written expected notices and exact duration equality. The six observed results agree with CASE.md §8. The well-formed and earlier brace inputs confirm the new control is not simply disabled. Reading the renderer and input establishes that the actual current dollar input escapes raw; the printed results alone do not.

### Unplug

BLOCKING — `tests/corpus/cases/PV-006/check/main.go:51-56`: the dollar case prints only three verdicts, not the visitor-visible notice, and never asserts that it contains `$wait`. In the false world where the product instead emits a different incorrect suffix, for example `30 sec`, `change.CheckText` still passes and `reference.CheckText` still fails against its hand-written `30 seconds` (`tests/corpus/cases/PV-006/reference/reference.go:12,17-19`). The exact same line prints while no raw token reaches the visitor. This instrument cannot discriminate the named defect from another mismatch. Remedy: assert and report the escaped raw token, and add an isolated negative control where removing the raw token raises an alarm. Likewise expose or assert the duration values rather than relying only on their input name. No code-mutating control was run on this read-only review.

### Closure

Verdict: **reject**. The current source implements the condition, but the check's printed evidence does not independently establish the required raw-token finding; the opposite causal explanation fits its output. Strengthen the checker without weakening the reference, then rerun the six inputs and record the new output. No product files were changed by this reviewer.

### Remainder

The constructed base is not the historical pre-change code. The case does not establish that other controls, another language, or an earlier source scan fail. The authored reference is a specification, not a recording of a deployed visitor. Declined to judge: historical private incident and licensing (outside scope).

### Profile

Test corpus / Go. `tests/corpus/cases/PV-006/check.sh:4-6` uses Bash (`BASH_SOURCE`), `dirname`, `pwd`, and Go's `-C` flag; no GNU-only flag or network call is present. A macOS run is still required to verify portability in that environment. No dependencies beyond the Go standard library or install-time scripts were introduced.

### Focus

The controls catch the earlier listed and unlisted brace inputs; the gap is the proof instrument, not an assertion loosened in an existing suite. The dollar input's `reference=fail` establishes inequality only. The current source corroborates the expected defect, but another mismatch gives identical printed evidence.

### Conditions

No accept conditions: the blocking checker remedy is in Unplug at `tests/corpus/cases/PV-006/check/main.go:51-56`.

### Unrun

macOS execution: no macOS host in this leg. Isolated code-mutating negative control: not performed in this read-only review. Historical incident: not accessible in this diff. Coverage/mutation scores: no command, threshold, or baseline declared for this case.

## The review status as filed

## REVIEW — corpus-pv-006
- state: recommendation-ready
- verdict: reject
- ref: _REVIEW_NOTES.md
- abstract: The fixture recreates the named conditions, but the printed check cannot prove that a raw dollar token reaches the visitor. The same verdicts fit a different non-raw text mismatch. Make the escaped notice observable before accepting the established-defect label.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-22T18:21:57Z

## REDACTED

none
