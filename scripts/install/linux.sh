#!/usr/bin/env bash
# User-local install on Linux: binary + .desktop + icon + per-DE
# keyboard shortcut registration.
set -euo pipefail

_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=../lib/log.sh
source "${_dir}/../lib/log.sh"
# shellcheck source=../lib/version.sh
source "${_dir}/../lib/version.sh"

bin_src="build/bin/shushingface"
desktop_src="build/linux/shushingface.desktop"
icon_src="build/appicon.png"

if [ ! -f "${bin_src}" ]; then
    error "${bin_src} not found - run 'just build' first"
    exit 1
fi

install -Dm755 "${bin_src}"     "${PREFIX}/bin/shushingface"
install -Dm644 "${desktop_src}" "${HOME}/.local/share/applications/shushingface.desktop"
install -Dm644 "${icon_src}"    "${HOME}/.local/share/icons/hicolor/512x512/apps/shushingface.png"

gtk-update-icon-cache -f -t "${HOME}/.local/share/icons/hicolor/" 2>/dev/null || true
update-desktop-database "${HOME}/.local/share/applications/" 2>/dev/null || true

# Desktop-environment specific Super+Ctrl+B shortcut registration.
register_shortcut() {
    case "${XDG_CURRENT_DESKTOP:-}" in
        COSMIC)
            local dir="${HOME}/.config/cosmic/com.system76.CosmicSettings.Shortcuts/v1"
            local file="${dir}/custom"
            mkdir -p "${dir}"
            if [ -f "${file}" ] && grep -q "shushingface" "${file}"; then
                info "shortcut already registered (COSMIC)"
                return
            fi
            if [ -f "${file}" ] && grep -q "Spawn" "${file}"; then
                sed -i 's/}$/    (\n        modifiers: [\n            Super,\n            Ctrl,\n        ],\n        key: "b",\n        description: Some("shushingface: toggle recording"),\n    ): Spawn("shushingface --toggle"),\n}/' "${file}"
            else
                cat > "${file}" <<'RON'
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
            info "registered Super+Ctrl+B shortcut (COSMIC)"
            info "log out and back in (or restart cosmic-comp) for it to take effect"
            ;;
        GNOME*)
            local path="/org/gnome/settings-daemon/plugins/media-keys/custom-keybindings/shushingface/"
            local base="org.gnome.settings-daemon.plugins.media-keys.custom-keybinding:${path}"
            gsettings set "${base}" name "shushingface" 2>/dev/null || return
            gsettings set "${base}" command "shushingface --toggle" 2>/dev/null || return
            gsettings set "${base}" binding "<Super><Ctrl>b" 2>/dev/null || return
            local existing
            existing="$(gsettings get org.gnome.settings-daemon.plugins.media-keys custom-keybindings 2>/dev/null || echo "[]")"
            if ! echo "${existing}" | grep -q "shushingface"; then
                local new
                new="$(echo "${existing}" | sed "s/]/, '${path}']/" | sed "s/\[, /[/")"
                gsettings set org.gnome.settings-daemon.plugins.media-keys custom-keybindings "${new}"
            fi
            info "registered Super+Ctrl+B shortcut (GNOME)"
            ;;
        *)
            info "bind 'shushingface --toggle' to Super+Ctrl+B in your desktop settings"
            ;;
    esac
}

register_shortcut

info "installed shushingface to ${PREFIX}/bin/shushingface"
