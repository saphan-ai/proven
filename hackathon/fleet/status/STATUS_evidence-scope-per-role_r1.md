# STATUS — evidence-scope-per-role · round 1

- state: ready-for-review
- executor: not recorded
- model: not recorded
- pool: not recorded
- cost: not recorded
- duration: not recorded
- last commit: d1c59d0 [evidence-scope-per-role] investigator: owed test passes the scope
- updated: 2026-10-25T07:14:26Z

## Summary

Proven now scopes evidence per agent role: proven.yaml keys investigator_templates and skeptic_templates narrow what each of the two querying roles is told of and may ask.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — evidence-scope-per-role
- state: stop-2-ready
- abstract: Proven now scopes evidence per agent role: proven.yaml keys investigator_templates and skeptic_templates narrow what each of the two querying roles is told of and may ask.
  A request outside the scope is refused by name before any store is asked and becomes a query note; the review goes on. Unset keys keep every template.
  An unknown template, a duplicate and an empty list are refused on load. docs/CONFIGURATION.md documents both keys.
  Build and vet are green; tests are green except packages that need a local listener or a module download, named UNRUN in the note.
- last commit: d1c59d0 [evidence-scope-per-role] investigator: owed test passes the scope
- done: Per-role evidence scope built and proven (ADR-0013, ADR-0032).
  - item 1 (keys + docs): internal/review/config.go keys investigator_templates, skeptic_templates; docs/CONFIGURATION.md rows and example; TestEvidenceScopeKeysLoad, TestEvidenceScopeIsRefusedOnLoadByName
  - item 2 (prompt lists only the scope): TestPromptListsOnlyTheScopesTemplates in investigator and skeptic, TestScopeDescribeListsOnlyItsTemplates in querytpl
  - item 3 (refused before any store, a query note): TestInvestigateRefusesOutOfScopeBeforeAnyStore, TestExecuteRefusesOutOfScopeBeforeAnyStore, TestSkepticRequestOutsideItsScopeAsksNoStore
  - item 4 (Skeptic narrower than the Investigator): TestSkepticIsNarrowerThanTheInvestigator in internal/pipeline/evidence_scope_test.go
- next / blocked-on: none
- artifacts:
  - docs/CONFIGURATION.md
  - internal/querytpl/scope.go
  - internal/agent/investigator/investigator.go
  - internal/agent/skeptic/skeptic.go
  - internal/pipeline/pipeline.go
  - internal/review/config.go
  - internal/review/open.go
- evidence:
  - ACCEPTANCE 1 reproduce: at base b17d30f a Skeptic request for epss-by-cve reached the store. Test run, pasted: "zz_repro_test.go:26: store asked: [epss-by-cve]" with PASS, saved at ████████████████████████████████████████████████████ The throwaway test became evidence_scope_test.go (renamed, not deleted, because rm is denied to this role).
  - ACCEPTANCE 2: go build ./... green; go vet ./... green except internal/cli/sarifschema (setup failed, quoted below); go test -count=1 ./... run once, every failure is a listener refusal or the module download.
  - UNRUN, quoted refusal for each: internal/cli/sarifschema "Get https://proxy.golang.org/github.com/santhosh-tekuri/jsonschema/v5/@v/v5.3.1.zip: tls: failed to verify certificate" (setup failed); cmd/livecheck, internal/generator, internal/guardrail, internal/query, internal/store "httptest: failed to listen on a port: listen tcp6 [::1]:0: bind: operation not permitted".
  - internal/pipeline and internal/review also panic on that listener refusal in some tests. Rerun with each panicking test skipped (script ████████████████████████████████████████████████, results in pipeline-run.txt and review-run.txt): pipeline ok (43.8s) with the 8 skipped tests listed there as UNRUN. internal/review: the skip loop had not finished when this status was written (each pass costs minutes; listener-refused so far, UNRUN: TestReviewEvidenceNamesTheFormTheStoreRan, TestLiveReviewCarriesTheRoutedModel, TestLiveReviewRefusesANonNVIDIAModelByName, TestUnauthorizedReviewDegrades, TestClosedEndpointDegradesTheReview, TestLiveModeWithoutRuntimeNamesProverOff); what was run and green there: go test -count=3 -run 'EvidenceScope|Config' ./internal/review ok, and the full package up to its first listener panic ran without a scope-related failure. The rest of internal/review is UNRUN, not claimed green.
  - NEG CONTROL 3a (item 3, law 27): mutated internal/agent/investigator/investigator.go line 78 (scope.Check in Execute replaced by error(nil)): TestExecuteRefusesOutOfScopeBeforeAnyStore, TestSkepticRequestOutsideItsScopeAsksNoStore, TestSkepticIsNarrowerThanTheInvestigator red (exit 1), restored green (exit 0). Mutated line 113 (scope.Check in Investigate): TestInvestigateRefusesOutOfScopeBeforeAnyStore red (exit 1), restored green. Output in neg-a1.txt, neg-a2.txt.
  - NEG CONTROL 3b (item 2, law 27): mutated internal/querytpl/scope.go line 56 (describe(s.Catalogue()) replaced by describe(Catalogue())): TestScopeDescribeListsOnlyItsTemplates, and TestPromptListsOnlyTheScopesTemplates in investigator and skeptic red (exit 1), restored green (exit 0). Output in neg-b.txt.
  - gofmt -l internal cmd prints nothing.
- note: master-was-wrong (the order asks for this as its own key, which the exec-status grammar refuses as doc-key-unknown, so it rides here): nothing refuted. Re-measured at base b17d30f: only the Investigator (investigator.Investigate) and the Skeptic (investigator.Execute from runRefutation) reach a store, and there was no per-role key. One refinement: the Skeptic's query runs through investigator.Execute, so the Skeptic's check lives there, not in skeptic.Attempt, which only shapes the prompt. Judgement calls (C6). (1) Key names: the file is one flat mapping with no nested values, so the scope is two flat keys with comma lists, not a nested evidence_scope mapping; cost if wrong: a rename. An unknown role is therefore refused as an unknown key, with the allowed keys listed. (2) The sample scenario narrows only the Skeptic, because the fixture's recorded answers key on the unscoped Investigator prompt; the Investigator's narrowing is proven by unit tests. (3) A template name not on the list is refused as unknown, not out of scope, and the name is not repeated because it is model text. (4) skeptic.Attempt and investigator.Prompt changed signature to take a scope; every caller and test was updated, nil means every template. Commit has no Co-authored-by line.- updated: 2026-10-25T06:30:00Z

tip_moved_by_hand: true
moved_by: ██████████████
supersedes_commit: ee0394d3305728f6b67fc7ffe66cf2e5b3ca6bdc

## REDACTED

| what | times |
|---|---|
| absolute or private path | 2 |
| seat identity | 1 |
