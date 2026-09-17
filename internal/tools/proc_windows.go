//go:build windows

package tools

import (
	"os/exec"
	"syscall"
)

func SetHiddenProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
}

func KillProcessByName(name string) {
	exec.Command("taskkill", "/F", "/IM", name).Run()
}
