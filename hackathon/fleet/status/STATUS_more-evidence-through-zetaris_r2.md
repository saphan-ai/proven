# STATUS — more-evidence-through-zetaris · round 2

- state: ready-for-review
- executor: cursor on runner-cursor
- model: grok-4.7-high
- pool: not recorded
- cost: not recorded
- duration: 24 min
- last commit: fb608bf [more-evidence-through-zetaris] merge: absorb main
- updated: 2026-10-25T03:53:24Z

## Summary

The four ADR-0039 sources are built-in adapters with templates, rules and sample scenarios. The named review is next.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — more-evidence-through-zetaris
- state: stop-2-ready
- abstract: The four ADR-0039 sources are built-in adapters with templates, rules and sample scenarios. The named review is next.
- last commit: fb608bf [more-evidence-through-zetaris] merge: absorb main
- done: Four ADR-0039 sources are built-in adapters beside OSV, KEV and EPSS
  - mal, scorecard, cvss and spdx register from internal/querytpl/adapters.go; go test -count=1 ./internal/querytpl ./internal/source exited 0
  - sample scenarios malicious-package, project-health, cvss-score and spdx-licence each keep a finding; go test -count=1 -run TestMaliciousPackageReportsTheAdvisory|TestProjectHealthReportsNoMaintainers|TestCVSSScoreSitsBesideEPSS|TestSPDXLicenceReportsADisallowedLicence ./cmd/proven exited 0
  - a missing table is a named down source on local, zetaris and couchdb, including mal-by-package; go test -count=1 -run TestStoresAgreeOnAMissingTable ./internal/store exited 0
  - negative control mutated internal/querytpl/adr0039.go line 90, the empty-list return in malAdapter.Rule, from return source.Evidence{}, false to return copyEvidence(in), true; TestMaliciousPackageReportsTheAdvisory exited 1 with "an empty row list was evidence"; the line was restored
- next / blocked-on: none
- artifacts:
  - cmd/proven/scenario_test.go
  - docs/ADAPTERS.md
  - docs/QUERIES.md
  - internal/osv/check.go
  - internal/osv/check_test.go
  - internal/querytpl/adapters.go
  - internal/querytpl/adr0039.go
  - internal/querytpl/adr0039_test.go
  - internal/querytpl/forms_test.go
  - internal/querytpl/mango.go
  - internal/querytpl/querytpl.go
  - internal/querytpl/querytpl_test.go
  - internal/source/builtin_test.go
  - internal/source/source.go
  - internal/store/missing_test.go
  - testdata/fixture/cvss-score/base/go.mod
  - testdata/fixture/cvss-score/change/go.mod
  - testdata/fixture/cvss-score/fixture.json
  - testdata/fixture/cvss-score/tables.sql
  - testdata/fixture/discount/fixture.json
  - testdata/fixture/fix/fixture.json
  - testdata/fixture/malicious-package/base/go.mod
  - testdata/fixture/malicious-package/change/go.mod
  - testdata/fixture/malicious-package/fixture.json
  - testdata/fixture/malicious-package/tables.sql
  - testdata/fixture/project-health/base/go.mod
  - testdata/fixture/project-health/change/go.mod
  - testdata/fixture/project-health/fixture.json
  - testdata/fixture/project-health/tables.sql
  - testdata/fixture/refuted/fixture.json
  - testdata/fixture/spdx-licence/base/go.mod
  - testdata/fixture/spdx-licence/change/go.mod
  - testdata/fixture/spdx-licence/fixture.json
  - testdata/fixture/spdx-licence/tables.sql
- evidence:
  - go version printed go1.26.10; go.mod pins go 1.26
  - baseline go test -count=1 ./... on 431f3c4 exited 0
  - go build ./... exited 0
  - go vet ./... exited 0
  - go test -count=1 ./... on the tip exited 0
  - negative control of TestMaliciousPackageReportsTheAdvisory exited 1, then the line was restored and the suite exited 0
- note: master-was-wrong: git cat-file -t 964b683 failed, not a valid object name, so the adapters were not ported from that commit. The order names base fa21bc8; the worktree tip at the start was 431f3c4, which contains fa21bc8. The interface claim held: Adapter is ID, Load, Templates, Rule, registered from internal/querytpl/adapters.go. Rulings: the four table names are the adapters' own, not settings; the SPDX permissive set (MIT, Apache-2.0, BSD-2-Clause, BSD-3-Clause, ISC, 0BSD, Unlicense) stands in for an organisation policy and the cell is compared whole; a release is only days old when release_age_days is under 14; CVSS binds only a CVE the dependency check already recorded, and EPSS is attached by the existing score step. Existing scenario prompts were updated to the new catalogue sentence; their recorded answers were not. Testing bar: no metric declared, UNRUN. Race detector UNRUN: the acceptance command is go test -count=1 and this change adds no goroutine. The closed status keys have no master-was-wrong field, so this note carries it.
- updated: 2026-10-25T01:49:09Z

CLOSE-OK · more-evidence-through-zetaris · commit fb608bf7034e79ffc4aa216680c6ab00f41c3e23 · branch fix/more-evidence-through-zetaris · worktree █████████████████████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
