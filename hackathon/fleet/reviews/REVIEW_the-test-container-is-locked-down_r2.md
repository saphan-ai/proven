# REVIEW — the-test-container-is-locked-down · round 2

- verdict: accept-with-conditions
- reviewer: claude-code on macm5, claude-opus-5-5
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-23T19:08:29Z

## Summary

Re-review on claude-opus-5-5 of 07b901d..65dddaf. C1 ADDRESSED in code: tmpfs exec, fixed HOME/GOCACHE on scratch, read grant, 512m; O1, O2 and O4 addressed. 14 of 15 mutants killed; the survivor is the untested size-overflow guard (O7). The real container run is the master's only, since docker is not in this leg's class: CR1 asks for an independent run before the gate. O6 is new and owed: other host users can now read the staged checkout. The NVIDIA half is owed.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — the-test-container-is-locked-down

Reviewer: claude-code (`claude-opus-5-5`), named REVIEW leg `the-test-container-is-locked-down-review-r2`, 2026-10-24.
Re-review (R14) of `07b901d..65dddaf`, two commits: `ce40863` (round 2) and `65dddaf` (round 3).
`git rev-parse --short=7 HEAD` printed `65dddaf`. Profile: security.
Verdict: **accept-with-conditions**. C1 is ADDRESSED in the code, and every new guard is killed by its own test when
unplugged. But the real container run that proves C1 is the master's alone; this leg could not repeat it (Unrun). Law #53
therefore rules out a plain `accept`.

### Ground

All measured on 65dddaf, 2026-10-24, on this leg's host (darwin, no container runtime reachable):

1. `go build ./... && go vet ./...`: exit 0.
2. `go test -count=1 ./...`: **28 ok, 6 FAIL, 5 no test files**. This is the same split as round 1. All 6 FAILs
   (`cmd/livecheck`, `internal/generator`, `guardrail`, `pipeline`, `query`, `review`) are the one sandbox panic
   `httptest: failed to listen on a port: listen tcp6 [::1]:0: bind: operation not permitted`. All three changed
   packages are green here: `internal/testrun` ok, `internal/change` ok, and the new
   `TestMaterializedTreesAreReadableByOthers` in `internal/pipeline` PASS when run alone with `-run`.
3. Changed tests, run verbose: these PASS: `TestArgvScratchIsBoundedTmpfs`, `TestArgvPassesOnlyTheFixedEnvironment`,
   `TestRuntimeProcessEnvironmentIsExactlyTheNamedOnes`, `TestArgvIsOneReadOnlyMountWithNoNetworkAndNoEnv`,
   `TestMakeReadableGrantsReadAndSearchAndNoWrite`, `TestStagedCheckoutIsReadableByOthers`,
   `TestLoadHardeningSettingsAreReadAndRefused` and `TestContractW1bT2`.
   `TestWrittenGoTestRunsInTheContainer` and `TestContractW1bT10` SKIP (no runtime).
4. The diff touches 12 files, +392/−43. No `go.mod` or `go.sum` change.
5. These were read in full or in their changed region: `internal/testrun/container.go`, `internal/testrun/written.go`
   (`stageWritten`, `copyTree`, `MakeReadable`), `internal/pipeline/trees.go` (`materialize`, `loadLiveTarget`,
   `workArea`) and `internal/change/archive.go` (the export modes, and "gitArchive … does not emit .git").
6. The order's H1 instruction (`# REVIEW — …-review-r2`) contradicts the engine-injected channel grammar, which binds the
   verdict to the reviewed slug. I followed the engine (`# REVIEW — the-test-container-is-locked-down`).

### Differential

**C1 (blocking, round 1): ADDRESSED.** Each cause is checked against the fix:

- **(a) `/tmp` noexec → ADDRESSED.**
  - Fix: `container.go:356` is now `rw,exec,nosuid,nodev,mode=1777,size=<TmpSize>`, and the comment at `:351-355`
    now matches the argv.
  - Test: `TestArgvScratchIsBoundedTmpfs` (`container_hardening_test.go:37-67`) parses the option set. It requires
    `exec`, `nosuid`, `nodev` and `size=512m`, refuses `noexec`, and requires exactly one tmpfs. All four mutants
    (`noexec`, `exec` dropped, `nodev` dropped, `nosuid` dropped) are killed (Unplug).
