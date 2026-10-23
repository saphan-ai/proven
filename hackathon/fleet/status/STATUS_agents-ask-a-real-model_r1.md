# STATUS — agents-ask-a-real-model · round 1

- state: ready-for-review
- executor: cursor
- model: grok-4.7-high
- pool: not shown to this leg
- cost: not recorded by Studio in this session
- duration: not measured; the code commit is 2026-10-22T07:23:57Z
- last commit: b28adc3e2b4046f2ff3f7ca5eb8cb37d2ce163c5 "[agents-ask-a-real-model] agent: ask for JSON and name a bad answer"
- updated: 2026-10-22T07:23:57Z

## Summary

The planner, investigator, prover and skeptic ask for one JSON shape and one example, and they ask for that JSON only. A shared reader takes the JSON from a fenced block or from surrounding prose, and it accepts a trailing comma. A bad answer is sent back once. A second failure names the agent and the review finishes. The sample review still reports F1 and F2, and it still drops F3.

## Acceptance, item by item

1. `go build ./...` exit 0. `go vet ./...` exit 0. `go test -count=1 ./...` exit 0. A counted run of that suite, before a test-only path check, was 139 pass, 0 fail, 1 skip. The skip is `TestContractW1bT10` in `internal/testrun` (`PROVEN_TEST_DOCKER=1` was not set). That test was not part of this change. The suite was run again after the path check and exited 0.
2. `proven scan --sample testdata/fixture/discount` was not run. The executor grant refused the `proven` binary. The allowed form was none. `go test -count=1 ./cmd/proven` ran `TestScanSamplePrintsF1AndF2NotF3`, which calls the same scan function. It passed. It requires F1, F2, the recorded test, the query row, and `failed services: none`, and it fails if F3 is printed. The CLI stdout itself is unverified.
3. Negative control. Mutated field: the skeptic answer. A generator returned the prose `1. this is not json` for every skeptic prompt, including the one repair. `TestProseTwiceDegradesByNameAndDropsUnevidenced` passed: the review returned no error, skeptic was called 6 times, the failed service was `skeptic` / `the model's answer did not follow the schema`, F1 and F2 stayed, and F3's verdict was `dropped: no evidence`. With that degrade return removed, the same test failed: `review ended: answer: invalid character ' ' after decimal point in numeric literal`. The return was put back and the test passed.

## Evidence

Baseline, before any edit: `go test -count=1 ./...` exit 0. Per-test counts were not taken on that run.

On the base, before the reader existed, a prose skeptic answer ended the review: `review ended: skeptic: response is not a refutation: invalid character ' ' after decimal point in numeric literal`.

`git log --format=%B 0e2e0c1..HEAD | grep -ci co-authored` printed `0`.

Go: `go1.26.8`. `go.mod` says `go 1.26` and has no toolchain line.

No coverage bar, mutation score, lint count or type-warning count is declared. Those metrics are UNRUN. No threshold, skip or assertion was weakened. `-race` was not run. This change does not share state across goroutines.

Session recording (ADR-0024): no published recording is in this repository. A pending local transcript segment was present (1043 events, sealed 2026-10-22T07:14:17.679Z, sha256 0bf63a6c2240d54bd4ae7e4954c424d2f55bef03261a6c82a3d866e4ad7a6dda). Publishing it is not this order.

## Rulings

- ruling: only the planner, investigator, prover and skeptic got live prompts — the judge and the writer do not call the generator — cost if wrong: a live review still does not ask those two for JSON.
- ruling: a JSON value that parses and then fails a later check (empty id, duplicate id) still ends the review — the order named a parse failure — cost if wrong: a blank id still ends the review with exit 2.
- ruling: sample mode uses a generator supplied on options when one is set, and the sample query and recorded tests stay — production leaves it unset — cost if wrong: none on the path Open uses.

## master-was-wrong

none. The prose answer reproduced the measured parse error on the base.

## Blocked on / next

none. The CLI stdout in acceptance 2 is unverified, as stated above.
