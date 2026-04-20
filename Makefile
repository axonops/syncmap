SHELL := /bin/bash

GO            ?= go
GOFMT         ?= gofmt
GOIMPORTS     ?= goimports
GOLANGCI_LINT ?= golangci-lint
GOVULNCHECK   ?= govulncheck
GORELEASER    ?= goreleaser

PKGS          := ./...
COVERAGE_OUT  ?= coverage.out
COVERAGE_HTML ?= coverage.html

GOIMPORTS_INSTALL   := $(GO) install golang.org/x/tools/cmd/goimports@latest
GOLANGCI_INSTALL    := https://golangci-lint.run/usage/install/
GOVULNCHECK_INSTALL := $(GO) install golang.org/x/vuln/cmd/govulncheck@latest
GORELEASER_INSTALL  := $(GO) install github.com/goreleaser/goreleaser/v2@latest

.DEFAULT_GOAL := help

.PHONY: help
help: ## Print available targets
	@grep -hE '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| sort \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

.PHONY: check
check: fmt-check vet lint test tidy-check security ## Run the full quality gate

.PHONY: test
test: ## Run unit tests with -race
	$(GO) test -race -count=1 $(PKGS)

.PHONY: test-bdd
test-bdd: ## Run BDD tests (no-op until tests/bdd/ exists)
	@if [ -d tests/bdd ]; then \
		$(GO) test -race -count=1 -tags bdd ./tests/bdd/...; \
	else \
		echo "tests/bdd/ not present yet; skipping BDD run"; \
	fi

.PHONY: lint
lint: ## Run golangci-lint
	@command -v $(GOLANGCI_LINT) >/dev/null 2>&1 || { \
		echo "golangci-lint not found on PATH."; \
		echo "Install: see $(GOLANGCI_INSTALL)"; \
		exit 1; \
	}
	$(GOLANGCI_LINT) run $(PKGS)

.PHONY: vet
vet: ## Run go vet
	$(GO) vet $(PKGS)

.PHONY: fmt
fmt: ## Auto-fix gofmt and goimports
	$(GOFMT) -s -w .
	@command -v $(GOIMPORTS) >/dev/null 2>&1 || { \
		echo "goimports not found on PATH."; \
		echo "Install: $(GOIMPORTS_INSTALL)"; \
		exit 1; \
	}
	$(GOIMPORTS) -w -local github.com/axonops/syncmap .

.PHONY: fmt-check
fmt-check: ## Fail on any gofmt/goimports diff
	@out=$$( $(GOFMT) -s -l . ); \
	if [ -n "$$out" ]; then \
		echo "gofmt diff:"; echo "$$out"; exit 1; \
	fi
	@if command -v $(GOIMPORTS) >/dev/null 2>&1; then \
		out=$$( $(GOIMPORTS) -l -local github.com/axonops/syncmap . ); \
		if [ -n "$$out" ]; then \
			echo "goimports diff:"; echo "$$out"; exit 1; \
		fi; \
	else \
		echo "goimports not found on PATH (skipping); install with: $(GOIMPORTS_INSTALL)"; \
	fi

.PHONY: bench
bench: ## Run benchmarks with allocation reports
	$(GO) test -race -bench=. -benchmem -run='^$$' $(PKGS)

.PHONY: coverage
coverage: ## Generate and summarise a coverage report
	$(GO) test -race -covermode=atomic -coverprofile=$(COVERAGE_OUT) $(PKGS)
	$(GO) tool cover -func=$(COVERAGE_OUT) | tail -1
	$(GO) tool cover -html=$(COVERAGE_OUT) -o $(COVERAGE_HTML)
	@echo "HTML report: $(COVERAGE_HTML)"

.PHONY: tidy
tidy: ## Run go mod tidy
	$(GO) mod tidy

.PHONY: tidy-check
tidy-check: ## Fail if go mod tidy would change go.mod or go.sum
	@tmp=$$(mktemp -d); \
	cp go.mod $$tmp/go.mod; \
	if [ -f go.sum ]; then cp go.sum $$tmp/go.sum; fi; \
	$(GO) mod tidy; \
	status=0; \
	if ! cmp -s go.mod $$tmp/go.mod; then \
		echo "go mod tidy changed go.mod:"; diff -u $$tmp/go.mod go.mod || true; status=1; \
	fi; \
	if [ -f $$tmp/go.sum ] || [ -f go.sum ]; then \
		if ! cmp -s $$tmp/go.sum go.sum 2>/dev/null; then \
			echo "go mod tidy changed go.sum"; status=1; \
		fi; \
	fi; \
	cp $$tmp/go.mod go.mod; \
	if [ -f $$tmp/go.sum ]; then cp $$tmp/go.sum go.sum; fi; \
	rm -rf $$tmp; \
	exit $$status

.PHONY: security
security: ## Run govulncheck (skips cleanly if tool is absent)
	@if command -v $(GOVULNCHECK) >/dev/null 2>&1; then \
		$(GOVULNCHECK) $(PKGS); \
	else \
		echo "govulncheck not found on PATH (skipping)."; \
		echo "Install: $(GOVULNCHECK_INSTALL)"; \
	fi

.PHONY: release-check
release-check: ## Validate GoReleaser config (skips cleanly if tool or config absent)
	@if [ ! -f .goreleaser.yml ] && [ ! -f .goreleaser.yaml ]; then \
		echo ".goreleaser config not present yet (lands in #4); skipping"; \
		exit 0; \
	fi; \
	if command -v $(GORELEASER) >/dev/null 2>&1; then \
		$(GORELEASER) check; \
	else \
		echo "goreleaser not found on PATH (skipping)."; \
		echo "Install: $(GORELEASER_INSTALL)"; \
	fi

.PHONY: clean
clean: ## Remove coverage artefacts and clear the Go test cache
	$(GO) clean -testcache
	rm -f $(COVERAGE_OUT) $(COVERAGE_HTML)
