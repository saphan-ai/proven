# The semantic layer in Zetaris

Proven's evidence lives in four places in Zetaris: the OSV package table, the OSV advisory-id view, the KEV view and the
EPSS view ([ADR-0023](adr/ADR-0023-how-proven-s-evidence-is-loaded-through-zetaris.md), [OSV_TABLE.md](OSV_TABLE.md)).
The layer puts one container, `px_usl`, in front of them. It names four entities and the relations between them, and one
data-quality rule decides whether Proven may be pointed at it. Proven's templates do not change: the layer's view names
are given through the name settings of [QUERIES.md](QUERIES.md).

## What this layer is, and what it is not

The repository carries no Zetaris documentation of a "Unified Semantic Layer" object. So this layer is plain SchemaStore
views in one container, in the form the registration runbook already uses (`CREATE SCHEMASTORE CONTAINER`,
`CREATE SCHEMASTORE VIEW ... WITH CONTAINER ... AS SELECT`). The relations are join keys, written in one view. The
data-quality rule is a query in its own file.

A real USL object would add three things that plain views cannot: a relationship that a client can list and follow, a
description on each column, and the rule bound to the entity instead of kept in `gate.sql`. None of that was measured on
our build, so none of it is used.

## The entities and the relations

| Entity | View | Source | Key | Template that reads it | Setting |
|---|---|---|---|---|---|
| package (one affected range of one advisory) | `px_usl.px_package` | `px_osvfile.px_osv_pkg` | `ecosystem`, `package` | `osv-by-package` | `PROVEN_OSV_TABLE` |
| OSV advisory | `px_usl.px_advisory` | `px_evidence.px_osv_id` | `id` | `osv-by-advisory` | `PROVEN_OSV_ID_VIEW` |
| KEV CVE | `px_usl.px_kev_cve` | `px_evidence.px_kev_vulns` | `cve_id` | `kev-by-cve` | `PROVEN_KEV_VIEW` |
| EPSS score | `px_usl.px_epss_score` | `px_evidence.px_epss_cve_rows` | `cve_id` | `epss-by-cve` | `PROVEN_EPSS_VIEW` |

Each entity view keeps the column names of its source, so the template SQL is the same text with another name in `FROM`.

The relations are in `px_usl.px_package_evidence`. Its grain is described under "The grain of the relation view" below:

| Relation | Join key |
|---|---|
| a package has advisories | `px_package.id = px_advisory.id` |
| an advisory names CVEs | `px_advisory.alias` is the CVE id |
| a CVE may be in KEV | CVE id `= px_kev_cve.cve_id` |
| a CVE has an EPSS score | CVE id `= px_epss_score.cve_id` |

Every join is a `LEFT JOIN`. A join is an inner join by default, and an inner join drops a CVE that has no KEV row, which
reads as "not exploited". Proven does not read this view; it is for people and tools that ask the layer a question of
their own.

### The grain of the relation view

One row is one package range, one alias of its advisory and one EPSS score date. The view does not collapse aliases or
score dates; it shows each. So a range whose advisory has three aliases is three rows, and a CVE with two score dates
doubles the rows of its alias. A range with no advisory row, an alias that is not a CVE (`GHSA-…`, `GO-…`) and a CVE with
no score are each one row with empty columns. The view adds `epss_score_date`, so a reader can tell the rows of one CVE
apart. To count package ranges, count distinct `(ecosystem, package, advisory_id, introduced, fixed)`; to read the latest
score, filter on the greatest `epss_score_date` of the CVE. The grain assumes `px_kev_cve` has one row per CVE: a
repeated KEV row repeats the rows of that CVE.

`TestRelationViewGrainIsPerAliasPerScoreDate` checks this on a fixture with two aliases and two score dates. It runs the
view's own SELECT in `python3`'s sqlite3, not in Zetaris, and is skipped without `python3`.

The other four templates (`mal-by-package`, `scorecard-by-package`, `cvss-by-cve`, `spdx-by-package`) are not served.
Their table names are the source adapters' own and are not settings ([QUERIES.md](QUERIES.md)), and the order for this
layer named four entities.

The definitions are in [`lab/zetaris/usl/layer.sql`](../lab/zetaris/usl/layer.sql). Each view is commented with its source and key.

## The gate

[`lab/zetaris/usl/gate.sql`](../lab/zetaris/usl/gate.sql) is rule `usl-epss-score-valid`. It is a query that returns the violating rows of
`px_usl.px_epss_score`. A row violates it when

- its `cve_id` is empty or is not `CVE-`, four digits, a hyphen and four to twelve digits (the format of [QUERIES.md](QUERIES.md)), or
- its `epss` is not a number from 0 to 1, or
- its `percentile` is not a number from 0 to 1.

`apply.sh` runs it after the views exist. While it returns any row, the script exits non-zero, names the rule and the count,
prints the first rows, and does not print the settings. Proven is not pointed at the layer.

**Why this rule.** EPSS is the one numeric source a finding quotes as evidence, and nothing checks it on the way in. Proven
reads `epss` with `strconv.ParseFloat` and keeps the highest value on a finding (`epss.Best`), with no upper bound, so a
score written as a percent (`42`) would outrank every real score on that finding. The percentile reader takes a
value above 1 as a percent already, so a bad percentile prints as a plausible rank. The `cve_id` is also the join key to
KEV, and an inner join drops a malformed id without a word. The KEV view has no number to check, and its `cve_id` is
already refused by the template's format when Proven asks.

