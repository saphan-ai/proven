# ORDER — more-evidence-through-zetaris · round 1

- change: more-evidence-through-zetaris
- round: 1
- base: main @ 8770336
- executor: claude-code
- written: 2026-10-24T03:39:08Z
- task: malicious packages, project health, CVSS and SPDX licences as Zetaris tables read through query templates (ADR-0039)

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — more-evidence-through-zetaris
track: proven-reviewer-core
- train: proven-reviewer-core
- wagon: W43
- review: named
- repo: proven
- base: main @ 8770336
- branch: fix/more-evidence-through-zetaris
- lands-on: main
- backend: claude-code
- model: claude-sonnet-5-5
- quote-usd: 6.00
- cap-usd: 12.00
- wall: 180m
- task: malicious packages, project health, CVSS and SPDX licences as Zetaris tables read through query templates (ADR-0039)

█████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████
ROLE-LOCK: YOU = EXECUTOR. One-shot headless: foreground commands only; your first act writes `_EXEC_STATUS.md`
(`- state: in-progress`). A permission denial is a fact: try once, then `- state: blocked`.
COMMIT EVERY PART BEFORE `stop-2-ready` · ZERO push/merge/PR · author = fleet identity · commit subject at most 72
characters, `[more-evidence-through-zetaris] <area>: <what>` included.
No AI attribution anywhere.
After `git commit`, `packed-refs.lock: Permission denied` means the commit LANDED — check `git log -1`.

### SKILLS
- ██████████████████████████████████
- ██████████████████████████████████████

Craft: `██████████████████████████` + `████████████████████████████` — read first.

### THE OWNER'S WORDS
2026-10-21 (translated from Polish): *"Working first, pretty later."* 2026-10-23: the judges read this code (ADR-0027).

### THE OWNER'S DECISION (ADR-0039, docs/adr/ADR-0039-more-evidence-sources.md) — read it whole:
Four more sources, each loaded into Zetaris and queried from a template like the first ones:
1. OpenSSF Malicious Packages (OSV entries with the `MAL-` prefix): a dependency published on purpose to harm is reported
   as such, with the advisory's row as evidence.
2. Project health (OpenSSF Scorecard, through deps.dev): a newly added dependency with no maintainers, unsigned
   releases, or a release only days old is reported, with its scores as evidence.
3. CVSS from NVD, shown beside EPSS: how severe a vulnerability is, next to how likely it is to be exploited.
4. Licences (SPDX): a new dependency whose licence the organisation's policy does not allow is reported (ADR-0036).
Each source has a sample table, so sample mode covers it with no network.
And ADR-0032: no model-written SQL reaches Zetaris; the model picks a template and values, Proven validates each value
against a strict format and builds the SQL itself.

### MASTER'S READINGS (claims, measured by the master 2026-10-24 at main 8770336 — re-measure, you may refute them)
- Templates live in `internal/querytpl/querytpl.go`: `Catalogue()`, `Build(req, rel)`, `Validate`, the relation
  defaults (`DefaultKEVView`, `DefaultEPSSView`, …) and the refusal reasons. New sources add templates and relations
  there, with the same value formats and refusals.
- Sample tables are built by the fixture SQL (`internal/fixture`).
- The live tables are loaded by the Zetaris track, not by this wagon; their names are not known yet. Use defaults that
  follow the existing `px_evidence.*` naming, overridable by settings like the existing relations.
- Whether ADR-0036's policy is in code at your base is UNMEASURED; if there is no policy, report licences as evidence
  with no allow/deny verdict and say so.

### WHAT TO BUILD
1. One template per source (malicious-by-package, health-by-package, cvss-by-cve, licence-by-package), each with
   strict value formats and relation defaults overridable by settings (ADR-0013).
2. Sample tables and one sample scenario per source; the findings carry the row as evidence.
3. A missing table or view is a named down source in the review's failed services, never a silent empty answer.
4. CVSS shown beside EPSS wherever EPSS is shown.

### ACCEPTANCE
1. Reproduce first: before changing anything, run `go test -count=1 ./...` at your base and paste the summary.
2. `go build ./... && go vet ./... && go test -count=1 ./...`: green. Paste the summary lines.
3. Each source has a template test that refuses a value outside its format and a sample scenario test.
4. **Negative control (law #27):** drop one sample table → that source is named down (test red if it is silent);
   restore. Name the mutated field in your status.

### COMMENTS FOR THE JUDGES (ADR-0027, a standing rule of this train)
(1) every package you touch keeps its package comment true; (2) every exported name has a doc comment beginning with the
name; (3) every step a reader would stop at (a guard, an order that matters, a value from a setting, a refusal) has a
comment saying WHY, never one restating the code; (4) code that carries out a decision names it (e.g. `ADR-0039: …`);
(5) every test says which property it proves; (6) English, plain; no AI attribution, no host names, internal paths or
credential names, no commented-out code.

### WHAT IS NOT IN SCOPE
Loading the live Zetaris tables (the Zetaris track); the organisation policy itself (ADR-0036, its own wagon).

### WHEN THIS ORDER CANNOT PROCEED
Write `SHOUT_<topic>.md` in the worktree, set `- state: blocked`, and stop.

### ARTIFACT CONFINEMENT
Every artifact this leg makes lives under `███████████████████████████████████████████` in the worktree, except the
files this order names.

### RETURN
`_EXEC_STATUS.md`: `- state: stop-2-ready`, `- last commit:` <sha>, an `- abstract:` of at most 5 lines, `- master-was-wrong:`,
the commands run and their exit codes, and the session recording or the reason there is none (ADR-0024). Check it with
`saphan doc check --family exec-status _EXEC_STATUS.md` before you end: its exit code is the verdict.

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| how an order is bound to the laws | 1 |
| private workspace path | 4 |
