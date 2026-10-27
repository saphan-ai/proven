# STATUS — the-source-adapter-interface · round 1

- state: ready-for-review
- executor: cursor on runner-cursor
- model: grok-4.7-high
- pool: not recorded
- cost: not recorded
- duration: not recorded
- last commit: 0189d9d [the-source-adapter-interface] docs: one row each for LAB and ADAPTERS
- updated: 2026-10-24T14:51:25Z

## Summary

RED: docs/README.md at 5faabc9 still has conflict markers. OSV, KEV and EPSS are source adapters; the catalogue reads that registry; docs/ADAPTERS.md has a banned-package example. The four templates' SQL is unchanged. QUERIES.md was not edited.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — the-source-adapter-interface
- state: stop-2-ready
- abstract: RED: docs/README.md at 5faabc9 still has conflict markers. OSV, KEV and EPSS are source adapters; the catalogue reads that registry; docs/ADAPTERS.md has a banned-package example. The four templates' SQL is unchanged. QUERIES.md was not edited.
- last commit: 0189d9d [the-source-adapter-interface] docs: one row each for LAB and ADAPTERS
- done: Source adapters own the catalogue, and a test-only source runs on the local store.
  - Fixed order osv, kev, epss and the four template names: TestRegistryOrderIsOSVThenKEVThenEPSS and TestCatalogueFollowsTheSourceRegistry, both inside go test -count=1 ./... exit 0.
  - The three call sites build through source.Build: internal/osv/check.go and internal/epss/epss.go. SQL text is unchanged: TestBuiltSQLIsTheTextThePipelineAlwaysSent passed in that suite.
  - A source added only by Register shows its template, runs on the local store, and its rule returns the row: TestAddedSourceShowsItsTemplateItsQueryAndItsRule, same suite, exit 0.
  - Negative control: bannedPackage{} removed from internal/source/extend_test.go line 78. go test -count=1 -v ./internal/source -run TestAddedSourceShowsItsTemplateItsQueryAndItsRule exit 1. extend_test.go:94: catalogue has no banned-by-package template. The element was restored before the suite above.
- next / blocked-on: none
- artifacts:
  - internal/source/source.go
  - internal/source/registry.go
  - internal/source/builtin_test.go
  - internal/source/extend_test.go
  - internal/querytpl/adapters.go
  - internal/querytpl/querytpl.go
  - internal/querytpl/querytpl_test.go
  - internal/osv/check.go
  - internal/epss/epss.go
  - docs/ADAPTERS.md
- evidence:
  - Baseline before edits, at 5faabc99489b979f7bc3016acd11f7b0e988bbcf: go test -count=1 ./... exit 0. Skip counts were not summed. No FAIL line was in the captured output.
  - go build ./... exit 0.
  - go vet ./... exit 0.
  - go test -count=1 ./... exit 0 after the code and the adapters page were written, before the two commits. Skip counts were not summed. No FAIL line was in the captured output. internal/source printed ok.
  - go test -race -count=1 ./internal/source ./internal/querytpl ./internal/osv ./internal/epss exit 0.
  - sh scripts/check-links.sh exit 0.
  - go version go1.26.10 linux/amd64. go.mod says go 1.26 and has no toolchain line.
  - Earlier commit 302080e4a06fd0fe33a90e3c9f70c50389362b8f [the-source-adapter-interface] source: register OSV, KEV and EPSS. Tip ef929f6aa776b880f98321a9da408c76ab349768. git log -1 showed no Co-authored-by on either.
- note: master-was-wrong: none. Readings at 5faabc9 matched: Catalogue at querytpl.go:135 with four templates, OSVPackageSQL at 354, KEVSQL at 384, EPSSSQL at 396, call sites internal/osv/check.go:207 and :297 and internal/epss/epss.go:152, tool.Adapter at internal/tool/tool.go:23, language Adapter in internal/manifest/adapter.go, store.Store at internal/store/store.go:26. RED: docs/README.md lines 17-21 are a committed conflict between main's LAB keys row and fix/a-language-is-one-adapter-php's language-adapter row. This wagon did not edit that file.
  - ruling: the three adapters are built in internal/querytpl/adapters.go and registered into internal/source, because the SQL stays beside literal and source cannot import querytpl. Cost if wrong: a reader looks for the adapter types under internal/source and finds the interface only.
  - ruling: the OSV rule is not the dependency finding decision, because it does not see the change's version. Cost if wrong: a later caller that treats the rule as that decision would report a version outside every range.
  - ruling: KEV and EPSS use the rule as the empty-row gate and still attach queryEvidence, so Bound stays set. Cost if wrong: none while the rule rejects only an empty row list, which is the previous check.
  - ruling: Register is in-process. There is no plugin file loader. Cost if wrong: an organisation expected to drop in a file.
  - ruling: the banned-package adapter is test-only, so the default catalogue stays four templates and docs/QUERIES.md is byte-identical. Cost if wrong: the worked example is not a source a review can pick.
  - ruling: Load.Snapshot is empty for OSV, KEV and EPSS because this program does not load those tables. Cost if wrong: a reader expected a fixed date.
  - testing-bar: no metric is declared, so coverage and mutation score are UNRUN. No threshold, skip or assertion was weakened.
  - Describe() was not captured before the edit. Template name, answer and format strings were copied from the previous Catalogue literals. The SQL golden test locks the query text. Full-suite -race and the lab workflow were not run.
  - A write of the commit message outside the worktree was refused (tool-before-hook-payload). The messages were written under ████████████ inside the worktree and were not committed. awk was refused (tool-command-not-in-class, allowed form none). Subject length was counted by hand. Both commits landed; git log -1 was the check.
- updated: 2026-10-24T14:38:57Z

CLOSE-OK · the-source-adapter-interface · commit 0189d9d99cb00f407b687c1137a4a6e86e393adc · branch fix/the-source-adapter-interface · worktree ████████████████████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 2 |
