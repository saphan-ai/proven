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
| The trace (`<user cache>/proven/logs/trace.jsonl`), H-MEM and the World Model on disk | The live guardrail: the sample guardrail answers `safe`, or the verdict a fixture records |

To check that the live services are ready instead, run `go run ./cmd/livecheck` (it never prints a secret).

## Run it

The trace is written to the user cache, in `proven/logs`, created so only that user can read it. `PROVEN_LOG_DIR`
overrides that directory. A path that resolves inside the repository under review is refused when Proven loads.

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
  SQL: SELECT id, aliases, ecosystem, package, introduced, fixed, last_affected, severity, published FROM px_osvfile.px_osv_pkg WHERE ecosystem = 'Go' AND package = 'evil.example/leftpad'
  columns: id, aliases, ecosystem, package, introduced, fixed, last_affected, severity, published
  row: GHSA-px-sample-0001 | CVE-2026-0001 | Go | evil.example/leftpad | 1.0.0 | 1.2.4 |  | HIGH | 2026-01-15
  row: GHSA-px-sample-0002 |  | Go | evil.example/leftpad | 2.0.0 | 2.1.0 |  | LOW | 2026-03-01
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
  command: go test -run '^TestDiscounted$' ./billing
  output: base: PASS
change: FAIL got 60 want 80
Agreed by two models, not proven
  none: one reviewer; a claim only one model makes stays dropped
