# One entry point for Proven and the judges' lab (ADR-0026).
# GNU make. Settings live in lab/versions.env.

SHELL := /bin/sh
.DEFAULT_GOAL := help

ROOT := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
export ROOT

.PHONY: help build build-all image seed-image scan push up down status logs sample livecheck e2e e2e-failed-run check-e2e-parse check-lab-gates check-lab-password check-review-gates check-links clean

help: ## Show every target, the version, and whether go and docker were found
	@sh $(ROOT)/lab/scripts/help.sh

build: ## Build ./bin/proven and ./bin/livecheck
	@sh $(ROOT)/lab/scripts/build.sh

build-all: ## Cross-compile darwin and linux, amd64 and arm64, into ./dist
	@sh $(ROOT)/lab/scripts/build-all.sh

image: ## Build Proven's image for this machine and load it (one platform; a local engine cannot load a manifest list)
	@sh $(ROOT)/lab/scripts/image.sh

seed-image: ## Build the lab seed for this machine and load it (one platform; make push publishes both)
	@sh $(ROOT)/lab/scripts/seed-image.sh

scan: ## Scan our images with Docker Scout; refuses when Scout is missing, fails, or finds a fixable critical or high
	@sh $(ROOT)/lab/scripts/scan.sh

push: ## Scan first (refuses on a Scout failure), then push linux/amd64 and linux/arm64, one tag per image, with provenance and an SBOM
	@sh $(ROOT)/lab/scripts/push.sh

up: ## Start the judges' lab
	@sh $(ROOT)/lab/scripts/up.sh

down: ## Stop the judges' lab
	@sh $(ROOT)/lab/scripts/down.sh

status: ## Show containers, health, the runner, the review, and the lab URL
	@sh $(ROOT)/lab/scripts/status.sh

logs: ## Follow lab logs. SERVICE= selects one service
	@sh $(ROOT)/lab/scripts/logs.sh

sample: ## Run a sample review outside the repository, with no keys and no Docker
	@sh $(ROOT)/lab/scripts/sample.sh

livecheck: ## Name which live services are ready. Reads lab/.env when present
	@sh $(ROOT)/lab/scripts/livecheck.sh

e2e: ## Start the lab and require Proven's review on the pull request
	@sh $(ROOT)/lab/scripts/e2e.sh

e2e-failed-run: ## Negative control: with an old review comment present and the current run failing, e2e must refuse
	@sh $(ROOT)/lab/scripts/e2e-failed-run.sh

check-e2e-parse: ## Check e2e's parsers offline against statuses a live Gitea served
	@sh $(ROOT)/lab/scripts/e2e-parse-check.sh

check-lab-gates: ## Check offline, with a stub docker, that make up refuses a non-loopback bind and make push refuses without a green scan
	@sh $(ROOT)/lab/scripts/check-lab-gates.sh

check-lab-password: ## Check offline, with a stub docker, that the admin password is never printed to a redirected stdout, is kept in a 0600 file, and a supplied one is used and not printed
	@sh $(ROOT)/lab/scripts/check-lab-password.sh

check-review-gates: ## Check offline five properties of the lab: loopback bind, password handling, Scout failure, push after scan, build context
	@sh $(ROOT)/lab/scripts/check-review-gates.sh

check-links: ## Check that every relative Markdown link and metadata.json path resolves
	@cd $(ROOT) && sh scripts/check-links.sh

clean: ## Remove the lab, its volumes, local images, ./bin and ./dist
	@sh $(ROOT)/lab/scripts/clean.sh
