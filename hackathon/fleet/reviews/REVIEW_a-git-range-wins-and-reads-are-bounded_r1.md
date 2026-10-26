# REVIEW — a-git-range-wins-and-reads-are-bounded · round 1

- verdict: accept-with-conditions
- reviewer: claude-code on macm5, claude-opus-5-5
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-23T15:08:04Z

## Summary

W31 at 0a4341c, reviewed on claude-opus-5-5: the git range decides the trees when a baseline is given, every repository read is bounded, and six unplug mutants each turn their test red.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — a-git-range-wins-and-reads-are-bounded

Reviewer: named REVIEW leg `a-git-range-wins-and-reads-are-bounded-review`, claude-code on `claude-opus-5-5`.
Reviewed `436ef04...0a4341c` (three commits). `git rev-parse --short=7 HEAD` printed `0a4341c`. Profile: security.
Date: 2026-10-23. Go: `go1.26.8 darwin/arm64`. I wrote no code under review. Every probe ran through `go test -overlay`
from files outside the worktree (see Remainder).

**Verdict: accept-with-conditions.** The two halves this stream claims hold, and each guard turns its own test red
when it is removed: a baseline makes the git range decide the trees, and every read of repository content is bounded.
One condition must land before merge (C1): the largest value the bound settings accept wraps `limit+1` to a negative
number, and every read then returns **empty content with no error**. Separately, the piece-3 property as a whole still
does NOT hold at the tip. The two symlink findings (p3-claude F2/F3 = p3-codex F1/F2) are untouched by this diff and
their four original tests are still red (C2).

### Ground

The diff is 13 files, +663/−68: `internal/change/{limits.go (new), archive.go, between.go, read.go, bounds_test.go (new)}`,
`internal/pipeline/{trees.go, trees_range_test.go (new)}`, `internal/review/{config.go, decide.go, open.go, bounds_test.go (new)}`,
`internal/settingspage/catalog.go` and `docs/CONFIGURATION.md`. I read every hunk, plus whole files where a guard lives:
`change/{read,archive,between}.go`, `pipeline/trees.go`, `review/decide.go:345-395` and `testrun/written.go:80-194`.

**The property, part by part (file · line · control · test):**

| Part | Line | Control | Test that proves it (and the unplug) |
|---|---|---|---|
| A baseline means the range decides the trees | `pipeline/trees.go:51-53` | `BaselineCommit != ""` goes to `materialize`, and `hasPreparedTrees` is never consulted | `TestBaselineCommitDecidesTheTreesNotTheRepositoryDirectories`. Red under M1 |
| No baseline means prepared trees (the demo) | `trees.go:57`, `:63` | unchanged | `TestPreparedTreesAreReadWhenNoBaselineIsGiven` |
| ExportCommit checks its revision | `change/archive.go:24` | `ensureCommit(..., "exported")` | `TestDashLeadingRevisionIsRefusedByEveryReader` asserts the `does not exist` sentence. Red under M3 |
| A dash-leading revision never becomes an option | `between.go:34,94,131,170,207` | `ensureCommit` plus `--end-of-options` on every revision argv | same test. The probe file is never created |
| Bound on the git output stream (archive, show, log, ls-files, diff) | `between.go:42-75`, `:35`, `:132`, `:173`, `:224` | `io.LimitReader(pipe, max+1)`, then kill | `TestExportRefusesAnArchiveOverTheArchiveBound`, `TestBlobAndMessagesRefuseOutputOverTheFileBound`. Both red under M4 |
| Per-file bound in untar | `archive.go:120` | `hdr.Size > File` | `TestExportRefusesAFileOverTheFileBound` (ExportCommit half). Red under M5 |
| Per-file and total bounds on the working-tree copy | `archive.go:226-250`, `:72-75` | Lstat size, then `LimitReader(limit+1)`; running total against Archive | same test (ExportWorktree half), red under M6; `TestExportWorktreeRefusesATreeOverTheArchiveBound` |
| Prepared-tree read | `read.go:88`, `:92-95`, `:107-124` | `ReadBounded` per file, Archive total per tree | `TestReadRefusesAFileOverTheFileBound`, `TestReadAcceptsAFileAtTheFileBound`. Red under M2 |
| proven.yaml in the target / explicit `--config` | `review/config.go:207`, `:161` | `ReadBounded(Config)` | `TestTargetConfigOverTheConfigBoundIsRefused`, `TestConfigFileOverTheConfigBoundIsRefused`. Both red under M2 |
| proven.yaml at the baseline | `config.go:182-185` | `len(b) > Config` after a File-bounded `Blob` | **no test in the diff.** My probe P4 passes |
| Author record, summed over its files | `review/decide.go:365-370` | `remaining` decremented per file | `TestAuthorRecordOverTheRecordBoundIsRefused` covers one file only. **Summing is untested in the diff.** My probe P3 passes |
| A setting cannot make a bound unbounded | `limits.go:117-123`, `:57-73` | `n <= 0` refused; zero fields keep their defaults | **no test in the diff.** My probe P2 passes. See C1 for the top end |
| The repository cannot set a bound | `limits.go:12-20` | env-only and closed; `proven.yaml`'s closed key set has no bound key | read; no test needed |

