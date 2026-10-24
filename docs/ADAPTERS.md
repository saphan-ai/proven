# Adapters

An adapter is how Proven grows without changing its core. A language adapter reads one ecosystem's files and versions ([ADR-0038](adr/ADR-0038-a-language-is-one-small-adapter.md)); a tool adapter turns a deterministic tool's run into evidence ([ADR-0040](adr/ADR-0040-a-tool-run-is-evidence.md)). An output format writes the review result. Each has one worked example below.

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

### govulncheck

The second tool is govulncheck ([ADR-0043](adr/ADR-0043-a-vulnerable-dependency-is-checked-for-reachability.md)). Its id is `govulncheck`. It applies to a changed Go file, `go.mod` or `go.sum`. Its command is:

```text
govulncheck -json -db file:///checkout/<database> ./...
```

`./...` is the whole module. A new dependency can be called from a file the diff did not touch. The database is a file URL under `/checkout`, the only directory the container can read. `PROVEN_GOVULNCHECK_DB` names that URL. Empty means the tool is not run. An `http` or `https` URL is refused: the container has no network, and a failed fetch must not be read as "not reached". The time limit is `PROVEN_TOOL_TIMEOUT`, the same setting as the race detector.

The JSON stream is one object per line. A trace whose first frame names a function and that has a caller is **reached**. The call path is that trace written from the caller down to the vulnerable function. Any other trace for the module is **imported but not reached**, including a module that is required but whose package is not imported. A module the stream does not mention is **not measured**: silence is not a measurement that the function is unreached.

Each dependency finding gains one label. A Go finding is **not measured** when govulncheck did not answer, including when the binary is not in the image and when the command exits without a config object (a missing database is that kind of miss, and it is not called a measurement). A finding for any other ecosystem is **not measured**. A finding is never dropped because it is not reached. Within one severity, a reached dependency finding ranks ahead of a dependency finding that is not reached. Scores from EPSS still order findings that share a label.

### Sample scenario

Sample mode runs no container. The recorded scenario `GovulncheckMixed` has a call stack for `example.com/reached` and a package-only trace for `example.com/quiet`. The version in that recording is `v1.1.4`. A live run that does not ask the container for a version records `unknown`.

That scenario labels the reached module with the call path and the quiet module as imported but not reached. A Packagist finding in the same list is not measured. The reached finding is ordered first.

### Adding the next tool

Fuzzing, gosec and coverage are later. A new tool is a new adapter with the four answers above, a recorded scenario, and tests for three outcomes: a result only on the change, the same result on the base, and a run that does not answer. It uses the same container and the same time-limit setting. It does not grow a second way to start a process.

## Adding an output format

A formatter writes one review result, for a person or for another program. text, JSON and SARIF are the three formats in this build ([ADR-0020](adr/ADR-0020-proven-command-line-and-action-interface.md)). The bytes of JSON and SARIF are the contract in [OUTPUT_FORMATS.md](OUTPUT_FORMATS.md). A new format is another writer. It does not change the review.

### The interface

`Formatter` is in `internal/cli`. It has one method.

```text
Write(w io.Writer, result review.Result) error
```

`w` is the destination. The shell does not hand `Write` a file. `render` hands it a memory buffer and keeps the bytes only when `Write` returns nil. A formatter does not open a path of its own.

`result` is the whole review: the findings, the opinions, the failed services, the class and the trace path. `--sample` swaps the reviewer for the sample adapters. The value that comes back is still a `review.Result`. `internal/cli/format.go` does not read the sample flag, so a formatter behaves the same in sample mode and in a live review.

Return nil only when `w` holds a complete body. A non-nil error means the body is not usable. Return the destination's error as it was given. `render` drops the buffer on that error. `writeOutputs` renders every body before it writes any of them, so a format that cannot be encoded publishes nothing. A failure while writing a later file leaves the first file as it was written. A formatter that writes straight to a destination the caller passed can leave a prefix in that destination. The shell's `render` does not pass such a destination.

`FormatterFor` is a closed switch over `text`, `json` and `sarif`. There is no registry. Any other name returns `format %q has no writer in this build` and writes nothing.

### Where the format is chosen

The standard-output format is one of `--text`, `--json` or `--sarif`. The command has no `--format` flag. Two of those three flags together are refused. When none is set, `PROVEN_OUTPUT_FORMAT` is used. The variable accepts `text`, `json` or `sarif` in any letter case. When the variable is empty, the format is `text`.

`-o` and `--output` write that chosen format to a file instead of standard output. They do not choose the format. When the flag is absent, `PROVEN_OUTPUT` names the file. When both are absent, the body goes to standard output.

`--text-output`, `--json-output` and `--sarif-output` each write that one format as well, in the same run, to their own file. Two outputs of one run must be two files. The collision rules are in [OUTPUT_FORMATS.md](OUTPUT_FORMATS.md).

### What you add

1. A type whose `Write` matches `Formatter`, and a case for its name in `FormatterFor` (`internal/cli/format.go`).
2. The flag. A format named `lines` is selectable only when every one of these exists. The flag sites are `internal/cli/config.go` and `internal/cli/run.go`:
   - a bool flag `--lines`, counted in `refuseExclusive` and read in `resolveFormat`
   - the name in `parseFormat`, so `PROVEN_OUTPUT_FORMAT` accepts it, and in `knownFlagNames`
   - a `--lines-output` flag, stored the way `--text-output` is, rendered in `writeOutputs`, and included in both collision checks
   - that output flag named in `isValueFlag`, so its path is not read as the command
   - one line in `helpText`

The worked example is not registered. Putting it on the command would edit `internal/cli`. This page shows the writer, and names the edits.

### Strings from the change

Every string the text writer prints that can come from the change goes through `visible` in `internal/cli/format.go`. C0 (U+0000 to U+001F), DEL (U+007F) and C1 (U+0080 to U+009F) are written as `\xNN`. A byte that is not valid UTF-8 is written as `\xNN` of that byte. It is not replaced with U+FFFD. A title that contains an escape then cannot move the cursor.

JSON and SARIF must not assume `encoding/json` does this. Both writers call `SetEscapeHTML(false)`. In Go 1.26.10 that encoder escapes a byte below 0x20. It copies DEL through: `safeSet` and `htmlSafeSet` both mark U+007F as safe, so turning HTML escaping off does not change DEL. It copies a C1 rune through as UTF-8. The only runes above ASCII that it escapes are U+2028 and U+2029. A byte that is not valid UTF-8 becomes U+FFFD, which is a different choice from `visible`. A consumer that prints a decoded JSON or SARIF string on a terminal still needs the `visible` rule. `go.mod` says `go 1.26` and names no toolchain. The reading above is the `encoding/json` in Go 1.26.10.

### Worked example: one line per finding

[`examples/formatter`](../examples/formatter) is a small format inside this module. `Lines.Write` prints one finding per line:

```text
path:line: title
```

`line` is `Location.StartLine`. The path and the title go through `Escaped`, which applies the same byte rule as `visible`, so a newline in the title stays inside its line. A review with no findings prints `findings: none`. An opinion is not a finding and is not a line.

The package assigns `Lines` to a `cli.Formatter`. That line fails the build when `Write` changes. `lines_test.go` renders one fixed `review.Result` and compares the bytes with `testdata/lines.golden`. A golden that differs by one byte fails the test.

```text
go test -count=1 ./examples/formatter
```

`go test ./...` runs the same test. No separate job is required.
