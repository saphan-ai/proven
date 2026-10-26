# REVIEW — the-prover-runs-only-its-test · round 1

- verdict: accept-with-conditions
- reviewer: claude-code on macm5, claude-opus-5-5
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-23T12:29:28Z

## Summary

Reviewed on claude-opus-5-5 at 29e566f: no model text or change content shapes the container argv or command; proven by type removal plus TestArgv at Container.Run, unplug controls red as claimed.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — the-prover-runs-only-its-test

Reviewer: claude-code, model `claude-opus-5-5` (Opus 5.5), one-shot headless, 2026-10-23. Reviewed `21af174..29e566f`
(one commit), HEAD verified `29e566f` before any act. Property under review (SECURITY_REVIEW.md piece 1, the W29 half):
nothing in the change, and nothing a model answers, can shape the container's arguments or the command it runs.
The owner's two-review rule is not met by this leg alone: the NVIDIA-through-Nightshift review is OWED (route down today).

### Ground

The property holds at `29e566f`. Where it is guaranteed, and the test that proves each part:

1. **No model text reaches the command.** `testrun.Request` no longer has a `Command` field
   (`internal/testrun/testrun.go:20-27`) and `candidate.Refutation.Command` is removed (`internal/candidate/candidate.go:72-80`),
   so a model's command can not travel to the runner as a type. The Prover refuses a whole answer that carries one
   (`internal/agent/prover/prover.go:85-88`); the Skeptic drops the proposed test (`internal/agent/skeptic/skeptic.go:69-72`).
   Tests: `TestAnswerThatCarriesACommandIsRefused` (written and named), `TestAnswerThatCarriesACommandProposesNoTest`,
   `TestValidNameReachesTheRunnerAsFileAndNameOnly` (exact `Request` equality, so nothing else rides along).
2. **Proven builds the command from a fixed template.** `TestArgv` (`internal/testrun/argv.go:61-70`) gives exactly
   `go test -run ^<name>$ ./<dir>` and only after `RefuseTest` (`argv.go:29-44`) accepts the name and path.
   `Container.Run` calls `TestArgv` itself (`internal/testrun/container.go:278-282`) for EVERY caller, named or written,
   so the guarantee does not depend on the Prover or the Skeptic having checked first. Tests:
   `TestArgvIsTheTemplateWithNoShell`, `TestRunnerRunsExactlyTheTemplate`, `TestRunnerStartsNoContainerForAHostileName`.
3. **No shell.** `sh -c` is gone; the words are appended as separate argv elements after the image
   (`container.go:237-239`) and run by `exec.CommandContext(argv[0], argv[1:]...)` (`container.go:421`).
4. **Defence in depth on the Skeptic's route:** `prover.Execute` checks `RefuseTest` before the runner
   (`prover.go:59-61`), and the container checks again. Test: `TestExecuteRefusesAnInvalidNameWithoutCallingTheRunner`.

Measured runs (this leg, 2026-10-23, this host):
- `go build ./... && go vet ./...` — clean.
- `go test -count=1 ./...` — 27 packages `ok`; 6 `FAIL` (`cmd/livecheck`, `internal/generator`, `internal/guardrail`,
  `internal/pipeline`, `internal/query`, `internal/review`), every one a panic `httptest: failed to listen on a port:
  listen tcp6 [::1]:0: bind: operation not permitted` — this leg's sandbox refuses a loopback listener. Not the code.
- Because a panic aborts the whole package, `internal/pipeline` and `internal/review` were re-run with only the
  listener tests skipped (26 names from `guardrail_test.go`, `live_generator_test.go`, `open_test.go`, `zetaris_test.go`):
  both `ok`. `internal/agent/prover`, `internal/agent/skeptic`, `internal/testrun` (the three packages the diff's
  behaviour lives in) are `ok` in the full run.

Original reviewers' tests (`████████████████████████`), run against this tip with `go test -overlay` (scratch files in
this run's artifacts directory, never in the worktree):

