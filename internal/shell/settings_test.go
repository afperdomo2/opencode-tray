package shell

import (
	"os"
	"path/filepath"
	"testing"
)

func tempStore(t *testing.T) *SettingsStore {
	t.Helper()
	return NewSettingsStore(filepath.Join(t.TempDir(), "settings.json"))
}

func TestLoadDefaultsWhenFileMissing(t *testing.T) {
	s := tempStore(t)
	got, err := s.Load()
	if err != nil {
		t.Fatalf("Load() error inesperado: %v", err)
	}
	want := DefaultSettings()
	if got != want {
		t.Errorf("Load() = %+v, want %+v", got, want)
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	s := tempStore(t)
	in := Settings{Language: "en", AutoStart: true, RefreshMin: 30}
	if err := s.Save(in); err != nil {
		t.Fatalf("Save() error: %v", err)
	}
	got, err := s.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if got != in {
		t.Errorf("round-trip = %+v, want %+v", got, in)
	}
}

func TestLoadNormalizesInvalidValues(t *testing.T) {
	s := tempStore(t)
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		t.Fatal(err)
	}
	// idioma inválido y refresco <= 0
	raw := `{"language":"fr","autoStart":true,"refreshMin":0}`
	if err := os.WriteFile(s.path, []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := s.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if got.Language != "es" {
		t.Errorf("Language = %q, want \"es\"", got.Language)
	}
	if got.RefreshMin != 15 {
		t.Errorf("RefreshMin = %d, want 15", got.RefreshMin)
	}
	if !got.AutoStart {
		t.Errorf("AutoStart = false, want true (se conserva)")
	}
}

func TestSaveCreatesParentDir(t *testing.T) {
	s := NewSettingsStore(filepath.Join(t.TempDir(), "a", "b", "settings.json"))
	if err := s.Save(DefaultSettings()); err != nil {
		t.Fatalf("Save() error: %v", err)
	}
	if _, err := os.Stat(s.path); err != nil {
		t.Errorf("archivo no creado: %v", err)
	}
}

func TestLoadCorruptJSON(t *testing.T) {
	s := tempStore(t)
	if err := os.WriteFile(s.path, []byte("{no es json"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Load(); err == nil {
		t.Error("Load() debería fallar con JSON corrupto")
	}
}
