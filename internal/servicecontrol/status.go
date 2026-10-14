package servicecontrol

import (
	"encoding/json"
	"fmt"
)

// Status describe el estado del servicio de fondo.
type Status struct {
	Running bool   `json:"running"`
	URL     string `json:"url"` // URL del servicio (solo si está corriendo)
}

// ServerInfo es la respuesta de /api/info.
type ServerInfo struct {
	Version string   `json:"version"`
	PID     int      `json:"pid"`
	URLs    []string `json:"urls"`
}

// Status reporta si el servicio está corriendo, usando `opencode service status`.
// A diferencia de `opencode api`, este comando NO arranca el servicio.
func (s *Service) Status() (Status, error) {
	stdout, _, err := s.run("service", "status")
	if err != nil {
		// Detenido u otro problema: lo tratamos como detenido.
		return Status{Running: false}, nil
	}
	return Status{Running: true, URL: trim(stdout)}, nil
}

// Info obtiene la información del servidor (/api/info), incluido el PID.
// Nota: `opencode api` puede arrancar el servicio si no hay uno sano; usarlo
// solo cuando se espere que esté corriendo (p. ej. dentro de ForceStop).
func (s *Service) Info() (ServerInfo, error) {
	stdout, _, err := s.run("api", "get", "/api/info")
	if err != nil {
		return ServerInfo{}, fmt.Errorf("api get /api/info: %w", err)
	}
	var info ServerInfo
	if err := json.Unmarshal([]byte(stdout), &info); err != nil {
		return ServerInfo{}, fmt.Errorf("parsear /api/info: %w", err)
	}
	return info, nil
}
