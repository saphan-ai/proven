# ADR-0038: A language is one small adapter: PHP, Java, C#, Rust and Ruby join Go, JavaScript and Python

**Status:** accepted · **Date:** 2026-10-23
**Decided by:** the owner · **Written by:** the advisor (claude-code, `claude-opus-5-5`), 2026-10-23
**Follows:** [ADR-0001](ADR-0001-go-behind-adapters.md) (Go behind adapters), [ADR-0023](ADR-0023-how-proven-s-evidence-is-loaded-through-zetaris.md) (OSV by package)
**Trigger:** the owner, 23 October: *"Which languages do we add? What is popular?"* and *"we can also say we give an adapter interface, so it is easy to extend."*

## Context

On 23 October Proven read dependencies from Go (`go.mod`, `go.sum`), JavaScript (npm) and Python (`requirements.txt`,
`poetry.lock`), and checked each changed package against OSV by version range, the known-exploited list and EPSS. OSV
covers every popular ecosystem; what a new language needs is only a reader for its manifest and lock file, and the
version comparison its ecosystem uses.

## Decision

1. **One public interface per ecosystem**, small enough to implement in one pull request: find the ecosystem's files
   in a change; list the packages and versions the change adds or raises; name the ecosystem as OSV names it; compare
   two versions by the ecosystem's own rules. Everything after that (OSV ranges, KEV, EPSS, the evidence, the report)
   is shared.
2. **Every adapter ships with a sample scenario and its tests**, so a contribution proves itself in sample mode.
3. **New languages, in this order:** PHP (Composer, `composer.lock`), Java (Maven, Gradle), C# (NuGet), Rust
   (`Cargo.lock`), Ruby (`Gemfile.lock`).
4. **`docs/ADAPTERS.md`** explains how to add a language, with the PHP adapter as the worked example.

## Consequences

- A user whose language is missing can add it without learning the rest of Proven.
- Each version scheme (Maven, NuGet, Composer, Cargo, RubyGems) has its own comparison rules; each adapter's tests
  cover them.
