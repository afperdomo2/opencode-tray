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

// Fetch consulta el consumo probando los candidatos en orden: export de
// OpenCode v2 → auth.json legado → token manual.
func (s *Service) Fetch() (Usage, error) {
	candidates := s.tokenCandidates()
	if len(candidates) == 0 {
		return Usage{}, ErrNoToken
	}

	var sawInvalid, sawEntitlement, sawForbidden bool
	for _, candidate := range candidates {
		u, err := s.fetchWithToken(candidate.Token)
		if err == nil {
			return u, nil
		}
		switch {
		case errors.Is(err, ErrTransport):
			return Usage{}, err
		case errors.Is(err, ErrInvalidToken):
			sawInvalid = true
		case errors.Is(err, ErrNoEntitlement):
			sawEntitlement = true
		case errors.Is(err, ErrForbidden):
			sawForbidden = true
		default:
			return Usage{}, err
		}
	}

	switch {
	case sawInvalid:
		return Usage{}, ErrInvalidToken
	case sawEntitlement:
		return Usage{}, ErrNoEntitlement
	case sawForbidden:
		return Usage{}, ErrForbidden
	default:
		return Usage{}, ErrNoToken
	}
}

// ValidateToken comprueba un token con una petición directa, sin tocar los
// candidatos configurados.
func (s *Service) ValidateToken(tok string) error {
	if tok == "" {
		return ErrNoToken
	}
	_, err := s.fetchWithToken(tok)
	return err
}

func (s *Service) fetchWithToken(key string) (Usage, error) {
	req, err := http.NewRequest(http.MethodGet, s.baseURL+"/zen/go/v1/usage", nil)
	if err != nil {
		return Usage{}, err
	}
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return Usage{}, fmt.Errorf("%w: %v", ErrTransport, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Usage{}, fmt.Errorf("%w: %v", ErrTransport, err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		// continuar
	case http.StatusUnauthorized:
		return Usage{}, ErrInvalidToken
	case http.StatusForbidden:
		var e struct {
			Error struct {
				Type string `json:"type"`
			} `json:"error"`
		}
		if json.Unmarshal(body, &e) == nil && e.Error.Type == "EntitlementError" {
			return Usage{}, ErrNoEntitlement
		}
		return Usage{}, ErrForbidden
	default:
		return Usage{}, fmt.Errorf("USAGE_HTTP_%d: respuesta inesperada", resp.StatusCode)
	}

	var parsed struct {
		Usage *struct {
			Rolling *apiWindow `json:"rolling"`
			Weekly  *apiWindow `json:"weekly"`
			Monthly *apiWindow `json:"monthly"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil || parsed.Usage == nil {
		return Usage{}, errors.New("USAGE_PARSE: no se pudo parsear la respuesta de uso")
	}

	return Usage{
		Rolling: parsed.Usage.Rolling.toWindow(),
		Weekly:  parsed.Usage.Weekly.toWindow(),
		Monthly: parsed.Usage.Monthly.toWindow(),
	}, nil
}
