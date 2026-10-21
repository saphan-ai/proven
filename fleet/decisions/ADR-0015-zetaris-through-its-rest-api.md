# ADR-0015: proven's Zetaris adapter speaks Zetaris's REST API, not JDBC, so proven carries no Java

**Status:** accepted · **Date:** 2026-10-20
**Decided by:** the owner · **Written by:** the advisor (claude-code, `claude-opus-5-5`), 2026-10-20
**Follows:** [ADR-0001](ADR-0001-go-behind-adapters.md) (Go behind adapters), [ADR-0002](ADR-0002-evidence-is-a-test-or-a-query-through-zetaris.md) (evidence through Zetaris), [ADR-0013](ADR-0013-no-hardcoded-settings.md) (no hardcoded settings)
**Trigger:** Zetaris access arrived ([`planning/CONVERSATION_2026-10-20.md`](../../planning/CONVERSATION_2026-10-20.md), hops 5 and 6)

## Context

On 20 October the sponsor gave us a hosted Zetaris instance and the whole platform as a Docker Compose stack. The sponsor's [getting-started guide](https://github.com/zetaris/getting-started) offers two ways for a program to query it:

- **JDBC**, which needs Java and the driver file shipped with the platform;
- **REST**, which needs an API key created by a user, the organisation's numeric ID and a fresh request ID per call. The guide: *"No Java runtime or JDBC driver is needed."*

proven is written in Go ([ADR-0001](ADR-0001-go-behind-adapters.md)), and Go has no JDBC. Reaching JDBC from Go would mean a Java process beside proven, in every image and on every judge's machine. The advisor proposed REST; the owner:

> *"Go with ADR-0015."*

## Decision

1. **The live Zetaris adapter calls Zetaris's REST API over HTTP.** proven contains no Java and no JDBC driver.
2. **Authentication comes from environment variables only**: the API key and the organisation ID ([ADR-0001](ADR-0001-go-behind-adapters.md), point 4). The adapter makes a new request ID for every call.
3. **The base URL, the organisation ID, the row limit and the timeouts are settings**, validated when they load ([ADR-0013](ADR-0013-no-hardcoded-settings.md)). Which instance a run uses, ours or the hosted one, is one of them.
4. **The adapter follows the API's query lifecycle as the platform's OpenAPI specification describes it**: start a query, read its pages, close it. It closes every query it starts, including on error and on timeout. The routes are taken from the specification during the build, not guessed.
5. **The evidence is the query, not the transport.** The SQL text and the rows it returned go into the review exactly as [ADR-0002](ADR-0002-evidence-is-a-test-or-a-query-through-zetaris.md) says, so anyone can run the same query again, over REST, over JDBC or in the platform's own SQL editor, and get the same rows.
6. **JDBC stays for people and coding agents exploring the data by hand.** It is not on proven's runtime path.
7. **The sample adapter does not change**: `SAMPLE_MODE=true` runs the same queries on a local engine over the samples.

## Consequences

- proven's images stay Go plus Python, with no Java layer.
- The API key carries the permissions of the user who created it, so proven gets its own non-admin user with query rights only, created by the owner (order [`proven-zetaris-local`](../orders/TRAIN_ORDER_proven-zetaris-local.md), W4).
- The OpenAPI specification is served behind the platform's user name and password, not the API key. Reading it is a build-time step, never something proven does at run time.
- The same SQL must run on Zetaris and on the local sample engine. Where the two dialects differ, the queries are written in what both accept; if that is not enough, it needs a decision of its own.
- An HTTP API can change under us during the hackathon. The adapter's tests run against our own instance and the hosted one, so a change shows up as a failing test, not as a review that silently finds nothing.
