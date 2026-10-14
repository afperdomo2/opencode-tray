package configviewer

import (
	"errors"
	"os"
	"path/filepath"
)

// ConfigDir devuelve el directorio de configuración de OpenCode.
func (s *Service) ConfigDir() string { return s.configDir }

// ConfigPath devuelve la ruta del config global (opencode.jsonc u opencode.json).
func (s *Service) ConfigPath() (string, error) {
	for _, name := range []string{"opencode.jsonc", "opencode.json"} {
		p := filepath.Join(s.configDir, name)
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	return "", errors.New("no se encontró opencode.json en " + s.configDir)
}

// ConfigJSON lee el config global y devuelve su contenido con secretos enmascarados.
func (s *Service) ConfigJSON() (string, error) {
	p, err := s.ConfigPath()
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return "", err
	}
	return MaskSecrets(string(data)), nil
}
