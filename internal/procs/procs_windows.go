//go:build windows

package procs

import (
	"os/exec"
	"syscall"

	"golang.org/x/sys/windows"
)

func suppressConsole(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.CreationFlags |= windows.CREATE_NO_WINDOW
}
