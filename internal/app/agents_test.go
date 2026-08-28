package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// In package app rather than app_test: what a client is told to spawn is
// decided here, and that decision is the whole of this.

// notInAContainer makes the ordinary case testable on a machine that is in
// one — this repository's own development happens inside distrobox, where
// every one of these would otherwise take the container branch.
func notInAContainer(t *testing.T) {
	t.Helper()
	original := inContainer
	t.Cleanup(func() { inContainer = original })
	inContainer = func() bool { return false }
}

func pretendContainer(t *testing.T, home string) {
	t.Helper()
	originalIn, originalDir := inContainer, sharedBinDir
	t.Cleanup(func() { inContainer, sharedBinDir = originalIn, originalDir })
	inContainer = func() bool { return true }
	sharedBinDir = func() string { return filepath.Join(home, ".local", "bin") }
}

func writeServer(t *testing.T, dir string) string {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, serverFileName())
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

// Outside a container, the server beside this binary — unchanged.
func TestTheCommandRegisteredIsTheServerBesideUs(t *testing.T) {
	notInAContainer(t)

	binary, err := os.Executable()
	if err != nil {
		t.Fatalf("os.Executable: %v", err)
	}
	if resolved, err := filepath.EvalSymlinks(binary); err == nil {
		binary = resolved
	}
	beside := filepath.Join(filepath.Dir(binary), serverFileName())

	if _, err := os.Stat(beside); err != nil {
		if err := os.WriteFile(beside, []byte("#!/bin/sh\n"), 0o755); err != nil {
			t.Skipf("cannot write beside the test binary: %v", err)
		}
		t.Cleanup(func() { _ = os.Remove(beside) })
	}

	got := locateServer()
	if got.Command != beside {
		t.Fatalf("registers %q, want the server beside us at %q", got.Command, beside)
	}
	if filepath.Base(got.Command) == filepath.Base(binary) {
		t.Error("registers the application itself, which cannot start where agents run")
	}
	if got.Where != "" {
		t.Errorf("says something about the environment when there is nothing to say: %q", got.Where)
	}
}

// In a container the path beside this binary is real and useless. The client
// reading the configuration runs on the host, where it holds nothing — the
// same $HOME that makes the configuration reachable from in here is what
// makes the container's own paths meaningless out there.
func TestInAContainerTheSharedCopyIsRegistered(t *testing.T) {
	home := t.TempDir()
	pretendContainer(t, home)

	// The container's own binary exists, and is the wrong answer.
	inside := writeServer(t, filepath.Join(t.TempDir(), "usr", "local", "bin"))
	shared := writeServer(t, filepath.Join(home, ".local", "bin"))

	got := locateServer()
	if got.Command != shared {
		t.Errorf("registers %q, want the shared copy at %q", got.Command, shared)
	}
	if got.Command == inside {
		t.Error("registers a path that only resolves inside the container")
	}
	if got.Problem != "" {
		t.Errorf("refused with a perfectly good command available: %s", got.Problem)
	}
	// The person has to be able to see which environment it resolves in.
	if !strings.Contains(strings.ToLower(got.Where), "container") {
		t.Errorf("does not say where the command runs: %q", got.Where)
	}
}

// With nothing on the shared home there is no command that resolves on both
// sides, and inventing one is the failure being fixed.
func TestInAContainerWithNothingSharedItRefuses(t *testing.T) {
	pretendContainer(t, t.TempDir())

	got := locateServer()
	if got.Command != "" {
		t.Fatalf("registers %q, which only the container can see", got.Command)
	}
	if got.Problem == "" {
		t.Fatal("refused without saying why")
	}
	for _, want := range []string{"container", "distrobox-export", mcpServerName} {
		if !strings.Contains(got.Problem, want) {
			t.Errorf("the refusal does not mention %q: %s", want, got.Problem)
		}
	}
}

// The plan says which environment the command will run in, before it is
// written rather than after it fails.
func TestThePlanSaysWhereTheCommandWillRun(t *testing.T) {
	home := t.TempDir()
	pretendContainer(t, home)
	writeServer(t, filepath.Join(home, ".local", "bin"))

	s := &BoardService{}
	plan := s.AgentPlan("claude-code", false)
	if plan.Error != "" {
		t.Fatalf("planning failed: %s", plan.Error)
	}
	if !strings.Contains(strings.ToLower(plan.Note), "container") {
		t.Errorf("the plan does not say where the command runs: %q", plan.Note)
	}
	// The client's own note survives beside it.
	if !strings.Contains(plan.Note, "every project") {
		t.Errorf("the client's own note was lost: %q", plan.Note)
	}
}

// With no server anywhere, nothing is written into anybody's configuration.
// Writing a command that does not exist is the failure this is about: it
// looks like it worked and fails later, inside a program that cannot explain
// why.
//
// The location is replaced rather than the filesystem arranged: whichbin
// searches absolute directories like /usr/local/bin that no environment
// variable can hide, so on any machine that has installed the package this
// would otherwise skip — which is every machine where it matters.
func TestConnectingIsRefusedWithoutTheServer(t *testing.T) {
	original := serverCommand
	t.Cleanup(func() { serverCommand = original })
	serverCommand = func() serverLocation {
		return serverLocation{Problem: missingServer}
	}

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

// This repository is developed inside a distrobox container, so the detection
// can be checked against something real rather than only against a stub.
func TestContainerDetectionAgreesWithTheMarksOnThisMachine(t *testing.T) {
	marks := 0
	for _, mark := range containerMarks {
		if _, err := os.Stat(mark); err == nil {
			marks++
		}
	}
	for _, name := range containerVars {
		if os.Getenv(name) != "" {
			marks++
		}
	}
	if got := inContainer(); got != (marks > 0) {
		t.Errorf("inContainer() = %v with %d marks present", got, marks)
	}
}
