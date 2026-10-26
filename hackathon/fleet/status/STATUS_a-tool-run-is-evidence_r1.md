# STATUS — a-tool-run-is-evidence · round 1

- state: ready-for-review
- executor: cursor on runner-cursor
- model: grok-4.7-high
- pool: not recorded
- cost: not recorded
- duration: 23 min
- last commit: 94a5b2c453e96f8b3d33f4ef61ac2188620b2a0f [a-tool-run-is-evidence] tool: add the race detector
- updated: 2026-10-24T03:03:42Z

## Summary

The evidence kind tool maps to proved-by-tool. The race detector's recorded scenario reports a finding when the change races and the base does not. A real container run is UNRUN on this seat.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — a-tool-run-is-evidence
- state: stop-2-ready
- abstract: The evidence kind tool maps to proved-by-tool. The race detector's recorded scenario reports a finding when the change races and the base does not. A real container run is UNRUN on this seat.
- last commit: 94a5b2c453e96f8b3d33f4ef61ac2188620b2a0f [a-tool-run-is-evidence] tool: add the race detector
- done: The Go race detector is the first tool evidence, compared on the base and the change.
  - TestSampleRaceReportsAFindingProvedByATool passed (go test -count=1 -run TestSampleRaceReportsAFindingProvedByATool ./internal/tool, exit 0): the recorded scenario is a finding the Judge keeps, at proved-by-tool.
  - TestUnrunToolIsNamed and TestTimeoutIsNamedEvenWhenTheBannerWasPrinted passed in the same package (exit 0): a tool that does not answer is named and is not evidence.
  - TestRaceOnTheBaseIsNotEvidence went red when Introduced ignored PresentOnBase (exit 1, "a race on the base was evidence"), then passed after that check was restored (exit 0).
  - go build ./... exit 0 and go vet ./... exit 0.
- next / blocked-on: none
- artifacts:
  - docs/ADAPTERS.md
  - docs/CONFIGURATION.md
  - docs/README.md
  - internal/tool/tool.go
  - internal/tool/race.go
  - internal/tool/sample.go
  - internal/tool/settings.go
  - internal/tool/host.go
  - internal/candidate/candidate.go
  - internal/review/review.go
- evidence:
  - go version exit 0: go1.26.8 linux/amd64. Module go line is 1.26. Race banner WARNING: DATA RACE measured in that toolchain's race runtime.
  - Baseline before edits: go test -count=1 ./... exit 1. The only failure was TestMaterializedTreesAreReadableByOthers: go.mod mode 664 is writable by group or others. Every other package was ok.
  - After the change: go test -count=1 ./... exit 1. The same test failed the same way. internal/tool, internal/testrun, internal/review, internal/cli, internal/trace, internal/agent/judge and internal/settingspage were ok.
  - Negative control: with PresentOnBase ignored, go test -count=1 -run TestRaceOnTheBaseIsNotEvidence ./internal/tool exit 1. After restore, the same command exit 0.
  - Docker run of go test -race: UNRUN. docker is not in this leg's program grant, so no container was started.
  - Session recording: none. This leg wrote no recording (ADR-0024).
- note: master-was-wrong: none on the decision. The order's base 5a9f2fe is an ancestor of this branch (git merge-base --is-ancestor 5a9f2fe HEAD, exit 0). The tip before this leg was 8770336, which already had ProvedByTool and no kind that set it; this leg added the kind and the ProofLevelOf rule. Rulings: the six-agent pipeline does not call the tool, so existing sample reviews stay at zero tool findings, and the cost is that proven scan does not yet report a race. The live version field is unknown because no second container asks go version; the sample records go1.26 from the module go line. The command is go test -race on the packages of changed Go files. review.Open does not load PROVEN_TOOL_TIMEOUT; tool.Load refuses a bad value when a caller uses it. No Co-authored-by trailer on either commit.
- updated: 2026-10-24T02:58:11Z

CLOSE-OK · a-tool-run-is-evidence · commit 94a5b2c453e96f8b3d33f4ef61ac2188620b2a0f · branch fix/a-tool-run-is-evidence · worktree ██████████████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
