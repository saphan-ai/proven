# ORDER — the-generator-through-nightshift · round 1

- change: the-generator-through-nightshift
- round: 1
- base: main @ 6affa00
- executor: cursor on runner-cursor
- written: 2026-10-22T06:02:37Z
- task: the live generator calls Nightshift (any OpenAI-compatible endpoint by setting), streams, reads the routed model from the response, refuses a non-NVIDIA answer, and degrades by name

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — the-generator-through-nightshift
track: proven-reviewer-core
- train: proven-reviewer-core
- wagon: W2
- review: master-verify-only
- repo: proven
- base: main @ 6affa00
- branch: fix/the-generator-through-nightshift
- lands-on: main
- worktree: ████████████████████████████████████████████████████████████████████
- backend: cursor
- model: grok-4.7-high
- quote-usd: 3.00
- cap-usd: 6.00
- wall: 120m
- task: the live generator calls Nightshift (any OpenAI-compatible endpoint by setting), streams, reads the routed model from the response, refuses a non-NVIDIA answer, and degrades by name

█████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████
ROLE-LOCK: YOU = EXECUTOR. One-shot headless: foreground commands only; your first act writes `_EXEC_STATUS.md`
(`- state: in-progress`). A permission denial is a fact: try once, then `- state: blocked`.
COMMIT EVERY PART BEFORE `stop-2-ready` · ZERO push/merge/PR · author = fleet identity · commit message
`[the-generator-through-nightshift] <area>: <what>`.
No AI attribution anywhere: no `Co-Authored-By`, no "Generated with", no tool or model name, no session link. After EVERY
commit, run `git log -1 --format=%B` and amend away any `Co-authored-by:` line the tool added; in your status paste
`git log --format=%B 6affa00..HEAD | grep -ci co-authored` — it must print `0`.
After `git commit`, `packed-refs.lock: Permission denied` means the commit LANDED — check `git log -1`.

### SKILLS
- ██████████████████████████████████
- ██████████████████████████████████████

Craft: `██████████████████████████` + `████████████████████████████` — read first.

### THE OWNER'S WORDS
2026-10-21 (translated from Polish): *"Working first, pretty later."*

### THE DECISION YOU BUILD — ADR-0019 (pasted, accepted by the owner)
1. Nightshift is the default route for every model call proven makes, through the generator of ADR-0018.
3. proven implements the Nightshift runtime contract: `Authorization: Bearer <token>`, the `meterless-app` header,
   `meterless-scope: user`, model discovery with `GET /v1/models`, and `POST /v1/chat/completions`.
4. Each review names the route and the model that answered, read from Nightshift's response, not assumed from the
   configuration. A review whose answering model is not an NVIDIA model (`nvidia/*`) is refused: proven does not show it
   as a verdict, it says which model answered and why the answer was refused. (Amendment 1: naming a model in the request
   does not pin it; the route is CHECKED, never trusted.)
5. Other routes stay settings: a team may point the generator at another OpenAI-compatible endpoint.
6. The settings are the base URL, the runtime token and the routing mode or model, from the environment, validated when
   they load (ADR-0013). The token is a secret and lives in the environment only. The runtime token's variable is
   `███████████████████████` (ADR-0020 §1).
7. Sample mode needs no Nightshift.

### WHAT THE MASTER MEASURED (2026-10-21/22; claims for you to test, law #44)
- A real streamed answer of Nightshift is provisioned at `█████████████████████████████████████████████████████████████` (the raw capture inside a fenced text block)
  (headers + SSE). The routed model and the receipt id ride in a chunk's top-level `meterless` object
  (`meterless.routed_model`, `meterless.receipt_id`); the chunk's `model` field changes from `meterless/auto` to the
  routed model. The provider (`winning_route.provider`) is NOT in the stream; it is only in Nightshift's stored receipt.
- Nightshift's hosted NVIDIA model returned EMPTY responses on ~25–30% of calls on 2026-10-22 04:40–05:01Z (receipts
  `…_empty_response`), more when two reviews ran at once; other failures read `ttfb_timeout`.

### WHAT TO BUILD
1. The live generator (the interface W1 built), OpenAI-compatible: `POST {base}/v1/chat/completions` with `stream: true`,
   the Bearer token, `meterless-app` and `meterless-scope: user`; it reads the SSE stream to the end and returns the text
   and the routed model and receipt id when present. Settings from the environment, a closed set, validated on load
   (`*ConfigError` naming the variable and the allowed values): base URL, app name, routing model, timeout.
2. The route check: an answer whose routed model is not `nvidia/*` is refused for every agent whose configuration
   requires NVIDIA (default: all); the review names the model that answered. An answer with NO routed model (another
   OpenAI-compatible endpoint, or no `meterless` block) is reported as "model unknown" — never guessed.
3. Degradation (ADR-0001): a wrong token (401), a timeout, an unreachable endpoint, or an empty answer after ONE retry
   → the generator returns a named service failure (`nightshift: <reason>`), the review completes degraded, never hangs.
4. `review.Open` with `Sample` false uses the live generator; `Result.Models` carries each agent's routed model and
   receipt id (`ModelUse.FromReceipt` true only when the stream named it).
5. The settings in one short section of `AGENTS.md`.

### ACCEPTANCE
1. `go build ./... && go vet ./... && go test -count=1 ./...`: green. Paste the summary lines.
2. Tests against a local fake endpoint (`httptest`) that replays the provisioned stream: the routed model and receipt id
   are read. **Negative controls (law #27):** the same stream with `routed_model` set to a non-NVIDIA model → refused by
   name; with the `meterless` object removed → "model unknown"; a 401 → named failure, review degraded; an empty stream →
   one retry, then named failure. Name the mutated field in each test.
3. **Unplug (law #41):** a test with the endpoint closed → the review completes degraded and names `nightshift`.

### THIS ORDER CAN BE WRONG (law #44)
The master's claims above are from one recorded answer. If the provisioned file does not parse as described, say so in
`- master-was-wrong:` with what it contains, and build to what it contains. If your environment refuses a local
listener for `httptest`, say so (`- state: blocked`) — do not test against a real network.

### WHAT IS NOT IN SCOPE
- Any call to the real Nightshift: the master runs the live review after the round.
- The guardrail (W3), H-MEM (W4), Zetaris (W5), the Skeptic's model family (W6).
- The seam types in `internal/review/review.go`.

### WHEN THIS ORDER CANNOT PROCEED
Write `SHOUT_<topic>.md` in the worktree, set `- state: blocked`, and stop.

### ARTIFACT CONFINEMENT
Every artifact this leg makes — captures, logs, evidence, scratch — lives under `██████████████████████████████████████████████`
in the worktree. Nothing outside the worktree; nothing committed unless this order names it (a test fixture copied from the
provisioned file into `testdata/` is named here and allowed).

### RETURN
`_EXEC_STATUS.md`: `- state: stop-2-ready`, `- last commit:` <sha>, an `- abstract:` of at most 5 lines, `- master-was-wrong:`,
the commands run and their exit codes, and the session recording or the reason there is none (ADR-0024). Check it with
`saphan doc check --family exec-status _EXEC_STATUS.md` before you end: its exit code is the verdict. Read every `saphan`
verb by its `RESULT:` line, its `--json` or its exit code, never by its sentences.

## REDACTED

| what | times |
|---|---|
| absolute or private path | 2 |
| how an order is bound to the laws | 1 |
| private credential or variable name | 1 |
| private workspace path | 5 |
