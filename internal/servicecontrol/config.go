package servicecontrol

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// Get lee una configuración del servicio (hostname, port, password, ...).
func (s *Service) Get(key string) (string, error) {
	stdout, stderr, err := s.run("service", "get", key)
	if err != nil {
		return "", fmt.Errorf("service get %s: %w (%s)", key, err, trim(stderr))
	}
	return trim(stdout), nil
}

// Snapshot lee toda la configuración del servicio en una sola invocación del
// CLI y la devuelve como mapa de strings (la contraseña no se incluye).
// El CLI omite las claves con valor por defecto (p. ej. "remote"), por lo que
// la ausencia debe interpretarse como el valor por defecto de OpenCode.
func (s *Service) Snapshot() (map[string]string, error) {
	stdout, stderr, err := s.run("service", "get")
	if err != nil {
		return nil, fmt.Errorf("service get: %w (%s)", err, trim(stderr))
	}

	var raw map[string]any
	if err := json.Unmarshal([]byte(stdout), &raw); err != nil {
		return nil, fmt.Errorf("parsear service get: %w", err)
	}

	out := make(map[string]string, len(raw))
	for key, value := range raw {
		switch v := value.(type) {
		case string:
			out[key] = v
		case bool:
			out[key] = strconv.FormatBool(v)
		case float64:
			out[key] = strconv.FormatFloat(v, 'f', -1, 64)
		case map[string]any:
			// OpenCode 2.x representa `remote` como objeto {route} (o {} cuando
			// el CLI redacta la ruta). Su presencia significa túnel activo; no
			// copiamos la ruta porque es tan sensible como la contraseña.
			if key == "remote" {
				out[key] = "true"
				continue
			}
			if encoded, err := json.Marshal(v); err == nil {
				out[key] = string(encoded)
			}
		}
	}
	return out, nil
}

// Set escribe una configuración del servicio.
// Nota: cambiar una configuración detiene el servicio; hay que hacer Start() después.
func (s *Service) Set(key, value string) error {
	_, stderr, err := s.run("service", "set", key, value)
	if err != nil {
		return fmt.Errorf("service set %s: %w (%s)", key, err, trim(stderr))
	}
	return nil
}

// SetRemote activa o desactiva OpenTunnel. Encender puede tardar hasta un
// minuto en crear el túnel, por eso usa un timeout mayor que el resto de
// operaciones del CLI.
func (s *Service) SetRemote(value string) error {
	if value != "true" && value != "false" {
		return fmt.Errorf("valor remoto inválido: %s", value)
	}
	_, stderr, err := s.slowRunner.Run([]string{"service", "set", "remote", value})
	if err != nil {
		return fmt.Errorf("service set remote %s: %w (%s)", value, err, trim(stderr))
	}
	return nil
}

// Unset restablece una configuración del servicio a su valor por defecto.
func (s *Service) Unset(key string) error {
	_, stderr, err := s.run("service", "unset", key)
	if err != nil {
		return fmt.Errorf("service unset %s: %w (%s)", key, err, trim(stderr))
	}
	return nil
}
