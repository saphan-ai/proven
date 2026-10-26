# REVIEW — a-fix-is-proved-by-the-same-test · round 1

- verdict: accept-with-conditions
- reviewer: claude-code on macm5, claude-opus-5-5
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-24T05:51:17Z

## Summary

Reviewed on claude-opus-5-5, alone. The NVIDIA/Nightshift half is still owed. Build and vet are green; tests are green except 6 packages where the sandbox refused a local listener (listed under Unrun).

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — a-fix-is-proved-by-the-same-test

### Ground
- Reviewer model: claude-opus-5-5 (claude-code). Ran alone. The second, independent NVIDIA-through-Nightshift review the owner's rule (2026-10-22) asks for is still OWED: the order records that route as down today.
- HEAD checked: `git rev-parse --short=7 HEAD` gave `446d6ac`, the reviewed SHA. Diff read three-dot: `fcbcb79...446d6ac`.
- Scope judged: the W42 leg commits (65b2f5b, 723d85e, 0eff8e4, 7689024, c7978e7, e3fd18d, 97676c5, f7e5ea0, 2954467) and the absorb 446d6ac. The nvidia-generator files that arrived through the absorbed main merges (fa2a0ee, c981d1c, 1a45a80, d57f1bd, ab00024, 286a299) are in the three-dot range but belong to another stream. They were built and tested here; they were not judged.
- `go build ./...`: exit 0. `go vet ./...`: exit 0.
- `go test -count=1 ./...`: FAIL in 6 packages. Every failure is the sandbox refusing a listener (`httptest: failed to listen on a port: listen tcp6 [::1]:0: bind: operation not permitted`); see Unrun. All other packages pass, including internal/agent/fixer, internal/testrun, internal/osv and cmd/proven (which holds fix_scenario_test.go).
- The panic in internal/pipeline stopped that package's run partway, so the six new pipeline fix tests were run again on their own: `go test -count=1 -v -run "TestEachProvedFindingGetsItsFixAttemptsInOrder|TestFixesOffAsksNeitherTheModelNorTheRunner|TestFixAttemptsBoundsTheModelCalls|TestADroppedFindingGetsNoFix|TestAFixIsNeverWrittenToTheChange|TestAFixIsRedactedLikeTheRestOfTheFinding" ./internal/pipeline/`. All 6 PASS.
- Probes ran through `go test -overlay` with their files in the run's artifact directory, so no file in the worktree changed.

### Differential
Each question in the order, with file:line and the control that answers it:
1. **Can a patch reach the reviewed tree, a commit, or the host?** The patch is applied only by `stagePatched` (internal/testrun/fix.go:208), which writes into `os.MkdirTemp` after `copyTree` and goes through `writeInside` (internal/testrun/written.go:185), a prefix-checked write. `RefuseEdits` and `Sample.RunFix` only read (fix.go:98, sample.go RunFix). Nothing in the diff runs git. The patched copy is mounted through the same `sideExit`/`argv` as the Prover's Run (container.go:603), so it gets the same no-network, no-secret container. Test: `TestAFixIsNeverWrittenToTheChange` PASS. Probe: after a patched staging, the file the patch targeted outside the tree was unchanged (`unchanged=true`). **Writes cannot escape.** Reads can, in the exported API only; see Remainder R-1.
2. **Can a patch be shown as proved without all four runs?** `judge` (internal/agent/fixer/fixer.go:228) needs base=passed, change=failed, change+patch=passed and package=passed, and a missing run fails it (nil checks). A run that did not answer (`could not run` or `timed out`) is not `passed`, so the patch fails; `TestEveryOtherRunOfTheProofIsNeeded` covers both cases. Not-proved records keep no Edits (fixer.go:262). `toFix` (review/open.go:326) and `writeFix` (cli/format.go) print edits only for Proved. **BUT every run's `passed` comes from the exit code alone**, and a patch can make the test binary exit 0 without running the proving test. See Conditions C1. This is the blocking finding.
3. **Is the model's patch treated as data? Does every call pass the guardrail?** Edits are exact text: `planEdits` (fix.go:107) needs a non-empty Old that appears exactly once, and refuses absolute paths, `..`, segments outside the allowed pattern, `_test.go` files, more than 8 edits and patches over the byte limit. A patch is refused whole, never half applied. An answer carrying a command is refused whole (fixer.go:197). The Fixer gets the guarded `gen` (pipeline.go:496 → :722 → :1063). `guardrail.Wrap` wraps every agent with no agent filter (internal/guardrail/wrap.go:17). That holds structurally, but no test asserts it for the agent "fixer"; see Conditions C2.
4. **Is the dependency fix's query template-built, with values format-checked?** No new query is added. `SmallestFix` (internal/osv/fix.go) reads only the rows the existing osv-by-package template already returned (check.go:226). It compares versions in Go with semver or pep440, and names nothing for an ecosystem it cannot order. No model and no SQL take part. `TestSmallestFixSkipsABoundAnotherAdvisoryStillHolds` and the others PASS. Holds.
5. Build, vet and test: see Ground and Unrun.

