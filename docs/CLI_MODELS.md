# Use your CLI account

For the short path covering Claude, Copilot, Codex, Cursor and Qwen Code, use **[Start with Proven](START.md)**. It handles customer images and login profiles through one Action. The sections below are the explicit transport reference.

Use your own installed, authenticated vendor CLI as a text model for Proven. You keep the account and subscription. Proven supplies prompts and validates answers; its restricted runner still performs the evidence execution. **The published Proven image contains no vendor CLI.** Install locally or follow the [customer image recipe](CLI_IMAGES.md).

These recipes use official CLIs through their normal sign-in, not extracted browser sessions or an unofficial subscription API. Account entitlement, organization policy, quota and vendor terms still apply. Proven does not provide a shared subscription gateway.

## Choose a model connection

Codex, Cursor and Qwen Code use the same [build/login/check helper](START.md#use-another-cli). The pinned customer recipes and text wrappers are in [examples/cli](../examples/cli/README.md). Codex uses a dedicated `CODEX_HOME`, ignores user configuration/rules, disables shell/code execution features and reads only `--output-last-message`; its JSONL envelope is discarded. Cursor runs in ask mode in a fresh directory with explicit shell/read/write denials. Qwen disables customizations and sets a zero tool-call budget. Each receives only its own profile and prompt, with no target checkout or Docker socket.

These controls depend on the pinned CLI version. Account-free wrapper checks verify argv/stdin/profile boundaries; image builds check the actual executable versions. Authenticate and validate your own model access before relying on a review.

| What you have | Connection |
| --- | --- |
| A compatible API and key | [API setup](CONNECT_API.md). |
| Claude Code subscription/account | Customer-installed `claude`, text wrapper and a dedicated login profile below. |
| GitHub Copilot access | Customer-installed `copilot`, text wrapper and a dedicated login profile below. |
| Both | [Two reviewers](TWO_MODELS.md), with independent command configurations. |

Copilot is the CLI product, not the old `gh copilot` extension or the PR comment token. GitHub documents CLI availability and organization controls in its [authentication guide](https://docs.github.com/en/copilot/how-tos/copilot-cli/set-up-copilot-cli/authenticate-copilot-cli).

## Claude Code locally

Install the official CLI using [Anthropic's setup instructions](https://code.claude.com/docs/en/setup). Record `claude --version`. The included wrapper flags were checked against local Claude Code 2.1.319; the customer build recipe pins an independently overridable npm version.

From the Proven checkout, remember the wrapper path, then authenticate a dedicated profile:

```sh
PROVEN_CLAUDE_WRAPPER="$PWD/examples/cli/claude-text.sh"
mkdir -p "$HOME/.config/proven/claude"
chmod 700 "$HOME/.config/proven/claude"
CLAUDE_CONFIG_DIR="$HOME/.config/proven/claude" claude auth login
CLAUDE_CONFIG_DIR="$HOME/.config/proven/claude" claude auth status --text
```

Complete the vendor's sign-in yourself, choosing the account/billing method you intend. This separate profile avoids loading a coding project's normal customizations. See [Claude authentication](https://code.claude.com/docs/en/authentication).

Probe the wrapper from a shell, with a model alias or ID your account supports:

```sh
printf '%s\n' 'Reply only with {"queries":[]} as JSON, without markdown.' \
  | sh "$PROVEN_CLAUDE_WRAPPER" sonnet
```

It should print the requested JSON alone. Configure Proven with an absolute wrapper path:

```sh
export PROVEN_GENERATOR_MODE=command
export PROVEN_GENERATOR_MODEL=sonnet
export PROVEN_GENERATOR_COMMAND='["sh","/absolute/path/to/proven/examples/cli/claude-text.sh","{model}"]'
export PROVEN_GENERATOR_TIMEOUT=300s
proven scan --baseline-commit <base-sha> /path/to/repository
```

The wrapper uses print mode and **text** output, disables tools and normal customizations, supplies no MCP servers, and avoids session persistence. JSON/stream-JSON output would wrap the agent answer in a vendor envelope. See the [official flag reference](https://code.claude.com/docs/en/cli-reference). This uses the CLI's own authenticated profile; Proven does not read or export its credentials.

## GitHub Copilot locally

Install the official [Copilot CLI](https://docs.github.com/en/copilot/how-tos/copilot-cli/set-up-copilot-cli/install-copilot-cli), then record `copilot --version`. The recipe and local help check use 1.0.112. Set up its dedicated profile:

```sh
PROVEN_COPILOT_WRAPPER="$PWD/examples/cli/copilot-text.sh"
mkdir -p "$HOME/.config/proven/copilot"
chmod 700 "$HOME/.config/proven/copilot"
COPILOT_HOME="$HOME/.config/proven/copilot" copilot login
COPILOT_HOME="$HOME/.config/proven/copilot" copilot config disableAllHooks true
```

Complete GitHub's browser/device login with the account that has Copilot access. A container can use `copilot login --device-code`. Credentials can use the OS keychain; on headless Linux, GitHub documents a file-storage fallback. Logging in inside the customer container avoids assuming a host keychain can be copied into Linux. See [authentication](https://docs.github.com/en/copilot/how-tos/copilot-cli/set-up-copilot-cli/authenticate-copilot-cli) and [profile settings](https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-config-dir-reference).

Choose a model available to your account; `auto` lets Copilot select one. A model's vendor and the transport are different: a Claude model selected *inside Copilot* is still a Copilot connection.

```sh
printf '%s\n' 'Reply only with {"queries":[]} as JSON, without markdown.' \
  | sh "$PROVEN_COPILOT_WRAPPER" auto

export PROVEN_GENERATOR_MODE=command
export PROVEN_GENERATOR_MODEL=auto
export PROVEN_GENERATOR_COMMAND='["sh","/absolute/path/to/proven/examples/cli/copilot-text.sh","{model}"]'
export PROVEN_GENERATOR_TIMEOUT=300s
proven scan --baseline-commit <base-sha> /path/to/repository
```

The adapter requests silent final text, no custom repository instructions, no built-in MCPs and an empty available-tool set; explicit denials add another tool boundary. The fresh working directory contains no reviewed checkout. It does not use allow-all or an autonomous coding mode. The [programmatic reference](https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-programmatic-reference) describes output and prompt piping.

## Credentials and CI

The command transport inherits ordinary process and login locations, not arbitrary API/database/comment tokens. The wrappers derive their private profiles from `HOME`. Log in those profiles once or mount a previously provisioned private profile at runtime. Do not put the login into the image build or commit profiles to Git.

GitHub also supports specific Copilot token types for automation; classic PATs are unsupported, and a workflow's ordinary posting token is not a substitute for Copilot entitlement. Those credentials need explicit customer provisioning. The included recipe uses vendor login/profile storage rather than broadening Proven's child environment.

## Limits and verification

`proven model check` is an HTTP API probe, so test these wrappers directly as above. CLI mode uses Proven's total timeout and shared answer-recovery policy; HTTP header/idle/token/body fields do not configure a vendor executable. Tune the CLI's own model flags in your customer wrapper if needed.

Proven records the requested CLI model; the answering model is **unknown**, because the generic text protocol has no verified response identity. It does not invent API timings or HTTP attempts for a CLI. Two subscriptions can also route to the same model family; see [Two models](TWO_MODELS.md).

Adapter tests use fake executables and cover stdin, argv, cleanup and tool flags. Flag/help checks and image version checks do not establish successful subscription authentication or model quality. Validate your own login with a short prompt, then known defect/control reviews. See [Evaluation](EVALUATION.md).
