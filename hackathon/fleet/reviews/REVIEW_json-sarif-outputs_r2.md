# REVIEW — json-sarif-outputs · round 2

- verdict: reject
- reviewer: claude-code on macm5, claude-opus-5-5
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-23T19:15:03Z

## Summary

f6b6b1f fails order property 3: -o and --sarif-output reaching one file through a symlinked directory exit 0, and the SARIF overwrites the JSON (reproduced; config.go:441-463 compares lexically only).

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — json-sarif-outputs

### Ground
- Reviewer: claude-code (claude-opus-5-5), REVIEW leg `json-sarif-outputs-review`, 2026-10-24.
- `git rev-parse --short=7 HEAD` printed `f6b6b1f`. Scope: `git diff main...f6b6b1f` (4 commits: 7281552, 250301b, 3aabc4b, f6b6b1f), 18 files, +4959/-12.
- `git diff --stat main...f6b6b1f -- AGENTS.md internal/review` printed nothing (boundary held).
- The author's and master's own claims were set aside (R1); every result below was re-measured on this seat.
- Seat limits, measured: the sandbox refuses network. (a) `go test ./internal/cli/` fails `[setup failed]`: `github.com/santhosh-tekuri/jsonschema/█████████` is not in the confined module cache and the TLS fetch to proxy.golang.org is refused (`x509: OSStatus -26276`). (b) Six tests in other packages (cmd/livecheck, generator, guardrail, pipeline, query, review) panic with `listen tcp6 [::1]:0: bind: operation not permitted`. These are outside W2 and come from the seat.
- To measure `internal/cli` without bypassing the sandbox, every cli run used `go test -overlay` with `██████████████████████████████████` mapped to empty. The source tree was not edited. Result: `ok github.com/saphan-ai/proven/internal/cli 23.349s`; `go vet` on the same overlay: clean.

### Differential
Mutants, each applied by `-overlay` (the file itself unchanged) and run as `go test -count=1 ./internal/cli/`:

| id | mutation | file:line | result |
|---|---|---|---|
| m1 | remove the `refuseCollidingOutputs` call | internal/cli/config.go:390 | KILLED: TestCollidingOutputPathsAreRefused, TestCollisionReproduction |
| m2 | swap the WARNING and ERROR arms | internal/cli/format_sarif.go:174-177 | KILLED: TestSARIFLevelFollowsSeverity |
| m3 | rename the json tag `class_rule` to `classRule` | internal/cli/format_json.go:36 | KILLED: TestJSONDocumentIsTheContract |
| m4 | drop SARIF result `properties.evidence` | internal/cli/format_sarif.go:137 | KILLED: TestSARIFShapes, TestSARIFPropertiesAreTheContract |
| m5 | drop SARIF run `family_sentence` | internal/cli/format_sarif.go:117 | KILLED: TestSARIFCarriesEveryTextOnlyLine, TestSARIFPropertiesAreTheContract |
| m6 | `reviewer_statements` without `stringSlice` (nil marshals as null) | internal/cli/format_json.go:149 | KILLED: TestJSONEmptyListsAreArraysNotNull |
| m7 | region guard `>= 1` changed to `>= 0` (startLine 0) | internal/cli/format_sarif.go:147 | KILLED: TestSARIFShapes ("line 0") |
| m8 | ignore the error of each `--*-output` write | internal/cli/run.go:237 | KILLED: TestOutputWriteFailureExits2 |
| m9 | `filepath.Abs` replaced by `filepath.Clean` | internal/cli/config.go:452 | **SURVIVED**: the whole package passes |
| m10 | JSON `unresolved` always false | internal/cli/format_json.go:169 | KILLED: TestJSONWriterEmitsFindings, TestJSONDocumentIsTheContract |

m9 surviving means no test pins the "relative and absolute spelling" case that order property 3 names. The suite's collision test uses absolute paths only (outputs_test.go:38-39).

### Unplug
- Property 3, probe (`review/probe_test.go`, added only by overlay as `████████████████████████████████████`). It runs `scan --json` against the capture reviewer and logs the exit:
  - **symlinked directory**: `-o <dir>/real/r.json --sarif-output <dir>/link/r.json`, where `link -> real`. Result: **exit=0, reviewRan=true**, and `real/r.json` starts `{"$schema": "https://docs.oasis-open.…`. The SARIF body replaced the JSON body. This is the round-2 BLOCKING defect, reproduced through the spelling the order names.
  - **relative and absolute** (`-o r.json` with `--sarif-output <cwd>/r.json`): exit=7, the review did not run, and the refusal names both flags. Correct today, but untested (m9).
  - **case only** (`R.json` and `r.json` in one directory, macOS default case-insensitive APFS): exit=0. The directory holds one entry, `R.json`, with the last write in it. The order does not name this spelling (see Remainder).
