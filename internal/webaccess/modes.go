package webaccess

import "fmt"

// ApplyMode aplica un modo de acceso usando service-control.
//   - local:   hostname 127.0.0.1 (solo esta máquina)
//   - network: hostname 0.0.0.0 (todas las interfaces)
//   - remote:  abre un túnel con `opencode pair --remote --yes`
//
// Nota: un `set` de configuración detiene el servicio; se vuelve a arrancar.
func (s *Service) ApplyMode(m Mode) error {
	switch m {
	case ModeLocal:
		if err := s.sc.Set("hostname", "127.0.0.1"); err != nil {
			return err
		}
	case ModeNetwork:
		if err := s.sc.Set("hostname", "0.0.0.0"); err != nil {
			return err
		}
	case ModeRemote:
		_, err := s.sc.RunRaw("pair", "--remote", "--yes")
		return err
	default:
		return fmt.Errorf("modo inválido: %s", m)
	}
	return s.sc.Start()
}
