# ADR-0023: Proven's evidence sources are Zetaris REST tables that an administrator registers and Proven's own key reads live; OSV by package needs another path

**Status:** accepted · **Date:** 2026-10-21
**Decided by:** the owner · **Written by:** the infrastructure master (claude-code, `claude-opus-5-5`), 2026-10-21
**Follows:** [ADR-0002](ADR-0002-evidence-is-a-test-or-a-query-through-zetaris.md) (evidence through Zetaris), [ADR-0004](ADR-0004-first-evidence-sources.md) (the first sources), [ADR-0013](ADR-0013-no-hardcoded-settings.md) (no hardcoded settings), [ADR-0015](ADR-0015-zetaris-through-its-rest-api.md) (REST, not JDBC), [ADR-0022](ADR-0022-gitea-on-premises-is-an-equal-scenario.md), point 3 ("How the sources are loaded is the next decision.")
**Trigger:** the evidence exploration on our own Zetaris platform, train [`proven-zetaris-local`](../orders/TRAIN_ORDER_proven-zetaris-local.md) ([`planning/CONVERSATION_2026-10-21.md`](../../planning/CONVERSATION_2026-10-21.md), hops 20 to 23)

## Context

[ADR-0004](ADR-0004-first-evidence-sources.md) named the sources: OSV, CISA KEV, FIRST EPSS and the repository's own history. [ADR-0002](ADR-0002-evidence-is-a-test-or-a-query-through-zetaris.md) says a finding counts only with a query through Zetaris and the rows it returned. On 21 October the owner asked for the loading to be explored, not built (translated from Polish):

> *"Now only exploration, no data in the database. Preparation only."*

The master measured on our own platform (train `proven-zetaris-local`), with Proven's own non-admin API key. Every statement that registers a source was refused for that key: `{"msg":"User is not permitted to run this command"}`. The sponsor's command reference marks `CREATE LIGHTNING DATABASE`, `CREATE LIGHTNING REST TABLE`, `TEST REST REQUEST`, `CREATE SCHEMASTORE CONTAINER` and the cache commands as *"Administrators only"*. The owner then allowed the next step (translated from Polish):

> *"And you may do b."*

So the administrator account registered the sources as definitions, every object named with the prefix `px_`, and Proven's own key measured them. No cache was made and no row was stored. After the result, the owner (translated from Polish):

> *"OK. Save it in decisions."*

**What was measured with Proven's own key** (each statement run twice; wall-clock time of the HTTP call):

| source | what the table asks | rows | time |
|---|---|---|---|
| CISA KEV | the whole feed, `cisa.gov` | 1754, the same as the feed's own count | about 0.6 s (0.56–0.87 s) |
| FIRST EPSS, one CVE | `api.first.org`, `cve=CVE-2021-44228`, fixed | 1 | about 0.2–0.3 s |
| FIRST EPSS, one page | `api.first.org`, `limit=100&offset=0`, fixed | 100 | about 0.2–0.3 s |
| OSV by id | `api.osv.dev`, GET of one advisory | 1 | about 0.5–0.6 s |
| KEV × EPSS | a join of the KEV view and the fixed-CVE EPSS view | 1 (CVE-2021-44228) | about 1.1–1.3 s |
| OSV by package | `api.osv.dev/v1/query`, POST with a nested JSON body | **does not work** | — |

**Why OSV by package fails.** A REST table can send a POST with `application/json`, but its body is a flat map of strings. OSV's query needs a nested `package` object. Every form tried was sent as a flat object, and OSV answered 400. The encoding `NONE`, listed in the platform's API specification, is refused by the SQL parser.

## Decision

1. **Each evidence source is a definition in Zetaris, not a copy.** An administrator registers it as a Lightning REST database with one REST table per request, and a SchemaStore view that turns the JSON into rows. No cache, no stored rows: the definition holds the URL and the request, nothing else.
2. **Registration is the administrator's act, never Proven's.** Proven's key only reads. It can read an object only after both of these, from the sponsor's command reference, §21:
   ```sql
   ASSIGN LIGHTNING DATABASE <database> TO USER '<proven user>'
   GRANT SELECT ON LIGHTNING DATABASE NAMESPACE <database> TO USER '<proven user>'
   GRANT SELECT ON VIEW NAMESPACE <container> TO USER '<proven user>'
   ```
   The assignment makes the database visible; without the grant, a SELECT is refused. Views need their own grant on the view namespace.
