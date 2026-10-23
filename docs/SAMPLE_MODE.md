# Sample mode

Sample mode runs a whole Proven review with **no network and no keys**. Every model answer, every test outcome and every
query row comes from a recorded fixture, so anyone can run it on a laptop and get the same result every time.

Every command and every output on this page was run on 23 October 2026 with Go 1.26.8 on macOS, on `main` at `26343c2`
with this page's four new scenarios added. Long output is shortened where marked.

## What sample mode exercises, and what it does not

| Exercised (the real code runs) | Not exercised (replaced by recorded data) |
|---|---|
| The six agents in order: Planner, Investigator, Prover, Skeptic, Judge, Writer | Live model calls through Nightshift |
| The Judge's rule: a finding without evidence is dropped | The Prover running a test in a container: the outcome is recorded |
| The Skeptic's refutation step and the Judge's verdict | Live Zetaris queries: the same SQL runs over local sample tables |
| The trace (`logs/trace.jsonl`), H-MEM and the World Model on disk | The live guardrail: the sample guardrail answers `safe`, or the verdict a fixture records |

To check that the live services are ready instead, run `go run ./cmd/livecheck` (it never prints a secret).

## Run it

Run Proven from a directory **outside** the repository. Proven keeps its memory in `./state` and its trace in `./logs`,
and it refuses to write its memory inside the repository it reviews.

```bash
git clone https://github.com/saphan-ai/proven.git
cd proven
go build -o /tmp/proven ./cmd/proven
mkdir -p /tmp/proven-run && cd /tmp/proven-run
/tmp/proven scan --sample "$OLDPWD/testdata/fixture/discount"
```

Output (exit code 0):

```text
review class: security
review rule: go.mod
author model not stated; default reviewer NVIDIA
reviewer NVIDIA; role default; family NVIDIA; author model not stated
special-review model: not configured
finding F2 ERROR Dependency leftpad v1.2.3 is on the known-exploited list
  EPSS 0.50000 (90th percentile, 2026-10-20)
  go.mod:5-5
  The change requires evil.example/leftpad v1.2.3, and the sample advisory table has a known-exploited row for that version.
  evidence query local-sample snapshot=sample-2026-10-22
  SQL: SELECT package, version, cve, kev FROM advisories WHERE package = 'evil.example/leftpad' AND version = '1.2.3'
  columns: package, version, cve, kev
  row: evil.example/leftpad | 1.2.3 | CVE-2026-0001 | true
  evidence query local-sample snapshot=sample-2026-10-22
  SQL: SELECT cve_id, epss, percentile, score_date FROM px_evidence.px_epss_cve_rows WHERE cve_id = 'CVE-2026-0001'
  columns: cve_id, epss, percentile, score_date
  row: CVE-2026-0001 | 0.50000 | 0.90000 | 2026-10-20
finding F4 ERROR CVE-2026-0001 for leftpad v1.2.3 is on the KEV list
  EPSS 0.50000 (90th percentile, 2026-10-20)
  go.mod:5-5
  The change requires evil.example/leftpad v1.2.3. The sample KEV view has CVE-2026-0001 for that advisory.
  evidence query local-sample snapshot=sample-2026-10-22
  SQL: SELECT cve_id, vendor_project, product, date_added, known_ransomware FROM px_evidence.px_kev_vulns WHERE cve_id = 'CVE-2026-0001'
  columns: cve_id, vendor_project, product, date_added, known_ransomware
  row: CVE-2026-0001 | evil.example | leftpad | 2026-01-15 | Known
  evidence query local-sample snapshot=sample-2026-10-22
  SQL: SELECT cve_id, epss, percentile, score_date FROM px_evidence.px_epss_cve_rows WHERE cve_id = 'CVE-2026-0001'
  columns: cve_id, epss, percentile, score_date
  row: CVE-2026-0001 | 0.50000 | 0.90000 | 2026-10-20
finding osv:evil.example/leftpad ERROR Dependency evil.example/leftpad v1.2.3 is in an OSV affected range
  EPSS 0.50000 (90th percentile, 2026-10-20)
  go.mod:5-5
  evil.example/leftpad v1.2.3 matches GHSA-px-sample-0001: introduced 1.0.0 <= v1.2.3 < fixed 1.2.4. GHSA-px-sample-0002 does not match: v1.2.3 < introduced 2.0.0. alias CVE-2026-0001 is on the KEV list
  evidence query local-sample snapshot=sample-2026-10-22
  SQL: SELECT id, aliases, ecosystem, package, introduced, fixed, last_affected, severity, published FROM px_osvfile.px_osv_pkg WHERE ecosystem = 'Go' AND package = 'evil.example/leftpad'
  columns: id, aliases, ecosystem, package, introduced, fixed, last_affected, severity, published
  row: GHSA-px-sample-0001 | CVE-2026-0001 | Go | evil.example/leftpad | 1.0.0 | 1.2.4 |  | HIGH | 2026-01-15
  row: GHSA-px-sample-0002 |  | Go | evil.example/leftpad | 2.0.0 | 2.1.0 |  | LOW | 2026-03-01
  evidence query local-sample snapshot=sample-2026-10-22
  SQL: SELECT cve_id, vendor_project, product, date_added, known_ransomware FROM px_evidence.px_kev_vulns WHERE cve_id = 'CVE-2026-0001'
  columns: cve_id, vendor_project, product, date_added, known_ransomware
  row: CVE-2026-0001 | evil.example | leftpad | 2026-01-15 | Known
  evidence query local-sample snapshot=sample-2026-10-22
  SQL: SELECT cve_id, epss, percentile, score_date FROM px_evidence.px_epss_cve_rows WHERE cve_id = 'CVE-2026-0001'
  columns: cve_id, epss, percentile, score_date
  row: CVE-2026-0001 | 0.50000 | 0.90000 | 2026-10-20
finding F1 ERROR Discount is applied twice
  billing/discount.go:5-5
  Discounted subtracts the percent twice, so 100 at 20 percent becomes 60 instead of 80.
  evidence test billing/discount_test.go TestDiscounted fails_on_change=true passes_on_base=true
  command: go test ./billing -run TestDiscounted
  output: base: PASS
change: FAIL got 60 want 80
failed services: none
trace: /tmp/proven-run/logs/trace.jsonl
```

