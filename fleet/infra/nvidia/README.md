# NVIDIA Nemotron through Nightshift

Proven never calls NVIDIA directly. Every Nemotron request goes to Meterless Nightshift, which runs in the macOS VM
([`../nightshift-vm/`](../nightshift-vm/)), and Nightshift calls NVIDIA's hosted API. The NVIDIA key is entered in
Nightshift's Providers tab inside the VM and never leaves it: no key is in Proven's code, in Cursor, or on the build
machines. Environment, not product code.

## How a request is sent

- Proven speaks the OpenAI chat protocol to Nightshift's local endpoint.
- Each request carries two headers that name who sent it: `meterless-app` (`proven` for the product,
  `proven-review` for the fleet's review legs) and `meterless-scope: user`.
- The model cannot be pinned from outside (see the nightshift-vm runbook): Nightshift records the model a request
  names but routes by its own mode. The answering model is read from the signed receipt, never from the request.

## How it is checked: the signed receipts

Nightshift writes one signed receipt per request to its `receipts.jsonl` inside the VM. Each line is
`{body: {receipt, manifest_hash, llama_cpp_ref, signer, ts_ms}, pubkey_b64, sig_b64}`. The signature is Ed25519 over
`body` serialised as compact JSON with sorted keys.

What we measured on 22 October 2026, over every line in the file (206 receipts, 21 Oct 13:24Z to 22 Oct 07:37Z):

| check | result |
|---|---|
| signature verifies against `pubkey_b64` | 206 of 206 |
| one signing key for the whole file | yes, one public key |
| answering route (`winning_route`) | 145 × `nvidia` / `nvidia/nemotron-3.5-super-vl-120b-a12b`; 61 failed, no route |
| receipt `status` | 145 `ok`, 61 `failed` |
| receipt `app` | 182 `proven-review`, 24 `proven` |
| receipt `request_id` equals the chat response `id` (`chatcmpl-…`) | 27 of 27 requests of one review's chat log |

The last row is what ties a receipt to a request: the `id` Proven receives in the chat response is the receipt's
`request_id`, so any answer Proven shows can be traced to the signed receipt that names the model that produced it.

## Not run

- Why 61 requests failed is not yet classified (timeouts, provider errors and refusals read the same in this count).
- The receipts are read inside the VM. Copying a dated extract out, with the key's public half, for the evidence
  folder is still to do.
- Nemotron reviews of the fleet's own changes are paused from 22 October 08:00Z until the NVIDIA review moves to an
  agent loop; the changes accepted meanwhile are listed in each track's conversation log, under *Owed an NVIDIA review*.
