# Exercise the language adapters

These small images test the shared proof/fix contract without model accounts. They are test fixtures, not complete runtimes for an arbitrary application.

From the Proven checkout, with Docker running:

```sh
docker build -f examples/runners/Dockerfile --target python -t proven-test-python:local .
docker build -f examples/runners/Dockerfile --target node -t proven-test-node:local .
PROVEN_TEST_LANGUAGE_IMAGES='{"pytest":"proven-test-python:local","node":"proven-test-node:local","vitest":"proven-test-node:local"}' \
  go test -count=1 -v ./internal/testrun -run TestLanguageContainers
```

The opt-in suite checks actual containers: base/head regression proof, passing controls, fix reruns, neighbouring tests, syntax/import failures, unittest classes, nested Node suites and Vitest hook errors. Ordinary `go test ./...` does not start this integration suite unless the environment variable is set.

For your application, prepare its trusted dependencies in the test image before review. The proof run has no network and does not install packages. The CLI Action can prepare a simple Go, npm or requirements.txt Python project automatically; use its `test-image` input for other layouts. See [language coverage](../../docs/LANGUAGES.md), [runner preparation](../../docs/RUNNER.md) and [the CLI Action](../../docs/CLI_ACTION.md).