Findings with an EPSS score come first within a severity; F1 has none, so it comes last. A scan writes `logs/trace.jsonl`
in the directory it starts from.

Run from inside the repository, it refuses (exit code 2):

```text
hmem: state directory <repository>/state/hmem is inside the repository
```

## The scenario: `testdata/fixture/discount`

| File | What it holds |
|---|---|
| `base/`, `change/` | Two copies of a tiny Go module: before and after the change |
| `fixture.json` | The recorded model answers (`planner_response`, `responses`) and test outcomes (`recorded_tests`) |
| `tables.sql` | The sample tables the queries run over: `advisories`, the Zetaris-shaped views `px_evidence.px_kev_vulns`, `px_epss_cve_rows`, `px_epss_page_rows` and `px_osv_id`, and the OSV package table `px_osvfile.px_osv_pkg` |

**The change** (`base/` → `change/`):

```diff
-	return cents - cents*percent/100
+	return cents - cents*percent/100 - cents*percent/100
+
+func note() string { return "paid" }
```
and in `go.mod`:
```diff
+require evil.example/leftpad v1.2.3
```

**What comes out, and why:**

| Finding | Verdict | Evidence |
|---|---|---|
| F1 Discount is applied twice | reported | a test: `TestDiscounted` fails on the change and passes on the base |
| F2 leftpad v1.2.3 is known-exploited | reported | a query and its row in `advisories` |
| F3 `note` may include a payment token | **dropped** | none: the Planner named it, nothing proved it; the trace says `dropped: no evidence` |
| F4 CVE-2026-0001 is on the KEV list | reported | a query and its row in `px_evidence.px_kev_vulns` |
| osv:evil.example/leftpad | reported | the OSV package query, both range rows, the comparison `introduced 1.0.0 <= v1.2.3 < fixed 1.2.4`, and the KEV row for the alias `CVE-2026-0001`. The row for `example.com/unused` is not returned |

F3 is the point of the scenario: a plausible risk with no evidence never reaches the review.

## Read the trace

`logs/trace.jsonl` holds one JSON line per event. The run above wrote 62 lines: 5 `claim`, 8 `evidence` (with the
EPSS rows), 4 `refutation`, 5 `verdict`, 13 `model` (one per model call, with its route, model and Nightshift receipt id
when live), 26 `guardrail` (each prompt and each answer checked) and 1 `selection`. F3's last line:

```json
{"finding_id":"F3","kind":"verdict","agent":"judge","text":"dropped: no evidence", …}
```

Run the same command a second time in the same directory: the findings are the same.

## Add a scenario

1. Make `testdata/fixture/<name>/base/` and `change/`, each a small module that builds on its own.
2. Write `tables.sql` with the rows your queries need, in the shapes above.
3. Write `fixture.json`: the Planner's answer (`planner_response`, a list of findings with `id`, `rule_id`, `title`,
   `message`, `severity`, `path`, `start_line`, `end_line`), each agent's recorded answer (`responses`) and each test's
   outcome (`recorded_tests`: `path`, `name`, `command`, `fails_on_change`, `passes_on_base`, `output`). Copy
   `discount/fixture.json` and change it; the Planner's prompt hash (`planner_prompt_sha256`) must match the new trees,
   and a run with a stale hash stops and prints the hash it computed
   (`fixture: planner prompt sha256 is <hash>`); put that hash in the file.
   An optional `guardrail` list records verdicts. A record with `"agent": "planner"` and an empty prompt is bound to
   this fixture's planner prompt. `"user": "unsafe"` and `categories` stop that step and the trace names the categories.
   A fixture with no `guardrail` list answers `safe`.
4. Run `proven scan --sample testdata/fixture/<name>` from outside the repository, as above.

The four scenarios below run the same way: from `/tmp/proven-run`, with the repository in `$OLDPWD` as in "Run it".

## The scenario: `testdata/fixture/refuted`

