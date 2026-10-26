# ORDER — the-reviewer-honours-the-request-tests · round 2

- change: the-reviewer-honours-the-request-tests
- round: 2
- base: main @ 8a356f6
- executor: cursor on runner-cursor
- written: 2026-10-22T03:24:25Z
- task: write the tests of the-reviewer-honours-the-request from its test plan alone, independently of its implementation

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — the-reviewer-honours-the-request-tests
track: proven-reviewer-core
- train: proven-reviewer-core
- wagon: W1ct
- review: master-verify-only
- repo: proven
- base: main @ 8a356f6
- branch: test/the-reviewer-honours-the-request
- lands-on: main
- worktree: ██████████████████████████████████████████████████████████████████████████
- backend: cursor
- model: grok-4.7-high
- quote-usd: 2.00
- cap-usd: 4.00
- wall: 90m
- task: write the tests of the-reviewer-honours-the-request from its test plan alone, independently of its implementation

█████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████
ROLE-LOCK: YOU = EXECUTOR, as TEST AUTHOR. One-shot headless: foreground commands only; your first act writes `_EXEC_STATUS.md`
(`- state: in-progress`). A permission denial is a fact: try once, then `- state: blocked`.
COMMIT EVERY PART BEFORE `stop-2-ready` · ZERO push/merge/PR · author = fleet identity · commit message `[the-reviewer-honours-the-request-tests] tests: <what>`.
No AI attribution anywhere: no `Co-Authored-By`, no "Generated with", no tool or model name, no session link.
After `git commit`, `packed-refs.lock: Permission denied` means the commit LANDED — check `git log -1`.

### SKILLS
- ██████████████████████████████████
- ██████████████████████████████████████

Craft: `██████████████████████████` + `████████████████████████████` — read first.

### THE OWNER'S WORDS
2026-10-21 (translated from Polish): *"Working first, pretty later."*

### THE TEST PLAN
- test files: `internal/review/request_contract_test.go`; every package-level name starts with `TestContractW1c` or `contractW1c`

### Contract
```go
package review
func Open(opts Options) (Reviewer, error)           // a *ConfigError = configuration refused on load (shell exit 7)
type Options struct{ ConfigPath string; Sample bool; LogDir string; Env func(string) string }
type Request struct{ Target, BaselineCommit, HeadCommit string; Include, Exclude []string; MinSeverity Severity }
Review(ctx, req) (Result, error)                      // non-nil error = Proven could not review (shell exit 2)
type ConfigError struct{ Field, Value string; Allowed []string; Reason string }
```
The fixture `testdata/fixture/discount` in sample mode yields F1 at `billing/discount.go` and F2 at `go.mod` (measured by
the master at main @ 7126744). A glob in `Include`/`Exclude` matches a changed file when it matches the file's slash path
or its base name (Go `path.Match`). The configuration file is YAML; its keys form a closed set; `timeout` is one of them (a
Go duration).

### Cases
| id | the behaviour that must hold | given | expected | negative control: the change that must turn it red |
|---|---|---|---|---|
| T1 | Exclude removes findings on excluded files | sample, fixture, `Exclude: ["*"]` | no findings, nil error | Review ignores Exclude |
| T2 | Include keeps only included files | sample, fixture, `Include: ["go.mod"]` | exactly F2's location `go.mod`; nothing at `billing/discount.go` | Review ignores Include |
| T3 | Exclude wins over Include | `Include: ["*"]`, `Exclude: ["discount.go"]` | only `go.mod` findings | Include checked after Exclude |
| T4 | a baseline that does not exist is a fatal error, never a silent full review | sample, fixture, `BaselineCommit: "does-not-exist"` | Review returns a non-nil error whose text names `does-not-exist` | Review ignores BaselineCommit |
| T5 | an unknown configuration key is refused on load by name | `ConfigPath` → a file `colour: red` | Open returns an error that `errors.As` reads as `*ConfigError`, `Field == "colour"`, `Allowed` non-empty and containing `timeout` | Open ignores ConfigPath |
| T6 | a bad value is refused on load by name | file `timeout: banana` | `*ConfigError` with `Field == "timeout"` | the value is not parsed |
| T7 | an explicit configuration path that does not exist is refused | `ConfigPath` → a path that does not exist | `*ConfigError` whose `Value` or `Reason` names the path | a missing file is read as defaults |
| T8 | an empty ConfigPath means the defaults | `ConfigPath: ""`, sample, fixture | Open ok; Review returns F1 and F2 | — |
| T9 | the configured timeout bounds a review whose context has no deadline | file `timeout: 1ns`; `context.Background()` | Review returns with `Partial == true`, or with no finding proved | the config timeout never applied |

### Out of scope
- Computing the changed files from a real git history (BaselineCommit/HeadCommit on a repository): the implementation's own tests.
- The Prover's container settings in the file: they arrive with W1b.

### WHAT YOU WORK FROM, AND WHAT YOU DO NOT READ
You work from the test plan and the contract at your base. You do not look for the implementation: no other branch, no other
worktree, no search of the record for this stream's code or its orders. If the behaviour already exists at your base, write
`SHOUT_<topic>.md` and stop.

### WHAT TO BUILD
1. One test per case id of the plan, named so the id can be found in it (`TestContractW1cT1…`).
2. Each test must FAIL at your base, for the reason its case names, and not for a compile error outside the contract. At
   your base the reviewer ignores the request's selection fields and the configuration path, so these tests fail on behaviour.
3. Write only in `internal/review/request_contract_test.go`. Configuration files are written into `t.TempDir()` at test time.

### ACCEPTANCE
1. Every case id of the plan has a test.
2. At the base, every one of them except T8 fails for its case's reason (T8 passes at the base: it guards the defaults). Paste each failure under `- evidence:`.

### THIS ORDER CAN BE WRONG (law #44)
The master wrote the plan without building it. If a case cannot be tested through the contract as written, write
`SHOUT_contract-<case id>.md` naming the case and why; that is a correct result, not a failure.

### WHAT IS NOT IN SCOPE
- Any change outside the one test file. A test that needs a change in product code is a SHOUT, not an edit.

### WHEN THIS ORDER CANNOT PROCEED
Write `SHOUT_<topic>.md` in the worktree, set `- state: blocked`, and stop.

### ARTIFACT CONFINEMENT
Every artifact this leg makes — captures, logs, evidence, scratch — lives under `████████████████████████████████████████████████████` in the worktree.
Nothing outside the worktree; nothing committed unless this order names it.

### RETURN
`_EXEC_STATUS.md`: `- state: stop-2-ready`, `- last commit:` <sha>, an `- abstract:` of at most 5 lines, every case id with its failure
at the base. Check it with `saphan doc check --family exec-status _EXEC_STATUS.md` before you end: its exit code is the verdict.
Read every `saphan` verb by its `RESULT:` line, its `--json` or its exit code, never by its sentences.

## REDACTED

| what | times |
|---|---|
| absolute or private path | 2 |
| how an order is bound to the laws | 1 |
| private workspace path | 4 |