**Question 2: reads of repository content without a bound.** `grep -rnE 'os\.ReadFile|io\.ReadAll|exec\.Command|\.Output\(\)|CombinedOutput'`
and `grep -rnE 'io\.Copy|os\.Open\(|bufio\.New|ReadDir\('` over non-test `internal cmd` code. These sites remain:
- `between.go:17-18` `WorkTree` (`rev-parse --is-inside-work-tree`), `:207-208` `ensureCommit` (`rev-parse --verify --quiet`) and `:284-286`
  `gitZText` (`rev-parse --show-toplevel`): fixed, tiny outputs, not repository content. Acceptable.
- `between.go:43-44`: **git's stderr goes to an unbounded `bytes.Buffer`** (C4).
- `testrun/written.go:169` `copyFile` `io.Copy` without a limit. Its source is the change tree, which an export or `Read` has already bounded
  (transitively bounded; not a new gap).
- `fixture/fixture.go:74` (sample mode only), `pipeline/memory.go:39` (the operator's cwd `go.mod`), and osv/world/hmem/query/cmd: these read operator state, not repository content.
- `manifest/*` parse bytes that `change.Read` returns, so they are bounded.

**Question 3: a value from the change reaching a git argument as an option.** No path was found. Revisions come from the
operator or CI, never from the change. Each passes `ensureCommit` and sits after `--end-of-options`. The one
repository-adjacent argument is `Blob`'s `rel`, which the only caller fixes to `"proven.yaml"` (`config.go:175`), behind the
`..`/`/` filter (`between.go:127`). `ls-files` and `diff` take no value from the change. The paths git prints are consumed
as output and never passed back as argv.

**The original reviewers' tests, run on 0a4341c** (overlay; command:
`go test -overlay=<scratch>/overlay.json -count=1 -v -run '<names>' ./internal/change ./internal/review ./internal/pipeline`):

| Test (origin) | Finding it pins | At 0a4341c |
|---|---|---|
| `TestReviewP3DashLeadingRevisionIsNotAnOption` (claude) | control | PASS |
| `TestReviewP3BaselineReviewIgnoresBaseAndChangeDirectoriesInTheRepository` (claude) | claude F1 | **PASS** (red at 436ef04) |
| `TestReviewP3ReadRefusesAnOversizedFile` (claude) | claude F4 | **PASS** |
| `TestReviewP3SettingsInTargetRefuseAnOversizedFile` (claude) | claude F4 | **PASS** |
| `TestReadRefusesOversizedRepositoryFile` (codex) | codex F3 | **PASS** |
| `TestExportCommitRefusesOversizedRepositoryFile` (codex) | codex F3 | **PASS** |
| `TestTargetConfigRefusesOversizedFile` (codex) | codex F3 | **PASS** |
| `TestAuthorRecordRefusesOversizedFile` (codex) | codex F3 | **PASS** |
| `TestReviewP3ExportWorktreeDoesNotFollowASymlinkedParent` (claude) | claude F3 | **FAIL**: `ExportWorktree copied a file from outside the target through a symlinked directory` |
| `TestExportWorktreeRefusesLinkedParent` (codex) | codex F1 | **FAIL**: `a tracked path through a linked parent was accepted` |
| `TestReviewP3SettingsInTargetRefuseASymlink` (claude) | claude F2 | **FAIL**: `a proven.yaml symlink to a file outside the target was loaded (timeout 7m0s)` |
| `TestReviewP3SettingsRefusalDoesNotQuoteAFileOutsideTheTarget` (claude) | claude F2 | **FAIL**: `... proven.yaml="marker-outside-the-target": missing colon ...` |
| `TestTargetConfigRefusesSymlinkOutsideTarget` (codex) | codex F2 | **FAIL**: `target config symlink outside the target was accepted` |

remark: the five FAILs are the symlink findings, and this diff does not claim them. They are red, so piece 3 is not closed (C2). They are not a defect of W31.

