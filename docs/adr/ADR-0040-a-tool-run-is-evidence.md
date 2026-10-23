# ADR-0040: A tool run is evidence: deterministic tools plug in through an adapter, like languages

**Status:** accepted · **Date:** 2026-10-23
**Decided by:** the owner · **Written by:** the advisor (claude-code, `claude-opus-5-5`), 2026-10-23
**Amends:** [ADR-0002](ADR-0002-evidence-is-a-test-or-a-query-through-zetaris.md) (evidence was a test or a query; now also a tool's run)
**Follows:** [ADR-0025](ADR-0025-the-prover-runs-the-change-with-no-network-and-no-secret.md) (the Prover's container), [ADR-0038](ADR-0038-a-language-is-one-small-adapter.md) (adapters)
**Trigger:** the owner, 23 October: *"deterministic — yes, as another adapter, for tools."*

## Decision

1. **A third kind of evidence: a deterministic tool's run**, compared between the base and the change: the tool, its
   version, the exact command, and the output that differs. A finding is reported when the change produces a result the
   base did not.
2. **Tools plug in through a public adapter interface**, as languages do: which files it applies to, the command, how
   to read its output into findings, and how to tell base from change.
3. **First tools:** the Go race detector (`go test -race`), fuzzing of the changed function (`go test -fuzz`; the
   evidence is the input that breaks it), a static analyser (gosec), and coverage of the changed lines.
4. **Every tool runs inside the Prover's container**, with no network and no secret (ADR-0025), with a time limit
   from the settings.
5. **Each tool adapter ships with a sample scenario and its tests**, documented in `docs/ADAPTERS.md`.

## Consequences

- Proven can prove a data race or a crash, not only a failing assertion.
- Tools take time; each has its own limit, and a tool that runs out of time is named, not counted as clean.