- **(b) build cache off scratch → ADDRESSED, by option 2 of the round-1 remedy (fixed `-e`), the master's ruling.**
  - Fix: `containerEnv` (`container.go:70`) is `HOME=/tmp` and `GOCACHE=/tmp/go-build`, appended at `:358-360`.
    `/.cache` is dropped from `scratchDirs` (`:61`).
  - Tests: `forbidEnv` (`container_test.go:481-496`) requires the `-e` values to equal the fixed list exactly, and
    refuses `--env`, `--env-file`, `--env=`, `--env-file=` and `-e=`. `contractW1bEnvLeak`
    (`container_contract_test.go:421-440`) allows `-e` only with one of the two fixed values.
  - Mutants: changing the GOCACHE value, appending an inherited value, and appending a bare `-e NAME` are all killed,
    by 4 tests each.
- **(c) round 3, found by the master's real run: the private checkout root (`stat .: permission denied`) →
  ADDRESSED.**
  - Fix: `MakeReadable` (`written.go:201-224`) is called from `stageWritten` (`written.go:124`) and from
    `materialize` (`trees.go:113`).
  - Tests: `TestMakeReadableGrantsReadAndSearchAndNoWrite`, `TestStagedCheckoutIsReadableByOthers` and
    `TestMaterializedTreesAreReadableByOthers`.
- **(d) round 3: ENOSPC at 64m → ADDRESSED by the default `512m` (`container.go:49`).**
  - Sufficiency is a measurement, not a unit property. The only instrument is the master's real run. The argv test
    pins `size=512m`.

**The master's dated claims (2026-10-23, █████████████, Docker 29.8.0, golang:1.26), checked:**

- "`TestWrittenGoTestRunsInTheContainer` PASS 23.30s, `TestContractW1bT10` PASS 16.81s at 65dddaf": **UNMEASURED
  here** (no runtime; Unrun). I judge the evidence by what each test exercises:
  - `TestWrittenGoTestRunsInTheContainer` runs `go test` on a staged checkout. It needs all three of: exec on
    `/tmp`, a writable GOCACHE, and a checkout readable by uid 65534.
  - `TestContractW1bT10` builds and runs `dial.go` under `/tmp`.
  - A PASS of both cannot happen if any of the three causes is still present. So the pasted output, if genuine,
    discriminates.
- "Negative control: read grant removed → `stat .: permission denied`": **UNMEASURED here, but consistent.** Unplugging
  the grant at the unit level turns `TestStagedCheckoutIsReadableByOthers` red (mutant `stagenogrant`). It also turns
  `TestMaterializedTreesAreReadableByOthers` red (mutant `noop`).
- "a cold build measured at about 135 MB": UNMEASURED here.
- master-was-wrong: none of the claims I could measure is wrong.

**Owed items from round 1 that rode along:**
- **O1: ADDRESSED.** `settings_hardening_test.go:35` adds `{EnvUser, "0:1000"}`.
- **O2: ADDRESSED.**
  - `environment_test.go:22-40` requires PATH and HOME in git's environment.
  - `TestRuntimeProcessEnvironmentIsExactlyTheNamedOnes` requires the exact set
    `[DOCKER_HOST HOME PATH]`, so an empty child output fails.
- **O4: ADDRESSED for PIDs and scratch.**
  - Bounds: `maxPidsLimit=4096` and `maxTmpBytes=1<<30` (`container.go:54-55`), enforced at `:184` and `:191`, with
    test cases `4097`, `2147483647`, `2g`, `1000000g` and `99999999999999999999g`.
  - Still open, as in round 1: CPU, Memory and the upper uid. These are outside C1.

### Unplug

