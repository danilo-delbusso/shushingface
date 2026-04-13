#!/usr/bin/env bash
set -euo pipefail

_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=../lib/log.sh
source "${_dir}/../lib/log.sh"
# shellcheck source=../lib/version.sh
source "${_dir}/../lib/version.sh"

rm -f "${PREFIX}/bin/shushingface"
rm -f "${HOME}/.local/share/applications/shushingface.desktop"
rm -f "${HOME}/.local/share/icons/hicolor/512x512/apps/shushingface.png"
update-desktop-database "${HOME}/.local/share/applications/" 2>/dev/null || true

cosmic_file="${HOME}/.config/cosmic/com.system76.CosmicSettings.Shortcuts/v1/custom"
if [ -f "${cosmic_file}" ] && grep -q "shushingface" "${cosmic_file}"; then
    # If the shortcut was the only entry, reset to empty map. Otherwise
    # leave the file alone — a future iteration can do surgical removal.
    echo "{}" > "${cosmic_file}"
    info "removed COSMIC shortcut"
fi

info "uninstalled shushingface"