3. **A REST table's URL and parameters are fixed when it is created.** A query cannot pass them. So an EPSS `cve` value or a page (`limit`, `offset`) cannot come from Proven's SQL: each fixed query or page is its own table, made by the administrator.
4. **Every SELECT fetches the source live.** The second run took about as long as the first for every source, as the sponsor's guide says (*"an uncached query fetches the endpoint again"*). The rows in a review are the source's answer at the time of the query ([ADR-0002](ADR-0002-evidence-is-a-test-or-a-query-through-zetaris.md)), and a join calls both sources.
5. **Two endpoints, two uses.** Registration statements go to `POST <REST base URL>/sql-editor/sqls/run`, as the administrator. `POST <REST base URL>/query/sql/start` accepts SELECT only; anything else returns 400 `Bad Select SQL`.
6. **The REST base URL, the organisation's numeric ID and the key are settings** ([ADR-0013](ADR-0013-no-hardcoded-settings.md), [ADR-0015](ADR-0015-zetaris-through-its-rest-api.md)), read from the environment. No URL, ID or key is written in Proven's code.
7. **The exploration objects stay, as definitions.** The `px_` database, its four REST tables (`px_kev`, `px_epss_cve`, `px_epss_page`, `px_osv_id`), the container `px_evidence` and its three views are kept. They survived a platform restart: after `docker compose down` and `up -d`, the KEV view still answered 1754. Today nothing else is registered and no data is stored. The build's own definitions are product work, from 22 October.

## Consequences

- KEV, EPSS for a fixed CVE or page, OSV by advisory id, and their joins can be evidence today, with no stored data and no Java.
- **OSV by package is not possible as a REST table on this build.** The owner has not chosen the path yet. The options:
  1. OSV's data in storage that Zetaris reaches, so the query runs through Zetaris as [ADR-0002](ADR-0002-evidence-is-a-test-or-a-query-through-zetaris.md) asks. This stores data, which is not allowed today.
  2. Proven calls OSV outside Zetaris for package queries. The evidence for those findings is then not a Zetaris query, which [ADR-0002](ADR-0002-evidence-is-a-test-or-a-query-through-zetaris.md) does not allow as it stands.
  3. OSV by advisory id only (GET, which works), one table per advisory, which needs the administrator for each new advisory.
- **The repository's history (ADR-0004) is not measured.** Note only: GitHub's REST API (`api.github.com`) as an administrator-registered table per fixed URL, or a file in storage.
- **A fixed URL per table limits what a review can ask.** A CVE that has no table yet needs the administrator first. How Proven asks about a CVE it meets for the first time is not decided.
- **The full EPSS feed (391,915 rows) does not fit a REST table**; one table holds one page. FIRST's daily CSV file in storage is the other way, and it stores data.
- **Live fetches cost time and touch rate limits.** No rate limit was met in the exploration (a few dozen calls per source). A review that runs many queries calls CISA, FIRST and OSV each time; the sponsor's `CACHE TABLE` is a session cache, administrator-controlled, and was not used.
- **The on-premises scenario** ([ADR-0022](ADR-0022-gitea-on-premises-is-an-equal-scenario.md), point 3) needs copies in storage, because a live REST table calls the public source. Not measured.
- **A join is an inner join by default** (the sponsor's guide): a CVE absent from one side is dropped, so a review must not read a missing row as "not exploited".
- **The hosted instance is not measured.** Everything above was measured on our own platform.

## REDACTED

| what | times | how |
|---|---|---|
| Proven's user name on the platform | 3 | written as `<proven user>` |
| the platform's address, the organisation's ID and the database's internal ID | — | left out; the address is a setting |

macm5, macmini and runner-cursor stay visible by the owner's rule of 2026-10-17.
