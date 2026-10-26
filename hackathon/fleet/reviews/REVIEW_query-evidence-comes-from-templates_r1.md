# REVIEW — query-evidence-comes-from-templates · round 1

- verdict: accept-with-conditions
- reviewer: claude-code on macm5, claude-opus-5-5
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-23T15:38:08Z

## Summary

claude-code claude-opus-5-5 named review of W32 at 291753a, verdict accept-with-conditions. The property holds: every value is refused or becomes one quote- and backslash-free ASCII literal (querytpl.go:238-248), and no model text reaches the SQL. Proven by the leg's tests and by two 300k/256-byte scratch probes. Original reviewers' tests: 6 pass (4 ported), and codex's gate-predicate test stays red by design.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — query-evidence-comes-from-templates

Reviewer: claude-code (`claude-opus-5-5`), REVIEW leg `query-evidence-comes-from-templates-review`, 2026-10-23.
Object: `741cc77...291753a`, five commits (`git rev-parse --short=7 HEAD` printed `291753a`). Profile: security (R9).
Property judged: **a value from the change or a model enters a query only as a literal and cannot change the query's
meaning; no query text written by a model reaches Zetaris** (SECURITY_REVIEW.md piece 4, ADR-0032).
Verdict: **accept-with-conditions**. The property holds, from the code and by test. Two owed conditions concern the
second line of defence and one untested guard. Neither blocks. No author self-assessment was used (R1); the leg's
commit messages and `docs/` prose were read as claims and checked against the code.

### Ground

Toolchain: `go.mod` pins `go 1.26`. Every `go test` ran with `-count=1`. The scratch files live OUTSIDE the worktree
in this run's artifact directory (`█████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████`)
and reach `go test` through `-overlay`, so no file of the package was created or changed (see Remainder for why
not `█████████████`).

**Item 5 — `go build ./... && go vet ./... && go test -count=1 ./...`:** build ok, vet ok. Test: **28 packages ok,
6 FAIL** (`cmd/livecheck`, `internal/generator`, `internal/guardrail`, `internal/pipeline`, `internal/query`,
`internal/review`). `grep -c 'httptest: failed to listen on a port: listen tcp6 [::1]:0: bind: operation not permitted'`
over the same run printed **6**: one listener panic per failing package. This seat refuses a loopback listener, and a
panic stops the whole package at its first httptest test. The diff touches three of those packages, so I re-ran them
with exactly the listener tests skipped (`-skip` naming the 16 tests in `pipeline/guardrail_test.go`,
`query/live_test.go`, `review/zetaris_test.go`, `review/live_generator_test.go`, `review/open_test.go` that call
`httptest.NewServer`) → `ok internal/pipeline`, `ok internal/query`, `ok internal/review`. `cmd/livecheck`,
`internal/generator` and `internal/guardrail` are not in the diff; their listener tests are Unrun.

**Every query call site at the tip** (`grep -rnI --include=*.go -E '\.Query\(' internal cmd demo fleet`, non-test,
non-hmem):

| site | SQL comes from | values |
|---|---|---|
| `internal/osv/check.go:164,172` | `querytpl.OSVPackageSQL` | package from the change's manifests / lock files; ecosystem a reader constant |
| `internal/osv/check.go:242,246` | `querytpl.KEVSQL` | CVE alias from an OSV row |
| `internal/epss/epss.go:152,156` | `querytpl.EPSSSQL` | CVE id from a kept finding's query cells |
| `internal/agent/investigator/investigator.go:94,108` | `querytpl.Build` | template + params from a **model** (Investigator) |
| `internal/agent/investigator/investigator.go:64,71` (`Execute`) | `querytpl.Build` | template + params from a **model** (Skeptic), via `pipeline.go:721` |
| `cmd/livecheck/main.go:170` | `"SELECT 1 FROM "+view+" LIMIT 1"` | view from a setting, `epss.ValidView` → `querytpl.ValidRelation`; no change or model value |

`hmem`'s and `meterless-conform`'s `Query` are a memory ranking, not SQL. `grep 'Refutation\.SQL'` returns nothing:
the field is gone (`candidate.go:480`, now `Query querytpl.Request`).

