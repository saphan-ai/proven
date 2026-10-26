# REVIEW — cli-over-the-sample-reviewer · round 9

- verdict: accept-with-conditions
- reviewer: claude-code on macm5, claude-opus-5-5
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-23T12:00:18Z

## Summary

Round 13 (19f292e..c57abe4) addresses all of round 12's owed items: bool-flag exit 7, ci head required on explicit baseline, rule-line test, no-secret/no-network test, dead GITHUB_WORKSPACE branch.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — cli-over-the-sample-reviewer

### Ground
- Worktree HEAD, measured first: `git rev-parse --short=7 HEAD` printed `c57abe4`. This matches the order's reviewed SHA.
- Scope: `git diff 90d6951...c57abe4` (13 files, +3171/−230). The focus is round 13, `git log --oneline 19f292e..c57abe4`, which is 6 commits: e396693, b79d301, 5329efe, b378da6, 2feea77, c57abe4.
- Toolchain: `go.mod` declares `go 1.26`; the local toolchain is `go1.26.8 darwin/arm64`.
- Canon read: ADR-0020 (fleet/decisions/ADR-0020-proven-command-line-and-action-interface.md) §1 and §2, plus `██████████████████████████` and both provisioned skills.
- The executor's claims were set aside under R1. Every claim below was measured again on this seat.
- Header note: the order asks for the H1 `# REVIEW — cli-over-the-sample-reviewer-review-r15-claude`. The engine's injected channel grammar asks for `# REVIEW — cli-over-the-sample-reviewer`, which is the reviewed slug. Its stated reason is that law #29's merge check reads the verdict only under that slug. I followed the engine. The two sources disagree, and the master should fix the order template.

### Differential
How round 13 compares with the six round-12 conditions:
- **C1 (claude-code F1): bad value on a boolean flag → exit 7 naming `true, false`.** ADDRESSED for every non-empty value. `config.go:105-109` records the refused value. `config.go:226-237` finds which flag refused it. `config.go:305-307` turns it into an exit-7 `refuse(...)` with the double-dash spelling.
  - Probe on `cli.Run`: `scan --sample=maybe` exits 7 with `configuration refused: --sample="maybe": value must be true or false; allowed: true, false`.
  - `-sample=maybe` (one dash) gives the same result. `SAMPLE_MODE=maybe` exits 7 with the same wording.
  - A residual is open: see F1 under Conditions.
- **Owed: test for the `rule` line.** ADDRESSED. `format_test.go:15` `TestTextCarriesOneRuleLinePerFinding` covers four cases: rule id differs from the id, same id, empty id, and two findings.
- **Owed: test that sample mode reads no key.** ADDRESSED. `run_test.go:691` `TestSampleModeReadsNoAdapterSecret` sets four decoy secrets and checks every read against the ADR-0020 §1 allowlist. A recording `http.DefaultTransport` covers network calls.
- **Owed: dead GITHUB_WORKSPACE branch.** ADDRESSED.
  - `run.go` (old lines 228-230) was removed.
  - It was dead because `resolve` already sets `cfg.target = GITHUB_WORKSPACE` for `ci` whenever the variable is non-empty (`config.go:333-336`). So `cfg.target == "."` together with a non-empty workspace could only happen when the workspace was `.` itself, and then the branch changed nothing.
- **codex OWED: `ci --baseline-commit=x` with no head proceeds with exit 0.** ADDRESSED.
  - The head guard moved above the `explicitBaseline` return (`run.go:260-266`).
  - Probe: `ci --baseline-commit=x` with an empty env exits 2 with `ci: no head commit: set GITHUB_SHA, or GITHUB_EVENT_PATH with a head SHA`.
  - Tests cover the empty head (`run_test.go:838`), a valid `GITHUB_SHA` head (`:850`) and a valid event head (`:867`). The event-head case uses explicit baseline mode, as asked.
- **codex NOTE: the no-secret/no-network claim needs its own instrument.** ADDRESSED by the same test as the sample-mode item above. Its reach is under Remainder.
- **JSON/SARIF writers.** Not this wagon (W2), as the order says. Not judged.
- **W19 scenario carried onto `cli.Run`** (`cmd/proven/scenario_test.go:92-108`). The property "sample mode reads no secret" still holds.
  - The test now admits the ADR-0020 §1 variables, which return empty, and fails on any other key.
  - The old predicate failed on every key, and the loader reads §1 variables, so it could not survive the move.
