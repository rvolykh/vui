SHELL := /bin/bash
.DEFAULT_GOAL := help

# Variables
VERSION?=dev
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.gitCommit=$(GIT_COMMIT)"

# Allow to pass args to targets
cmd := $(firstword $(MAKECMDGOALS))
args := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
%::
	@ true
# (c) https://askubuntu.com/a/1448109

##@ Build targets

deps: ## Download and tidy dependencies
	@ go mod download
	@ go mod tidy
.PHONY: deps

fmt: ## Format source code
	@ go fmt ./...
.PHONY: fmt

vet: ## Examine source code
	@ go vet ./...
.PHONY: vet

build: ## Build the application
	@ go build $(LDFLAGS) -o vui ./cmd/vui
.PHONY: build

##@ Test targets

test: ## Run tests, e.g. make test, make test TestCoalesce
	@ go test -v $(if $(args), -run '$(args)') ./...
.PHONY: test

coverage: ## Run tests with coverage
	@ go test -coverprofile=coverage.out ./...
	@ go tool cover -html=coverage.out -o coverage.html
.PHONY: coverage

##@ Sandbox targets

sbx-build: ## Build sandbox init image(s)
	@ $(MAKE) -C sandbox build
.PHONY: sbx-build

sbx-up: ## Create sandbox, e.g. make sbx-up, make sbx-up vault
	@ $(MAKE) -C sandbox up $(args)
.PHONY: sbx-up

sbx-logs: ## Show logs for sandbox, e.g. make sbx-logs, make sbx-logs vault
	@ $(MAKE) -C sandbox logs $(args)
.PHONY: sbx-logs

sbx-ps: ## Show sandbox services
	@ $(MAKE) -C sandbox ps
.PHONY: sbx-ps

sbx-run: build
sbx-run: ## Run vui in sandbox
	@ $(MAKE) -C sandbox env
	@ echo "# Run the application"
	@ eval $$(cat ./sandbox/.env) && ./vui
.PHONY: sbx-run

sbx-down: ## Destroy sandbox
	@ $(MAKE) -C sandbox down
	@ $(MAKE) -C sandbox clean
.PHONY: sbx-down

##@ Other targets

clean: ## Clean temporary files
	@ go clean
	@ rm -f ./vui
	@ rm -f *.log
	@ rm -f coverage.*
.PHONY: clean

help: ## Show help message
	@awk 'BEGIN { \
		FS = ":.*##"; \
		printf "Usage:\n  make \033[36m<target>\033[0m\n" \
		} \
		/^[a-zA-Z0-9_-]+:.*?##/ { \
			desc = $$2; \
			if (match(desc, / e\.g\. /)) { \
				before = substr(desc, 1, RSTART - 1); \
				after = substr(desc, RSTART); \
				printf "  \033[36m%-20s\033[0m %s\033[32m%s\033[0m\n", $$1, before, after \
			} else { \
				printf "  \033[36m%-20s\033[0m %s\n", $$1, desc \
			} \
		} \
		/^##@/ { \
			printf "\n\033[1m%s\033[0m\n", substr($$0, 5) \
		}' $(MAKEFILE_LIST)
.PHONY: help
