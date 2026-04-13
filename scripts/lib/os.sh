# shellcheck shell=bash
# OS / package-manager / display-server detection helpers. Sourced.

# detect_os prints: linux | windows | unknown
detect_os() {
    case "$(uname -s 2>/dev/null)" in
        Linux*)                  echo linux ;;
        MINGW*|MSYS*|CYGWIN*)    echo windows ;;
        *)                       echo unknown ;;
    esac
}

# detect_pkg_manager prints the first package manager found on PATH.
# Linux: apt | dnf | pacman. Windows: winget. Empty string if none.
detect_pkg_manager() {
    if command -v apt-get >/dev/null 2>&1; then
        echo apt
    elif command -v dnf >/dev/null 2>&1; then
        echo dnf
    elif command -v pacman >/dev/null 2>&1; then
        echo pacman
    elif command -v winget >/dev/null 2>&1; then
        echo winget
    else
        echo ""
    fi
}

# detect_display_server prints: wayland | x11 | unknown (Linux only).
detect_display_server() {
    case "${XDG_SESSION_TYPE:-}" in
        wayland)  echo wayland ;;
        x11)      echo x11 ;;
        *)        echo unknown ;;
    esac
}
