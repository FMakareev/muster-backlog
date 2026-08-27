package backlogcli_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/FMakareev/muster-backlog/internal/backlogcli"
)

// fakeCLI writes an executable that answers --version, so resolution can be
// tested without the real binary and without a version to match.
func fakeCLI(t *testing.T, dir, version string) string {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	path := filepath.Join(dir, backlogcli.BinaryName)
	script := "#!/bin/sh\necho " + version + "\n"
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("write: %v", err)
	}
	return path
}

// An application started from a desktop launcher inherits the session
// environment rather than a shell's, so a CLI installed by a package manager
// into a directory only a shell rc file mentions is invisible on PATH. It is
// still where that package manager puts it.
func TestFoundInAPackageManagerBinWithNothingOnPath(t *testing.T) {
	home := t.TempDir()
	want := fakeCLI(t, filepath.Join(home, ".local", "share", "pnpm"), "1.50.1")

	t.Setenv("HOME", home)
	t.Setenv("PATH", filepath.Join(t.TempDir(), "empty"))
	t.Setenv("XDG_DATA_HOME", "")
	t.Setenv("PNPM_HOME", "")
	t.Setenv("BUN_INSTALL", "")
	t.Setenv("NPM_CONFIG_PREFIX", "")

	found, searched := backlogcli.Find()
	if len(found) == 0 || found[0] != want {
		t.Fatalf("found %v, want %q first (looked in %v)", found, want, searched)
	}
	if len(searched) == 0 || searched[0] != "PATH" {
		t.Errorf("PATH was not tried first: %v", searched)
	}
}

func TestAnEnvironmentVariableWins(t *testing.T) {
	home := t.TempDir()
	fakeCLI(t, filepath.Join(home, ".local", "share", "pnpm"), "1.50.1")
	elsewhere := t.TempDir()
	want := fakeCLI(t, elsewhere, "1.50.1")

	t.Setenv("HOME", home)
	t.Setenv("PATH", filepath.Join(t.TempDir(), "empty"))
	t.Setenv("PNPM_HOME", elsewhere)

	found, searched := backlogcli.Find()
	if len(found) == 0 || found[0] != want {
		t.Fatalf("found %v, want the one PNPM_HOME names first (looked in %v)",
			found, searched)
	}
}

// Being told it is missing is no help to someone who has installed it. Being
// told where the search went is.
func TestNotFoundNamesEverywhereItLooked(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("PATH", filepath.Join(t.TempDir(), "empty"))
	t.Setenv("PNPM_HOME", "")
	t.Setenv("BUN_INSTALL", "")
	t.Setenv("NPM_CONFIG_PREFIX", "")

	found, searched := backlogcli.Find()
	if len(found) != 0 {
		t.Fatalf("found %v where nothing was installed", found)
	}
	if len(searched) < 3 {
		t.Fatalf("only looked in %v", searched)
	}

	_, err := backlogcli.NewAt("")
	if err == nil {
		t.Fatal("resolution succeeded with nothing installed")
	}
	for _, want := range []string{"PATH", ".local"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the error does not mention %q: %v", want, err)
		}
	}
}

// An explicit path is a person saying where it is. It is not second-guessed,
// and a wrong one is reported about itself rather than about a search.
func TestAnExplicitPathIsUsedAndItsFailureNamesIt(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "nowhere", "backlog")

	_, err := backlogcli.NewAt(missing)
	if err == nil {
		t.Fatal("a path that is not there was accepted")
	}
	if !strings.Contains(err.Error(), missing) {
		t.Errorf("the error does not name the path given: %v", err)
	}
	if strings.Contains(err.Error(), "Looked in") {
		t.Errorf("an explicit path fell back to searching: %v", err)
	}
}

// Whatever is found still has to be the right thing.
func TestAFoundBinaryIsStillVersionChecked(t *testing.T) {
	dir := t.TempDir()
	old := fakeCLI(t, dir, "1.20.0")

	_, err := backlogcli.NewAt(old)
	if err == nil {
		t.Fatal("an old version was accepted")
	}
	if !strings.Contains(err.Error(), backlogcli.MinimumVersion) {
		t.Errorf("the error does not say what is needed: %v", err)
	}
}

// The binary pnpm installs is a shell script that execs node. A desktop
// launcher's environment has neither on PATH, so finding the script is not
// enough - what runs it has to be findable too.
func TestTheCLIIsRunWithAPathThatCanFindNode(t *testing.T) {
	home := t.TempDir()
	nodeDir := filepath.Join(home, ".nvm", "versions", "node", "v24.0.0", "bin")
	if err := os.MkdirAll(nodeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// A "node" that answers with a version, and a wrapper that can only work
	// if it can find that node.
	if err := os.WriteFile(filepath.Join(nodeDir, "node"),
		[]byte("#!/bin/sh\necho 1.50.1\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	wrapperDir := filepath.Join(home, ".local", "share", "pnpm")
	if err := os.MkdirAll(wrapperDir, 0o755); err != nil {
		t.Fatal(err)
	}
	wrapper := filepath.Join(wrapperDir, backlogcli.BinaryName)
	if err := os.WriteFile(wrapper,
		[]byte("#!/bin/sh\nexec node \"$@\"\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("HOME", home)
	t.Setenv("PATH", filepath.Join(t.TempDir(), "empty"))
	t.Setenv("PNPM_HOME", "")
	t.Setenv("BUN_INSTALL", "")
	t.Setenv("NPM_CONFIG_PREFIX", "")
	t.Setenv("XDG_DATA_HOME", "")

	runner, err := backlogcli.NewAt("")
	if err != nil {
		t.Fatalf("the CLI could not be run in a launcher-like environment: %v", err)
	}
	if runner.Version() != "1.50.1" {
		t.Errorf("version is %q", runner.Version())
	}
}

// A candidate that exists but cannot run is not the answer; the search goes on.
func TestABinaryThatCannotRunIsPassedOver(t *testing.T) {
	home := t.TempDir()
	broken := filepath.Join(home, ".local", "share", "pnpm")
	if err := os.MkdirAll(broken, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(broken, backlogcli.BinaryName),
		[]byte("#!/bin/sh\nexec definitely-not-here\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	working := filepath.Join(home, ".local", "bin")
	fakeCLI(t, working, "1.50.1")

	t.Setenv("HOME", home)
	t.Setenv("PATH", filepath.Join(t.TempDir(), "empty"))
	t.Setenv("PNPM_HOME", "")
	t.Setenv("XDG_DATA_HOME", "")

	runner, err := backlogcli.NewAt("")
	if err != nil {
		t.Fatalf("the working binary was not reached: %v", err)
	}
	if runner.Version() != "1.50.1" {
		t.Errorf("version is %q", runner.Version())
	}
}
