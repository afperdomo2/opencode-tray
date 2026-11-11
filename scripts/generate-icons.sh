#!/usr/bin/env bash
# Genera los iconos de la app a partir de las fuentes SVG de `brand/`.
#
# Fuente única:
#   brand/icon.svg  → tile de app (fondo #1E293B + glifo verde)
#   brand/mark.svg  → glifo suelto con `currentColor` (bandeja)
#
# Salidas (se commitean):
#   build/appicon.png          1024  máster de Wails
#   build/windows/icon.ico     16..256  .exe e instalador (tile oscuro)
#   build/windows/tray.ico     16..256  bandeja de Windows (tile verde)
#   build/linux/icon.png       64    bandeja en Linux/macOS (PNG embebido)
#   build/linux/icon-256.png   256   .deb y menú de aplicaciones
#
# Requiere Google Chrome/Chromium (rasteriza SVG) y Python 3 + Pillow
# (supersampling y ensamblado del ICO). Es manual: no corre en CI ni en hooks.
set -Eeuo pipefail
trap 'echo "❌ Error en ${BASH_SOURCE[0]} línea $LINENO" >&2' ERR

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BRAND_DIR="$ROOT_DIR/brand"
BUILD_DIR="$ROOT_DIR/build"
TMP_DIR="$(mktemp -d)"
cleanup() { rm -rf "$TMP_DIR"; }
trap cleanup EXIT

die() {
  echo "❌ $*" >&2
  exit 1
}

find_chrome() {
  local bin
  for bin in \
    google-chrome \
    google-chrome-stable \
    chromium \
    chromium-browser \
    "/c/Program Files/Google/Chrome/Application/chrome.exe" \
    "/c/Program Files (x86)/Google/Chrome/Application/chrome.exe"; do
    if command -v "$bin" >/dev/null 2>&1; then
      command -v "$bin"
      return 0
    fi
  done
  return 1
}

find_python() {
  local bin
  for bin in python3 python; do
    if command -v "$bin" >/dev/null 2>&1; then
      command -v "$bin"
      return 0
    fi
  done
  return 1
}

# Chrome es un binario nativo: en Windows/Git Bash hay que pasarle rutas
# nativas aunque el script use rutas POSIX.
chrome_path() {
  if command -v cygpath >/dev/null 2>&1; then
    cygpath -w "$1"
  else
    printf '%s' "$1"
  fi
}

# URL file:// válida tanto en Linux/macOS como en Windows/Git Bash.
file_url() {
  if command -v cygpath >/dev/null 2>&1; then
    local win
    win="$(cygpath -m "$1")"
    printf 'file:///%s' "${win#/}"
  else
    printf 'file://%s' "$1"
  fi
}

[[ -f "$BRAND_DIR/icon.svg" ]] || die "No existe $BRAND_DIR/icon.svg"
[[ -f "$BRAND_DIR/mark.svg" ]] || die "No existe $BRAND_DIR/mark.svg"

CHROME_BIN="$(find_chrome)" \
  || die "No encontré Chrome/Chromium. Instalalo para rasterizar los SVG."
PYTHON_BIN="$(find_python)" || die "Falta python3/python."
"$PYTHON_BIN" -c 'import PIL' 2>/dev/null \
  || die "Falta Pillow: python3 -m pip install pillow"

ICON_SVG="$(cat "$BRAND_DIR/icon.svg")"
MARK_SVG="$(cat "$BRAND_DIR/mark.svg")"

# Renderiza un SVG a PNG cuadrado con Chrome headless.
#
#   render <modo> <px> <salida>
#
#   app  → brand/icon.svg (tile oscuro)
#   tray → brand/mark.svg sobre un tile verde redondeado
render() {
  local mode="$1"
  local size="$2"
  local out="$3"
  local wrap svg

  case "$mode" in
    app)
      wrap="display:grid;place-items:center"
      svg="$ICON_SVG"
      ;;
    tray)
      # El radio 21.875% reproduce los 112/512 del tile de la app.
      wrap="display:grid;place-items:center;background:#22C55E;border-radius:21.875%;color:#0F172A"
      svg="$MARK_SVG"
      ;;
    *)
      die "Modo de render desconocido: $mode"
      ;;
  esac

  local html="$TMP_DIR/render-${mode}-${size}.html"
  cat >"$html" <<HTML
<!doctype html>
<html>
<head>
<meta charset="utf-8">
<style>
html, body { margin:0; padding:0; width:100%; height:100%; overflow:hidden; background:transparent; }
.wrap { position:fixed; inset:0; ${wrap}; }
svg { width:100%; height:100%; display:block; }
</style>
</head>
<body><div class="wrap">${svg}</div></body>
</html>
HTML

  local url
  url="$(file_url "$html")"

  "$CHROME_BIN" \
    --headless \
    --disable-gpu \
    --disable-dev-shm-usage \
    --hide-scrollbars \
    --no-first-run \
    --no-default-browser-check \
    --user-data-dir="$(chrome_path "$TMP_DIR/chrome-profile")" \
    --force-device-scale-factor=1 \
    --default-background-color=00000000 \
    --window-size="${size},${size}" \
    --screenshot="$(chrome_path "$out")" \
    "$url" >/dev/null 2>&1 \
    || die "Chrome falló al renderizar $mode a ${size}px"

  [[ -s "$out" ]] || die "Chrome no generó la imagen $out"
}

