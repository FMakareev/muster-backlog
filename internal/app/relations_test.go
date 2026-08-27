package app_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/FMakareev/muster-backlog/internal/app"
)

func taskWith(id, title, status string, extra string) string {
	return "---\nid: " + id + "\ntitle: " + title + "\nstatus: " + status + "\n" +
		extra + "---\n\n## Description\n\n<!-- SECTION:DESCRIPTION:BEGIN -->\nx\n" +
		"<!-- SECTION:DESCRIPTION:END -->\n"
}

func ordinalsIn(t *testing.T, s *app.BoardService, project string) map[string]int {
	t.Helper()
	out := map[string]int{}
	for _, view := range s.Tasks(app.QueryInput{Projects: []string{project}}) {
		if view.Entity.Ordinal != nil {
			out[view.ID] = *view.Entity.Ordinal
		}
	}
	return out
}

func TestDependenciesCanBeSetAndCleared(t *testing.T) {
	if _, err := exec.LookPath("backlog"); err != nil {
		t.Skip("the backlog CLI is not installed")
	}
	one := newProject(t, "one", map[string]string{
		"task-1 - a.md": task("TASK-1", "First", "To Do"),
		"task-2 - b.md": task("TASK-2", "Second", "To Do"),
		"task-3 - c.md": task("TASK-3", "Third", "To Do"),
	})
	s := startService(t, withRegistry(t, one))

	if result := s.SetDependencies(one, "TASK-3", []string{"TASK-1", "TASK-2"}); !result.OK {
		t.Fatalf("SetDependencies: %+v", result.Problem)
	}
	view, ok := s.Task(one, "task", "active", "TASK-3")
	if !ok {
		t.Fatal("TASK-3 vanished")
	}
	if strings.Join(view.Entity.Dependencies, ",") != "TASK-1,TASK-2" {
		t.Errorf("dependencies are %v", view.Entity.Dependencies)
	}

	if result := s.SetDependencies(one, "TASK-3", nil); !result.OK {
		t.Fatalf("clearing: %+v", result.Problem)
	}
	view, _ = s.Task(one, "task", "active", "TASK-3")
	if len(view.Entity.Dependencies) != 0 {
		t.Errorf("dependencies survived clearing: %v", view.Entity.Dependencies)
	}
}

// Ids resolve inside their own project only, so the commonest mistake is
// pointing at a task that exists somewhere else. It is caught before the CLI
// is asked, with a message that says why.
func TestADependencyThatCannotResolveIsRefusedBeforeWriting(t *testing.T) {
	if _, err := exec.LookPath("backlog"); err != nil {
		t.Skip("the backlog CLI is not installed")
	}
	one := newProject(t, "one", map[string]string{
		"task-1 - a.md": task("TASK-1", "Here", "To Do"),
	})
	two := newProject(t, "two", map[string]string{
		"task-9 - b.md": task("TASK-9", "Elsewhere", "To Do"),
	})
	s := startService(t, withRegistry(t, one, two))

	result := s.SetDependencies(one, "TASK-1", []string{"TASK-9"})
	if result.OK {
		t.Fatal("a dependency on another project's task was accepted")
	}
	if result.Problem == nil || !strings.Contains(result.Problem.Detail, "own project") {
		t.Errorf("refused without the reason: %+v", result.Problem)
	}

	if result := s.SetDependencies(one, "TASK-1", []string{"TASK-1"}); result.OK {
		t.Error("a task was allowed to wait on itself")
	}

	view, _ := s.Task(one, "task", "active", "TASK-1")
	if len(view.Entity.Dependencies) != 0 {
		t.Errorf("a refused edit still wrote: %v", view.Entity.Dependencies)
	}
}

func TestReferencesDocumentationAndFilesRoundTrip(t *testing.T) {
	if _, err := exec.LookPath("backlog"); err != nil {
		t.Skip("the backlog CLI is not installed")
	}
	one := newProject(t, "one", map[string]string{
		"task-1 - a.md": task("TASK-1", "Subject", "To Do"),
	})
	s := startService(t, withRegistry(t, one))

	cases := []struct {
		name  string
		set   func([]string) app.WriteResult
		read  func(app.TaskView) []string
		value []string
	}{
		{"references", func(v []string) app.WriteResult { return s.SetReferences(one, "TASK-1", v) },
			func(v app.TaskView) []string { return v.Entity.References },
			[]string{"https://example.com/a", "notes/b.md"}},
		{"documentation", func(v []string) app.WriteResult { return s.SetDocumentation(one, "TASK-1", v) },
			func(v app.TaskView) []string { return v.Entity.Documentation },
			[]string{"docs/spec.md"}},
	}
	for _, c := range cases {
		if result := c.set(c.value); !result.OK {
			t.Fatalf("%s: %+v", c.name, result.Problem)
		}
		view, _ := s.Task(one, "task", "active", "TASK-1")
		if strings.Join(c.read(view), ",") != strings.Join(c.value, ",") {
			t.Errorf("%s round-tripped as %v, want %v", c.name, c.read(view), c.value)
		}
		if result := c.set(nil); !result.OK {
			t.Fatalf("clearing %s: %+v", c.name, result.Problem)
		}
		view, _ = s.Task(one, "task", "active", "TASK-1")
		if len(c.read(view)) != 0 {
			t.Errorf("%s survived clearing: %v", c.name, c.read(view))
		}
	}
}

