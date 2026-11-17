# Start with Proven

Proven reviews a pull request and attaches evidence to its findings. Bring a CLI account you already use, build your own image and connect one workflow.

**No model account?** [Start with the generic image](MODEL_FREE.md): dependency/data rules, source refresh and description checks already work without a model.

Before you start: [what Proven checks and proves](WHAT_PROVEN_CHECKS.md) explains the evidence in a table. [These screenshots](GITHUB_TOUR.md) show the workflow and PR comment.

## 1. Choose where it runs

| Where | How the CLI logs in | Start here |
| --- | --- | --- |
| **Your own runner** | Complete the CLI's browser/device login once. Its session stays on that machine, outside the image. | Follow steps 2–4 below. This is the simplest subscription path. |
| **GitHub-hosted runner** | Each job starts fresh. It needs credentials provisioned for that job; an image alone does not carry a login. | See [GitHub-hosted Claude](#github-hosted-claude). |
| **API endpoint** | Use your endpoint's API credential. | [Connect an API](CONNECT_API.md). |

To see a report without connecting anything, try the [recorded demo](QUICKSTART.md).

## 2. Prepare your runner once

Use a **Linux or macOS self-hosted Actions runner** with Docker, Git and Python 3. In your repository, open **Settings → Actions → Runners → New self-hosted runner**. Follow GitHub's commands and add the label **`proven`**. [GitHub's setup instructions](https://docs.github.com/en/actions/how-tos/manage-runners/self-hosted-runners/add-runners) cover installation.

Run this on that same machine, as the account running the Actions service:

```sh
git clone https://github.com/saphan-ai/proven.git
cd proven
sh scripts/proven-cli.sh build claude
sh scripts/proven-cli.sh login claude
sh scripts/proven-cli.sh check claude
```

- **Build** installs the licensed CLI into an image you build and control. Proven publishes recipes; it does not publish those vendor-CLI images.
- **Login** starts Claude's own OAuth flow. Finish it in your browser. The session stays in the private `proven-claude-home` Docker volume.
- **Check** makes one short model request using your account quota. It checks the connection, without reviewing a project.

Image and login volume must be on the Docker daemon used by the runner. Complete the login inside that image, including on a Mac; a desktop keychain login is a separate environment. Scan your own image before deployment: [image checks](CLI_IMAGES.md#2-scan-the-image-you-built).

## 3. Add the workflow

In the repository **you want reviewed**, save `.github/workflows/proven.yml`:

```yaml
name: Proven
on: [pull_request]
permissions:
  contents: read
  pull-requests: write
jobs:
  review:
    if: github.event.pull_request.head.repo.full_name == github.repository
    runs-on: [self-hosted, linux, proven]
    timeout-minutes: 35
    steps:
      - uses: actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1
        with:
          ref: ${{ github.event.pull_request.head.sha }}
          fetch-depth: 0
          persist-credentials: false
      - uses: saphan-ai/proven/cli@main
        with:
          backend: claude
          image: proven-claude:local
```

Open a PR. The Action checks the model connection, prepares offline test dependencies from the **base revision**, reviews the change, updates one comment and keeps report artifacts.

For a Mac runner, change `linux` to `macOS` in `runs-on`. The Action prepares its own reviewer toolchain; generated project tests still run in restricted Docker containers.

The example uses `main` for the first setup. Pin a reviewed full Proven commit SHA for retained deployments. Use a dedicated self-hosted runner for repositories and contributors you trust; this example skips fork PRs. Proof isolation covers tests started by Proven, not unrelated workflow steps.

## 4. Read the result

| Result | Meaning |
| --- | --- |
| A finding with test/query/tool evidence | A specific claim met Proven's evidence rules. Inspect its attached evidence. |
| No findings, complete review | No defect was proved by the checks that ran. This is not exhaustive coverage. |
| Incomplete review / exit 3 | Some required work is unsettled. Read failed services, unavailable data and remaining hypotheses. |
| CLI connection check failed | Refresh that runner's login; check model access and account quota. |

The comment is on your PR; JSON, SARIF and the redacted trace are under the workflow's **Artifacts**. [The evidence table](WHAT_PROVEN_CHECKS.md) explains what each check establishes.

To compare implementation with declared scope, add **`check-delivery: true`**. The PR title/description, each commit message and every collected discussion/review/inline/commit comment keep separate findings. The [illustrated scope guide](PR_DELIVERY.md) explains description checks, code indications, contradictions and uncertainty. Add **`refresh-data: true`** to acquire current provider data before review.

## Use another CLI

The commands and workflow stay the same. Change the backend name and image:

| Account/CLI | Backend | Image | Login |
| --- | --- | --- | --- |
| Claude Code | `claude` | `proven-claude:local` | Claude browser OAuth. |
| GitHub Copilot | `copilot` | `proven-copilot:local` | GitHub device login. |
| OpenAI Codex | `codex` | `proven-codex:local` | Codex device login to your ChatGPT account. |
| Cursor | `cursor` | `proven-cursor:local` | Cursor browser login. |
| Qwen Code | `qwen` | `proven-qwen:local` | Qwen's OAuth/device flow. |

For example:

```sh
sh scripts/proven-cli.sh build codex
sh scripts/proven-cli.sh login codex
sh scripts/proven-cli.sh check codex
```

Every backend gets its own `proven-BACKEND-home` volume. Run `login` again to refresh it; rebuilding an image does not refresh authentication. Choose a model available to your account with the Action's `model:` input. Defaults are `sonnet` for Claude, `auto` for Copilot and the CLI's default for the others.

## GitHub-hosted Claude

A fresh `ubuntu-latest` runner has no saved OAuth session. Claude provides a separate CI route: run **`claude setup-token`**, complete its browser flow and save the printed OAuth token in the reviewed repository's **Settings → Secrets and variables → Actions**, named **`CLAUDE_CODE_OAUTH_TOKEN`**. [Claude's instructions](https://code.claude.com/docs/en/authentication#generate-a-long-lived-token) explain that credential and its lifetime.

If Claude is installed only in your own image, run `docker run --rm -it --entrypoint claude proven-claude:local setup-token` instead. Complete that vendor flow yourself and keep its printed credential private.

In the workflow above, use `runs-on: ubuntu-latest` and replace the Proven step with:

```yaml
      - uses: saphan-ai/proven/cli@main
        with:
          backend: claude
          claude-oauth-token: ${{ secrets.CLAUDE_CODE_OAUTH_TOKEN }}
```

With `image` omitted, the Action builds the recipe in **your job**. Authentication is supplied at runtime; no credential enters an image layer. To reuse an image, set `image:` to your own registry reference and arrange registry login in the preceding step if needed.

This CI route is specific to Claude. The other backends in this guide use the saved login on your own runner. A desktop keychain or another machine's profile is not automatically portable.

## After the first run

The generated-test adapters cover **Go, Python/pytest, TypeScript/Vitest and JavaScript/node:test**. Automatic preparation handles one Go module, an npm project with `package-lock.json`, or Python with `requirements.txt`. For mixed projects, private dependencies, Poetry-only projects or extra system libraries, provide a prepared **`test-image:`**. That image contains your project's test tools; **`image:`** contains the model CLI.

Use `refresh-data: true` to acquire relevant public data automatically, or [import your own snapshots](SOURCES.md). The [refresh report](SOURCE_REFRESH.md) records check/observation dates and provider gaps.

Next: [two models](TWO_MODELS.md), [custom test runtime](RUNNER.md), [language coverage](LANGUAGES.md), or the [Action reference](ACTION.md).
