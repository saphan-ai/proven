# STATUS — json-sarif-outputs · round 1

- state: ready-for-review
- executor: claude-code on macm5
- model: claude-sonnet-5-5
- pool: not recorded
- cost: not recorded
- duration: 11 min
- last commit: 3aabc4b [json-sarif-outputs] cli: carry text-only lines; test the contract
- updated: 2026-10-23T18:20:36Z

## Summary

Round 3 of W2 is done in 2 commits: colliding output paths are refused at load with exit 7, and C1 to C8 are done. UNRUN here, for the master: the two schema tests `TestSampleSARIFMatchesTheSchema` and `TestSARIFShapesMatchTheSchema`, because this seat's module cache has no `jsonschema/█████████`. The rest of `internal/cli` ran green with `go test -overlay`, which blanks only `format_schema_test.go`. 6 packages this round did not touch fail on sandbox-refused httptest listeners.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — json-sarif-outputs
- state: stop-2-ready
- abstract: Round 3 of W2 is done in 2 commits: colliding output paths are refused at load with exit 7, and C1 to C8 are done. UNRUN here, for the master: the two schema tests `TestSampleSARIFMatchesTheSchema` and `TestSARIFShapesMatchTheSchema`, because this seat's module cache has no `jsonschema/█████████`. The rest of `internal/cli` ran green with `go test -overlay`, which blanks only `format_schema_test.go`. 6 packages this round did not touch fail on sandbox-refused httptest listeners.
- last commit: 3aabc4b [json-sarif-outputs] cli: carry text-only lines; test the contract
- done: Collisions refused (exit 7, ADR-0013 form), the C1 to C8 asks, and a red with the check removed for each of the collision, SM2, JM1, SM1 and family-line unplugs.
  - collision: `refuseCollidingOutputs` in `internal/cli/config.go`, tested by `TestCollidingOutputPathsAreRefused` for all 6 pairs, `TestCollisionReproduction` and `TestDistinctOutputPathsAreNotCollisions`
  - C1: the five lines are in the SARIF run properties and `family_sentence` is in the JSON, proven by `TestSARIFCarriesEveryTextOnlyLine` and `TestJSONCarriesEveryTextOnlyLine`, one subtest per key
  - C2: `sarifShapes` has 8 shapes, checked by `TestSARIFShapes` (run) and by `TestSARIFShapesMatchTheSchema` (UNRUN, needs the module)
  - C3, C4, C5: `TestSARIFLevelFollowsSeverity` by ruleId, `TestJSONDocumentIsTheContract`, `TestJSONEmptyListsAreArraysNotNull`, `TestSARIFPropertiesAreTheContract`, and `docs/OUTPUT_FORMATS.md` names every key
  - C6, C7, C8: `TestOutputWriteFailureExits2`, the two `// Package cli is` comments are now plain file comments, and `artifactURI` escapes a colon in the first segment, tested by `TestArtifactURIEscapesAColonInTheFirstSegment`
