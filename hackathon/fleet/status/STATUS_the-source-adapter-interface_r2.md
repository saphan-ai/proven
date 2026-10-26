# STATUS — the-source-adapter-interface · round 2

- state: ready-for-review
- executor: not recorded
- model: not recorded
- pool: not recorded
- cost: not recorded
- duration: not recorded
- last commit: 7b302eb [the-source-adapter-interface] merge: absorb main
- updated: 2026-10-25T00:11:44Z

## Summary

Round 2 lands C1-C4: Proven checks every parameter value (machine format plus byte floor) before any builder runs, Register refuses a parameter with no format and a non-plain Load.Table, a builder error becomes a refusal and a QueryNote, and the owning adapter's Rule decides evidence. Touched packages are green; seven packages plus sarifschema are UNRUN here (listener or TLS refusal) and the master runs them.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — the-source-adapter-interface
- state: stop-2-ready
- abstract: Round 2 lands C1-C4: Proven checks every parameter value (machine format plus byte floor) before any builder runs, Register refuses a parameter with no format and a non-plain Load.Table, a builder error becomes a refusal and a QueryNote, and the owning adapter's Rule decides evidence. Touched packages are green; seven packages plus sarifschema are UNRUN here (listener or TLS refusal) and the master runs them.
- last commit: 7b302eb [the-source-adapter-interface] merge: absorb main
- done: C1-C4 landed in three commits (a4ef4c7 source, 803aa59 investigator, e9c1118 docs)
  - C1 property: Proven refuses quote, NUL and newline values for a registered concatenating builder through Build and Validate, ReasonBadValue. Tests TestRegisteredTemplateValueIsHeldToItsFormatByProven, TestRegisteredTemplateFormatMismatchIsRefusedByProven, TestRegisterRefusesAParameterWithNoFormat. Before: querytpl test package did not compile (unknown field Match in source.Param), exit 1. After: ok, exit 0. Revert the CheckValue call in querytpl.Build (mutated field = parameter value): both value tests red. Remove the Match nil check in checkNew: the no-format test red.
  - C2 property: a registered builder's plain error reaches the Investigator as a *querytpl.Refusal and becomes a QueryNote. Test TestRefusalByARegisteredBuilderIsAQueryNote. Before: FAIL, the review failed on the refused value. After: ok. Revert the wrap in querytpl.Build (return err): test red. The worked example in extend_test.go and ADAPTERS.md return a *querytpl.Refusal.
  - C3 property: Register and SetBuiltins refuse a Load.Table that is not one identifier or schema.name (ValidRelation moved to source, aliased in querytpl). Test TestRegisterRefusesALoadTableThatIsNotAPlainName, mutated field = Load.Table. Before: FAIL on five tables, After: ok. Disable the check: test red.
  - C4 property: the owning adapter's Rule decides evidence on the Investigator path, landed in this stream. Test TestOwningAdaptersRuleDecidesEvidence, mutated field = the rule's verdict. Before: FAIL (evidence 1, rule not consulted). After: ok. Ignore the verdict: test red.
- note: master-was-wrong (the closed key set refuses it as a key): nothing found wrong in the order. One refinement: source.Build also checks values and returns *source.ValueError, which osv/check.go now reads beside *querytpl.Refusal, so an odd package name stays a trace note there.
- next / blocked-on: the master runs the UNRUN packages and the re-review of the fix diff.
- artifacts:
  - internal/source/source.go
  - internal/source/registry.go
  - internal/querytpl/querytpl.go
  - internal/querytpl/adapters.go
  - internal/agent/investigator/investigator.go
  - docs/ADAPTERS.md
- evidence:
  - go build ./... exit 0. go vet ./... refused for internal/cli/sarifschema (TLS: x509 OSStatus -26276 fetching jsonschema/v5); vet of the touched packages exit 0.
  - go test -count=1 ./... UNRUN, each refused by the environment: internal/cli/sarifschema [setup failed] same TLS refusal; cmd/livecheck, internal/generator, internal/guardrail, internal/pipeline, internal/query, internal/review, internal/store panic "httptest: failed to listen on a port: listen tcp6 [::1]:0: bind: operation not permitted". Rerun with -skip of the named listener tests still panics in other listener tests, so the rest of those packages is unrun, not green.
  - Every other package passed, including all packages the diff touches.
  - Not run: go test -race, scripts/check-links.sh.
- updated: 2026-10-24T17:10:00Z

## REDACTED

none
