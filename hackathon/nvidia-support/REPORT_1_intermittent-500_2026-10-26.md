# Report 1 to NVIDIA support: intermittent HTTP 500 and silent hangs on Nemotron

Filed 26 October 2026 by team Saphan.ai, Open Agent Hackathon 2026, for Proven. Published here with the account, the key id
and network details removed. The request ids (`nvcf-reqid`) are NVIDIA's own ids for our calls and are kept, so the
failures can be traced. **As of 27 October 2026 this report has not been resolved.**

## What fails

- **Endpoint:** `POST https://integrate.api.nvidia.com/v1/chat/completions`
- **Model requested:** `nvidia/nemotron-3.5-super-vl-120b-a12b`. The responses name the model
  `private/nvidia/nemotron-3.5-super-120b-a12b`.
- **Two symptoms:**
  1. **HTTP 500** with this body: `{"message":"Failed to generate completions","type":"Internal Server Error","code":500}`
  2. **No response at all** for long requests: the connection stays open and nothing comes back until our client gives up.
- **What still works:** `GET /v1/models` answers 200 every time, and some chat calls succeed between the failures. We never
  received HTTP 401, 403 or 429, so this is not authentication and not rate limiting.
- **Earlier, separate:** on 25 October around 16:20 UTC the model we used before, `nvidia/llama-3.3-nemotron-super-49b-v1.5`,
  answered HTTP 410 (retired). We moved to the model above.

## Measured failures, with NVIDIA's request ids

### A minimal test, 26 October 2026, 00:37 UTC

Five identical calls, 3 to 4 seconds apart, one at a time, prompt `Reply with the single word OK`, `max_tokens` 8:

| time (UTC) | HTTP | nvcf-reqid | body |
|---|---|---|---|
| 2026-10-26 00:37:12 | 500 | `0f319903-b68e-4ca8-bbfa-2b8ff55c58a8` | `Failed to generate completions` |
| 2026-10-26 00:37:16 | 500 | `9ea1d0c7-f98c-4c7f-bbfb-ee7b1a7a69e1` | same |
| 2026-10-26 00:37:19 | 500 | `799cc4ae-acc0-40b1-aaff-423aef589e43` | same |
| 2026-10-26 00:37:23 | 200 | `c9bf5c21-7ed1-41c7-ad6b-4090eec9ed2a` | `OK` |
| 2026-10-26 00:37:27 | 200 | `3036eb64-5eb5-4f52-bde6-3da71ef51729` | `OK` |

Three of five identical, tiny, sequential calls failed within 15 seconds. Nothing in the request changed between a failing
and a succeeding call.

### A batch run, 26 October 2026, 00:07 to 00:15 UTC

A recording of our test corpus: about 40 review calls in a row. Every one failed, with HTTP 500 or a client timeout.

### Long requests with no answer, 25 October 2026, 15:04 to 15:26 UTC

One code review sent a request of 412,733 bytes. It was sent 7 times, about every 3.5 minutes. **None of the 7 got any
response**: no status code, no body, no headers. Before 14:57 UTC the same day every call succeeded, and other calls
succeeded at 15:22 and 16:01 UTC.

## What we asked

1. Is there a known incident on this model, and is there an ETA?
2. Can you look up the failing request ids above and tell us what failed on your side?
3. Is there a size or time limit for a single request that we should stay under?
4. Is there a stable alternative endpoint or model id we should use until it is fixed?
