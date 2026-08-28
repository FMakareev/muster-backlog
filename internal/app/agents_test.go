package app

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// In package app rather than app_test: what a client is told to spawn is
// decided by mcpCommand, and that decision is the whole of this fix.

func serverName() string {
	if runtime.GOOS == "windows" {
		return mcpServerName + ".exe"
	}
	return mcpServerName
}

// What a client is told to spawn has to be a program that exists and starts.
// It was neither: the path recorded was the desktop binary, which links a
// browser engine and cannot start in the sandbox or container an agent runs
// in — and which then vanished when the package was reinstalled elsewhere,
// leaving a configuration pointing at nothing.
func TestTheCommandRegisteredIsTheServerBesideUs(t *testing.T) {
	binary, err := os.Executable()
	if err != nil {
		t.Fatalf("os.Executable: %v", err)
	}
	if resolved, err := filepath.EvalSymlinks(binary); err == nil {
		binary = resolved
	}
	beside := filepath.Join(filepath.Dir(binary), serverName())

	if _, err := os.Stat(beside); err != nil {
		if err := os.WriteFile(beside, []byte("#!/bin/sh\n"), 0o755); err != nil {
			t.Skipf("cannot write beside the test binary: %v", err)
		}
		t.Cleanup(func() { _ = os.Remove(beside) })
	}

	got := mcpCommand()
	if got != beside {
		t.Fatalf("registers %q, want the server beside us at %q", got, beside)
	}
	if filepath.Base(got) == filepath.Base(binary) {
		t.Error("registers the application itself, which cannot start where agents run")
	}
}

// With no server on the machine, nothing is written into anybody's
// configuration. Writing a command that does not exist is the failure this
// whole change is about: it looks like it worked and fails later, inside a
// program that cannot explain why.
//
// The command is replaced rather than the filesystem arranged: whichbin
// searches absolute directories like /usr/local/bin that no environment
// variable can hide, so on any machine that has installed the package this
// would otherwise skip - which is every machine where it matters.
func TestConnectingIsRefusedWithoutTheServer(t *testing.T) {
	original := serverCommand
	t.Cleanup(func() { serverCommand = original })
	serverCommand = func() string { return "" }

	s := &BoardService{}

	plan := s.AgentPlan("claude-code", false)
	if plan.Error == "" {
		t.Error("planned a connection with no server to connect to")
	}
	if !strings.Contains(plan.Error, mcpServerName) {
		t.Errorf("the refusal does not name what is missing: %q", plan.Error)
	}

	if result := s.AgentApply("claude-code", false); result.OK || result.Error == "" {
		t.Errorf("applied a connection with no server: %+v", result)
	}

	// Disconnecting still works: somebody whose server has gone is exactly
	// who needs to remove the entry pointing at it.
	if plan := s.AgentPlan("claude-code", true); plan.Error == missingServer {
		t.Error("refused to disconnect for want of the server being disconnected")
	}
}
