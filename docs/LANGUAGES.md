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

The live Prover uses the registered language adapters below. It receives the candidate risk, bounded source excerpts from base/head and nearby tests. A finding requires the same selected test to pass on base and fail on head in the restricted runner.

| Language | Adapter / framework | Accepted test files | Test selection |
| --- | --- | --- | --- |
| Go | Native `go test -json` | `*_test.go` | A `TestName` function. |
| Python | pytest, including unittest-style classes | `test_*.py`, `*_test.py` | `test_function` or `ClassName::test_method`; parametrized cases are retained. |
| TypeScript | Vitest | `.test.ts`, `.spec.ts`, TSX/MTS/CTS variants, TypeScript files under `__tests__/` | Exact full test name, including enclosing `describe` names. |
| JavaScript / Node | Native `node:test` | `.test.js`, `.test.mjs`, `.test.cjs` | Exact full test name, including nested suite names. |
| JavaScript / Vitest | Vitest | `.spec.js`, `.spec.mjs`, `.spec.cjs` | Exact full test name. This naming keeps the adapter choice explicit. |

### Shared proof and fix contract

| Capability | Go | Python | TypeScript | JavaScript/Node |
| --- | --- | --- | --- | --- |
| Same generated test staged on base and head | Yes | Yes | Yes | Yes |
| Native pass/failure observations; exit zero alone is insufficient | Yes | Yes | Yes | Yes |
| Build/import/collection/setup errors stay distinct from an observed failing test | Yes | Yes | Yes | Yes |
| Proving test rerun on base, head and patched head | Yes | Yes | Yes | Yes |
| Additional package/project tests and touched-area rechecks | Yes | Yes | Yes | Yes |
| Read-only source, no network, non-root execution and secret filtering | Yes | Yes | Yes | Yes |
| Shared CPU/memory/process/time bounds and container cleanup | Yes | Yes | Yes | Yes |

The entry point remains `testrun.Adapter`; `Container` also implements `FixRunner` and `BreakageRunner`. [`LanguageAdapters`](../internal/testrun/languages.go) supplies fixed commands and naming rules. Embedded reporters consume native events or framework reports; model text never selects a shell command. New-language results retain the existing test-evidence shape in reports and packs.

Account-free real-container tests use pytest `9.2.0`, Vitest `5.1.0` and Node 24. They cover a deliberate regression, a correct change, fix reruns, neighbouring tests, syntax/import failure, unittest classes, nested Node suites and Vitest hook failure. This establishes the tested contract, not compatibility with every framework release or project setup.

The framework and offline dependencies must exist in the selected runtime. Node/Vitest use the prepared `node_modules` at `/deps/node_modules`, or dependencies already present in the staged project. Helpers copy source to bounded scratch; the original source mount stays read-only. pytest runs without its cache plugin. No dependency install occurs during proof.

Kotlin/Gradle, Rust/Cargo and Swift/XCTest are follow-on generated-test adapters. Their manifest support is a separate capability. SwiftUI requires macOS/Xcode and an appropriate isolated execution adapter; a Linux Docker image cannot supply that environment.

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
