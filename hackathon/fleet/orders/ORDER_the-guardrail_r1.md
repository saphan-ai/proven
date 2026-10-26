# ORDER — the-guardrail · round 1

- change: the-guardrail
- round: 1
- base: main @ 6db4272
- executor: cursor on runner-cursor
- written: 2026-10-22T10:44:27Z
- task: every prompt and every answer passes the guardrail adapter, on by default, with two backends (our model server, NVIDIA's hosted API); unsafe stops the step and is written with its categories

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — the-guardrail
track: proven-reviewer-core
- train: proven-reviewer-core
- wagon: W3
- review: master-verify-only
- repo: proven
- base: main @ 6db4272
- branch: fix/the-guardrail
- lands-on: main
- worktree: █████████████████████████████████████████████████
- backend: cursor
- model: grok-4.7-high
- quote-usd: 3.00
- cap-usd: 6.00
- wall: 120m
- task: every prompt and every answer passes the guardrail adapter, on by default, with two backends (our model server, NVIDIA's hosted API); unsafe stops the step and is written with its categories

█████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████
ROLE-LOCK: YOU = EXECUTOR. One-shot headless: foreground commands only; your first act writes `_EXEC_STATUS.md`
(`- state: in-progress`). A permission denial is a fact: try once, then `- state: blocked`.
COMMIT EVERY PART BEFORE `stop-2-ready` · ZERO push/merge/PR · author = fleet identity · commit message
`[the-guardrail] <area>: <what>`.
No AI attribution anywhere: no `Co-Authored-By`, no "Generated with", no tool or model name, no session link. After EVERY
commit, run `git log -1 --format=%B` and amend away any `Co-authored-by:` line the tool added; in your status paste the count
of `co-authored` lines in your commits (it must be `0`).
After `git commit`, `packed-refs.lock: Permission denied` means the commit LANDED — check `git log -1`.

### SKILLS
- ██████████████████████████████████
- ██████████████████████████████████████

Craft: `██████████████████████████` + `████████████████████████████` — read first.

### THE OWNER'S WORDS
2026-10-21 (translated from Polish): *"Working first, pretty later."*

### THE DECISION YOU BUILD — ADR-0005 (pasted, accepted by the owner)
- The guardrail is on by default. Every prompt an agent sends to a model, and every answer it gets back, passes the
  guardrail first.
- One guardrail adapter, two backends: our own model server, and NVIDIA's hosted API.
- A setting turns the guardrail off, for anyone with neither. The trace says that the run had no guardrail.
- An `unsafe` verdict stops that step and is written to the trace with its categories. It is never silently dropped.
- Every model call costs one guardrail call more. Its latency is measured and shown in the trace.

### WHAT THE GUARDRAIL TRACK MEASURED (2026-10-21; claims to test, law #44)
- Our server: OpenAI-compatible `POST {base}/v1/chat/completions`, model id `nemotron-3.5-content-safety`, `temperature: 0`.
  `messages`: one `user` turn = the prompt to judge; an optional `assistant` turn = the answer to judge.
  `chat_template_kwargs`: `{"enable_thinking": false, "request_categories": "/categories"}`.
- Its answer is TEXT in the completion, lines:
  `User Safety: safe|unsafe` · `Response Safety: safe|unsafe` (only with an assistant turn) ·
  `Safety Categories: <comma-separated names>` (only when unsafe). Measured examples:
  `User Safety: unsafe` + `Safety Categories: Criminal Planning/Confessions, Other`;
  `User Safety: safe` + `Response Safety: unsafe` + `Safety Categories: Malware, PII/Privacy` (706 ms).
  With thinking on, a `</think>` block can come before the verdict lines — skip it.
- NVIDIA's hosted API: the same OpenAI-style endpoint at `https://integrate.api.nvidia.com/v1`, same answer format; its
  model id is NOT confirmed (third parties say `nvidia/nemotron-3.5-content-safety`) — make it a setting.
- Warm latencies: 156–706 ms. No literal raw JSON pair was recorded; build the parser to the verdict lines above.

### WHAT TO BUILD
1. A guardrail adapter (interface + live implementation for each backend + a sample implementation answering from recorded
   verdicts): settings from the environment, closed set, validated on load: mode (`server` | `nvidia` | `off`, default
   `server`), base URL, model id, key variable (secret, never printed), timeout.
2. The generator (W2) passes every prompt and every answer through it. `unsafe` stops that agent's step; the trace records
   the verdict, the categories and the latency of every check; `off` writes "no guardrail" to the trace once per run.
3. The guardrail unreachable or timing out → the step is REFUSED by name (`guardrail: <reason>`), never passed through.
4. Sample mode: the sample guardrail answers safe for the fixture; one fixture prompt is recorded as unsafe for the tests.

### ACCEPTANCE
1. `go build ./... && go vet ./... && go test -count=1 ./...`: green. Paste the summary lines.
2. Tests against a fake server (httptest) replaying the verdict lines above: unsafe prompt → the step stops, trace names
   the categories; unsafe answer (`Response Safety: unsafe`) → the answer is not used. **Negative control (law #27):** the
   same run with mode `off` completes and the trace says "no guardrail".
3. **Unplug (law #41):** fake server closed → the step is refused by name, not passed through.
4. Sample scan unchanged (F1, F2; F3 absent). Paste it.

### THIS ORDER CAN BE WRONG (law #44)
If a real answer shape differs from the verdict lines above in a way your tests reveal, write it in `- master-was-wrong:`.

### WHAT IS NOT IN SCOPE
- Any call to the real guardrail server or NVIDIA (the master runs the live check). The seam types; the shell; Zetaris.

### WHEN THIS ORDER CANNOT PROCEED
Write `SHOUT_<topic>.md` in the worktree, set `- state: blocked`, and stop.

### ARTIFACT CONFINEMENT
Every artifact this leg makes — captures, logs, evidence, scratch — lives under `███████████████████████████` in the
worktree. Nothing outside the worktree; nothing committed unless this order names it.

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
| private workspace path | 4 |
