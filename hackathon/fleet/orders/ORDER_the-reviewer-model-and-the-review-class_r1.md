# ORDER — the-reviewer-model-and-the-review-class · round 1

- change: the-reviewer-model-and-the-review-class
- round: 1
- base: main @ 7ad0803
- executor: cursor on runner-cursor
- written: 2026-10-22T12:02:45Z
- task: the author's model is read from the change, the reviewer is chosen against its family, the review class comes from configured path rules, and every list is closed and validated on load

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — the-reviewer-model-and-the-review-class
track: proven-reviewer-core
- train: proven-reviewer-core
- wagon: W7
- review: named
- repo: proven
- base: main @ 7ad0803
- branch: fix/the-reviewer-model-and-the-review-class
- lands-on: main
- worktree: ███████████████████████████████████████████████████████████████████████████
- backend: cursor
- model: grok-4.7-high
- quote-usd: 3.00
- cap-usd: 6.00
- wall: 120m
- task: the author's model is read from the change, the reviewer is chosen against its family, the review class comes from configured path rules, and every list is closed and validated on load

█████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████
ROLE-LOCK: YOU = EXECUTOR. One-shot headless: foreground commands only; your first act writes `_EXEC_STATUS.md`
(`- state: in-progress`). A permission denial is a fact: try once, then `- state: blocked`.
COMMIT EVERY PART BEFORE `stop-2-ready` · ZERO push/merge/PR · author = fleet identity · commit message
`[the-reviewer-model-and-the-review-class] <area>: <what>`.
No AI attribution anywhere: no `Co-Authored-By`, no "Generated with", no tool or model name, no session link. After EVERY
commit, run `git log -1 --format=%B` and amend away any `Co-authored-by:` line the tool added; in your status paste
`git log --format=%B 7ad0803..HEAD | grep -ci co-authored` — it must print `0`.
After `git commit`, `packed-refs.lock: Permission denied` means the commit LANDED — check `git log -1`.

### SKILLS
- ██████████████████████████████████
- ██████████████████████████████████████

Craft: `██████████████████████████` + `████████████████████████████` — read first.

### THE OWNER'S WORDS
2026-10-21 (translated from Polish): *"Working first, pretty later."*

### THE DECISIONS (in your worktree; read them first)
`fleet/decisions/` ADR-0008, ADR-0009, ADR-0010, ADR-0011, ADR-0012, ADR-0013 (point 4a: validated on load, the refusal
says what was wrong, what is allowed, how to fix it), ADR-0014 (all nine points), and ADR-0020's exit-code table (exit 7:
the configuration was refused when it loaded). They are the specification; this order only cuts it.

### WHAT TO BUILD
1. **The author's model** (ADR-0011): read `Assisted-by: <agent>:<model>` trailers from every commit between the request's
   baseline and head (W1c's `internal/change/between.go` already walks that range with git), and the same line in a
   pull-request description when the request carries one. Sources and trailer keys are settings (defaults: commit
   trailers on, description on, record file off; key `Assisted-by:`). Proven never writes a trailer.
2. **The family table** (ADR-0011 point 4): a configurable map of model-name patterns to families, with the defaults the ADR
   names. A model the table does not know is "family unknown".
3. **The reviewer models** (ADR-0008/0009/0010): a default reviewer (NVIDIA by default) and a special-review model; every
   review uses the default; doubled classes add the special-review model. The review states, per reviewer, its model and
   whether it is the author's family (or "author model not stated" and which default it used).
4. **The review class** (ADR-0012/0014): path rules → class; the highest-risk class in a configured order wins; a fallback
   class (backend); the classes are a closed list (security, backend, user interface, mobile, docs); security is doubled
   by default; the review states its class and the rule that chose it.
5. **Validation on load** (ADR-0013 4a): an unknown family in the table, a rule naming an unknown class, a priority order
   that leaves a class out, an unknown key → refused before any work, exit 7, the message naming the field, the value and
   the allowed values. Extend the existing loader in `internal/review/config.go`; do not add a second one.
6. Where the seam (`internal/review/review.go`) has no field to carry the class, the reviewer models and the
   author-model statement, add them ADDITIVELY (new fields only; nothing renamed or removed) and say so in your abstract —
   the shell's track reads this seam.

### ACCEPTANCE
1. `go build ./... && go vet ./... && go test -count=1 ./...`: green. Paste the summary lines.
2. A table test per rule of items 1–5.
3. **Negative controls (law #27), each pasted:** an unknown family in the table → refused on load, exit 7; a path rule
   naming an unknown class → refused on load; a change with no trailer → the review says "author model not stated" and
   which default it used. And one unplug: remove the highest-risk-wins step → its table test turns red; restore.
4. `go run ./cmd/proven scan --sample testdata/fixture/discount` still prints F1 and F2 and drops F3, and now prints the
   class and the reviewer statement. Paste it.

### THIS ORDER CAN BE WRONG (law #44)
The master's claims, 2026-10-22, read off `main` at 7ad0803 and not built: (M1) `internal/review/config.go` reads only
`timeout` today; (M2) the seam has no field for the class or the reviewer models (`git grep -n Class internal/review`).
Check both BEFORE you build; each one that is wrong goes into `- master-was-wrong:` with the measurement.

### WHAT IS NOT IN SCOPE
- The Skeptic and the Judge (W6); the shell's flags (the shell's track); calling the second reviewer model live (this
  wagon decides and states; the generator call is the existing one).

### WHEN THIS ORDER CANNOT PROCEED
Write `SHOUT_<topic>.md` in the worktree, set `- state: blocked`, and stop.

### ARTIFACT CONFINEMENT
Every artifact this leg makes lives under `█████████████████████████████████████████████████████` in the worktree.

### RETURN
`_EXEC_STATUS.md`: `- state: stop-2-ready`, `- last commit:` <sha>, an `- abstract:` of at most 5 lines, `- master-was-wrong:`,
the commands run and their exit codes, and the session recording or the reason there is none (ADR-0024). Check it with
`saphan doc check --family exec-status _EXEC_STATUS.md` before you end: its exit code is the verdict.

## REDACTED

| what | times |
|---|---|
| absolute or private path | 2 |
| how an order is bound to the laws | 1 |
| private workspace path | 4 |