# Renderiza a `size` px con supersampling 4× para los tamaños chicos y
# normaliza el PNG final (sin metadatos) para que el script sea idempotente.
#
#   render_png <modo> <size> <salida>
render_png() {
  local mode="$1"
  local size="$2"
  local out="$3"
  local px="$size"

  if (( size <= 64 )); then
    px=$(( size * 4 ))
  fi

  local raw="$TMP_DIR/${mode}-${size}-${px}.png"
  render "$mode" "$px" "$raw"

  "$PYTHON_BIN" - "$raw" "$size" "$out" <<'PY'
import sys
from PIL import Image

src, size, out = sys.argv[1], int(sys.argv[2]), sys.argv[3]
try:
    resample = Image.Resampling.LANCZOS
except AttributeError:  # Pillow < 10
    resample = Image.LANCZOS

with Image.open(src) as img:
    img = img.convert("RGBA")
    if img.size != (size, size):
        img = img.resize((size, size), resample)
    img.save(out, format="PNG", optimize=True)
PY
}

echo "🎨 Renderizando iconos desde brand/…"
mkdir -p "$BUILD_DIR/linux" "$BUILD_DIR/windows"

# Icono de la app y assets de Linux.
render_png app 1024 "$BUILD_DIR/appicon.png"
render_png app 256 "$BUILD_DIR/linux/icon-256.png"
render_png tray 64 "$BUILD_DIR/linux/icon.png"

# Ensambla un ICO multi-tamaño a partir de renders del modo indicado.
#
#   build_ico <modo> <salida>
build_ico() {
  local mode="$1"
  local out="$2"
  local sizes=(16 24 32 48 64 128 256)
  local dir="$TMP_DIR/ico-${mode}"

  mkdir -p "$dir"
  for size in "${sizes[@]}"; do
    render_png "$mode" "$size" "$dir/${size}.png"
  done

  "$PYTHON_BIN" - "$dir" "$out" "${sizes[@]}" <<'PY'
import io
import struct
import sys
from pathlib import Path

from PIL import Image

tmp = Path(sys.argv[1])
out = Path(sys.argv[2])
sizes = [int(s) for s in sys.argv[3:]]


def bmp_entry(img):
    """Entrada ICO en BMP 32-bit (XOR + máscara AND), para tamaños chicos."""
    img = img.convert("RGBA")
    w, h = img.size
    px = img.load()

    xor = bytearray()
    for y in range(h - 1, -1, -1):
        for x in range(w):
            r, g, b, a = px[x, y]
            xor += bytes((b, g, r, a))

    # La máscara AND va al final, 1 bit por píxel (1 = transparente),
    # filas de abajo hacia arriba y rellenas a múltiplos de 4 bytes.
    stride = ((w + 31) // 32) * 4
    and_mask = bytearray()
    for y in range(h - 1, -1, -1):
        row = bytearray(stride)
        for x in range(w):
            if px[x, y][3] == 0:
                row[x >> 3] |= 0x80 >> (x & 7)
        and_mask += row

    header = struct.pack(
        "<IiiHHIIiiII",
        40,          # biSize
        w,           # biWidth
        h * 2,       # biHeight: XOR + AND
        1,           # biPlanes
        32,          # biBitCount
        0,           # biCompression (BI_RGB)
        len(xor) + len(and_mask),
        0, 0, 0, 0,
    )
    return header + bytes(xor) + bytes(and_mask)


def png_entry(img):
    """Entrada ICO en PNG (Vista+), para 128 y 256 px."""
    buf = io.BytesIO()
    img.save(buf, format="PNG", optimize=True)
    return buf.getvalue()


entries = []
for size in sizes:
    with Image.open(tmp / f"{size}.png") as img:
        img = img.convert("RGBA")
        data = bmp_entry(img) if size <= 64 else png_entry(img)
    entries.append((size, data))

header = struct.pack("<HHH", 0, 1, len(entries))
offset = 6 + 16 * len(entries)
directory = bytearray()
payload = bytearray()
for size, data in entries:
    dim = 0 if size == 256 else size
    directory += struct.pack(
        "<BBBBHHII", dim, dim, 0, 0, 1, 32, len(data), offset
    )
    offset += len(data)
    payload += data

out.write_bytes(header + bytes(directory) + bytes(payload))

# Verificación: Pillow debe leer el ICO y reconocer todos los tamaños.
with Image.open(out) as img:
    got = sorted(img.info.get("sizes", []))
    expected = [(s, s) for s in sizes]
    if got != expected:
        raise SystemExit(f"ICO inválido: {got} != {expected}")

print(f"   {out.name} → {', '.join(f'{s}px' for s in sizes)}")
PY
}

build_ico app "$BUILD_DIR/windows/icon.ico"
build_ico tray "$BUILD_DIR/windows/tray.ico"

echo
echo "✅ Iconos generados:"
echo "   build/appicon.png         (1024×1024)"
echo "   build/windows/icon.ico    (multi-tamaño, app/.exe)"
echo "   build/windows/tray.ico    (multi-tamaño, bandeja)"
echo "   build/linux/icon.png      (64×64, bandeja)"
echo "   build/linux/icon-256.png  (256×256, menú/.deb)"
echo
echo "Siguiente: recompilar con ./scripts/build-linux.sh o wails build."
