# REVIEW — security-piece-8-fixes · round 1

- verdict: reject
- reviewer: claude-code on macm5, claude-opus-5-5
- cost: not recorded
- conditions: 1 BLOCKING · 2 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-24T08:46:58Z

## Summary

Reject (claude-code, claude-opus-5-5; the NVIDIA second review is owed). The leg's disk-property guards hold, and unplugging each one turns its test red.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — security-piece-8-fixes

Reviewer: claude-code, model claude-opus-5-5 (one review; the owner's second, NVIDIA review through Nightshift is
OWED, its route being down on 2026-10-24). Profile: security. PRIVATE until merged.
`git rev-parse --short=7 HEAD` printed `37de26a`. Range: `656dbcb...37de26a`; `git merge-base 656dbcb 37de26a`
printed `656dbcb`, so the three-dot and two-dot diffs are the same.

Verdict: **reject**. The five data-on-disk properties hold on the paths the leg built, and each guard I unplugged turns
its test red, except one. The branch as a whole cannot land, for three reasons:
- **B1:** the absorb merge reverts main's Rust adapter.
- **B2:** property 4 fails on a secret that main added.
- **B3:** property 1 fails on a case-insensitive file system.

### Ground

Measured on this seat (darwin, the go toolchain in the sandbox), 2026-10-24:

- `go build ./...`: clean. `go vet ./...`: clean.
- `go test -count=1 ./...`: 41 packages `ok`. Six FAIL, all at an `httptest` listener:
  `cmd/livecheck`, `internal/generator`, `internal/guardrail`, `internal/pipeline`, `internal/query` and
  `internal/review`. Quoted refusal:
  `panic: httptest: failed to listen on a port: listen tcp6 [::1]:0: bind: operation not permitted`.
  A panic stops the package, so I re-ran the five internal ones with the listener tests skipped
  (`-skip 'TestLive|TestServer|TestClosed|TestUnauthorized|TestSilent|TestRedirect|TestUnsafe|TestGuardrailOff|TestRecordedFixturePrompt|TestSafeGuardrail|TestZetaris|Echoed|TestAllServicesReady|TestGeneratorErrorIsNotReady|TestNVIDIAUsesItsModel|TestABadVerdict|TestNonSelectIsNotSent|TestQueryUsesTheConfiguredView|TestContainerChecks|TestMissingImage|TestBadEvidenceView|TestSecretFromTheEnvironment|TestCancelledContext|TestLoadRefusesABadSettingAndDoesNotPrintTheKey'`).
  All five were `ok`; `cmd/livecheck` still panics at `main_test.go:360`.
- I ran the leg's new tests by name. All PASS:
  - generator, guardrail and query `…TooShortToBeRedacted` and `…OfTheMinimumLength`;
  - `TestTraceOmitsAConfiguredSecretOfAnyLength`;
  - `TestAConfiguredValueIsRedactedWhateverItsLength` and `TestFromEnvKeepsAShortOrganisationIdOutOfTheRedactor`;
  - hmem and world `TestOpenRefusesAStateDirectoryReachedThroughALinkIntoTheRepository`,
    `TestOpenNarrowsAnExistingWideStateFile` and `TestAtomicWriteDoesNotCarryAWiderTemporaryMode`;
  - hmem `TestOpenRefusesAStateDirectoryBehindALinkThatDoesNotResolve`;
  - pathguard `TestInsideComparesResolvedPaths` and `TestResolveRefusesALinkThatDoesNotResolve`;
  - review `TestReviewRefuses{LogsAndStateInsideTheReviewedTarget,ATraceInsideAnEmptyTarget,ALogsDirectoryReachedThroughALinkIntoTheTarget}`;
  - pipeline `TestMaterializedTreesAreNotGroupWritableUnderAPermissiveUmask`. The older
    `TestMaterializedTreesAreReadableByOthers` now PASSes too; it was red at d82cac2.
- The security reviewers' tests in `████████████████████████` have shipped as committed equivalents. Each one's
  assertion now lives in a leg test that passes:
  - B1 / BLOCKING 1 → `internal/review/placement_test.go:19,42`;
  - B2 / BLOCKING 2 → `hmem/placement_test.go:14`, `world/placement_test.go:24`;
  - O1 / BLOCKING 3 → `hmem/placement_test.go:47`, `world/placement_test.go:44`;
  - O2 → `…:71` / `…:68`;
  - BLOCKING 4 → `internal/trace/short_secret_test.go:15`;
  - O3 → `pipeline/trees_mode_test.go:17`.
