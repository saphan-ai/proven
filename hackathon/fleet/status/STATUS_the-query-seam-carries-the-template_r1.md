# STATUS — the-query-seam-carries-the-template · round 1

- state: ready-for-review
- executor: cursor on runner-cursor
- model: grok-4.7-high
- pool: not recorded
- cost: not recorded
- duration: 27 min
- last commit: 04172a7 [the-query-seam-carries-the-template] merge: absorb main
- updated: 2026-10-25T03:46:20Z

## Summary

Live reviews pass a template request to the configured store. CouchDB evidence names mango and its rows; the local store names sql. Sample reviews are unchanged.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — the-query-seam-carries-the-template
- state: stop-2-ready
- abstract: Live reviews pass a template request to the configured store. CouchDB evidence names mango and its rows; the local store names sql. Sample reviews are unchanged.
- last commit: 04172a7 [the-query-seam-carries-the-template] merge: absorb main
- done: The live seam passes a template request and the evidence names the form that ran
  - base reproduction exit 1: couchdb down, reason "the review's query seam sends SQL and CouchDB runs a Mango form; the seam is not moved to templates yet", findings 0
  - after the fix, TestReviewEvidenceNamesTheFormTheStoreRan names mango and row GHSA-1 on couchdb, and sql and the same row on the local store
  - negative control mutated internal/store/couchdb.go:191 Form from querytpl.FormMango to querytpl.FormSQL; the couchdb case exited 1 with form "sql"; the line was restored to FormMango
  - go build ./... exit 0; go vet ./... exit 0; go test -count=1 -timeout 300s ./... exit 0
  - go test -race -count=1 -run 'TestReviewEvidenceNamesTheFormTheStoreRan|TestEmptyLocalStore|TestLocalFileInside|TestCouchDB' ./internal/review/ exit 0
- next / blocked-on: named review of this stream
- artifacts:
  - internal/review/form_seam_test.go
  - internal/review/open.go
  - internal/review/review.go
  - ████████████████████████████████████
  - internal/pipeline/pipeline.go
  - internal/query/query.go
  - internal/query/sample.go
  - internal/agent/investigator/investigator.go
  - internal/osv/check.go
  - internal/epss/epss.go
  - internal/store/local.go
  - internal/store/settings.go
  - docs/ADAPTERS.md
  - docs/CONFIGURATION.md
  - ███████████████████████████████████████████████████████████████
- evidence:
  - go version go1.26.10 linux/amd64; go.mod pins go 1.26
  - base command: go test -count=1 -timeout 120s -run TestReviewEvidenceNamesTheFormTheStoreRan ./internal/review/ exit 1
  - failure: no dependency finding; failed service couchdb, reason the review's query seam sends SQL and CouchDB runs a Mango form; the seam is not moved to templates yet
  - go build ./... exit 0
  - go vet ./... exit 0
  - go test -count=1 -timeout 300s ./... exit 0, including cmd/proven
  - race subset of ./internal/review/ exit 0; full-module -race UNRUN
- note: master-was-wrong: none. The three readings held at the base (SQL-only Adapter, Open installing lazyLocal and couchdbSQL, registered templates without a mango form). Sample output was not changed; cmd/proven passed. Local.Query was removed because the pipeline no longer calls it; the missing-table proof stays on Run. A down store still refuses a bad template before it names itself down. ████████████████████████████████████ could not be deleted (rm denied, Delete rejected) and now only records that the old seam test is gone. No testing-bar metric is declared. No Co-authored-by on the commit.
- updated: 2026-10-25T01:50:35Z

CLOSE-OK · the-query-seam-carries-the-template · commit 04172a7f3b63cb3caf4fc9714fc7aa57cfd449bb · branch fix/the-query-seam-carries-the-template · worktree ███████████████████████████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 2 |
| file and line in Saphan Studio's private code | 2 |
