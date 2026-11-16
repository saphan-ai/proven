# Contributing

Proven welcomes fixes, adapters, reproducible bug reports and documentation improvements.

## Development

Install the Go version named in `go.mod`, then run:

```sh
go build ./cmd/proven
go test ./...
go vet ./...
make check-links
```

Ordinary tests use local fixtures and test servers. They need no model key or sponsor account. Container and external-service tests are opt-in and name the setting that enables them. Development checks also use Python 3 and Node.js. Run the shell Action checks with `sh action/test.sh` and documentation-script tests with `node --test docs-site/scripts/*.test.mjs` when changing those surfaces.

## Changes

Open an issue for a substantial change before implementing it. A pull request should explain the problem, the resulting behavior and the checks performed. Include a regression test for a bug fix and update the relevant user documentation.

Keep model transports, evidence stores, analyzers and output formats behind the existing adapter boundaries. Prefer small packages whose logic and tests can be reused independently. Report unsupported and incomplete analysis explicitly.

Generated documentation is refreshed with:

```sh
sh docs-site/scripts/generate.sh docs-site/generated
```

## Reporting a review problem

Include the Proven version, a minimal base/change example, the expected finding and the actual result. Remove credentials and private source before sharing logs or evidence packs. For vulnerabilities in Proven itself, follow [SECURITY.md](SECURITY.md).

Code contributions use [Apache-2.0](LICENSE). Documentation uses [CC BY 4.0](LICENSE-docs). Preserve applicable copyright and attribution notices.
