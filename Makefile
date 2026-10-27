# One entry point for Proven and the judges' lab (ADR-0026).
# GNU make. Settings live in lab/versions.env.

SHELL := /bin/sh
.DEFAULT_GOAL := help

ROOT := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
export ROOT

.PHONY: help check-flow password demo-check reports try nightshift-receipts check-banner nightshift zetaris check-golive build build-all image image-check seed-image scan push up down status logs sample livecheck e2e trace e2e-failed-run check-e2e-parse check-lab-gates check-lab-secrets check-lab-password check-review-gates baseline-dependabot baseline-dependabot-manifests check-links zetaris-usl-parity clean

help: ## Show every target, the version, and whether go and docker were found
	@sh $(ROOT)/lab/scripts/help.sh

build: ## Build ./bin/proven and ./bin/livecheck
	@sh $(ROOT)/lab/scripts/build.sh

build-all: ## Cross-compile darwin and linux, amd64 and arm64, into ./dist
	@sh $(ROOT)/lab/scripts/build-all.sh

image: ## Build Proven's image for this machine and load it (one platform; a local engine cannot load a manifest list)
	@sh $(ROOT)/lab/scripts/image.sh

image-check: build image ## Check that the image prints the same version and sample review as ./bin/proven, and print its size
	@sh $(ROOT)/lab/scripts/image-check.sh

seed-image: ## Build the lab seed for this machine and load it (one platform; make push publishes both)
	@sh $(ROOT)/lab/scripts/seed-image.sh

scan: ## Scan our images with Docker Scout; refuses when Scout is missing, fails, or finds a fixable critical or high
	@sh $(ROOT)/lab/scripts/scan.sh

push: ## Build once for linux/amd64 and linux/arm64, scan every platform with Docker Scout, then publish those scanned digests as the final tags (refuses, publishing nothing, on any scan failure)
	@sh $(ROOT)/lab/scripts/push.sh

up: ## Start the judges' lab. On a terminal it asks sample or live, and for a judge, ours or own; settings answer in advance (docs/LAB.md); with no terminal it is a sample lab
	@sh $(ROOT)/lab/scripts/up.sh

down: ## Stop the judges' lab
	@sh $(ROOT)/lab/scripts/down.sh

status: ## Show containers, health, the runner, the review on each pull request, and the lab login
	@sh $(ROOT)/lab/scripts/status.sh

password: ## Print the lab login: the start page, the Gitea user and password (on a terminal), the dashboard
	@sh $(ROOT)/lab/scripts/password.sh

demo-check: ## Require Proven's review on every demo pull request (a clean one says clean) and a trace per review
	@sh $(ROOT)/lab/scripts/demo-check.sh

reports: ## Build the report page of every demo pull request again: its review as a page set like a GitHub comment
	@sh $(ROOT)/lab/scripts/reports.sh

nightshift-receipts: ## Load the public Nightshift receipts (lab/nightshift/receipts.public.jsonl) into the lab's Loki, for the Nightshift dashboard
	@sh $(ROOT)/lab/scripts/nightshift-receipts.sh

try: ## Try one thing Proven does and see it print: SCENARIO=forged (a changed evidence pack is refused) or SCENARIO=prover (no network, no secret)
	@sh $(ROOT)/lab/scripts/try.sh "$(SCENARIO)"

check-banner: ## Check the closing banner offline: the exact lines, the password on a terminal, the path when redirected
	@sh $(ROOT)/lab/scripts/banner-check.sh

logs: ## Follow lab logs. SERVICE= selects one service
	@sh $(ROOT)/lab/scripts/logs.sh

sample: ## Run a sample review outside the repository, with no keys and no Docker
	@sh $(ROOT)/lab/scripts/sample.sh

livecheck: ## Name which live services are ready. Reads lab/.env when present
	@sh $(ROOT)/lab/scripts/livecheck.sh

e2e: ## Start the lab and require Proven's review on the pull request
	@sh $(ROOT)/lab/scripts/e2e.sh

trace: ## Ask the dashboard's trace store for the latest review's trace (spans and the two reviewers)
	@sh $(ROOT)/lab/scripts/trace.sh

e2e-failed-run: ## Negative control: with an old review comment present and the current run failing, e2e must refuse
	@sh $(ROOT)/lab/scripts/e2e-failed-run.sh

check-e2e-parse: ## Check e2e's parsers offline against statuses a live Gitea served
	@sh $(ROOT)/lab/scripts/e2e-parse-check.sh

zetaris-usl-parity: ## Ask Zetaris the same requests with and without the semantic layer and require equal rows (skipped by name when the Zetaris settings are not set)
	@cd $(ROOT) && go test -count=1 -run 'TestParityLive$$' -v ./internal/usl/

check-lab-gates: ## Check offline, with a stub docker, that make up refuses the retired LAB_BIND and that make push publishes only digests that were all scanned
	@sh $(ROOT)/lab/scripts/check-lab-gates.sh

check-flow: ## Check offline the start-up flow of make up: each path by its settings with no terminal, the no-terminal default, the port check, where the credential goes, and the terminal after Ctrl-C
	@cd $(ROOT) && go test -count=1 -run 'TestFlow' -v ./lab/

check-lab-secrets: ## Live, on a Docker host after make up with a sentinel in each credential: no sentinel in compose config, inspect or process arguments; each is an Actions secret; the transfer file is gone
	@sh $(ROOT)/lab/scripts/check-lab-secrets.sh

check-lab-password: ## Check offline, with a stub docker, that the admin password is never printed to a redirected stdout, is kept in a 0600 file, and a supplied one is used and not printed
	@sh $(ROOT)/lab/scripts/check-lab-password.sh

check-review-gates: ## Check offline five properties of the lab: loopback in the compose file, password handling, Scout failure, push after scan, build context
	@sh $(ROOT)/lab/scripts/check-review-gates.sh

baseline-dependabot: ## Run Proven's dependency check at the commit Dependabot's 51 alerts are pinned to, on the store PROVEN_EVIDENCE_STORE names, and write test/corpus/reference/dependabot-fbdf7d3/REPORT.md
	@sh $(ROOT)/lab/scripts/baseline-dependabot.sh evidence

baseline-dependabot-manifests: ## The same run with no evidence store: which alerted packages Proven reads from the manifests at the pinned commit
	@sh $(ROOT)/lab/scripts/baseline-dependabot.sh manifests

check-links: ## Check that every relative Markdown link and metadata.json path resolves
	@cd $(ROOT) && sh scripts/check-links.sh

clean: ## Remove the lab, its volumes, local images, ./bin and ./dist
	@sh $(ROOT)/lab/scripts/clean.sh

nightshift: ## Start Meterless Nightshift in a Tart VM (macOS on Apple silicon); on any other host say so and exit 0
	@sh $(ROOT)/lab/scripts/nightshift.sh

zetaris: ## Start Zetaris on this machine from your own bundle and register the evidence (same as make up ZETARIS=local, without the lab)
	@ZETARIS=local sh $(ROOT)/lab/scripts/zetaris.sh

check-golive: ## Check offline: lab/.env keys, the workflow's live/sample choice and the loopback bind of the start page
	@sh $(ROOT)/lab/scripts/golive-check.sh
