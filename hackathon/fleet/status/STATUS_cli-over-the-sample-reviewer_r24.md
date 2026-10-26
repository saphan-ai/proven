# STATUS — cli-over-the-sample-reviewer · round 24

- state: ready-for-review
- executor: cursor on runner-cursor
- model: grok-4.7-high
- pool: not recorded
- cost: not recorded
- duration: not recorded
- last commit: c57abe40036b1ef03372060ea223f9072ee6eb53 [cli-over-the-sample-reviewer] cli: drop the dead ci workspace branch
- updated: 2026-10-23T10:09:15Z

## Summary

A bad boolean flag exits 7 and names true, false. ci with an explicit baseline and no head exits 2. W19 scenarios call cli.Run. Ready for master verification.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — cli-over-the-sample-reviewer
- state: stop-2-ready
- abstract: A bad boolean flag exits 7 and names true, false. ci with an explicit baseline and no head exits 2. W19 scenarios call cli.Run. Ready for master verification.
- last commit: c57abe40036b1ef03372060ea223f9072ee6eb53 [cli-over-the-sample-reviewer] cli: drop the dead ci workspace branch
- done: Round 13 pays both reviews of round 12. Each claim names the command that proves it.
  - W19 scenarios call cli.Run and allow only the ADR-0020 §1 names: `go test -count=1 -v ./cmd/proven -run 'TestRefutedScenarioDropsTheFinding|TestUnsafePromptStopsAndNamesCategories|TestOSVGoModuleShowsTheRangeAndTheEPSSScore|TestUnknownAuthorSaysTheDefaultReviewer'` exit 0. PASS TestRefutedScenarioDropsTheFinding 0.20s, TestUnsafePromptStopsAndNamesCategories 0.04s, TestOSVGoModuleShowsTheRangeAndTheEPSSScore 0.11s, TestUnknownAuthorSaysTheDefaultReviewer 0.06s.
  - Before the carry, `go test -count=1 -c -o /tmp/proven-scenario.test ./cmd/proven` exit 1: `cmd/proven/scenario_test.go:90:10: undefined: run`.
  - Before the boolean fix, `go test -count=1 -v ./internal/cli -run 'TestBadValuesAreRefused|TestCIMissingContextExits2'` showed `--sample` as `code 2, want 7: invalid boolean value "maybe" for -sample: strconv.ParseBool: parsing "maybe": invalid syntax`, and `explicit baseline without a head` as `exit 0 opened true stderr "" stdout "review class: \nreview rule: \nfailed services: none\n"`.
  - After the fix, the same tests pass. With the boolean mapping reverted, `TestBadValuesAreRefused/--sample` fails: `code 2, want 7: invalid boolean value "maybe" for -sample: strconv.ParseBool: parsing "maybe": invalid syntax`. Restored.
  - With the empty-head check moved back below the explicit-baseline return, `TestCIMissingContextExits2/explicit_baseline_without_a_head` fails: `exit 0 opened true stderr "" stdout "review class: \nreview rule: \nfailed services: none\n"`. Restored.
  - `TestTextCarriesOneRuleLinePerFinding` passes. With the `  rule` write removed, it fails: `"  rule discount.twice\n" count` missing. Restored.
  - `TestSampleModeReadsNoAdapterSecret` passes: decoys ███████████████████████, ███████████████, ZETARIS_ORG_ID and ████████████████████ were not read, and the default HTTP client made 0 RoundTrips.
  - `TestCIWorkspaceIsTheTarget` and `TestCIDefaultConfigIsTheWorkspaceFile` pass after the dead GITHUB_WORKSPACE branch in applyCI was removed. `TestCIMergeBaseIsTheForkPoint` passes.
  - `git log --oneline 19f292e..HEAD` is the six commits e396693, b79d301, 5329efe, b378da6, 2feea77, c57abe4.
  - `git diff --stat 90d6951..HEAD -- AGENTS.md internal/review` printed nothing.
  - `go vet ./...` exit 0, no output. `go test -count=1 ./...` exit 0: 31 packages ok, 8 with no test files, 0 FAIL. Go is go1.26.8; go.mod says `go 1.26` and has no toolchain line.
- next / blocked-on: none
- updated: 2026-10-23T10:00:44Z
- artifacts:
  - cmd/proven/scenario_test.go
  - internal/cli/config.go
  - internal/cli/config_test.go
  - internal/cli/format_test.go
  - internal/cli/run.go
  - internal/cli/run_test.go
- evidence:
  - Baseline before edits: internal/cli `go test -count=1 ./internal/cli` ok. cmd/proven did not compile, so the full suite was not run before the carry.
  - Reproduction, before the fix: scan --sample=maybe exits 2 with the strconv.ParseBool text and one dash. ci --baseline-commit=from-flag with no event and no GITHUB_SHA exits 0 and opens a review of an empty head.
  - Reverted-fix reds, then restored: the --sample row exits 2; the explicit baseline with no head exits 0.
  - Rule-line deletion red, then restored: TestTextCarriesOneRuleLinePerFinding fails on a missing `  rule discount.twice` line.
  - No Makefile and no .github workflow found. The acceptance commands are the ones above.
  - testing-bar: no metric, command or baseline is declared. Nothing to ratchet. No bar-weakening edit found. The checks run are a project check (go vet) and suite checks (go test).
  - `go test -race` UNRUN. This round adds no goroutine and no shared mutable state.
  - Per-test pass, fail and skip counts UNRUN. `go test` without -v does not print them. No package printed FAIL.
  - `saphan doc check --family exec-status _EXEC_STATUS.md` exit 0. The verb printed no RESULT line; the exit code is the verdict.
- note: master-was-wrong: —. `date` is denied for this role, so updated is the last commit's committer time (2026-10-23T10:00:44Z); the suite ran after that commit. The old subtest that expected an explicit baseline with no head to review was the defect, so it now expects exit 2; the other missing-baseline subtests are unchanged. runCIOpen supplies a head so the config-path test still opens. KnownCLIEnv is exported so the scenario test and the secret test share the ADR-0020 §1 list. remark: the network count is RoundTrips on http.DefaultTransport, so a dial on another Transport reads 0 in a quiet run and in a dialling run. remark: KnownCLIEnv does not include GITHUB_WORKSPACE, which ADR-0020 §1 does not name; resolve still reads it for ci, and the scenario tests are scan.

```owner-block-v1
version: 1
stream: cli-over-the-sample-reviewer
elements:
  - kind: none
    abstract: Nothing on your hand. This round is master-verify-only.
```

CLOSE-OK · cli-over-the-sample-reviewer · commit c57abe40036b1ef03372060ea223f9072ee6eb53 · branch fix/cli-over-the-sample-reviewer · worktree ████████████████████████████████████████████████████████████████

remark: the network count is RoundTrips on http.DefaultTransport, so a dial on another Transport reads 0 in a quiet run and in a dialling run.
remark: KnownCLIEnv does not include GITHUB_WORKSPACE, which ADR-0020 §1 does not name; resolve still reads it for ci, and the scenario tests are scan.

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| private credential or variable name | 3 |