The gate answers one question: no row means the EPSS view held no bad row at the moment of the query. The EPSS view is a
live REST table ([ADR-0023](adr/ADR-0023-how-proven-s-evidence-is-loaded-through-zetaris.md), point 4), so the answer says
nothing about a row the source publishes later. A count that cannot be read is a refusal, not a pass.

## Apply it

The registration is the administrator's act: Proven's own key is refused for it (ADR-0023, point 2). The script reads the
Zetaris address, the numeric organisation id and the key from the environment, with the names
[CONFIGURATION.md](CONFIGURATION.md) lists for Zetaris. It takes no argument and no file. It prints none of the three:
the key and the organisation id are removed from every refusal and every transport error before it is shown (`[key]`, `[org]`).
It hands the key to `curl` on a pipe, so it is never on a command line.

1. Source the settings of the **administrator** in your shell, from the mode 600 file that holds them. Never put them in a repository.
2. Read what would be sent: `DRY=1 sh lab/zetaris/usl/apply.sh`. It sends nothing and needs no setting.
3. Apply: `sh lab/zetaris/usl/apply.sh`. It sends each statement on its own, drops and creates each view so a second run replaces the first, runs the gate, and prints the settings.
4. As the administrator, grant Proven read on the container: `GRANT SELECT ON VIEW NAMESPACE px_usl TO USER '<proven user>'` (ADR-0023, point 2). The script does not run it: the user name is not Proven's to print.
5. Give Proven the settings the script printed:

```
PROVEN_OSV_TABLE=px_usl.px_package
PROVEN_KEV_VIEW=px_usl.px_kev_cve
PROVEN_EPSS_VIEW=px_usl.px_epss_score
PROVEN_OSV_ID_VIEW=px_usl.px_advisory
```

To go back, unset the four settings. The defaults read the sources directly.

| Exit | Meaning |
|---|---|
| 0 | applied, and the gate returned no row (or `DRY=1`) |
| 2 | a setting is missing or malformed, or `python3` is not on the path; nothing was sent |
| 3 | Zetaris refused a call; the refusal is printed without the key; the layer may be half applied |
| 4 | the gate returned rows; Proven is not pointed at the layer |
| 5 | the gate's answer was not exactly one count (`lab/zetaris/usl/count.py` names what was wrong); Proven is not pointed at the layer |

`DRY` is read by this script only. It is not a Proven setting.

## The parity check

Parity means Proven gets the same rows through the layer as it gets without it.

- **Offline, in `go test ./...`.** `internal/usl` holds one request for each template the layer serves and two that find nothing. The answers sample mode gives are recorded in `testdata/zetaris-usl/answers.json`, and a test holds the recording to sample mode. A second test builds the layer on the local engine from the entity views of `layer.sql`, over the same sample tables, and requires the same rows. The comparer is proven red when one expected cell, row or column changes.
- **Live: `make zetaris-usl-parity`.** It asks the same requests twice through Proven's own adapter, once with the default names and once with the layer's names, and requires equal rows. It needs the layer applied, the grant of step 4, and the three Zetaris settings of Proven. Without them it is skipped, and the skip names the settings that are missing. The data in Zetaris is not the sample data, so the live check compares the layer with its sources and not with the recording. A source that changes between the two calls (EPSS at the day's rollover) can make it red once; run it again.

## When the gate refuses

`apply.sh` ends with `REFUSED. Gate usl-epss-score-valid returned N violating row(s)` and exit code 4.

1. Read the rows it printed: `cve_id`, `epss`, `percentile`, `score_date`.
2. Decide where the bad value comes from. A REST table repeats the source's answer, so a bad row is the source's, or a mistake in the `CREATE LIGHTNING REST TABLE` of the view (a wrong field behind `epss`, or a page of the wrong feed).
3. Fix the source table or its definition as the administrator. Do not edit `gate.sql` to let a row through: a rule that is relaxed is a proposal with an owner, not a repair.
4. Run `apply.sh` again. It replaces the views and runs the gate again.
5. Until it exits 0, leave Proven's four settings unset. A layer that is already applied is not in use until they are set.

Exit 5 means the gate's answer was not exactly the expected shape: a JSON object with `headers` of one name (`violations`),
`data` of exactly one row with one non-negative whole number, and no `error`, `errors`, `exception`, `message` or `msg`
field. Anything else, including a count beside an error field, is not a pass, and the message names what was wrong. Read what Zetaris returned for the query: `DRY=1` prints the exact text, and the SQL editor of the console runs it. Exit 3 names the statement Zetaris refused; its refusal is printed.

## What was not run

- The tests send nothing to Zetaris. On 2026-10-25 the owner's side applied the layer live as administrator: the first apply was refused at `DROP SCHEMASTORE VIEW px_usl.<v>`, Zetaris parses `DROP SCHEMASTORE VIEW <v> FROM px_usl`, and after that fix the apply exited 0 with the gate at 0 rows and `make zetaris-usl-parity` passed. That is the owner's measurement, written here as a report; this repository's tests do not reproduce it. The relation view's `LEFT JOIN`s were created by that apply, but its grain on live data, and the `epss_score_date` column added after it, were not run on Zetaris. The manual of the sponsor is not in this repository.
- The gate's text is not run offline. The tests run its meaning: the same rule in Go, with the pattern taken from `gate.sql`, over the sample rows.
- `make zetaris-usl-parity` is skipped in a build with no Zetaris.
- The tolerated refusals ("exist" for a container that is there and a view that is not yet) assume Zetaris's refusal says so. If it says something else, the first apply stops at that statement.
