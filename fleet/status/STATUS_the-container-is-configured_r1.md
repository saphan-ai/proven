# STATUS — the-container-is-configured · round 1

- state: ready-for-review
- executor: cursor
- model: grok-4.7-high
- pool: unverified (this seat did not receive a pool reading)
- cost: unverified (no Saphan Studio cost row was handed to this seat; the order's quote is $2.00 and its cap is $4.00)
- duration: about 20 minutes; the prompt timestamp was 2026-10-22T14:54:00Z and the end was measured at 2026-10-22T15:14:14Z
- last commit: b07a184fc6da979238335a24346c20fdf523301c "[the-container-is-configured] testrun: load the container settings"
- updated: 2026-10-22T15:14:14Z

## Summary

The Prover's container runtime, image, CPU, memory, time limit and written-test size load from the environment. Empty values keep the defaults in `DefaultContainerConfig`. A bad value is refused on load, by the variable's name, with the allowed form. A live review, the pipeline's live fallback, and `cmd/livecheck` use that load. Sample mode does not read the variables.

## Acceptance, item by item

1. `go build ./...` exit 0. `go vet ./...` exit 0. `go test -count=1 ./...` exit 0. The three commands were run separately. Summary lines are under Evidence. Baseline before any edit was the same test command, exit 0, every package `ok`.
2. Negative control. Mutated field: `PROVEN_TEST_IMAGE`, which the loader was made to ignore. `go test -count=1 -v ./internal/testrun -run TestLoadImage` exit 1. `TestLoadImage/default` passed. `TestLoadImage/other` failed (`PROVEN_TEST_IMAGE "golang:1.26"`). `TestLoadImage/refused` failed (`got <nil>`). The read was restored. The same command then exited 0, and all three subtests passed.
3. `PROVEN_TEST_MEMORY=lots go run ./cmd/livecheck` was not run. The executor grant refused the command: it treated `PROVEN_TEST_MEMORY=lots` as the program name. Allowed form: none. The command was not rebuilt. `TestBadContainerMemoryIsRefusedByName` calls the same `run` function with `PROVEN_TEST_MEMORY=lots` and requires the container line to contain `container: not ready:`, `PROVEN_TEST_MEMORY`, `allowed` and `512m`, and requires the runtime record to stay empty. It passed inside the suite above.

## Evidence

Toolchain: `go.mod` says `go 1.26` and has no `toolchain` line. `go version` reported `go1.26.8 linux/amd64`. No `.github` workflow was in the tree. The acceptance commands are the ones that were run.

`go test -count=1 ./...` exit 0:

```
ok  	github.com/saphan-ai/proven/cmd/livecheck	0.051s
?   	github.com/saphan-ai/proven/cmd/meterless-conform	[no test files]
ok  	github.com/saphan-ai/proven/cmd/proven	2.508s
ok  	github.com/saphan-ai/proven/conformance/meterless	0.007s
ok  	github.com/saphan-ai/proven/demo	0.523s
?   	github.com/saphan-ai/proven/demo/live/review	[no test files]
ok  	github.com/saphan-ai/proven/internal/agent/answer	0.003s
?   	github.com/saphan-ai/proven/internal/agent/investigator	[no test files]
ok  	github.com/saphan-ai/proven/internal/agent/judge	0.001s
ok  	github.com/saphan-ai/proven/internal/agent/planner	0.002s
ok  	github.com/saphan-ai/proven/internal/agent/prover	0.003s
ok  	github.com/saphan-ai/proven/internal/agent/skeptic	0.002s
?   	github.com/saphan-ai/proven/internal/agent/writer	[no test files]
?   	github.com/saphan-ai/proven/internal/candidate	[no test files]
ok  	github.com/saphan-ai/proven/internal/change	6.756s
ok  	github.com/saphan-ai/proven/internal/fixture	0.009s
ok  	github.com/saphan-ai/proven/internal/generator	0.216s
ok  	github.com/saphan-ai/proven/internal/guardrail	0.214s
ok  	github.com/saphan-ai/proven/internal/hmem	0.371s
?   	github.com/saphan-ai/proven/internal/limit	[no test files]
ok  	github.com/saphan-ai/proven/internal/pipeline	33.440s
ok  	github.com/saphan-ai/proven/internal/query	0.006s
ok  	github.com/saphan-ai/proven/internal/review	27.758s
?   	github.com/saphan-ai/proven/internal/service	[no test files]
ok  	github.com/saphan-ai/proven/internal/testrun	12.175s
ok  	github.com/saphan-ai/proven/internal/trace	0.135s
ok  	github.com/saphan-ai/proven/internal/world	0.826s
```

Per-test pass, fail and skip counts were not printed by that command. `TestContractW1bT10` skips unless `PROVEN_TEST_DOCKER=1`. That variable was not set. Whether it skipped on this run is UNMEASURED.

RED before the loader existed: `go test -count=1 ./internal/testrun -run TestLoad` exit 1. The compiler named the missing symbols `Load`, `EnvRuntime`, `EnvImage`, `EnvCPUs`, `EnvMemory`, `EnvTimeout` and `EnvMaxWrittenBytes`.

`go test -race` UNRUN. This change does not add or alter goroutines or shared mutable state.

Testing bar: no metric, command or baseline is declared for this order. UNRUN. No bar was edited.

Session recording (ADR-0024): none was provisioned for this seat, and none is published in this repository for this run.

## master-was-wrong

No. On 2026-10-22, at the worktree tip before this change (`39225b0`), `internal/testrun/container.go` held the six defaults and `ContainerConfig.Validate`, and both `internal/pipeline/pipeline.go` and `internal/review/open.go` passed `DefaultContainerConfig()` to `NewContainer`. The reading did not name `cmd/livecheck` or the prover's `MaxWrittenBytes` argument. Both were still the hardcoded default. They now use the loaded settings. The order named base `main` at `31fbb97`. That commit is an ancestor of the worktree. `main` at the same measurement was `f924fb9`. The four files this change edits had no diff against that `main`.

## Rulings

- ruling: the loader's refusal names the environment variable — the other loaders do that, and `Validate` still names the struct field — a caller that looks for `Memory` instead of `PROVEN_TEST_MEMORY` will not match.
- ruling: a bad setting refuses the review — a wrong value is a configuration error, and a missing runtime still names the Prover off — a typo stops the review instead of degrading it.
- ruling: the loaded size limit is also the prover's pre-check — the container and the prover would otherwise disagree when the limit is raised — zero still means the default, so sample mode is unchanged.
- ruling: a duration or an integer that does not parse is refused before `Validate` — those strings cannot be stored on the config — zero and other out-of-range values still go through `Validate`.

## NOTICED

`AGENTS.md` lists the container defaults and does not name `PROVEN_TEST_RUNTIME`, `PROVEN_TEST_IMAGE`, `PROVEN_TEST_CPUS`, `PROVEN_TEST_MEMORY`, `PROVEN_TEST_TIMEOUT` or `PROVEN_TEST_MAX_WRITTEN_BYTES`. The order leaves the docs pages to the master. They were not edited.

The executor grant cannot run the acceptance command that sets `PROVEN_TEST_MEMORY` on the same line as `go run`. Allowed form was none. This is a grant fact, not a product failure. The unit test above is the proof that was run.

## Blocked on / next

None for this leg. The master verifies. No push, no merge, no pull request.
