# Optional safety guardrails around model calls

The reviewer model and the safety model are independent connections. You can use your own API or subscription CLI for review and enable a configured content-safety service around its prompts and answers.

```text
prompt → safety check → reviewer model → safety check → agent uses the answer
```

The first check happens before Proven sends the prompt to the reviewer. The second checks the prompt/answer pair before the answer reaches the agent. An unsafe verdict, malformed verdict, timeout or unavailable safety service refuses that model step. Proven records the failure and preserves other available evidence. A refusal is never converted to a successful safety check.

## NVIDIA-hosted service

Choose a content-safety model available through your account. Set the model ID explicitly rather than relying on an unverified default:

```yaml
env:
  PROVEN_GUARDRAIL_MODE: nvidia
  PROVEN_GUARDRAIL_BASE_URL: https://integrate.api.nvidia.com/v1
  PROVEN_GUARDRAIL_MODEL: ${{ vars.PROVEN_SAFETY_MODEL }}
  PROVEN_GUARDRAIL_KEY: ${{ secrets.PROVEN_SAFETY_KEY }}
  PROVEN_GUARDRAIL_TIMEOUT: 120s
```

The safety credential is separate from your reviewer credential. Store it as a runtime secret. The safety request contains the prompt/answer content; enabling a hosted service sends that content to its operator.

NVIDIA's [NemoGuard content-safety NIM](https://docs.nvidia.com/nim/llama-3-1-nemoguard-8b-contentsafety/latest/index.html) and [NeMo Guardrails content-safety rails](https://docs.nvidia.com/nemo/guardrails/latest/configure-guardrails/guardrail-catalog/content-safety) are related components. Proven supplies a **direct content-safety API adapter**: its shipped interface is not a NeMo Colang/orchestration configuration loader. Use an endpoint compatible with the adapter or add a guardrail adapter for a different service contract.

The hosted default model ID in the configuration reference has not been confirmed by a live call. Account access, deployed model identity and the exact response contract must be checked for your service. Ordinary tests use fixtures; they do not establish a connection to your account.

## A self-hosted endpoint

```sh
export PROVEN_GUARDRAIL_MODE=server
export PROVEN_GUARDRAIL_BASE_URL=http://127.0.0.1:8081
export PROVEN_GUARDRAIL_MODEL=your-deployed-safety-model
```

Set `PROVEN_GUARDRAIL_KEY` if your server requires it. Use HTTPS for remote authenticated endpoints. The adapter calls the configured OpenAI-style chat-completion endpoint and requires the supported structured safety verdict; ordinary reviewer text is not a safety verdict. See `internal/guardrail/parse.go` and the contract tests when implementing a compatible server.

`PROVEN_NETWORK=local` refuses NVIDIA's public hosted service; permitted internal endpoints follow the same destination policy as other connections. A safety service does not create an exception to that policy.

## Reports and security boundaries

Safety checks are **off by default**. Disable them explicitly with `PROVEN_GUARDRAIL_MODE=off`. A deterministic run needs no safety model. Every output identifies whether safety checks ran, were off or had no calls to check, and records checked/refused/unanswered counts and the configured service identity. Recorded demos label these checks as replays, not live service calls.

A classifier verdict does not prove code correctness, remove every prompt-injection risk or grant permission to execute model-written commands. Restricted test containers, fixed command construction, registered query templates, resource bounds, redaction and cleanup remain separate controls. [Architecture](ARCHITECTURE.md), [restricted runner](RUNNER.md), [configuration](CONFIGURATION.md#optional-model-safety-service) and [security reporting](../SECURITY.md) describe these boundaries.
