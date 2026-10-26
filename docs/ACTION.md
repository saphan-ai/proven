# The GitHub Action

`action.yml` at the root of this repository makes Proven a step a team swaps in for a Semgrep or an Opengrep step. The names and meanings are decided in [ADR-0020](adr/ADR-0020-proven-command-line-and-action-interface.md), points 3 to 5. This page is the use, the two before-and-after workflows, and the runbook of the swap test the owner runs on GitHub.

What is measured here, and what is not: the Action's script and `action.yml` are tested without GitHub (`sh action/test.sh`, see [Testing](#testing)). **No run on GitHub has happened.** The workflows below were written from the two tools' published documentation, not run; the runbook is how they get run.

## Use

```yaml
- id: scan
  uses: saphan-ai/proven@<full commit SHA>
  with:
    target: .
    sample-mode: true
```

A container action runs on Linux runners only: a macOS or Windows runner runs the command line in a `run:` step instead ([ADR-0020](adr/ADR-0020-proven-command-line-and-action-interface.md) §6).

### Inputs

| input | default | maps to |
|---|---|---|
| `target` | `.` | the `proven scan` target, taken from the workspace |
| `config` | none | `--config`, a path in the workspace to **Proven's** configuration file |
| `output-format` | `json/sarif` | `json/sarif`, `json`, `sarif` or `text`: which file `results-file` names and which format the log shows (below) |
| `severity` | `INFO` | `--severity` |
| `include` · `exclude` | none | `--include` · `--exclude`; comma-separated globs, one flag each; a space beside a comma is part of the glob |
| `timeout` | none | `--timeout`, seconds |
| `strict` | `true` | `true` → `--error`, `false` → `--no-error` |
| `baseline-commit` | the pull request's base (`github.event.pull_request.base.sha`) | `--baseline-commit`; empty reviews the whole target |
| `sample-mode` | `false` | `true` → `--sample` |
| `comment` | `false` | `true` posts ONE summary comment on the pull request, edited in place on every run ([below](#the-comment)); `false` posts nothing and reads no token; with either value the `proven scan` process runs without the runner's credentials |
| `github-token` | the workflow's token (`github.token`) | the token the summary comment is posted with; read only when `comment` is `true` |

`github-token` is the twelfth input: ADR-0020 §3 gained it on 2026-10-25, because §7 posts the comment and posting needs a token.

`true` and `false` are the only values `strict`, `sample-mode` and `comment` take; any other value, and any other `output-format`, ends the step with exit 2 and a message naming the input.

The Opengrep action's `patterns`, `jobs` and `opengrep-version` are not inputs. GitHub only warns about an input an action does not define, so a workflow that still passes one runs and the input does nothing; remove it.

`config` keeps its name, not its meaning: Opengrep's `config` is a rules file, Proven's is `proven.yaml` ([ADR-0013](adr/ADR-0013-no-hardcoded-settings.md)). A Semgrep registry value such as `auto` or `p/…`, or a rules file, is refused by Proven with exit 7, and the step fails with that code.

### Outputs

| output | value |
|---|---|
| `results-file` | the file for `output-format`: `json` → the JSON file; `text` → the text file; `sarif` and `json/sarif` → the SARIF file |
| `json-file` | `proven-results.json` |
| `sarif-file` | `proven-results.sarif` |
| `findings-count` | the number of findings in the JSON file |
| `unresolved-count` | the findings in the JSON file whose `unresolved` is `true`: their evidence did not settle them |

Every path is **relative to the workspace**: the step runs inside a container, where the workspace is at another path than on the runner, and a later step on the runner reads these. The JSON and SARIF files are written on every run, whatever the `output-format`, because the counts are read from the JSON; the text file is written for `text`. A file left by an earlier run is removed first. When Proven could not review (exit 2 or 7) no JSON is written and both counts are `0`: **read the step's outcome before the count**, because a count of `0` is also what a clean review prints.

The log shows the text review for `json/sarif` and `text`, the JSON for `json`, the SARIF for `sarif`.

### Exit codes

The step's exit code is Proven's ([ADR-0020](adr/ADR-0020-proven-command-line-and-action-interface.md) §2): `0` done, with no findings or with `strict: false`; `1` findings with `strict: true`; `2` Proven could not review; `7` the configuration was refused. The outputs are written before the step exits, so a step with `if: always()` can read them after a failure.

### Permissions

GitHub grants permissions to a **job**, never to a step. So the scan job has `contents: read` and nothing else; `pull-requests: write` is added to that job only when `comment` is `true`; and `security-events: write` belongs to a **second job**, the SARIF upload, that `needs:` the scan job and receives the SARIF as an artifact. Checkout and the scan never run with `security-events: write`. The Action asks for nothing more.

### The comment

With `comment: true` the Action writes **one comment per pull request**, the owner's order of 2026-10-25: the summary below, written by the Action's own poster (`action/post-comment.sh`) and edited in place on every run. It writes no line comments; the SARIF file is the place for per-line results (code scanning, as in the swap test below). The Action works the same on GitHub (and GitHub Enterprise Server) and on Gitea Actions, whose `GITEA_ACTIONS=true` only matters to the posting account (below). The token is `github-token`, by default the workflow's own, and the job needs `pull-requests: write`:

```yaml
    permissions:
      contents: read
      pull-requests: write
    # ...
      - id: scan
        uses: saphan-ai/proven@<full commit SHA>
        with:
          comment: true
```

The comment is Proven's [Markdown](OUTPUT_FORMATS.md#markdown) (a verdict line, a findings table, each finding's evidence in a collapsible block, and always the statements the text review prints and every service that failed). It is edited in place on every run, never added again. `action/post-comment.sh` lists the pull request's comments page by page until a page holds none, takes the first one written by the posting account whose body starts with `<!-- proven-review`, and edits it (`PATCH /repos/{repo}/issues/comments/{id}`); it creates one only when the whole list was read and there is none. A walk that is cut off (more than `PC_MAX_PAGES` pages, 500 by default, or a server that repeats a page) creates nothing and fails the step with the reason. A comment with that marker by another account is never edited. A result too large for the platforms' 65,536-character limit is cut per finding, with the cut counted in the comment; the full evidence is in the JSON and SARIF files of the run. The comment's first line is `<!-- proven-review run=<run number> sha=<head commit of the pull request> -->`, read from `GITHUB_RUN_NUMBER` and the event file. The posting account is the one `GET /user` names; where the token may not read it (GitHub's installation token) it is `github-actions[bot]`, and on Gitea, where there is no such default, set `PROVEN_COMMENT_ACCOUNT` in the step's `env`. The same token rules apply: the summary is posted after the scan with the same `github-token`, which the scan never sees. The image carries `curl` for the `PATCH`. When the summary cannot be posted the step fails with exit 2 and the line `posting the summary comment on the pull request failed`; a strict review with findings keeps exit 1. A run with no pull request in its event says so on standard error and posts nothing. Not measured: a post on GitHub or on Gitea (the tests use a stand-in for the server, with the JSON shapes of both).

