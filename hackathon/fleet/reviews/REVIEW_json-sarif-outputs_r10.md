# REVIEW — json-sarif-outputs · round 10

- verdict: accept-with-conditions
- reviewer: qwen-code, glm-5.3
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 1 NOTE
- reviewed: see the notes below
- updated: 2026-10-24T14:00:42Z

## Summary

The JSON and SARIF writers meet every mapping docs/OUTPUT_FORMATS.md names; the mandated go test -count=1 ./internal/cli/... is green and the SARIF schema test is self-guarding. Three conditions: guardrail_statement is promised by the docs' SARIF sentence but carried by neither machine writer; tool evidence reaches JSON and SARIF as a contentless kind; the SARIF message priority is untested and its docs sentence is ambiguous.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — json-sarif-outputs

### Ground

- Role lock: `git rev-parse --short=7 HEAD` printed `386bbf0` — run twice, at the start and again before this file was written; both runs printed `386bbf0`. It matches the order's `- reviewed-sha: 386bbf0`.
- Read: `internal/cli/format_json.go`, `internal/cli/format_sarif.go`, `internal/cli/format.go` (the seam both writers share: `FormatterFor`, `familySentence`), `internal/cli/format_json_test.go`, `internal/cli/format_sarif_test.go`, `internal/cli/output_contract_test.go` (the contract tests both writers share), `internal/cli/guardrail_statement_test.go`, `internal/cli/format_tool_test.go`, `internal/cli/sarifschema/doc.go` and `schema_test.go`, `docs/OUTPUT_FORMATS.md`, `internal/review/review.go` (Result, Severity, Evidence), `internal/pipeline/pipeline.go` and `internal/review/open.go` (where GuardrailStatement and tool evidence are produced), `testdata/sarif/sarif-schema-2.1.0.json`.
- Toolchain (lang-go, C1): `go.mod` pins `go 1.26`; one dependency, `github.com/santhosh-tekuri/jsonschema/v5 v5.3.1`. No `.github/workflows` files exist in this repository, so no CI-gated command was found to mirror; the order's own command is the gate. The local `go` version is UNMEASURED (permission refusal, Unrun).
- Order point 4, run verbatim: `go test -count=1 ./internal/cli/...` →

  ```
  ok  	github.com/saphan-ai/proven/internal/cli	13.556s
  ok  	github.com/saphan-ai/proven/internal/cli/sarifschema	0.454s
  ```

  Exit code 0. No in-scope test file carries a build tag, `t.Skip` or a `testing.Short()` guard (measured by reading each file), so the `ok` covers them.
- The vendored schema is the OASIS SARIF 2.1.0 errata01 schema, read from the file: draft-04, title "Static Analysis Results Format (SARIF) Version 2.1.0 JSON Schema", id `https://docs.oasis-open.org/sarif/sarif/v2.1.0/errata01/os/schemas/sarif-schema-2.1.0.json`, `version` enum `["2.1.0"]`, `startLine` minimum 1 (line 1716), top-level `additionalProperties` false.

### Differential

