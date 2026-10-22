# AGENTS.md — for every agent working on this repository

You are working on **proven**, a public, open-source repository (Apache-2.0). Everything you commit here will be read by strangers, including the judges of the Open Agent Hackathon 2026. Read this file before you do anything else.

## 1. Two places, one change

This repository is a sub-repository of our workspace. The fleet keeps working exactly as it works on every other track.

1. **Write and register every document in the bridge, as usual.** Orders, status reports, reviews, test plans and decisions are written in our internal format and registered in the record. The record stays the source.
2. **Then put a public copy here**, in the public template:

   | document | goes to | template |
   |---|---|---|
   | order | `fleet/orders/ORDER_<slug>_r<n>.md` | [`fleet/orders/README.md`](fleet/orders/README.md) |
   | status report | `fleet/status/STATUS_<slug>_r<n>.md` | [`fleet/status/README.md`](fleet/status/README.md) |
   | review | `fleet/reviews/REVIEW_<slug>_r<n>.md` | [`fleet/reviews/README.md`](fleet/reviews/README.md) |
   | decision | `fleet/decisions/DECISION_<slug>_r<n>.md` | [`fleet/decisions/README.md`](fleet/decisions/README.md) |

3. **The public copy says the same thing as the original.** Same findings, same verdict, same numbers. Translate our internal format into the public template; do not soften, shorten or improve it.
4. **Every document carries its model and its cost**, to the cent, as recorded by Saphan Studio. A decision adds up the whole change and adds its row to [`fleet/LEDGER.md`](fleet/LEDGER.md): models, rounds, repair rounds, BLOCKING / OWED / NOTE, cost of building and of reviewing.
5. **The copy lands in the same change as the work it describes.** A change without its documents is not finished.

## 2. What never goes into this repository

Black it out (█, one block per character it covers) and list it in a `REDACTED` section at the end of the document, with what kind of thing it was and how many times. Never delete it silently, and never reword around it.

- credentials, tokens, keys, and the names of the variables that hold them;
- names of our machines, host names, IP addresses, user names;
- absolute paths and paths into our private record or workspace;
- file names and line numbers in Saphan Studio's private code;
- how our platform signs, verifies, pins or licenses anything, and how orders are bound to policies;
- the exact input or condition that gets past one of our own guards;
- table and column names of our private store.

You may keep: slugs, commit hashes of this repository, test names, law numbers, and anything that is already public on https://saphan.ai.

## 3. The rules of this repository

- **English only.** Short sentences, plain words.
- **All project code is written between 22 and 27 October 2026.** Nothing written before 22 October may be copied in.
- **Executors are Cursor agents.** They write the product code. Tests are written by a test author from the test plan, apart from the code. Every other role works as described in [`SAPHAN.md`](SAPHAN.md).
- **Record the model that really ran**, including the model Auto chose, and the share of the monthly pool before and after the run.
- **The owner writes no code.** He decides and merges.
- **Never push.** The owner pushes once a day at 22:00 Bangkok time (15:00 UTC), and once more before the submission on 27 October.
- **Never change `.github/`, `LICENSE` or this file.** Those are the owner's.
- **No document claims more than was run.** If you did not run it, it goes under "Not run" with the reason.
- **Do not write about Saphan Studio's internals.** Saphan is a pre-existing component. What may be said about it is in [`SAPHAN.md`](SAPHAN.md).

## Package layout

The reviewer sits behind `internal/review`. `review.Open` wires the packages below. Sample mode swaps in the sample adapters. The agents stay the same.

