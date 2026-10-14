package servicecontrol

import (
	"bytes"
	"context"
	"os/exec"
	"time"
)

// execRunner ejecuta el binario de OpenCode con un timeout.
type execRunner struct {
	bin     string
	timeout time.Duration
}

// Run ejecuta `bin args...` y devuelve stdout, stderr y el error
// (no nulo cuando el código de salida es distinto de cero).
func (e execRunner) Run(args []string) (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, e.bin, args...)
	suppressConsole(cmd)
	var out, errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut

	err := cmd.Run()
	return out.String(), errOut.String(), err
}
