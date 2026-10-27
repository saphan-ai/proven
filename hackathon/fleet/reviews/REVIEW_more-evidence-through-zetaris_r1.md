# REVIEW — more-evidence-through-zetaris · round 1

- verdict: accept-with-conditions
- reviewer: claude-code on macm5, claude-opus-5-5
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-25T03:40:14Z

## Summary

Named review of W43 r2 (main...0aecadc) on claude-code claude-opus-5-5: all four properties hold on every path read; 3 mutants red, the noteRaised unplug red.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — more-evidence-through-zetaris

Reviewer: claude-code, `claude-opus-5-5`, headless, 2026-10-25. Diff judged: `main...0aecadc` (three-dot), HEAD verified `0aecadc`.
The master's self-measurement (order §"THE MASTER CAN BE WRONG") was set aside as evidence (R1) and re-measured where the sandbox allowed.

### Ground

- `go build ./...` — exit 0 (this seat, 2026-10-25).
- `go vet ./...` — exit 1, but NOT on this diff: `internal/cli/sarifschema/schema_test.go:12:2` cannot download `github.com/santhosh-tekuri/jsonschema/█████████` (`tls: failed to verify certificate: x509: OSStatus -26276`). Sandbox network, not code.
- `go test -count=1 ./...` — every package `ok` except eight, all refused by the sandbox (listed under Unrun). Packages the diff touches that ran green in full: `internal/querytpl`, `internal/source`, `internal/osv`, `cmd/proven`.
- Targeted, verbose, all PASS: `TestADR0039SQLIsBuiltFromTheTemplates`, `TestADR0039Rules`, `TestCatalogueFollowsTheSourceRegistry`, `TestRegistryOrderIsFixed`, `TestBuiltinRulesKeepRowsAndDropNone`, `TestRaisedPackageBindsTheADR0039Templates`, `TestMaliciousPackageReportsTheAdvisory`, `TestProjectHealthReportsNoMaintainers`, `TestCVSSScoreSitsBesideEPSS`, `TestSPDXLicenceReportsADisallowedLicence`, `TestLocalMissingTableIsADownSource`, `TestSampleReviewNamesEachFindingsLevel` (the last two are in packages that otherwise panic on a listener; run with `-run`).
- Dependency delta (R4): no root `go.mod`/`go.sum` change. The eight `go.mod` files added are fixture inputs under `testdata/fixture/*/{base,change}/`.

### Differential

Property by property, FILE·LINE·CONTROL.

**P1 — each of the four is a source adapter (load + snapshot, checked templates, rule; empty list → nothing).** HOLDS.
- Registration: `internal/querytpl/adapters.go:17-20`; order held by `TestRegistryOrderIsFixed` and `TestCatalogueFollowsTheSourceRegistry`.
- Load: `Origin/Refresh/Table` set; `Snapshot` left empty for all four, same as OSV/KEV/EPSS, and the `source.Load.Snapshot` contract says empty means "this process did not load the table; the evidence carries the store's read date". The rendered evidence does carry it (`evidence query local-sample snapshot=sample-2026-10-25`, seen in the CVSS scenario body). That satisfies ADR-0039 "each finding states the snapshot date of its row".
- Parameters: every Param carries `Match` (`ecosystemFormat`, `anyPackageFormat`, `cveFormat`); the package value goes through `ecosystemKey` → `checked(... packageFormat(ecosystem))` (`internal/querytpl/querytpl.go` `ecosystemKey`), the CVE through `literal(... cveFormat)`; the table name through `ValidRelation`. sql and mango forms share `ecosystemKey`, so they key the same rows (`mango.go` `packageFind`).
- Rules: empty list → `false` at `adr0039.go:89,137,200,238`; a filtered-to-nothing list → `false` via `keepRows` (`adr0039.go:276`). Controls: `TestADR0039Rules` (loop over all four with no rows), plus per-rule positives/negatives.

**P2 — a missing table is a named down source on every store.** HOLDS by construction; tested for one of the four.
- Local: `TestLocalMissingTableIsADownSource` PASS here.
- All three stores: `internal/store/missing_test.go` `TestStoresAgreeOnAMissingTable` now requests `kev-by-cve` and `mal-by-package` on local, Zetaris (recorded 404) and CouchDB (fake server, reason prefix `no table `). UNRUN here (listener refused). `scorecard`/`spdx` share `packageFind`/`packageSQL` with `mal`; `cvss` shares `cveFind` with `kev`. So the property follows from the shared builders, but no test names scorecard/cvss/spdx → condition C1.

