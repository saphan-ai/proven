# Examples

Each extension point adds its own row. This page starts with the output format. A later change adds the generator, the guardrail, the reviewer, the test runner, the tool and the evidence query.

| Name | Extension point | Explained in | Run |
|---|---|---|---|
| formatter | output format (`Formatter`) | [Adding an output format](../docs/ADAPTERS.md#adding-an-output-format) | `go test -count=1 ./examples/formatter` |
