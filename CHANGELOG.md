# Changelog

## Unreleased

- Support an independently configured command transport for the second reviewer, including API + CLI and two CLI configurations. Add API credential setup, customer-built CLI recipes, Claude Code/Copilot examples and real GitHub screenshots with labelled recorded evidence.

- Preserve final unresolved bounded-planning hypotheses in all report formats and traces while continuing evidence checks for identified risks; required completion exits 3 and never counts those hypotheses as findings.
- Make model request limits, three clocks, provider options and shared recovery configurable, with per-agent and second-reviewer overrides and per-request stream observations.
- Pin updated git/musl runtime packages above the Alpine base and gate image publication on successful checks and HIGH/CRITICAL scans for both architectures.
- Document readiness, source adapter versus feed-connector status, bounded-planning behavior and the public source roadmap.
- Add guided recorded and real Git/SQLite tutorials, model profiles, runner preparation, operations, troubleshooting, observability, fix validation, language coverage, design rationale and evaluation guidance. Retain the detailed adapter/evidence/sample references and organize the documentation site around learning paths.

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
