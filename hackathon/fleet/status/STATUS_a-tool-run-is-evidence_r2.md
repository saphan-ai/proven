# STATUS — a-tool-run-is-evidence · round 2

- state: ready-for-review
- executor: cursor on runner-cursor
- model: grok-4.7-high
- pool: not recorded
- cost: not recorded
- duration: 11 min
- last commit: 0a4efb3ebf4f71c0750d1e0028b7f2e65e1e4954 [a-tool-run-is-evidence] absorb main
- updated: 2026-10-24T04:02:43Z

## Summary

A Docker-gated test runs the race detector in the Prover's container on a clean base and a change that races. Without PROVEN_TEST_DOCKER=1 the test skips. The master runs that test on a host with Docker.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — a-tool-run-is-evidence
- state: stop-2-ready
- abstract: A Docker-gated test runs the race detector in the Prover's container on a clean base and a change that races. Without PROVEN_TEST_DOCKER=1 the test skips. The master runs that test on a host with Docker.
- last commit: 0a4efb3ebf4f71c0750d1e0028b7f2e65e1e4954 [a-tool-run-is-evidence] absorb main
- done: The race tool has one real container test, and it skips unless PROVEN_TEST_DOCKER=1.
  - go test -count=1 -v -run TestRaceToolRunInTheContainerProvesTheChange ./internal/tool exit 0. The test printed SKIP and PROVEN_TEST_DOCKER=1.
  - go build ./... exit 0. go vet ./... exit 0. go test -count=1 ./internal/tool exit 0. The verbose package run was 14 pass, 1 skip, 0 fail.
  - The test requires the command go test -race . and a non-empty version. A live run records unknown, because the adapter does not ask the container.
- next / blocked-on: the master, from the repository root, with image golang:1.26 already present, runs PROVEN_TEST_DOCKER=1 go test -count=1 -timeout 15m -run TestRaceToolRunInTheContainerProvesTheChange ./internal/tool
- artifacts:
  - internal/tool/race_container_test.go
  - docs/ADAPTERS.md
- evidence:
  - go version exit 0: go1.26.8 linux/amd64. Module go line is 1.26. No toolchain line.
  - Baseline before the edit: go test -count=1 ./internal/tool exit 0.
  - go build ./... exit 0. go vet ./... exit 0. go test -count=1 ./internal/tool exit 0.
  - go test -count=1 -v ./internal/tool exit 0: 14 PASS, 1 SKIP (TestRaceToolRunInTheContainerProvesTheChange, reason PROVEN_TEST_DOCKER=1), 0 FAIL.
  - go test -race -count=1 ./internal/tool exit 0. The gated test still skipped.
  - go test -count=1 ./... exit 1. The only failure was TestMaterializedTreesAreReadableByOthers: go.mod mode 664 is writable by group or others. internal/tool was ok. The same failure was already on the base in round 1.
  - remark: docker on PATH is UNMEASURED. The test returns at the env gate before LookPath, so a present docker and a missing docker print the same SKIP. The program command is not in this grant, so PATH was not probed.
  - remark: whether a cold race compile finishes inside the default 120s side limit is UNMEASURED. This seat never started a container.
  - Negative control UNRUN. No container was started, so PresentOnBase could not be shown true. The master makes the base tree race the same way as the change; the test must then fail, and the file is restored.
  - Testing bar UNRUN. No metric is declared.
  - Session recording: none. This leg wrote no recording.
- note: master-was-wrong: none. Before this commit no test in internal/tool started a container. Rulings: filled version means non-empty, and the live value stays unknown because Race.Version is empty — cost if wrong, the master wanted the image's go version. The trees are passed through MakeReadable because RunCommand mounts the directory the test wrote, and a temp directory is not searchable by the container user — cost if wrong, an extra mode change. docs/ADAPTERS.md no longer says that no test starts a container — cost if wrong, a sentence the order did not name. No Co-authored-by trailer on the commit.
- updated: 2026-10-24T03:31:00Z

CLOSE-OK · a-tool-run-is-evidence · commit 48be6dcffbed12f033038fc8e6d2dedf9cea69b9 · branch fix/a-tool-run-is-evidence · worktree ██████████████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
