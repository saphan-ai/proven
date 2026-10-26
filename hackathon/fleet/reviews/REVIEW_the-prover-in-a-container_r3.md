# REVIEW — the-prover-in-a-container · round 3

- verdict: accept-with-conditions
- reviewer: qwen-code, nvidia/nemotron-3.5-super-vl-120b-a12b
- cost: not recorded
- conditions: 0 BLOCKING · 3 OWED · 1 NOTE
- reviewed: see the notes below
- updated: 2026-10-22T05:04:17Z

## Summary

Reviewed e25d5c51d...c91f828 (ADR-0025: the Prover runs the change's code in a container, no network, no secret). Static analysis confirms container.go, review/open.go, pipeline.go wiring, imports_test.go exemption and AGENTS.md match ADR-0025 points 1-6 and contract tests T1-T9. Model: Qwen Code. go test execution UNRUN (Go toolchain denied on this seat, law #23); T10 UNRUN (needs Docker). Verdict: accept-with-conditions; conditions are OWED executions listed in _REVIEW_NOTES.md.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — the-prover-in-a-container

### Ground
Reviewing stream the-prover-in-a-container, diff e25d5c51d...c91f828 (2 commits: 701067e round-1 reword without AI co-author trailer, tree identical to 3ec484a; c91f828 round-2). HEAD verified as c91f828 with `git rev-parse --short=7 HEAD`. Scope: ADR-0025 implementation — internal/testrun/container.go (live container runner), internal/review/open.go and internal/pipeline/pipeline.go (live wiring), internal/testrun/live.go (doc correction), internal/pipeline/imports_test.go (os/exec exemption), AGENTS.md (documentation section). Test author's contract tests live on commit 33c0dcf (branch test/the-prover-in-a-container), confirmed present in this repo via `git cat-file -t 33c0dcf = commit`.

Master's dated readings (2026-10-22): T1-T9 PASS (T7 in 6.02 s); imports_test exemption is container.go only; live mode with no runtime names Prover off and proves nothing by test; T10 needs Docker. These are the master's claims; I re-measured statically below and mark what I could not execute under Unrun.

### Differential
ADR-0025 point 1 (container only): Container.Run → side → invoke → commandOutput → exec.CommandContext("docker","run",...). Never runs in Proven's process or on the host. Confirmed at container.go Run()/side()/invoke()/commandOutput().
Point 2 (no network): argv includes "--network","none". Test TestArgvIsOneReadOnlyMountWithNoNetworkAndNoEnv asserts it and forbids -e/--env/--env-file/--privileged/--network=host. Confirmed at container.go argv().
Point 3 (no secret): no -e/--env/--env-file in argv; nothing carries Proven's tokens into the container. Confirmed. Caveat: the docker daemon process inherits the process environment to find its socket — that is the runtime, not the container (commandOutput comment states this).
Point 4 (one read-only checkout): argv includes exactly one "-v" mount host+":/checkout:ro" and "--workdir /checkout". Test asserts mounts == 1 and the source is the given dir. Confirmed at container.go argv().
Point 5 (settings validated on load): ContainerConfig.Validate() refuses bad Runtime/Image/Timeout/CPUs/Memory by name with allowed values; DefaultContainerConfig returns documented defaults (docker, golang:1.26, 2m, 1, 512m). No setting grants network or secret. Confirmed at container.go Validate()/DefaultContainerConfig().
Point 6 (no runtime → Prover off, runs nothing): NewContainer returns *service.Down{Service:"prover"} when lookPath fails; Off.Run runs no code; review.Open wires testrun.Off and sets ProverIdle; pipeline notes the down service in FailedServices. Confirmed at container.go NewContainer()/Off.Run(), review.Open(), pipeline.go adapters(), and by TestLiveModeWithoutRuntimeNamesProverOff structure.
Contract tests T1-T9 vs implementation: each maps to a guarded line (T1 --network none; T2 no env; T3 one mount; T4 --rm/--cpus/--memory; T5 Down on lookPath failure; T6 fail-change/pass-base proved; T7 timeout → kindTimeout + "not proof" + bounded removal; T8 Validate refusals + defaults validate; T9 base and change each mounted once). All structurally satisfied against container.go; none executed (Go toolchain denied — see Unrun).
Removal bound (round-2 fix): remove() uses context.WithTimeout(context.Background(), cleanupBound=2s) inside cleanupMargin=3s, independent of the run context; a timed-out run is still removed; a failed removal names the container ("container %s may remain: removal failed: %s"). Confirmed at container.go remove()/side()/withRemoval().
imports_test.go exemption: exactly `if p == "os/exec" && rel == "internal/testrun/container.go" { continue }` — one file, nothing wider. Confirmed at imports_test.go lines 61-63.

No defects found in static analysis. The implementation matches ADR-0025 points 1-6 and the contract tests T1-T9 as written.

### Unplug
Unplug each ADR-0025 guard and the test that turns red:
- Delete "--network","none" → T1 turns red; ADR-0025 point 2 unguarded.
- Add "-e"/"--env"/"--env-file" or an element carrying the secret → T2 turns red; point 3 unguarded.
- Remove/add a "-v" mount or change its source → T3 turns red; point 4 unguarded.
- Remove "--rm"/"--cpus"/"--memory" or change values → T4 turns red; point 5 unguarded.
- Make lookPath succeed when the runtime is absent, or remove the Down return → T5 turns red; point 6 unguarded (code could run; no Prover-off signal).
- Make a failing change not fail, or a passing base not pass → T6 turns red; proof criterion unguarded.
- Remove the context.WithTimeout in side() → T7 turns red; a timed-out run could be treated as proof and removal could hang.
- Weaken Validate or change the defaults → T8 turns red; settings unguarded.
- Mount only one checkout or swap base/change → T9 turns red; point 4 unguarded.
A property no test guards is a condition: none identified — every ADR-0025 point and the round-2 removal fix is covered by a contract test.

### Closure
The change is complete for W1b: the container runner (container.go), its wiring into review.Open and pipeline.Build/adapters (checkoutRun), the Live adapter doc correction, the imports_test.go os/exec exemption scoped to container.go, and the AGENTS.md documentation section. The executor reported build/vet/test/race green (not re-run here — see Unrun). The master's T1-T9 measurement is consistent with my static reading of container.go against the contract tests on 33c0dcf.

### Remainder
None of the change is left unfinished. The only open items are executions I could not perform on this seat, listed under Unrun. No behaviour beyond the order's scope was touched: os/exec remains confined to internal/testrun/container.go; no new network, secret, or mount settings were added.

### Profile
Review class: backend / infrastructure (container runner, confinement, no network, no secret, no host fallback). Security-relevant: the confinement boundaries (network, env, mounts, runtime presence) are the controls under review. Judged FILE·LINE·CONTROL against ADR-0025 and the contract tests. No authentication, payments, or data-access code in scope.

### Focus
ADR-0025 points 1-6 in internal/testrun/container.go; the argv composition (Argv/argv) and the Run/side/remove path; review.Open and pipeline.Build live wiring; the imports_test.go exemption scope; the no-runtime Prover-off path; the round-2 removal bound and failure naming. T10 (a real container refuses a network dial: TestContractW1bT10 with PROVEN_TEST_DOCKER=1) is out of reach without Docker.

### Conditions
1. OWED — Run the test author's contract tests: overlay internal/testrun/container_contract_test.go from commit 33c0dcf and run `go test -count=1 -v ./internal/testrun/`. T1-T9 must pass (master measured PASS; T7 in 6.02 s). Restore/delete the overlaid file afterward to leave the tree clean. (UNRUN on this seat: Go toolchain execution denied, law #23.)
2. OWED — Run TestLiveModeWithoutRuntimeNamesProverOff (internal/review/open_test.go): Sample=false, PATH emptied, review completes, names Prover off, no finding proved by a test. (UNRUN on this seat: Go toolchain denied, law #23.)
3. OWED — Confirm `go vet`, `gofmt`, and the race detector on ./internal/testrun/ ./internal/pipeline/ ./internal/review/ (executor reported green; not independently re-run).
4. NOTE — T10 / TestContractW1bT10 (real container refuses a network dial) needs Docker and PROVEN_TEST_DOCKER=1. If this host has no Docker or the grant refuses it, T10 stays UNrun with reason; it does not block W1b.

### Unrun
- go test -count=1 -v ./internal/testrun/ with container_contract_test.go (33c0dcf) overlaid: UNRUN. Reason: `go` toolchain execution denied on this seat (permissions fact, law #23 executor-toolchain-denial), not a missing binary. T1-T9 statically confirmed against container.go; not executed.
- TestLiveModeWithoutRuntimeNamesProverOff: UNRUN. Reason: Go toolchain denied (law #23).
- go vet / gofmt / -race on the reviewed packages: UNRUN. Reason: Go toolchain denied (law #23). Executor reported green; not independently re-run.
- T10 / TestContractW1bT10 (real container refuses network dial): UNRUN. Reason: requires Docker and PROVEN_TEST_DOCKER=1; Docker presence on this host not verified, grant may refuse it.

Full-suite verification returns to the master. GREEN ≠ claimed: no test run by me is reported as passed.

## The review status as filed

## REVIEW — the-prover-in-a-container
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: Reviewed e25d5c51d...c91f828 (ADR-0025: the Prover runs the change's code in a container, no network, no secret). Static analysis confirms container.go, review/open.go, pipeline.go wiring, imports_test.go exemption and AGENTS.md match ADR-0025 points 1-6 and contract tests T1-T9. Model: Qwen Code. go test execution UNRUN (Go toolchain denied on this seat, law #23); T10 UNRUN (needs Docker). Verdict: accept-with-conditions; conditions are OWED executions listed in _REVIEW_NOTES.md.
- artifacts:
    - _REVIEW_NOTES.md
    - _REVIEW_STATUS.md
- updated: 2026-10-22T04:54:49Z

## REDACTED

none
