# ADR-0043: A vulnerable dependency is checked for reachability: does the code call the vulnerable function?

**Status:** accepted · **Date:** 2026-10-23
**Decided by:** the owner · **Written by:** the advisor (claude-code, `claude-opus-5-5`), 2026-10-23
**Follows:** [ADR-0040](ADR-0040-a-tool-run-is-evidence.md) (a tool run is evidence)
**Trigger:** the owner, 23 October: *"we have to show something more."*

## Context

A dependency in an affected version range is a risk; a vulnerable function the code actually calls is a much larger
one. For Go, the Go project's own `govulncheck` answers that question from the call graph.

## Decision

1. **`govulncheck` is a tool adapter (ADR-0040)**: for a Go change, Proven runs it on the base and on the change, in
   the Prover's container, with a local copy of its vulnerability database (no network).
2. **Each dependency finding says whether the vulnerable function is reached**: *reached* (with the call path
   `govulncheck` printed), *imported but not reached*, or *not measured* for ecosystems without such a tool.
3. **A reached vulnerability ranks above one that is only present**; a finding is never dropped because it is not
   reached, only ranked lower and labelled.

## Consequences

- Less noise: the reader sees first the vulnerabilities their code actually calls.
- Go first; other ecosystems say *not measured* until a tool adapter exists for them.
