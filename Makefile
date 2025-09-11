.PHONY: help dev test audit build release clean install deps fmt vet lint coverage check complexity deadcode vulncheck tidy

# Default target
help: ## Show this help message
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*##/ {printf "  %-12s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Development cycle targets
check: deps fmt vet lint test ## Run full development cycle (deps, fmt, vet, lint, test)

test: ## Run tests with coverage
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

build: ## Build the library (check compilation)
	go build ./...

release: audit test build ## Full release cycle (audit, test, build)
	@echo "Release checks passed!"

# Individual targets
deps: ## Download and verify dependencies
	go mod download
	go mod verify

install: ## Install development dependencies
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	@which gocyclo > /dev/null || (echo "Installing gocyclo..." && go install github.com/fzipp/gocyclo/cmd/gocyclo@latest)
	@which deadcode > /dev/null || (echo "Installing deadcode..." && go install golang.org/x/tools/cmd/deadcode@latest)
	@echo "Development dependencies installed"

fmt: ## Format all Go source files
	go fmt ./...

vet: ## Run go vet
	go vet ./...

lint: ## Run golangci-lint
	@which golangci-lint > /dev/null || (echo "golangci-lint not found, run 'make install'" && exit 1)
	golangci-lint run

tidy: ## Run go mod tidy to clean up dependencies
	go mod tidy

complexity: ## Check code complexity
	@which gocyclo > /dev/null || (echo "gocyclo not found, run 'make install'" && exit 1)
	@echo "Checking cyclomatic complexity (>10 is flagged):"
	@gocyclo -over 10 . || true

deadcode: ## Check for dead code
	@which deadcode > /dev/null || (echo "deadcode not found, run 'make install'" && exit 1)
	@echo "Checking for dead code:"
	@deadcode -test ./... || true

vulncheck: ## Check for known vulnerabilities
	@echo "Checking for known vulnerabilities:"
	@go run golang.org/x/vuln/cmd/govulncheck@latest ./...

audit: fmt tidy vet lint complexity deadcode vulncheck test ## Run comprehensive code quality and security audit

coverage: test ## Generate and view test coverage report
	@echo "Coverage report generated: coverage.html"
	@which open > /dev/null && open coverage.html || echo "Open coverage.html in your browser"

clean: ## Clean build artifacts and temporary files
	go clean -testcache
	rm -f coverage.out coverage.html
	rm -rf dist/ build/

# Benchmarks
bench: ## Run benchmarks
	go test -bench=. -benchmem ./...

# Examples
examples: ## Run all examples
	@find examples -name "main.go" -exec go run {} \;

# Documentation
docs: ## Generate and serve documentation
	@echo "Go documentation server starting at http://localhost:6060"
	godoc -http=:6060

# Git hooks setup
hooks: ## Setup git pre-commit hooks
	@echo "#!/bin/sh" > .git/hooks/pre-commit
	@echo "make check" >> .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "Git pre-commit hook installed (runs 'make check')"

# Development helpers
watch: ## Watch for changes and run tests
	@which fswatch > /dev/null || (echo "fswatch not found. Install with: brew install fswatch" && exit 1)
	fswatch -o . | xargs -n1 -I{} make test

init: install hooks ## Initialize development environment
	@echo "Development environment initialized!"
