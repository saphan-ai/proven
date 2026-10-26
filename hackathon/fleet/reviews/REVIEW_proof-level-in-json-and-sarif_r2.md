# REVIEW — proof-level-in-json-and-sarif · round 2

- verdict: accept-with-conditions
- reviewer: claude-code on macm5, claude-opus-5-5
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-25T04:53:32Z

## Summary

540616b holds all five properties: the core's values copied without re-deriving, JSON = SARIF via one helper, the base-vs-tip sample diff is only the new fields, and 7/7 writer mutants die.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — proof-level-in-json-and-sarif

### Ground
- Reviewer: claude-code leg (Opus 5.5), stream proof-level-in-json-and-sarif-review-r2, 2026-10-25, darwin/arm64, go1.26.10.
- `git rev-parse --short=7 HEAD` printed `540616b` (role-lock satisfied). Scope `git diff d4785e2...540616b`: docs/OUTPUT_FORMATS.md, internal/cli/format_json.go, internal/cli/format_sarif.go, internal/cli/format_prooflevel_test.go (new), internal/cli/output_contract_test.go. 5 files, +230/−16.
- Core read, not under review: internal/review/review.go:102-105 (ProofCounts), :169-209 (ProofLevel values, ProofLevels), :221-250 (ProofLevelOf), :269-270 (Finding.ProofLevel, "never empty in a Result").
- The author's self-description (commit message, the comments in the diff) was not used as evidence (R1).
- H1 note: the order's return section asks for `# REVIEW — proof-level-in-json-and-sarif-review-claude`. The engine-injected channel grammar names `# REVIEW — proof-level-in-json-and-sarif` and says that one binds the verdict to the reviewed stream. I used the engine's form. The master should correct the order template.
- Re-measure of the master's claim "build, vet ok, 49 packages ok, 0 FAIL":
  - `go build ./...` ok.
  - `go vet` stopped at internal/cli/sarifschema: the module github.com/santhosh-tekuri/jsonschema/v5 v5.3.1 is not in the cache, and the proxy fetch failed (`tls: failed to verify certificate: x509: OSStatus -26276`, sandbox).
  - `go test ./...`: 41 ok, 8 FAIL. Every FAIL is environmental and outside the diff: sarifschema setup failed (the same fetch); cmd/livecheck, internal/generator, internal/guardrail, internal/pipeline, internal/query, internal/review and internal/store panic with `httptest: failed to listen on a port: listen tcp6 [::1]:0: bind: operation not permitted`. 41+8 = 49, which matches the master's package count. I could not confirm the master's 0 FAIL on this host; see Unrun.
  - `go vet ./internal/cli` ok; `go test ./internal/cli -count=1` ok (76.8s) at the restored tip.

### Differential
Property 3 (nothing else moved), measured. A temporary in-package test captured `scan --sample --json|--sarif testdata/fixture/discount` twice: with the d4785e2 writers swapped in (`git checkout d4785e2 -- internal/cli/format_json.go internal/cli/format_sarif.go`) and with the 540616b writers. Then `diff base tip`:
- JSON: `+proof_counts {1,0,3,0}` before `findings`, and `+"proof_level"` on each of 4 findings (3× proved-by-data, 1× proved-by-test). The only other changed line is `trace_path`, a per-run t.TempDir name that also differs between two runs at the same SHA.
- SARIF: `+"proof_level"` in each of the 4 result `properties`, and `+proof_counts {1,0,3,0}` in run `properties`. The only other changed line is the same `trace_path` temp dir.
- Captures kept: █████████████████████████████████████████████████████████████████████,sarif}.{base,tip}.
- The contract-test goldens (output_contract_test.go) differ from the base only by the new expected keys, plus ProofLevel set on the two fixture inputs. No existing expectation was removed or relaxed.

Property 2 (JSON and SARIF agree): JSON (format_json.go:164) and SARIF (format_sarif.go:119) both call the one function `proofCountsFrom`, so the counts are identical by construction. Per finding, both write `string(finding.ProofLevel)` (format_json.go:193, format_sarif.go:135). Measured on the sample: the per-id levels and the counts are equal across the two outputs (format_prooflevel_test.go:35, :58).

### Unplug
I made each mutation at the tip, ran `go test ./internal/cli -count=1 [-run 'ProofLevel|Contract|SARIF|JSON']`, and restored the file with `git checkout 540616b -- <file>`:
- M0, the whole feature unplugged (base writers): TestSampleJSONAndSARIFCarryProofLevel, TestJSONDocumentIsTheContract and TestSARIFPropertiesAreTheContract are RED.
- M1, JSON findingFrom drops the `ProofLevel:` line: TestSample… and TestJSONDocumentIsTheContract are RED.
- M2, JSON tag renamed `proof_level`→`prooflevel`: TestSample… and TestJSONDocumentIsTheContract are RED.
- M3, `omitempty` on `proved-by-tool`: TestSample… is RED ("JSON/SARIF proof_counts missing proved-by-tool"). The contract tests stay green because their tool count is 1.
- M4, the count read from the wrong level (`ProvedByTool: counts[ProvedByData]`): all three are RED.
- M5, JSON derives the level from the evidence (`review.ProofLevelOf`): TestJSONDocumentIsTheContract is RED. The F1 fixture with test evidence but a tool level does its job.
- M6, SARIF result key renamed `proof_level`→`prooflevel`: TestSample… and TestSARIFPropertiesAreTheContract are RED.
- M7, SARIF run drops `proof_counts`: TestSample… and TestSARIFPropertiesAreTheContract are RED.

