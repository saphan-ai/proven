# REVIEW — json-sarif-outputs · round 9

- verdict: accept-with-conditions
- reviewer: claude-code on macm5, claude-opus-5-5
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-24T08:30:46Z

## Summary

W2 at e289682 holds properties 2–6 and every listed alias class of property 3 (internal/cli ok; 12/12 mutants killed, U1–U12).

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — json-sarif-outputs

Reviewer: claude-code (Anthropic family), stream json-sarif-outputs-review-r9, 2026-10-24. Scope: `git diff main...e289682`
(merge-base ce61e44), 21 files, +5909/−18. One of two independent reviews. I have not seen the other one.

### Ground

- `git rev-parse --short=7 HEAD` printed `e289682` (ROLE-LOCK satisfied). `git merge-base main HEAD` = `ce61e444ead43c21a846f2123853d03881316a13`; `main` itself has moved to `fc44de2`, so the three-dot diff is measured against ce61e44.
- Read in full: `internal/cli/config.go` (the W2 hunk, lines 391–650), `run.go` (the W2 hunks plus `Run` and `writeOutputs`), `format.go`, `format_json.go`, `format_sarif.go`, `docs/OUTPUT_FORMATS.md`, `outputs_test.go`, `outputs_alias_test.go`, `output_contract_test.go`, `format_sarif_test.go` (lines 75–335), `sarifschema/schema_test.go`. I also read `review.Result`/`Finding` in `internal/review/review.go:99-148` (the input) and the text writer (the comparison baseline for property 2).
- `go vet` over the 41 packages other than `internal/cli/sarifschema`: clean.
- `go test -count=1` over those 41: 34 ok, 7 FAIL. All 7 FAILs are `httptest: failed to listen on a port: listen tcp6 [::1]:0: bind: operation not permitted`: livecheck `TestAllServicesReady`, generator `TestLiveReadsTheProvisionedStream`, guardrail `TestServerPostsTheMeasuredRequest`, pipeline `TestUnsafePromptStopsTheStepAndNamesCategories`, query `TestLiveSelectReadsRowsFromTheSQLEditor`, review `TestLiveReviewCarriesTheRoutedModel`. I re-ran each one alone and the panic was the same. None of those packages is in the diff. This is the seat's sandbox, not a regression. **`internal/cli`: ok (35.3 s).**
- `internal/cli/sarifschema` did not build here: `github.com/santhosh-tekuri/jsonschema/█████████: Get "https://proxy.golang.org/...": tls: failed to verify certificate: x509: OSStatus -26276`. The module is not in the seat's cache (only a `.lock` exists). See Unrun.
- `git diff --stat main...e289682 -- AGENTS.md internal/review` is empty. The commit bodies in `ce61e44..e289682` have no AI attribution (`grep -ciE "co-authored|generated with|claude|codex|openai|anthropic"` → 0).
- Dependency delta (R4): one new module, `github.com/santhosh-tekuri/jsonschema/v5 v5.3.1`, imported only by `internal/cli/sarifschema/schema_test.go`. go.sum has 2 lines. Go has no install-time scripts. The vendored `testdata/sarif/sarif-schema-2.1.0.json` names the OASIS errata01 `id` in its header. Its sha256 is UNRUN (see Unrun).

### Differential

Each order property, with the file, line and control.

