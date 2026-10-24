# Adapters

An adapter is how Proven grows without changing its core. A language adapter reads one ecosystem's files and versions ([ADR-0038](adr/ADR-0038-a-language-is-one-small-adapter.md)); a tool adapter turns a deterministic tool's run into evidence ([ADR-0040](adr/ADR-0040-a-tool-run-is-evidence.md)). Each has one worked example below.

## Adding a language


One language is one adapter ([ADR-0038](adr/ADR-0038-a-language-is-one-small-adapter.md)). The adapter finds the language's files, lists the packages a change adds or raises, names the ecosystem as OSV names it, and compares two versions by that ecosystem's rules. OSV ranges, the known-exploited list, EPSS and the review are shared. A new language does not change them.

The type is `Adapter` in `internal/manifest`. It is the same type as `Reader`. The dependency check runs `Readers()`. A version lands in a range through `InRangeEcosystem`, which calls the adapter's `Compare` and `Explain`.

The languages are Go, JavaScript, Python, PHP and Rust.

### What you add

1. A comparer for the ecosystem's versions, with tests for equal, lower, higher, and for a string the comparer refuses. Put the refusal in a sentence. Do not guess a version for a string you refuse.
2. A manifest reader and a lock reader. Each is an `Adapter`. Both use the same ecosystem name, the same `Compare` and the same `Explain`. The lock reader sets `Lock` and `NoteErr`, so a lock file that cannot be parsed is a trace note and not a failed review.
3. `Match` accepts the file's own name at any depth, and refuses a backup such as `name.bak`.
4. `Diff` reports a package the change adds, and a package whose version went up. A lower version is not a finding. A pair that does not order is still reported, so the trace can say why.
5. One line in `Readers` for each reader.
6. The ecosystem name in the `osv-by-package` query (`internal/querytpl`), and the package name's format. Update [QUERIES.md](QUERIES.md) in the same change: a test holds that page to the list.
7. The files and the version rules in [OSV_TABLE.md](OSV_TABLE.md).
8. A sample scenario under `testdata/fixture/` and a section in [SAMPLE_MODE.md](SAMPLE_MODE.md). The scenario's test fails when the adapter is taken out of `Readers`.

`Compare` on the adapter and the order `Diff` uses are the same function. A raise and a range then cannot disagree.

### Worked example: PHP

Composer is the first language added this way. The OSV ecosystem name is `Packagist`.

| | manifest | lock file |
|---|---|---|
| constructor | `Composer` | `ComposerLock` |
| file | `composer.json` | `composer.lock` |
| what it reads | `require` and `require-dev` | `packages` and `packages-dev` |
| `Lock` | no | yes |
| `NoteErr` | no | yes |

A platform package is not a Packagist package. `php`, `hhvm`, `ext-*`, `lib-*`, `php-*`, `composer-plugin-api` and `composer-runtime-api` are named in the trace and not queried. The package name is stored in lower case. `Monolog/monolog` and `monolog/monolog` are one package.

The version comparison is the subset in `internal/composerver`. One to four numeric components, missing components equal to zero. A leading `v` is ignored. Text after `+` is ignored. Stability, from lowest to highest: dev, alpha (`a`), beta (`b`), RC (`rc`), stable, patch (`p`, `pl`, `patch`). A constraint and a branch name (`dev-main`, `1.0.x-dev`) are refused with a sentence. The string in the file is not rewritten. [OSV_TABLE.md](OSV_TABLE.md) records the same subset.

When `composer.json` and `composer.lock` in one directory both name a package, the lock file's version is the one compared, and the finding names the lock file.

The sample is `testdata/fixture/osv-php-composer`. It raises `monolog/monolog` from 2.0.0 to 2.9.1. 2.9.1 is inside one sample range and outside another. The finding names `composer.lock`.

`composer.json` and `composer.lock` are security-class path rules, the same as the other dependency files. A change that touches either is a security review.

### What stays shared

You do not add a new query, a new range rule, or a new place in the report. `InRangeEcosystem` already asks the adapter. An ecosystem that is not in `Readers` is not compared, and the trace says so.


## Adding a tool

### A tool is a third kind of evidence

[ADR-0040](adr/ADR-0040-a-tool-run-is-evidence.md). A finding can be proved by a test, by a query, or by a tool. A tool proves a finding when it reports a result on the change and does not report that result on the base. The evidence records the tool, its version, the exact command, and the output that differs.

The proof level is `proved-by-tool`. It sits between a test and a query ([ADR-0044](adr/ADR-0044-every-finding-names-its-level-of-proof.md)).

A result the base already had is not evidence. A tool that does not answer is named, and it is not counted as a clean run. A run past its time limit is that kind of miss.

### What an adapter implements

One adapter is one tool. It answers four questions:

1. **Which files.** `Applies` is true when the changed paths are ones this tool can read. The race detector applies to a changed Go file. A change with no Go file does not run it.
2. **The command.** `Argv` returns the words Proven runs. There is no shell. A path that would mean something else to the tool, such as `..` or `...`, is refused and nothing is started.
3. **The output.** `Read` classifies one side. For the race detector, the line `WARNING: DATA RACE` means a race. A test that fails without that line is not a race. The spelling is the one in the race runtime of Go 1.26.
4. **The comparison.** The runner does this, not the adapter. The change runs first. Evidence is a result on the change and not on the base.

The runner lives in `internal/tool`. The sample implementation replays a recorded outcome and starts no container. The live implementation runs the command in the Prover's container.

### The container

The command is the tail of the Prover's run line ([ADR-0025](adr/ADR-0025-the-prover-runs-the-change-with-no-network-and-no-secret.md)): no network, no secret, a read-only checkout, a CPU limit and a memory limit. The same argv builder builds that line for a test and for a tool. A checkout that is a symbolic link is not mounted.

A real run needs a container runtime on the host. One test starts that run, and it is skipped unless `PROVEN_TEST_DOCKER` is `1`. The other tests check the run line and replay recorded output.

### The time limit

`PROVEN_TOOL_TIMEOUT` is the limit for one side, the base or the change. Empty means `120s`. A value that is set and is not a Go duration greater than zero is refused before a run, and it is not replaced by the default ([ADR-0013](adr/ADR-0013-no-hardcoded-settings.md)). Sample mode does not read the variable.

The limit is the tool's own. It does not change the limit of a test run. A side that passes the limit is named `timeout` and is not evidence, even when its output already contains a result.

### The race detector

The first tool is the Go race detector. Its id is `go-race`. Its command is:

```text
go test -race ./<package>
```

One `./<package>` for each directory of a changed Go file, sorted. A Go file at the checkout root is the package `.`.

### Sample scenario

Sample mode runs no container. The recorded scenario `RaceIntroduced` has the detector's banner on the change and a passing line on the base. The version in that recording is `go1.26`, the module's go line, not a measurement from a container. A live run that does not ask the container for a version records `unknown`.

That scenario reports one finding. The Judge keeps it. Its proof level is `proved-by-tool`.

Two recordings are the controls:

- `RaceOnBoth` has the banner on both sides. It is not a finding.
- `RaceTimeout` is a change that did not answer. The reason names `go-race` and `timeout`.

### Adding the next tool

Fuzzing, gosec and coverage are later. A new tool is a new adapter with the four answers above, a recorded scenario, and tests for three outcomes: a result only on the change, the same result on the base, and a run that does not answer. It uses the same container and the same time-limit setting. It does not grow a second way to start a process.
