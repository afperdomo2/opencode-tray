package procs

import "os/exec"

// SuppressConsole evita que Windows abra una ventana de consola al ejecutar un
// hijo de consola (opencode.exe, taskkill.exe, code.exe, ...) desde la app de
// bandeja, que no tiene consola propia. Sin esto, cada comando provoca un
// parpadeo de consola. En no-Windows es un no-op.
func SuppressConsole(cmd *exec.Cmd) {
	suppressConsole(cmd)
}
