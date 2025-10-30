.PHONY: help dev test audit build release clean install deps fmt vet lint coverage check complexity deadcode vulncheck tidy

# Default target
help: ## Show this help message
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*##/ {printf "  %-12s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Development cycle targets
check: deps fmt vet lint test ## Run full development cycle (deps, fmt, vet, lint, test)

test: ## Run tests with coverage
	go test -v -race -coverpkg=./pkg/... -coverprofile=coverage.out ./...
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
bench: ## Run basic benchmarks
	go test -bench=. -benchmem ./...

bench-full: ## Run comprehensive benchmark analysis
	@echo "Running comprehensive benchmark analysis..."
	go test -bench=. -benchmem -count=3 ./tests

bench-compare: ## Run comparison benchmarks (gopantic vs standard library)
	go test -bench="Benchmark.*Parsing.*" -benchmem -count=3 ./tests

bench-memory: ## Run memory allocation benchmarks
	go test -bench="BenchmarkMemory" -benchmem -count=3 ./tests

bench-cache: ## Run cache performance benchmarks
	go test -bench="Benchmark.*Cached" -benchmem -count=3 ./tests

bench-concurrent: ## Run concurrent processing benchmarks
	go test -bench="BenchmarkTypeCoercion" -benchmem -count=3 -cpu=1,2,4 ./tests

bench-profile: ## Run benchmarks with CPU and memory profiling
	@mkdir -p benchmark_results
	go test -bench=. -benchmem -cpuprofile=benchmark_results/cpu.prof -memprofile=benchmark_results/mem.prof ./tests
	@echo "Profiles generated in benchmark_results/"
	@echo "View CPU profile: go tool pprof benchmark_results/cpu.prof"
	@echo "View memory profile: go tool pprof benchmark_results/mem.prof"

# Examples
examples: ## Run all examples
	@find examples -name "main.go" -exec go run {} \;

# Documentation
docs: ## Generate and serve documentation
	@echo "Go documentation server starting at http://localhost:6060"
	godoc -http=:6060

# Git hooks setup
hooks: ## Setup git pre-commit hooks
	@echo "Installing enhanced pre-commit hook..."
	@cp .git/hooks/pre-commit .git/hooks/pre-commit.backup 2>/dev/null || true
	@chmod +x .git/hooks/pre-commit
	@echo "Pre-commit hook installed successfully!"
	@echo "Hook performs: security check, formatting, vetting, linting, and conditional testing"

# Development helpers
watch: ## Watch for changes and run tests
	@which fswatch > /dev/null || (echo "fswatch not found. Install with: brew install fswatch" && exit 1)
	fswatch -o . | xargs -n1 -I{} make test

init: install hooks ## Initialize development environment
	@echo "Development environment initialized!"