Instrument: mutant copies of `container.go` and `written.go`, with a selector constant in an added file. All are
overlaid with `go test -overlay` from the run's artifact directory, so nothing is written in the worktree.
- Command per mutant: `go test -count=1 ./internal/testrun`. The tree-grant mutants also run
  `./internal/pipeline -run 'Readable|Staged|Materialized'`.
- Sanity run with selector `none`: both packages ok.

| Mutant | Effect | Result |
|---|---|---|
| noexec | tmpfs `noexec` | killed: `TestArgvScratchIsBoundedTmpfs` |
| noexecword | `exec` dropped | killed: `TestArgvScratchIsBoundedTmpfs` |
| nonodev | `nodev` dropped | killed: `TestArgvScratchIsBoundedTmpfs` |
| nonosuid | `nosuid` dropped | killed: `TestArgvScratchIsBoundedTmpfs` |
| gocache | `GOCACHE=/.cache/go-build` | killed: `…NoNetworkAndNoEnv`, `…OnlyTheFixedEnvironment`, `ContractW1bT2`, `TestRunReadsBothCheckouts…` |
| leak | extra `-e GOFLAGS=$GOFLAGS` | killed: same four |
| leakbare | extra `-e ███████████████████████` (inherit form) | killed: same four |
| pidsmax | upper PIDs bound removed | killed: `TestLoadHardeningSettingsAreReadAndRefused` |
| tmpmax | upper scratch bound removed | killed: `TestLoadHardeningSettingsAreReadAndRefused` |
| stagenogrant | `MakeReadable` call removed from `stageWritten` | killed: `TestStagedCheckoutIsReadableByOthers` |
| noop | `MakeReadable` does nothing | killed: all three grant tests (testrun + pipeline) |
| nosearch | directories get no search bit | killed: all three grant tests |
| writebit | grant adds `g+w,o+w` | killed: all three grant tests |
| rootonly | grant on the root only | killed: `TestMakeReadable…` only; the staged and materialized tests stay green |
| **overflow** | **`sizeBytes` overflow guard removed** | **SURVIVES** (O7) |

- **The overflow guard is correct.** I probed it with an overlaid test, `TestW28R2WrappingTmpSizeRefused`. On the
  real code, `17179869184g`, `17592186044416m` and `18014398509481984k` (each 2^64 bytes, which wraps to 0) are refused
  by name, and `1g`/`1024m` are accepted. Under the mutant, all three are accepted (`17179869184g accepted`, FAIL).
- remark: `rootonly` survives the two path tests because `copyTree` and the export already write 0755 directories and
  0644 files. Only the `MkdirTemp` root needed the grant. This is an observation about what those two tests can see,
  not a gap: the unit test kills it.

### Closure

The closure members are the change's code and the model's command, which run inside the container, and the operator's
`PROVEN_TEST_*` settings.

- **Question 2: does the read grant expose anything beyond the checkout? To the container: no.**
  - The container sees only the one `-v <tree>:/checkout:ro` mount and its own image.
  - The trees hold the tracked files only. `archive.go` notes that `gitArchive` does not emit `.git`, so no remote URL
    or credential helper config is included. Tracked symlinks are skipped (`trees.go:86`, `copyTree` at
    `written.go:154`). `MakeReadable` touches regular files and directories only.
  - `filepath.Walk` uses `Lstat`, so `Chmod` never follows a link out of the tree.
- **Question 2: is anything in the mounted trees writable by the container user? No.**
  - The mount is `:ro` (`container.go:362`).
  - No group or other write bit is created (export and copy modes 0755/0644), and `MakeReadable` adds none.
    The `writebit` mutant is killed.
  - The container user is uid 65534, which is not the owner. The only writable place is the bounded `/tmp` tmpfs.
- **Question 3: can either fixed `-e` value carry anything from Proven's environment or the change? No.**
  - Both are Go string literals in `NAME=value` form (`container.go:70`). Docker reads a value from the client's
    environment only for the bare `-e NAME` form, and that form is refused by `forbidEnv` (mutant `leakbare` killed).
  - No setting or request field reaches `containerEnv`.
  - The runtime process itself still gets only PATH, HOME and the daemon names (`:540`).
  - remark: `containerEnv` is a package `var`, not a constant. Only code in the package could change it; the tests pin
    its value.
