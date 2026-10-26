# REVIEW — the-prover-in-a-container · round 1

- verdict: reject
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-22T03:34:17Z

## Summary

Codex review (exact model ID unavailable): reject; cleanup exceeds run limit and may fail silently; the executor changed owner-only AGENTS.md. Independent suite and real Docker were unrun.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — the-prover-in-a-container

### Ground

RED — BLOCKING: `AGENTS.md:57` and `AGENTS.md:70` were changed by the executor, despite the repository rule that only the owner changes this file. Revert that part of this commit and have the owner make any documentation change separately. The commit message also contains a prohibited AI co-author trailer; the master must address that before landing. These are conformance findings, not reasons to skip the code review.

Reviewed `e25d5c51d...3ec484a` (one commit, nine files). The local Go toolchain is `go1.26.5`; `go.mod` requests Go 1.26. No dependency or lockfile changes. `go test -count=1 ./...`, `go vet ./...`, and `go test -race -count=1 ./internal/testrun ./internal/pipeline ./internal/review` passed. `go test -count=1 -v ./internal/testrun/` passed the tests on this branch; it did not run the independent test author's tests. The relevant Go import ban has a single exact-path exception at `internal/pipeline/imports_test.go:61` and passed its test. No declared numerical quality baseline or metric command was found; do not infer a coverage bar from the passing suite. Check strength: external real-container test (unrun) > repository import rule (passed) > author-owned unit suite (passed); the unit suite alone does not establish container isolation.

### Differential

The master's 2026-10-22 T1–T9 and T7 observations cannot be repeated directly: `git cat-file -t 33c0dcf` returned `fatal: Not a valid object name 33c0dcf`, and no test branch is present locally. No fetch is permitted by this order. A separate 10 ms timeout reproduction using an overlay test failed after 10.02 s: `TestReviewCleanupExceedsRunTimeout` reported `10ms run held for 10.019245854s by cleanup`. This independently supports the *class* of the master's T7 reading, not its exact 1 s/10 s numbers or the independent suite's other results.

The live-path test `TestLiveModeWithoutRuntimeNamesProverOff` passed with `Sample: false` and an empty executable search path: the review completed, named the Prover off, and attached no proven test evidence. `TestSampleModeOpensNoNetwork` and `TestSampleReviewDoesNotReadTheEnvironment` passed. This verifies the existing sample checks, not a byte-for-byte comparison of all W1 output.

### Unplug

Scratch Go overlays left the reviewed source untouched. Removing `--network none` at `internal/testrun/container.go:215` made `TestArgvIsOneReadOnlyMountWithNoNetworkAndNoEnv` fail at line 121. Removing `:ro` at line 218 failed at line 133. Removing `--rm` at line 214 failed at line 120. Adding a container environment flag alongside line 215 failed at line 137. Removing the missing-runtime refusal at lines 170–179 failed `TestNewContainerMissingRuntimeIsProverDown` at line 102. These are argv and constructor controls, not a real-container proof. `TestRunReadsBothCheckoutsAndDoesNotProveATimeout` counts cleanup calls but uses a cleanup fake that returns immediately, so a slow or failed cleanup survives that test. A targeted overlay test with cleanup honoring its context produced the 10.02 s failure above.

### Closure

REJECT. The owner-only file change is prohibited. More importantly, `internal/testrun/container.go:280-286` synchronously invokes `remove` after a run timeout, and `internal/testrun/container.go:306-309` gives cleanup a separate 10 s and discards its outcome. Each side can outlive its configured limit; a failed cleanup is silently reported as a mere timeout even if a container remains. Fix the total timeout/cleanup bound and make failed removal observable as an unproved, incomplete cleanup; test a cleanup that waits for cancellation and one that fails. Re-run the independent contract suite when its object is available.

### Remainder

`internal/testrun/container.go:209-222` constructs `docker run` with a single checkout mount, read-only `:ro`, `--network none`, `--rm`, CPU and memory flags, and no `-e`/`--env` flag; `internal/testrun/container.go:319-331` executes that argv with `exec.CommandContext`, not by shelling out on the host. The shell command is inside the image. `internal/testrun/container.go:163-180` uses executable discovery and returns `service.Down` if absent; `internal/review/open.go:26-36` then chooses `testrun.Off`, whose `Run` starts nothing at `internal/testrun/container.go:186-192`. `internal/pipeline/pipeline.go:229-253` supplies the base/change checkouts only for the container adapter. The configured timeout is per side; the pair therefore has no single two-minute total limit. A missing image may cause a host-side pull, acknowledged in the source, and is not adjudicated as a no-host-network guarantee here. Declined to judge licensing and image provenance, outside this order.

### Profile

Security profile: the caller supplies the test command; container flags limit its network, mount, CPU, and memory, but the image/runtime and Docker daemon are outside this unit suite's coverage. No environment transfer flag is built. Failure of runtime discovery is named; failure of daemon operation is treated as not-run for exit 125–127, while failure of forced removal is hidden. No new dependencies, grants, data stores, or explicit secret reads are in the diff. No actual container dial or resource-limit measurement was possible.

### Focus

The key review boundary is *client timeout versus container lifetime*. `--rm` removes a container after it exits, not necessarily when its CLI client times out. `rm -f` is the only attempted forced cleanup and its error is dropped. The separate independent T7 test is unavailable locally; the reproduced overlay failure provides a falsifiable implementation-level control instead.

### Conditions

Blocking: revert the `AGENTS.md` change and correct the prohibited attribution in the existing commit under the owner's process; bound cleanup within the promised run limit or explicitly revise the contract through its owner; surface and test a failed `rm -f` so an unremoved container is not silently accepted. Run the independent contract tests at the reviewed implementation and settle each red test against the contract. If T7's fake ignores context on every call, that fake also needs correction to honor cancellation; making its `rm` return immediately is useful for a fast-path test but does not resolve the real slow-cleanup defect. No downward reclassification is assumed.

### Unrun

The test author's T1–T9/T7 suite was not run: commit `33c0dcf` is absent locally and this order disallows fetching. T10 real container network-dial test was not run: no `docker` executable is on this host's path, and the order forbids extra network. Real image availability, daemon cleanup, and image-based Go compilation remain unmeasured. Base-commit test metrics were not run because no metric or recorded baseline was declared; sample-mode parity was checked with existing tests, not exhaustive byte comparison.

## The review status as filed

## REVIEW — the-prover-in-a-container
- state: recommendation-ready
- verdict: reject
- ref: _REVIEW_NOTES.md
- abstract: Codex review (exact model ID unavailable): reject; cleanup exceeds run limit and may fail silently; the executor changed owner-only AGENTS.md. Independent suite and real Docker were unrun.
- artifacts:
    - _REVIEW_NOTES.md
    - _REVIEW_STATUS.md
- updated: 2026-10-22T03:31:08Z

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

     # REVIEW NOTES — the-prover-in-a-container

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
