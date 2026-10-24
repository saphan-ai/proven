# What shipped on day 1, 22 October 2026: a walk through a terminal

Everything below was run on 22 October 2026 on a clean clone of `main` (`39225b0` and `16e269a`, before the
[`day-1`](https://github.com/saphan-ai/proven/tree/day-1) tag), with Go 1.26.8 on macOS, and recorded the same day in
the documentation. The output is pasted as it came, shortened where marked, and each step has a note for a person.
The next day: [what shipped on day 2](SHIPPED_DAY_2.md).

## 1. A whole review, with no network and no keys

```text
$ go build -o /tmp/proven ./cmd/proven
$ mkdir -p /tmp/proven-run && cd /tmp/proven-run
$ /tmp/proven scan --sample "$OLDPWD/testdata/fixture/discount"
finding F1 ERROR Discount is applied twice
  billing/discount.go:5-5
  Discounted subtracts the percent twice, so 100 at 20 percent becomes 60 instead of 80.
  evidence test billing/discount_test.go TestDiscounted fails_on_change=true passes_on_base=true
  command: go test ./billing -run TestDiscounted
  output: base: PASS
change: FAIL got 60 want 80
finding F2 ERROR Dependency leftpad v1.2.3 is on the known-exploited list
  go.mod:5-5
  evidence query local-sample snapshot=sample-2026-10-22
  SQL: SELECT package, version, cve, kev FROM advisories WHERE package = 'evil.example/leftpad' AND version = '1.2.3'
  row: evil.example/leftpad | 1.2.3 | CVE-2026-0001 | true
finding F4 ERROR CVE-2026-0001 for leftpad v1.2.3 is on the KEV list
  ...  (shortened)
failed services: none
trace: /tmp/proven-run/logs/trace.jsonl
```

**Note.** On day one, all six agents ran end to end for the first time: the Planner names what might be wrong, the
Investigator and the Prover look for proof, the Skeptic tries to refute each claim with a model from another family,
the Judge decides by evidence, and the Writer reports. In sample mode every model answer, test result and query row
comes from a recording, so anyone gets the same result on a laptop.

Two kinds of proof are visible: **F1** is a bug proved by a test that passes on the old code and fails on the new one;
**F2** and **F4** are a library version proved risky by a query and the row it returned.

## 2. What is not reported

The Planner also suggested a fourth problem, F3: *"`note` may include a payment token"*. It is not in the output.
Nothing proved it, so the Judge dropped it, and the trace says why:

```json
{"finding_id":"F3","kind":"verdict","agent":"judge","text":"dropped: no evidence", …}
```

**Note.** This is the point of Proven: a plausible worry without proof never reaches the review. The trace keeps
every claim, piece of evidence, refutation and verdict, one line each.

## 3. Proven refuses to write into the code it reviews

```text
$ cd proven && /tmp/proven scan --sample testdata/fixture/discount
hmem: state directory <repository>/state/hmem is inside the repository
exit: 2
```

**Note.** Proven keeps its memory and its trace outside the repository it reviews, so a review never changes the code
under review.

## 4. Memory between reviews: Meterless's engines, checked by Meterless's own tests

```text
H-MEM conformance
  PASS rv-001 a record with no source is rejected by construction
  PASS fd-003 wrong subtracts exactly 0.20 and adds the review tag
  PASS ro-001 superseded penalty dominates: a superseded record ranks below a fresh one ...
  ...  (shortened)
  9/13 checks passed

World Model conformance
  PASS prov-001 writes without provenance are rejected
  PASS idem-001 ingest idempotency: same batch twice produces an empty diff
  ...  (shortened)
  7/8 checks passed
```

**Note.** H-MEM remembers which findings were refuted, so they do not come back in the next review; the World Model
remembers what each review has seen. Both are built in Go to Meterless's specifications and run against Meterless's
own conformance suites. The checks that do not pass yet are the sections not built in the first day.

## 5. Is the live setup ready?

```text
$ go run ./cmd/livecheck
generator: not ready: configuration refused: METERLESS_RUNTIME_TOKEN="": the live generator needs a runtime token; …
guardrail: not ready: configuration refused: PROVEN_GUARDRAIL_KEY="": the guardrail needs a key; …
query: not ready: configuration refused: ZETARIS_REST_URL="": the live query adapter needs a base URL; …
container: not ready: prover: Prover off: container runtime "docker" not found: …
summary: 0 ready, 4 not ready
```

**Note.** One command checks every live service a review needs: Meterless Nightshift for the models, NVIDIA's safety
model as the guardrail, Zetaris for the evidence, and a container runtime for the tests. It names each one that is not
ready, and never prints a secret.

## 6. A live demo a judge can run

```text
$ bash demo/live/run.sh
```

**Note.** A small Go change with one bug a test can prove and one dependency a known-exploited query can confirm. The
script builds Proven, checks the live services first, and only then runs the live review: the models through
Meterless Nightshift, each answer from an NVIDIA model, every prompt checked by the guardrail, and the evidence from
Zetaris. See [demo/live/README.md](../demo/live/README.md).
