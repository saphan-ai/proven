# REVIEW — the-test-container-is-locked-down · round 1

- verdict: reject
- reviewer: claude-code on macm5, claude-opus-5-5
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-23T12:12:18Z

## Summary

claude-code claude-opus-5-5 named review of 21af174...07b901d. Both security properties hold: argv flags and explicit child env, tests pass, 11/12 mutants killed, the three in-scope original reviewer tests now PASS.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — the-test-container-is-locked-down

Reviewer: claude-code (`claude-opus-5-5`), named REVIEW leg `the-test-container-is-locked-down-review`, 2026-10-23.
Reviewed: `21af174...07b901d`, one commit (`git rev-parse --short=7 HEAD` printed `07b901d`). Profile: security.
Verdict: **reject**. The security properties hold, and the tests prove them. But on the default image the hardened
container can no longer run a Go test, so after this change the Prover proves nothing. That is one blocking
finding with two causes, and the fix is small.

### Ground

Searches and runs, all on 07b901d, 2026-10-23:

1. `grep -rn -e exec.Command -e os.StartProcess -e syscall.Exec -e '.Env = ' --include=*.go .` finds two
   production process starts. Each has an explicit environment:
   - `internal/change/between.go:17-21` (`gitCommand`, `cmd.Env = childenv.Minimal()`). All seven git call sites
     in the package now go through it (`:26,42,111,147,189,204,273`).
   - `internal/testrun/container.go:487-488` (`commandOutput`, `cmd.Env = childenv.Minimal(runtimeEnvNames...)`).
     It serves both the `run` and the `rm -f` calls, through `invoke` (`:471-476`, `:438`).
   Every other hit is in a `_test.go`, `demo/live_test.go` (sets its own `offlineEnv()`) or
   `conformance/meterless/run_test.go`. **Q3: no process Proven starts inherits the whole environment.**
2. `go build ./... && go vet ./...` exit 0. `go test -count=1 ./...` gives **28 ok, 6 FAIL, 5 no test files**.
   The 6 failures (`cmd/livecheck`, `internal/generator`, `guardrail`, `pipeline`, `query`, `review`) are all the
   same panic: `httptest: failed to listen on a port: listen tcp6 [::1]:0: bind: operation not permitted`. This
   sandbox refuses local listeners. With `-skip 'Live|Server|...'`, `internal/pipeline` passes. The changed
   packages `internal/testrun`, `internal/change` and `internal/childenv` all pass.
3. **Property 1, from the code** (`container.go:292-314`):
   - `--network none` (`:294`)
   - `--user <cfg.User>` (`:297`), default `65534:65534` (`:40`)
   - `--pids-limit` (`:298`), `--cap-drop ALL` (`:299`), `--security-opt no-new-privileges` (`:300`),
     `--read-only` (`:301`)
   - two `--tmpfs` mounts with `size=<TmpSize>` (`:303-308`)
   - `-v host:/checkout:ro` (`:310`)
   - no `-e`/`--env` anywhere.
   Proved by `TestArgvHardensTheContainer` (`container_hardening_test.go:18-27`, asserts flag and value),
   `TestArgvScratchIsBoundedTmpfs` (`:32-48`) and the existing `TestContractW1bT1..T5`.
4. **Property 2, from the code:**
   - The container has no `-e` flag, so it receives no environment.
   - The runtime process gets PATH, HOME and four daemon-locating names (`container.go:56`).
   - git gets PATH and HOME (`childenv.go:12`).
   Proved by `TestRuntimeProcessEnvironmentOmitsSecrets` (`container_hardening_test.go:53-63`),
   `TestGitEnvironmentOmitsSecrets` (`environment_test.go:11-22`, a git alias `!env` really runs under git) and
   `TestMinimalFromPassesOnlyNamedVariables`.