failed services: none
trace: /tmp/proven-run/logs/trace.jsonl
```

Findings with an EPSS score come first within a severity; F1 has none, so it comes last. A scan writes `trace.jsonl`
in the user cache, under `proven/logs`, unless `PROVEN_LOG_DIR` names another directory. A path that resolves inside
the repository under review is refused at load (exit 7). The `trace:` lines in the recorded output on this page are
from the run that wrote the page.

## The scenario: `testdata/fixture/discount`

| File | What it holds |
|---|---|
| `base/`, `change/` | Two copies of a tiny Go module: before and after the change |
| `fixture.json` | The recorded model answers (`planner_response`, `responses`) and test outcomes (`recorded_tests`) |
| `tables.sql` | The sample tables the queries run over: `advisories` (a plain table that no query template reads; it stays as the local engine's example of a table that is not a Zetaris view), the Zetaris-shaped views `px_evidence.px_kev_vulns`, `px_epss_cve_rows`, `px_epss_page_rows` and `px_osv_id`, and the OSV package table `px_osvfile.px_osv_pkg` |

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

| Finding | Verdict | Level of proof (ADR-0044) | Evidence |
|---|---|---|---|
| F1 Discount is applied twice | reported | proved by a test | a test: `TestDiscounted` fails on the change and passes on the base |
| F2 leftpad v1.2.3 is known-exploited | reported | proved by data | a query and its rows in `px_osvfile.px_osv_pkg` |
| F3 `note` may include a payment token | **dropped** | none: dropped | none: the Planner named it, nothing proved it; the trace says `dropped: no evidence` |
| F4 CVE-2026-0001 is on the KEV list | reported | proved by data | a query and its row in `px_evidence.px_kev_vulns` |
| osv:evil.example/leftpad | reported | proved by data | the OSV package query, both range rows, the comparison `introduced 1.0.0 <= v1.2.3 < fixed 1.2.4`, and the KEV row for the alias `CVE-2026-0001`. The row for `example.com/unused` is not returned |

Totals by level on this scan: 1 proved by a test, 0 proved by a tool, 3 proved by data, 0 agreed by two models and not proven.

F3 is the point of the scenario: a plausible risk with no evidence never reaches the review. This scan has one reviewer, so the section "Agreed by two models, not proven" is empty and says why. An agreed opinion needs both reviewers to name it at the same place, and it is not a finding and not one of these totals (ADR-0041).

## Read the trace

`<user cache>/proven/logs/trace.jsonl` holds one JSON line per event. The run above wrote 62 lines: 5 `claim`, 8 `evidence` (with the
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
   outcome (`recorded_tests`: `path`, `name`, `command`, `fails_on_change`, `passes_on_base`, `output`). A recorded
   answer names a test by `path` and `name` only: an answer that carries a `command` is refused. The `command` of a
   recorded test is the line Proven builds from them, `go test -json -run '^<name>$' ./<directory of path>`. Copy
   `discount/fixture.json` and change it; the Planner's prompt hash (`planner_prompt_sha256`) must match the new trees,
   and a run with a stale hash stops and prints the hash it computed
   (`fixture: planner prompt sha256 is <hash>`); put that hash in the file.
   An optional `guardrail` list records verdicts. A record with `"agent": "planner"` and an empty prompt is bound to
   this fixture's planner prompt. `"user": "unsafe"` and `categories` stop that step and the trace names the categories.
   A fixture with no `guardrail` list answers `safe`.
   The Investigator's and the Skeptic's recorded answers name a query template and its values, never SQL
   ([QUERIES.md](QUERIES.md)): `{"template":"kev-by-cve","params":{"cve_id":"CVE-2026-0001"}}`. An empty `template`
   means no query. A recorded answer is found by its agent and the exact prompt text the code builds, so after a change
   to a prompt the recorded prompts are regenerated. In every prompt the change's text and each model-written title
   sit in one data section after the line `DATA:`, written as a single line of JSON, so no byte of the change can start
   a new line, field or file record of the prompt itself. A query's rows count as evidence only when Proven ties them
   to a package or an advisory the change is affected by; a recorded query for anything else keeps its rows on the
   finding as context and the finding is dropped for lack of evidence.
   A fixture that shows a fix (ADR-0042) adds two things. The Fixer's recorded answers (`responses`, agent `fixer`)
   are a patch as a list of edits, `{"what":"…","why":"…","edits":[{"path":"…","old":"…","new":"…"}]}`; a retry's
   prompt carries the failed attempt's record, so its recorded prompt is regenerated from a run. And `recorded_fixes`
   lists the runs of each attempt (`finding_id`, `attempt` from 1, and `runs`, each with `label`, `command`, `result`,
   `exit_code` and `output`). Sample mode still checks every recorded patch against the real trees: an edit whose old text
   is not in the file once is refused as it would be live. A finding with no recorded Fixer answer shows no fix.
4. Run `proven scan --sample testdata/fixture/<name>` from outside the repository, as above.

The scenarios below run the same way: from `/tmp/proven-run`, with the repository in `$OLDPWD` as in "Run it".

## The scenario: `testdata/fixture/refuted`

The Skeptic's recorded answer names `TestTotal`. The recorded outcome passes on the change, so F1 is dropped. The trace
verdict is `dropped: refuted: test TestTotal passes on the change`. F1 is not printed. The query row (the Investigator's
`osv-by-advisory` template, built into SQL by Proven) would have kept the finding: with the refutation's path and name
emptied, the same test prints F1 and fails.

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
  fix named from the advisory table, not run by a test: raise example.com/widget v1.4.0 to 1.5.0
  fix closes the problem because: 1.5.0 is the fixed version of advisory GHSA-w19-sample-0001, and it is the smallest fixed version of the rows returned that no affected range of those rows holds. No test proves a version change, so this fix is named from the table, not run.
  fix advisory GHSA-w19-sample-0001 columns: id, aliases, ecosystem, package, introduced, fixed, last_affected, severity, published
  fix advisory row: GHSA-w19-sample-0001 | CVE-2026-4242 | Go | example.com/widget | 1.3.0 | 1.5.0 |  | HIGH | 2026-04-01
failed services: none
trace: /tmp/proven-run/logs/trace.jsonl
```

The last four lines are the dependency fix (ADR-0042 point 5): the smallest fixed version, taken from the advisory rows
the query returned, with the row that shows it. Proven compares versions in Go, never in SQL, and no model names the
version. When no row names a fixed version (an advisory with only a last-affected bound), the review says so instead of
guessing the next version.

## The scenario: `testdata/fixture/fix`

Two findings, each proved by a test that fails on the change and passes on the base. F1 is `Discounted` taking the
percent off twice. F2 is `shipping.Fee` returning 0 for every order. The recorded model answers a patch for each, and the
recorded runs are what a container run printed: the fixture's `recorded_fixes` were made by running the real
`go test` commands on copies of the trees (a container run is the one thing sample mode does not do).