| Test (origin) | Adaptation | Result at 29e566f | In W29 scope |
|---|---|---|---|
| `TestReviewModelTextDoesNotBecomeTheContainerCommand` (claude D2) | `run.last.Command` no longer exists; asserts instead the runner is never called and the note is the command refusal | PASS | yes |
| `TestReviewRunnerRefusesUnboundedCommand` (codex T2) | free text moved into the only model fields left (name, path) | PASS | yes |
| `TestReviewProcessInsideTheContainerIsBounded` (claude D1) | `Argv(dir, []string{...})` signature only | FAIL: `--user`, `--pids-limit`, `--cap-drop ALL`, `--security-opt no-new-privileges`, `--read-only` all absent | no (hardening) |
| `TestReviewContainerBoundsProcessAndFilesystem` (codex T1) | same | FAIL, same five flags | no (hardening) |
| `TestReviewExportWorktreeStaysInsideTheTarget` (claude D3) | `ExportWorktree` now returns two values | FAIL: file read from outside the target through a linked directory (err=nil) | no |
| `TestReviewCheckoutThatIsALinkIsNotMounted` (claude D4a) | `Command` field dropped, `Path` added | FAIL: runtime handed a link as mount source | no |
| `TestReviewLinkedTreesAreNotPreparedTrees` (claude D4b) | verbatim | FAIL | no |

The four red rows are piece-1 findings this commit does not claim to fix; they are listed under Remainder.

### Differential

- **The master's claim "33 packages ok" (2026-10-23) is NOT reproduced on this host:** 27 ok + 6 FAIL. All six failures
  are the sandbox's loopback refusal, and the two of them that matter for this diff (`pipeline`, `review`) are green
  with the listener tests skipped. The master's figure is plausible on a host that allows a listener; I could not
  measure it. Not a code defect.
- **The leg's negative control holds:** unplugging the Prover's command check turns
  `TestAnswerThatCarriesACommandIsRefused` red (both subtests). Measured, see Unplug.
- **The leg's account says the command field is removed from the prompts and answers and the fixtures regenerated:**
  holds — both specs (`prover.go:27-28`, `skeptic.go:25-26`) and both fixtures' recorded prompts and responses changed;
  the sample review prints the same findings (Closure, item 4).
- **The code cites ADR-0025 for a rule ADR-0025 does not contain.** `argv.go:28`, `argv.go:58`, `prover.go:82`,
  `skeptic.go:66`, `container.go:237,276` and the commit body say the file-and-name rule and the fixed template are
  ADR-0025. ADR-0025's six points (no network, no secret, own checkout, settings, no runtime → no code) say nothing
  about who composes the command. ADR-0027 point 4 asks a comment to name the decision it carries out; these name one
  that does not hold the rule. See Conditions C1.

### Unplug

Each guard was removed by an overlay copy of its file (the tree was not edited) and the suite re-run:

| Guard unplugged | Turned red |
|---|---|
| Prover command check `prover.go:85` (`if false && ...`) | `TestAnswerThatCarriesACommandIsRefused/written`, `/named` (runner called; evidence produced with no note) |
| Skeptic command check `skeptic.go:69` | `TestAnswerThatCarriesACommandProposesNoTest` |
| `RefuseTest` returns "" and `refusePackage` returns "" (`argv.go`) | `TestRefuseTestKeepsNamesAndPathsInsideThePattern`, `TestRunnerStartsNoContainerForAHostileName`, `TestContainerRefusesAWrittenPathBeforeItRuns/path_with_a_pattern`, `TestNamedTestThatFailsTheChecksIsRefused`, `TestExecuteRefusesAnInvalidNameWithoutCallingTheRunner` |

remark: `internal/pipeline` stayed green with `RefuseTest` unplugged — no pipeline-level test drives a hostile name through to the container; the property is proven at the prover and testrun layers only.

### Closure

