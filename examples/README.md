# Examples

Small adapters implement interfaces without real services. Their tests need no accounts. They are teaching implementations, not production model/runner replacements. The [adapter guide](../docs/ADAPTERS.md) explains wiring and failures.

| Name | Extension point | Explained in | Run |
|---|---|---|---|
| formatter | output format (`Formatter`) | [Adding an output format](../docs/ADAPTERS.md#adding-an-output-format) | `go test -count=1 ./examples/formatter` |
| generator | fixed-response model | [Model provider](../docs/ADAPTERS.md#adding-a-model-provider) | `go test -count=1 ./examples/generator` |
| guardrail | fixed safety classifier | [Guardrail](../docs/ADAPTERS.md#adding-a-guardrail) | `go test -count=1 ./examples/guardrail` |
| reviewer | fixed verdict | [Reviewer](../docs/ADAPTERS.md#adding-a-reviewer) | `go test -count=1 ./examples/reviewer` |
| testrun | recorded outcomes | [Test run](../docs/ADAPTERS.md#adding-a-test-run) | `go test -count=1 ./examples/testrun` |
| tool | fixed detector | [Tool](../docs/ADAPTERS.md#adding-a-tool) | `go test -count=1 ./examples/tool` |
| data | synthetic snapshot | [Git/SQLite walkthrough](../docs/DATA_WALKTHROUGH.md) | Follow the walkthrough; no model or Docker. |
| cli | customer-installed Claude Code / Copilot | [CLI recipes](cli/README.md) | `python3 examples/cli/test_wrappers.py` (fake executables, no accounts). |

Run all adapter examples with `go test ./examples/...`. For recorded reviews, use the [sample catalogue](../docs/SAMPLE_MODE.md). The data example is invented teaching data and must not be used as an operational feed.
