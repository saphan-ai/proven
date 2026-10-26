# ORDER — rust-is-the-second-new-language · round 1

- change: rust-is-the-second-new-language
- round: 1
- base: main @ 2afad66
- executor: cursor on runner-cursor
- written: 2026-10-24T07:34:26Z
- task: Rust (Cargo.toml and Cargo.lock) through the language adapter, with its sample scenario (ADR-0038)

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — rust-is-the-second-new-language
track: proven-reviewer-core
- train: proven-reviewer-core
- wagon: W39
- review: master-verify-only
- repo: proven
- base: main @ 2afad66
- branch: fix/rust-is-the-second-new-language
- lands-on: main
- backend: cursor
- model: grok-4.7-high
- quote-usd: 3.00
- cap-usd: 6.00
- wall: 120m
- task: Rust (Cargo.toml and Cargo.lock) through the language adapter, with its sample scenario (ADR-0038)

█████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████
ROLE-LOCK: YOU = EXECUTOR. One-shot headless: foreground commands only; your first act writes `_EXEC_STATUS.md`
(`- state: in-progress`). A permission denial is a fact: try once, then `- state: blocked`.
COMMIT EVERY PART BEFORE `stop-2-ready` · ZERO push/merge/PR · author = fleet identity · commit subject at most 72
characters, `[rust-is-the-second-new-language] <area>: <what>` included — the slug is long, so keep `<what>` short.
No AI attribution anywhere. The tool re-adds a `Co-authored-by:` line on amend; if it is there at the end, say so (the
master removes it).
After `git commit`, `packed-refs.lock: Permission denied` means the commit LANDED — check `git log -1`.

### SKILLS
- ██████████████████████████████████
- ██████████████████████████████████████

Craft: `██████████████████████████` + `████████████████████████████` — read first.


### THE OWNER'S WORDS
2026-10-21 (translated from Polish): *"Working first, pretty later."* 2026-10-23: the judges read this code (ADR-0027).

### THE OWNER'S DECISION (ADR-0038) — one language is one small adapter; Rust is the second new language after PHP.

### THE RECIPE (docs/ADAPTERS.md "Adding a language", landed with W38 — read it whole and follow it step by step)
1. A comparer for Cargo versions (semver), with tests for equal, lower, higher and a refused string.
2. A manifest reader (`Cargo.toml`: `[dependencies]`, `[dev-dependencies]`) and a lock reader (`Cargo.lock`:
   `[[package]]` name/version; `Lock` and `NoteErr` set). OSV ecosystem name: `crates.io`.
3. `Match` on the file's own name at any depth; a backup such as `Cargo.lock.bak` refused.
4. `Diff`: added packages and raised versions only.
5. One line in `Readers` for each reader.
6. The ecosystem in the `osv-by-package` template (`internal/querytpl`) with the crate-name format; QUERIES.md updated.
   Recorded prompts that list the ecosystems change: regenerate the affected fixtures (a stale planner hash prints the
   new one; recorded prompts are matched by exact text) and say which.
7. OSV_TABLE.md; a sample scenario under `testdata/fixture/osv-rust-cargo` and a section in SAMPLE_MODE.md; the
   scenario's test fails when the adapter is taken out of `Readers`.
8. A section "Worked example: Rust" is NOT needed in ADAPTERS.md (PHP is the worked example); add Rust to its list of
   languages only.

### ACCEPTANCE
1. Reproduce first: before changing anything, run `go test -count=1 ./...` at your base and paste the summary.
2. `go build ./... && go vet ./... && go test -count=1 ./...`: green.
3. **Negative control (law #27):** take the two Cargo readers out of `Readers` → the Rust scenario test red; restore.

### COMMENTS FOR THE JUDGES (ADR-0027, a standing rule of this train)
Every exported name has a doc comment; every test says which property it proves; English, plain; no AI attribution,
host names, internal paths or credential names. Commit subjects `[rust-is-the-second-new-language] <area>: <what>`, the
area ONE token, AT MOST 72 CHARACTERS IN TOTAL (the merge refuses anything else — this slug is long, keep <what> short).

### WHEN THIS ORDER CANNOT PROCEED
Write `SHOUT_<topic>.md` in the worktree, set `- state: blocked`, and stop.

### ARTIFACT CONFINEMENT
Every artifact this leg makes lives under `█████████████████████████████████████████████` in the worktree, except the
files this order names.

### RETURN
`_EXEC_STATUS.md`: `- state: stop-2-ready`, `- last commit:` <sha>, an `- abstract:` of at most 5 lines, `- master-was-wrong:`,
the commands run and their exit codes. Check it with `saphan doc check --family exec-status _EXEC_STATUS.md` before you
end: its exit code is the verdict.

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| how an order is bound to the laws | 1 |
| private workspace path | 4 |