### Unplug
- Master's control (law #44), not repeated: they report that turning off the package check made `TestAPatchThatBreaksAnotherTestIsNotShownAsAFix` red. I did not run that mutant again.
- My mutant M1: I removed the "fails on the change when run again" check from `judge` (overlay file `probe/fixer_m1.go`). Running `go test -count=1 -overlay …/overlay_m1.json ./internal/agent/fixer/` gave **FAIL** in `TestEveryOtherRunOfTheProofIsNeeded/change_passes_when_run_again`, with a Proved:true fix over a change run that passed. The mutant was killed, so that guard has a test.
- Exit-0 probe (it underlies C1): in a throwaway module, a non-test file holds `func init() { os.Exit(0) }` and the test file holds a test that always fails. `go -C …/exit0 test -count=1 -run '^TestPrice$' ./p` printed `ok exit0/p`. The same applies to `go test ./p`, the package run. The test never ran, and an exit-code judge would read both runs as `passed`.

### Closure
- The patched tree is written only inside a scratch dir: holds (fix.go:208, written.go:185).
- The patched runs go in the Prover's container through the same argv builder as the Prover: holds (container.go:603). No real container run was measured; see Unrun.
- A failed patch is retried up to the setting, then "fix not proved" with no patch shown: holds (fixer.go:125-166, config.go fix_attempts 1..5; TestTheAttemptsSettingBoundsTheModelCalls and TestOneAttemptSettingEndsTheFailedFixAtOnce PASS).
- Shown only when the same test passes: **does NOT hold against an exit-0 patch** (C1).
- Every Fixer call passes the guardrail: holds structurally, with no test (C2).
- The dependency fix is template-only, the smallest exit is computed in Go, and its row is shown: holds.

### Remainder
- R-1 (owed, not blocking). `readRegular` (internal/testrun/fix.go:170) says "A link is refused" but runs `Lstat` only on the last path component, so a symlinked **parent directory** is followed. Overlay probe `probe/zz_review_probe_test.go` (change/lib → outside dir): `ReadEditable` returned the outside file's text, `RefuseEdits` returned "", and `stagePatched` wrote the edited outside text into the scratch copy. The outside file itself stayed unchanged. Through the pipeline I believe this cannot be reached: git mode leaves symlinks out (pipeline/trees.go:87-89, change/linked_parent_test.go), and directory mode refuses any symlink in `change.Read` (change/read.go:77) through `changedFiles`. I did not reproduce that pipeline path end to end. `Container.RunFix`, `ReadEditable` and `RefuseEdits` are exported, though, and their comments promise more than they check. Remedy: resolve each parent segment (`Lstat` every prefix, or `filepath.EvalSymlinks` and check the result is under `dir`), and add a negative-control test in which only a parent is a link.
- R-2 (judgement, order-mandated scope). The package run covers only the proving test's package. A patch may edit a file in another package, such as a package the test imports, and break that package's tests unseen. ADR-0042 asks for "the package's other tests", so this is labelled order-mandated; the owner decides.
- R-3 (judgement). `fixEvents` puts the model's `What` into the trace before redaction, as other agents' text already goes. Redaction of the finding is covered by `TestAFixIsRedactedLikeTheRestOfTheFinding`; whether the trace path is redacted was not traced here.
- Declined to judge (R11): the nvidia-generator files that came in through absorbed main; the licence files (out of scope by order); the docs prose in docs/SAMPLE_MODE.md, LAB.md and ARCHITECTURE.md beyond a skim.