1. **SARIF is valid 2.1.0 for every result shape.** `format_sarif.go:147` emits a region only when `startLine >= 1`, and `endLine` only when it is ≥ startLine. A result with no path gets no location (`:140`). The message is never empty (`sarifMessageText`). `results` and `failed_services` are `[]`, never null. Structure is proven by `TestSARIFShapes` (8 shapes: no evidence, unresolved, failed service, empty path, line 0, end before start, unknown severity, empty result). Schema validity is claimed by `sarifschema.TestSARIFShapesMatchTheSchema` over the same 8 shapes, and by `TestSampleSARIFMatchesTheSchema`, whose negative control deletes `version` and requires the rejection to name it. **The schema half is UNRUN on this seat.** It holds by code reading. Not measured here.
2. **Evidence and `unresolved` reach both formats, and no text line is lost.** JSON: `findings[].evidence[]` and `findings[].unresolved` (`format_json.go:73-84`). SARIF: `properties.evidence` (the same `evidenceFrom` objects) and `properties.unresolved` (`format_sarif.go:136-137`). `TestSARIFPropertiesAreTheContract` asserts that the SARIF evidence equals the JSON evidence. I compared every text line in `textFormatter.Write`/`writeFinding`/`writeEvidence` against both outputs. Class, rule, settings, author-model and reviewer statements, failed services, partial, family sentence, trace, and each finding's id, severity, title, unresolved, location, message, rule and evidence fields are present in both. Holds. The one derived text line, the EPSS phrases from `epss.Phrases(columns, rows)`, is not carried literally. Its inputs (columns and rows) are carried in both formats, so the information survives and only the rendering is lost. See Remainder.
3. **One result to every output; a write failure is exit 2; collisions are refused at load with exit 7.** `writeOutputs` renders every body from the one filtered `result` before writing anything (`run.go:204-235`). A write error is exit 2 (`run.go:106-108`, guarded by `TestOutputWriteFailureExits2` for all four flags). The load check is `refuseCollidingOutputs` (`config.go:448`), called from `resolve` at `:391`. It refuses every pair of the four flags when they share an identity from `outputIdentity` (`:568`, a component-by-component kernel walk) or when `os.SameFile` is true for existing files (`:478`). Standard output is compared in `refuseStdoutOnAnOutput` (`run.go:54`). A second, write-time check sits at `run.go:250-262`. Exercised: every pair through `sub/..` (`TestCollidingOutputPathsAreRefused`, 6 pairs); a symlinked directory; a symlinked directory followed by `..`; chained links; a dangling final link for all 12 ordered pairs; link→link; a hard link; a trailing slash; relative against absolute; a link loop or more than 40 hops (exit 7); `missing/../x` (exit 2 when absent, exit 7 when present). **Holds for every case the order lists.** Two residuals are outside the listed cases: letter case on a case-insensitive volume (caught at write, exit 2, first output intact; measured, see Remainder) and an output that names standard output's own non-regular file (condition C2).
4. **The JSON field names match `docs/OUTPUT_FORMATS.md`.** `TestJSONDocumentIsTheContract` decodes the whole document and compares it exactly with a literal. `assertNamesDocumented` fails on any key the doc does not name. I checked the doc table by hand against the `jsonDocument` struct tags: 16 top-level keys and every nested key match. Holds.
5. **Each claim has a test that goes red when reverted.** Measured with 12 mutants. All 12 were killed (see Unplug).
6. **ADR-0027.** Every new function and type has a doc comment, including the unexported ones. WHY comments sit at each choice (`format_json.go:19-21` on HTML escaping, `format_sarif.go:131-133` on evidence under properties instead of codeFlows, `:144-146` on region bounds, `config.go:540-566` THE RULE / WHY, `run.go:244-248` on the second check). The decision is named (ADR-0020 §1, ADR-0013). Every test comment names its property and its mutated field. Holds.

Claims in the order's own record that this review refutes (law #44):
- The order says "e289682 is the master's absorb of main ce61e44". **Refuted:** e289682 is `cli: a missing directory blocks parent dots`. The absorb is `4313347 [json-sarif-outputs] merge: absorb main ce61e44, keep both docs rows`.
- The order says "Round 4 (5fa7d24, e289682)". **Refuted:** `git merge-base --is-ancestor 5fa7d24 e289682` is false. 5fa7d24 is contained only by `review/json-sarif-outputs-review-r6`. On the reviewed branch the commit with that subject (`cli: refuse outputs that are one file via a link`) is `c2564b1`. The code under review does not change, but the order's commit map is stale.
- I re-measured the master's "38 packages ok, 0 FAIL" on this seat: 34 ok and 7 FAIL, all of them socket-bind denials in the sandbox, and sarifschema did not build. That is not a contradiction of the master's seat, only a reading that this seat cannot reproduce.

### Unplug

Each mutant replaced one source file through `go test -count=1 -overlay <map> ./internal/cli`. The checkout's files were never edited. Red tests are listed per mutant.

