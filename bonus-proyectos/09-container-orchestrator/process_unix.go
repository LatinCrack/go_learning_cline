//go:build !windows

package main

import (
	"os/exec"
	"syscall"
)

// SetProcessGroup configures the command to run in its own process group (Unix only).
func SetProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
}

// KillProcessGroup sends a signal to the entire process group.
func KillProcessGroup(pid int, sig syscall.Signal) error {
	return syscall.Kill(-pid, sig)
}
