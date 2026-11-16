# GitHub Action

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
          fetch-depth: 0
          persist-credentials: false
      - uses: saphan-ai/proven@<full commit SHA>
        with:
          baseline-commit: ${{ github.event.pull_request.base.sha }}
          strict: true
          require-complete: true
          comment: true
        env:
          PROVEN_EVIDENCE_STORE: sqlite
          PROVEN_EVIDENCE_FILE: /path/to/prepared/evidence.sqlite
          PROVEN_GENERATOR_MODE: openai
          PROVEN_GENERATOR_BASE_URL: ${{ vars.PROVEN_MODEL_URL }}
          PROVEN_GENERATOR_MODEL: ${{ vars.PROVEN_MODEL }}
          PROVEN_GENERATOR_KEY: ${{ secrets.PROVEN_MODEL_KEY }}
```

Prepare and mount an evidence database outside the repository under review. The path in the example is a placeholder, not a bundled feed. Omit model configuration to select deterministic analysis. For an account-free recorded demonstration, set `sample-mode: true` and `target: /opt/proven/fixture/discount`.

`strict` fails on findings. `require-complete` independently fails with exit 3 on partial review or a failed service. It defaults to true; set it false only for an explicitly advisory check. The result is still written when incomplete.

Outputs are `results-file`, `json-file`, `sarif-file`, `findings-count` and `unresolved-count`. Comments are opt-in and update one summary comment. The posting token is removed from the environment of the review process. Fork pull requests need a workflow that does not expose secrets to untrusted code.

Gitea Actions can use the same entrypoint and outputs. Containerized proving requires a correctly configured restricted runtime; the Action does not grant that automatically.

## Proven reviewing itself

This repository's `self-review.yml` workflow reviews same-repository pull requests with NVIDIA Nemotron through the ordinary OpenAI-compatible adapter. `NVIDIA_API_KEY` is a repository secret. `PROVEN_REVIEW_IMAGE` is a repository variable containing a published `saphan/proven@sha256:...` image digest; updating it explicitly selects the reviewer version.

The workflow extracts the CLI from that image and runs the base revision's Action entrypoint on the host, so the restricted test runner can mount the correct host paths. The PR is checked out separately at its exact head SHA. Dependencies are prepared from the trusted base with `Dockerfile.runner` before tests run offline. Only the review step receives the model key and comment token; test containers receive neither.

Findings fail the check, as does incomplete analysis. One summary comment is updated on each run, and JSON, SARIF and the redacted trace are retained as artifacts. The SQLite database starts without advisory feeds: a requested missing source is reported as unavailable, rather than treated as a clean dataset. This workflow exercises code review; comprehensive dependency review requires imported advisory snapshots. Fork PRs are skipped because they have no access to the model key.