| package | role |
|---|---|
| `internal/generator` | one generator interface; the sample answers are recorded by agent and prompt |
| `internal/query` | query adapter; sample mode runs SQL on a local engine over `tables.sql` |
| `internal/testrun` | test-run adapter; sample mode returns recorded outcomes and executes no code; the live adapter runs the test in a container, or names the Prover off when no runtime is on the path |
| `internal/trace` | append-only JSONL trace at `logs/trace.jsonl` |
| `internal/candidate` | the finding record the agents pass along |
| `internal/agent/planner` | reads the change and names the risks |
| `internal/agent/investigator` | turns a risk into a query and runs it |
| `internal/agent/prover` | asks the test-run adapter for a test that fails on the change and passes on the base |
| `internal/agent/skeptic` | one refutation attempt per finding |
| `internal/agent/judge` | keeps a finding only when it has evidence |
| `internal/agent/writer` | fills the review result |
| `internal/pipeline` | runs the six agents in that order |
| `cmd/proven` | `proven scan --sample <fixture>` |

The sample change is `testdata/fixture/discount`.

## Configuration

The reviewer's file is `proven.yaml`. `review.Open` reads `Options.ConfigPath` when that path is set. An empty path reads `proven.yaml` in the review target when the file is present, and otherwise uses the defaults. An explicit path that does not exist is refused. The file is one flat YAML mapping of scalar values (comments and quotes are fine; a nested value is refused). The key set is closed.

| key | default | meaning |
|---|---|---|
| `timeout` | `0s` | A Go duration. When it is greater than zero, it bounds a review whose context has no deadline. Zero adds no deadline. The context's own deadline wins when it has one. |

An unknown key, or a `timeout` that is not a Go duration of zero or greater, is refused with `*ConfigError`. The error names the field and the allowed values.

## Generator

A live review calls an OpenAI-compatible endpoint. Sample mode does not, and it does not read these variables. `review.Open` with sample mode off loads them and refuses a bad value with `*ConfigError` before the review starts. The token is a secret: it is read from the environment only, and a refusal does not print it.

| variable | default | meaning |
|---|---|---|
| `METERLESS_RUNTIME_TOKEN` | none; required when sample mode is off | Bearer token |
| `PROVEN_GENERATOR_BASE_URL` | `http://127.0.0.1:8088` | Origin. A path is empty or `/v1`. The call is `POST {base}/v1/chat/completions` with `stream: true` |
| `PROVEN_GENERATOR_APP` | `proven` | `meterless-app` header. `meterless-scope` is `user` |
| `PROVEN_GENERATOR_MODEL` | `meterless/auto` | Model field on the request. The endpoint may route elsewhere. The review records the model the stream names |
| `PROVEN_GENERATOR_TIMEOUT` | `60s` | Bound on one attempt. An empty answer is tried once more, then named as a failure |
| `PROVEN_GENERATOR_REQUIRE_NVIDIA` | `all` | `all`, `none`, or a comma-separated list of `planner`, `investigator`, `prover`, `skeptic`, `judge`, `writer` |

The same headers are sent on `GET {base}/v1/models`. A review does not call that path. Each completion reads `meterless.routed_model` and `meterless.receipt_id` from the stream. The chunk's own `model` field is not the receipt. A routed model that is not `nvidia/*` is refused for an agent in the list above: the review names that model and does not use the answer as a verdict. A stream with no routed model is recorded as model unknown. A 401, a timeout, an unreachable endpoint, or an empty answer after one retry is a failed service named `nightshift`, and the review finishes degraded.

## Prover container defaults

The live Prover runs each test in a container. `DefaultContainerConfig` in `internal/testrun` returns these defaults, and `Validate` refuses a bad value on load. No setting turns the network on or passes a secret in.

| setting | default |
|---|---|
| runtime | `docker` |
| image | `golang:1.26` |
| time limit | 2 minutes for the base run, and 2 minutes for the change run |
| CPUs | `1` |
| memory | `512m` |

Each run is removed when it exits, has no network, has a CPU limit and a memory limit, and has one read-only mount of that one checkout. No environment variable is passed in. A run that passes its time limit is not proof. Removal after that has its own bound of 2 seconds, inside a margin of 3 seconds, so one side returns within its time limit plus 3 seconds even when the runtime hangs on every call. A failed removal is written into the outcome and names the container that may remain. The flags follow docker-run(1) from Docker 29.9.1. That manual says a missing image is pulled on the host before the container starts; it does not document a flag that stops the pull.
