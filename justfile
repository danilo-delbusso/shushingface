# Sussurro Workflow Integration

set shell := ["bash", "-c"]

prefix := env("PREFIX", env("HOME") / ".local")

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

# Install sussurro (binary + desktop entry + icon + shortcut). Defaults to ~/.local
install: build
    #!/bin/bash
    set -e
    install -Dm755 build/bin/sussurro "{{prefix}}/bin/sussurro"
    install -Dm644 build/linux/sussurro.desktop "$HOME/.local/share/applications/sussurro.desktop"
    install -Dm644 build/appicon.png "$HOME/.local/share/icons/hicolor/512x512/apps/sussurro.png"
    update-desktop-database "$HOME/.local/share/applications/" 2>/dev/null || true
    just _install-shortcut
    echo "Installed sussurro to {{prefix}}/bin/sussurro"

# Register Ctrl+Shift+R shortcut in the current desktop environment
_install-shortcut:
    #!/bin/bash
    case "${XDG_CURRENT_DESKTOP:-}" in
      COSMIC)
        dir="$HOME/.config/cosmic/com.system76.CosmicSettings.Shortcuts/v1"
        file="$dir/custom"
        mkdir -p "$dir"
        if [ -f "$file" ] && grep -q "sussurro" "$file"; then
          echo "Shortcut already registered (COSMIC)"
        else
          # Write a fresh custom shortcuts file with our entry
          # If user had other custom shortcuts, we merge
          if [ -f "$file" ] && grep -q "Spawn" "$file"; then
            # Append our entry before the closing brace
            sed -i 's/}$/    (\n        modifiers: [\n            Ctrl,\n            Shift,\n        ],\n        key: "r",\n        description: Some("Sussurro: toggle recording"),\n    ): Spawn("sussurro --toggle"),\n}/' "$file"
          else
            cat > "$file" << 'RON'
    {
        (
            modifiers: [
                Ctrl,
                Shift,
            ],
            key: "r",
            description: Some("Sussurro: toggle recording"),
        ): Spawn("sussurro --toggle"),
    }
    RON
          fi
          echo "Registered Ctrl+Shift+R shortcut (COSMIC)"
          echo "Log out and back in, or restart cosmic-comp, for the shortcut to take effect"
        fi
        ;;
      GNOME*)
        path="/org/gnome/settings-daemon/plugins/media-keys/custom-keybindings/sussurro/"
        base="org.gnome.settings-daemon.plugins.media-keys.custom-keybinding:$path"
        gsettings set "$base" name "Sussurro" 2>/dev/null &&
        gsettings set "$base" command "sussurro --toggle" 2>/dev/null &&
        gsettings set "$base" binding "<Ctrl><Shift>r" 2>/dev/null &&
        # Add to the list of custom keybindings
        existing=$(gsettings get org.gnome.settings-daemon.plugins.media-keys custom-keybindings 2>/dev/null || echo "[]")
        if ! echo "$existing" | grep -q "sussurro"; then
          new=$(echo "$existing" | sed "s/]/, '$path']/" | sed "s/\[, /[/")
          gsettings set org.gnome.settings-daemon.plugins.media-keys custom-keybindings "$new"
        fi &&
        echo "Registered Ctrl+Shift+R shortcut (GNOME)" ||
        echo "Could not register GNOME shortcut (gsettings not available)"
        ;;
      *)
        echo "Tip: bind 'sussurro --toggle' to Ctrl+Shift+R in your desktop settings"
        ;;
    esac

# Remove sussurro
uninstall:
    #!/bin/bash
    rm -f "{{prefix}}/bin/sussurro"
    rm -f "$HOME/.local/share/applications/sussurro.desktop"
    rm -f "$HOME/.local/share/icons/hicolor/512x512/apps/sussurro.png"
    update-desktop-database "$HOME/.local/share/applications/" 2>/dev/null || true
    # Remove COSMIC shortcut
    file="$HOME/.config/cosmic/com.system76.CosmicSettings.Shortcuts/v1/custom"
    if [ -f "$file" ] && grep -q "sussurro" "$file"; then
      # If it's the only entry, write an empty map
      echo "{}" > "$file"
      echo "Removed COSMIC shortcut"
    fi
    echo "Uninstalled sussurro"

# Re-generates TypeScript bindings from Go structs
bindings:
    wails generate module
