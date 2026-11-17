# Evidence sources and acquisition

A database backend stores evidence. A source supplies facts. Proven has seven built-in source adapters and eight registered query templates, plus public-data acquisition through the [refresh connectors](SOURCE_REFRESH.md).

Use `proven scan --refresh-data --baseline-commit HEAD~1 .` to refresh data for the dependency changes being reviewed, or `proven data sync` to populate the configured SQL store separately. No model or sponsor account is needed. Existing operator-prepared table imports remain supported.

## Available source support

| Source | Data acquired and checked | Boundary |
| --- | --- | --- |
| OSV | Version-scoped API queries, ID/modified-time advisory delta cache, normalized package ranges and advisory index. | Dependency readers cover six ecosystems; unresolved versions and unsupported syntax are not guessed. |
| CISA KEV | CISA-maintained catalogue mirror, conditional refresh, CVE lookup. | Known-exploitation context, not proof of an executed application exploit. |
| FIRST EPSS | Dated scores for relevant CVEs, lookup and enrichment. | A probability observation has its own date. It does not determine whether a package is affected. |
| Malicious packages (`mal`) | Relevant OSV/OpenSSF `MAL-` records and the registered package rule. | Package-level metadata must not be described as proving every release malicious; OSV checks version applicability separately. |
| Project health (`scorecard`) | deps.dev repository mappings, available OpenSSF maintenance/signing checks and known release age. | Missing checks and partial signing scores stay unknown. These heuristics are not a vulnerability verdict or a full Scorecard assessment. |
| CVSS (`cvss`) | Available deps.dev CVSS v3 scores associated with known CVE identities. | Missing/unsupported score versions are not zero severity. A severity score alone does not establish package applicability. |
| Licence metadata (`spdx`) | Reported package-version licence metadata through deps.dev and the existing fixed permissive-set rule. | A package-to-licence mapping is needed; the SPDX identifier list does not supply it. General SPDX-expression policy is not implemented. |

Acquisition coverage varies by provider. In particular, deps.dev does not cover every ecosystem supported by the manifest readers. A missing licence or Scorecard observation is reported as unavailable/partial coverage, not converted to a permissive or healthy value.

In model-free mode, bindable registered source rules run deterministically against the dependency changes. The report's `source_checks` keeps these observations separate from proved code findings. OSV's affected-range comparison and its KEV/EPSS enrichment remain the dedicated dependency-evidence path. With a model, refresh also enables the deterministic source-rule pass alongside model investigation.

The source registry is in `internal/source`; built-ins are registered from `internal/querytpl/adapters.go` and `adr0039.go`. Acquisition is in `internal/feeds`. Query schemas are in [QUERIES.md](QUERIES.md); import formats and backend setup are in [DATABASES.md](DATABASES.md).

## Refresh, provenance and completion

`source_refresh` records actual origin, check time, available upstream observation time, status and row/reuse counts. OSV modified timestamps and conditional requests avoid downloading unchanged records where supported. APIs without delta/validator support are queried again; Proven does not invent a delta endpoint or a new observation date.

A failed provider response is not replaced with a successful empty dataset. Existing source tables are retained; successfully refreshed tables remain available and the requested refresh is marked incomplete. Snapshots identify loaded content, not exhaustive coverage, accuracy or permanent freshness. See [refresh behavior and limits](SOURCE_REFRESH.md).

A fresh SQLite database contains no public advisory data until refreshed or imported. An empty query result means the loaded scoped snapshot returned no matching row; it does not guarantee that a package is safe. A missing table/query or unmeasured tool is unavailable analysis.

Automatic source import uses SQLite or PostgreSQL. CouchDB, local fixture files and Zetaris retain their native loading procedures. A backend selection does not configure an upstream feed or an arbitrary write adapter.

## Further connections

| Direction | Additional work needed |
| --- | --- |
| Wider metadata coverage | More package/project mappings, licensing providers and ecosystem coverage. |
| CVSS v4 and additional NVD acquisition | Score-version/source handling, access/rate limits and normalization tests. |
| A full malicious-package mirror | Global advisory acquisition, deduplication and explicit scope beyond selected-version queries. |
| CWE taxonomy | Explanatory classification separate from defect proof. |
| Repository history, prior fixes and reverts | Bounded history evidence and reproducible comparisons. Reading Git snapshots/messages alone does not implement this analysis. |
| Issue/PR evidence search | An explicit search/selection adapter and reproducible evidence rules. The current [scope collector](PR_DELIVERY.md) reads declarations; it is not general issue search. |

## Add another source

Implement acquisition/provenance where needed, register query templates and a deterministic rule, and test positive, negative, missing-data and stale/error cases. Models do not supply SQL. There is no runtime plugin loader. See [the adapter guide](ADAPTERS.md#adding-a-source), [refresh connector guide](SOURCE_REFRESH.md#adding-a-connector) and [roadmap](../ROADMAP.md).
