# ORDER — org-policy · round 6

- change: org-policy
- round: 6
- base: main @ 16cf7b7
- executor: claude-code
- written: 2026-10-24T10:10:51Z
- task: round 2 — a repository cannot lower the organisation policy across fields, and the two negative controls and a time-zone boundary are proved

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — org-policy
track: proven-enterprise
- train: proven-enterprise
- wagon: W3
- review: named
- repo: proven
- base: main @ 16cf7b7
- branch: fix/org-policy
- lands-on: main
- backend: claude-code
- quote-usd: 12.00
- cap-usd: 24.00
- wall: 150m
- task: round 2 — a repository cannot lower the organisation policy across fields, and the two negative controls and a time-zone boundary are proved

█████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████
ROLE-LOCK: YOU = EXECUTOR. One-shot headless: foreground commands only; your first act writes `_EXEC_STATUS.md`
(`- state: in-progress`). A permission denial is a fact: try once, then `- state: blocked`.
COMMIT EVERY PART BEFORE `stop-2-ready` · ZERO push/merge/PR · author = fleet identity · commit message `[org-policy] <area>: <what>`.
No AI attribution anywhere: no `Co-authored-by:` line, no tool or model name in a commit.
After `git commit`, `packed-refs.lock: Permission denied` means the commit LANDED — check `git log -1`.

### SKILLS
- ██████████████████████████████████
- ██████████████████████████████████████

Craft: `██████████████████████████` + `████████████████████████████` — read first.

### THE OWNER'S WORDS
2026-10-23: the decision is ADR-0036 — *"A repository may make the policy stricter, never weaker; a weaker local rule is
refused by name."* 2026-10-24: Cursor seats first for coding.

### THE REVIEW OF ROUND 1 (tip c7ebc39) — PASTED (law #44)
Verdict: reject. Its findings, verbatim in substance:
- **BLOCKING · cross-field downgrade:** `internal/review/policy.go:285`, `:265`, callers `internal/review/open.go:222`
  and `internal/review/decide.go:93`: policy `fallback_class: security` and local `path_rules: **=docs` pass
  `parseLocal`, because `weaker("path_rules", …)` checks the local rules only against the policy's rules, not against the
  policy's fallback. A repository lowers every unmatched path below the organisation's floor. The existing test
  `TestEachPolicyKeyRefusesAWeakerValueAndAcceptsAnEqualOrStricterOne` varies one key at a time, so it cannot see it.
- **BLOCKING · negative controls not recorded:** `internal/review/policy_test.go:77`, `:170` — no isolated mutation of
  the weaker-than check and of the expiry comparison, with the failing test output, was recorded.
- **OWED · time zone:** `policy_test.go:15`, `policy.go:420` — the expiry boundary is tested only in UTC; add two
  non-UTC zones on both sides of midnight.
- **BLOCKING · JSON:** accepted findings must appear in JSON; `FormatterFor("json")` refuses (no writer on main). NOT in
  this round: the master asked the owner to rule on it; do not build a JSON writer.

### WHAT TO BUILD
1. **Rebase first:** `git rebase --onto main 16cf7b7`. Do not merge.
2. **Effective policy comparison:** compare the EFFECTIVE class every path would get (local rules + fallback) against the
   policy's effective class, not key by key: a local rule, a local fallback or a local class order that gives any path
   a lower class than the policy would is refused by name (key, the path pattern, both classes, the policy file).
3. A test for the reviewer's case (`fallback_class: security` in the policy, `path_rules: **=docs` locally) and one
   more cross-field case of your choice.
4. Run and record the two isolated negative controls: remove the weaker-than check, then (restored) the expiry
   comparison; paste each failing test name and output into your status.
5. The time-zone tests (two zones, both sides of midnight).
6. Update `docs/POLICY.md`: the effective-class rule, with the reviewer's example.

### ACCEPTANCE
1. `go test -count=1 ./...` (whole module) green; `go vet ./...` clean.
2. The cross-field tests fail on c7ebc39's behaviour and pass after.
3. The two negative controls are recorded by name with output; the time-zone tests pass.
4. Subjects match `^\[org-policy\] [a-z0-9/._-]+: .+$`, at most 72 characters; no `Co-authored-by` line.

### WHAT IS NOT IN SCOPE
- A JSON, SARIF or PR-comment writer (owner's ruling pending; the CLI train's writers). W1, W2. Licences.

### WHEN THIS ORDER CANNOT PROCEED
Write `SHOUT_<topic>.md` in the worktree, set `- state: blocked`, and stop.

### ARTIFACT CONFINEMENT
Every artifact this leg makes — captures, logs, evidence, scratch — lives under `████████████████████████` in the worktree.
Nothing outside the worktree; nothing committed unless this order names it.

### RETURN
`_EXEC_STATUS.md`: `- state: stop-2-ready`, `- last commit:` <sha>, an `- abstract:` of at most 5 lines, the cross-field
rule, the negative-control outputs, the whole-module counts. Check it with
`saphan doc check --family exec-status _EXEC_STATUS.md` before you end.
Read every `saphan` verb by its `RESULT:` line, its `--json` or its exit code, never by its sentences (`craft/COMMON.md`).

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| how an order is bound to the laws | 1 |
| private workspace path | 4 |