- **Executor claim checked:** `git diff --stat 90d6951..c57abe4 -- AGENTS.md internal/review` is empty (measured).

### Unplug
Each mutant was applied in the worktree, the named test was run with `-count=1`, and then the file was restored with `git checkout -- <file>`.

| # | mutation (file:line) | test that went red | result |
|---|---|---|---|
| M1 | `config.go:305` `if …; ok && false` (bool-flag refusal disabled) | `TestBadValuesAreRefused/--text … /--version` (all 10 flag subtests) | killed |
| M2 | `run.go:260` `if cfg.head == "" && !explicitBaseline` (the pre-fix logic) | `TestCIMissingContextExits2/explicit_baseline_without_a_head` | killed |
| M3 | `format.go:159` `if finding.RuleID != ""` (rule line always printed) | `TestTextCarriesOneRuleLinePerFinding` | killed |
| M4 | `format.go:159` `if false` (rule line never printed) | `TestTextCarriesOneRuleLinePerFinding` | killed |
| M5 | `run.go:150` sample path also calls `getenv("███████████████")` | `TestSampleModeReadsNoAdapterSecret` ("███████████████ was read 1 times") **and** `cmd/proven TestRefutedScenarioDropsTheFinding` ("sample mode read ███████████████") | killed |
| M6 | `run.go:150` sample path also calls `http.Get("http://example.invalid/")` | `TestSampleModeReadsNoAdapterSecret` ("sample mode sent GET …", "default HTTP client calls: 1") | killed |

After the restores, `git status --porcelain internal cmd` was empty.

### Closure
- `go vet ./...`: clean.
- `go test -count=1 ./...`: 31 packages with tests. 25 `ok` and 6 `FAIL`: cmd/livecheck, internal/generator, internal/guardrail, internal/pipeline, internal/query, internal/review.
  - Every one of the 6 fails with `httptest: failed to listen on a port: listen tcp6 [::1]:0: bind: operation not permitted`. That is this seat's sandbox refusing to open a local port, not a code defect.
  - The 31 package count matches the executor's claim. I could not confirm their "0 FAIL" on this seat.
  - Of those six, the diff touches only `internal/pipeline/imports_test.go`: a 3-line exemption for `os/exec` in `internal/cli/run.go`, which is git merge-base and does not run the change.
- `go test -count=1 -race ./internal/cli ./cmd/proven`: both `ok`, no data race.
- `internal/cli` and `cmd/proven` pass on this seat.

### Remainder
remark: the no-network instrument (`recordingTransport`) sees only `http.DefaultTransport`. I grepped the non-test code: every `http.Client{}` it builds (generator/live.go:54, guardrail/live.go:85, query/live.go:36) has a nil Transport, so all of them go through the default. Its reach today is therefore wider than its own comment claims. A future adapter with its own `Transport` or a raw `net.Dial` would not be seen.
remark: the no-secret instrument sees reads made through the injected `getenv`. A grep of non-test code finds `os.Getenv` only at `cmd/proven/main.go:11`, which hands it to `cli.Run`, so no direct read escapes it today. A child process (git, the test runner) inherits the real environment, and this instrument cannot see that.
- Declared, not judged (R11, outside round 13):
  - `applyCI` prefers `GITHUB_SHA` over the event's `pull_request.head.sha` (`run.go:248-252`). On a `pull_request` trigger, `GITHUB_SHA` is the merge commit, not the PR head. This behaviour predates round 13, and the master should rule on it against ADR-0020 §1 ("reviews only the change against the merge base").
  - An unknown flag exits 2, not 7 (`config.go:300-301`). ADR-0020 §1 names the list of known flags but no exit code. This was not changed in round 13.

### Profile
- Profile: default (cli / configuration surface).
- The security acts S1–S4 do not apply: no authorization, grant or egress code is in round 13. The diff only reads env names and adds an allowlist.
- ADR-0027 readability:
  - Doc comment on the exported `KnownCLIEnv`: present (`config.go:27-33`).
  - WHY comments at the new guards: present (`config.go:92-95`, `:187-188`, `:303-304`, `run.go:261-264`).
  - Tests name their property: `TestTextCarriesOneRuleLinePerFinding` and `TestSampleModeReadsNoAdapterSecret` have doc comments, and `TestBadValuesAreRefused` names its mutated field (`config_test.go:623-624`).
  - The unexported `refusedBoolFlag` (`config.go:226`) has no comment. ADR-0027 asks for comments on exported names, so this is not a breach.

