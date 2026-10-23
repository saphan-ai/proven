# One entry point for Proven and the judges' lab (ADR-0026).
# GNU make. Settings live in lab/versions.env.

SHELL := /bin/sh
.DEFAULT_GOAL := help

ROOT := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
export ROOT

.PHONY: help build build-all image seed-image scan push up down status logs sample livecheck e2e clean

help: ## Show every target, the version, and whether go and docker were found
	@sh $(ROOT)/lab/scripts/help.sh

build: ## Build ./bin/proven and ./bin/livecheck
	@sh $(ROOT)/lab/scripts/build.sh

build-all: ## Cross-compile darwin and linux, amd64 and arm64, into ./dist
	@sh $(ROOT)/lab/scripts/build-all.sh

image: ## Build Proven's image for this machine and load it
	@sh $(ROOT)/lab/scripts/image.sh

seed-image: ## Build the one-shot lab seed image
	@sh $(ROOT)/lab/scripts/seed-image.sh

scan: ## Scout the images we build, and record third-party images in docs/LAB.md
	@sh $(ROOT)/lab/scripts/scan.sh

push: ## Scout, then push linux/amd64 and linux/arm64 with provenance and an SBOM
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

clean: ## Remove the lab, its volumes, local images, ./bin and ./dist
	@sh $(ROOT)/lab/scripts/clean.sh