```bash
/tmp/proven scan --sample "$OLDPWD/testdata/fixture/fix"
```

Output for F1, whose patch is proved on the first attempt (exit code 0):

```text
finding F1 ERROR Discount is applied twice
  billing/discount.go:5-5
  Discounted subtracts the percent twice, so 100 at 20 percent becomes 60 instead of 80.
  rule discount-applied-twice
  evidence test billing/discount_test.go TestDiscounted fails_on_change=true passes_on_base=true
  command: go test -json -run '^TestDiscounted$' ./billing
  output: base: PASS
change: FAIL got 60 want 80
  fix proved after 1 attempt: a suggestion for you to apply; Proven changes nothing
  fix changes: Subtract the discount once.
  fix closes the problem because: Discounted took the percent off twice, so 100 cents at 20 percent came to 60. Taking it off once gives 80.
  fix edit billing/discount.go
  - return cents - cents*percent/100 - cents*percent/100
  + return cents - cents*percent/100
  fix run base: go test -json -run '^TestDiscounted$' ./billing exit=0 passed
  fix run change: go test -json -run '^TestDiscounted$' ./billing exit=1 failed
  fix run change+patch: go test -json -run '^TestDiscounted$' ./billing exit=0 passed
  fix run package on change+patch: go test -json ./billing exit=0 passed
```

The four `fix run` lines are the validation steps: each has the command, its exit code and how it ended, so a reader can
run them again. The patch is a suggestion for a person to apply; nothing in the review is written to the change.

Output for F2, whose first patch (`return 50`) fails the proving test, and whose retry (`return 500`) makes it pass and
breaks `TestFreeShipping` in the same package. The model got the failed run's output with its retry:

```text
finding F2 ERROR Shipping is always free
  shipping/shipping.go:5-5
  Fee returns 0 for every order, so an order of 1000 cents pays no shipping instead of 500.
  rule shipping-fee-dropped
  evidence test shipping/shipping_test.go TestFee fails_on_change=true passes_on_base=true
  command: go test -json -run '^TestFee$' ./shipping
  output: base: PASS
change: FAIL got 0 want 500
  fix not proved after 2 attempts: the patch breaks another test of the package
  fix run base: go test -json -run '^TestFee$' ./shipping exit=0 passed
  fix run change: go test -json -run '^TestFee$' ./shipping exit=1 failed
  fix run change+patch: go test -json -run '^TestFee$' ./shipping exit=0 passed
  fix run package on change+patch: go test -json ./shipping exit=1 failed
```

A run counts as passed only when its output shows the named test ran and passed (ADR-0042). Every proof run is
`go test -json`, and a run that exits 0 without a pass event for the test is recorded as not run. The sample applies the
same rule to a recorded run: its output must hold a `--- PASS:` line for the test, and a package run an `ok` line too.

A patch that is not proved is not shown: there is no `fix edit` for F2, only the record of the last attempt. Two settings
in `proven.yaml` change this ([CONFIGURATION.md](CONFIGURATION.md)): `fixes: off` shows no fix at all, and
`fix_attempts: 1` ends F2 after its first patch.

## The scenarios: `testdata/fixture/autofix-fixed` and `testdata/fixture/autofix-rejected`

Both are the `fix` change (F1: the discount is taken off twice; F2: shipping is always free) with the whole-change
attempt of ADR-0050 recorded. Neither prints the attempt in the review text: the attempt is on the `Result`, and the
review comment that shows it is the shell's. Turn it on with a `proven.yaml` that holds `autofix: on`, then read the
trace, where each verdict and each re-run is a line of `kind autofix`:

```sh
printf 'autofix: on\n' > autofix.yaml
proven scan --sample --config autofix.yaml testdata/fixture/autofix-fixed
```

- `autofix-fixed` records one patch for both findings: subtract the discount once, and charge 500 cents below 5000. Each
  finding's test, run again on the patched change, passes, and so do the tests of its package. The verdict is
  `fixed automatically, waiting for a person`, and the patch is the diff of `billing/discount.go` and
  `shipping/shipping.go`.