**Question 4: `go build ./... && go vet ./... && go test -count=1 ./...`.** Build: ok. Vet: ok. Test: 27 packages `ok`
and 6 `FAIL`: `cmd/livecheck`, `internal/generator`, `internal/guardrail`, `internal/pipeline`, `internal/query` and
`internal/review`. Each failure is the same panic, `httptest: failed to listen on a port: listen tcp6 [::1]:0: bind:
operation not permitted`. The seat's sandbox refuses a localhost bind. The cause is the environment, not the diff. A panic
aborts the rest of its package, so I re-ran the two packages the diff touches with the binding tests skipped:
`go test -count=1 -skip 'Live|Unsafe|Guardrail|Server|HTTP|Http' ./internal/pipeline` → `ok` (28.2s), and
`go test -count=1 -skip 'Live|Unauthorized|Closed|Cancelled|Zetaris|AdapterTimeout|Server' ./internal/review` → `ok` (49.6s).
`internal/change` (196.7s), `internal/settingspage` and `cmd/settings-doc` (the generated-doc check) passed in the full run.

### Differential

**C1 · BLOCKING (silent wrong result) · `change/read.go:116`, `change/between.go:52`, `change/archive.go:243` (with
`limits.go:117-123`) · the largest accepted bound wraps to a negative limit, and every read returns empty.** `LoadLimits`
accepts any positive `int64`, `9223372036854775807` included. Each reader then computes `limit+1`, which overflows to
`math.MinInt64`, and an `io.LimitReader` with a negative N reads nothing. Size pre-checks (`info.Size() > limit`) never
fire, so the reader returns an empty body and no error. The review then runs over empty files, and nothing says so.
Measured with the overlay probes (fail on 0a4341c):
`zz_w31probe_test.go:37: file a.txt read as base="" change=""` (`Read` with File=MaxInt64) and
`zz_w31probe_test.go:47: Blob returned 0 bytes ok=true err=<nil>` (`Blob` with File=MaxInt64);
`TestW31ProbeMaxInt64BoundIsAccepted` passes, so the value reaches `Bounds()` from `PROVEN_MAX_FILE_BYTES`.
For git output larger than the pipe buffer, `outputBounded` then waits on a git that blocks writing. That is a probable
hang, inferred and not measured. Trigger: an operator who types "the biggest number" to mean "no limit". The value comes
from the operator's environment, not the repository, so it falls outside the change-injection property. The result is
still wrong and silently trusted, which is the blocking class under R6. Remedy: refuse values above a ceiling in
`LoadLimits` (anything at or below `MaxInt64-1` avoids the wrap; a product ceiling such as 1 TiB is saner), or compute
the +1 with a saturating guard. Add a test of the probe's shape that asserts content survives, or a refusal, at the
largest accepted value.

**C2 · OWED (scope, for the master) · piece 3 is not closed by W31.** p3-claude F2/F3 and p3-codex F1/F2 have five red
tests at the tip (table above): a `proven.yaml` symlink in a no-baseline review, and a tracked path read through a
symlinked parent in `ExportWorktree` (`archive.go:66-68` → `copyTracked` Lstat checks the last component only). This
diff does not claim them. `SECURITY_REVIEW.md` must not mark piece 3 closed on this merge. The stream that carries them
has to be named. Note also that `ReadBounded` follows links by contract (`read.go:105-106`), so `config.go:207` still
opens through a link. The bound is applied to the link target's bytes.

**C3 · OWED (tests the diff lacks; R3 survivors in waiting).** None of these guards has a test, so deleting any of them
turns nothing red:
(a) `LoadLimits` refusing `0`, `-1`, `abc`, `1.5`. Probe P2 passes; commit it.
(b) the baseline `proven.yaml` Config bound at `config.go:182-185`. Probe P4 passes (2 MiB at the baseline refused, content not quoted).
(c) the author record summed over several files at `decide.go:365-370`. Probe P3 passes (two 700 KiB files under `*.txt` refused).
(d) `Open` wiring `Options.Env` into `UseLimits` (`open.go:41-47`), plus `readLimitConfig` turning a bad value into a `ConfigError`. Unmeasured.
(e) the `Messages` bound on its own: in `TestBlobAndMessagesRefuseOutputOverTheFileBound` the Blob half fatals first under M4, so the Messages half is unobserved.

**C4 · OWED (R8) · `between.go:43-44, :65`.** The doc comment says the refusal "never [names] the output". The failure
path, though, makes git's stdout the error message when git exits non-zero, and `gitFailure.Error()` prints it. That can
be up to the bound of repository content (128 MiB for an archive). For `Blob` this is new: the old code used stderr only.
stderr itself is an unbounded buffer. Remedy: prefer stderr, clip both (as `clipValue` does), and never fall back to stdout.

