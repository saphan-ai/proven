# ORDER — generator-adapter-example · round 1

- change: generator-adapter-example
- round: 1
- base: main @ 931e2e8
- executor: agent-loop
- written: 2026-10-25T07:22:13Z
- task: a small, tested example of a model-provider adapter that implements Proven's Generator interface, with its section in docs/ADAPTERS.md, so a reader can copy it to plug in their own model

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — generator-adapter-example
track: proven-meterless-engines
- train: proven-meterless-engines
- wagon: W53
- review: master-verify-only
- repo: proven
- base: main @ 931e2e8
- branch: fix/generator-adapter-example
- lands-on: main
- backend: agent-loop
- model: nvidia/nemotron-3.5-super-vl-120b-a12b
- quote-usd: 4.00
- cap-usd: 8.00
- wall: 90m
- task: a small, tested example of a model-provider adapter that implements Proven's Generator interface, with its section in docs/ADAPTERS.md, so a reader can copy it to plug in their own model

█████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████
ROLE-LOCK: YOU = EXECUTOR. One-shot headless: foreground commands only; your first act writes `_EXEC_STATUS.md`
(`- state: in-progress`). A permission denial is a fact: try once, then `- state: blocked`.
FIRST `git rev-parse --short=7 HEAD` must print `931e2e8`; else `- state: blocked`, stop.
COMMIT EVERY PART BEFORE `stop-2-ready` · ZERO push/merge/PR · author = fleet identity · commit subject at most 72
characters, matching `^\[[a-z0-9-]+\] [a-z0-9/._-]+: .+$`, e.g. `[generator-adapter-example] examples: a generator adapter`.
The area is ONE lowercase token. No AI attribution anywhere: no `Co-authored-by:`, no "Generated with", no tool or model
name. After the last commit paste `git log --format=%B 931e2e8..HEAD | grep -ci co-authored` — it must print `0`.
Never `git add -A` or `git add .`: stage the files you name.

### SKILLS
- ██████████████████████████████████
- ██████████████████████████████████████

Craft: `██████████████████████████` + `████████████████████████████` — read first.

### WHY
The owner wants each of Proven's extension points to have a section in `docs/ADAPTERS.md` and one small example that
compiles and is tested by `go test ./...`. This order is the first of them: the generator, the one model interface every
model-calling agent goes through. It is also the first development run on an NVIDIA executor seat, so keep it small and
measurable.

### THE INTERFACE (main @ 931e2e8, `internal/generator/generator.go`, pasted)
```go
// Generator answers one prompt for one agent.
type Generator interface {
	Complete(ctx context.Context, agent, prompt string) (string, error)
}
```
The package comment there names the existing implementations: `Live` (Meterless Nightshift, or NVIDIA's hosted API in
nvidia mode), `Sample` (recorded responses, no network), `Absent` (the service is down, named), `Second` (any
OpenAI-compatible API, the second reviewer). `PROVEN_GENERATOR_MODE` chooses. Read `internal/generator/` before you write.

### CLAIMS FROM THE MASTER — TEST THEM (law #44)
- The master claims `examples/formatter/` (`lines.go`, `lines_test.go`, `testdata/lines.golden`) is the shape an example
  takes on main, and that `go test ./...` from the repository root already runs it. Measure both; follow what you find.
- The master claims `docs/ADAPTERS.md` has sections for languages, tools, sources, the evidence store and output formats,
  and none for the generator. Measure it.

### WHAT TO BUILD
1. `examples/generator/`: a package with one small type that implements `generator.Generator`. Make it something a reader
   would copy: an adapter that answers from a fixed table keyed by agent name, or one that wraps another Generator and
   counts calls. No network, no secret, no environment variable read. If importing `internal/generator` from `examples/`
   is refused by Go's internal-package rule, say so, and declare a local interface with the same method set, with a
   compile-time check in the test that `generator.Generator` and it agree.
2. A test that compiles the type against the interface (`var _ generator.Generator = …`) and checks one answer and one
   refusal (an unknown agent returns a named error).
3. A section `## Adding a model provider` in `docs/ADAPTERS.md`: what the interface is for, its one method and its
   arguments, the settings that select an implementation (`PROVEN_GENERATOR_MODE` and what it reads), how sample mode
   treats it, and a pointer to the example. The facts come from the code you read; name the file of each.

### ACCEPTANCE
1. `go build ./... && go vet ./... && go test -count=1 ./examples/...`: paste the summary lines.
2. `go test -count=1 ./...`: paste the summary; name any test that cannot bind a port on this seat.
3. **Negative control (law #27):** rename the method in the example → the build turns red; paste the error; restore.

### COMMENTS FOR THE JUDGES (ADR-0027)
Every exported name has a doc comment beginning with the name; the test says which property it proves; English, plain; no
host names, internal paths or credential names; no AI attribution.

### WHAT IS NOT IN SCOPE
- Any change under `internal/` (the interface and its implementations stay as they are).
- `examples/README.md` (the CLI master's file).
- The other adapter sections (guardrail, reviewer, testrun, tool); later orders.

### WHEN THIS ORDER CANNOT PROCEED
Write `SHOUT_<topic>.md` in the worktree, set `- state: blocked`, and stop.

### ARTIFACT CONFINEMENT
Every artifact this leg makes lives under `███████████████████████████████████████` in the worktree, except the files this
order names.

### RETURN
`_EXEC_STATUS.md`: `- state: stop-2-ready`, `- last commit:` <sha>, an `- abstract:` of at most 5 lines, each master claim a
measurement contradicted (in `- note:`), the commands run and their exit codes.
Check it with `saphan doc check --family exec-status _EXEC_STATUS.md` before you end: its exit code is the verdict.

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| how an order is bound to the laws | 1 |
| private workspace path | 4 |
