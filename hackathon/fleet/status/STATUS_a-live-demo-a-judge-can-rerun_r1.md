# STATUS — a-live-demo-a-judge-can-rerun · round 1

- state: ready-for-review
- executor: not recorded
- model: not recorded
- pool: not recorded
- cost: not recorded
- duration: not recorded
- last commit: 7ec8f27446cd2e10ba2f8c7f14db605d7be2fc94 [a-live-demo-a-judge-can-rerun] demo: add a rerunnable live review
- updated: 2026-10-22T14:50:11Z

## Summary

demo/live is a base/change review a judge can rerun: TestShare fails on the change, and change/go.mod requires golang.org/x/net v0.16.0 for CVE-2023-44487. With no settings, run.sh stops at livecheck.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — a-live-demo-a-judge-can-rerun
- state: stop-2-ready
- abstract: demo/live is a base/change review a judge can rerun: TestShare fails on the change, and change/go.mod requires golang.org/x/net v0.16.0 for CVE-2023-44487. With no settings, run.sh stops at livecheck.
- last commit: 7ec8f27446cd2e10ba2f8c7f14db605d7be2fc94 [a-live-demo-a-judge-can-rerun] demo: add a rerunnable live review
- done: The demo is committed and the acceptance commands passed.
  - TestShare passes on a temp copy of base/ and fails on a temp copy of change/: go test -count=1 ./demo exit 0
  - Equal trees turn that test red: go test -count=1 ./demo -run TestShareFailsOnChangeAndPassesOnBase exit 1, line "change: go test passed"; the tree was restored and the same package then exited 0
  - Parent gate go build ./... && go vet ./... && go test -count=1 ./... exit 0, including demo and demo/live/review
  - bash demo/live/run.sh with no settings exit 1 and names generator, guardrail, query and container not ready; the same exit from another directory
  - Import ban stays green inside the pipeline package, and demo/ is outside internal/ and cmd/
- next / blocked-on: none
- artifacts:
  - demo/live/README.md
  - demo/live/fixture.json
  - demo/live/run.sh
  - demo/live/tables.sql
  - demo/live/base/go.mod
  - demo/live/base/split/share.go
  - demo/live/base/split/share_test.go
  - demo/live/change/go.mod
  - demo/live/change/split/share.go
  - demo/live/change/split/share_test.go
  - demo/live/review/main.go
  - demo/live_test.go
- evidence:
  - toolchain go1.26.8; module go line 1.26. Baseline before the edit: go test -count=1 ./... exit 0
  - go build ./... && go vet ./... && go test -count=1 ./... exit 0. Build and vet printed nothing. Test lines included ok github.com/saphan-ai/proven/demo 0.445s and ? github.com/saphan-ai/proven/demo/live/review [no test files]
  - bash demo/live/run.sh exit 1. generator, guardrail, query and container each "not ready", summary 0 ready, 4 not ready. No secret was set, and none was printed
  - The same script from another directory, by absolute path, exit 1 with the same four lines
  - Negative control mutated change/ to a copy of base/. The demo test exited 1. Restored. go test -count=1 ./demo exit 0
  - No go.sum under demo/live/base or demo/live/change. The demo test runs go build and go test with GOPROXY=off and GOSUMDB=off
- note: master-was-wrong: no. An empty baseline reads base/ and change/ (change.Read, prover checkouts). Git is used only when BaselineCommit is set. fixture.Load still requires fixture.json and tables.sql for a live review; both are in the demo and live mode does not use the recorded response or the SQL file. proven scan accepts only --sample (cmd/proven, TestScanWithoutSampleIsRefused), so run.sh builds that binary and reviews through demo/live/review. remark: CVE-2023-44487 and the v0.17.0 fix were not read from the catalogue or the advisory in this run. A KEV query for that id returns no row if the CVE is absent, and this run cannot tell those worlds apart, so the pairing is UNVERIFIED. The offline KEV row in this tree is CVE-2021-44228, Apache Log4j2, which is not a Go module. ruling: the driver is not a new proven flag — flags are the shell's track — cost if wrong: a later live scan leaves a second entry point. ruling: no declared testing-bar metric, so the bar is UNRUN; no skip, threshold or assertion was weakened. -race UNRUN: the change adds no shared state and the order's gate does not include it. A live Review was not run: livecheck was not ready, and tests must not call a live service. No session recording was given to this leg and none is in the worktree (ADR-0024). The commit carries a Co-authored-by trailer this leg did not write. It was not amended: the order says the tool puts the trailer back on amend. The master removes it.
- updated: 2026-10-22T14:21:38Z

CLOSE-OK · a-live-demo-a-judge-can-rerun · commit de390a30a8f7be46c281c9236f22b22bf050bcd1 · branch fix/a-live-demo-a-judge-can-rerun · worktree █████████████████████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