**C5 · OWED (R8) · `limits.go:43` process-global limits.** `Open` sets the package-wide pointer, and `Open` in sample mode
leaves the previous live `Open`'s values in place. Two reviewers in one process with different `Env` share whichever
wrote last. No test runs in parallel today (`grep t.Parallel` over change/review/pipeline: none), so the tests are not
racy. Remedy: carry `Limits` on the request/reader, or reset to defaults in sample mode.

**C6 · judgement (R6, does not block) · `config.go:175-185`.** The baseline `proven.yaml` is read up to the File bound
(8 MiB) before the 1 MiB Config check. The read is bounded, at 8×. Passing the Config bound to the blob read would
make the check exact.

### Unplug

Each guard was removed once, through an overlay that replaced one source file. No tracked file was edited.

| Mutant | Guard removed | Result |
|---|---|---|
| M1 `pipeline/trees.go` | baseline first → prepared trees first (the 436ef04 order) | `trees_range_test.go:37: a review with a baseline commit read the repository's own base/ and change/ directories`. **RED** |
| M2 `change/read.go` | `ReadBounded` = `io.ReadAll(f)` | change `bounds_test.go:73`; review `bounds_test.go:20`, `:35`, `:48`. **RED ×4** |
| M3 `change/archive.go` | `ensureCommit` in `ExportCommit` | `bounds_test.go:46: ExportCommit refused the revision without the commit check: git: fatal: not a valid object name: --output=...`. **RED** |
| M4 `change/between.go` | `LimitReader` in `outputBounded` | `bounds_test.go:110` (archive), `:132` (blob). **RED ×2** |
| M5 `change/archive.go` | `hdr.Size > File` in `untar` | `bounds_test.go:85: ExportCommit wrote a committed file over the file bound`. **RED** |
| M6 `change/archive.go` | both file-bound checks in `copyTracked` | `bounds_test.go:88: ExportWorktree copied a tracked file over the file bound`. **RED**; the archive-total test stays green, as it should (a separate guard) |

remark: M3 shows that `--end-of-options` alone also stops the option (`not a valid object name: --output=...`). The two layers are redundant by design, and the test pins the outer one.
remark: the guards in C3 were not unplugged. They have no test, so the unplug would show nothing red. That absence is the finding.

### Closure

What the change controls: tracked paths and bytes, file sizes, commit messages, the `base/` and `change/` directory
names, the bytes of `proven.yaml` at the baseline and in the target, the files `author_record` matches, and, in a
headless review, working-tree links. What it does not control: `target`, `baseline`, `head`, `ConfigPath`, the
`PROVEN_MAX_*` environment and the container settings. With this diff, the change can no longer choose the trees when a
baseline is given. It can no longer make any repository read unbounded through size. It cannot reach a git argv as an
option. It still can reach outside the target through a link (C2). C1 is an operator-side input.

### Remainder

- **Artifact confinement deviation (named, for the master).** The order puts artifacts under `██████████████████>/` in the worktree and asks
  that `git status --porcelain` show nothing of mine. `████████` is not ignored (`cat .gitignore`; `git status --porcelain --ignored` shows
  only `.claude/` ignored), and this seat's grant has no `rm`. Anything I wrote there would have stayed as untracked files. I put every
  scratch file in the run's artifacts directory instead, outside the worktree:
  `█████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████`
  (`overlay.json`, `m1.json`–`m6.json`, `m1/`–`m6/` mutants, the five original-test files and two `*_w31probe_test.go`). The worktree holds only
  the channel files.
- **H1 discrepancy, order vs engine.** The order's return manifest names the H1 `# REVIEW — a-git-range-wins-and-reads-are-bounded-review`. The
  engine's injected grammar says the H1 is the reviewed slug without `-review`, and that a row under the leg's own slug is never seen by
  `merge`. I followed the engine. The master should correct the order template.
- **Declined to judge (R11):**
  (a) a git repository with no baseline and its own `base/`+`change/` is still read as prepared trees (`trees.go:63`). That is the operator
  omitting the baseline, and the order's property is scoped "with a baseline". A CI that forgets the baseline lets the change choose the trees again.
  A refusal for "git repo + no baseline" whatever the directories hold would close it, at the cost of the demo layout inside a git checkout.
  (b) `recordPaths` puts no limit on the NUMBER of matches. Many empty files cost CPU and file opens, not bytes.
  (c) the bound values themselves (128 MiB, 8 MiB, 1 MiB, 1 MiB) are product choices. I did not judge them.
- The NVIDIA half of the owner's two-review rule is **owed**. The route is down today, per the order.