### Focus
- Property 1: holds for the round-13 surface. Every boolean flag now resolves like its variable, and the 10 flag subtests sit in `TestBadValuesAreRefused`.
- Property 2: holds.
  - Bad bool values give 7.
  - Missing ci head gives 2, with the WHY at `run.go:264`.
  - Reviewer panic gives 2 (`run.go:131-134`, `:157-160`), unchanged.
- Property 3: holds for non-empty bad values. The allowed `true, false` is printed. The empty value is open (F1).
- Property 4: the no-key / no-network half has its own instrument now, and the instrument kills both mutants (M5, M6). "The text output carries the evidence" is not changed in round 13 and was not re-judged.
- Property 5: the event head is used when `GITHUB_SHA` is absent (`run_test.go:867`). The precedence is under Remainder.
- Property 6: each round-13 claim's test fails when its fix is reverted (M1–M6).
- Property 7: holds (see Profile).

### Conditions
- **F1 (owed, non-blocking; R6):** an empty value on a boolean flag still exits 2, with the flag package's one-dash wording.
  - File and line: `internal/cli/config.go:232`. `refusedBoolFlag` uses `b.bad != ""` as its "was refused" signal, and `Set("")` fails `strconv.ParseBool` while leaving `bad == ""`.
  - Trigger, measured through `cli.Run`: `scan --sample=` exits 2 with `invalid boolean value "" for -sample: strconv.ParseBool: parsing "": invalid syntax`. `scan --error=` behaves the same.
  - This is the C1 refusal path, but the empty spelling escapes it.
  - Remedy: record the refusal with its own field (`refused bool`) set in `Set` on error, test it in `refusedBoolFlag`, and add a `--sample=` row to `TestBadValuesAreRefused`.
  - Note: `SAMPLE_MODE=""` is "unset" by design, so the variable and the flag behave differently only because an explicit empty flag value is a bad value.
- **F2 (owed, non-blocking; docs):** the help text describes ci's exit-2 cases (`run.go:309-313`) but not the new one, "no head commit". Remedy: add one sentence naming GITHUB_SHA / GITHUB_EVENT_PATH head.
- **N1 (judgement, R6, no remedy owed this wagon):** `KnownCLIEnv` is an exported production symbol that only tests call (`run_test.go:714`, `cmd/proven/scenario_test.go:102`). It is a second, hand-kept list of the variables `resolve` reads. It drifts in the loud direction, because a new read makes the scenario test fail, so it is acceptable. A test-only helper would be the cleaner home.

### Unrun
- `gofmt -l` was not run: gofmt is not in this review class's command allowlist (██████████████████████████). Not rebuilt from other commands (R15, law #57).
- The full `./...` suite could not run green on this seat. Six packages need a local listener, which the sandbox refuses (`bind: operation not permitted`). They are listed under Closure.
- The built binary from outside the repository was not run. Running the binary with an exit-code capture was refused by the hook's construct rule, so the same cases were measured through `cli.Run` by a probe under `██████████████████████████████████████████████████████████████████`.
- Confinement: the order wants the worktree left as found, and `rm` is not in this class's allowlist. So the untracked `████████` directory I created, holding the probe and a built binary, is still there. The master must remove it.

## The review status as filed

## REVIEW — cli-over-the-sample-reviewer
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: Round 13 (19f292e..c57abe4) addresses all of round 12's owed items: bool-flag exit 7, ci head required on explicit baseline, rule-line test, no-secret/no-network test, dead GITHUB_WORKSPACE branch.
  Six mutants (M1–M6) were each killed by a named test. go vet is clean; internal/cli and cmd/proven pass, also under -race.
  Owed, non-blocking: F1, `--sample=` (empty value) still exits 2 (config.go:232, `bad != ""` sentinel); F2, help text omits the ci no-head exit 2.
  Unrun: gofmt (not in class), 6 listener-bound packages (sandbox bind denied); untracked ████████ probe dir left for the master to remove.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-23T12:30:00Z

## REDACTED

| what | times |
|---|---|
| absolute or private path | 3 |
| policy binding name | 1 |
| private credential or variable name | 3 |
| private workspace path | 1 |
