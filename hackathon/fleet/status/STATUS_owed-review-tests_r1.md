# STATUS — owed-review-tests · round 1

- state: ready-for-review
- executor: cursor on runner-cursor
- model: grok-4.7-high
- pool: not recorded
- cost: not recorded
- duration: not recorded
- last commit: 53e365e [owed-review-tests] merge: absorb main
- updated: 2026-10-25T05:38:54Z

## Summary

Pinned the checks W43, W47, W49 and W52 left owed, labelled a miss note with its query form, and bound a raised PyPI name on the normalised spelling.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — owed-review-tests
- state: stop-2-ready
- abstract: Pinned the checks W43, W47, W49 and W52 left owed, labelled a miss note with its query form, and bound a raised PyPI name on the normalised spelling.
- last commit: 53e365e [owed-review-tests] merge: absorb main
- done: The owed tests and the two named fixes are on fix/owed-review-tests
  - go build ./... exited 0
  - go vet ./... exited 0
  - go test -count=1 ./... exited 0, with no listener refusal
  - 095fee524920a2fb4c7e189e882d7f5b25353306 pins source.Build, the kept row copy, and the byte floor
  - 975be987e4dd0c1e854072733602d05112cb4aa8 pins http://user@host and keeps the defensive blanking
  - af17eec85c03b54ae7092b8d2945e9a75870b35a pins the down-store refusal, the zetaris form seam, the query-form label, and deletes the empty seam file
  - ee2c8b4f9f26c745f062564c63d4af9730bd7043 pins every catalogue template, proved-by-data on the malicious-package fixture, and the normalised PyPI bind key
- next / blocked-on: none
- note: master-was-wrong: none; the sites are the named checks, and R4's two prints stood at the no-rows note and missNote (lines 262 and 390 before this change), not 237 and 365
  - Packagist is folded to lower case on the same key, because the query key already does that; a mixed-case Packagist name now binds
  - sample goldens did not move: cmd/proven, the formatter golden and guardcount exited 0, so nothing was regenerated
  - testing-bar: no metric is declared in the order, so the comparison is UNRUN
  - race detector UNRUN: the acceptance command has no -race, and the edits add no goroutine
- evidence:
  - W47-O1 TestBuildRefusesAValueBeforeTheBuilderRuns — CheckValue loop in source.Build (registry.go 119-123) — before exit 1, after exit 0
  - W47-O2 TestRuleKeepsOneRowOfTwoOnTheEvidence — res.Columns, res.Rows = kept.Columns, kept.Rows (investigator.go 146) — before exit 1, after exit 0
  - W47-O3 TestRegisteredTemplateValueIsHeldToItsFormatByProven — byte floor c > 0x7e and c == '\\' in CheckValue (source.go 71) — before exit 1, after exit 0
  - W49-RC2 TestLoadCouchEmptyIsADownStoreAndWrongIsRefused — @ blanking of the address value (couchsettings.go) — before exit 1, after exit 0
  - W52-R1 TestDownStoreRefusesBeforeItNamesItselfDown — BuildForms call in downStore.Run (open.go 296) — before exit 1, after exit 0
  - W52-R3 TestReviewEvidenceNamesTheFormTheStoreRan/zetaris — store.NewZetaris(zset) (open.go 172) — before exit 1, after exit 0
  - W52-R4 TestMissNoteNamesTheQueryForm — queryText label in the no-rows note and missNote — before exit 1, after exit 0
  - W52-R5 deletion of ████████████████████████████████████ — the file held no test, so no line turns red; the suite exited 0 with the file and exited 0 after git rm
  - W43-C1 TestStoresAgreeOnAMissingTable — the request list cut to one catalogue template — before exit 1, after exit 0
  - W43-C2 TestMaliciousPackageFixtureIsProvedByData — EvidenceQuery mapped to ProvedByTool in ProofLevelOf — before exit 1, after exit 0
  - W43-O2 TestBindsKeysBothSidesOnTheNormalisedPyPIName — Binds keyed on the raw package spelling — before exit 1, after exit 0
- artifacts:
  - internal/source/build_value_test.go
  - internal/querytpl/registered_test.go
  - internal/agent/investigator/investigator_test.go
  - internal/store/couchsettings.go
  - internal/store/settings_test.go
  - internal/review/downstore_test.go
  - internal/review/form_seam_test.go
  - internal/osv/check.go
  - internal/osv/check_test.go
  - internal/store/missing_test.go
  - internal/review/malicious_level_test.go

CLOSE-OK · owed-review-tests · commit 53e365e858b339b65be14f583e743eab33cc5a46 · branch fix/owed-review-tests · worktree █████████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| file and line in Saphan Studio's private code | 1 |
