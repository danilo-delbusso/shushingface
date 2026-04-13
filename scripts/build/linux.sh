#!/usr/bin/env bash
# Linux build dispatcher. First arg: `dev` (runs `wails dev`) or `build`
# (runs `wails build`). Remaining args are forwarded to wails verbatim.
set -euo pipefail

_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=../lib/version.sh
source "${_dir}/../lib/version.sh"

mode="${1:-build}"
shift || true

sync_wails_product_version

case "${mode}" in
    dev)
        wails dev -tags webkit2_41 -ldflags "${LDFLAGS}" "$@"
        ;;
    build)
        wails build -tags webkit2_41 -ldflags "${LDFLAGS}" "$@"
        ;;
    *)
        # Pass-through for any wails subcommand (generate, doctor, etc.).
        wails "${mode}" -tags webkit2_41 -ldflags "${LDFLAGS}" "$@"
        ;;
esac