Every mutant died. Property 4's mutated fields: findings[].proof_level (JSON), result properties.proof_level (SARIF), proof_counts (JSON), and run properties.proof_counts (SARIF).

### Closure
- P1, the values are the core's: HOLDS. The writers copy `finding.ProofLevel` and `result.ProofCounts[level]` keyed by the review.* constants (format_json.go:184-189). No level is derived in either writer. M5 proves a re-derivation is caught.
- P2, JSON and SARIF agree: HOLDS (see Differential).
- P3, nothing else moved: HOLDS (see Differential).
- P4, a test reddens when a writer omits or renames a field: HOLDS for omission and renaming in both writers (M1–M7). "The SARIF schema test still passes": UNRUN on this host (see Unrun).
- P5, the docs: HOLDS as written. docs/OUTPUT_FORMATS.md:28 and :31 describe the keys, the zero-filled levels and the value set. docs :53 says both SARIF placements carry the same values as the JSON, and that is true by construction. One precision gap is owed (O2).

### Remainder
These are owed findings, not blocking (R6):
- O1 · internal/cli/format_json.go:50-51 and :169-189 · `jsonProofCounts` hard-codes the four levels as struct fields, while its comment says "in the order review.ProofLevels lists them".
  - No test binds the writer's keys to `review.ProofLevels`. format_prooflevel_test.go:136-141 is also a hard-coded list.
  - Trigger: a fifth level added to the core. Both writers would drop its count silently, and every test would stay green.
  - Remedy: a test that walks `review.ProofLevels` and asserts that each string is a key of both `proof_counts` objects. Or build the counts from `review.ProofLevels` (an ordered encoding).
  - Today the risk is low, because the values are a released contract.
- O2 · docs/OUTPUT_FORMATS.md:31 · "one of the four keys of `proof_counts`" is true as a set. But per review.go:182-185 and :224-250, a finding never carries `agreed-by-two-models-not-proven`, and that count is always 0 (opinions are not findings, ADR-0041).
  - A reader may expect findings at that level.
  - Remedy: one sentence saying the agreed level is reserved and stays 0.
- O3 · internal/cli/format_json.go:78-79 · the comment "always written, including when it is empty" sits beside the core's guarantee "never empty in a Result" (review.go:269). It is harmless, but it describes a state the core rules out. Cosmetic.

### Profile
None of the special profiles apply (R9): the diff touches output serialization and docs only, with no auth, no data access and no configuration. I used the base review acts R1–R3, R7, R10, R13 (the paired JSON/SARIF writers judged together, with a shared-shape test at format_prooflevel_test.go) and the docs act D1. There is no dependency delta (R4): go.mod and go.sum are untouched.

### Focus
The five properties in the order: the core's values, JSON/SARIF agreement, nothing else moved, an unplug test per field, and the docs' exactness.
Declined to judge (R11), for the master to rule on:
- the text renderer's proof output (not in this diff);
- whether `proof_counts` should precede `findings` in the JSON key order (additive, and no consumer contract names the order);
- the order-template H1 inconsistency (see Ground), which is the order's defect, not the diff's.

### Conditions
- C1 (verification, owed before the gate) · The order requires the SARIF schema test to pass, and I could not run it here. The master runs `go test ./internal/cli/sarifschema -count=1` on a host with module access and records ok. Class: verification-owed.
- C2 (verification, owed before the gate) · The master re-runs `go test ./...` on a host with a loopback listener and confirms 0 FAIL for cmd/livecheck, internal/generator, internal/guardrail, internal/pipeline, internal/query, internal/review and internal/store. None of them is in the diff's write-set. Class: verification-owed.
- O1, O2 and O3 (see Remainder) are owed work, not conditions of the gate.

### Unrun
- internal/cli/sarifschema (the SARIF schema test): the setup failed because module github.com/santhosh-tekuri/jsonschema/v5 v5.3.1 is not cached and the fetch was refused: `tls: failed to verify certificate: x509: OSStatus -26276`. `go vet ./...` stopped on the same package.
- 7 packages panicked on `httptest: failed to listen on a port: listen tcp6 [::1]:0: bind: operation not permitted`: cmd/livecheck, internal/generator, internal/guardrail, internal/pipeline, internal/query, internal/review and internal/store.
- Cleanup not done: deleting my temporary test file ██████████████████████████████████████ was refused (`"rm" is NOT in the ██████████████ list of role review … deny by default`). I emptied it to `package cli` plus a comment, and did not rebuild the deletion from other commands (law #57). The master deletes this one untracked file at teardown. Tracked files are byte-identical to 540616b (`git diff --stat` is empty).

## The review status as filed

## REVIEW — proof-level-in-json-and-sarif
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: 540616b holds all five properties: the core's values copied without re-deriving, JSON = SARIF via one helper, the base-vs-tip sample diff is only the new fields, and 7/7 writer mutants die.
  Conditions are verification only: the SARIF schema test and 7 loopback packages were UNRUN here (sandbox: no module fetch, no bind). The master runs them before the gate.
  Owed, non-blocking: the proof_counts keys are hard-coded rather than bound to review.ProofLevels; the docs don't say the agreed level is always 0.
- artifacts:
    - _REVIEW_NOTES.md
    - _REVIEW_STATUS.md
- updated: 2026-10-25

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| file and line in Saphan Studio's private code | 1 |
| policy binding name | 1 |
