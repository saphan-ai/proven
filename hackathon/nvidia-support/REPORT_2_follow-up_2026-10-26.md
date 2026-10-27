# Report 2 to NVIDIA support (follow-up): not our prompts, not our key; Nemotron fails in windows

Filed 26 October 2026 by team Saphan.ai, following [report 1](REPORT_1_intermittent-500_2026-10-26.md). Published here with
the account and key id removed. **As of 27 October 2026 this report has not been resolved.**

## Prompt size is not the cause

On 26 October 2026, 00:58 to 01:02 UTC, one fixed prompt at 8, 16, 32, 48, 64, 96, 128, 192 and 256 KB: 37 calls, all
HTTP 200, in 2 to 26 seconds. Our first guess of a size limit was wrong.

## A new API key does not fix it

At about 14:05 UTC we created a new key and ran a small test (prompt "Reply with the single word OK", `max_tokens` 8, calls 3
seconds apart, one at a time):

| call | HTTP | time | nvcf-reqid |
|---|---|---|---|
| tiny 1 | 200 | 5.16 s | `4d9d570f-a95f-49a9-bf04-ba16afc05708` |
| tiny 2 | 200 | 0.98 s | `b05807d9-50ec-45f2-8ede-169af6a4c847` |
| tiny 3 | 200 | 0.60 s | `4fce764b-5f87-4938-8377-bb96dd2014e3` |
| tiny 4 | 200 | 0.58 s | `b848ff15-61e0-464a-8a21-fcdca1c12bc1` |
| tiny 5 | 200 | 0.58 s | `41ac4eb7-eb18-40cf-8dac-cf3db0f8f7ab` |
| ~30 KB prompt | 200 | 2.67 s | `b06a63a8-beee-4e23-8f16-aca828cecc78` |
| ~30 KB prompt | 200 | 2.48 s | `15b85674-89f7-45fc-8873-b20355f47275` |

Minutes later, the same test on a newer key (the key had been replaced again in between):

| call | HTTP | time | nvcf-reqid |
|---|---|---|---|
| tiny 1 | 200 | 0.75 s | `bfcf3f8e-acf4-4395-8285-a3245849c34c` |
| tiny 2 | 200 | 0.58 s | `48ca38dc-d81b-4f2f-83ca-fd7e3e40467e` |
| tiny 3 | 200 | 1.31 s | `40c5f3e5-a5b0-40ec-bc82-f30af8293732` |
| tiny 4 | — | **no response at all** | — |

A third run: `tiny 1` 200 in 0.81 s (`8f097cc2-1c30-4245-b0c1-1fe7ff7025d9`), then the next call hung with no response.

## The pattern we see

- The service works in quiet windows and fails in others, on tiny and large prompts alike, on every key we used.
- **Latency rises just before a failure**: 0.58 s, then 1.31 s, then no answer at all.
- Failure windows so far (UTC): 25 Oct 15:04 to 15:26; 26 Oct about 00:07 to 00:15, 00:37, 00:41 to 00:50, and around 14:10.
- `GET /v1/models` answers 200 the whole time, so authentication and routing to the API are fine.

## What we asked

1. Is there capacity trouble on this model right now, and is there an ETA?
2. Can you check the request ids above, especially the ones just before the hangs?
3. Is there a more stable Nemotron model id or endpoint we should use until judging ends?
4. When the service is overloaded, could it answer with HTTP 429 or 503 instead of hanging?
