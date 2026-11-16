# Changelog

## Unreleased

- Add standalone development, configuration and contribution guides.
- Add neutral OpenAI-compatible and operator-configured CLI model transports, including optional authentication for local endpoints.
- Permit deterministic analysis with model stages disabled. Retain optional legacy providers and local engines.
- Add SQLite and PostgreSQL evidence adapters, transactional table import and content-based source identities.
- Extract reusable report types and outcome rules; evidence packs no longer depend on reviewer initialization.
- Prevent unrelated denial rows from suppressing findings and prevent lexical query refutations overriding independent test/tool proof.
- Decode govulncheck as a stream of JSON values, including indented output; reject malformed streams as measurements.
- Separate completion policy from findings policy, with exit 3 for incomplete required reviews.
- Build the GitHub Action from its referenced source; add account-free CI checks and tagged build artifacts.

Existing JSON fields, SARIF, Markdown, HTML, evidence packs, fixes, autofix, policy and local engines remain supported. JSON adds `completion_status`. CI and Action completion defaults change; see [configuration](docs/CONFIGURATION.md).
