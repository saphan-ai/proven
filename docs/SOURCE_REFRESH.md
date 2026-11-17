# Refresh source data before review

Use **`--refresh-data`** to acquire the public data needed for the dependency changes being reviewed. The same code is available separately as **`proven data sync`**. Neither needs a model or sponsor account. Source acquisition and the existing evidence/query adapters remain separate.

## One run, current provider responses

```sh
PROVEN_GENERATOR_MODE=off proven scan --refresh-data --baseline-commit HEAD~1 .
```

Or refresh the configured store first, then review:

```sh
proven data sync --baseline-commit HEAD~1 .
PROVEN_GENERATOR_MODE=off proven scan --baseline-commit HEAD~1 .
```

The first command is easier to enforce: refresh results appear in the review's `source_refresh` field and its human-readable report. A failed requested refresh makes the analysis incomplete. The standalone sync command writes a JSON refresh report and exits 3 on incomplete acquisition/import.

In either [GitHub Action](ACTION.md), add `refresh-data: true`. Retain the source cache on your own runner, or restore/save it with your normal CI cache mechanism. Do not share one writable evidence database between simultaneous reviews; use a database per job and share the HTTP cache instead.

## What “delta” means here

| Provider | Refresh behavior | Scope and boundary |
| --- | --- | --- |
| OSV | Query selected package versions in batches. Download full advisory records only for new IDs or changed `modified` values. | Handles pagination, withdrawn records and empty current matches. The package/version set is the current review, not a complete worldwide OSV mirror. |
| CISA KEV | Conditional HTTP GET with retained ETag/Last-Modified validators where supplied. | Retains catalogue publication date and counts. The configured CISA-maintained catalogue mirror is an upstream copy, not a freshness guarantee. |
| FIRST EPSS | Query only relevant CVEs and retain each score's date. | EPSS publishes dated observations. A check today can return yesterday's score. |
| deps.dev | Conditional requests where supported; otherwise compare fetched JSON with retained responses. | Licences for indexed package versions, repository mappings/Scorecard observations and available CVSS v3 scores. Coverage is not universal. |

OSV batch timestamps can have microsecond precision while full records retain nanoseconds. The importer accepts that measured precision difference, but rejects another microsecond, an unexpected ID or an invalid timestamp. It keeps the original record unchanged.

HTTP APIs do not all offer an actual incremental stream. Proven does not claim that every provider serves a delta endpoint: it reuses unchanged responses/records and rebuilds the relevant normalized tables. Withdrawn or no-longer-matching advisories disappear from the current materialized scope; their cached historical JSON does not remain active evidence.

## The seven source adapters

| Evidence table | Acquired data | Important limitation |
| --- | --- | --- |
| OSV package ranges and advisory index | OSV records, normalized by the same converter as `osv-to-csv`. | A version must be concrete enough for the registered ecosystem reader. |
| `mal` | Relevant OSV records with `MAL-` IDs. | A package-level advisory is not proof that every version is malicious. |
| KEV | CISA catalogue entries. | Membership is known-exploitation context, not an executed exploit in your application. |
| EPSS | FIRST scores for relevant CVEs. | Probability score and score date, not a yes/no vulnerability verdict. |
| `spdx` | Licence metadata reported for a package version by deps.dev. | Missing metadata is not permissive. Multiple reported licences are retained as metadata; SPDX-expression policy is not implemented. |
| `scorecard` | Source-repository mapping, available `Maintained`/`Signed-Releases` results and known package publication age. | Partial signing scores and missing checks are not converted to a fabricated boolean. Unverified repository mappings remain a provider limitation. |
| `cvss` | Available CVSS v3 numeric scores associated with known CVE aliases. | Unsupported/missing score versions remain unavailable, not zero severity. |

The refresh report records `id`, actual origin, `checked_at`, available `upstream_at`, status, row count and downloaded/reused counts. Provider-level counters can be shared by derived tables. Statuses distinguish updated, unchanged, not-needed, partial measurement coverage and errors. Underlying query evidence retains the imported table's content identity.

**Checking now and having an observation from now are different facts.** For example, a current deps.dev response can legitimately contain an older Scorecard result. Proven shows the observation date where supplied; it does not relabel an old observation as today's measurement.

## Configuration and storage

SQLite is the default. Automatic import also supports PostgreSQL through the existing SQL adapter. CouchDB, local fixture files and Zetaris keep their native import procedures; selecting a different store does not magically supply a refresh writer.

