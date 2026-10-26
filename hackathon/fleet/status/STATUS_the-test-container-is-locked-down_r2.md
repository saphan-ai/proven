# STATUS — the-test-container-is-locked-down · round 2

- state: ready-for-review
- executor: not recorded
- model: not recorded
- pool: not recorded
- cost: not recorded
- duration: not recorded
- last commit: f7e9eef [the-test-container-is-locked-down] testrun: hardening tests pass argv words (the master: main absorbed, three conflicts hand-resolved; the real container run passes at this tip on █████████████; 37 packages ok)
- updated: 2026-10-23T19:36:10Z

## Summary

Round 3 done in one commit: every tree Proven mounts (staged checkout, live base/change trees) gets read and search for others and no write bit; the scratch default is 512m (measured build: about 135 MB cold), upper bound 1g kept.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — the-test-container-is-locked-down
- state: stop-2-ready
- abstract: Round 3 done in one commit: every tree Proven mounts (staged checkout, live base/change trees) gets read and search for others and no write bit; the scratch default is 512m (measured build: about 135 MB cold), upper bound 1g kept.
  The mandatory real Docker run is UNRUN: `docker` is outside this leg's permission class (and `which` was refused). The master runs `PROVEN_TEST_DOCKER=1 go test -count=1 -v -run 'TestWrittenGoTestRunsInTheContainer|TestContractW1bT10' ./internal/testrun`.
  `go build`, `go vet` exit 0; `go test ./...`: 6 packages FAIL, all the sandbox's refused local listener (as in the review); testrun, change, childenv ok.
- last commit: f7e9eef [the-test-container-is-locked-down] testrun: hardening tests pass argv words (the master: main absorbed, three conflicts hand-resolved; the real container run passes at this tip on █████████████; 37 packages ok)
- done: round 3 build items 1 and 2; round-1 and round-2 bounds untouched
  - MakeReadable (testrun/written.go) grants r and search to group and others, adds no write bit; used in stageWritten and Pipeline.materialize; proven by TestMakeReadableGrantsReadAndSearchAndNoWrite, TestStagedCheckoutIsReadableByOthers, TestMaterializedTreesAreReadableByOthers (all pass)
  - negative control (law #27, mode-assertion stand-in because no Docker): the grant removed from stageWritten, TestStagedCheckoutIsReadableByOthers FAIL ("mode 700 is not readable by others"); restored, passes
  - scratch: sample go test build, cold cache, measured on the host with -work: cache 66 MB + work dir 67 MB, about 135 MB; default 64m to 512m, TestArgvScratchIsBoundedTmpfs updated, docs/CONFIGURATION.md updated
- next / blocked-on: the master runs the Docker proof run; if it still fails, paste its output back
- artifacts:
  - internal/testrun/written.go
  - internal/testrun/readable_test.go
  - internal/pipeline/trees.go
  - internal/pipeline/trees_readable_test.go
- note: master-was-wrong: claim 1 (the user cannot read the checkout) is consistent with the code (MkdirTemp makes a 0700 root, proven by the failing negative control) but was not measured under Docker here. Claim 2 (scratch too small) is consistent with the measurement, 135 MB against 64m, not run in a container. GOTMPDIR is not set: TMPDIR is unset in the container so Go uses /tmp, which is the same scratch. Judgement calls: the memory default stays 512m (contract T4 pins it); tmpfs pages count against it, so a large build could hit the memory limit before the 512m scratch limit; the Docker run will show it. Commands and exit codes: go build ./... 0; go vet ./... 0; gofmt -l empty; go test ./internal/testrun ./internal/change ./internal/childenv 0; go test ./... 1 (cmd/livecheck, generator, guardrail, pipeline, query, review: listener panics; pipeline's new test passes alone). git commit printed packed-refs.lock Operation not permitted; the commit landed (git log -1). No Co-authored-by line. No session recording: headless one-shot, none provisioned (ADR-0024).
- updated: 2026-10-23T14:00:00Z

## REDACTED

| what | times |
|---|---|
| machine name | 2 |