- **Host side (new reach, outside the container): the grant opens the staged checkout root to other local accounts.**
  - `stageWritten` uses `os.MkdirTemp("", …)` (`written.go:110`), which is `/tmp` on Linux when `TMPDIR` is unset.
    Before round 3, the 0700 root blocked other users. Now the root is 0755 (measured here under the run's TMPDIR:
    parent 755, root 755).
  - So while a run lasts, any local user on a shared host can read the copy of the code under review. That is the
    checkout itself, not a Proven key. Owed as O6.
- **Settings:** the upper bounds now close PIDs and scratch. The overflow path is refused (probe).

S2: each closed cause is reproduced by an argv or mode test that fails when the fix is removed. No payload was used.

### Remainder

Owed, not blocking. Each is outside C1 or a statement (R14: never another round).

- **O6 (owed, host disclosure).** `written.go:110` and `:124`, `trees.go:90` and `:113`: the read grant makes the
  staged and materialized roots readable by every local user, under `/tmp` on Linux.
  - Remedy: create the tree one level below a private 0700 parent (`MkdirTemp`, then a child directory that gets the
    grant), and mount the child. The daemon resolves the bind source as root, so a private parent does not block the
    mount, and inside the container the parent is not visible. Alternatively, state the exposure in
    `docs/CONFIGURATION.md`.
  - Not measured: whether rootless Podman or Docker Desktop resolve the bind source the same way.
- **O7 (owed, test gap).** No case wraps the 64-bit size arithmetic, so the overflow guard at `container.go:233` can be
  deleted while the suite stays green. Add `{EnvTmpSize, "17179869184g"}` to `settings_hardening_test.go:39`.
- **O8 (owed, paired path, R13).** `loadLiveTarget` uses prepared `base/` and `change/` trees as they are
  (`trees.go:49-51`), with no grant.
  - When the Prover runs a command with no written test (`req.Content == ""`, `container.go:397`), those trees are
    mounted directly. If they are owner-only, both sides fail on `stat .` with a non-125 exit, so they are read as
    `failed`. `PassesOnBase` is false, so nothing false is proven, but the Prover again proves nothing for that layout.
  - Proven should not chmod an operator's tree. The remedy is to state the mode requirement, or to stage a copy.
- **O9 (owed, statement).** The default scratch (`512m`, `container.go:49`) equals the default memory limit
  (`defaultMemory = "512m"`, `:36`). tmpfs pages are charged to the container's memory cgroup. That is kernel
  behaviour I did not measure here.
  - Consequence: the scratch can never actually be filled under the default memory limit. A build near the limit is
    OOM-killed (137). `side` classifies 137 as `kindFail` (`:471-472`), the same class it gives ENOSPC.
  - Either state that the memory bound is the effective scratch bound, or raise the default memory.
- **Order-mandated (R12): the ADR-0025 point 3 amendment** ("no -e" → two fixed values) is owed after the merge, per
  the master's ruling. The comment at `container.go:63-69` already says it amends the ADR.
- **O3 and O5 (round 1): untouched by this diff and still owed.** O3: `runtimeEnvNames` is Docker-only. O5:
  `/dev/shm` is unbounded by Proven.
- Declined to judge (R11):
  - whether `golang:1.26` holds every module a written test needs offline: GOPATH comes from the image's `/go` on a
    read-only root, and `HOME=/tmp` does not move it;
  - supplementary groups from the image;
  - licences.

### Profile

security. S1, change → container:
- spoofing: n/a
- tampering: argv comes from validated settings and constants (holds)
- disclosure: container env is only the two literals (holds); the runtime process env is exact (holds); the trees carry
  no `.git` (holds); host-side readability is new (O6)
- denial: PIDs ≤4096 and scratch ≤1g, bounded and tested; scratch inside the memory limit (O9)
- escalation: `nosuid`, `nodev` and `exec` on scratch; non-root; caps dropped; no-new-privileges; read-only root (holds)

