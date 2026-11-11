package main

import (
	"bytes"
	"encoding/binary"
	"image/png"
	"os"
	"testing"
)

// TestTrayIconPNG comprueba el PNG embebido para Linux/macOS: energye/systray
// lo decodifica con image.Decode, así que debe ser un PNG válido y cuadrado.
func TestTrayIconPNG(t *testing.T) {
	img, err := png.Decode(bytes.NewReader(trayIconPNG))
	if err != nil {
		t.Fatalf("el icono de bandeja no es un PNG válido: %v", err)
	}
	if got := img.Bounds().Size(); got.X != 64 || got.Y != 64 {
		t.Fatalf("el icono de bandeja mide %dx%d, quiero 64x64", got.X, got.Y)
	}
}

// TestTrayIconICO valida el ICO multi-tamaño que usa la bandeja en Windows.
func TestTrayIconICO(t *testing.T) {
	checkICO(t, "bandeja", trayIconICO)
}

// TestAppIconICO valida el ICO del .exe y del instalador. No se embebe en Go
// (lo consumen Wails e Inno Setup), así que se lee del repo.
func TestAppIconICO(t *testing.T) {
	data, err := os.ReadFile("build/windows/icon.ico")
	if err != nil {
		t.Fatalf("no pude leer build/windows/icon.ico: %v", err)
	}
	checkICO(t, "app", data)
}

// TestAppIconPNG comprueba el máster de Wails.
func TestAppIconPNG(t *testing.T) {
	for _, tc := range []struct {
		name string
		path string
		want int
	}{
		{"máster Wails", "build/appicon.png", 1024},
		{"menú/.deb", "build/linux/icon-256.png", 256},
	} {
		data, err := os.ReadFile(tc.path)
		if err != nil {
			t.Fatalf("%s: no pude leer %s: %v", tc.name, tc.path, err)
		}
		img, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			t.Fatalf("%s: PNG inválido: %v", tc.name, err)
		}
		if got := img.Bounds().Size(); got.X != tc.want || got.Y != tc.want {
			t.Fatalf("%s: mide %dx%d, quiero %dx%d", tc.name, got.X, got.Y, tc.want, tc.want)
		}
	}
}

// checkICO valida la cabecera, el directorio y los tamaños requeridos de un ICO.
func checkICO(t *testing.T, name string, data []byte) {
	t.Helper()

	const (
		headerSize = 6
		entrySize  = 16
	)
	if len(data) < headerSize {
		t.Fatalf("%s: ICO truncado (%d bytes)", name, len(data))
	}
	reserved := binary.LittleEndian.Uint16(data[0:2])
	typ := binary.LittleEndian.Uint16(data[2:4])
	count := int(binary.LittleEndian.Uint16(data[4:6]))
	if reserved != 0 || typ != 1 {
		t.Fatalf("%s: cabecera ICO inválida (reserved=%d type=%d)", name, reserved, typ)
	}
	if count == 0 {
		t.Fatalf("%s: el ICO no tiene entradas", name)
	}
	dirEnd := headerSize + count*entrySize
	if len(data) < dirEnd {
		t.Fatalf("%s: directorio ICO truncado (%d entradas, %d bytes)", name, count, len(data))
	}

	required := map[int]bool{16: true, 24: true, 32: true, 48: true, 64: true, 128: true, 256: true}
	seen := make(map[int]bool)
	for i := 0; i < count; i++ {
		off := headerSize + i*entrySize
		w := int(data[off])
		h := int(data[off+1])
		if w == 0 {
			w = 256
		}
		if h == 0 {
			h = 256
		}
		if w != h {
			t.Errorf("%s: entrada %d no es cuadrada (%dx%d)", name, i, w, h)
		}
		length := int(binary.LittleEndian.Uint32(data[off+8 : off+12]))
		offset := int(binary.LittleEndian.Uint32(data[off+12 : off+16]))
		if length == 0 {
			t.Errorf("%s: entrada %d sin datos", name, i)
		}
		if offset < dirEnd || offset+length > len(data) {
			t.Errorf("%s: entrada %d fuera del archivo (offset=%d length=%d total=%d)", name, i, offset, length, len(data))
		}
		seen[w] = true
	}
	for size := range required {
		if !seen[size] {
			t.Errorf("%s: falta el tamaño %dpx en el ICO", name, size)
		}
	}
}
