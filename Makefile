SHELL := /bin/sh
.DEFAULT_GOAL := help
ROOT := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
IMAGE ?= proven:local
REVISION := $(shell git rev-parse HEAD 2>/dev/null || echo unknown)

.PHONY: help build build-all test vet sample image image-check scan publish-image check-links clean

help: ## Show available development commands
	@awk 'BEGIN { FS = ":.*## " } /^[a-z-]+:.*## / { printf "%-16s %s\n", $$1, $$2 }' "$(ROOT)/Makefile"

build: ## Build the CLI into bin/
	@mkdir -p "$(ROOT)/bin"
	@cd "$(ROOT)" && go build -trimpath -o bin/proven ./cmd/proven

build-all: ## Build macOS and Linux binaries for Intel and ARM
	@sh "$(ROOT)/scripts/build-all.sh"

test: ## Run the test suite without external accounts
	@cd "$(ROOT)" && go test ./...

vet: ## Check Go code
	@cd "$(ROOT)" && go vet ./...

sample: build ## Run a recorded example without keys or Docker
	@sh "$(ROOT)/scripts/sample.sh"

image: ## Build the product container locally
	@docker build --label org.opencontainers.image.revision=$(REVISION) -t "$(IMAGE)" "$(ROOT)"

image-check: build image ## Compare the container and native sample review
	@sh "$(ROOT)/scripts/image-check.sh" "$(IMAGE)" "$(ROOT)/bin/proven"

scan: ## Scan a local image with Trivy
	@trivy image --scanners vuln --severity HIGH,CRITICAL --exit-code 1 "$(IMAGE)"

publish-image: ## Build both Linux architectures, test, scan and publish only on success
	@BUILDX_BUILDER="$(BUILDX_BUILDER)" sh "$(ROOT)/scripts/publish-image.sh" "$(IMAGE)"

check-links: ## Check documentation links
	@cd "$(ROOT)" && sh scripts/check-links.sh

clean: ## Remove local build output
	@rm -rf "$(ROOT)/bin" "$(ROOT)/dist"