### Profile
security, because the diff adds a path where model-written text changes files and runs code inside a confinement. S1 boundaries: model answer → edit planner (data validation holds); edit → scratch tree (writes confined; parent-link reads see R-1); scratch tree → container (same no-network, no-secret argv as the Prover); runner record → judge (**exit-code trust, C1**). S4: no new dependencies in go.mod and no install scripts. Secrets: the new `Content` field is redacted (pipeline.go redactFindings). Invisible or bidirectional characters were not scanned by tool; nothing was seen while reading.

### Focus
The property as worded in the order: "a patch is shown as a fix only when the proving test passes on the base, fails on the change, passes on the change with the patch, and the package's other tests pass". It holds for every result except one: a run that exits 0 without running the test. The patch controls that result, and the model writing the patch has the attacker-written change in its prompt.

### Conditions
- **C1 (blocking, must land before merge).** A run's `passed` must mean the named test RAN and passed, not only that the process exited 0. Location: internal/testrun/argv.go:69 (`go test -run ^Name$ pkg`), internal/testrun/fix.go:331-341 (`resultOf(kind)` from the exit code), internal/agent/fixer/fixer.go:244-255 (judge trusts `Result`). Trigger: a patch adding `func init() { os.Exit(0) }` (or an equivalent early exit) to any non-test file of the package. Both the change+patch run and the package run exit 0, and the review prints "fix proved". The exit-0 probe in Unplug measures this. Remedy, one option: run the proof with `-v` (or `-json`) and record `passed` only when the output carries `--- PASS: <Name>` for the proving test, and an `ok` line or `PASS` for the package run. Also refuse "no tests to run". Proof owed: a test feeding RunFix a run that exits 0 with output that lacks the PASS line, which must fail before the remedy and pass after, plus a negative control that removes only that check.
- **C2 (owed, not blocking).** Add a test showing the guardrail refuses a "fixer" prompt or answer and that the Fixer then shows no patch, so a later re-wiring of `gen` in `p.fix` turns a test red. Location: internal/pipeline/fix_test.go (none today; pipeline/guardrail_test.go has no "fixer").
- **C3 (owed, not blocking).** R-1: check symlinks on parent segments in `readRegular`, or correct the comment's claim.
- **C4 (owed by the owner's rule).** The independent NVIDIA-through-Nightshift review has not run.

### Unrun
- Real container run of the proof (base, change, change+patch, package) with docker or podman: UNRUN. No runtime was used in this sandbox; the master also records this as unrun.
- Packages whose tests need a local listener, refused by the sandbox with `httptest: failed to listen on a port: listen tcp6 [::1]:0: bind: operation not permitted`: cmd/livecheck (TestAllServicesReady), internal/generator (TestLiveReadsTheProvisionedStream), internal/guardrail (TestServerPostsTheMeasuredRequest), internal/pipeline (TestUnsafePromptStopsTheStepAndNamesCategories; the fix tests were run again on their own and PASS), internal/query (TestLiveSelectReadsRowsFromTheSQLEditor), internal/review (TestLiveReviewCarriesTheRoutedModel). The panic stops each package's run, so the tests after it in those packages also did not run here.
- R-1 through the full pipeline in directory mode: not reproduced end to end.
- Mutation of the package-run check: not re-run (the master's measurement is quoted, not repeated).

## The review status as filed

## REVIEW — a-fix-is-proved-by-the-same-test
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: Reviewed on claude-opus-5-5, alone. The NVIDIA/Nightshift half is still owed. Build and vet are green; tests are green except 6 packages where the sandbox refused a local listener (listed under Unrun).
  BLOCKING C1: a fix's runs count as passed on exit code 0 alone. A patch adding init(){os.Exit(0)} to a non-test file is shown as "fix proved" although the test never ran (measured: go test printed ok).
  Holds: patches are written only in a scratch copy and never committed; four runs, judged strictly; edits are exact text; the Fixer uses the guarded generator; the dependency fix is template-only.
  Owed: C2, a guardrail test for the fixer; C3, readRegular follows a symlinked parent directory (exported API only).
- artifacts:
    - _REVIEW_NOTES.md
    - _REVIEW_STATUS.md
- updated: 2026-10-24T13:20:00Z

## REDACTED

none
