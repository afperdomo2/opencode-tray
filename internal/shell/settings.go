package shell

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Settings es la configuración persistente de la aplicación (local, no de OpenCode).
type Settings struct {
	Language   string `json:"language"`             // "es" | "en"
	AutoStart  bool   `json:"autoStart"`            // arrancar con Windows
	RefreshMin int    `json:"refreshMin"`           // intervalo de refresco (min)
	UsageToken string `json:"usageToken,omitempty"` // blob DPAPI del token manual de OpenCode Go
}

// Public devuelve una copia sin secretos, apta para enviar al frontend.
func (s Settings) Public() Settings {
	s.UsageToken = ""
	return s
}

// DefaultSettings devuelve los valores por defecto.
func DefaultSettings() Settings {
	return Settings{
		Language:   "es",
		AutoStart:  false,
		RefreshMin: 15,
	}
}

// SettingsStore carga y guarda Settings en un archivo JSON.
// La ruta es inyectable para facilitar los tests.
type SettingsStore struct {
	path string
}

// NewSettingsStore crea un store que persiste en path.
func NewSettingsStore(path string) *SettingsStore {
	return &SettingsStore{path: path}
}

// DefaultPath devuelve la ruta por defecto de los ajustes
// (%APPDATA%\OpenCodeTray\settings.json en Windows).
func DefaultPath() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "OpenCodeTray", "settings.json"), nil
}

// Load lee los ajustes. Si el archivo no existe, devuelve los valores por defecto.
func (s *SettingsStore) Load() (Settings, error) {
	st := DefaultSettings()
	data, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return st, nil
		}
		return st, err
	}
	if err := json.Unmarshal(data, &st); err != nil {
		return st, fmt.Errorf("settings corruptos en %s: %w", s.path, err)
	}
	return normalize(st), nil
}

// Save persiste los ajustes, creando el directorio si hace falta.
func (s *SettingsStore) Save(st Settings) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(normalize(st), "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o600)
}

// normalize corrige valores fuera de rango para que siempre sean usables.
func normalize(st Settings) Settings {
	if st.Language != "es" && st.Language != "en" {
		st.Language = "es"
	}
	if st.RefreshMin <= 0 {
		st.RefreshMin = 15
	}
	return st
}
