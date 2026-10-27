# Proven for organisations

What an organisation asks before it lets a code reviewer near its code, and what Proven answers. Each line links to the
page or the decision that holds the detail. Where something is still being built, it says so.

## Your code stays yours

- **It runs where your code is.** On GitHub as an [Action](ACTION.md), on your own Gitea
  ([ADR-0022](adr/ADR-0022-gitea-on-premises-is-an-equal-scenario.md)), or from the command line on your own machines.
- **The change under review runs with no network and no secret.** The Prover runs a test in a container with the
  network off, no environment passed in and one read-only mount of the checkout; the command it runs is built by Proven,
  never taken from a model ([ADR-0025](adr/ADR-0025-the-prover-runs-the-change-with-no-network-and-no-secret.md)).
- **A model's answer is untrusted input.** Every prompt and every answer goes through a guardrail, on by default
  ([ADR-0005](adr/ADR-0005-guardrail-on-by-default.md)), and every query Proven sends to your data is built from a
  fixed template, never written by a model ([ADR-0032](adr/ADR-0032-proven-builds-every-query-from-a-template.md)).
- **Air-gapped mode**, where nothing leaves your network: set `PROVEN_NETWORK=local` and Proven refuses any address
  outside your network, when it reads its settings and again when it dials ([AIR_GAPPED.md](AIR_GAPPED.md),
  [ADR-0034](adr/ADR-0034-proven-runs-air-gapped.md)).

## Every finding can be checked, later, by someone else

- **Evidence, not opinion.** A finding is reported only with a test that fails on the change and passes before it, or
  a query and the rows it returned ([ADR-0002](adr/ADR-0002-evidence-is-a-test-or-a-query-through-zetaris.md)), and
  each finding names its level of proof ([ADR-0044](adr/ADR-0044-every-finding-names-its-level-of-proof.md)).
- **A signed evidence pack for every review**, and `proven verify` re-derives every verdict from the recorded evidence
  offline, so an auditor does not have to trust the stored result ([EVIDENCE_PACK.md](EVIDENCE_PACK.md),
  [ADR-0035](adr/ADR-0035-every-review-leaves-a-signed-evidence-pack.md)).
- **Model receipts kept in custody**: the signed receipts of the model calls behind a review are checked and kept
  ([RECEIPT_CUSTODY.md](RECEIPT_CUSTODY.md)).
- **Machine-readable output for your tools**: JSON and SARIF for code scanning ([OUTPUT_FORMATS.md](OUTPUT_FORMATS.md)).

## Your rules, your data, your models

- **An organisation policy file** sets the floor for every repository, with exceptions that accept one known finding
  until a date, with an owner and a reason ([POLICY.md](POLICY.md),
  [ADR-0036](adr/ADR-0036-an-organisation-policy-and-exceptions-that-expire.md)).
- **No hidden settings.** Every setting is listed with its default in [CONFIGURATION.md](CONFIGURATION.md)
  ([ADR-0013](adr/ADR-0013-no-hardcoded-settings.md)).
- **Your data layer through Zetaris**: vulnerability and exploit data today, and the same path for your own sources
  ([ZETARIS_USL.md](ZETARIS_USL.md), [ADR-0046](adr/ADR-0046-a-data-source-is-one-adapter.md)). Plugging in your own
  lists (banned packages, service criticality, code owners) is *still in the plan*.
- **Your evidence store**: where review evidence is kept is an adapter
  ([ADR-0047](adr/ADR-0047-the-evidence-store-is-an-adapter.md)).
- **Every model is configurable**, and the reviewer comes from a different model family than the author
  ([ADR-0008](adr/ADR-0008-reviewer-model-chosen-against-the-authors-model.md),
  [ADR-0010](adr/ADR-0010-every-reviewer-model-is-configurable.md)). Meterless Nightshift is the primary AI provider
  ([ADR-0019](adr/ADR-0019-meterless-nightshift-is-the-primary-ai-provider.md)).
- **Bring your own second reviewer** from any OpenAI-compatible provider
  ([ADR-0037](adr/ADR-0037-bring-your-own-second-reviewer.md)): *being built.*

## You can watch it and measure it

- **Every review is observable through OpenTelemetry**, into Grafana or the monitoring you already run
  ([ADR-0045](adr/ADR-0045-a-review-is-observable-through-opentelemetry.md)).
- **Proven's own code is security-reviewed as a whole**, piece by piece, and its weaknesses are published once fixed
  ([security review](security/SECURITY_REVIEW.md), [SECURITY.md](../SECURITY.md)).
- **What it does not do yet** is listed, measured and dated, on the front page:
  [Honest limits](../README.md#honest-limits).

## Try it

One command starts a lab on your machine with a Git server, demo pull requests reviewed by Proven and a live dashboard:
[LAB.md](LAB.md).
