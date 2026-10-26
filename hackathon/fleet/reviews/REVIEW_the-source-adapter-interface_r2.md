# REVIEW — the-source-adapter-interface · round 2

- verdict: accept-with-conditions
- reviewer: claude-code on macm5, claude-opus-5-5
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-24T23:29:06Z

## Summary

claude-opus-5-5 re-review of W47 round 2 (0189d9d..a56a7e8): C1–C4 all ADDRESSED; mutants M1 and M3–M8 each turn a named test red.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — the-source-adapter-interface

Reviewer: claude-opus-5-5, headless, named re-review of W47 round 2 (R14: the round-1 findings and the fix diff only).
HEAD measured `a56a7e8` (ROLE-LOCK satisfied). Fix diff `0189d9d..a56a7e8`.

### Ground
- Round-2 commits under judgement: `a4ef4c7` (source: Proven checks each parameter), `803aa59` (investigator: the
  owner's rule decides), `e9c1118` (docs: parameter formats and the rule path), `a56a7e8` (absorb main).
- `177bba6` touches `docs/README.md` only. The absorb `a56a7e8` brings in main's own content (formatter example,
  SECURITY.md line, ADAPTERS.md "Adding an output format"). That content is main's and is NOT judged here (R14).
- Code read: `internal/source/source.go` (Param.Match, CheckValue, ValueError, ValidRelation),
  `internal/source/registry.go` (Build, checkNew), `internal/querytpl/querytpl.go` (Build, Validate),
  `internal/querytpl/adapters.go` (built-in Match patterns, Rules), `internal/agent/investigator/investigator.go`
  (Investigate, Execute), `internal/osv/check.go:201-330` (the only direct caller of `source.Build`).
- Commands run in this sandbox (2026-10-25):
  - `go build ./...`: green.
  - `go vet` over every package except `internal/cli/sarifschema`: green.
  - `go test -count=1` over every package except `internal/cli/sarifschema`: every package touched by the diff
    is green (`querytpl`, `source`, `agent/investigator`, `osv`), plus every other package that does not bind a
    listener. The packages that panicked on the sandbox's listener refusal are listed under Unrun.

### Differential
| Mutation (one at a time, restored with `git checkout --`) | Site | Result |
|---|---|---|
| M1 the per-parameter check in `querytpl.Build` disabled | `internal/querytpl/querytpl.go:258` | RED: `TestRegisteredTemplateValueIsHeldToItsFormatByProven`, `TestRegisteredTemplateFormatMismatchIsRefusedByProven` |
| M2 the per-parameter check in `source.Build` disabled | `internal/source/registry.go:119-123` | SURVIVES (only the sandbox `httptest` panic in `pipeline`, which is unrelated) |
| M3 the `*source.ValueError` branch in osv disabled | `internal/osv/check.go:221-224` | RED: `TestCheckSendsNoQueryForAPackageOutsideItsFormat` |
| M4 the byte floor in `CheckValue` disabled | `internal/source/source.go:70-74` | RED: quote, NUL and newline cases, through both Build and Validate (6 errors) |
| M5 the `Param.Match == nil` refusal in `checkNew` disabled | `internal/source/registry.go:179-183` | RED: `TestRegisterRefusesAParameterWithNoFormat` |
| M6 the `Load.Table` `ValidRelation` refusal in `checkNew` disabled | `internal/source/registry.go:162-164` | RED: `TestRegisterRefusesALoadTableThatIsNotAPlainName` (all 5 cases) |
| M7 a builder's plain error returned raw, not as a `*Refusal` | `internal/querytpl/querytpl.go:278` | RED: `TestRefusalByARegisteredBuilderIsAQueryNote` ("the review failed on a refused value") |
| M8 the owning adapter's `Rule` not consulted | `internal/agent/investigator/investigator.go:132` | RED: `TestOwningAdaptersRuleDecidesEvidence` ("sent 1, evidence 1") |

This confirms the master's claim (law #44) that disabling the check in `querytpl.Build` turns
`TestRegisteredTemplateValueIsHeldToItsFormatByProven` red. Measured again here: M1, red.

