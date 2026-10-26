# STATUS — more-evidence-through-zetaris · round 1

- state: ready-for-review
- executor: claude-code
- model: claude-sonnet-5-5
- pool: not recorded
- cost: not recorded
- duration: 15 min
- last commit: 964b683 [more-evidence-through-zetaris] sources: ADR-0039 evidence as templates
- updated: 2026-10-24T03:55:15Z

## Summary

Four ADR-0039 sources (malicious packages, project health, CVSS beside EPSS, SPDX licences) are read from Zetaris views through four templates with strict formats and settings; each has a sample table and scenario. A missing table is a named failed service, never an empty answer. Licences carry no verdict: ADR-0036's policy is not in this code. Build and vet are green; one pipeline test fails at base and after, same cause (see note).

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — more-evidence-through-zetaris
- state: stop-2-ready
- abstract: Four ADR-0039 sources (malicious packages, project health, CVSS beside EPSS, SPDX licences) are read from Zetaris views through four templates with strict formats and settings; each has a sample table and scenario. A missing table is a named failed service, never an empty answer. Licences carry no verdict: ADR-0036's policy is not in this code. Build and vet are green; one pipeline test fails at base and after, same cause (see note).
- last commit: 964b683 [more-evidence-through-zetaris] sources: ADR-0039 evidence as templates
- done: Four sources built, tested and documented in one commit; every acceptance item except a fully green go test is met, and that one failure is not from this change.
  - 1 template per source with strict formats and settings: █████████████████████████████████ (builds, refuses a bad value, refuses a bad relation)
  - 2 sample tables and one scenario per source: ███████████████████████████████████ (malicious-package, weak-health, licence, and CVSS in osv-go-module)
  - 3 a missing table is a named failed service: TestADroppedSampleTableNamesItsSourceDown and TestMissingTableIsANamedDownSource
  - 4 CVSS beside EPSS: TestCVSSIsShownBesideEPSS checks the CVSS line is the line after the EPSS line
  - negative control (law 27): mutated field is the sample table named px_evidence.px_osv_malicious in tables.sql of malicious-package (likewise px_scorecard_health, px_spdx_licence, px_nvd_cvss), dropped by dropTable; the source is named down; with the missing-table branch of ask() made silent the test went red for 3 of 4 cases and the sources test went red; restored, green
- next / blocked-on: none
- artifacts:
  - ███████████████████████████
  - ████████████████████████
  - ████████████████████████████
  - internal/querytpl/querytpl.go
  - ███████████████████████████████████
  - docs/QUERIES.md
  - docs/SAMPLE_MODE.md
- evidence:
  - baseline at base, go test -count=1 ./... exit 1: all packages ok except internal/pipeline, FAIL TestMaterializedTreesAreReadableByOthers (go.mod mode 664 is writable by group or others)
  - final, go build ./... exit 0
  - final, go vet ./... exit 0
  - final, gofmt -l . prints nothing
  - final, go test -count=1 ./... exit 1: 38 packages ok, 1 FAIL internal/pipeline, the same single test as the baseline, 0 new failures
  - also run once each: go run ./cmd/settings-doc exit 0 (regenerated docs/CONFIGURATION.md, 4 rows)
  - the session recording: none, this is a one-shot headless run and no recording tool was available (ADR-0024)
  - not run: no live Zetaris, no docker, no push
- note: master-was-wrong: (1) the master said sample tables are built by the fixture SQL in internal/fixture; they are each fixture's tables.sql, loaded by query.OpenSample, and internal/fixture only checks the file exists. (2) Not in the readings: the recorded model prompts in testdata/fixture/discount and refused carry the template list word for word, so adding templates broke 12 scan tests until those prompts were extended; answers are unchanged. (3) The remaining go test failure is not from this change: TestMaterializedTreesAreReadableByOthers fails at base 8770336 too, because this host creates files mode 664 (umask 002). Confirmed by running the suite before any edit. The ADR-0036 reading was right: no policy in code, so licences have no verdict. Judgement calls: health and licence run on added packages only, malicious on added and raised; a release under 14 days old is recent (RecentReleaseDays); a Scorecard score of -1 is never read as bad; a Zetaris call that fails ends the sources pass and is named zetaris, and a source pass is skipped when zetaris is already named down. The names px_osv_malicious, px_scorecard_health, px_nvd_cvss and px_spdx_licence are guesses for the Zetaris track to confirm or override by setting. Live Zetaris cannot tell a missing table from another HTTP error, so there it is named zetaris with its status.
- updated: 2026-10-24T04:30:00Z

CLOSE-OK · more-evidence-through-zetaris · commit 964b683 · branch fix/more-evidence-through-zetaris · worktree █████████████████████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| file and line in Saphan Studio's private code | 6 |
