//go:build windows

package main

import (
	"fmt"
	"os/exec"
	"syscall"
)

// SetProcessGroup configures the command to create a new process group (Windows).
func SetProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
}

// KillProcessGroup terminates the process group on Windows.
func KillProcessGroup(pid int, sig syscall.Signal) error {
	// On Windows, we use taskkill to forcefully terminate the process tree.
	kill := exec.Command("taskkill", "/F", "/T", "/PID", fmt.Sprintf("%d", pid))
	return kill.Run()
}
