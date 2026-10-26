# ORDER — formatter-docs-and-example · round 1

- change: formatter-docs-and-example
- round: 1
- base: main @ 9b11366
- executor: cursor on runner-cursor
- written: 2026-10-24T13:11:21Z
- task: Proven's documentation explains how to add an output format, with a small working example under a new examples/ directory that the test suite builds and runs, so an example that stops matching the code turns the build red; nothing else

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — formatter-docs-and-example
track: proven-cli-and-action
- train: proven-cli-and-action
- wagon: W11
- review: master-verify-only
- repo: proven
- base: main @ 9b11366
- branch: fix/formatter-docs-and-example
- lands-on: main
- backend: cursor
- model: grok-4.7-high
- quote-usd: 2.00
- cap-usd: 4.00
- wall: 120m
- task: Proven's documentation explains how to add an output format, with a small working example under a new examples/ directory that the test suite builds and runs, so an example that stops matching the code turns the build red; nothing else

█████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████
ROLE-LOCK: YOU = EXECUTOR. One-shot headless: foreground commands only; your first act writes `_EXEC_STATUS.md`
(`- state: in-progress`). A permission denial is a fact: try once, then `- state: blocked`.
COMMIT EVERY PART BEFORE `stop-2-ready` · ZERO push/merge/PR · author = fleet identity.
Commit subjects: `[formatter-docs-and-example] <area>: <what>`, the area ONE token, at most 72 characters, English
(the merge refuses others). No AI attribution anywhere: no `Co-authored-by:` line, no tool or model name.

### SKILLS
- ██████████████████████████████████
- ██████████████████████████████████████

Craft: `██████████████████████████` + `████████████████████████████` — read first.

### WHERE YOU START (master, 2026-10-24; claims to test, law #44)
- The owner wants Proven's extension points documented for people, each with a runnable example in a new top-level
  `examples/` directory. This wagon owns ONE extension point: the output format. Other trains write the others (generator,
  guardrail, reviewer, test runner, tool, evidence query); they add their own rows to `examples/README.md` later.
- `internal/cli/format.go:18` (main 9b11366): `type Formatter interface { Write(w io.Writer, result review.Result) error }`;
  `FormatterFor(format string)` is a closed `switch` over `text`, `json`, `sarif` — there is no registry, so a new format also
  needs a case there and a value the `--format` flag accepts. `render` writes to memory first so a failure is known before any
  output is written. The JSON and SARIF contracts are in `docs/OUTPUT_FORMATS.md`.
- `docs/ADAPTERS.md` today has "Adding a language" and "Adding a tool"; keep its tone (plain words, one worked example).
- Another wagon (W8) is rewriting the text writer in `███████████████████████████`; do not touch `internal/cli` at all.

### WHAT TO DO
1. **Reproduce first:** confirm the interface and `FormatterFor` as stated above at your base, and paste `go doc ./internal/cli Formatter`.
   If they differ, write the difference in `- master-was-wrong:` and document what the code really does.
2. `docs/ADAPTERS.md`: a section "Adding an output format": what a formatter is for; the interface and its one method, every
   argument and the error contract; where the format is chosen (`--format`, `-o`, the `--*-output` flags) and the two places a
   new format must be wired (`FormatterFor` and the flag); that sample mode produces the same `review.Result`, so a formatter
   behaves identically there; the rule that every string from the change must be escaped before it reaches a terminal
   (point at `visible`), and that JSON/SARIF writers must not assume the encoder escapes DEL and C1. One worked example.
3. `examples/formatter/`: a small, complete formatter (e.g. a one-line-per-finding CSV or "path:line: title" writer) as its own
   Go package inside this module, with a `_test.go` that renders a fixed `review.Result` and compares it to a golden file. It is
   built and run by `go test ./...` — no separate CI wiring needed. Escape what it prints from the change.
4. `examples/README.md`: the index of examples — one row per example (name, extension point, the docs section that explains it,
   how to run it). Start it with the formatter row and a sentence that each extension point adds its own row.
5. Negative control: change one byte of the golden and paste the red; restore it.
6. Nothing else: `internal/`, `cmd/`, `AGENTS.md` unchanged.

### ACCEPTANCE
1. `git log --oneline 9b11366..HEAD` shows only this wagon's commits; paste it.
2. `go vet ./...` clean; `go test -count=1 ./...` green (including `./examples/...`), counts per package; `gofmt -l` empty.
3. The negative control's red, pasted.
4. A `- master-was-wrong:` line: `—`, or what you measured that contradicts this order.

### CODE IS WRITTEN TO BE READ (ADR-0027)
The example is read by people learning the interface: doc comments on every exported name, a WHY comment where a reader would
stop, and its test says which property it proves. English, plain.

### HEADLESS TURNS
A headless leg never ends its turn with prose: a message without a tool call ends the session. Write `_EXEC_STATUS.md`
before your last message; if you cannot finish, write it with `- state: blocked`. Quote `git rev-parse` output verbatim.

### WHEN THIS ORDER CANNOT PROCEED
Write `SHOUT_<topic>.md` in the worktree, set `- state: blocked`, and stop.

### ARTIFACT CONFINEMENT
Every artifact this leg makes lives under `████████████████████████████████████████` in the worktree, except the files this order
names. Never run `proven scan` inside the checkout.

### RETURN
`_EXEC_STATUS.md`: `- state: stop-2-ready`, `- last commit:` <sha>, an `- abstract:` of at most 5 lines, the commands run and
their exit codes, the red. Check it with `saphan doc check --family exec-status _EXEC_STATUS.md`: its exit code is the verdict.

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| file and line in Saphan Studio's private code | 1 |
| how an order is bound to the laws | 1 |
| private workspace path | 4 |