Without `comment: true` nothing is posted and the token input is not read. With either value, the `proven scan` process runs with those credentials removed from its environment (`github-token` in both spellings, `GITHUB_TOKEN`, `ACTIONS_RUNTIME_TOKEN`, `ACTIONS_ID_TOKEN_REQUEST_TOKEN` and its URL). The token is still in the container: the runner sets the inputs on its first process. The code under review runs only in Proven's own test container, which receives no token (ADR-0025). The posting token goes only to the poster, as the `Authorization` header of its requests. When posting fails the step fails with exit 2 and the line `posting the summary comment on the pull request failed`; a strict review with findings keeps exit 1. When Proven could not review (exit 2 or 7) there is no Markdown and nothing is posted. Not measured: any post on GitHub or on Gitea (the tests use a stand-in server; runbook step 6 is the real check). A pull request from a fork gets a read-only token, so there the post fails with the message above.

### The image

`action.yml` builds `action/Dockerfile`: Proven's published image `saphan/proven` plus `curl` plus the scripts `action/entrypoint.sh` and `action/post-comment.sh`, run as root. Root is GitHub's rule for a container action, which must write the workspace and `$GITHUB_OUTPUT` that the runner owns; the published image's user 65532 cannot.

The image reference is `ARG PROVEN_IMAGE` in `action/Dockerfile`, and it reads `saphan/proven:latest` in this commit because nothing has been published yet to name. **A tag moves, a digest does not**: pin it to the digest (`saphan/proven@sha256:…`) before anyone writes `uses: saphan-ai/proven@<full commit SHA>` for real, so the commit SHA fixes the image as well as the script. Runbook step 1 does it.

