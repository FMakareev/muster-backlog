package registry_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/FMakareev/muster-backlog/internal/registry"
)

// annotated is a registry as a person would keep one: comments at the top, an
// inline note, a hand-chosen order and a section that has nothing to do with
// projects.
const annotated = `# The projects I actually work on.
projects:
  - path: /tmp/one   # the game
    name: One
  - path: /tmp/two
    color: "#7aa2f7"

# advisory only, nothing is blocked
wip_limits:
  In Progress: 3
`

func writeFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), registry.FileName)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	return path
}

func read(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	return string(raw)
}

func TestAddKeepsCommentsAndOrder(t *testing.T) {
	path := writeFile(t, annotated)

	if err := registry.Add(path, registry.Entry{
		Path: "/tmp/three", Name: "Three", Colour: "#abc123",
	}); err != nil {
		t.Fatalf("Add: %v", err)
	}

	got := read(t, path)
	for _, want := range []string{
		"# The projects I actually work on.",
		"# advisory only, nothing is blocked",
		"# the game",
		"wip_limits:",
		"In Progress: 3",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("lost %q:\n%s", want, got)
		}
	}
	// Appended last, and the entries before it untouched.
	if !strings.Contains(got, "- path: /tmp/three") {
		t.Errorf("the new project is missing:\n%s", got)
	}
	if strings.Index(got, "/tmp/one") > strings.Index(got, "/tmp/three") {
		t.Errorf("the new project was not appended last:\n%s", got)
	}
	// The list indentation a person wrote is kept.
	if !strings.Contains(got, "  - path: /tmp/one") {
		t.Errorf("list indentation changed:\n%s", got)
	}
}

func TestAddRefusesAFolderTwice(t *testing.T) {
	path := writeFile(t, annotated)

	// The same folder written the other way round still resolves to the same
	// place, and must not be registered twice.
	err := registry.Add(path, registry.Entry{Path: "/tmp/./one"})
	if !errors.Is(err, registry.ErrAlreadyRegistered) {
		t.Fatalf("err = %v, want ErrAlreadyRegistered", err)
	}
	if read(t, path) != annotated {
		t.Error("a refused add still wrote to the file")
	}
}

