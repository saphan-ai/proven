# REVIEW — cli-over-the-sample-reviewer · round 7

- verdict: accept-with-conditions
- reviewer: claude-code on macm5, claude-opus-5-5
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-23T08:12:00Z

## Summary

26343c2...09a52cd holds all six properties. 10 of 10 mutants were killed. vet is clean, and internal/cli and cmd/proven are green at 09a52cd.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — cli-over-the-sample-reviewer

Reviewer: claude-code (claude-opus-5-5), round r14 (the order's own name: `cli-over-the-sample-reviewer-review-r13-claude`).
Scope: `git diff 26343c2...09a52cd`, 14 commits. All measurements below are mine, taken 2026-10-23 on this seat
(darwin/arm64, go1.26.8). The master's figures in the order were set aside as claims (R1) and re-measured where I could.

### Ground

- `git rev-parse --short=7 HEAD` printed `09a52cd` (ROLE-LOCK satisfied).
- `git log --oneline 26343c2..09a52cd`: 14 commits, 1ca207a … 09a52cd.
- `git diff --stat 26343c2 09a52cd -- AGENTS.md internal/review`: empty output. The owner file and the reviewer seam are untouched.
- `go vet ./...`: exit 0, no output.
- `go test -count=1 ./...`: every package in the diff passes — `internal/cli` ok, `cmd/proven` ok, `internal/pipeline` passes
  everything except one test (below). 6 packages FAIL, all on one cause, and none of them is in the diff's write-set.
  `go test … | grep -c 'bind: operation not permitted'` printed `6`: this seat's sandbox refuses `httptest` a localhost
  port (`cmd/livecheck`, `internal/generator`, `internal/guardrail`, `internal/pipeline` (TestUnsafePromptStopsTheStepAndNamesCategories),
  `internal/query`, `internal/review`). The only diff-touched file in those packages is `internal/pipeline/imports_test.go`
  (one exemption row). Its test, the import allow-list, passed.
- Earlier codex round (ae34b74, "repair three shell findings"): `git diff ae34b74 09a52cd -- internal/cli/run.go internal/cli/config.go`
  shows only additions on top of it: the timeout bound, ci reading `GITHUB_WORKSPACE`, exit 7 on a review `ConfigError`, and the
  ci no-baseline refusal. All three repairs are still in place after the carry: the merge base (run.go:271), only the
  winning source is parsed (config.go:439-447 and siblings), and the panic during open is recovered (run.go:130-137).
  Each one is mutation-checked under Unplug.

### Differential

Item by item against WHAT TO JUDGE:

1. **Flags/variables, flag > variable > default.** Every ADR-0020 §1 flag is defined (config.go:149-176) and listed for the
   unknown-flag refusal (config.go:123-147; `TestKnownFlagsMatchTheFlagSet` keeps the two in sync). Precedence is resolved per
   setting (config.go:354-492). `TestPrecedence` (config_test.go:66) has a default, variable and flag case for every row,
   including all three baseline variables in their order and `-o`. HOLDS.
2. **Exit codes.** 0/1 (run.go:91-95), 2 for a reviewer error or a panic in open or in review (run.go:50-59, 74-82, 130-163),
   and 7 for a `review.ConfigError` from either open or review and for the shell's own refusals (config.go:517-520). A failed
   outside service is printed by name (`failed service: …`, format.go:74-84) and does not change the exit code
   (`TestFailedServiceIsNamedAndDoesNotExit2`). HOLDS, with **F1** below.
3. **Refusals name what is allowed.** An unknown flag lists the known flags (config.go:240). A bad `PROVEN_SEVERITY` lists
   INFO, WARNING, ERROR (config.go:454), and so do the format, boolean-variable, timeout and `--config` refusals. HOLDS for
   both of the order's examples. **F1**: a bad value on a *boolean flag* does not hold.
4. **Sample mode: no environment key, no network.** In the CLI, `--sample`/`SAMPLE_MODE` always goes through `review.Open`, never
   the injected opener (run.go:150; `TestSampleFixtureExitCodes` fails the test if the opener is called). On the reviewer side,
   which is not in this diff, live settings are loaded only when `!opts.Sample` (internal/review/open.go:42) and the pipeline reads
   `p.env` only when it is not in sample mode (internal/pipeline/pipeline.go:642-645). The text output carries the
   evidence: test command and output, query SQL, columns, rows and EPSS phrases (format.go:130-209). HOLDS. **F3**: no test
   at the CLI level.
5. **ci takes baseline and head from the event.** run.go:224-277. The event's base SHA is used, then `GITHUB_BASE_REF`, then
   `git merge-base` with the head. `TestCIMergeBaseIsTheForkPoint` asserts baseline = fork point (≠ base tip) and head = the
   event's head. No baseline at all exits 2 and names all three sources (`TestCIMissingContextExits2`). HOLDS. **F4** is about
   the fixture.
6. **A test fails per claim when reverted.** See Unplug: 10 mutants, 10 killed, plus one probe that measured F1.

The **`  rule <rule-id>` line** (format.go:159-163), which the order asks me to judge: **it belongs.** It is additive. It prints
only when the rule ID differs from the finding's display name, so it never prints the same value twice. It names the stable key
a drop-in replacement for Semgrep/Opengrep reports (ADR-0016), and main's output assertions are all `Contains`, so they still
hold (`cmd/proven` ok). The diff does not ask for it, so R2 applies, but this is presentation and not a behaviour change. It
is untested (**F2**).

### Unplug

Each mutant was edited in, run with `go test -count=1 ./internal/cli` (and `./cmd/proven` where named), then reverted with
`git checkout --`. `git status --porcelain` was clean of tracked changes after each one.

| # | mutant (file:line) | killed by |
|---|---|---|
| M1 | ci no-baseline guard disabled (run.go:256) | TestCIMissingContextExits2 |
| M2 | timeout overflow bound disabled (config.go:473) | TestTimeoutDurationBound |
| M3 | merge base skipped (run.go:263 `|| true`) | TestCIMergeBaseIsTheForkPoint, TestCIBaseRefWhenTheEventHasNoSHA, TestCICommitsComeFromTheFixtureEvent, TestCIUnresolvableBaseExits2 |
| M4 | severity variable before flag (config.go:440-445) | TestPrecedence/--severity_overrides_PROVEN_SEVERITY, TestValidFlagBeatsInvalidVariable/severity |
| M5 | event head preferred over GITHUB_SHA (run.go:251) | TestCIFlagAndSHAOverrideTheEvent |
| M6 | review-time ConfigError → 2 (run.go:76) | TestReviewConfigErrorExits7/config |
| M7 | known-flag list dropped from refusal (config.go:240) | TestUnknownFlagListsKnownFlags, TestSecretIsNotAFlag |
| M8 | ci GITHUB_WORKSPACE in resolve removed (config.go:268) | TestCIDefaultConfigIsTheWorkspaceFile/default |
| M9 | EPSS phrases suppressed (format.go:142) | cmd/proven TestScanSamplePrintsF1AndF2NotF3 |
| M10 | recover in openCall removed (run.go:131) | TestOpenerPanicExits2 (panics, FAIL) |
| M11 | — probe, not a mutant: added `--sample=maybe`, `--error=maybe` to TestBadValuesAreRefused | measured `code 2, want 7` — this is F1 |

Not unplugged, because no test pins them: the `  rule` line (F2) and a sample-mode key read (F3).

### Closure

Findings, most severe first. Each one has its location, trigger and remedy (R7/R10).

- **F1 · owed, CONDITION C1 · config.go:70-77 with config.go:238-243.** Trigger: `proven scan --sample=maybe` (likewise
  `--error=`, `--no-error=`, `--quiet=`, `--verbose=`, `--text=`/`--json=`/`--sarif=` with a non-boolean value). Measured:
  exit **2**, message `invalid boolean value "maybe" for -sample: strconv.ParseBool: parsing "maybe": invalid syntax`.
  The same setting refused through its variable (`SAMPLE_MODE=maybe`, `PROVEN_ERROR=maybe`) exits **7** and lists
  `true, false` (config.go:494-499). Value flags such as `--timeout=soon` and `--severity=CRITICAL` also exit 7. So one setting
  gets two exit codes depending on its source. The boolean-flag refusal also names no allowed values, which breaks item 3 for
  these flags, and it spells the flag `-sample` instead of `--sample`. Remedy: return a `refuse("--<name>", v, "value must be
  true or false", []string{"true","false"})` from `boolPresence.Set`, or map the parse error to it in `resolve`. Then
  add the two probe rows above to `TestBadValuesAreRefused` as the negative control (law #27: the mutated field is the flag's value).
- **F2 · owed, test · format.go:159-163.** The `  rule <rule-id>` line, and the ID → RuleID fallback in the header
  (format.go:131-134), have no test. Delete them and every test still passes. Remedy: a `writeFinding` test with ID≠RuleID
  (line present), ID==RuleID (absent), and empty ID (the header carries the RuleID).
- **F3 · owed, test · run_test.go:49-81.** Item 4 is enforced only inside `internal/review`/`internal/pipeline`, which belong
  to the other train. No CLI test asserts that a sample run reads no `METERLESS_*`/`ZETARIS_*` key. Remedy: give
  `TestSampleFixtureExitCodes` a recording `getenv` that fails on any key outside `PROVEN_*`/`SAMPLE_MODE`/`GITHUB_*`.
- **F4 · note · tests/fixtures/pull_request.json.** The committed fixture is reached only on the unresolvable-merge-base path
  (TestCICommitsComeFromTheFixtureEvent asserts only the exit-2 message). Its head SHA is never asserted. The positive
  event path is proven on generated events (run_test.go:697-722), so item 5 holds.
- **F5 · note, R8 dead code · run.go:228-230.** `applyCI` sets `cfg.target = GITHUB_WORKSPACE` when it is still `.`, but
  `resolve` already did that (config.go:268-270). The branch can never change anything.

### Remainder

Declined to judge (R11), for the master to rule on:
- `--json`, `--sarif`, `--json-output`, `--sarif-output` and `PROVEN_OUTPUT_FORMAT=json|sarif` are defined but exit 2:
  `format "json" has no writer in this build` (format.go:26-27, checked before the reviewer opens, run.go:46-49). The help
  text says so (run.go:317). I did not judge whether the writers belong to this wagon.
- In GitHub Actions, `GITHUB_SHA` on a `pull_request` event is the synthetic merge commit, and the code prefers it over the
  event's head SHA (run.go:251-255; this is pinned by TestCIFlagAndSHAOverrideTheEvent). With `actions/checkout`'s default
  `fetch-depth: 1`, the event's base SHA is not in the clone, so `git merge-base` fails and ci exits 2. That is correct by §2, but
  it will meet every default checkout. This belongs to the Action wagon, and nothing here measured it.
- `--quiet` has no progress line to suppress. Today it only cancels `--verbose`, which is refused alongside it anyway.
- `writeOutputs` writes standard output before the `--*-output` files (run.go:196-207). A failing extra file then exits 2 after a
  complete report has already been printed.
- Testing bar: the project declares no numeric metric, so there is nothing to ratchet. Every check here is a *suite* check.
  The external and project checks were `go vet` only (clean). NOTE per testing-bar step 10.

### Profile

security-adjacent configuration: R9 applies, because the diff touches configuration and secrets handling. S-acts done on the
CLI boundary: secrets are not flags (`TestSecretIsNotAFlag`, which also checks the value is not echoed). The one new `os/exec`
(`git -C <target> merge-base <base> <head>`, run.go:280) is named on the import allow-list (imports_test.go). Its arguments
are passed as an argv, never through a shell. `base`/`head` come from the event, a flag or the environment, and a value
beginning with `-` would reach git as an option. That is a trust-boundary note only: the pipeline's operator controls those
values. Remedy if wanted: add `--end-of-options` before `base`. I did not run a dependency delta (R4): there is no
`go.mod`/`go.sum` change in `git diff --stat`. I did not scan for invisible characters beyond the read.

### Focus

The order's six properties, the ae34b74 repairs after the carry, the `rule` line, and round 12's two changes (ci with no
baseline exits 2; AGENTS.md restored). Nothing outside `26343c2...09a52cd` was judged, and licences are out of scope.

### Conditions

- **C1 (F1):** a bad value on a boolean flag is refused with exit 7 and the allowed values `true, false`, the same as its
  variable. Add the negative-control rows to `TestBadValuesAreRefused`.
- **C2 (Unrun):** the master confirms the three unrun items below on macm5 before the gate. The master's order already
  claims them, and I could not re-measure them here.
- F2, F3 and F5 are owed work and do not reopen a round (R6).

### Unrun

- **The binary run of docs/SAMPLE_MODE.md from outside the repository** (exit 0, 4 findings, 3 EPSS lines, 26 guardrail
  events), and the binary `--error` / `--config=auto` / `--timeout=9223372037` runs. Not run: the binary writes `logs/` into
  its working directory, and this role's permission class admits no removal program, so I could not leave the worktree as I
  found it. Partial equivalent that DID run: `cli.Run` from a `t.TempDir()` outside the repository
  (`TestSampleFixtureExitCodes`, `TestSampleFlagIgnoresInvalidSampleMode`), plus `TestRunConfigAutoExits7` and
  `TestTimeoutDurationBound`, all green. The *counts* (4 / 3 / 26) are UNMEASURED.
- **Output parity with main's own binary** on the same fixture. Not run, because it needs two built binaries; see above.
- **6 localhost-listener tests** in packages outside the diff (`bind: operation not permitted`, sandbox). The test in
  `internal/pipeline`, whose only diff-touched file is `imports_test.go`, is among them; the allow-list test in that file passed.

## The review status as filed

## REVIEW — cli-over-the-sample-reviewer
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: 26343c2...09a52cd holds all six properties. 10 of 10 mutants were killed. vet is clean, and internal/cli and cmd/proven are green at 09a52cd.
  C1: a bad value on a boolean flag (--sample=maybe) exits 2 without the allowed values, while the same setting through its variable exits 7 and lists true, false.
  C2: the binary SAMPLE_MODE run and the parity runs are UNRUN on this seat; the master confirms them.
  Owed: tests for the `rule` line (it belongs) and for sample mode reading no key; one dead GITHUB_WORKSPACE branch.
- artifacts:
    - _REVIEW_NOTES.md
    - _REVIEW_STATUS.md
- updated: 2026-10-23T08:30:00Z

## REDACTED

none
