# Model-free product and Action image check

Measured on 17 November 2026 with Docker Buildx and Docker Scout 1.23.1. These are local development candidates, not published release digests. The eventual release must be rebuilt from a clean committed revision and scanned again.

| Image | Platform | Packages indexed | HIGH | CRITICAL | Retained scan |
| --- | --- | --- | --- | --- | --- |
| Generic Proven | linux/amd64 | 72 | 0 | 0 | [Scout output](product-amd64.scout.txt) |
| Generic Proven | linux/arm64 | 72 | 0 | 0 | [Scout output](product-arm64.scout.txt) |
| GitHub Action | linux/amd64 | 93 | 0 | 0 | [Scout output](action-amd64.scout.txt) |
| GitHub Action | linux/arm64 | 93 | 0 | 0 | [Scout output](action-arm64.scout.txt) |

Local archive path names have been shortened and trailing whitespace removed in the retained output. The scans used explicit single-platform archives, `--exit-code --ignore-suppressed=false --only-severity critical,high`. No base exclusion, suppression or fix-availability filter was applied. The counts concern this scan date and selected severity levels; they are not an assurance against undiscovered vulnerabilities.

Candidate image indexes:

- Generic: `sha256:e19152e257417f15fb60900690e4ee8a509ea04cb8c006dba2a34e2001a45426`.
- Action: `sha256:2b7b217eaf1c427eebc4710858feeed6bb315631a70d855bc861a3350cbd7299`.

Both generic image architectures passed the bundled offline smoke check using actual Git commits, SQLite imports, OSV matching, deterministic source rules and per-commit description checks, with no network or model. The Linux arm64 image also passed [two live-provider runs](model-free-refresh-linux-arm64.json), including reuse of unchanged records. A controlled dependency fixture is not coverage of an arbitrary application or test framework.

## A dependency introduced by the Action helper

The initial Action image installed Python for comment collection. Alpine's Python brought `sqlite-libs 3.53.4-r0`, which Scout matched to **CVE-2026-119586 (HIGH)**. The initial scan failed and was not used as a passing result.

`Dockerfile.action` now explicitly installs the signed `sqlite-libs 3.54.1-r0` package through the existing tagged security repository, checks the exact package version and exercises Python's SQLite binding during the build. It also verifies that the `sqlite3` command-line shell is absent. Both architectures were rebuilt and rescanned; the table above records the completed candidates.

This is a package update and measured scanner result, not a claim that an upstream CLI issue was patched. The [SQLite project's CVE analysis](https://www.sqlite.org/cves.html) says this particular issue concerns piping attacker-controlled `.recover` output into another SQLite CLI without safe mode, rather than a bug in the core library. Proven's comment collector does not use SQLite or that command flow. [Alpine package identity](https://pkgs.alpinelinux.org/package/edge/main/x86_64/sqlite-libs).

The generic Proven image does not install Python or native SQLite packages. Its embedded Go SQLite implementation remains a separate dependency, covered by its own build and scanner inventory.

See [container publication](../../CONTAINERS.md) for the release gate and [CLI supply-chain audit](../../CLI_SUPPLY_CHAIN_AUDIT.md) for the separate customer-owned vendor images.
