# ADR-0039: More evidence sources through Zetaris: malicious packages, project health, CVSS and licences

**Status:** accepted · **Date:** 2026-10-23
**Decided by:** the owner · **Written by:** the advisor (claude-code, `claude-opus-5-5`), 2026-10-23
**Follows:** [ADR-0004](ADR-0004-first-evidence-sources.md) (the first sources), [ADR-0023](ADR-0023-how-proven-s-evidence-is-loaded-through-zetaris.md) (loaded through Zetaris), [ADR-0032](ADR-0032-proven-builds-every-query-from-a-template.md) (queries from templates)
**Trigger:** the owner, 23 October: *"and other databases as sources?"*

## Decision

Four more sources, each loaded into Zetaris and queried from a template like the first ones:

1. **OpenSSF Malicious Packages** (OSV entries with the `MAL-` prefix): a dependency published on purpose to harm is
   reported as such, with the advisory's row as evidence.
2. **Project health** (OpenSSF Scorecard, through deps.dev): a newly added dependency with no maintainers, unsigned
   releases, or a release only days old is reported, with its scores as evidence.
3. **CVSS from NVD**, shown beside EPSS: how severe a vulnerability is, next to how likely it is to be exploited.
4. **Licences (SPDX)**: a new dependency whose licence the organisation's policy does not allow is reported
   ([ADR-0036](ADR-0036-an-organisation-policy-and-exceptions-that-expire.md)).

Each source has a sample table, so sample mode covers it with no network.

## Consequences

- Proven sees supply-chain attacks, not only known vulnerabilities.
- Four more tables to keep fresh; each finding states the snapshot date of its row.
