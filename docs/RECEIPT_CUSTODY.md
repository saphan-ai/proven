# Receipt custody

A model answer counts only when it is in custody. Three things must hold:

1. A Nightshift receipt exists whose `request_id` is the answer's response id, and its Ed25519 signature
   verifies under a public key you pinned.
2. The model that answered is the model that was requested.
3. The answer's trace line names its response id, so the join can be made at all.

Each live model call writes one `model` event to `logs/trace.jsonl`. The event carries the requested model, the routed
model, the receipt id and the response id. Older traces lack the requested model and the response id. An older event
that carries a receipt id is a live answer, and the audit fails it by name (`no response id recorded`); it does not
guess. An older event that carries neither a requested model nor a receipt id cannot be told from a sample call, so the
audit skips it and counts it under "trace events skipped". Such an answer is not judged either way.

## What a receipt looks like

Nightshift writes one JSON line per request:

```
{"body":{...,"receipt":{"request_id":"chatcmpl-...","status":"ok","winning_route":{"provider":"nvidia","model":"nvidia/..."}},"ts_ms":...},"pubkey_b64":"...","sig_b64":"..."}
```

The signature is over `body` written as sorted-key compact JSON. The key inside each line is not trusted: you pin
the key out of band, and a line signed under another key is refused as an unknown key.

## Run the audit

```
go run ./cmd/receiptaudit -receipts receipts.jsonl -pubkey <base64 public key> logs/trace.jsonl [more traces...]
```

`-pubkey-file` reads the key from a file instead. The command opens only the files you give it. It dials nothing.

Per UTC day it prints:

| line | meaning |
|---|---|
| receipts | lines of the receipts file written that day |
| signatures verified | how many of them verify under the pinned key |
| answers | live model answers in the traces that day |
| answers matched to a receipt by request id | the answer's response id equals the request id of a receipt that verifies |
| answers whose routed model equals the requested model | out of the answers that asked for a specific model |
| auto-routed | answers that asked for `meterless/auto`; the router chooses, so there is nothing to compare, and they are counted apart |
| answers in custody | answers that pass every check |

Then it prints every answer that fails, with its reason, and every receipts line it refused.

| reason | what happened |
|---|---|
| `no receipt matches response id ...` | no receipt in the file names that response |
| `receipt refused: signature does not verify` | the receipt body was changed after it was signed |
| `receipt refused: unknown key` | the receipt is signed under a key other than the pinned one |
| `receipt refused: malformed receipt` | the line is not a receipt |
| `receipt status is "failed", not "ok"` | the receipt verifies, but it records a request that did not succeed |
| `response id already matched at <source>` | one receipt vouches for one answer; this answer repeats a response id already matched |
| `model mismatch: requested A, answered by B` | the answering model is not the one asked for |
| `model mismatch: the stream names A, the signed receipt names B` | the trace and the signed receipt disagree |
| `no receipt: the answer came from an API that signs nothing` | a `nvidia` mode answer; the hosted API gives no receipt |
| `no response id recorded ...` | the trace predates the response id |

Exit codes: `0` every answer is in custody, `1` at least one answer fails, `2` the audit could not run.

## Limits

- Nightshift loses its receipts when it restarts. An answer whose receipts file was lost fails with no matching receipt.
- The audit proves what the receipts file and the trace say. It does not prove the receipts file is complete.
- Sample-mode trace events name no request and no receipt. They are counted as skipped, not failed.
