package servicecontrol

import "fmt"

// Get lee una configuración del servicio (hostname, port, password, ...).
func (s *Service) Get(key string) (string, error) {
	stdout, stderr, err := s.run("service", "get", key)
	if err != nil {
		return "", fmt.Errorf("service get %s: %w (%s)", key, err, trim(stderr))
	}
	return trim(stdout), nil
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

// Unset restablece una configuración del servicio a su valor por defecto.
func (s *Service) Unset(key string) error {
	_, stderr, err := s.run("service", "unset", key)
	if err != nil {
		return fmt.Errorf("service unset %s: %w (%s)", key, err, trim(stderr))
	}
	return nil
}
