# STATUS — query-evidence-comes-from-templates · round 1

- state: ready-for-review
- executor: not recorded
- model: not recorded
- pool: not recorded
- cost: not recorded
- duration: not recorded
- last commit: b4bd425 [query-evidence-comes-from-templates] absorb main (the master: W30 and W31 absorbed; the redactor redacts the refutation template; 36 packages ok)
- updated: 2026-10-23T18:07:29Z

## Summary

The Investigator and the Skeptic now name one of four query templates and its values; Proven checks each value against a strict ASCII format and builds the SQL itself (internal/querytpl, docs/QUERIES.md, ADR-0032). An answer with SQL text, an unknown template or a bad value sends no query and the trace names the reason. OSV, KEV and EPSS use the same builder; the live gate is kept as a second, stricter check.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — query-evidence-comes-from-templates
- state: stop-2-ready
- abstract: The Investigator and the Skeptic now name one of four query templates and its values; Proven checks each value against a strict ASCII format and builds the SQL itself (internal/querytpl, docs/QUERIES.md, ADR-0032). An answer with SQL text, an unknown template or a bad value sends no query and the trace names the reason. OSV, KEV and EPSS use the same builder; the live gate is kept as a second, stricter check.
- last commit: b4bd425 [query-evidence-comes-from-templates] absorb main (the master: W30 and W31 absorbed; the redactor redacts the refutation template; 36 packages ok)
- done: five commits on fix/query-evidence-comes-from-templates, build and vet green, every test that can run in this sandbox green, sample scan on testdata/fixture/discount gives the same four findings as main
  - step-1 red at base 741cc77: 7 failing tests, output in ████████████████████████████████████████████████████████████████ (go test -v, exit 1 for osv and query)
  - each format refuses a quote, a backslash, a control byte and invalid UTF-8: TestEachFormatRefusesAHostileValue and TestEveryBuiltLiteralIsOneLiteralInEitherDialect in internal/querytpl, go test exit 0
  - a model answer carrying SQL is refused: TestInvestigatorSendsNothingForAnAnswerThatIsNotATemplate, TestAnswerThatIsNotATemplateProposesNoQuery, TestRefutingQueryThatIsNotATemplateIsRefused, go test exit 0
  - negative control: the SQL check in investigator.go row.request removed, go test ./internal/agent/investigator exit 1 (sql_text and sql_beside_a_template red), check restored, exit 0
  - proven scan --sample on testdata/fixture/discount from a scratch directory outside the repo, base binary and new binary, diff shows only F2's own query evidence changed: 3 lines of free-text advisories query replaced by the 2-row osv-by-package query
- next / blocked-on: master verifies and runs the named review; the Zetaris read-only role (ADR-0032 point 4) is the Zetaris track's measurement and is not claimed anywhere
- artifacts:
  - internal/querytpl/querytpl.go
  - internal/querytpl/querytpl_test.go
  - internal/querytpl/literal_test.go
  - docs/QUERIES.md
  - ████████████████████████████████████████████████████████████████
  - ██████████████████████████████████████████████████████████████
  - █████████████████████████████████████████████████████████████
  - █████████████████████████████████████████████████████████████
- evidence:
  - go build ./... ; go vet ./... : exit 0, no output
  - go test -count=1 ./... at tip: ok for every package except six that panic on the sandbox refusing a loopback listener (cmd/livecheck, internal/generator, internal/guardrail, internal/pipeline, internal/query, internal/review); the same six panic at base 741cc77, see full-tip.txt and the base run
  - the 68 tests of the 8 files that start a listener were run one by one (run_each.py): at base PASS 18, NOLISTEN 49, FAIL 1 (guardrail TestSilentServerIsATimeout); at tip PASS 17, NOLISTEN 49, FAIL 2; the extra FAIL at tip is internal/review TestOpenSampleReviewKeepsF1AndF2 which was fixed after that run (expects the new F2 rows), re-run green below
  - the rest of those six packages with the listener tests skipped (run_skip.py): ok for cmd/livecheck (no tests left), generator, guardrail, pipeline 11.9s, query, review 61.3s, all exit 0
  - go test -count=1 ./internal/querytpl ./internal/agent/... ./internal/osv ./internal/epss ./cmd/proven : all ok
  - UNRUN: 49 listener tests (named in each-tip.txt as NOLISTEN), among them the live Zetaris client tests in internal/query/live_test.go and review/zetaris_test.go, whose recorded answer I changed to a template but could not execute
  - UNRUN: -race (no goroutine or shared state touched), coverage and mutation bars (none declared for this repository)
  - baseline E10: full suite on a git-archive copy of 741cc77 under ██████████████████████████████████████████████████████ same six listener-panic packages red, everything else ok
- note: master-was-wrong: (1) the order lists a version format; no template takes a version, because the OSV table is read by package and the version is compared in Go by osv.Check, never in SQL, so I wrote none and docs/QUERIES.md says so. (2) the second reviewer's test TestModelValueCannotChangeQueryPredicate asked the live gate to refuse `WHERE cve_id = 'X' OR ...`, which is a legal single SELECT: no gate can refuse it; the property holds where the value enters (querytpl refuses the quote), so that test moved there as part of TestEveryBuiltLiteralIsOneLiteralInEitherDialect and was removed from internal/query. (3) the measured claims about quoteLiteral, the gate and the missing refusals were all reproduced red at the base before any change. JUDGEMENT CALLS: (a) a package name from the change outside its format is a trace note and the check goes on, where quoteLiteral used to return an error that ended the check; cost if wrong: one package silently unchecked, but the note names it. (b) the old free-text gate is KEPT as a second check, made strict (no backslash, no control byte, no semicolon or comment outside a literal), not removed; the liveness probe still uses it. (c) Refutation.SQL is replaced by Refutation.Query, a template request; a candidate gets QueryNote for a refused Investigator answer. (d) osv-by-advisory has no setting for its view name, only the default px_evidence.px_osv_id. (e) the refuted fixture needed a row a template can read, so its tables.sql now holds a px_osv_id row instead of the free-text checks table; discount F1 and F3 name a package with no row to keep their story. (f) the advisories table stays in discount/tables.sql, unread by any template. (g) the sample F2 evidence now shows two OSV rows instead of one advisories row; the findings are the same four. HOUSEKEEPING: no AI attribution in any commit and no Co-authored-by trailer (git log 741cc77..HEAD checked). A commit printed packed-refs.lock Permission denied and git log -1 showed it landed. I could not delete or move files (rm and git worktree add are denied), so scratch lives under ████████████ and an artifact directory; osv/literal_test.go keeps its file name though it now tests the check's refusals. No session recording was made by this leg (ADR-0024): the tool provides none; the guardrail transcript is under .saphan-guardrail. Commands and exit codes are in the evidence list above.
- updated: 2026-10-23T14:55:00Z

CLOSE-OK · query-evidence-comes-from-templates · commit 291753a · branch fix/query-evidence-comes-from-templates · worktree ███████████████████████████████████████████████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 8 |
