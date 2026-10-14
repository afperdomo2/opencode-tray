//go:build !windows

package servicecontrol

import "os/exec"

// suppressConsole no hace nada fuera de Windows.
func suppressConsole(cmd *exec.Cmd) {}
