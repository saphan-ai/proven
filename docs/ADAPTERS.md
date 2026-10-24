# Adding a language

One language is one adapter ([ADR-0038](adr/ADR-0038-a-language-is-one-small-adapter.md)). The adapter finds the language's files, lists the packages a change adds or raises, names the ecosystem as OSV names it, and compares two versions by that ecosystem's rules. OSV ranges, the known-exploited list, EPSS and the review are shared. A new language does not change them.

The type is `Adapter` in `internal/manifest`. It is the same type as `Reader`. The dependency check runs `Readers()`. A version lands in a range through `InRangeEcosystem`, which calls the adapter's `Compare` and `Explain`.

## What you add

1. A comparer for the ecosystem's versions, with tests for equal, lower, higher, and for a string the comparer refuses. Put the refusal in a sentence. Do not guess a version for a string you refuse.
2. A manifest reader and a lock reader. Each is an `Adapter`. Both use the same ecosystem name, the same `Compare` and the same `Explain`. The lock reader sets `Lock` and `NoteErr`, so a lock file that cannot be parsed is a trace note and not a failed review.
3. `Match` accepts the file's own name at any depth, and refuses a backup such as `name.bak`.
4. `Diff` reports a package the change adds, and a package whose version went up. A lower version is not a finding. A pair that does not order is still reported, so the trace can say why.
5. One line in `Readers` for each reader.
6. The ecosystem name in the `osv-by-package` query (`internal/querytpl`), and the package name's format. Update [QUERIES.md](QUERIES.md) in the same change: a test holds that page to the list.
7. The files and the version rules in [OSV_TABLE.md](OSV_TABLE.md).
8. A sample scenario under `testdata/fixture/` and a section in [SAMPLE_MODE.md](SAMPLE_MODE.md). The scenario's test fails when the adapter is taken out of `Readers`.

`Compare` on the adapter and the order `Diff` uses are the same function. A raise and a range then cannot disagree.

## Worked example: PHP

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

## What stays shared

You do not add a new query, a new range rule, or a new place in the report. `InRangeEcosystem` already asks the adapter. An ecosystem that is not in `Readers` is not compared, and the trace says so.
