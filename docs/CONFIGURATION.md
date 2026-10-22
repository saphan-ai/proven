# Configuration

Every setting the reviewer reads, its default, and what it changes (ADR-0013: nothing is hardcoded; secrets come from the
environment only). The names and defaults below were read from the code on `main` on 22 October 2026
(`internal/generator/settings.go`, `internal/guardrail/settings.go`, `internal/query/settings.go`,
`internal/review/config.go`, `cmd/livecheck`). A wrong value stops Proven before any work, and the refusal names the
setting, the value and what is allowed.

The shell's own flags (`proven scan …`) are documented by the shell's track. Sample mode (`--sample`) reads none of the
settings below except `proven.yaml`'s `timeout`: see [SAMPLE_MODE.md](SAMPLE_MODE.md).

<!-- generated: settings -->
## The model generator: Meterless Nightshift

| Environment variable | Default | What it changes |
|---|---|---|
| `METERLESS_RUNTIME_TOKEN` | none (required live) | the runtime token sent to Nightshift. Secret: never printed |
| `PROVEN_GENERATOR_BASE_URL` | `http://127.0.0.1:8088` | Nightshift's address |
| `PROVEN_GENERATOR_APP` | `proven` | the application name sent with each call |
| `PROVEN_GENERATOR_MODEL` | `meterless/auto` | the model asked for; Nightshift's router chooses when `auto` |
| `PROVEN_GENERATOR_TIMEOUT` | `60s` | the limit for one model call |
| `PROVEN_GENERATOR_REQUIRE_NVIDIA` | `all` | which agents must be answered by an `nvidia/*` model (by the stream's receipt); `all`, `none`, or a list of agents. A non-NVIDIA answer to such an agent is refused by name |
| `PROVEN_SKEPTIC_MODEL` | `nvidia/nemotron-3.5-super-vl-120b-a12b` | the Skeptic's model (ADR-0006: another family than the finder's) |

## The guardrail: NVIDIA Nemotron content safety

| Environment variable | Default | What it changes |
|---|---|---|
| `PROVEN_GUARDRAIL_MODE` | `server` | `server` (your own model server), `nvidia` (NVIDIA's hosted API) or `off` (the trace says "no guardrail") |
| `PROVEN_GUARDRAIL_BASE_URL` | `http://127.0.0.1:8081` for `server`; `https://integrate.api.nvidia.com/v1` for `nvidia` | the guardrail's address |
| `PROVEN_GUARDRAIL_MODEL` | `nemotron-3.5-content-safety` for `server`; `nvidia/nemotron-3.5-content-safety` for `nvidia` | the safety model. The hosted id is not yet confirmed by a live call |
| `PROVEN_GUARDRAIL_KEY` | none | the guardrail's key. Secret: never printed |
| `PROVEN_GUARDRAIL_TIMEOUT` | `5s` | the limit for one check |

Every prompt and every answer is checked; an `unsafe` verdict, or a guardrail that does not answer, stops that step by name.

## Evidence: Zetaris over REST

| Environment variable | Default | What it changes |
|---|---|---|
| `ZETARIS_REST_URL` | none (required live) | Zetaris's REST address; an absolute `http` or `https` URL with no user info, query or fragment |
| `ZETARIS_API_KEY` | none (required live) | Proven's Zetaris key. Secret: never printed |
| `ZETARIS_ORG_ID` | none (required live) | the organisation id sent with each query |
| `ZETARIS_ROW_LIMIT` | `100` | the most rows one query returns |
| `ZETARIS_TIMEOUT` | `60s` | the limit for one query |
| `PROVEN_EVIDENCE_VIEW` | `px_evidence.px_kev_vulns` | the view `livecheck` queries to prove Zetaris answers (livecheck only) |

## The Prover's container

| Environment variable | Default | What it changes |
|---|---|---|
| `PROVEN_TEST_RUNTIME` | `docker` | the container command line, by name or absolute path |
| `PROVEN_TEST_IMAGE` | `golang:1.26` | the image each test runs in |
| `PROVEN_TEST_CPUS` | `1` | the CPU limit of one test run |
| `PROVEN_TEST_MEMORY` | `512m` | the memory limit of one test run; a positive size with suffix b, k, m or g |
| `PROVEN_TEST_TIMEOUT` | `120s` | the limit for one test run, base or change |
| `PROVEN_TEST_MAX_WRITTEN_BYTES` | `65536` | the largest test file the Prover may write |

Each test runs with no network and no secret (ADR-0025). With no container runtime the Prover is off, and the review says so. Sample mode reads none of these.
<!-- end generated -->

## `proven.yaml`

A flat mapping in the reviewed repository. On `main` it has one key:

| Key | Default | What it changes |
|---|---|---|
| `timeout` | `0s` (no deadline) | the review's deadline when the shell set none; at the deadline the review returns what the Judge has kept. A configuration file the shell names explicitly is read instead of this one |

The reviewer model, the model-family table and the review classes (ADR-0008 to ADR-0014) add keys here when that work
lands; this page is updated with them.

## Check the live settings

```bash
go run ./cmd/livecheck
```

With nothing set, it reports every service `not ready` with the reason and exits 1. Its output never contains a secret.
