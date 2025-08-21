//go:build unix

package manager

import (
    "os/exec"
    "syscall"
)

func (m *Manager) setProcAttr(cmd *exec.Cmd) {
    cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

func (m *Manager) killProcessGroup(pid int) error {
    pgid, err := syscall.Getpgid(pid)
    if err != nil {
        return syscall.Kill(-pid, syscall.SIGTERM)
    }
    return syscall.Kill(-pgid, syscall.SIGTERM)
}

func (m *Manager) isProcessAlive(pid int) bool {
    return syscall.Kill(pid, 0) == nil
}
