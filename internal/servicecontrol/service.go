package servicecontrol

import (
	"strings"
	"time"

	"opencode-tray/internal/cli"
)

// runner ejecuta el CLI de OpenCode. Es una interfaz para poder mockear el
// comando en los tests sin invocar `opencode` real.
type runner interface {
	Run(args []string) (stdout string, stderr string, err error)
}

// Service controla el servicio de fondo de OpenCode.
type Service struct {
	runner     runner
	slowRunner runner // operaciones que pueden tardar (p. ej. crear el túnel remoto)
	kill       func(pid int) error
}

// New crea un Service que invoca el CLI de OpenCode (se resuelve con
// cli.Resolve para funcionar también desde el menú o el auto-inicio).
func New() *Service {
	bin := cli.Resolve()
	return &Service{
		runner:     execRunner{bin: bin, timeout: 15 * time.Second},
		slowRunner: execRunner{bin: bin, timeout: 2 * time.Minute},
		kill:       killByPID,
	}
}

// NewWithRunner crea un Service con runner y kill inyectados (para tests).
func NewWithRunner(r runner, kill func(int) error) *Service {
	if kill == nil {
		kill = killByPID
	}
	return &Service{runner: r, slowRunner: r, kill: kill}
}

// run ejecuta el CLI con los argumentos dados.
func (s *Service) run(args ...string) (string, string, error) {
	return s.runner.Run(args)
}

func trim(s string) string { return strings.TrimSpace(s) }