### Differential

#### Item 1 — the property, from the code

- **No model text reaches the SQL, beyond checked values.** `investigator.go:51-56` refuses an answer with a `sql`
  field (`ReasonModelSQL`). `investigator.go:94` and `:64` build the only text that is sent, with `querytpl.Build`.
  `skeptic.go:357-365` refuses `sql` and validates the request. `pipeline.go:721` rebuilds the request with the
  review's settings through `Execute`, which calls `Build` again. `Build` (`querytpl.go:266-294`) never pastes
  `req.Template`. It pastes a value only through `literal` (`:238-248`). A table or view name goes in only after
  `ValidRelation` (`:201-212`). The ecosystem goes in only after the exact `Go|npm|PyPI` switch (`:349-353`).
- **A value is one literal holding exactly itself.** `literal` matches an anchored ASCII allow-list. All five formats
  (`querytpl.go:187-191`) exclude `'`, `\`, `"`, backtick, `;`, `*`, space and every byte ≥ 0x80 or < 0x20. A second
  loop (`:242-246`) refuses `< 0x20`, `> 0x7e`, `'` and `\`. So the bytes between the quotes contain no quote and no
  backslash. **This is dialect-independent**: with neither byte present, an engine that reads `\` as an escape and
  one that reads it as a plain byte end the literal at the same place. That closes the earlier F1
  (`quoteLiteral`, backslash) without measuring the Zetaris dialect.
- **Tests that prove it** (all green at `291753a`): `querytpl.TestEveryBuiltLiteralIsOneLiteralInEitherDialect`,
  `TestEachFormatRefusesAHostileValue`, `TestOnlyFormedValuesAreBuilt`, `TestRequestRefusalsAreNamed`,
  `TestRelationsAreCheckedAgainAtBuild`; `investigator.TestInvestigatorSendsNothingForAnAnswerThatIsNotATemplate`,
  `TestInvestigatorRunsTheQueryProvenBuilds`, `TestExecuteRefusesBeforeItSends`;
  `skeptic.TestAnswerThatIsNotATemplateProposesNoQuery`; `osv.TestCheckSendsNoQueryForAPackageOutsideItsFormat`;
  `pipeline.TestRefutingQueryThatIsNotATemplateIsRefused`.

#### Item 1 — the original reviewers' tests against this tip

Command: `go test -count=1 -v -overlay <scratch>/overlay-tests.json -run '…' ./internal/query/ ./internal/querytpl/`.

| original test (source) | at 291753a | how it was run |
|---|---|---|
| `TestOneSelectRefusesASecondStatementBehindACommentOrEscape` (p4-claude, text 2) | **PASS** | verbatim |
| `TestLocalEngineRefusesTheSameTexts` (p4-claude, text 2) | **PASS** | verbatim, renamed `…Orig`, since the leg landed one with the same name; the leg's copy (`select_gate_test.go:46`) also passes |
| `TestQuoteLiteralIsOneLiteralInEitherDialect` (p4-claude, text 1) | **PASS** | **does not compile verbatim**: `osv.quoteLiteral` was deleted. Ported: same value list and same assertion, applied to the literal `OSVPackageSQL` builds in Go, npm and PyPI |
| `TestModuleSQLKeepsThePackageInOneLiteral` (p4-claude, text 1) | **PASS** | ported: `moduleSQL` → `OSVPackageSQL` (deleted likewise). The value `trailing\` is now refused |
| `TestNonUTF8PackageIsRefusedBeforeQuery` (p4-codex) | **PASS** | ported: `quoteLiteral` → `OSVPackageSQL`, all three ecosystems |
| `TestOtherControlByteIsRefusedBeforeQuery` (p4-codex) | **PASS** | ported as above |
| `TestModelValueCannotChangeQueryPredicate` (p4-codex) | **FAIL**: `model value changed the WHERE predicate but the live adapter accepted it` | verbatim. **Expected, and not a defect of this change.** The test builds `… WHERE cve_id = 'CVE-2026-0001' OR cve_id = 'CVE-2026-0002'` by hand and asks the live gate to refuse it. That text is one well-formed SELECT, and no lexical gate can tell it from an intended query. ADR-0032 moves the property upstream: a model can no longer write that text. The same input as a model value (`"CVE-2026-0001' OR cve_id = 'CVE-2026-0002"`) is refused with `ReasonBadValue` by `investigator_test.go:223` (green) |

#### Item 2 — every path from a model answer, a manifest, a lock file or proven.yaml to a query

- **Model answer (Investigator):** `answer.Read` → `row` → `request()` → `Build`. The JSON keys match
  case-insensitively, so `"SQL"` and `"Sql"` are refused too. A duplicate `"sql"` key whose last value is empty is
  not refused, but the `SQL` field is never sent anywhere, so the property does not depend on that refusal.
- **Model answer (Skeptic):** `skeptic.Attempt` → `Validate` (default relations) → `Refutation.Query` →
  `pipeline.runRefutation` → `Execute` → `Build` with the review's relations. Checked twice, built once.
- **Manifests and lock files (Go, go.sum, npm, npm lock, PyPI, Poetry):** `manifest.Collect` → `osv.Check` → `oneModule`
  → `OSVPackageSQL`. The version never enters SQL (compared in Go). A refused name becomes a trace note quoted with
  `%q` (`check.go:168`).
- **OSV rows → KEV / evidence cells → EPSS:** each id goes through `ValidCVE` and again through `literal(cveFormat)`.
- **Settings (env, and proven.yaml where it feeds them):** table and view names go through `ValidRelation` at load
  (`osv/settings.go`, `epss/epss.go`) and again at build (`querytpl.go:281,346,366,378`). `pipeline.relations`
  passes the KEV view unvalidated (`osv.KEVViewOf`), and `Build` refuses it if it is not plain.
- **Result: no path reaches the SQL text other than through a checked template value**, a plain relation name or a
  closed ecosystem constant. Confirmed by `TestReviewEveryBuiltQueryPassesTheGateAndCarriesOnlyItsAlphabet`
  (scratch, package `query`). It ran 300 000 random values over a hostile alphabet (`'`, `''`, `\`, `"`, backtick,
  `;`, `--`, `/*`, `*/`, `#`, NUL, TAB, LF, CR, ESC, DEL, `\xff`, overlong `\xc0\xaf`, é, U+2028, U+FF1B, …) in every
  parameter of every template. Built: epss 1310, kev 1310, osv-by-advisory 1582, osv-by-package 10847. Every built
  query held only `[A-Za-z0-9 ._~/@=,*-']`, held exactly 2 quotes (4 for osv-by-package, which includes the
  ecosystem constant), and passed `oneSelect`. **PASS.**

