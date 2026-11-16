# Working on Proven

Proven is an independent Apache-2.0 Go project. Write code and documentation in plain English.

- Keep changes focused. Preserve CLI, report and evidence-pack compatibility, or document a migration.
- Keep decision rules deterministic. A model proposes hypotheses; executable evidence determines findings.
- Keep useful logic in small packages with explicit inputs. Read configuration and perform I/O at the application boundary. Selected code must be easy to reuse with its tests.
- Models, databases, safety services and container runtimes belong behind adapters. No sponsor account is required to build or test the project.
- Never execute model-provided commands. Preserve the runner's isolation, resource limits, secret filtering and cleanup reporting.
- Never commit credentials, private infrastructure details or unrelated workspace files.
- Test changed behavior, including failures and incomplete runs. Use fixtures for ordinary tests; live integration tests must be explicitly enabled.
- Run `go test ./...`, `go vet ./...`, `make check-links` and the relevant Action or documentation checks before delivery. Report checks that could not run.
- Do not push or publish a release unless the user requests it.