S3: every new guard was removed once. 14 of 15 mutants were killed; the survivor is a test gap with a correct guard (O7).

S4: no dependency delta. No secrets (test markers only). No new destructive path: `MakeReadable` only adds bits, and
it skips links. Hidden characters: I did not re-run the Cf/bidi rune scan this round (Unrun).

### Focus

1. **Is C1 closed?** In the code, yes: (a), (b), the round-3 read grant, and the scratch size. Every new guard is tested
   and killed when unplugged. The only proof that a real Go test now runs is the master's pasted run, which this leg
   could not repeat (Unrun, condition CR1).
2. **Does round 3 weaken a property?** Not property 1 or property 2: the container sees only the `:ro` checkout, and
   nothing there is writable by its user. It does widen host-side read access to the checkout (O6).
3. **The two fixed `-e` values:** literals in `NAME=value` form. They can carry nothing from Proven or the change, and
   the inherit form is refused by tests.
4. **Build, vet and test:** green except 6 packages that cannot open a listener in this sandbox, all outside the
   diff's write-set. The new pipeline test passes when run alone.

### Conditions

**CR1 (before the gate): an independent real-container run.**
- Owner: the master, a second leg on a host with a runtime, or the owed NVIDIA review.
- Act: re-run `PROVEN_TEST_DOCKER=1 go test -count=1 -v -run 'TestWrittenGoTestRunsInTheContainer|TestContractW1bT10' ./internal/testrun`
  at `65dddaf` with `golang:1.26`, and paste the output.
- Why: the C1 closure rests on one hand's measurement. This leg's permission class does not hold `docker` (see
  Unrun).
- If it is red, C1 is NOT ADDRESSED and this verdict becomes `reject`.

**CR2 (after the merge, order-mandated): the ADR-0025 point 3 amendment for the two fixed `-e` values.**

O6–O9 are owed and never open another round. O3 and O5 stay owed from round 1.

The NVIDIA half of the owner's two-review rule is **owed**. This review ran alone, because the route is down (order).

### Unrun

- **A real hardened container run, and the master's negative control.** Not run: `docker` is not in this leg's
  permission class. The hook refusal printed the class as
  `[saphan doc check go git ls cat grep find wc head tail diff pwd cut sort]`.
  - Setting `PROVEN_TEST_DOCKER=1` as a command prefix was also refused in round 1 (same class).
  - So these remain UNMEASURED here: `TestWrittenGoTestRunsInTheContainer`, `TestContractW1bT10`, the 135 MB figure,
    and the `stat .: permission denied` control.
- **Full `go test` of 6 packages** (`cmd/livecheck`, `internal/generator`, `guardrail`, `pipeline`, `query`,
  `review`). They panic with `listen tcp6 [::1]:0: bind: operation not permitted`, so their full result is
  UNMEASURED here. Only `TestMaterializedTreesAreReadableByOthers` was run in `pipeline`.
- **Hidden-character scan of the diff.** Not repeated this round.
- **tmpfs-in-memory-cgroup accounting (O9).** Kernel behaviour, not measured.
- **Independent NVIDIA review.** Owed (route down).

## The review status as filed

## REVIEW — the-test-container-is-locked-down
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: Re-review on claude-opus-5-5 of 07b901d..65dddaf. C1 ADDRESSED in code: tmpfs exec, fixed HOME/GOCACHE on scratch, read grant, 512m; O1, O2 and O4 addressed. 14 of 15 mutants killed; the survivor is the untested size-overflow guard (O7). The real container run is the master's only, since docker is not in this leg's class: CR1 asks for an independent run before the gate. O6 is new and owed: other host users can now read the staged checkout. The NVIDIA half is owed.
- artifacts:
    - _REVIEW_NOTES.md
    - _REVIEW_STATUS.md
- updated: 2026-10-24T00:00:00Z

## REDACTED

| what | times |
|---|---|
| machine name | 1 |
| private credential or variable name | 1 |