- Property 1, negative control: the schema test's own control (version removed, format_schema_test.go:60-79) is sound by reading. On this seat it is UNRUN; see Unrun.

### Closure
Order property by property:
1. **SARIF valid 2.1.0 for every shape.** The edge shapes (no evidence, unresolved, failed service, empty path, line 0, end before start, unknown severity, empty result) are tabled once in format_sarif_test.go:117-218 and used by both TestSARIFShapes and TestSARIFShapesMatchTheSchema. The structural half passes and kills m4 and m7. The schema half is UNRUN here, so this property is **not verified by this reviewer**. By reading, the writer meets the schema requirements: `version`, `runs`, `tool.driver.name`, `message.text` (always non-empty, format_sarif.go:188-199), and `startLine` ≥ 1.
2. **Evidence and `unresolved` reach both formats; nothing from the text is lost.** Met. The evidence is the same `evidenceFrom` array in JSON and SARIF (format_sarif.go:137), and output_contract_test.go:228-232 compares them. Every text-only line (settings, author statement, reviewer statements, trace path, family sentence) is in the JSON and in the SARIF run properties, and m5 is killed. The EPSS lines the text prints are derived from query `columns`/`rows`, which both formats carry.
3. **One run, all outputs; write failure is exit 2; colliding outputs refused at load for every pair and spelling.** One run and exit 2: met (m8 is killed). Pairs and `..`: met (TestCollidingOutputPathsAreRefused covers all 6 pairs). **Symlinked directory: NOT MET**; it was reproduced above with exit 0 and a silent overwrite. config.go:438-440 and docs/OUTPUT_FORMATS.md:5 openly state that links are not followed, but the order names this spelling as required. **Relative and absolute: met by the code, unpinned by tests (m9).**
4. **JSON field names are the documented ones.** Met. TestJSONDocumentIsTheContract does an exact decode-and-compare plus `assertNamesDocumented`, and m3 is killed. Every top-level key in format_json.go:35-50 appears in docs/OUTPUT_FORMATS.md:13-37.
5. **Each claim has a test that fails on revert.** Met for m1-m8 and m10. Not met for the relative-and-absolute spelling (m9) or for symlinked directories (no test, because no code).
6. **ADR-0027.** Met by reading. Every new type and function carries a WHY comment (e.g. config.go:434-439, format_sarif.go:144-146, 159-160, 168-171, 201-206; format_json.go:30-33, 91-93), the decisions are named (ADR-0020 §1, ADR-0013, RFC 3986 §4.2), and test names state their property. The new names are unexported apart from the existing `Formatter`/`FormatterFor`, whose comments were updated.

R4, dependency delta: one module was added, `github.com/santhosh-tekuri/jsonschema/v5 v5.3.1` (go.mod:5, go.sum:1-2). It has no transitive requirements and no install-time scripts, and it is imported only by `██████████████████████████████████`. Side effect, measured: on a seat without that module cached and without network, the **whole** `internal/cli` package fails `[setup failed]`, not just the schema test.

### Remainder
- **Leftovers in the worktree (the class prevented clean-up).** `review/` (README, mutant copies, overlay maps, probe_test.go) and `███████████████████████████████████████████████████████████` are untracked files of mine. The bash class for role review has no `rm`, and its write-set is `review/`, not the order's `████████████████`. I could not delete them. The master removes both at teardown. None of them is under `internal/` or tracked.
- The order's H1 for `_REVIEW_STATUS.md` (`# REVIEW — json-sarif-outputs-review-claude`) conflicts with the engine-injected channel grammar (`# REVIEW — json-sarif-outputs`). I followed the engine, which states that the verdict binds to the reviewed stream's slug.
- Declined to judge (R11), for the master to rule on:
  - (a) Case-insensitive collisions (`R.json` and `r.json` on APFS/NTFS): exit 0 with a silent overwrite. The order does not name this spelling. The same fix as C1 (`os.SameFile` when both exist) only catches it when the file already exists.
  - (b) The end line is dropped from SARIF when it is before the start line (format_sarif.go:149). The text and JSON keep it. This is documented.
  - (c) If an `--*-output` write fails, `-o` has already been written (run.go:233-239). The comment at run.go:196-197 only claims protection against render failure, which is accurate.
  - (d) The sha256 prefix of the vendored schema (testdata/sarif/README.md:4) was not recomputed; there is no hash tool in the class.

### Profile
None of the named profiles applies, because the diff touches no authentication, payments or data access. The output-path handling is configuration and a write path (R9). It was judged for silent overwrite (the destructive outcome) under property 3, not as a security profile. The checks were ranked (testing-bar step 10): every check here is a **suite** check, plus one **external** check (the OASIS schema) that is UNRUN on this seat. The project declares no numeric bar, and no bar-weakening moves were found. The assertions changed in run_test.go:470-482 and 624-625 turn an expected refusal into an expected success with non-empty output, which is a stronger assertion, not a looser one.

