# Tests and evaluation cases

`go test ./...` runs unit, regression and fixture tests without model accounts. External-service and container tests require explicit opt-in settings.

- `testdata/fixture/` holds recorded review scenarios and base/change pairs.
- `test/corpus/cases/` holds focused regression cases and their expected observations.
- `test/corpus/held-out/` holds separate evaluation cases.
- `test/conformance/meterless/` provides optional compatibility runners for the local context engines.

Fixture success does not establish real-world recall or precision. Keep expected findings tied to observable evidence. Never treat an unavailable tool or empty dataset as a clean result. See [contributing](../CONTRIBUTING.md).
