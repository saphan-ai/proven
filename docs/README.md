# Proven documentation

Read in this order.

| Page | For | Status |
|---|---|---|
| [WALKTHROUGH.md](WALKTHROUGH.md) | a judge's walkthrough: start the lab, read one review and its evidence, watch it on the dashboards, check the image, then Proven on your own GitHub repository | first version |
| [TOUR.md](TOUR.md) | Proven in pictures: the lab's start page, a review as a page, and the six dashboards | first version |
| [LAB.md](LAB.md) | the judges' lab (`make up`, the pull request, and how to stop it) and the keys a live review needs, per generator mode | first version |
| [LAB_GO_LIVE_TEST.md](LAB_GO_LIVE_TEST.md) | the owner's step by step test of the three ways to take the lab live, with the output to expect | first version |
| [INSTALL.md](INSTALL.md) | the quick start: prerequisites, build, the first sample review, livecheck | first version |
| [SAMPLE_MODE.md](SAMPLE_MODE.md) | a whole review with no network and no keys: the scenario, what comes out, how to add one | first version |
| [../demo/live/README.md](../demo/live/README.md) | a live review a judge can rerun: one finding proved by a test, one by Zetaris KEV | first version; the live run needs keys |
| [ARCHITECTURE.md](ARCHITECTURE.md) | the six agents and the pipeline, on one page | first version |
| [ADAPTERS.md](ADAPTERS.md) | how to add a language (PHP with Composer, and Kotlin with Gradle, as the worked examples), a deterministic tool (the Go race detector), a data source and an output format | first version |
| [CONFIGURATION.md](CONFIGURATION.md) | every setting the reviewer reads, its default and what it changes, including every `proven.yaml` key | first version; the shell's flags are added by the shell's track |
| [ENTERPRISE.md](ENTERPRISE.md) | Proven for organisations: where your code runs, how a finding is checked later, your rules, data and models, and what is still being built | first version |
| [POLICY.md](POLICY.md) | one organisation policy for every repository's `proven.yaml`, and exceptions that name an owner and expire | first version |
| [QUERIES.md](QUERIES.md) | every question Proven asks Zetaris: the query templates, each parameter's format, and what a refusal does | first version |
| [OUTPUT_FORMATS.md](OUTPUT_FORMATS.md) | the five writings of one review result: text, Markdown and HTML for people, JSON and SARIF for machines | first version |
| [ZETARIS_USL.md](ZETARIS_USL.md) | the semantic layer in Zetaris over the evidence: four entities and their relations, the data-quality gate, how to apply it, the parity check, and what to do when the gate refuses | first version; not yet applied to a live Zetaris |
| [OSV_TABLE.md](OSV_TABLE.md) | the OSV package table, the CSV converter, and how an administrator loads and refreshes it | first version |
| [RECEIPT_CUSTODY.md](RECEIPT_CUSTODY.md) | the custody check of model answers and the audit command that counts it over a receipts file and a trace | first version |
| [JUDGES_GATEWAY.md](JUDGES_GATEWAY.md) | the owner's runbook for the gateway that gives a judge's lab our model service and our evidence, with the judge's own token: run it, the tunnel and Access, issue and revoke a token, read a refusal | first version; the Cloudflare steps are not yet run |
| SETUP.md (not written yet) | running Proven live: Zetaris, Meterless Nightshift, the NVIDIA guardrail | written by the infrastructure track; not here yet |

The design and its reasons are in [DESIGN.md](../hackathon/DESIGN.md); every decision is in [fleet/decisions/](adr/).
