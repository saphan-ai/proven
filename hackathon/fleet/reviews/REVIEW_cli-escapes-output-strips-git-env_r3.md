# REVIEW — cli-escapes-output-strips-git-env · round 3

- verdict: accept-with-conditions
- reviewer: claude-code on macm5, claude-opus-5-5
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-24T03:48:47Z

## Summary

The diff holds: every text-output field is escaped (C0, DEL, C1, invalid UTF-8), the newline is kept only behind "    | ", and the merge-base git child gets childenv.Minimal(). 12 escaping mutants and the env unplug each turn a named test red.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — cli-escapes-output-strips-git-env

### Ground
- Reviewer: claude-code (Opus 5.5), the Anthropic-family review of two parallel reviews. Profile: security (R9: output encoding and subprocess environment).
- `git rev-parse --short=7 HEAD` printed `40a9570`. Scope: `git diff 0b954a9...40a9570`, which touches 4 files: `internal/cli/format.go`, `internal/cli/run.go`, `internal/cli/terminal_escape_test.go`, `internal/cli/mergebase_env_test.go`.
- I set the executor's and the master's self-assessment aside (R1) and re-measured it at the tip 40a9570. I did not measure the master's merge with main 8770336.
- `go vet ./...` is clean. `go test -count=1 ./...` at 40a9570: `internal/cli` ok, `internal/childenv` ok, and 6 packages FAIL (`cmd/livecheck`, `internal/generator`, `internal/guardrail`, `internal/pipeline`, `internal/query`, `internal/review`). Every one of the 6 failures is `httptest: failed to listen on a port: listen tcp6 [::1]:0: bind: operation not permitted`, a sandbox refusal of this leg. None of those packages is in the diff. So the master's "39 ok, 0 FAIL" is neither confirmed nor refuted here.
- Named tests, run with `-v`: `TestGitMergeBaseChildCarriesNoKeyFromProven`, `TestCIMergeBaseIsTheForkPoint` (real git, 9.9s), `TestTextOutputShowsEveryPrintedFieldWithoutControlBytes`, `TestMultiLineFieldCannotForgeProvenLine` and `TestControlContractBytes` all PASS.
- No vendor co-author or AI trailer appears on 025fb8c, 1e77598 or 40a9570; all three are authored by Marcin Marzec. `git diff --stat 0b954a9...40a9570 -- AGENTS.md internal/review` is empty.
- Mismatch between the order and the engine: the order asks for the H1 `# REVIEW — cli-escapes-output-strips-git-env-review-claude`, but the engine-injected channel grammar requires `# REVIEW — cli-escapes-output-strips-git-env` and says to trust it. I followed the engine, so the verdict binds to the reviewed stream (law #29).

### Differential
Item by item, each with FILE·LINE·CONTROL:
1. **Every text-output field is escaped. HOLDS for the text formatter.** I read every `Fprint*` in `format.go:55-228`:
   - Through `visible`: class, rule, the settings, author-model and reviewer statements, failed service and reason, the family sentence, the trace path, finding name/severity/title, EPSS phrases, location path, message, rule id, evidence test path and name, command, query engine and snapshot, SQL, columns, rows.
   - Through `visibleLines`: the test output.
   - Printed without escaping, which is correct: only literals ("failed services: none", "  unresolved", "review partial: …") and integers (line numbers, `%t` booleans).
   - No field is printed raw. Control: `TestTextOutputShowsEveryPrintedFieldWithoutControlBytes` carries 23 named fields, the family sentence and the trace path.
   - **Outside the formatter, the property does NOT hold for stderr.** See Conditions C1.
2. **The escaping is total over its claimed range. HOLDS.** `format.go:260-281`:
   - `utf8.DecodeRuneInString`, plus the `RuneError && size==1` branch, prints each malformed byte as `\xNN` of the original byte.
   - `isControl` covers U+0000–001F, U+007F and U+0080–009F.
   - A valid U+FFFD passes unchanged, because size 3 does not take the error branch.
   - Only `visibleLines` keeps the newline, and it appends `"    | "` after it. `strings.TrimSpace` runs first, which removes leading and trailing whitespace controls (removal is allowed by P4-F5).
   - The only multi-line field is the test output (`format.go:204`). Every other field escapes `\n`, so no single-line field can start a new line.
   - Forgery: a forged line can only appear behind `"    | "`, and nothing Proven writes begins with four spaces and a bar. Control: `TestMultiLineFieldCannotForgeProvenLine`.
3. **Subprocess environment. HOLDS for the merge-base child.** `run.go:333` sets `cmd.Env = childenv.Minimal()`, so the child gets PATH and HOME only (`childenv.go:12`). Git still works: `TestCIMergeBaseIsTheForkPoint` is green against real git.
   - The only other git spawner on the CLI's path, `change.gitCommand` (`between.go:19-23`), already uses `childenv.Minimal()`.
   - The container runtime (`testrun/container.go:553`) uses `childenv.Minimal(runtimeEnvNames...)`.
   - `internal/cli` has no other `exec.Command` outside tests.
4. **JSON and SARIF byte-identical. HOLDS, but vacuously.** No JSON or SARIF writer exists in this build: `FormatterFor` refuses both (`format.go:27-28`), and the diff touches neither path. The doc comment's claim about those encoders is wrong; see Remainder R2.
5. **Tests have different false-world answers. HOLDS.** See Unplug: every mutant I ran turned a named test red.
6. **ADR-0027. HOLDS.** There is a WHY comment at the escaping (`format.go:230-258`, which writes the byte contract once) and at the env line (`run.go:330-332`, "Property: no key reaches a subprocess …"). Every new test's comment names its property (ADR-0027 rule 5).

### Unplug
I mutated without editing the checkout, using `go test -overlay` with a faithful copy of the escaping (`████████████████████████████████████████████████████████████████████████████████,helpers.go,mutant.go,overlay.json}`), one mutant per run via `-ldflags -X …cli.mutant=<name>`.
- Instrument control: the faithful copy (`control`) is green, so the instrument itself does not redden.

| mutant | what is unplugged | tests red |
|---|---|---|
| identity | `visible` returns its input | TextOutput…, ControlContractBytes |
| sentence-raw | family sentence printed raw | TextOutput… (names `FamilySentence`) |
| testname-raw | evidence test name raw | TextOutput… (names `TestName`) |
| row-raw | query row raw | TextOutput… (names `QueryRow`) |
| output-raw | test output raw | TextOutput… + MultiLineField… |
| output-visible | output newlines escaped (no lines kept) | MultiLineField… |
| no-prefix | newline kept without `"    | "` | MultiLineField…, ControlContractBytes |
| no-invalid | malformed-byte branch removed | TextOutput…, ControlContractBytes |
| no-c1 | C1 range not escaped | TextOutput…, ControlContractBytes |
| no-del | DEL not escaped | TextOutput…, ControlContractBytes |
| tab-kept | tab passed raw | TextOutput…, ControlContractBytes |
| c1-edge | U+009F excluded | ControlContractBytes |

- P2-F3 unplug: an overlay of `run.go` without `cmd.Env = childenv.Minimal()` (`review/run_noenv.go`, `review/overlay_noenv.json`; marked `//go:build ignore` after the run) turns `TestGitMergeBaseChildCarriesNoKeyFromProven` red with "the git child inherited Proven's key: ████████████████████=marker-9f3a1c".

### Closure
- Within the diff, every claim of the executor's return is re-measured true at 40a9570:
  - C0, DEL, C1 and invalid UTF-8 are shown as `\xNN`.
  - The newline is kept only in the test output, behind a constant prefix.
  - Tab is escaped.
  - The merge-base child gets `childenv.Minimal()`.
  - Each property has a test that is red with its fix removed.
- The round-2 answers (continuation prefix, tab, original malformed byte) are each pinned by a test that a mutant reddens.
- P4-F5 as worded ("output to a terminal") is not yet closed, because of the stderr channel (C1).

### Remainder
- **R1 (owed, not blocking): characters outside the contract.** Bidirectional overrides (U+202A–202E, U+2066–2069), zero-width characters and U+2028/2029 pass unchanged. They cannot drive a terminal the way ESC or CSI can, but they can reorder or hide what a reader sees in a path or title. They are outside the C0/DEL/C1 contract the wagon claims; this is a candidate for a follow-on decision.
- **R2 (owed, not blocking): a false comment.** `format.go:247` says "JSON and SARIF need no such step: their encoders already escape". Measured (`review/json_probe_test.go`, `-tags reviewprobe`): `json.Marshal("a\x1b\x7f\u009bb")` yields `"a\u001b\x7f\u009bb"`, which escapes ESC but leaves DEL and C1 CSI (U+009B) raw. The day a JSON writer prints to a terminal, this comment would mislead its author. Remedy: correct the sentence to say encoding/json escapes C0 only, or require the future writers to escape DEL and C1.
- **R3 (owed, not blocking): Windows.** `childenv.Minimal` passes only PATH and HOME. On Windows, git usually needs `SYSTEMROOT`/`USERPROFILE`. Not measured, because I have no Windows host; whether Windows is a target is for the master to rule.
- **Declined to judge (R11):** licences (out of scope per the order); whether the `--verbose` echo of target/baseline/head (`run.go:99`) should be escaped (those values come from flags and the environment, not from the change).

### Profile
security. S1 trust boundaries:
- the change's bytes reaching the operator's terminal: stdout is closed by this diff; stderr is open (C1);
- Proven's environment reaching a git child: closed for merge-base by this diff and already closed in `change.gitCommand`.

S2: I reproduced the blocking-class finding with a failing probe. S3 does not apply (no authorization check in this diff). S4: no dependency delta (only the stdlib `unicode/utf8` is added); no secrets in the diff; no invisible or bidirectional characters in the diff (the test file carries escapes as Go literals).

### Focus
Two properties were in focus: P4-F5 (no control byte from the change reaches the terminal) and P2-F3 (no key reaches the merge-base git child). I spent most of the effort on the mutation matrix for the escaping, and then on sweeping every terminal write the command line makes, not only the formatter.

### Conditions
- **C1 (condition; the class of property P4-F5): stderr prints errors that carry the change's bytes raw.**
  - Location: `internal/cli/run.go:92` (and `:38`, `:50`, `:56`, `:65`, `:68`, `:89`, `:102`, the same pattern) prints `err.Error()` unescaped.
  - Trigger: the errors carry file names from the change through `%s`. `internal/change/read.go:78` (`change: symlink %s`), `internal/change/archive.go:145` and `:224` (`change: unsupported file %s`), and `OverBound("the file "+clean, …)` (`archive.go:121`, `:228` → `limits.go:168`).
  - Evidence (`review/stderr_probe_test.go`, run with `go test -tags reviewprobe ./review/`):
    - (a) `cli.Run(["scan","--sample",<fixture>])` on a fixture whose `change/` holds a symlink named `x\x1b[2Jy` exits 2 with stderr `"change: symlink …/change/x\x1b[2Jy\n"`, where the raw ESC is in the buffer.
    - (b) On the git path, `change.ExportCommit` of a commit holding `big\x1b[2J.txt` over a 4-byte file bound returns `change: the file big\x1b[2J.txt is larger than the 4-byte bound …` with the raw ESC. That is the error the pipeline returns (`pipeline/trees.go:106`) and `run.go:92` prints.
  - Remedy: pass the stderr sentences through `visible` at the print sites in `run.go`, which keeps one encoder, or quote paths with `%q` at the error sites. Add a test that drives `Run` into one of these errors and asserts no raw control byte on stderr.
  - The master rules whether this lands in this wagon or as a named follow-on. Either way, P4-F5 ("output to a terminal") is not closed until it lands.
- Owed items R1–R3 (Remainder) do not block.

### Unrun
- The full suite is not green here. Six packages outside the diff (`cmd/livecheck`, `internal/generator`, `internal/guardrail`, `internal/pipeline`, `internal/query`, `internal/review`) fail on `listen tcp6 [::1]:0: bind: operation not permitted`, a sandbox refusal, so the master's "39 ok, 0 FAIL" stays the master's measurement.
- The master's merge with main 8770336 was not re-measured. Everything here was measured at 40a9570 alone.
- The end-to-end git-path stderr (C1 b) was measured only up to the error string returned by `change.ExportCommit`. A full `Run` in non-sample mode needs a live reviewer, which this sandbox cannot run.
- Windows child environment (R3): no host available.

## The review status as filed

## REVIEW — cli-escapes-output-strips-git-env
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: The diff holds: every text-output field is escaped (C0, DEL, C1, invalid UTF-8), the newline is kept only behind "    | ", and the merge-base git child gets childenv.Minimal(). 12 escaping mutants and the env unplug each turn a named test red.
  C1: stderr still prints error sentences carrying the change's file names raw (run.go:92 ← read.go:78, archive.go:121/145). Measured with a failing probe; P4-F5 as worded ("output to a terminal") stays open until those prints are escaped or a follow-on stream is named.
  Owed, not blocking: format.go:247 claims JSON encoders escape, but encoding/json leaves DEL and C1 raw; bidi characters are outside the contract; the Windows child env is unmeasured.
- artifacts:
    - _REVIEW_NOTES.md
    - _REVIEW_STATUS.md
- updated: 2026-10-24T04:10:00Z

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| private credential or variable name | 1 |
