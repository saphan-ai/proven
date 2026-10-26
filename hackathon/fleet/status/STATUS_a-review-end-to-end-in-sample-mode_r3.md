# STATUS — a-review-end-to-end-in-sample-mode · round 3

- state: ready-for-review
- executor: cursor on runner-cursor
- model: grok-4.7-high
- pool: not recorded
- cost: not recorded
- duration: 31 min
- last commit: 041b3f24db8f84259da56d5c34d5fd072a6d5f5e [a-review-end-to-end-in-sample-mode] docs: sample review status
- updated: 2026-10-22T01:36:29Z

## Summary

Sample mode runs the six agents behind the review seam. F1 and F2 stay with evidence. F3 is dropped. No network.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — a-review-end-to-end-in-sample-mode
- state: stop-2-ready
- abstract: Sample mode runs the six agents behind the review seam. F1 and F2 stay with evidence. F3 is dropped. No network.
- last commit: 041b3f24db8f84259da56d5c34d5fd072a6d5f5e [a-review-end-to-end-in-sample-mode] docs: sample review status
- done: Sample mode returns F1 and F2 with evidence and drops F3.
  - go test -count=1 ./... exit 0, 22 passed, 0 failed, 0 skipped
  - F3 verdict text is dropped: no evidence
  - live test-run names prover off and F1 is not proved
  - sample review dials 0, and a direct dial on the guard is counted
- next / blocked-on: none
- artifacts:
  - AGENTS.md
  - cmd/proven/main.go
  - cmd/proven/main_test.go
  - internal/agent
  - internal/candidate
  - internal/change
  - internal/fixture
  - internal/generator
  - internal/limit
  - internal/netguard
  - internal/pipeline
  - internal/query
  - internal/review/open.go
  - internal/review/open_test.go
  - internal/review/review.go
  - internal/service
  - internal/testrun
  - internal/trace
  - testdata/fixture/discount
  - fleet/status/STATUS_a-review-end-to-end-in-sample-mode_r1.md
- evidence:
  - go build ./... exit 0
  - go vet ./... exit 0
  - go test -count=1 ./... exit 0
  - go test -race -count=1 ./... exit 0
  - baseline before edits: go test -count=1 ./... exit 0, internal/review had no test files
  - RED before Open was wired: review: not implemented
  - work commit 8e5147d6574ac2344050a43d9d22303bc3369a2b
  - toolchain go1.26.8, module go 1.26
- note: master-was-wrong: none. M1 held (seam already has both evidence kinds and FailedServices). M2 held (local SELECT returns the sample row, no cgo). M3 held (six agent packages, go build exit 0, no import cycle). Session recording: none; no recording file was handed to this leg. The built binary was not executed: the permission class refused it and the allowed form was none. go test ran the same run function. Rulings are in the public status.
- updated: 2026-10-22T01:33:25Z

## REDACTED

none
