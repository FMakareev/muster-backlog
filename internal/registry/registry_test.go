package registry_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/FMakareev/muster-backlog/internal/project"
	"github.com/FMakareev/muster-backlog/internal/registry"
)

// newProject creates a minimal but real Backlog.md project.
func newProject(t *testing.T, root, name string) string {
	t.Helper()
	dir := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Join(dir, "backlog", "tasks"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	cfg := filepath.Join(dir, "backlog", "config.yml")
	if err := os.WriteFile(cfg, []byte("project_name: \""+name+"\"\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return dir
}

func writeRegistry(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), registry.FileName)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write registry: %v", err)
	}
	return path
}

func TestLoadFromResolvesProjects(t *testing.T) {
	work := t.TempDir()
	alpha := newProject(t, work, "alpha")
	beta := newProject(t, work, "beta")

	path := writeRegistry(t, ""+
		"projects:\n"+
		"  - path: "+beta+"\n"+
		"    name: Beta Project\n"+
		"    color: \"#7aa2f7\"\n"+
		"  - path: "+alpha+"\n"+
		"wip_limits:\n"+
		"  In Progress: 3\n")

	reg, err := registry.LoadFrom(path)
	if err != nil {
		t.Fatalf("LoadFrom: %v", err)
	}
	if len(reg.Projects) != 2 {
		t.Fatalf("got %d projects, want 2", len(reg.Projects))
	}

	// File order is display order; the loader must not sort.
	if reg.Projects[0].Path != beta || reg.Projects[1].Path != alpha {
		t.Errorf("order = [%s %s], want the order written in the file",
			reg.Projects[0].DisplayName, reg.Projects[1].DisplayName)
	}
	if got := reg.Projects[0].DisplayName; got != "Beta Project" {
		t.Errorf("DisplayName = %q, want the name from the file", got)
	}
	if got := reg.Projects[0].Colour; got != "#7aa2f7" {
		t.Errorf("Colour = %q, want #7aa2f7", got)
	}
	// An entry with no name falls back to the folder name.
	if got := reg.Projects[1].DisplayName; got != "alpha" {
		t.Errorf("DisplayName = %q, want the folder name", got)
	}
	if got := reg.WIPLimits["In Progress"]; got != 3 {
		t.Errorf("WIPLimits[In Progress] = %d, want 3", got)
	}
	for _, p := range reg.Projects {
		if !p.OK() {
			t.Errorf("%s did not resolve: %v", p.DisplayName, p.Err)
		}
		if p.Location.Layout != project.LayoutStandard {
			t.Errorf("%s layout = %q", p.DisplayName, p.Location.Layout)
		}
	}
}

// The registry describes where projects are, never how they work. Statuses,
// priorities and types belong to each project's own config.
func TestEntryCarriesNoBacklogConfiguration(t *testing.T) {
	work := t.TempDir()
	alpha := newProject(t, work, "alpha")

	path := writeRegistry(t, ""+
		"projects:\n"+
		"  - path: "+alpha+"\n"+
		"    statuses: [\"Nope\"]\n"+
		"    task_prefix: nope\n")

	reg, err := registry.LoadFrom(path)
	if err != nil {
		t.Fatalf("LoadFrom: %v", err)
	}
	if len(reg.Projects) != 1 || !reg.Projects[0].OK() {
		t.Fatalf("project did not resolve: %+v", reg.Projects)
	}
	// The entry type has no field for either key, so they are simply not
	// carried. What the project declares is read from its own config instead.
	if reg.Projects[0].Location.ConfigPath == "" {
		t.Error("resolved project has no config path to read its own settings from")
	}
}

