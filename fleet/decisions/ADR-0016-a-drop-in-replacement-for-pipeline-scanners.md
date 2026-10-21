# ADR-0016: proven drops into a pipeline in place of a scanner such as Semgrep or Opengrep, with the same flags, outputs and exit codes where they mean the same

**Status:** accepted · **Date:** 2026-10-20
**Decided by:** the owner · **Written by:** the advisor (claude-code, `claude-opus-5-5`), 2026-10-20
**Follows:** [ADR-0013](ADR-0013-no-hardcoded-settings.md) (no hardcoded settings)
**Trigger:** the owner, on how teams adopt proven ([`planning/CONVERSATION_2026-10-20.md`](../../planning/CONVERSATION_2026-10-20.md), hop 7)

## Context

The owner:

> *"Proven should be a drop-in replacement for popular PR GitHub Actions ([create-pull-request](https://github.com/marketplace/actions/create-pull-request)), Semgrep, Opengrep and others. If possible, the same command-line parameters, inputs and outputs, with as few changes to the customer's pipeline as possible. It should be easy to do; we do not need too much. At least we should support these three output formats at the start: text, JSON and SARIF."*

What the pipelines look like today:

- **Semgrep** runs as `semgrep scan` locally and `semgrep ci` in pipelines, which on a pull request reports only what the pull request introduced. Its output flags are `--text`, `--json` and `--sarif`, each with a `--<format>-output=<file>` twin, plus `-o`/`--output`. `--error` makes `scan` exit 1 on findings; `ci` fails on findings by default. Exit code 2 is a fatal error. `--baseline-commit` (environment `SEMGREP_BASELINE_COMMIT`) shows only results not in that commit. ([CLI reference](https://docs.semgrep.dev/cli-reference))
- **Opengrep** is a fork of Semgrep 1.100.0 under LGPL 2.1, run as `opengrep scan`, with JSON and SARIF output and `--sarif-output=<file>`. ([opengrep](https://github.com/opengrep/opengrep))
- **SARIF** is what GitHub Code Scanning and most security dashboards import.
- The linked **create-pull-request** action is of a different kind: it opens pull requests from changes a workflow made, and reviews nothing. proven does not replace it. proven reviews the pull requests such an action opens, so the two sit in the same pipeline without changing each other.

## Decision

1. **proven has two commands in the scanners' shape**: `proven scan` for a local run and `proven ci` for a pipeline. On a pull request, `ci` reviews only the change.
2. **Three output formats from the start**, chosen by the same flags as Semgrep's:

   | format | flag | file twin | what it is |
   |---|---|---|---|
   | text | `--text` (default) | `--text-output=<file>` | readable in a terminal: each finding with its code and its evidence |
   | json | `--json` | `--json-output=<file>` | everything, machine-readable: findings, evidence, errors, what the run cost |
   | sarif | `--sarif` | `--sarif-output=<file>` | SARIF 2.1.0, for GitHub Code Scanning and other dashboards |

   `-o`/`--output=<file>` writes the chosen format to a file. Several twins may be given at once, so one run writes all three.
3. **The flags keep Semgrep's names where the meaning is the same**: the output flags above, `--error` and `--no-error`, `--baseline-commit`, `--quiet`, `--verbose`. proven's own environment variable for each is `PROVEN_<NAME>`, and the Semgrep name (`SEMGREP_BASELINE_COMMIT`) is read as a fallback, so an existing pipeline's variables keep working.
4. **Exit codes keep Semgrep's numbers where they apply**: 0 OK, 1 findings (with `--error` in `scan`; by default in `ci`), 2 fatal error. A failing outside service is not a fatal error: the review degrades and says so ([ADR-0001](ADR-0001-go-behind-adapters.md), point 3).
5. **Every finding carries its evidence in every format.** In SARIF the test or the query and its rows go into the result, so the proof travels into the dashboard with the finding ([ADR-0002](ADR-0002-evidence-is-a-test-or-a-query-through-zetaris.md)).
6. **The JSON keeps Semgrep's top-level layout** (`results`, `errors`, `version`) where the meaning matches, so pipeline steps that read those keys need the fewest changes. proven's own fields are documented and versioned. Field-for-field compatibility is not promised.
7. **What is not copied**: Semgrep's rule registry, its rule syntax, its login and its app token. `--config` points to proven's own configuration file; a registry value such as `auto` or `p/<name>` is refused when the configuration loads, with a message that says what `--config` takes in proven ([ADR-0013](ADR-0013-no-hardcoded-settings.md)). An unknown flag is refused with the list of known ones.
8. **proven ships as a container image and as a GitHub Action** whose inputs are the same flags. The documentation shows, for Semgrep and for Opengrep, the pipeline before and after, and the lines that changed.

## Consequences

- Moving from Semgrep or Opengrep should change the image and the command name, and little else. The before-and-after examples are how that claim is checked; the claim is not measured until they run.
- SARIF puts proven's findings in the same place as a scanner's, with the evidence inside each one.
- proven is not a pattern scanner. It is slower and calls models, so a team may run both, the scanner for breadth and proven for proof. Nothing in this decision prevents that.
- More flags mean more to test; each one is tested ([ADR-0013](ADR-0013-no-hardcoded-settings.md), Consequences).
- Other output formats (GitLab SAST, JUnit, Checkstyle) and other tools' flags are not part of this decision.
