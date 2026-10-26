# REVIEW — corpus-pv-019 · round 4

- verdict: accept
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-23T22:22:07Z

## Summary

PV-019 reproduces its nine recorded lines byte for byte; the round-2 path defect is addressed, and the answer key follows from the differential.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — corpus-pv-019

### Ground

Reviewed `e3ffdafd336e45b79b98837b8ab8141d6989da48` (measured with `git rev-parse HEAD`, matching the ordered tip). `git diff a8977d8..e3ffdaf` changes only the two old `tests/corpus` references in `test/corpus/cases/PV-019/CASE.md:95-96` and one in `test/corpus/cases/PV-019/check.sh:3` to `test/corpus`; it changes no executable behaviour. `git diff --name-status 9303b2d...e3ffdaf` lists only PV-019 case files and `test/corpus/manifest.json`. `rg -n 'tests/|fleet/|conformance/' test/corpus/cases/PV-019` returned no matches. The changed paths exist. No private identifier was found in the reviewed diff; no material outside the named lane was changed.

`bash test/corpus/cases/PV-019/check.sh` exited 0 on Go `go1.26.5 linux/amd64`. `cmp -s` of its stdout against the nine-line block in `CASE.md` §8 returned `SECTION8_BYTE_MATCH`. Actual stdout:

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

`go -C test/corpus/cases/PV-019 test -count=1 ./...` passed the one test in `change`; the other three packages report no test files. `git diff --check 9303b2d...e3ffdaf` exited 0.

### Differential

The shipped row passes both change and reference, whereas the base refuses its unfamiliar status (`table/table.go`, `base/validator.go`, `change/validator.go:18-31`). Mutating the preview session switch or plugin folder separately leaves the changed validator passing but makes the independent reference refuse (`check/main.go`); likewise an install route on unsupported. A preview install route is refused by both. A wrong nonempty failure channel passes the changed validator while the content test refuses that known value. The empty channel is refused by both. These outcomes cover all seven contract inputs, and the content test's two extra lines.

### Unplug

The check's false-world answer is visible: if the changed validator closed all preview delivery fields and the unsupported install route, the three named mutated inputs would show `change: fail`, not `change: pass`. If the channel were closed in the validator, the wrong-nonempty input would show `change: fail`; the separate content test already turns red for that particular value. A check that only ran the shipped table would produce the same pass even with those holes and would be insufficient. The recorded mutant lines distinguish the two worlds; `check/main.go` computes reference results separately from either validator.

### Closure

The sole round-2 condition, the retired path in both documentation and command comment, is ADDRESSED at `CASE.md:95-96` and `check.sh:3`. The shell check executes and its output exactly matches §8. The expected answer, read after the code and check, calls this an established defect in the fixture's changed validator, not a defect in this repair: the three latent holes and the distinct content-test failure match `expected/EXPECTED.md`. This `accept` of the corpus repair is not acceptance of the fixture's defective validator.

### Remainder

No blocking or owed findings in this round. Declined to judge: licensing, downstream behaviour of a real integration, and future review model performance; the fixture is a local synthetic table and the order excludes licensing. The documented limitation that no shipped row is wrong remains accurate (`CASE.md:118`).

### Profile

Docs and Go fixture, with no runtime code changed in this repair. The Go 1.26 module uses only the standard library (`go.mod`); no dependency or lockfile delta. The repository-root command and isolated module test both ran. The Go toolchain used here is `go1.26.5 linux/amd64`, not the documented macOS run's `go1.26.8`. No performance assertion is part of the repair. The full-repository suite was outside this path-only re-review and was not run.

### Focus

FILE · LINE · CONTROL: `CASE.md:95-96` and `check.sh:3` now name a runnable existing location; `check/main.go` drives all contract mutations; `change/validator.go:20-30` leaves the intended latent holes; `change/content_test.go:9-13` pins the shipped channel. Verdict: `accept` for the corrected corpus case.

### Conditions

none

### Unrun

none

## The review status as filed

## REVIEW — corpus-pv-019
- state: recommendation-ready
- verdict: accept
- ref: _REVIEW_NOTES.md
- abstract: PV-019 reproduces its nine recorded lines byte for byte; the round-2 path defect is addressed, and the answer key follows from the differential.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-23T22:20:00Z

## REDACTED

none
