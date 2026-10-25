# Adapters

An adapter is how Proven grows without changing its core. A language adapter reads one ecosystem's files and versions ([ADR-0038](adr/ADR-0038-a-language-is-one-small-adapter.md)); a tool adapter turns a deterministic tool's run into evidence ([ADR-0040](adr/ADR-0040-a-tool-run-is-evidence.md)); a source adapter loads a table and turns a query's rows into evidence ([ADR-0046](adr/ADR-0046-a-data-source-is-one-adapter.md)). An output format writes the review result. Each has one worked example below.

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

## Adding a source

A source adapter feeds the evidence store and plugs into the decision ([ADR-0046](adr/ADR-0046-a-data-source-is-one-adapter.md)). It declares four things.

1. **How it loads.** Where the data comes from, how often it is refreshed, the table name it is registered as, and the snapshot date of that load. The adapter does not name a store ([ADR-0047](adr/ADR-0047-the-evidence-store-is-an-adapter.md)). `PROVEN_EVIDENCE_STORE` runs the query.
2. **Its query templates.** Each one is a template the Investigator may pick, plus the builder that writes the query. Each parameter carries a format as a pattern; `Register` refuses a parameter with none. A model never writes the query ([ADR-0032](adr/ADR-0032-proven-builds-every-query-from-a-template.md)). `Catalogue` and `Build` in `internal/querytpl` read the registry, so the template is written in one place.
3. **Its rule.** What a returned row means. Rows that mean a finding become evidence: the query and those rows. Rows that mean nothing are not evidence. The Judge decides, as it does for any other finding. On the Investigator's path the rule of the adapter that owns the template decides whether the rows become evidence.
4. **A sample table and a test.** The test adds the adapter by registering it.

The type is `Adapter` in `internal/source`. The built-in adapters are OSV, KEV, EPSS, then the four of [ADR-0039](adr/ADR-0039-more-evidence-sources.md): malicious packages (`mal`), project health (`scorecard`), CVSS (`cvss`) and SPDX licences (`spdx`), in that order. OSV owns `osv-by-package` and `osv-by-advisory`. KEV owns `kev-by-cve`. EPSS owns `epss-by-cve`. `mal` owns `mal-by-package`. `scorecard` owns `scorecard-by-package`. `cvss` owns `cvss-by-cve`. `spdx` owns `spdx-by-package`. Their query text stays in `internal/querytpl`, beside the check that refuses a value outside its format. `internal/source` does not import `internal/querytpl`. The catalogue reads the registry, so the import goes the other way.

None of these is loaded by this program. An administrator puts the table where the store can read it. A finding cites the date the store stamps on the read. The load's own snapshot date is empty, because this program did not load the table.

A malicious-package row is evidence only when its advisory id starts with `MAL-`. A project-health row is evidence when maintainers score `0`, releases are unsigned (`signed` is `false`), or `release_age_days` is under 14. A CVSS row is a severity score and is evidence; the Investigator's finding then receives the EPSS score of the same CVE, so the two sit on one finding. A licence row is evidence when the `licence` cell is outside a closed permissive set (MIT, Apache-2.0, BSD-2-Clause, BSD-3-Clause, ISC, 0BSD, Unlicense). The cell is compared whole. That set stands in for an organisation policy, which this program does not store. A missing table is a down source named for the store, on the local store, on Zetaris and on CouchDB.

The sample scenarios are `testdata/fixture/malicious-package`, `testdata/fixture/project-health`, `testdata/fixture/cvss-score` and `testdata/fixture/spdx-licence`. Each scenario's test fails when that finding is absent from the review. The malicious-package test also fails when an empty row list is treated as evidence.

The OSV rule does not decide the dependency finding. A returned row is an advisory range. The dependency check compares the change's version, and a version outside every range is a trace line, not a finding. The rule does not see the version. KEV and EPSS use their rules: a row is evidence, and no row is not.

### Worked example: a banned-package list

