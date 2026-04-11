# Sussurro Workflow Integration

set shell := ["bash", "-c"]

prefix := env("PREFIX", "/usr/local")

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
    #!/bin/bash
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        wails dev -tags webkit2_41
    else
        wails dev
    fi

# Platform-aware build (auto-detects OS)
build:
    #!/bin/bash
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        wails build -tags webkit2_41
    else
        wails build
    fi

# Build for Linux explicitly
build-linux:
    wails build -tags webkit2_41

# Build for Linux with system tray support (requires libayatana-appindicator3-dev)
build-linux-tray:
    wails build -tags "webkit2_41 systray"

# Build for macOS explicitly
build-darwin:
    wails build

# --- Install & Uninstall ---

# Install sussurro system-wide (binary + desktop entry + icon)
install: build
    #!/bin/bash
    set -e
    install -Dm755 build/bin/sussurro "{{prefix}}/bin/sussurro"
    install -Dm644 build/linux/sussurro.desktop "$HOME/.local/share/applications/sussurro.desktop"
    install -Dm644 build/appicon.png "$HOME/.local/share/icons/hicolor/512x512/apps/sussurro.png"
    update-desktop-database "$HOME/.local/share/applications/" 2>/dev/null || true
    echo "Installed sussurro to {{prefix}}/bin/sussurro"
    echo "Tip: bind 'sussurro --toggle' to a keyboard shortcut in your desktop settings"

# Remove sussurro
uninstall:
    #!/bin/bash
    rm -f "{{prefix}}/bin/sussurro"
    rm -f "$HOME/.local/share/applications/sussurro.desktop"
    rm -f "$HOME/.local/share/icons/hicolor/512x512/apps/sussurro.png"
    update-desktop-database "$HOME/.local/share/applications/" 2>/dev/null || true
    echo "Uninstalled sussurro"

# Re-generates TypeScript bindings from Go structs
bindings:
    wails generate module
