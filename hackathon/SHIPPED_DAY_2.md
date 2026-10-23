# What shipped on day 2, 23 October 2026: a walk through a terminal

Everything below was run by Marcin on 23 October 2026, about 14:00 UTC, on a fresh clone of `main` at `ae34d44`, on
macOS, with **no keys and no network** (`env -i`). The output is pasted as it came, shortened where marked, and each
step has a note for a person. To run it yourself, see [docs/INSTALL.md](../docs/INSTALL.md).

## 1. The command line

```text
$ proven --help
proven — review a change

usage:
  proven scan [flags] [TARGET]
  proven ci [flags]
  ...
exit codes:
  0  review done; no findings, or findings without --error
  1  review done; findings, with --error (the default in ci)
  2  fatal error: proven could not review
  7  the configuration was refused when it loaded
```

**What is new.** Proven now has its own command line, shaped like the scanners a CI pipeline already runs: `scan` for a
developer, `ci` for a pipeline, flags for format, baseline, severity and paths, each with its environment variable, and
exit codes a pipeline can act on. JSON and SARIF output are being finished; the help says so instead of pretending.

```text
$ proven scan --sample --no-such-flag
flag provided but not defined: -no-such-flag
known flags: --baseline-commit, --config, --error, --exclude, --help, --include, --json, ...
exit: 2
```

**Note.** A mistake names itself and lists what is allowed.

## 2. A review in sample mode: a bug proved by a test

```text
$ proven scan --sample testdata/fixture/discount
...
finding F1 ERROR Discount is applied twice
  billing/discount.go:5-5
  Discounted subtracts the percent twice, so 100 at 20 percent becomes 60 instead of 80.
  rule discount-applied-twice
  evidence test billing/discount_test.go TestDiscounted fails_on_change=true passes_on_base=true
  command: go test -run '^TestDiscounted$' ./billing
  output: base: PASS
change: FAIL got 60 want 80
failed services: none
```

**Note.** This is the heart of Proven: the finding is reported because a test **passes on the old code and fails on the
new one**. Look at the command: since today, Proven builds it itself from the test's file and name. Before, it could
come from a model's answer; the security review found that, and the fix landed the same day
([SECURITY.md](../SECURITY.md)).

## 3. The same review: a dependency on the known-exploited list

```text
finding F2 ERROR Dependency leftpad v1.2.3 is on the known-exploited list
  EPSS 0.50000 (90th percentile, 2026-10-20)
  go.mod:5-5
  evidence query local-sample snapshot=sample-2026-10-22
  SQL: SELECT package, version, cve, kev FROM advisories WHERE package = 'evil.example/leftpad' AND version = '1.2.3'
  row: evil.example/leftpad | 1.2.3 | CVE-2026-0001 | true
  ...  (shortened: F4 and the OSV finding add the KEV row and the OSV range for the same package)
```

**Note.** In plain words: the change adds a library version that is on the known-exploited list, its advisory says it
is fixed in 1.2.4, and EPSS puts it among the 10% most likely to be exploited. The evidence is the query and the row it
returned, so anyone can run it again. Today this reads like a log; tomorrow the default output becomes a review for a
person, with the queries one flag away ([ADR-0033](../docs/adr/ADR-0033-the-text-review-is-written-for-a-person.md)).

## 4. The guardrail stops an unsafe prompt

```text
$ proven scan --sample testdata/fixture/unsafe-prompt
...
failed service: guardrail: unsafe: Criminal Planning/Confessions, Other
```

**Note.** Every prompt and every answer goes through NVIDIA's Nemotron content-safety model. Here the change carries a
prompt the guardrail calls unsafe, and the review stops that step and says why.

## 5. Two more scenarios: nothing proved, nothing reported

```text
$ proven scan --sample testdata/fixture/refuted
...
failed services: none
$ proven scan --sample testdata/fixture/unknown-author
...
failed services: none
```

**Note.** No finding. A claim that the evidence does not back never reaches the review; the trace
(`logs/trace.jsonl`) shows each dropped claim and why.

## 6. Is the live setup ready?

```text
$ livecheck
generator: not ready: nightshift: METERLESS_RUNTIME_TOKEN is empty
guardrail: not ready: guardrail: PROVEN_GUARDRAIL_KEY is empty
query: not ready: configuration refused: ZETARIS_REST_URL="": the live query adapter needs a base URL; ...
container: not ready: prover: image golang:1.26 is not present: ... No such image: golang:1.26
summary: 0 ready, 4 not ready
exit: 1
```

**Note.** With nothing configured, every live service is named with the reason, and no secret is printed. Two things
found by this walk, both ordered for 24 October: `livecheck` will print the next step (`docker pull golang:1.26`); and a
live `proven scan` with no keys still stops at the first missing setting instead of naming every down service and
reviewing with what is left. Tomorrow a single NVIDIA API key will be enough to run a live review
([ADR-0031](../docs/adr/ADR-0031-anyone-with-an-nvidia-key-can-run-a-live-review.md)).
