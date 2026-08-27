package whichbin

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// Every test here starts from a home nobody has installed anything into and a
// PATH that holds nothing, which is the state a launcher hands a desktop app.
func sandbox(t *testing.T) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("the executable bit and the folder layout are a unix story")
	}
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("PATH", filepath.Join(home, "empty"))
	for _, k := range []string{"PNPM_HOME", "BUN_INSTALL", "VOLTA_HOME", "CARGO_HOME", "GOBIN", "GOPATH", "NPM_CONFIG_PREFIX"} {
		t.Setenv(k, "")
	}
	return home
}

func put(t *testing.T, dir, name string, mode os.FileMode) string {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte("#!/bin/sh\nexit 0\n"), mode); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestPATHWinsAndSaysSo(t *testing.T) {
	home := sandbox(t)
	dir := filepath.Join(home, "onpath")
	want := put(t, dir, "toolish", 0o755)
	t.Setenv("PATH", dir)

	got, onPATH, err := Look("toolish")
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("found %q, wanted %q", got, want)
	}
	if !onPATH {
		t.Error("a command on PATH must be reported as such: the caller prints the bare name for it")
	}
}

// The bug this package exists for: the command is installed, and PATH has never
// heard of it because no shell profile ever ran.
func TestFoundWhereInstallersActuallyPutThings(t *testing.T) {
	home := sandbox(t)
	want := put(t, filepath.Join(home, ".local", "bin"), "claude", 0o755)

	got, onPATH, err := Look("claude")
	if err != nil {
		t.Fatalf("a command in ~/.local/bin must be found: %v", err)
	}
	if got != want {
		t.Errorf("found %q, wanted %q", got, want)
	}
	if onPATH {
		t.Error("it was not on PATH, and the caller has to know that to print a usable command")
	}
}

func TestASymlinkIsFollowed(t *testing.T) {
	home := sandbox(t)
	real := put(t, filepath.Join(home, ".local", "share", "claude", "versions"), "2.1.238", 0o755)
	dir := filepath.Join(home, ".local", "bin")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "claude")
	if err := os.Symlink(real, link); err != nil {
		t.Fatal(err)
	}
	// exactly the shape a real install has
	if got, _, err := Look("claude"); err != nil || got != link {
		t.Errorf("found %q (%v), wanted the link at %q", got, err, link)
	}
}

func TestAFileThatCannotRunIsNotACommand(t *testing.T) {
	home := sandbox(t)
	put(t, filepath.Join(home, ".local", "bin"), "notes", 0o644)

	if got, _, err := Look("notes"); err == nil {
		t.Errorf("a file without an execute bit was offered as a command: %q", got)
	}
}

func TestTheNewestVersionOfAVersionManagerWins(t *testing.T) {
	home := sandbox(t)
	put(t, filepath.Join(home, ".nvm", "versions", "node", "v20.0.0", "bin"), "node", 0o755)
	want := put(t, filepath.Join(home, ".nvm", "versions", "node", "v24.15.0", "bin"), "node", 0o755)

	got, _, err := Look("node")
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("found %q, wanted the newer %q", got, want)
	}
}

func TestAnInstallerThatExportsItsOwnVariableIsBelieved(t *testing.T) {
	home := sandbox(t)
	dir := filepath.Join(home, "somewhere", "odd")
	want := put(t, dir, "pnpm", 0o755)
	t.Setenv("PNPM_HOME", dir)

	got, _, err := Look("pnpm")
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("found %q, wanted %q", got, want)
	}
}

func TestNothingAnywhereIsAnHonestError(t *testing.T) {
	sandbox(t)
	_, _, err := Look("no-such-command-anywhere-at-all")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("wanted ErrNotFound, got %v", err)
	}
}

// A name with a path in it is the caller's business, not ours: guessing a
// second directory for it would be running something they did not name.
func TestAPathIsNotAName(t *testing.T) {
	home := sandbox(t)
	put(t, filepath.Join(home, ".local", "bin"), "tool", 0o755)

	if _, _, err := Look("bin/tool"); err == nil {
		t.Error("a relative path must not be resolved against the guess list")
	}
}
