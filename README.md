# Proven

**An AI code reviewer that has to show its evidence.** Every problem it reports comes with proof you can check
yourself: a test that fails on the change, or a data query and the rows it returned. If it cannot show the evidence, it
does not report the problem.

> [!IMPORTANT]
> **NOTICE: the reviewer is now GLM-5.3, not NVIDIA Nemotron. NVIDIA's safety model stays the guardrail.**
>
> Proven was built around Nemotron, and Nemotron through Meterless Nightshift remains supported. But since 25 October
> NVIDIA's hosted API fails in windows: HTTP 500 "Failed to generate completions", and calls that never answer. We
> measured it, with NVIDIA's own request ids, and reported it twice; neither report has been resolved
> ([report 1](hackathon/nvidia-support/REPORT_1_intermittent-500_2026-10-26.md),
> [report 2](hackathon/nvidia-support/REPORT_2_follow-up_2026-10-26.md)).
>
> **A reviewer must not be weaker than the developer it checks.** The code here is written by strong coding models; a
> junior cannot judge a senior. A review that times out half-way is weaker than none, because it looks like a review.
> So the pull-request reviews here use a strong model that answers: **GLM-5.3, called directly on QwenCloud, outside
> Nightshift**, and every review comment names the model that answered. This stays the default. Nemotron through
> Nightshift remains a supported choice for anyone who wants it, and NVIDIA's safety model still checks every prompt and
> answer.

## The problem

Coding agents now write code faster than people can read it. The hard question is no longer "can this be written?" but
**"can this change be trusted, and on what evidence?"**

A green test run does not answer it. One example from real work: an agent wrote a protection against deleting a
repository's main branch, plus 16 tests. All 16 passed. A reviewer then **removed the protection** and ran them again.
All 16 still passed: git itself was stopping the deletions one step earlier. The tests proved git, not the fix
([source](hackathon/planning/answers/master-the-bridge-commits-without-a-shared-lock.md)).

An AI reviewer that only gives its opinion has the same problem one level up: it sounds sure, and you cannot check it.

## How Proven works

Proven reviews a pull request with six agents: a **Planner** names the risks, an **Investigator** gathers facts, a
**Prover** tries to prove each problem, a **Skeptic** tries to refute it, a **Judge** keeps only what has evidence, and a
**Writer** writes the report.

- **No evidence, no finding.** A finding counts only with a test that fails on the change and passes on the code before
  it, or a query and the rows it returned (known vulnerabilities from OSV, CISA KEV and FIRST EPSS, through Zetaris).
  The Judge counts evidence by fixed rules; a model's confidence is never enough.
- **The change runs locked down.** The Prover runs the test in a container with no network and no secret, and the
  command is built by Proven, never taken from a model
  ([ADR-0025](docs/adr/ADR-0025-the-prover-runs-the-change-with-no-network-and-no-secret.md)).
- **Checkable later, by someone else.** Every review leaves a signed evidence pack, and `proven verify` re-derives every
  verdict from it offline ([EVIDENCE_PACK.md](docs/EVIDENCE_PACK.md)).
- **More than one model family.** A security review can run on a second model family, and every review names the
  models that answered. A safety model checks every prompt and every answer.
- **It runs where your code is**: GitHub, your own Gitea, your own servers, or fully air-gapped
  ([AIR_GAPPED.md](docs/AIR_GAPPED.md)).

## See a real review