`include`, `exclude` and the rest are tested at their mapping with a stand-in for `proven`, because the sample fixture's four findings are all `ERROR` and one fixture cannot show a flag whose effect it does not exercise. The sample reviewer reads `fixture.json` in the target, so `sample-mode: true` needs those files in the target (the fixture is at `testdata/fixture/discount`).

## Before and after

Both workflows are real shapes of the tools' documented jobs. `<full commit SHA>` stands for the 40-character SHA of the action named, which is the form GitHub recommends; it is not filled in here because this page was written without a network. The `fetch-depth: 0` line is in both "before" workflows: both tools, like Proven, need history to review only what a pull request changed.

### Semgrep → Proven

Before:

```yaml
name: Code review scan
on:
  pull_request:
jobs:
  scan:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      security-events: write
    container:
      image: semgrep/semgrep
    steps:
      - uses: actions/checkout@<full commit SHA>
        with:
          fetch-depth: 0
      - run: semgrep ci --sarif-output=results.sarif
        env:
          SEMGREP_APP_TOKEN: ${{ secrets.SEMGREP_APP_TOKEN }}
      - if: always()
        uses: github/codeql-action/upload-sarif@<full commit SHA>
        with:
          sarif_file: results.sarif
```

After (the whole file, 39 lines; the SARIF upload is its own job, the only one with `security-events: write`):

```yaml
name: Code review scan
on:
  pull_request:
jobs:
  scan:
    runs-on: ubuntu-latest
    permissions:
      contents: read
    container:
      image: saphan/proven
    steps:
      - uses: actions/checkout@<full commit SHA>
        with:
          fetch-depth: 0
      - run: proven ci --sarif-output=results.sarif
        env:
          METERLESS_RUNTIME_TOKEN: ${{ secrets.METERLESS_RUNTIME_TOKEN }}
          ZETARIS_API_KEY: ${{ secrets.ZETARIS_API_KEY }}
          ZETARIS_ORG_ID: ${{ secrets.ZETARIS_ORG_ID }}
      - if: always()
        uses: actions/upload-artifact@<full commit SHA>
        with:
          name: sarif
          path: results.sarif
  upload:
    needs: scan
    if: always()
    runs-on: ubuntu-latest
    permissions:
      contents: read
      security-events: write
    steps:
      - uses: actions/download-artifact@<full commit SHA>
        with:
          name: sarif
      - uses: github/codeql-action/upload-sarif@<full commit SHA>
        with:
          sarif_file: results.sarif
          category: proven
```

