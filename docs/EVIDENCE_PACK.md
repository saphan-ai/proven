# The evidence pack

ADR-0035. A review can leave one signed file that an auditor can check long after the run: what was reviewed, when, by which models, and on what evidence.

```
proven scan --evidence-pack review.pack [TARGET]
proven verify --public-key pack.pub review.pack
```

## What is in a pack

A pack is one tar archive. It has no folders. Every file below is listed in the manifest with its SHA-256.

| file | holds |
|---|---|
| `findings.json` | the review result: every finding with its evidence, the proof counts, the failed services, the review class and the reviewer models |
| `trace.jsonl` | this run's lines of the trace: claims, evidence, refutations, verdicts, guardrail checks and one line per model call |
| `tests.json` | every test and every deterministic tool run behind a finding: the command, whether it failed on the change and passed on the base, and its output |
| `queries.json` | every query behind a finding, the engine, the snapshot, the columns and every row |
| `receipts.json` | one entry per model use: agent, route, model and the Nightshift receipt id |
| `commits.json` | the target, the baseline and the head as named, and what git resolved each name to |
| `settings.json` | the settings the run used. A secret setting is written as `set` or `unset`, never its value |
| `MANIFEST.json` | the version, the time, the id of the signing key, the commits, the models, and each file's path, size and SHA-256 |
| `MANIFEST.sig` | an Ed25519 signature over the exact bytes of `MANIFEST.json`, in base64 |

Every file passes through the same redactor as the trace before it is hashed, so the value of a configured secret is replaced by a marker that names its variable. The private key is read from its file, signs the manifest, and is not written anywhere in the pack. A value is replaced in every spelling the pack can hold it in: as written, and as a JSON string with or without HTML escaping. A configured secret shorter than 8 bytes cannot be kept out of a pack without damaging ordinary text, so a pack is not written while one is set; the refusal names the variable, never its value.

An empty `baseline` means the whole target was reviewed. An empty `head` means the working tree, which git has no commit for, so `head_resolved` is empty too.

## Settings

| setting | meaning |
|---|---|
| `PROVEN_PACK_KEY` | path to the PEM file of the organisation's Ed25519 private key (PKCS #8). Required with `--evidence-pack`. A missing or unreadable file, or one that group or others can reach (fix: `chmod 600`), is named and the review does not start (exit 7) |
| `PROVEN_PACK_PUBLIC_KEY` | path to the PEM file of the Ed25519 public key (PKIX). `proven verify` reads it unless `--public-key` is given |

Both settings hold a path. A refusal names the setting and the path, and never prints a byte of a key. The rows are also in [CONFIGURATION.md](CONFIGURATION.md).

Making a key pair with OpenSSL 3:

```
openssl genpkey -algorithm ed25519 -out pack.key
openssl pkey -in pack.key -pubout -out pack.pub
```

Keep `pack.key` where only the people and the jobs that sign may read it. Hand `pack.pub` to every auditor.

**A readable key is refused.** A private key file with any group or other permission bit set is refused, the same as a missing one: the review does not start, the refusal names the setting, the path and the mode, and the run exits 7. To fix it, run `chmod 600 pack.key`, then run again. Mode `0600` and `0400` are accepted. The check does not apply on Windows, where file modes do not say who may read. The public key is not secret and is not checked.

## How an auditor verifies

`proven verify <pack>` checks, in this order, and stops at the first failure:

1. **Format.** The file is a tar of plain files: no folder, no link, no repeated name. It is also the exact archive the writer makes from those files. A changed header byte, changed padding or any trailing byte is refused, so the file as a whole is tamper-evident and a stored digest of it stays true. The signature file must be the base64 and one newline, with no other whitespace.
2. **Signature.** `MANIFEST.sig` is a valid Ed25519 signature over `MANIFEST.json` under the public key you hold, and the manifest names that key.
3. **Hash of every file.** Each listed file's SHA-256 is the manifest's.
4. **No file missing, none extra.** The pack holds the listed files and the manifest and signature, and nothing else. The manifest must list every file of the table above, by name; a signed manifest that leaves one out is refused. Each listed size is the file's size.
5. **Receipts.** `receipts.json` is the manifest's own model list. Each model use that says it came from a receipt has a receipt id and a model, and the trace holds a model event with that id and that model.

It then prints what was reviewed, when, the models with their receipt ids, and the evidence: findings by level of proof, test and tool runs, queries and trace events. Exit code 0 means the pack verified. A refusal names the check and the file, for example `evidence pack refused: hash check failed on findings.json`, and exits 2. Every text that comes from the pack or from its path is shown with control characters, bidirectional marks and invalid bytes written as `\xNN` or `\uNNNN`, so a pack cannot move the cursor or recolour your terminal.

`proven scan --evidence-pack` verifies the pack it has just built before it writes the file. A secret the redactor catches is blacked out in `MANIFEST.json` as in every other file, and a pack that would not verify is not written. A key setting that cannot be used exits 7.

**What verify does not check**, and says every time: the provider's own signature on each receipt, because the pack keeps the receipt id, the route and the model, not the signed body; and whether the review was right. A pack proves what was recorded and that it has not been changed since it was signed. It does not prove the findings are true.

A sample-mode review has no provider receipts, so its models read `unknown` with no receipt id, and verify reports them as models without a receipt.

## Where a pack is kept

A pack holds source excerpts in test output and trace lines, and every row a query returned. Treat it as you treat the repository and the data it was reviewed against. Where it is stored, who may read it and how long it stays is the organisation's decision. Proven writes the file with mode 0600 and does nothing else with it.