### Unplug
- C1: unplug Proven's check (M1) or the floor (M4). The registered concatenating builder then receives
  `x' OR '1'='1`, NUL or a newline, and `TestRegisteredTemplateValueIsHeldToItsFormatByProven` turns red on Build
  and on Validate. Unplug the `Register` refusal (M5) and `TestRegisterRefusesAParameterWithNoFormat` turns red.
- C2: unplug the wrapping (M7) and `TestRefusalByARegisteredBuilderIsAQueryNote` turns red, because the review fails.
- C3: unplug the table check (M6) and `TestRegisterRefusesALoadTableThatIsNotAPlainName` turns red.
- C4: unplug the rule call (M8) and `TestOwningAdaptersRuleDecidesEvidence` turns red.
- Not observable: the duplicate check in `source.Build` (M2). See Remainder O1.

### Closure
- **C1 (F1): ADDRESSED.** `querytpl.Build` holds every declared parameter to `Param.Match` and to the byte floor
  of `source.CheckValue` before `spec.Build` runs (`querytpl.go:255-261`), for built-in and registered templates
  alike. `Validate` is `Build` with empty relations (`querytpl.go:285-288`), so it refuses the same values.
  `checkNew` refuses a param with `Match == nil` (`registry.go:179-183`), and `CheckValue` also refuses when
  `Match` is nil. The floor does not depend on the adapter's pattern. The DoD test holds the format at
  `(?s)^.*$` and mutates only the value: quote, NUL, newline. Its red at `0189d9d` is a compile failure, because
  `Param.Match` did not exist there. M1 and M4 are the behavioural red. Built-in params received `Match` patterns
  (`adapters.go`). `anyPackageFormat` covers every ecosystem's package spelling, and the builder still applies
  each ecosystem's own format.
- **C2 (F2): ADDRESSED.** A builder's non-`Refusal` error becomes `&Refusal{Reason: ReasonBuilder}`, and the
  builder's text is not repeated (`querytpl.go:269-279`). A `*Refusal` passes through. The Investigator turns a
  `*Refusal` into `QueryNote = "refused: …"` (`investigator.go:101-106`). The worked example (`extend_test.go`
  banned-by-package) now returns `*querytpl.Refusal`, and ADAPTERS.md says so. M7 is red.
- **C3 (F4): ADDRESSED.** `checkNew` refuses a `Load.Table` that fails `ValidRelation` (`registry.go:162-164`).
  `Register` and `SetBuiltins` both go through `checkNew` (`registry.go:26,39`). The DoD mutates only
  `Load.Table`, in 5 cases plus a positive `schema.name` case. M6 is red. `ValidRelation` moved to `source`, and
  `querytpl.ValidRelation` delegates to it, so one implementation remains.
- **C4 (F3): ADDRESSED in this stream, not deferred.** On the Investigator path, `source.Lookup` finds the owning
  adapter and its `Rule` decides. When it returns false, no evidence is added. Otherwise the kept columns and rows
  replace the result's (`investigator.go:129-141`). The built-in rules keep every row when there is at least one
  (`adapters.go:65,107,144`), so the built-in behaviour is unchanged once the `len(Rows)==0` early return at
  `investigator.go:120` has run. ADAPTERS.md point 3 states the rule path. M8 is red.

### Remainder
Owed, non-blocking (R6/R14). None of these opens another round:
- **O1 (test gap).** The check in `source.Build` (`internal/source/registry.go:117-123`) is not pinned by any test
  (M2 survives). Today its only callers are the built-in OSV and KEV builders (`osv/check.go:209,310`), and those
  builders refuse a bad value themselves, so the duplicate layer is hidden. A registered concatenating builder
  reached through `source.Build` would lose its only guard silently. Remedy: put the `concatAdapter` probe
  through `source.Build` and assert `*source.ValueError`. Mutated field: the parameter value.
