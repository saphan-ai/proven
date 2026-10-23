# STATUS — the-doubled-review-runs-twice · round 1

- state: ready-for-review
- executor: cursor
- model: grok-4.7-high
- pool: not shown to this leg
- cost: not recorded by Studio in this session
- duration: not measured
- last commit: 3c0f08688a03b63644fad95a45f934f7c23e180c "[the-doubled-review-runs-twice] pipeline: run the special reviewer"
- updated: 2026-10-22T14:25:53Z

## Summary

When a review class is doubled and a special-review model is configured, the planner and the skeptic run a second time on that model. The second run does not receive the first run's findings. Both sets of findings go through one investigator, one prover and one judge. A finding is kept only when it has evidence. A finding from both runs is reported once and names both models. A finding from one run names that model. When the special model is not configured, the review says so and runs once. The trace names `default` or `special` on each claim.

## Acceptance, item by item

1. `go build ./...` exit 0. `go vet ./...` exit 0, no output. `go test -count=1 ./...` exit 0. Every package printed `ok` or `[no test files]`. Go is `go1.26.8`. `go.mod` says `go 1.26` and has no toolchain line.
2. `TestDoubledSecurityRunsBothGenerators` uses two fake generators on the discount fixture, which is class security. The special generator is called for the planner and the skeptic, and not for the investigator or the prover. A finding from both runs is kept once with models `NVIDIA` and `qwen-special` and with query evidence. A finding from only the default run names `NVIDIA`. A finding from only the special run names `qwen-special`. Two findings with no evidence are absent. The trace claims are `default`, `special`, or both. `TestNonDoubledClassRunsOnce`: path rules that leave the fixture as backend, special generator call count 0. `TestDoubledWithoutSpecialModelRunsOnce`: security, no special model, the statement `special-review model: not configured`, special generator call count 0.
3. Negative control. Mutated field: the generator the second run calls, in `secondGenerator`. It returned the first generator. `go test -count=1 ./internal/review/ -run TestDoubledSecurityRunsBothGenerators` exit 1: `special model was not called: []review.scriptCall(nil)`. The return of the special generator was restored. The full suite then exited 0.
4. `go run ./cmd/proven scan --sample testdata/fixture/discount` from the repository root exit 2: `hmem: state directory …/state/hmem is inside the repository`. That guard was already in the tree. This change does not edit the shell, hmem, or the world store. `go test -count=1 ./cmd/proven` includes `TestScanSamplePrintsF1AndF2NotF3`, which calls the same scan function with a working directory outside the module. It passed. It requires F1, F2, F4, and `failed services: none`, and it fails if F3 is printed.

## Evidence

Baseline, before any edit: `go test -count=1 ./...` exit 0. No package failed.

The code commit message contains a `Co-authored-by` trailer. This leg did not write it. The master removes it.

No coverage bar, mutation score, lint count or type-warning count is declared. Those metrics are UNRUN. No threshold, skip or assertion was weakened. `-race` was not run. This change does not share state across goroutines.

Session recording (ADR-0024): none. This seat was not given a recording path, and no proxy receipt was present.

## Rulings

- ruling: a finding found by both runs is the same planner id — cost if wrong: two wordings of one bug stay two findings.
- ruling: the investigator and the prover run once, on the default generator; the special generator is the planner and the skeptic — cost if wrong: a special-only finding is evidenced by the default model's query.
- ruling: sample mode with a special reviewer and no injected second generator calls the sample generator again — cost if wrong: the sample answers do not differ by model.
- ruling: `Finding.Models` is an additive field on the review seam; the shell does not print it — cost if wrong: a reader of the command output does not see which model found the finding.
- ruling: when both skeptics run, the stored refutation is the default run's; the special attempt is in the trace and does not replace it — cost if wrong: a later memory match sees the default refutation only.

## NOTICED

The sample command from the repository root exits 2 because the state directory would sit inside the module. The shell is out of scope for this order.
