# REVIEW — json-sarif-outputs · round 8

- verdict: accept-with-conditions
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 0 BLOCKING · 1 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-24T07:26:08Z

## Summary

JSON, SARIF and output collision claims pass measured tests; correct one inaccurate path-resolution sentence and remove the temporary base archive before merge.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — json-sarif-outputs

### Ground

Reviewed `remotes/origin/main...e289682`; this checkout has no local `main` ref. `git rev-parse --short=7 HEAD` printed `e289682`. The diff changes 21 files. `go version` printed `go1.26.5 linux/amd64`; `go.mod` requires Go 1.26. The author's reported results were treated as claims.

Fresh measurements: `go test -count=1 ./internal/cli/...` passed two packages on the tip. `go test -count=1 ./...` passed every tested package on the tip with no FAIL. `go vet ./...` and `git diff --check remotes/origin/main...e289682` passed. An archive of `remotes/origin/main` gave `go test -count=1 ./internal/cli/...`: one package passed. No cached result was used. The archive remains under `█████████████████████████████████████████████████` because the cleanup command was refused by the tool policy.

### Differential

`internal/cli/format_json.go:21-215` maps the documented JSON keys to `review.Result`, including both evidence types, unresolved, and empty arrays. `internal/cli/format_sarif.go:50-218` makes one SARIF run, a result per finding, evidence and unresolved in result properties, and failed services and text-only statements in run properties. `internal/cli/run.go:200-274` renders requested formats from one result and reports a write error as exit 2. `internal/cli/config.go:435-629` compares output identities at load, following symlinks before parent components, and compares existing hard links. `internal/cli/run.go:249-272` checks aliases introduced before a later write. The paired text writer at `internal/cli/format.go:72-181` was read against both structured writers.

### Unplug

`TestSampleSARIFMatchesTheSchema` removes only `version` from a generated log and verifies schema rejection. `TestSARIFShapesMatchTheSchema` covers no evidence, unresolved, failed service, empty path, line zero, reversed lines, unknown severity, and empty result. `TestJSONDocumentIsTheContract` and `TestSARIFPropertiesAreTheContract` compare decoded values; a renamed JSON key or dropped SARIF evidence or unresolved property fails. `TestSARIFLevelFollowsSeverity` checks level by finding ID. `TestCollidingOutputPathsAreRefused`, `TestOutputsThroughASymlinkedDirectoryAreRefused`, `TestDanglingAliasOutputsAreRefused`, and `TestLinkLoopAndLongChainAreRefusedAtLoad` exercise the path guards. I read and ran these controls; I did not create additional production-code mutants.

### Closure

The earlier same-file collision is addressed by load-time comparison and pairwise and symlink tests. The missing structured evidence, unresolved mark, text-only statements, and failed-service shape are addressed by decoded contract and schema tests. No blocking result defect was found in the reviewed diff.

### Remainder

OWED — `docs/OUTPUT_FORMATS.md:5` says paths are made absolute and cleaned before symlinks are resolved. The implementation deliberately resolves components first (`internal/cli/config.go:543-552`), because cleaning `link/..` first can identify the wrong file. Correct that sentence. This is a documentation error; the alias tests guard the behavior. No behavior was declined within the stated review scope. No test or lint threshold was weakened in the diff; no numeric quality bar was declared in the order. The schema is an external check, vet a toolchain check, and the contract tests are suite checks.

### Profile

Backend and docs. New I/O errors reach exit 2 and stderr. No new runtime dependency or install script was added; `github.com/santhosh-tekuri/jsonschema/v5 v5.3.1` is imported by the schema test package. The schema file is local. No performance claim or concurrency change was made. Findings are limited to this diff and the Linux filesystem used for the test run.

### Focus

FILE · LINE · CONTROL: SARIF validity — `internal/cli/format_sarif.go:90-218` · schema tests and negative control. Evidence and text parity — `internal/cli/format_json.go:113-215`, `internal/cli/format_sarif.go:90-153` · decoded contract tests. Output writes and collisions — `internal/cli/run.go:200-274`, `internal/cli/config.go:435-629` · output path tests. JSON names — `docs/OUTPUT_FORMATS.md:7-40` · `TestJSONDocumentIsTheContract`. ADR-0027 comments and WHY notes were checked in the added code; no defect found.

### Conditions

Correct `docs/OUTPUT_FORMATS.md:5` before merge. Remove the temporary base archive from the review worktree when authorized; my `rm -rf` cleanup command was refused by the tool policy, and law #57 bars respelling that denied act. The documentation condition does not require another product-code review.

### Unrun

Manual production-code mutation and platform-specific case-insensitive filesystem tests were not run. Neither is claimed as a passed measurement. The documented case-insensitive new-file exception remains a limit of the path guard. Cleanup of the temporary base archive was refused; the worktree is therefore not yet restored to its starting state.

## The review status as filed

## REVIEW — json-sarif-outputs
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: JSON, SARIF and output collision claims pass measured tests; correct one inaccurate path-resolution sentence and remove the temporary base archive before merge.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-24T08:00:00Z

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