An organisation keeps a table of packages it does not allow. The adapter id is `banned-package`. The table name is `banned_pkg`. The snapshot date in the example is `2026-10-24`. The origin is the organisation's own list. The refresh is whenever that list changes.

The template is `banned-by-package`. It takes `package`. The builder writes:

```sql
SELECT name, reason FROM banned_pkg WHERE name = '<package>'
```

The parameter declares its format as a pattern, `^[A-Za-z0-9._-]{1,200}$`. Proven holds every value to that pattern, and to a floor of its own, before the builder runs: no quote, no backslash, no control byte and nothing outside printable ASCII. A value outside either is refused with a named reason and never escaped. The builder is not the only guard. A builder that refuses a value for its own reason returns a `*querytpl.Refusal`; any other error it returns is turned into one, so the review records a query note and goes on. The table name is the adapter's own. `Register` refuses it unless it is a plain name, one identifier or `schema.name`.

The rule is: a row means the package is banned, and the evidence is that query and the row. No row means the package is not on the list, and there is no finding.

The columns of the sample table are `name` and `reason`. One sample row is `left-pad` with reason `not allowed`.

The test `TestAddedSourceShowsItsTemplateItsQueryAndItsRule` registers this adapter, sees `banned-by-package` in the catalogue, runs the query on the local store, and checks that the rule returns the row. It also queries `right-pad`, which is not in the table, and checks that the rule returns nothing. Taking `bannedPackage` out of the added list makes that test fail. The adapter is not one of the built-in sources. There is no plugin loader: `Register` is the door, and this program calls it from the test.

## The evidence store

[ADR-0047](adr/ADR-0047-the-evidence-store-is-an-adapter.md). A review's queries run on one store. `PROVEN_EVIDENCE_STORE` chooses it: `zetaris` (the default), `local` or `couchdb`. `local` is one file and no server, and it opens no network. The file of the local store is `PROVEN_EVIDENCE_FILE`. It is required when the store is `local`: an empty value is refused by name, because a default would sit in the repository under review. A path that resolves inside that repository, links included, is refused by name too. A table the file does not hold is a down source named `local`, as a missing view on Zetaris is a down source named `zetaris`, and a missing database on CouchDB is a down source named `couchdb`. Sample mode does not read either variable: it opens the fixture's tables file as that same local store.

`couchdb` runs a Mango selector over CouchDB's HTTP API. One registered table is one database. Its name is `PROVEN_COUCHDB_PREFIX` (default `proven_`), then the table name in lower case with `.` written as `-`, so `px_evidence.px_kev_vulns` is `proven_px_evidence-px_kev_vulns`. The address is `PROVEN_COUCHDB_URL`; the user and password are `PROVEN_COUCHDB_USER` and `PROVEN_COUCHDB_PASSWORD`, sent as HTTP Basic. The password is a secret: it is read from the environment only, a refusal never prints it, and the redactor removes it from the trace and the result. A value that is set wrongly is refused by name. An empty address, user or password is a down store named `couchdb`, and the review goes on without it. The client does not follow redirects, so the credential goes to the configured address and nowhere else. Rows come back in the column order of the `sql` form, though CouchDB returns the fields of a document in no fixed order. Each cell is text.

A source registers a table on the configured store, then runs a query Proven built from a template. Zetaris keeps the rule that Proven's key only reads: registering a table there is the administrator's act, and Proven records the name and sends no statement. The local store writes the table into its file. The CouchDB store creates the table's database and writes one document for each row; registering the same table twice is refused.

A template has one form for each query language: `sql` for Zetaris and the local store, and `mango` for CouchDB. Both are built from the same validated values (ADR-0032), by the same checks, so a value one form refuses the other refuses. A `mango` form is one JSON object: the table, a selector with one equality for each parameter, and the fields in column order. The evidence of a query names the form that ran, with the text of that form and the rows it returned.

The review's query seam passes a template request and the relation names. The configured store builds the form of its own language and runs it: `sql` on Zetaris and the local store, `mango` on CouchDB. The evidence names that form, the text that ran and the rows. A template a source adapter registered has the sql form only, so CouchDB names itself down for that template (`no mango form`) instead of sending SQL. Sample mode does not read the store setting: it opens the fixture file and runs the sql form.

