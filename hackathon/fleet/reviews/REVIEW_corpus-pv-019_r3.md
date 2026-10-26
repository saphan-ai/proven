# REVIEW — corpus-pv-019 · round 3

- verdict: accept-with-conditions
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-23T17:53:44Z

## Summary

The case reproduces the latent validator defects and channel split; correct the answer key's false shipped-table claim and obsolete check paths before acceptance. Exact round-one condition closure remains unverified.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — corpus-pv-019

### Ground

`git rev-parse --short=7 HEAD` printed `a8977d8`. `git diff 9303b2d...a8977d8` changes only this case outside `expected/` and PV-019's manifest status object. Local `go version`: `go1.26.5 linux/amd64`, matching the case's major/minor pin. `go -C test/corpus/cases/PV-019 test ./...` exits 0: one passing test in `change`, three packages with no test files. `git diff --check 9303b2d...a8977d8` exits 0. The changed files contain no real credentials, email addresses, private repository names, private host/seat names or absolute private paths. No product files were edited.

### Differential

`bash test/corpus/cases/PV-019/check.sh` exited 0 and printed exactly:

```text
shipped-table | base: fail [ed-cedar: unknown status "preview"] | change: pass | reference: pass
preview-with-install-route | base: fail [ed-cedar: unknown status "preview"] | change: fail [ed-cedar: preview row has an install route] | reference: fail [ed-cedar: preview row has a delivery field]
preview-with-session-switch | base: fail [ed-cedar: unknown status "preview"] | change: pass | reference: fail [ed-cedar: preview row has a delivery field]
preview-with-plugin-folder | base: fail [ed-cedar: unknown status "preview"] | change: pass | reference: fail [ed-cedar: preview row has a delivery field]
unsupported-with-install-route | base: fail [ed-cedar: unknown status "preview"] | change: pass | reference: fail [ed-birch: unsupported row has an install route]
channel-wrong-nonempty | base: fail [ed-cedar: unknown status "preview"] | change: pass | reference: fail [ed-cedar: failure channel "ignored silently" is not in the closed set]
channel-wrong-nonempty | shipped content test: fail [ed-cedar failure channel = "ignored silently", want "exit code 2"]
channel-empty | base: fail [ed-cedar: unknown status "preview"] | change: fail [ed-cedar: failure channel is empty] | reference: fail [ed-cedar: failure channel "" is not in the closed set]
channel-empty | shipped content test: fail [ed-cedar failure channel = "", want "exit code 2"]
```

`check/main.go:67-75` covers all seven inputs from CASE §2. Base rejects every variant because each retains a preview row (`base/validator.go:14-18`); change/reference disagree precisely on the two unchecked preview fields, unsupported install route and wrong nonempty channel. The content test fails on both changed-channel inputs. The shipped table is not defective: change and reference pass it.

### Unplug

`change/validator.go:20-30` has no guard for unsupported installation, preview switch or preview folder; the three isolated mutations in `check/main.go:70-72` yield change=pass/reference=fail. `change/validator.go:29` guards only an empty channel; a wrong nonempty channel passes validation and fails `change/content.go:15-27`. The reference is separate code but authored in the same case: shared misinterpretation of the contract could yield a false agreement. False-world answer for the *exit status* of `check.sh` is still zero when validation differs: `check/main.go:76-85` prints, never asserts. The lines, not exit zero, are the evidence. The single shipped Go test is more circular than the project check, which is in turn less independent than the prose contract; no external standards check or numeric quality baseline was provided.

### Closure

**CONDITION — false answer-key requirement.** `test/corpus/cases/PV-019/expected/EXPECTED.md:6-8` demands a shipped-table line "showing all agree". The observed base=fail/change=pass/reference=pass line contradicts it, and `base/validator.go:14-18` must reject the new preview status under §2. Correct the key to require change/reference agreement and base rejection before using it to grade. The core established-defect outcome follows from the observed differential and CASE §2; the same lines cannot show a validator that enforces all three delivery constraints. Severity comes from the recorded ruling, not the program exit code.

**CONDITION — stale reproduction commands.** `test/corpus/cases/PV-019/CASE.md:95-96` and `test/corpus/cases/PV-019/check.sh:3` still name `tests/corpus/` instead of `test/corpus/`. The actual command in this review succeeds at the latter path. Update the literal instructions before a judge uses them. Neither correction belongs to this read-only review leg.

### Remainder

Base's global unknown-preview refusal masks an unsupported-row-only base/change comparison, but the change/reference pair isolates that defect. Bash, `dirname`, `cd` and `go run` require Bash and Go 1.26+, but have no GNU-only calls; Linux was executed, macOS was not. The manifest's `independent_check: not done` is a prior status, not a claim that this review has already been incorporated. Declined to judge the private original incident, production integrations, licensing and severity beyond the recorded ruling.

### Profile

Go corpus fixture plus docs. No new skip, suppression, removed assertion or lowered threshold occurs in the diff. Shipped Go test alone does not detect any of the three delivery-field holes; the printed reference comparison does. No base-versus-tip quantitative metric is defined for the case; claiming one would invent a bar.

### Focus

FILE · LINE · CONTROL: `change/validator.go:20-30` against `check/main.go:70-72` and CASE §2; `change/validator.go:29` against `change/content.go:15-27` and `check/main.go:73-74`; `expected/EXPECTED.md:6-8` against the measured first line; `CASE.md:95-96` and `check.sh:3` against the actual layout. Ruling: accept with conditions because substantive reproduction is measured while answer-key wording and literal instructions require repair; if the grader strictly demands "all agree", hold the case until corrected.

### Conditions

Correct the answer key's shipped-table requirement and both obsolete path directions. Round two restores the PV-019 manifest status object and CASE §8's nine lines match the fresh output. The original round-one reviews' verbatim conditions were not included in this provisioned order or the reviewed diff; exact condition-by-condition closure is not certifiable from the available evidence.

### Unrun

macOS execution: no macOS host on this seat. Original round-one review conditions: their literal text was not provided, so exact closure is unverified. Base-commit case checks: `9303b2d` predates the case module. This review did execute the seven-input check and the shipped Go suite at the reviewed tip.

## The review status as filed

## REVIEW — corpus-pv-019
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: The case reproduces the latent validator defects and channel split; correct the answer key's false shipped-table claim and obsolete check paths before acceptance. Exact round-one condition closure remains unverified.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-23T17:48:33Z

## REDACTED

none
