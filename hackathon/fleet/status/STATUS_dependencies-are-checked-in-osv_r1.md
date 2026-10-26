# STATUS — dependencies-are-checked-in-osv · round 1

- state: ready-for-review
- executor: not recorded
- model: not recorded
- pool: not recorded
- cost: not recorded
- duration: not recorded
- last commit: ee6e55f730e4138ef735b0c4eaf8d9121edba3d7 [dependencies-are-checked-in-osv] pipeline: check added modules in osv
- updated: 2026-10-22T16:52:07Z

## Summary

A module the change adds or raises is checked against the OSV package table; a matching range is a finding whose evidence is the query, the rows and the Go comparison.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — dependencies-are-checked-in-osv
- state: stop-2-ready
- abstract: A module the change adds or raises is checked against the OSV package table; a matching range is a finding whose evidence is the query, the rows and the Go comparison.
- last commit: ee6e55f730e4138ef735b0c4eaf8d9121edba3d7 [dependencies-are-checked-in-osv] pipeline: check added modules in osv
- done: OSV package check is in the sample review and the suite is green
  - go version go1.26.8 linux/amd64
  - go build ./... && go vet ./... && go test -count=1 ./... exited 0
  - negative control: InRange returned true for every version; TestCheckTracesAnUnaffectedVersion and TestInRange/outside_every_range failed; the forced return was removed; the suite then passed
  - go test -count=1 -v ./cmd/proven -run TestScanSamplePrintsF1AndF2NotF3 exited 0 from a temp directory and printed the leftpad OSV finding
- next / blocked-on: none
- artifacts:
  - cmd/osv-to-csv
  - internal/osv
  - internal/manifest
  - internal/semver
  - internal/pipeline/pipeline.go
  - internal/pipeline/osv_test.go
  - internal/pipeline/guardrail_test.go
  - internal/review/open.go
  - internal/review/live_generator_test.go
  - internal/review/selection_test.go
  - internal/review/zetaris_test.go
  - cmd/proven/main_test.go
  - testdata/osv
  - testdata/fixture/discount/tables.sql
  - docs/OSV_TABLE.md
  - docs/SAMPLE_MODE.md
  - docs/ARCHITECTURE.md
  - docs/CONFIGURATION.md
  - docs/README.md
- evidence:
  - gofmt -w on the new Go files exited 0
  - go test -count=1 ./internal/semver ./internal/manifest ./internal/osv ./cmd/osv-to-csv exited 0
  - go test -count=1 ./internal/review first failed TestZetarisUnauthorizedReviewCompletes (zetaris calls 2, wanted 1); the assertion now expects 2 calls; go test -count=1 ./internal/review then exited 0
  - go build ./... && go vet ./... && go test -count=1 ./... exited 0: cmd/proven 1.092s, internal/pipeline 18.780s, internal/review 15.630s, internal/osv 0.014s, internal/semver 0.002s, internal/manifest 0.003s, cmd/osv-to-csv 0.005s
  - negative control command go test -count=1 ./internal/osv -run 'TestInRange|TestCheckTracesAnUnaffectedVersion|TestCheckKeepsAMatch' exited 1 while InRange began with return true
  - sample finding line: finding osv:evil.example/leftpad ERROR Dependency evil.example/leftpad v1.2.3 is in an OSV affected range
  - sample comparison: evil.example/leftpad v1.2.3 matches GHSA-px-sample-0001: introduced 1.0.0 <= v1.2.3 < fixed 1.2.4. GHSA-px-sample-0002 does not match: v1.2.3 < introduced 2.0.0. alias CVE-2026-0001 is on the KEV list
  - scratch go run was not achieved: go -C the repo from a scratch directory exited 2 (hmem saw the repository as the process cwd); go run of the absolute cmd/proven path from the scratch directory exited 1 (go.mod file not found); executing a built binary was refused as not in the executor allow list
- note: master-was-wrong: no. M1 (2026-10-22) holds. The ██████ notes say a SELECT returns every row of the filestore table. They do not say to compare versions in SQL. The check filters ecosystem and package in SQL and compares the range in Go.
  - ruling: PROVEN_OSV_TABLE is an environment variable, not a proven.yaml key, because the yaml key set is closed and holds only timeout. Cost if wrong: an operator who looks only in proven.yaml misses the setting. It is documented in docs/OSV_TABLE.md and docs/CONFIGURATION.md.
  - ruling: one finding per module, id osv:<module>. Cost if wrong: a second advisory is not its own finding. Its id, comparison and rows stay in the message and the query evidence.
  - ruling: a replace or exclude directive is not a version change. Cost if wrong: a replace that raises the selected module without changing require is not checked.
  - ruling: an empty introduced, or introduced 0, means no lower bound.
  - ruling: when both fixed and last_affected are set, fixed wins and is exclusive.
  - ruling: a versions list with no ranges becomes one row per version, with introduced and last_affected set to that version.
  - ruling: aliases are sorted and joined with a space.
  - ruling: the check does not call a model, so it still runs when the planner is stopped. Guardrail tests that require no model finding ignore ids that start with osv:.
  - ruling: comparison sentences use ASCII <=.
  - ruling: the finding severity is ERROR.
  - ruling: sample mode does not read PROVEN_OSV_TABLE.
  - both commits carry a Co-authored-by trailer the tool added. The messages written for the commits did not contain it. The master removes it.
  - session recording (ADR-0024): none in this worktree. Run id run-20261022-153601-e5c9bbfa has an empty evidence directory. No recording was provisioned for this seat, and publishing one is not this order.
  - testing bar: no coverage or mutation metric is declared for this stream. Those numbers are UNRUN. go test -race is UNRUN; the acceptance command did not include it and the new packages start no goroutines.
  - quote 3.00 USD, cap 6.00 USD. The dollar actual is unmeasured on this seat.
- updated: 2026-10-22T16:11:35Z

CLOSE-OK · dependencies-are-checked-in-osv · commit 5caee0a88911c8f9fc3d79f6276231b5a7069e19 · branch fix/dependencies-are-checked-in-osv · worktree ███████████████████████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| private exchange | 1 |
