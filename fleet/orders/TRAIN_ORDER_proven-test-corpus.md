# TRAIN ORDER — proven-test-corpus

> **Public copy.** The original is in our private record. This copy says the same thing; the owner's words written in
> Polish are given in English only. Nothing needed blacking out: the order names no machine, path or private mechanism.
>
> **State of this document:** a draft. The scribe wrote it; the master may change it; the owner has not signed it yet.
>
> - written by: the scribe, 2026-10-19

- master: (to be appointed by the owner)
- status: BACKLOG — written 2026-10-19 by the scribe from the owner's order; the master may change it, then hands it to
  the owner to sign
- track: proven-test-corpus
- repo: proven
- base: main at the cut
- author: SCRIBE
- deadline: **before 2026-10-22 00:00 UTC, documentation only — no code at all** (owner, 2026-10-19). Executable
  fixtures (W2, W3, W7) start when the build window opens, and are dated and declared as code written in the window.

## §0 THE OWNER, VERBATIM

2026-10-19, to the scribe (written in Polish; given in English): *"Do the measurements yourself and prepare the train
that builds the test data for Proven,"* with the owner's written order
"SCRIBE ORDER — Turn real review disagreements into Proven's public test corpus" (19 October 2026), kept in full in the
private record. Its binding points, in short:
- extract the recorded disagreements; account for all 31; report the exact gap; never invent or force the count;
- every finding needs evidence; confidence, model identity, majority or the eventual merge are never the truth label;
- small synthetic Go and TypeScript cases with a contract, provenance, both claims and their evidence, an expected
  outcome, a reproduction plan, and limits; the expected answer kept apart from the reviewer's input;
- a dated preparation corpus, declared as pre-existing evaluation material; no Proven product code before the window;
- `tests/README.md` as the public entry point; the owner reviews and publishes; the scribe owns extraction,
  traceability and assembly; existing roles build and check fixtures; no new roles, no new dispatch authority.

2026-10-19, later (written in Polish; given in English): *"But NO CODING — documentation only!"*
⇒ Before the window opens, this train writes documents only: contracts, provenance, manifest, README, the boundary
declaration. No fixture, check script or test file is written before 2026-10-22 00:00 UTC.

2026-10-19, later (written in Polish; given in English): *"I mean: we prepare all the material now — documentation,
specifications and so on — in a neutral language. And the coding starts at the same time as proven's implementation.
As we do it."*
⇒ Before the window: EVERY case contract (all 21), written language-neutral — behaviour, inputs, observable results and
the expected outcome, with no function signature, package or toolchain of any language. From 2026-10-22 00:00 UTC the
fixtures are coded in parallel with proven's implementation, by test authors who are not proven's implementers — the
same separation as our own rule that tests are written apart from the implementation.

2026-10-19, last word (written in Polish; given in English): *"Agreed, we build now; the test code is 'prior art', but it
must exist to show how Proven works."*
⇒ THIS RULING REPLACES the "no code before the window" lines above. The case code (fixtures and their checks) is built
NOW, before the window, by test-author legs from the language-neutral contracts, and declared in `tests/README.md` and
the submission as **prior art**: pre-existing evaluation material, dated, not judged as hackathon work, never re-dated.
Proven's own code still starts 2026-10-22 00:00 UTC, and nothing in the corpus imports or prescribes it.

## §1 WHAT THIS TRAIN IS

The measurement is done (scribe, 2026-10-19, the private reconciliation; its public, redacted form is [`tests/corpus/PROVENANCE.md`](../../tests/corpus/PROVENANCE.md)):
- The owner's 31 are exactly the published population: concurrent cross-family review runs on one target, ≤30 min apart.
- Re-read by hand, **3 of the 31 are not disagreements** (two were both-reject, one review died before a verdict). The
  published count was wrong: **28 of 40 readable pairs differ (70 %), not 31 of 41 (76 %)**. Cause: the 2026-10-16
  extractor took a template line as the verdict.
- The 28 come from **19 incidents** and yield **21 distinct case candidates** (PV-001 … PV-021). Five pairs are
  documented without a fixture: three environment artefacts and two repeat rounds.
- Mechanisms of disagreement: policy or threshold (both saw the same fact, weighed it differently), a test that checks
  the wrong mechanism, findings outside what the order pointed at, contradiction on fact, different assigned subjects,
  and environment (a reviewer that could not run a check).
- Technical truth, established by later measurement or round: 27 pairs an established defect, 3 context-dependent,
  1 a human scope decision.
- Language: most cases are Go by nature; PV-001, PV-002, PV-004 and PV-005 suit TypeScript.

When the train is done, `proven/tests/` holds a dated, redacted, reproducible corpus that exists before Proven's code,
and a reader can rerun every executable case and challenge every expected outcome.

## §2 WAGONS

### W0 — correct the published count (owner act, before anything is published)
- **lands:** the corrected figures (28 of 40, 70 %; qwen-code reject rate 19 %) in proven's README and
  `planning/answers/README.md`, with one dated line saying what was wrong and why. The video and social posts are the
  owner's to correct.
- **dod:** `grep -rn '31 of 41\|31/41' proven/` returns nothing outside the dated correction note.
- **review:** master-verify-only

