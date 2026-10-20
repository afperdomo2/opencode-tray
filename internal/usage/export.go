package usage

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Runner ejecuta el CLI de OpenCode. Se inyecta para poder testear sin CLI real.
type Runner interface {
	Run(ctx context.Context, args []string) ([]byte, error)
}

type execRunner struct {
	bin string
}

func (e execRunner) Run(ctx context.Context, args []string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, e.bin, args...)
	hideConsole(cmd)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	// stderr se descarta a propósito: podría contener rutas o detalles internos.
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	return stdout.Bytes(), nil
}

var errNoExportedToken = errors.New("sin credenciales de OpenCode Go en auth export")

type exportValue struct {
	Type   string `json:"type"`
	Key    string `json:"key"`
	Access string `json:"access"`
}

type exportItem struct {
	IntegrationID string      `json:"integrationID"`
	Provider      string      `json:"provider"`
	Active        *bool       `json:"active"`
	Value         exportValue `json:"value"`

	// Forma plana por si el export cambia a un objeto tipo auth.json.
	Type   string `json:"type"`
	Key    string `json:"key"`
	Access string `json:"access"`
}

func (i exportItem) providerID() string {
	if p := strings.TrimSpace(i.IntegrationID); p != "" {
		return p
	}
	return strings.TrimSpace(i.Provider)
}

func (i exportItem) token() string {
	if k := strings.TrimSpace(i.Value.Key); k != "" {
		return k
	}
	if a := strings.TrimSpace(i.Value.Access); a != "" {
		return a
	}
	if k := strings.TrimSpace(i.Key); k != "" {
		return k
	}
	return strings.TrimSpace(i.Access)
}

// ReloadToken invalida el token exportado y lo vuelve a leer desde OpenCode.
// No valida contra el endpoint de uso; el siguiente Fetch lo hará.
func (s *Service) ReloadToken(ctx context.Context) (TokenStatus, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	s.clearCachedExport()

	key, err := s.exportFromCLI(ctx)
	if err != nil {
		if errors.Is(err, errNoExportedToken) {
			return s.TokenStatus(), nil
		}
		return s.TokenStatus(), err
	}
	s.setCachedExport(key)
	return s.TokenStatus(), nil
}

func (s *Service) exportFromCLI(ctx context.Context) (AuthKey, error) {
	if s.runner == nil {
		return AuthKey{}, errors.New("no hay lector de credenciales de OpenCode")
	}

	fastCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	out, err := s.runner.Run(fastCtx, []string{"auth", "export"})
	cancel()
	if err == nil && len(bytes.TrimSpace(out)) > 0 {
		if key, perr := parseAuthExport(out); perr == nil {
			return key, nil
		}
	}

	slowCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	out, err = s.runner.Run(slowCtx, []string{"auth", "export", "--standalone"})
	cancel()
	if err != nil {
		return AuthKey{}, fmt.Errorf("no se pudo leer el token desde OpenCode: %w", err)
	}
	return parseAuthExport(out)
}

// parseAuthExport acepta un array de credenciales o un objeto indexado por
// proveedor. Elige primero opencode-go y luego opencode, solo si están activos.
func parseAuthExport(data []byte) (AuthKey, error) {
	trimmed := bytes.TrimPrefix(bytes.TrimSpace(data), []byte{0xEF, 0xBB, 0xBF})
	if len(trimmed) == 0 {
		return AuthKey{}, errNoExportedToken
	}

	items := make([]exportItem, 0, 4)
	if err := json.Unmarshal(trimmed, &items); err != nil {
		var byProvider map[string]exportItem
		if err2 := json.Unmarshal(trimmed, &byProvider); err2 != nil {
			return AuthKey{}, errors.New("respuesta de auth export no reconocida")
		}
		for provider, item := range byProvider {
			if item.IntegrationID == "" && item.Provider == "" {
				item.Provider = provider
			}
			items = append(items, item)
		}
	}

	for _, preferred := range []string{providerGo, providerLegacy} {
		for _, item := range items {
			if item.providerID() != preferred {
				continue
			}
			if item.Active != nil && !*item.Active {
				continue
			}
			if tok := item.token(); tok != "" {
				return AuthKey{Provider: preferred, Token: tok, Source: sourceExport}, nil
			}
		}
	}
	return AuthKey{}, errNoExportedToken
}