1. **The property** — holds; lines and tests under Ground.
2. **Name and path checks** — `namePattern` `^Test([A-Z0-9_][A-Za-z0-9_]*)?$` and `segmentPattern`
   `^[A-Za-z0-9_][A-Za-z0-9_.-]*$` plus the `..` substring ban. Probe `TestReviewHostileNamesAndPathsGetNoCommand`
   (32 inputs: leading dash in a directory and in the file name, `--`, `..` and `..b`, `./`, `.hidden`, absolute path,
   backslash, NUL, tab, newline (in path and name), trailing slash, leading/trailing space in the name, `$|.`, non-ASCII
   letter, NBSP, `-run`, `-test.v`, `TestX/sub`, full-width letter, `~`, `*`, `@v1`, `:`, `=`, `...`, empty name): every
   one refused, no argv. Probe `TestReviewAcceptedArgvHasNoRoomForMore`: every accepted input yields exactly five words,
   only `-run` begins with a dash, the package word always starts `./` or is `.`, and no word holds whitespace or a
   shell or regexp metacharacter beyond the template's own `^`/`$`. A **symlinked directory**: the trees the container
   mounts come from `git archive` with links skipped (`internal/change/archive.go:107-108`) or are refused by
   `change.Read` (`read.go:75-76`), and a written test's staging copy skips links (`written.go:141-150`); so `./<dir>`
   can not name a link inside a tree Proven built. The mount-source-is-a-link case (D4) is a separate, still-open
   finding (Remainder).
3. **Other paths to a runtime argument** — none found from model text or change content. Image and runtime:
   `PROVEN_TEST_*` environment only (`settings.go:55-66`); `proven.yaml` from the change carries no container key
   (`internal/review/config.go:281-389`). Working directory: the constant `checkoutDir`. Mount source:
   `fx.Root/base|change` or Proven's own temp copy (`pipeline.go:806-811`, `container.go:287-298`). Environment: no
   `-e`/`--env` (`container.go:212-240`). `--name`: pid + atomic counter (`container.go:433-435`). The Skeptic's refuting
   test: same `RefuseTest` + `TestArgv` path. Probe `TestReviewPrefixDoesNotDependOnTheRequest`: two requests with
   different names and paths give a byte-identical argv up to and including the image (name value normalised).
   Probe `TestReviewCommandKeyCaseVariantsAreRefused`: `Command`, `COMMAND`, `cOmMaNd` keys (encoding/json is
   case-insensitive) are all refused — PASS.
4. **Sample review** — from a scratch directory outside the repository (`go -C <scratch> run` through a `go.work`),
   `scan --sample testdata/fixture/discount` printed F2, F4, `osv:evil.example/leftpad`, F1, `failed services: none`:
   identical to the output recorded in `docs/SAMPLE_MODE.md` at main `21af174` except the one `command:` line, which
   is now the built template. `scan --sample testdata/fixture/refuted` printed no finding and its trace carries
   `dropped: refuted: test TestTotal passes on the change`, identical to main's recorded verdict line. I did not build
   main's own binary (no second checkout was writable here); the comparison is against main's recorded output.
5. **ADR-0027 and finding text** — every new or touched exported name has a doc comment starting with its name; every
   new test states its property; comments are plain English with no attribution or host paths. One defect: the
   ADR-0025 citations (Differential, C1). Finding text: the commit subject and body, comments and test names describe
   the behaviour and name no finding id or reviewer wording. The test strings `make proof`
   (`skeptic/command_test.go:22`) and a shell pipeline in `prover/command_test.go:15` are hostile-input fixtures; the
   first matches the private reviewer's probe value but discloses no finding by itself.
6. **Build, vet, test** — under Ground.

### Remainder

- **Out of W29's scope, still open at 29e566f (piece 1 is not closed by this commit):** D1/T1 hardening flags (no
  `--user`, `--pids-limit`, `--cap-drop ALL`, `--security-opt no-new-privileges`, `--read-only`); D3 export reads
  through a linked directory; D4 a linked checkout is handed to the runtime as a mount source, and
  `hasPreparedTrees` accepts linked trees. All four tests red today (Ground table).
- **What change content still decides, by design:** the code `go test` runs inside the container — the test body,
  `TestMain`, `init`, `go.mod` (a `toolchain` line would try a download, which `--network none` refuses). The argv does
  not move; the bound on that code is ADR-0025's container, whose hardening is D1.