- **O2 (test gap).** In the Investigator, the assignment `res.Columns, res.Rows = kept.Columns, kept.Rows`
  (`investigator.go:140`) has no test. No probe rule keeps a strict subset, so a mutant that ignores `kept` would
  survive. Remedy: a rule that keeps one row of two, with an assertion on the evidence row count.
- **O3 (test gap, minor).** Two members of the byte floor in `CheckValue`, backslash and bytes above 0x7e, have
  no case of their own in `registered_test.go`. Quote, NUL and newline do.
- **O4 (behaviour note, declared not judged).** In the OSV trace note, a package name outside `anyPackageFormat`
  now reads `query parameter value is outside its format: osv-by-package package` (`ValueError`). Before, it
  carried the builder's refusal text. Both are notes and the review continues
  (`TestCheckSendsNoQueryForAPackageOutsideItsFormat` is green). Whether that wording is wanted is the master's
  call.

### Profile
security (R9): data access. A query is built from text a model chose, and the trust boundary is
model → template parameter → SQL literal. The fix moves the guard from the adapter, which an organisation writes,
to Proven. The floor rejects every byte that could end a single-quoted literal (quote, backslash) or alter a log
or terminal (control bytes, non-ASCII). Each authorization-shaped check was removed once (M1, M4–M8) and a test
failed, except M2 (O1). No dependency delta. `go.mod`/`go.sum` are untouched by the fix diff.

### Focus
The four round-1 conditions and the fix diff only (R14). Declared not judged (R11): main's absorbed content in
`a56a7e8` (`examples/formatter`, ADAPTERS.md "Adding an output format", SECURITY.md); `docs/README.md`
conflict-resolution wording at `177bba6`; documentation link checking (not run by this leg; the master reports
check-links green).

### Conditions
1. **Unrun closure (law #53):** before the gate, the master pastes a green `go test -count=1 ./...` at `a56a7e8`
   run outside a sandbox that refuses loopback listeners and module downloads. It must cover the packages listed
   under Unrun. The master states a green run of this kind at `a56a7e8` from 2026-10-24 on macm5, and this leg
   did not reproduce it.
2. O1–O3 are owed test work. They may land in this stream or be recorded by the master as a follow-on. They do
   not block.

### Unrun
- `github.com/saphan-ai/proven/internal/cli/sarifschema` (test build, and `go vet ./...`): refused on a module
  download. Quoted: `github.com/santhosh-tekuri/jsonschema/█████████: Get "https://proxy.golang.org/github.com/santhosh-tekuri/jsonschema/v5/@v/v5.3.1.zip": tls: failed to verify certificate: x509: OSStatus -26276`
- Panicked on a loopback listener in this sandbox. Quoted:
  `panic: httptest: failed to listen on a port: listen tcp6 [::1]:0: bind: operation not permitted`. Packages:
  `cmd/livecheck` (TestAllServicesReady), `internal/generator` (TestLiveReadsTheProvisionedStream),
  `internal/guardrail` (TestServerPostsTheMeasuredRequest), `internal/pipeline`
  (TestUnsafePromptStopsTheStepAndNamesCategories), `internal/query` (TestLiveSelectReadsRowsFromTheSQLEditor),
  `internal/review` (TestLiveReviewCarriesTheRoutedModel), `internal/store` (TestStoresAgreeOnOneScenario).
  Each package stops at its first panic, so its remaining tests did not run either.
- Documentation link check: not run by this leg.

## The review status as filed

## REVIEW — the-source-adapter-interface
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: claude-opus-5-5 re-review of W47 round 2 (0189d9d..a56a7e8): C1–C4 all ADDRESSED; mutants M1 and M3–M8 each turn a named test red.
  Condition: the master pastes a green unsandboxed `go test ./...` at a56a7e8 covering the Unrun packages (sarifschema download; 7 packages with httptest listeners).
  Owed, non-blocking: the check in source.Build is unpinned (M2 survives); the Rule row-subset path is untested; the backslash and high-byte floor members have no case of their own.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-25T00:00:00Z

## REDACTED

| what | times |
|---|---|
| e-mail or user name | 1 |