#### Item 3 — each format

`TestReviewEveryByteInEveryPosition` (scratch, package `querytpl`) placed each of the 256 bytes, six multi-byte runes
(U+00E9, U+2028, U+202E, U+FF07, U+FF1B, U+FEFF), `--`, `/*`, `''` and an overlong `\xc0\xaf` at the start, middle
and end of a valid value of `package` (Go), `cve_id` (kev, epss) and `advisory_id`. It built 342 values; **every
built value held only `[A-Za-z0-9._~/@-]`. PASS.** So no value that passes a format can carry a quote, a backslash,
a control byte, invalid UTF-8 or a `;`. A comment marker `--` can pass the Go, npm and PyPI formats
(`example.com/a--b`); `cveFormat` and `advisoryFormat` do not admit it (each `-` must be followed by a letter or digit). Inside a literal
that holds no quote it is plain text in every dialect. The gate admits it inside a literal by design, and
`select_gate_test.go:29` pins that. `/*` cannot pass: no format admits `*`.

#### Item 4 — the sample review

A scratch test (`zz_review_sample_test.go`, package `main`, through `-overlay`) ran `run([]string{"scan","--sample",…})`
from a fresh `t.TempDir()` for every fixture at `291753a`:

```
discount exit=0       | finding F2 … | finding F4 … | finding osv:evil.example/leftpad … | finding F1 …
osv-go-module exit=0  | finding osv:example.com/widget …
refuted exit=0        | (no finding)
unknown-author exit=0 | (no finding)
unsafe-prompt exit=0  | (no finding)
```