5. **Q2, settings** (`settings.go:75-103`, `container.go:158-191`):
   - User must match `^\d+:\d+$` with uid>0 and gid>0. `0:0`, `00:00`, `0:1000`, `1000:0` and `root` are all
     refused (my probe `TestW28RootUIDAnyGroupRefused` PASS).
   - Pids must be >0: `0`, `-1` and `many` are refused.
   - TmpSize uses the memory grammar with n>0: `0` and `lots` are refused.
   - Every refusal names its variable (`envName`, `settings.go:159-164`).
   **No setting can produce root, an unlimited process count or an unbounded tmpfs.** The upper end is open (see
   Remainder).
6. **Original reviewers' tests, run against 07b901d.** Each test is overlaid with `go test -overlay`. Its source
   is copied verbatim from `████████████████████████`, except one call in D3 (below). All inputs live outside the
   worktree, under the run's artifact dir `scratch/`.

   | Test (origin) | Scope | At 90d6951 | At 07b901d |
   |---|---|---|---|
   | `TestReviewProcessInsideTheContainerIsBounded` (p1-claude D1) | **W28** | FAIL | **PASS** |
   | `TestReviewContainerBoundsProcessAndFilesystem` (p1-codex F2) | **W28** | FAIL | **PASS** |
   | `TestReviewSecretNeverReachesRuntimeProcess` (p2-codex) | **W28** | FAIL | **PASS** |
   | `TestReviewRunnerRefusesUnboundedCommand` (p1-codex F1) | other wagon | FAIL | FAIL (`Found:true`) |
   | `TestReviewModelTextDoesNotBecomeTheContainerCommand` (p1-claude D2) | other wagon | FAIL | FAIL |
   | `TestReviewExportWorktreeStaysInsideTheTarget` (p1-claude D3) | other wagon | FAIL | FAIL |
   | `TestReviewCheckoutThatIsALinkIsNotMounted` (p1-claude D4) | other wagon | FAIL | FAIL |
   | `TestReviewLinkedTreesAreNotPreparedTrees` (p1-claude D4) | other wagon | FAIL | FAIL |
   | `TestReviewConfiguredSecretNeverReachesTraceFromReceipt` (p2-codex) | other wagon | FAIL | FAIL |

   D3 was adapted only because `ExportWorktree` now returns two values (`_, err :=`). The livecheck tests from
   p2-codex were not run (that package panics on the listener refusal here).
7. **Q5, comments and naming.**
   - ADR-0027 holds: new package comment (`childenv.go:1-4`); doc comments on `Minimal`, `MinimalFrom` and
     `gitCommand`; each new test states its property and its mutated field.
   - One comment is false (B1, `container.go:304-305`).
   - The commit subject and body, the comments and the test names carry no finding text. Test names state
     properties ("OmitsSecrets", "HardensTheContainer").
   - Hidden-character scan: a Go rune scan over all 10 diff files found no Cf, bidi or BOM rune. Its positive
     control (a planted U+202E) was caught. Only one non-ASCII rune is present, in `docs/CONFIGURATION.md`.

### Differential

**B1 — BLOCKING. On the default image, the hardened container cannot run `go test`.** Location:
`internal/testrun/container.go:48-51,303-308`. Classed blocking because the Prover silently proves nothing. The
change side and the base side both "fail", so `PassesOnBase` is false. No refutation can stand, and no service is
marked down. Before this change the same run worked (root, writable layer). Two causes, and either one alone is
fatal:

- **(a) `/tmp` is noexec.** The option string is `rw,nosuid,mode=1777,size=64m` and does not name `exec`. Docker
  and Podman both mount `--tmpfs` with `noexec,nosuid,nodev` unless the option string overrides it. Podman's
  manual states those defaults, and moby merges them in its tmpfs mount setup. `go test` links the test binary
  under `$TMPDIR` or `/tmp` and runs it from there, so the run ends in
  `fork/exec /tmp/go-build…/x.test: permission denied`. The comment at `:304-305` ("exec stays allowed") states
  the opposite of what the argv does. My probe `TestW28ScratchTmpNamesExec` **FAILS** at 07b901d:
  `/tmp tmpfs options ["rw" "nosuid" "mode=1777" "size=64m"] do not name exec`.