- **A directory whose name ends in `.go`** (accepted: `a.go/b_test.go` → `./a.go`) is read by `go test` as a file list,
  not a package; the run fails rather than doing anything more. Not measured against a real `go` binary.
- **The fixture's `recorded_tests.command`** is copied verbatim to the review by the `Sample` adapter
  (`internal/testrun/sample.go:53`); no test ties it to the template. Display text only; sample mode runs no code.
- **`docs/SAMPLE_MODE.md:6`** still says every output on the page was run at main `26343c2`; line 86 now shows an
  output that commit could not produce (C2).
- **The owner's second review (NVIDIA through Nightshift) is owed** for this change.

### Profile

Security boundary: model answer (Prover, Skeptic) → test file and name → template argv → container runtime. Spoofing:
n/a. Tampering with argv or the in-container command: closed by type removal plus a single builder checked at the
runner. Information disclosure: unchanged (no `-e`, read-only mount). Denial of service: a refused name costs no
container. Elevation inside the container: unchanged (D1 open). No new dependency, store, reader or persistence.
No invisible characters or instruction-like comments in the 21 files of the diff (read in full for the six code files).

### Focus

master-was-wrong: the "33 packages ok" measurement does not reproduce here (27 ok, 6 sandbox listener panics); the
negative control the master did not repeat was repeated and holds. Reviewer rulings: (1) the order confines artifacts
to `████████████████` in the worktree, but that directory is not git-ignored and this leg holds no delete; scratch files
went to this run's artifacts directory instead, and the worktree is left as found. (2) The order's H1 for
`_REVIEW_STATUS.md` names my own slug; the engine's channel grammar names the reviewed slug and says it binds the
verdict, so the H1 follows the engine. (3) The order's state token `recommendation-ready` is the one the review
skeleton lists as write-legal.

### Conditions

Verdict `accept-with-conditions`. Neither condition blocks the property; both are documentation.

- **C1 (documentation, ADR-0027 point 4)** — the file-and-name rule and the fixed template are cited as ADR-0025 in
  `internal/testrun/argv.go:28,58`, `internal/agent/prover/prover.go:82`, `internal/agent/skeptic/skeptic.go:66`,
  `internal/testrun/container.go:237,276`, but ADR-0025 does not carry that rule. Either amend ADR-0025 in place with a
  point stating it (law #21), or change the citations to the decision that does carry it. Check: `grep -n` of the
  cited ADR shows the rule.
- **C2 (documentation)** — `docs/SAMPLE_MODE.md:6` names the commit every output on the page was run at; re-run the
  page's commands at the merge commit and re-stamp the line, or the page states a provenance its line 86 contradicts.

### Unrun

- Real container execution (Docker): not run; no runtime established on this leg. Every argv claim here is from a
  recording fake runtime, never a daemon.
- The listener-dependent tests in six packages: not run (sandbox refuses a loopback listener); their packages' other
  tests were run where the panic would otherwise have hidden them (`pipeline`, `review`).
- Main's own binary for the sample comparison: not built; compared against main's recorded output instead.
- `go test` on a real module with a `<x>.go` directory name: not run.
- The NVIDIA review: not run (route down), owed.

## The review status as filed

## REVIEW — the-prover-runs-only-its-test
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: Reviewed on claude-opus-5-5 at 29e566f: no model text or change content shapes the container argv or command; proven by type removal plus TestArgv at Container.Run, unplug controls red as claimed.
  Two documentation conditions: ADR-0025 is cited for a rule it does not hold (C1); SAMPLE_MODE.md provenance line stale (C2).
  Suite: 27 ok, 6 FAIL from this sandbox's loopback-listen refusal only; pipeline and review green with listener tests skipped.
  Original reviewers' D1/T1/D3/D4 tests still red; out of W29 scope, piece 1 not closed. NVIDIA second review owed.
- artifacts:
    - _REVIEW_NOTES.md
    - _REVIEW_STATUS.md
- updated: 2026-10-23T12:40:00Z

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| private workspace path | 1 |
