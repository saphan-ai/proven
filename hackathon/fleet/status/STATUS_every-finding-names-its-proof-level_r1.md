# STATUS — every-finding-names-its-proof-level · round 1

- state: ready-for-review
- executor: not recorded
- model: not recorded
- pool: not recorded
- cost: not recorded
- duration: not recorded
- last commit: 1c09608 [every-finding-names-its-proof-level] absorb main (the master: clean merge)
- updated: 2026-10-24T02:29:23Z

## Summary

review.Finding now carries ProofLevel (4 named values, strongest first; test and data are set today, tool and two-model values exist for ADR-0040/0041). review.Result.ProofCounts counts by level; a finding with no evidence level is refused in fill. Sample review: 1 test, 0 tool, 3 data, 0 agreed. Docs updated. Six listener-bind tests cannot run in this sandbox (not caused by this change).

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — every-finding-names-its-proof-level
- state: stop-2-ready
- abstract: review.Finding now carries ProofLevel (4 named values, strongest first; test and data are set today, tool and two-model values exist for ADR-0040/0041). review.Result.ProofCounts counts by level; a finding with no evidence level is refused in fill. Sample review: 1 test, 0 tool, 3 data, 0 agreed. Docs updated. Six listener-bind tests cannot run in this sandbox (not caused by this change).
- last commit: 1c09608 [every-finding-names-its-proof-level] absorb main (the master: clean merge)
- done: ProofLevel type, ProofLevelOf, Finding.ProofLevel, Result.ProofCounts, refusal of a levelless finding, sample totals, docs
  - TestProofLevelOfEachKind, TestProofLevelOfRefusesNoEvidence, TestFillRefusesAFindingWithNoLevel, TestFillCountsByLevel, TestSampleReviewNamesEachFindingsLevel pass (go test -run, exit 0)
  - negative control (law #27): mapping EvidenceTest to ProvedByData turned TestProofLevelOfEachKind, TestFillCountsByLevel and TestSampleReviewNamesEachFindingsLevel red; restored, green
  - go build ./... and go vet ./... exit 0; gofmt -l prints nothing
- next / blocked-on: none
- note: master-was-wrong: the master's reading said the sample review has findings F1, F2, F4; it also has the OSV finding osv:evil.example/leftpad, so the sample totals are 1/0/3/0, not 1/0/2/0 (measured by running the sample review). Also, fill had no error path; it now returns (Result, error).
- evidence:
  - FIELD FOR THE SHELL TRAIN: render review.Finding.ProofLevel (string values proved-by-test, proved-by-tool, proved-by-data, agreed-by-two-models-not-proven) first in text, Markdown, HTML and the PR comment, as a field in JSON and SARIF; render review.Result.ProofCounts (map[ProofLevel]int, every level present) as the totals; review.ProofLevels gives the order. Noted in docs/ARCHITECTURE.md.
  - go test -count=1 ./... exit 1: FAIL in cmd/livecheck, internal/generator, internal/guardrail, internal/pipeline, internal/query, internal/review; every one is "httptest: failed to listen on a port: bind: operation not permitted" (sandbox denies listeners). All other packages ok (cmd/proven, internal/cli, demo, agent/*, change, etc.). The panic aborts internal/review, so tests after it did not run there; the review tests I ran by -run (proof, sample, open) pass. Base was not re-run; the failing tests listen on a port and do not touch this change.
  - commit: git commit printed packed-refs.lock Operation not permitted; git log -1 shows the commit landed. No Co-authored-by line.
  - session recording: none available to this leg (headless, no recorder provided).
- artifacts:
  - internal/review/review.go
  - internal/review/open.go
  - internal/review/prooflevel_test.go
  - internal/review/prooflevel_sample_test.go
  - docs/ARCHITECTURE.md
  - docs/SAMPLE_MODE.md
- updated: 2026-10-24T00:00:00Z

CLOSE-OK · every-finding-names-its-proof-level · commit c5747573d320a8535476a749e0dded6a7e502eca · branch fix/every-finding-names-its-proof-level · worktree ███████████████████████████████████████████████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
