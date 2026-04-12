#!/usr/bin/env bash
# Build a .deb package from an already-built binary.
# Usage: ./scripts/build-deb.sh <version>
# Expects build/bin/shushingface to exist (run build-release.sh first).
set -euo pipefail

VERSION="${1:?Usage: build-deb.sh <version>}"
# Strip leading 'v' for deb versioning
DEB_VERSION="${VERSION#v}"
ARCH="amd64"
PKG_NAME="shushingface"
PKG_DIR="dist/${PKG_NAME}_${DEB_VERSION}_${ARCH}"

echo "Building .deb for ${PKG_NAME} ${DEB_VERSION}..."

# Clean
rm -rf "$PKG_DIR"

# Directory structure
mkdir -p "${PKG_DIR}/DEBIAN"
mkdir -p "${PKG_DIR}/usr/bin"
mkdir -p "${PKG_DIR}/usr/share/applications"
mkdir -p "${PKG_DIR}/usr/share/icons/hicolor/512x512/apps"

# Binary
cp build/bin/shushingface "${PKG_DIR}/usr/bin/shushingface"
chmod 755 "${PKG_DIR}/usr/bin/shushingface"

# Desktop entry
cp build/linux/shushingface.desktop "${PKG_DIR}/usr/share/applications/shushingface.desktop"

# Icon
cp build/appicon.png "${PKG_DIR}/usr/share/icons/hicolor/512x512/apps/shushingface.png"

# Control file
cat > "${PKG_DIR}/DEBIAN/control" << EOF
Package: ${PKG_NAME}
Version: ${DEB_VERSION}
Section: utils
Priority: optional
Architecture: ${ARCH}
Depends: libwebkit2gtk-4.1-0, libgtk-3-0
Maintainer: dbus <dbus@noreply.codeberg.org>
Description: Voice-to-text refinement tool
 Speak naturally, get polished text. Records your voice,
 transcribes via AI, refines the transcript into clean text,
 and types it where your cursor is.
Homepage: https://codeberg.org/dbus/shushingface
EOF

# Post-install: update icon cache and desktop database
cat > "${PKG_DIR}/DEBIAN/postinst" << 'EOF'
#!/bin/sh
set -e
gtk-update-icon-cache -f -t /usr/share/icons/hicolor/ 2>/dev/null || true
update-desktop-database /usr/share/applications/ 2>/dev/null || true
EOF
chmod 755 "${PKG_DIR}/DEBIAN/postinst"

# Post-remove: same cleanup
cat > "${PKG_DIR}/DEBIAN/postrm" << 'EOF'
#!/bin/sh
set -e
gtk-update-icon-cache -f -t /usr/share/icons/hicolor/ 2>/dev/null || true
update-desktop-database /usr/share/applications/ 2>/dev/null || true
EOF
chmod 755 "${PKG_DIR}/DEBIAN/postrm"

# Build the .deb
dpkg-deb --build --root-owner-group "$PKG_DIR"
DEB_FILE="dist/${PKG_NAME}_${DEB_VERSION}_${ARCH}.deb"
echo "Built: ${DEB_FILE}"
