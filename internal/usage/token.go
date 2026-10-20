package usage

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
)

// Errores estables: el frontend usa el prefijo USAGE_* para traducir el estado
// y decidir si conviene reintentar leyendo el token desde OpenCode.
var (
	ErrNoToken       = errors.New("USAGE_NO_TOKEN: sin token de OpenCode Go")
	ErrInvalidToken  = errors.New("USAGE_INVALID_TOKEN: API key de OpenCode Go inválida")
	ErrNoEntitlement = errors.New("USAGE_NO_ENTITLEMENT: sin suscripción OpenCode Go activa")
	ErrForbidden     = errors.New("USAGE_FORBIDDEN: sin permiso para consultar el uso")
	ErrTransport     = errors.New("USAGE_TRANSPORT")
)

const (
	sourceExport = "opencode-export"
	sourceJSON   = "opencode-json"
	sourceManual = "manual"
	sourceNone   = "none"

	providerGo     = "opencode-go"
	providerLegacy = "opencode"
)

// AuthKey es un candidato de token con su origen.
type AuthKey struct {
	Provider string
	Token    string
	Source   string
}

type legacyCredential struct {
	Type   string `json:"type"`
	Key    string `json:"key"`
	Access string `json:"access"`
}

func (c legacyCredential) token() string {
	if k := strings.TrimSpace(c.Key); k != "" {
		return k
	}
	return strings.TrimSpace(c.Access)
}

// ResolveToken devuelve el primer candidato disponible (export → auth.json → manual).
// No ejecuta red ni el CLI de export.
func (s *Service) ResolveToken() (string, error) {
	candidates := s.tokenCandidates()
	if len(candidates) == 0 {
		return "", ErrNoToken
	}
	return candidates[0].Token, nil
}

// SetManualToken fija un token manual (respaldo) en memoria.
func (s *Service) SetManualToken(tok string) {
	s.manualMu.Lock()
	defer s.manualMu.Unlock()
	s.manualToken = strings.TrimSpace(tok)
}

// ClearManualToken elimina el token manual de memoria.
func (s *Service) ClearManualToken() {
	s.manualMu.Lock()
	defer s.manualMu.Unlock()
	s.manualToken = ""
}

func (s *Service) manualKey() string {
	s.manualMu.RLock()
	defer s.manualMu.RUnlock()
	return s.manualToken
}

// TokenStatus describe el origen del token sin hacer red.
func (s *Service) TokenStatus() TokenStatus {
	if key, ok := s.cachedExport(); ok {
		return TokenStatus{Configured: true, Source: sourceExport, Provider: key.Provider}
	}
	if keys := s.readAuthKeys(); len(keys) > 0 {
		return TokenStatus{Configured: true, Source: sourceJSON, Provider: keys[0].Provider}
	}
	if s.manualKey() != "" {
		return TokenStatus{Configured: true, Source: sourceManual}
	}
	return TokenStatus{Configured: false, Source: sourceNone}
}

// tokenCandidates devuelve los tokens a probar, sin duplicados.
func (s *Service) tokenCandidates() []AuthKey {
	var out []AuthKey
	if key, ok := s.cachedExport(); ok {
		out = append(out, key)
	}
	out = append(out, s.readAuthKeys()...)
	if manual := s.manualKey(); manual != "" {
		out = append(out, AuthKey{Provider: "manual", Token: manual, Source: sourceManual})
	}

	seen := make(map[string]bool, len(out))
	unique := make([]AuthKey, 0, len(out))
	for _, k := range out {
		if k.Token == "" || seen[k.Token] {
			continue
		}
		seen[k.Token] = true
		unique = append(unique, k)
	}
	return unique
}

func (s *Service) cachedExport() (AuthKey, bool) {
	s.exportMu.RLock()
	defer s.exportMu.RUnlock()
	if s.exportToken == "" {
		return AuthKey{}, false
	}
	return AuthKey{Provider: s.exportProvider, Token: s.exportToken, Source: sourceExport}, true
}

func (s *Service) setCachedExport(key AuthKey) {
	s.exportMu.Lock()
	defer s.exportMu.Unlock()
	s.exportToken = key.Token
	s.exportProvider = key.Provider
}

func (s *Service) clearCachedExport() {
	s.exportMu.Lock()
	defer s.exportMu.Unlock()
	s.exportToken = ""
	s.exportProvider = ""
}

// readAuthKeys lee auth.json (formato legado) y devuelve las credenciales de
// OpenCode Go. En v2 el archivo puede estar ausente o desactualizado; por eso
// también existe el export bajo demanda.
func (s *Service) readAuthKeys() []AuthKey {
	var out []AuthKey
	seenFiles := make(map[string]bool, len(s.authFiles))
	for _, path := range s.authFiles {
		if path == "" || seenFiles[path] {
			continue
		}
		seenFiles[path] = true
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		// Estructura: {"opencode-go":{"type":"api","key":"..."}}
		var auth map[string]legacyCredential
		if err := json.Unmarshal(data, &auth); err != nil {
			continue
		}
		for _, provider := range []string{providerGo, providerLegacy} {
			if tok := auth[provider].token(); tok != "" {
				out = append(out, AuthKey{Provider: provider, Token: tok, Source: sourceJSON})
			}
		}
	}
	return out
}