| Variable | Purpose |
| --- | --- |
| `PROVEN_SOURCE_CACHE` | Persistent HTTP/advisory cache, outside the reviewed checkout. Default: the user's cache directory under `proven/feeds`. |
| `PROVEN_EVIDENCE_STORE`, `PROVEN_EVIDENCE_FILE` | Select the store and SQLite file. Keep the file outside the target checkout. |
| `PROVEN_OSV_URL` | OSV API root. |
| `PROVEN_KEV_URL` | Complete KEV catalogue JSON address. |
| `PROVEN_EPSS_URL` | FIRST API root. |
| `PROVEN_DEPS_URL` | deps.dev API root. |

The default upstream addresses are in the [generated configuration reference](CONFIGURATION.md). Operator-defined mirrors pass through the same network policy. URLs do not embed credentials. `PROVEN_NETWORK=local` refuses the default public addresses before acquisition; configure permitted internal mirrors or use a previously imported snapshot without refresh.

Package coordinates are sent to the configured providers. Repository source code, GitHub posting tokens and model credentials are not part of these requests. The HTTP client refuses redirects. There is no package-manager install, shell command or target build during source acquisition.

## Failures and bounds

- A failed download, invalid response or normalization/import error does not replace that source with a successful empty table.
- The cached response is not used as a fresh success after a network failure. A 304 or unchanged OSV revision explicitly validates reuse.
- Successfully refreshed sources remain usable; other sources retain their prior tables. The review records the failure instead of claiming all data is current.
- Missing package licences, unsupported ecosystems and absent Scorecard/CVSS measurements stay explicit as partial coverage. A missing observation is not a healthy result.
- Source tables are replaced atomically one table at a time. This is not a transaction across all providers. Keep separate evidence databases for concurrent jobs.

Current bounds are 200 selected package versions, 20 OSV pages, a 32 MiB response limit, a 128 MiB aggregate response budget, 100,000 rows/64 MiB of normalized content per table and 8 KiB per cell. Refresh is bounded at three minutes within the scan's overall deadline; HTTP requests have a 20-second timeout. A limit is an incomplete refresh, not a clean scan. The deterministic source-rule pass also records its query-limit/deadline failures.

## Measured model-free flow

On 17 November 2026, a controlled Git change adding the public npm package `lodash 4.17.20` was reviewed twice against the live providers, with `PROVEN_GENERATOR_MODE=off`. Both runs completed with one data-proved dependency finding, seven additional data-rule observations, a separate placeholder commit-description warning and **zero model calls**. The repeated run reused all four unchanged OSV advisory records and received a conditional KEV response. [Retained measurement summary](audits/2026-11-17/model-free-refresh.json).

The Scorecard observation in that response was dated 24 August, despite a current provider check. The report retained that older date. This is why retrieval time must not be presented as a fresh upstream measurement. The case demonstrates the selected package and recorded provider responses; it is not a full npm scan, exploit demonstration or general quality benchmark.

The same two-run case also passed through the standalone **Linux arm64 image**, using the README's read-only checkout mount and persistent data volume. Both runs completed with the same counts and zero model calls; the second reused unchanged OSV records and the KEV catalogue. [Container measurement and candidate image identity](audits/2026-11-17/model-free-refresh-linux-arm64.json). Both Linux architectures additionally pass the bundled offline Git/SQLite smoke check; only arm64 was used for this live-provider container measurement.

## Adding a connector

`internal/feeds.Connector` acquires data and returns named normalized snapshots plus refresh metadata. `Connectors()` is the controlled provider registry. `Session.Client` supplies policy-aware HTTP and the retained cache. Connectors do not choose a database; `Sync` receives an import function.

To add a provider:

1. Implement acquisition, pagination, identity checks, normalization and explicit missing-data behavior.
2. Retain actual origin, upstream observation/revision and check time. State whether the provider supplies deltas or only conditional/full responses.
3. Return registered table shapes. The existing `source.Adapter` owns queries and evidence rules; models never write SQL.
4. Test unchanged responses, updates/withdrawals, malformed pages, limits, outage handling and imports through an evidence adapter.
5. Update source coverage and the README in the same change.

The ordinary connector tests use local fixtures and require no accounts. Live acquisition is a separately reported integration check. See [adding a source](ADAPTERS.md#adding-a-source) and [source semantics](SOURCES.md).

Provider references: [OSV API](https://google.github.io/osv.dev/api/), [OSV batch queries](https://google.github.io/osv.dev/post-v1-querybatch/), [CISA-maintained KEV mirror](https://github.com/cisagov/kev-data), [FIRST EPSS](https://www.first.org/epss/) and [deps.dev API](https://docs.deps.dev/api/v3/).