func TestUpdateRewritesOnlyItsOwnEntry(t *testing.T) {
	path := writeFile(t, annotated)

	if err := registry.Update(path, "/tmp/two", registry.Entry{
		Path: "/tmp/two", Name: "Renamed", Colour: "#ff0000", Hidden: true,
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got := read(t, path)
	if !strings.Contains(got, "name: Renamed") || !strings.Contains(got, `color: "#ff0000"`) {
		t.Errorf("the edit did not land:\n%s", got)
	}
	if !strings.Contains(got, "hidden: true") {
		t.Errorf("hidden was not written:\n%s", got)
	}
	// The neighbour keeps its own inline comment.
	if !strings.Contains(got, "# the game") {
		t.Errorf("an untouched entry lost its comment:\n%s", got)
	}
	if !strings.Contains(got, "# The projects I actually work on.") {
		t.Errorf("the header comment was lost:\n%s", got)
	}
}

// An entry with no name or colour is written as short as a person would write
// it, rather than padded with empty keys.
func TestUpdateOmitsEmptyFields(t *testing.T) {
	path := writeFile(t, annotated)

	if err := registry.Update(path, "/tmp/one", registry.Entry{Path: "/tmp/one"}); err != nil {
		t.Fatalf("Update: %v", err)
	}
	got := read(t, path)
	if strings.Contains(got, "name: \"\"") || strings.Contains(got, "color: \"\"") ||
		strings.Contains(got, "hidden: false") {
		t.Errorf("empty fields were written out:\n%s", got)
	}
}

func TestRemoveLeavesTheRestAlone(t *testing.T) {
	path := writeFile(t, annotated)

	if err := registry.Remove(path, "/tmp/one"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	got := read(t, path)
	if strings.Contains(got, "/tmp/one") {
		t.Errorf("the entry is still there:\n%s", got)
	}
	if !strings.Contains(got, "/tmp/two") || !strings.Contains(got, "wip_limits:") {
		t.Errorf("removal took more than its entry:\n%s", got)
	}
}

func TestMoveReordersWithoutLosingEntries(t *testing.T) {
	path := writeFile(t, annotated)
	if err := registry.Add(path, registry.Entry{Path: "/tmp/three"}); err != nil {
		t.Fatalf("Add: %v", err)
	}

	if err := registry.Move(path, "/tmp/three", 0); err != nil {
		t.Fatalf("Move: %v", err)
	}
	reg, err := registry.LoadFrom(path)
	if err != nil {
		t.Fatalf("LoadFrom: %v", err)
	}
	var order []string
	for _, p := range reg.Projects {
		order = append(order, p.Written)
	}
	want := []string{"/tmp/three", "/tmp/one", "/tmp/two"}
	if strings.Join(order, ",") != strings.Join(want, ",") {
		t.Errorf("order is %v, want %v", order, want)
	}

	// Past the end means last, not an error.
	if err := registry.Move(path, "/tmp/three", 99); err != nil {
		t.Fatalf("Move to the end: %v", err)
	}
	reg, _ = registry.LoadFrom(path)
	if reg.Projects[len(reg.Projects)-1].Written != "/tmp/three" {
		t.Error("moving past the end did not put it last")
	}
}

func TestEditsFindAProjectWrittenAnyWay(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("no home directory")
	}
	path := writeFile(t, "projects:\n  - path: ~/Dev/thing\n")

	// Addressed by its expanded form, written in its tilde form.
	if err := registry.Update(path, filepath.Join(home, "Dev", "thing"),
		registry.Entry{Path: "~/Dev/thing", Name: "Thing"}); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if got := read(t, path); !strings.Contains(got, "name: Thing") {
		t.Errorf("the entry was not found by its resolved path:\n%s", got)
	}
}

func TestEditingAnUnknownProjectIsRefused(t *testing.T) {
	path := writeFile(t, annotated)
	for name, err := range map[string]error{
		"Update": registry.Update(path, "/tmp/nowhere", registry.Entry{Path: "/tmp/nowhere"}),
		"Remove": registry.Remove(path, "/tmp/nowhere"),
		"Move":   registry.Move(path, "/tmp/nowhere", 0),
	} {
		if !errors.Is(err, registry.ErrNoSuchProject) {
			t.Errorf("%s: err = %v, want ErrNoSuchProject", name, err)
		}
	}
	if read(t, path) != annotated {
		t.Error("a refused edit still wrote to the file")
	}
}

// A registry that will not parse is someone's work in progress. Refusing is
// the only safe answer; overwriting it loses whatever they were writing.
func TestAnUnparseableRegistryIsNotOverwritten(t *testing.T) {
	const broken = "projects:\n  - path: [unclosed\n"
	path := writeFile(t, broken)

	if err := registry.Add(path, registry.Entry{Path: "/tmp/new"}); err == nil {
		t.Fatal("Add accepted a registry it could not parse")
	}
	if read(t, path) != broken {
		t.Error("the broken file was rewritten")
	}
}

func TestAddCreatesTheFileOnFirstRun(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", registry.FileName)

	if err := registry.Add(path, registry.Entry{Path: "/tmp/first", Name: "First"}); err != nil {
		t.Fatalf("Add: %v", err)
	}
	reg, err := registry.LoadFrom(path)
	if err != nil {
		t.Fatalf("LoadFrom: %v", err)
	}
	if len(reg.Projects) != 1 || reg.Projects[0].DisplayName != "First" {
		t.Fatalf("got %d projects, want the one just added", len(reg.Projects))
	}
}

// A registry that sets only wip_limits has no projects list to append to.
func TestAddToARegistryWithNoProjectsKey(t *testing.T) {
	path := writeFile(t, "# limits only\nwip_limits:\n  In Progress: 2\n")

	if err := registry.Add(path, registry.Entry{Path: "/tmp/one"}); err != nil {
		t.Fatalf("Add: %v", err)
	}
	reg, err := registry.LoadFrom(path)
	if err != nil {
		t.Fatalf("LoadFrom: %v", err)
	}
	if len(reg.Projects) != 1 {
		t.Fatalf("got %d projects, want 1", len(reg.Projects))
	}
	if reg.WIPLimits["In Progress"] != 2 {
		t.Error("the limits were lost")
	}
	if !strings.Contains(read(t, path), "# limits only") {
		t.Error("the comment was lost")
	}
}

// "projects:" with nothing under it parses as null rather than as a list.
func TestAddToAnEmptyProjectsKey(t *testing.T) {
	path := writeFile(t, "projects:\n")

	if err := registry.Add(path, registry.Entry{Path: "/tmp/one"}); err != nil {
		t.Fatalf("Add: %v", err)
	}
	reg, err := registry.LoadFrom(path)
	if err != nil {
		t.Fatalf("LoadFrom: %v", err)
	}
	if len(reg.Projects) != 1 {
		t.Fatalf("got %d projects, want 1:\n%s", len(reg.Projects), read(t, path))
	}
}

func TestHiddenSurvivesARoundTrip(t *testing.T) {
	path := writeFile(t, annotated)

	if err := registry.Update(path, "/tmp/one",
		registry.Entry{Path: "/tmp/one", Name: "One", Hidden: true}); err != nil {
		t.Fatalf("Update: %v", err)
	}
	reg, err := registry.LoadFrom(path)
	if err != nil {
		t.Fatalf("LoadFrom: %v", err)
	}
	if !reg.Projects[0].Hidden {
		t.Error("hidden did not survive being written and read back")
	}
	if reg.Projects[1].Hidden {
		t.Error("hiding one project hid another")
	}
}

// An empty registry is written as an inline list, which cannot hold a block
// entry. Appending to one used to produce a file the parser rejected, and it
// went to disk.
func TestAddToAnInlineEmptyList(t *testing.T) {
	path := writeFile(t, "# nothing yet\nprojects: []\n")

	if err := registry.Add(path, registry.Entry{Path: "/tmp/one", Name: "One"}); err != nil {
		t.Fatalf("Add: %v", err)
	}
	reg, err := registry.LoadFrom(path)
	if err != nil {
		t.Fatalf("the written registry does not parse: %v\n%s", err, read(t, path))
	}
	if len(reg.Projects) != 1 || reg.Projects[0].DisplayName != "One" {
		t.Fatalf("got %d projects:\n%s", len(reg.Projects), read(t, path))
	}
	if !strings.Contains(read(t, path), "# nothing yet") {
		t.Error("the comment was lost")
	}
}

// A list written inline with entries in it is rebuilt as a block list, which
// changes the style but keeps every project.
func TestAddToAnInlineListWithEntries(t *testing.T) {
	path := writeFile(t, "projects: [{path: /tmp/one, name: One}]\n")

	if err := registry.Add(path, registry.Entry{Path: "/tmp/two"}); err != nil {
		t.Fatalf("Add: %v", err)
	}
	reg, err := registry.LoadFrom(path)
	if err != nil {
		t.Fatalf("the written registry does not parse: %v\n%s", err, read(t, path))
	}
	if len(reg.Projects) != 2 {
		t.Fatalf("got %d projects, want 2:\n%s", len(reg.Projects), read(t, path))
	}
	if reg.Projects[0].DisplayName != "One" {
		t.Errorf("the entry that was already there changed: %+v", reg.Projects[0])
	}
}
