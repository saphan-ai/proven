# REVIEW — the-source-adapter-interface · round 1

- verdict: reject
- reviewer: claude-code on macm5, claude-opus-5-5
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-24T15:36:48Z

## Summary

REJECT (claude-opus-5-5, at 0189d9d). Property 4 does not hold. querytpl.Build passes a registered template's model-supplied value to the adapter's builder unchecked, so it can carry a quote, NUL or newline (probe reproduced); Register admits a non-plain Load.Table; the worked example's plain-error refusal is fatal to the review.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — the-source-adapter-interface

Reviewer: REVIEW leg `the-source-adapter-interface-review`, model claude-opus-5-5, 2026-10-24.
Reviewed: `5faabc9...0189d9d` (three-dot), HEAD verified `0189d9d`. Profile: **security** (R9: the diff opens a
new path by which text reaches a store's query; ADR-0032 is the boundary), plus **docs** for ADAPTERS.md.

### Ground

- Diff: 11 files, +831 −95. New: `internal/source/{source.go,registry.go,builtin_test.go,extend_test.go}`,
  `internal/querytpl/adapters.go`. Changed: `internal/querytpl/querytpl.go` (Catalogue/Build read the registry;
  the `switch` and `lookup` are gone), `internal/osv/check.go` and `internal/epss/epss.go` (build through
  `source.Build`, decide KEV/EPSS no-row through the adapter's `Rule`), `docs/ADAPTERS.md`, `docs/README.md`
  (conflict markers removed — verified in the diff: the `<<<<<<<`/`=======`/`>>>>>>>` lines are deleted and one
  merged ADAPTERS row stands).
- `internal/source` non-test files import only `fmt` and `sync` (read) — the import direction the ADR asks for holds.
- No dependency manifest changed (R4): `go.mod`/`go.sum` are not in the diff stat.
- `docs/QUERIES.md`, `internal/agent/**`, `test/**`, `demo/**`, `internal/fixture/**`: zero diff
  (`git diff --stat 5faabc9 0189d9d -- …` printed nothing).
- Own runs at 0189d9d, in this leg's sandbox: `go build ./...` exit 0. `go vet ./...` and `go test ./...` — see Unrun.
- The author's self-assessment and the master's measurements were set aside as evidence (R1); each was re-measured
  or listed under Unrun.

### Differential

What the diff changes in behaviour, per property of the order.

**P1 — one place for a template: HOLDS.** `querytpl.Catalogue` (querytpl.go:136) iterates `source.All()`;
`querytpl.Build` (querytpl.go:263) resolves through `source.Lookup`; osv/epss call `source.Build`
(check.go:209, check.go:306, epss.go:154); prompts are written from `Describe()` → `Catalogue()`. The template
literals exist once, in `adapters.go`.

**P2 — behaviour unchanged: HOLDS.** `TestBuiltSQLIsTheTextThePipelineAlwaysSent` (querytpl_test.go:112) is
untouched by the diff and pins the four texts byte for byte through `Build`; `docs/QUERIES.md` and every
prompt/fixture/scenario file have zero diff; the built-in order OSV, KEV, EPSS is the old `Catalogue` order, so
`Describe()` prints the same string. osv/epss move from a direct call to `source.Build` with the same builder
function underneath.

**P3 — an added source shows its template, runs on the local store, its rule produces evidence: HOLDS IN THE
TEST, NOT IN THE PIPELINE.** `TestAddedSourceShowsItsTemplateItsQueryAndItsRule` (extend_test.go:76) proves the
catalogue entry and the local-store run, then calls `sample.Rule(...)` **itself** (extend_test.go:127). No
production code calls a registered adapter's `Rule`: the only production `Rule` callers are osv/check.go:333 (KEV,
by fixed id) and epss.go:174 (EPSS, by fixed id). The Investigator, which is the path by which a registered template
reaches a review, decides on `len(res.Rows) == 0` (investigator.go:119) and never looks up the owning adapter. See
finding F3.

**P4 — a registered source cannot widen what reaches a store: DOES NOT HOLD.** Answer to the order's question
"can a registered template bypass the identifier and literal checks?": **yes — there is no Proven-side check for it
to bypass.** `querytpl.Build` (querytpl.go:260–283) runs `checkParams` (names present, no extras — the *set*,
never the *value*), then hands the raw model-supplied map to `spec.Build`. `source.Param.Format` is prose for the
prompt; it carries no matcher. `literal` (querytpl.go:228) is unexported, so an adapter outside `querytpl` cannot
even reach the one literal builder. `source.checkNew` (registry.go:131) checks id, origin/refresh/table
non-empty, names unique and builder non-nil — it never checks `Load.Table` is a plain name, though `source.go:58`
tells an adapter it "may ignore rel and use Load.Table". Reproduced — see Closure. See findings F1, F2, F4.

### Unplug

Which control turns red when each property is removed. Every row below was **run by me** unless marked READ.

| Property | Mutation (out of tree, `go test -overlay`; worktree untouched) | Red controls |
|---|---|---|
| P1 + P2 | `adapters.go`: KEV and EPSS swapped in `SetBuiltins`; KEV builder writes `EPSSSQL` | querytpl: `TestBuiltSQLIsTheTextThePipelineAlwaysSent`, `TestCatalogueFollowsTheSourceRegistry`, `TestEveryIdParameterRefusesLineBreaksNulAndAForeignId`; source: `TestRegistryOrderIsOSVThenKEVThenEPSS`; osv: `TestCheckKeepsAMatchAndLooksUpTheCVE`, `TestCheckUsesTheKEVViewOnTheContext`; investigator: `TestInvestigatorRunsTheQueryProvenBuilds` — 7 tests in 4 packages |
| P3 | remove `bannedPackage{}` from extend_test.go:78 | READ, not re-run: the test then `t.Fatal`s at extend_test.go:94 (catalogue has no `banned-by-package`). The master reports running it (FAIL); I did not repeat it. |
| P4 | — | **No control exists.** Nothing turns red when a registered template places an unchecked value; my probe (Closure) is the first test of it and it fails at the tip. |

remark: the P1/P2 mutant shows that **order** of the catalogue is guarded by `TestCatalogueFollowsTheSourceRegistry` and `TestRegistryOrderIsOSVThenKEVThenEPSS`, not by `TestQueriesDocumentListsEveryTemplate`, which stayed green under the swap (it checks presence, not order).
remark: the mutant swaps two built-ins only; it does not show that prompt golden files would catch a `Describe()` change, because the prompt-bearing packages outside investigator/skeptic were not in that run.

### Closure

Out-of-tree probe `probe/registered_probe_test.go` in this run's artifact directory, overlaid into
`internal/source` with `go test -count=1 -overlay …/probe/overlay.json -run Probe -v ./internal/source/`. Output,
verbatim:

```
--- FAIL: TestProbeRegisteredParamIsHeldToItsFormat
    PROBE: a registered template built "SELECT a FROM t WHERE n = 'x' OR '1'='1\x00\n'" from a value with a quote, NUL and newline; no refusal
    PROBE: Validate (the Skeptic's gate) admitted the same value
--- FAIL: TestProbeWorkedExampleRefusalIsANamedRefusal
    PROBE: worked-example refusal is *errors.errorString (banned-by-package: package is outside its format), not *querytpl.Refusal
--- FAIL: TestProbeRegisteredLoadTableIsAPlainName
    PROBE: Register admitted Load.Table "t; DROP TABLE x", which is not a plain name
FAIL	github.com/saphan-ai/proven/internal/source
```

#### Findings (most severe first)

**F1 · BLOCKING · `internal/querytpl/querytpl.go:277` (with `internal/source/source.go:43–48`, `internal/source/registry.go:159–170`).**
Defect class: unvalidated model-supplied value reaching query text (ADR-0032 points 1 and 3). Trigger: any adapter
added through `source.Register` whose builder concatenates a parameter; the Investigator's prompt lists that
template (`Describe` → `Catalogue`), a model names it, `querytpl.Build` passes the value through unchecked, the store
runs the text. Evidence: probe 1 above; `checkParams` (querytpl.go:294) checks keys only; `Param.Format` is prose.
Property 4 of the order fails. Before this diff the list was closed and every builder used `literal`; the diff opens
the list without moving the guarantee to the registry. **Remedy:** make the format a machine check owned by Proven,
not by the adapter — e.g. `source.Param` gains a matcher (an anchored `*regexp.Regexp`, or a closed set of named
formats Proven owns), `checkNew` refuses a param without one, and `querytpl.Build` runs the format check **and** the
`literal` byte floor (no quote, backslash, control byte, non-ASCII) on every value before calling `spec.Build`,
returning `*Refusal{ReasonBadValue}`. Proof expected: a test that registers a builder which concatenates its
parameter, sends a quote/NUL/newline value through `querytpl.Build` and `querytpl.Validate`, and requires
`ReasonBadValue` — red at 0189d9d (my probe is that test), green after.

**F2 · OWED (blocking once a production `Register` caller exists) · `internal/agent/investigator/investigator.go:101–107`, `internal/pipeline/pipeline.go:651–657, 1336–1348`; taught by `docs/ADAPTERS.md` § Worked example and `internal/source/extend_test.go:49`.**
A registered builder that refuses with a plain error — exactly what the worked example does (`fmt.Errorf`) — is not a
`*querytpl.Refusal`, so `Investigate` returns it, `soften` finds neither `service.Down` nor a cancelled context and
returns it as **fatal**, and the review ends. One malformed value from a model ends the whole review instead of
becoming a `QueryNote`. Evidence: probe 2 above, and the code read at the cited lines. **Remedy:** F1's remedy
removes the format case (Proven refuses first, as a `*Refusal`); additionally, `querytpl.Build` should wrap any
non-`*Refusal` builder error into a named refusal, and the worked example and ADAPTERS.md should say a builder
returns `*querytpl.Refusal`. Proof: the probe's second test green.

**F3 · OWED · `internal/agent/investigator/investigator.go:118–119`; order property 3 / ADR-0046 "its rule".**
A registered source's `Rule` is never called by the pipeline; the Investigator keeps any non-empty row set. The rule
an organisation writes therefore decides nothing in a review — it is exercised only by the test that calls it
directly. Same for the OSV built-in rule (by the leg's ruling; see Focus). **Remedy:** either the Investigator looks
up the owning adapter (`source.Lookup` already returns it) and keeps evidence only when `Rule` returns ok, with a test
in which a rule returning false on a non-empty row set yields no evidence; or ADAPTERS.md and source.go:107–110 state
that the rule is advisory today. The first is what the ADR describes.

**F4 · OWED · `internal/source/registry.go:151–154`.** `checkNew` does not apply `querytpl.ValidRelation`-equivalent
checking to `Load.Table`, while `source.go:58–59` invites a builder to place `Load.Table` in the query unquoted. Probe
3 above. **Remedy:** refuse a `Load.Table` that is not one identifier or `schema.name` at `Register`/`SetBuiltins`
(the check cannot import querytpl; a copy of `identifier` in `source`, or moving `ValidRelation` down into `source`
and aliasing it from querytpl, keeps one author). Proof: probe 3 green.

**F5 · JUDGEMENT (R6, does not block) · `docs/ADAPTERS.md` § Worked example.** "a quote, a space or a semicolon is
refused and never escaped" describes the example adapter's own loop, but reads as a property of the system. After F1
it becomes true of the system; until then it should say the adapter refuses it.

### Remainder

- Concurrency: `registry.go` holds one mutex; `Lookup` calls `a.Templates()` under it; `Build` releases it before
  calling the builder (registry.go:109–118). Read; `-race` not run.
- `Unregister` builds a new slice and leaves built-ins in place (registry.go:50–63) — matches its comment.
- `SetBuiltins` checks the new list against itself and against `added` before swapping (registry.go:22–31) — read.
- osv/epss now return a hard error when a built-in template is missing (`found == false`). Unreachable while `init`
  succeeds; `builtinErr` is only surfaced by a test, so a failed built-in registration would show up as these errors
  at review time rather than at start. Not a finding.
- `copyEvidence` keeps `Rows` non-nil on an empty input; `Rule` never reaches it with no rows. Fine.

### Profile

security (S1–S4) + docs.
- S1 boundary map: model answer (`template`, `params`) → `querytpl.Build` → builder → store. Spoof/tamper: the param
  value — guarded for built-ins by `literal`, **unguarded for registered templates (F1)**. The template name — guarded
  (unknown name refused; name never echoed). The relation names — guarded by `ValidRelation` inside each built-in
  builder; `Load.Table` unguarded (F4). Denial: one bad value ends a review through a registered builder (F2).
- S2: every blocking finding reproduced (Closure).
- S3: no authorisation check in the diff.
- S4: no dependency delta; no secrets in the diff (read); no destructive path. Invisible/bidirectional characters in
  the diff: `grep -P` is not an allowed program here, so this is UNRUN (see Unrun).
- docs: D1 ADAPTERS.md stands alone; D3 the worked example does not say what it does not do (F5); D4 the commands it
  quotes — none.

### Focus

The leg's rulings, judged:
1. **Adapters built in `internal/querytpl` because `internal/source` cannot import it — HONEST.** The import graph
   proves it (source imports fmt/sync only; querytpl imports source). Not a defect.
2. **The OSV rule is not the version decision — HONEST, and documented** (adapters.go:66–72, ADAPTERS.md). Consequence
   worth naming: the OSV `Rule` has no production caller, which is F3's pattern for a built-in.
3. **`Register` is in-process — HONEST**; it has no production caller today (grep: only extend_test.go:80). That lowers
   F1/F2 from a live exposure to a latent one, but it does not make property 4 hold: the order asked for the property
   of the door, and the door is public API documented for organisations. Per R12 the low exposure does not lower the
   class.
4. **The worked-example source is test-only — HONEST**, and stated in ADAPTERS.md.
5. **Empty snapshot dates for the three built-ins — HONEST.** source.go:24–27 and ADAPTERS.md say the store stamps
   the read; ADR-0046 line 28 asks that each finding state the snapshot date of its row, which the store stamp
   provides. Not a defect.

Declined to judge (R11), for the master to rule on: whether `Rule` should also be applied in osv for OSV rows; whether
`TestQueriesDocumentListsEveryTemplate` should also pin order (it stayed green under a KEV/EPSS swap).

### Conditions

Verdict `reject` under R2: property 4 of the order does not hold. Merge after these, re-reviewed against the fix diff
only (R14):
1. **C1 (F1, blocking):** Proven, not the adapter, checks every parameter value of every template — built-in or
   registered — against a machine format and the literal byte floor before the builder runs; `Register` refuses a
   param with no format. DoD: a test registering a concatenating builder; mutated field = the parameter value
   (quote, NUL, newline); `querytpl.Build` and `querytpl.Validate` both return `ReasonBadValue`; red at 0189d9d.
2. **C2 (F2):** a builder's error reaches the Investigator as a `*querytpl.Refusal`; the worked example and
   ADAPTERS.md return one. DoD: a value refused by a registered builder yields a `QueryNote`, not a failed review.
3. **C3 (F4):** `Register`/`SetBuiltins` refuse a `Load.Table` that is not a plain name. DoD: mutated field =
   `Load.Table`.
4. **C4 (F3), owed — may land in this stream or be recorded as a follow-on by the master:** the owning adapter's
   `Rule` decides evidence on the Investigator path, or the docs state the rule is not consulted there.

### Unrun

- `go vet ./...` — refused by the environment when fetching a test dependency for `internal/cli/sarifschema`:
  `github.com/santhosh-tekuri/jsonschema/█████████: Get "https://proxy.golang.org/github.com/santhosh-tekuri/jsonschema/v5/@v/v5.3.1.zip": tls: failed to verify certificate: x509: OSStatus -26276`. Vet therefore did not finish for any package.
- `go test -count=1 ./...` — these packages did not complete in this sandbox, every one for an environment reason, none
  touched by the diff:
  - `internal/cli/sarifschema` — `[setup failed]`, same TLS refusal as above.
  - `cmd/livecheck`, `internal/generator`, `internal/guardrail`, `internal/pipeline`, `internal/query`,
    `internal/review`, `internal/store` — each `panic: httptest: failed to listen on a port: listen tcp6 [::1]:0: bind: operation not permitted`
    (in `TestAllServicesReady`, `TestLiveReadsTheProvisionedStream`, `TestServerPostsTheMeasuredRequest`,
    `TestUnsafePromptStopsTheStepAndNamesCategories`, `TestLiveSelectReadsRowsFromTheSQLEditor`,
    `TestLiveReviewCarriesTheRoutedModel`, `TestStoresAgreeOnOneScenario`). Measured twice with the same result.
  remark: a panic ends its package's test binary, so the other tests in these 7 packages (including the local-store tests in `internal/store` and every `internal/pipeline` test) are unrun here, not green.
  remark: the full-suite exit status is FAIL for environment reasons only; it cannot tell a diff-caused failure inside those 7 packages from none.
  Every other package passed, including every package the diff touches that has tests: `internal/querytpl`,
  `internal/source`, `internal/epss`, `internal/osv`, `internal/agent/investigator`, `internal/agent/skeptic`.
- `sh scripts/check-links.sh` — refused: `the program "sh" is NOT in the ██████████████ list of role review under permission class ██████████████████████████ … [tool-command-not-in-class]`. Not rebuilt from other commands (law #57).
- Scan of the diff for invisible/bidirectional characters (S4) — not run: no allowed program does a byte-class scan.
- P3 unplug (remove `bannedPackage{}`) — read, not re-run (see Unplug).
- `go test -race` on `internal/source` — not run.

## The review status as filed

## REVIEW — the-source-adapter-interface
- state: recommendation-ready
- verdict: reject
- ref: _REVIEW_NOTES.md
- abstract: REJECT (claude-opus-5-5, at 0189d9d). Property 4 does not hold. querytpl.Build passes a registered template's model-supplied value to the adapter's builder unchecked, so it can carry a quote, NUL or newline (probe reproduced); Register admits a non-plain Load.Table; the worked example's plain-error refusal is fatal to the review.
  P1 and P2 hold and are well guarded (one mutant turned 7 tests red in 4 packages). P3 holds only in its test: the pipeline never calls a registered Rule. Conditions C1–C4 are in the notes.
  Unrun: go vet (TLS module fetch); 8 packages whose tests could not complete in this sandbox (httptest bind refusal, sarifschema setup); check-links (`sh` denied).
- artifacts:
    - _REVIEW_NOTES.md
    - _REVIEW_STATUS.md
- updated: 2026-10-24T16:10:00Z

## REDACTED

| what | times |
|---|---|
| e-mail or user name | 1 |
| policy binding name | 2 |
