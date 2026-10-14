//go:build !windows

package servicecontrol

import (
	"fmt"
	"os/exec"
	"strconv"
)

// killByPID mata el proceso con el PID dado (Unix: kill -9).
func killByPID(pid int) error {
	out, err := exec.Command("kill", "-9", strconv.Itoa(pid)).CombinedOutput()
	if err != nil {
		return fmt.Errorf("kill %d: %w (%s)", pid, err, string(out))
	}
	return nil
}
