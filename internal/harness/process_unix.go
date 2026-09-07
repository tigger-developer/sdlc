//go:build unix

package harness

import (
	"errors"
	"os"
	"os/exec"
	"syscall"
)

func configureProcessGroup(process *exec.Cmd) {
	process.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

func terminateProcessGroup(process *exec.Cmd) error {
	if process.Process == nil {
		return nil
	}
	err := syscall.Kill(-process.Process.Pid, syscall.SIGKILL)
	if errors.Is(err, os.ErrProcessDone) || errors.Is(err, syscall.ESRCH) {
		return nil
	}
	return err
}
