//go:build !unix

package whichbin

import "os/exec"

// detach does nothing where there is no process group to make. The login
// shell is not asked on those platforms anyway.
func detach(*exec.Cmd) {}
