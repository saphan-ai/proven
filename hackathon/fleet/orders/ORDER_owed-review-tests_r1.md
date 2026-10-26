# ORDER — owed-review-tests · round 1

- change: owed-review-tests
- round: 1
- base: main @ b17d30f
- executor: cursor on runner-cursor
- written: 2026-10-25T04:53:19Z
- task: the tests and small fixes four named reviews left owed (W43, W47, W49, W52), each with a test that turns red when its line is removed

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — owed-review-tests
track: proven-reviewer-core
- train: proven-reviewer-core
- wagon: W56
- review: master-verify-only
- repo: proven
- base: main @ b17d30f
- branch: fix/owed-review-tests
- lands-on: main
- backend: cursor
- model: grok-4.7-high
- quote-usd: 5.00
- cap-usd: 10.00
- wall: 180m
- task: the tests and small fixes four named reviews left owed (W43, W47, W49, W52), each with a test that turns red when its line is removed

█████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████
ROLE-LOCK: YOU = EXECUTOR. One-shot headless: foreground commands only; your first act writes `_EXEC_STATUS.md`
(`- state: in-progress`). A permission denial is a fact: try once, then `- state: blocked`.
COMMIT EVERY PART BEFORE `stop-2-ready` · ZERO push/merge/PR · author = fleet identity · commit subject at most 72
characters, `[owed-review-tests] <area>: <what>` included, the area ONE token. No AI attribution anywhere. The tool re-adds a
`Co-authored-by:` line on amend; if it is there at the end, say so (the master removes it).
After `git commit`, `packed-refs.lock: Permission denied` means the commit LANDED — check `git log -1`.

### SKILLS
- ██████████████████████████████████
- ██████████████████████████████████████

Craft: `██████████████████████████` + `████████████████████████████` — read first.

### WHY THIS WAGON
Four named reviews (claude-opus-5-5, 2026-10-24/25) accepted W43, W47, W49 and W52 with owed, non-blocking items. They are
pasted below, verbatim in substance. Their FILE:LINE are the reviewers' readings at the reviewed commits; main has moved
since (absorbs only) — re-measure each site, and say in `- master-was-wrong:` where one is not where it was claimed.

### THE ITEMS (do each; one commit per source review is fine)
From W47 (the-source-adapter-interface):
- O1: the check in `source.Build` (internal/source/registry.go ~117-123) is pinned by no test. Put a registered
  concatenating builder through `source.Build` and assert `*source.ValueError`. Mutated field: the parameter value.
- O2: in the Investigator, `res.Columns, res.Rows = kept.Columns, kept.Rows` (investigator.go ~140) has no test. A rule
  that keeps one row of two, with an assertion on the evidence row count.
- O3: the byte floor in `CheckValue` — add cases for a backslash and for a byte above 0x7e (quote, NUL, newline exist).
From W49 (couchdb-evidence-store):
- RC2: the parse-error blanking in internal/store/couchsettings.go (~180-182) protects nothing the `@` guard does not;
  keep it with a one-line comment that it is defensive only. Add a case pinning `http://user@host` (user name, no
  password) refused with an empty Value.
From W52 (the-query-seam-carries-the-template):
- R1: a unit test that `downStore.Run` (internal/review/open.go ~296) given a refused request (unknown template, bad CVE)
  returns `*querytpl.Refusal`, not `*service.Down`. Mutated: the BuildForms call.
- R3: a review-level `zetaris` subtest in internal/review/form_seam_test.go with a recorded Zetaris server, asserting form
  `sql`, the Zetaris engine and the advisory row, so replacing `store.NewZetaris` in open.go turns it red.
- R4: internal/osv/check.go (~237 and missNote ~365) prints `SQL: <res.SQL>`, which on CouchDB is Mango JSON. Label with
  the form (`query (mango): …` / `query (sql): …`). Sample goldens: if any move, regenerate and say so.
- R5: ████████████████████████████████████ is a package clause plus a comment: delete it.
From W43 (more-evidence-through-zetaris):
- C1: extend `TestStoresAgreeOnAMissingTable` (internal/store/missing_test.go) to iterate every template of
  `querytpl.Catalogue()`, not only `mal-by-package`.
- C2: a review test over the `malicious-package` fixture asserting `ProofLevel == proved-by-data` and its count is 1.
  Negative control: map `EvidenceQuery` to `ProvedByTool` in `ProofLevelOf` → red.
- O2: internal/osv/check.go `Binds`: mal/scorecard/spdx key on the model's raw `package` param while `noteRaised` uses the
  manifest spelling; key both sides on the ecosystem-normalised name, with a PyPI case whose spellings differ.

### ACCEPTANCE
1. Reproduce first: for each new test, show it red against the mutation it names (or, for a pure test gap, show the
   mutant surviving at your base), then green.
2. `go build ./... && go vet ./... && go test -count=1 ./...`: green, or each listener-refused package named UNRUN with
   the quoted refusal (the master runs those).

### WHAT IS NOT IN SCOPE
The text renderer's level of proof (the CLI track's); any change of behaviour beyond R4's label and O2's key.

### COMMENTS FOR THE JUDGES (ADR-0027, a standing rule of this train)
(1) every package you touch keeps its package comment true; (2) every exported name has a doc comment beginning with the
name; (3) every step a reader would stop at has a comment saying WHY; (4) code that carries out a decision names it;
(5) every test says which property it proves; (6) English, plain; no AI attribution, no host names, internal paths or
credential names, no commented-out code.

### WHEN THIS ORDER CANNOT PROCEED
Write `SHOUT_<topic>.md` in the worktree, set `- state: blocked`, and stop.

### ARTIFACT CONFINEMENT
Every artifact this leg makes lives under `███████████████████████████████` in the worktree, except the files this order names.

### RETURN
`_EXEC_STATUS.md`: `- state: stop-2-ready`, `- last commit:` <sha>, an `- abstract:` of at most 5 lines, `- master-was-wrong:`,
one line per item (item id, test name, the mutated line, before/after exit codes). Check it with
`saphan doc check --family exec-status _EXEC_STATUS.md` before you end: its exit code is the verdict.

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| file and line in Saphan Studio's private code | 1 |
| how an order is bound to the laws | 1 |
| private workspace path | 4 |
