//go:build windows

package servicecontrol

import (
	"os/exec"
	"syscall"

	"golang.org/x/sys/windows"
)

// suppressConsole evita que Windows abra una ventana de consola al ejecutar un
// hijo de consola (opencode.exe, taskkill.exe) desde la app de bandeja, que no
// tiene consola propia. Sin esto, cada comando provoca un parpadeo de consola.
func suppressConsole(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.CreationFlags |= windows.CREATE_NO_WINDOW
}
