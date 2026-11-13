package webaccess

import (
	"fmt"
	"time"
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
		// `service set remote true` es la vía oficial: crea el túnel y detiene
		// el servicio. No usamos `pair` porque genera códigos de un solo uso
		// que esta vista no muestra.
		if err := s.sc.SetRemote("true"); err != nil {
			return err
		}
	default:
		return fmt.Errorf("modo inválido: %s", m)
	}
	if err := s.sc.Start(); err != nil {
		return err
	}
	if m == ModeRemote {
		s.waitForRemoteURL()
	}
	return nil
}

// waitForRemoteURL espera a que /api/info incluya la URL del túnel. El
// servicio arranca antes de que OpenTunnel termine de engancharse; sin esta
// espera el panel mostraría solo la URL local justo tras activar el remoto.
func (s *Service) waitForRemoteURL() {
	for attempt := 0; attempt < s.remoteWaitAttempts; attempt++ {
		if info, err := s.sc.Info(); err == nil {
			for _, raw := range info.URLs {
				if classify(raw) == KindRemote {
					return
				}
			}
		}
		if attempt+1 < s.remoteWaitAttempts {
			time.Sleep(s.remoteWaitInterval)
		}
	}
}

// disableRemote apaga OpenTunnel si estaba encendido. Es idempotente.
func (s *Service) disableRemote() error {
	remote, err := s.sc.Get("remote")
	if err != nil || !remoteEnabled(remote) {
		return nil
	}
	return s.sc.SetRemote("false")
}