The Skeptic's recorded answer names `TestTotal`. The recorded outcome passes on the change, so F1 is dropped. The trace
verdict is `dropped: refuted: test TestTotal passes on the change`. F1 is not printed. The query row would have kept
the finding: with the refutation's path, name, command and sql emptied, the same test prints F1 and fails.

```bash
/tmp/proven scan --sample "$OLDPWD/testdata/fixture/refuted"
```

Output (exit code 0):

```text
review class: security
review rule: go.mod
author model not stated; default reviewer NVIDIA
reviewer NVIDIA; role default; family NVIDIA; author model not stated
special-review model: not configured
failed services: none
trace: /tmp/proven-run/logs/trace.jsonl
```

The verdict line in that trace:

```json
{"finding_id":"F1","kind":"verdict","agent":"judge","text":"dropped: refuted: test TestTotal passes on the change"}
```

## The scenario: `testdata/fixture/unsafe-prompt`

The fixture records one guardrail verdict: the planner prompt is `unsafe`, categories `Criminal Planning/Confessions`
and `Other`. That step stops. No finding is printed.

```bash
/tmp/proven scan --sample "$OLDPWD/testdata/fixture/unsafe-prompt"
```

Output (exit code 0):

```text
review class: security
review rule: go.mod
author model not stated; default reviewer NVIDIA
reviewer NVIDIA; role default; family NVIDIA; author model not stated
special-review model: not configured
failed service: guardrail: unsafe: Criminal Planning/Confessions, Other
trace: /tmp/proven-run/logs/trace.jsonl
```

The guardrail line in that trace:

```json
{"finding_id":"guardrail","kind":"guardrail","agent":"planner","text":"prompt unsafe latency=0ms categories=Criminal Planning/Confessions, Other","guardrail":{"side":"prompt","verdict":"unsafe","categories":["Criminal Planning/Confessions","Other"],"latency_ms":0}}
```

## The scenario: `testdata/fixture/osv-go-module`

`go.mod` raises `example.com/widget` from v1.2.3 to v1.4.0. v1.4.0 is inside one sample OSV range and outside another.
The alias `CVE-2026-4242` has an EPSS row. The row for `example.com/other` is not returned.

```bash
/tmp/proven scan --sample "$OLDPWD/testdata/fixture/osv-go-module"
```

Output (exit code 0):

```text
review class: security
review rule: go.mod
author model not stated; default reviewer NVIDIA
reviewer NVIDIA; role default; family NVIDIA; author model not stated
special-review model: not configured
finding osv:example.com/widget ERROR Dependency example.com/widget v1.4.0 is in an OSV affected range
  EPSS 0.42000 (91st percentile, 2026-10-21)
  go.mod:5-5
  example.com/widget v1.4.0 matches GHSA-w19-sample-0001: introduced 1.3.0 <= v1.4.0 < fixed 1.5.0. GHSA-w19-sample-0002 does not match: v1.4.0 < introduced 2.0.0. alias CVE-2026-4242 is on the KEV list
  evidence query local-sample snapshot=sample-2026-10-23
  SQL: SELECT id, aliases, ecosystem, package, introduced, fixed, last_affected, severity, published FROM px_osvfile.px_osv_pkg WHERE ecosystem = 'Go' AND package = 'example.com/widget'
  columns: id, aliases, ecosystem, package, introduced, fixed, last_affected, severity, published
  row: GHSA-w19-sample-0001 | CVE-2026-4242 | Go | example.com/widget | 1.3.0 | 1.5.0 |  | HIGH | 2026-04-01
  row: GHSA-w19-sample-0002 |  | Go | example.com/widget | 2.0.0 | 2.1.0 |  | LOW | 2026-05-01
  evidence query local-sample snapshot=sample-2026-10-23
  SQL: SELECT cve_id, vendor_project, product, date_added, known_ransomware FROM px_evidence.px_kev_vulns WHERE cve_id = 'CVE-2026-4242'
  columns: cve_id, vendor_project, product, date_added, known_ransomware
  row: CVE-2026-4242 | example.com | widget | 2026-04-01 | Unknown
  evidence query local-sample snapshot=sample-2026-10-23
  SQL: SELECT cve_id, epss, percentile, score_date FROM px_evidence.px_epss_cve_rows WHERE cve_id = 'CVE-2026-4242'
  columns: cve_id, epss, percentile, score_date
  row: CVE-2026-4242 | 0.42000 | 0.91000 | 2026-10-21
failed services: none
trace: /tmp/proven-run/logs/trace.jsonl
```

## The scenario: `testdata/fixture/unknown-author`

Sample mode does not read commit history on this command: the scan sets no baseline commit, and the fixture is two
trees, not a git history. The sample request's description is empty, and an empty description names no author model.
The review says which default it used.

```bash
/tmp/proven scan --sample "$OLDPWD/testdata/fixture/unknown-author"
```

Output (exit code 0):

```text
review class: security
review rule: go.mod
author model not stated; default reviewer NVIDIA
reviewer NVIDIA; role default; family NVIDIA; author model not stated
special-review model: not configured
failed services: none
trace: /tmp/proven-run/logs/trace.jsonl
```
