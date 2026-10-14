package shell

import (
	"errors"
	"testing"
)

type mockRunKey struct{ values map[string]string }

func newMockRunKey() *mockRunKey { return &mockRunKey{values: map[string]string{}} }

func (m *mockRunKey) Set(name, value string) error { m.values[name] = value; return nil }
func (m *mockRunKey) Delete(name string) error {
	delete(m.values, name)
	return nil
}
func (m *mockRunKey) Get(name string) (string, bool, error) {
	v, ok := m.values[name]
	return v, ok, nil
}

func TestAutoStartEnableDisable(t *testing.T) {
	key := newMockRunKey()
	a := NewAutoStart(key, func() (string, error) {
		return `C:\apps\opencode-tray.exe`, nil
	})

	if on, err := a.Enabled(); err != nil || on {
		t.Fatalf("debería iniciar deshabilitado: on=%v err=%v", on, err)
	}

	if err := a.SetEnabled(true); err != nil {
		t.Fatalf("SetEnabled(true): %v", err)
	}
	if on, _ := a.Enabled(); !on {
		t.Error("debería estar habilitado")
	}
	if got := key.values["OpenCodeTray"]; got != `"C:\apps\opencode-tray.exe"` {
		t.Errorf("valor clave = %q, want con comillas", got)
	}

	if err := a.SetEnabled(false); err != nil {
		t.Fatalf("SetEnabled(false): %v", err)
	}
	if on, _ := a.Enabled(); on {
		t.Error("debería estar deshabilitado")
	}
	if _, ok := key.values["OpenCodeTray"]; ok {
		t.Error("la clave debería eliminarse")
	}
}

func TestAutoStartExePathError(t *testing.T) {
	a := NewAutoStart(newMockRunKey(), func() (string, error) {
		return "", errors.New("fallo")
	})
	if err := a.SetEnabled(true); err == nil {
		t.Error("debería propagar el error de exePath")
	}
}
