# shellcheck shell=bash
# Structured console logging for dev-infra scripts. All output goes to
# stderr so it never contaminates stdout (which some scripts reserve
# for machine-readable output). Sourced, not run.

# Detect colour support: TTY on stderr, TERM not "dumb", NO_COLOR unset.
if [ -t 2 ] && [ "${TERM:-}" != "dumb" ] && [ -z "${NO_COLOR:-}" ]; then
    _log_c_dim=$'\033[2m'
    _log_c_blue=$'\033[34m'
    _log_c_yellow=$'\033[33m'
    _log_c_red=$'\033[31m'
    _log_c_bold=$'\033[1m'
    _log_c_reset=$'\033[0m'
else
    _log_c_dim=''
    _log_c_blue=''
    _log_c_yellow=''
    _log_c_red=''
    _log_c_bold=''
    _log_c_reset=''
fi

section() {
    printf '%s\n==> %s%s\n' "${_log_c_bold}" "$*" "${_log_c_reset}" >&2
}

info() {
    printf '%s[info]%s %s\n' "${_log_c_blue}" "${_log_c_reset}" "$*" >&2
}

warn() {
    printf '%s[warn]%s %s\n' "${_log_c_yellow}" "${_log_c_reset}" "$*" >&2
}

error() {
    printf '%s[error]%s %s\n' "${_log_c_red}" "${_log_c_reset}" "$*" >&2
}

dim() {
    printf '%s%s%s\n' "${_log_c_dim}" "$*" "${_log_c_reset}" >&2
}