- `autofix-rejected` records a patch that returns 500 for every order. F1 closes. F2's own test passes, but the package's
  `TestFreeShipping` fails (the trace says `test proof open: go test -json ./shipping: the patch breaks another test of
  the package`), so F2 stays open and the verdict is `rejected`. The patch is kept in the record and is not a suggestion.

Both fixtures hold one recorded answer for the `autofix` agent and one recorded run of the proof per finding. The sample
applies the patch's edits to the real trees, so a stale recording whose edits no longer apply is refused as it would be
live. Without `autofix: on`, the same fixtures make no attempt, and the review text is the review text of `fix`.

## The scenario: `testdata/fixture/osv-php-composer`

`composer.json` and `composer.lock` raise `monolog/monolog` from 2.0.0 to 2.9.1. 2.9.1 is inside one sample OSV range and outside another. The lock file's version is the one reported. The alias `CVE-2026-9090` has an EPSS row. The row for `other/unused` is not returned.

The lines below are the stdout of `scan --sample testdata/fixture/osv-php-composer`, captured on 24 October 2026 with Go 1.26.10 by the scenario test, which runs that command. The trace path is the directory the scan starts from, so it is not copied here. The fix lines are in this capture because the fix chooser uses the language adapter's version order.

```text
review class: security
review rule: composer.json
author model not stated; default reviewer NVIDIA
reviewer NVIDIA; role default; family NVIDIA; author model not stated
special-review model: not configured
finding osv:monolog/monolog ERROR Dependency monolog/monolog 2.9.1 is in an OSV affected range
  EPSS 0.25000 (91st percentile, 2026-10-24)
  composer.lock:5-5
  monolog/monolog 2.9.1 matches GHSA-php-sample-0001: introduced 2.4.0 <= 2.9.1 < fixed 2.10.0. GHSA-php-sample-0002 does not match: 2.9.1 < introduced 3.0.0. alias CVE-2026-9090 is on the KEV list
  rule osv-affected-package
  evidence query local-sample snapshot=sample-2026-10-24
  SQL: SELECT id, aliases, ecosystem, package, introduced, fixed, last_affected, severity, published FROM px_osvfile.px_osv_pkg WHERE ecosystem = 'Packagist' AND package = 'monolog/monolog'
  columns: id, aliases, ecosystem, package, introduced, fixed, last_affected, severity, published
  row: GHSA-php-sample-0001 | CVE-2026-9090 | Packagist | monolog/monolog | 2.4.0 | 2.10.0 |  | HIGH | 2026-10-01
  row: GHSA-php-sample-0002 |  | Packagist | monolog/monolog | 3.0.0 | 3.1.0 |  | LOW | 2026-10-02
  evidence query local-sample snapshot=sample-2026-10-24
  SQL: SELECT cve_id, vendor_project, product, date_added, known_ransomware FROM px_evidence.px_kev_vulns WHERE cve_id = 'CVE-2026-9090'
  columns: cve_id, vendor_project, product, date_added, known_ransomware
  row: CVE-2026-9090 | example.com | monolog | 2026-10-24 | Unknown
  evidence query local-sample snapshot=sample-2026-10-24
  SQL: SELECT cve_id, epss, percentile, score_date FROM px_evidence.px_epss_cve_rows WHERE cve_id = 'CVE-2026-9090'
  columns: cve_id, epss, percentile, score_date
  row: CVE-2026-9090 | 0.25000 | 0.91000 | 2026-10-24
  fix named from the advisory table, not run by a test: raise monolog/monolog 2.9.1 to 2.10.0
  fix closes the problem because: 2.10.0 is the fixed version of advisory GHSA-php-sample-0001, and it is the smallest fixed version of the rows returned that no affected range of those rows holds. No test proves a version change, so this fix is named from the table, not run.
  fix advisory GHSA-php-sample-0001 columns: id, aliases, ecosystem, package, introduced, fixed, last_affected, severity, published
  fix advisory row: GHSA-php-sample-0001 | CVE-2026-9090 | Packagist | monolog/monolog | 2.4.0 | 2.10.0 |  | HIGH | 2026-10-01
failed services: none
```

