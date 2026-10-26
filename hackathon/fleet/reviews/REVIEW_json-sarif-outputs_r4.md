# REVIEW — json-sarif-outputs · round 4

- verdict: reject
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 2 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-24T03:19:58Z

## Summary

BLOCKING: a directory symlink followed by .. escapes load-time collision detection; the review runs, writes one output, and exits 2 instead of refusing at load with exit 7. Fix path traversal and test the raw alias before gate.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — json-sarif-outputs

**RED — BLOCKING: a symbolic directory followed by `..` bypasses the load-time collision refusal.**

### Ground
At `3d54504` (`git rev-parse --short=7 HEAD` printed `3d54504`), reviewed the W2 changes against the recorded main ancestor `ce61e44`. `main` is not a local ref; the currently available `origin/main` is `84cf11d`, so a literal `main...3d54504` was unrun rather than silently replaced with today's remote main. Inspected the writers, load and write guards, contract tests, schema tests, and the documented field names. The JSON document carries the documented keys, evidence, and unresolved marker. The SARIF document carries the evidence and unresolved marker in result properties, and text-only statements in run properties. The vendored OASIS schema is a stronger independent check than the project-authored literal contract tests.

### Differential
BLOCKING — `internal/cli/config.go:505-514`, `internal/cli/run.go:243-255`, `internal/cli/outputs_test.go:33-68`. At load, `filepath.Abs` cleans `link/../report` before resolving `link`. The filesystem instead resolves the directory link before traversing `..`. With `link -> elsewhere/dir`, `-o elsewhere/report` and `--sarif-output link/../report` address one file, but their calculated identities differ. In a confined reproduction with both output paths absent, `cli.Run` returned `exit=2` and `stderr="-o and --sarif-output are the same file (.../link/../report); -o is left as it was written"`; the first file existed and retained JSON. The write-time `os.SameFile` check saves the first file from replacement, but only after the reviewer has run and the first output has been written. The required behavior is a load-time exit 7 with nothing written. Remedy: resolve symlinks in lexical path order before applying subsequent `..` segments (or use an equivalent filesystem-aware identity), and add a test that supplies the raw `link/../report` string, not `filepath.Join`'s pre-cleaned equivalent; assert exit 7, zero reviewer calls and no output file. The current same-file write guard should remain for changes after load.

### Unplug
`TestSampleSARIFMatchesTheSchema` removes only `version` and observes schema rejection; `TestSARIFShapesMatchTheSchema` validates empty evidence, unresolved, failed service, invalid/absent region and unknown severity. `TestJSONDocumentIsTheContract` and `TestSARIFPropertiesAreTheContract` compare decoded fields to literal maps: renaming a JSON tag or removing SARIF evidence turns them red. `TestJSONEmptyListsAreArraysNotNull` detects a nil array; `TestSARIFLevelFollowsSeverity` detects a swapped severity mapping. `TestCollidingOutputPathsAreRefused` covers all six output-flag pairs but joins its `sub/../report` path with `filepath.Join`, removing the parent segment before the CLI receives it; `TestOutputsThroughASymlinkedDirectoryAreRefused` has no `..`. Neither fires for the blocking case. The independent reproduction supplies the raw `..` and returns 2 rather than 7.

### Closure
Round 3's documented text-only fields, schema checking, level mapping, literal JSON fields and array shape are addressed by code and tests. Round 4's directory-link collision without `..` and dangling final-link cases are addressed; the general promise including a directory link followed by `..` is not addressed. The write-time guard catches this alias only after partial output, not at load. Verdict: reject until the load-time guarantee and its negative-control test hold.

### Remainder
No further finding from the inspected diff. Nonexistent-file spelling collisions on case-insensitive volumes are explicitly documented as not detected. Symlink retargeting between the final write-time stat and open is not established by this test; do not treat its absence as a measured exploit. Declined to judge licenses, `internal/review` and the agents, and unrelated work on `origin/main`, as ordered.

### Profile
Backend/output boundary and documentation. No dependency downgrade or suppressed test seen in the reviewed changes. Independent published schema check ranks above project-owned literal maps, which rank above the suite's own happy-path checks. `go version` printed `go1.26.5 linux/amd64` against `go 1.26`; `go test -count=1 ./internal/cli/...` passed both packages; `go test -count=1 ./...` passed all reporting packages (including the two CLI packages), with no reported failures; `go vet ./...` passed. These are tip results, not base-versus-tip performance metrics.

### Focus
SARIF edge shapes: schema-validated. Both formats: evidence and unresolved fields decoded and checked. All four output flag combinations: six pairs tested; the path-order edge above fails the exit-7/no-write claim. JSON documented names: literal decoded contract and documentation census present. ADR-0027: exported names and choice comments inspected in changed writers/guard; tests state the property and negative control for their covered cases. The missing `..`-after-symlink test is the material coverage gap.

### Conditions
BLOCKING: correct `outputIdentity` to preserve filesystem path traversal semantics for a symlinked directory followed by `..`; add a reproduction that fails on this tip and passes after the fix, asserting load-time exit 7 before the reviewer runs and before any output is written. Re-run the CLI and schema tests and the full affected suite. No class reclassification is proposed.

### Unrun
Baseline `go test -count=1 ./...` at `ce61e44`: not run; the confined review worktree was not switched to another commit. A literal `git diff main...3d54504` could not run because this checkout has no local `main`; `ce61e44...3d54504` was inspected instead. External SARIF service validation and a live sample scan were not run; the latter refused the state directory inside this repository, while the in-process sample and vendored-schema tests passed. Cleanup of the confined reproduction artifacts under `██████████████████████████████████████████████` was attempted once but the tool refused the deletion command (`rm -f style commands are not permitted. Use a safer approach`); by the denied-act law no alternate spelling was attempted, so the scratch artifacts remain untracked. No private Saphan Studio model/cost/pool reading was available in this worktree; none is inferred.

## The review status as filed

## REVIEW — json-sarif-outputs
- state: recommendation-ready
- verdict: reject
- ref: _REVIEW_NOTES.md
- abstract: BLOCKING: a directory symlink followed by .. escapes load-time collision detection; the review runs, writes one output, and exits 2 instead of refusing at load with exit 7. Fix path traversal and test the raw alias before gate.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-24T03:12:26Z

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
