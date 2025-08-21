//go:build windows

package manager

import (
	"os"
    "os/exec"
)

func (m *Manager) setProcAttr(cmd *exec.Cmd) {
    // Windows doesn't support process groups in the same way
    // cmd.SysProcAttr can be left nil or set other Windows-specific attributes
}

func (m *Manager) killProcessGroup(pid int) error {
    // On Windows, use os.Process to kill the process
    process, err := os.FindProcess(pid)
    if err != nil {
        return err
    }
    return process.Kill()
}

func (m *Manager) isProcessAlive(pid int) bool {
    process, err := os.FindProcess(pid)
    if err != nil {
        return false
    }
    // On Windows, FindProcess always succeeds, so we need to check if we can signal it
    err = process.Signal(os.Interrupt)
    return err == nil
}