- remark: the green packages were measured with the listener tests skipped; a regression that shows only in a live
  adapter test is UNMEASURED here (see Unrun).
- remark: a passing placement test shows that a refusal happens on its layout; it does not show that every spelling of
  a path is compared (see B3).

### Differential

**B1 · BLOCKING · the absorb merge `5b89e3e` reverts main's Rust adapter (ADR-0038) and more.**
- `git diff --stat 656dbcb 5b89e3e` shows the merge deleting main content that has nothing to do with this stream:
  - `internal/manifest/cargo.go` (433 lines), `cargolock.go`, their tests, and `internal/cargover/` (whole package);
  - `testdata/fixture/osv-rust-cargo/`, and the Rust scenario in `cmd/proven/scenario_test.go` (−29);
  - Rust lines in `docs/ADAPTERS.md`, `docs/SAMPLE_MODE.md`, `docs/OSV_TABLE.md`, `docs/QUERIES.md`,
    `demo/live/README.md` and `hackathon/SAPHAN.md`.
- `internal/manifest/adapter.go:20` loses `Cargo(), CargoLock()`.
- `internal/querytpl/querytpl.go` loses the `crates.io` ecosystem and `crateFormat`.
- `internal/osv/fix.go` `compareFor` goes back to a hard-coded `switch` over Go, npm and PyPI. That also drops version
  ordering for **Packagist**, which main orders through the adapter's `Compare`, so smallest-fix ordering regresses for
  PHP as well as Rust.
- `git ls-tree -r --name-only 5b89e3e internal/manifest` holds no cargo file. The merge's second parent is `656dbcb`,
  so git records Rust as merged while the tree has none of it. A later merge of main will not restore it.
- The order says the absorb had "one hand-resolved conflict in internal/generator/settings.go". The tree shows a much
  wider resolution, and the merge message (`[security-piece-8-fixes] merge: absorb main`) does not name it.
- R2: the diff changes behaviour the order did not ask for. Remedy: redo the absorb from `7d6ea62` onto `656dbcb`,
  keeping main's side for every file this stream did not touch. Then re-apply `37de26a`. Check with
  `git diff --stat 656dbcb <new tip>`, which must list only this stream's files.

**B2 · BLOCKING · property 4 fails for `██████████████████████████`. Main added it, and the merge did not apply the
floor to it.**
- `internal/generator/second.go:66` checks only `\r\n` and `> 4096`. It has no `service.TooShortSecret`, so a 3-byte
  key loads.
- `redact.Names` (`internal/redact/redact.go:29`) includes `EnvSecondKey`, and `New` has no length floor any more. A
  short second-reviewer key is therefore replaced as a substring in every trace text. That is the text damage the
  floor exists to prevent, and the load does not refuse it by name.
