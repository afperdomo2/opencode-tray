package configviewer

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"opencode-tray/internal/cli"
	"opencode-tray/internal/procs"
)

// runner ejecuta el CLI de OpenCode. Interfaz para mockear en tests.
// IMPORTANTE: no usar métodos variádicos (rompen los bindings de Wails).
type runner interface {
	Run(args []string) (stdout string, stderr string, err error)
}

// MCPServer es un servidor MCP con su estado de conexión.
type MCPServer struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

// Service da acceso a la configuración de OpenCode (MCP, config global).
type Service struct {
	runner    runner
	configDir string
	lookPath  func(string) (string, error)
	open      func(args ...string) error
}

// New crea un Service apuntando a ~/.config/opencode y al CLI de OpenCode
// (resuelto con cli.Resolve para funcionar desde el menú y el auto-inicio).
func New() *Service {
	home, _ := os.UserHomeDir()
	return &Service{
		runner:    execRunner{bin: cli.Resolve(), timeout: 15 * time.Second},
		configDir: filepath.Join(home, ".config", "opencode"),
		lookPath:  exec.LookPath,
		open:      openExec,
	}
}

// NewWithRunner crea un Service con runner y configDir inyectados (para tests).
func NewWithRunner(r runner, configDir string) *Service {
	return &Service{
		runner:    r,
		configDir: configDir,
		lookPath:  exec.LookPath,
		open:      openExec,
	}
}

// execRunner ejecuta el binario de OpenCode con un timeout, sin ventana de consola.
type execRunner struct {
	bin     string
	timeout time.Duration
}

func (e execRunner) Run(args []string) (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, e.bin, args...)
	procs.SuppressConsole(cmd)
	var out, errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	err := cmd.Run()
	return out.String(), errOut.String(), err
}