// Dragging a card is what manual order means, so the ordinal is computed from
// where it landed rather than typed.
func TestReorderingPlacesATaskWhereItWasDropped(t *testing.T) {
	if _, err := exec.LookPath("backlog"); err != nil {
		t.Skip("the backlog CLI is not installed")
	}
	one := newProject(t, "one", map[string]string{
		"task-1 - a.md": taskWith("TASK-1", "First", "To Do", "ordinal: 1000\n"),
		"task-2 - b.md": taskWith("TASK-2", "Second", "To Do", "ordinal: 2000\n"),
		"task-3 - c.md": taskWith("TASK-3", "Third", "To Do", "ordinal: 3000\n"),
	})
	s := startService(t, withRegistry(t, one))

	// Third dropped in front of Second: between 1000 and 2000.
	if result := s.Reorder(one, "TASK-3", "TASK-2"); !result.OK {
		t.Fatalf("Reorder: %+v", result.Problem)
	}
	got := ordinalsIn(t, s, one)
	if got["TASK-3"] != 1500 {
		t.Errorf("TASK-3 has ordinal %d, want the midpoint 1500", got["TASK-3"])
	}
	if order := orderOf(s, one); strings.Join(order, ",") != "TASK-1,TASK-3,TASK-2" {
		t.Errorf("order is %v", order)
	}

	// Dropped at the end: past the last one.
	if result := s.Reorder(one, "TASK-1", ""); !result.OK {
		t.Fatalf("Reorder to the end: %+v", result.Problem)
	}
	if order := orderOf(s, one); strings.Join(order, ",") != "TASK-3,TASK-2,TASK-1" {
		t.Errorf("order after moving to the end is %v", order)
	}
}

// Ordinals are neither unique nor mandatory, so a column can run out of room.
// It is restacked rather than refusing to reorder.
func TestReorderingRestacksAColumnWithNoRoomLeft(t *testing.T) {
	if _, err := exec.LookPath("backlog"); err != nil {
		t.Skip("the backlog CLI is not installed")
	}
	one := newProject(t, "one", map[string]string{
		"task-1 - a.md": taskWith("TASK-1", "First", "To Do", "ordinal: 1000\n"),
		"task-2 - b.md": taskWith("TASK-2", "Second", "To Do", "ordinal: 1001\n"),
		"task-3 - c.md": taskWith("TASK-3", "Third", "To Do", "ordinal: 1002\n"),
	})
	s := startService(t, withRegistry(t, one))

	if result := s.Reorder(one, "TASK-3", "TASK-2"); !result.OK {
		t.Fatalf("Reorder: %+v", result.Problem)
	}
	if order := orderOf(s, one); strings.Join(order, ",") != "TASK-1,TASK-3,TASK-2" {
		t.Errorf("order is %v", order)
	}
	// Every ordinal in the column is distinct afterwards, which is what makes
	// the next drag possible.
	seen := map[int]bool{}
	for id, ordinal := range ordinalsIn(t, s, one) {
		if seen[ordinal] {
			t.Errorf("%s shares ordinal %d after a restack", id, ordinal)
		}
		seen[ordinal] = true
	}
}

// A task with no ordinal at all sorts after those that have one; dropping it
// among them has to give it one.
func TestReorderingATaskWithNoOrdinal(t *testing.T) {
	if _, err := exec.LookPath("backlog"); err != nil {
		t.Skip("the backlog CLI is not installed")
	}
	one := newProject(t, "one", map[string]string{
		"task-1 - a.md": taskWith("TASK-1", "First", "To Do", "ordinal: 1000\n"),
		"task-2 - b.md": taskWith("TASK-2", "Second", "To Do", "ordinal: 2000\n"),
		"task-3 - c.md": task("TASK-3", "Unordered", "To Do"),
	})
	s := startService(t, withRegistry(t, one))

	if result := s.Reorder(one, "TASK-3", "TASK-1"); !result.OK {
		t.Fatalf("Reorder: %+v", result.Problem)
	}
	if order := orderOf(s, one); order[0] != "TASK-3" {
		t.Errorf("order is %v, want the unordered task first", order)
	}
}

func orderOf(s *app.BoardService, project string) []string {
	var out []string
	for _, view := range s.Tasks(app.QueryInput{Projects: []string{project}}) {
		out = append(out, view.ID)
	}
	return out
}

var _ = os.ReadDir
var _ = filepath.Join

// The files a task touches can be set but never emptied: Backlog.md has no
// --clear-modified-files, and --modified-file "" exits 0 having done nothing.
// A control that silently does nothing is worse than one that says why.
func TestModifiedFilesCanBeSetButNotEmptied(t *testing.T) {
	if _, err := exec.LookPath("backlog"); err != nil {
		t.Skip("the backlog CLI is not installed")
	}
	one := newProject(t, "one", map[string]string{
		"task-1 - a.md": task("TASK-1", "Subject", "To Do"),
	})
	s := startService(t, withRegistry(t, one))

	if result := s.SetModifiedFiles(one, "TASK-1", []string{"src/a.go", "src/b.go"}); !result.OK {
		t.Fatalf("SetModifiedFiles: %+v", result.Problem)
	}
	view, _ := s.Task(one, "task", "active", "TASK-1")
	if strings.Join(view.Entity.ModifiedFiles, ",") != "src/a.go,src/b.go" {
		t.Fatalf("files are %v", view.Entity.ModifiedFiles)
	}

	result := s.SetModifiedFiles(one, "TASK-1", nil)
	if result.OK {
		t.Fatal("emptying the list was reported as working")
	}
	if result.Problem == nil || !strings.Contains(result.Problem.Detail, "no") {
		t.Errorf("refused without saying why: %+v", result.Problem)
	}
	view, _ = s.Task(one, "task", "active", "TASK-1")
	if len(view.Entity.ModifiedFiles) != 2 {
		t.Errorf("the list changed anyway: %v", view.Entity.ModifiedFiles)
	}
}
