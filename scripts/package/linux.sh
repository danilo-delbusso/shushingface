#!/usr/bin/env bash
# Build and package a Linux release: tar.gz + .deb.
# Paths inside the .deb follow FHS (/usr/bin, /usr/share/...), NOT $PREFIX —
# .deb is a system package, not a user install.
#
# Usage: scripts/package/linux.sh [version]
set -euo pipefail

_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=../lib/log.sh
source "${_dir}/../lib/log.sh"
# shellcheck source=../lib/version.sh
source "${_dir}/../lib/version.sh"

# Allow positional override but default to the shared VERSION.
if [ $# -gt 0 ] && [ -n "${1:-}" ]; then
    VERSION="$1"
    LDFLAGS="-X codeberg.org/dbus/shushingface/internal/version.version=${VERSION}"
fi

section "packaging shushingface ${VERSION} for linux"

rm -rf build/bin/
wails build -tags webkit2_41 -ldflags "${LDFLAGS}"

mkdir -p dist

# --- tar.gz -----------------------------------------------------------------
staging="dist/staging"
rm -rf "${staging}"
mkdir -p "${staging}"
cp build/bin/shushingface "${staging}/"
cp build/linux/shushingface.desktop "${staging}/" 2>/dev/null || true
cp build/appicon.png "${staging}/"

archive="dist/shushingface-${VERSION}-linux-amd64.tar.gz"
tar czf "${archive}" -C "${staging}" .
rm -rf "${staging}"
info "packaged ${archive}"

# --- .deb -------------------------------------------------------------------
deb_version="${VERSION#v}"
arch="amd64"
pkg="shushingface"
pkg_dir="dist/${pkg}_${deb_version}_${arch}"

rm -rf "${pkg_dir}"
mkdir -p "${pkg_dir}/DEBIAN" \
         "${pkg_dir}/usr/bin" \
         "${pkg_dir}/usr/share/applications" \
         "${pkg_dir}/usr/share/icons/hicolor/512x512/apps"

install -Dm755 build/bin/shushingface               "${pkg_dir}/usr/bin/shushingface"
install -Dm644 build/linux/shushingface.desktop     "${pkg_dir}/usr/share/applications/shushingface.desktop"
install -Dm644 build/appicon.png                    "${pkg_dir}/usr/share/icons/hicolor/512x512/apps/shushingface.png"

cat > "${pkg_dir}/DEBIAN/control" <<EOF
Package: ${pkg}
Version: ${deb_version}
Section: utils
Priority: optional
Architecture: ${arch}
Depends: libwebkit2gtk-4.1-0, libgtk-3-0
Maintainer: dbus <dbus@noreply.codeberg.org>
Description: Voice-to-text refinement tool
 Speak naturally, get polished text. Records your voice,
 transcribes via AI, refines the transcript into clean text,
 and types it where your cursor is.
Homepage: https://codeberg.org/dbus/shushingface
EOF

cat > "${pkg_dir}/DEBIAN/postinst" <<'EOF'
#!/bin/sh
set -e
gtk-update-icon-cache -f -t /usr/share/icons/hicolor/ 2>/dev/null || true
update-desktop-database /usr/share/applications/ 2>/dev/null || true
EOF
chmod 755 "${pkg_dir}/DEBIAN/postinst"

cat > "${pkg_dir}/DEBIAN/postrm" <<'EOF'
#!/bin/sh
set -e
gtk-update-icon-cache -f -t /usr/share/icons/hicolor/ 2>/dev/null || true
update-desktop-database /usr/share/applications/ 2>/dev/null || true
EOF
chmod 755 "${pkg_dir}/DEBIAN/postrm"

dpkg-deb --build --root-owner-group "${pkg_dir}"
deb_file="dist/${pkg}_${deb_version}_${arch}.deb"
info "packaged ${deb_file}"
