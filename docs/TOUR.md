# A tour of Proven, in pictures

What you see when you run the judges' lab on your own machine (`make up`, see [LAB.md](LAB.md)). Every picture was taken from
a lab started in sample mode: recorded answers, no keys, nothing leaves the machine. Click a picture to open it at full size.

## The start page

The lab ends `make up` by pointing you here. It holds the login, links to the pull requests, the dashboards and these
documents, six scenarios you can try yourself, and the code and image security reviews.

[![The lab's start page: the pull requests, the dashboards and the documentation, then the scenarios to try](images/tour/lab.png)](images/tour/lab.png)

## A review, as a page

Each demo pull request has its review as a page, set the way GitHub shows Proven's comment: the verdict on top, the
findings table, and each finding's evidence folded under it.

A change with defects: blocked, with a test that fails on the change and passes on the base, and the advisory rows the
dependency check matched.

[![The review of a change that applies a discount twice: blocked, with its findings and their evidence](images/tour/report-defect.png)](images/tour/report-defect.png)

A clean change: a short verdict that says what was checked. No false alarm.

[![The review of a clean change: a clean verdict and what was checked](images/tour/report-clean.png)](images/tour/report-clean.png)

## The dashboards

Six Grafana dashboards, each opening with one sentence you can read in three seconds.

The overview: every pull request, how its review ended, and how well each kept finding is proved.

[![The overview dashboard: 5 of 7 pull requests blocked, 7 of 7 reviews finished, no false alarm on the clean changes](images/tour/dashboard-overview.png)](images/tour/dashboard-overview.png)

One review, step by step: each agent in order, how long it took, and how the two reviewers judged each finding.

[![The one-review dashboard: the agents of one review in order, and the two reviewers' verdicts](images/tour/dashboard-one-review.png)](images/tour/dashboard-one-review.png)

The evidence behind each finding: what the Investigator queried and what the Prover ran.

[![The evidence dashboard: findings kept and dropped, and the Investigator and Prover runs](images/tour/dashboard-evidence.png)](images/tour/dashboard-evidence.png)

Which models answered, and what the guardrail checked.

[![The models dashboard: calls by agent and reviewer, the guardrail's checks, and what Proven refused to keep](images/tour/dashboard-models.png)](images/tour/dashboard-models.png)

The four Meterless engines and what each did.

[![The engines dashboard: what H-MEM, Scout Intent, Markovian and the World Model did in each review](images/tour/dashboard-engines.png)](images/tour/dashboard-engines.png)

The model gateway, read from Nightshift's signed receipts. In a sample lab no receipt is loaded, and the panels say so.

[![The Nightshift dashboard: calls and receipts, here with none loaded yet](images/tour/dashboard-nightshift.png)](images/tour/dashboard-nightshift.png)

## The image

The lab runs Proven's published container image. Its Docker Scout result, software bill of materials and build provenance
are on [Docker Hub](https://hub.docker.com/r/saphan/proven/tags).

The tag the judges pull, for both architectures: no critical, high, medium or low vulnerability, one left unrated.

[![Docker Hub: the judging tag of saphan/proven for amd64 and arm64, with no critical, high, medium or low vulnerability](images/tour/docker-scout.png)](images/tour/docker-scout.png)
