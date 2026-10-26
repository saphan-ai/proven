# REVIEW — the-prover-in-a-container · round 2

- verdict: reject
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 1 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-22T04:51:19Z

## Summary

Reject: failed container removal can again become silent when timed-out output exceeds 8192 bytes. T1–T9 pass; T10 unrun without Docker. Review performed by OpenAI Codex (model identifier not exposed).

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — the-prover-in-a-container

### Ground
Reviewed e25d5c51d...c91f8282d456402e9defc54526618e914f31be82 (2026-10-22), the two commits and their callers. Re-ran `go test -count=1 -v ./internal/testrun/` with the independent test author's 33c0dcf file temporarily overlaid, then removed that overlay; T1–T9 PASS, T10 SKIP. Re-ran `go test -count=1 ./internal/review ./internal/pipeline`; both packages PASS. Evidence: `████████████████████████████████████████████████████████████` and `wiring.txt` beside it. `command -v docker` returned no path. The reviewed worktree's tracked files are unchanged.
remark: The simulated runtime tests measure argv and timeout handling, not Docker's actual isolation or image behavior.

### Differential
The master's 2026-10-22 T1–T9 claim holds on this host; T7 took 6.04 s here rather than 6.02 s. The no-runtime live test and sample test pass. Round-2 commit c91f828 bounds simulated cleanup at two seconds and reports small failed-removal output; the commit no longer carries an AI co-author trailer. The wider claim that a failed removal always names the residual container does not hold for long runtime output (Closure).

### Unplug
`internal/testrun/container.go:224` supplies `--network none`: removing this pair is caught by `TestArgvIsOneReadOnlyMountWithNoNetworkAndNoEnv` (its `assertPair`). `container.go:223` supplies `--rm`: removing it is caught by that test's `assertFlag`. `container.go:227` supplies the one `:ro` checkout volume: removing or changing it is caught by that test and independent T4/T9. `container.go:225-226` carries CPU and memory: their omission is caught by `TestArgvUsesTheConfiguredLimits` and T2/T3. No environment flags are appended in `container.go:218-231`; adding an env flag is caught by `TestArgvIsOneReadOnlyMountWithNoNetworkAndNoEnv`, but there is no single line to delete to unplug an absence. `container.go:174-188` reports missing runtime as `service.Down`, and `review/open.go:25-36` selects `testrun.Off`; `TestLiveModeWithoutRuntimeNamesProverOff` and T1 guard this route. `container.go:286-293` bounds a timed-out run and requests removal; T7 and `TestRunReturnsWithinTenSecondsWhenTheRuntimeHangs` cover their observable behavior. These are test assertions and code-path inspections, NOT executed deletion mutants: the patch command was unavailable in this seat. A direct actual-Docker network negative control is T10, unrun here.

### Closure
BLOCKING — `internal/testrun/container.go:328-335,376-385,403-409`: timeout cleanup appends `container <name> may remain` AFTER the runtime's output, but `formatSides` sends that string through `trimOutput`, which keeps only its first 8192 bytes. If the timed-out runtime emits more than 8192 bytes before `rm -f` fails, the reported outcome loses the removal failure and its container name. `internal/testrun/container_test.go:272-305` checks only a short `still running` response, so it remains green. The prior review's silent-cleanup defect is therefore NOT ADDRESSED for ordinary large test output. Remedy: preserve cleanup status separately from truncated stdout (or reserve output budget for it), with a long-output failed-removal regression on both sides. This is a code-path proof, not a run-time reproducer; the false-world output is a truncated normal timeout with no `may remain`, indistinguishable from successful removal.

ADDRESSED within tested bounds — `internal/testrun/container.go:318-325` bounds `rm -f` to two seconds and includes the name on failure; T7 and `TestRunNamesTheContainerWhenRemovalFails` pass on short output. Commit 701067e carries no co-author trailer. No unrelated behavior accepted by this verdict.

### Remainder
`internal/testrun/container.go:20-26` correctly discloses that a missing image may be pulled on the host; `--network none` does not measure host egress. Runtime-present image execution and actual network refusal remain unmeasured. The suite has no test that proves a timed-out invocation leaves no container when removal itself fails; reporting its name is the specified fallback. No dependency manifest or lockfile changed. The added `os/exec` exemption in `internal/pipeline/imports_test.go:61-63` matches exactly `internal/testrun/container.go`, not an entire package. Declined to judge: licences (excluded by order), host image-pull policy, real Docker isolation, and behavior beyond the reviewed diff.

### Profile
Security: the untrusted generated command in `internal/agent/prover/prover.go:30-54` reaches `sh -c` at `internal/testrun/container.go:230` through `pipeline.go:231-250`; there is no host command fallback in this path. Host paths derive from fixture root; the adapter accepts explicit request directories if supplied, so the one-checkout guarantee depends on this caller never populating them. No new dependency or secret-setting field. Check ranking: published Docker manual claims (external but not a run), project import guard (project), simulated container and live-off tests (suite). No declared numerical test bar was located in this diff; no baseline inferred.

### Focus
ADR-0025 points 1–6: independent container invocation per side at `container.go:261-267,284-288`; no in-container network at 224; no `-e`/`--env` and no config field for secrets at 47-56,218-231; one read-only checkout mount at 227; normal removal flag at 223 and timeout removal at 289-293; CPU/memory/time config validation at 98-135 and timeout at 286; no-runtime Prover-off path at 174-188,249-254 and `review/open.go:25-36`. `pipeline.go:231-250` fills the base/change directories. The block in Closure prevents endorsing the failed-removal guarantee.

### Conditions
Do not close the previous cleanup finding until a long-output timeout plus failed-removal test observes the container name and removal error. Run independent T10 with a real Docker runtime before claiming actual network isolation or that a stock image runs Go tests. No condition downgrades a blocking finding here.

### Unrun
T10 skipped: Docker executable absent (`command -v docker` returned no path), and the test requires `PROVEN_TEST_DOCKER=1`; do not treat its skip as proof. Actual deletion-mutant executions unrun: no `apply_patch` command was available to edit the scratch copy; mapping above states expected test alarms, not measured mutant reds. Full repository build, vet and race suite not run: the order scoped this review to the listed checks. Live image pull and executable Go inside the container unrun: no Docker runtime.

## The review status as filed

## REVIEW — the-prover-in-a-container
- state: recommendation-ready
- verdict: reject
- ref: _REVIEW_NOTES.md
- abstract: Reject: failed container removal can again become silent when timed-out output exceeds 8192 bytes. T1–T9 pass; T10 unrun without Docker. Review performed by OpenAI Codex (model identifier not exposed).
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-22T04:48:15Z

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
