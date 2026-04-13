# shellcheck shell=bash
# Shared version / ldflags / install-prefix computation. Sourced, not run.
#
# Exports:
#   VERSION   - git-describe string, or "dev" if git describe fails
#   LDFLAGS   - Go linker flags embedding VERSION into the binary
#   PREFIX    - user install root; $PREFIX env wins, else $HOME/.local,
#               else $USERPROFILE/.local (Windows fallback)

# Idempotent — safe to source multiple times in the same shell.
if [ -z "${VERSION:-}" ]; then
    VERSION="$(git describe --tags --always --dirty --exclude 'backup/*' 2>/dev/null || echo dev)"
    export VERSION
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
