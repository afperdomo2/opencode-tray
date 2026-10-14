//go:build windows

package servicecontrol

import (
	"fmt"
	"os/exec"
	"strconv"

	"opencode-tray/internal/procs"
)

// killByPID mata el proceso con el PID dado (Windows: taskkill /F).
func killByPID(pid int) error {
	cmd := exec.Command("taskkill", "/PID", strconv.Itoa(pid), "/F")
	procs.SuppressConsole(cmd)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("taskkill %d: %w (%s)", pid, err, string(out))
	}
	return nil
}