| # | mutation (file:line) | red |
|---|---|---|
| U1 | `refuseCollidingOutputs` returns nil (config.go:448) | 11 tests: AliasTable×2, MissingDirectory, CollidingOutputPaths, SymlinkedDirectory, SymlinkedFileAndHardLink, RelativeAndAbsolute, CollisionReproduction, DanglingAlias, DanglingAliasChains, LinkLoop |
| U2 | symlinks treated as plain components (config.go:609) | AliasTable×2 (6 link rows each), OutputsThroughASymlinkedDirectory, DanglingAlias (12/12), DanglingAliasChains, LinkLoop (2/2) |
| U3 | `..` after a missing component collapsed lexically (config.go:591) | TestMissingDirectoryThenDotDotIsUnwritable/absent |
| U4 | load-time `os.SameFile` half off (config.go:478) | AliasTable×2/hard_link, TestSymlinkedFileAndHardLinkAreRefused/hard_link |
| U5 | `refuseStdoutOnAnOutput` always nil (config.go:510) | TestStdoutRedirectedOntoAnOutputIsRefused (8/8) |
| U6 | write-time `os.SameFile` guard off (run.go:260) | TestAliasCreatedAfterLoadIsRefusedAtWrite, TestStdoutAliasCreatedAfterLoadIsRefusedAtWrite |
| U7 | WARNING↔ERROR swapped in `sarifLevel` (format_sarif.go:174) | TestSARIFLevelFollowsSeverity |
| U8 | SARIF `properties.evidence` dropped (format_sarif.go:137) | TestSARIFShapes/no_evidence, TestSARIFPropertiesAreTheContract |
| U9 | SARIF `properties.unresolved` dropped (format_sarif.go:136) | TestSARIFShapes/unresolved, TestSARIFPropertiesAreTheContract |
| U10 | JSON tag `unresolved`→`unresolve` (format_json.go:79) | TestJSONWriterEmitsFindings, TestJSONDocumentIsTheContract |
| U11 | region admits startLine 0 (format_sarif.go:147) | TestSARIFShapes/line_0 (the schema twin is UNRUN here) |
| U12 | a write failure exits 0 (run.go:108) | TestOutputWriteFailureExits2 (4/4), MissingDirectory/absent, both AliasCreatedAfterLoad tests |

Surviving mutants: none. The review's round-2 conditions are therefore guarded as claimed: C3 by U7, C4 by U8 and U10, C1 by `TestSARIFCarriesEveryTextOnlyLine`/`TestJSONCarriesEveryTextOnlyLine` (one subtest per key), and the codex BLOCKING by U1 and `TestCollisionReproduction`.

### Closure

The round-3 and round-4 answers, checked against the fix diff (R14):
- codex BLOCKING (`--json -o X --sarif-output X` exits 0): **ADDRESSED.** `TestCollisionReproduction` returns exit 7 and X is not written. U1 turns it red.
- Both reviews' BLOCKING on e289682's predecessor (two outputs through a symlinked directory exit 0): **ADDRESSED.** `TestOutputsThroughASymlinkedDirectoryAreRefused` and the `aliasTable` rows `symlinked directory`, `symlinked directory then ..`, `chain of two directory links` and `trailing slash`. U2 turns them red.
- claude-code C1 (text-only lines in SARIF run properties and in JSON): **ADDRESSED** (`format_sarif.go` run properties, `TestSARIFCarriesEveryTextOnlyLine`).
- C2 (table-driven schema test over the edge shapes): **ADDRESSED in code**, with the run UNRUN on this seat (`sarifschema/schema_test.go:93-136`).
- C3 (`level` per result id, unplug by swapping): **ADDRESSED** (U7).
- C4 (decode-and-compare test over every field, `[]` not null): **ADDRESSED** (`TestJSONDocumentIsTheContract`, `TestJSONEmptyListsAreArraysNotNull`; U8, U10).
- C5 (every top-level JSON key in the doc): **ADDRESSED** (`assertNamesDocumented`, and checked by hand).
- Schema tests moved to `internal/cli/sarifschema`: **ADDRESSED** (`cddf918`). `internal/cli` no longer imports the schema module.

