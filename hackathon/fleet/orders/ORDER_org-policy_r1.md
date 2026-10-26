# ORDER — org-policy · round 1

- change: org-policy
- round: 1
- base: main @ 9303b2d
- executor: claude-code on macm5
- written: 2026-10-23T18:59:17Z
- task: an organisation policy file sets the floor for every repository's proven.yaml, and a known risk is accepted only by an exception with an owner and an expiry

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — org-policy
track: proven-enterprise
- train: proven-enterprise
- wagon: W3
- review: named
- repo: proven
- base: main @ 9303b2d
- branch: fix/org-policy
- lands-on: main
- quote-usd: 12.00
- cap-usd: 24.00
- wall: 150m
- task: an organisation policy file sets the floor for every repository's proven.yaml, and a known risk is accepted only by an exception with an owner and an expiry

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
2026-10-23: features closer to enterprise, built on what Proven already has. The decision is ADR-0036
(`docs/adr/ADR-0036-an-organisation-policy-and-exceptions-that-expire.md`). Its four points, verbatim:
1. An organisation policy file, named by a setting (a path or a repository), is read before each repository's
   `proven.yaml`. A repository may make the policy stricter, never weaker; a weaker local rule is refused by name.
2. An exception names the finding it covers (rule and location), a reason, an owner and an expiry date. Until it
   expires the finding is reported as *accepted* with those four fields, never hidden; after it expires the finding
   counts again, and the review says the exception expired.
3. An exception without an owner or an expiry is refused.
4. The proof is a test for each: a weaker local rule refused, an accepted finding shown with its owner, an expired
   exception no longer accepted.

### THE MASTER'S READING (2026-10-23, at 9303b2d — CLAIMS to re-measure, law #44)
- `proven.yaml` is read in `internal/review/config.go`: `loadConfigFile` (~140), `loadSettingsAtBaseline` (~173, the
  baseline commit's copy), `loadSettingsInTarget` (~186), parsed by `parseConfig` (~212) and `applyKey` (~300). Its keys
  are listed in `docs/CONFIGURATION.md` § `proven.yaml`.
- What makes a review fail in CI (severity, `--error`) is read in `internal/cli/config.go` and `internal/cli/run.go`.
- **Step 1, before any change:** write in your status a table of every setting a policy can set, and for each what
  "stricter" means (for example: `doubled_classes` — a superset is stricter; `path_rules` — a path moved to a higher
  class is stricter; a failing severity — a lower threshold is stricter). A setting with no clear order is not
  policy-settable in this wagon; name it. If the reading above is wrong, write `- master-was-wrong:`.

### WHAT TO BUILD
1. A setting naming the organisation policy file (e.g. `PROVEN_POLICY`): a local path in this wagon. A path inside the
   reviewed repository is refused (the change cannot carry its own policy). Missing or unreadable: refused at load, named.
2. The policy file uses the same flat format as `proven.yaml`, restricted to the policy-settable keys of step 1, plus
   exceptions. It is read before `proven.yaml`; each `proven.yaml` value is compared with the policy: stricter or equal
   is applied, weaker is refused at load, naming the key, both values and the policy file.
3. Exceptions live in the policy file only (never in the reviewed repository): each names a rule id, a location
   (path, or path and line), a reason, an owner and an expiry date (YYYY-MM-DD). Missing owner or expiry: refused at
   load. Format and keys: the leg chooses, consistent with the flat format, and documents it.
4. In a review: a finding matched by an unexpired exception is reported as `accepted` with its reason, owner and expiry,
   in every output format (text, JSON, SARIF as a suppression with justification, PR comment); it does not count toward
   failure. An expired exception: the finding counts again, and the review prints that the exception expired on <date>.
   The date compared is the review's own clock, passed in so tests can fix it.
5. `docs/POLICY.md`: the setting, the format, the stricter-than table, exceptions, and an example. New rows in
   `docs/CONFIGURATION.md` via `cmd/settings-doc`.
6. Code comments in the style of the surrounding code (ADR-0027).

### ACCEPTANCE
1. `go test ./...` green; `go vet ./...` clean.
2. Tests, one each: a weaker local rule refused by name; an accepted finding shown with its owner (text and JSON); an
   expired exception no longer accepted and the expiry printed; an exception with no owner refused; one with no expiry
   refused; a policy path inside the target refused.
3. Negative control, named in your status: remove the weaker-than check and its test turns red by name; remove the
   expiry comparison and the expired-exception test turns red by name. Restore both.
4. No policy setting: every existing test passes unchanged.

### WHAT IS NOT IN SCOPE
- A policy read from a repository (a path only in this wagon; ADR-0036 allows a repository later).
- The air-gapped mode (W1) and the evidence pack (W2).
- Licences.

### WHEN THIS ORDER CANNOT PROCEED
Write `SHOUT_<topic>.md` in the worktree, set `- state: blocked`, and stop.

### ARTIFACT CONFINEMENT
Every artifact this leg makes — captures, logs, evidence, scratch — lives under `████████████████████████` in the worktree.
Nothing outside the worktree; nothing committed unless this order names it.

### RETURN
`_EXEC_STATUS.md`: `- state: stop-2-ready`, `- last commit:` <sha>, an `- abstract:` of at most 5 lines, the step-1 table,
the commands run and their counts, the negative controls, and `- master-was-wrong:` if any. Check it with
`saphan doc check --family exec-status _EXEC_STATUS.md` before you end: its exit code is the verdict.
Read every `saphan` verb by its `RESULT:` line, its `--json` or its exit code, never by its sentences (`craft/COMMON.md`).

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| how an order is bound to the laws | 1 |
| private workspace path | 4 |
