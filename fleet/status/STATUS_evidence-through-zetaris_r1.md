# STATUS — evidence-through-zetaris · round 1

- state: ready-for-review
- executor: cursor
- model: grok-4.7-high
- pool: unverified (this seat did not receive a pool reading)
- cost: unverified (no Saphan Studio cost row was handed to this seat; the order's quote is $3.00 and its cap is $6.00)
- duration: 37 minutes, wall time from the spawn timestamp 2026-10-22T10:42:00Z to 2026-10-22T11:19:17Z
- last commit: 02aa89f9635b3f092ca2a4fd20cfa8dbc0735a55 "[evidence-through-zetaris] query: read KEV rows through Zetaris"
- updated: 2026-10-22T11:19:17Z

## Summary

The Investigator can carry a finding whose evidence is a query against `px_evidence.px_kev_vulns` and the row it returned. Sample mode runs that SQL on local tables shaped like the measured views. A live review uses the SQL-editor route the 2026-10-21 measurement proved. A 401 or a closed endpoint names `zetaris` and the review still finishes. F1 and F2 stay. F3 stays absent.

## Acceptance, item by item

1. `go build ./...` exit 0. `go vet ./...` exit 0. `go test -count=1 ./...` exit 0. Summary lines are under Evidence. Baseline before any edit was the same command, exit 0, every package `ok`.
2. Sample scan. The program `proven` is denied for this role (allowed form: none), so the binary was not run. `go test -count=1 ./cmd/proven` runs `TestScanSamplePrintsF1AndF2NotF3`, which calls the same `run` function. It passed inside the suite above. The lines that test requires, and fails if any is missing, are:

```
finding F1 ERROR Discount is applied twice
evidence test billing/discount_test.go TestDiscounted fails_on_change=true passes_on_base=true
finding F2 ERROR Dependency leftpad v1.2.3 is on the known-exploited list
evidence query local-sample snapshot=sample-2026-10-22
row: evil.example/leftpad | 1.2.3 | CVE-2026-0001 | true
finding F4 ERROR CVE-2026-0001 for leftpad v1.2.3 is on the KEV list
SQL: SELECT cve_id, vendor_project, product, date_added, known_ransomware FROM px_evidence.px_kev_vulns WHERE cve_id = 'CVE-2026-0001'
row: CVE-2026-0001 | evil.example | leftpad | 2026-01-15 | Known
failed services: none
```

The test fails if the output contains `finding F3`.

3. Negative controls. Each passed inside `go test -count=1 ./...`.
   - Mutated field: `version` on the advisories sample row, from `1.2.3` to `9.9.9`. `TestNegativeControl_VersionTheAdvisoryDoesNotCover`: F2 is absent. F1 stays. F4 stays, because the KEV view has no version column.
   - Mutated field: `cve_id` on the EPSS sample row that was `CVE-2026-0001`, changed to `CVE-0000-0000`. `TestNegativeControl_CVEAbsentFromEPSSIsNotCalledNotExploited`: while the CVE is on both sides, F5 is kept with the join row; after the mutation F5 is absent, the verdict is `dropped: no evidence`, and the trace does not contain `not exploited`.
   - Mutated field: the HTTP status, set to 401. `TestZetarisUnauthorizedReviewCompletes`: one call, failed service `zetaris` reason `unauthorized`, the review returns no error and is not partial. The API key is not in the failure.
4. Unplug. Mutated field: the listening socket, closed before the review. `TestZetarisClosedEndpointDegradesTheReview`: failed service `zetaris` (`unreachable` or `timeout`), the review is not partial.

## Evidence

`go test -count=1 ./...` exit 0:

```
ok  	github.com/saphan-ai/proven/cmd/proven	0.197s
ok  	github.com/saphan-ai/proven/internal/agent/answer	0.002s
?   	github.com/saphan-ai/proven/internal/agent/investigator	[no test files]
ok  	github.com/saphan-ai/proven/internal/agent/judge	0.001s
ok  	github.com/saphan-ai/proven/internal/agent/planner	0.003s
?   	github.com/saphan-ai/proven/internal/agent/prover	[no test files]
?   	github.com/saphan-ai/proven/internal/agent/skeptic	[no test files]
?   	github.com/saphan-ai/proven/internal/agent/writer	[no test files]
?   	github.com/saphan-ai/proven/internal/candidate	[no test files]
ok  	github.com/saphan-ai/proven/internal/change	0.044s
ok  	github.com/saphan-ai/proven/internal/fixture	0.003s
ok  	github.com/saphan-ai/proven/internal/generator	0.210s
?   	github.com/saphan-ai/proven/internal/limit	[no test files]
ok  	github.com/saphan-ai/proven/internal/netguard	0.002s
ok  	github.com/saphan-ai/proven/internal/pipeline	2.151s
ok  	github.com/saphan-ai/proven/internal/query	0.004s
ok  	github.com/saphan-ai/proven/internal/review	2.323s
?   	github.com/saphan-ai/proven/internal/service	[no test files]
ok  	github.com/saphan-ai/proven/internal/testrun	12.140s
ok  	github.com/saphan-ai/proven/internal/trace	0.027s
```

`go test -race -count=1` on `./internal/query`, `./internal/review`, `./internal/pipeline`, `./internal/fixture`, and `./cmd/proven`: exit 0. Race on `./internal/testrun` and the packages this change does not touch: UNRUN.

Local toolchain: `go1.26.8`. `go.mod` says `go 1.26` and has no toolchain line.

Testing bar: no metric is declared in this repository. UNRUN. No threshold, skip, or assertion was loosened. The import census gained one exemption, `internal/query/live.go` for `net/http`, the same shape as the generator client. `sample.go` is still refused.

`co-authored` lines in the landed commit: 0. `git commit` and `git commit --amend` both had the trailer re-added by the commit tool. The landed object was written with `git commit-tree` from the message file, then the branch was moved to it. That object has no trailer.

Session recording (ADR-0024): none in this repository. No recording was provisioned for this seat. Publishing one is not this order.

## Rulings

- ruling: the live call is `POST {ZETARIS_REST_URL}/sql-editor/sqls/run` — the 2026-10-21 measurement proved that route and did not prove `/query/sql/start` or a close route — cost if wrong: a later measurement of the start/close lifecycle would move the client.
- ruling: row limit and timeout are `ZETARIS_ROW_LIMIT` (empty means 100) and `ZETARIS_TIMEOUT` (empty means 60s) — the order named the settings and not the variables — cost if wrong: live Open refuses until the environment uses these names.
- ruling: sample evidence keeps the fixture snapshot; the live adapter sets the snapshot to the UTC time of the call — the seam's existing sample tests pin `sample-2026-10-22` — cost if wrong: a reader who wants the clock time on a sample row will not see it.
- ruling: the version guard is the `version` column of the sample advisory row (F2). The new finding filters `cve_id` on the KEV view — that view has no version column in the measurement — cost if wrong: a version outside the advisory still leaves the KEV finding in place.

## master-was-wrong

The order and ADR-0015 say a query started at `POST /query/sql/start` is always closed. The provisioned measurement does not record that route's request, response, or close call. It records `POST /sql-editor/sqls/run` (headers `Authorization: Bearer`, `X-Org-ID`, `X-Request-ID`; body `queryId`, `sql`, `source` `SqlEditor`, `limit`). The adapter uses that route. The HTTP body is closed on success, on a non-200, and when the call ends. There is no second close call.

`px_epss_cve` is frozen to CVE-2021-44228 and `px_osv_id` is frozen to GHSA-jfh8-c2jp-5v3q. A WHERE for any other id returns no row on the live platform. The KEV view is the full feed, so a WHERE on `cve_id` can match a CVE the planner named. Sample tables hold a few rows, including the fixture CVE.

The measured OSV columns are nested (`aliases`, `database_specific.severity`). The local engine has no arrays or nested fields. The sample table stores `id`, `alias`, `published`, and `severity` as plain strings.

## Blocked on / next

none for this leg. The live review against the real service is the master's, after this round. Not run here: any call to the real Zetaris.

## REDACTED

none