The docs contract against the two writers at `386bbf0`, property by property (the order's four points).

1. **JSON keys.** The table's 16 top-level keys are all fields of `jsonDocument` (`internal/cli/format_json.go:34`) with no omitempty; the finding's 9 keys, the location's 3, the test's 6 and the query's 5 are all present; no key outside the table is emitted — the document is exactly the table's shape. "Of one evidence entry, only the object named by its kind is present" is the two omitempty pointers on `test` and `query`. Empty lists are `[]`, never `null`, through `stringSlice` and the length-zero `make`s. `TestJSONDocumentIsTheContract` pins the whole decoded document against a literal whose every value is unique per field, and `assertNamesDocumented` fails for any name the docs do not carry. MET for the key set; the kind *value* divergence is finding 2 below.
2. **SARIF mapping.** `level` = `sarifLevel` (`internal/cli/format_sarif.go:172`): WARNING→warning, ERROR→error, INFO and any other→note — the docs' line for line. `ruleId` = `sarifRuleID` (161): `rule_id`, or `id` when the finding has no rule. Locations (123): no path → no location (`artifactURI` at 207 returns `""`); a start line below 1 → no region; an end line before the start line → omitted. Result properties (131): exactly `id`, `rule_id`, `title`, `severity`, `unresolved`, `evidence` — and `TestSARIFPropertiesAreTheContract` asserts the evidence array equals the JSON's own. Run properties (107): exactly the nine keys the docs name, each asserted equal to the JSON key of that name. One result per finding, one run, version `2.1.0`, driver name `proven`, and no field the docs do not name. MET; the completeness sentence is finding 1, the message sentence is finding 3.
3. **Schema test:** Unplug below.
4. **Suite:** Ground above, pasted.

### Unplug

- The instrument is real: `internal/cli/sarifschema/schema_test.go` compiles the vendored OASIS schema from disk (no network; see `doc.go`) and validates the writer's own output — the sample scan's file and every edge shape.
- Measured, inside the mandated green run: `TestSampleSARIFMatchesTheSchema`'s negative control removes `version` from a real sample log and requires the schema to refuse it, naming `version`. Green means the refusal fired. The mutated field is `version` — exactly the field the control guards.
- Named unplug (law #41): unplug the validation — `validateSARIF` a no-op, or the vendored schema swapped for a permissive file — and `TestSampleSARIFMatchesTheSchema` fails with "schema accepted a log with version removed"; a writer that emits a malformed log (startLine 0, say) fails `TestSARIFShapesMatchTheSchema` and the sample check with "schema: …" plus the log body. The surface that turns red is `go test -count=1 ./internal/cli/...` (exit FAIL); the one woken is whoever runs the suite — the executor and the master at the gate.
- Writer guards, read-level (R3): swapping `sarifLevel`'s warning and error arms dies in `TestSARIFLevelFollowsSeverity` (levels read per ruleId); dropping the startLine guard dies in the "line 0" shape (Region must be nil) and in the schema (minimum 1); dropping the endLine guard dies in the "end line before start line" shape; dropping the ruleId fallback dies in `TestSARIFLevelFollowsSeverity`'s four rule-less findings; dropping `stringSlice` dies in `TestJSONEmptyListsAreArraysNotNull`; dropping one run-property key dies in its own subtest of `TestSARIFCarriesEveryTextOnlyLine`. **One mutant survives:** inverting `sarifMessageText`'s precedence (title before message) — measured by grep, no test in `internal/cli` asserts `message.text` at all. That is finding 3's second half.

### Closure

Verdict: **accept-with-conditions**. Every mapping the docs name is what the writers do; the mandated suite is green; the schema test validates against the genuine SARIF 2.1.0 and is self-guarding. The three conditions are owed work: two are cross-format gaps opened around the writers by later streams at the shared base, one is a missing test plus an ambiguous sentence. None of the three makes the W2 writers unfaithful to the written contract on the shapes the docs describe; all three carry locations and remedies in Remainder.

### Remainder

Findings for the master (FILE·LINE·CONTROL):

1. **WARNING — guardrail_statement is promised by the docs and carried by no machine writer.** `docs/OUTPUT_FORMATS.md`, SARIF section, last sentence: "…so the lines the text output prints beside the findings are in a SARIF file too." The text writer prints `result.GuardrailStatement` (`internal/cli/format.go:73`), set when the guardrail mode is off (`internal/pipeline/pipeline.go:280`, reached from `internal/review/open.go:123` when `gs.Mode == guardrail.ModeOff`). `internal/cli/format_json.go:34` (`jsonDocument`) and `internal/cli/format_sarif.go:107` (run properties) carry no `guardrail_statement`, and the docs' JSON table lists none. Control, read-level: the field's whole path is pipeline → `Result` (`internal/review/review.go:118`) → text print; the two machine writers have no field for it. A JSON-only or SARIF-only consumer of a guardrail-off review never sees the line the text reader sees. Remedy, the master's ruling: either add `guardrail_statement` to both writers, the JSON table and the SARIF key list, with a per-key subtest like `TestSARIFCarriesEveryTextOnlyLine`'s, or narrow the docs sentence to the nine keys it names. Cost if unfixed: a machine consumer takes an unguarded review for a guarded one. Attribution: the line was added to the text writer by stream `security-piece-5-fixes` (8bdf2c5), which touched neither machine writer nor docs.
2. **WARNING — tool evidence is dropped to a contentless kind.** `internal/review/review.go:331` defines kind `tool` (ADR-0040); the live pipeline produces it (`internal/pipeline/pipeline.go:1554`, `internal/tool/tool.go:89`) and the text writer prints the whole run (`internal/cli/format.go:324`, `format_tool_test.go`). `evidenceFrom` (`internal/cli/format_json.go:178`) copies only `Test` and `Query`, both omitempty, so a tool entry marshals as `{"kind":"tool"}` with no object — in JSON `findings[].evidence` and in SARIF `properties.evidence` alike. The docs table states kind is `test` or `query` and names objects only for those two; the table's key set stays closed, but the value set the writer can emit is wider than the documented one, and the tool's command and both outputs are lost to both machine formats. Remedy, the master's ruling: extend the table and both writers with the tool object (a row in both shape tables), or write into the docs that tool evidence is text-only. Cost if unfixed: a proved-by-tool finding reaches machine consumers with evidence that carries nothing. Attribution: the kind was added by stream `a-tool-run-is-evidence` (5b3fe22), which touched the text writer only.
3. **NOTE, owed — the SARIF message order is untested and the sentence is ambiguous.** `docs/OUTPUT_FORMATS.md`: "The message is the finding's message, then its title, then its id." `sarifMessageText` (`internal/cli/format_sarif.go:188`) implements first-non-empty precedence — message, else title, else id, else "finding" — which is one reading of "then"; a concatenation reading is also possible and the docs pick neither. No test asserts `message.text` (grep over `internal/cli`: zero assertions on it), so an inverted precedence ships silently — the surviving mutant of the Unplug section. Remedy: pin `message.text` with three cases (message set; message empty and title set; both empty and id set) and reword the sentence to say which wins.

Declined to judge (R11):

- `artifactURI` trims the path (`internal/cli/format_sarif.go:207`), so a path of only spaces is "no path" and gets no location; the docs speak only of no path.
- `ruleId` is omitted entirely when both `rule_id` and `id` are empty; the docs do not state that case and SARIF allows the omission.
- `severity` is written as `string(finding.Severity)` with no enum check; the domain defines exactly INFO/WARNING/ERROR (`internal/review/review.go`) and the planner copies into one of these, so the table's value set holds for every input the domain can produce today.
- `review.Result` also carries `ProofCounts`, `Reviewers`, `Disagreements` and `SkippedPaths`; no machine format carries them and no docs row promises them. Noted at the boundary, not judged.

### Profile

No profile is named by the order; the judged surface is its four properties over the two writers, their tests and the docs contract. Check ranking (skill testing-bar): external — the vendored OASIS SARIF 2.1.0 schema, a published rule set no working-but-wrong writer can satisfy, the strongest check the change has; suite — the repository's own tests (`TestJSONDocumentIsTheContract`, `TestSARIFPropertiesAreTheContract`, the shape tables), author-satisfiable, the weaker half; project-level — none: no linter, threshold or CI configuration exists in the repository. No numeric bar is declared anywhere (metrics default none), so there is nothing to ratchet; no bar-weakening move was found in scope — no skip, no loosened assertion, no suppression, no deleted test.

### Focus

In scope, per the order: the JSON table and the SARIF section of `docs/OUTPUT_FORMATS.md` against `format_json.go` and `format_sarif.go` and their tests (`format_json_test.go`, `format_sarif_test.go`, `output_contract_test.go`, `internal/cli/sarifschema`); the schema test's reality; the mandated suite run. Out of scope, per the order: output-path identity and file-writing rules (docs paragraph 2 and `config.go`), control or escape bytes in any output, the text writer, licences. The text writer was read only at the seam the machine writers share (`FormatterFor`, `familySentence`) and where its lines define the cross-format promise the docs make (the guardrail line, tool evidence).

### Conditions

accept-with-conditions carries these; each is owed work with a named remedy (detail in Remainder):

1. guardrail_statement: the master rules writer-extension against docs-narrowing (finding 1).
2. tool evidence: the master rules tool-object extension against an explicit text-only ruling in the docs (finding 2).
3. message order: add the message.text-pinning cases and reword the docs sentence (finding 3).

Changing a condition's class upward is the master's act, with a sentence in the gate note.

### Unrun

- `saphan doc check _REVIEW_STATUS.md --family review-status`: refused by the permission layer, verbatim: `Qwen Code requires permission to use "run_shell_command", but that permission was declined (non-interactive mode cannot prompt for confirmation).` The order mandates this check; it could not be run. The status file was written to the engine-emitted skeleton's closed form — the exact H1 and the six keys — which is the shape that check reads.
- `go version`: refused by the permission layer, verbatim: `Qwen Code requires permission to use "run_shell_command", but that permission was declined (non-interactive mode cannot prompt for confirmation).` The local toolchain version is unmeasured; the pinned version is `go 1.26` (`go.mod`). The mandated test command itself ran and passed, so the build-and-test evidence stands.
- `go vet` and `gofmt`: not run — the repository has no CI configuration to mirror, the order names only the test command, and after the `go version` refusal I made no further attempts at go subcommands beyond the one the order mandates (law #23: a denial is a permissions fact, not a missing binary; do not hunt).
- Executed mutation runs: not run — a mutation harness needs `go build` or `go test` in forms the permission layer did not admit; the guard analysis in Unplug is read-level, and the one executed negative control is the schema test's own, which ran green inside the mandated suite. The surviving mutant is named as finding 3, not measured as a kill.

## The review status as filed

## REVIEW — json-sarif-outputs
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: The JSON and SARIF writers meet every mapping docs/OUTPUT_FORMATS.md names; the mandated go test -count=1 ./internal/cli/... is green and the SARIF schema test is self-guarding. Three conditions: guardrail_statement is promised by the docs' SARIF sentence but carried by neither machine writer; tool evidence reaches JSON and SARIF as a contentless kind; the SARIF message priority is untested and its docs sentence is ambiguous.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-24T13:53:07Z

## REDACTED

none
