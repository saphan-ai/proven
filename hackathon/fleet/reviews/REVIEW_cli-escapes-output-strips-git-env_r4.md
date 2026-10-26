# REVIEW — cli-escapes-output-strips-git-env · round 4

- verdict: accept-with-conditions
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 0 BLOCKING · 1 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-24T05:37:31Z

## Summary

Terminal text and stderr escape controls; git child receives a minimal environment. Focused positive and negative checks pass. Owed: correct the JSON/SARIF comment and add a package-name integration proof. Temporary overlay cleanup was refused by the tool.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — cli-escapes-output-strips-git-env

### Ground
Reviewed `a78d7d8d951ba4b8d8af7a91e58345bacfd6e6c6` against `0b954a9` on 2026-10-24. Inspected `git diff 0b954a9...a78d7d8` and the changed CLI files; the three-dot diff also includes unrelated mainline merges, which are not findings on this wagon. Text writer: `internal/cli/format.go:55-108,141-275`; terminal errors, verbose fields and git invocation: `internal/cli/run.go:30-117,334-350`; environment allowlist: `internal/childenv/childenv.go:10-34`. No changes to `go.mod`, `go.sum`, dependency versions or the JSON/SARIF dispatch in this wagon.

### Differential
`go test -count=1 -v ./internal/cli -run 'Test(TextOutputShowsEveryPrintedFieldWithoutControlBytes|MultiLineFieldCannotForgeProvenLine|ControlContractBytes|StderrCarriesNoRawControlByteFromTheChange|GitMergeBaseChildCarriesNoKeyFromProven|CIMergeBaseIsTheForkPoint|JSONFormatHasNoWriter)'`: all seven named tests passed, including two stderr subtests. `go vet ./internal/cli`: passed. `git diff --check 0b954a9...a78d7d8 -- internal/cli`: passed. `go test -count=1 ./...` first failed only `internal/pipeline/TestMaterializedTreesAreReadableByOthers`: the seat umask was `0002` and the fixture's created files were group-writable. With `umask 0022`, that test passed and `go test -count=1 ./...` passed for every package (no failures). This is an environment-dependent test outside this wagon, not evidence that the CLI change failed. The project has no declared numeric metric/baseline/tolerance in this order; no quality-bar number is claimed. Check independence: compiler/vet and diff hygiene are project checks; the focused and full Go tests are author-controlled suite checks, so test green alone is not independent proof.

### Unplug
Temporary Go overlays inside `████████████████████████████████████████████████████████████` changed only the reviewed functions, not tracked source. Replacing `visible` with an identity function made `TestTextOutputShowsEveryPrintedFieldWithoutControlBytes`, `TestControlContractBytes` and the two `TestStderrCarriesNoRawControlByteFromTheChange` subtests fail on raw controls. Removing `continuationPrefix` after preserved newlines made `TestMultiLineFieldCannotForgeProvenLine` fail: forged status and finding lines appeared at column zero. Appending the parent's environment after `childenv.Minimal()` made `TestGitMergeBaseChildCarriesNoKeyFromProven` fail because the marker reached the child. Each false-world test executed, failed for its intended reason, and did not fail to compile. The green fork-point test uses real git; the environment test substitutes a local executable on `PATH`. A tool refusal prevented removal of the temporary overlay directory; it remains under the allowed review work path, and is declared in the return manifest rather than silently discarded.

### Closure
No BLOCKING finding in these two properties. Text fields that can carry change bytes are encoded before writing: class/rule/statements, finding ID or fallback rule/severity/title, EPSS phrases, path/message/rule, test path/name/command/output, query engine/snapshot/SQL/columns/rows, failed-service name/reason, trace family sentence and trace path (`internal/cli/format.go:55-108,141-225`). Verbose target/baseline/head and all error branches use `visible` or `printErr` (`internal/cli/run.go:36-117`). The only unencoded output is fixed program wording, numbers/booleans, help/version text, or text already prefixed and encoded. `visible` escapes C0 (including tab and newline), DEL, C1 and each invalid UTF-8 byte; `visibleLines` alone retains line breaks, with an unambiguous constant continuation prefix. `gitMergeBase` uses `childenv.Minimal()` and the closed set passed through it is `PATH` and `HOME`; the child does not receive the key variable tested here. The merge-base/fork-point test passed.

### Remainder
OWED documentation at `internal/cli/format.go:249`: the WHY comment says JSON and SARIF encoders already escape, but `FormatterFor` rejects both formats at `internal/cli/format.go:22-31`, unchanged from base. Replace the assertion with the current unsupported-format boundary; do not claim a byte-for-byte JSON/SARIF result that cannot exist. OWED proof at `internal/cli/terminal_escape_test.go:55`: the test constructs a finding carrying ESC in `Location.Path` and `Title`, but does not ingest a package name from a change into an actual printed finding. Add the requested sample-to-finding integration case, without weakening the direct field and stderr tests. `internal/cli/format.go:204` trims leading/trailing whitespace from test output before escaping; such whitespace bytes are removed, not printed as `\xNN`. This preserves the security property (escaped **or removed**), but limits the executor's stronger fidelity claim. Do not interpret the test as proof that all source bytes remain visible. A base-revision whole-suite run was not made: no numeric base metric was declared for this scoped check. `saphan doc check` validated the status; it does not have a review-notes family, so the notes' H2 order was checked against the injected grammar directly.

### Profile
Security boundary: change-controlled strings can spoof terminal status or alter display, and inherited parent environment can disclose a key to git or a child it launches. Writer-to-terminal and process-to-git boundaries were inspected; the negative controls above exercise both. `PATH` and `HOME` remain intentionally inherited, so this review does not claim isolation from tool configuration in those values or from an executable deliberately placed on `PATH`. No new dependencies, authorization decisions, network listeners, or destructive operations in the wagon. No bidi characters or new suppression/skip markers found in the five changed CLI files.

### Focus
Verdict: accept-with-conditions. P4-F5's terminal-control safety is met by the formatter and stderr path; P2-F3's named key does not reach the git child and real git still works. JSON/SARIF serialization is NOT APPLICABLE: both base and tip refuse these formats before any finding can be rendered (`FormatterFor` and `TestJSONFormatHasNoWriter`); neither produces bytes to compare. Their refusal behavior is unchanged, but byte identity for an actual finding cannot be asserted. The comment that says otherwise is OWED. Declined to judge: other subprocess adapters, live endpoint behavior, policy for Unicode format characters outside C0/DEL/C1, licences and unrelated mainline merges.

### Conditions
Before this verdict is used as evidence of the *full* order, correct the unsupported JSON/SARIF comment and add a test showing an ESC-bearing package name from the sample change in a printed finding. These are documentation and coverage debts, not a known bypass of the writer or the child environment. The master decides whether to require the integration case before the gate; any change to a review condition's class must be recorded as an upward reclassification with a reason.

### Unrun
none

## The review status as filed

## REVIEW — cli-escapes-output-strips-git-env
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: Terminal text and stderr escape controls; git child receives a minimal environment. Focused positive and negative checks pass. Owed: correct the JSON/SARIF comment and add a package-name integration proof. Temporary overlay cleanup was refused by the tool.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
    - cli-escapes-output-strips-git-env-review-codex.tar
- updated: 2026-10-24T05:31:09Z

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

     # REVIEW NOTES — cli-escapes-output-strips-git-env

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
