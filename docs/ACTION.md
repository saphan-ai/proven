# GitHub Action

Start **[without a model](MODEL_FREE.md)** for deterministic dependency checks, source observations and description checks. For subscription CLIs, use **[Start with Proven](START.md)**. The [`proven/cli` preset](../cli/action.yml) handles your customer image, saved login, test preparation, reports and cleanup. The reference below describes the generic API/model-free/sample Action and this repository's API self-review.

See the [CLI Action input reference](CLI_ACTION.md) for advanced settings and runner boundaries.

Proven's Action builds the CLI from the source revision referenced by `uses`. It does not pull a mutable prebuilt Proven image. Pin the Action to a full reviewed commit SHA.

```yaml
on: [pull_request]
permissions:
  contents: read
  pull-requests: write
jobs:
  review:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v7
        with:
          ref: ${{ github.event.pull_request.head.sha }}
          fetch-depth: 0
          persist-credentials: false
      - uses: saphan-ai/proven@<full commit SHA>
        with:
          baseline-commit: ${{ github.event.pull_request.base.sha }}
          refresh-data: true
          check-delivery: true
          strict: true
          require-complete: true
          comment: true
        env:
          PROVEN_EVIDENCE_STORE: sqlite
          PROVEN_GENERATOR_MODE: openai
          PROVEN_GENERATOR_BASE_URL: ${{ vars.PROVEN_MODEL_URL }}
          PROVEN_GENERATOR_MODEL: ${{ vars.PROVEN_MODEL }}
          PROVEN_GENERATOR_KEY: ${{ secrets.PROVEN_MODEL_KEY }}
```

`refresh-data: true` acquires and imports scoped public data into the default SQLite store outside the reviewed checkout. See [source refresh](SOURCE_REFRESH.md) for cache persistence, provenance, partial coverage and PostgreSQL. You can instead import snapshots yourself and omit the switch. Set `PROVEN_GENERATOR_MODE: off` for deterministic analysis without model credentials. For an account-free recorded demonstration, set `sample-mode: true` and `target: /opt/proven/fixture/discount`; do not refresh in sample mode.

`check-delivery: true` collects the PR description, commit messages and discussion/review/inline/commit comments. Each gets its own assessment and source identity. With a model, Proven checks claims against exact head-code citations and existing proved findings. Without a model, it reports missing, placeholder or very short descriptions mechanically. See [declared scope](PR_DELIVERY.md). Check out the exact PR head for this feature, rather than GitHub's synthetic merge commit:

```yaml
      - uses: actions/checkout@v7
        with:
          ref: ${{ github.event.pull_request.head.sha }}
          fetch-depth: 0
          persist-credentials: false
```

Both switches default to `false`. `github-token` defaults to the job token. It needs read access for context collection and `pull-requests: write` only when posting the summary. Collection failures and limits remain visible as incomplete coverage.

`strict` fails on findings. `require-complete` independently fails with exit 3 on interrupted execution, a failed service or remaining planning hypotheses. It defaults to true; set it false only for an explicitly advisory check. The result is still written for these incomplete outcomes. Fatal failures such as mid-plan carry overflow can still return exit 2 with only stderr and a trace. See [readiness](READINESS.md).

Outputs are `results-file`, `json-file`, `sarif-file`, `findings-count` and `unresolved-count`. Comments are opt-in and update one summary comment. The posting token is removed from the environment of the review process. Fork pull requests need a workflow that does not expose secrets to untrusted code.

Gitea Actions can use the same entrypoint and outputs. Containerized proving requires a correctly configured restricted runtime; the Action does not grant that automatically.

## Proven reviewing itself

For a guided visual introduction, see [GitHub screenshots](GITHUB_TOUR.md). [API credentials](CONNECT_API.md) includes secret creation and replacement; [CLI images](CLI_IMAGES.md) describes customer-owned subscription tooling rather than a bundled official image.

