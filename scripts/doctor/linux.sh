#!/usr/bin/env bash
set -euo pipefail

_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=./common.sh
source "${_dir}/common.sh"
# shellcheck source=../lib/os.sh
source "${_dir}/../lib/os.sh"

section "checking build / runtime dependencies (linux)"

check_go
check_bun
check_wails
check_git
check_just

# webkit2gtk-4.1 is needed to build the GUI. We can only probe via
# pkg-config since headers live in a dev package.
if pkg-config --exists webkit2gtk-4.1 2>/dev/null; then
    ver="$(pkg-config --modversion webkit2gtk-4.1 2>/dev/null || echo '?')"
    row "webkit2gtk-4.1" "ok" "${ver}" "required" ""
else
    row "webkit2gtk-4.1" "missing" "" "required" "install libwebkit2gtk-4.1-dev (apt) / webkit2gtk4.1-devel (dnf)"
fi

# Paste helpers: install both; runtime picks based on $XDG_SESSION_TYPE.
display="$(detect_display_server)"
for tool in wtype xdotool; do
    if command -v "${tool}" >/dev/null 2>&1; then
        row "${tool}" "ok" "" "recommended" ""
    else
        note="paste helper"
        case "${tool}" in
            wtype)   [ "${display}" = "wayland" ] && note="paste helper (active on current session)" ;;
            xdotool) [ "${display}" = "x11" ]     && note="paste helper (active on current session)" ;;
        esac
        row "${tool}" "missing" "" "recommended" "install (${note})"
    fi
done

check_golangci_lint

doctor_print_table
exit "${_doctor_exit}"
