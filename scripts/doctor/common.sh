# shellcheck shell=bash
# Checks shared across all OSes: go, bun, wails, git. Populates the
# global `_doctor_rows` array with tab-separated rows of:
#
#     tool<TAB>status<TAB>version<TAB>classification<TAB>action
#
# where status is one of: ok | missing | outdated, classification is
# one of: required | recommended | optional. Caller appends OS-specific
# rows and then calls doctor_print_table.

# --- logging / version setup -------------------------------------------------

_doctor_lib_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/../lib" && pwd)"
# shellcheck source=../lib/log.sh
source "${_doctor_lib_dir}/log.sh"
# shellcheck source=../lib/versions.env
source "${_doctor_lib_dir}/versions.env"

# Populated by row_* helpers, printed by doctor_print_table.
# Uses `|` as field separator so empty fields survive `read -r` splitting
# (bash collapses consecutive whitespace-IFS characters).
_doctor_rows=()

# row <tool> <status> <version> <class> <action>
row() {
    _doctor_rows+=("$1|$2|$3|$4|$5")
}

# --- individual checks -------------------------------------------------------

check_go() {
    if ! command -v go >/dev/null 2>&1; then
        row "go" "missing" "" "required" "install go ${GO_VERSION}+"
        return
    fi
    local v
    v="$(go env GOVERSION 2>/dev/null | sed 's/^go//')"
    row "go" "ok" "${v}" "required" ""
}

check_bun() {
    if ! command -v bun >/dev/null 2>&1; then
        row "bun" "missing" "" "required" "install bun ${BUN_VERSION}+"
        return
    fi
    local v
    v="$(bun --version 2>/dev/null)"
    row "bun" "ok" "${v}" "required" ""
}

check_wails() {
    if ! command -v wails >/dev/null 2>&1; then
        row "wails" "missing" "" "required" "go install github.com/wailsapp/wails/v2/cmd/wails@${WAILS_VERSION}"
        return
    fi
    local v
    v="$(wails version 2>/dev/null | head -1 | awk '{print $NF}')"
    row "wails" "ok" "${v}" "required" ""
}

check_git() {
    if ! command -v git >/dev/null 2>&1; then
        row "git" "missing" "" "required" "install git"
        return
    fi
    local v
    v="$(git --version 2>/dev/null | awk '{print $3}')"
    row "git" "ok" "${v}" "required" ""
}

check_just() {
    if ! command -v just >/dev/null 2>&1; then
        row "just" "missing" "" "required" "install just (cargo install just, or system package)"
        return
    fi
    local v
    v="$(just --version 2>/dev/null | awk '{print $2}')"
    row "just" "ok" "${v}" "required" ""
}

# --- optional dev-tool checks ------------------------------------------------

check_golangci_lint() {
    if ! command -v golangci-lint >/dev/null 2>&1; then
        row "golangci-lint" "missing" "" "optional" "optional: go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest"
        return
    fi
    local v
    v="$(golangci-lint version --short 2>/dev/null || golangci-lint version 2>/dev/null | awk '{print $4}')"
    row "golangci-lint" "ok" "${v}" "optional" ""
}

# --- table rendering ---------------------------------------------------------

# doctor_print_table: pretty-prints _doctor_rows and sets global
# _doctor_exit to 0 if every required row is "ok", else 1.
doctor_print_table() {
    _doctor_exit=0
    printf '\n%-16s  %-12s  %-14s  %-12s  %s\n' "Tool" "Status" "Version" "Class" "Action"
    printf -- '%s\n' "──────────────────────────────────────────────────────────────────────"
    local row tool status version class action mark
    for row in "${_doctor_rows[@]}"; do
        IFS='|' read -r tool status version class action <<<"${row}"
        case "${status}" in
            ok)       mark="✓ ok" ;;
            missing)  mark="✗ missing" ;;
            outdated) mark="! outdated" ;;
            *)        mark="${status}" ;;
        esac
        printf '%-16s  %-12s  %-14s  %-12s  %s\n' \
            "${tool}" "${mark}" "${version:-—}" "${class}" "${action}"
        if [ "${status}" != "ok" ] && [ "${class}" = "required" ]; then
            _doctor_exit=1
        fi
    done
    echo
    if [ "${_doctor_exit}" -eq 0 ]; then
        info "all required dependencies present"
    else
        warn "missing required dependencies — run 'just bootstrap' to install"
    fi
}
