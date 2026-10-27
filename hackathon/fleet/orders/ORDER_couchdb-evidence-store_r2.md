# ORDER — couchdb-evidence-store · round 2

- change: couchdb-evidence-store
- round: 2
- base: main @ 5faabc9
- executor: claude-code
- written: 2026-10-24T16:13:21Z
- task: round 2: a refused CouchDB URL never carries user info; settings text says a live review names couchdb down; a review-level test of that (ADR-0047)

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — couchdb-evidence-store
track: proven-reviewer-core
- train: proven-reviewer-core
- wagon: W49
- review: named
- repo: proven
- base: main @ 5faabc9
- branch: fix/couchdb-evidence-store
- lands-on: main
- backend: claude-code
- quote-usd: 4.00
- cap-usd: 8.00
- wall: 120m
- task: round 2: a refused CouchDB URL never carries user info; settings text says a live review names couchdb down; a review-level test of that (ADR-0047)

█████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████
ROLE-LOCK: YOU = EXECUTOR. One-shot headless: foreground commands only; your first act writes `_EXEC_STATUS.md`
(`- state: in-progress`). A permission denial is a fact: try once, then `- state: blocked`.
COMMIT EVERY PART BEFORE `stop-2-ready` · ZERO push/merge/PR · author = fleet identity · commit subject at most 72
characters, `[couchdb-evidence-store] <area>: <what>` included, the area ONE token. No AI attribution anywhere.
After `git commit`, `packed-refs.lock: Permission denied` means the commit LANDED — check `git log -1`.

### SKILLS
- ██████████████████████████████████
- ██████████████████████████████████████

Craft: `██████████████████████████` + `████████████████████████████` — read first.

### WHY THIS ROUND
Your branch is at ae8167d. The named review (claude-opus-5-5, 2026-10-24) returned the conditions below, pasted verbatim.
Each one is yours in this round. PRIVATE until merged: no review wording in commit messages, test names, comments or docs
— name the PROPERTY that now holds.

### THE REVIEW'S CONDITIONS (pasted)

- **C1 — BLOCKING (property 4).** `internal/store/couchsettings.go:173-174`: a refused `PROVEN_COUCHDB_URL` must not carry user info in its Value. Remedy: when `u.User != nil` (or the parse fails), refuse with Value `""` or the URL with user info replaced (`u.Redacted()` / scheme+host only), and add to `TestLoadCouchEmptyIsADownStoreAndWrongIsRefused` an assertion that `http://u:p@c`'s refusal text contains neither `u:p` nor `p@`. Negative control: restore `clip(raw)` → the assertion fails. Mutated field: `PROVEN_COUCHDB_URL`.
- **C2 — owed, user-facing.** `internal/store/settings.go` (PROVEN_EVIDENCE_STORE description), `internal/settingspage/catalog.go:44` (`evidenceAfter`) and the generated `docs/CONFIGURATION.md`: say that with `couchdb` a live review names couchdb down on every query until the query seam passes templates, as `docs/ADAPTERS.md` already does; drop "its evidence is a Mango query" for the live review or qualify it.
- **C3 — owed, test.** Add a review-level test: Open with `PROVEN_EVIDENCE_STORE=couchdb` and valid settings, run a query through the installed adapter (or a live review over a fixture that asks one), and assert a `service.Down{Service:"couchdb"}` with the seam reason. Negative control: `couchdbSQL.Query` returning `query.Result{}, nil` → the test fails.

### ACCEPTANCE
1. Reproduce first: each condition's test red at ae8167d; paste the failure.
2. `go build ./... && go vet ./... && go test -count=1 ./...`: green, or each listener-refused package named UNRUN with
   the quoted refusal (the master runs those).
3. Each fix reverted → its test red (law #27); name the mutated field or line.

### COMMENTS FOR THE JUDGES (ADR-0027)
Every test says which property it proves; English, plain; no AI attribution, host names, internal paths or credential
names.

### BOUNDARY
Own code only; no network beyond what `go test` needs locally; describe and prove the property — a failing test is the
proof.

### WHEN THIS ORDER CANNOT PROCEED
Write `SHOUT_<topic>.md` in the worktree, set `- state: blocked`, and stop.

### ARTIFACT CONFINEMENT
Every artifact this leg makes lives under `████████████████████████████████████` in the worktree, except the files this order names.

### RETURN
`_EXEC_STATUS.md`: `- state: stop-2-ready`, `- last commit:` <sha>, an `- abstract:` of at most 5 lines, `- master-was-wrong:`,
one line per condition (property, test name, before/after exit codes). Check it with
`saphan doc check --family exec-status _EXEC_STATUS.md` before you end: its exit code is the verdict.

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| how an order is bound to the laws | 1 |
| private workspace path | 4 |
