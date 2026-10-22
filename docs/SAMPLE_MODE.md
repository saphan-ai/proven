# Sample mode

Sample mode runs a whole Proven review with **no network and no keys**. Every model answer, every test outcome and every
query row comes from a recorded fixture, so anyone can run it on a laptop and get the same result every time.

Every command and every output on this page was run on 22 October 2026 on a clean clone of `main` at `39225b0`, with
Go 1.26.8 on macOS. Long output is shortened where marked.

## What sample mode exercises, and what it does not

| Exercised (the real code runs) | Not exercised (replaced by recorded data) |
|---|---|
| The six agents in order: Planner, Investigator, Prover, Skeptic, Judge, Writer | Live model calls through Nightshift |
| The Judge's rule: a finding without evidence is dropped | The Prover running a test in a container: the outcome is recorded |
| The Skeptic's refutation step and the Judge's verdict | Live Zetaris queries: the same SQL runs over local sample tables |
| The trace (`logs/trace.jsonl`), H-MEM and the World Model on disk | The live guardrail: a sample guardrail answers `safe` |

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
finding F1 ERROR Discount is applied twice
  billing/discount.go:5-5
  Discounted subtracts the percent twice, so 100 at 20 percent becomes 60 instead of 80.
  evidence test billing/discount_test.go TestDiscounted fails_on_change=true passes_on_base=true
  command: go test ./billing -run TestDiscounted
  output: base: PASS
change: FAIL got 60 want 80
finding F2 ERROR Dependency leftpad v1.2.3 is on the known-exploited list
  go.mod:5-5
  The change requires evil.example/leftpad v1.2.3, and the sample advisory table has a known-exploited row for that version.
  evidence query local-sample snapshot=sample-2026-10-22
  SQL: SELECT package, version, cve, kev FROM advisories WHERE package = 'evil.example/leftpad' AND version = '1.2.3'
  columns: package, version, cve, kev
  row: evil.example/leftpad | 1.2.3 | CVE-2026-0001 | true
finding F4 ERROR CVE-2026-0001 for leftpad v1.2.3 is on the KEV list
  go.mod:5-5
  (shortened)
  SQL: SELECT cve_id, vendor_project, product, date_added, known_ransomware FROM px_evidence.px_kev_vulns WHERE cve_id = 'CVE-2026-0001'
  row: CVE-2026-0001 | evil.example | leftpad | 2026-01-15 | Known
failed services: none
trace: /tmp/proven-run/logs/trace.jsonl
```

Run from inside the repository, it refuses (exit code 2):

```text
hmem: state directory <repository>/state/hmem is inside the repository
```

## The scenario: `testdata/fixture/discount`

| File | What it holds |
|---|---|
| `base/`, `change/` | Two copies of a tiny Go module: before and after the change |
| `fixture.json` | The recorded model answers (`planner_response`, `responses`) and test outcomes (`recorded_tests`) |
| `tables.sql` | The sample tables the queries run over: `advisories`, and the Zetaris-shaped views `px_evidence.px_kev_vulns`, `px_epss_cve_rows`, `px_epss_page_rows` and `px_osv_id` |

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

F3 is the point of the scenario: a plausible risk with no evidence never reaches the review.

## Read the trace

`logs/trace.jsonl` holds one JSON line per event. The run above wrote 56 lines: 4 `claim`, 4 `evidence`,
4 `refutation`, 4 `verdict`, 13 `model` (one per model call, with its route, model and Nightshift receipt id when live),
26 `guardrail` (each prompt and each answer checked) and 1 `selection`. F3's last line:

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
4. Run `proven scan --sample testdata/fixture/<name>` from outside the repository, as above.
