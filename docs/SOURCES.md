# Evidence sources: implemented and planned

A database backend stores evidence. A source supplies facts. Selecting SQLite, PostgreSQL, CouchDB or Zetaris does not connect a vulnerability feed or populate its tables.

Proven currently has seven built-in source adapters and eight registered query templates. They query operator-prepared tables. **There is no automatic feed downloader, scheduler or freshness enforcement.** A fresh SQLite database contains no public advisory snapshots. Importing an empty table does not establish coverage of an upstream dataset.

## Current source support

| Source | What is implemented | What the operator must supply |
| --- | --- | --- |
| OSV advisories | Package/version-range matching, six package ecosystems, package and advisory-ID templates. `osv-to-csv` converts a directory of advisory JSON files to CSV. | An acquired, normalized snapshot and the required tables. CSV is not directly accepted by `proven data import`; that command takes the table JSON format. |
| CISA KEV | CVE lookup and evidence/enrichment from prepared rows. | A normalized copy of the catalogue and its source revision. |
| FIRST EPSS | CVE score lookup and ordering/enrichment of findings. | A dated score snapshot, including `score_date`. |
| Malicious packages (`mal`) | Package lookup; rows with a `MAL-` advisory ID can become evidence. Fixture coverage exists. | A normalized malicious-package snapshot. No live OpenSSF feed connector is included. |
| Project health (`scorecard`) | Package lookup and fixed heuristics for maintenance, unsigned releases and young releases. Fixture coverage exists. | Package-to-project mapping and normalized health rows. These heuristics are not a vulnerability verdict or a complete Scorecard assessment. |
| CVSS (`cvss`) | CVE score lookup and severity evidence. Fixture coverage exists. | Normalized scores and their origin. A CVSS score alone does not establish that a particular package/version is affected. |
| Licence metadata (`spdx`) | Package licence lookup and whole-cell comparison against a fixed permissive set. Fixture coverage exists. | Package licence metadata. The SPDX identifier list does not supply package licence mappings. SPDX expression parsing and configurable organisation licence policy are not implemented. |

The built-ins are registered in `internal/querytpl/adapters.go` and `internal/querytpl/adr0039.go`. Their declared origins describe intended provenance; `source.Load` is metadata, not executable ingestion. Query schemas and exact relation names are in [QUERIES.md](QUERIES.md). Import format, snapshot identities and database setup are in [DATABASES.md](DATABASES.md).

An empty query result means the loaded snapshot returned no matching row. A missing source or failed query is unavailable analysis. Neither proves that a package is safe. Snapshot hashes identify imported content; they do not certify accuracy, completeness or freshness. Track upstream revisions and acquisition dates alongside imports.

## Planned connections

The source expansion plan remains part of the project. The following work is not shipped as a live integration:

| Planned work | Purpose | Required before calling it available |
| --- | --- | --- |
| OSV, KEV and EPSS acquisition and refresh | Make the existing first sources usable without hand-built snapshots. | Documented upstream access and data terms, conversion, dated manifests, bounded downloads, failure tests and a reproducible import. |
| OpenSSF malicious-package ingestion | Populate the existing malicious-package adapter. | Advisory identity/deduplication and a tested upstream-to-table conversion. |
| deps.dev / OpenSSF Scorecard ingestion | Populate project-health evidence. | Ecosystem/package/version-to-repository mapping, supported metrics and explicit missing-data semantics. |
| NVD / CVSS ingestion | Populate severity metadata and retain provenance. | CVE identity, score version/source, updates, access requirements and normalization tests. |
| Package licence acquisition | Populate the existing licence adapter. | Package/version provenance, expression handling and an explicit organisation policy. |
| GitHub Advisory Database | Additional advisory acquisition where it adds coverage. | Deduplication with OSV/GHSA aliases and source-specific update handling; no separate live connector exists today. |
| CWE taxonomy | Explain weakness classes. | A taxonomy adapter and rules that keep explanatory classifications separate from proof of a defect. |
| Repository history, prior fixes and reverts | Compare a change with earlier repository evidence. | A bounded history evidence adapter and reproducible comparisons. Reading Git snapshots and commit metadata today does not implement this feature. |
| Issues and pull requests | Bring relevant reported behavior into review context. | An explicit forge connector, permissions, provenance and evidence validation. No general issue/PR evidence connector is shipped. |

The intended order is to make acquisition and freshness reliable for the existing sources, then expand coverage. This is a roadmap, with no promised release dates. Source terms, authentication and availability must be checked when each connector is implemented.

## Adding a source

Use the source adapter interface and registered templates; models do not supply SQL. New sources need a small reproducible fixture, negative controls, provenance and freshness documentation, and a rule that explains what a row does and does not prove. There is no dynamic plugin loader. See the [source adapter example](ADAPTERS.md#adding-a-source) and [roadmap](../ROADMAP.md).
