//go:build !windows

package procs

import "os/exec"

func suppressConsole(cmd *exec.Cmd) {}