**P3 — no existing scenario, fixture prompt or generated doc changes beyond the four sources; QUERIES.md follows the registry.** HOLDS.
- Word diff of `testdata/fixture/{discount,fix,refuted}/fixture.json`: 7 removed chunks, all the word-boundary token `CVE.\n\nReply`; 14 added chunks, all the same four-template catalogue sentence (`mal-by-package … spdx-by-package …`). No response, planner hash or other field moved. No other existing fixture file changed.
- `docs/QUERIES.md` gains the four rows and the four table-name rows; `TestQueriesDocumentListsEveryTemplate` (iterates `Catalogue()`) PASS.
- Renamed test `TestRegistryOrderIsOSVThenKEVThenEPSS` → `TestRegistryOrderIsFixed`: rename only, assertion extended, not weakened.

**P4 — a finding from these sources shows the level its rows give (ADR-0044), never more.** HOLDS on every path read; not pinned by a test for these sources.
- The level is never set by an adapter: `internal/review/open.go:457` → `ProofLevelOf` (`internal/review/review.go:224`) maps `EvidenceQuery` → `proved-by-data`; only test/tool evidence ranks higher, and none of the four adapters emits either.
- Rows count only when bound: `internal/agent/judge/judge.go:73` keeps a query only if `Bound && len(Rows) > 0`; binding is `internal/osv/check.go` `Facts.Binds` (new cases for the four), wired at `internal/pipeline/pipeline.go:979`. An unbound query becomes context with a QueryNote and doesn't count.
- New binding is narrow: mal/scorecard/spdx bind only a package the change added or raised (`noteRaised`, `check.go`), and `osv-by-package` still needs an affected range; cvss binds only a CVE the check recorded. Control: `TestRaisedPackageBindsTheADR0039Templates`.
- Gap: none of the four scenario tests and no review-level test asserts `proved-by-data` for a finding from these sources → condition C2.

### Unplug

Mutants run on this seat, each restored with `git checkout -- <file>`; `git status --porcelain` matches the starting state.

| Mutant | Location | Result |
|---|---|---|
| MAL- filter bypassed (`idAt >= -1 \|\| …`) | `adr0039.go:97` | RED — `TestADR0039Rules`: "a row whose id does not start with MAL- was evidence". `cmd/proven` scenarios stayed green (fixture has only a MAL- row). |
| young-release bound `<` → `<=` | `adr0039.go:166` | RED — `TestADR0039Rules`: "a release of 14 days was treated as only days old". |
| `out.Facts.noteRaised(...)` removed | `internal/osv/check.go` Check loop | RED — `TestMaliciousPackageReportsTheAdvisory`, `TestProjectHealthReportsNoMaintainers`, `TestSPDXLicenceReportsADisallowedLicence` (finding absent; judge drops unbound rows). CVSS scenario unaffected, as designed (it binds through `ids`). |
| master's claim: mal empty-list return made to accept | `adr0039.go:89` | Not re-run. Read: the direct `ad.Rule(... Rows: nil)` call in `TestMaliciousPackageReportsTheAdvisory` and the loop in `TestADR0039Rules` both assert `!ok`, so turning the return into accept goes red. Deleting the guard (rather than inverting it) is an EQUIVALENT mutant for mal/scorecard/spdx, because `keepRows` also returns false on no rows. |

### Closure

What a reader can rely on after this diff:
- The four templates go through the same closed catalogue, the same parameter checks, the same plain-name relation check and the same sql/mango key function as OSV.
- A row from these tables becomes evidence only if the template binds to a package the change added or raised, or to a CVE the dependency check recorded; anything else stays context.
- The strongest level these findings can carry is `proved-by-data`.

### Remainder