- **(b) The build cache is mounted where `golang:1.26` does not look.** `scratchDirs` assumes the user has no
  home, so HOME is `/`. The default image `golang:1.26` is Debian-based, and its `/etc/passwd` maps `65534` to
  `nobody` with home `/nonexistent`. The runtime sets HOME from that entry, so Go's `GOCACHE` defaults to
  `/nonexistent/.cache/go-build`. That path is on the read-only root, so the build fails with
  `failed to initialize build cache … read-only file system`. The `/.cache` assumption fits an image where 65534
  has no passwd entry, or Alpine (`nobody` home `/`). It does not fit the shipped default. The comment at
  `:38-39` ("exists in common images") contradicts the one at `:49-50` ("an identity that has none").

Both causes come from documented runtime and image behaviour. **Neither was measured live** (see Unrun). The
leg's own docker-backed tests would decide it: `TestWrittenGoTestRunsInTheContainer` (`written_test.go:148`,
`go test -count=1 -run TestAdd .`) and `TestContractW1bT10` (`go run dial.go`, which also builds under `/tmp`
and execs from there).

**Master's dated claims (2026-10-23), re-measured:**
- "34 packages ok": **UNMEASURED here.** 28 ok, plus 6 that cannot run in this sandbox for the listener reason.
  28+6=34, so the claim is consistent but not confirmed.
- "Removing `--cap-drop ALL` turns `TestArgvHardensTheContainer` red": **holds.** Mutant `capdrop` gave
  `--- FAIL: TestArgvHardensTheContainer`.
- "Letting the runtime process inherit the environment turns `TestRuntimeProcessEnvironmentOmitsSecrets` red":
  **holds.** Mutant `runtimeenv` gave that test FAIL.
- "No container runtime on the master's host, so a real hardened run is UNRUN": holds here too.
- master-was-wrong: none of the measured claims. The leg's correction (git calls are in `between.go`) holds.

### Unplug

Mutants are an env-free selector constant overlaid on a copy of `container.go` (or `childenv.go`). The leg's
own tests were run against each, with `go test -count=1 ./internal/testrun` (plus `./internal/change` and
`./internal/childenv` for `inherit`). The sanity run with no mutant switched on was `ok`.

| Mutant | Effect | Result |
|---|---|---|
| capdrop | drop `--cap-drop ALL` | killed: `TestArgvHardensTheContainer` |
| user | drop `--user` | killed: `TestArgvHardensTheContainer` |
| pids | drop `--pids-limit` | killed: `TestArgvHardensTheContainer` |
| nnp | drop `no-new-privileges` | killed: `TestArgvHardensTheContainer` |
| readonly | drop `--read-only` | killed: `TestArgvHardensTheContainer` |
| tmpnosize | tmpfs without `size=` | killed: `TestArgvScratchIsBoundedTmpfs` |
| runtimeenv | runtime inherits env | killed: `TestRuntimeProcessEnvironmentOmitsSecrets` |
| inherit (childenv) | `Minimal` returns `os.Environ()` | killed: `TestGitEnvironmentOmitsSecrets` and `TestRuntimeProcessEnvironmentOmitsSecrets`; `TestMinimalFromPassesOnlyNamedVariables` stays green (it tests `MinimalFrom` only) |
| gid0 | accept gid 0 | killed: `TestLoadHardeningSettingsAreReadAndRefused` |
| pidsguard | accept pids 0 | killed: `TestLoadHardeningSettingsAreReadAndRefused` |
| tmpguard | skip TmpSize validation | killed: `TestLoadHardeningSettingsAreReadAndRefused` |
| **uid0** | **accept uid 0 when gid ≠ 0** | **SURVIVES** (O1) |

remark: the env tests have no positive control. An empty child output passes them exactly as a clean one does.
I measured the positive half separately (`TestW28RuntimeEnvironmentHasPath` PASS, names `[PATH HOME DOCKER_HOST]`).

### Closure

The closure members are the code of the change and the model's command, which run inside the container, and the
operator's `PROVEN_TEST_*` settings.

- **From inside the container:** no network, and no Proven environment. The checkout is read-only and the root
  is read-only. The user is non-root with every capability dropped and no privilege gain. Process count and the
  two scratch mounts are bounded (by argv).
