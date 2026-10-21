# TRAIN ORDER — proven-cli-and-action

- master: (to be appointed by the owner)
- status: DRAFT — written by the advisor on 2026-10-21, before the build window; the master may change it, then gives it to
  the owner to sign. **No wagon is dispatched before 2026-10-22 00:00 UTC.**
- track: proven-cli-and-action
- repo: proven (`saphan-ai/proven`), the product code
- base: main at the cut, 2026-10-22 or later
- author: ADVISOR of the hackathon track "The Agent That Can Explain Why"

## §0 THE OWNER, VERBATIM

2026-10-20: *"Proven should be a drop-in replacement for popular PR GitHub Actions — semgrep, opengrep, others. If
possible the same command-line parameters, inputs, outputs, as few changes to the customer's pipeline as possible ... At
least we should support those three output formats at the start [text/json/sarif]."*

2026-10-21 (translated from Polish): *"But self-hosted too. Not everyone will want to hand over their model API keys.
Locally too, or even above all."* · *"Gitea as an equal on-premises scenario. That is important. Not everyone has their
source code on the internet."* · *"Working first, pretty later."*

2026-10-21, asked whether the call of Proven as a drop-in replacement is planned (translated from Polish): *"Write it."*

## §1 WHAT THIS TRAIN IS

The outer shell of Proven: the command line, its three output formats, the GitHub Action, the comments on a pull
request, and the self-hosted runner demo. It is everything a pipeline touches, so that a team moving from Semgrep or the
Opengrep action changes as few lines as possible. **It does not build the reviewer itself** (finding, proving, the
Skeptic, Zetaris and Nightshift). That is another train. The seam between the two is one Go interface, defined in W1.
Until the reviewer exists, this train runs on the **sample reviewer**: findings from a fixture, no network, no keys
(`--sample`, [ADR-0001](../decisions/ADR-0001-go-behind-adapters.md)).

**The decisions this train builds, all accepted:** [ADR-0016](../decisions/ADR-0016-a-drop-in-replacement-for-pipeline-scanners.md)
(drop-in), [ADR-0020](../decisions/ADR-0020-proven-command-line-and-action-interface.md) (the exact interface: read it in
full; its tables are the specification), [ADR-0021](../decisions/ADR-0021-a-self-hosted-runner-demo-set-up-by-one-script.md)
(the runner script), [ADR-0022](../decisions/ADR-0022-gitea-on-premises-is-an-equal-scenario.md) (Gitea equal to GitHub),
[ADR-0013](../decisions/ADR-0013-no-hardcoded-settings.md) (no hardcoded setting; refused on load with the allowed list),
[ADR-0024](../decisions/ADR-0024-every-cursor-coding-session-is-recorded-and-published.md) (every Cursor session recorded).

