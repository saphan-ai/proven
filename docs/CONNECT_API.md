# Connect your own API credentials

Use this path for a streaming OpenAI-compatible chat-completions endpoint. A subscription login for Claude Code or Copilot is a different connection: use [CLI models](CLI_MODELS.md). An API implementing another protocol needs a compatible gateway or adapter.

## 1. Select the recipient and model

In Bash, replace these two values with the API root and model from your provider:

```sh
export PROVEN_GENERATOR_MODE=openai
export PROVEN_GENERATOR_BASE_URL=https://api.example.com/v1
export PROVEN_GENERATOR_MODEL=your-model-id
```

For the NVIDIA endpoint used by this project's tests, the root is `https://integrate.api.nvidia.com/v1` and the measured model is `nvidia/nemotron-3.5-lightning-30b-a3b`. Other compatible providers can use the same adapter. Select options for that provider rather than copying a different model's thinking fields.

## 2. Enter the key without putting it in shell history

```bash
read -r -s -p 'Model API key: ' PROVEN_GENERATOR_KEY
printf '\n'
export PROVEN_GENERATOR_KEY
```

The prompt hides input; the shell history records the variable name, not the key. The key remains in this shell's environment and child process environment. Do not enable shell tracing or paste a literal key into a Dockerfile, URL or committed file. Local endpoints that require no authentication can leave the key unset.

## 3. Check one request, then review

```sh
export PROVEN_GENERATOR_TIMEOUT=300s
export PROVEN_GENERATOR_HEADER_TIMEOUT=40s
export PROVEN_GENERATOR_IDLE_TIMEOUT=90s
proven model check
proven scan --baseline-commit <base-sha> /path/to/repository
```

The first command probes one Investigator-shaped answer without retry. It does not validate data or the runner. Inspect the real report's completion, failed services and evidence after the scan. [Models](MODELS.md) explains tokens, reasoning and clocks; [Runner](RUNNER.md) and [Databases](DATABASES.md) cover the other components.

## 4. Store or replace the key on GitHub

With GitHub CLI signed in and permission to manage that repository's Actions secrets:

```bash
read -r -s -p 'Model API key: ' PROVEN_GENERATOR_KEY
printf '\n'
printf '%s' "$PROVEN_GENERATOR_KEY" | gh secret set PROVEN_MODEL_KEY --repo OWNER/REPOSITORY
unset PROVEN_GENERATOR_KEY
```

Running the same command replaces the existing secret value. It does not print or retrieve the old key. Re-run the workflow after replacement; an already running process retains its previous environment. The secret name must match the workflow:

```yaml
env:
  PROVEN_GENERATOR_MODE: openai
  PROVEN_GENERATOR_BASE_URL: ${{ vars.PROVEN_MODEL_URL }}
  PROVEN_GENERATOR_MODEL: ${{ vars.PROVEN_MODEL }}
  PROVEN_GENERATOR_KEY: ${{ secrets.PROVEN_MODEL_KEY }}
```

Set `PROVEN_MODEL_URL` and `PROVEN_MODEL` as repository variables. See the [Action guide](ACTION.md) for a complete job and fork handling. This repository's self-review instead uses `PROVEN_REVIEW_API_KEY`, with the historical `NVIDIA_API_KEY` fallback; use those names only when configuring that particular workflow.

## Two keys or local operation

The second API reviewer has its own endpoint, model and `PROVEN_SECOND_REVIEWER_KEY`; [Two models](TWO_MODELS.md) gives the complete recipe. For a local API under a restricted network, follow [Air-gapped operation](AIR_GAPPED.md). Never reuse the PR comment token as a model credential.
