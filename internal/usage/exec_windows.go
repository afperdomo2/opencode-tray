//go:build windows

package usage

import (
	"os/exec"
	"syscall"
)

// hideConsole evita que se abra una consola al ejecutar `opencode auth export`.
func hideConsole(cmd *exec.Cmd) {
	const createNoWindow = 0x08000000
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: createNoWindow,
	}
}
