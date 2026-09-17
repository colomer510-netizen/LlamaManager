//go:build !windows

package tools

import (
	"os/exec"
)

func SetHiddenProcess(cmd *exec.Cmd) {
	// No-op en Linux / macOS
}

func KillProcessByName(name string) {
	exec.Command("pkill", "-f", name).Run()
}
