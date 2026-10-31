//go:build linux

package shell

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestXDGRunKeyRoundTrip(t *testing.T) {
	dir := t.TempDir()
	a := NewAutoStart(newXDGRunKey(dir), func() (string, error) {
		return "/opt/opencode tray/opencode-tray", nil
	})

	if on, err := a.Enabled(); err != nil || on {
		t.Fatalf("debería iniciar deshabilitado: on=%v err=%v", on, err)
	}

	if err := a.SetEnabled(true); err != nil {
		t.Fatalf("SetEnabled(true): %v", err)
	}
	if on, err := a.Enabled(); err != nil || !on {
		t.Fatalf("debería estar habilitado: on=%v err=%v", on, err)
	}

	data, err := os.ReadFile(filepath.Join(dir, desktopFileName))
	if err != nil {
		t.Fatalf("leer .desktop: %v", err)
	}
	entry := string(data)
	for _, want := range []string{
		"[Desktop Entry]",
		"Type=Application",
		`Exec="/opt/opencode tray/opencode-tray"`,
		"Terminal=false",
	} {
		if !strings.Contains(entry, want) {
			t.Errorf("falta %q en la entrada:\n%s", want, entry)
		}
	}

	if err := a.SetEnabled(false); err != nil {
		t.Fatalf("SetEnabled(false): %v", err)
	}
	if on, _ := a.Enabled(); on {
		t.Error("debería estar deshabilitado tras borrar la entrada")
	}
	if _, err := os.Stat(filepath.Join(dir, desktopFileName)); !os.IsNotExist(err) {
		t.Errorf("la entrada debería eliminarse: %v", err)
	}
}

func TestXDGRunKeySetSinComando(t *testing.T) {
	if err := newXDGRunKey(t.TempDir()).Set("OpenCodeTray", "  "); err == nil {
		t.Error("un comando vacío debería devolver error")
	}
}
