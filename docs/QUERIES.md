# Queries

Every question Proven asks Zetaris is one of the templates below, and no other. A model never writes a query
([ADR-0032](adr/ADR-0032-proven-builds-every-query-from-a-template.md)): the Investigator and the Skeptic
pick a template and give the values of its parameters, and Proven checks each value against the format on this page and
builds the SQL itself. The OSV check and the EPSS check build their queries from the same templates, so every query
Proven sends has one author. The list lives in `internal/querytpl`; a test holds this page to it.

## The templates

| Template | It answers | Parameters |
|---|---|---|
| `osv-by-package` | the OSV advisory rows of one package | `ecosystem`, `package` |
| `osv-by-advisory` | the OSV advisory with one advisory id | `advisory_id` |
| `kev-by-cve` | the CISA known-exploited entry of one CVE | `cve_id` |
| `epss-by-cve` | the EPSS score of one CVE | `cve_id` |

A template that is not in this table is refused. A request names its template and gives exactly its parameters: a
missing parameter and an extra one are each refused.

## The formats

A value is checked against its format before any SQL is built. Each format is an allow-list of ASCII, so a single quote,
a backslash, a control byte, a byte that is not UTF-8 and a space or a semicolon are outside every one of them. A value
outside its format is refused with a named reason. It is never escaped, so the text between the two quotes of a literal
is exactly the value.

| Parameter | Format |
|---|---|
| `ecosystem` | exactly `Go`, `npm`, `PyPI`, `Packagist` or `crates.io` |
| `package`, ecosystem `Go` | a module path of letters, digits and `. _ ~ / -`, at most 300 characters, starting with a letter or digit |
| `package`, ecosystem `npm` | an optional `@scope/` and a name of letters, digits and `. _ ~ -`, each starting with a letter or digit, at most 214 characters |
| `package`, ecosystem `PyPI` | letters, digits and `. _ -`, starting and ending with a letter or digit, at most 214 characters. It is put in its PEP 503 form first, because the OSV table is keyed by that form |
| `package`, ecosystem `Packagist` | `vendor/name` in lower case. Each part is letters, digits and `. _ -`, starts and ends with a letter or digit, and is at most 214 characters. A name is put in lower case first, because Packagist is not case-sensitive |
| `package`, ecosystem `crates.io` | a crate name of 1 to 64 characters. It starts with a letter. The rest are letters, digits, `_` or `-`. The spelling is kept: case is not folded, and `_` is not turned into `-` |
| `cve_id` | `CVE-` and four digits, a hyphen and four to twelve digits |
| `advisory_id` | an upper-case prefix of two to sixteen letters or digits, then one to six hyphen-separated parts of letters and digits, such as `GHSA-xxxx-xxxx-xxxx`, `GO-2021-0113` or `PYSEC-2021-1` |

A package version is not a parameter of any template. The OSV table is read by package, and the version is compared in
Go by the dependency check, never in SQL.

## The names of tables and views

A table or view name is a setting ([CONFIGURATION.md](CONFIGURATION.md)), not a value from a model or from the change.
It goes into the SQL unquoted, so it is checked again when the query is built: one identifier, or `schema.name`, of
letters, digits and underscores. A setting that is not a plain name is refused at load and again at build. The names
and their defaults:

| Template | Name | Setting | Default |
|---|---|---|---|
| `osv-by-package` | table | `PROVEN_OSV_TABLE` | `px_osvfile.px_osv_pkg` |
| `kev-by-cve` | view | `PROVEN_KEV_VIEW` | `px_evidence.px_kev_vulns` |
| `epss-by-cve` | view | `PROVEN_EPSS_VIEW` | `px_evidence.px_epss_cve_rows` |
| `osv-by-advisory` | view | none | `px_evidence.px_osv_id` |

## What a refusal does

A refused request sends no query. The finding has no query evidence from it, and the trace names the reason, one of:

| Reason | Meaning |
|---|---|
| `query template is not on the list` | the name is not one of the templates above |
| `query template parameter is missing` | a parameter of the template was not given |
| `query template parameter is not one the template takes` | a parameter the template does not have was given |
| `query parameter value is outside its format` | a value does not match its format |
| `table or view name is not a plain name` | a setting that names a table or view is not a plain name |
| `a model's answer carries SQL text` | the answer had a `sql` field; SQL from a model is never run |

The refusal names the template and the parameter and never repeats the model's text. The Judge then decides as it does
for any finding without query evidence. A package name from the change that is outside its format is a trace note
("osv query refused") and the check goes on to the next package.

## The second check

`internal/query` keeps a gate in front of the live client. A query Proven builds passes it by construction; it holds for
a statement from any other caller, such as the `livecheck` probe. It reads only the single quote as a quote. It refuses a
text with a backslash, a double quote, a backtick or a control byte anywhere, because a dialect may read any of them as
a quote or an escape and a single quote inside one would hide a second statement from a gate that tracks the single
quote alone. It also refuses invalid UTF-8, an unclosed single quote, and, outside a single-quoted literal, a semicolon
or the start of a comment.

## What this page does not claim

Proven does not rely on a read-only Zetaris key. Whether Zetaris offers one is measured separately; until that is
measured and recorded, the protection is the template list, the formats and the gate above.

## Adding a template

A new kind of evidence is a new template: a change to `internal/querytpl` and to this page, reviewed like any other
change. The test `TestQueriesDocumentListsEveryTemplate` turns red if a template or a parameter is missing from this
page.
