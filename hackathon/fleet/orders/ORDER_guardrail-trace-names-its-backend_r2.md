# ORDER — guardrail-trace-names-its-backend · round 2

- change: guardrail-trace-names-its-backend
- round: 2
- base: main @ db9f7c2
- executor: claude-code
- written: 2026-10-25T00:18:25Z
- task: each guardrail check on the trace names the backend that answered, the model, and a structured down reason; a short instrument counts checks per UTC day, side, verdict and backend, with the median latency

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — guardrail-trace-names-its-backend
track: proven-meterless-engines
- train: proven-meterless-engines
- wagon: W54
- review: master-verify-only
- repo: proven
- base: main @ db9f7c2
- branch: fix/guardrail-trace-names-its-backend
- lands-on: main
- backend: claude-code
- model: claude-opus-5-5
- quote-usd: 10.00
- cap-usd: 20.00
- wall: 120m
- task: each guardrail check on the trace names the backend that answered, the model, and a structured down reason; a short instrument counts checks per UTC day, side, verdict and backend, with the median latency

█████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████
ROLE-LOCK: YOU = EXECUTOR. One-shot headless: foreground commands only; your first act writes `_EXEC_STATUS.md`
(`- state: in-progress`). A permission denial is a fact: try once, then `- state: blocked`.
COMMIT EVERY PART BEFORE `stop-2-ready` · ZERO push/merge/PR · author = fleet identity · commit subject at most 72
characters, matching `^\[[a-z0-9-]+\] [a-z0-9/._-]+: .+$`, e.g. `[guardrail-trace-names-its-backend] trace: backend on check`.
The area is ONE lowercase token. No AI attribution anywhere: no `Co-authored-by:`, no "Generated with", no tool or model
name. After the last commit paste `git log --format=%B db9f7c2..HEAD | grep -ci co-authored` — it must print `0`.
After `git commit`, `packed-refs.lock: Permission denied` means the commit LANDED — check `git log -1`.

### SKILLS
- ██████████████████████████████████
- ██████████████████████████████████████

Craft: `██████████████████████████` + `████████████████████████████` — read first.

### WHY
The owner wants to count guardrail checks per backend: how many ran on the self-hosted server and how many on NVIDIA's
hosted model, how many were blocked, how many could not be made, and how long they took. Today the trace cannot answer
that: it does not say which backend or model answered, and a check that could not be made is free text only.

### CLAIMS FROM THE MASTER — TEST THEM (law #44)
- MC-4 (master core, from the code at 5faabc9; re-read by this master on main d9083c0 on 2026-10-25): every guardrail check
  (pass, block, down) is one `kind="guardrail"` event, written by `guardEvent` in `internal/pipeline/pipeline.go`
  (about line 1437). `trace.Check` (`internal/trace/trace.go`, about line 65) has side, verdict, categories and latency, and
  no backend and no model. A down check sets only `Text = "guardrail: <reason>"`: no `Guardrail` block, so not even its
  side is structured.
- The master claims the down reasons form a closed set, all made by `guardrail.Refuse` in `internal/guardrail/{live,sample,
  absent}.go`: `not configured`, `off`, `bad verdict`, `timeout`, `unauthorized`, `empty answer`, `unreachable`,
  `no recorded verdict`, and the HTTP reasons of `httpReason`. The one open door is `asDown` in `wrap.go`, which passes
  any other error's text through. Measure the set; if it is not closed, say which reason escapes.
- The master claims `guardrail.Observe` does not carry the backend or the model, so the wrapper cannot report them
  without a change to its signature or to the `Adapter`. Choose the smallest change and say why.

### WHAT TO BUILD
0. **Step 0 — the test must be able to fail.** Write the trace tests of point 3 first and run them on base: paste the red.
1. `trace.Check` gains `backend` (the adapter that answered: `server`, `nvidia`, or the sample adapter's name, as the
   code calls it), `model` (the configured model id; empty only where no model is configured, and the test says where),
   and, for a check that could not be made, `down` (one value of the closed set above). A down check also carries its side.
   The existing `text` stays as it is, so nothing that reads it today breaks.
2. `asDown` maps an error outside the closed set to one named value (for example `other`), with the original text kept in
   `text` only. Never a free-text value in `down`.
3. Tests, one per property: a `server` check names `server` and its model; an `nvidia` check names `nvidia` and its model;
   a sample-mode check names the sample backend; each down reason of the set lands as its structured value with its side;
   an unknown error lands as the named fallback value.
4. The instrument: a small Go command under `cmd/` (name it; justify it in one line) that reads one or more `trace.jsonl`
   files and prints, per UTC day: side, verdict (`safe`, `unsafe`, `down:<reason>`), backend, count, median latency. Plain
   text table, stable column order. A test runs it on a committed sample trace and compares the output to a golden file.
5. A short section in `docs/` (beside the guardrail settings in `docs/CONFIGURATION.md`, or where you show it belongs):
   the new fields, the instrument's command line, and its output on the sample trace.

### ACCEPTANCE
1. `go build ./... && go vet ./... && go test -count=1 ./...`: green. Paste the summary lines.
2. Step 0's red run, pasted.
3. **Negative control (law #27):** leave `backend` empty in `guardEvent` → the backend tests turn red; paste it; restore.
4. **Unplug (law #41):** the guardrail switched off → the trace's one "no guardrail" line is unchanged, and the instrument
   prints no check rows for that run, with no error.
5. No secret in the new fields: the key never appears on the trace (a test greps the written trace for the configured key).

### COMMENTS FOR THE JUDGES (ADR-0027, a standing rule of this train)
(1) every package you touch keeps its package comment true; (2) every exported name has a doc comment beginning with the
name; (3) every step a reader would stop at (a guard, an order that matters, a value from a setting, a refusal) has a
comment saying WHY, never one restating the code; (4) code that carries out a decision names it (e.g. `ADR-0005: …`);
(5) every test says which property it proves; (6) English, plain; no AI attribution, no host names, internal paths or
credential names, no commented-out code.

### WHAT IS NOT IN SCOPE
- The guardrail's checks themselves, its settings, and its live client.
- The generator interface (`internal/generator`): unchanged. If you need a change there, stop and say so.
- Markovian (W50) and Scout Intent (W51), parallel legs in other worktrees.

### WHEN THIS ORDER CANNOT PROCEED
Write `SHOUT_<topic>.md` in the worktree, set `- state: blocked`, and stop.

### ARTIFACT CONFINEMENT
Every artifact this leg makes lives under `███████████████████████████████████████████████` in the worktree, except the
files this order names (the trace fields, the guardrail change, the command, its sample trace and golden file, the tests,
the docs section).

### RETURN
`_EXEC_STATUS.md`: `- state: stop-2-ready`, `- last commit:` <sha>, an `- abstract:` of at most 5 lines, `- master-was-wrong:`
(each claim above that a measurement contradicted, with the measurement, or `none`), the commands run and their exit codes.
Check it with `saphan doc check --family exec-status _EXEC_STATUS.md` before you end: its exit code is the verdict.

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| how an order is bound to the laws | 1 |
| private workspace path | 4 |
