# ORDER — the-prover-in-a-container-tests · round 1

- change: the-prover-in-a-container-tests
- round: 1
- base: main @ 0ed349b
- executor: cursor on runner-cursor
- written: 2026-10-22T02:33:05Z
- task: write the tests of the-prover-in-a-container from its test plan alone, independently of its implementation

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — the-prover-in-a-container-tests
track: proven-reviewer-core
- train: proven-reviewer-core
- wagon: W1bt
- review: master-verify-only
- repo: proven
- base: main @ 0ed349b
- branch: test/the-prover-in-a-container
- lands-on: main
- worktree: ███████████████████████████████████████████████████████████████████
- backend: cursor
- model: grok-4.7-high
- quote-usd: 2.00
- cap-usd: 4.00
- wall: 90m
- task: write the tests of the-prover-in-a-container from its test plan alone, independently of its implementation

█████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████
ROLE-LOCK: YOU = EXECUTOR, as TEST AUTHOR. One-shot headless: foreground commands only; your first act writes `_EXEC_STATUS.md`
(`- state: in-progress`). A permission denial is a fact: try once, then `- state: blocked`.
COMMIT EVERY PART BEFORE `stop-2-ready` · ZERO push/merge/PR · author = fleet identity · commit message `[the-prover-in-a-container-tests] tests: <what>`.
No AI attribution anywhere: no `Co-Authored-By`, no "Generated with", no tool or model name, no session link.
After `git commit`, `packed-refs.lock: Permission denied` means the commit LANDED — check `git log -1`.

### SKILLS
- ██████████████████████████████████
- ██████████████████████████████████████

Craft: `██████████████████████████` + `████████████████████████████` — read first.

### THE OWNER'S WORDS
2026-10-21, on ADR-0025 (the change's code runs only in a container with no network and no secret): *"OK."*

### THE TEST PLAN
- test files: `internal/testrun/container_contract_test.go`; every package-level name starts with `TestContractW1b` or `contractW1b`

### Contract
```go
package testrun
type Request struct{ FindingID, Path, Name, Command, BaseDir, ChangeDir string }
type Outcome struct{ Found bool; Path, Name, Command, Output string; FailsOnChange, PassesOnBase bool }
type ContainerConfig struct{ Runtime, Image string; Timeout time.Duration; CPUs, Memory string }
func DefaultContainerConfig() ContainerConfig
func (c ContainerConfig) Validate() error
func NewContainer(cfg ContainerConfig, lookPath func(string) (string, error)) (*Container, error)
func (c *Container) Argv(dir, command string) []string
func (c *Container) Run(ctx context.Context, req Request) (Outcome, error)
```
`service.Down{Service, Reason}` (package `internal/service`) is the error that names a failed service.
A fake runtime is any executable the test writes into `t.TempDir()` and returns from `lookPath`: it receives the argv,
can record it, and exits as the test decides. No test needs Docker except T10.

### Cases
| id | the behaviour that must hold | given | expected | negative control: the change that must turn it red |
|---|---|---|---|---|
| T1 | the container has no network | `Argv(dir, cmd)` with defaults | the argv carries `--network none` or `--network=none` | remove the network flag from Argv |
| T2 | the container gets no environment from Proven | process env holds `███████████████████████=s3cret-T2`; `Argv(dir, cmd)` | no `-e`, `--env`, `--env-file` in the argv, and no element contains `s3cret-T2` | Argv passes one `-e` |
| T3 | the container sees only the one checkout | `Argv(dir, cmd)` | exactly one `-v` or `--mount`; its source is `dir`; no other host path appears in the argv | Argv mounts the home directory as well |
| T4 | the container is one-shot and bounded | config with CPUs `1`, Memory `512m` | argv carries `--rm`, the CPU limit `1` and the memory limit `512m` | drop `--rm` |
| T5 | no runtime → the Prover runs no code and is named off | `lookPath` returns an error | `NewContainer` returns an error that `errors.As` reads as `*service.Down` with `Service == "prover"`; the fake runtime file was never executed | NewContainer returns a Container that falls back to `sh -c` |
| T6 | a test failing on the change and passing on the base is proved | fake runtime exits 0 when the mounted source is BaseDir and 1 when it is ChangeDir | `Outcome.FailsOnChange && Outcome.PassesOnBase`, `Found`; the fake ran twice | the fake exits 0 for both → the outcome is not proved |
| T7 | a run past its time limit is not proof and does not hang | `Timeout` 1s, the fake sleeps 30s | `Run` returns within 10s; the outcome is not proved, or the error names the time limit | Timeout ignored → the test exceeds its own deadline |
| T8 | settings are refused by name | `Image ""`; `Timeout 0`; then `DefaultContainerConfig()` | the first two return errors naming `Image` and `Timeout`; the defaults return nil and carry a non-empty Image and a positive Timeout | Validate returns nil always |
| T9 | the change runs in its own checkout, not the base's | fake records the mounted source per call | the two calls mount BaseDir and ChangeDir, one each | Run mounts BaseDir twice |
| T10 | a real container has no network (runs only with `PROVEN_TEST_DOCKER=1`, else `t.Skip` with that reason) | Docker present; a command that dials `███████:53` | the dial fails inside the container | — (the master runs it on runner-cursor) |

### Out of scope
- How the Prover agent and `review.Open` choose this adapter in live mode: the implementation's own tests.
- Image contents and Go module caching inside the container.

### WHAT YOU WORK FROM, AND WHAT YOU DO NOT READ
You work from the test plan and the contract at your base. You do not look for the implementation: no other branch, no other
worktree, no search of the record for this stream's code or its orders. If the behaviour already exists at your base, write
`SHOUT_<topic>.md` and stop.

### WHAT TO BUILD
1. One test per case id of the plan, named so the id can be found in it (`TestContractW1bT1…`).
2. Each test must FAIL at your base, for the reason its case names, and not for a compile error outside the contract. At
   your base the stubs return `ErrContainerNotImplemented` or zero values; T10 skips without `PROVEN_TEST_DOCKER=1`.
3. Write only in `internal/testrun/container_contract_test.go`. Fake runtimes are written into `t.TempDir()` at test time.

### ACCEPTANCE
1. Every case id of the plan has a test.
2. At the base, every one of them (except T10, which skips) fails for its case's reason. Paste each failure under `- evidence:`.

### THIS ORDER CAN BE WRONG (law #44)
The master wrote the plan without building it. If a case cannot be tested through the contract as written, write
`SHOUT_contract-<case id>.md` naming the case and why; that is a correct result, not a failure.

### WHAT IS NOT IN SCOPE
- Any change outside the one test file. A test that needs a change in product code is a SHOUT, not an edit.

### WHEN THIS ORDER CANNOT PROCEED
Write `SHOUT_<topic>.md` in the worktree, set `- state: blocked`, and stop.

### ARTIFACT CONFINEMENT
Every artifact this leg makes — captures, logs, evidence, scratch — lives under `█████████████████████████████████████████████` in the worktree.
Nothing outside the worktree; nothing committed unless this order names it.

### RETURN
`_EXEC_STATUS.md`: `- state: stop-2-ready`, `- last commit:` <sha>, an `- abstract:` of at most 5 lines, every case id with its failure
at the base. Check it with `saphan doc check --family exec-status _EXEC_STATUS.md` before you end: its exit code is the verdict.
Read every `saphan` verb by its `RESULT:` line, its `--json` or its exit code, never by its sentences.

## REDACTED

| what | times |
|---|---|
| IP address | 1 |
| absolute or private path | 2 |
| how an order is bound to the laws | 1 |
| private credential or variable name | 1 |
| private workspace path | 4 |
