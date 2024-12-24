# List all available commands
default:
    @just --list

# Run all tests
test:
    go test -v ./...

# Run tests with race detection
test-race:
    go test -v -race ./...

# Run tests with coverage report
test-coverage:
    go test -v -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html

# Run linter
lint:
    golangci-lint run

# Format code
fmt:
    go fmt ./...

# Tidy go modules
tidy:
    go mod tidy

# Run all CI checks
ci: fmt tidy lint test-race

# Build the binary
build:
    go build -o bin/outline-cli

# Clean build artifacts
clean:
    rm -rf bin coverage.out coverage.html

# Install dependencies
setup:
    go mod download
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest 