### W1 — ALL 21 case contracts (scribe), language-neutral, written before the window
- **lands:** `tests/corpus/cases/PV-<NNN>/CASE.md` for all 21 cases, language-neutral (no signature, package or
  toolchain; the intended fixture language is a hint in the manifest only): story, contract, provenance (public, redacted), synthetic transformation, both claims and what their evidence
  established, expected outcome, reproduction plan marked **not run**, limits. Expected-outcome material goes under
  `expected/`, never beside the reviewer's input. `tests/corpus/manifest.json` with every PV and its status, and
  `tests/corpus/PROVENANCE.md`: the public, redacted reconciliation of all 31.
- **dod:** every CASE.md answers the order's nine points; the manifest counts match the reconciliation; a redaction pass
  (W5) finds no internal name, path or mechanism.
- **review:** named (advisor reads the contracts against the reconciliation)

### W2 — first executable batch (test-author legs, one leg per case) — NOW, as declared prior art
- **lands:** for PV-001 (TypeScript, a guard that reads the last duplicate attribute where an HTML parser keeps the
  first), PV-004 (TypeScript, a "strict" fenced-block extractor that accepts malformed fences), PV-013 (Go, a refusal
  message that prints the rejected secret), PV-014 (Go, a client that replays a token-bearing body across a redirect to
  another host), PV-020 (Go, a required check that cannot run in the reviewer's environment; expected outcome: a human
  scope decision, not a defect): `base/`, `change/`, a readable diff, and a local check script with exact commands.
  Standard library only (Go), a minimal pinned TypeScript toolchain; no model keys, no network beyond localhost test
  servers, no fleet access. The leg writes the fixture from CASE.md alone and never sees `expected/`.
- **dod:** each check run on base and on change, commands and exit codes recorded in CASE.md §8; for a regression the
  check is green on base and red on change; for a false alarm the refuting measurement is shown; a deliberately failing
  probe is labelled as expected evidence.
- **review:** named

### W3 — independent check of the first batch (a reviewer of another model family than W2's legs)
- **lands:** a review per case: does the fixture reproduce the technical condition; does the expected outcome follow from
  the evidence; is anything private left. Corrections go back to W1/W2, never edited by the reviewer.
- **dod:** each case's review verdict and its corrections recorded; a case that does not reproduce keeps status
  `unresolved`, never a weakened check.
- **review:** named (this wagon is the review)

### W4 — the public entry point (scribe, owner reviews)
- **lands:** `tests/README.md` with the eight sections of the owner's order (why, where from, the work, what is
  available now as a dated table derived from the manifest, how to use, what it contributes, what it does not
  establish, chronology and the build-window declaration); `tests/corpus/CHANGELOG.md`; one short factual paragraph and
  a link in the main README and the planning page. Chronology, measured: the fleet's record runs from 2026-07-06 (first
  record commit) and 2026-07-16 (first run); the disagreements date 2026-09-01 … 2026-09-28; the corpus is prepared
  2026-10-19 … 2026-10-21.
- **dod:** every count in the README is produced by one command over `manifest.json`, printed beside the table.
- **review:** named

### W5 — redaction pass (advisor)
- **lands:** a check of every public file against proven/AGENTS.md: no internal host names, paths, seat names,
  private repository names, numbered laws or internal recipes; source links that work for an outside reader.
- **dod:** the deny-list check returns nothing; the REDACTED lists are complete.
- **review:** master-verify-only

### W6 — the boundary declaration and the open eligibility question (advisor)
- **lands:** in `tests/README.md` and the submission inventory: the corpus DOCUMENTS (contracts, provenance, manifest)
  are pre-existing material, dated, independent of Proven's APIs and architecture. No code exists before the window; the
  fixtures are written in it and dated as such. The question for the organisers, if any remains: may case contracts
  written before the window be used as evaluation input during it?
- **dod:** the question is written in the owner's decision list; nothing in the corpus imports or prescribes Proven code.
- **review:** master-verify-only

### W7 — the remaining reconstructable cases
- **lands:** PV-002, PV-003, PV-005 … PV-012, PV-015 … PV-019, PV-021 the same way (W1 contract → W2 fixture → W3 check).
  What lands after 2026-10-22 00:00 UTC is dated as such and declared, never back-dated.
- **dod:** the manifest shows every PV with a status; the README table is regenerated.
- **review:** named

## §3 ORDER OF WORK
- Before the window (documents only): W0, W1, W4, W5, W6, in parallel.
- Now, as declared prior art: W2 per case as soon as its contract lands, W3 after W2 per case, W7 last.
- The owner publishes: nothing in this train pushes to the public repository.

## §4 DEFINITION OF DONE
- All 31 source pairs accounted for in `PROVENANCE.md` with one status each.
- Before the window: at least one Go and one TypeScript case reproduced by its check and independently confirmed by
  W3, declared as prior art with its real dates.
- `tests/README.md` published by the owner, with counts tied to the manifest and the chronology stated.
- Every executed command and its outcome recorded; everything not run says "not run".

## §5 OUT OF SCOPE
- Proven's reviewer, orchestration, adapters and product test harness (they are built in the window, not here).
- Calling the corpus a benchmark or a blind test: product development will see these examples.
- Correcting the video and social posts (the owner's).
