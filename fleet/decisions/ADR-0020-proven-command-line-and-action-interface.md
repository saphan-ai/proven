# ADR-0020: Proven's command line and GitHub Action: the exact flags, inputs, outputs and exit codes, taken from Semgrep and the Opengrep action

**Status:** proposed · **Date:** 2026-10-21
**Proposed by:** the advisor (claude-code, `claude-opus-5-5`), 2026-10-21, at the owner's request; accepted when the owner says so
**Extends:** [ADR-0016](ADR-0016-a-drop-in-replacement-for-pipeline-scanners.md) (a drop-in replacement for pipeline scanners)
**Follows:** [ADR-0013](ADR-0013-no-hardcoded-settings.md) (no hardcoded settings), [ADR-0017](ADR-0017-a-gitea-lab-for-the-judges.md) (the Gitea lab)
**Trigger:** the owner, on what is decided before the build window and what is built in it ([`planning/CONVERSATION_2026-10-21_advisor.md`](../../planning/CONVERSATION_2026-10-21_advisor.md), hop 12)

## Context

[ADR-0016](ADR-0016-a-drop-in-replacement-for-pipeline-scanners.md) decided that Proven drops into a pipeline in place of Semgrep or Opengrep. It did not list the exact names. The owner (translated from Polish): *"The same goes for checking which call parameters Proven should accept, the drop-in replacement in GitHub Actions. All of it, the data sets and the call, goes in as decisions."* The code is written from 22 October; the interface is decided now.

How the two tools are run in GitHub Actions today, read on 21 October:

- **Semgrep** documents one job: the container `semgrep/semgrep`, `actions/checkout`, then `run: semgrep ci`, with `SEMGREP_APP_TOKEN` from the repository's secrets. On a `pull_request` trigger it scans only what the pull request changed, with no baseline variable set. Its free variant runs `semgrep scan --config auto`. ([sample CI configs](https://docs.semgrep.dev/semgrep-ci/sample-ci-configs), [CLI reference](https://docs.semgrep.dev/cli-reference))
- **Opengrep** has a community action, `platform-sec/opengrep-action`, with inputs `target`, `patterns`, `config`, `output-format`, `severity`, `include`, `exclude`, `timeout`, `jobs`, `strict`, `baseline-commit`, and outputs `results-file`, `json-file`, `sarif-file`, `findings-count`. ([opengrep-action](https://github.com/platform-sec/opengrep-action))
- **SARIF into GitHub** goes through `github/codeql-action/upload-sarif`, input `sarif_file`, optional `category`, with the job permission `security-events: write`. ([upload-sarif](https://github.com/github/codeql-action/blob/main/upload-sarif/action.yml))

## Decision

### 1. The command line

`proven scan [TARGET]` reviews a directory or the changes since a baseline, locally. `proven ci` reviews a pull request in a pipeline: it reads the pipeline's context (`GITHUB_EVENT_PATH`, `GITHUB_BASE_REF`, `GITHUB_SHA`; Gitea Actions sets the same names) and reviews only the change against the merge base.

| flag | environment variable | default | meaning (same as Semgrep's unless said) |
|---|---|---|---|
| `--text` · `--json` · `--sarif` | `PROVEN_OUTPUT_FORMAT` | `text` | format on standard output |
| `--text-output=<file>` · `--json-output=<file>` · `--sarif-output=<file>` | — | none | also write that format to a file; any number at once |
| `-o`, `--output=<file>` | `PROVEN_OUTPUT` | standard output | write the chosen format to a file |
| `--error` · `--no-error` | `PROVEN_ERROR` | `scan`: off · `ci`: on | exit 1 when there are findings |
| `--baseline-commit=<ref>` | `PROVEN_BASELINE_COMMIT`, then `SEMGREP_BASELINE_COMMIT`, then `SEMGREP_BASELINE_REF` | `ci`: the pull request's base | review only what changed since that commit |
| `--config=<file>` | `PROVEN_CONFIG` | `proven.yaml` if present, else the defaults | **Proven's** configuration file ([ADR-0013](ADR-0013-no-hardcoded-settings.md)); a Semgrep registry value such as `auto` or `p/…` is refused with a message saying what `--config` takes |
| `--include=<glob>` · `--exclude=<glob>` | `PROVEN_INCLUDE` · `PROVEN_EXCLUDE` | all files the change touches | which changed files are reviewed |
| `--severity=<level>` | `PROVEN_SEVERITY` | `INFO` | lowest severity reported: `INFO`, `WARNING`, `ERROR` (SARIF `note`, `warning`, `error`) |
| `--timeout=<seconds>` | `PROVEN_TIMEOUT` | from the configuration | the whole review's time limit; a review that runs out returns what it proved so far |
| `--sample` | `SAMPLE_MODE=true` | off | sample adapters: no network, no keys ([ADR-0001](ADR-0001-go-behind-adapters.md)) |
| `--quiet` · `--verbose` | — | neither | as Semgrep |
| `--version` · `--help` | — | — | as usual |

Unknown flags are refused with the list of known ones. Secrets are environment variables only and never flags: `METERLESS_RUNTIME_TOKEN` ([ADR-0019](ADR-0019-meterless-nightshift-is-the-primary-ai-provider.md)), `ZETARIS_API_KEY` and `ZETARIS_ORG_ID` ([ADR-0015](ADR-0015-zetaris-through-its-rest-api.md)), and a token for posting on pull requests where that is switched on.

### 2. Exit codes

| code | meaning | as Semgrep |
|---|---|---|
| 0 | review done; no findings, or findings without `--error` | yes |
| 1 | review done; findings, with `--error` (the default in `ci`) | yes |
| 2 | fatal error: Proven could not review | yes |
| 7 | the configuration was refused when it loaded ([ADR-0013](ADR-0013-no-hardcoded-settings.md)) | Semgrep's 7 is "missing configuration" |

An outside service that fails is not exit 2: the review degrades, says which service failed, and exits 0 or 1 ([ADR-0001](ADR-0001-go-behind-adapters.md)).

### 3. The GitHub Action

An `action.yml` at the root of this repository, a container action on Proven's published image, used as `uses: saphan-ai/proven@<full commit SHA>`. Its inputs use the Opengrep action's names where the meaning is the same:

| input | default | maps to |
|---|---|---|
| `target` | `.` | the `scan` target |
| `config` | none | `--config` |
| `output-format` | `json/sarif` | `json/sarif`, `json`, `sarif` or `text`, as in the Opengrep action |
| `severity` | `INFO` | `--severity` |
| `include` · `exclude` | none | `--include` · `--exclude` |
| `timeout` | none | `--timeout` |
| `strict` | `true` | `--error` |
| `baseline-commit` | the pull request's base | `--baseline-commit` |
| `sample-mode` | `false` | `--sample` |
| `comment` | `false` | post the review on the pull request; needs `pull-requests: write` |

Outputs, with the Opengrep action's names: `results-file`, `json-file`, `sarif-file`, `findings-count`; and one of Proven's own, `unresolved-count`, the findings whose evidence did not settle them.

The Opengrep action's `patterns`, `jobs` and `opengrep-version` have no meaning in Proven and are not inputs. GitHub only warns about an input an action does not define, so Proven cannot refuse one; the documentation names them.

### 4. Before and after, in the documentation

**Semgrep → Proven:** the container `semgrep/semgrep` becomes Proven's image; `run: semgrep ci` becomes `run: proven ci`; `SEMGREP_APP_TOKEN` becomes the Proven secrets above. A `upload-sarif` step, if there is one, stays and only its file name changes.

**Opengrep action → Proven:** `uses: platform-sec/opengrep-action@<sha>` becomes `uses: saphan-ai/proven@<sha>`; `target`, `config`, `output-format`, `severity`, `include`, `exclude`, `timeout`, `strict` and `baseline-commit` stay as they are; `patterns` is removed. The outputs a later step reads keep their names.

**SARIF into GitHub code scanning:** `github/codeql-action/upload-sarif` with `sarif_file` set to Proven's SARIF and `category: proven`, so it does not replace another tool's analysis of the same commit.

### 5. Permissions

`contents: read` always. `security-events: write` only for the SARIF upload step. `pull-requests: write` only when `comment` is on. Proven asks for nothing more.

### 6. Runners

The owner pointed to GitHub's documentation on [GitHub-hosted](https://docs.github.com/en/actions/how-tos/manage-runners/github-hosted-runners/use-github-hosted-runners) and [self-hosted](https://docs.github.com/en/actions/how-tos/manage-runners/self-hosted-runners/add-runners) runners. Three facts from it shape the Action:

- A container action *"can only execute on runners with a Linux operating system"*, and *"self-hosted runners must use a Linux operating system and have Docker installed"* ([about custom actions](https://docs.github.com/en/actions/sharing-automations/creating-actions/about-custom-actions)). So `uses: saphan-ai/proven@…` runs on Linux runners. On macOS and Windows runners, the documentation shows the command line in a `run:` step instead, from a release binary.
- A GitHub-hosted runner is a fresh virtual machine in GitHub's cloud. It cannot reach a team's private network, so a Zetaris or a Nightshift that is not reachable from the internet is out of its reach. There Proven runs in sample mode, or against endpoints the team makes reachable.
- GitHub *"recommend[s] that you only use self-hosted runners with private repositories"*, because *"forks of your public repository can potentially run dangerous code on your self-hosted runner machine"*. So, for Proven's own public repository: its workflow runs on GitHub-hosted runners in sample mode, and live reviews with our Zetaris and Nightshift run in the Gitea lab or in private repositories on our own runners. The documentation tells teams the same: a self-hosted runner for live mode, never for pull requests from forks of a public repository.

### 7. Comments on the pull request, through reviewdog

[reviewdog](https://github.com/reviewdog/reviewdog) reads SARIF 2.1.0 (`-f=sarif`) and posts it as review comments on a pull request: `-reporter=github-pr-review` on GitHub, `-reporter=gitea-pr-review` on Gitea, plus GitHub's checks and annotations. Many pipelines already use it for their linters.

- Proven writes SARIF; posting comments is reviewdog's job. The Action's `comment` input runs reviewdog on Proven's SARIF, and a team that already runs reviewdog adds one more input to it.
- This replaces the Gitea comment adapter of [ADR-0017](ADR-0017-a-gitea-lab-for-the-judges.md), point 4, if the owner accepts it: less code of our own, and the same path on GitHub and Gitea.

### 8. Other pull-request review tools

AI reviewers (for example Qodo's PR-Agent, CodeRabbit, GitHub Copilot's code review, Anthropic's Claude Code action) are usually installed as GitHub Apps or actions that post comments, each with its own commands and settings. Proven does not copy their commands. It meets them where their output meets the pipeline: comments on the pull request (point 7), SARIF in code scanning, and a failing check. Matching one of them input for input is not part of this decision; their documentation was not read for it.

## Consequences

- Moving from either tool changes the image or the `uses:` line and the secrets, and keeps every other line. The lab ([ADR-0017](ADR-0017-a-gitea-lab-for-the-judges.md)) runs the same workflow on Gitea, which checks that claim.
- Every flag and input is a setting with a documented default ([ADR-0013](ADR-0013-no-hardcoded-settings.md)); each is tested at its default and at one other value.
- Semgrep's rule registry, its app token and Opengrep's rule patterns are not copied: Proven is not a pattern scanner.
- The exact JSON fields and the SARIF properties that carry each finding's evidence are written with the code, from 22 October, within [ADR-0016](ADR-0016-a-drop-in-replacement-for-pipeline-scanners.md)'s rules.