### Profile

security. S1: the trust map is Closure. S2: every finding is reproduced (C1 by two red probes; C2 by five red original
tests). S3: there is no authorization check in scope; the remove-the-guard act is the Unplug table, with six mutants and
six reds. S4: there is no dependency delta (no `go.mod`/`go.sum` in the diff). No destructive operation on a derived path
is new. `RemoveAll` is still only on the `MkdirTemp` dir (`trees.go:94`). `git diff 436ef04...0a4341c | grep -c '[^ -~<TAB>]'`
printed `0`, so the diff has no non-ASCII, invisible or bidi bytes. I found no instruction-like comments.

### Focus

*Can a git range, a path, a manifest, a lock file or proven.yaml from the change inject arguments, lead Proven to read or
write outside its places, or read without bound?*
- **Inject arguments: no** (Ground Q3; M3; the dash-leading control passes).
- **Read without bound: no** from the change. Yes from the operator's top-end value, which reads as empty rather than unbounded (C1).
- **Read outside its places: the trees, no**, with a baseline (M1). **A link, yes**, still (C2; not this stream's claim).

**master-was-wrong (law #44), each claim re-measured:**
1. "33 packages ok". **Not reproduced on this seat, and not refuted.** 33 test packages ran: 27 ok and 6 failed on the
   sandbox's localhost-bind refusal. The two in-scope packages among the six pass with the binding tests skipped. The other four
   (`cmd/livecheck`, `generator`, `guardrail`, `query`) are not touched by the diff. Their network tests are UNMEASURED here.
2. "prepared trees allowed to win with a baseline → red". **Holds** (M1, same test, same sentence).
3. "the file bound lifted → red". **Holds** (M2: four reds across change and review). M4–M6 also hold for the other bound sites.
4. Order text: "every read of repository content has a named size bound and is refused over it". **Holds**, except at the
   accepted maximum value, where the read is not refused but emptied (C1).
5. "The leg's own step-1 red output was never recorded". I could not supply it. The original reviewers' tests at 436ef04 are
   the recorded reds, and they pass now (table in Ground).

### Conditions

1. **C1 (blocking, before merge):** `LoadLimits` refuses a bound that would overflow `limit+1` (or every reader saturates the +1), with a test
   that sets the largest accepted value and asserts content is read or the value is refused, never emptied.
2. **C2 (owed, before piece 3 is closed):** the master names the stream that carries the symlink findings (p3-claude F2/F3, p3-codex F1/F2).
   Piece 3 stays open until those five tests pass.
3. **C3 (owed):** commit tests for (a) `LoadLimits` refusals, (b) the baseline Config bound, (c) record summing, (d) `Open` env wiring,
   (e) the `Messages` bound observed on its own.
4. **C4 (owed):** git's failure message comes from clipped stderr, never stdout; the stderr buffer is bounded.
5. **C5 (owed):** limits stop being last-writer-wins process state, or sample-mode `Open` resets them.
6. **C6 (judgement):** read the baseline `proven.yaml` with the Config bound.

### Unrun

- `git version` / `git --version` / `git check-ignore`: refused on this seat (`This command requires approval`). I do not know the git version the
  `--end-of-options` results were measured on.
- The network-binding tests of `cmd/livecheck`, `internal/generator`, `internal/guardrail`, `internal/query`, and the skipped binding tests of
  `internal/pipeline` and `internal/review`: refused by the sandbox (`bind: operation not permitted`). They are owed to a seat that can bind localhost.
- The hang in C1 for git output over the pipe buffer at a MaxInt64 bound: inferred from `os/exec` semantics, not run.
- Race detector (`go test -race`): not run. The order did not ask for it.
- The NVIDIA independent review: owed (route down).

## The review status as filed

## REVIEW — a-git-range-wins-and-reads-are-bounded
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: W31 at 0a4341c, reviewed on claude-opus-5-5: the git range decides the trees when a baseline is given, every repository read is bounded, and six unplug mutants each turn their test red.
  BLOCKING C1: the largest bound LoadLimits accepts (MaxInt64) wraps limit+1, and Read/Blob return EMPTY content with no error (two red probes).
  C2: piece 3 is not closed; the symlink findings (5 original tests) are still red and outside this diff.
  Build and vet ok. go test: 27 packages ok; 6 hit a sandbox localhost-bind refusal; pipeline and review are ok with those tests skipped.
- artifacts:
    - _REVIEW_NOTES.md
    - _REVIEW_STATUS.md
- updated: 2026-10-23T15:40:00Z

## REDACTED

| what | times |
|---|---|
| absolute or private path | 3 |
