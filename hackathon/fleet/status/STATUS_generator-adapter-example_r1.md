# STATUS — generator-adapter-example · round 1

- state: ready-for-review
- executor: agent-loop
- model: nvidia/nemotron-3.5-super-vl-120b-a12b
- pool: not recorded
- cost: not recorded
- duration: not recorded
- last commit: 653c006cddd88ff97fe225c71266ecbc467cd306
- updated: 2026-10-25T07:40:00Z

## Summary

examples/generator holds Table, an adapter that answers Complete from a fixed table keyed by agent name, with no network, no secret and no environment variable; its test checks one answer and one refusal (unknown agent returns ErrUnknownAgent) and a compile-time check binds Table to internal/generator.Generator. docs/ADAPTERS.md gains "Adding a model provider": the Generator interface and its Complete(ctx, agent, prompt) method, PROVEN_GENERATOR_MODE choosing nightshift or nvidia, how sample mode treats it, and a pointer to the example; each fact names its file. go build, go vet, go test -count=1 ./examples/... and go test -count=1 ./... all exit 0. The negative control (method renamed) made the build fail with "missing method Complete", then was restored.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — generator-adapter-example
- state: stop-2-ready
- last commit: 653c006cddd88ff97fe225c71266ecbc467cd306
- done: 653c006 [generator-adapter-example] docs: a model-provider adapter section — the generator example plus its ADAPTERS.md section, build, vet and all tests green
- abstract: examples/generator holds Table, an adapter that answers Complete from a fixed table keyed by agent name, with no network, no secret and no environment variable; its test checks one answer and one refusal (unknown agent returns ErrUnknownAgent) and a compile-time check binds Table to internal/generator.Generator. docs/ADAPTERS.md gains "Adding a model provider": the Generator interface and its Complete(ctx, agent, prompt) method, PROVEN_GENERATOR_MODE choosing nightshift or nvidia, how sample mode treats it, and a pointer to the example; each fact names its file. go build, go vet, go test -count=1 ./examples/... and go test -count=1 ./... all exit 0. The negative control (method renamed) made the build fail with "missing method Complete", then was restored.
- note: master claim 1 — examples/formatter/ is the example shape and go test ./... runs it: measured `go test -count=1 ./examples/formatter` → ok, and `go test -count=1 ./...` lists `ok github.com/saphan-ai/proven/examples/formatter`. Confirmed at base and tip, not contradicted. Master claim 2 — docs/ADAPTERS.md has sections for languages, tools, sources, the evidence store and output formats, and none for the generator: measured before the change the page had exactly those five `## Adding ...` headings and no generator heading; after the change it has six, including "Adding a model provider". Confirmed at base; the "none for the generator" part no longer holds, by design. Commands run: `git rev-parse --short=7 HEAD` → exit 0, prints 931e2e8 (base required by the order); `go build ./...` → exit 0, on base 931e2e8 and tip 653c006; `go vet ./...` → exit 0, on base and tip; `go test -count=1 ./examples/...` → exit 0: ok examples/formatter 0.003s; ok examples/generator 0.002s; `go test -count=1 ./...` → exit 0: every package ok, no failures and no test that cannot bind a port on this seat; negative control — `Complete` renamed to `CompleteX` in examples/generator/table.go, `go build ./...` → exit 1: `examples/generator/table.go:19:29: ... Table does not implement "github.com/saphan-ai/proven/internal/generator".Generator (missing method Complete)`; file restored, then `go build`, `go vet`, `go test -count=1 ./examples/...` → exit 0 again.
- artifacts:
  - examples/generator/table.go
  - examples/generator/table_test.go
  - docs/ADAPTERS.md

## REDACTED

none
