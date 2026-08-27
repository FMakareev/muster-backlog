//go:build unix

package whichbin

import (
	"os/exec"
	"syscall"
)

// detach puts the shell in a process group of its own, so that killing it on
// timeout also reaches whatever its profile started.
func detach(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return nil
		}
		// The negative pid is the group. Falling back to the process alone
		// matters when Setpgid did not take.
		if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL); err == nil {
			return nil
		}
		return cmd.Process.Kill()
	}
}