One scenario runs on every store that is built. The columns, the rows and the snapshot match on all three; the form differs as the language does. Zetaris and CouchDB are played by recorded HTTP servers in that test, and the local store is the real file. A second test runs against a real CouchDB when `PROVEN_COUCHDB_LIVE_URL` names one (with the user and password set), and skips by name otherwise.

### Evidence records and their viewer

The tables above are the evidence a review reads. A review's own result is written back to CouchDB when `PROVEN_COUCHDB_RECORDS` names a database on the server `PROVEN_COUCHDB_URL` names (ADR-0013). Unset, nothing is written. Set, each finding becomes one document: the pull request and commit, the finding's id, rule, severity and place, its proof level, the CVE ids and package names its rows hold, and its evidence as it ran. A query keeps the form that ran, its text and its rows. A test keeps its command and whether it fails on the change and passes on the base. A tool keeps its command and both outputs. A second review of the same commit adds records; it never overwrites the first, so the records are a history.

The records are written with their own account, `PROVEN_COUCHDB_RECORDS_USER` and `PROVEN_COUCHDB_RECORDS_PASSWORD`. That account must be a different user from `PROVEN_COUCHDB_USER`; the loader refuses the same name. In CouchDB a member of a database's security document reads every document and also writes every document that is not a design document, so membership alone does not make an account read-only. The administrator therefore adds the read-only account as a member of the records database and the writing account as a database admin (the `admins` of its `_security` document), and grants the writing account admin rights on the records database only. The design document carries a `validate_doc_update` function that refuses every write unless the account is a server admin or a database admin, by name or by role, so the read-only account is refused a write even though it is a member. Creating the accounts and setting `_security` is the administrator's act: this program does neither.

The first write also stores one design document named `evidence`. It holds four views, whose keys start with the fields a reader narrows by: `by_review` (pull request, commit, finding, run), `by_finding` (finding id), `by_cve_package` (CVE, package) and `by_proof_level`. A record with a CVE and no package, or the reverse, is keyed with an empty text on the missing side. The same document carries the viewer as an attachment, `viewer.html`. CouchDB serves it itself at `/<database>/_design/evidence/viewer.html`; there is no other server. The page is one static file of HTML and JavaScript. It loads nothing from outside, and it makes only `GET` requests with the account the reader opened it with, so a read-only account is enough and holds no write right. It lists a review's findings (`?pr=<number>&commit=<sha>`) and shows one record (`?doc=<id>`): the query in the form that ran, its rows as a table, and the test command with its base and change result. Every value it shows goes through one escaping function, so a record that carries markup is shown as text. The design document is written again only when its views or its page change.

`PROVEN_EVIDENCE_LINK_BASE_URL` is the address a reader reaches CouchDB at. With it and the records setting both set, each stored finding carries `EvidenceRecordURL`, the address of its record in the viewer page: the base, the database, `_design/evidence/viewer.html`, and `?doc=` with the document id. The JSON output carries it as `findings[].evidence_record_url` (docs/OUTPUT_FORMATS.md); the pull request comment is rendered from that output. Unset, the field is empty and the comment shows no link. A finding whose record was not stored has no link either, and when the read-only account is configured, the review fetches the page as that account first and shows no link for a page it is not served. A CouchDB failure is a failed service named `couchdb`, and the review goes on. The link never goes through `/_utils`, the admin console, which can edit and delete. In the lab the base address is the lab's own CouchDB. Sample mode reads these four variables and no other CouchDB one, so the lab's sample review can write its records.

The tests play CouchDB as a recorded server that checks each request's account: the writing account may write, the read-only account may only read. They cover what the four views return for the sample scenario's records (the map functions are run in Node, as CouchDB would run them), that the design document and the page are written once and only by the writing account, that the page is served to the read-only account and a write with that account is refused, that the link exists only when the base address is set, and that a value carrying HTML is shown as text.

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
