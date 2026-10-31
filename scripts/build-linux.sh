#!/usr/bin/env bash
# Build de producción para Linux (Ubuntu 24.04+).
#
# Requisitos: go, pnpm, wails CLI y las librerías de desarrollo:
#   sudo apt install build-essential pkg-config libgtk-3-dev libwebkit2gtk-4.1-dev
#
# El flag webkit2_41 es obligatorio en Ubuntu 24.04+/26.04, que ya no empaquetan
# WebKitGTK 4.0. El binario queda en build/bin/opencode-tray.
set -euo pipefail

cd "$(dirname "$0")/.."

export CGO_ENABLED=1

if pkg-config --exists webkit2gtk-4.1; then
  WAILS_TAGS="webkit2_41"
elif pkg-config --exists webkit2gtk-4.0; then
  WAILS_TAGS=""
else
  echo "Faltan las librerías de WebKitGTK. Instala:" >&2
  echo "  sudo apt install build-essential pkg-config libgtk-3-dev libwebkit2gtk-4.1-dev" >&2
  exit 1
fi

args=(-clean)
if [[ -n "$WAILS_TAGS" ]]; then
  args+=(-tags "$WAILS_TAGS")
fi

wails build "${args[@]}" "$@"