This repository's `self-review.yml` workflow reviews same-repository pull requests through the ordinary OpenAI-compatible adapter. It currently defaults to NVIDIA Nemotron. `PROVEN_REVIEW_BASE_URL` and `PROVEN_REVIEW_MODEL` select the endpoint and model. The secret `PROVEN_REVIEW_API_KEY` takes precedence over the existing `NVIDIA_API_KEY`. `PROVEN_REVIEW_IMAGE` is a repository variable containing a published `saphan/proven@sha256:...` image digest; updating it explicitly selects the reviewer version.

The self-review profile uses 300s per attempt, 40s to headers, 90s idle, 8192 tokens (Planner 4096, Investigator/Skeptic 2048), a 60s finalization reserve and three rate-limit retries. Repository variables `PROVEN_REVIEW_CALL_TIMEOUT`, `PROVEN_REVIEW_HEADER_TIMEOUT`, `PROVEN_REVIEW_IDLE_TIMEOUT`, `PROVEN_REVIEW_MAX_TOKENS`, `PROVEN_REVIEW_AGENT_MAX_TOKENS`, `PROVEN_REVIEW_DEADLINE_RESERVE`, `PROVEN_REVIEW_RATE_LIMIT_RETRIES` and `PROVEN_REVIEW_RECOVERY_RETRIES` override those choices. `PROVEN_REVIEW_INCLUDE_USAGE`, `PROVEN_REVIEW_EXTRA_BODY` and `PROVEN_REVIEW_AGENT_EXTRA_BODY` enable endpoint-specific options after checking compatibility. The whole review remains bounded at 1500s and the job at 35 minutes. See [call limits](CONFIGURATION.md#call-limits-and-provider-options).

`PROVEN_REVIEW_MAX_TOKENS_FIELD` selects `max_tokens` or `max_completion_tokens` for both probes and review. `PROVEN_REVIEW_MARKOVIAN=off` selects whole-change planning when the model can hold the changed files together; `on` (default) retains bounded steps. This switch does not change the selected files, proof requirements or completion policy. Final open claims are reported as `unresolved_hypotheses` while known risks continue to evidence checks. Their presence makes completion incomplete and fails a required check with exit 3. See [bounded planning](MARKOVIAN.md).

`PROVEN_REVIEW_RETRY_BACKOFF` and `PROVEN_REVIEW_RETRY_MAX_BACKOFF` override the initial and maximum pre-stream retry delays (defaults 250ms and 5s). A burst-limited endpoint can need longer intervals; for example, 5s initial and 30s maximum yield 5/10/20s for three rate-limit retries when no Retry-After is supplied. Valid server hints and the attempt/review deadlines still take precedence. These settings do not retry a timeout or a broken stream.

An optional `PROVEN_MODEL_PROBES` variable enables serial compatibility probes. It is a JSON array such as `[{"name":"variant-a","model":"your-model","extra_body":"{\"temperature\":0}"}]`. Each entry makes one request with no retries and saves measurements as an artifact. HTTP failures or invalid answers are recorded as unsuccessful measurements; malformed configuration fails the job. Remove this variable after the experiment to avoid repeating live probes on every PR. A probe establishes API compatibility and latency for a short prompt, not review quality.

The workflow extracts the CLI from that image and runs the base revision's Action entrypoint on the host, so the restricted test runner can mount the correct host paths. The PR is checked out separately at its exact head SHA. Dependencies are prepared from the trusted base with `Dockerfile.runner` before tests run offline. Only the review step receives the model key and comment token; test containers receive neither.

Findings fail the check, as does incomplete analysis. One summary comment is updated on each run, and JSON, SARIF and the redacted trace are retained as artifacts. The SQLite database starts without advisory feeds: a requested missing source is reported as unavailable, rather than treated as a clean dataset. This workflow exercises code review; comprehensive dependency review requires imported advisory snapshots. Fork PRs are skipped because they have no access to the model key.
