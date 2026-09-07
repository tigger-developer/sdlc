//go:build !unix

package harness

import "os/exec"

func configureProcessGroup(*exec.Cmd) {}

func terminateProcessGroup(process *exec.Cmd) error {
	if process.Process == nil {
		return nil
	}
	return process.Process.Kill()
}
