//go:build !windows

package usage

import "os/exec"

func hideConsole(_ *exec.Cmd) {}
