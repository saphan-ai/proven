# REVIEW — corpus-pv-009 · round 3

- verdict: reject
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 1 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-23T17:45:44Z

## Summary

The case runner fails after the corpus directory moved; the recorded nine-line output is not reproducible at this tip. Repair its paths and rerun before relying on this case.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — corpus-pv-009

### Ground

RED — BLOCKING: `test/corpus/cases/PV-009/check.sh:4` executes a path absent after the corpus move. `test/corpus/cases/PV-009/check.go:9-10` imports that same retired directory. The advertised nine-line check cannot run at the reviewed tip. Remedy: correct the runner, imports and `CASE.md:96` to use `test/corpus`, rerun, and record the new measurement.

`git rev-parse --short=7 HEAD` printed `627907d`. Reviewed `9303b2d...627907d`. The diff lists only six case files outside `expected/`, plus the case's own status fields in `test/corpus/manifest.json`. I attempted the literal check before reading `expected/EXPECTED.md`. No reviewed code was edited.

### Differential

`bash test/corpus/cases/PV-009/check.sh` printed exactly:

```text
stat ./tests/corpus/cases/PV-009/check.go: no such file or directory
```

Exit 1; zero result lines. A second control, `go run ./test/corpus/cases/PV-009/check.go`, exited 1 because its `base` and `change` imports point to nonexistent `tests/corpus` packages. `go test ./test/corpus/cases/PV-009/...` exited 1 for that setup failure; the six existing `change` subtests passed and `base` has no tests. The case did not exist at the selected base; no comparable base case-suite count exists. The master's claimed exit 0 and `CASE.md:96`'s exit 0 cannot be reproduced at this tip.

### Unplug

The runner fails before testing any case; a live negative control on its output cannot be run. Source inspection: `change/guard.go:22-24` discards conversion errors and equals-pinned values together; `change/guard_test.go:10-22` has no conversion-failure input. Changing the error branch to refuse would not be caught by these six existing tests (source-derived, not a measured mutation). Independent reference `check.go:15-29` compares closing digits as text. In the false world where the change correctly refuses numbers above MaxInt64, `change=fail reference=fail`; the documented `change=pass reference=fail` would distinguish it if a working check printed it. A broken runner prints the same error in both worlds, so its current result establishes neither outcome.

### Closure

**Reject:** `CASE.md:96-109` records a successful nine-line check, but the shipped runner exits 1. The fixture sources agree with §2: base reads only the first pair (`base/guard.go:19`), change reads every pair (`change/guard.go:19`), and both skip conversion errors (`guard.go:22-24`). `expected/EXPECTED.md` agrees with the source-derived overflow distinction and the recorded lines, but these lines are not independently printed by the current runner. No check was bent to force success: nine required inputs are present and the independent reference compares text rather than reusing the guard. This finding targets execution and evidence, not the deliberately flawed guard.

### Remainder

Round-2 changes in `CASE.md:94-98` describe the toolchain caveat and acknowledge macOS was not checked. The moved-layout path correction remains unpaid in `CASE.md:96`, `check.sh:4`, and `check.go:9-10`. Earlier reviewers' verbatim condition lists were not provided, so their other conditions cannot be individually certified. `git diff --check` passed. Inspection of the case diff and its manifest entry found no private host/seat names, private paths/repository names, tokens, or e-mail addresses.

### Profile

Corpus review. The author-controlled suite is the most circular check; source inspection of the text-based reference is a distinct control, but the reference is not executing. The Go-only subtest passes while the entire case fails to set up. No dependencies or performance claims changed. Bash, `dirname`, `cd` and `go run` are expected on Linux and macOS with Bash/Go installed; the broken relative path is platform-independent. No macOS run was made.

### Focus

`test/corpus/cases/PV-009/check.sh:4` and `test/corpus/cases/PV-009/check.go:9-10`: broken runner after move. `test/corpus/cases/PV-009/CASE.md:96`: the recorded command does not produce its stated output. `test/corpus/cases/PV-009/change/guard_test.go:10-22`: deliberate omission of conversion-failure inputs. Licences, original unpublished code and the other two historical findings are declined as outside scope.

