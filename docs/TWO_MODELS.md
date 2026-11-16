# Configure two reviewers

For two CLI accounts on your runner, build and log each one in with the [start guide](START.md), then add the second backend to the same Action:

```yaml
- uses: saphan-ai/proven/cli@main
  with:
    backend: claude
    image: proven-claude:local
    second-backend: copilot
    second-image: proven-copilot:local
```

Each uses its own login volume. `second-model` selects a model available to the second account; `second-home-volume` selects another dedicated profile. Both connections are checked before review. The sections below describe explicit adapter configuration for custom deployments.

The primary and second reviewer can each use an API or a customer-installed CLI. They receive independent review prompts; their identities, evidence and disagreements remain visible. Two model opinions without evidence do not become a proved finding.

The Skeptic model on one endpoint is a separate setting. Selecting `PROVEN_SKEPTIC_MODEL` alone does not create a second independently configured reviewer.

## Two APIs

Configure the primary as in [API setup](CONNECT_API.md), then set:

```sh
export PROVEN_SECOND_REVIEWER_MODE=openai
export PROVEN_SECOND_REVIEWER_BASE_URL=https://second-api.example.com/v1
export PROVEN_SECOND_REVIEWER_MODEL=second-model-id
read -r -s -p 'Second model key: ' PROVEN_SECOND_REVIEWER_KEY
printf '\n'
export PROVEN_SECOND_REVIEWER_KEY
export PROVEN_SECOND_REVIEWER_EXTRA_BODY='{}'
export PROVEN_SECOND_REVIEWER_AGENT_EXTRA_BODY='{}'
```

Use Bash for the hidden prompt. Local unauthenticated APIs can omit the key. Clearing inherited provider-body maps matters when the endpoints accept different options. Primary call controls otherwise carry over; `PROVEN_SECOND_REVIEWER_` overrides customize them.

## API plus a subscription CLI

Keep the primary API configuration and authenticate the [Copilot wrapper](CLI_MODELS.md#github-copilot-locally), then:

```sh
export PROVEN_SECOND_REVIEWER_MODE=command
export PROVEN_SECOND_REVIEWER_MODEL=auto
export PROVEN_SECOND_REVIEWER_COMMAND='["sh","/absolute/path/to/proven/examples/cli/copilot-text.sh","{model}"]'
export PROVEN_SECOND_REVIEWER_TIMEOUT=300s
```

The second command has its own argv and model; it cannot silently borrow the first command/model. API address/key settings are ignored in command mode. HTTP-only controls do not change CLI flags.

## Claude Code plus GitHub Copilot

Install and authenticate both dedicated profiles using [CLI models](CLI_MODELS.md). Use absolute paths on a host, or the installed wrapper paths in your [customer-built dual image](CLI_IMAGES.md):

```sh
export PROVEN_GENERATOR_MODE=command
export PROVEN_GENERATOR_MODEL=sonnet
export PROVEN_GENERATOR_COMMAND='["sh","/absolute/path/to/proven/examples/cli/claude-text.sh","{model}"]'
export PROVEN_GENERATOR_TIMEOUT=300s

export PROVEN_SECOND_REVIEWER_MODE=command
export PROVEN_SECOND_REVIEWER_MODEL=auto
export PROVEN_SECOND_REVIEWER_COMMAND='["sh","/absolute/path/to/proven/examples/cli/copilot-text.sh","{model}"]'
export PROVEN_SECOND_REVIEWER_TIMEOUT=300s

PROVEN_REQUIRE_COMPLETE=true proven scan \
  --baseline-commit <base-sha> --timeout 1800 \
  --json-output review.json /path/to/repository
```

Choose concrete model IDs instead of aliases/`auto` when your providers allow it and reproducibility matters. `sonnet` is interpreted by Claude Code, `auto` by Copilot; neither is a Proven model catalogue. Review time and quota must accommodate both paths.

## Which changes get both reviews?

Connecting a second reviewer widens the default doubled classes, while an explicit repository `doubled_classes` policy can narrow them. Security obligations remain. A validated prose-only plan can select no model at all, so connecting two models does not imply two paid calls for every documentation change. See [Architecture](ARCHITECTURE.md#review-planning-and-security-floors), [Policy](POLICY.md) and the exact keys in [Configuration](CONFIGURATION.md#provenyaml).

Inspect reviewer statements and trace roles (`default` / `special`) to confirm both actually ran. CLI answering identities stay unknown; configured names are not proof of independence. A failed second service is a missing answer, not a disagreement. Its failure is attributed to `second-reviewer` and required completion can fail with exit 3.

## Disable or change the second connection

`PROVEN_SECOND_REVIEWER_MODE=off` explicitly disconnects it, including when old connection variables remain in the environment. Without a mode, a configured second command selects command mode; otherwise the established API key or endpoint-plus-model activation remains compatible. Explicit `openai` requires its endpoint and model.

For independent source data and test execution, the same [database](DATABASES.md) and [runner](RUNNER.md) boundaries apply. A second model does not supply an absent vulnerability feed or fix an unprepared toolchain.
