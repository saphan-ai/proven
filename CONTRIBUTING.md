# Contributing to Proven

Thank you for looking. During the hackathon week (22–27 October 2026) Proven is built by a fleet of agents under one
owner, and outside changes are not merged until the week is over. Issues and ideas are welcome now.

## Reporting

- **A vulnerability:** privately, never in an issue. See [SECURITY.md](SECURITY.md).
- **A bug:** open an issue with the command you ran, what you expected, what happened, and the commit (`git rev-parse
  HEAD`). If it came from a review, attach `<user cache>/proven/logs/trace.jsonl`, or the file in the directory `PROVEN_LOG_DIR` names, with any secret removed.
- **A wrong finding:** open an issue with the finding as Proven printed it and why you think it is wrong. A finding is
  only reported with evidence, so the evidence line is the most useful part.

## Your language is missing?

A language is one small adapter: where the dependencies are declared and which ecosystem to ask about. PHP, Rust and
Kotlin were each added that way. [docs/ADAPTERS.md](docs/ADAPTERS.md) walks through it with worked examples. Open a pull
request with yours: Proven will review it, and it is merged after the hackathon week.

## Building and testing

```bash
go build ./...
go vet ./...
go test ./...
```

A first review needs no network and no keys: [docs/INSTALL.md](docs/INSTALL.md) and
[docs/SAMPLE_MODE.md](docs/SAMPLE_MODE.md).

## How a change is made here

- Every design decision is written down first, as an ADR in [docs/adr/](docs/adr/).
- Every change gets two independent reviews from different model families before it is merged
  ([ADR-0009](docs/adr/ADR-0009-two-reviewer-models-and-the-same-model-fallback.md)).
- Code is written to be read: a comment for every package and exported name, and a comment that says why at every step
  a reader would stop at ([ADR-0027](docs/adr/ADR-0027-the-code-is-written-to-be-read-by-the-judges.md)).
- No setting is hard-coded ([ADR-0013](docs/adr/ADR-0013-no-hardcoded-settings.md)); each one is listed in
  [docs/CONFIGURATION.md](docs/CONFIGURATION.md).
- Commits carry no AI attribution lines.
