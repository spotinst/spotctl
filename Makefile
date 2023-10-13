.DEFAULT_GOAL := help
.EXPORT_ALL_VARIABLES:

# Utilities.
V := 0
Q := $(if $(filter 1,$(V)),,@)
T := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Versioning.
GIT_DIRTY := $(shell git status --porcelain)
VERSION   := $(shell cat internal/version/VERSION)

# Shell.
# Setting SHELL to bash allows bash commands to be executed by recipes. Options
# are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

# Go.
GO := GO111MODULE=on go

##@ General

.PHONY: help
help: ## Display this help
	$(Q) awk 'BEGIN {FS = ":.*##"; printf "Usage: make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) }' $(MAKEFILE_LIST)

.PHONY: todo
todo: ## Show to-do items per file
	$(Q) grep \
		--exclude=Makefile \
		--exclude-dir=vendor \
		--exclude-dir=dist \
		--exclude-dir=.idea \
		--exclude-dir=.git \
		--text \
		--color \
		-nRo \
		-E 'TODO.*' \
		.

##@ Development

.PHONY: build
build: fmt ## Build all commands
	$(Q) $(GO) build -trimpath -race -o dist/spotctl cmd/spotctl/main.go
	cp dist/spotctl /home/pathcl/go/bin

.PHONY: test
test: fmt ## Run all tests
	$(Q) $(GO) test \
		-v $$($(GO) list ./... | grep -v vendor) $(TESTARGS) \
		-covermode=atomic \
		-coverprofile=dist/coverage.txt \
		-race \
		-timeout=30s \
		-parallel=4

.PHONY: cover
cover: test ## Run all tests and open the coverage report
	$(Q) $(GO) tool cover -html=dist/coverage.txt

.PHONY: tidy
tidy: ## Add missing and remove unused modules
	$(Q) $(GO) mod tidy

.PHONY: fmt
fmt: ## Format all .go files
	$(Q) $(GO) fmt ./...

.PHONY: vet
vet: ## Analyze all .go files
	$(Q) $(GO) vet ./...

.PHONY: clean
clean: ## Clean the generated artifacts
	$(Q) rm -rf dist

##@ Release

.PHONY: release
release: fmt ## Release a new version
ifneq ($(strip $(GIT_DIRTY)),)
	$(Q) echo "Git is currently in a dirty state. Please commit your changes or stash them before you release." ; exit 1
else
	$(Q) read -p "Release version: $(VERSION) →  " version ;\
		 echo $$version > internal/version/VERSION ;\
		 git commit -a -m "chore(release): v$$version" ;\
		 git tag -f -m    "chore(release): v$$version" v$$version ;\
		 git push --follow-tags
endif
