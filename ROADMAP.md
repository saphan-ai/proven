# Roadmap

This roadmap distinguishes planned work from shipped behavior. There are no promised release dates. The current project is a development preview; see [readiness](docs/READINESS.md).

## Reliable operation

- Preserve final bounded-planning hypotheses in reports and traces while completing evidence checks for identified risks. This is implemented in the unreleased code; compare bounded and whole-change quality before recommending an enforcement profile.
- Improve reporting and recovery for mid-plan carry overflow and other failures that currently leave only a trace.
- Evaluate a final reconciliation step for cross-file claims, with an explicit call/time budget and tests showing that no unresolved claim disappears silently.
- Publish scanned images with updated runtime packages, SBOM/provenance and a recorded digest. Maintain per-model configuration instead of assuming one provider's defaults.
- Measure detection and misses on controlled defects and changes without defects. Extend generated-test proving beyond Go through runner adapters.

## Evidence coverage

1. Add reproducible acquisition, normalization and refresh for OSV, CISA KEV and FIRST EPSS. Existing query adapters and manual imports are available; automated feeds are not.
2. Populate the existing malicious-package, project-health, CVSS and licence adapters with validated upstream connectors. Add package/version mappings, source terms and snapshot/freshness rules.
3. Add complementary GitHub advisory, CWE, repository-history and issue/PR sources where they provide useful evidence. Deduplicate advisory identities and keep explanatory context separate from proof.
4. Add typed source-specific contradiction rules and configurable licence policies; the current text and licence-cell heuristics have documented limits.

The [source status table](docs/SOURCES.md) gives the exact boundary for each source. A new source is ready only with its connector/import path, a small fixture, negative controls, provenance, missing-data behavior and user documentation. A new database backend does not complete this work.

## Reusable review components

Keep deterministic adjudication, report types, model-call observations and their tests easy to take into another reviewer. Proven remains independently runnable. Reusing selected code in another product does not require connecting that product's core directly to the Proven application.