Main's record (`git show 741cc77:docs/SAMPLE_MODE.md`, lines 41, 53, 65, 82, 169, 238 and the scenario sections)
lists the same finding headers, in the same order, with the same exit codes. **Same findings as main.** F2's
*evidence* changed by design: it was a model-written query over `advisories`, and it is now the
`osv-by-package` template over `px_osvfile.px_osv_pkg` with both range rows. `cmd/proven/main_test.go:56-66` pins
that. The instrument is weaker than a side-by-side run (Unrun U3).

#### Findings

**O1 — OWED (second line, not reachable from a model or the change) — `internal/query/live.go:185-187,197-220`;
`docs/QUERIES.md` § "The second check".**
- *Claim:* "read strictly enough that **no dialect's** comment or escape rule can hide a second statement".
- *Defect:* the gate tracks only `'`. In a dialect that reads `"…"` as a string literal (Spark SQL's default, MySQL)
  or `` `…` `` as a quoted identifier, a `'` inside them is a plain byte, so the gate's single-quote parity and the
  engine's disagree. The scratch test `TestReviewGateSeesAStatementBehindAnotherQuoteCharacter` **fails** at
  `291753a`: `oneSelect("SELECT \"'\" FROM t; DELETE FROM t WHERE a = \"'\"") = true`,
  `oneSelect("SELECT `'` FROM t; DELETE FROM t WHERE `'` = 1") = true`, and `oneSelect("SELECT \"'\" FROM t -- '\n") = true`
  (a comment the gate says it refuses).
- *Reach:* none from a model or the change at this tip. No format admits `"` or backtick, and the 300 000-value
  property above shows that no built query holds one. The gate now guards only `livecheck`'s constant probe. That is
  why this is owed and not blocking.
- *Remedy:* refuse `"` and `` ` `` anywhere in the text (no template needs either), then land the scratch test.
  Reword `live.go:185-187` and the QUERIES.md paragraph to say what the gate actually reads.

**O2 — OWED (R3, an unguarded guard) — `internal/querytpl/querytpl.go:242-246`.** The byte loop in `literal`
exists so that "a format edited carelessly cannot let one through" (ADR-0032 point 3). Removing it turns nothing
red (mutant `backstop`, below). *Remedy:* one test that calls `literal` with a permissive format such as
`regexp.MustCompile("(?s)^.*$")` and asserts that `'`, `\`, a control byte and a byte ≥ 0x80 are each refused.

### Unplug

Mutants run through `-overlay` over `internal/querytpl`, `osv`, `epss`, `agent/investigator`, `agent/skeptic`. The
querytpl mutant is the tip's file with four guards switchable by an overlaid constant. The equivalence control
(nothing switched) gave **all five packages ok**.

| guard removed | what turned red |
|---|---|
| `investigator.go:52` SQL-text refusal (`if false && …`) | `TestInvestigatorSendsNothingForAnAnswerThatIsNotATemplate/sql_text` and `/sql_beside_a_template` — **the master's claim holds** |
| `querytpl.go:239` format match | `TestEachFormatRefusesAHostileValue`, `TestPackageFormatsFollowTheEcosystem` |
| `querytpl.go:242-246` byte backstop | **nothing — survives** (O2) |
| `querytpl.go:349-353` ecosystem allow-list | `TestPackageFormatsFollowTheEcosystem` |
| `querytpl.go:201` `ValidRelation` (always true) | `TestRelationsAreCheckedAgainAtBuild`, `osv.TestLoadTableRefusesABadName`, `osv.TestLoadViewEmptyIsTheDefaultAndABadNameIsRefused`, `epss.TestLoadViewEmptyIsTheDefaultAndABadNameIsRefused` |

Not unplugged by a run: the Skeptic's `Validate` (`skeptic.go:361`), `pipeline.go:724-731` (a settings refusal at
`Execute`), and each new branch of `oneSelect`. My reading of the code says `TestAnswerThatIsNotATemplateProposesNoQuery`
reddens on the first, and the three `TestSelectGate*` tests on the third. **UNMEASURED**: each needs a full mutant
copy of `skeptic.go`, `pipeline.go` or `live.go`, and I spent the budget on the querytpl mutants. I know of no test
that pins the second: the Skeptic's early `Validate` clears a bad request before `Execute` can refuse it, so that
branch fires only for a bad relation setting.

