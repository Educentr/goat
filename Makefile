.PHONY: help build test coverage goat install-lint lint lint-full lint-fix fmt vet clean

# Variables
GOLANGCI_LINT_VERSION ?= v1.62.0
TEST_TIMEOUT ?= 300s
GO ?= go
GOLANGCI_LINT ?= golangci-lint

# Default target
.DEFAULT_GOAL := help

## help: Show this help message
help:
	@echo 'Usage:'
	@echo '  make <target>'
	@echo ''
	@echo 'Targets:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## build: Build the project (verify compilation)
build:
	@echo "Building project..."
	$(GO) build -v ./...

## test: Run unit tests (excludes services/* integration tests)
test:
	@echo "Running unit tests (timeout: $(TEST_TIMEOUT))..."
	$(GO) test -v -timeout=$(TEST_TIMEOUT) -race ./... -run '.*' -skip 'Test.*_Integration'

## coverage: Run tests with coverage report
coverage:
	@echo "Running tests with coverage..."
	$(GO) test -v -timeout=$(TEST_TIMEOUT) -race -coverprofile=coverage.out -covermode=atomic ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## goat: Run integration tests with Docker (services/* only)
goat:
	@echo "Running integration tests with Docker..."
	$(GO) test -v -timeout=$(TEST_TIMEOUT) -race ./services/... ./...

## install-lint: Install golangci-lint
install-lint:
	@echo "Installing golangci-lint $(GOLANGCI_LINT_VERSION)..."
	@which $(GOLANGCI_LINT) > /dev/null 2>&1 || \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \
		sh -s -- -b $$(go env GOPATH)/bin $(GOLANGCI_LINT_VERSION)
	@echo "Installed: $$($(GOLANGCI_LINT) --version)"

## lint: Run linter on changed files (diff with origin/main)
lint:
	@echo "Running linter on changed files..."
	@if git rev-parse --verify origin/main >/dev/null 2>&1; then \
		$(GOLANGCI_LINT) run --new-from-rev=origin/main; \
	else \
		echo "Warning: origin/main not found, running lint-full instead"; \
		$(MAKE) lint-full; \
	fi

## lint-full: Run linter on entire codebase
lint-full:
	@echo "Running linter on entire codebase..."
	$(GOLANGCI_LINT) run ./...

## lint-fix: Auto-fix linting issues
lint-fix:
	@echo "Auto-fixing linting issues..."
	$(GOLANGCI_LINT) run --fix ./...

## fmt: Format code with go fmt
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...

## vet: Run go vet
vet:
	@echo "Running go vet..."
	$(GO) vet ./...

## clean: Clean build artifacts and coverage reports
clean:
	@echo "Cleaning build artifacts..."
	$(GO) clean -cache -testcache -modcache
	rm -f coverage.out coverage.html
	@echo "Clean complete"

## tidy: Tidy and verify go.mod
tidy:
	@echo "Tidying go.mod..."
	$(GO) mod tidy
	$(GO) mod verify

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GO) mod download