- evidence:
  - git log --oneline 7281552..HEAD gives 3aabc4b [json-sarif-outputs] cli: carry text-only lines; test the contract, then 250301b [json-sarif-outputs] cli: refuse two outputs that name one file
  - git diff --stat 7281552..HEAD -- AGENTS.md internal/review is empty
  - gofmt -l . is empty. go vet -overlay (schema file blanked) ./... is clean (exit 0)
  - go test -count=1 ./internal/cli with the overlay is ok (23s). The unoverlaid run cannot compile here, setup failed, because jsonschema/v5 is missing and the proxy answered Forbidden
  - go test -count=1 ./... with the overlay is ok in 28 packages. FAIL in 6 packages this round did not touch: cmd/livecheck, internal/generator, internal/guardrail, internal/pipeline, internal/query and internal/review. Each failure is the sandbox refusing a listen (bind operation not permitted) inside an httptest server. I did not run the base to compare
  - RED 1, collision check disabled: TestCollisionReproduction got "exit 0, want 7"; all 6 TestCollidingOutputPathsAreRefused rows failed, e.g. "exit 2, want 7"
  - RED 2, SM2 (WARNING and ERROR swapped in sarifLevel): TestSARIFLevelFollowsSeverity failed with `result w: level "error", want "warning"` and `result e: level "warning", want "error"`
  - RED 3, JM1 (the json tag rule_id renamed to ruleId): TestJSONDocumentIsTheContract failed with "document differs from the contract"
  - RED 4, SM1 (SARIF properties.evidence dropped): TestSARIFPropertiesAreTheContract failed with "SARIF evidence differs from the JSON evidence", and TestSARIFShapes/no_evidence failed
  - RED 5, family_sentence dropped from the SARIF run properties: TestSARIFCarriesEveryTextOnlyLine/family_sentence failed with `run properties have no "family_sentence"`
  - every unplug was reverted, and the final tree is the committed one
  - laws echoed: operator-paste-law, command-delivery-law, commands-anchored-in-canonical-root, dogfood-gate-sequence, sacred-injection-hierarchy, red-escalation-law, fix-or-shout-law, pre-flight-budget-check, backend-routing-by-work-class, plan-default-opus-max, law-at-act, literal-executability, owner-next-block, gate-dual-write, verify-before-assert, cap-is-2x-estimate, fresh-checkout-law, full-surface-verify, commandment-01-archive-before-teardown, worktree-fork-from-current-HEAD, amendment-edits-the-text-it-amends, law-refresh-has-a-dod, executor-toolchain-denial, quarantine-on-irregular-ingest, remediation-is-a-projection, addressee-pure-queue, negative-control-in-DoD, fresh-run-over-resume, review-is-named-where-the-record-is-touched, reveal-on-act, channel-namespace-files-only, guard-outlives-what-it-guards, bounce-your-own-gate-first, watcher-scope-is-your-own, environment-remediation-is-offered-measured-and-recorded, state-machine-is-projected-not-read, remote-spawn-instruction-is-served, completion-signal, human-ceremony-must-live-in-docs, language-of-the-fleet, unplug-test, lanes-are-mechanical, source-repo-is-the-only-assumed-repo, an-order-must-be-able-to-refute-its-author, placement-is-derived-not-named, a-chooser-prints-its-arithmetic, balance-is-per-resource-budget-never-rotation, one-authoring-shape-regardless-of-where-it-runs, a-priced-act-never-runs-implicitly, a-precondition-states-its-own-applicability, a-seat-declares-what-its-work-costs-and-nothing-infers-it, a-reading-carries-its-own-reach, review-return-refusals, a-compacted-master-re-reads-its-track-order, self-describing-production-refusal, security-orders-name-the-defect, a-denied-act-stays-denied-in-every-spelling, a-known-next-act-is-performed-not-announced, a-head-refuses-work-outside-its-track
- note: master-was-wrong: nothing in the order was contradicted. Two measured differences: this seat cannot run the schema tests (no `jsonschema/v5` in the module cache), and `mv` and `rmdir` are denied for the executor class, so I did not move files aside and used `go test -overlay` for local runs.
- artifacts:
  - internal/cli/config.go
  - internal/cli/outputs_test.go
  - internal/cli/format_json.go
  - internal/cli/format_sarif.go
  - internal/cli/format_sarif_test.go
  - ██████████████████████████████████
  - internal/cli/output_contract_test.go
  - docs/OUTPUT_FORMATS.md
- next / blocked-on: the master runs `go test -count=1 -run 'TestSampleSARIFMatchesTheSchema|TestSARIFShapesMatchTheSchema' -v ./internal/cli/` on a seat whose cache holds `jsonschema/█████████`, and runs the full suite there.
- updated: 2026-10-24T00:00:00Z

CLOSE-OK · json-sarif-outputs · commit 3aabc4ba11165cf32970632652051e3cccfb5f2d · branch fix/json-sarif-outputs · worktree ██████████████████████████████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| e-mail or user name | 3 |
| file and line in Saphan Studio's private code | 1 |
