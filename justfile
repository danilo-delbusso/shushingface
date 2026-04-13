# shushing face — speech-to-text desktop app
# Dependencies: wtype (Wayland auto-paste) or xdotool (X11 auto-paste)

set shell := ["bash", "-c"]
set windows-shell := ["powershell.exe", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command"]

default:
    @just --list

# --- Dev environment ---

# Report missing / outdated build and runtime dependencies
[unix]
doctor:
    @bash scripts/doctor/linux.sh

[windows]
doctor:
    @bash scripts/doctor/windows.sh

# Install missing dependencies (pass --yes to skip prompts)
[unix]
bootstrap *args:
    @bash scripts/bootstrap/linux.sh {{args}}

[windows]
bootstrap *args:
    @powershell.exe -NoProfile -ExecutionPolicy Bypass -File scripts/bootstrap/windows.ps1 {{args}}

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
    golangci-lint run --build-tags webkit2_41

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

# Runs the Wails Desktop application in development mode
[unix]
dev:
    @bash scripts/build/linux.sh dev

[windows]
dev:
    @powershell.exe -NoProfile -ExecutionPolicy Bypass -File scripts/build/windows.ps1 dev

# Platform-aware build (auto-detects OS)
[unix]
build:
    @bash scripts/build/linux.sh build

[windows]
build:
    @powershell.exe -NoProfile -ExecutionPolicy Bypass -File scripts/build/windows.ps1 build

# Produce a Windows NSIS installer in build/bin
[windows]
build-windows:
    @powershell.exe -NoProfile -ExecutionPolicy Bypass -File scripts/build/windows.ps1 build -nsis

# --- Install & Uninstall ---
# Binary goes to $PREFIX/bin ($HOME/.local/bin by default on both OSes).

# Install shushingface for the current user (builds first)
[unix]
install: build
    @bash scripts/install/linux.sh

[windows]
install: build
    @powershell.exe -NoProfile -ExecutionPolicy Bypass -File scripts/install/windows.ps1

# Uninstall shushingface
[unix]
uninstall:
    @bash scripts/uninstall/linux.sh

[windows]
uninstall:
    @powershell.exe -NoProfile -ExecutionPolicy Bypass -File scripts/uninstall/windows.ps1

# Muscle-memory aliases for the old per-OS recipes
alias install-windows   := install
alias uninstall-windows := uninstall

# Produce installable artifact(s) for the current OS (tar.gz + .deb on Linux, NSIS on Windows)
[unix]
package:
    @bash scripts/package/linux.sh

[windows]
package:
    @powershell.exe -NoProfile -ExecutionPolicy Bypass -File scripts/package/windows.ps1

# Re-generates TypeScript bindings from Go structs
bindings:
    wails generate module

# Generate THIRD_PARTY_LICENSES.md from Go + frontend deps + assets
licenses:
    ./scripts/release/licenses.sh --update

# Verify THIRD_PARTY_LICENSES.md is up to date (for CI)
licenses-check:
    ./scripts/release/licenses.sh