- The comment at `redact.go:101-102` ("The three credentials … a short one is refused by the settings loaders, so none
  reaches here") is false at the tip: five credentials are listed, and one loader does not refuse.
- Proof: probe `TestProbeSecondReviewerKeyBelowTheFloorIsRefused` FAILS on 37de26a:
  `a 3-byte second-reviewer key was accepted at load (connected=true)`. Its control,
  `TestProbeSecondReviewerKeyControlLoads` (16-byte key), PASSES. The full text is in Closure.
- Remedy: add `|| service.TooShortSecret(key)` at `second.go:66`, with a refusal that names the floor. Add a committed
  twin of the probe. Correct the redact comment.

**B3 · BLOCKING on darwin · property 1 (and 2) compares paths case-sensitively on a case-insensitive file system.**
- `internal/pathguard/pathguard.go:56-68` compares the `EvalSymlinks` results with `filepath.Rel`.
- On APFS and HFS+ default volumes, `EvalSymlinks` keeps the case of each component as it was given, so
  `<base>/REPO/logs` and `<base>/repo` are judged "not inside" although they are the same directory.
- `pipeline.begin` (`memory.go:67-79`) then lets the trace and the state be written inside the target. The same applies
  to hmem and world `refuseInside`.
- Trigger in practice: a working directory spelled in one case (`cd ~/code/Repo`, so `os.Getwd` returns `$PWD`) and a
  TARGET spelled in another. darwin is this fleet's primary seat.
- Proof: probe `TestProbeLogsSpelledInAnotherCaseAreRefused` FAILS on 37de26a:
  `written inside the reviewed target through a case-folded spelling: [logs state]`. Its control,
  `TestProbeLogsSpelledExactlyAreRefused`, PASSES.
- Remedy: decide inside-ness by identity, not spelling. Walk the resolved `dir` upward and compare each ancestor with
  the resolved root using `os.SameFile`, which compares device and inode. For a path not yet created, use the nearest
  existing ancestor. Add a committed test that skips on a case-sensitive volume, as the probe does.

**O1 · OWED · the working-tree export's mode fix has no test.**
- `internal/change/archive.go:284-291` (`copyTracked`, used by `ExportWorktree` when no head commit is given) is
  correct by reading.
- Unplugged, nothing turns red: I reverted only `copyTracked` to the pre-leg form (source mode, umask, no Chmod) by
  overlay, and `internal/change` and `internal/pipeline` both stayed `ok`.
- `trees_mode_test.go` drives two commits, so only `writeTarFile` is covered. Remedy: a twin test with
  `HeadCommit: ""` and a 0664 tracked file under umask 002.

**O2 · OWED · doc drift after the merge.**
- `redact.go:24-29` says Names lists values "Proven must never write", including the organisation id. Since `7d6390c`,
  an id shorter than 8 bytes *is* written.
- The sentence should say what the code does, and ADR-0013/0005 should carry the ruling (judged under Focus).

**N1 · NOTE · `begin` returns early on `p.trace != nil` (`memory.go:64`).** A second `Review` on one `Reviewer`, with
a different target, skips the target comparison and writes to a trace that is already closed. Before the leg, the
trace was opened in `Build` and closed after the first review, so the code was single-use before too. No caller
reviews twice (the `review/open.go:289` path runs once per `Review`; a `DoubleRun` stays inside one pipeline call).
Remedy, optional: refuse a second `Review` by name.

### Unplug

Each control was run by `go test -overlay`, with mutants in the artifact directory. No worktree file was edited.
- Target comparison off (`if false && target != ""` in `memory.go:67`): `TestReviewRefusesLogsAndStateInsideTheReviewedTarget`,
  `…ATraceInsideAnEmptyTarget` and `…ALogsDirectoryReachedThroughALinkIntoTheTarget` all FAIL. Unplugging the guard
  raises the alarm.
- Link resolution off (`pathguard.Resolve` = `filepath.Abs` only):
  - FAIL: pathguard `TestInsideComparesResolvedPaths`, `TestResolveRefusesALinkThatDoesNotResolve`; hmem and world
    `TestOpenRefusesAStateDirectoryReachedThroughALinkIntoTheRepository`; review `…ALogsDirectoryReachedThroughALinkIntoTheTarget`.
  - Survived: hmem `TestOpenRefusesAStateDirectoryBehindALinkThatDoesNotResolve`. A dangling link is refused later by
    `MkdirAll` either way, so that test does not pin the resolver. The pathguard unit test does.
- `copyTracked` mode fix reverted: nothing turns red (O1).
- Short-secret floor: the master's measurement (2026-10-24, outside the sandbox) is that forcing it off turns the three
  loader refusal tests and the organisation-id test red. I did not repeat it. The tests exist and pass here, and B2
  shows a fourth loader the unplug never covered.
- Not unplugged by me: the `os.Chmod(state.json, 0o600)` on Open, and the `O_EXCL` temporary-file create. The leg's
  tests `TestOpenNarrowsAnExistingWideStateFile` and `TestAtomicWriteDoesNotCarryAWiderTemporaryMode` assert exactly
  those fields. The first matches the security reviewers' O1 test, which was red at d82cac2.

### Closure

The probes were run through `-overlay`; none was committed. Sources are in this run's artifact directory, `probe/`.

```go
// ██████████████████████████████████████████████ (overlay)
package generator_test
func TestProbeSecondReviewerKeyBelowTheFloorIsRefused(t *testing.T) {
	env := map[string]string{
		generator.EnvSecondKey:     "abc",
		generator.EnvSecondBaseURL: "https://api.example.com/v1",
		generator.EnvSecondModel:   "some-model",
	}
	_, connected, err := generator.LoadSecond(func(k string) string { return env[k] })
	if err == nil {
		t.Fatalf("a 3-byte second-reviewer key was accepted at load (connected=%v)", connected)
	}
}
// control: same env with "abcdefghijklmnop" loads without error.
```

```go
// █████████████████████████████████████████ (overlay)
func TestProbeLogsSpelledInAnotherCaseAreRefused(t *testing.T) {
	base := t.TempDir()
	target := filepath.Join(base, "repo")
	os.Mkdir(target, 0o755)
	alt := filepath.Join(base, "REPO")
	if _, err := os.Stat(alt); err != nil { t.Skip("case-sensitive file system") }
	rev, err := review.Open(review.Options{Sample: true, LogDir: filepath.Join(alt, "logs")})
	if err == nil { _, err = rev.Review(context.Background(), review.Request{Target: target}) }
	// FAIL if target/logs or target/state exists
}
// control: LogDir = target/logs (exact spelling) is refused and writes nothing.
```

Output on 37de26a:
```
zz_probe_second_key_test.go:20: a 3-byte second-reviewer key was accepted at load (connected=true)
--- FAIL: TestProbeSecondReviewerKeyBelowTheFloorIsRefused
--- PASS: TestProbeSecondReviewerKeyControlLoads
zz_probe_casefold_test.go:36: review err lstat …/repo/base: no such file or directory; written inside the reviewed target through a case-folded spelling: [logs state]
--- FAIL: TestProbeLogsSpelledInAnotherCaseAreRefused
--- PASS: TestProbeLogsSpelledExactlyAreRefused
```

B1 is shown by `git diff --stat 656dbcb 5b89e3e` and `git ls-tree -r --name-only 5b89e3e internal/manifest internal/cargover`,
both pasted in substance under Differential.

### Remainder

- **Check-then-use (question 2).** `pipeline.begin` resolves and compares paths, then `trace.Open`, `hmem.Open` and
  `world.Open` create them by name. A process that can write an ancestor of the logs or state directory could swap a
  segment for a link into the target between the two. That process needs the user's own write access to the
  ancestors; the reviewed target's content cannot do it, because the target is compared after resolution and the
  directories are outside it. Not blocking. Closing it fully needs `openat`-style creation with no-follow on each
  segment.
- **`os.Chmod(state.json)` follows a link** (`hmem.go:179`, `world.go:165`). A `state.json` that is a link narrows
  and reads its link target. This was the case before the leg too (`ReadFile` follows), and the state directory is
  0700 and the user's own. The next `persist` renames over the link itself. Remedy, optional: `Lstat` and refuse a
  non-regular `state.json`.
- **Secrets configured but unused by the active mode.** `EnvToken` in nvidia mode and `EnvKey` in nightshift mode are
  still read by `redact.FromEnv`, but no loader floor-checks them. A short unused value is redacted as a substring.
  This is the same class as B2, at lower likelihood.
- **Declined to judge (R11):** Windows path behaviour; the CLI's default target (the CLI train's, per the order); the
  container checkout under `os.MkdirTemp("")` (TMPDIR inside the target), which is outside the five properties;
  licences.