**What the advisor read on 2026-10-20 and 21. CLAIMS FOR YOU TO TEST, not ground (law #44).** ADR-0020 rests on them.
Re-measure each one before a wagon builds on it. Where the advisor is wrong, the wagon's `master-was-wrong:` field says so,
and ADR-0020 gets a new ADR, never a silent difference.

| # | claim (advisor) | source | instrument |
|---|---|---|---|
| CA-1 | Semgrep's exit codes: 0 no findings or findings without `--error`; 1 findings with `--error`; 2 fatal; 7 missing configuration. | Semgrep CLI reference | the reference page, quoted with its date |
| CA-2 | Semgrep's flags `--json`, `--sarif`, `--text`, `--json-output=`, `--sarif-output=`, `--text-output=`, `-o`, `--error`, `--baseline-commit`, `--include`, `--exclude`, `--severity`, `--timeout` mean what ADR-0020 §1 says. | Semgrep CLI reference | the same |
| CA-3 | The Opengrep action's inputs `target`, `config`, `output-format`, `severity`, `include`, `exclude`, `timeout`, `strict`, `baseline-commit` and outputs `results-file`, `json-file`, `sarif-file`, `findings-count` are as ADR-0020 §3 says. | the action's `action.yml` | its `action.yml` at a pinned commit, quoted |
| CA-4 | Gitea Actions sets `GITHUB_EVENT_PATH`, `GITHUB_BASE_REF` and `GITHUB_SHA` with the same meaning as GitHub. | Gitea documentation | measured in W5 on the lab's Gitea |
| CA-5 | A container action runs only on Linux runners. | GitHub documentation | quoted |
| CA-6 | reviewdog reads SARIF and posts a review through `github-pr-review` and `gitea-pr-review`. | reviewdog documentation | measured in W5 |
| CA-7 | GitHub only warns about an action input the action does not define; it does not fail the step. | advisor's reading | measured in W4 |

## §2 WAGONS

Executors: Cursor seats, `grok-4.7-high` by default, the master may step a wagon to `medium` or `xhigh` (the owner,
2026-10-21). `--model` is on the dispatch line. Every session is recorded (ADR-0024); a session that was not recorded says
so in its report, with the reason. Reviews on `codex` with `--role review`, the focus phrased as the property that must
hold (law #9).

### W1 `the-command-line`
- **what:** a Go module, `proven scan [TARGET]` and `proven ci`, every flag and environment variable of ADR-0020 §1, the
  exit codes of §2, and the interface between the shell and the reviewer: a request (target, baseline, include/exclude,
  severity, timeout, sample) and a result (findings, each with its evidence and an `unresolved` mark; the services that
  failed). The sample reviewer behind it returns findings from a fixture. The layout of the module is the master's choice
  and is written down in the repository's `AGENTS.md`.
- **dod:** a table test per flag and per variable, flag over variable over default; `proven ci` reads the three
  `GITHUB_*` variables from a fixture event. **Negative controls (law #27):** an unknown flag → refused with the list of
  known ones; `--config=auto` → refused with what `--config` takes, exit 7; an unknown value of `PROVEN_SEVERITY` →
  refused with the allowed list, exit 7; a fixture with findings under `--error` → exit 1, without → 0; a reviewer that
  panics → exit 2; a reviewer that reports a failed service → exit 0 or 1, never 2, and the service is named.
- **master-was-wrong:** —

### W2 `three-formats`
- **what:** text, JSON and SARIF 2.1.0 from the same result; `-o` and the three `--*-output=` files at once.
- **dod:** the SARIF validates against the official 2.1.0 schema in the test; severity maps to `note`/`warning`/`error`;
  every finding's evidence is in the SARIF result (where it goes is the wagon's choice, written down). **Negative
  control:** a SARIF with a removed required field fails the schema test. **Unplug (law #41):** a format writer that
  writes nothing turns the test red, not green on an empty file.
- **master-was-wrong:** —

### W3 `the-image`
- **what:** Proven's container image, built from the repository, with nothing but the binary and what it needs; tagged by
  commit.
- **dod:** `docker run <image> --version` and `docker run <image> scan --sample /fixture` give the same output as the
  binary. The image's size printed.
- **master-was-wrong:** —

### W4 `the-github-action-and-the-swap`
- **what:** `action.yml` at the repository root as ADR-0020 §3 says: inputs, outputs, `unresolved-count`, permissions of
  §5. Then **the swap test, the point of the whole train:** a real Semgrep workflow and a real Opengrep-action workflow,
  each changed only in the lines ADR-0020 §4 names, run on GitHub against a fixture pull request in sample mode.
- **dod:** both swapped workflows pass; the diff of each workflow is pasted and counted in lines; the outputs a later
  step reads keep their names; SARIF reaches code scanning under `category: proven`. CA-7 measured: the step with
  `patterns:` still set warns and runs. **Negative control:** a later step that reads `findings-count` sees the fixture's
  count; with the output renamed, that step fails.
- **master-was-wrong:** —

### W5 `comments-on-github-and-gitea`
- **what:** with `comment: true`, reviewdog posts the review from Proven's SARIF; on GitHub with `github-pr-review`, on
  Gitea with `gitea-pr-review`. Proven has no comment code of its own (ADR-0020, point 7). CA-4 and CA-6 measured.
- **dod:** the same fixture pull request gets the same comments on GitHub and on the lab's Gitea, screenshots of both.
  **Negative control:** with `comment: false`, no comment and no `pull-requests: write` asked for.
- **master-was-wrong:** —

### W6 `the-self-hosted-runner-script`
- **what:** ADR-0021's one script: takes the repository's name and a token, starts an ephemeral runner in Docker,
  registers it, commits the workflow, opens a pull request, shows the review. The runner's properties of ADR-0020 §6:
  one job then gone; pull requests from forks wait for a maintainer; no secret beyond Proven's own; a narrow network.
- **dod:** run once against a test repository, the pull request's review shown. The token appears in no file, no log, no
  argument of a process, measured by a search after the run. **Negative controls:** a second job on the same runner is
  refused because the runner is gone; a token without the needed scope is refused by name before anything is changed.
  The numbered runbook of the human steps is in the documentation (law #39).
- **review:** named — the last wagon carries the stream's review (law #29); intermediate wagons are master-verified.
- **master-was-wrong:** —

## §3 DECISIONS OWED BY THE OWNER

- **C1 — the master.** Recommended: the infrastructure master, who already runs the Cursor seats and the test corpus,
  so the build and the seats have one head.
- **C2 — the test repository for W4 and W6** on GitHub: a new public repository under `saphan-ai`, or a private one.
  Recommended: private, because a self-hosted runner on a public repository runs code from forks (ADR-0020 §6).

## §4 DEFINITION OF DONE

- A Semgrep pipeline and an Opengrep-action pipeline both run Proven after changing only the lines ADR-0020 §4 names, on
  GitHub; the diffs are published.
- Text, JSON and a schema-valid SARIF from one run.
- The same review is commented on GitHub and on Gitea, through reviewdog.
- One script turns a judge's repository name and token into a reviewed pull request, and the token is nowhere afterwards.
- Every refusal above prints what was refused and what is allowed; every wagon's report has its run id and start time
  (the evidence that it ran inside the window), and its session recording or the reason it has none.

## §5 COST

- Cursor seats: subscription; the scarce quantity is each seat's allowance (law #51). Each wagon's quote is the master's,
  before dispatch, and its cap is twice the quote (law #16).
- Reviews on `codex`: quoted by the master at the review order.
- GitHub Actions minutes on the test repository: within the free allowance of the account; measured in W4.
