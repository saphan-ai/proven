# REVIEW — cli-escapes-output-strips-git-env · round 2

- verdict: reject
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 2 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-23T21:35:18Z

## Summary

Two blocking text-output defects remain: a preserved evidence newline forges a Proven status line, and tab plus malformed UTF-8 violate the claimed escape range. Git child environment isolation and real fork-point tests pass. See the review notes for the unmeasured JSON/SARIF formats and the umask-sensitive suite rerun.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — cli-escapes-output-strips-git-env

### Ground
Review of `0b954a9...1e77598` at HEAD `1e77598` (verified with `git rev-parse --short=7 HEAD`). Diff: four files, two commits; no dependency or lockfile change. Read `internal/cli/format.go`, `internal/cli/run.go`, `internal/childenv/childenv.go`, the new tests, CLI callers and ADR-0027. The executor's and master's statements in the order were treated as claims, not evidence. No package outside the named CLI diff was changed.

### Differential
- BLOCKING — `internal/cli/format.go:203-205`, `internal/cli/format.go:239-264`: `visibleLines` leaves an attacker-controlled test-output newline intact and does not indent subsequent lines. A test-output value containing `first\nfailed services: none\nlast` prints an unprefixed `failed services: none` even when `result.FailedServices` contains `prover: unavailable`. The review-local `TestReviewProbeMultilineCannotForgeServiceVerdict` fails on the tip with `test output forged the review's service verdict`. This violates the no-forged-Proven-line property and can conceal a service failure in a human reading. Remedy: encode newline or prefix every continuation line in a way that cannot reproduce a top-level Proven line; add a negative test with an actual failed service.
- BLOCKING — `internal/cli/format.go:229-264`: `isControl` explicitly preserves tab (C0 U+0009), contrary to the requested C0 range, and Go's rune iteration replaces malformed UTF-8 bytes with U+FFFD rather than rendering their original `\xNN` values. The review-local `TestReviewProbeEscapesTabAndMalformedUTF8` fails on the tip with `unescaped tab or malformed byte`. Remedy: define the byte-level contract, escape tab if claiming all C0, and display each malformed byte rather than silently replacing it; add distinct tab and malformed-byte tests. This is a contract violation, not proof that malformed bytes themselves survive in output.
- `internal/cli/format.go:54-106`, `internal/cli/format.go:143-225`: all result fields printed by the text formatter are passed through `visible` or `visibleLines`; no raw finding title, location, rule, evidence, SQL, row, failed-service reason, trace path, or reviewer statement was found. `visibleLines(test.Output)` is the unsafe exception above. This review does not certify every diagnostic on stderr (outside the changed formatter).
- `internal/cli/run.go:329-334`: the merge-base child is assigned `childenv.Minimal()`, whose closed default set is `PATH` and `HOME` (`internal/childenv/childenv.go:12-35`). `TestCIMergeBaseIsTheForkPoint` passes with real git. No other subprocess was added by this diff. This measures direct environment assignment, not external git configuration loaded through HOME.
- `internal/cli/format.go:21-30`: JSON and SARIF have no writer at either revision. No successful finding can be rendered in these formats, so byte-identity of their output is UNMEASURED, not a pass; the diff does not touch their unimplemented branches. `TestJSONFormatHasNoWriter` passes.

### Unplug
With the base `format.go` overlaid onto the tip (no tracked change), `TestTextOutputCarriesNoControlByteFromTheChange` fails on raw ESC U+001B. With the base `run.go` overlaid onto the tip, `TestGitMergeBaseChildCarriesNoKeyFromProven` fails because the fake git receives the marker. Both tests pass on the tip; their false-world answers differ. The terminal test expressly permits any newline, and neither test covers the forged-service-line, tab or malformed-byte cases; the two review-local overlay tests fail on the tip as described above. These are review-only probes, not changes to the product.

### Closure
Verdict: **reject**. The merge-base environment fix and the control-byte test's ESC/CR/DEL/C1 coverage work, but an evidence-output newline can forge Proven's own status line and the claimed C0/malformed-byte escaping is incomplete. Recheck both failures after a fix; do not report the terminal property as closed from the existing green test. No owner gate is requested here.

### Remainder
`go test ./...` initially failed in an unchanged pipeline test because this shell's umask was `0002`: the test observed group-writable materialized `go.mod`. Under a command-local `umask 022`, `go test ./...` passed: 37 packages passed, five had no tests, zero failed. `go vet ./...` passed with no output; targeted tests for terminal escaping, merge-base child, fork point and missing JSON writer passed. These commands ran with offline module resolution and a review-local build cache. The initial failure is environment-sensitive, not evidence of a change in the four reviewed files. No threshold, baseline, suppression, dependency, skip or test assertion was changed by the diff. No numerical project bar was declared in the order, so there are no base/tip metric values or tolerance to claim. Check ranking: the real-git fork-point test is stronger than the fake-git environment observation for git behavior; both and the formatter suite are project-owned suite checks, and no independent external security check ran. No exception was supplied or granted.
Review-local scratch logs contain subprocess environment output (including host-specific paths); they are not return artifacts. Cleanup was refused by the command tool, so the master must remove them during authorized teardown; do not publish these logs.

### Profile
Security trust boundaries: change-derived strings cross from review evidence into a human terminal; `visible` handles single-line fields, but preserved evidence newlines can impersonate output. Process environment crosses into git; direct inheritance of non-allowlisted variables is prevented, subject to `PATH` and `HOME` values and git's own configuration. Diff does not touch authorization, destructive paths, dependency manifests or external network. No new invisible/bidirectional source-code character or instruction-like comment was identified in the diff. No credentials or host-specific values from verbose test logs are copied here.

### Focus
Properties assessed: (1) text fields covered except multi-line evidence output; (2) C0 and malformed-byte claims fail, and newlines can forge output; (3) git child direct environment limited to PATH/HOME and real merge-base green; (4) JSON/SARIF output equality unmeasurable because both writers are absent; (5) both new happy tests go red with their fix overlaid away, but do not detect the residual terminal failures; (6) WHY comments and property-named tests exist, but the new escaping comment does not name the decision it carries out as required by ADR-0027 point 4. This last item is owed documentation, not a separate safety proof.

### Conditions
Fix both BLOCKING findings and add tests that fail with the unsafe behavior present: a newline-forged top-level status despite an actual failed service; a tab and malformed byte in a change-derived field. Clarify the documented byte contract and add the relevant ADR reference to the escaping comment. No condition was reclassified downward.

### Unrun
JSON and SARIF byte-for-byte result comparison: neither writer exists in this build. Base full-suite and numerical bar comparison: no base checkout test was run and no numeric bar was declared. External git config and arbitrary subprocess behavior beyond this diff were not audited. Network and live service execution were outside this local review. No unsupported claims are inferred from these omissions.

## The review status as filed

## REVIEW — cli-escapes-output-strips-git-env
- state: recommendation-ready
- verdict: reject
- ref: _REVIEW_NOTES.md
- abstract: Two blocking text-output defects remain: a preserved evidence newline forges a Proven status line, and tab plus malformed UTF-8 violate the claimed escape range. Git child environment isolation and real fork-point tests pass. See the review notes for the unmeasured JSON/SARIF formats and the umask-sensitive suite rerun.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-23T21:32:21Z

## REDACTED

none
