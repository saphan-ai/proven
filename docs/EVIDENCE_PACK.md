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
| `findings.json` | the review result: every finding with its evidence, the facts of the Skeptic's refutation that the Judge's rule reads, whether query rows are bound to the change, the proof counts, the failed services, the review class and the reviewer models |
| `trace.jsonl` | this run's lines of the trace: claims, evidence, refutations, verdicts, guardrail checks and one line per model call |
| `tests.json` | every test and every deterministic tool run behind a finding: the command, whether it failed on the change and passed on the base, and its output |
| `queries.json` | every query behind a finding, the engine, the snapshot, the columns, every row and whether the rows are bound to the change |
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
6. **Outcomes.** Each finding's outcome is worked out again from the evidence the pack records for it, with the Judge's own rule and the review's own function for the proof level (`review.Rederive`; the code is shared, not copied). `verify` runs no test and no query and opens no connection: it reads the recorded results only. The pack must say, for each finding:
   - its **proof level**: `proved-by-test` when a recorded test failed on the change and passed on the base, `proved-by-tool` when a recorded tool result was on the change and not on the base, `proved-by-data` when recorded query rows are marked as bound to the change. The level is that of the strongest evidence that counts under the Judge's rule (one shared predicate, `judge.EvidenceCounts`). A test that did not fail on the change does not lift a finding above the bound query that does count;
   - whether it is **reported**: its own evidence counts under that rule, and the recorded refutation (the Skeptic's test result and query rows, in `findings.json`) does not hold;
   - whether it is **left open** (`unresolved`): the refuting test or query could not run;
   - the **proof counts**, which must be the counts of the findings recorded.

   `tests.json` and `queries.json` must be the evidence `findings.json` holds for each finding, value for value, so one of them cannot be changed alone. A refusal names the finding, the recorded outcome and the re-derived one, for example `evidence pack refused: outcome check failed on findings.json: finding F3 records proof level "proved-by-test", the rule gives "proved-by-data" for the evidence recorded`. A finding whose pack records no refutation still verifies; `verify` prints it under `not re-derived`, because whether it was refuted or left open cannot be worked out, and does not count it in `outcomes: N of M finding(s) re-derived`. The same holds for a pack that does not record `Bound` for a query (a version-1 pack written before it was recorded): absent is not false, so the pack is not refused as forged. The rule is run with the absent `Bound` read both ways: every component both readings agree on is checked, and only the components they disagree on are named as not re-derived. A finding the manifest's `redaction_moved` lists (redaction changed the text the rule reads for it) is treated the same way: the only thing the rule reads that text for is whether a refuting query's rows are about the finding, so `verify` runs the rule with the rows read as about it and as not, checks every component both readings agree on (the proof level and the proof counts always; the reported and open state unless the rows decide them), and names only the others as not re-derived. The marker never skips a check the recorded evidence still decides. The line `outcome of every finding re-derived` is printed only when none is named.

It then prints what was reviewed, when, the models with their receipt ids, and the evidence: findings by level of proof, test and tool runs, queries and trace events. Exit code 0 means the pack verified. A refusal names the check and the file, for example `evidence pack refused: hash check failed on findings.json`, and exits 2. Every text that comes from the pack or from its path is shown with control characters, bidirectional marks and invalid bytes written as `\xNN` or `\uNNNN`, so a pack cannot move the cursor or recolour your terminal.

`proven scan --evidence-pack` verifies the pack it has just built before it writes the file. A secret the redactor catches is blacked out in `MANIFEST.json` as in every other file, and a pack that would not verify is not written. A key setting that cannot be used exits 7.

The pack's path must be its own. `--evidence-pack` naming the same file as `-o`, `--text-output`, `--json-output` or `--sarif-output` — by the same spelling, by a link, or by a link to its directory — is refused before the review starts. If a link appears after that check, the run stops with the pack left as it was written.

**What a pack proves.** A pack proves that each verdict follows from the evidence it records, not that the evidence is true. Whoever holds the signing key can sign a pack that records a test result that never happened; the re-derivation then agrees with it. What stops this is the key, who holds it, and the trace, the commits and the receipts a reader can compare with the repository and the provider.

**What verify does not check**, and says every time: the provider's own signature on each receipt, because the pack keeps the receipt id, the route and the model, not the signed body; and that the recorded evidence is true. Two more limits: the Judge's rule is the only rule re-derived (a finding the Judge dropped is not in the pack, so a pack cannot show that a finding was left out), and the pack does not record the model's own claims, only the facts the rule reads.

## The forgery bench

`TestForgeryBench` in `internal/evidence` takes a sample pack of four findings (one proved by a test, one by a tool, one by data, and one left open), lists every scalar value of every record (each scalar leaf of every JSON file, each scalar leaf of every trace line, the manifest's scalar values and its signature), and changes one at a time. It does not change an array or an object as a whole. Each value is replaced in the way that matters for it: a text gets a letter, a flag is turned over, a number goes up by one, and a proof level becomes each of the other levels.

| part | the forger | the bench requires |
|---|---|---|
| (a) hashes left | changes one value and leaves the archive as it is | every change refused by the hash check or the signature check |
| (b) hashes recomputed | changes one value, recomputes every hash in the manifest, and signs the manifest again with the key | every change that moves a finding's outcome refused by the outcome check or the format check |

A change moves an outcome when the pack, read by a path that does not go through `verify`, states a different proof level, reported state, open state or count for any finding, or gives a different outcome under the rule (`review.Rederive`: level, reported state, open state) for the evidence `findings.json` holds, or states them in a way no reader can read. The bench fails if any change in (a), or any outcome-moving change in (b), verifies. The signature file has one value that is not JSON, and it is changed in (a) only: signing again would put the original back.

A text is also replaced by the word `no`, which is what the Judge reads as a denial. In a refutation row it moves an outcome through the Judge's contradiction rule (`TestForgeryBenchChangesQueryRowToDenial` pins that the bench makes this change, and `TestADenialWrittenIntoARefutationRowIsRefused` pins that `verify` refuses it). A null becomes a text.

**What the bench measures, and what it does not.** The bench decides "this change moves an outcome" from three readings: the proof level and open state that `findings.json` itself records; the whole outcome that `review.Rederive` (the function `verify` calls) gives over `findings.json`, that is its level, its reported state and its open state, so a change that moves the rule's level and not the recorded one counts; and `review.Rederive` over the evidence read back from `tests.json` and `queries.json`, with the recorded proof counts. So the count shows that `verify` refuses the one-value changes the bench makes, and that part (c) refuses the four field-removal changes it makes. It does not show that no change escapes. The bench changes one value at a time, with these exceptions. It never empties a row list, never makes a refutation null, never sets a `Kind` to another valid kind, and removes a `Bound` field only in part (c). It does not show that the rule is right: a wrong rule would be wrong in both. The rule is pinned by its own tests in `internal/review` and `internal/agent/judge`.

**What a redaction marker hides.** A finding in `redaction_moved` is checked for every component the recorded evidence decides, by the same two-readings rule as a missing `Bound`. The pack does not record which components redaction moved: `verify` derives it from the rule, so a signer who adds an id to the list gains nothing for a component the evidence decides. Where the finding's refutation holds no query rows, the text decides nothing and every component is checked. Where it holds rows, the reported state and the open state may be named not re-derived; a signer who lists an id can hide a wrong value of those, and `verify` names the finding. The refusal is of the outcome as a whole: the pack records every finding as reported, so `verify` refuses a finding when no reading of the text (rows not about it, rows about it) gives the recorded level and open state together with the finding reported. A signer cannot hide an outcome that no reading gives; only a value that some reading gives can be hidden, and `verify` names the finding for it (`TestReviewR8RedactionCannotHideAnOutcomeNoReadingGives`, `TestRedactionMovedFindingMatchedByOneReadingVerifiesAndIsNamed`).

**What a missing `Bound` hides.** When a query's `Bound` is absent, `verify` runs the rule twice, with that `Bound` read as false and as true. A component of the outcome (proof level, reported, open) on which the two readings agree is checked as for any finding: the rest of the recorded evidence decides it, for example a test that counts. Only a component on which they differ is named as not re-derived, and a finding with any such component is not counted in `outcomes: N of M finding(s) re-derived`. The check is of the outcome as a whole, not of each component alone: among the readings, `verify` refuses a finding when none gives its recorded outcome (reported, proof level and open state together; the pack records every finding as reported), counts it re-derived only when every reading agrees and matches the record, and names it not re-derived when the readings that match the record differ from one another. So a signer who removes a `Bound` can hide a value that one reading gives, and `verify` names the finding; a signer cannot hide an outcome that no reading gives (`TestReviewR8MissingBoundCannotHideAnOutcomeNoReadingGives`). A pack at format version 1 cannot tell a pack written before `Bound` was recorded from a later pack that had the field removed, and `verify` treats both alike: the format version was not changed.

The counts of the last run (`go test -count=1 -v -run TestForgeryBench ./internal/evidence`, 26 October 2026, after the sample gained main's fields), tallied by the check that refused:

| part | changes | result |
|---|---|---|
| (a) hashes left | 430 | 370 refused by the hash check, 60 by the signature check, 0 verified |
| (b) hashes recomputed | 429 | 118 move an outcome, and 118 of them were refused: 84 by the outcome check and 34 by the format check (a value that no reader can read, such as a text where a null stood). The other 311 leave every outcome as the bench reads it (a title, a trace line, a time): 162 were refused by another check (102 by the outcome check, 22 by the receipts check, 21 by the hash check, 14 as a missing file, 2 by the signature check, 1 by the format check), and 149 verify, as they should. The 102 outcome refusals are the cross-file agreement check (`tests.json` or `queries.json` no longer matches `findings.json`). Before the oracle read the rule's whole outcome over `findings.json`, the bench counted 112 moving and 305 not; four changes that move the rule's outcome moved to the first group (F4's `FailsOnChange` and `PassesOnBase` in `findings.json` among them), and all four are refused. `TestBenchCountsRuleLevelChange` pins this |
| (c) `Bound` removed | 4 | a finding's `Bound` fields removed from `findings.json` and `queries.json`, with a result the rest of the evidence decides changed in the same pack (F4's open state; F4's level; F3's open state; F3's level, with the counts kept consistent): 4 refused by the outcome check, 0 verified. Print: `TestForgeryBenchBoundRemoved` |
| (d) redaction marked | 16 | a finding listed in `redaction_moved` and, in the same pack, its open state flipped or its level moved to each other level (counts kept consistent), for each of the 4 findings: 16 refused by the outcome check, 0 verified. In the sample, the rows F3's refutation holds contain no denial, so a change of text does not change an outcome and no change is excused. Print: `TestForgeryBenchRedactionMarked`. The case where the text does decide a component is `TestRedactionNamesOnlyTheComponentsTheTextDecides` |

Negative controls, 25 October 2026. With the blanket skip of a redaction-moved finding restored, `TestReviewRedactionMarkerCannotHideDecidableOpenState`, `TestRedactionNamesOnlyTheComponentsTheTextDecides` and `TestForgeryBenchRedactionMarked` fail (16 of 16 changes verified). Without the re-derivation, (b) fails by name: with the call removed from `verify`, `TestForgeryBench/b-hashes-recomputed` reported 111 of the 112 outcome-moving changes as verified. With the skip restored (a finding with an absent `Bound` named not re-derived whole), `TestMissingIrrelevantBoundDoesNotHideWrongOutcome` and all 4 cases of `TestForgeryBenchBoundRemoved` fail. Negative controls, round 5, 25 October 2026. With the whole-outcome check turned back into one check per component, `TestReviewR8MissingBoundCannotHideAnOutcomeNoReadingGives` and `TestReviewR8RedactionCannotHideAnOutcomeNoReadingGives` fail. With the bench's oracle turned back to the recorded level and open state, `TestBenchCountsRuleLevelChange` fails in both subtests. With the "related" text reading removed from `review.TextReadings`, `TestTextReadingsGiveTheTwoEndsOfWhatRedactionHides` and `TestRedactionMovedFindingMatchedByOneReadingVerifiesAndIsNamed` fail. With the reason clause for redacted text removed, `TestRedactionNamesOnlyTheComponentsTheTextDecides` fails. With the proof level taken from the kind of evidence and not from the evidence that counts, `TestVerifyRejectsProofFromNonCountingEvidence` and `TestALevelNeedsEvidenceOfThatLevelThatCounts` fail.

A sample-mode review has no provider receipts, so its models read `unknown` with no receipt id, and verify reports them as models without a receipt.

## Where a pack is kept

A pack holds source excerpts in test output and trace lines, and every row a query returned. Treat it as you treat the repository and the data it was reviewed against. Where it is stored, who may read it and how long it stays is the organisation's decision. Proven writes the file with mode 0600 and does nothing else with it.