### Profile

Security. S1, data on disk:
- **Tampering with the reviewed checkout:** holds by spelling and by link, but does not hold across case spellings
  on darwin (B3).
- **Disclosure to another local user:** holds. Existing state files are narrowed, the temporary file is created
  exclusively at 0600, and the trees are not group-writable on the commit path. The worktree path is correct but
  untested (O1).
- **Disclosure of a secret:** holds for the four loaders the leg touched. It does not hold for the second-reviewer
  key (B2).

S3: each guard was removed once (Unplug). S4: the dependency delta is main's, through the merge, and B1 is that delta
taken out. No new modules: `go.mod` and `go.sum` are not in the diff. Destructive paths: `writeAtomic` removes only
`<state>.tmp`, and only when it is not a directory.

### Focus

1. **Outside the target:** holds except B3. The comparison is against `Request.Target`, not the module root
   (`memory.go:67-79`), and runs before any write.
2. **Links:** a link on any segment is resolved (`EvalSymlinks` on the whole path, or on the nearest existing
   ancestor), and an unresolved link is refused. A link created between check and use is a same-user race (Remainder).
3. **Narrow on open, atomic write:** holds (`hmem.go:178-181,560-583`; world twins).
4. **Secrets:** holds except B2.
   - **The organisation-id ruling: I accept it.** Zetaris authenticates with `███████████████`. The org id is a
     numeric account identifier (`query/settings.go:136,204`), so withholding it is hygiene, not a credential barrier.
     Redacting a 1–7 digit value as a substring would corrupt commit ids, line numbers and counts in every trace, and
     refusing short ids at load would refuse genuine Zetaris ids.
   - The cost is that a short id appears in the trace. Ids of 8 or more digits are still redacted.
   - The ruling needs its doc made true (O2). The owner may still overrule it on account-privacy grounds.
