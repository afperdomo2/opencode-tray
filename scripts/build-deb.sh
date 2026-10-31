#!/usr/bin/env bash
# Genera el paquete .deb de OpenCode Tray para Ubuntu 24.04+.
#
# Uso:
#   ./scripts/build-deb.sh            # compila y empaqueta
#   sudo apt install ./dist/opencode-tray_<versión>_<arch>.deb
#
# El paquete instala:
#   /usr/bin/opencode-tray
#   /usr/share/applications/opencode-tray.desktop   (icono en el menú)
#   /usr/share/icons/hicolor/256x256/apps/opencode-tray.png
set -euo pipefail

cd "$(dirname "$0")/.."

PKG="opencode-tray"
MAINTAINER="Andrés Felipe Perdomo <alvarado0925@gmail.com>"
ARCH="$(dpkg --print-architecture)"

# La versión sale de wails.json (misma fuente que el .exe de Windows).
VERSION="$(sed -n 's/.*"productVersion": *"\([^"]*\)".*/\1/p' wails.json | head -1)"
if [[ -z "$VERSION" ]]; then
  echo "No se pudo leer Info.productVersion de wails.json" >&2
  exit 1
fi

echo "→ Compilando el binario (versión $VERSION)…"
./scripts/build-linux.sh

STAGE="build/deb/${PKG}_${VERSION}_${ARCH}"
rm -rf "$STAGE"
mkdir -p \
  "$STAGE/DEBIAN" \
  "$STAGE/usr/bin" \
  "$STAGE/usr/share/applications" \
  "$STAGE/usr/share/icons/hicolor/256x256/apps" \
  "$STAGE/usr/share/doc/$PKG"

install -m 0755 build/bin/opencode-tray "$STAGE/usr/bin/$PKG"
install -m 0644 build/linux/opencode-tray.desktop "$STAGE/usr/share/applications/$PKG.desktop"
install -m 0644 build/linux/icon-256.png "$STAGE/usr/share/icons/hicolor/256x256/apps/$PKG.png"
install -m 0644 README.md "$STAGE/usr/share/doc/$PKG/README.md"

SIZE="$(du -sk "$STAGE/usr" | cut -f1)"

cat > "$STAGE/DEBIAN/control" <<EOF
Package: $PKG
Version: $VERSION
Architecture: $ARCH
Maintainer: $MAINTAINER
Installed-Size: $SIZE
Depends: libgtk-3-0t64 | libgtk-3-0, libwebkit2gtk-4.1-0
Recommends: gnome-shell-extension-appindicator
Section: utils
Priority: optional
Description: App de bandeja para OpenCode
 Controla el servicio de OpenCode, gestiona el acceso web (local, red y
 remoto), muestra el consumo de OpenCode Go y permite ver la configuración
 de MCP, skills y agents. Necesita el CLI de OpenCode instalado.
EOF

cat > "$STAGE/DEBIAN/copyright" <<EOF
Format: https://www.debian.org/doc/packaging-manuals/copyright-format/1.0/
Upstream-Name: OpenCode Tray
Copyright: $(date +%Y) Andrés Felipe Perdomo
License: propietario (sin licencia publicada)
EOF

desktop-file-validate "$STAGE/usr/share/applications/$PKG.desktop"

mkdir -p dist
OUT="dist/${PKG}_${VERSION}_${ARCH}.deb"
dpkg-deb --root-owner-group --build "$STAGE" "$OUT" >/dev/null

echo
echo "✓ Paquete generado: $OUT"
dpkg-deb --info "$OUT" | sed -n '1,12p'
echo
echo "Instalar:  sudo apt install ./$OUT"