### Conditions

BLOCKING: fix the three retired path references; run the literal check from the repository on Linux; verify every printed result against §2 and `expected/EXPECTED.md`; correct §8's command, exit status and output. Run on macOS before claiming it works there, or continue to state it is unverified. Do not fix the guard's intentional overflow defect in this corpus case.

### Unrun

Nine live case results: the runner fails before executing an input. macOS: no macOS host available. Condition-by-condition comparison of earlier reviews: their full condition lists were not provisioned. Mutation of reviewed files: review-only role, runner cannot execute.

## The review status as filed

## REVIEW — corpus-pv-009
- state: recommendation-ready
- verdict: reject
- ref: _REVIEW_NOTES.md
- abstract: The case runner fails after the corpus directory moved; the recorded nine-line output is not reproducible at this tip. Repair its paths and rerun before relying on this case.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-23T17:39:02Z

<!-- ENGINE-EMITTED REVIEW SKELETON, rendered from ██████████████████████████████.
     Everything below this line contains WORKED FORMS to copy as applicable.
     inside an HTML comment, and this engine's parsers skip comment interiors,
     so nothing here is read as a declaration — copy the applicable form OUT.

     Write-legal `- state:` tokens: spawned · recommendation-ready.
     Last act before done: set `- state:` above to recommendation-ready and add every
     return-time key shown below. Their names, value classes, and vocabularies
     are projected from the review family's rows, never copied here.

     `verdict`: exactly one write-legal verdict token, alone on the line — mandatory for `recommendation-ready`; no parenthetical and no conditions after it; the conditions belong in the review body the `- artifacts:` manifest names
     - verdict: <accept | accept-with-conditions | reject>

     `ref`: the review's returned artifact — mandatory for `recommendation-ready`; one of this document's own `- artifacts:` entries, VERBATIM (a basename that is not a declared entry is refused at delivery); never a git ref
     - ref: _REVIEW_NOTES.md

     `artifacts`: NESTED DASH LIST of paths the review returns; mandatory for `recommendation-ready`, including an explicit empty list when nothing is returned
     - artifacts:
       - _REVIEW_NOTES.md
       - _REVIEW_STATUS.md

     An entry names a FILE, or a DIRECTORY: a directory is delivered as ONE
     deterministic `.tar` archive (sorted entries, fixed mtimes — the same
     bytes, and so the same digest, every time) recorded under the entry's
     own name plus that suffix, which is the name the projected manifest and
     the teardown gate then read. A path that EXISTS and is neither is
     refused by what it IS, never as `missing`.

     Explicit empty manifest form when the review returns no files:
     - artifacts:

     🔴 THE ANALYSIS DOES NOT GO IN THIS FILE. It goes in _REVIEW_NOTES.md, whose
     mandatory members are H2 HEADINGS — never top-level dash bullets in
     _REVIEW_STATUS.md. A section heading written as `- ground:` here is an unknown
     key: it is now reported as a WARNING rather than refused, so your
     verdict still lands — but the analysis lands nowhere at all.

     The complete _REVIEW_NOTES.md form, projected from the same one section source
     this engine renders it from (██████████████████████████████):

     # REVIEW NOTES — corpus-pv-009

     ## Ground
     <replace with measured review content; write `none` only when none is the measured answer>

     ## Differential
     <replace with measured review content; write `none` only when none is the measured answer>

     ## Unplug
     <replace with measured review content; write `none` only when none is the measured answer>

     ## Closure
     <replace with measured review content; write `none` only when none is the measured answer>

     ## Remainder
     <replace with measured review content; write `none` only when none is the measured answer>

     ## Profile
     <replace with measured review content; write `none` only when none is the measured answer>

     ## Focus
     <replace with measured review content; write `none` only when none is the measured answer>

     ## Conditions
     <replace with measured review content; write `none` only when none is the measured answer>

     ## Unrun
     <replace with measured review content; write `none` only when none is the measured answer>
-->

## REDACTED

| what | times |
|---|---|
| file and line in Saphan Studio's private code | 2 |