## The scenario: `testdata/fixture/osv-rust-cargo`

`Cargo.toml` and `Cargo.lock` raise `regex` from 1.5.0 to 1.9.1. The manifest writes the constraint `^1.9`. The lock file's version is the one reported. 1.9.1 is inside one sample OSV range and outside another. `libc` goes from 0.2.150 to 0.2.140 and is not a finding. The alias `CVE-2026-9191` has an EPSS row. The row for `unused_crate` is not returned. 1.10.0 is the smallest fixed version that leaves every returned range.

The lines below are the stdout of `scan --sample testdata/fixture/osv-rust-cargo`, captured on 24 October 2026 with Go 1.26.10 by the scenario test, which runs that command. The trace path is the directory the scan starts from, so it is not copied here.

```text
review class: security
review rule: Cargo.toml
author model not stated; default reviewer NVIDIA
reviewer NVIDIA; role default; family NVIDIA; author model not stated
special-review model: not configured
finding osv:regex ERROR Dependency regex 1.9.1 is in an OSV affected range
  EPSS 0.33000 (91st percentile, 2026-10-24)
  Cargo.lock:15-15
  regex 1.9.1 matches GHSA-rs-sample-0001: introduced 1.7.0 <= 1.9.1 < fixed 1.10.0. GHSA-rs-sample-0002 does not match: 1.9.1 < introduced 1.11.0. alias CVE-2026-9191 is on the KEV list
  rule osv-affected-package
  evidence query local-sample snapshot=sample-2026-10-24
  SQL: SELECT id, aliases, ecosystem, package, introduced, fixed, last_affected, severity, published FROM px_osvfile.px_osv_pkg WHERE ecosystem = 'crates.io' AND package = 'regex'
  columns: id, aliases, ecosystem, package, introduced, fixed, last_affected, severity, published
  row: GHSA-rs-sample-0001 | CVE-2026-9191 | crates.io | regex | 1.7.0 | 1.10.0 |  | HIGH | 2026-10-01
  row: GHSA-rs-sample-0002 |  | crates.io | regex | 1.11.0 | 1.12.0 |  | LOW | 2026-10-02
  evidence query local-sample snapshot=sample-2026-10-24
  SQL: SELECT cve_id, vendor_project, product, date_added, known_ransomware FROM px_evidence.px_kev_vulns WHERE cve_id = 'CVE-2026-9191'
  columns: cve_id, vendor_project, product, date_added, known_ransomware
  row: CVE-2026-9191 | example.com | regex | 2026-10-24 | Unknown
  evidence query local-sample snapshot=sample-2026-10-24
  SQL: SELECT cve_id, epss, percentile, score_date FROM px_evidence.px_epss_cve_rows WHERE cve_id = 'CVE-2026-9191'
  columns: cve_id, epss, percentile, score_date
  row: CVE-2026-9191 | 0.33000 | 0.91000 | 2026-10-24
  fix named from the advisory table, not run by a test: raise regex 1.9.1 to 1.10.0
  fix closes the problem because: 1.10.0 is the fixed version of advisory GHSA-rs-sample-0001, and it is the smallest fixed version of the rows returned that no affected range of those rows holds. No test proves a version change, so this fix is named from the table, not run.
  fix advisory GHSA-rs-sample-0001 columns: id, aliases, ecosystem, package, introduced, fixed, last_affected, severity, published
  fix advisory row: GHSA-rs-sample-0001 | CVE-2026-9191 | crates.io | regex | 1.7.0 | 1.10.0 |  | HIGH | 2026-10-01
failed services: none
```

## The scenario: `testdata/fixture/osv-kotlin-gradle`

`build.gradle.kts` and `gradle.lockfile` raise `com.example:widget` from 1.2.0 to 1.4.0. The catalog `gradle/libs.versions.toml` leaves the library versions as they were and changes a plugin version, which is a note. 1.4.0 is inside one sample OSV range and outside another. The lock file's version is the one reported. `com.example:quiet` goes from 2.0.0 to 1.9.0 and is not a finding. The alias `CVE-2026-9291` has an EPSS row. The row for `com.example:unused` is not returned. 1.5.0 is the smallest fixed version that leaves every returned range.

