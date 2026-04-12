#!/usr/bin/env bash
# Build and package a release for the current platform.
# Usage: ./scripts/build-release.sh [version]
# If no version is given, uses git describe.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
VERSION="${1:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}"
LDFLAGS="-X codeberg.org/dbus/shushingface/internal/version.version=$VERSION"

echo "Building shushingface $VERSION..."

# Build
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    wails build -tags webkit2_41 -ldflags "$LDFLAGS"
else
    wails build -ldflags "$LDFLAGS"
fi

# Package tarball
mkdir -p dist/staging
cp build/bin/shushingface dist/staging/
cp build/linux/shushingface.desktop dist/staging/ 2>/dev/null || true
cp build/appicon.png dist/staging/

ARCHIVE="dist/shushingface-${VERSION}-linux-amd64.tar.gz"
tar czf "$ARCHIVE" -C dist/staging .
rm -rf dist/staging
echo "Packaged: $ARCHIVE"

# Package .deb (Linux only)
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    "$SCRIPT_DIR/build-deb.sh" "$VERSION"
fi