Owed, non-blocking (R6/R8):
- O1 `adr0039.go:89,137,238`: the explicit `len(in.Rows) == 0` guards in mal/scorecard/spdx are redundant with `keepRows`. They read clearly but make "remove the guard" an equivalent mutant; the guarding test is really `keepRows`'s.
- O2 `internal/osv/check.go` `Binds`: mal/scorecard/spdx are keyed on the model's raw `package` param, but `noteRaised` uses the manifest spelling. A PyPI/Packagist name the model sends in a different spelling than the manifest (the builder itself normalises with PEP 503/lowercase) does not bind, and the finding is silently dropped. This only fails toward weaker, same as the existing OSV `packages` key. Remedy: key both sides on `ecosystemKey` output.
- O3 `adr0039.go:156`: `maintained == "0"` matches only that spelling; `"0.0"` is not the signal. That fails toward weaker; the docs name `0`.
- O4 the CVSS scenario renders two findings for one dependency/CVE (the CVSS finding F1 and `osv:evil.example/scored`, each with the same EPSS line). That is the order's design ("beside EPSS"), but a reader sees the CVE twice.
- O5 (pre-existing, outside this diff) the text renderer prints no level of proof; ADR-0044 §2 says text shows it first. `ProofLevel` lives on `review.Result`, and nothing under `internal/cli` reads it.

Declined to judge (R11):
- The SPDX policy content: the permissive set, whole-cell comparison (`"MIT OR Apache-2.0"` is reported as disallowed), `NOASSERTION` reported as disallowed, empty cell skipped. The order puts licences out of scope. I judged only the adapter mechanics.
- Whether the four table names should be settings rather than adapter constants. The docs record the decision; that is an ADR matter.

### Profile

backend (data access through templated queries) with security touch on the query-building boundary: values reach SQL only after `Match` + `CheckValue` byte floor + `checked`; relation names through `ValidRelation`. The new builders add no path around those checks (`packageSQL`, `cvssSQL`, `packageFind`). No N+1: one query per template call. No new store, cache or persistence.

### Focus

The four properties in the order; the binding gate (`Facts.Binds`) as the place where a new source could overstate a finding; the existing-fixture byte change; the store miss behaviour.

### Conditions

- C1 (P2, test coverage): extend `TestStoresAgreeOnAMissingTable` (`internal/store/missing_test.go`) to iterate all four ADR-0039 templates (or all of `querytpl.Catalogue()`), not only `mal-by-package`. Negative control: point `packageFind` or `cveFind` at an existing table for one of them → that store's half goes red.
- C2 (P4, test coverage): assert the level for a finding from these sources. For example, a `review_test` over the `malicious-package` fixture asserting `ProofLevel == proved-by-data` and `ProofCounts[proved-by-data] == 1`. Negative control: map `EvidenceQuery` to `ProvedByTool` in `ProofLevelOf` → red.
- C3 (verification, master): run the Unrun packages below outside the sandbox at `0aecadc` and paste the result in the gate note; that is where `TestStoresAgreeOnAMissingTable` (P2 across Zetaris and CouchDB) actually runs.

### Unrun

- `internal/store` (includes `TestStoresAgreeOnAMissingTable`, the P2 three-store control), `internal/review` (full package), `internal/pipeline`, `internal/query`, `internal/generator`, `internal/guardrail`, `cmd/livecheck`. Refusal, quoted: `panic: httptest: failed to listen on a port: listen tcp6 [::1]:0: bind: operation not permitted`. The panic aborts each package, so its later tests also did not run. Subsets run with `-run` are listed under Ground.
- `internal/cli/sarifschema` and `go vet ./...` as a whole. Refusal: `Get "https://proxy.golang.org/github.com/santhosh-tekuri/jsonschema/v5/@v/v5.3.1.zip": tls: failed to verify certificate: x509: OSStatus -26276` (`[setup failed]`).
- Running the built `proven` binary directly against the four fixtures. Refusal: the program "…/proven" is NOT in the ██████████████ list of role review (`tool-command-not-in-class`). Not rebuilt from other commands; the scenario bodies were read through `go test` instead.

## The review status as filed

## REVIEW — more-evidence-through-zetaris
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: Named review of W43 r2 (main...0aecadc) on claude-code claude-opus-5-5: all four properties hold on every path read; 3 mutants red, the noteRaised unplug red.
  Conditions: C1 extend the three-store missing-table test to scorecard/cvss/spdx; C2 assert proved-by-data for an ADR-0039 finding; C3 master runs the sandbox-refused packages.
  Unrun: internal/store, review, pipeline, query, generator, guardrail, cmd/livecheck (httptest bind refused); sarifschema and go vet (module TLS).
- artifacts:
    - _REVIEW_NOTES.md
    - _REVIEW_STATUS.md
- updated: 2026-10-25T03:40:00Z

## REDACTED

| what | times |
|---|---|
| e-mail or user name | 1 |
| policy binding name | 1 |
