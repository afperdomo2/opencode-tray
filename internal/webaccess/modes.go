package webaccess

import (
	"fmt"
	"strings"
)

// ApplyMode aplica un modo de acceso usando service-control.
//   - local:   hostname 127.0.0.1 (solo esta máquina)
//   - network: hostname 0.0.0.0 (todas las interfaces)
//   - remote:  túnel OpenTunnel sobre un bind local
//
// Los tres modos son exclusivos: al pasar a local/red se desactiva el remoto
// para no dejar el servidor expuesto a Internet sin que el usuario lo vea.
//
// Nota: un `set` de configuración detiene el servicio; se vuelve a arrancar.
func (s *Service) ApplyMode(m Mode) error {
	switch m {
	case ModeLocal:
		if err := s.disableRemote(); err != nil {
			return err
		}
		if err := s.sc.Set("hostname", "127.0.0.1"); err != nil {
			return err
		}
	case ModeNetwork:
		if err := s.disableRemote(); err != nil {
			return err
		}
		if err := s.sc.Set("hostname", "0.0.0.0"); err != nil {
			return err
		}
	case ModeRemote:
		// El túnel llega al puerto local: bind loopback para no exponer además
		// la red LAN, que sería un efecto colateral del modo "remoto".
		if err := s.sc.Set("hostname", "127.0.0.1"); err != nil {
			return err
		}
		if _, err := s.sc.RunRaw("pair", "--remote", "--yes"); err != nil {
			return err
		}
	default:
		return fmt.Errorf("modo inválido: %s", m)
	}
	return s.sc.Start()
}

// disableRemote apaga OpenTunnel si estaba encendido. Es idempotente.
func (s *Service) disableRemote() error {
	remote, err := s.sc.Get("remote")
	if err != nil || !strings.EqualFold(strings.TrimSpace(remote), "true") {
		return nil
	}
	return s.sc.Set("remote", "false")
}
