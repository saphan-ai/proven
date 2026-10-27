# STATUS — a-scenario-per-kind-of-evidence · round 1

- state: ready-for-review
- executor: cursor on runner-cursor
- model: grok-4.7-high
- pool: not recorded
- cost: not recorded
- duration: not recorded
- last commit: 7bf80260e0ddfa5f16fa119649564bde448d15e7 [a-scenario-per-kind-of-evidence] absorb main (the master: the leg's 64f54cf rewritten without the tool trailer as e3d28b8, then main bd9ed31 absorbed; fixture.go and pipeline.go hand-resolved)
- updated: 2026-10-23T07:56:59Z

## Summary

Four sample scenarios show a dropped refutation, an unsafe prompt, an OSV range with an EPSS score, and an unstated author model.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — a-scenario-per-kind-of-evidence
- state: stop-2-ready
- abstract: Four sample scenarios show a dropped refutation, an unsafe prompt, an OSV range with an EPSS score, and an unstated author model.
- last commit: 7bf80260e0ddfa5f16fa119649564bde448d15e7 [a-scenario-per-kind-of-evidence] absorb main (the master: the leg's 64f54cf rewritten without the tool trailer as e3d28b8, then main bd9ed31 absorbed; fixture.go and pipeline.go hand-resolved)
- done: Four scenarios run in sample mode, and a fixture can record one unsafe guardrail verdict.
  - refuted: go test -count=1 -run TestRefutedScenarioDropsTheFinding ./cmd/proven exit 0; the trace verdict is dropped: refuted: test TestTotal passes on the change; F1 is not printed
  - unsafe-prompt: go test -count=1 -run TestUnsafePromptStopsAndNamesCategories ./cmd/proven exit 0; failed service guardrail unsafe: Criminal Planning/Confessions, Other; the trace names those categories
  - osv-go-module: go test -count=1 -run TestOSVGoModuleShowsTheRangeAndTheEPSSScore ./cmd/proven exit 0; the finding prints the OSV rows, introduced 1.3.0 <= v1.4.0 < fixed 1.5.0, and EPSS 0.42000 (91st percentile, 2026-10-21)
  - unknown-author: go test -count=1 -run TestUnknownAuthorSaysTheDefaultReviewer ./cmd/proven exit 0; the scan prints author model not stated; default reviewer NVIDIA
  - negative control: the refuted skeptic response path, name, command and sql were emptied; go test -count=1 -run TestRefutedScenarioDropsTheFinding ./cmd/proven exit 1 with F1 was printed; the response was restored; the same command then exit 0
  - suite: go build ./... exit 0, go vet ./... exit 0, go test -count=1 ./... exit 0 on the tip, and the same test command exit 0 on the base before any edit
- next / blocked-on: none
- artifacts:
  - cmd/proven/scenario_test.go
  - docs/SAMPLE_MODE.md
  - internal/fixture/fixture.go
  - internal/pipeline/pipeline.go
  - testdata/fixture/refuted
  - testdata/fixture/unsafe-prompt
  - testdata/fixture/osv-go-module
  - testdata/fixture/unknown-author
- evidence:
  - baseline before edits: go version go1.26.8; go.mod says go 1.26 and has no toolchain line; go build ./... exit 0; go vet ./... exit 0; go test -count=1 ./... exit 0; every package ok or no test files
  - tip: go test -count=1 ./... exit 0; cmd/proven 2.370s; internal/pipeline 26.355s; internal/review 33.553s; internal/testrun 12.179s; no package failed
  - gofmt -l on the three Go files printed nothing
  - each scenario module: go build -C of base and of change exit 0 for refuted, unsafe-prompt, osv-go-module and unknown-author
  - proven binary: go build of ./cmd/proven exit 0; executing that binary was denied, allowed form none; the scan output is from go test calling run after changing to a scratch directory
  - negative control mutated field: the skeptic response path, name, command and sql; exit 1; failure line F1 was printed; restored; exit 0
  - race detector: UNRUN; the change does not add a goroutine
  - testing bar: UNRUN; the repository declares no coverage, mutation, lint or type metric
- note: master-was-wrong: sample mode does not read commit history on proven scan --sample, because the request sets no baseline commit and the fixture is two trees. The description field exists and the scan leaves it empty, so the review says author model not stated; default reviewer NVIDIA. unknown-author was not skipped. Cost if wrong: the scenario repeats a sentence the discount scan already prints. Ruling: a fixture guardrail list is how an unsafe prompt can stop a step, because the sample guardrail otherwise answers safe. Cost if wrong: a fixture that records verdicts replaces the safe sample for that review; discount records none. The commit message carries Co-authored-by: Cursor <██████████████████████>, added by the tool; the master removes it. Session recording (ADR-0024): none. This seat was not given a recording path, and no proxy receipt was present. Cost unverified: no cost row was handed to this seat; the order quote is 3.00 USD and the cap is 6.00 USD.
- updated: 2026-10-23T01:37:30Z

CLOSE-OK · a-scenario-per-kind-of-evidence · commit 64f54cf70c4f14445cd9088bd8d4313c0405d44e · branch fix/a-scenario-per-kind-of-evidence · worktree ███████████████████████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| e-mail or user name | 1 |
