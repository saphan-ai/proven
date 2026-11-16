# CLI Action reference

Start with [the short guide](START.md). Use `saphan-ai/proven/cli@<reviewed SHA>` for your retained workflow. The preset runs trusted reviewer tooling on the runner and the model CLI in your customer image. Generated project tests use separate restricted containers.

## Inputs

| Input | Default | Meaning |
| --- | --- | --- |
| `backend` | Required | `claude`, `copilot`, `codex`, `cursor` or `qwen`. |
| `image` | Build in your job | Your local or registry CLI image. Registry access is configured before this step. |
| `home-volume` | `proven-BACKEND-home` | The named Docker volume authenticated with that CLI. It must already exist. |
| `model` | CLI selection | `sonnet` for Claude, `auto` for Copilot, otherwise the CLI default. The report does not invent a confirmed answering-model identity. |
| `claude-oauth-token` | Empty | Explicit Claude CI OAuth credential from `claude setup-token`. Use this or a login volume. |
| `second-backend` | Disabled | Optional second CLI reviewer. |
| `second-image` | Build in your job | Customer image for that second CLI. |
| `second-home-volume` | `proven-SECOND_BACKEND-home` | Its own authenticated profile. |
| `second-model` | Second CLI selection | Account-accessible model on the second transport. |
| `test-image` | Automatic preparation | Prepared offline project test runtime. The automatic path supports one Go module, npm lockfile project, or `requirements.txt` Python project. |
| `check-delivery` | `false` | Compare stated PR/commit scope with checked code citations. See [delivery assessment](PR_DELIVERY.md). |
| `target` | `.` | Target checkout/module path relative to the workspace. |
| `baseline-commit` | PR base SHA | Full commit SHA present in that clone. |
| `timeout` | `1500` | Whole review seconds. Preflight preparation is additionally bounded by the job limit. |
| `call-timeout` | `300` | CLI completion seconds; an internal container timeout ends five seconds earlier, with a two-second kill margin. |
| `strict` | `true` | Findings fail the check. |
| `require-complete` | `true` | Incomplete required analysis fails with exit 3 independently of findings. |
| `comment` | `true` | Post/update one PR comment; needs `pull-requests: write`. |
| `github-token` | Job token | Comment credential, removed before review/model invocation. |

Outputs are `json-file`, `sarif-file` and `findings-count`. JSON, SARIF and the redacted trace are uploaded even when a started review fails. Preparation/connection failures may have no report; their step log identifies the failed phase.

## Runtime and credential boundaries

Linux and macOS runners need Docker, Git and Python 3. The preset prepares the Go toolchain and builds the controller from the pinned Action source. It does not execute a target build script on the host. Dependency preparation reads trusted base manifests and disables npm lifecycle scripts. Private dependencies, additional system libraries, Poetry-only projects and mixed runtimes use an explicitly prepared `test-image`.

Each model container receives its own login volume or the primary Claude CI credential through a private runtime file. It receives no target checkout or host Docker socket. Containers have bounded time, memory, CPU and process counts and run with dropped capabilities and a read-only root. The token-file route keeps the credential out of Docker's recorded environment and uses a non-root identity.

The controller retains the CI token value for redaction, while the command adapter's environment allowlist excludes it. The second CLI gets only its own profile. A write token for PR comments does not authenticate Copilot.

Every connection gets one short preflight completion, bounded at 60 seconds. These requests use your account quota. A failed preflight stops before review and directs you to refresh that runner's login, model access or quota. Passing it establishes a small protocol check, not review accuracy.

The preset refuses `PROVEN_NETWORK=local` before dependency/toolchain preparation, because an arbitrary vendor CLI cannot enforce Proven's HTTP destination policy. Running on your own machine is different from air-gapped network mode.

## Persistent and interrupted runs

Images and login volumes remain yours. Temporary review state and CI credential files are removed; login volumes are preserved for the next run. Containers carry a unique `io.proven.cli-run` label and cleanup targets only that invocation. A cleanup failure reports container identities.

Killing a Docker client does not necessarily stop its container. The internal deadline remains even if a job is interrupted before cleanup; a crashed/offline runner can still leave temporary files or labelled containers. Inspect that runner before reusing it. The full runtime trust and mount-path requirements are in [Runner](RUNNER.md); a containerized self-hosted runner needs shared paths that resolve identically on its Docker daemon host.

After upgrading a vendor CLI or runtime, rebuild, scan and run `check` again. The [measured supply-chain audit](CLI_SUPPLY_CHAIN_AUDIT.md) shows why checking the completed customer image matters. Keep accounts and runtime profiles outside source control and image layers.
