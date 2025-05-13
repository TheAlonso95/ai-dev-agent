# AI Dev Agent Makefile

.PHONY: build clean test lint run fmt help dev install-tools init

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GORUN=$(GOCMD) run
GOFMT=$(GOCMD) fmt
GOLINT=golangci-lint

# Binary name
BINARY_NAME=aiagent
MAIN_PATH=./main.go

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	$(GOBUILD) -o $(BINARY_NAME) $(MAIN_PATH)

# Run the application
run:
	@echo "Running $(BINARY_NAME)..."
	$(GORUN) $(MAIN_PATH) $(ARGS)

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -cover -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) ./...

# Run linter
lint:
	@echo "Running linter..."
	$(GOLINT) run

# Build and run
dev: build
	@echo "Running $(BINARY_NAME) in dev mode..."
	./$(BINARY_NAME) $(ARGS)

# Install necessary development tools
install-tools:
	@echo "Installing development tools..."
	$(GOGET) -u github.com/golangci/golangci-lint/cmd/golangci-lint
	$(GOGET) -u golang.org/x/tools/cmd/goimports

# Initialize a new project using AI Dev Agent
init:
	@echo "Initializing a new project..."
	$(GORUN) $(MAIN_PATH) init $(ARGS)

# Show help
help:
	@echo "Make targets:"
	@echo "  build          - Build the application"
	@echo "  run            - Run the application (use ARGS=\"your args\" to pass arguments)"
	@echo "  clean          - Clean build artifacts"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  fmt            - Format code"
	@echo "  lint           - Run linter"
	@echo "  dev            - Build and run the application"
	@echo "  install-tools  - Install development tools"
	@echo "  init           - Initialize a new project using AI Dev Agent (use ARGS=\"your idea\" to pass arguments)"
	@echo "  help           - Show this help message"