- **Writable places:** `/tmp` and `/.cache` (tmpfs, bounded) are the declared ones. Also writable, and not
  declared: `/dev/shm` (O5) and `/dev/mqueue`. `/etc/hosts`, `/etc/hostname` and `/etc/resolv.conf` are mounted
  read-only once `--read-only` is set (runtime behaviour, not measured).
- **Settings:** they cannot produce root, zero or negative limits, or an unsized tmpfs. Every refusal names its
  variable.
- **git:** gets PATH and HOME only. The host's `~/.gitconfig` is still read (HOME), but that is the operator's
  config, not the change's.
- **Runtime process:** its output still reaches the trace, but it can no longer hold a Proven key.

S2: the blocking finding is reproduced by a failing argv test, which covers (a) only. (b) has no test that can
run without a runtime. Neither cause uses a payload.

### Remainder

- **O1 (owed, test gap).** `settings_hardening_test.go:35-37` has no case with uid 0 and a non-zero gid, so the
  `uid > 0` guard at `container.go:190` can be deleted and the suite stays green. Add `{EnvUser, "0:1000"}`.
- **O2 (owed, instrument).** `container_hardening_test.go:56-63` and `environment_test.go:16-21` should assert
  that `PATH=` is present. As written, a child that printed nothing passes.
- **O3 (owed, compatibility).** `runtimeEnvNames` (`container.go:56`) is Docker-only, and three setups lose what
  they need to reach the daemon:
  - Podman remote: `CONTAINER_HOST`, `CONTAINERS_CONF`
  - Docker over TLS: `DOCKER_TLS_VERIFY`, `DOCKER_CERT_PATH`
  - `ssh://` `DOCKER_HOST`: `SSH_AUTH_SOCK`
  On those hosts every run would exit 125, the side reads "could not run", and the Prover is not marked down.
  Either name the supported setups in `docs/CONFIGURATION.md` or add the non-credential names. The env-free
  default (local socket) is unaffected.
- **O4 (owed, statement).** There is no upper bound on the new limits. Measured: `PidsLimit=2147483647`,
  `TmpSize=1000000g` and `User=4294967295:4294967295` all give Validate → `<nil>`. CPU and memory had the same
  open end before this change. tmpfs pages are charged to the container's memory cgroup, so `--memory` caps
  them in practice (swap aside). The operator sets these values; this is not a crossing.
- **O5 (owed).** `/dev/shm` stays writable. Its size comes from the daemon's default (64m on Docker), not from
  Proven. Q4 asks whether the scratch is the only writable path; strictly, it is not. Pass `--shm-size` from
  `TmpSize`, or state it.
- The `container_contract_test.go:504-509` change skips `--tmpfs` values in the host-path check. That is right,
  because a tmpfs value names a path inside the container. The pair is still judged: the mount-count test T3
  still counts only `-v`.
- Declined to judge (R11): supplementary groups the runtime adds from the image's `/etc/group` for a uid that has
  a name there (depends on the image); whether `golang:1.26` holds the module cache a test needs offline; licences.

### Profile

security. S1, change → container:
- spoofing: n/a
- tampering: argv is settings-only, validated (holds)
- disclosure: env (holds, both child kinds)
- denial: pids and scratch bounded, shm by daemon default
- escalation: non-root, caps dropped, no-new-privileges, read-only root (holds)

S1, Proven → git: disclosure holds. S3: each guard was removed once (table under Unplug), and 11 of 12 mutants
were killed. S4: no `go.mod` or `go.sum` change in the diff, so no dependency delta. No secrets in the diff (test
markers only). No destructive path added. Hidden-character scan clean, with a positive control.

### Focus

1. **Property 1** holds at the argv: `container.go:294-310`, tests `TestArgvHardensTheContainer`,
   `TestArgvScratchIsBoundedTmpfs` and `TestContractW1bT1..T5`. **But the bounded container cannot do the
   Prover's job on the default image (B1).**