### Remainder

Measured by reviewer probe tests, applied by overlay as `████████████████████████████████████`, which is not in the checkout:
- **Letter case on this seat's case-insensitive temp volume.** `-o R.json --sarif-output r.json` → exit 2, `-o and --sarif-output are the same file (r.json); -o is left as it was written`, and R.json still holds the JSON body. `-o Real/r.json --sarif-output real/r.json` with `real/` existing → also exit 2 at write, not 7 at load. No silent overwrite in either case. The write-time guard does its job. The doc says such spellings "are not caught", which understates what happens (see C3).
- `PROVEN_OUTPUT=r.json` with `--sarif-output r.json` → exit 7, but the refusal names `-o` even though no `-o` was typed. Wording only.
- Standard output redirected to a regular file together with `-o` on that same file (`proven scan --json -o out.json > out.json`) → exit 7, "same file as standard output". With `-o` set, nothing is written to standard output, so this refusal is conservative rather than necessary. Judgement only. It loses no data.
- `--sarif-output /dev/fd/N` aimed at a pipe used as standard output: the probe was blocked by the sandbox (`open /dev/fd/5: operation not permitted`), so this is UNMEASURED. By code reading it is condition C2.

Declined to judge (R11), for the master to rule on:
- `review.Result.SkippedPaths` is printed by neither the text writer nor JSON nor SARIF. It is outside "nothing from the text output is lost", because text does not print it either.
- The EPSS phrases are derived in text and not carried literally in JSON or SARIF (their inputs are carried).
- The shape table is duplicated in `format_sarif_test.go:117` and `sarifschema/schema_test.go:93`, with a comment asking that both be kept in sync. Drift risk, owed as structure (R6).
- SARIF omits the per-finding `models`, `doubled`, `author_models`, `default_reviewer`, `special_reviewer`, `reviewer_models` and run `models`. The text output prints none of them as separate lines, so property 2 holds. A SARIF-only consumer gets less than a JSON consumer, which the doc states implicitly by listing the SARIF properties.

### Profile

Backend/CLI with a filesystem-write path. Under R9 (a destructive path on derived paths), the write path carries part of the security profile. S4 "allowed root after link resolution" does not apply: outputs are user-named and may legitimately be anywhere. The property judged is that no output may silently replace another, and that holds for every listed alias class (Unplug U1–U6). Testing-bar: the project declares no numeric bar, so no number was invented. Check ranking: the SARIF schema check is **external** (OASIS schema) and UNRUN here. Everything else is **suite** checks, strengthened by the 12/12 mutant kill measured above. I found no bar-weakening moves in the diff: no skip markers, no loosened assertions, and the rewritten `run_test.go` assertions are stricter (exit 2 → exit 0 plus a non-empty body with named keys).

### Focus

1. Whether `outputIdentity` really follows the kernel's order (link before `..`). Verified by reading and by U2/U3.
2. Whether every pair of outputs, standard output included, is compared. Verified by U1, U4, U5 and U6.
3. Whether SARIF stays schema-shaped on edge inputs. Verified structurally (U8, U9, U11). The schema run is UNRUN.
4. Whether JSON and SARIF have parity with the text output. Verified by reading against `textFormatter`.

### Conditions

Verdict `accept-with-conditions`. **None of these is blocking.** C1 is owed before the gate because this seat could not measure it (R7). C2 and C3 are owed work.

