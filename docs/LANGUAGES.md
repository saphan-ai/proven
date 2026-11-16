# Language and tool coverage

Proven has separate capabilities for dependency parsing, generated-test proof and deterministic tools. A language appearing in the dependency table does not mean every program in that language can be tested or automatically repaired.

## Dependency readers

The deterministic path checks packages added or raised by the change. It is not a complete inventory scan of every dependency already present on base, and the parsers do not perform a package manager's full dependency resolution.

| Language / ecosystem | Files read | Important boundary |
| --- | --- | --- |
| Go / `Go` | `go.mod`, `go.sum` | `require` lines; `replace`/`exclude` are not resolved. Checksum entries are not full module selection. |
| JavaScript / `npm` | `package.json`, `package-lock.json` | `dependencies`/`devDependencies`; lock formats 2/3 `packages`. Ranges such as `^1.2.3` are not resolved as exact versions. |
| Python / `PyPI` | `requirements.txt`, `poetry.lock` | Exact `name==version` pins and package name/version lock entries. PEP 440 subset, not arbitrary constraints or environment resolution. |
| PHP / `Packagist` | `composer.json`, `composer.lock` | Composer version subset; platform packages are skipped. Branch names and constraints are not guessed. |
| Rust / `crates.io` | `Cargo.toml`, `Cargo.lock` | Literal version entries and package locks; semantic-version subset. A caret requirement is not a resolved pin. |
| Kotlin / `Maven` | `libs.versions.toml`, `build.gradle.kts`, `gradle.lockfile` | Literal coordinates/catalog references supported by the reader; no Groovy evaluation, dynamic versions or arbitrary build execution. |

When a manifest and its paired lock file in the **same directory** name the same package, the lock's version wins. A Gradle catalog in a subdirectory is not automatically paired with a root lock. PyPI names use PEP 503 normalization; Packagist names are lower-cased. Other ecosystem spellings are retained.

Unsupported values and unreadable lock forms can produce trace notes rather than findings. That is a coverage boundary, not evidence that the ignored requirement is safe. [OSV_TABLE.md](OSV_TABLE.md) specifies exact supported subsets and refusal behavior. [Adapters](ADAPTERS.md#adding-a-language) explains how to add coverage with tests.

## Advisory evaluation

Prepared OSV rows identify ecosystem, package and affected version bounds. Go performs the comparison; the database does not decide whether an arbitrary version string is inside a range. A matching affected range can become data evidence. KEV membership, EPSS and other registered sources add context under their own rules.

Version presence is not runtime reachability. A dependency finding must not be described as an executed exploit or a proved application call path. No matching row in the imported snapshot is not a guarantee that a package has no vulnerability. See [Sources](SOURCES.md), [Queries](QUERIES.md) and [Data walkthrough](DATA_WALKTHROUGH.md).

## Generated tests

The current live Prover generates Go tests. It receives the candidate risk, bounded source excerpts from base/head and existing tests from the relevant package. The test must establish the required base-pass/head-fail differential outcome in the restricted runner.

Models can discuss other languages, but text is not a runner implementation. Wider proving coverage needs a language-appropriate test protocol, controlled command construction, prepared toolchain, isolation and failure-path tests. See [Runner](RUNNER.md) and [test adapter](ADAPTERS.md#adding-a-test-run).

## Deterministic tools

| Adapter | What it observes | Preparation |
| --- | --- | --- |
| Go race detector | Controlled base/head race-test outcomes, with evidence level retained | Compatible Go/CGO/compiler environment, dependencies and bounded runtime. |
| govulncheck | Structured base/head Go vulnerability results | Prepared executable and local database; no public fetch during the run. |

Tool presence and execution are separate. A configured tool that cannot run contributes no proof. Consult [tool adapters](ADAPTERS.md#adding-a-tool) for parsing, differential and rerun rules.

## Adding another language

Implement a manifest/lock reader and comparator only if supported syntax can be stated and tested precisely. Share source/query and report paths. To add generated-test proving, separately define safe commands, test placement, package discovery, differential outcome classification and fix rechecks. Keep no-run/timeout distinct from a real test failure. Small interface examples live under [examples](../examples/README.md).

The [roadmap](../ROADMAP.md) tracks broader source and evaluation work. It is a direction, not a claim that planned runners or feed connectors are already delivered.