The lines below are the stdout of `scan --sample testdata/fixture/osv-kotlin-gradle`, captured on 25 October 2026 with Go 1.26.10 by the scenario test, which runs that command. The trace path is the directory the scan starts from, so it is not copied here. The fix lines are in this capture because the fix chooser uses the language adapter's version order.

```text
review class: security
review rule: build.gradle.kts
author model not stated; default reviewer NVIDIA
reviewer NVIDIA; role default; family NVIDIA; author model not stated
special-review model: not configured
finding osv:com.example:widget ERROR Dependency com.example:widget 1.4.0 is in an OSV affected range
  EPSS 0.27000 (91st percentile, 2026-10-25)
  gradle.lockfile:5-5
  com.example:widget 1.4.0 matches GHSA-kt-sample-0001: introduced 1.3.0 <= 1.4.0 < fixed 1.5.0. GHSA-kt-sample-0002 does not match: 1.4.0 < introduced 2.0.0. alias CVE-2026-9291 is on the KEV list Reachability: not measured.
  rule osv-affected-package
  evidence query local-sample snapshot=sample-2026-10-25 form=sql
  SQL: SELECT id, aliases, ecosystem, package, introduced, fixed, last_affected, severity, published FROM px_osvfile.px_osv_pkg WHERE ecosystem = 'Maven' AND package = 'com.example:widget'
  columns: id, aliases, ecosystem, package, introduced, fixed, last_affected, severity, published
  row: GHSA-kt-sample-0001 | CVE-2026-9291 | Maven | com.example:widget | 1.3.0 | 1.5.0 |  | HIGH | 2026-10-01
  row: GHSA-kt-sample-0002 |  | Maven | com.example:widget | 2.0.0 | 2.1.0 |  | LOW | 2026-10-02
  evidence query local-sample snapshot=sample-2026-10-25 form=sql
  SQL: SELECT cve_id, vendor_project, product, date_added, known_ransomware FROM px_evidence.px_kev_vulns WHERE cve_id = 'CVE-2026-9291'
  columns: cve_id, vendor_project, product, date_added, known_ransomware
  row: CVE-2026-9291 | example.com | widget | 2026-10-25 | Unknown
  evidence query local-sample snapshot=sample-2026-10-25 form=sql
  SQL: SELECT cve_id, epss, percentile, score_date FROM px_evidence.px_epss_cve_rows WHERE cve_id = 'CVE-2026-9291'
  columns: cve_id, epss, percentile, score_date
  row: CVE-2026-9291 | 0.27000 | 0.91000 | 2026-10-25
  fix named from the advisory table, not run by a test: raise com.example:widget 1.4.0 to 1.5.0
  fix closes the problem because: 1.5.0 is the fixed version of advisory GHSA-kt-sample-0001, and it is the smallest fixed version of the rows returned that no affected range of those rows holds. No test proves a version change, so this fix is named from the table, not run.
  fix advisory GHSA-kt-sample-0001 columns: id, aliases, ecosystem, package, introduced, fixed, last_affected, severity, published
  fix advisory row: GHSA-kt-sample-0001 | CVE-2026-9291 | Maven | com.example:widget | 1.3.0 | 1.5.0 |  | HIGH | 2026-10-01
Agreed by two models, not proven
  none: one reviewer; a claim only one model makes stays dropped
failed services: none
```

## The scenario: `testdata/fixture/discount-clean`

The correct version of the change `discount` botches: `Discounted` is written with a named `off` and the discount is
rounded down to a whole cent, and a second test says so. The Planner's recorded answer is an empty list, so no query
runs and no test is run, and the review keeps no finding. The judges' lab reviews both and shows them side by side:
[LAB.md](LAB.md#the-demo-pull-requests). A clean review prints no `finding` line; the lab's workflow adds one line
above it that says so.

```bash
/tmp/proven scan --sample "$OLDPWD/testdata/fixture/discount-clean"
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
