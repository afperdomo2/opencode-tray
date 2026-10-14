package servicecontrol

import "fmt"

// Start arranca el servicio de fondo.
func (s *Service) Start() error {
	_, stderr, err := s.run("service", "start")
	if err != nil {
		return fmt.Errorf("service start: %w (%s)", err, trim(stderr))
	}
	return nil
}

// Stop intenta detener el servicio normalmente (puede fallar por el bug conocido).
func (s *Service) Stop() error {
	_, stderr, err := s.run("service", "stop")
	if err != nil {
		return fmt.Errorf("service stop: %w (%s)", err, trim(stderr))
	}
	return nil
}

// Restart reinicia el servicio.
func (s *Service) Restart() error {
	_, stderr, err := s.run("service", "restart")
	if err != nil {
		return fmt.Errorf("service restart: %w (%s)", err, trim(stderr))
	}
	return nil
}

// ForceStop detiene el servicio de forma fiable, sorteando el bug de
// `opencode service stop` (que a veces no mata el proceso):
//  1. intenta un stop normal,
//  2. verifica con Status,
//  3. si sigue vivo, obtiene el PID de /api/info y lo mata.
func (s *Service) ForceStop() error {
	_ = s.Stop()

	st, err := s.Status()
	if err == nil && !st.Running {
		return nil
	}

	info, err := s.Info()
	if err != nil {
		return fmt.Errorf("no se pudo obtener el PID: %w", err)
	}
	if info.PID <= 0 {
		return fmt.Errorf("PID inválido: %d", info.PID)
	}
	return s.kill(info.PID)
}
