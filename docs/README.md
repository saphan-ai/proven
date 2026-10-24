# Proven documentation

Read in this order.

| Page | For | Status |
|---|---|---|
| [LAB.md](LAB.md) | the judges' lab: `make up`, the pull request, and how to stop it | first version |
| [INSTALL.md](INSTALL.md) | the quick start: prerequisites, build, the first sample review, livecheck | first version |
| [SAMPLE_MODE.md](SAMPLE_MODE.md) | a whole review with no network and no keys: the scenario, what comes out, how to add one | first version |
| [../demo/live/README.md](../demo/live/README.md) | a live review a judge can rerun: one finding proved by a test, one by Zetaris KEV | first version; the live run needs keys |
| [ARCHITECTURE.md](ARCHITECTURE.md) | the six agents and the pipeline, on one page | first version |
| [ADAPTERS.md](ADAPTERS.md) | how to add a language (PHP with Composer as the worked example) and how a deterministic tool plugs in (the Go race detector) | first version |
| [CONFIGURATION.md](CONFIGURATION.md) | every setting the reviewer reads, its default and what it changes, including every `proven.yaml` key | first version; the shell's flags are added by the shell's track |
| [QUERIES.md](QUERIES.md) | every question Proven asks Zetaris: the query templates, each parameter's format, and what a refusal does | first version |
| [OUTPUT_FORMATS.md](OUTPUT_FORMATS.md) | JSON and SARIF, two writings of one review result | first version |
| [OSV_TABLE.md](OSV_TABLE.md) | the OSV package table, the CSV converter, and how an administrator loads and refreshes it | first version |
| [LAB.md](LAB.md) | the keys a live review needs, per generator mode, and where to get them | first version |
| SETUP.md (not written yet) | running Proven live: Zetaris, Meterless Nightshift, the NVIDIA guardrail | written by the infrastructure track; not here yet |

The design and its reasons are in [DESIGN.md](../hackathon/DESIGN.md); every decision is in [fleet/decisions/](adr/).
