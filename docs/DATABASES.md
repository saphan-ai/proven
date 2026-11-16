# Evidence databases

The store interface registers source tables and runs validated query-template requests. The reviewer never accepts model-written SQL. SQLite and PostgreSQL use the same source and template definitions as the existing adapters.

## SQLite

The default database is `<user cache>/proven/evidence.sqlite`. Set `PROVEN_EVIDENCE_FILE` to use another file. Keep it outside the reviewed repository. SQLite needs no server. The old `local` store is a separate fixture SQL implementation, not SQLite.

## Importing snapshots

Prepare a JSON document with `version: 1` and a `tables` array. Each table has `name`, `columns`, `rows` and `snapshot`; rows are arrays of strings. For example:

```json
{
  "version": 1,
  "tables": [{
    "name": "px_evidence.px_epss_cve_rows",
    "columns": ["cve_id", "epss", "percentile", "score_date"],
    "rows": [["CVE-2021-44228", "0.9", "0.99", "2026-11-16"]],
    "snapshot": "example only; replace with your source revision"
  }]
}
```

The example row illustrates the format; it is not current vulnerability data. Use the configured relation names from the settings reference. The source schemas and upstream origins are described in [OSV_TABLE.md](OSV_TABLE.md), [QUERIES.md](QUERIES.md) and the source adapter catalogue.

```sh
proven data import dataset.json
```

The importer replaces each table and its snapshot metadata in one transaction. An invalid table cannot replace that table's previous data. Tables are published individually; an import containing several tables is not an atomic release of the entire dataset set. Every registered SQL table records a SHA-256 identity over its name, columns, rows and supplied revision. Query evidence reads rows and identity in the same transaction. Reading the same table later does not invent a new dataset identity.

The importer does not automatically download or refresh public feeds. Source preparation and freshness remain the operator's responsibility. A missing table or failed read is reported explicitly.

## PostgreSQL

Set `PROVEN_EVIDENCE_STORE=postgres` and `PROVEN_POSTGRES_URL` to your connection URL. Credentials are read from the environment and errors do not print the URL. TLS follows PostgreSQL connection options; configure it for your deployment. PostgreSQL connections use Proven's network policy. The same `proven data import` command applies.

## Existing backends

- `couchdb`: the existing Mango adapter and evidence-record viewer remain supported.
- `zetaris`: the existing REST adapter remains available for installations that already use it.
- `local`: the existing fixture SQL file adapter remains available with `PROVEN_EVIDENCE_FILE`.

These adapters retain their native loading procedures. Selecting SQLite or PostgreSQL reads no Zetaris credentials. Sample mode continues to use its recorded local fixtures.
