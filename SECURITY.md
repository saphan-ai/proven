# Security

Report vulnerabilities privately through [GitHub security advisories](https://github.com/saphan-ai/proven/security/advisories/new). Include the Proven version, a minimal reproduction and the relevant configuration, with credentials removed.

## Trust boundaries

Repository content, model answers and generated tests are untrusted. Models propose hypotheses and edits. Proven constructs queries from registered templates and constructs runner arguments in code. It does not execute model-written commands.

Generated tests run in restricted containers with read-only checkouts, no network or inherited secrets, resource and time limits, and bounded cleanup. The host container runtime and its prepared images are trusted dependencies. Offline images must already contain required toolchains and project dependencies.

A model-based content-safety service is optional. Its absence is reported. Content classification does not replace process isolation, query validation or evidence rules.

The command model adapter executes an operator-configured program. That CLI owns its authentication and permissions; configure it for text generation without coding tools. Proven filters unrelated service credentials and cancels timed-out commands, but a local CLI is not equivalent to an isolated remote model endpoint.

## Data and evidence

Keep evidence stores, logs and persistent state outside the reviewed repository. Model keys and database credentials belong in the environment. Source adapters and imported datasets are operator-trusted inputs; review freshness and source provenance before relying on them.

A signed evidence pack binds bytes to a signing key and permits offline re-derivation of recorded verdicts. It does not prove that a run occurred or that every defect was found. Trust the verification key through an independent channel.

Query refutation uses a restricted same-row text rule and cannot override an observed test or tool proof. It remains a heuristic for query-only findings. Passing generated tests alone also does not establish that a change matches its intended specification.

## Incomplete reviews

JSON reports completion separately from findings. Required CI reviews fail with exit 3 if services fail or the review ends early. An unavailable analyzer or absent dataset is not a clean measurement.

Current changes and limitations are recorded in [CHANGELOG.md](CHANGELOG.md).