// A folder that moved, or was never initialised, must not take the others with
// it: one broken entry is a degraded row, not a failed load.
func TestOneBadEntryDoesNotSinkTheRest(t *testing.T) {
	work := t.TempDir()
	good := newProject(t, work, "good")
	notAProject := filepath.Join(work, "empty")
	if err := os.MkdirAll(notAProject, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	path := writeRegistry(t, ""+
		"projects:\n"+
		"  - path: "+notAProject+"\n"+
		"  - path: "+filepath.Join(work, "gone")+"\n"+
		"  - path: \"   \"\n"+
		"  - path: "+good+"\n")

	reg, err := registry.LoadFrom(path)
	if err != nil {
		t.Fatalf("LoadFrom returned an error for a partially broken registry: %v", err)
	}
	if len(reg.Projects) != 4 {
		t.Fatalf("got %d projects, want all 4 reported", len(reg.Projects))
	}
	for i, p := range reg.Projects[:3] {
		if p.OK() {
			t.Errorf("entry %d resolved but should not have", i+1)
		}
		if p.Err == nil || p.Err.Error() == "" {
			t.Errorf("entry %d has no explanation", i+1)
		}
	}
	if !reg.Projects[3].OK() {
		t.Errorf("the valid project did not resolve: %v", reg.Projects[3].Err)
	}
}

func TestDuplicatePathIsReported(t *testing.T) {
	work := t.TempDir()
	alpha := newProject(t, work, "alpha")

	path := writeRegistry(t, "projects:\n  - path: "+alpha+"\n  - path: "+alpha+"\n")

	reg, err := registry.LoadFrom(path)
	if err != nil {
		t.Fatalf("LoadFrom: %v", err)
	}
	if !reg.Projects[0].OK() {
		t.Errorf("first entry should resolve: %v", reg.Projects[0].Err)
	}
	if reg.Projects[1].OK() {
		t.Error("second entry duplicates the first and should be reported")
	} else if !strings.Contains(reg.Projects[1].Err.Error(), "already registered") {
		t.Errorf("Err = %v, want it to name the duplication", reg.Projects[1].Err)
	}
}

// First run is not a failure; it is an invitation to add a project.
func TestMissingRegistryIsADistinctState(t *testing.T) {
	path := filepath.Join(t.TempDir(), registry.FileName)

	reg, err := registry.LoadFrom(path)
	if !errors.Is(err, registry.ErrNoRegistry) {
		t.Fatalf("err = %v, want ErrNoRegistry", err)
	}
	if reg.Path != path {
		t.Errorf("Path = %q, want the attempted path so the UI can name it", reg.Path)
	}
	if len(reg.Projects) != 0 {
		t.Errorf("got %d projects, want none", len(reg.Projects))
	}
}

func TestMalformedRegistryNamesTheFile(t *testing.T) {
	path := writeRegistry(t, "projects:\n  - path: [unclosed\n")

	_, err := registry.LoadFrom(path)
	if err == nil {
		t.Fatal("want an error for malformed YAML")
	}
	if errors.Is(err, registry.ErrNoRegistry) {
		t.Fatal("malformed is not the same state as missing")
	}
	if !strings.Contains(err.Error(), path) {
		t.Errorf("Err = %v, want it to name %q", err, path)
	}
}

func TestTildeIsExpanded(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("no home directory: %v", err)
	}
	path := writeRegistry(t, "projects:\n  - path: \"~/definitely-not-a-project-xyz\"\n")

	reg, err := registry.LoadFrom(path)
	if err != nil {
		t.Fatalf("LoadFrom: %v", err)
	}
	got := reg.Projects[0].Path
	if !strings.HasPrefix(got, home) {
		t.Errorf("Path = %q, want it expanded under %q", got, home)
	}
}

func TestDefaultPathIsUnderTheXDGConfigHome(t *testing.T) {
	got := registry.DefaultPath()
	want := filepath.Join(registry.AppDir, registry.FileName)
	if !strings.HasSuffix(got, want) {
		t.Errorf("DefaultPath = %q, want it to end in %q", got, want)
	}
	if !filepath.IsAbs(got) {
		t.Errorf("DefaultPath = %q, want an absolute path", got)
	}
}

// The path in the status bar is shown, not used. Under a home directory the
// full form carries a username into every screenshot anybody takes.
func TestAPathIsAbbreviatedForShowing(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("no home directory")
	}

	under := filepath.Join(home, ".config", "muster", "projects.yml")
	if got, want := registry.Abbreviate(under), filepath.Join("~", ".config", "muster", "projects.yml"); got != want {
		t.Errorf("Abbreviate(%q) = %q, want %q", under, got, want)
	}
	if got := registry.Abbreviate(home); got != "~" {
		t.Errorf("the home directory itself abbreviates to %q", got)
	}

	// Anything outside is left exactly as it is: shortening it would be
	// inventing a path rather than shortening one.
	outside := filepath.Join("/etc", "muster", "projects.yml")
	if got := registry.Abbreviate(outside); got != outside {
		t.Errorf("Abbreviate(%q) = %q, want it untouched", outside, got)
	}
	if got := registry.Abbreviate(""); got != "" {
		t.Errorf("Abbreviate(\"\") = %q", got)
	}

	// A prefix that merely looks like the home directory is not one.
	sibling := home + "-backup/projects.yml"
	if got := registry.Abbreviate(sibling); got != sibling {
		t.Errorf("Abbreviate(%q) = %q, want it untouched", sibling, got)
	}
}
