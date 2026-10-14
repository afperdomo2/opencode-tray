package usage

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// apiWindow es la forma de una ventana en la respuesta de /zen/go/v1/usage.
type apiWindow struct {
	Percent  *float64 `json:"percent"`
	ResetsAt *string  `json:"resetsAt"`
}

func (w *apiWindow) toWindow() Window {
	if w == nil || w.Percent == nil {
		return Window{}
	}
	out := Window{Percent: *w.Percent}
	if w.ResetsAt != nil {
		if t, err := time.Parse(time.RFC3339Nano, *w.ResetsAt); err == nil {
			out.ResetsAt, out.HasReset = t, true
		}
	}
	return out
}

// Fetch consulta GET /zen/go/v1/usage y devuelve las tres ventanas de consumo.
func (s *Service) Fetch() (Usage, error) {
	key, err := s.ResolveToken()
	if err != nil {
		return Usage{}, err
	}

	req, err := http.NewRequest(http.MethodGet, s.baseURL+"/zen/go/v1/usage", nil)
	if err != nil {
		return Usage{}, err
	}
	req.Header.Set("Authorization", "Bearer "+key)

	resp, err := s.client.Do(req)
	if err != nil {
		return Usage{}, fmt.Errorf("consultar uso: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Usage{}, fmt.Errorf("leer respuesta: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		// continuar
	case http.StatusUnauthorized:
		return Usage{}, errors.New("API key de OpenCode Go inválida")
	case http.StatusForbidden:
		var e struct {
			Error struct {
				Type string `json:"type"`
			} `json:"error"`
		}
		if json.Unmarshal(body, &e) == nil && e.Error.Type == "EntitlementError" {
			return Usage{}, errors.New("sin suscripción OpenCode Go activa")
		}
		return Usage{}, errors.New("sin permiso para consultar el uso")
	default:
		return Usage{}, fmt.Errorf("respuesta inesperada: HTTP %d", resp.StatusCode)
	}

	var parsed struct {
		Usage *struct {
			Rolling *apiWindow `json:"rolling"`
			Weekly  *apiWindow `json:"weekly"`
			Monthly *apiWindow `json:"monthly"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil || parsed.Usage == nil {
		return Usage{}, errors.New("no se pudo parsear la respuesta de uso")
	}

	return Usage{
		Rolling: parsed.Usage.Rolling.toWindow(),
		Weekly:  parsed.Usage.Weekly.toWindow(),
		Monthly: parsed.Usage.Monthly.toWindow(),
	}, nil
}
