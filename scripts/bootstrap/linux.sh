#!/usr/bin/env bash
# Install missing dev dependencies via the system package manager.
# Accepts --yes / -y to skip confirmation. Dry-run prints commands.
set -euo pipefail

_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=../lib/log.sh
source "${_dir}/../lib/log.sh"
# shellcheck source=../lib/os.sh
source "${_dir}/../lib/os.sh"
# shellcheck source=../lib/versions.env
source "${_dir}/../lib/versions.env"

assume_yes=0
dry_run="${DRY_RUN:-0}"
for arg in "$@"; do
    case "${arg}" in
        --yes|-y) assume_yes=1 ;;
        --dry-run) dry_run=1 ;;
        *) error "unknown argument: ${arg}"; exit 1 ;;
    esac
done

pm="$(detect_pkg_manager)"
if [ -z "${pm}" ]; then
    error "no supported package manager found (apt, dnf, pacman)"
    exit 1
fi
info "using package manager: ${pm}"

# run_install <description> <cmd...> — prompts (or auto-confirms) and
# executes. Respects DRY_RUN.
run_install() {
    local desc="$1"; shift
    section "${desc}"
    printf '  $ %s\n' "$*"
    if [ "${dry_run}" = "1" ]; then
        dim "(dry run — not executing)"
        return 0
    fi
    if [ "${assume_yes}" -ne 1 ]; then
        read -r -p "  proceed? [y/N] " ans
        case "${ans}" in y|Y|yes|YES) ;; *) warn "skipped"; return 0 ;; esac
    fi
    "$@"
}

# --- system packages ---------------------------------------------------------

case "${pm}" in
    apt)
        run_install "system build deps (webkit, build tools, paste helpers)" \
            sudo apt-get install -y \
                build-essential pkg-config \
                libwebkit2gtk-4.1-dev \
                libgtk-3-dev \
                wtype xdotool
        ;;
    dnf)
        run_install "system build deps (webkit, build tools, paste helpers)" \
            sudo dnf install -y \
                gcc pkgconfig \
                webkit2gtk4.1-devel \
                gtk3-devel \
                wtype xdotool
        ;;
    pacman)
        run_install "system build deps (webkit, build tools, paste helpers)" \
            sudo pacman -S --needed --noconfirm \
                base-devel pkgconf \
                webkit2gtk-4.1 \
                gtk3 \
                wtype xdotool
        ;;
    *)
        warn "${pm}: no recipe. install webkit2gtk-4.1, gtk3, build tools, wtype, xdotool manually."
        ;;
esac

# --- language toolchains -----------------------------------------------------

if ! command -v go >/dev/null 2>&1; then
    warn "go not found. install Go ${GO_VERSION}+ from https://go.dev/dl/ (distro packages are often too old)."
fi

if ! command -v bun >/dev/null 2>&1; then
    run_install "bun (frontend runtime)" \
        bash -c 'curl -fsSL https://bun.sh/install | bash'
fi

if ! command -v wails >/dev/null 2>&1; then
    run_install "wails CLI ${WAILS_VERSION}" \
        go install "github.com/wailsapp/wails/v2/cmd/wails@${WAILS_VERSION}"
fi

info "bootstrap complete — run 'just doctor' to verify"
