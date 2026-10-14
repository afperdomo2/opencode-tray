package usage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

// ResolveToken devuelve el token a usar: el manual si está configurado, si no
// el de `opencode-go` en auth.json (tras hacer /connect en OpenCode).
func (s *Service) ResolveToken() (string, error) {
	if s.manualToken != "" {
		return s.manualToken, nil
	}
	return s.readAuthKey()
}

// SetManualToken fija un token manual (desde Ajustes).
func (s *Service) SetManualToken(tok string) { s.manualToken = tok }

// ClearManualToken elimina el token manual (vuelve a auth.json).
func (s *Service) ClearManualToken() { s.manualToken = "" }

// readAuthKey lee la key de opencode-go desde auth.json.
func (s *Service) readAuthKey() (string, error) {
	data, err := os.ReadFile(s.authFile)
	if err != nil {
		return "", fmt.Errorf("leer auth.json: %w", err)
	}
	// Estructura: {"opencode-go":{"type":"api","key":"..."}}
	var auth map[string]struct {
		Key string `json:"key"`
	}
	if err := json.Unmarshal(data, &auth); err != nil {
		return "", fmt.Errorf("parsear auth.json: %w", err)
	}
	if k := strings.TrimSpace(auth["opencode-go"].Key); k != "" {
		return k, nil
	}
	return "", errors.New("sin token de opencode-go en auth.json")
}
