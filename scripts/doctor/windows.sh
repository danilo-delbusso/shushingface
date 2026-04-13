#!/usr/bin/env bash
set -euo pipefail

_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=./common.sh
source "${_dir}/common.sh"

section "checking build / runtime dependencies (windows)"

# Hard-required: Git Bash. We're running under it right now, so just
# verify sh is sane and PATH has the mingw tools we expect.
if [ -z "${MSYSTEM:-}" ] && [ -z "${MINGW_PREFIX:-}" ] && ! command -v bash >/dev/null 2>&1; then
    row "git-bash" "missing" "" "required" "install Git for Windows (provides bash + git describe)"
else
    row "git-bash" "ok" "${MSYSTEM:-bash}" "required" ""
fi

check_go
check_bun
check_wails
check_git
check_just

# CGO toolchain: gcc or clang must be on PATH for wails to build.
# `gcc -dumpversion` gives a clean "15.0.0" without the vendor suffix.
if command -v gcc >/dev/null 2>&1; then
    ver="$(gcc -dumpversion 2>/dev/null || echo '?')"
    row "gcc" "ok" "${ver}" "required" ""
elif command -v clang >/dev/null 2>&1; then
    ver="$(clang -dumpversion 2>/dev/null || echo '?')"
    row "clang" "ok" "${ver}" "required" ""
else
    row "cgo-compiler" "missing" "" "required" "install llvm-mingw (winget install -e --id mstorsjo.LLVM-MinGW.UCRT)"
fi

# winget for bootstrap.
if command -v winget >/dev/null 2>&1; then
    row "winget" "ok" "" "recommended" ""
else
    row "winget" "missing" "" "recommended" "install App Installer from Microsoft Store"
fi

check_golangci_lint

doctor_print_table
exit "${_doctor_exit}"
