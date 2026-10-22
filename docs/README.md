# Proven documentation

Read in this order.

| Page | For | Status |
|---|---|---|
| [INSTALL.md](INSTALL.md) | the quick start: prerequisites, build, the first sample review, livecheck | first version |
| [SAMPLE_MODE.md](SAMPLE_MODE.md) | a whole review with no network and no keys: the scenario, what comes out, how to add one | first version |
| [../demo/live/README.md](../demo/live/README.md) | a live review a judge can rerun: one finding proved by a test, one by Zetaris KEV | first version; the live run needs keys |
| [ARCHITECTURE.md](ARCHITECTURE.md) | the six agents and the pipeline, on one page | first version |
| [CONFIGURATION.md](CONFIGURATION.md) | every setting the reviewer reads, its default and what it changes, including every `proven.yaml` key | first version; the shell's flags are added by the shell's track |
| SETUP.md (not written yet) | running Proven live: Zetaris, Meterless Nightshift, the NVIDIA guardrail | written by the infrastructure track; not here yet |

The design and its reasons are in [DESIGN.md](../DESIGN.md); every decision is in [fleet/decisions/](../fleet/decisions/).
