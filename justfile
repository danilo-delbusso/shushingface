# Sussurro Workflow Integration

set shell := ["bash", "-c"]

default:
    @just --list

# --- Frontend Tasks ---

# Formats frontend code using Biome
format-ui:
    cd frontend && bun run biome format --write .

# Lints frontend code using Biome
lint-ui:
    cd frontend && bun run biome lint .

# Runs UI tests using Vitest
test-ui:
    cd frontend && bun run vitest run

# --- Backend Tasks ---

# Formats Go code
format-go:
    go fmt ./...

# Lints Go code (requires golangci-lint)
lint-go:
    golangci-lint run

# Runs backend tests
test-go:
    go test ./...

# --- Combined Tasks ---

# Run all formatters
format: format-go format-ui

# Run all linters
lint: lint-go lint-ui

# Run all tests
test: test-go test-ui

# Checks the entire project (linting + testing)
check: lint test

# --- Build & Run ---

# Runs the TUI interface
run-tui:
    go run cmd/sussurro/main.go

# Runs the Wails Desktop application in development mode
dev:
    wails dev

# Builds the Wails Desktop application binary
build:
    wails build

# Re-generates TypeScript bindings from Go structs
bindings:
    wails generate module
