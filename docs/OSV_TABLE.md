# OSV package table

Proven defines this table. One row is one affected range of one advisory. The file is CSV with a header. A real OSV dump is not loaded here; an administrator puts the file in Zetaris's storage.

## Columns

| column | meaning |
|---|---|
| `id` | the advisory id |
| `aliases` | other ids for the same advisory, separated by spaces, sorted |
| `ecosystem` | the package ecosystem: `Go`, `npm` or `PyPI` |
| `package` | the module path, or the package name in another ecosystem |
| `introduced` | the first affected version; empty or `0` means there is no lower bound |
| `fixed` | the first version that is not affected; the bound is exclusive |
| `last_affected` | the last affected version; the bound is inclusive; used only when `fixed` is empty |
| `severity` | a short severity when the advisory has one, otherwise the first score string |
| `published` | the advisory's published time, as the source wrote it |

`cmd/osv-to-csv` writes this header. It reads a directory of OSV JSON files, one advisory per file, and does not walk subdirectories. A file that does not end in `.json` is skipped. A `versions` list with no ranges becomes one row per version, with `introduced` and `last_affected` set to that version.

```bash
go run ./cmd/osv-to-csv <advisory-dir> <out.csv>
```

`<out.csv>` may be `-` to write to stdout.

## The name Proven queries

The setting is the environment variable `PROVEN_OSV_TABLE`. The default is `px_osvfile.px_osv_pkg`. An empty value means the default. Any other value must be a table name or `schema.table`, made of letters, digits and underscores. Sample mode uses the default and does not read the variable.

For each package the change adds, or whose required version goes up, Proven runs one query. The ecosystem is `Go`, `npm` or `PyPI`.

```sql
SELECT id, aliases, ecosystem, package, introduced, fixed, last_affected, severity, published
FROM px_osvfile.px_osv_pkg
WHERE ecosystem = '<ecosystem>' AND package = '<package>'
```

The version comparison is done in Go, not in SQL. A version is inside a range when `introduced` ≤ version < `fixed`. When `fixed` is empty, the test is `introduced` ≤ version ≤ `last_affected`. When both ends are empty, every version the comparer accepts at or after `introduced` matches. A matching row becomes a finding. The evidence is that query, the rows it returned, and the comparison. No matching row is not a finding; the trace says which comparison excluded the version, or that the query returned no rows. An alias that is a CVE is also looked up in `px_evidence.px_kev_vulns`, the same query the review already uses for a CVE. A missing KEV row does not drop the finding.

Go and npm use semantic versions: `MAJOR.MINOR.PATCH`, an optional leading `v`, and an optional pre-release. Build metadata is ignored. A pre-release is lower than the release it precedes. A string that is not a semantic version, including an npm range such as `^1.2.3`, gives no finding. The trace says it is not a semantic version. The check does not rewrite the string.

PyPI uses PEP 440 for a release (`1.2.3`, any number of numeric segments), a pre-release (`1.0a1`, `1.0b1`, `1.0rc1`, and the spellings `alpha`, `beta`, `rc`, `pre`, `preview`, `c`) and a post-release (`1.0.post1` and `1.0-1`). `1.0` and `1.0.0` are the same version. A pre-release is lower than its release. A post-release is higher than its release. An epoch (`1!1.0`), a dev release (`1.0.dev1`), a local version (`1.0+local`) or any other string gives no finding. The trace says which of those it is. The check does not guess a value for it.

A `replace` or an `exclude` in `go.mod` is not read. Only `require` lines are.

`package.json` contributes `dependencies` and `devDependencies`. Other sections are not read. A dependency whose value is not a string is skipped, and the trace names it.

`requirements.txt` contributes pinned `name==version` lines. The name may carry extras in brackets; the extras are not part of the package name. A blank line or a comment is ignored. Any other line is skipped, and the trace names the line.

## What the administrator runs

Measured on 22 October 2026: a Lightning filestore table over a CSV with a header answers `SELECT` for every row, with Proven's own key, in about 0.10–0.15 seconds. The statements below are those measurements, with this table's name and file. The storage path is filled in by the administrator.

```sql
CREATE LIGHTNING DATABASE PX_OSVFILE DESCRIBE BY "exploration";
CREATE LIGHTNING FILESTORE TABLE px_osv_pkg FROM PX_OSVFILE FORMAT CSV
  OPTIONS (PATH "<customer-data>/shared/px_osv_pkg/px_osv_pkg.csv", header "true", inferSchema "true");
ASSIGN LIGHTNING DATABASE PX_OSVFILE TO USER '<proven>';
GRANT SELECT ON LIGHTNING DATABASE NAMESPACE PX_OSVFILE TO USER '<proven>';
```

Before the grant, the same query was refused.

JSON Lines is not this file's format. A JSON Lines probe returned only the first record. Use CSV, or a JSON array.

## How to refresh it

1. Build a new CSV with `osv-to-csv`.
2. Put that file at the path registered in the `CREATE` statement, replacing the previous file or appending rows.
3. Run the `SELECT` again.

Reads are live. On 22 October, a row appended after the table was registered appeared in the next query. No separate refresh statement was measured. A full file replace uses the same path; that shape was not measured on its own.
