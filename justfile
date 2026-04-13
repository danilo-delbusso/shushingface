# shushing face — speech-to-text desktop app
# Dependencies: wtype (Wayland auto-paste) or xdotool (X11 auto-paste)

set shell := ["bash", "-c"]
set windows-shell := ["powershell.exe", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command"]

prefix := env("PREFIX", env("HOME", env("USERPROFILE", "")) / ".local")

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

VERSION := if os_family() == "windows" {
    `try { (git describe --tags --always --dirty 2>$null) } catch { 'dev' }`
} else {
    `git describe --tags --always --dirty 2>/dev/null || echo dev`
}
LDFLAGS := "-X codeberg.org/dbus/shushingface/internal/version.version=" + VERSION

# Runs the Wails Desktop application in development mode
[unix]
dev:
    wails dev -tags webkit2_41 -ldflags '{{LDFLAGS}}'

[windows]
dev:
    & scripts/build.ps1 dev

# Platform-aware build (auto-detects OS)
[unix]
build:
    wails build -tags webkit2_41 -ldflags '{{LDFLAGS}}'

[windows]
build:
    & scripts/build.ps1 build

# Build for Linux explicitly
build-linux:
    wails build -tags webkit2_41 -ldflags '{{LDFLAGS}}'

# Build for macOS explicitly
build-darwin:
    wails build -ldflags '{{LDFLAGS}}'

# Build for Windows explicitly (produces NSIS installer in build/bin)
build-windows:
    & scripts/build.ps1 build -nsis

# --- Install & Uninstall (Linux) ---
# macOS: `wails build` produces a .app bundle in build/bin/
# Windows: use `just install-windows` (copies exe + Start Menu shortcut)
#          or `just build-windows` to produce an NSIS installer

# Install shushingface on Linux (binary + desktop entry + icon + shortcut)
install: build
    #!/bin/bash
    set -e
    install -Dm755 build/bin/shushingface "{{prefix}}/bin/shushingface"
    install -Dm644 build/linux/shushingface.desktop "$HOME/.local/share/applications/shushingface.desktop"
    install -Dm644 build/appicon.png "$HOME/.local/share/icons/hicolor/512x512/apps/shushingface.png"
    gtk-update-icon-cache -f -t "$HOME/.local/share/icons/hicolor/" 2>/dev/null || true
    update-desktop-database "$HOME/.local/share/applications/" 2>/dev/null || true
    just _install-shortcut
    echo "Installed shushingface to {{prefix}}/bin/shushingface"

# Register Super+Ctrl+B shortcut in the current desktop environment
_install-shortcut:
    #!/bin/bash
    case "${XDG_CURRENT_DESKTOP:-}" in
      COSMIC)
        dir="$HOME/.config/cosmic/com.system76.CosmicSettings.Shortcuts/v1"
        file="$dir/custom"
        mkdir -p "$dir"
        if [ -f "$file" ] && grep -q "shushingface" "$file"; then
          echo "Shortcut already registered (COSMIC)"
        else
          if [ -f "$file" ] && grep -q "Spawn" "$file"; then
            sed -i 's/}$/    (\n        modifiers: [\n            Super,\n            Ctrl,\n        ],\n        key: "b",\n        description: Some("shushingface: toggle recording"),\n    ): Spawn("shushingface --toggle"),\n}/' "$file"
          else
            cat > "$file" << 'RON'
    {
        (
            modifiers: [
                Super,
                Ctrl,
            ],
            key: "b",
            description: Some("shushingface: toggle recording"),
        ): Spawn("shushingface --toggle"),
    }
    RON
          fi
          echo "Registered Super+Ctrl+B shortcut (COSMIC)"
          echo "Log out and back in, or restart cosmic-comp, for the shortcut to take effect"
        fi
        ;;
      GNOME*)
        path="/org/gnome/settings-daemon/plugins/media-keys/custom-keybindings/shushingface/"
        base="org.gnome.settings-daemon.plugins.media-keys.custom-keybinding:$path"
        gsettings set "$base" name "shushingface" 2>/dev/null &&
        gsettings set "$base" command "shushingface --toggle" 2>/dev/null &&
        gsettings set "$base" binding "<Super><Ctrl>b" 2>/dev/null &&
        existing=$(gsettings get org.gnome.settings-daemon.plugins.media-keys custom-keybindings 2>/dev/null || echo "[]")
        if ! echo "$existing" | grep -q "shushingface"; then
          new=$(echo "$existing" | sed "s/]/, '$path']/" | sed "s/\[, /[/")
          gsettings set org.gnome.settings-daemon.plugins.media-keys custom-keybindings "$new"
        fi &&
        echo "Registered Super+Ctrl+B shortcut (GNOME)" ||
        echo "Could not register GNOME shortcut (gsettings not available)"
        ;;
      *)
        echo "Tip: bind 'shushingface --toggle' to Super+Ctrl+B in your desktop settings"
        ;;
    esac

# Remove shushingface
uninstall:
    #!/bin/bash
    rm -f "{{prefix}}/bin/shushingface"
    rm -f "$HOME/.local/share/applications/shushingface.desktop"
    rm -f "$HOME/.local/share/icons/hicolor/512x512/apps/shushingface.png"
    update-desktop-database "$HOME/.local/share/applications/" 2>/dev/null || true
    # Remove COSMIC shortcut
    file="$HOME/.config/cosmic/com.system76.CosmicSettings.Shortcuts/v1/custom"
    if [ -f "$file" ] && grep -q "shushingface" "$file"; then
      # If it's the only entry, write an empty map
      echo "{}" > "$file"
      echo "Removed COSMIC shortcut"
    fi
    echo "Uninstalled shushingface"

# --- Install & Uninstall (Windows) ---

# Install shushingface on Windows (binary to LocalAppData + Start Menu shortcut)
install-windows: build
    & scripts/install-windows.ps1

# Remove shushingface on Windows
uninstall-windows:
    & scripts/uninstall-windows.ps1

# Re-generates TypeScript bindings from Go structs
bindings:
    wails generate module

# Generate THIRD_PARTY_LICENSES.md from Go + frontend deps + assets
licenses:
    ./scripts/check-licenses.sh --update

# Verify THIRD_PARTY_LICENSES.md is up to date (for CI)
licenses-check:
    ./scripts/check-licenses.sh
