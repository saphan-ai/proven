// Package sarifschema holds the tests that check Proven's SARIF output against
// the vendored OASIS SARIF 2.1.0 schema.
//
// They live in their own package because they need a test-only module (the
// JSON Schema validator). In internal/cli that module made the whole package
// fail to build on a machine that has not cached it; here only these tests
// depend on it, and they reach the writer through the exported seam
// (cli.Run and cli.FormatterFor).
package sarifschema