### Closure

- **Property (item 1):** holds. Guaranteed at `querytpl.go:238-248` (`literal`), `:266-294` (`Build`) and
  `investigator.go:51-56,94,64`; `skeptic.go:357-365`. Proven by the tests named in Differential and by the
  two scratch property tests.
- **Earlier F1 (backslash in `quoteLiteral`):** closed. The function is gone, and a backslash is refused by every
  format and by the backstop. Ported test text 1 passes.
- **Earlier F2 (lexically incomplete live gate):** closed for model text, because none reaches the gate. The gate
  itself still fails another spelling (O1), which is not reachable today.
- **Codex's blocking condition ("bound values or a constrained query representation"):** met by the template
  representation. Its verbatim test stays red by design (see Differential).
- **Codex's owed (invalid UTF-8 and ESC in deterministic literals):** closed (ported tests pass).
- **The master's 2026-10-23 claims (law #44):** "SQL-text refusal disabled → `TestInvestigatorSendsNothingForAnAnswerThatIsNotATemplate`
  red" — **re-measured, holds**. "34 packages ok" — **not reproducible on this seat**: 28 ok and 6 FAIL, and all 6
  failures are the loopback-listener panic. 28 + 6 = 34 packages with tests, so the count is consistent with the
  master's, but I did not see the 6 pass. master-was-wrong: none measured.

### Remainder

