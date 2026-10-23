# TRAIN ORDER — proven-enterprise
<!-- Public copy. Original: the train's order in the fleet's private record (identical except this line). -->

- master: MASTER (claude-code, claude-opus-5-5), the master of proven-security-review, handed this train by the owner on 2026-10-23
- status: DRAFT — written by the master 2026-10-23; waits for the owner's signature
- track: proven-enterprise
- repo: proven (`saphan-ai/proven`), the product code
- base: main at each wagon's cut (main was c3917a4 when this order was written)
- author: MASTER of proven-enterprise

## §0 THE OWNER, VERBATIM

2026-10-23: features closer to enterprise, built on what Proven already has (trigger of ADR-0034, ADR-0035, ADR-0036).
2026-10-23, relayed by the advisor: *"START TODAY, not Saturday."* The owner confirmed the hand-over to this master in
this master's session on 2026-10-23.

## §1 WHAT THIS TRAIN IS

Three features that an enterprise asks for before it runs a code reviewer: a mode in which nothing leaves the
organisation's network ([ADR-0034](../../../docs/adr/ADR-0034-proven-runs-air-gapped.md)), a signed evidence pack per
review that an auditor can verify ([ADR-0035](../../../docs/adr/ADR-0035-every-review-leaves-a-signed-evidence-pack.md)),
and one organisation policy with exceptions that have an owner and an expiry
([ADR-0036](../../../docs/adr/ADR-0036-an-organisation-policy-and-exceptions-that-expire.md)).

When the train is done, each of the three works in a live review and in sample mode, each has its test as its proof,
and each has an operator page in `docs/`.

## §2 HOW EVERY WAGON IS BUILT AND REVIEWED

1. **One wagon per ADR.** Each wagon order carries the ADR's decision points as its acceptance, word for word, and the
   test the ADR names as its proof. The test is written with the implementation plan, before code.
2. **The code lives in other trains' packages.** `internal/` is proven-reviewer-core's, `cmd/proven` and `internal/cli`
   are proven-cli-and-action's. Before each wagon is cut, the master sends both masters the files the wagon will
   touch; a wagon is cut from current `main` and absorbs `main` before it lands.
3. **Executors:** claude-code or codex seats; Cursor is optional (owner, 2026-10-23). W1 and W2 are security class
   (network refusal, signatures): claude-code or codex, never qwen. Orders word security work as the property that must
   hold and its proof (law #56).
4. **Reviews are independent of this master.** Every wagon gets a named review, two for W1 and W2 (security class), one
   for W3. This master does not order them: the master of proven-reviewer-core orders them (C11). The security pieces
   of proven-security-review also read this code once it is on `main`, like any other code.
5. **Quote per executor round:** 12 USD, cap 24 USD (the executor class p90), wall 150 minutes. One fix round per wagon
   at most is budgeted. Three wagons, two rounds each: cap 144 USD.
6. **Signatures:** the owner signs this order and each wagon order.

## §3 WAGONS

### W1 — Proven runs air-gapped (ADR-0034)
- **lands:** one setting `PROVEN_NETWORK` (`public` by default, `local` for air-gapped); in `local`, every configured
  address (generator, guardrail, Zetaris, and any other outbound client) is checked at load: loopback, private ranges,
  or a host named in a new allow-list setting; any other is refused at load and named; the hosted NVIDIA guardrail mode
  is refused in `local`. Rows in `docs/CONFIGURATION.md`; a new `docs/AIR_GAPPED.md` (what runs locally, the settings,
  how to check the mode holds).
- **dod:** a test that loads `local` with one public address and asserts the refusal names it; a test that runs a
  sample-mode review in `local` with a dialer that refuses every non-allowed host and records every dial, and asserts
  the review completes and no dial left the allowed hosts. Negative control: remove the load check and the first test
  reddens by name.
- **review:** named (two, security class)

### W2 — every review can leave a signed evidence pack (ADR-0035)
- **lands:** `--evidence-pack <file>` writes one archive: findings, trace, each test command and output, each query and
  its rows, the Nightshift receipts, the reviewed commits, the settings used with every secret removed; a manifest with
  a SHA-256 per file; a signature over the manifest with a key the organisation configures (a setting naming a key
  file; a missing key is named, never printed). `proven verify <pack>` checks the signature, every hash and every
  receipt, and prints what was reviewed, when, by which models and on what evidence. `docs/EVIDENCE_PACK.md`.
- **dod:** a test that writes a pack in sample mode and verifies it; a test that changes one byte of any file and
  asserts `verify` refuses and names the file; a test that asserts no configured secret appears anywhere in the pack.
- **review:** named (two, security class)

### W3 — one organisation policy, exceptions with an owner and an expiry (ADR-0036)
- **lands:** a setting naming an organisation policy file, read before each repository's `proven.yaml`; a local rule
  weaker than the policy is refused by name; an exception names the finding (rule and location), a reason, an owner and
  an expiry date; until expiry the finding is reported as *accepted* with those four fields; after expiry it counts
  again and the review says the exception expired; an exception without an owner or an expiry is refused.
  `docs/POLICY.md`.
- **dod:** one test each: a weaker local rule refused; an accepted finding shown with its owner; an expired exception no
  longer accepted; an exception with no expiry refused.
- **review:** named (one)

## §4 ORDER OF WORK

- W1 first, dispatched after the owner's 15:00Z push on 2026-10-23: the smallest, and it touches only settings.
- W3 next: it touches the review's settings and its output, not the network.
- W2 last: it reads the trace and the output of the other two.
- A wagon does not wait for another train's wagon unless both touch the same file; then the later one absorbs `main`.

## §5 DEFINITION OF DONE

1. W1–W3 merged on `main`, each with its tests green and its negative control shown.
2. `docs/AIR_GAPPED.md`, `docs/EVIDENCE_PACK.md`, `docs/POLICY.md` on `main`, and every new setting in
   `docs/CONFIGURATION.md`.
3. Each wagon passed its named review(s), ordered by the master of proven-reviewer-core.
4. The record carries the closing line:
   `DONE · proven-enterprise · nothing in flight · nothing owed to the owner · THANK YOU`.

## §6 LEDGER

| date | act |
|---|---|
| 2026-10-23 | hand-over: the advisor relayed the owner's assignment (messages bb32b90c9fff7f1c, eca6d07c6a764212); the owner confirmed it to this master in its session |
| 2026-10-23 | order written by the master |

## §7 OUT OF SCOPE

- The security review of Proven's code; that is proven-security-review.
- Licences. The owner decides them once; no review checks them.
- A read-only Zetaris key and the query templates of ADR-0032; those are proven-reviewer-core's.