- **On GitHub:** [pull request #1](https://github.com/saphan-ai/proven/pull/1), reviewed by Proven's own Action, with the
  comment on the pull request.
- **On your machine, in minutes:** `make up` starts a lab with Gitea, sample pull requests (clean ones and ones with
  planted defects), Proven's reviews on them and Grafana dashboards. Step by step: [docs/LAB.md](docs/LAB.md).
- **With no network and no keys:** [docs/SAMPLE_MODE.md](docs/SAMPLE_MODE.md).

## Use it on your repository

Add Proven as a GitHub Action (Gitea Actions works the same way):

```yaml
on:
  pull_request:
permissions:
  contents: read
  pull-requests: write
jobs:
  review:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - uses: saphan-ai/proven@<full commit SHA>
        with:
          comment: true
```

The model, the guardrail and the data sources are settings, never built in: [docs/ACTION.md](docs/ACTION.md) and
[docs/CONFIGURATION.md](docs/CONFIGURATION.md). Output is a pull-request comment, JSON, SARIF for code scanning, and
text ([OUTPUT_FORMATS.md](docs/OUTPUT_FORMATS.md)).

## What is measured

- **Meterless conformance**, on Meterless's own runner: H-MEM 13 of 13, World Model 8 of 8, Markovian 8 of 8
  ([scorecards](test/conformance/meterless/)).
- **Against Dependabot** on this repository: Proven found all 6 of Dependabot's Go alerts, each with its evidence, plus
  15 advisories Dependabot did not list. Its 45 npm alerts are marked "not checked", because the evidence store held no
  npm advisories at the time ([report](test/corpus/reference/dependabot-fbdf7d3/REPORT.md)).
- **Forged evidence is refused:** on the forgery bench every single changed value in a pack is refused, and when the
  forger recomputes the hashes and signs again, every change that moves a finding's outcome is still refused
  ([how it is tested](docs/EVIDENCE_PACK.md#the-forgery-bench)).
- **The problem, measured on the project itself:** while Proven was built, 240 independent reviews raised 201 blocking
  problems in work an agent had called finished, and 116 were stopped before reaching the main branch (25 October).

What is not measured yet is listed below, with dates.

## Honest limits

What Proven does not do yet, or does not do well, measured and dated. This list is updated with every push.

- **Scout Intent: how well it sorts changes is measured; whether it improves reviews is not.** On a blind set of 46
  changes frozen before the run (English, Polish and Thai, including disguised ones) it got 26 of 46 right, and seven
  of the frozen labels are disputed ([report](test/conformance/meterless/scout-intent/blind/v1/REPORT.md)). In some
  misses a security change was taken for a documentation change and would have got a lighter review; a guard added on
  26 October now runs those as security, and none of the six such misses runs a lighter plan any more. It also asks for clarification more often than its own ceiling allows. Whether its plans
  give the same findings for less work is not measured yet. Since ADR-0053 a change Scout called documentation or dependency that is not verified prose or a verified manifest runs in class security, doubled; a change Scout called feature, refactor or configuration, or an ambiguous one, still runs in the class its rules give it.
- **Markovian: its effect on reviews is not measured yet.** It passes its conformance checks (8 of 8), but whether
  reviewing a large change in bounded steps finds as much as one whole review is still to be measured.
- **Scout Intent and Markovian are not measured yet.** Their benefit on review quality is unknown. All four engines are
  ON by default (ADR-0053), in Proven and in the lab's sample reviews, and each has its own off setting:
  `PROVEN_MARKOVIAN=off`, `scout_intent: off` in `proven.yaml`, `PROVEN_HMEM=off` and `PROVEN_WORLD_MODEL=off`. Whether
  Markovian and Scout Intent help is UNMEASURED.
- **H-MEM reads its whole state file** when it opens and rewrites it on every write. A size cap is being added; until
  then it is meant for one repository's memory, not a large organisation's.
- **The corpus score was recorded while NVIDIA's API was failing** (HTTP 500 and timeouts on 26 October), so the NVIDIA
  column shows mostly failed calls, every failure named. A second column on GLM-5.3 was tried on 27 October and stopped:
  the model's reasoning made most calls run past 300 seconds, and the planner was sent unchanged files too. Both are
  being fixed; until a run completes, there is no corpus score on GLM-5.3.
- **Evidence first means some real bugs go unreported.** Proven reports a finding only with a failing test or a query
  and its rows. A real concurrency bug, design flaw or business-logic error that it cannot prove during the review is
  left out, not guessed at. This trades recall for precision, and the trade is not measured yet.
- **One judging rule is a heuristic.** When a query's rows are read as refuting a finding (`QueryContradicts`), the
  rule looks for related text and a denial word such as "false" or "clean" in a cell. A badly shaped row could refute a
  valid finding. Typed, query-specific checks with adversarial tests are the planned fix.
- **`proven verify` checks the record, not the run.** It proves that each verdict follows from the evidence recorded in
  the signed pack; it does not re-run the tests, and it cannot show that a review found every defect.
- **The lab's demo runs on sample data.** Live mode needs Zetaris and Meterless Nightshift, your own or ours through an
  authenticated gateway.
- **The review spine is one long function** (`pipeline.review`, about 500 lines). Splitting it is accepted debt until
  after 6 November ([what is still in the plan](BUILD_LOG.md#next-everything-still-in-the-plan)).

## Built in public, in six days

Proven was built from 22 to 27 October 2026 at the [Open Agent Hackathon 2026](https://hackathon.genai.works/event/open-agent-hackathon-2026),
organised by GenAI Works, in the track *The Agent That Can Explain Why*, by team Saphan.ai: one person,
[Marcin "Wato" Marzec](https://wato.org/), and a fleet of AI coding agents run by [Saphan Studio](https://saphan.ai/).
The agents write the code; Marcin decides and approves every merge. **Coding started on 22 October 2026, at 00:00 UTC**,
and everything before it was preparation, all of it public.

<a href="https://vimeo.com/1236933110"><img src="assets/proven-loop.gif" alt="Proven: a pull-request reviewer that has to prove what it says. The opening slides of the introduction video, looping; click to watch it." width="100%"></a>

<p align="center"><a href="https://vimeo.com/1236933110"><b>▶ Watch the introduction</b></a> (6 min).</p>

- **[How it was built, day by day](BUILD_LOG.md)**: what landed each day, what broke, what each push brought.
- **[Every change in the open](hackathon/fleet/)**: each order, the builder's report, two independent reviews and the
  decision; [the ledger](hackathon/fleet/LEDGER.md) shows each change's rounds and outcome.
- **[The hackathon guide](hackathon/README.md)**, **[the design](hackathon/DESIGN.md)** and **[how the fleet works](hackathon/SAPHAN.md)**.
- **[Design decisions](docs/adr/)**: what was decided, and why.
- **[Security](SECURITY.md)** and **[the security review of Proven's own code](docs/security/SECURITY_REVIEW.md)**.

**Contact:** Marcin Marzec, [marcin.marzec@wato.org](mailto:marcin.marzec@wato.org), [LinkedIn](https://www.linkedin.com/in/marcinmarzec).

**Found a vulnerability?** Please report it privately: [Report a vulnerability](https://github.com/saphan-ai/proven/security/advisories/new) on this repository's Security tab (GitHub's private vulnerability reporting is on). Not in a public issue. See [SECURITY.md](SECURITY.md).

## Repository layout

The repository holds two things: **the product**, and, in one folder, **the record of how it was built in the hackathon
week** ([ADR-0030](docs/adr/ADR-0030-the-product-and-the-hackathon-live-apart.md)).

| Path | What it is |
|---|---|
| `cmd/`, `internal/` | Proven's code: the command line and the reviewer |
| `docs/` | how to install, run and configure Proven; `docs/adr/` holds the design decisions, `docs/security/` the security review |
| `docs-site/` | the documentation website, built from this repository's Markdown and from the code |
| `SECURITY.md` | what Proven protects, and how to report a vulnerability |
| `test/` | the test corpus (`test/corpus/`) and the Meterless conformance runner (`test/conformance/`) |
| `testdata/`, `demo/`, `assets/` | sample scenarios, the live demo, images |
| `BUILD_LOG.md` | how Proven was built, day by day, in the hackathon week |
| `hackathon/` | the build week's record: `fleet/` (every change's order, report, review and decision), `planning/` (the topic, the daily conversations), `SAPHAN.md`, `DESIGN.md` |
| `metadata.json`, `metadata.schema.json` | the hackathon's submission metadata, at the root until the hackathon ends |

## Documentation site

The documentation website is built from this repository's own Markdown and from the code, and packaged as one container image. Run it locally:

```bash
docker build -f docs-site/Dockerfile -t proven-docs .
docker run -p 3000:3000 proven-docs
```

Then open `http://localhost:3000/docs`. The command-line, settings and adapter pages are written from the code during the build, and the build fails when a link is broken, a page matches a known shape of an internal path, a private host name or a credential (the list is in `docs-site/scripts/guards.mjs`; a value outside those shapes is not caught), or a generated page is out of date. After a flag, a setting or an adapter interface changes, regenerate the committed copies with `sh docs-site/scripts/generate.sh docs-site/generated`.

Deploy it to Cloudflare, from `docs-site/`, with one command:

```bash
CLOUDFLARE_ACCOUNT_ID=<your account id> CLOUDFLARE_API_TOKEN=<a token that may edit Workers, Containers and DNS> npx wrangler@4.45.0 deploy
```

It needs Docker running, and it serves `proven.saphan.ai`: the route in `docs-site/wrangler.jsonc` and its `DOCS_HOST` setting name that host, and a different host is changed in both.

## For technical readers

| | |
|---|---|
| [docs/README.md](docs/README.md) | how to run Proven, its architecture and every setting |
| [docs/SAMPLE_MODE.md](docs/SAMPLE_MODE.md) | try Proven with no network and no keys: the sample review, its findings and its trace |
| [DESIGN.md](hackathon/DESIGN.md) | Proven's design, the sponsors' technology, evidence sources, open questions |
| [SAPHAN.md](hackathon/SAPHAN.md) | the platform that runs the fleet, its roles and vocabulary |
| [AGENTS.md](AGENTS.md) | the rules every agent follows in this repository |
| [planning/](hackathon/planning/) | how the topic was chosen, and the fleet's answers |
| [fleet/](hackathon/fleet/) | orders, reports, reviews, decisions and the ledger of every change, with their templates |

**Pre-existing components:** [Saphan Studio](https://saphan.ai/) runs the fleet; it is not part of this submission and not in this repository. The [Meterless engine specifications](https://github.com/Meterless/Meterless) (Apache-2.0) are open source; we build to them and do not copy their reference code. Model providers, tools and data sources are listed as they are used.

Before publication Marcin reviews every document. Credentials, internal host names and private infrastructure details are blacked out, and each document says where.

**Licence:** the code is Apache-2.0 ([LICENSE](LICENSE)); the documentation and the build record (the Markdown pages,
`docs/` and `hackathon/`) are CC BY 4.0 ([LICENSE-docs](LICENSE-docs)).
