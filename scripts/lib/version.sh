# shellcheck shell=bash
# Shared version / ldflags / install-prefix computation. Sourced, not run.
#
# Exports:
#   VERSION         - git-describe string, or "dev" if git describe fails.
#                     Used in -ldflags for internal/version.version.
#   PRODUCT_VERSION - numeric MAJOR.MINOR.PATCH from the nearest tag
#                     (v-prefix stripped), or "0.0.0" if no tag. NSIS
#                     VIProductVersion requires a numeric form.
#   LDFLAGS         - Go linker flags embedding VERSION into the binary.
#   PREFIX          - user install root; $PREFIX env wins, else $HOME/.local,
#                     else $USERPROFILE/.local (Windows fallback).

# Idempotent — safe to source multiple times in the same shell.
if [ -z "${VERSION:-}" ]; then
    VERSION="$(git describe --tags --always --dirty --exclude 'backup/*' 2>/dev/null || echo dev)"
    export VERSION
fi

if [ -z "${PRODUCT_VERSION:-}" ]; then
    _tag="$(git describe --tags --abbrev=0 --exclude 'backup/*' 2>/dev/null || true)"
    _tag="${_tag#v}"
    if [[ "${_tag}" =~ ^([0-9]+\.[0-9]+\.[0-9]+) ]]; then
        PRODUCT_VERSION="${BASH_REMATCH[1]}"
    else
        PRODUCT_VERSION="0.0.0"
    fi
    unset _tag
    export PRODUCT_VERSION
fi

if [ -z "${LDFLAGS:-}" ]; then
    LDFLAGS="-X codeberg.org/dbus/shushingface/internal/version.version=${VERSION}"
    export LDFLAGS
fi

if [ -z "${PREFIX:-}" ]; then
    if [ -n "${HOME:-}" ]; then
        PREFIX="${HOME}/.local"
    elif [ -n "${USERPROFILE:-}" ]; then
        PREFIX="${USERPROFILE}/.local"
    else
        PREFIX="./.local"
    fi
    export PREFIX
fi

# Patch wails.json's info.productVersion in-place to match PRODUCT_VERSION,
# then register a trap that restores the original content on EXIT. Safe to
# call multiple times; only the first call registers the trap.
#
# Usage: sync_wails_product_version [path/to/wails.json]
sync_wails_product_version() {
    local file="${1:-wails.json}"
    if [ ! -f "${file}" ]; then
        return 0
    fi

    # Snapshot once so the trap always restores the pre-build contents.
    if [ -z "${_WAILS_JSON_BACKUP:-}" ]; then
        _WAILS_JSON_BACKUP="$(mktemp)"
        cp "${file}" "${_WAILS_JSON_BACKUP}"
        _WAILS_JSON_PATH="${file}"
        export _WAILS_JSON_BACKUP _WAILS_JSON_PATH
        # shellcheck disable=SC2064
        trap "cp '${_WAILS_JSON_BACKUP}' '${_WAILS_JSON_PATH}'; rm -f '${_WAILS_JSON_BACKUP}'" EXIT
    fi

    # Replace the productVersion value. GNU sed (Linux + mingw/git-bash) supports -i.
    sed -i.tmp -E "s/(\"productVersion\"[[:space:]]*:[[:space:]]*\")[^\"]*(\")/\1${PRODUCT_VERSION}\2/" "${file}"
    rm -f "${file}.tmp"
}