5. **Materialized trees:** holds on both export paths by reading. A test covers only the commit path (O1).

master-was-wrong:
- The order says the absorb had one hand-resolved conflict, in `internal/generator/settings.go`. The merged tree
  differs from main in 54 files. The cargo, cargover, crates.io, Rust-fixture and Rust-doc deletions listed in B1 are
  main content this stream never touched.
- The master's "merged tree green" is consistent with this, because the merge also deleted the Rust tests.

### Conditions

1. **B1 · BLOCKING.** Redo the absorb so the merged tree keeps main's Rust adapter, the Packagist `compareFor` and
   every other main-only file. `git diff --stat 656dbcb <tip>` must list only this stream's files. Re-run
   `go test ./internal/manifest ./internal/osv ./internal/querytpl ./cmd/proven` green.
2. **B2 · BLOCKING.** `LoadSecond` refuses a key below `service.MinSecretBytes` by name. A committed test fails
   without it.
3. **B3 · BLOCKING.** `pathguard.Inside` decides by file identity, not spelling. A committed case-fold test fails
   without it on darwin.
4. **O1 · OWED.** A worktree-export umask test, red when `copyTracked`'s Chmod is removed.
5. **O2 · OWED.** The redact `Names` comment and ADR text say that a short organisation id is written, and the
   "three credentials" comment is corrected.
6. **OWED by the owner's rule.** The second, independent NVIDIA review through Nightshift, not run today.

### Unrun

- **Listener tests.** Every `httptest` listener test in `cmd/livecheck`, `internal/generator`, `internal/guardrail`,
  `internal/pipeline`, `internal/query` and `internal/review` was refused by the sandbox:
  `panic: httptest: failed to listen on a port: listen tcp6 [::1]:0: bind: operation not permitted`. Those packages
  were re-run with the listener tests skipped (list in Ground). `cmd/livecheck` was not re-run green.
- **Short-secret unplug.** Not repeated by me; I rely on the master's measurement, labelled as theirs.
- **Narrow-on-open and `O_EXCL` unplugs.** Not run by mutant (Unplug).
- **Rust and Packagist behaviour.** No test was run on the reverted Rust/Packagist behaviour. B1 rests on the tree
  diff, not on a failing test, because the merge deleted those tests.
- **Case-fold on a case-sensitive volume.** Not measured; the probe skips there.
- **Live review.** A live review (sample mode off) was not run.

## The review status as filed

## REVIEW — security-piece-8-fixes
- state: recommendation-ready
- verdict: reject
- ref: _REVIEW_NOTES.md
- abstract: Reject (claude-code, claude-opus-5-5; the NVIDIA second review is owed). The leg's disk-property guards hold, and unplugging each one turns its test red.
  Three findings block: the absorb merge 5b89e3e deletes main's Rust adapter and Packagist fix ordering; ██████████████████████████ has no length floor; and pathguard compares paths case-sensitively, so state and trace land inside the target on darwin.
  Build and vet are clean. Listener tests are Unrun because the sandbox refuses the bind. Remedies are in _REVIEW_NOTES.md Conditions.
- artifacts:
    - _REVIEW_NOTES.md
    - _REVIEW_STATUS.md
- updated: 2026-10-24T09:20:00Z

## REDACTED

| what | times |
|---|---|
| file and line in Saphan Studio's private code | 2 |
| private credential or variable name | 3 |
| private workspace path | 1 |