### Focus
Order properties 1-6. The weight went to property 3, because the round-2 BLOCKING finding was a silent overwrite and the order lists the spellings the fix must cover. I also checked whether the conditions C1-C5 from the round-2 claude-code review were addressed (R14):
- C1 (text-only lines in SARIF and JSON): ADDRESSED, m5 killed.
- C2 (table-driven schema test over edge shapes): ADDRESSED in structure. Its schema half is UNRUN here.
- C3 (`level` per result id, unplug swap): ADDRESSED, m2 killed.
- C4 (decode-and-compare, `[]` not null, rename and drop unplugs): ADDRESSED, m3, m4 and m6 killed.
- C5 (top-level keys documented): ADDRESSED, `assertNamesDocumented`.
- codex BLOCKING (path collisions): ADDRESSED for lexical spellings, **NOT ADDRESSED for a symlinked directory**, which the order names.

### Conditions
- **C1 · BLOCKING · internal/cli/config.go:441-463.** Two outputs that reach one file through a symlinked directory are not refused. Measured: exit 0, and the SARIF body replaced the JSON body in `real/r.json`. Remedy: resolve each path's existing parent directory with `filepath.EvalSymlinks` before comparing. Additionally, where both targets exist, compare them with `os.Stat` + `os.SameFile`, which also covers a symlink to a file and case-only spellings on case-insensitive volumes. Update docs/OUTPUT_FORMATS.md:5 and the comment at config.go:434-439, which currently document the gap. Test: a row with `link -> real` that must exit 7; unplug it by reverting the resolution, which turns the row red.
- **C2 · OWED · internal/cli/outputs_test.go.** Add a relative-vs-absolute row (`t.Chdir(dir)`, `-o r.json`, `--sarif-output <dir>/r.json` must exit 7). Unplug: m9 (`filepath.Abs` replaced by `filepath.Clean`) must turn it red; today m9 survives.
- **C3 · OWED · go.mod:5 / █████████████████████████████████████.** The test-only module makes the whole `internal/cli` package `[setup failed]` on a seat without it cached. Remedy (pick one): move the schema test behind a build tag or into its own package, or record that the module must be in the provisioned cache. This is a judgement finding (R6) and does not block alone.

### Unrun
- `TestSampleSARIFMatchesTheSchema` and `TestSARIFShapesMatchTheSchema` (██████████████████████████████████). The module is not cached, and the sandbox refuses the fetch: `Get "https://proxy.golang.org/github.com/santhosh-tekuri/jsonschema/v5/@v/v5.3.1.zip": tls: failed to verify certificate: x509: OSStatus -26276`. Not bypassed (law 57). Property 1's schema half and its version-removed negative control are therefore not verified by this reviewer.
- Network-binding tests outside W2: cmd/livecheck TestAllServicesReady, generator TestLiveReadsTheProvisionedStream, guardrail TestServerPostsTheMeasuredRequest, pipeline TestUnsafePromptStopsTheStepAndNamesCategories, query TestLiveSelectReadsRowsFromTheSQLEditor, review TestLiveReviewCarriesTheRoutedModel. Each fails with `bind: operation not permitted`. On this seat the master's "36 packages ok" is not re-measurable; 29 packages printed `ok` in the full run, plus `internal/cli` `ok` under the overlay.
- The real `proven` binary was not built and run for the exit-7 check. `Run` was exercised in-process instead, through the same `resolve` path.
- The sha256 of the vendored schema was not recomputed; no hash program is in the class.

## The review status as filed

## REVIEW — json-sarif-outputs
- state: recommendation-ready
- verdict: reject
- ref: _REVIEW_NOTES.md
- abstract: f6b6b1f fails order property 3: -o and --sarif-output reaching one file through a symlinked directory exit 0, and the SARIF overwrites the JSON (reproduced; config.go:441-463 compares lexically only).
  Mutants m1-m8 and m10 are killed. m9 (Abs→Clean) survives, so the relative-vs-absolute spelling is untested.
  JSON/SARIF field contract, evidence, unresolved and the text-only lines are met. The schema tests are UNRUN: the jsonschema module is uncached and the sandbox blocks the fetch.
  C1 BLOCKING: resolve symlinked dirs (EvalSymlinks/SameFile) and add a test. C2 owed: relative/absolute row. C3 owed: schema dep breaks the whole cli package offline.
- artifacts:
    - _REVIEW_NOTES.md
    - _REVIEW_STATUS.md
- updated: 2026-10-24T00:00:00Z

## REDACTED

| what | times |
|---|---|
| absolute or private path | 2 |
| e-mail or user name | 1 |
| file and line in Saphan Studio's private code | 5 |
