package configviewer

import (
	"os/exec"

	"opencode-tray/internal/procs"
)

// OpenInEditor abre un archivo: VS Code si está en el PATH, si no el editor
// por defecto del SO.
func (s *Service) OpenInEditor(path string) error {
	if code, err := s.lookPath("code"); err == nil {
		return s.open(append([]string{code}, path)...)
	}
	return s.open(defaultOpenArgs(path)...)
}

// OpenConfigDir abre el directorio de configuración en el explorador.
func (s *Service) OpenConfigDir() error {
	return s.open(defaultOpenArgs(s.configDir)...)
}

// openExec ejecuta el comando de apertura sin ventana de consola.
func openExec(args ...string) error {
	cmd := exec.Command(args[0], args[1:]...)
	procs.SuppressConsole(cmd)
	return cmd.Start()
}