- **C1 (owed before gate, measurement).** This seat could not build `internal/cli/sarifschema`: the module fetch failed with a TLS error. The master re-runs `go test -count=1 ./internal/cli/sarifschema` on a seat with module access and pastes the result. A stronger control: apply U11 (`format_sarif.go:147`, `>= 1` → `>= 0`) and show `TestSARIFShapesMatchTheSchema/line_0` go red. The master reports these tests PASS on e289682. This review neither confirms nor contradicts that.
- **C2 (owed, PLAUSIBLE, not reproduced: the sandbox denied `/dev/fd`).** An output path that is standard output's own non-regular file is not refused. Trigger: `proven scan --json --sarif-output /dev/stdout | consumer`. `refuseStdoutOnAnOutput` returns nil when standard output is not a regular file (`config.go:510`, via `regularFileInfo` `:496`), and the write-time guard records standard output only when it is regular (`run.go:253`). Both bodies are then written to the one pipe and the run exits 0. This is the same silent-mixing class the round-4 BLOCKING covered, reached through a device path. Remedy: when `cfg.output == ""`, compare each output's `os.Stat` with standard output's `Stat` for any `*os.File` (pipe and character device included) and refuse with exit 7. Test with `os.Pipe` and `/dev/fd/<w.Fd()>`, skipping where `/dev/fd` cannot be opened.
- **C3 (owed, docs).** `docs/OUTPUT_FORMATS.md:5` says paths are compared "after they are made absolute and cleaned and after every symbolic link ... is resolved". That describes clean-then-resolve, the order `config.go:540-566` (THE RULE) explicitly rejects because it misnames `link/..`. Remedy: state the kernel-walk rule, meaning links are resolved as they are met and before `..`, and a `..` after a missing directory names no file and fails at write (exit 2). Also replace "are not caught" for letter-case spellings with what was measured: caught when the second file is written, exit 2, and the first output is left intact.

### Unrun

- `go test ./internal/cli/sarifschema` (`TestSampleSARIFMatchesTheSchema`, `TestSARIFShapesMatchTheSchema`): the package does not build on this seat. `go: downloading github.com/santhosh-tekuri/jsonschema/v5 v5.3.1 ... tls: failed to verify certificate: x509: OSStatus -26276`. Setting `GOPROXY`/`GOFLAGS` for one command was refused by the role's permission class (`"GOPROXY=off" is NOT listed in permission class ██████████████████████████`). Owed as C1.
- `go vet ./internal/cli/sarifschema`: same cause.
- The `/dev/fd` probe for C2: `open /dev/fd/5: operation not permitted` (sandbox).
- The vendored schema's sha256 against the README's prefix `c3b4bb2d60938974`, and against upstream: this class has no hashing program and the seat has no network.
- The built `proven` binary (the master's exit-7 claim): running a binary is outside this class's program list. The same path is exercised in-process through `cli.Run` by `TestCollisionReproduction`, which passes.
- The six network-binding tests in packages outside the diff (listed in Ground): sandbox bind denial.
- Seat hygiene, reported rather than fixed. This leg created `review/mut/` (scratch: overlay maps, plus mutant and probe files that were emptied to `//go:build ignore` stubs after use; `go list -e ./...` still lists the same 42 packages) and `█████████████████████████████████████████████████████` (written before the role's write-set `[_REVIEW_STATUS.md _REVIEW_NOTES.md review/]` was known). The class has no `rm`, so `git status --porcelain` shows `?? ████████` and `?? review/` from this leg. Owed to the master at teardown: delete both directories.
- The order's H1 for `_REVIEW_STATUS.md` (`# REVIEW — json-sarif-outputs-review-claude`) contradicts the engine's channel grammar (`# REVIEW — json-sarif-outputs`, which binds the verdict to the reviewed stream). I followed the engine.

## The review status as filed

## REVIEW — json-sarif-outputs
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: W2 at e289682 holds properties 2–6 and every listed alias class of property 3 (internal/cli ok; 12/12 mutants killed, U1–U12).
  No blocking finding. C1 owed before gate: the sarifschema tests are UNRUN here (module fetch TLS failure), so the master re-runs them.
  C2 owed (PLAUSIBLE): --sarif-output /dev/stdout with stdout on a pipe mixes both bodies and exits 0 (config.go:510, run.go:253).
  C3 owed: docs/OUTPUT_FORMATS.md:5 describes clean-then-resolve and says case variants are "not caught" (measured: exit 2 at write).
  Scratch review/ and ████████ are left untracked (this role has no rm); the master deletes them at teardown.
- artifacts:
    - _REVIEW_NOTES.md
    - _REVIEW_STATUS.md
- updated: 2026-10-24T09:10:00Z

## REDACTED

| what | times |
|---|---|
| absolute or private path | 3 |
| e-mail or user name | 1 |
| file and line in Saphan Studio's private code | 1 |
| policy binding name | 1 |