**Counted with `diff -U3` against the "before": 27 changed lines, 5 removed and 22 added, in a file that goes from 22 lines to 39.** ADR-0020 §4's lines are in it: the image, `semgrep ci` → `proven ci`, `SEMGREP_APP_TOKEN` → Proven's secrets, and `category: proven`. The rest is the job split: `security-events: write` leaves the scan job, and the upload becomes a second job that receives the SARIF as an artifact. The first draft of this page left the permission on the whole job and counted 9 lines; the split is the honest price of giving checkout and the scan only `contents: read`. To post a comment too, use the Action and add `pull-requests: write` to the scan job (see [The comment](#the-comment)). This workflow runs `proven ci` directly in a job container, not the Action: it is the other half of the swap, the image. A team on sample mode has no secrets to set and drops the `env:` block.

Not measured, and the swap test looks at both: that a job container started from the published image, whose user is 65532, can run `actions/checkout` and write the workspace (Semgrep's image runs as root). If it cannot, the fix is `options: --user root` under `container:`, one more line, and the swap test records it as a line ADR-0020 §4 did not name.

### Opengrep action → Proven

Before:

```yaml
name: Code review scan
on:
  pull_request:
jobs:
  scan:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      security-events: write
    steps:
      - uses: actions/checkout@<full commit SHA>
        with:
          fetch-depth: 0
      - id: scan
        uses: platform-sec/opengrep-action@<full commit SHA>
        with:
          target: src
          patterns: rules/
          output-format: json/sarif
          severity: WARNING
          exclude: vendor/*
          timeout: 600
          strict: true
          baseline-commit: ${{ github.event.pull_request.base.sha }}
      - run: echo "${{ steps.scan.outputs.findings-count }} findings"
      - if: always()
        uses: github/codeql-action/upload-sarif@<full commit SHA>
        with:
          sarif_file: ${{ steps.scan.outputs.sarif-file }}
```

After (the whole file, 43 lines):

```yaml
name: Code review scan
on:
  pull_request:
jobs:
  scan:
    runs-on: ubuntu-latest
    permissions:
      contents: read
    steps:
      - uses: actions/checkout@<full commit SHA>
        with:
          fetch-depth: 0
      - id: scan
        uses: saphan-ai/proven@<full commit SHA>
        with:
          target: src
          output-format: json/sarif
          severity: WARNING
          exclude: vendor/*
          timeout: 600
          strict: true
          baseline-commit: ${{ github.event.pull_request.base.sha }}
      - run: echo "${{ steps.scan.outputs.findings-count }} findings"
      - if: always()
        uses: actions/upload-artifact@<full commit SHA>
        with:
          name: sarif
          path: ${{ steps.scan.outputs.sarif-file }}
  upload:
    needs: scan
    if: always()
    runs-on: ubuntu-latest
    permissions:
      contents: read
      security-events: write
    steps:
      - uses: actions/download-artifact@<full commit SHA>
        with:
          name: sarif
      - uses: github/codeql-action/upload-sarif@<full commit SHA>
        with:
          sarif_file: proven-results.sarif
          category: proven
```

**Counted with `diff -U3` against the "before": 24 changed lines, 5 removed and 19 added, in a file that goes from 29 lines to 43.** The Proven-specific changes are 4 of them (the `uses:` line, `patterns` removed, `category: proven`; the `uses:` change counts once removed and once added); the other 20 are the job split that keeps `security-events: write` off the scan job. The first draft of this page counted 4 and left the permission on the whole job. The upload job names `proven-results.sarif` because that is the file name the Action always writes; the artifact carries it under that name. To post a comment, add `pull-requests: write` to the scan job's permissions and `comment: true` to the step (see [The comment](#the-comment)).

`target`, `output-format`, `severity`, `exclude`, `timeout`, `strict` and `baseline-commit` stay as they are; `patterns` goes; the outputs a later step reads (`findings-count`, `sarif-file`) keep their names. A workflow that passes `config:` keeps the line but changes its value to Proven's configuration file (see Inputs).

### SARIF into code scanning

`github/codeql-action/upload-sarif` with `sarif_file` set to Proven's SARIF and `category: proven`, so it does not replace another tool's analysis of the same commit. The upload job, and only it, has `security-events: write`. Code scanning for a private repository needs GitHub Advanced Security on it.

## The swap test the owner runs

This is an act of the owner's hand, on GitHub, after the push. It is not run here. It answers one question: does a team change the lines ADR-0020 §4 names, plus the job split this page counts, and get the same job?

1. **Publish the image, pin it.** (a) The image carries no downloaded tool to pin: `curl` comes from the base image's package index. (b) `make push` publishes `saphan/proven` (`docs/LAB.md`). Read the digest of the tag it wrote: `docker buildx imagetools inspect saphan/proven:<tag>`; the `Digest:` line is the value. In `action/Dockerfile` set `ARG PROVEN_IMAGE=saphan/proven@sha256:<that digest>`. (c) Commit, push, and read the commit SHA with `git rev-parse HEAD`. **That SHA is the one every `uses: saphan-ai/proven@<full commit SHA>` below carries.** The repository must be one the test repository may read (public, or the same organisation).
2. **Make the test repository.** A new repository on GitHub, **public, or with GitHub Advanced Security on**, because step 5 reads Code scanning. Put the fixture at its root: `cp -R testdata/fixture/discount/. <the new repository>/` (it has `base`, `change`, `fixture.json` and `tables.sql`), committed on `main`.
3. **Add the Opengrep-shaped workflow first.** In `.github/workflows/proven.yml`, the Opengrep "after" workflow above (both jobs), with these lines changed for the test (they are test settings, not part of the swap): `target: .`, remove `exclude`, add `sample-mode: true` and `strict: false`; leave `baseline-commit` as it is, so step 5 runs the default. Sample mode needs no secret. Commit it on `main`.
4. **Open the fixture pull request.** Branch from `main`, add a file `note.txt` (any change), open a pull request. Wait for both jobs, `scan` and `upload`.
5. **Read each step.**
   - *Checkout*: green.
   - *The Proven step, with the default baseline*: this is the path every real pull request takes: `git` runs as root in the container on a checkout the runner owns. Record which of these happens: the step resolves the pull request's base and reviews only `note.txt`'s change (green; the fixture's files did not change, so `findings-count` is `0`, which is correct here); or `git` refuses the checkout as owned by someone else ("dubious ownership"); or the step exits 2. Either of the last two is a defect to record, not to patch quietly. Then re-run once with `baseline-commit: ''` added to the step: it must be green and its log must show the text review with a line `finding F1 …` among four findings (`F2`, `F4`, `osv:evil.example/leftpad`, `F1`), and the `echo` step prints `4 findings`. **How to read `findings-count`:** `0` also means Proven could not review (the step would then be red, exit 2 or 7, so look there first). `unresolved-count` is `0` for this fixture.
   - *The `upload` job*: green, and the repository's Security tab, Code scanning, lists the four alerts under the category `proven`. In the workflow run's summary, the `scan` job's token permissions list `Contents: read` and nothing else, and the `upload` job's list `Security events: write`.
6. **Check the comment.** In the workflow add `pull-requests: write` to the `scan` job's permissions and `comment: true` to the Proven step (keep `baseline-commit: ''`), push. Push and wait for the run. On the pull request's Conversation tab exactly ONE comment by the workflow's bot appears, starting with the verdict line (`Proven: blocked`, with the findings table and a collapsible block per finding), and no line comment. Push again: the same comment is edited, no second one appears. Record: whether the comment appears; if it does not, the exact text after `posting the summary comment on the pull request failed` in the step's log. On a Gitea instance run the same workflow and record the same two things; the Gitea path is not measured here. Then remove `comment: true` and push: the comment is not edited and no new one appears.
7. **Run it strict.** Remove `strict: false` (the default is `true`) on the pull request's branch and push. The Proven step must turn **red** with exit code `1` in its log, and the `upload` job, which has `if: always()`, must still run and stay green. A strict step that stays green on four findings is a defect: record it.
8. **Run the swap itself, the other way.** In a second pull request, replace the workflow with the Semgrep workflow in its "after" form (the 27-line change above) and run it. It runs `proven ci`; for the test add `SAMPLE_MODE: "true"` to the step's `env:` so it needs no secret. What `proven ci` prints on the fixture pull request is not measured here: record it. This tests the image half of the swap; record whether the job container needed `options: --user root` and any other line beyond the 27.
9. **Record the result.** For each step: green or red, and the exact text of any difference from what this page says. A line the swap needed that ADR-0020 §4 did not name goes back to the ADR as a defect in its claim, not into the workflow quietly.


## Testing

`sh action/test.sh` runs from any directory. It builds `proven` (or uses `PROVEN_BIN`), checks `action.yml` for exactly the inputs and outputs of ADR-0020 §3, then runs `action/entrypoint.sh` against the sample fixture with a temporary `GITHUB_WORKSPACE` and `GITHUB_OUTPUT` and asserts each output, the exit codes, and each input's flag. With `comment` true, a run with findings is shown to write exactly one comment-writing request (no line comments), and a missing token is refused; a stand-in `proven` records any token in its environment, which shows the scan runs with none whether `comment` is true or false; a stand-in server (`COMMENT_HTTP`) records each request of the summary comment, which is how "one comment, edited in place" is tested: an earlier comment of the account is edited and none is added, another account's marker comment is not edited, a failed post fails the step; a working directory holding `vendor/a` shows that a glob in `include` or `exclude` reaches Proven unexpanded. Its negative control: rename `findings-count` in the script's `GITHUB_OUTPUT` writes and seven checks turn red; the script then goes back.