- **Write-set conflict, shouted (law #7).** The order puts artifacts under `██████████████` in the worktree and
  also requires `git status --porcelain` to show nothing of mine. `████████` is not gitignored here (`.gitignore`
  read), and this role has no `rm`. The previous reviewer met the same conflict and left `?? ████████` behind. I
  used the run's artifact directory outside the worktree instead (an admitted additional working directory). The
  worktree is unchanged except for my two channel files. The master decides which path stands.
- **The order's H1 for `_REVIEW_STATUS.md`** reads `# REVIEW — query-evidence-comes-from-templates-review`, while
  the engine-injected grammar demands `# REVIEW — query-evidence-comes-from-templates` (the reviewed stream). I
  followed the engine, because the verdict binds to the H1 slug.
- **"Every query Proven sends has one author" (QUERIES.md:6, querytpl.go:9-10)** is broad. `cmd/livecheck/main.go:170`
  builds its probe outside `querytpl`, from a validated view and no external value. QUERIES.md § "The second check"
  names it, so this is a remark, not a finding.
- **`osv-by-advisory`'s view has no setting** (`OSVIDView` is never set from config; QUERIES.md says "none"). It was
  a hard-coded name in the old prompt too, so this is no regression. Noted against ADR-0013.
- **Declined to judge (R11):** licences (out of scope by order); the trace and terminal encoding of the `SQL:` lines
  beyond the `%q` note at `check.go:168`; `hmem`; the R8 size of `querytpl.go` (386 lines, one responsibility);
  whether legitimate package names exist that the formats refuse (e.g. a Go path with `+`, an npm name of more than
  214 characters). A refused name is a visible trace note ("osv query refused"), not a silent pass.
- **Dependency delta (R4):** none. `go.mod` is untouched and there is no new import outside the module.
- **Scan for invisible or bidirectional characters in the diff (S4):** not run as a separate instrument. The scratch
  probes put U+202E and U+FEFF into values and both were refused.

### Profile

security, per R9: a boundary between untrusted input (model answers, the change's manifests, evidence rows) and a
query engine.
- **S1 trust-boundary map.** (a) Model → template + params → `Build`. Tampering with query meaning is closed by the
  closed template list and the ASCII formats. A model can still choose WHICH of four questions to ask and about
  WHICH id; that is the intended affordance. (b) The change → manifest readers → `OSVPackageSQL`, closed the same
  way. (c) Evidence rows → CVE ids → KEV/EPSS, closed by `cveFormat`. (d) Settings → relation names, closed by
  `ValidRelation`, checked twice. (e) Built SQL → `Live.Query` → Zetaris. The gate is a second line, imperfect (O1)
  but unreachable. ADR-0032 point 4: the docs claim no read-only key (`SECURITY.md:64`, `QUERIES.md` § "What this
  page does not claim"), and I confirmed that sentence is present.
- **S2.** O1 is reproduced by a failing test of the expected behaviour, not by an exploit (law #56).
- **S3.** Each new refusal was removed once, in Unplug. One guard survives (O2).
- **S4.** No dependency delta. No secrets in the diff: `git diff 741cc77...291753a | grep -E '^\+' | grep -iE
  'api[_-]?key|token|secret|password|bearer'` returned 3 lines. They are a file name (`missing_secret_test.go`) and
  two fixture prompts that hold the sample title "note may include a payment token". No destructive path.

### Focus

*Can a value from the change or a model change a query's meaning, or can model-written query text reach Zetaris?* —
**No**, at `291753a`. Every value is refused or becomes one quote-free, backslash-free ASCII literal, identically in
standard and backslash-escaping readings. Model text other than validated values never enters the SQL, and the
`sql` field of an answer is refused. What remains is the live gate's overbroad claim (O1) and one untested guard (O2).

### Conditions

- **O1 — owed.** `oneSelect` refuses `"` and `` ` ``, and `TestReviewGateSeesAStatementBehindAnotherQuoteCharacter`
  (text in this run's scratch directory) lands in `internal/query` and passes. `live.go:185-187` and `docs/QUERIES.md`
  § "The second check" are reworded to what the gate reads.
- **O2 — owed.** A test in `internal/querytpl` calls `literal` with a permissive format and shows that the byte
  backstop refuses `'`, `\`, a control byte and a byte ≥ 0x80. Removing `querytpl.go:242-246` must turn it red.
- **Owed by the order, not by the code:** the second, independent NVIDIA review through Nightshift (the owner's rule
  of 2026-10-22). Its route is down today, so this review ran alone.

### Unrun

- **U1** — the httptest-based tests of `cmd/livecheck`, `internal/generator`, `internal/guardrail` and the 16 skipped
  in `internal/pipeline`, `internal/query`, `internal/review` (including `TestNonSelectIsNotSent`). They failed with
  `httptest: failed to listen on a port: listen tcp6 [::1]:0: bind: operation not permitted`.
- **U2** — the mutants of `skeptic.go:361`, `pipeline.go:724-731` and the `oneSelect` branches (see Unplug).
  Reason: budget.
- **U3** — a side-by-side run of main's binary on main's fixtures. `git worktree add` into the scratch directory was
  refused: `fatal: could not create directory of '██████████████████████████████████████████████████████████████████████████████': Operation not permitted`.
  Main's side was read from its record (`docs/SAMPLE_MODE.md` at `741cc77`) instead.
- **U4** — any Zetaris dialect measurement and the read-only-key question (ADR-0032 point 4). No network and no
  credentials in this order's boundary. The property no longer depends on the dialect (Differential, item 1).
- **U5** — the independent NVIDIA review (order: owed, route down).

## The review status as filed

## REVIEW — query-evidence-comes-from-templates
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: claude-code claude-opus-5-5 named review of W32 at 291753a, verdict accept-with-conditions. The property holds: every value is refused or becomes one quote- and backslash-free ASCII literal (querytpl.go:238-248), and no model text reaches the SQL. Proven by the leg's tests and by two 300k/256-byte scratch probes. Original reviewers' tests: 6 pass (4 ported), and codex's gate-predicate test stays red by design.
  Owed O1: the live gate's comment overclaims, since double quotes and backticks hide a second statement (test red, not reachable from a model or the change). Owed O2: the literal byte backstop is unguarded (mutant survives).
  go build and vet ok. go test: 28 ok and 6 FAIL, all 6 from the seat's loopback-listener refusal; touched packages ok with the listener tests skipped. Sample findings match main's record. The NVIDIA second review is owed (route down).
- artifacts:
    - _REVIEW_NOTES.md
    - _REVIEW_STATUS.md
- updated: 2026-10-23T23:00:00Z

## REDACTED

| what | times |
|---|---|
| absolute or private path | 6 |
