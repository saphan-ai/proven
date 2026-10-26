# ORDER — the-generator-calls-nvidia-hosted · round 1

- change: the-generator-calls-nvidia-hosted
- round: 1
- base: main @ 0b954a9
- executor: claude-code on macm5
- written: 2026-10-23T20:34:05Z
- task: the generator gains NVIDIA's hosted API as a second provider, PROVEN_GENERATOR_MODE=nightshift|nvidia (ADR-0031)

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — the-generator-calls-nvidia-hosted
track: proven-reviewer-core
- train: proven-reviewer-core
- wagon: W35
- review: named
- repo: proven
- base: main @ 0b954a9
- branch: fix/the-generator-calls-nvidia-hosted
- lands-on: main
- backend: claude-code
- model: claude-sonnet-5-5
- quote-usd: 4.00
- cap-usd: 8.00
- wall: 150m
- task: the generator gains NVIDIA's hosted API as a second provider, PROVEN_GENERATOR_MODE=nightshift|nvidia (ADR-0031)

█████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████
ROLE-LOCK: YOU = EXECUTOR. One-shot headless: foreground commands only; your first act writes `_EXEC_STATUS.md`
(`- state: in-progress`). A permission denial is a fact: try once, then `- state: blocked`.
COMMIT EVERY PART BEFORE `stop-2-ready` · ZERO push/merge/PR · author = fleet identity · commit subject at most 72
characters, `[the-generator-calls-nvidia-hosted] <area>: <what>` included — the slug is long, so keep `<what>` short.
No AI attribution anywhere. The tool re-adds a `Co-authored-by:` line on amend; if it is there at the end, say so (the
master removes it).
After `git commit`, `packed-refs.lock: Permission denied` means the commit LANDED — check `git log -1`.

### SKILLS
- ██████████████████████████████████
- ██████████████████████████████████████

Craft: `██████████████████████████` + `████████████████████████████` — read first.

### THE OWNER'S WORDS
2026-10-21 (translated from Polish): *"Working first, pretty later."* 2026-10-23: the judges read this code (ADR-0027).

### THE OWNER'S DECISION (ADR-0031, docs/adr/ADR-0031-anyone-with-an-nvidia-key-can-run-a-live-review.md), pasted
1. The generator has a second provider, NVIDIA's hosted API (https://integrate.api.nvidia.com/v1, OpenAI-compatible chat
   completions), chosen by `PROVEN_GENERATOR_MODE=nightshift|nvidia`. 2. Nightshift stays the default (ADR-0019).
3. In `nvidia` mode the key is `████████████████████`, the address and models are settings with defaults, the key is
   never printed. Every agent answered by an `nvidia/*` model where `PROVEN_GENERATOR_REQUIRE_NVIDIA` asks for it, the
   Skeptic from another family, every call checked by the guardrail. Where the hosted API gives no signed receipt the
   trace says *no receipt* instead of inventing one. 6. The docs say per mode which keys are needed and where to get
   them (docs/CONFIGURATION.md, docs/LAB.md).

### THE MASTER'S RULING ON A CONFLICT (2026-10-24)
On main, `internal/generator/live.go` `finish` refuses an answer with no receipt when NVIDIA is required (ADR-0019
point 4, Nightshift's receipt). That rule stays for `nightshift` mode. In `nvidia` mode Proven calls NVIDIA directly:
the answering model is the `model` field of the hosted API's response, recorded with "no receipt", and the
`nvidia/*` requirement is checked against that field. Say in a comment why the two modes differ.

### WHAT TO BUILD
1. The mode setting and the hosted provider (an adapter beside the Nightshift one, the same `Generator` interface),
   streaming or not as the hosted API serves; empty key = the service down and named (ADR-0013 point 4a), a wrong
   value refused; the key in W30's redactor.
2. livecheck checks the mode in use with one cheap call and names a missing or invalid key.
3. docs/CONFIGURATION.md (regenerate), docs/LAB.md: per mode, the keys and where to get them (build.nvidia.com).
4. Defaults for the hosted model ids: pick from NVIDIA's published list and say so; a live call with a real key is
   UNRUN unless one is provided to you (never search for one).

### ACCEPTANCE
1. `go build ./... && go vet ./... && go test -count=1 ./...`: green. Paste the summary lines.
2. Recorded-response tests (an httptest server): an nvidia/* answer kept, a non-nvidia answer refused where required,
   no receipt recorded as "no receipt", the key never in any output, an empty key down and named.
3. **Negative control (law #27):** let a non-nvidia model through in nvidia mode where required → its test red; restore.

### COMMENTS FOR THE JUDGES (ADR-0027, a standing rule of this train)
(1) every package you touch keeps its package comment true; (2) every exported name has a doc comment beginning with the
name; (3) every step a reader would stop at (a guard, an order that matters, a value from a setting, a refusal) has a
comment saying WHY, never one restating the code; (4) code that carries out a decision names it (e.g. `ADR-0025: …`);
(5) every test says which property it proves; (6) English, plain; no AI attribution, no host names, internal paths or
credential names, no commented-out code.

### WHAT IS NOT IN SCOPE
The second reviewer (ADR-0037, W36); the CLI flags beyond reading the setting; any real key.

### WHEN THIS ORDER CANNOT PROCEED
Write `SHOUT_<topic>.md` in the worktree, set `- state: blocked`, and stop.

### ARTIFACT CONFINEMENT
Every artifact this leg makes lives under `███████████████████████████████████████████████` in the worktree, except the
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
| private credential or variable name | 1 |
| private workspace path | 4 |
