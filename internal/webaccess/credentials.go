package webaccess

// Username devuelve el usuario del servidor web de OpenCode (fijo).
func (s *Service) Username() string { return "opencode" }

// CurrentPassword lee la contraseña actual del servicio.
func (s *Service) CurrentPassword() (string, error) {
	return s.sc.Get("password")
}

// ChangePassword cambia la contraseña y reinicia el servicio para aplicarla.
func (s *Service) ChangePassword(newPassword string) error {
	if err := s.sc.Set("password", newPassword); err != nil {
		return err
	}
	return s.sc.Start()
}
