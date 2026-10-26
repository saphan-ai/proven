# REVIEW — json-sarif-outputs · round 7

- verdict: reject
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 1 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-24T05:32:46Z

## Summary

BLOCKING: a path through a missing directory and `..` is falsely treated as the same writable file, so a write failure is reported as configuration exit 7. The full Go suite and vet passed; repair the output identity check and add the negative control before gate.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — json-sarif-outputs

### Ground

Reviewed HEAD `e446790` (`git rev-parse --short=7 HEAD` printed `e446790`). `main` is absent locally; `git merge-base origin/main HEAD` printed `ce61e444ead43c21a846f2123853d03881316a13`, so the inspected W2 diff was `ce61e44...HEAD`. The Go module pins 1.26; `go version` printed `go1.26.5 linux/amd64`. There is no Makefile. `go test -count=1 ./internal/cli/...` passed both CLI packages. `go test -count=1 ./...` exited 0, and `go vet ./...` exited 0. The tests use the vendored SARIF 2.1.0 schema and remove `version` as its negative control. These commands measure the tip, not the base.

### Differential

The W2 diff adds JSON and SARIF writers, a schema test, three extra output paths, load-time collision checks and write-time guards. The round-four repair in `internal/cli/config.go` adds `outputIdentity`; the prior text-only fields appear in both writers. I set the author's verification claims aside and reran the CLI suite and full suite.

### Unplug

BLOCKING — `internal/cli/config.go:576-604` (`outputIdentity`), caller at `internal/cli/config.go:465-470`. Trigger: the two output names `report.json` and `missing/../report.json`, with `missing` absent. The resolver sets `missing = true` after `Lstat(missing)` fails, then collapses `..` lexically and reports both names as one file. The filesystem cannot traverse the absent directory; the second name is not writable. From the scratch directory, the built binary command `proven scan --sample --json-output report.json --sarif-output missing/../report.json` exited 7 and said the flags name the same file. This is a false collision refusal before review. The write-failure contract calls for exit 2. Control: retain the unresolved component until a write, or classify this path as unwritable instead of equating it with a reachable file; add a test that keeps the missing directory absent and asserts exit 2. An unplug that changes only that component to an existing directory should then assert exit 7. The current collision tests do not cover the absent-component-and-parent case.

The existing tests do have useful unplug controls: removing SARIF `version` makes the schema check fail; swapping warning and error levels fails `TestSARIFLevelFollowsSeverity`; dropping evidence or unresolved fails `TestSARIFPropertiesAreTheContract`; removing symlink resolution fails `TestOutputsThroughASymlinkedDirectoryAreRefused`; removing the write-time guard fails `TestAliasCreatedAfterLoadIsRefusedAtWrite`. I read these assertions and the writers, but did not edit the product to run those mutants.

### Closure

Reject until the false exit-7 case is repaired and its test passes. The previously reported same-file overwrite through a symlinked directory is addressed by the load check and its regression test. The JSON contract test compares documented fields, including evidence arrays and empty lists; SARIF tests cover no evidence, unresolved findings, failed services and malformed `version`. No finding about the writer's evidence mapping was admitted.

### Remainder

The run's SARIF properties do not include every JSON-only model field; the documented SARIF contract only promises the text-only statements, which are present. I did not turn that difference into a finding. No numeric coverage or mutation bar is declared by the order, so no bar value is asserted. The new dependency is `github.com/santhosh-tekuri/jsonschema/v5 v5.3.1`, used by the schema tests; the production writer has no new third-party import. The checked-in schema's provenance was not independently verified.

### Profile

Backend and docs: the output writer adds file I/O and a machine-readable contract. The external schema check is less circular than the repository's own contract tests; the suite is the most circular check. No authentication or network path is changed in this diff.

### Focus

Conditions 1, 2, 4 and 6 were checked through the writers, docs, comments and tests. Condition 3 fails for the trigger above. Condition 5 lacks a negative control for that trigger. The `outputIdentity` path comment claims the missing remainder can be appended lexically; that choice causes the false identity. The exported formatter names and package comments have doc comments, and the new tests name their properties.

### Conditions

Repair the blocking path identity classification and add the two-sided missing-directory test before this stream's gate. The fix must preserve exit 7 for true aliases and exit 2 for an unwritable output path.

### Unrun

Base-commit tests and a revert-the-fix mutant were not run: this review leg was confined to the current checkout and did not change the code under review. Independent verification of the vendored schema's upstream bytes was not run. No coverage or mutation command was declared. A scratch binary remained after the cleanup command was refused by the command guard. `git status --porcelain` therefore lists `████████` as untracked, in addition to the provisioned channel files. No source file was changed.

## The review status as filed

## REVIEW — json-sarif-outputs
- state: recommendation-ready
- verdict: reject
- ref: _REVIEW_NOTES.md
- abstract: BLOCKING: a path through a missing directory and `..` is falsely treated as the same writable file, so a write failure is reported as configuration exit 7. The full Go suite and vet passed; repair the output identity check and add the negative control before gate.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-24T05:27:34Z

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

     # REVIEW NOTES — json-sarif-outputs

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
| absolute or private path | 1 |
| file and line in Saphan Studio's private code | 2 |