2. **Property 2** holds: `container.go:488`, `between.go:19`, `childenv.go:12-31`; tests
   `TestRuntimeProcessEnvironmentOmitsSecrets`, `TestGitEnvironmentOmitsSecrets` and
   `TestMinimalFromPassesOnlyNamedVariables`. Unplugging each one turns its test red.
3. **Q2** holds for root, zero, negative and unsized values; the uid-only root case is untested (O1).
4. **Q3** holds: two production process starts, both explicit.
5. **Q4** partly holds: the declared scratch is bounded, and `/dev/shm` is not declared (O5).
6. **Q5** holds, except the false `exec` comment (B1).
7. **Q6:** build and vet green; tests 28 ok, 6 not runnable here.

### Conditions

**C1 (blocking, B1).** Remedy owner: an executor on claude-code (security class).
- (a) Add `exec` (and `nodev`) to the `/tmp` tmpfs options at `container.go:307`, fix the comment at `:304-305`,
  and assert `exec` on `/tmp` in `TestArgvScratchIsBoundedTmpfs`.
- (b) Make the build cache land on scratch for the default image. The master or owner chooses one way, because
  each touches a decision:
  1. Mount the scratch at the home the image gives the user (`/nonexistent` for `golang:1.26`).
  2. Hand the container a fixed, non-secret `HOME=/tmp` or `GOCACHE=/tmp/...`. This amends ADR-0025's "no `-e`"
     contract (T2).
  3. Default the user to a uid with no passwd entry in the image.
- Proof owed: `PROVEN_TEST_DOCKER=1 go test -count=1 -run 'TestWrittenGoTestRunsInTheContainer|TestContractW1bT10' ./internal/testrun`
  passes on a host with a runtime and `golang:1.26`, and its output is pasted. A re-review judges only the C1
  diff and that output (R14).

**O1–O5 are owed, not blocking.** O1 and O2 are small enough to ride along with C1.

The NVIDIA half of the owner's two-review rule is **owed**. This review ran alone, because the route is down
today (order).

### Unrun

- **A real hardened container run.** Not run: `docker` is not in this leg's permission class
  (`saphan doc check go git ls cat grep find wc head tail diff pwd cut sort`). Listing the runtime path was
  refused: `ls in '/Applications/Docker.app/Contents/Resources/bin/docker' was blocked`. So B1 (a) and (b), and
  the runtime's handling of `/dev/shm` and `/etc/*` under `--read-only`, are read from runtime and image
  documentation, not measured. The two `PROVEN_TEST_DOCKER=1` tests were not run.
- **Full `go test ./...` in 6 packages.** Not run: they panic with
  `listen tcp6 [::1]:0: bind: operation not permitted`. Their full result is UNMEASURED here.
- **p2-codex's `cmd/livecheck` review tests.** Not run, for the same reason (other wagon in any case).
- **Mutant selection by environment variable.** Refused by the hook: `the program "REVIEW_MUTANT=capdrop" is NOT
  listed`. I replaced it with an overlaid constant file, which is a different instrument with the same effect on
  the code, so I did not rebuild a denied act.
- **Independent NVIDIA review.** Owed (route down).

## The review status as filed

## REVIEW — the-test-container-is-locked-down
- state: recommendation-ready
- verdict: reject
- ref: _REVIEW_NOTES.md
- abstract: claude-code claude-opus-5-5 named review of 21af174...07b901d. Both security properties hold: argv flags and explicit child env, tests pass, 11/12 mutants killed, the three in-scope original reviewer tests now PASS.
  BLOCKING B1: on golang:1.26 the hardened container cannot run go test. The /tmp tmpfs is noexec (no `exec` option), and nobody's HOME is /nonexistent, so the build cache is not on scratch.
  Owed O1-O5: uid-0 case untested, env tests lack a positive control, Docker-only runtime env names, open upper bounds, /dev/shm undeclared.
  Unrun: a real container run (docker not in this class), 6 packages blocked by sandbox listeners, the NVIDIA half.
- artifacts:
    - _REVIEW_NOTES.md
    - _REVIEW_STATUS.md
- updated: 2026-10-23T13:05:00Z

## REDACTED

| what | times |
|---|---|
| private workspace path | 1 |
