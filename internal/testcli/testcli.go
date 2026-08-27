// Package testcli decides what a test does when the backlog CLI is absent.
//
// Thirty-eight tests need it: every write in this application goes through
// that CLI, so every test of a write runs the real thing against a real
// project on disk. On a machine without it those tests skip, which is right —
// somebody reading the code should not need the whole toolchain to run what
// they can.
//
// On continuous integration it is exactly wrong. A skip is not a failure, so
// a pipeline with no CLI installed reports green having tested none of the
// writes: the most valuable half of the suite disappears and nothing says so.
// Setting MUSTER_REQUIRE_BACKLOG_CLI turns the skip into a failure, and CI
// sets it.
package testcli

import (
	"os"
	"os/exec"
	"testing"
)

// RequireEnv is the variable that makes a missing CLI fatal.
const RequireEnv = "MUSTER_REQUIRE_BACKLOG_CLI"

// Require skips the test when the backlog CLI is not installed, or fails it
// when the environment says the CLI has to be there.
func Require(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("backlog"); err == nil {
		return
	}
	if os.Getenv(RequireEnv) != "" {
		t.Fatalf("the backlog CLI is not installed, and %s is set: "+
			"this test exercises a real write and skipping it would report "+
			"green having tested nothing", RequireEnv)
	}
	t.Skip("the backlog CLI is not installed")
}
